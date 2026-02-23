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

package kibana

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/slo"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/stretchr/testify/require"
)

func Test_sloResponseToModel(t *testing.T) {
	syncDelay := "2m"

	tests := []struct {
		name          string
		spaceID       string
		sloResponse   *slo.SloWithSummaryResponse
		expectedModel *models.Slo
	}{
		{
			name:    "should return a model with the correct values",
			spaceID: "space-id",
			sloResponse: &slo.SloWithSummaryResponse{
				Id:          "slo-id",
				Name:        "slo-name",
				Description: "slo-description",
				Indicator: slo.SloWithSummaryResponseIndicator{
					IndicatorPropertiesApmAvailability: &slo.IndicatorPropertiesApmAvailability{
						Type: "sli.apm.transactionErrorRate",
						Params: slo.IndicatorPropertiesApmAvailabilityParams{
							Service:         "slo-service",
							Environment:     "slo-environment",
							TransactionType: "slo-transaction-type",
							TransactionName: "slo-transaction-name",
							Index:           "slo-index",
						},
					},
				},
				TimeWindow: slo.TimeWindow{
					Duration: "7d",
					Type:     "rolling",
				},
				BudgetingMethod: "occurrences",
				Settings: slo.Settings{
					SyncDelay: &syncDelay,
				},
				Revision:  5.0,
				Enabled:   true,
				CreatedAt: "2023-08-11T00:05:36.567Z",
				UpdatedAt: "2023-08-11T00:05:36.567Z",
			},
			expectedModel: &models.Slo{
				SloID:       "slo-id",
				SpaceID:     "space-id",
				Name:        "slo-name",
				Description: "slo-description",
				Indicator: slo.SloWithSummaryResponseIndicator{
					IndicatorPropertiesApmAvailability: &slo.IndicatorPropertiesApmAvailability{
						Type: "sli.apm.transactionErrorRate",
						Params: slo.IndicatorPropertiesApmAvailabilityParams{
							Service:         "slo-service",
							Environment:     "slo-environment",
							TransactionType: "slo-transaction-type",
							TransactionName: "slo-transaction-name",
							Index:           "slo-index",
						},
					},
				},
				TimeWindow: slo.TimeWindow{
					Duration: "7d",
					Type:     "rolling",
				},
				BudgetingMethod: "occurrences",
				Settings: &slo.Settings{
					SyncDelay: &syncDelay,
				},
				GroupBy: nil,
			},
		},
		{
			name:    "should return tags if available",
			spaceID: "space-id",
			sloResponse: &slo.SloWithSummaryResponse{
				Id:          "slo-id",
				Name:        "slo-name",
				Description: "slo-description",
				Indicator: slo.SloWithSummaryResponseIndicator{
					IndicatorPropertiesApmAvailability: &slo.IndicatorPropertiesApmAvailability{
						Type: "sli.apm.transactionErrorRate",
						Params: slo.IndicatorPropertiesApmAvailabilityParams{
							Service:         "slo-service",
							Environment:     "slo-environment",
							TransactionType: "slo-transaction-type",
							TransactionName: "slo-transaction-name",
							Index:           "slo-index",
						},
					},
				},
				TimeWindow: slo.TimeWindow{
					Duration: "7d",
					Type:     "rolling",
				},
				BudgetingMethod: "occurrences",
				Settings: slo.Settings{
					SyncDelay: &syncDelay,
				},
				Tags:      []string{"tag-1", "another_tag"},
				Revision:  5.0,
				Enabled:   true,
				CreatedAt: "2023-08-11T00:05:36.567Z",
				UpdatedAt: "2023-08-11T00:05:36.567Z",
			},
			expectedModel: &models.Slo{
				SloID:       "slo-id",
				SpaceID:     "space-id",
				Name:        "slo-name",
				Description: "slo-description",
				Indicator: slo.SloWithSummaryResponseIndicator{
					IndicatorPropertiesApmAvailability: &slo.IndicatorPropertiesApmAvailability{
						Type: "sli.apm.transactionErrorRate",
						Params: slo.IndicatorPropertiesApmAvailabilityParams{
							Service:         "slo-service",
							Environment:     "slo-environment",
							TransactionType: "slo-transaction-type",
							TransactionName: "slo-transaction-name",
							Index:           "slo-index",
						},
					},
				},
				TimeWindow: slo.TimeWindow{
					Duration: "7d",
					Type:     "rolling",
				},
				BudgetingMethod: "occurrences",
				Settings: &slo.Settings{
					SyncDelay: &syncDelay,
				},
				Tags:    []string{"tag-1", "another_tag"},
				GroupBy: nil,
			},
		},
		{
			name:          "nil response should return a nil model",
			spaceID:       "space-id",
			sloResponse:   nil,
			expectedModel: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := sloResponseToModel(tt.spaceID, tt.sloResponse)
			require.Equal(t, tt.expectedModel, model)
		})
	}
}
