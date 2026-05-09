// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package kibanaoapi

import (
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/stretchr/testify/require"
)

// kbapiTestArtifacts is a test helper; nested field name matches generated kbapi.
//
//nolint:revive // var-naming: `Id` matches kbapi SLOsArtifacts OpenAPI
func kbapiTestArtifacts(dashboardID string) *kbapi.SLOsArtifacts {
	return &kbapi.SLOsArtifacts{Dashboards: &[]struct {
		Id string `json:"id"`
	}{{Id: dashboardID}}}
}

func makeApmAvailabilityIndicator(t *testing.T) kbapi.SLOsSloWithSummaryResponse_Indicator {
	t.Helper()
	ind := kbapi.SLOsIndicatorPropertiesApmAvailability{
		Type: "sli.apm.transactionErrorRate",
		Params: struct {
			Environment     string  `json:"environment"`
			Filter          *string `json:"filter,omitempty"`
			Index           string  `json:"index"`
			Service         string  `json:"service"`
			TransactionName string  `json:"transactionName"`
			TransactionType string  `json:"transactionType"`
		}{
			Service:         "slo-service",
			Environment:     "slo-environment",
			TransactionType: "slo-transaction-type",
			TransactionName: "slo-transaction-name",
			Index:           "slo-index",
		},
	}
	var result kbapi.SLOsSloWithSummaryResponse_Indicator
	require.NoError(t, result.FromSLOsIndicatorPropertiesApmAvailability(ind))
	return result
}

func makeApmLatencyIndicator(t *testing.T) kbapi.SLOsSloWithSummaryResponse_Indicator {
	t.Helper()
	ind := kbapi.SLOsIndicatorPropertiesApmLatency{
		Type: "sli.apm.transactionDuration",
		Params: struct {
			Environment     string  `json:"environment"`
			Filter          *string `json:"filter,omitempty"`
			Index           string  `json:"index"`
			Service         string  `json:"service"`
			Threshold       float64 `json:"threshold"`
			TransactionName string  `json:"transactionName"`
			TransactionType string  `json:"transactionType"`
		}{
			Service:         "svc",
			Environment:     "prod",
			TransactionType: "request",
			TransactionName: "GET /",
			Index:           "metrics-*",
			Threshold:       200,
		},
	}
	var result kbapi.SLOsSloWithSummaryResponse_Indicator
	require.NoError(t, result.FromSLOsIndicatorPropertiesApmLatency(ind))
	return result
}

func makeCustomKqlIndicator(t *testing.T) kbapi.SLOsSloWithSummaryResponse_Indicator {
	t.Helper()
	var ind kbapi.SLOsIndicatorPropertiesCustomKql
	require.NoError(t, json.Unmarshal([]byte(`{
		"type": "sli.kql.custom",
		"params": {"index":"logs-*","timestampField":"@timestamp","good":"status:200","total":"*"}
	}`), &ind))
	var result kbapi.SLOsSloWithSummaryResponse_Indicator
	require.NoError(t, result.FromSLOsIndicatorPropertiesCustomKql(ind))
	return result
}

func makeCustomMetricIndicator(t *testing.T) kbapi.SLOsSloWithSummaryResponse_Indicator {
	t.Helper()
	var ind kbapi.SLOsIndicatorPropertiesCustomMetric
	require.NoError(t, json.Unmarshal([]byte(`{
		"type": "sli.metric.custom",
		"params": {
			"index": "metrics-*",
			"timestampField": "@timestamp",
			"good":  {"equation":"A","metrics":[{"aggregation":"sum","field":"good_field","name":"A"}]},
			"total": {"equation":"A","metrics":[{"aggregation":"sum","field":"total_field","name":"A"}]}
		}
	}`), &ind))
	var result kbapi.SLOsSloWithSummaryResponse_Indicator
	require.NoError(t, result.FromSLOsIndicatorPropertiesCustomMetric(ind))
	return result
}

func makeHistogramIndicator(t *testing.T) kbapi.SLOsSloWithSummaryResponse_Indicator {
	t.Helper()
	var ind kbapi.SLOsIndicatorPropertiesHistogram
	require.NoError(t, json.Unmarshal([]byte(`{
		"type": "sli.histogram.custom",
		"params": {
			"index": "metrics-*",
			"timestampField": "@timestamp",
			"good":  {"aggregation":"value_count","field":"latency"},
			"total": {"aggregation":"value_count","field":"latency"}
		}
	}`), &ind))
	var result kbapi.SLOsSloWithSummaryResponse_Indicator
	require.NoError(t, result.FromSLOsIndicatorPropertiesHistogram(ind))
	return result
}

func makeTimesliceMetricIndicator(t *testing.T) kbapi.SLOsSloWithSummaryResponse_Indicator {
	t.Helper()
	var ind kbapi.SLOsIndicatorPropertiesTimesliceMetric
	require.NoError(t, json.Unmarshal([]byte(`{
		"type": "sli.metric.timeslice",
		"params": {
			"index": "metrics-*",
			"timestampField": "@timestamp",
			"metric": {
				"comparator": "GT",
				"equation": "A",
				"metrics": [{"aggregation":"avg","field":"latency","name":"A"}],
				"threshold": 100
			}
		}
	}`), &ind))
	var result kbapi.SLOsSloWithSummaryResponse_Indicator
	require.NoError(t, result.FromSLOsIndicatorPropertiesTimesliceMetric(ind))
	return result
}

func Test_ResponseIndicatorToCreateIndicator(t *testing.T) {
	tests := []struct {
		name      string
		indicator kbapi.SLOsSloWithSummaryResponse_Indicator
		wantErr   bool
	}{
		{name: "apm availability", indicator: makeApmAvailabilityIndicator(t)},
		{name: "apm latency", indicator: makeApmLatencyIndicator(t)},
		{name: "custom kql", indicator: makeCustomKqlIndicator(t)},
		{name: "custom metric", indicator: makeCustomMetricIndicator(t)},
		{name: "histogram", indicator: makeHistogramIndicator(t)},
		{name: "timeslice metric", indicator: makeTimesliceMetricIndicator(t)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResponseIndicatorToCreateIndicator(tt.indicator)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				b, jsonErr := got.MarshalJSON()
				require.NoError(t, jsonErr)
				require.NotEmpty(t, b)
			}
		})
	}
}

func Test_ResponseIndicatorToUpdateIndicator(t *testing.T) {
	tests := []struct {
		name      string
		indicator kbapi.SLOsSloWithSummaryResponse_Indicator
		wantErr   bool
	}{
		{name: "apm availability", indicator: makeApmAvailabilityIndicator(t)},
		{name: "apm latency", indicator: makeApmLatencyIndicator(t)},
		{name: "custom kql", indicator: makeCustomKqlIndicator(t)},
		{name: "custom metric", indicator: makeCustomMetricIndicator(t)},
		{name: "histogram", indicator: makeHistogramIndicator(t)},
		{name: "timeslice metric", indicator: makeTimesliceMetricIndicator(t)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResponseIndicatorToUpdateIndicator(tt.indicator)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				b, jsonErr := got.MarshalJSON()
				require.NoError(t, jsonErr)
				require.NotEmpty(t, b)
			}
		})
	}
}

func Test_SloResponseToModel(t *testing.T) {
	syncDelay := "2m"

	tests := []struct {
		name          string
		spaceID       string
		sloResponse   *kbapi.SLOsSloWithSummaryResponse
		expectedModel *models.Slo
	}{
		{
			name:    "should return a model with the correct values",
			spaceID: "space-id",
			sloResponse: &kbapi.SLOsSloWithSummaryResponse{
				Id:              "slo-id",
				Name:            "slo-name",
				Description:     "slo-description",
				Indicator:       makeApmAvailabilityIndicator(t),
				TimeWindow:      kbapi.SLOsTimeWindow{Duration: "7d", Type: "rolling"},
				BudgetingMethod: "occurrences",
				Settings: kbapi.SLOsSettings{
					SyncDelay: &syncDelay,
				},
			},
			expectedModel: &models.Slo{
				SloID:           "slo-id",
				SpaceID:         "space-id",
				Name:            "slo-name",
				Description:     "slo-description",
				Indicator:       makeApmAvailabilityIndicator(t),
				TimeWindow:      kbapi.SLOsTimeWindow{Duration: "7d", Type: "rolling"},
				BudgetingMethod: "occurrences",
				Settings:        &kbapi.SLOsSettings{SyncDelay: &syncDelay},
				GroupBy:         nil,
			},
		},
		{
			name:    "should return tags if available",
			spaceID: "space-id",
			sloResponse: &kbapi.SLOsSloWithSummaryResponse{
				Id:              "slo-id",
				Name:            "slo-name",
				Description:     "slo-description",
				Indicator:       makeApmAvailabilityIndicator(t),
				TimeWindow:      kbapi.SLOsTimeWindow{Duration: "7d", Type: "rolling"},
				BudgetingMethod: "occurrences",
				Settings: kbapi.SLOsSettings{
					SyncDelay: &syncDelay,
				},
				Tags: []string{"tag-1", "another_tag"},
			},
			expectedModel: &models.Slo{
				SloID:           "slo-id",
				SpaceID:         "space-id",
				Name:            "slo-name",
				Description:     "slo-description",
				Indicator:       makeApmAvailabilityIndicator(t),
				TimeWindow:      kbapi.SLOsTimeWindow{Duration: "7d", Type: "rolling"},
				BudgetingMethod: "occurrences",
				Settings:        &kbapi.SLOsSettings{SyncDelay: &syncDelay},
				Tags:            []string{"tag-1", "another_tag"},
				GroupBy:         nil,
			},
		},
		{
			name:          "nil response should return a nil model",
			spaceID:       "space-id",
			sloResponse:   nil,
			expectedModel: nil,
		},
		{
			name:    "maps artifacts from get SLO",
			spaceID: "space-id",
			sloResponse: &kbapi.SLOsSloWithSummaryResponse{
				Id:              "slo-id",
				Name:            "slo-name",
				Description:     "slo-description",
				Indicator:       makeApmAvailabilityIndicator(t),
				TimeWindow:      kbapi.SLOsTimeWindow{Duration: "7d", Type: "rolling"},
				BudgetingMethod: "occurrences",
				Settings:        kbapi.SLOsSettings{SyncDelay: &syncDelay},
				Artifacts:       kbapiTestArtifacts("dashboard-1"),
			},
			expectedModel: &models.Slo{
				SloID:           "slo-id",
				SpaceID:         "space-id",
				Name:            "slo-name",
				Description:     "slo-description",
				Indicator:       makeApmAvailabilityIndicator(t),
				TimeWindow:      kbapi.SLOsTimeWindow{Duration: "7d", Type: "rolling"},
				BudgetingMethod: "occurrences",
				Settings:        &kbapi.SLOsSettings{SyncDelay: &syncDelay},
				Artifacts:       kbapiTestArtifacts("dashboard-1"),
				GroupBy:         nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := SloResponseToModel(tt.spaceID, tt.sloResponse)
			require.Equal(t, tt.expectedModel, model)
		})
	}
}
