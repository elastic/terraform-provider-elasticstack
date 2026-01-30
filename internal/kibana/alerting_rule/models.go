package alerting_rule

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type tfModel struct {
	ID                  types.String         `tfsdk:"id"`
	KibanaConnection    types.List           `tfsdk:"kibana_connection"`
	RuleID              types.String         `tfsdk:"rule_id"`
	SpaceID             types.String         `tfsdk:"space_id"`
	Name                types.String         `tfsdk:"name"`
	Consumer            types.String         `tfsdk:"consumer"`
	NotifyWhen          types.String         `tfsdk:"notify_when"`
	Params              jsontypes.Normalized `tfsdk:"params"`
	RuleTypeID          types.String         `tfsdk:"rule_type_id"`
	Interval            types.String         `tfsdk:"interval"`
	Actions             types.List           `tfsdk:"actions"`
	Enabled             types.Bool           `tfsdk:"enabled"`
	Tags                types.List           `tfsdk:"tags"`
	Throttle            types.String         `tfsdk:"throttle"`
	ScheduledTaskID     types.String         `tfsdk:"scheduled_task_id"`
	LastExecutionStatus types.String         `tfsdk:"last_execution_status"`
	LastExecutionDate   types.String         `tfsdk:"last_execution_date"`
	AlertDelay          types.Float64        `tfsdk:"alert_delay"`
}

type actionTfModel struct {
	Group        types.String         `tfsdk:"group"`
	ID           types.String         `tfsdk:"id"`
	Params       jsontypes.Normalized `tfsdk:"params"`
	Frequency    *frequencyTfModel    `tfsdk:"frequency"`
	AlertsFilter *alertsFilterTfModel `tfsdk:"alerts_filter"`
}

type frequencyTfModel struct {
	Summary    types.Bool   `tfsdk:"summary"`
	NotifyWhen types.String `tfsdk:"notify_when"`
	Throttle   types.String `tfsdk:"throttle"`
}

type alertsFilterTfModel struct {
	Kql       types.String      `tfsdk:"kql"`
	Timeframe *timeframeTfModel `tfsdk:"timeframe"`
}

type timeframeTfModel struct {
	Days       types.List   `tfsdk:"days"`
	Timezone   types.String `tfsdk:"timezone"`
	HoursStart types.String `tfsdk:"hours_start"`
	HoursEnd   types.String `tfsdk:"hours_end"`
}

func (model tfModel) GetID() (*clients.CompositeId, diag.Diagnostics) {
	compId, sdkDiags := clients.CompositeIdFromStr(model.ID.ValueString())
	if sdkDiags.HasError() {
		var diags diag.Diagnostics
		for _, d := range sdkDiags {
			diags.AddError(d.Summary, d.Detail)
		}
		return nil, diags
	}

	return compId, nil
}

func (model tfModel) toAPIModel(ctx context.Context) (models.AlertingRule, diag.Diagnostics) {
	var diags diag.Diagnostics

	rule := models.AlertingRule{
		SpaceID:    model.SpaceID.ValueString(),
		Name:       model.Name.ValueString(),
		Consumer:   model.Consumer.ValueString(),
		RuleTypeID: model.RuleTypeID.ValueString(),
		Schedule: models.AlertingRuleSchedule{
			Interval: model.Interval.ValueString(),
		},
	}

	// Set rule ID if provided
	if utils.IsKnown(model.RuleID) && model.RuleID.ValueString() != "" {
		rule.RuleID = model.RuleID.ValueString()
	}

	// Parse params JSON
	if utils.IsKnown(model.Params) {
		params := map[string]interface{}{}
		if err := json.Unmarshal([]byte(model.Params.ValueString()), &params); err != nil {
			diags.AddError("Failed to parse params JSON", err.Error())
			return models.AlertingRule{}, diags
		}
		rule.Params = params
	}

	// Enabled
	if utils.IsKnown(model.Enabled) {
		enabled := model.Enabled.ValueBool()
		rule.Enabled = &enabled
	}

	// Throttle
	if utils.IsKnown(model.Throttle) {
		throttle := model.Throttle.ValueString()
		rule.Throttle = &throttle
	}

	// NotifyWhen
	if utils.IsKnown(model.NotifyWhen) && model.NotifyWhen.ValueString() != "" {
		notifyWhen := model.NotifyWhen.ValueString()
		rule.NotifyWhen = &notifyWhen
	}

	// AlertDelay
	if utils.IsKnown(model.AlertDelay) {
		alertDelay := float32(model.AlertDelay.ValueFloat64())
		rule.AlertDelay = &alertDelay
	}

	// Tags
	if utils.IsKnown(model.Tags) && !model.Tags.IsNull() {
		var tags []string
		diags.Append(model.Tags.ElementsAs(ctx, &tags, false)...)
		if diags.HasError() {
			return models.AlertingRule{}, diags
		}
		rule.Tags = tags
	}

	// Actions
	if utils.IsKnown(model.Actions) && !model.Actions.IsNull() {
		var actionModels []actionTfModel
		diags.Append(model.Actions.ElementsAs(ctx, &actionModels, false)...)
		if diags.HasError() {
			return models.AlertingRule{}, diags
		}

		actions := make([]models.AlertingRuleAction, 0, len(actionModels))
		for _, actionModel := range actionModels {
			action, actionDiags := actionModel.toAPIModel(ctx)
			diags.Append(actionDiags...)
			if diags.HasError() {
				return models.AlertingRule{}, diags
			}
			actions = append(actions, action)
		}
		rule.Actions = actions
	}

	return rule, diags
}

func (actionModel actionTfModel) toAPIModel(ctx context.Context) (models.AlertingRuleAction, diag.Diagnostics) {
	var diags diag.Diagnostics

	action := models.AlertingRuleAction{
		Group: actionModel.Group.ValueString(),
		ID:    actionModel.ID.ValueString(),
	}

	// Parse params JSON
	if utils.IsKnown(actionModel.Params) {
		params := map[string]interface{}{}
		if err := json.Unmarshal([]byte(actionModel.Params.ValueString()), &params); err != nil {
			diags.AddError("Failed to parse action params JSON", err.Error())
			return models.AlertingRuleAction{}, diags
		}
		action.Params = params
	}

	// Frequency
	if actionModel.Frequency != nil {
		// Validate that summary and notify_when are both provided when frequency block is specified
		if actionModel.Frequency.Summary.IsNull() || actionModel.Frequency.Summary.IsUnknown() {
			diags.AddError("Missing required attribute", "The 'summary' attribute is required when 'frequency' block is specified")
			return models.AlertingRuleAction{}, diags
		}
		if actionModel.Frequency.NotifyWhen.IsNull() || actionModel.Frequency.NotifyWhen.IsUnknown() || actionModel.Frequency.NotifyWhen.ValueString() == "" {
			diags.AddError("Missing required attribute", "The 'notify_when' attribute is required when 'frequency' block is specified")
			return models.AlertingRuleAction{}, diags
		}

		frequency := models.ActionFrequency{
			Summary:    actionModel.Frequency.Summary.ValueBool(),
			NotifyWhen: actionModel.Frequency.NotifyWhen.ValueString(),
		}
		if utils.IsKnown(actionModel.Frequency.Throttle) && actionModel.Frequency.Throttle.ValueString() != "" {
			throttle := actionModel.Frequency.Throttle.ValueString()
			frequency.Throttle = &throttle
		}
		action.Frequency = &frequency
	}

	// AlertsFilter
	if actionModel.AlertsFilter != nil {
		filter := models.ActionAlertsFilter{}

		if utils.IsKnown(actionModel.AlertsFilter.Kql) {
			kql := actionModel.AlertsFilter.Kql.ValueString()
			filter.Kql = &kql
		}

		if actionModel.AlertsFilter.Timeframe != nil {
			var days []int64
			diags.Append(actionModel.AlertsFilter.Timeframe.Days.ElementsAs(ctx, &days, false)...)
			if diags.HasError() {
				return models.AlertingRuleAction{}, diags
			}

			days32 := make([]int32, len(days))
			for i, d := range days {
				days32[i] = int32(d)
			}

			filter.Timeframe = &models.AlertsFilterTimeframe{
				Days:       days32,
				Timezone:   actionModel.AlertsFilter.Timeframe.Timezone.ValueString(),
				HoursStart: actionModel.AlertsFilter.Timeframe.HoursStart.ValueString(),
				HoursEnd:   actionModel.AlertsFilter.Timeframe.HoursEnd.ValueString(),
			}
		}

		action.AlertsFilter = &filter
	}

	return action, diags
}

func (model *tfModel) populateFromAPI(ctx context.Context, apiModel *models.AlertingRule, compositeID *clients.CompositeId) diag.Diagnostics {
	var diags diag.Diagnostics

	model.ID = types.StringValue(compositeID.String())
	model.RuleID = types.StringValue(apiModel.RuleID)
	model.SpaceID = types.StringValue(apiModel.SpaceID)
	model.Name = types.StringValue(apiModel.Name)
	model.Consumer = types.StringValue(apiModel.Consumer)
	model.RuleTypeID = types.StringValue(apiModel.RuleTypeID)
	model.Interval = types.StringValue(apiModel.Schedule.Interval)

	// Enabled
	if apiModel.Enabled != nil {
		model.Enabled = types.BoolValue(*apiModel.Enabled)
	}

	// NotifyWhen
	if apiModel.NotifyWhen != nil {
		model.NotifyWhen = types.StringValue(*apiModel.NotifyWhen)
	} else {
		model.NotifyWhen = types.StringNull()
	}

	// Throttle
	if apiModel.Throttle != nil {
		model.Throttle = types.StringValue(*apiModel.Throttle)
	} else {
		model.Throttle = types.StringNull()
	}

	// ScheduledTaskID
	if apiModel.ScheduledTaskID != nil {
		model.ScheduledTaskID = types.StringValue(*apiModel.ScheduledTaskID)
	} else {
		model.ScheduledTaskID = types.StringNull()
	}

	// Execution status
	if apiModel.ExecutionStatus.Status != nil {
		model.LastExecutionStatus = types.StringValue(*apiModel.ExecutionStatus.Status)
	} else {
		model.LastExecutionStatus = types.StringNull()
	}

	if apiModel.ExecutionStatus.LastExecutionDate != nil {
		model.LastExecutionDate = types.StringValue(apiModel.ExecutionStatus.LastExecutionDate.Format("2006-01-02 15:04:05.999 -0700 MST"))
	} else {
		model.LastExecutionDate = types.StringNull()
	}

	// AlertDelay
	if apiModel.AlertDelay != nil {
		model.AlertDelay = types.Float64Value(float64(*apiModel.AlertDelay))
	} else {
		model.AlertDelay = types.Float64Null()
	}

	// Params
	paramsBytes, err := json.Marshal(apiModel.Params)
	if err != nil {
		diags.AddError("Failed to marshal params", err.Error())
		return diags
	}
	model.Params = jsontypes.NewNormalizedValue(string(paramsBytes))

	// Tags
	if len(apiModel.Tags) > 0 {
		tagValues := make([]attr.Value, len(apiModel.Tags))
		for i, tag := range apiModel.Tags {
			tagValues[i] = types.StringValue(tag)
		}
		model.Tags = types.ListValueMust(types.StringType, tagValues)
	} else {
		model.Tags = types.ListNull(types.StringType)
	}

	// Actions
	if len(apiModel.Actions) > 0 {
		actionValues, actionDiags := actionsFromAPI(ctx, apiModel.Actions)
		diags.Append(actionDiags...)
		if diags.HasError() {
			return diags
		}
		model.Actions = actionValues
	} else {
		model.Actions = types.ListNull(actionObjectType())
	}

	return diags
}

func actionsFromAPI(ctx context.Context, apiActions []models.AlertingRuleAction) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	actionValues := make([]attr.Value, len(apiActions))
	for i, apiAction := range apiActions {
		// Params
		paramsBytes, err := json.Marshal(apiAction.Params)
		if err != nil {
			diags.AddError("Failed to marshal action params", err.Error())
			return types.ListNull(actionObjectType()), diags
		}

		actionAttrs := map[string]attr.Value{
			"group":  types.StringValue(apiAction.Group),
			"id":     types.StringValue(apiAction.ID),
			"params": jsontypes.NewNormalizedValue(string(paramsBytes)),
		}

		// Frequency
		if apiAction.Frequency != nil {
			frequencyAttrs := map[string]attr.Value{
				"summary":     types.BoolValue(apiAction.Frequency.Summary),
				"notify_when": types.StringValue(apiAction.Frequency.NotifyWhen),
			}
			if apiAction.Frequency.Throttle != nil {
				frequencyAttrs["throttle"] = types.StringValue(*apiAction.Frequency.Throttle)
			} else {
				frequencyAttrs["throttle"] = types.StringNull()
			}
			actionAttrs["frequency"] = types.ObjectValueMust(frequencyObjectType().AttrTypes, frequencyAttrs)
		} else {
			actionAttrs["frequency"] = types.ObjectNull(frequencyObjectType().AttrTypes)
		}

		// AlertsFilter
		if apiAction.AlertsFilter != nil {
			filterAttrs := map[string]attr.Value{}

			if apiAction.AlertsFilter.Kql != nil {
				filterAttrs["kql"] = types.StringValue(*apiAction.AlertsFilter.Kql)
			} else {
				filterAttrs["kql"] = types.StringNull()
			}

			if apiAction.AlertsFilter.Timeframe != nil {
				days := make([]attr.Value, len(apiAction.AlertsFilter.Timeframe.Days))
				for j, d := range apiAction.AlertsFilter.Timeframe.Days {
					days[j] = types.Int64Value(int64(d))
				}

				timeframeAttrs := map[string]attr.Value{
					"days":        types.ListValueMust(types.Int64Type, days),
					"timezone":    types.StringValue(apiAction.AlertsFilter.Timeframe.Timezone),
					"hours_start": types.StringValue(apiAction.AlertsFilter.Timeframe.HoursStart),
					"hours_end":   types.StringValue(apiAction.AlertsFilter.Timeframe.HoursEnd),
				}
				filterAttrs["timeframe"] = types.ObjectValueMust(timeframeObjectType().AttrTypes, timeframeAttrs)
			} else {
				filterAttrs["timeframe"] = types.ObjectNull(timeframeObjectType().AttrTypes)
			}

			actionAttrs["alerts_filter"] = types.ObjectValueMust(alertsFilterObjectType().AttrTypes, filterAttrs)
		} else {
			actionAttrs["alerts_filter"] = types.ObjectNull(alertsFilterObjectType().AttrTypes)
		}

		actionValues[i] = types.ObjectValueMust(actionObjectType().AttrTypes, actionAttrs)
	}

	return types.ListValueMust(actionObjectType(), actionValues), diags
}

func actionObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"group":         types.StringType,
			"id":            types.StringType,
			"params":        jsontypes.NormalizedType{},
			"frequency":     frequencyObjectType(),
			"alerts_filter": alertsFilterObjectType(),
		},
	}
}

func frequencyObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"summary":     types.BoolType,
			"notify_when": types.StringType,
			"throttle":    types.StringType,
		},
	}
}

func alertsFilterObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"kql":       types.StringType,
			"timeframe": timeframeObjectType(),
		},
	}
}

func timeframeObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"days":        types.ListType{ElemType: types.Int64Type},
			"timezone":    types.StringType,
			"hours_start": types.StringType,
			"hours_end":   types.StringType,
		},
	}
}
