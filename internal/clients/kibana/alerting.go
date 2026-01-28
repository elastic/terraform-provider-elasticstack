package kibana

import (
	"context"
	"fmt"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func ruleResponseToModel(spaceID string, res *kbapi.GetAlertingRuleIdResponse) *models.AlertingRule {
	if res == nil || res.JSON200 == nil {
		return nil
	}

	data := res.JSON200
	actions := []models.AlertingRuleAction{}
	for _, action := range data.Actions {

		a := models.AlertingRuleAction{
			Group:  *action.Group,
			ID:     action.Id,
			Params: action.Params,
		}

		if action.Frequency != nil {
			frequency := action.Frequency

			a.Frequency = &models.ActionFrequency{
				Summary:    frequency.Summary,
				NotifyWhen: string(frequency.NotifyWhen),
				Throttle:   frequency.Throttle,
			}
		}

		if action.AlertsFilter != nil {
			filter := action.AlertsFilter

			a.AlertsFilter = &models.ActionAlertsFilter{}

			if filter.Query != nil {
				a.AlertsFilter.Kql = &filter.Query.Kql
			}

			if filter.Timeframe != nil {
				timeframe := filter.Timeframe
				days := make([]int32, len(timeframe.Days))
				for i, d := range timeframe.Days {
					days[i] = int32(d)
				}
				a.AlertsFilter.Timeframe = &models.AlertsFilterTimeframe{
					Days:       days,
					Timezone:   timeframe.Timezone,
					HoursStart: timeframe.Hours.Start,
					HoursEnd:   timeframe.Hours.End,
				}
			}
		}

		actions = append(actions, a)
	}

	var alertDelay *float32
	if data.AlertDelay != nil {
		alertDelay = &data.AlertDelay.Active
	}

	var notifyWhen *string
	if data.NotifyWhen != nil {
		nw := string(*data.NotifyWhen)
		notifyWhen = &nw
	}

	var throttle *string
	if data.Throttle != nil {
		throttle = data.Throttle
	}

	var lastExecutionDate *time.Time
	if data.ExecutionStatus.LastExecutionDate != "" {
		t, err := time.Parse(time.RFC3339, data.ExecutionStatus.LastExecutionDate)
		if err == nil {
			lastExecutionDate = &t
		}
	}

	status := string(data.ExecutionStatus.Status)

	return &models.AlertingRule{
		RuleID:   data.Id,
		SpaceID:  spaceID,
		Name:     data.Name,
		Consumer: data.Consumer,

		// DEPRECATED
		NotifyWhen: notifyWhen,

		Params:     data.Params,
		RuleTypeID: data.RuleTypeId,
		Schedule: models.AlertingRuleSchedule{
			Interval: data.Schedule.Interval,
		},
		Enabled:         &data.Enabled,
		Tags:            data.Tags,
		Throttle:        throttle,
		ScheduledTaskID: data.ScheduledTaskId,
		ExecutionStatus: models.AlertingRuleExecutionStatus{
			LastExecutionDate: lastExecutionDate,
			Status:            &status,
		},
		Actions:    actions,
		AlertDelay: alertDelay,
	}
}

// Maps the rule actions to the struct required by the request model (Actions array)
func ruleActionsToActionsInner(ruleActions []models.AlertingRuleAction) *[]struct {
	AlertsFilter *struct {
		Query *struct {
			Dsl     *string `json:"dsl,omitempty"`
			Filters []struct {
				State *struct {
					Store kbapi.PostAlertingRuleIdJSONBodyActionsAlertsFilterQueryFiltersStateStore `json:"store"`
				} `json:"$state,omitempty"`
				Meta  map[string]interface{}  `json:"meta"`
				Query *map[string]interface{} `json:"query,omitempty"`
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
	Group                   *string                 `json:"group,omitempty"`
	Id                      string                  `json:"id"`
	Params                  *map[string]interface{} `json:"params,omitempty"`
	UseAlertDataForTemplate *bool                   `json:"use_alert_data_for_template,omitempty"`
	Uuid                    *string                 `json:"uuid,omitempty"`
} {
	actions := []struct {
		AlertsFilter *struct {
			Query *struct {
				Dsl     *string `json:"dsl,omitempty"`
				Filters []struct {
					State *struct {
						Store kbapi.PostAlertingRuleIdJSONBodyActionsAlertsFilterQueryFiltersStateStore `json:"store"`
					} `json:"$state,omitempty"`
					Meta  map[string]interface{}  `json:"meta"`
					Query *map[string]interface{} `json:"query,omitempty"`
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
		Group                   *string                 `json:"group,omitempty"`
		Id                      string                  `json:"id"`
		Params                  *map[string]interface{} `json:"params,omitempty"`
		UseAlertDataForTemplate *bool                   `json:"use_alert_data_for_template,omitempty"`
		Uuid                    *string                 `json:"uuid,omitempty"`
	}{}

	for index := range ruleActions {
		action := ruleActions[index]

		group := action.Group
		params := action.Params

		actionToAppend := struct {
			AlertsFilter *struct {
				Query *struct {
					Dsl     *string `json:"dsl,omitempty"`
					Filters []struct {
						State *struct {
							Store kbapi.PostAlertingRuleIdJSONBodyActionsAlertsFilterQueryFiltersStateStore `json:"store"`
						} `json:"$state,omitempty"`
						Meta  map[string]interface{}  `json:"meta"`
						Query *map[string]interface{} `json:"query,omitempty"`
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
			Group                   *string                 `json:"group,omitempty"`
			Id                      string                  `json:"id"`
			Params                  *map[string]interface{} `json:"params,omitempty"`
			UseAlertDataForTemplate *bool                   `json:"use_alert_data_for_template,omitempty"`
			Uuid                    *string                 `json:"uuid,omitempty"`
		}{
			Group:  &group,
			Id:     action.ID,
			Params: &params,
		}

		if action.Frequency != nil {
			actionToAppend.Frequency = &struct {
				NotifyWhen kbapi.PostAlertingRuleIdJSONBodyActionsFrequencyNotifyWhen `json:"notify_when"`
				Summary    bool                                                       `json:"summary"`
				Throttle   *string                                                    `json:"throttle,omitempty"`
			}{
				Summary:    action.Frequency.Summary,
				NotifyWhen: kbapi.PostAlertingRuleIdJSONBodyActionsFrequencyNotifyWhen(action.Frequency.NotifyWhen),
				Throttle:   action.Frequency.Throttle,
			}
		}

		if action.AlertsFilter != nil {
			filter := struct {
				Query *struct {
					Dsl     *string `json:"dsl,omitempty"`
					Filters []struct {
						State *struct {
							Store kbapi.PostAlertingRuleIdJSONBodyActionsAlertsFilterQueryFiltersStateStore `json:"store"`
						} `json:"$state,omitempty"`
						Meta  map[string]interface{}  `json:"meta"`
						Query *map[string]interface{} `json:"query,omitempty"`
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
						Meta  map[string]interface{}  `json:"meta"`
						Query *map[string]interface{} `json:"query,omitempty"`
					} `json:"filters"`
					Kql string `json:"kql"`
				}{
					Kql: *action.AlertsFilter.Kql,
					Filters: []struct {
						State *struct {
							Store kbapi.PostAlertingRuleIdJSONBodyActionsAlertsFilterQueryFiltersStateStore `json:"store"`
						} `json:"$state,omitempty"`
						Meta  map[string]interface{}  `json:"meta"`
						Query *map[string]interface{} `json:"query,omitempty"`
					}{},
				}
			}

			if action.AlertsFilter.Timeframe != nil {
				timeframe := action.AlertsFilter.Timeframe
				days := make([]kbapi.PostAlertingRuleIdJSONBodyActionsAlertsFilterTimeframeDays, len(timeframe.Days))
				for i, d := range timeframe.Days {
					days[i] = kbapi.PostAlertingRuleIdJSONBodyActionsAlertsFilterTimeframeDays(d)
				}
				filter.Timeframe = &struct {
					Days  []kbapi.PostAlertingRuleIdJSONBodyActionsAlertsFilterTimeframeDays `json:"days"`
					Hours struct {
						End   string `json:"end"`
						Start string `json:"start"`
					} `json:"hours"`
					Timezone string `json:"timezone"`
				}{
					Timezone: timeframe.Timezone,
					Days:     days,
					Hours: struct {
						End   string `json:"end"`
						Start string `json:"start"`
					}{
						Start: timeframe.HoursStart,
						End:   timeframe.HoursEnd,
					},
				}
			}

			actionToAppend.AlertsFilter = &filter
		}

		actions = append(actions, actionToAppend)
	}
	return &actions
}

//go:generate go tool go.uber.org/mock/mockgen -destination=./alerting_mocks.go -package=kibana -source ./alerting.go ApiClient
type ApiClient interface {
	GetKibanaOapiClient() (*kibana_oapi.Client, error)
}

// enableAlertingRule enables an alerting rule using the Kibana API
func enableAlertingRule(ctx context.Context, client *kibana_oapi.Client, ruleID, spaceID string) diag.Diagnostics {
	fwDiags := kibana_oapi.EnableAlertingRule(ctx, client, spaceID, ruleID)
	return diagutil.SDKDiagsFromFramework(fwDiags)
}

// disableAlertingRule disables an alerting rule using the Kibana API
func disableAlertingRule(ctx context.Context, client *kibana_oapi.Client, ruleID, spaceID string) diag.Diagnostics {
	fwDiags := kibana_oapi.DisableAlertingRule(ctx, client, spaceID, ruleID)
	return diagutil.SDKDiagsFromFramework(fwDiags)
}

func CreateAlertingRule(ctx context.Context, apiClient ApiClient, rule models.AlertingRule) (*models.AlertingRule, diag.Diagnostics) {
	client, err := apiClient.GetKibanaOapiClient()
	if err != nil {
		return nil, diag.FromErr(err)
	}

	var alertDelay *struct {
		Active float32 `json:"active"`
	}

	if rule.AlertDelay != nil {
		alertDelay = &struct {
			Active float32 `json:"active"`
		}{
			Active: *rule.AlertDelay,
		}
	}

	var notifyWhen *kbapi.PostAlertingRuleIdJSONBodyNotifyWhen
	if rule.NotifyWhen != nil {
		nw := kbapi.PostAlertingRuleIdJSONBodyNotifyWhen(*rule.NotifyWhen)
		notifyWhen = &nw
	}

	var params *kbapi.PostAlertingRuleIdJSONBody_Params
	if rule.Params != nil {
		p := kbapi.PostAlertingRuleIdJSONBody_Params{
			AdditionalProperties: rule.Params,
		}
		params = &p
	}

	var tags *[]string
	if rule.Tags != nil {
		tags = &rule.Tags
	}

	reqModel := kbapi.PostAlertingRuleIdJSONRequestBody{
		Consumer:   rule.Consumer,
		Actions:    ruleActionsToActionsInner(rule.Actions),
		Enabled:    rule.Enabled,
		Name:       rule.Name,
		NotifyWhen: notifyWhen,
		Params:     params,
		RuleTypeId: rule.RuleTypeID,
		Schedule: struct {
			Interval string `json:"interval"`
		}{
			Interval: rule.Schedule.Interval,
		},
		Tags:       tags,
		Throttle:   rule.Throttle,
		AlertDelay: alertDelay,
	}

	resp, fwDiags := kibana_oapi.CreateAlertingRule(ctx, client, rule.SpaceID, rule.RuleID, reqModel)
	diags := diagutil.SDKDiagsFromFramework(fwDiags)
	if diags.HasError() {
		return nil, diags
	}

	if resp == nil || resp.JSON200 == nil {
		return nil, diag.Diagnostics{diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Create rule returned an empty response",
			Detail:   fmt.Sprintf("Create rule returned an empty response with HTTP status code [%d].", resp.StatusCode()),
		}}
	}

	rule.RuleID = resp.JSON200.Id

	// Re-fetch the rule to get the full response with the correct type
	return GetAlertingRule(ctx, apiClient, rule.RuleID, rule.SpaceID)
}

func UpdateAlertingRule(ctx context.Context, apiClient ApiClient, rule models.AlertingRule) (*models.AlertingRule, diag.Diagnostics) {
	client, err := apiClient.GetKibanaOapiClient()
	if err != nil {
		return nil, diag.FromErr(err)
	}

	var alertDelay *struct {
		Active float32 `json:"active"`
	}

	if rule.AlertDelay != nil {
		alertDelay = &struct {
			Active float32 `json:"active"`
		}{
			Active: *rule.AlertDelay,
		}
	}

	var notifyWhen *kbapi.PutAlertingRuleIdJSONBodyNotifyWhen
	if rule.NotifyWhen != nil {
		nw := kbapi.PutAlertingRuleIdJSONBodyNotifyWhen(*rule.NotifyWhen)
		notifyWhen = &nw
	}

	// Convert actions to the proper type for PUT request
	var actions *[]struct {
		AlertsFilter *struct {
			Query *struct {
				Dsl     *string `json:"dsl,omitempty"`
				Filters []struct {
					State *struct {
						Store kbapi.PutAlertingRuleIdJSONBodyActionsAlertsFilterQueryFiltersStateStore `json:"store"`
					} `json:"$state,omitempty"`
					Meta  map[string]interface{}  `json:"meta"`
					Query *map[string]interface{} `json:"query,omitempty"`
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
		Group                   *string                 `json:"group,omitempty"`
		Id                      string                  `json:"id"`
		Params                  *map[string]interface{} `json:"params,omitempty"`
		UseAlertDataForTemplate *bool                   `json:"use_alert_data_for_template,omitempty"`
		Uuid                    *string                 `json:"uuid,omitempty"`
	}

	if len(rule.Actions) > 0 {
		convertedActions := []struct {
			AlertsFilter *struct {
				Query *struct {
					Dsl     *string `json:"dsl,omitempty"`
					Filters []struct {
						State *struct {
							Store kbapi.PutAlertingRuleIdJSONBodyActionsAlertsFilterQueryFiltersStateStore `json:"store"`
						} `json:"$state,omitempty"`
						Meta  map[string]interface{}  `json:"meta"`
						Query *map[string]interface{} `json:"query,omitempty"`
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
			Group                   *string                 `json:"group,omitempty"`
			Id                      string                  `json:"id"`
			Params                  *map[string]interface{} `json:"params,omitempty"`
			UseAlertDataForTemplate *bool                   `json:"use_alert_data_for_template,omitempty"`
			Uuid                    *string                 `json:"uuid,omitempty"`
		}{}

		postActions := ruleActionsToActionsInner(rule.Actions)
		for _, postAction := range *postActions {
			putAction := struct {
				AlertsFilter *struct {
					Query *struct {
						Dsl     *string `json:"dsl,omitempty"`
						Filters []struct {
							State *struct {
								Store kbapi.PutAlertingRuleIdJSONBodyActionsAlertsFilterQueryFiltersStateStore `json:"store"`
							} `json:"$state,omitempty"`
							Meta  map[string]interface{}  `json:"meta"`
							Query *map[string]interface{} `json:"query,omitempty"`
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
				Group                   *string                 `json:"group,omitempty"`
				Id                      string                  `json:"id"`
				Params                  *map[string]interface{} `json:"params,omitempty"`
				UseAlertDataForTemplate *bool                   `json:"use_alert_data_for_template,omitempty"`
				Uuid                    *string                 `json:"uuid,omitempty"`
			}{
				Group:  postAction.Group,
				Id:     postAction.Id,
				Params: postAction.Params,
			}

			if postAction.Frequency != nil {
				putAction.Frequency = &struct {
					NotifyWhen kbapi.PutAlertingRuleIdJSONBodyActionsFrequencyNotifyWhen `json:"notify_when"`
					Summary    bool                                                      `json:"summary"`
					Throttle   *string                                                   `json:"throttle,omitempty"`
				}{
					NotifyWhen: kbapi.PutAlertingRuleIdJSONBodyActionsFrequencyNotifyWhen(postAction.Frequency.NotifyWhen),
					Summary:    postAction.Frequency.Summary,
					Throttle:   postAction.Frequency.Throttle,
				}
			}

			if postAction.AlertsFilter != nil {
				putFilter := &struct {
					Query *struct {
						Dsl     *string `json:"dsl,omitempty"`
						Filters []struct {
							State *struct {
								Store kbapi.PutAlertingRuleIdJSONBodyActionsAlertsFilterQueryFiltersStateStore `json:"store"`
							} `json:"$state,omitempty"`
							Meta  map[string]interface{}  `json:"meta"`
							Query *map[string]interface{} `json:"query,omitempty"`
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

				if postAction.AlertsFilter.Query != nil {
					filters := make([]struct {
						State *struct {
							Store kbapi.PutAlertingRuleIdJSONBodyActionsAlertsFilterQueryFiltersStateStore `json:"store"`
						} `json:"$state,omitempty"`
						Meta  map[string]interface{}  `json:"meta"`
						Query *map[string]interface{} `json:"query,omitempty"`
					}, len(postAction.AlertsFilter.Query.Filters))
					for i, f := range postAction.AlertsFilter.Query.Filters {
						filters[i].Meta = f.Meta
						filters[i].Query = f.Query
						if f.State != nil {
							filters[i].State = &struct {
								Store kbapi.PutAlertingRuleIdJSONBodyActionsAlertsFilterQueryFiltersStateStore `json:"store"`
							}{
								Store: kbapi.PutAlertingRuleIdJSONBodyActionsAlertsFilterQueryFiltersStateStore(f.State.Store),
							}
						}
					}
					putFilter.Query = &struct {
						Dsl     *string `json:"dsl,omitempty"`
						Filters []struct {
							State *struct {
								Store kbapi.PutAlertingRuleIdJSONBodyActionsAlertsFilterQueryFiltersStateStore `json:"store"`
							} `json:"$state,omitempty"`
							Meta  map[string]interface{}  `json:"meta"`
							Query *map[string]interface{} `json:"query,omitempty"`
						} `json:"filters"`
						Kql string `json:"kql"`
					}{
						Kql:     postAction.AlertsFilter.Query.Kql,
						Filters: filters,
					}
				}

				if postAction.AlertsFilter.Timeframe != nil {
					days := make([]kbapi.PutAlertingRuleIdJSONBodyActionsAlertsFilterTimeframeDays, len(postAction.AlertsFilter.Timeframe.Days))
					for i, d := range postAction.AlertsFilter.Timeframe.Days {
						days[i] = kbapi.PutAlertingRuleIdJSONBodyActionsAlertsFilterTimeframeDays(d)
					}
					putFilter.Timeframe = &struct {
						Days  []kbapi.PutAlertingRuleIdJSONBodyActionsAlertsFilterTimeframeDays `json:"days"`
						Hours struct {
							End   string `json:"end"`
							Start string `json:"start"`
						} `json:"hours"`
						Timezone string `json:"timezone"`
					}{
						Days:     days,
						Timezone: postAction.AlertsFilter.Timeframe.Timezone,
						Hours: struct {
							End   string `json:"end"`
							Start string `json:"start"`
						}{
							End:   postAction.AlertsFilter.Timeframe.Hours.End,
							Start: postAction.AlertsFilter.Timeframe.Hours.Start,
						},
					}
				}

				putAction.AlertsFilter = putFilter
			}

			convertedActions = append(convertedActions, putAction)
		}
		actions = &convertedActions
	}

	var params *map[string]interface{}
	if rule.Params != nil {
		params = &rule.Params
	}

	var tags *[]string
	if rule.Tags != nil {
		tags = &rule.Tags
	}

	reqModel := kbapi.PutAlertingRuleIdJSONRequestBody{
		Actions:    actions,
		Name:       rule.Name,
		NotifyWhen: notifyWhen,
		Params:     params,
		Schedule: struct {
			Interval string `json:"interval"`
		}{
			Interval: rule.Schedule.Interval,
		},
		Tags:       tags,
		Throttle:   rule.Throttle,
		AlertDelay: alertDelay,
	}

	resp, fwDiags := kibana_oapi.UpdateAlertingRule(ctx, client, rule.SpaceID, rule.RuleID, reqModel)
	diags := diagutil.SDKDiagsFromFramework(fwDiags)
	if diags.HasError() {
		return nil, diags
	}

	if resp == nil || resp.JSON200 == nil {
		return nil, diag.Diagnostics{diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Update rule returned an empty response",
			Detail:   fmt.Sprintf("Update rule returned an empty response with HTTP status code [%d].", resp.StatusCode()),
		}}
	}

	rule.RuleID = resp.JSON200.Id

	shouldBeEnabled := rule.Enabled != nil && *rule.Enabled

	if shouldBeEnabled && !resp.JSON200.Enabled {
		if diags := enableAlertingRule(ctx, client, rule.RuleID, rule.SpaceID); diags.HasError() {
			return nil, diags
		}
	}

	if !shouldBeEnabled && resp.JSON200.Enabled {
		if diags := disableAlertingRule(ctx, client, rule.RuleID, rule.SpaceID); diags.HasError() {
			return nil, diags
		}
	}

	// Re-fetch the rule to get the full response with the correct type
	return GetAlertingRule(ctx, apiClient, rule.RuleID, rule.SpaceID)
}

func GetAlertingRule(ctx context.Context, apiClient ApiClient, id, spaceID string) (*models.AlertingRule, diag.Diagnostics) {
	client, err := apiClient.GetKibanaOapiClient()
	if err != nil {
		return nil, diag.FromErr(err)
	}

	resp, fwDiags := kibana_oapi.GetAlertingRule(ctx, client, spaceID, id)
	diags := diagutil.SDKDiagsFromFramework(fwDiags)
	if diags.HasError() {
		return nil, diags
	}

	if resp == nil {
		return nil, nil
	}

	return ruleResponseToModel(spaceID, resp), nil
}

func DeleteAlertingRule(ctx context.Context, apiClient ApiClient, ruleId string, spaceId string) diag.Diagnostics {
	client, err := apiClient.GetKibanaOapiClient()
	if err != nil {
		return diag.FromErr(err)
	}

	fwDiags := kibana_oapi.DeleteAlertingRule(ctx, client, spaceId, ruleId)
	return diagutil.SDKDiagsFromFramework(fwDiags)
}
