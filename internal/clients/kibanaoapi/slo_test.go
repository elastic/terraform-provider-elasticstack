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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/stretchr/testify/require"
)

const testTimestampField = "@timestamp"

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

func makeKqlGood(t *testing.T, kql string) kbapi.SLOsKqlWithFiltersGood {
	t.Helper()
	var g kbapi.SLOsKqlWithFiltersGood
	require.NoError(t, g.FromSLOsKqlWithFiltersGood0(kql))
	return g
}

func makeKqlTotal(t *testing.T, kql string) kbapi.SLOsKqlWithFiltersTotal {
	t.Helper()
	var to kbapi.SLOsKqlWithFiltersTotal
	require.NoError(t, to.FromSLOsKqlWithFiltersTotal0(kql))
	return to
}

func makeApmAvail(t *testing.T) kbapi.SLOsIndicatorPropertiesApmAvailability {
	t.Helper()
	return kbapi.SLOsIndicatorPropertiesApmAvailability{
		Type: "sli.apm.transactionErrorRate",
		Params: struct {
			Environment     string  `json:"environment"`
			Filter          *string `json:"filter,omitempty"`
			Index           string  `json:"index"`
			Service         string  `json:"service"`
			TransactionName string  `json:"transactionName"`
			TransactionType string  `json:"transactionType"`
		}{Service: "s", Environment: "e", TransactionType: "t", TransactionName: "n", Index: "i"},
	}
}

func makeApmLatency(t *testing.T) kbapi.SLOsIndicatorPropertiesApmLatency {
	t.Helper()
	return kbapi.SLOsIndicatorPropertiesApmLatency{
		Type: "sli.apm.transactionDuration",
		Params: struct {
			Environment     string  `json:"environment"`
			Filter          *string `json:"filter,omitempty"`
			Index           string  `json:"index"`
			Service         string  `json:"service"`
			Threshold       float64 `json:"threshold"`
			TransactionName string  `json:"transactionName"`
			TransactionType string  `json:"transactionType"`
		}{Service: "s", Environment: "e", TransactionType: "t", TransactionName: "n", Index: "i", Threshold: 100},
	}
}

func makeCustomKql(t *testing.T) kbapi.SLOsIndicatorPropertiesCustomKql {
	t.Helper()
	ind := kbapi.SLOsIndicatorPropertiesCustomKql{Type: "sli.kql.custom"}
	ind.Params.Index = "i"
	ind.Params.TimestampField = testTimestampField
	ind.Params.Good = makeKqlGood(t, "status:200")
	ind.Params.Total = makeKqlTotal(t, "status:*")
	return ind
}

func makeCustomMetric(t *testing.T) kbapi.SLOsIndicatorPropertiesCustomMetric {
	t.Helper()
	ind := kbapi.SLOsIndicatorPropertiesCustomMetric{Type: "sli.metric.custom"}
	ind.Params.Index = "i"
	ind.Params.TimestampField = testTimestampField
	ind.Params.Good.Equation = "A"
	ind.Params.Total.Equation = "A"
	return ind
}

func makeHistogram(t *testing.T) kbapi.SLOsIndicatorPropertiesHistogram {
	t.Helper()
	ind := kbapi.SLOsIndicatorPropertiesHistogram{Type: "sli.histogram.custom"}
	ind.Params.Index = "i"
	ind.Params.TimestampField = testTimestampField
	ind.Params.Good.Aggregation = "value_count"
	ind.Params.Good.Field = "f"
	ind.Params.Total.Aggregation = "value_count"
	ind.Params.Total.Field = "f"
	return ind
}

func makeTimesliceMetric(t *testing.T) kbapi.SLOsIndicatorPropertiesTimesliceMetric {
	t.Helper()
	ind := kbapi.SLOsIndicatorPropertiesTimesliceMetric{Type: "sli.metric.timeslice"}
	ind.Params.Index = "i"
	ind.Params.TimestampField = testTimestampField
	ind.Params.Metric.Comparator = "GT"
	ind.Params.Metric.Equation = "A"
	ind.Params.Metric.Threshold = 0.99
	return ind
}

func makeResponseIndicator(t *testing.T, ind any) kbapi.SLOsSloWithSummaryResponse_Indicator {
	t.Helper()
	var result kbapi.SLOsSloWithSummaryResponse_Indicator
	var err error
	switch v := ind.(type) {
	case kbapi.SLOsIndicatorPropertiesApmAvailability:
		err = result.FromSLOsIndicatorPropertiesApmAvailability(v)
	case kbapi.SLOsIndicatorPropertiesApmLatency:
		err = result.FromSLOsIndicatorPropertiesApmLatency(v)
	case kbapi.SLOsIndicatorPropertiesCustomKql:
		err = result.FromSLOsIndicatorPropertiesCustomKql(v)
	case kbapi.SLOsIndicatorPropertiesCustomMetric:
		err = result.FromSLOsIndicatorPropertiesCustomMetric(v)
	case kbapi.SLOsIndicatorPropertiesHistogram:
		err = result.FromSLOsIndicatorPropertiesHistogram(v)
	case kbapi.SLOsIndicatorPropertiesTimesliceMetric:
		err = result.FromSLOsIndicatorPropertiesTimesliceMetric(v)
	default:
		t.Fatalf("unhandled indicator type in test helper: %T", ind)
	}
	require.NoError(t, err)
	return result
}

func Test_applyResponseIndicator_allTypes(t *testing.T) {
	cases := []struct {
		name string
		ind  any
	}{
		{name: "apm availability", ind: makeApmAvail(t)},
		{name: "apm latency", ind: makeApmLatency(t)},
		{name: "custom kql", ind: makeCustomKql(t)},
		{name: "custom metric", ind: makeCustomMetric(t)},
		{name: "histogram", ind: makeHistogram(t)},
		{name: "timeslice metric", ind: makeTimesliceMetric(t)},
	}

	for _, tc := range cases {
		t.Run(tc.name+"/create", func(t *testing.T) {
			src := makeResponseIndicator(t, tc.ind)
			got, err := ResponseIndicatorToCreateIndicator(src)
			require.NoError(t, err)
			require.NotEmpty(t, got)
		})
		t.Run(tc.name+"/update", func(t *testing.T) {
			src := makeResponseIndicator(t, tc.ind)
			got, err := ResponseIndicatorToUpdateIndicator(src)
			require.NoError(t, err)
			require.NotEmpty(t, got)
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
