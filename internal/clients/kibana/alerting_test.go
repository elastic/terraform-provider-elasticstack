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
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/stretchr/testify/require"
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
				Enabled:    makePtr(true),
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
				NotifyWhen: makePtr(alerting.NotifyWhen("broken")),
				Actions: []alerting.ActionsInner{
					{
						Group:  makePtr("group-1"),
						Id:     makePtr("id"),
						Params: map[string]interface{}{},
					},
					{
						Group:  makePtr("group-2"),
						Id:     makePtr("id"),
						Params: map[string]interface{}{},
					},
				},
				ExecutionStatus: alerting.RuleResponsePropertiesExecutionStatus{
					Status:            makePtr("firing"),
					LastExecutionDate: &now,
				},
				ScheduledTaskId: makePtr("scheduled-task-id"),
				Schedule: alerting.Schedule{
					Interval: makePtr("1m"),
				},
				Throttle: *alerting.NewNullableString(makePtr("throttle")),
			},
			expectedModel: &models.AlertingRule{
				RuleID:          "id",
				SpaceID:         "space-id",
				Name:            "name",
				Consumer:        "consumer",
				Params:          map[string]interface{}{},
				RuleTypeID:      "rule-type-id",
				Enabled:         makePtr(true),
				Tags:            []string{"hello"},
				NotifyWhen:      "broken",
				Schedule:        models.AlertingRuleSchedule{Interval: "1m"},
				Throttle:        makePtr("throttle"),
				ScheduledTaskID: makePtr("scheduled-task-id"),
				ExecutionStatus: models.AlertingRuleExecutionStatus{
					LastExecutionDate: &now,
					Status:            makePtr("firing"),
				},
				Actions: []models.AlertingRuleAction{
					{
						Group:  "group-1",
						ID:     "id",
						Params: map[string]interface{}{},
					},
					{
						Group:  "group-2",
						ID:     "id",
						Params: map[string]interface{}{},
					},
				},
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
	getApiClient := func() ApiClient {
		return &FakeApiClient{
			&FakeAlertingAPI{
				RuleResponseProperties: nil,
				HttpResponse: &http.Response{
					StatusCode: 401,
					Body:       io.NopCloser(strings.NewReader("some error")),
				},
			},
		}
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
			name:        "CreateAlertingRule should not crash when client returns 4xx",
			testFunc:    CreateAlertingRule,
			client:      getApiClient(),
			rule:        models.AlertingRule{},
			expectedRes: nil,
			expectedErr: "some error",
		},
		{
			name:        "UpdateAlertingRule should not crash when client returns 4xx",
			testFunc:    UpdateAlertingRule,
			client:      getApiClient(),
			rule:        models.AlertingRule{},
			expectedRes: nil,
			expectedErr: "some error",
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

type FakeApiClient struct {
	Alerting alerting.AlertingAPI
}

func (f *FakeApiClient) SetAlertingAuthContext(ctx context.Context) context.Context { return ctx }

func (f *FakeApiClient) GetAlertingClient() (alerting.AlertingAPI, error) {
	return f.Alerting, nil
}
