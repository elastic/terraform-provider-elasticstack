package kibana_oapi

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/stretchr/testify/require"
)

func createGroupBySlice() []string {
	return []string{"service.name"}
}

func createGroupByForResponse() kbapi.SLOsGroupBy {
	var groupBy kbapi.SLOsGroupBy
	err := groupBy.FromSLOsGroupBy1([]string{"service.name"})
	if err != nil {
		panic(err)
	}
	return groupBy
}

func createApmAvailabilityIndicatorForWithSummary() kbapi.SLOsSloWithSummaryResponse_Indicator {
	indicator := kbapi.SLOsIndicatorPropertiesApmAvailability{
		Params: struct {
			Environment     string  `json:"environment"`
			Filter          *string `json:"filter,omitempty"`
			Index           string  `json:"index"`
			Service         string  `json:"service"`
			TransactionName string  `json:"transactionName"`
			TransactionType string  `json:"transactionType"`
		}{
			Environment:     "production",
			Service:         "service",
			TransactionName: "transaction",
			TransactionType: "request",
			Index:           "apm-*",
		},
		Type: "sli.apm.transactionDuration",
	}

	var result kbapi.SLOsSloWithSummaryResponse_Indicator
	err := result.FromSLOsIndicatorPropertiesApmAvailability(indicator)
	if err != nil {
		panic(err)
	}
	return result
}

func createApmAvailabilityIndicatorForDefinition() kbapi.SLOsSloDefinitionResponse_Indicator {
	indicator := kbapi.SLOsIndicatorPropertiesApmAvailability{
		Params: struct {
			Environment     string  `json:"environment"`
			Filter          *string `json:"filter,omitempty"`
			Index           string  `json:"index"`
			Service         string  `json:"service"`
			TransactionName string  `json:"transactionName"`
			TransactionType string  `json:"transactionType"`
		}{
			Environment:     "production",
			Service:         "service",
			TransactionName: "transaction",
			TransactionType: "request",
			Index:           "apm-*",
		},
		Type: "sli.apm.transactionDuration",
	}

	var result kbapi.SLOsSloDefinitionResponse_Indicator
	err := result.FromSLOsIndicatorPropertiesApmAvailability(indicator)
	if err != nil {
		panic(err)
	}
	return result
}

func TestConvertSloWithSummaryResponseToModel(t *testing.T) {
	syncDelay := "1m"

	tests := []struct {
		name          string
		sloResponse   *kbapi.SLOsSloWithSummaryResponse
		expectedModel *models.Slo
	}{
		{
			name: "convert SLO with summary response",
			sloResponse: &kbapi.SLOsSloWithSummaryResponse{
				Id:          "slo-id",
				Name:        "slo-name",
				Description: "slo-description",
				Indicator:   createApmAvailabilityIndicatorForWithSummary(),
				TimeWindow: kbapi.SLOsTimeWindow{
					Duration: "7d",
					Type:     "rolling",
				},
				BudgetingMethod: "occurrences",
				Objective: kbapi.SLOsObjective{
					Target: 0.95,
				},
				Settings: kbapi.SLOsSettings{
					SyncDelay: &syncDelay,
				},
				Tags:    []string{"tag-1", "tag-2"},
				GroupBy: createGroupByForResponse(),
			},
			expectedModel: &models.Slo{
				SloID:       "slo-id",
				Name:        "slo-name",
				Description: "slo-description",
				Indicator:   createApmAvailabilityIndicatorForDefinition(),
				TimeWindow: kbapi.SLOsTimeWindow{
					Duration: "7d",
					Type:     "rolling",
				},
				BudgetingMethod: "occurrences",
				Objective: kbapi.SLOsObjective{
					Target: 0.95,
				},
				Settings: &kbapi.SLOsSettings{
					SyncDelay: &syncDelay,
				},
				SpaceID: "",
				Tags:    []string{"tag-1", "tag-2"},
				GroupBy: createGroupBySlice(),
			},
		},

		{
			name:          "nil response should return nil model",
			sloResponse:   nil,
			expectedModel: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := convertSloWithSummaryResponseToModel(tt.sloResponse)

			require.Equal(t, tt.expectedModel, model)
		})
	}
}

func TestConvertSloDefinitionResponseToModel(t *testing.T) {
	syncDelay := "1m"

	tests := []struct {
		name          string
		sloResponse   *kbapi.SLOsSloDefinitionResponse
		expectedModel *models.Slo
	}{
		{
			name: "convert SLO definition response",
			sloResponse: &kbapi.SLOsSloDefinitionResponse{
				Id:          "slo-id",
				Name:        "slo-name",
				Description: "slo-description",
				Indicator:   createApmAvailabilityIndicatorForDefinition(),
				TimeWindow: kbapi.SLOsTimeWindow{
					Duration: "7d",
					Type:     "rolling",
				},
				BudgetingMethod: "occurrences",
				Objective: kbapi.SLOsObjective{
					Target: 0.95,
				},
				Settings: kbapi.SLOsSettings{
					SyncDelay: &syncDelay,
				},
				Tags:    []string{"tag-1", "tag-2"},
				GroupBy: createGroupByForResponse(),
			},
			expectedModel: &models.Slo{
				SloID:       "slo-id",
				Name:        "slo-name",
				Description: "slo-description",
				Indicator:   createApmAvailabilityIndicatorForDefinition(),
				TimeWindow: kbapi.SLOsTimeWindow{
					Duration: "7d",
					Type:     "rolling",
				},
				BudgetingMethod: "occurrences",
				Objective: kbapi.SLOsObjective{
					Target: 0.95,
				},
				Settings: &kbapi.SLOsSettings{
					SyncDelay: &syncDelay,
				},
				SpaceID: "",
				Tags:    []string{"tag-1", "tag-2"},
				GroupBy: createGroupBySlice(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := convertSloDefinitionResponseToModel(tt.sloResponse)

			require.Equal(t, tt.expectedModel, model)
		})
	}
}

func TestConvertIndicatorFromDefinitionToCreateRequest(t *testing.T) {
	tests := []struct {
		name         string
		setupInput   func() kbapi.SLOsSloDefinitionResponse_Indicator
		expectError  bool
		validateFunc func(*testing.T, *kbapi.SLOsCreateSloRequest_Indicator)
	}{
		{
			name: "APM Availability indicator conversion",
			setupInput: func() kbapi.SLOsSloDefinitionResponse_Indicator {
				var indicator kbapi.SLOsSloDefinitionResponse_Indicator
				apmIndicator := kbapi.SLOsIndicatorPropertiesApmAvailability{
					Type: "sli.apm.transactionErrorRate",
					Params: struct {
						Environment     string  `json:"environment"`
						Filter          *string `json:"filter,omitempty"`
						Index           string  `json:"index"`
						Service         string  `json:"service"`
						TransactionName string  `json:"transactionName"`
						TransactionType string  `json:"transactionType"`
					}{
						Environment:     "production",
						Service:         "test-service",
						Index:           "apm-*",
						TransactionName: "GET /api/test",
						TransactionType: "request",
					},
				}
				err := indicator.FromSLOsIndicatorPropertiesApmAvailability(apmIndicator)
				require.NoError(t, err)
				return indicator
			},
			expectError: false,
			validateFunc: func(t *testing.T, indicator *kbapi.SLOsCreateSloRequest_Indicator) {
				apmIndicator, err := indicator.AsSLOsIndicatorPropertiesApmAvailability()
				require.NoError(t, err)
				require.Equal(t, "sli.apm.transactionErrorRate", apmIndicator.Type)
				require.Equal(t, "production", apmIndicator.Params.Environment)
				require.Equal(t, "test-service", apmIndicator.Params.Service)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourceIndicator := tt.setupInput()

			result, err := convertIndicatorFromDefinitionToCreateRequest(sourceIndicator)

			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.validateFunc != nil {
					tt.validateFunc(t, result)
				}
			}
		})
	}
}

func TestConvertIndicatorFromDefinitionToUpdateRequest(t *testing.T) {
	tests := []struct {
		name        string
		setupInput  func() kbapi.SLOsSloDefinitionResponse_Indicator
		expectError bool
	}{
		{
			name: "Custom KQL indicator conversion should not fail",
			setupInput: func() kbapi.SLOsSloDefinitionResponse_Indicator {
				var indicator kbapi.SLOsSloDefinitionResponse_Indicator
				var good kbapi.SLOsKqlWithFiltersGood
				err := good.FromSLOsKqlWithFiltersGood0("response_time < 100")
				require.NoError(t, err)
				var total kbapi.SLOsKqlWithFiltersTotal
				err = total.FromSLOsKqlWithFiltersTotal0("*")
				require.NoError(t, err)

				kqlIndicator := kbapi.SLOsIndicatorPropertiesCustomKql{
					Type: "sli.kql.custom",
					Params: struct {
						DataViewId     *string                       `json:"dataViewId,omitempty"`
						Filter         *kbapi.SLOsKqlWithFilters     `json:"filter,omitempty"`
						Good           kbapi.SLOsKqlWithFiltersGood  `json:"good"`
						Index          string                        `json:"index"`
						TimestampField string                        `json:"timestampField"`
						Total          kbapi.SLOsKqlWithFiltersTotal `json:"total"`
					}{
						Index:          "logs-*",
						TimestampField: "@timestamp",
						Good:           good,
						Total:          total,
					},
				}
				err = indicator.FromSLOsIndicatorPropertiesCustomKql(kqlIndicator)
				require.NoError(t, err)
				return indicator
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourceIndicator := tt.setupInput()

			result, err := convertIndicatorFromDefinitionToUpdateRequest(sourceIndicator)

			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
			}
		})
	}
}

func TestConvertIndicatorFromResponseToDefinition(t *testing.T) {
	tests := []struct {
		name        string
		setupInput  func() kbapi.SLOsSloWithSummaryResponse_Indicator
		expectError bool
	}{
		{
			name: "APM Availability indicator conversion should succeed",
			setupInput: func() kbapi.SLOsSloWithSummaryResponse_Indicator {
				var indicator kbapi.SLOsSloWithSummaryResponse_Indicator
				apmIndicator := kbapi.SLOsIndicatorPropertiesApmAvailability{
					Type: "sli.apm.transactionErrorRate",
					Params: struct {
						Environment     string  `json:"environment"`
						Filter          *string `json:"filter,omitempty"`
						Index           string  `json:"index"`
						Service         string  `json:"service"`
						TransactionName string  `json:"transactionName"`
						TransactionType string  `json:"transactionType"`
					}{
						Environment:     "production",
						Service:         "test-service",
						Index:           "apm-*",
						TransactionName: "GET /api/test",
						TransactionType: "request",
					},
				}
				err := indicator.FromSLOsIndicatorPropertiesApmAvailability(apmIndicator)
				require.NoError(t, err)
				return indicator
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourceIndicator := tt.setupInput()

			targetIndicator, err := convertIndicatorFromResponseToDefinition(&sourceIndicator)

			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, targetIndicator)
			} else {
				require.NoError(t, err)
				require.NotNil(t, targetIndicator)
				// Verify that the conversion worked by checking if we can extract the same indicator type
				apmIndicator, extractErr := targetIndicator.AsSLOsIndicatorPropertiesApmAvailability()
				require.NoError(t, extractErr)
				require.Equal(t, "sli.apm.transactionErrorRate", apmIndicator.Type)
				require.Equal(t, "production", apmIndicator.Params.Environment)
				require.Equal(t, "test-service", apmIndicator.Params.Service)
			}
		})
	}
}
