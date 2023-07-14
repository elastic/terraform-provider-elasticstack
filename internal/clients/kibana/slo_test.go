package kibana

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/slo"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/stretchr/testify/require"
)

func Test_sloResponseToModel(t *testing.T) {
	// now := time.Now()
	tests := []struct {
		name          string
		spaceId       string
		sloResponse   *slo.SloResponse
		expectedModel *models.Slo
	}{
		{
			name:          "nil response should return a nil model",
			spaceId:       "space-id",
			sloResponse:   nil,
			expectedModel: nil,
		},
		{
			name:    "nil optional fields should not blow up the slo",
			spaceId: "space-id",
			sloResponse: &slo.SloResponse{
				Id:          makePtr("id"),
				Name:        makePtr("name"),
				Description: makePtr("description"),
				Indicator: &slo.SloResponseIndicator{
					IndicatorPropertiesCustomKql: &slo.IndicatorPropertiesCustomKql{
						Type: "sli.kql.custom",
						Params: slo.IndicatorPropertiesCustomKqlParams{
							Index:          "index",
							TimestampField: "timestamp-field",
						},
					},
				},
				TimeWindow: &slo.TimeWindow{
					Duration: "1m",
					Type:     "rolling",
				},
				BudgetingMethod: (*slo.BudgetingMethod)(makePtr("budgeting-method")),
				Objective: &slo.Objective{
					Target: 0.99,
				},
			},
			expectedModel: &models.Slo{
				ID:          "id",
				SpaceID:     "space-id",
				Name:        "name",
				Description: "description",
				Indicator: slo.SloResponseIndicator{
					IndicatorPropertiesCustomKql: &slo.IndicatorPropertiesCustomKql{
						Type: "sli.kql.custom",
						Params: slo.IndicatorPropertiesCustomKqlParams{
							Index:          "index",
							TimestampField: "timestamp-field",
						},
					},
				},
				TimeWindow: slo.TimeWindow{
					Duration: "1m",
					Type:     "rolling",
				},
				BudgetingMethod: "budgeting-method",
				Objective: slo.Objective{
					Target: 0.99,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := sloResponseToModel(tt.spaceId, tt.sloResponse)

			require.Equal(t, tt.expectedModel, model)
		})
	}
}
