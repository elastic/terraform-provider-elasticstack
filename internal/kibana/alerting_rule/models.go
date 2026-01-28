package alerting_rule

import (
	"context"
	"encoding/json"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type alertingRuleModel struct {
	ID                  types.String         `tfsdk:"id"`
	RuleID              types.String         `tfsdk:"rule_id"`
	SpaceID             types.String         `tfsdk:"space_id"`
	Name                types.String         `tfsdk:"name"`
	Consumer            types.String         `tfsdk:"consumer"`
	NotifyWhen          types.String         `tfsdk:"notify_when"`
	Params              jsontypes.Normalized `tfsdk:"params"`
	RuleTypeID          types.String         `tfsdk:"rule_type_id"`
	Interval            customtypes.Duration `tfsdk:"interval"`
	Actions             types.List           `tfsdk:"actions"` //> actionModel
	Enabled             types.Bool           `tfsdk:"enabled"`
	Tags                types.List           `tfsdk:"tags"` //> string
	Throttle            customtypes.Duration `tfsdk:"throttle"`
	ScheduledTaskID     types.String         `tfsdk:"scheduled_task_id"`
	LastExecutionStatus types.String         `tfsdk:"last_execution_status"`
	LastExecutionDate   types.String         `tfsdk:"last_execution_date"`
	AlertDelay          types.Float64        `tfsdk:"alert_delay"`
}

type actionModel struct {
	Group        types.String         `tfsdk:"group"`
	ID           types.String         `tfsdk:"id"`
	Params       jsontypes.Normalized `tfsdk:"params"`
	Frequency    types.Object         `tfsdk:"frequency"`     //> frequencyModel
	AlertsFilter types.Object         `tfsdk:"alerts_filter"` //> alertsFilterModel
}

type frequencyModel struct {
	Summary    types.Bool           `tfsdk:"summary"`
	NotifyWhen types.String         `tfsdk:"notify_when"`
	Throttle   customtypes.Duration `tfsdk:"throttle"`
}

type alertsFilterModel struct {
	Kql       types.String `tfsdk:"kql"`
	Timeframe types.Object `tfsdk:"timeframe"` //> timeframeModel
}

type timeframeModel struct {
	Days       types.List   `tfsdk:"days"` //> int64
	Timezone   types.String `tfsdk:"timezone"`
	HoursStart types.String `tfsdk:"hours_start"`
	HoursEnd   types.String `tfsdk:"hours_end"`
}

func (m *alertingRuleModel) populateFromAPI(ctx context.Context, resp *kbapi.GetAlertingRuleIdResponse) diag.Diagnostics {
	var diags diag.Diagnostics

	if resp == nil || resp.JSON200 == nil {
		return diags
	}

	data := resp.JSON200

	compositeID := clients.CompositeId{
		ClusterId:  m.SpaceID.ValueString(),
		ResourceId: data.Id,
	}

	m.ID = types.StringValue(compositeID.String())
	m.RuleID = types.StringValue(data.Id)
	m.SpaceID = types.StringValue(m.SpaceID.ValueString())
	m.Name = types.StringValue(data.Name)
	m.Consumer = types.StringValue(data.Consumer)
	m.NotifyWhen = typeutils.StringishPointerValue(data.NotifyWhen)
	m.RuleTypeID = types.StringValue(data.RuleTypeId)
	m.Enabled = types.BoolValue(data.Enabled)

	// Handle interval
	if data.Schedule.Interval != "" {
		m.Interval = customtypes.NewDurationValue(data.Schedule.Interval)
	}

	// Handle throttle
	if data.Throttle != nil {
		m.Throttle = customtypes.NewDurationValue(*data.Throttle)
	}

	// Handle params - marshal to JSON
	if data.Params != nil {
		paramsJSON, err := json.Marshal(data.Params)
		if err != nil {
			diags.AddError("Error marshaling params", err.Error())
			return diags
		}
		m.Params = jsontypes.NewNormalizedValue(string(paramsJSON))
	}

	// Handle tags
	if len(data.Tags) > 0 {
		tags := make([]types.String, len(data.Tags))
		for i, tag := range data.Tags {
			tags[i] = types.StringValue(tag)
		}
		tagsList, tagsDiags := types.ListValueFrom(ctx, types.StringType, tags)
		diags.Append(tagsDiags...)
		m.Tags = tagsList
	} else {
		// Preserve existing null/unknown state
		if !utils.IsKnown(m.Tags) {
			m.Tags = types.ListNull(types.StringType)
		} else if len(m.Tags.Elements()) == 0 {
			m.Tags = types.ListNull(types.StringType)
		}
	}

	// Handle actions
	if len(data.Actions) > 0 {
		actions := make([]actionModel, len(data.Actions))
		for i, action := range data.Actions {
			actionParams, err := json.Marshal(action.Params)
			if err != nil {
				diags.AddError("Error marshaling action params", err.Error())
				return diags
			}

			actions[i] = actionModel{
				Group:  typeutils.StringishPointerValue(action.Group),
				ID:     types.StringValue(action.Id),
				Params: jsontypes.NewNormalizedValue(string(actionParams)),
			}

			// Handle frequency
			if action.Frequency != nil {
				freqModel := frequencyModel{
					Summary:    types.BoolValue(action.Frequency.Summary),
					NotifyWhen: typeutils.StringishValue(action.Frequency.NotifyWhen),
				}
				if action.Frequency.Throttle != nil {
					freqModel.Throttle = customtypes.NewDurationValue(*action.Frequency.Throttle)
				}
				freqObj, freqDiags := types.ObjectValueFrom(ctx, getFrequencyAttrTypes(), freqModel)
				diags.Append(freqDiags...)
				actions[i].Frequency = freqObj
			} else {
				actions[i].Frequency = types.ObjectNull(getFrequencyAttrTypes())
			}

			// Handle alerts_filter
			if action.AlertsFilter != nil {
				filterModel := alertsFilterModel{}

				if action.AlertsFilter.Query != nil {
					filterModel.Kql = types.StringValue(action.AlertsFilter.Query.Kql)
				}

				if action.AlertsFilter.Timeframe != nil {
					days := make([]types.Int64, len(action.AlertsFilter.Timeframe.Days))
					for j, day := range action.AlertsFilter.Timeframe.Days {
						days[j] = types.Int64Value(int64(day))
					}
					daysList, daysDiags := types.ListValueFrom(ctx, types.Int64Type, days)
					diags.Append(daysDiags...)

					tfModel := timeframeModel{
						Days:       daysList,
						Timezone:   types.StringValue(action.AlertsFilter.Timeframe.Timezone),
						HoursStart: types.StringValue(action.AlertsFilter.Timeframe.Hours.Start),
						HoursEnd:   types.StringValue(action.AlertsFilter.Timeframe.Hours.End),
					}
					tfObj, tfDiags := types.ObjectValueFrom(ctx, getTimeframeAttrTypes(), tfModel)
					diags.Append(tfDiags...)
					filterModel.Timeframe = tfObj
				} else {
					filterModel.Timeframe = types.ObjectNull(getTimeframeAttrTypes())
				}

				filterObj, filterDiags := types.ObjectValueFrom(ctx, getAlertsFilterAttrTypes(), filterModel)
				diags.Append(filterDiags...)
				actions[i].AlertsFilter = filterObj
			} else {
				actions[i].AlertsFilter = types.ObjectNull(getAlertsFilterAttrTypes())
			}
		}

		actionsList, actionsDiags := types.ListValueFrom(ctx, getActionElemType(), actions)
		diags.Append(actionsDiags...)
		m.Actions = actionsList
	} else {
		// Preserve existing null/unknown state
		if !utils.IsKnown(m.Actions) {
			m.Actions = types.ListNull(getActionElemType())
		} else if len(m.Actions.Elements()) == 0 {
			m.Actions = types.ListNull(getActionElemType())
		}
	}

	// Handle computed fields
	if data.ScheduledTaskId != nil {
		m.ScheduledTaskID = types.StringValue(*data.ScheduledTaskId)
	}

	m.LastExecutionStatus = typeutils.StringishValue(data.ExecutionStatus.Status)

	if data.ExecutionStatus.LastExecutionDate != "" {
		// Parse the time and format it
		t, err := time.Parse(time.RFC3339, data.ExecutionStatus.LastExecutionDate)
		if err == nil {
			m.LastExecutionDate = types.StringValue(t.Format("2006-01-02 15:04:05.999 -0700 MST"))
		}
	}

	if data.AlertDelay != nil {
		m.AlertDelay = types.Float64Value(float64(data.AlertDelay.Active))
	}

	return diags
}

func (m alertingRuleModel) toAPICreateModel(ctx context.Context) (kbapi.PostAlertingRuleIdJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics

	body := kbapi.PostAlertingRuleIdJSONRequestBody{
		Name:       m.Name.ValueString(),
		Consumer:   m.Consumer.ValueString(),
		RuleTypeId: m.RuleTypeID.ValueString(),
		Schedule: struct {
			Interval string `json:"interval"`
		}{
			Interval: m.Interval.ValueString(),
		},
	}

	// Handle params
	if utils.IsKnown(m.Params) {
		var params map[string]interface{}
		if err := json.Unmarshal([]byte(m.Params.ValueString()), &params); err != nil {
			diags.AddError("Error unmarshaling params", err.Error())
			return body, diags
		}
		// PostAlertingRuleId uses a special union type for Params
		// We need to marshal it back to JSON and use it as RawMessage
		paramsBytes, err := json.Marshal(params)
		if err != nil {
			diags.AddError("Error marshaling params", err.Error())
			return body, diags
		}
		// Create the union type and set its internal union field via JSON unmarshaling
		var paramsUnion kbapi.PostAlertingRuleIdJSONBody_Params
		if err := json.Unmarshal(paramsBytes, &paramsUnion); err != nil {
			diags.AddError("Error creating params union", err.Error())
			return body, diags
		}
		body.Params = &paramsUnion
	}

	// Handle enabled
	if utils.IsKnown(m.Enabled) {
		enabled := m.Enabled.ValueBool()
		body.Enabled = &enabled
	}

	// Handle notify_when
	if utils.IsKnown(m.NotifyWhen) {
		notifyWhen := kbapi.PostAlertingRuleIdJSONBodyNotifyWhen(m.NotifyWhen.ValueString())
		body.NotifyWhen = &notifyWhen
	}

	// Handle throttle
	if utils.IsKnown(m.Throttle) {
		throttle := m.Throttle.ValueString()
		body.Throttle = &throttle
	}

	// Handle tags
	if utils.IsKnown(m.Tags) && len(m.Tags.Elements()) > 0 {
		tags := utils.ListTypeToSlice_String(ctx, m.Tags, path.Root("tags"), &diags)
		body.Tags = &tags
	}

	// Handle alert_delay
	if utils.IsKnown(m.AlertDelay) {
		body.AlertDelay = &struct {
			Active float32 `json:"active"`
		}{
			Active: float32(m.AlertDelay.ValueFloat64()),
		}
	}

	// Handle actions
	if utils.IsKnown(m.Actions) && len(m.Actions.Elements()) > 0 {
		actions := utils.ListTypeToSlice(ctx, m.Actions, path.Root("actions"), &diags,
			func(item actionModel, meta utils.ListMeta) struct {
				AlertsFilter            *struct{}               `json:"alerts_filter,omitempty"`
				Frequency               *struct{}               `json:"frequency,omitempty"`
				Group                   *string                 `json:"group,omitempty"`
				Id                      string                  `json:"id"`
				Params                  *map[string]interface{} `json:"params,omitempty"`
				UseAlertDataForTemplate *bool                   `json:"use_alert_data_for_template,omitempty"`
				Uuid                    *string                 `json:"uuid,omitempty"`
			} {
				action := struct {
					AlertsFilter            *struct{}               `json:"alerts_filter,omitempty"`
					Frequency               *struct{}               `json:"frequency,omitempty"`
					Group                   *string                 `json:"group,omitempty"`
					Id                      string                  `json:"id"`
					Params                  *map[string]interface{} `json:"params,omitempty"`
					UseAlertDataForTemplate *bool                   `json:"use_alert_data_for_template,omitempty"`
					Uuid                    *string                 `json:"uuid,omitempty"`
				}{
					Group: utils.ValueStringPointer(item.Group),
					Id:    item.ID.ValueString(),
				}

				if utils.IsKnown(item.Params) {
					var params map[string]interface{}
					if err := json.Unmarshal([]byte(item.Params.ValueString()), &params); err != nil {
						meta.Diags.AddError("Error unmarshaling action params", err.Error())
						return action
					}
					action.Params = &params
				}

				return action
			})

		// Convert to the actual API type - we need to build this properly
		apiActions := make([]struct {
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
		}, len(actions))

		actionsModels := utils.ListTypeToSlice(ctx, m.Actions, path.Root("actions"), &diags, func(item actionModel, meta utils.ListMeta) actionModel { return item })

		for i, action := range actionsModels {
			apiActions[i].Group = utils.ValueStringPointer(action.Group)
			apiActions[i].Id = action.ID.ValueString()

			if utils.IsKnown(action.Params) {
				var params map[string]interface{}
				if err := json.Unmarshal([]byte(action.Params.ValueString()), &params); err != nil {
					diags.AddError("Error unmarshaling action params", err.Error())
					continue
				}
				apiActions[i].Params = &params
			}

			// Handle frequency
			if utils.IsKnown(action.Frequency) {
				freq := utils.ObjectTypeAs[frequencyModel](ctx, action.Frequency, path.Root("actions").AtListIndex(i).AtName("frequency"), &diags)
				if freq != nil {
					apiActions[i].Frequency = &struct {
						NotifyWhen kbapi.PostAlertingRuleIdJSONBodyActionsFrequencyNotifyWhen `json:"notify_when"`
						Summary    bool                                                       `json:"summary"`
						Throttle   *string                                                    `json:"throttle,omitempty"`
					}{
						NotifyWhen: kbapi.PostAlertingRuleIdJSONBodyActionsFrequencyNotifyWhen(freq.NotifyWhen.ValueString()),
						Summary:    freq.Summary.ValueBool(),
					}
					if utils.IsKnown(freq.Throttle) {
						throttle := freq.Throttle.ValueString()
						apiActions[i].Frequency.Throttle = &throttle
					}
				}
			}

			// Handle alerts_filter
			if utils.IsKnown(action.AlertsFilter) {
				filter := utils.ObjectTypeAs[alertsFilterModel](ctx, action.AlertsFilter, path.Root("actions").AtListIndex(i).AtName("alerts_filter"), &diags)
				if filter != nil {
					apiActions[i].AlertsFilter = &struct {
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

					if utils.IsKnown(filter.Kql) {
						apiActions[i].AlertsFilter.Query = &struct {
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
							Kql: filter.Kql.ValueString(),
						}
					}

					if utils.IsKnown(filter.Timeframe) {
						tf := utils.ObjectTypeAs[timeframeModel](ctx, filter.Timeframe, path.Root("actions").AtListIndex(i).AtName("alerts_filter").AtName("timeframe"), &diags)
						if tf != nil {
							days := utils.ListTypeToSlice(ctx, tf.Days, path.Root("actions").AtListIndex(i).AtName("alerts_filter").AtName("timeframe").AtName("days"), &diags,
								func(item types.Int64, meta utils.ListMeta) kbapi.PostAlertingRuleIdJSONBodyActionsAlertsFilterTimeframeDays {
									return kbapi.PostAlertingRuleIdJSONBodyActionsAlertsFilterTimeframeDays(item.ValueInt64())
								})

							apiActions[i].AlertsFilter.Timeframe = &struct {
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
									End:   tf.HoursEnd.ValueString(),
									Start: tf.HoursStart.ValueString(),
								},
								Timezone: tf.Timezone.ValueString(),
							}
						}
					}
				}
			}
		}

		body.Actions = &apiActions
	}

	return body, diags
}

func (m alertingRuleModel) toAPIUpdateModel(ctx context.Context) (kbapi.PutAlertingRuleIdJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics

	body := kbapi.PutAlertingRuleIdJSONRequestBody{
		Name: m.Name.ValueString(),
		Schedule: struct {
			Interval string `json:"interval"`
		}{
			Interval: m.Interval.ValueString(),
		},
	}

	// Handle params
	if utils.IsKnown(m.Params) {
		var params map[string]interface{}
		if err := json.Unmarshal([]byte(m.Params.ValueString()), &params); err != nil {
			diags.AddError("Error unmarshaling params", err.Error())
			return body, diags
		}
		body.Params = &params
	}

	// Handle notify_when
	if utils.IsKnown(m.NotifyWhen) {
		notifyWhen := kbapi.PutAlertingRuleIdJSONBodyNotifyWhen(m.NotifyWhen.ValueString())
		body.NotifyWhen = &notifyWhen
	}

	// Handle throttle
	if utils.IsKnown(m.Throttle) {
		throttle := m.Throttle.ValueString()
		body.Throttle = &throttle
	}

	// Handle tags
	if utils.IsKnown(m.Tags) && len(m.Tags.Elements()) > 0 {
		tags := utils.ListTypeToSlice_String(ctx, m.Tags, path.Root("tags"), &diags)
		body.Tags = &tags
	}

	// Handle alert_delay
	if utils.IsKnown(m.AlertDelay) {
		body.AlertDelay = &struct {
			Active float32 `json:"active"`
		}{
			Active: float32(m.AlertDelay.ValueFloat64()),
		}
	}

	// Handle actions - similar to create
	if utils.IsKnown(m.Actions) && len(m.Actions.Elements()) > 0 {
		apiActions := make([]struct {
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
		}, 0)

		actionsModels := utils.ListTypeToSlice(ctx, m.Actions, path.Root("actions"), &diags, func(item actionModel, meta utils.ListMeta) actionModel { return item })

		for i, action := range actionsModels {
			apiAction := struct {
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
				Group: utils.ValueStringPointer(action.Group),
				Id:    action.ID.ValueString(),
			}

			if utils.IsKnown(action.Params) {
				var params map[string]interface{}
				if err := json.Unmarshal([]byte(action.Params.ValueString()), &params); err != nil {
					diags.AddError("Error unmarshaling action params", err.Error())
					continue
				}
				apiAction.Params = &params
			}

			// Handle frequency
			if utils.IsKnown(action.Frequency) {
				freq := utils.ObjectTypeAs[frequencyModel](ctx, action.Frequency, path.Root("actions").AtListIndex(i).AtName("frequency"), &diags)
				if freq != nil {
					apiAction.Frequency = &struct {
						NotifyWhen kbapi.PutAlertingRuleIdJSONBodyActionsFrequencyNotifyWhen `json:"notify_when"`
						Summary    bool                                                      `json:"summary"`
						Throttle   *string                                                   `json:"throttle,omitempty"`
					}{
						NotifyWhen: kbapi.PutAlertingRuleIdJSONBodyActionsFrequencyNotifyWhen(freq.NotifyWhen.ValueString()),
						Summary:    freq.Summary.ValueBool(),
					}
					if utils.IsKnown(freq.Throttle) {
						throttle := freq.Throttle.ValueString()
						apiAction.Frequency.Throttle = &throttle
					}
				}
			}

			// Handle alerts_filter
			if utils.IsKnown(action.AlertsFilter) {
				filter := utils.ObjectTypeAs[alertsFilterModel](ctx, action.AlertsFilter, path.Root("actions").AtListIndex(i).AtName("alerts_filter"), &diags)
				if filter != nil {
					apiAction.AlertsFilter = &struct {
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

					if utils.IsKnown(filter.Kql) {
						apiAction.AlertsFilter.Query = &struct {
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
							Kql: filter.Kql.ValueString(),
						}
					}

					if utils.IsKnown(filter.Timeframe) {
						tf := utils.ObjectTypeAs[timeframeModel](ctx, filter.Timeframe, path.Root("actions").AtListIndex(i).AtName("alerts_filter").AtName("timeframe"), &diags)
						if tf != nil {
							days := utils.ListTypeToSlice(ctx, tf.Days, path.Root("actions").AtListIndex(i).AtName("alerts_filter").AtName("timeframe").AtName("days"), &diags,
								func(item types.Int64, meta utils.ListMeta) kbapi.PutAlertingRuleIdJSONBodyActionsAlertsFilterTimeframeDays {
									return kbapi.PutAlertingRuleIdJSONBodyActionsAlertsFilterTimeframeDays(item.ValueInt64())
								})

							apiAction.AlertsFilter.Timeframe = &struct {
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
									End:   tf.HoursEnd.ValueString(),
									Start: tf.HoursStart.ValueString(),
								},
								Timezone: tf.Timezone.ValueString(),
							}
						}
					}
				}
			}

			apiActions = append(apiActions, apiAction)
		}

		body.Actions = &apiActions
	}

	return body, diags
}

func (m *alertingRuleModel) getRuleIDAndSpaceID() (ruleID string, spaceID string) {
	ruleID = m.RuleID.ValueString()
	spaceID = m.SpaceID.ValueString()

	// Try to parse composite ID if present
	resourceID := m.ID.ValueString()
	if resourceID != "" {
		maybeCompositeID, _ := clients.CompositeIdFromStr(resourceID)
		if maybeCompositeID != nil {
			ruleID = maybeCompositeID.ResourceId
			spaceID = maybeCompositeID.ClusterId
		}
	}

	return
}
