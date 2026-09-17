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

package securitydetectionrule

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Helper function to process actions configuration for all rule types
func (d Data) actionsToAPI(ctx context.Context) ([]kbapi.SecurityDetectionsAPIRuleAction, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !typeutils.IsKnown(d.Actions) || len(d.Actions.Elements()) == 0 {
		return nil, diags
	}

	apiActions := typeutils.ListTypeToSlice(ctx, d.Actions, path.Root("actions"), &diags,
		func(action ActionModel, meta typeutils.ListMeta) kbapi.SecurityDetectionsAPIRuleAction {
			apiAction := kbapi.SecurityDetectionsAPIRuleAction{
				ActionTypeId: action.ActionTypeID.ValueString(),
				Id:           action.ID.ValueString(),
			}

			// Convert params: unmarshal the JSON string into a free-form map.
			if typeutils.IsKnown(action.Params) && !action.Params.IsNull() {
				var paramsMap map[string]any
				if err := json.Unmarshal([]byte(action.Params.ValueString()), &paramsMap); err != nil {
					meta.Diags.AddError("Error unmarshaling action params", err.Error())
				} else {
					apiAction.Params = kbapi.SecurityDetectionsAPIRuleActionParams(paramsMap)
				}
			}

			// Set optional fields
			if typeutils.IsKnown(action.Group) {
				group := action.Group.ValueString()
				apiAction.Group = &group
			}

			if typeutils.IsKnown(action.UUID) {
				uuidStr := action.UUID.ValueString()
				apiAction.Uuid = &uuidStr
			}

			apiAction.AlertsFilter = expandActionAlertsFilter(ctx, action.AlertsFilter, meta.Diags)

			// Handle frequency using ObjectTypeToStruct
			if typeutils.IsKnown(action.Frequency) {
				frequency := typeutils.ObjectTypeToStruct(ctx, action.Frequency, meta.Path.AtName("frequency"), meta.Diags,
					func(frequencyModel ActionFrequencyModel, freqMeta typeutils.ObjectMeta) kbapi.SecurityDetectionsAPIRuleActionFrequency {
						apiFreq := kbapi.SecurityDetectionsAPIRuleActionFrequency{
							NotifyWhen: kbapi.SecurityDetectionsAPIRuleActionNotifyWhen(frequencyModel.NotifyWhen.ValueString()),
							Summary:    frequencyModel.Summary.ValueBool(),
						}

						// Handle throttle - can be string or specific values
						if typeutils.IsKnown(frequencyModel.Throttle) {
							throttleStr := frequencyModel.Throttle.ValueString()
							var throttle kbapi.SecurityDetectionsAPIRuleActionThrottle
							if throttleStr == "no_actions" || throttleStr == "rule" {
								// Use the enum value
								var throttle0 kbapi.SecurityDetectionsAPIRuleActionThrottle0
								if throttleStr == "no_actions" {
									throttle0 = kbapi.SecurityDetectionsAPIRuleActionThrottle0NoActions
								} else {
									throttle0 = kbapi.SecurityDetectionsAPIRuleActionThrottle0Rule
								}
								err := throttle.FromSecurityDetectionsAPIRuleActionThrottle0(throttle0)
								if err != nil {
									freqMeta.Diags.AddError("Error setting throttle enum", err.Error())
								}
							} else {
								// Use the time interval string
								err := throttle.FromSecurityDetectionsAPIRuleActionThrottle1(throttleStr)
								if err != nil {
									freqMeta.Diags.AddError("Error setting throttle interval", err.Error())
								}
							}
							apiFreq.Throttle = throttle
						}

						return apiFreq
					})

				if frequency != nil {
					apiAction.Frequency = frequency
				}
			}

			return apiAction
		})

	// Filter out empty actions (where ActionTypeID or Id was null)
	validActions := make([]kbapi.SecurityDetectionsAPIRuleAction, 0)
	for _, action := range apiActions {
		if action.ActionTypeId != "" && action.Id != "" {
			validActions = append(validActions, action)
		}
	}

	return validActions, diags
}

// convertActionsToModel converts kbapi.SecurityDetectionsAPIRuleAction slice to Terraform model
func convertActionsToModel(ctx context.Context, apiActions []kbapi.SecurityDetectionsAPIRuleAction) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(apiActions) == 0 {
		return types.ListNull(getActionElementType()), diags
	}

	actions := make([]ActionModel, 0)

	for _, apiAction := range apiActions {
		action := ActionModel{
			ActionTypeID: types.StringValue(apiAction.ActionTypeId),
			ID:           types.StringValue(apiAction.Id),
		}

		// Convert params: serialize the whole object as normalized JSON.
		if apiAction.Params != nil {
			jsonBytes, err := json.Marshal(map[string]any(apiAction.Params))
			if err != nil {
				diags.AddError("Error marshaling action params", err.Error())
			} else {
				action.Params = jsontypes.NewNormalizedValue(string(jsonBytes))
			}
		} else {
			action.Params = jsontypes.NewNormalizedNull()
		}

		// Set optional fields
		action.Group = types.StringPointerValue(apiAction.Group)

		action.UUID = typeutils.StringishPointerValue(apiAction.Uuid)

		action.AlertsFilter = flattenActionAlertsFilter(ctx, apiAction.AlertsFilter, &diags)

		// Convert frequency
		if apiAction.Frequency != nil {
			var throttleStr string
			if throttle0, err := apiAction.Frequency.Throttle.AsSecurityDetectionsAPIRuleActionThrottle0(); err == nil {
				throttleStr = string(throttle0)
			} else if throttle1, err := apiAction.Frequency.Throttle.AsSecurityDetectionsAPIRuleActionThrottle1(); err == nil {
				throttleStr = throttle1
			}

			frequencyModel := ActionFrequencyModel{
				NotifyWhen: typeutils.StringishValue(apiAction.Frequency.NotifyWhen),
				Summary:    types.BoolValue(apiAction.Frequency.Summary),
				Throttle:   types.StringValue(throttleStr),
			}

			frequencyObj, frequencyDiags := types.ObjectValueFrom(ctx, getActionFrequencyType(), frequencyModel)
			diags.Append(frequencyDiags...)
			action.Frequency = frequencyObj
		} else {
			action.Frequency = types.ObjectNull(getActionFrequencyType())
		}

		actions = append(actions, action)
	}

	listValue, listDiags := types.ListValueFrom(ctx, getActionElementType(), actions)
	diags.Append(listDiags...)
	return listValue, diags
}

func (d *Data) updateActionsFromAPI(ctx context.Context, actions []kbapi.SecurityDetectionsAPIRuleAction) diag.Diagnostics {
	actionsListValue, diags := convertActionsToModel(ctx, actions)
	if !diags.HasError() {
		d.Actions = actionsListValue
	}
	return diags
}
