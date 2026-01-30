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
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// The datasource front-end sends domainnames (to graph) as a string. OPEN API POST request needs a domainname list.
func domainListFromDomain(domainName string) []string {
	domainName = strings.Replace(domainName, " ", "", -1) // remove spaces
	domainName = strings.Replace(domainName, ",", "", -1) // remove commas

	var cleanList []string
	if len(domainName) > 0 {
		cleanList = append(cleanList, domainName)
	}
	return cleanList
}

// The datasource configuration supplied by the front-end.
type dataSourceSettingsJson struct {
	ClientSecret string `json:"clientSecret"`
	Host         string `json:"host"`
	AccessToken  string `json:"accessToken"`
	ClientToken  string `json:"clientToken"`
}

// Query information supplied by the front-end
type dataQueryJson struct {
	DataSourceId  uint   `json:"dataSourceId"`
	IntervalMs    uint   `json:"intervalMs"`
	MaxDataPoints uint   `json:"maxDataPoints"`
	DomainName    string `json:"domainName"`
	MetricName    string `json:"metricName"`
}

// Grafana structures and functions
func newDataSourceInstance(ctx context.Context, setting backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	return &AkamaiEdgeDnsDatasource{
		httpClient: &http.Client{},
	}, nil
}

type instanceSettings struct {
	httpClient *http.Client
}

// Called before creating a new instance to allow plugin to cleanup.
func (s *instanceSettings) Dispose() {
}

type AkamaiEdgeDnsDatasource struct {
	im         instancemgmt.InstanceManager
	httpClient *http.Client
}

// QueryData handles multiple queries and returns multiple responses.
// 'req' contains the queries []DataQuery (where each query contains RefID as a unique identifer).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response contains Frames ([]*Frame).
func (td *AkamaiEdgeDnsDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {

	// create response struct
	response := backend.NewQueryDataResponse()

	log.DefaultLogger.Info("QueryData", "Login", req.PluginContext.User.Login)
	log.DefaultLogger.Info("QueryData", "Role", req.PluginContext.User.Role)

	var dss dataSourceSettingsJson
	err := json.Unmarshal(req.PluginContext.DataSourceInstanceSettings.JSONData, &dss)
	if err != nil {
		return response, err
	}

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res, err := td.query(ctx, q, dss)
		if err != nil {
			// Create an error frame so the error shows in the panel
			errorFrame := data.NewFrame("Error")
			errorFrame.Meta = &data.FrameMeta{
				Type: data.FrameTypeTable,
			}
			errorFrame.Fields = append(errorFrame.Fields,
				data.NewField("Error", nil, []string{err.Error()}),
			)

			response.Responses[q.RefID] = backend.DataResponse{
				Frames: data.Frames{errorFrame},
				Error:  err,
			}
		} else {
			response.Responses[q.RefID] = *res
		}
	}

	return response, nil
}

func errorFrame(msg string) *backend.DataResponse {
	frame := data.NewFrame("Error")
	frame.Fields = append(frame.Fields, data.NewField("message", nil, []string{msg}))
	frame.Meta = &data.FrameMeta{
		Type: data.FrameTypeTable,
	}

	return &backend.DataResponse{
		Frames: data.Frames{frame},
		Error:  errors.New(msg),
	}
}

func (td *AkamaiEdgeDnsDatasource) query(ctx context.Context, query backend.DataQuery, dss dataSourceSettingsJson) (*backend.DataResponse, error) {
	log.DefaultLogger.Info("QueryData", "RefID", query.RefID)

	// Unmarshal the query JSON into your struct
	var dqj dataQueryJson
	if err := json.Unmarshal(query.JSON, &dqj); err != nil {
		return errorFrame("Failed to parse query JSON: " + err.Error()), nil
	}

	log.DefaultLogger.Info("query", "query.TimeRange.From", query.TimeRange.From)
	log.DefaultLogger.Info("query", "query.TimeRange.To", query.TimeRange.To)
	log.DefaultLogger.Info("query", "maxDataPoints", dqj.MaxDataPoints)
	log.DefaultLogger.Info("query", "domainName", dqj.DomainName)
	log.DefaultLogger.Info("query", "metricName", dqj.MetricName)

	// Validate domain name presence
	if len(dqj.DomainName) == 0 {
		return errorFrame("Please enter a domain name"), nil
	}

	// Calculate interval and adjust times
	interval := calculateInterval(query.TimeRange.From, query.TimeRange.To, dqj.MaxDataPoints)
	fromRounded, toRounded, err := adjustQueryTimes(query.TimeRange.From, query.TimeRange.To, interval)
	if err != nil {
		return errorFrame("Failed to adjust query times: " + err.Error()), nil
	}

	domainNameList := domainListFromDomain(dqj.DomainName)
	if len(domainNameList) == 0 {
		return errorFrame("Enter at least one valid domain name"), nil
	}

	// Query the OPEN API
	openApiRspDto, err := gtmOpenApiQuery(domainNameList, fromRounded, toRounded, interval, dss.ClientSecret, dss.Host, dss.AccessToken, dss.ClientToken)
	if err != nil {
		return errorFrame("Failed to fetch data from API: " + err.Error()), nil
	}

	numDataRows := len(openApiRspDto.Data)
	log.DefaultLogger.Info("query", "numDataRows", numDataRows)

	sampletime := make([]time.Time, numDataRows)
	hitspersec := make([]float64, numDataRows)

	for i, datum := range openApiRspDto.Data {
		unixms, err := strconv.ParseInt(datum.StartDateTime, 10, 64)
		if err != nil {
			log.DefaultLogger.Error("Error parsing time", "err", err)
			return errorFrame("Invalid time format in API response: " + err.Error()), nil
		}
		sampletime[i] = time.Unix(unixms/1000, 0)
		hitspersec[i], _ = strconv.ParseFloat(datum.Hits, 64) // ignore parse errors, treat as zero
	}

	frame := data.NewFrame("response")
	metricName := dqj.MetricName
	if len(metricName) == 0 {
		metricName = dqj.DomainName + " hits"
	}

	frame.Fields = append(frame.Fields, data.NewField("time", nil, sampletime))
	frame.Fields = append(frame.Fields, data.NewField(metricName, nil, hitspersec))

	response := &backend.DataResponse{}
	response.Frames = append(response.Frames, frame)

	return response, nil
}

// The 'Save & Test' button on the datasource configuration page allows users to verify that the datasource is working as expected.
func (td *AkamaiEdgeDnsDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {

	var ds dataSourceSettingsJson
	err := json.Unmarshal(req.PluginContext.DataSourceInstanceSettings.JSONData, &ds)
	if err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusUnknown,
			Message: "Internal error. Failed to unmarshal healthcheck JSON",
		}, err
	}

	// Verify that the OPEN API responds.
	message, status := gtmOpenApiHealthCheck(ds.ClientSecret, ds.Host, ds.AccessToken, ds.ClientToken)

	return &backend.CheckHealthResult{
		Status:  status,
		Message: message,
	}, nil
}
