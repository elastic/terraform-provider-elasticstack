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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func CreateAlertingRule(ctx context.Context, client *Client, spaceID string, rule models.AlertingRule) (*models.AlertingRule, diag.Diagnostics) {
	body := buildCreateRequestBody(rule)
	resp, err := client.API.PostAlertingRuleIdWithResponse(
		ctx,
		rule.RuleID,
		body,
		SpaceAwarePathRequestEditor(spaceID),
	)
	if err != nil {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("HTTP request failed", err.Error())}
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, diag.Diagnostics{diag.NewErrorDiagnostic(
				"Create rule returned an empty response",
				fmt.Sprintf("Create rule returned an empty response with HTTP status code [%d].", resp.StatusCode()),
			)}
		}
		return ConvertResponseToModel(spaceID, resp.JSON200)
	case http.StatusConflict:
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic(
			"Rule ID conflict",
			fmt.Sprintf("Status code [%d], Saved object [%s/%s] conflict (Rule ID already exists in this Space)", resp.StatusCode(), spaceID, rule.RuleID),
		)}
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

func GetAlertingRule(ctx context.Context, client *Client, spaceID string, ruleID string) (*models.AlertingRule, diag.Diagnostics) {
	resp, err := client.API.GetAlertingRuleIdWithResponse(
		ctx,
		ruleID,
		SpaceAwarePathRequestEditor(spaceID),
	)
	if err != nil {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Unable to get alerting rule", err.Error())}
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, diag.Diagnostics{diag.NewErrorDiagnostic(
				"Get rule returned an empty response",
				fmt.Sprintf("Get rule returned an empty response with HTTP status code [%d].", resp.StatusCode()),
			)}
		}
		return ConvertResponseToModel(spaceID, resp.JSON200)
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

func UpdateAlertingRule(ctx context.Context, client *Client, spaceID string, rule models.AlertingRule) (*models.AlertingRule, diag.Diagnostics) {
	body := buildUpdateRequestBody(rule)

	resp, err := client.API.PutAlertingRuleIdWithResponse(
		ctx,
		rule.RuleID,
		body,
		SpaceAwarePathRequestEditor(spaceID),
	)
	if err != nil {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Unable to update alerting rule", err.Error())}
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, diag.Diagnostics{diag.NewErrorDiagnostic(
				"Update rule returned an empty response",
				fmt.Sprintf("Update rule returned an empty response with HTTP status code [%d].", resp.StatusCode()),
			)}
		}

		var wasEnabled bool
		if data, err := json.Marshal(resp.JSON200); err == nil {
			var temp struct {
				Enabled bool `json:"enabled"`
			}
			if err := json.Unmarshal(data, &temp); err == nil {
				wasEnabled = temp.Enabled
			}
		}

		shouldBeEnabled := rule.Enabled != nil && *rule.Enabled

		if shouldBeEnabled && !wasEnabled {
			if diags := EnableAlertingRule(ctx, client, spaceID, rule.RuleID); diags.HasError() {
				return nil, diags
			}
		}

		if !shouldBeEnabled && wasEnabled {
			if diags := DisableAlertingRule(ctx, client, spaceID, rule.RuleID); diags.HasError() {
				return nil, diags
			}
		}

		returnedRule, convDiags := ConvertResponseToModel(spaceID, resp.JSON200)
		if convDiags.HasError() {
			return nil, convDiags
		}

		if rule.Enabled != nil {
			returnedRule.Enabled = rule.Enabled
		}

		return returnedRule, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

func DeleteAlertingRule(ctx context.Context, client *Client, spaceID string, ruleID string) diag.Diagnostics {
	resp, err := client.API.DeleteAlertingRuleIdWithResponse(
		ctx,
		ruleID,
		SpaceAwarePathRequestEditor(spaceID),
	)
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Unable to delete alerting rule", err.Error())}
	}

	switch resp.StatusCode() {
	case http.StatusNoContent, http.StatusOK:
		return nil
	case http.StatusNotFound:
		return nil
	default:
		return reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

func EnableAlertingRule(ctx context.Context, client *Client, spaceID string, ruleID string) diag.Diagnostics {
	resp, err := client.API.PostAlertingRuleIdEnableWithResponse(
		ctx,
		ruleID,
		SpaceAwarePathRequestEditor(spaceID),
	)
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Unable to enable alerting rule", err.Error())}
	}

	switch resp.StatusCode() {
	case http.StatusNoContent, http.StatusOK:
		return nil
	default:
		return reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

func DisableAlertingRule(ctx context.Context, client *Client, spaceID string, ruleID string) diag.Diagnostics {
	body := kbapi.PostAlertingRuleIdDisableJSONRequestBody{}
	resp, err := client.API.PostAlertingRuleIdDisableWithResponse(
		ctx,
		ruleID,
		body,
		SpaceAwarePathRequestEditor(spaceID),
	)
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Unable to disable alerting rule", err.Error())}
	}

	switch resp.StatusCode() {
	case http.StatusNoContent, http.StatusOK:
		return nil
	default:
		return reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

func ConvertResponseToModel(spaceID string, resp any) (*models.AlertingRule, diag.Diagnostics) {
	if resp == nil {
		return nil, nil
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Failed to marshal response", err.Error())}
	}

	var intermediate struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		Consumer   string `json:"consumer"`
		Enabled    bool   `json:"enabled"`
		RuleTypeID string `json:"rule_type_id"`
		Schedule   struct {
			Interval string `json:"interval"`
		} `json:"schedule"`
		Params          map[string]any `json:"params"`
		Tags            []string       `json:"tags"`
		NotifyWhen      *string        `json:"notify_when"`
		Throttle        *string        `json:"throttle"`
		ScheduledTaskID *string        `json:"scheduled_task_id"`
		ExecutionStatus struct {
			LastExecutionDate string `json:"last_execution_date"`
			Status            string `json:"status"`
		} `json:"execution_status"`
		AlertDelay *struct {
			Active float32 `json:"active"`
		} `json:"alert_delay"`
		Actions []struct {
			Group     *string        `json:"group"`
			ID        string         `json:"id"`
			Params    map[string]any `json:"params"`
			Frequency *struct {
				NotifyWhen string  `json:"notify_when"`
				Summary    bool    `json:"summary"`
				Throttle   *string `json:"throttle"`
			} `json:"frequency"`
			AlertsFilter *struct {
				Query *struct {
					Kql string `json:"kql"`
				} `json:"query"`
				Timeframe *struct {
					Days  []int `json:"days"`
					Hours struct {
						Start string `json:"start"`
						End   string `json:"end"`
					} `json:"hours"`
					Timezone string `json:"timezone"`
				} `json:"timeframe"`
			} `json:"alerts_filter"`
		} `json:"actions"`
	}

	if err := json.Unmarshal(data, &intermediate); err != nil {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Failed to unmarshal response", err.Error())}
	}

	if intermediate.ID == "" || intermediate.Name == "" || intermediate.Consumer == "" || intermediate.RuleTypeID == "" {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic(
			"Invalid rule response",
			"Response is missing required fields (id, name, consumer, or rule_type_id)",
		)}
	}

	actions := []models.AlertingRuleAction{}
	for _, action := range intermediate.Actions {
		a := models.AlertingRuleAction{
			Group:  valueOrDefault(action.Group, "default"),
			ID:     action.ID,
			Params: action.Params,
		}

		if action.Frequency != nil {
			a.Frequency = &models.ActionFrequency{
				Summary:    action.Frequency.Summary,
				NotifyWhen: action.Frequency.NotifyWhen,
				Throttle:   action.Frequency.Throttle,
			}
		}

		if action.AlertsFilter != nil {
			a.AlertsFilter = &models.ActionAlertsFilter{}

			if action.AlertsFilter.Query != nil {
				a.AlertsFilter.Kql = &action.AlertsFilter.Query.Kql
			}

			if action.AlertsFilter.Timeframe != nil {
				days := make([]int32, len(action.AlertsFilter.Timeframe.Days))
				for i, d := range action.AlertsFilter.Timeframe.Days {
					days[i] = int32(d)
				}
				a.AlertsFilter.Timeframe = &models.AlertsFilterTimeframe{
					Days:       days,
					Timezone:   action.AlertsFilter.Timeframe.Timezone,
					HoursStart: action.AlertsFilter.Timeframe.Hours.Start,
					HoursEnd:   action.AlertsFilter.Timeframe.Hours.End,
				}
			}
		}

		actions = append(actions, a)
	}

	var alertDelay *float32
	if intermediate.AlertDelay != nil {
		alertDelay = &intermediate.AlertDelay.Active
	}

	var lastExecutionDate *time.Time
	if intermediate.ExecutionStatus.LastExecutionDate != "" {
		if parsed, err := time.Parse(time.RFC3339, intermediate.ExecutionStatus.LastExecutionDate); err == nil {
			lastExecutionDate = &parsed
		}
	}

	var status *string
	if intermediate.ExecutionStatus.Status != "" {
		status = &intermediate.ExecutionStatus.Status
	}

	return &models.AlertingRule{
		RuleID:     intermediate.ID,
		SpaceID:    spaceID,
		Name:       intermediate.Name,
		Consumer:   intermediate.Consumer,
		NotifyWhen: intermediate.NotifyWhen,
		Params:     intermediate.Params,
		RuleTypeID: intermediate.RuleTypeID,
		Schedule: models.AlertingRuleSchedule{
			Interval: intermediate.Schedule.Interval,
		},
		Enabled:         &intermediate.Enabled,
		Tags:            intermediate.Tags,
		Throttle:        intermediate.Throttle,
		ScheduledTaskID: intermediate.ScheduledTaskID,
		ExecutionStatus: models.AlertingRuleExecutionStatus{
			LastExecutionDate: lastExecutionDate,
			Status:            status,
		},
		Actions:    actions,
		AlertDelay: alertDelay,
	}, nil
}

func buildCreateRequestBody(rule models.AlertingRule) kbapi.PostAlertingRuleIdJSONRequestBody {
	body := kbapi.PostAlertingRuleIdJSONRequestBody{
		Consumer:   rule.Consumer,
		Name:       rule.Name,
		RuleTypeId: rule.RuleTypeID,
		Schedule: struct {
			Interval string `json:"interval"`
		}{
			Interval: rule.Schedule.Interval,
		},
	}

	if rule.Params != nil {
		params := kbapi.AlertingRuleAPIParams{
			AdditionalProperties: rule.Params,
		}
		body.Params = &params
	}

	if rule.Enabled != nil {
		body.Enabled = rule.Enabled
	}

	if rule.NotifyWhen != nil && *rule.NotifyWhen != "" {
		notifyWhen := kbapi.PostAlertingRuleIdJSONBodyNotifyWhen(*rule.NotifyWhen)
		body.NotifyWhen = &notifyWhen
	}

	if rule.Throttle != nil {
		body.Throttle = rule.Throttle
	}

	if rule.Tags != nil {
		tags := rule.Tags
		body.Tags = &tags
	}

	if rule.AlertDelay != nil {
		body.AlertDelay = &struct {
			Active float32 `json:"active"`
		}{
			Active: *rule.AlertDelay,
		}
	}

	if len(rule.Actions) > 0 {
		actions := make([]struct {
			AlertsFilter *struct {
				Query *struct {
					Dsl     *string `json:"dsl,omitempty"`
					Filters []struct {
						State *struct {
							Store kbapi.PostAlertingRuleIdJSONBodyActionsAlertsFilterQueryFiltersStateStore `json:"store"`
						} `json:"$state,omitempty"`
						Meta  map[string]any  `json:"meta"`
						Query *map[string]any `json:"query,omitempty"`
					} `json:"filters"`
					Kql string `json:"kql"`
				} `json:"query,omitempty"`
				Timeframe *struct {
					Days  []kbapi.PostAlertingRuleIdJSONBodyActionsAlertsFilterTimeframeDays `json:"days"`
					Hours struct {
						End   string `json:"end"`
						Start string `json:"start"`
					} `json:"hours"`
					Timezone string `json:"timezone"`
				} `json:"timeframe,omitempty"`
			} `json:"alerts_filter,omitempty"`
			Frequency *struct {
				NotifyWhen kbapi.PostAlertingRuleIdJSONBodyActionsFrequencyNotifyWhen `json:"notify_when"`
				Summary    bool                                                       `json:"summary"`
				Throttle   *string                                                    `json:"throttle,omitempty"`
			} `json:"frequency,omitempty"`
			Group                   *string         `json:"group,omitempty"`
			Id                      string          `json:"id"` //nolint:revive // var-naming: API struct field
			Params                  *map[string]any `json:"params,omitempty"`
			UseAlertDataForTemplate *bool           `json:"use_alert_data_for_template,omitempty"`
			Uuid                    *string         `json:"uuid,omitempty"` //nolint:revive // var-naming: API struct field
		}, len(rule.Actions))

		for i, action := range rule.Actions {
			actions[i].Id = action.ID
			if action.Group != "" {
				group := action.Group
				actions[i].Group = &group
			}
			if action.Params != nil {
				actions[i].Params = &action.Params
			}

			if action.Frequency != nil {
				actions[i].Frequency = &struct {
					NotifyWhen kbapi.PostAlertingRuleIdJSONBodyActionsFrequencyNotifyWhen `json:"notify_when"`
					Summary    bool                                                       `json:"summary"`
					Throttle   *string                                                    `json:"throttle,omitempty"`
				}{
					NotifyWhen: kbapi.PostAlertingRuleIdJSONBodyActionsFrequencyNotifyWhen(action.Frequency.NotifyWhen),
					Summary:    action.Frequency.Summary,
					Throttle:   action.Frequency.Throttle,
				}
			}

			if action.AlertsFilter != nil {
				filter := &struct {
					Query *struct {
						Dsl     *string `json:"dsl,omitempty"`
						Filters []struct {
							State *struct {
								Store kbapi.PostAlertingRuleIdJSONBodyActionsAlertsFilterQueryFiltersStateStore `json:"store"`
							} `json:"$state,omitempty"`
							Meta  map[string]any  `json:"meta"`
							Query *map[string]any `json:"query,omitempty"`
						} `json:"filters"`
						Kql string `json:"kql"`
					} `json:"query,omitempty"`
					Timeframe *struct {
						Days  []kbapi.PostAlertingRuleIdJSONBodyActionsAlertsFilterTimeframeDays `json:"days"`
						Hours struct {
							End   string `json:"end"`
							Start string `json:"start"`
						} `json:"hours"`
						Timezone string `json:"timezone"`
					} `json:"timeframe,omitempty"`
				}{}

				if action.AlertsFilter.Kql != nil {
					filter.Query = &struct {
						Dsl     *string `json:"dsl,omitempty"`
						Filters []struct {
							State *struct {
								Store kbapi.PostAlertingRuleIdJSONBodyActionsAlertsFilterQueryFiltersStateStore `json:"store"`
							} `json:"$state,omitempty"`
							Meta  map[string]any  `json:"meta"`
							Query *map[string]any `json:"query,omitempty"`
						} `json:"filters"`
						Kql string `json:"kql"`
					}{
						Kql: *action.AlertsFilter.Kql,
						Filters: []struct {
							State *struct {
								Store kbapi.PostAlertingRuleIdJSONBodyActionsAlertsFilterQueryFiltersStateStore `json:"store"`
							} `json:"$state,omitempty"`
							Meta  map[string]any  `json:"meta"`
							Query *map[string]any `json:"query,omitempty"`
						}{},
					}
				}

				if action.AlertsFilter.Timeframe != nil {
					days := make([]kbapi.PostAlertingRuleIdJSONBodyActionsAlertsFilterTimeframeDays, len(action.AlertsFilter.Timeframe.Days))
					for j, d := range action.AlertsFilter.Timeframe.Days {
						days[j] = kbapi.PostAlertingRuleIdJSONBodyActionsAlertsFilterTimeframeDays(d)
					}

					filter.Timeframe = &struct {
						Days  []kbapi.PostAlertingRuleIdJSONBodyActionsAlertsFilterTimeframeDays `json:"days"`
						Hours struct {
							End   string `json:"end"`
							Start string `json:"start"`
						} `json:"hours"`
						Timezone string `json:"timezone"`
					}{
						Days: days,
						Hours: struct {
							End   string `json:"end"`
							Start string `json:"start"`
						}{
							Start: action.AlertsFilter.Timeframe.HoursStart,
							End:   action.AlertsFilter.Timeframe.HoursEnd,
						},
						Timezone: action.AlertsFilter.Timeframe.Timezone,
					}
				}

				actions[i].AlertsFilter = filter
			}
		}

		body.Actions = &actions
	}

	return body
}

func buildUpdateRequestBody(rule models.AlertingRule) kbapi.PutAlertingRuleIdJSONRequestBody {
	body := kbapi.PutAlertingRuleIdJSONRequestBody{
		Name: rule.Name,
		Schedule: struct {
			Interval string `json:"interval"`
		}{
			Interval: rule.Schedule.Interval,
		},
	}

	if rule.Params != nil {
		body.Params = &rule.Params
	}

	if rule.NotifyWhen != nil && *rule.NotifyWhen != "" {
		notifyWhen := kbapi.PutAlertingRuleIdJSONBodyNotifyWhen(*rule.NotifyWhen)
		body.NotifyWhen = &notifyWhen
	}

	if rule.Throttle != nil {
		body.Throttle = rule.Throttle
	}

	if rule.Tags != nil {
		tags := rule.Tags
		body.Tags = &tags
	}

	if rule.AlertDelay != nil {
		body.AlertDelay = &struct {
			Active float32 `json:"active"`
		}{
			Active: *rule.AlertDelay,
		}
	}

	if len(rule.Actions) > 0 {
		actions := make([]struct {
			AlertsFilter *struct {
				Query *struct {
					Dsl     *string `json:"dsl,omitempty"`
					Filters []struct {
						State *struct {
							Store kbapi.PutAlertingRuleIdJSONBodyActionsAlertsFilterQueryFiltersStateStore `json:"store"`
						} `json:"$state,omitempty"`
						Meta  map[string]any  `json:"meta"`
						Query *map[string]any `json:"query,omitempty"`
					} `json:"filters"`
					Kql string `json:"kql"`
				} `json:"query,omitempty"`
				Timeframe *struct {
					Days  []kbapi.PutAlertingRuleIdJSONBodyActionsAlertsFilterTimeframeDays `json:"days"`
					Hours struct {
						End   string `json:"end"`
						Start string `json:"start"`
					} `json:"hours"`
					Timezone string `json:"timezone"`
				} `json:"timeframe,omitempty"`
			} `json:"alerts_filter,omitempty"`
			Frequency *struct {
				NotifyWhen kbapi.PutAlertingRuleIdJSONBodyActionsFrequencyNotifyWhen `json:"notify_when"`
				Summary    bool                                                      `json:"summary"`
				Throttle   *string                                                   `json:"throttle,omitempty"`
			} `json:"frequency,omitempty"`
			Group                   *string         `json:"group,omitempty"`
			Id                      string          `json:"id"` //nolint:revive // var-naming: API struct field
			Params                  *map[string]any `json:"params,omitempty"`
			UseAlertDataForTemplate *bool           `json:"use_alert_data_for_template,omitempty"`
			Uuid                    *string         `json:"uuid,omitempty"` //nolint:revive // var-naming: API struct field
		}, len(rule.Actions))

		for i, action := range rule.Actions {
			actions[i].Group = &action.Group
			actions[i].Id = action.ID
			if action.Params != nil {
				actions[i].Params = &action.Params
			}

			if action.Frequency != nil {
				actions[i].Frequency = &struct {
					NotifyWhen kbapi.PutAlertingRuleIdJSONBodyActionsFrequencyNotifyWhen `json:"notify_when"`
					Summary    bool                                                      `json:"summary"`
					Throttle   *string                                                   `json:"throttle,omitempty"`
				}{
					NotifyWhen: kbapi.PutAlertingRuleIdJSONBodyActionsFrequencyNotifyWhen(action.Frequency.NotifyWhen),
					Summary:    action.Frequency.Summary,
					Throttle:   action.Frequency.Throttle,
				}
			}

			if action.AlertsFilter != nil {
				filter := &struct {
					Query *struct {
						Dsl     *string `json:"dsl,omitempty"`
						Filters []struct {
							State *struct {
								Store kbapi.PutAlertingRuleIdJSONBodyActionsAlertsFilterQueryFiltersStateStore `json:"store"`
							} `json:"$state,omitempty"`
							Meta  map[string]any  `json:"meta"`
							Query *map[string]any `json:"query,omitempty"`
						} `json:"filters"`
						Kql string `json:"kql"`
					} `json:"query,omitempty"`
					Timeframe *struct {
						Days  []kbapi.PutAlertingRuleIdJSONBodyActionsAlertsFilterTimeframeDays `json:"days"`
						Hours struct {
							End   string `json:"end"`
							Start string `json:"start"`
						} `json:"hours"`
						Timezone string `json:"timezone"`
					} `json:"timeframe,omitempty"`
				}{}

				if action.AlertsFilter.Kql != nil {
					filter.Query = &struct {
						Dsl     *string `json:"dsl,omitempty"`
						Filters []struct {
							State *struct {
								Store kbapi.PutAlertingRuleIdJSONBodyActionsAlertsFilterQueryFiltersStateStore `json:"store"`
							} `json:"$state,omitempty"`
							Meta  map[string]any  `json:"meta"`
							Query *map[string]any `json:"query,omitempty"`
						} `json:"filters"`
						Kql string `json:"kql"`
					}{
						Kql: *action.AlertsFilter.Kql,
						Filters: []struct {
							State *struct {
								Store kbapi.PutAlertingRuleIdJSONBodyActionsAlertsFilterQueryFiltersStateStore `json:"store"`
							} `json:"$state,omitempty"`
							Meta  map[string]any  `json:"meta"`
							Query *map[string]any `json:"query,omitempty"`
						}{},
					}
				}

				if action.AlertsFilter.Timeframe != nil {
					days := make([]kbapi.PutAlertingRuleIdJSONBodyActionsAlertsFilterTimeframeDays, len(action.AlertsFilter.Timeframe.Days))
					for j, d := range action.AlertsFilter.Timeframe.Days {
						days[j] = kbapi.PutAlertingRuleIdJSONBodyActionsAlertsFilterTimeframeDays(d)
					}
					filter.Timeframe = &struct {
						Days  []kbapi.PutAlertingRuleIdJSONBodyActionsAlertsFilterTimeframeDays `json:"days"`
						Hours struct {
							End   string `json:"end"`
							Start string `json:"start"`
						} `json:"hours"`
						Timezone string `json:"timezone"`
					}{
						Days: days,
						Hours: struct {
							End   string `json:"end"`
							Start string `json:"start"`
						}{
							Start: action.AlertsFilter.Timeframe.HoursStart,
							End:   action.AlertsFilter.Timeframe.HoursEnd,
						},
						Timezone: action.AlertsFilter.Timeframe.Timezone,
					}
				}

				actions[i].AlertsFilter = filter
			}
		}
		body.Actions = &actions
	}

	return body
}

func valueOrDefault(val *string, def string) string {
	if val != nil && *val != "" {
		return *val
	}
	return def
}
