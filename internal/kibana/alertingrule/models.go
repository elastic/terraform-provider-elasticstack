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

package alertingrule

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/kibanacustomtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// alertingRuleModel is the Terraform model for an alerting rule.
type alertingRuleModel struct {
	entitycore.ResourceTimeoutsField
	entitycore.KibanaConnectionField
	ID     types.String `tfsdk:"id"`
	RuleID types.String `tfsdk:"rule_id"`
	entitycore.SpaceIDField
	Name            types.String                       `tfsdk:"name"`
	Consumer        types.String                       `tfsdk:"consumer"`
	NotifyWhen      types.String                       `tfsdk:"notify_when"`
	Params          jsontypes.Normalized               `tfsdk:"params"`
	RuleTypeID      types.String                       `tfsdk:"rule_type_id"`
	Interval        kibanacustomtypes.AlertingDuration `tfsdk:"interval"`
	Enabled         types.Bool                         `tfsdk:"enabled"`
	Tags            types.Set                          `tfsdk:"tags"`
	Throttle        kibanacustomtypes.AlertingDuration `tfsdk:"throttle"`
	ScheduledTaskID types.String                       `tfsdk:"scheduled_task_id"`
	AlertDelay      types.Int64                        `tfsdk:"alert_delay"`
	Flapping        types.Object                       `tfsdk:"flapping"`
	Artifacts       types.Object                       `tfsdk:"artifacts"`
	Actions         types.List                         `tfsdk:"actions"`
}

// flappingModel is the Terraform model for rule-level flapping detection.
type flappingModel struct {
	Enabled               types.Bool  `tfsdk:"enabled"`
	LookBackWindow        types.Int64 `tfsdk:"look_back_window"`
	StatusChangeThreshold types.Int64 `tfsdk:"status_change_threshold"`
}

// populateFromAPI populates the model from the API response.
func (m *alertingRuleModel) populateFromAPI(ctx context.Context, rule *models.AlertingRule) diag.Diagnostics {
	var diags diag.Diagnostics

	if rule == nil {
		return diags
	}
	previousParams := m.Params

	compositeID := clients.CompositeID{
		ClusterID:  rule.SpaceID,
		ResourceID: rule.RuleID,
	}

	m.ID = types.StringValue(compositeID.String())
	m.RuleID = types.StringValue(rule.RuleID)
	m.SpaceID = types.StringValue(rule.SpaceID)
	m.Name = types.StringValue(rule.Name)
	m.Consumer = types.StringValue(rule.Consumer)
	m.RuleTypeID = types.StringValue(rule.RuleTypeID)
	m.Interval = kibanacustomtypes.NewAlertingDurationValue(rule.Schedule.Interval)

	if rule.NotifyWhen != nil && *rule.NotifyWhen != "" {
		m.NotifyWhen = types.StringValue(*rule.NotifyWhen)
	} else if m.NotifyWhen.IsUnknown() {
		m.NotifyWhen = types.StringNull()
	}

	// Params as JSON string
	normalizedParams, d := normalizeRuleParamsForState(rule.Params, previousParams)
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
		m.Throttle = kibanacustomtypes.NewAlertingDurationValue(*rule.Throttle)
	} else {
		m.Throttle = kibanacustomtypes.NewAlertingDurationNull()
	}

	// Scheduled task ID - update if API returns a value, or resolve unknown to null
	// (preserves existing known value when API doesn't return this field on re-reads)
	if rule.ScheduledTaskID != nil {
		m.ScheduledTaskID = types.StringValue(*rule.ScheduledTaskID)
	} else if m.ScheduledTaskID.IsUnknown() {
		m.ScheduledTaskID = types.StringNull()
	}

	// Alert delay - update if API returns a value, or resolve unknown to null
	// (preserves existing known value when API doesn't return this field on re-reads)
	if rule.AlertDelay != nil {
		m.AlertDelay = types.Int64Value(int64(*rule.AlertDelay))
	} else if m.AlertDelay.IsUnknown() {
		m.AlertDelay = types.Int64Null()
	}

	// Flapping — same partial-response pattern as alert_delay / scheduled_task_id
	if rule.Flapping != nil {
		fm := flappingModel{
			LookBackWindow:        types.Int64Value(rule.Flapping.LookBackWindow),
			StatusChangeThreshold: types.Int64Value(rule.Flapping.StatusChangeThreshold),

			Enabled: types.BoolPointerValue(rule.Flapping.Enabled)}
		flObj, d := types.ObjectValueFrom(ctx, getFlappingAttrTypes(), fm)
		diags.Append(d...)
		m.Flapping = flObj
	} else if m.Flapping.IsUnknown() {
		m.Flapping = types.ObjectNull(getFlappingAttrTypes())
	}

	// Artifacts (investigation guide)
	diags.Append(m.populateArtifactsFromAPI(ctx, rule)...)
	if diags.HasError() {
		return diags
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

// Version thresholds for feature support
var (
	frequencyMinSupportedVersion       = version.Must(version.NewVersion("8.6.0"))
	alertsFilterMinSupportedVersion    = version.Must(version.NewVersion("8.9.0"))
	alertDelayMinSupportedVersion      = version.Must(version.NewVersion("8.13.0"))
	flappingMinSupportedVersion        = version.Must(version.NewVersion("8.16.0"))
	flappingEnabledMinSupportedVersion = version.Must(version.NewVersion("9.3.0"))
	artifactsMinSupportedVersion       = version.Must(version.NewVersion("9.1.0"))
)

func (m alertingRuleModel) GetID() types.String         { return m.ID }
func (m alertingRuleModel) GetResourceID() types.String { return m.RuleID }

var _ entitycore.KibanaResourceModel = alertingRuleModel{}

func (m alertingRuleModel) GetVersionRequirements(ctx context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics) {
	var diags diag.Diagnostics
	var reqs []entitycore.VersionRequirement

	// 8.6.0 when any action has Frequency set
	// 8.9.0 when any action has AlertsFilter set
	if typeutils.IsKnown(m.Actions) && !m.Actions.IsNull() {
		var actions []actionModel
		diags.Append(m.Actions.ElementsAs(ctx, &actions, false)...)
		if diags.HasError() {
			return nil, diags
		}
		var hasFrequency, hasAlertsFilter bool
		for _, action := range actions {
			if !hasFrequency && typeutils.IsKnown(action.Frequency) && !action.Frequency.IsNull() {
				reqs = append(reqs, entitycore.SingleVersionRequirement(
					*frequencyMinSupportedVersion,
					"actions.frequency is only supported for Kibana v8.6 or higher",
				)...)
				hasFrequency = true
			}
			if !hasAlertsFilter && typeutils.IsKnown(action.AlertsFilter) && !action.AlertsFilter.IsNull() {
				reqs = append(reqs, entitycore.SingleVersionRequirement(
					*alertsFilterMinSupportedVersion,
					"actions.alerts_filter is only supported for Kibana v8.9 or higher",
				)...)
				hasAlertsFilter = true
			}
			if hasFrequency && hasAlertsFilter {
				break
			}
		}
	}

	// 8.13.0 when AlertDelay is set
	if typeutils.IsKnown(m.AlertDelay) && !m.AlertDelay.IsNull() {
		reqs = append(reqs, entitycore.SingleVersionRequirement(
			*alertDelayMinSupportedVersion,
			"alert_delay is only supported for Kibana v8.13 or higher",
		)...)
	}

	// 8.16.0 when Flapping is set
	if typeutils.IsKnown(m.Flapping) && !m.Flapping.IsNull() {
		reqs = append(reqs, entitycore.SingleVersionRequirement(
			*flappingMinSupportedVersion,
			"flapping is only supported for Kibana v8.16 or higher",
		)...)

		fm := typeutils.ObjectTypeAs[flappingModel](ctx, m.Flapping, path.Root("flapping"), &diags)
		if diags.HasError() {
			return nil, diags
		}
		// 9.3.0 when Flapping.Enabled is set
		if typeutils.IsKnown(fm.Enabled) && !fm.Enabled.IsNull() {
			reqs = append(reqs, entitycore.SingleVersionRequirement(
				*flappingEnabledMinSupportedVersion,
				"flapping.enabled is only supported for Elastic Stack 9.3 or higher",
			)...)
		}
	}

	// 9.1.0 when Artifacts (investigation guide) is set
	if typeutils.IsKnown(m.Artifacts) && !m.Artifacts.IsNull() {
		reqs = append(reqs, entitycore.SingleVersionRequirement(
			*artifactsMinSupportedVersion,
			"artifacts (investigation guide / dashboards) is only supported for Elastic Stack 9.1 or higher",
		)...)
	}

	return reqs, diags
}

var _ entitycore.WithVersionRequirements = alertingRuleModel{}

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

	// Params from JSON string.
	// Note: params validation is handled exclusively by ValidateConfig during
	// the plan phase. We intentionally do not re-validate here to avoid
	// duplicate error messages.
	if typeutils.IsKnown(m.Params) {
		params := map[string]any{}
		diags.Append(m.Params.Unmarshal(&params)...)
		if diags.HasError() {
			return models.AlertingRule{}, diags
		}

		// Compatibility: older Kibana versions reject `.index-threshold` rule params
		// when `groupBy` is omitted (server-side expects a string, but sees undefined).
		// Defaulting to "all" preserves Kibana's effective behavior while avoiding 400s.
		if rule.RuleTypeID == ruleTypeIndexThreshold {
			if v, ok := params[paramsKeyGroupBy]; !ok || v == nil {
				params[paramsKeyGroupBy] = paramsGroupByAll
			}

			// Compatibility: some Kibana versions reject `.index-threshold` params
			// when `aggType` is omitted (server-side expects a string, but sees undefined).
			// Defaulting to `count` preserves Kibana behavior while avoiding 400s.
			if v, ok := params[paramsKeyAggType]; !ok || v == nil {
				params[paramsKeyAggType] = paramsAggTypeCount
				// aggField is only valid for non-count aggregations.
				delete(params, "aggField")
			}
		}

		rule.Params = params
	}

	// Enabled
	if typeutils.IsKnown(m.Enabled) {
		rule.Enabled = m.Enabled.ValueBoolPointer()
	}

	// NotifyWhen
	if typeutils.IsKnown(m.NotifyWhen) && m.NotifyWhen.ValueString() != "" {
		rule.NotifyWhen = m.NotifyWhen.ValueStringPointer()
	}

	// Throttle
	if typeutils.IsKnown(m.Throttle) && m.Throttle.ValueString() != "" {
		rule.Throttle = m.Throttle.ValueStringPointer()
	}

	// Tags
	if typeutils.IsKnown(m.Tags) && !m.Tags.IsNull() {
		var tags []string
		diags.Append(m.Tags.ElementsAs(ctx, &tags, false)...)
		rule.Tags = tags
	}

	// Alert delay
	if typeutils.IsKnown(m.AlertDelay) && !m.AlertDelay.IsNull() {
		alertDelay := float32(m.AlertDelay.ValueInt64())
		rule.AlertDelay = &alertDelay
	}

	// Flapping
	if typeutils.IsKnown(m.Flapping) && !m.Flapping.IsNull() {
		fm := typeutils.ObjectTypeAs[flappingModel](ctx, m.Flapping, path.Root("flapping"), &diags)
		if diags.HasError() {
			return models.AlertingRule{}, diags
		}
		rule.Flapping = &models.AlertingRuleFlapping{
			LookBackWindow:        fm.LookBackWindow.ValueInt64(),
			StatusChangeThreshold: fm.StatusChangeThreshold.ValueInt64(),
		}
		if typeutils.IsKnown(fm.Enabled) && !fm.Enabled.IsNull() {
			rule.Flapping.Enabled = fm.Enabled.ValueBoolPointer()
		}
	}

	// Artifacts (investigation guide)
	if typeutils.IsKnown(m.Artifacts) && !m.Artifacts.IsNull() {
		artifacts, d := m.artifactsToAPI(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return models.AlertingRule{}, diags
		}
		rule.Artifacts = artifacts
	}

	// Actions
	if typeutils.IsKnown(m.Actions) && !m.Actions.IsNull() {
		actions, d := convertActionsToAPI(ctx, m.Actions)
		diags.Append(d...)
		rule.Actions = actions
	}

	return rule, diags
}
