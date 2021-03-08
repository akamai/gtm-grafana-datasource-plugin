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
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// The datasource front-end sends zonenames (to graph) as a string. OPEN API POST request needs a zonename list.
func zonesListFromZones(zoneNames string) []string {
        zoneNames = strings.Replace(zoneNames, " ", "", -1) // remove spaces
        rawList := strings.Split(zoneNames, ",")            // split on commas, may contain empty entries

        var cleanList []string
        for _, z := range rawList {
                if len(z) > 0 {
                        cleanList = append(cleanList, z)
                }
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
	DataSourceId   uint               `json:"dataSourceId"`
	IntervalMs     uint               `json:"intervalMs"`
	MaxDataPoints  uint               `json:"maxDataPoints"`
	ZoneNames      string             `json:"zoneNames"`
	MetricName     string             `json:"metricName"`
}

// Grafana structures and functions
func newDataSourceInstance(setting backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	return &instanceSettings{
		httpClient: &http.Client{},
	}, nil
}

type instanceSettings struct {
	httpClient *http.Client
}

// Called before creating a new instance to allow plugin to cleanup.
func (s *instanceSettings) Dispose() {
}

func newDatasource() datasource.ServeOpts {
	// Creates a instance manager for the plugin. The function passed
	// into `NewInstanceManger` is called when the instance is created
	// for the first time or when datasource configuration changes.
	im := datasource.NewInstanceManager(newDataSourceInstance)

	ds := &AkamaiEdgeDnsDatasource{
		im: im,
	}

	return datasource.ServeOpts{
		QueryDataHandler:    ds,
		CheckHealthHandler:  ds,
	}
}

type AkamaiEdgeDnsDatasource struct {
	im instancemgmt.InstanceManager
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
		res := td.query(ctx, q, dss)

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

func (td *AkamaiEdgeDnsDatasource) query(ctx context.Context, query backend.DataQuery, dss dataSourceSettingsJson) backend.DataResponse {
	// log.DefaultLogger.Info("QueryData", "clientSecret", dss.ClientSecret)
	// log.DefaultLogger.Info("QueryData", "host", dss.Host)
	// log.DefaultLogger.Info("QueryData", "accessToken", dss.AccessToken)
	// log.DefaultLogger.Info("QueryData", "clientToken", dss.ClientToken)

	log.DefaultLogger.Info("QueryData", "RefID", query.RefID)

	response := backend.DataResponse{}

	// Unmarshal the (query request input) json into the 'dataQueryJson' structure
	var dqj dataQueryJson
	response.Error = json.Unmarshal(query.JSON, &dqj)
	if response.Error != nil {
		return response
	}

	log.DefaultLogger.Info("query", "query.TimeRange.From", query.TimeRange.From)
	log.DefaultLogger.Info("query", "query.TimeRange.To", query.TimeRange.To)
	log.DefaultLogger.Info("query", "maxDataPoints", dqj.MaxDataPoints)
	log.DefaultLogger.Info("query", "zoneNames", dqj.ZoneNames)
	log.DefaultLogger.Info("query", "metricName", dqj.MetricName)

	// If ZoneNames is empty then ignore the query
	if len(dqj.ZoneNames) == 0 {
		response.Error = errors.New("Enter zone names")
		return response

	}

	// 'interval' and fixed-up 'from' and 'to' times are needed to make the OPEN API POST URL
	interval := calculateInterval(query.TimeRange.From, query.TimeRange.To, dqj.MaxDataPoints)
	fromRounded, toRounded, err := adjustQueryTimes(query.TimeRange.From, query.TimeRange.To, interval)
	if err != nil {
		response.Error = err
		return response
	}

	// 'zoneNamesList' is needed for the OPEN API POST body
	zoneNamesList := zonesListFromZones(dqj.ZoneNames)
	if len(zoneNamesList) == 0 {
		response.Error = errors.New("Enter one zone name")
		return response
	}

	// The OPEN API returns the data to graph.
	openApiRspDto, err := gtmOpenApiQuery(zoneNamesList, fromRounded, toRounded, interval, dss.ClientSecret, dss.Host, dss.AccessToken, dss.ClientToken)
	if err != nil {
		response.Error = err
		return response
	}

	// The number of datapoints in the response
	numDataRows := len(openApiRspDto.Data)
	log.DefaultLogger.Info("query", "numDataRows", numDataRows)

	// Create slices that will be added to the dataframe.
	sampletime := make([]time.Time, numDataRows)
	hitspersec := make([]float64, numDataRows)

	// The response contains data for 'hits'.

	// Loop through the OPEN API response. Put data items into the dataframe slices.
	for i, datum := range openApiRspDto.Data {
		unixms, err := strconv.ParseInt(datum.StartDateTime, 10, 64)
		if err != nil {
			log.DefaultLogger.Error("Error parsing time", "err", err)
			response.Error = err
			return response
		}
		sampletime[i] = time.Unix(unixms/1000, 0)

		// Ignore the error. Some data will be "N/A", in which case hits will be zero.
		hitspersec[i], _ = strconv.ParseFloat(datum.Hits, 64)
	}

	// Create the response data frame.
	frame := data.NewFrame("response")

	// If the user configured a metric name then use that. Else generate a metric name.
	metricName := dqj.MetricName
	if len(metricName) == 0 {
		// Metric name not configured. Create the default name.
		metricName = dqj.ZoneNames + " hits" 
	}

	// Add data to the response data frame.
	frame.Fields = append(frame.Fields, data.NewField("time", nil, sampletime))     // add the time dimension to dataframe
	frame.Fields = append(frame.Fields, data.NewField(metricName, nil, hitspersec)) // add values to dataframe

	// Add the dataframe to the response
	response.Frames = append(response.Frames, frame)

	return response
}

// The 'Save & Test' button on the datasource configuration page allows users to verify that the datasource is working as expected.
func (td *AkamaiEdgeDnsDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	// log.DefaultLogger.Info("CheckHealth", "clientSecret", ds.ClientSecret)
	// log.DefaultLogger.Info("CheckHealth", "host", ds.Host)
	// log.DefaultLogger.Info("CheckHealth", "accessToken", ds.AccessToken)
	// log.DefaultLogger.Info("CheckHealth", "clientToken", ds.ClientToken)

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

