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
	"fmt"
	"regexp"
	"strconv"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

const kqlQueryLanguageKuery = "kuery"

// getKQLQueryLanguage maps language string to kbapi.SecurityDetectionsAPIKqlQueryLanguage
func (d Data) getKQLQueryLanguage() *kbapi.SecurityDetectionsAPIKqlQueryLanguage {
	if !typeutils.IsKnown(d.Language) {
		return nil
	}
	var language kbapi.SecurityDetectionsAPIKqlQueryLanguage
	switch d.Language.ValueString() {
	case kqlQueryLanguageKuery:
		language = kqlQueryLanguageKuery
	case "lucene":
		language = "lucene"
	default:
		language = kqlQueryLanguageKuery
	}
	return &language
}

// buildOsqueryResponseAction creates an Osquery response action from the terraform model
func (d Data) buildOsqueryResponseAction(ctx context.Context, params ResponseActionParamsModel) (kbapi.SecurityDetectionsAPIResponseAction, diag.Diagnostics) {
	var diags diag.Diagnostics

	osqueryAction := kbapi.SecurityDetectionsAPIOsqueryResponseAction{
		ActionTypeId: kbapi.SecurityDetectionsAPIOsqueryResponseActionActionTypeId(".osquery"),
		Params:       kbapi.SecurityDetectionsAPIOsqueryParams{},
	}

	// Set osquery-specific params
	if typeutils.IsKnown(params.Query) {
		osqueryAction.Params.Query = params.Query.ValueStringPointer()
	}
	if typeutils.IsKnown(params.PackID) {
		osqueryAction.Params.PackId = params.PackID.ValueStringPointer()
	}
	if typeutils.IsKnown(params.SavedQueryID) {
		osqueryAction.Params.SavedQueryId = params.SavedQueryID.ValueStringPointer()
	}
	if typeutils.IsKnown(params.Timeout) {
		timeout := float32(params.Timeout.ValueInt64())
		osqueryAction.Params.Timeout = &timeout
	}
	if typeutils.IsKnown(params.EcsMapping) {

		// Convert map to ECS mapping structure
		ecsMappingElems := make(map[string]basetypes.StringValue)
		elemDiags := params.EcsMapping.ElementsAs(ctx, &ecsMappingElems, false)
		if !elemDiags.HasError() {
			ecsMapping := make(kbapi.SecurityDetectionsAPIEcsMapping)
			for key, value := range ecsMappingElems {
				if stringVal := value; typeutils.IsKnown(value) {
					ecsMapping[key] = struct {
						Field *string                                      `json:"field,omitempty"`
						Value *kbapi.SecurityDetectionsAPIEcsMapping_Value `json:"value,omitempty"`
					}{
						Field: stringVal.ValueStringPointer(),
					}
				}
			}
			osqueryAction.Params.EcsMapping = &ecsMapping
		} else {
			diags.Append(elemDiags...)
		}
	}
	if typeutils.IsKnown(params.Queries) {
		queries := make([]OsqueryQueryModel, len(params.Queries.Elements()))
		queriesDiags := params.Queries.ElementsAs(ctx, &queries, false)
		if !queriesDiags.HasError() {
			apiQueries := make([]kbapi.SecurityDetectionsAPIOsqueryQuery, 0)
			for _, query := range queries {
				apiQuery := kbapi.SecurityDetectionsAPIOsqueryQuery{
					Id:    query.ID.ValueString(),
					Query: query.Query.ValueString(),
				}
				if typeutils.IsKnown(query.Platform) {
					apiQuery.Platform = query.Platform.ValueStringPointer()
				}
				if typeutils.IsKnown(query.Version) {
					apiQuery.Version = query.Version.ValueStringPointer()
				}
				if typeutils.IsKnown(query.Removed) {
					apiQuery.Removed = query.Removed.ValueBoolPointer()
				}
				if typeutils.IsKnown(query.Snapshot) {
					apiQuery.Snapshot = query.Snapshot.ValueBoolPointer()
				}
				if typeutils.IsKnown(query.EcsMapping) {
					// Convert map to ECS mapping structure for queries
					queryEcsMappingElems := make(map[string]basetypes.StringValue)
					queryElemDiags := query.EcsMapping.ElementsAs(ctx, &queryEcsMappingElems, false)
					if !queryElemDiags.HasError() {
						queryEcsMapping := make(kbapi.SecurityDetectionsAPIEcsMapping)
						for key, value := range queryEcsMappingElems {
							if stringVal := value; typeutils.IsKnown(value) {
								queryEcsMapping[key] = struct {
									Field *string                                      `json:"field,omitempty"`
									Value *kbapi.SecurityDetectionsAPIEcsMapping_Value `json:"value,omitempty"`
								}{
									Field: stringVal.ValueStringPointer(),
								}
							}
						}
						apiQuery.EcsMapping = &queryEcsMapping
					}
				}
				apiQueries = append(apiQueries, apiQuery)
			}
			osqueryAction.Params.Queries = &apiQueries
		} else {
			diags = append(diags, queriesDiags...)
		}
	}

	var apiResponseAction kbapi.SecurityDetectionsAPIResponseAction
	err := apiResponseAction.FromSecurityDetectionsAPIOsqueryResponseAction(osqueryAction)
	if err != nil {
		diags.AddError("Error converting osquery response action", err.Error())
	}

	return apiResponseAction, diags
}

// buildEndpointResponseAction creates an Endpoint response action from the terraform model
func (d Data) buildEndpointResponseAction(ctx context.Context, params ResponseActionParamsModel) (kbapi.SecurityDetectionsAPIResponseAction, diag.Diagnostics) {
	var diags diag.Diagnostics

	endpointAction := kbapi.SecurityDetectionsAPIEndpointResponseAction{
		ActionTypeId: kbapi.SecurityDetectionsAPIEndpointResponseActionActionTypeId(".endpoint"),
	}

	// Determine the type of endpoint action based on the command
	if typeutils.IsKnown(params.Command) {
		command := params.Command.ValueString()
		switch command {
		case "isolate":
			// Use DefaultParams for isolate command
			defaultParams := kbapi.SecurityDetectionsAPIDefaultParams{
				Command: kbapi.SecurityDetectionsAPIDefaultParamsCommand("isolate"),
			}
			if typeutils.IsKnown(params.Comment) {
				defaultParams.Comment = params.Comment.ValueStringPointer()
			}
			err := endpointAction.Params.FromSecurityDetectionsAPIDefaultParams(defaultParams)
			if err != nil {
				diags.AddError("Error setting endpoint default params", err.Error())
				return kbapi.SecurityDetectionsAPIResponseAction{}, diags
			}

		case "kill-process", "suspend-process":
			// Use ProcessesParams for process commands
			processesParams := kbapi.SecurityDetectionsAPIProcessesParams{
				Command: kbapi.SecurityDetectionsAPIProcessesParamsCommand(command),
			}
			if typeutils.IsKnown(params.Comment) {
				processesParams.Comment = params.Comment.ValueStringPointer()
			}

			// Set config if provided
			if typeutils.IsKnown(params.Config) {
				config := typeutils.ObjectTypeToStruct(ctx, params.Config, path.Root("response_actions").AtName("params").AtName("config"), &diags,
					func(item EndpointProcessConfigModel, _ typeutils.ObjectMeta) EndpointProcessConfigModel {
						return item
					})

				processesParams.Config = struct {
					Field     string `json:"field"`
					Overwrite *bool  `json:"overwrite,omitempty"`
				}{
					Field: config.Field.ValueString(),
				}
				if typeutils.IsKnown(config.Overwrite) {
					processesParams.Config.Overwrite = config.Overwrite.ValueBoolPointer()
				}
			}

			err := endpointAction.Params.FromSecurityDetectionsAPIProcessesParams(processesParams)
			if err != nil {
				diags.AddError("Error setting endpoint processes params", err.Error())
				return kbapi.SecurityDetectionsAPIResponseAction{}, diags
			}
		default:
			diags.AddError(
				"Unsupported params type",
				fmt.Sprintf("Params type '%s' is not supported", params.Command.ValueString()),
			)
		}
	}

	var apiResponseAction kbapi.SecurityDetectionsAPIResponseAction
	err := apiResponseAction.FromSecurityDetectionsAPIEndpointResponseAction(endpointAction)
	if err != nil {
		diags.AddError("Error converting endpoint response action", err.Error())
	}

	return apiResponseAction, diags
}

// Helper function to process threshold configuration for threshold rules
func (d Data) thresholdToAPI(ctx context.Context, diags *diag.Diagnostics) *kbapi.SecurityDetectionsAPIThreshold {
	if !typeutils.IsKnown(d.Threshold) {
		return nil
	}

	threshold := typeutils.ObjectTypeToStruct(ctx, d.Threshold, path.Root("threshold"), diags,
		func(item ThresholdModel, meta typeutils.ObjectMeta) kbapi.SecurityDetectionsAPIThreshold {
			threshold := kbapi.SecurityDetectionsAPIThreshold{
				Value: kbapi.SecurityDetectionsAPIThresholdValue(item.Value.ValueInt64()),
			}

			// Handle threshold field(s)
			if typeutils.IsKnown(item.Field) {
				fieldList := typeutils.ListTypeToSliceString(ctx, item.Field, meta.Path.AtName("field"), meta.Diags)
				if len(fieldList) > 0 {
					var thresholdField kbapi.SecurityDetectionsAPIThresholdField
					if len(fieldList) == 1 {
						err := thresholdField.FromSecurityDetectionsAPIThresholdField0(fieldList[0])
						if err != nil {
							meta.Diags.AddError("Error setting threshold field", err.Error())
						} else {
							threshold.Field = thresholdField
						}
					} else {
						err := thresholdField.FromSecurityDetectionsAPIThresholdField1(fieldList)
						if err != nil {
							meta.Diags.AddError("Error setting threshold fields", err.Error())
						} else {
							threshold.Field = thresholdField
						}
					}
				}
			}

			// Handle cardinality (optional)
			if typeutils.IsKnown(item.Cardinality) {
				cardinalityList := typeutils.ListTypeToSlice(ctx, item.Cardinality, meta.Path.AtName("cardinality"), meta.Diags,
					func(item CardinalityModel, _ typeutils.ListMeta) struct {
						Field string `json:"field"`
						Value int    `json:"value"`
					} {
						return struct {
							Field string `json:"field"`
							Value int    `json:"value"`
						}{
							Field: item.Field.ValueString(),
							Value: int(item.Value.ValueInt64()),
						}
					})
				if len(cardinalityList) > 0 {
					threshold.Cardinality = &cardinalityList
				}
			}

			return threshold
		})

	return threshold
}

// Helper function to convert alert suppression from TF data to API type
func (d Data) alertSuppressionToAPI(ctx context.Context, diags *diag.Diagnostics) *kbapi.SecurityDetectionsAPIAlertSuppression {
	if !typeutils.IsKnown(d.AlertSuppression) {
		return nil
	}

	var model AlertSuppressionModel
	objDiags := d.AlertSuppression.As(ctx, &model, basetypes.ObjectAsOptions{})
	diags.Append(objDiags...)
	if diags.HasError() {
		return nil
	}

	suppression := &kbapi.SecurityDetectionsAPIAlertSuppression{}

	// Handle group_by (required)
	if typeutils.IsKnown(model.GroupBy) {
		groupByList := typeutils.ListTypeToSliceString(ctx, model.GroupBy, path.Root("alert_suppression").AtName("group_by"), diags)
		if len(groupByList) > 0 {
			suppression.GroupBy = groupByList
		}
	}

	// Handle duration (optional)
	if typeutils.IsKnown(model.Duration) {
		duration, durationDiags := parseDurationToAPI(model.Duration)
		diags.Append(durationDiags...)
		if !durationDiags.HasError() {
			suppression.Duration = &duration
		}
	}

	// Handle missing_fields_strategy (optional)
	if typeutils.IsKnown(model.MissingFieldsStrategy) {
		strategy := kbapi.SecurityDetectionsAPIAlertSuppressionMissingFieldsStrategy(model.MissingFieldsStrategy.ValueString())
		suppression.MissingFieldsStrategy = &strategy
	}

	return suppression
}

// Helper function to convert alert suppression from TF data to threshold-specific API type
func (d Data) alertSuppressionToThresholdAPI(ctx context.Context, diags *diag.Diagnostics) *kbapi.SecurityDetectionsAPIThresholdAlertSuppression {
	if !typeutils.IsKnown(d.AlertSuppression) {
		return nil
	}

	var model AlertSuppressionModel
	objDiags := d.AlertSuppression.As(ctx, &model, basetypes.ObjectAsOptions{})
	diags.Append(objDiags...)
	if diags.HasError() {
		return nil
	}

	suppression := &kbapi.SecurityDetectionsAPIThresholdAlertSuppression{}

	// Handle duration (required for threshold alert suppression)
	if !typeutils.IsKnown(model.Duration) {
		diags.AddError(
			"Duration required for threshold alert suppression",
			"Threshold alert suppression requires a duration to be specified",
		)
		return nil
	}

	duration, durationDiags := parseDurationToAPI(model.Duration)
	diags.Append(durationDiags...)
	if !durationDiags.HasError() {
		suppression.Duration = duration
	}

	// Note: Threshold alert suppression only supports duration field.
	// GroupBy and MissingFieldsStrategy are not supported for threshold rules.

	return suppression
}

// Helper function to process threat mapping configuration for threat match rules
func (d Data) threatMappingToAPI(ctx context.Context) (kbapi.SecurityDetectionsAPIThreatMapping, diag.Diagnostics) {
	var diags diag.Diagnostics

	threatMapping := make([]TfDataItem, len(d.ThreatMapping.Elements()))

	threatMappingDiags := d.ThreatMapping.ElementsAs(ctx, &threatMapping, false)
	if threatMappingDiags.HasError() {
		diags.Append(threatMappingDiags...)
		return nil, diags
	}

	apiThreatMapping := make(kbapi.SecurityDetectionsAPIThreatMapping, 0)
	for _, mapping := range threatMapping {
		if !typeutils.IsKnown(mapping.Entries) {
			continue
		}

		entries := make([]TfDataItemEntry, len(mapping.Entries.Elements()))
		entryDiag := mapping.Entries.ElementsAs(ctx, &entries, false)
		diags = append(diags, entryDiag...)

		apiThreatMappingEntries := make([]kbapi.SecurityDetectionsAPIThreatMappingEntry, 0)
		for _, entry := range entries {

			apiMapping := kbapi.SecurityDetectionsAPIThreatMappingEntry{
				Field: entry.Field.ValueString(),
				Type:  kbapi.SecurityDetectionsAPIThreatMappingEntryType(entry.Type.ValueString()),
				Value: entry.Value.ValueString(),
			}
			apiThreatMappingEntries = append(apiThreatMappingEntries, apiMapping)

		}

		apiThreatMapping = append(apiThreatMapping, struct {
			Entries []kbapi.SecurityDetectionsAPIThreatMappingEntry `json:"entries"`
		}{Entries: apiThreatMappingEntries})
	}

	return apiThreatMapping, diags
}

// Helper function to convert MITRE ATT&CK threat data from Terraform to API format
func (d Data) threatToAPI(ctx context.Context) (kbapi.SecurityDetectionsAPIThreatArray, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !typeutils.IsKnown(d.Threat) || len(d.Threat.Elements()) == 0 {
		return nil, diags
	}

	threats := make([]ThreatModel, len(d.Threat.Elements()))
	threatDiags := d.Threat.ElementsAs(ctx, &threats, false)
	diags.Append(threatDiags...)
	if threatDiags.HasError() {
		return nil, diags
	}

	apiThreats := make(kbapi.SecurityDetectionsAPIThreatArray, 0)
	for _, threat := range threats {
		apiThreat := kbapi.SecurityDetectionsAPIThreat{
			Framework: threat.Framework.ValueString(),
		}

		// Convert tactic
		var tacticModel ThreatTacticModel
		tacticDiags := threat.Tactic.As(ctx, &tacticModel, basetypes.ObjectAsOptions{})
		diags.Append(tacticDiags...)
		if tacticDiags.HasError() {
			continue
		}

		apiThreat.Tactic = kbapi.SecurityDetectionsAPIThreatTactic{
			Id:        tacticModel.ID.ValueString(),
			Name:      tacticModel.Name.ValueString(),
			Reference: tacticModel.Reference.ValueString(),
		}

		// Convert techniques (optional)
		if typeutils.IsKnown(threat.Technique) && len(threat.Technique.Elements()) > 0 {
			techniques := make([]ThreatTechniqueModel, len(threat.Technique.Elements()))
			techniqueDiags := threat.Technique.ElementsAs(ctx, &techniques, false)
			diags.Append(techniqueDiags...)
			if techniqueDiags.HasError() {
				continue
			}

			apiTechniques := make([]kbapi.SecurityDetectionsAPIThreatTechnique, 0)
			for _, technique := range techniques {
				apiTechnique := kbapi.SecurityDetectionsAPIThreatTechnique{
					Id:        technique.ID.ValueString(),
					Name:      technique.Name.ValueString(),
					Reference: technique.Reference.ValueString(),
				}

				// Convert subtechniques (optional)
				if typeutils.IsKnown(technique.Subtechnique) && len(technique.Subtechnique.Elements()) > 0 {
					subtechniques := make([]ThreatSubtechniqueModel, len(technique.Subtechnique.Elements()))
					subtechniqueDiags := technique.Subtechnique.ElementsAs(ctx, &subtechniques, false)
					diags.Append(subtechniqueDiags...)
					if subtechniqueDiags.HasError() {
						continue
					}

					apiSubtechniques := make([]kbapi.SecurityDetectionsAPIThreatSubtechnique, 0)
					for _, subtechnique := range subtechniques {
						apiSubtechnique := kbapi.SecurityDetectionsAPIThreatSubtechnique{
							Id:        subtechnique.ID.ValueString(),
							Name:      subtechnique.Name.ValueString(),
							Reference: subtechnique.Reference.ValueString(),
						}
						apiSubtechniques = append(apiSubtechniques, apiSubtechnique)
					}
					apiTechnique.Subtechnique = &apiSubtechniques
				}

				apiTechniques = append(apiTechniques, apiTechnique)
			}
			apiThreat.Technique = &apiTechniques
		}

		apiThreats = append(apiThreats, apiThreat)
	}

	return apiThreats, diags
}

// Helper function to process response actions configuration for all rule types
func (d Data) responseActionsToAPI(ctx context.Context, client clients.MinVersionEnforceable) ([]kbapi.SecurityDetectionsAPIResponseAction, diag.Diagnostics) {
	var diags diag.Diagnostics

	if client == nil {
		diags.AddError(
			"Client is not initialized",
			"Response actions require a valid API client",
		)
		return nil, diags
	}

	if !typeutils.IsKnown(d.ResponseActions) || len(d.ResponseActions.Elements()) == 0 {
		return nil, diags
	}

	// Check version support for response actions
	if supported, versionDiags := client.EnforceMinVersion(ctx, MinVersionResponseActions); versionDiags.HasError() {
		diags.Append(diagutil.FrameworkDiagsFromSDK(versionDiags)...)
		return nil, diags
	} else if !supported {
		// Version is not supported, return nil without error
		diags.AddError("Response actions are unsupported",
			fmt.Sprintf("Response actions require server version %s or higher", MinVersionResponseActions.String()))
		return nil, diags
	}

	apiResponseActions := typeutils.ListTypeToSlice(ctx, d.ResponseActions, path.Root("response_actions"), &diags,
		func(responseAction ResponseActionModel, meta typeutils.ListMeta) kbapi.SecurityDetectionsAPIResponseAction {

			actionTypeID := responseAction.ActionTypeID.ValueString()

			params := typeutils.ObjectTypeToStruct(ctx, responseAction.Params, meta.Path.AtName("params"), &diags,
				func(item ResponseActionParamsModel, _ typeutils.ObjectMeta) ResponseActionParamsModel {
					return item
				})

			if params == nil {
				return kbapi.SecurityDetectionsAPIResponseAction{}
			}

			switch actionTypeID {
			case ".osquery":
				apiAction, actionDiags := d.buildOsqueryResponseAction(ctx, *params)
				diags.Append(actionDiags...)
				return apiAction

			case ".endpoint":
				apiAction, actionDiags := d.buildEndpointResponseAction(ctx, *params)
				diags.Append(actionDiags...)
				return apiAction

			default:
				diags.AddError(
					"Unsupported action_type_id in response actions",
					fmt.Sprintf("action_type_id '%s' is not supported", actionTypeID),
				)
				return kbapi.SecurityDetectionsAPIResponseAction{}
			}
		})

	return apiResponseActions, diags
}

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

			// Convert params map
			if typeutils.IsKnown(action.Params) {
				paramsStringMap := make(map[string]string)
				paramsDiags := action.Params.ElementsAs(ctx, &paramsStringMap, false)
				if !paramsDiags.HasError() {
					paramsMap := make(map[string]any)
					for k, v := range paramsStringMap {
						paramsMap[k] = v
					}
					apiAction.Params = kbapi.SecurityDetectionsAPIRuleActionParams(paramsMap)
				}
				meta.Diags.Append(paramsDiags...)
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

			if typeutils.IsKnown(action.AlertsFilter) {
				alertsFilterStringMap := make(map[string]string)
				alertsFilterDiags := action.AlertsFilter.ElementsAs(ctx, &alertsFilterStringMap, false)
				if !alertsFilterDiags.HasError() {
					alertsFilterMap := make(map[string]any)
					for k, v := range alertsFilterStringMap {
						alertsFilterMap[k] = v
					}
					apiAlertsFilter := kbapi.SecurityDetectionsAPIRuleActionAlertsFilter(alertsFilterMap)
					apiAction.AlertsFilter = &apiAlertsFilter
				}
				meta.Diags.Append(alertsFilterDiags...)
			}

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

// Helper function to process exceptions list configuration for all rule types
func (d Data) exceptionsListToAPI(ctx context.Context) ([]kbapi.SecurityDetectionsAPIRuleExceptionList, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !typeutils.IsKnown(d.ExceptionsList) || len(d.ExceptionsList.Elements()) == 0 {
		return nil, diags
	}

	apiExceptionsList := typeutils.ListTypeToSlice(ctx, d.ExceptionsList, path.Root("exceptions_list"), &diags,
		func(exception ExceptionsListModel, _ typeutils.ListMeta) kbapi.SecurityDetectionsAPIRuleExceptionList {

			apiException := kbapi.SecurityDetectionsAPIRuleExceptionList{
				Id:            exception.ID.ValueString(),
				ListId:        exception.ListID.ValueString(),
				NamespaceType: kbapi.SecurityDetectionsAPIRuleExceptionListNamespaceType(exception.NamespaceType.ValueString()),
				Type:          kbapi.SecurityDetectionsAPIExceptionListType(exception.Type.ValueString()),
			}

			return apiException
		})

	// Filter out empty exceptions (where required fields were null)
	validExceptions := make([]kbapi.SecurityDetectionsAPIRuleExceptionList, 0)
	for _, exception := range apiExceptionsList {
		if exception.Id != "" && exception.ListId != "" {
			validExceptions = append(validExceptions, exception)
		}
	}

	return validExceptions, diags
}

// Helper function to process risk score mapping configuration for all rule types
func (d Data) riskScoreMappingToAPI(ctx context.Context) (kbapi.SecurityDetectionsAPIRiskScoreMapping, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !typeutils.IsKnown(d.RiskScoreMapping) || len(d.RiskScoreMapping.Elements()) == 0 {
		return nil, diags
	}

	apiRiskScoreMapping := typeutils.ListTypeToSlice(ctx, d.RiskScoreMapping, path.Root("risk_score_mapping"), &diags,
		func(mapping RiskScoreMappingModel, _ typeutils.ListMeta) struct {
			Field     string                                              `json:"field"`
			Operator  kbapi.SecurityDetectionsAPIRiskScoreMappingOperator `json:"operator"`
			RiskScore *kbapi.SecurityDetectionsAPIRiskScore               `json:"risk_score,omitempty"`
			Value     string                                              `json:"value"`
		} {
			apiMapping := struct {
				Field     string                                              `json:"field"`
				Operator  kbapi.SecurityDetectionsAPIRiskScoreMappingOperator `json:"operator"`
				RiskScore *kbapi.SecurityDetectionsAPIRiskScore               `json:"risk_score,omitempty"`
				Value     string                                              `json:"value"`
			}{
				Field:    mapping.Field.ValueString(),
				Operator: kbapi.SecurityDetectionsAPIRiskScoreMappingOperator(mapping.Operator.ValueString()),
				Value:    mapping.Value.ValueString(),
			}

			// Set optional risk score if provided
			if typeutils.IsKnown(mapping.RiskScore) {
				riskScore := kbapi.SecurityDetectionsAPIRiskScore(mapping.RiskScore.ValueInt64())
				apiMapping.RiskScore = &riskScore
			}

			return apiMapping
		})

	// Return the mappings (any empty mappings were filtered out during creation)
	return apiRiskScoreMapping, diags
}

// Helper function to process investigation fields configuration for all rule types
func (d Data) investigationFieldsToAPI(ctx context.Context) (*kbapi.SecurityDetectionsAPIInvestigationFields, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !typeutils.IsKnown(d.InvestigationFields) || len(d.InvestigationFields.Elements()) == 0 {
		return nil, diags
	}

	fieldNames := make([]string, len(d.InvestigationFields.Elements()))
	fieldDiag := d.InvestigationFields.ElementsAs(ctx, &fieldNames, false)
	if fieldDiag.HasError() {
		diags.Append(fieldDiag...)
		return nil, diags
	}

	// Convert to API type
	apiFieldNames := make([]kbapi.SecurityDetectionsAPINonEmptyString, len(fieldNames))
	copy(apiFieldNames, fieldNames)

	return &kbapi.SecurityDetectionsAPIInvestigationFields{
		FieldNames: apiFieldNames,
	}, diags
}

// Helper function to process related integrations configuration for all rule types
func (d Data) relatedIntegrationsToAPI(ctx context.Context) (*kbapi.SecurityDetectionsAPIRelatedIntegrationArray, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !typeutils.IsKnown(d.RelatedIntegrations) || len(d.RelatedIntegrations.Elements()) == 0 {
		return nil, diags
	}

	apiRelatedIntegrations := typeutils.ListTypeToSlice(ctx, d.RelatedIntegrations, path.Root("related_integrations"), &diags,
		func(integration RelatedIntegrationModel, _ typeutils.ListMeta) kbapi.SecurityDetectionsAPIRelatedIntegration {

			apiIntegration := kbapi.SecurityDetectionsAPIRelatedIntegration{
				Package: integration.Package.ValueString(),
				Version: integration.Version.ValueString(),
			}

			// Set optional integration field if provided
			if typeutils.IsKnown(integration.Integration) {
				integrationName := integration.Integration.ValueString()
				apiIntegration.Integration = &integrationName
			}

			return apiIntegration
		})

	return &apiRelatedIntegrations, diags
}

// Helper function to process required fields configuration for all rule types
func (d Data) requiredFieldsToAPI(ctx context.Context) (*[]kbapi.SecurityDetectionsAPIRequiredFieldInput, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !typeutils.IsKnown(d.RequiredFields) || len(d.RequiredFields.Elements()) == 0 {
		return nil, diags
	}

	apiRequiredFields := typeutils.ListTypeToSlice(ctx, d.RequiredFields, path.Root("required_fields"), &diags,
		func(field RequiredFieldModel, _ typeutils.ListMeta) kbapi.SecurityDetectionsAPIRequiredFieldInput {

			return kbapi.SecurityDetectionsAPIRequiredFieldInput{
				Name: field.Name.ValueString(),
				Type: field.Type.ValueString(),
			}
		})

	return &apiRequiredFields, diags
}

// Helper function to process severity mapping configuration for all rule types
func (d Data) severityMappingToAPI(ctx context.Context) (*kbapi.SecurityDetectionsAPISeverityMapping, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !typeutils.IsKnown(d.SeverityMapping) || len(d.SeverityMapping.Elements()) == 0 {
		return nil, diags
	}

	apiSeverityMapping := typeutils.ListTypeToSlice(ctx, d.SeverityMapping, path.Root("severity_mapping"), &diags,
		func(mapping SeverityMappingModel, _ typeutils.ListMeta) struct {
			Field    string                                             `json:"field"`
			Operator kbapi.SecurityDetectionsAPISeverityMappingOperator `json:"operator"`
			Severity kbapi.SecurityDetectionsAPISeverity                `json:"severity"`
			Value    string                                             `json:"value"`
		} {
			return struct {
				Field    string                                             `json:"field"`
				Operator kbapi.SecurityDetectionsAPISeverityMappingOperator `json:"operator"`
				Severity kbapi.SecurityDetectionsAPISeverity                `json:"severity"`
				Value    string                                             `json:"value"`
			}{
				Field:    mapping.Field.ValueString(),
				Operator: kbapi.SecurityDetectionsAPISeverityMappingOperator(mapping.Operator.ValueString()),
				Severity: kbapi.SecurityDetectionsAPISeverity(mapping.Severity.ValueString()),
				Value:    mapping.Value.ValueString(),
			}
		})

	// Convert to the expected slice type
	severityMappingSlice := make(kbapi.SecurityDetectionsAPISeverityMapping, len(apiSeverityMapping))
	copy(severityMappingSlice, apiSeverityMapping)

	return &severityMappingSlice, diags
}

// filtersToAPI converts the Terraform filters field to the API type
func (d Data) filtersToAPI(ctx context.Context) (*kbapi.SecurityDetectionsAPIRuleFilterArray, diag.Diagnostics) {
	var diags diag.Diagnostics
	_ = ctx

	if !typeutils.IsKnown(d.Filters) {
		return nil, diags
	}

	// Unmarshal the JSON string to []interface{}
	var filters kbapi.SecurityDetectionsAPIRuleFilterArray
	unmarshalDiags := d.Filters.Unmarshal(&filters)
	diags.Append(unmarshalDiags...)

	if diags.HasError() {
		return nil, diags
	}

	return &filters, diags
}

// threatFiltersToAPI converts the Terraform threat_filters list (JSON strings) into the API type ([]interface{}).
func (d Data) threatFiltersToAPI(ctx context.Context) (*kbapi.SecurityDetectionsAPIThreatFilters, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !typeutils.IsKnown(d.ThreatFilters) {
		return nil, diags
	}

	filtersJSON := typeutils.ListTypeAs[string](ctx, d.ThreatFilters, path.Root("threat_filters"), &diags)
	if diags.HasError() {
		return nil, diags
	}

	apiThreatFilters := make(kbapi.SecurityDetectionsAPIThreatFilters, 0, len(filtersJSON))
	for i, filterStr := range filtersJSON {
		var filter any
		if err := json.Unmarshal([]byte(filterStr), &filter); err != nil {
			diags.AddError(
				"Invalid threat_filters JSON",
				fmt.Sprintf("threat_filters[%d] is not valid JSON: %s", i, err.Error()),
			)
			continue
		}
		apiThreatFilters = append(apiThreatFilters, filter)
	}

	return &apiThreatFilters, diags
}

// parseDurationToAPI converts a customtypes.Duration to the API structure
func parseDurationToAPI(duration customtypes.Duration) (kbapi.SecurityDetectionsAPIAlertSuppressionDuration, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !typeutils.IsKnown(duration) {
		diags.AddError("Duration Parse error", "duration string value is unknown")
		return kbapi.SecurityDetectionsAPIAlertSuppressionDuration{}, diags
	}

	// Get the raw duration string (e.g. "5m", "1h", "30s")
	durationStr := duration.ValueString()

	// Parse the duration string using regex to extract value and unit
	durationRegex := regexp.MustCompile(`^(\d+)([smhd])$`)
	matches := durationRegex.FindStringSubmatch(durationStr)

	if len(matches) != 3 {
		diags.AddError(
			"Invalid duration format",
			fmt.Sprintf("Duration '%s' is not in valid format. Expected format: number followed by unit (s, m, h)", durationStr),
		)
		return kbapi.SecurityDetectionsAPIAlertSuppressionDuration{}, diags
	}

	// Parse the numeric value
	value, err := strconv.Atoi(matches[1])
	if err != nil {
		diags.AddError(
			"Invalid duration value",
			fmt.Sprintf("Failed to parse duration value '%s': %s", matches[1], err.Error()),
		)
		return kbapi.SecurityDetectionsAPIAlertSuppressionDuration{}, diags
	}

	// Map the unit from the string to the API unit type
	var unit kbapi.SecurityDetectionsAPIAlertSuppressionDurationUnit
	switch matches[2] {
	case "s":
		unit = kbapi.SecurityDetectionsAPIAlertSuppressionDurationUnitS
	case "m":
		unit = kbapi.SecurityDetectionsAPIAlertSuppressionDurationUnitM
	case "h":
		unit = kbapi.SecurityDetectionsAPIAlertSuppressionDurationUnitH
	case "d":
		// Convert days to hours since API doesn't support days unit
		value *= 24
		unit = kbapi.SecurityDetectionsAPIAlertSuppressionDurationUnitH
	default:
		diags.AddError(
			"Unsupported duration unit",
			fmt.Sprintf("Unit '%s' is not supported. Supported units: s, m, h", matches[2]),
		)
		return kbapi.SecurityDetectionsAPIAlertSuppressionDuration{}, diags
	}

	return kbapi.SecurityDetectionsAPIAlertSuppressionDuration{
		Value: value,
		Unit:  unit,
	}, diags
}
