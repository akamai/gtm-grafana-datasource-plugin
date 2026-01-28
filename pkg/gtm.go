/*
 * Copyright 2021 Akamai Technologies, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v12/pkg/edgegrid"
	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v12/pkg/session"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

// GTM "load-balancing-dns-traffic-all-properties" OPEN API documentation
// https://developer.akamai.com/api/core_features/reporting/load-balancing-dns-traffic-all-properties.html

// Akamai OPEN EdgeGrid for GoLang v1
// https://github.com/akamai/AkamaiOPEN-edgegrid-golang/

const GTM_POST_URL_FORMAT = "/reporting-api/v1/reports/load-balancing-dns-traffic-all-properties/versions/2/report-data?start=%v&end=%v&interval=%v"
const GTM_TEST_URL_FORMAT = "/reporting-api/v1/reports/load-balancing-dns-traffic-all-properties/versions/2/report-data?start=%v&end=%v&interval=%v&objectIds=%v"
const FOUR_WEEKS = 4 * 7 * 24 // four weeks as hours
const NINETY_DAYS = 90 * 24 * time.Hour

type Interval string

const (
	HOUR         Interval = "HOUR"
	FIVE_MINUTES          = "FIVE_MINUTES"
)

const (
	maxFiveMinuteRange = 48 * time.Hour
	maxHourlyRange     = 31 * 24 * time.Hour
	apiLatency         = 15 * time.Minute
)

func parseAkamaiTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

// Enforce Akamai row limits
func enforceIntervalLimits(from, to time.Time, interval Interval) (time.Time, time.Time) {
	switch interval {
	case FIVE_MINUTES:
		if to.Sub(from) > maxFiveMinuteRange {
			from = to.Add(-maxFiveMinuteRange)
		}
	case HOUR:
		if to.Sub(from) > maxHourlyRange {
			from = to.Add(-maxHourlyRange)
		}
	}
	return from, to
}

// Canonical Akamai-safe normalization
func normalizeQueryTimes(from, to time.Time, interval Interval) (time.Time, time.Time, error) {
	if from.IsZero() || to.IsZero() {
		return time.Time{}, time.Time{}, errors.New("from/to must be set")
	}

	nowSafe := time.Now().UTC().Add(-apiLatency)

	// Enforce latency
	if to.After(nowSafe) {
		to = nowSafe
	}

	// Enforce retention (90 days)
	oldest := time.Now().UTC().Add(-NINETY_DAYS)
	if to.Before(oldest) {
		return time.Time{}, time.Time{}, errors.New("time range before available data")
	}
	if from.Before(oldest) {
		from = oldest
	}

	// Enforce interval limits
	from, to = enforceIntervalLimits(from, to, interval)

	// Align to interval boundaries
	from = roundupTimeForInterval(from, interval)
	to = roundupTimeForInterval(to, interval)

	if !from.Before(to) {
		return time.Time{}, time.Time{}, errors.New("invalid time range after normalization")
	}

	return from, to, nil
}

func calculateInterval(from time.Time, to time.Time, maxDataPoints uint) Interval {
	// Must use HOUR interval for time ranges over 4 weeks.
	timeRangeHours := uint(to.Sub(from).Hours())
	if timeRangeHours > FOUR_WEEKS {
		return HOUR
	}

	// If there are enough 1-hour datapoints to fill the graph then use HOUR
	if timeRangeHours >= maxDataPoints {
		return HOUR
	}

	// Else use FIVE_MINUTES
	return FIVE_MINUTES
}

// GTM OPEN API insists that start and end times must be on interval boundaries.
func roundupTimeForInterval(t time.Time, interval Interval) time.Time {
	// 1. Force to UTC immediately
	t = t.UTC()

	switch interval {
	case FIVE_MINUTES:
		// Truncate(5m) turns 13:54:32 into 13:50:00 (Perfectly clean)
		return t.Truncate(5 * time.Minute)
	case HOUR:
		// Truncate(1h) turns 13:54:32 into 13:00:00
		return t.Truncate(time.Hour)
	default:
		return t.Truncate(time.Minute)
	}
}

// Is the time before the oldest available data?
func timeBeforeOldestData(t time.Time, oldestDataTime time.Time) bool {
	return t.Before(oldestDataTime)
}

// Start time cannot be before the oldest available data.  If it it, fix it.
func limitTimeToOldestData(timeRounded time.Time, oldestDataTime time.Time) time.Time {
	if timeRounded.Before(oldestDataTime) {
		return oldestDataTime
	}
	return timeRounded
}

// Adjust the start (from) and end (to) times
func adjustQueryTimes(from time.Time, to time.Time, interval Interval) (time.Time, time.Time, error) {

	if from.IsZero() || to.IsZero() {
		err := errors.New("from or to time is not set (zero time)")
		log.DefaultLogger.Warn("adjustQueryTimes", "from", from, "to", to, "err", err)
		return time.Time{}, time.Time{}, err
	}

	fromRounded := roundupTimeForInterval(from, interval)
	toRounded := roundupTimeForInterval(to, interval)

	// Data is available from the OPEN API for 90 days
	ninetyDaysAgo := roundupTimeForInterval(time.Now().Add(-NINETY_DAYS), interval)

	// Is the 'to' (end) time before data is available?  If so, that's an error.
	if timeBeforeOldestData(toRounded, ninetyDaysAgo) {
		err := errors.New("Time range is before available data")
		log.DefaultLogger.Info("adjustQueryTimes", "err", err)
		return fromRounded, toRounded, err
	}

	// Limit the 'from' (start) time to when the oldest data is available.
	fromLimited := limitTimeToOldestData(fromRounded, ninetyDaysAgo)

	// Returned the fixed 'to' and 'from' times.
	return fromLimited, toRounded, nil
}

// The time format required by OPEN API
func openApiUrlTimeFormat(t time.Time) string {
	return url.QueryEscape(t.UTC().Format("2006-01-02T15:04:00Z"))
}

// OPEN API URLs
func createPostOpenUrl(fromRounded time.Time, toRounded time.Time, interval Interval) string {
	return fmt.Sprintf(GTM_POST_URL_FORMAT, openApiUrlTimeFormat(fromRounded), openApiUrlTimeFormat(toRounded), interval)
}

func createTestOpenUrl(fromRounded time.Time, toRounded time.Time, interval Interval, zone string) string {
	return fmt.Sprintf(GTM_TEST_URL_FORMAT, openApiUrlTimeFormat(fromRounded), openApiUrlTimeFormat(toRounded), interval, zone)
}

// EdgeGrid configuration structure constructor
func NewEdgegridConfig(clientSecret string, host string, accessToken string, clientToken string) *edgegrid.Config {
	return &edgegrid.Config{
		ClientSecret: clientSecret,
		Host:         host,
		AccessToken:  accessToken,
		ClientToken:  clientToken,
		MaxBody:      131072,
		Debug:        false,
	}
}

// OPEN API REQUEST

// Example request bodies:
// {"objectType": "fpdomain", "objectIds": ["akamccare.akadns.net"], "metrics": ["startdatetime", "hits"]}

// OPEN API request body contructor
func NewGtmDnsTrafficAllPropertiesReqDto(zoneName []string) *GtmDnsTrafficAllPropertiesReqDto {
	return &GtmDnsTrafficAllPropertiesReqDto{
		ObjectType: "fpdomain",
		ObjectIds:  zoneName,
		Metrics:    []string{"startdatetime", "hits"},
	}
}

type GtmDnsTrafficAllPropertiesReqDto struct {
	ObjectType string   `json:"objectType"`
	ObjectIds  []string `json:"objectIds"`
	Metrics    []string `json:"metrics"`
}

// OPEN API NORMAL RESPONSE

type Datum struct {
	StartDateTime string `json:"startdatetime"`
	Hits          string `json:"hits"`
}

type Metadata struct {
	AvailableDataEnds string   `json:"availableDataEnds"`
	End               string   `json:"end"`
	Interval          string   `json:"interval"`
	Name              string   `json:"name"`
	ObjectIds         []string `json:"objectIds"`
	ObjectType        string   `json:"objectType"`
	OutputType        string   `json:"outputType"`
	RowCount          int      `json:"rowCount"`
	Start             string   `json:"start"`
	Version           string   `json:"version"`
}

type GtmDnsTrafficAllPropertiesRspDto struct {
	Data     []Datum  `json:"data"`
	Metadata Metadata `json:"metadata"`
	// `json:"summaryStatistics"`
}

// OPEN API ERROR RESPONSE

type Error struct {
	Title  string `json:"title"`
	Type   string `json:"type"`
	Detail string `json:"detail"`
	//AvailableStartDate string `json:"availableStartDate"`
	//AvailableEndDate   string `json:"availableEndDate"`
	AvailableStartDate string `json:"availableStartDate"`
	AvailableEndDate   string `json:"availableEndDate"`

	RequestedStartDate string `json:"requestedStartDate"`
	RequestedEndDate   string `json:"requestedEndDate"`
}

type OpenApiErrorRspDto struct {
	Errors   []Error `json:"errors"`
	Instance string  `json:"instance"`
	Title    string  `json:"title"`
	Type     string  `json:"type"`
}

// OPEN API REQUEST METHODS

// Verify that the datasource can reach the OPEN API
func gtmOpenApiHealthCheck(clientSecret string, host string, accessToken string, clientToken string) (string, backend.HealthStatus) {

	to := time.Now().Add(-24 * time.Hour) // now
	from := to.Add(-1 * time.Hour)        // five minutes ago
	interval := Interval(FIVE_MINUTES)

	fromRounded := roundupTimeForInterval(from, interval)
	toRounded := roundupTimeForInterval(to, interval)
	openurl := createTestOpenUrl(fromRounded, toRounded, interval, "-fake-") // The URL
	log.DefaultLogger.Info("gtmOpenApiHealthCheck", "openurl", openurl)

	config := NewEdgegridConfig(clientSecret, host, accessToken, clientToken)

	sess, err := session.New(
		session.WithSigner(config),
		session.WithHTTPTracing(true),
	)

	if err != nil {
		log.DefaultLogger.Error("Error creating session", "err", err)
		return err.Error(), backend.HealthStatusError
	}
	// Send GET request to the OPEN API
	apireq, err := http.NewRequest(http.MethodGet, openurl, nil)
	if err != nil {
		log.DefaultLogger.Error("Error creating GET request", "err", err)
		return err.Error(), backend.HealthStatusError
	}
	apiresp, err := sess.Exec(apireq, nil)
	if err != nil {
		log.DefaultLogger.Error("OPEN API communication error", "err", err)
		return err.Error(), backend.HealthStatusError
	}

	log.DefaultLogger.Info("gtmOpenApiHealthCheck", "Status (403 expected)", apiresp.Status)

	// 403 Forbidden is expected because -test- is not a valid zone name.

	// Not a 403 response: datasource failed.
	if apiresp.StatusCode != 403 {
		var rspDto OpenApiErrorRspDto
		err := json.NewDecoder(apiresp.Body).Decode(&rspDto)
		msg := "Unexpected status code. Datasource failed: "
		if err != nil { // A JSON decode error. Not the expected body. Use the response status for the error message.
			msg += apiresp.Status
		} else {
			msg += rspDto.Errors[0].Title
		}
		log.DefaultLogger.Error("gtmOpenApiTest", "msg", msg)
		return msg, backend.HealthStatusError // RETURN
	}

	// 403 response
	var rspDto OpenApiErrorRspDto
	err = json.NewDecoder(apiresp.Body).Decode(&rspDto)

	// 403 response but not the expected body: datasource failed.
	if err != nil {
		msg := "Unexpected response format. Datasource failed: " + apiresp.Status
		log.DefaultLogger.Error("gtmOpenApiTest", "msg", msg)
		return msg, backend.HealthStatusError // RETURN
	}

	// 403 response with the expected body
	errorTitle := rspDto.Errors[0].Title

	// 403 response but not the expected error: datasource failed.
	if errorTitle != "Some of the requested objects are unauthorized: [-fake-]" {
		msg := "Unexpected error type. Datasource failed: " + errorTitle
		log.DefaultLogger.Error("gtmOpenApiTest", "msg", msg)
		return msg, backend.HealthStatusError // RETURN
	}

	return "Data source is working", backend.HealthStatusOk
}

func gtmOpenApiQuery(
	zoneNamesList []string,
	from time.Time,
	to time.Time,
	interval Interval,
	clientSecret string,
	host string,
	accessToken string,
	clientToken string,
) (*GtmDnsTrafficAllPropertiesRspDto, error) {

	fromRounded, toRounded, err := normalizeQueryTimes(from, to, interval)
	if err != nil {
		log.DefaultLogger.Error("Time normalization failed", "err", err)
		return nil, err
	}

	reqDto := NewGtmDnsTrafficAllPropertiesReqDto(zoneNamesList)
	openurl := createPostOpenUrl(fromRounded, toRounded, interval)

	log.DefaultLogger.Info(
		"gtmOpenApiQuery",
		"start", fromRounded,
		"end", toRounded,
		"interval", interval,
		"url", openurl,
	)

	postBodyJson, err := json.Marshal(reqDto)
	if err != nil {
		return nil, err
	}

	config := NewEdgegridConfig(clientSecret, host, accessToken, clientToken)
	sess, err := session.New(
		session.WithSigner(config),
		session.WithHTTPTracing(true),
	)
	if err != nil {
		return nil, err
	}

	apireq, err := http.NewRequest(http.MethodPost, openurl, bytes.NewBuffer(postBodyJson))
	if err != nil {
		return nil, err
	}

	apiresp, err := sess.Exec(apireq, nil)
	if err != nil {
		return nil, fmt.Errorf("OPEN API communication error: %w", err)
	}
	defer apiresp.Body.Close()

	bodyBytes, err := io.ReadAll(apiresp.Body)
	if err != nil {
		return nil, err
	}

	if apiresp.StatusCode != http.StatusOK {
		var errDto OpenApiErrorRspDto
		_ = json.Unmarshal(bodyBytes, &errDto)

		if len(errDto.Errors) > 0 {
			errDetail := errDto.Errors[0]

			if errDetail.AvailableEndDate != "" {
				availableEnd, err := parseAkamaiTime(errDetail.AvailableEndDate)
				if err == nil && toRounded.After(availableEnd) {

					timeGap := toRounded.Sub(availableEnd)

					newFrom := fromRounded.Add(-timeGap)

					log.DefaultLogger.Warn(
						"Clamping query and shifting window",
						"oldEnd", toRounded,
						"newEnd", availableEnd,
						"newFrom", newFrom,
					)

					// Re-run query with the whole window moved back
					return gtmOpenApiQuery(
						zoneNamesList,
						newFrom,
						availableEnd,
						interval,
						clientSecret,
						host,
						accessToken,
						clientToken,
					)
				}
			}

			return nil, fmt.Errorf(
				"Akamai error: %s",
				errDetail.Detail,
			)
		}

		return nil, fmt.Errorf("API error: %s", apiresp.Status)
	}

	var rspDto GtmDnsTrafficAllPropertiesRspDto
	if err := json.Unmarshal(bodyBytes, &rspDto); err != nil {
		return nil, err
	}

	log.DefaultLogger.Info(
		"gtmOpenApiQuery success",
		"rows", rspDto.Metadata.RowCount,
	)

	return &rspDto, nil
}
