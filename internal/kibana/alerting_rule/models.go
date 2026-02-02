package alerting_rule

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// alertingRuleModel is the Terraform model for an alerting rule.
type alertingRuleModel struct {
	ID                  types.String         `tfsdk:"id"`
	RuleID              types.String         `tfsdk:"rule_id"`
	SpaceID             types.String         `tfsdk:"space_id"`
	Name                types.String         `tfsdk:"name"`
	Consumer            types.String         `tfsdk:"consumer"`
	NotifyWhen          types.String         `tfsdk:"notify_when"`
	Params              jsontypes.Normalized `tfsdk:"params"`
	RuleTypeID          types.String         `tfsdk:"rule_type_id"`
	Interval            types.String         `tfsdk:"interval"`
	Enabled             types.Bool           `tfsdk:"enabled"`
	Tags                types.List           `tfsdk:"tags"`
	Throttle            types.String         `tfsdk:"throttle"`
	ScheduledTaskID     types.String         `tfsdk:"scheduled_task_id"`
	LastExecutionStatus types.String         `tfsdk:"last_execution_status"`
	LastExecutionDate   types.String         `tfsdk:"last_execution_date"`
	AlertDelay          types.Int64          `tfsdk:"alert_delay"`
	Actions             types.List           `tfsdk:"actions"`
}

// actionModel is the Terraform model for a rule action.
type actionModel struct {
	Group        types.String         `tfsdk:"group"`
	ID           types.String         `tfsdk:"id"`
	Params       jsontypes.Normalized `tfsdk:"params"`
	Frequency    types.List           `tfsdk:"frequency"`
	AlertsFilter types.List           `tfsdk:"alerts_filter"`
}

// frequencyModel is the Terraform model for action frequency.
type frequencyModel struct {
	Summary    types.Bool   `tfsdk:"summary"`
	NotifyWhen types.String `tfsdk:"notify_when"`
	Throttle   types.String `tfsdk:"throttle"`
}

// alertsFilterModel is the Terraform model for action alerts filter.
type alertsFilterModel struct {
	Kql       types.String `tfsdk:"kql"`
	Timeframe types.List   `tfsdk:"timeframe"`
}

// timeframeModel is the Terraform model for alerts filter timeframe.
type timeframeModel struct {
	Days       types.List   `tfsdk:"days"`
	Timezone   types.String `tfsdk:"timezone"`
	HoursStart types.String `tfsdk:"hours_start"`
	HoursEnd   types.String `tfsdk:"hours_end"`
}

// populateFromAPI populates the model from the API response.
func (m *alertingRuleModel) populateFromAPI(ctx context.Context, rule *models.AlertingRule) diag.Diagnostics {
	var diags diag.Diagnostics

	if rule == nil {
		return diags
	}

	compositeID := clients.CompositeId{
		ClusterId:  rule.SpaceID,
		ResourceId: rule.RuleID,
	}

	m.ID = types.StringValue(compositeID.String())
	m.RuleID = types.StringValue(rule.RuleID)
	m.SpaceID = types.StringValue(rule.SpaceID)
	m.Name = types.StringValue(rule.Name)
	m.Consumer = types.StringValue(rule.Consumer)
	m.RuleTypeID = types.StringValue(rule.RuleTypeID)
	m.Interval = types.StringValue(rule.Schedule.Interval)

	if rule.NotifyWhen != nil && *rule.NotifyWhen != "" {
		m.NotifyWhen = types.StringValue(*rule.NotifyWhen)
	} else {
		m.NotifyWhen = types.StringNull()
	}

	// Params as JSON string
	paramsJSON, err := json.Marshal(rule.Params)
	if err != nil {
		diags.AddError("Failed to marshal params", err.Error())
		return diags
	}
	m.Params = jsontypes.NewNormalizedValue(string(paramsJSON))

	if rule.Enabled != nil {
		m.Enabled = types.BoolValue(*rule.Enabled)
	} else {
		m.Enabled = types.BoolValue(true)
	}

	// Tags
	if len(rule.Tags) > 0 {
		tags, d := types.ListValueFrom(ctx, types.StringType, rule.Tags)
		diags.Append(d...)
		m.Tags = tags
	} else {
		m.Tags = types.ListNull(types.StringType)
	}

	// Throttle
	if rule.Throttle != nil {
		m.Throttle = types.StringValue(*rule.Throttle)
	} else {
		m.Throttle = types.StringNull()
	}

	// Scheduled task ID
	if rule.ScheduledTaskID != nil {
		m.ScheduledTaskID = types.StringValue(*rule.ScheduledTaskID)
	} else {
		m.ScheduledTaskID = types.StringNull()
	}

	// Execution status
	if rule.ExecutionStatus.Status != nil {
		m.LastExecutionStatus = types.StringValue(*rule.ExecutionStatus.Status)
	} else {
		m.LastExecutionStatus = types.StringNull()
	}

	if rule.ExecutionStatus.LastExecutionDate != nil {
		m.LastExecutionDate = types.StringValue(rule.ExecutionStatus.LastExecutionDate.Format("2006-01-02 15:04:05.999 -0700 MST"))
	} else {
		m.LastExecutionDate = types.StringNull()
	}

	// Alert delay
	if rule.AlertDelay != nil {
		m.AlertDelay = types.Int64Value(int64(*rule.AlertDelay))
	} else {
		m.AlertDelay = types.Int64Null()
	}

	// Actions
	if len(rule.Actions) > 0 {
		actionsList, d := convertActionsFromAPI(ctx, rule.Actions)
		diags.Append(d...)
		m.Actions = actionsList
	} else {
		m.Actions = types.ListNull(types.ObjectType{AttrTypes: getActionsAttrTypes()})
	}

	return diags
}

// toAPIModel converts the Terraform model to the API model.
func (m alertingRuleModel) toAPIModel(ctx context.Context) (models.AlertingRule, diag.Diagnostics) {
	var diags diag.Diagnostics

	rule := models.AlertingRule{
		RuleID:     m.RuleID.ValueString(),
		SpaceID:    m.SpaceID.ValueString(),
		Name:       m.Name.ValueString(),
		Consumer:   m.Consumer.ValueString(),
		RuleTypeID: m.RuleTypeID.ValueString(),
		Schedule: models.AlertingRuleSchedule{
			Interval: m.Interval.ValueString(),
		},
	}

	// Params from JSON string
	if utils.IsKnown(m.Params) {
		params := map[string]interface{}{}
		if err := json.Unmarshal([]byte(m.Params.ValueString()), &params); err != nil {
			diags.AddError("Failed to unmarshal params", err.Error())
			return models.AlertingRule{}, diags
		}
		rule.Params = params
	}

	// Enabled
	if utils.IsKnown(m.Enabled) {
		enabled := m.Enabled.ValueBool()
		rule.Enabled = &enabled
	}

	// NotifyWhen
	if utils.IsKnown(m.NotifyWhen) && m.NotifyWhen.ValueString() != "" {
		notifyWhen := m.NotifyWhen.ValueString()
		rule.NotifyWhen = &notifyWhen
	}

	// Throttle
	if utils.IsKnown(m.Throttle) && m.Throttle.ValueString() != "" {
		throttle := m.Throttle.ValueString()
		rule.Throttle = &throttle
	}

	// Tags
	if utils.IsKnown(m.Tags) && !m.Tags.IsNull() {
		var tags []string
		diags.Append(m.Tags.ElementsAs(ctx, &tags, false)...)
		rule.Tags = tags
	}

	// Alert delay
	if utils.IsKnown(m.AlertDelay) && !m.AlertDelay.IsNull() {
		alertDelay := float32(m.AlertDelay.ValueInt64())
		rule.AlertDelay = &alertDelay
	}

	// Actions
	if utils.IsKnown(m.Actions) && !m.Actions.IsNull() {
		actions, d := convertActionsToAPI(ctx, m.Actions)
		diags.Append(d...)
		rule.Actions = actions
	}

	return rule, diags
}

// getRuleIDAndSpaceID extracts rule ID and space ID from the composite ID or model fields.
func (m alertingRuleModel) getRuleIDAndSpaceID() (ruleID string, spaceID string) {
	resourceID := m.ID.ValueString()
	maybeCompositeID, _ := clients.CompositeIdFromStr(resourceID)
	if maybeCompositeID != nil {
		ruleID = maybeCompositeID.ResourceId
		spaceID = maybeCompositeID.ClusterId
	} else {
		ruleID = m.RuleID.ValueString()
		spaceID = m.SpaceID.ValueString()
	}
	return
}

// convertActionsFromAPI converts API actions to Terraform list.
func convertActionsFromAPI(ctx context.Context, apiActions []models.AlertingRuleAction) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	actions := make([]actionModel, 0, len(apiActions))

	for _, apiAction := range apiActions {
		action := actionModel{
			Group: types.StringValue(apiAction.Group),
			ID:    types.StringValue(apiAction.ID),
		}

		// Params as JSON
		paramsJSON, err := json.Marshal(apiAction.Params)
		if err != nil {
			diags.AddError("Failed to marshal action params", err.Error())
			continue
		}
		action.Params = jsontypes.NewNormalizedValue(string(paramsJSON))

		// Frequency - convert to list with single element
		if apiAction.Frequency != nil {
			freq := frequencyModel{
				Summary:    types.BoolValue(apiAction.Frequency.Summary),
				NotifyWhen: types.StringValue(apiAction.Frequency.NotifyWhen),
			}
			if apiAction.Frequency.Throttle != nil {
				freq.Throttle = types.StringValue(*apiAction.Frequency.Throttle)
			} else {
				freq.Throttle = types.StringNull()
			}
			freqList, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: getFrequencyAttrTypes()}, []frequencyModel{freq})
			diags.Append(d...)
			action.Frequency = freqList
		} else {
			action.Frequency = types.ListNull(types.ObjectType{AttrTypes: getFrequencyAttrTypes()})
		}

		// Alerts filter - convert to list with single element
		if apiAction.AlertsFilter != nil {
			filter := alertsFilterModel{}

			if apiAction.AlertsFilter.Kql != nil {
				filter.Kql = types.StringValue(*apiAction.AlertsFilter.Kql)
			} else {
				filter.Kql = types.StringNull()
			}

			if apiAction.AlertsFilter.Timeframe != nil {
				tf := apiAction.AlertsFilter.Timeframe
				days := make([]int64, len(tf.Days))
				for i, d := range tf.Days {
					days[i] = int64(d)
				}
				daysList, d := types.ListValueFrom(ctx, types.Int64Type, days)
				diags.Append(d...)

				timeframe := timeframeModel{
					Days:       daysList,
					Timezone:   types.StringValue(tf.Timezone),
					HoursStart: types.StringValue(tf.HoursStart),
					HoursEnd:   types.StringValue(tf.HoursEnd),
				}
				tfList, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: getTimeframeAttrTypes()}, []timeframeModel{timeframe})
				diags.Append(d...)
				filter.Timeframe = tfList
			} else {
				filter.Timeframe = types.ListNull(types.ObjectType{AttrTypes: getTimeframeAttrTypes()})
			}

			filterList, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: getAlertsFilterAttrTypes()}, []alertsFilterModel{filter})
			diags.Append(d...)
			action.AlertsFilter = filterList
		} else {
			action.AlertsFilter = types.ListNull(types.ObjectType{AttrTypes: getAlertsFilterAttrTypes()})
		}

		actions = append(actions, action)
	}

	actionsList, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: getActionsAttrTypes()}, actions)
	diags.Append(d...)
	return actionsList, diags
}

// convertActionsToAPI converts Terraform actions list to API actions.
func convertActionsToAPI(ctx context.Context, actionsList types.List) ([]models.AlertingRuleAction, diag.Diagnostics) {
	var diags diag.Diagnostics

	if actionsList.IsNull() || actionsList.IsUnknown() {
		return nil, diags
	}

	var actions []actionModel
	diags.Append(actionsList.ElementsAs(ctx, &actions, false)...)
	if diags.HasError() {
		return nil, diags
	}

	apiActions := make([]models.AlertingRuleAction, 0, len(actions))
	for i, action := range actions {
		apiAction := models.AlertingRuleAction{
			Group: action.Group.ValueString(),
			ID:    action.ID.ValueString(),
		}

		// Params from JSON
		if utils.IsKnown(action.Params) {
			var params map[string]interface{}
			if err := json.Unmarshal([]byte(action.Params.ValueString()), &params); err != nil {
				diags.AddAttributeError(path.Root("actions").AtListIndex(i).AtName("params"), "Failed to unmarshal action params", err.Error())
				continue
			}
			apiAction.Params = params
		}

		// Frequency - extract first element from list
		if utils.IsKnown(action.Frequency) && !action.Frequency.IsNull() && len(action.Frequency.Elements()) > 0 {
			var freqList []frequencyModel
			diags.Append(action.Frequency.ElementsAs(ctx, &freqList, false)...)
			if len(freqList) > 0 {
				freq := freqList[0]
				apiAction.Frequency = &models.ActionFrequency{
					Summary:    freq.Summary.ValueBool(),
					NotifyWhen: freq.NotifyWhen.ValueString(),
				}
				if utils.IsKnown(freq.Throttle) && freq.Throttle.ValueString() != "" {
					throttle := freq.Throttle.ValueString()
					apiAction.Frequency.Throttle = &throttle
				}
			}
		}

		// Alerts filter - extract first element from list
		if utils.IsKnown(action.AlertsFilter) && !action.AlertsFilter.IsNull() && len(action.AlertsFilter.Elements()) > 0 {
			var filterList []alertsFilterModel
			diags.Append(action.AlertsFilter.ElementsAs(ctx, &filterList, false)...)
			if len(filterList) > 0 {
				filter := filterList[0]
				apiAction.AlertsFilter = &models.ActionAlertsFilter{}

				if utils.IsKnown(filter.Kql) {
					kql := filter.Kql.ValueString()
					apiAction.AlertsFilter.Kql = &kql
				}

				if utils.IsKnown(filter.Timeframe) && !filter.Timeframe.IsNull() && len(filter.Timeframe.Elements()) > 0 {
					var tfList []timeframeModel
					diags.Append(filter.Timeframe.ElementsAs(ctx, &tfList, false)...)
					if len(tfList) > 0 {
						tf := tfList[0]
						var days []int64
						diags.Append(tf.Days.ElementsAs(ctx, &days, false)...)

						int32Days := make([]int32, len(days))
						for j, d := range days {
							int32Days[j] = int32(d)
						}

						apiAction.AlertsFilter.Timeframe = &models.AlertsFilterTimeframe{
							Days:       int32Days,
							Timezone:   tf.Timezone.ValueString(),
							HoursStart: tf.HoursStart.ValueString(),
							HoursEnd:   tf.HoursEnd.ValueString(),
						}
					}
				}
			}
		}

		apiActions = append(apiActions, apiAction)
	}

	return apiActions, diags
}
