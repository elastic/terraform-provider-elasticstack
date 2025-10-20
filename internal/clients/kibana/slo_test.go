package kibana

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/slo"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/stretchr/testify/require"
)

func Test_sloResponseToModel(t *testing.T) {
	var syncDelay = "2m"

	tests := []struct {
		name          string
		spaceId       string
		sloId         string
		sloResponse   *slo.SloWithSummaryResponse
		expectedModel *models.Slo
	}{
		{
			name:    "should return a model with the correct values",
			spaceId: "space-id",
			sloId:   "slo-id",
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
				SpaceID: "space-id",
				GroupBy: nil,
			},
		},

		{
			name:    "should return tags if available",
			spaceId: "space-id",
			sloId:   "slo-id",
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
				SpaceID: "space-id",
				Tags:    []string{"tag-1", "another_tag"},
				GroupBy: nil,
			},
		},

		{
			name:          "nil response should return a nil model",
			spaceId:       "space-id",
			sloId:         "slo-id",
			sloResponse:   nil,
			expectedModel: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := sloResponseToModel(tt.spaceId, tt.sloResponse)

			require.Equal(t, tt.expectedModel, model)
		})
	}
}
