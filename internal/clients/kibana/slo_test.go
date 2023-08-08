package kibana

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/slo"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/stretchr/testify/require"
)

func Test_sloResponseToModel(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := sloResponseToModel(tt.spaceId, tt.sloResponse)

			require.Equal(t, tt.expectedModel, model)
		})
	}
}
