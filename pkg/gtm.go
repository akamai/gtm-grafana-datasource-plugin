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
	"net/url"
	"time"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/client-v1"
	"github.com/akamai/AkamaiOPEN-edgegrid-golang/edgegrid"
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
	switch interval {
	case FIVE_MINUTES:
		return t.Round(5 * time.Minute)
	case HOUR:
		return t.Round(time.Hour)
	default:
		log.DefaultLogger.Error("roundupTimeForInterval", "unsupported interval:", interval)
		return t
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
	return url.QueryEscape(t.Format(time.RFC3339))
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
	Title string `json:"title"`
	Type  string `json:"type"`
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

	to := time.Now()                 // now
	from := to.Add(-5 * time.Minute) // five minutes ago
	interval := Interval(FIVE_MINUTES)

	fromRounded := roundupTimeForInterval(from, interval)
	toRounded := roundupTimeForInterval(to, interval)
	openurl := createTestOpenUrl(fromRounded, toRounded, interval, "-fake-") // The URL
	log.DefaultLogger.Info("gtmOpenApiHealthCheck", "openurl", openurl)

	config := NewEdgegridConfig(clientSecret, host, accessToken, clientToken)

	// Send GET request to the OPEN API
	apireq, err := client.NewRequest(*config, "GET", openurl, nil)
	if err != nil {
		log.DefaultLogger.Error("Error creating GET request", "err", err)
		return err.Error(), backend.HealthStatusError
	}
	apiresp, err := client.Do(*config, apireq)
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

// Get data needed to populate the graph.
func gtmOpenApiQuery(zoneNamesList []string, fromRounded time.Time, toRounded time.Time, interval Interval,
	clientSecret string, host string, accessToken string, clientToken string) (*GtmDnsTrafficAllPropertiesRspDto, error) {
	reqDto := NewGtmDnsTrafficAllPropertiesReqDto(zoneNamesList)   // the POST body
	openurl := createPostOpenUrl(fromRounded, toRounded, interval) // the POST URL
	log.DefaultLogger.Info("gtmOpenApiQuery", "openurl", openurl)

	// POST to the OPEN API
	postBodyJson, err := json.Marshal(reqDto)
	if err != nil {
		log.DefaultLogger.Error("Error marshaling POST request JSON", "err", err)
		return nil, err
	}
	config := NewEdgegridConfig(clientSecret, host, accessToken, clientToken)

	apireq, err := client.NewRequest(*config, "POST", openurl, bytes.NewBuffer(postBodyJson))
	if err != nil {
		log.DefaultLogger.Error("Error creating POST request", "err", err)
		return nil, err
	}
	apiresp, err := client.Do(*config, apireq)
	if err != nil {
		log.DefaultLogger.Error("OPEN API communication error", "err", err)
		return nil, err
	}
	defer apiresp.Body.Close()
	log.DefaultLogger.Info("gtmOpenApiQuery", "Status", apiresp.Status)

	// OPEN API error response
	if apiresp.StatusCode != 200 {
		var rspDto OpenApiErrorRspDto // the expected "error" response body
		err := json.NewDecoder(apiresp.Body).Decode(&rspDto)
		if err != nil { // A JSON decode error. Not the expected body. Use the response status for the error message.
			err = errors.New(apiresp.Status)
		} else {
			err = errors.New(rspDto.Errors[0].Title) // E.g. "Some of the requested objects are unauthorized: [foo.bar.com]"
		}
		log.DefaultLogger.Info("gtmOpenApiQuery", "err", err)
		return nil, err
	}

	// OPEN API normal response
	var rspDto GtmDnsTrafficAllPropertiesRspDto // the POST response body
	json.NewDecoder(apiresp.Body).Decode(&rspDto)
	return &rspDto, nil
}
