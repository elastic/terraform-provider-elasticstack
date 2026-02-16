package alerting_rule

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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
	Tags                types.Set            `tfsdk:"tags"`
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
	Frequency    types.Object         `tfsdk:"frequency"`
	AlertsFilter types.Object         `tfsdk:"alerts_filter"`
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
	Timeframe types.Object `tfsdk:"timeframe"`
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
	previousParams := m.Params

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
	normalizedParams, d := normalizeRuleParamsForState(ctx, rule.Params, previousParams)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	paramsJSON, err := json.Marshal(normalizedParams)
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
		tags, d := types.SetValueFrom(ctx, types.StringType, rule.Tags)
		diags.Append(d...)
		m.Tags = tags
	} else {
		m.Tags = types.SetNull(types.StringType)
	}

	// Throttle
	if rule.Throttle != nil {
		m.Throttle = types.StringValue(*rule.Throttle)
	} else {
		m.Throttle = types.StringNull()
	}

	// Scheduled task ID - update if API returns a value, or resolve unknown to null
	// (preserves existing known value when API doesn't return this field on re-reads)
	if rule.ScheduledTaskID != nil {
		m.ScheduledTaskID = types.StringValue(*rule.ScheduledTaskID)
	} else if m.ScheduledTaskID.IsUnknown() {
		m.ScheduledTaskID = types.StringNull()
	}

	// Execution status
	m.LastExecutionStatus = types.StringPointerValue(rule.ExecutionStatus.Status)

	if rule.ExecutionStatus.LastExecutionDate != nil {
		m.LastExecutionDate = types.StringValue(rule.ExecutionStatus.LastExecutionDate.Format("2006-01-02 15:04:05.999 -0700 MST"))
	} else {
		m.LastExecutionDate = types.StringNull()
	}

	// Alert delay - update if API returns a value, or resolve unknown to null
	// (preserves existing known value when API doesn't return this field on re-reads)
	if rule.AlertDelay != nil {
		m.AlertDelay = types.Int64Value(int64(*rule.AlertDelay))
	} else if m.AlertDelay.IsUnknown() {
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

// normalizeRuleParamsForState strips API-returned params keys that the user
// never specified. Kibana injects server-side defaults (e.g. aggType, groupBy)
// into the response even when the user's config omitted them. Without this,
// Terraform sees the extra keys as drift and produces "inconsistent result
// after apply" errors. The approach is generic: any key present in the API
// response but absent from the user's prior state params is removed.
func normalizeRuleParamsForState(ctx context.Context, apiParams map[string]interface{}, previousParams jsontypes.Normalized) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	if apiParams == nil {
		return apiParams, diags
	}

	priorParams, d := parsePriorParams(previousParams)
	diags.Append(d...)
	if diags.HasError() {
		return apiParams, diags
	}
	if priorParams == nil {
		// No prior state to compare against (first create); keep everything.
		return apiParams, diags
	}

	normalized, ok := removeInjectedDefaultsRecursive(apiParams, priorParams).(map[string]interface{})
	if !ok {
		// Defensive fallback: params are expected to be a JSON object, but if we
		// can't reconcile shapes, keep everything to avoid accidental data loss.
		return apiParams, diags
	}

	return normalized, diags
}

// parsePriorParams returns the decoded params JSON from the previous Terraform
// state, or nil if there is no usable prior state.
func parsePriorParams(previousParams jsontypes.Normalized) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !utils.IsKnown(previousParams) || previousParams.IsNull() {
		return nil, diags
	}

	var prior map[string]interface{}
	diags.Append(previousParams.Unmarshal(&prior)...)
	if diags.HasError() {
		return nil, diags
	}

	return prior, diags
}

// removeInjectedDefaultsRecursive removes API-injected keys at any nesting level.
//
// It walks the API params structure and removes keys that are present in the API
// payload but absent from the corresponding object in the prior state params.
//
// NOTE: This function assumes the prior state params represent what the user
// configured (i.e. prior-state keys are a subset of configured keys). It only
// iterates keys present in the API payload; therefore, keys present in the prior
// state but missing from the API response will not be preserved in the output.
//
//   - For JSON objects (map[string]interface{}), keys not present in the prior
//     object are dropped, and shared keys are recursed into.
//   - For JSON arrays ([]interface{}), elements are matched by index when the prior
//     is also an array; element values are recursed into when there is a matching
//     prior element, otherwise the element is kept as-is (defensive fallback).
//   - For scalars, the API value is returned unchanged.
func removeInjectedDefaultsRecursive(api any, prior any) any {
	switch apiTyped := api.(type) {
	case map[string]interface{}:
		priorMap, ok := prior.(map[string]interface{})
		if !ok {
			// Can't compare nested keys without a prior object; keep as-is.
			return apiTyped
		}

		out := make(map[string]interface{}, len(apiTyped))
		for k, v := range apiTyped {
			priorV, exists := priorMap[k]
			if !exists {
				// API injected this key but user never had it — drop it.
				continue
			}
			out[k] = removeInjectedDefaultsRecursive(v, priorV)
		}
		return out

	case []interface{}:
		priorSlice, ok := prior.([]interface{})
		if !ok {
			// Can't compare element structure without a prior array; keep as-is.
			return apiTyped
		}

		out := make([]interface{}, len(apiTyped))
		for i, v := range apiTyped {
			if i < len(priorSlice) {
				out[i] = removeInjectedDefaultsRecursive(v, priorSlice[i])
			} else {
				// Defensive fallback: no matching prior element; keep.
				out[i] = v
			}
		}
		return out

	default:
		return api
	}
}

// Version thresholds for feature support
var (
	frequencyMinSupportedVersion    = version.Must(version.NewVersion("8.6.0"))
	alertsFilterMinSupportedVersion = version.Must(version.NewVersion("8.9.0"))
	alertDelayMinSupportedVersion   = version.Must(version.NewVersion("8.13.0"))
)

// toAPIModel converts the Terraform model to the API model.
// It also validates version-specific requirements based on the provided server version.
func (m alertingRuleModel) toAPIModel(ctx context.Context, serverVersion *version.Version) (models.AlertingRule, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Validate version-specific requirements
	if serverVersion != nil {
		// notify_when is required until v8.6
		if !utils.IsKnown(m.NotifyWhen) || m.NotifyWhen.ValueString() == "" {
			if serverVersion.LessThan(frequencyMinSupportedVersion) {
				diags.AddError(
					"notify_when is required until v8.6",
					"notify_when is required until v8.6",
				)
				return models.AlertingRule{}, diags
			}
		}

		// alert_delay is only supported from v8.13+
		if utils.IsKnown(m.AlertDelay) && !m.AlertDelay.IsNull() {
			if serverVersion.LessThan(alertDelayMinSupportedVersion) {
				diags.AddError(
					"alert_delay is only supported for Elasticsearch v8.13 or higher",
					"alert_delay is only supported for Elasticsearch v8.13 or higher",
				)
				return models.AlertingRule{}, diags
			}
		}

		// Validate version-specific requirements for actions
		if utils.IsKnown(m.Actions) && !m.Actions.IsNull() {
			var actions []actionModel
			diags.Append(m.Actions.ElementsAs(ctx, &actions, false)...)
			if diags.HasError() {
				return models.AlertingRule{}, diags
			}

			for _, action := range actions {
				// Check frequency version requirement
				if utils.IsKnown(action.Frequency) && !action.Frequency.IsNull() {
					if serverVersion.LessThan(frequencyMinSupportedVersion) {
						diags.AddError(
							"actions.frequency is only supported for Kibana v8.6 or higher",
							"actions.frequency is only supported for Kibana v8.6 or higher",
						)
						return models.AlertingRule{}, diags
					}
				}

				// Check alerts_filter version requirement
				if utils.IsKnown(action.AlertsFilter) && !action.AlertsFilter.IsNull() {
					if serverVersion.LessThan(alertsFilterMinSupportedVersion) {
						diags.AddError(
							"actions.alerts_filter is only supported for Kibana v8.9 or higher",
							"actions.alerts_filter is only supported for Kibana v8.9 or higher",
						)
						return models.AlertingRule{}, diags
					}
				}
			}
		}
	}

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

	// Params from JSON string.
	// Note: params validation is handled exclusively by ValidateConfig during
	// the plan phase. We intentionally do not re-validate here to avoid
	// duplicate error messages.
	if utils.IsKnown(m.Params) {
		params := map[string]interface{}{}
		diags.Append(m.Params.Unmarshal(&params)...)
		if diags.HasError() {
			return models.AlertingRule{}, diags
		}

		// Compatibility: older Kibana versions reject `.index-threshold` rule params
		// when `groupBy` is omitted (server-side expects a string, but sees undefined).
		// Defaulting to "all" preserves Kibana's effective behavior while avoiding 400s.
		if rule.RuleTypeID == ".index-threshold" {
			if v, ok := params["groupBy"]; !ok || v == nil {
				params["groupBy"] = "all"
			}
		}

		rule.Params = params
	}

	// Enabled
	if utils.IsKnown(m.Enabled) {
		rule.Enabled = m.Enabled.ValueBoolPointer()
	}

	// NotifyWhen
	if utils.IsKnown(m.NotifyWhen) && m.NotifyWhen.ValueString() != "" {
		rule.NotifyWhen = m.NotifyWhen.ValueStringPointer()
	}

	// Throttle
	if utils.IsKnown(m.Throttle) && m.Throttle.ValueString() != "" {
		rule.Throttle = m.Throttle.ValueStringPointer()
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

// getRuleIDAndSpaceID extracts rule ID and space ID from the composite ID.
// The state upgrade ensures the ID is always in composite format.
func (m alertingRuleModel) getRuleIDAndSpaceID() (ruleID string, spaceID string) {
	compositeID, _ := clients.CompositeIdFromStr(m.ID.ValueString())
	return compositeID.ResourceId, compositeID.ClusterId
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

		// Frequency - convert to single object
		if apiAction.Frequency != nil {
			freq := frequencyModel{
				Summary:    types.BoolValue(apiAction.Frequency.Summary),
				NotifyWhen: types.StringValue(apiAction.Frequency.NotifyWhen),
				Throttle:   types.StringPointerValue(apiAction.Frequency.Throttle),
			}
			freqObj, d := types.ObjectValueFrom(ctx, getFrequencyAttrTypes(), freq)
			diags.Append(d...)
			action.Frequency = freqObj
		} else {
			action.Frequency = types.ObjectNull(getFrequencyAttrTypes())
		}

		// Alerts filter - convert to single object
		if apiAction.AlertsFilter != nil {
			filter := alertsFilterModel{}

			filter.Kql = types.StringPointerValue(apiAction.AlertsFilter.Kql)

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
				tfObj, d := types.ObjectValueFrom(ctx, getTimeframeAttrTypes(), timeframe)
				diags.Append(d...)
				filter.Timeframe = tfObj
			} else {
				filter.Timeframe = types.ObjectNull(getTimeframeAttrTypes())
			}

			filterObj, d := types.ObjectValueFrom(ctx, getAlertsFilterAttrTypes(), filter)
			diags.Append(d...)
			action.AlertsFilter = filterObj
		} else {
			action.AlertsFilter = types.ObjectNull(getAlertsFilterAttrTypes())
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

	if !utils.IsKnown(actionsList) {
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

		// Frequency - extract from object
		if utils.IsKnown(action.Frequency) && !action.Frequency.IsNull() {
			var freq frequencyModel
			diags.Append(action.Frequency.As(ctx, &freq, basetypes.ObjectAsOptions{})...)
			// Only create Frequency if both required fields are present
			if utils.IsKnown(freq.Summary) && utils.IsKnown(freq.NotifyWhen) {
				apiAction.Frequency = &models.ActionFrequency{
					Summary:    freq.Summary.ValueBool(),
					NotifyWhen: freq.NotifyWhen.ValueString(),
				}
				if utils.IsKnown(freq.Throttle) && freq.Throttle.ValueString() != "" {
					apiAction.Frequency.Throttle = freq.Throttle.ValueStringPointer()
				}
			}
		}

		// Alerts filter - extract from object
		if utils.IsKnown(action.AlertsFilter) && !action.AlertsFilter.IsNull() {
			var filter alertsFilterModel
			diags.Append(action.AlertsFilter.As(ctx, &filter, basetypes.ObjectAsOptions{})...)
			apiAction.AlertsFilter = &models.ActionAlertsFilter{}

			if utils.IsKnown(filter.Kql) {
				kql := filter.Kql.ValueString()
				apiAction.AlertsFilter.Kql = &kql
			}

			if utils.IsKnown(filter.Timeframe) && !filter.Timeframe.IsNull() {
				var tf timeframeModel
				diags.Append(filter.Timeframe.As(ctx, &tf, basetypes.ObjectAsOptions{})...)
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

		apiActions = append(apiActions, apiAction)
	}

	return apiActions, diags
}
