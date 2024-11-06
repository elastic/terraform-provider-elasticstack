package kibana

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/generated/alerting"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

func Test_ruleResponseToModel(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name          string
		spaceId       string
		ruleResponse  *alerting.RuleResponseProperties
		expectedModel *models.AlertingRule
	}{
		{
			name:          "nil response should return a nil model",
			spaceId:       "space-id",
			ruleResponse:  nil,
			expectedModel: nil,
		},
		{
			name:    "nil optional fields should not blow up the transform",
			spaceId: "space-id",
			ruleResponse: &alerting.RuleResponseProperties{
				Id:         "id",
				Name:       "name",
				Consumer:   "consumer",
				Params:     map[string]interface{}{},
				RuleTypeId: "rule-type-id",
				Enabled:    true,
				Tags:       []string{"hello"},
			},
			expectedModel: &models.AlertingRule{
				RuleID:     "id",
				SpaceID:    "space-id",
				Name:       "name",
				Consumer:   "consumer",
				Params:     map[string]interface{}{},
				RuleTypeID: "rule-type-id",
				Enabled:    utils.Pointer(true),
				Tags:       []string{"hello"},
				Actions:    []models.AlertingRuleAction{},
			},
		},
		{
			name:    "a full response should be successfully transformed",
			spaceId: "space-id",
			ruleResponse: &alerting.RuleResponseProperties{
				Id:         "id",
				Name:       "name",
				Consumer:   "consumer",
				Params:     map[string]interface{}{},
				RuleTypeId: "rule-type-id",
				Enabled:    true,
				Tags:       []string{"hello"},
				NotifyWhen: *alerting.NewNullableString(utils.Pointer("broken")),
				Actions: []alerting.ActionsInner{
					{
						Group:  "group-1",
						Id:     "id",
						Params: map[string]interface{}{},
						Frequency: utils.Pointer(alerting.ActionsInnerFrequency{
							Summary:    true,
							NotifyWhen: "onThrottleInterval",
							Throttle:   *alerting.NewNullableString(utils.Pointer("10s")),
						}),
						AlertsFilter: utils.Pointer(alerting.ActionsInnerAlertsFilter{
							Query: &alerting.ActionsInnerAlertsFilterQuery{
								Kql: utils.Pointer("foobar"),
							},
							Timeframe: &alerting.ActionsInnerAlertsFilterTimeframe{
								Days:     []int32{3, 5, 7},
								Timezone: utils.Pointer("UTC+1"),
								Hours: &alerting.ActionsInnerAlertsFilterTimeframeHours{
									Start: utils.Pointer("00:00"),
									End:   utils.Pointer("08:00"),
								},
							},
						}),
					},
					{
						Group:  "group-2",
						Id:     "id",
						Params: map[string]interface{}{},
						Frequency: utils.Pointer(alerting.ActionsInnerFrequency{
							Summary:    true,
							NotifyWhen: "onActionGroupChange",
						}),
					},
					{
						Group:  "group-3",
						Id:     "id",
						Params: map[string]interface{}{},
					},
				},
				ExecutionStatus: alerting.RuleResponsePropertiesExecutionStatus{
					Status:            utils.Pointer("firing"),
					LastExecutionDate: &now,
				},
				ScheduledTaskId: utils.Pointer("scheduled-task-id"),
				Schedule: alerting.Schedule{
					Interval: utils.Pointer("1m"),
				},
				Throttle: *alerting.NewNullableString(utils.Pointer("throttle")),
				AlertDelay: &alerting.AlertDelay{
					Active: float32(4),
				},
			},
			expectedModel: &models.AlertingRule{
				RuleID:          "id",
				SpaceID:         "space-id",
				Name:            "name",
				Consumer:        "consumer",
				Params:          map[string]interface{}{},
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
						Params: map[string]interface{}{},
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
						Params: map[string]interface{}{},
						Frequency: &models.ActionFrequency{
							Summary:    true,
							NotifyWhen: "onActionGroupChange",
						},
					},
					{
						Group:  "group-3",
						ID:     "id",
						Params: map[string]interface{}{},
					},
				},
				AlertDelay: utils.Pointer(float32(4)),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := ruleResponseToModel(tt.spaceId, tt.ruleResponse)

			require.Equal(t, tt.expectedModel, model)
		})
	}
}

func Test_CreateUpdateAlertingRule(t *testing.T) {
	ctrl := gomock.NewController(t)

	getApiClient := func() (ApiClient, *alerting.MockAlertingAPI) {
		apiClient := NewMockApiClient(ctrl)
		apiClient.EXPECT().SetAlertingAuthContext(gomock.Any()).DoAndReturn(func(ctx context.Context) context.Context { return ctx })
		alertingClient := alerting.NewMockAlertingAPI(ctrl)
		apiClient.EXPECT().GetAlertingClient().DoAndReturn(func() (alerting.AlertingAPI, error) { return alertingClient, nil })
		return apiClient, alertingClient
	}

	tests := []struct {
		name        string
		testFunc    func(ctx context.Context, apiClient ApiClient, rule models.AlertingRule) (*models.AlertingRule, diag.Diagnostics)
		client      ApiClient
		rule        models.AlertingRule
		expectedRes *models.AlertingRule
		expectedErr string
	}{
		{
			name:     "CreateAlertingRule should not crash when backend returns 4xx",
			testFunc: CreateAlertingRule,
			client: func() ApiClient {
				apiClient, alertingClient := getApiClient()
				alertingClient.EXPECT().CreateRuleId(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, spaceId string, ruleId string) alerting.ApiCreateRuleIdRequest {
					return alerting.ApiCreateRuleIdRequest{ApiService: alertingClient}
				})
				alertingClient.EXPECT().CreateRuleIdExecute(gomock.Any()).DoAndReturn(func(r alerting.ApiCreateRuleIdRequest) (*alerting.RuleResponseProperties, *http.Response, error) {
					return nil, &http.Response{
						StatusCode: 401,
						Body:       io.NopCloser(strings.NewReader("some error")),
					}, nil
				})
				return apiClient
			}(),
			rule:        models.AlertingRule{},
			expectedRes: nil,
			expectedErr: "some error",
		},
		{
			name:     "UpdateAlertingRule should not crash when backend returns 4xx",
			testFunc: UpdateAlertingRule,
			client: func() ApiClient {
				apiClient, alertingClient := getApiClient()
				alertingClient.EXPECT().UpdateRule(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, spaceId string, ruleId string) alerting.ApiUpdateRuleRequest {
					return alerting.ApiUpdateRuleRequest{ApiService: alertingClient}
				})
				alertingClient.EXPECT().UpdateRuleExecute(gomock.Any()).DoAndReturn(func(r alerting.ApiUpdateRuleRequest) (*alerting.RuleResponseProperties, *http.Response, error) {
					return nil, &http.Response{
						StatusCode: 401,
						Body:       io.NopCloser(strings.NewReader("some error")),
					}, nil
				})
				return apiClient
			}(),
			rule:        models.AlertingRule{},
			expectedRes: nil,
			expectedErr: "some error",
		},
		{
			name:     "CreateAlertingRule should not crash when backend returns an empty response and HTTP 200",
			testFunc: CreateAlertingRule,
			client: func() ApiClient {
				apiClient, alertingClient := getApiClient()
				alertingClient.EXPECT().CreateRuleId(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, spaceId string, ruleId string) alerting.ApiCreateRuleIdRequest {
					return alerting.ApiCreateRuleIdRequest{ApiService: alertingClient}
				})
				alertingClient.EXPECT().CreateRuleIdExecute(gomock.Any()).DoAndReturn(func(r alerting.ApiCreateRuleIdRequest) (*alerting.RuleResponseProperties, *http.Response, error) {
					return nil, &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(strings.NewReader("everything seems fine")),
					}, nil
				})
				return apiClient
			}(),
			rule:        models.AlertingRule{},
			expectedRes: nil,
			expectedErr: "empty response",
		},
		{
			name:     "UpdateAlertingRule should not crash when backend returns an empty response and HTTP 200",
			testFunc: UpdateAlertingRule,
			client: func() ApiClient {
				apiClient, alertingClient := getApiClient()
				alertingClient.EXPECT().UpdateRule(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, spaceId string, ruleId string) alerting.ApiUpdateRuleRequest {
					return alerting.ApiUpdateRuleRequest{ApiService: alertingClient}
				})
				alertingClient.EXPECT().UpdateRuleExecute(gomock.Any()).DoAndReturn(func(r alerting.ApiUpdateRuleRequest) (*alerting.RuleResponseProperties, *http.Response, error) {
					return nil, &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(strings.NewReader("everything seems fine")),
					}, nil
				})
				return apiClient
			}(),
			rule:        models.AlertingRule{},
			expectedRes: nil,
			expectedErr: "empty response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule, diags := tt.testFunc(context.Background(), tt.client, tt.rule)

			if tt.expectedRes == nil {
				require.Nil(t, rule)
			} else {
				require.Equal(t, tt.expectedRes, rule)
			}

			if tt.expectedErr != "" {
				require.NotEmpty(t, diags)
				if !strings.Contains(diags[0].Detail, tt.expectedErr) {
					require.Fail(t, fmt.Sprintf("Diags ['%s'] should contain message ['%s']", diags[0].Detail, tt.expectedErr))
				}
			}
		})
	}
}
