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

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/alertingactions"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/kibanacustomtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
	Summary    types.Bool                         `tfsdk:"summary"`
	NotifyWhen types.String                       `tfsdk:"notify_when"`
	Throttle   kibanacustomtypes.AlertingDuration `tfsdk:"throttle"`
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
				Throttle:   kibanacustomtypes.NewAlertingDurationPointerValue(apiAction.Frequency.Throttle),
			}
			freqObj, d := types.ObjectValueFrom(ctx, getFrequencyAttrTypes(), freq)
			diags.Append(d...)
			action.Frequency = freqObj
		} else {
			action.Frequency = types.ObjectNull(getFrequencyAttrTypes())
		}

		// Alerts filter - convert to single object
		if apiAction.AlertsFilter != nil {
			filter := alertsFilterModel{

				Kql: types.StringPointerValue(apiAction.AlertsFilter.Kql)}

			if apiAction.AlertsFilter.Timeframe != nil {
				tf := apiAction.AlertsFilter.Timeframe
				days, err := alertingactions.DaysFromAPI(tf.Days)
				if err != nil {
					diags.AddError("Failed to convert alerts_filter timeframe days", err.Error())
					continue
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

	if !typeutils.IsKnown(actionsList) {
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
		if params := typeutils.NormalizedTypeToMap[any](action.Params, path.Root("actions").AtListIndex(i).AtName("params"), &diags); params != nil {
			apiAction.Params = params
		}

		// Frequency - extract from object
		if typeutils.IsKnown(action.Frequency) && !action.Frequency.IsNull() {
			freq := typeutils.ObjectTypeAs[frequencyModel](ctx, action.Frequency, path.Root("actions").AtListIndex(i).AtName("frequency"), &diags)
			// Only create Frequency if both required fields are present
			if typeutils.IsKnown(freq.Summary) && typeutils.IsKnown(freq.NotifyWhen) {
				apiAction.Frequency = &models.ActionFrequency{
					Summary:    freq.Summary.ValueBool(),
					NotifyWhen: freq.NotifyWhen.ValueString(),
				}
				if typeutils.IsKnown(freq.Throttle) && freq.Throttle.ValueString() != "" {
					apiAction.Frequency.Throttle = freq.Throttle.ValueStringPointer()
				}
			}
		}

		// Alerts filter - extract from object
		if typeutils.IsKnown(action.AlertsFilter) && !action.AlertsFilter.IsNull() {
			alertsFilterPath := path.Root("actions").AtListIndex(i).AtName("alerts_filter")
			filter := typeutils.ObjectTypeAs[alertsFilterModel](ctx, action.AlertsFilter, alertsFilterPath, &diags)
			apiAction.AlertsFilter = &models.ActionAlertsFilter{}

			if typeutils.IsKnown(filter.Kql) {
				kql := filter.Kql.ValueString()
				apiAction.AlertsFilter.Kql = &kql
			}

			if typeutils.IsKnown(filter.Timeframe) && !filter.Timeframe.IsNull() {
				tf := typeutils.ObjectTypeAs[timeframeModel](ctx, filter.Timeframe, alertsFilterPath.AtName("timeframe"), &diags)
				var days []int64
				diags.Append(tf.Days.ElementsAs(ctx, &days, false)...)

				apiAction.AlertsFilter.Timeframe = &models.AlertsFilterTimeframe{
					Days:       alertingactions.Int32FromInt64(days),
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
