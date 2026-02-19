package kibana_oapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/stretchr/testify/require"
)

func Test_convertResponseToModel(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name          string
		spaceId       string
		response      any
		expectedModel *models.AlertingRule
	}{
		{
			name:          "nil response should return a nil model",
			spaceId:       "space-id",
			response:      nil,
			expectedModel: nil,
		},
		{
			name:    "nil optional fields should not blow up the transform",
			spaceId: "space-id",
			response: map[string]any{
				"id":           "id",
				"name":         "name",
				"consumer":     "consumer",
				"params":       map[string]any{},
				"rule_type_id": "rule-type-id",
				"enabled":      true,
				"tags":         []string{"hello"},
				"schedule": map[string]any{
					"interval": "1m",
				},
			},
			expectedModel: &models.AlertingRule{
				RuleID:     "id",
				SpaceID:    "space-id",
				Name:       "name",
				Consumer:   "consumer",
				Params:     map[string]any{},
				RuleTypeID: "rule-type-id",
				Enabled:    utils.Pointer(true),
				Tags:       []string{"hello"},
				Schedule:   models.AlertingRuleSchedule{Interval: "1m"},
				Actions:    []models.AlertingRuleAction{},
				ExecutionStatus: models.AlertingRuleExecutionStatus{
					LastExecutionDate: nil,
					Status:            nil,
				},
			},
		},
		{
			name:    "a full response should be successfully transformed",
			spaceId: "space-id",
			response: map[string]any{
				"id":           "id",
				"name":         "name",
				"consumer":     "consumer",
				"params":       map[string]any{},
				"rule_type_id": "rule-type-id",
				"enabled":      true,
				"tags":         []string{"hello"},
				"notify_when":  "broken",
				"schedule": map[string]any{
					"interval": "1m",
				},
				"throttle":          "throttle",
				"scheduled_task_id": "scheduled-task-id",
				"execution_status": map[string]any{
					"last_execution_date": now.Format(time.RFC3339),
					"status":              "firing",
				},
				"alert_delay": map[string]any{
					"active": float64(4),
				},
				"actions": []any{
					map[string]any{
						"group":  "group-1",
						"id":     "id",
						"params": map[string]any{},
						"frequency": map[string]any{
							"summary":     true,
							"notify_when": "onThrottleInterval",
							"throttle":    "10s",
						},
						"alerts_filter": map[string]any{
							"query": map[string]any{
								"kql": "foobar",
							},
							"timeframe": map[string]any{
								"days":     []any{float64(3), float64(5), float64(7)},
								"timezone": "UTC+1",
								"hours": map[string]any{
									"start": "00:00",
									"end":   "08:00",
								},
							},
						},
					},
					map[string]any{
						"group":  "group-2",
						"id":     "id",
						"params": map[string]any{},
						"frequency": map[string]any{
							"summary":     true,
							"notify_when": "onActionGroupChange",
						},
					},
					map[string]any{
						"group":  "group-3",
						"id":     "id",
						"params": map[string]any{},
					},
				},
			},
			expectedModel: &models.AlertingRule{
				RuleID:          "id",
				SpaceID:         "space-id",
				Name:            "name",
				Consumer:        "consumer",
				Params:          map[string]any{},
				RuleTypeID:      "rule-type-id",
				Enabled:         utils.Pointer(true),
				Tags:            []string{"hello"},
				NotifyWhen:      utils.Pointer("broken"),
				Schedule:        models.AlertingRuleSchedule{Interval: "1m"},
				Throttle:        utils.Pointer("throttle"),
				ScheduledTaskID: utils.Pointer("scheduled-task-id"),
				ExecutionStatus: models.AlertingRuleExecutionStatus{
					LastExecutionDate: &now,
					Status:            utils.Pointer("firing"),
				},
				Actions: []models.AlertingRuleAction{
					{
						Group:  "group-1",
						ID:     "id",
						Params: map[string]any{},
						Frequency: &models.ActionFrequency{
							Summary:    true,
							NotifyWhen: "onThrottleInterval",
							Throttle:   utils.Pointer("10s"),
						},
						AlertsFilter: &models.ActionAlertsFilter{
							Kql: utils.Pointer("foobar"),
							Timeframe: &models.AlertsFilterTimeframe{
								Days:       []int32{3, 5, 7},
								Timezone:   "UTC+1",
								HoursStart: "00:00",
								HoursEnd:   "08:00",
							},
						},
					},
					{
						Group:  "group-2",
						ID:     "id",
						Params: map[string]any{},
						Frequency: &models.ActionFrequency{
							Summary:    true,
							NotifyWhen: "onActionGroupChange",
						},
					},
					{
						Group:  "group-3",
						ID:     "id",
						Params: map[string]any{},
					},
				},
				AlertDelay: utils.Pointer(float32(4)),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, diags := kibana_oapi.ConvertResponseToModel(tt.spaceId, tt.response)

			require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

			if tt.expectedModel == nil {
				require.Nil(t, model)
			} else {
				require.NotNil(t, model)
				require.Equal(t, tt.expectedModel.RuleID, model.RuleID)
				require.Equal(t, tt.expectedModel.SpaceID, model.SpaceID)
				require.Equal(t, tt.expectedModel.Name, model.Name)
				require.Equal(t, tt.expectedModel.Consumer, model.Consumer)
				require.Equal(t, tt.expectedModel.RuleTypeID, model.RuleTypeID)
				require.Equal(t, tt.expectedModel.Enabled, model.Enabled)
				require.Equal(t, tt.expectedModel.Tags, model.Tags)
				require.Equal(t, tt.expectedModel.NotifyWhen, model.NotifyWhen)
				require.Equal(t, tt.expectedModel.Schedule, model.Schedule)
				require.Equal(t, tt.expectedModel.Throttle, model.Throttle)
				require.Equal(t, tt.expectedModel.ScheduledTaskID, model.ScheduledTaskID)
				require.Equal(t, tt.expectedModel.AlertDelay, model.AlertDelay)

				// Check execution status
				if tt.expectedModel.ExecutionStatus.LastExecutionDate != nil {
					require.NotNil(t, model.ExecutionStatus.LastExecutionDate)
					// Allow for minor time precision differences
					require.WithinDuration(t, *tt.expectedModel.ExecutionStatus.LastExecutionDate, *model.ExecutionStatus.LastExecutionDate, time.Second)
				} else {
					require.Nil(t, model.ExecutionStatus.LastExecutionDate)
				}
				require.Equal(t, tt.expectedModel.ExecutionStatus.Status, model.ExecutionStatus.Status)

				// Check actions
				require.Len(t, model.Actions, len(tt.expectedModel.Actions))
				for i, expectedAction := range tt.expectedModel.Actions {
					require.Equal(t, expectedAction.Group, model.Actions[i].Group)
					require.Equal(t, expectedAction.ID, model.Actions[i].ID)
					require.Equal(t, expectedAction.Frequency, model.Actions[i].Frequency)
					require.Equal(t, expectedAction.AlertsFilter, model.Actions[i].AlertsFilter)
				}
			}
		})
	}
}

func Test_CreateUpdateAlertingRule_ErrorHandling(t *testing.T) {
	tests := []struct {
		name         string
		testFunc     func(ctx context.Context, client *kibana_oapi.Client, spaceID string, rule models.AlertingRule) (*models.AlertingRule, string)
		statusCode   int
		responseBody string
		expectedErr  string
	}{
		{
			name: "CreateAlertingRule should not crash when backend returns 4xx",
			testFunc: func(ctx context.Context, client *kibana_oapi.Client, spaceID string, rule models.AlertingRule) (*models.AlertingRule, string) {
				result, diags := kibana_oapi.CreateAlertingRule(ctx, client, spaceID, rule)
				if diags.HasError() {
					return result, diags[0].Detail()
				}
				return result, ""
			},
			statusCode:   401,
			responseBody: "some error",
			expectedErr:  "some error",
		},
		{
			name: "UpdateAlertingRule should not crash when backend returns 4xx",
			testFunc: func(ctx context.Context, client *kibana_oapi.Client, spaceID string, rule models.AlertingRule) (*models.AlertingRule, string) {
				result, diags := kibana_oapi.UpdateAlertingRule(ctx, client, spaceID, rule)
				if diags.HasError() {
					return result, diags[0].Detail()
				}
				return result, ""
			},
			statusCode:   401,
			responseBody: "some error",
			expectedErr:  "some error",
		},
		{
			name: "CreateAlertingRule should not crash when backend returns an empty response and HTTP 200",
			testFunc: func(ctx context.Context, client *kibana_oapi.Client, spaceID string, rule models.AlertingRule) (*models.AlertingRule, string) {
				result, diags := kibana_oapi.CreateAlertingRule(ctx, client, spaceID, rule)
				if diags.HasError() {
					return result, diags[0].Detail()
				}
				return result, ""
			},
			statusCode:   200,
			responseBody: "{}",
			expectedErr:  "missing required fields",
		},
		{
			name: "UpdateAlertingRule should not crash when backend returns an empty response and HTTP 200",
			testFunc: func(ctx context.Context, client *kibana_oapi.Client, spaceID string, rule models.AlertingRule) (*models.AlertingRule, string) {
				result, diags := kibana_oapi.UpdateAlertingRule(ctx, client, spaceID, rule)
				if diags.HasError() {
					return result, diags[0].Detail()
				}
				return result, ""
			},
			statusCode:   200,
			responseBody: "{}",
			expectedErr:  "missing required fields",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client, err := kibana_oapi.NewClient(kibana_oapi.Config{
				URL:      server.URL,
				Username: "test",
				Password: "test",
			})
			require.NoError(t, err)

			rule := models.AlertingRule{
				RuleID:     "test-rule-id",
				Name:       "test",
				Consumer:   "alerts",
				RuleTypeID: ".index-threshold",
				Schedule:   models.AlertingRuleSchedule{Interval: "1m"},
				Params:     map[string]any{},
			}

			result, errDetail := tt.testFunc(context.Background(), client, "default", rule)

			require.Nil(t, result)
			require.NotEmpty(t, errDetail)
			require.Contains(t, errDetail, tt.expectedErr)
		})
	}
}
