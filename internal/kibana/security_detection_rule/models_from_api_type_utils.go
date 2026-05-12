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
	"strconv"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Utilities to convert various API types to Terraform model types

// commonAPIRuleFields holds the common fields extracted from any API rule response.
// Each updateFrom*Rule function populates this struct and calls updateCommonRuleFieldsFromAPI.
// Fields not applicable to a rule type (e.g. DataViewId for ESQL/ML) should be left nil.
type commonAPIRuleFields struct {
	ResourceID  string // rule.Id.String() — used to build the composite ID
	RuleID      string
	Name        string
	Type        string
	Enabled     bool
	From        string
	To          string
	Interval    string
	Description string
	RiskScore   int64
	Severity    string
	MaxSignals  int64
	Version     int64
	Revision    int64
	CreatedAt   time.Time
	CreatedBy   string
	UpdatedAt   time.Time
	UpdatedBy   string

	TimelineID                        *kbapi.SecurityDetectionsAPITimelineTemplateId
	TimelineTitle                     *kbapi.SecurityDetectionsAPITimelineTemplateTitle
	DataViewID                        *kbapi.SecurityDetectionsAPIDataViewId // nil for ESQL/ML → sets DataViewID to null
	Namespace                         *kbapi.SecurityDetectionsAPIAlertsIndexNamespace
	RuleNameOverride                  *kbapi.SecurityDetectionsAPIRuleNameOverride
	TimestampOverride                 *kbapi.SecurityDetectionsAPITimestampOverride
	TimestampOverrideFallbackDisabled *kbapi.SecurityDetectionsAPITimestampOverrideFallbackDisabled
	BuildingBlockType                 *kbapi.SecurityDetectionsAPIBuildingBlockType
	License                           *kbapi.SecurityDetectionsAPIRuleLicense
	Note                              *kbapi.SecurityDetectionsAPIInvestigationGuide

	Index          *[]string // nil for ESQL/ML → sets Index to empty list
	Author         []string
	Tags           []string
	FalsePositives []string
	References     []string
	Setup          kbapi.SecurityDetectionsAPISetupGuide

	Actions             []kbapi.SecurityDetectionsAPIRuleAction
	ExceptionsList      []kbapi.SecurityDetectionsAPIRuleExceptionList
	RiskScoreMapping    kbapi.SecurityDetectionsAPIRiskScoreMapping
	InvestigationFields *kbapi.SecurityDetectionsAPIInvestigationFields
	Threat              kbapi.SecurityDetectionsAPIThreatArray
	SeverityMapping     kbapi.SecurityDetectionsAPISeverityMapping
	RelatedIntegrations kbapi.SecurityDetectionsAPIRelatedIntegrationArray
	RequiredFields      kbapi.SecurityDetectionsAPIRequiredFieldArray
	// AlertSuppression is nil for Threshold rules (which use a different API type handled separately).
	AlertSuppression *kbapi.SecurityDetectionsAPIAlertSuppression
	ResponseActions  *[]kbapi.SecurityDetectionsAPIResponseAction
}

// updateCommonRuleFieldsFromAPI populates the Data fields that are shared across all rule types.
func (d *Data) updateCommonRuleFieldsFromAPI(ctx context.Context, fields commonAPIRuleFields) diag.Diagnostics {
	var diags diag.Diagnostics

	compID := clients.CompositeID{
		ClusterID:  d.SpaceID.ValueString(),
		ResourceID: fields.ResourceID,
	}
	d.ID = types.StringValue(compID.String())
	d.RuleID = types.StringValue(fields.RuleID)
	d.Name = types.StringValue(fields.Name)
	d.Type = types.StringValue(fields.Type)
	d.Enabled = types.BoolValue(fields.Enabled)
	d.From = types.StringValue(fields.From)
	d.To = types.StringValue(fields.To)
	d.Interval = types.StringValue(fields.Interval)
	d.Description = types.StringValue(fields.Description)
	d.RiskScore = types.Int64Value(fields.RiskScore)
	d.Severity = types.StringValue(fields.Severity)
	d.MaxSignals = types.Int64Value(fields.MaxSignals)
	d.Version = types.Int64Value(fields.Version)
	d.CreatedAt = typeutils.TimeToStringValue(fields.CreatedAt)
	d.CreatedBy = types.StringValue(fields.CreatedBy)
	d.UpdatedAt = typeutils.TimeToStringValue(fields.UpdatedAt)
	d.UpdatedBy = types.StringValue(fields.UpdatedBy)
	d.Revision = types.Int64Value(fields.Revision)

	d.TimelineID = typeutils.StringishPointerValue(fields.TimelineID)
	d.TimelineTitle = typeutils.StringishPointerValue(fields.TimelineTitle)
	d.DataViewID = typeutils.StringishPointerValue(fields.DataViewID)
	d.Namespace = typeutils.StringishPointerValue(fields.Namespace)
	d.RuleNameOverride = typeutils.StringishPointerValue(fields.RuleNameOverride)
	d.TimestampOverride = typeutils.StringishPointerValue(fields.TimestampOverride)
	if fields.TimestampOverrideFallbackDisabled != nil {
		d.TimestampOverrideFallbackDisabled = types.BoolValue(*fields.TimestampOverrideFallbackDisabled)
	} else {
		d.TimestampOverrideFallbackDisabled = types.BoolNull()
	}
	d.BuildingBlockType = typeutils.StringishPointerValue(fields.BuildingBlockType)
	d.License = typeutils.StringishPointerValue(fields.License)
	d.Note = typeutils.StringishPointerValue(fields.Note)
	d.Setup = typeutils.NonEmptyStringishValue(fields.Setup)

	diags.Append(d.updateIndexFromAPI(ctx, fields.Index)...)
	diags.Append(d.updateAuthorFromAPI(ctx, fields.Author)...)
	diags.Append(d.updateTagsFromAPI(ctx, fields.Tags)...)
	diags.Append(d.updateFalsePositivesFromAPI(ctx, fields.FalsePositives)...)
	diags.Append(d.updateReferencesFromAPI(ctx, fields.References)...)

	diags.Append(d.updateActionsFromAPI(ctx, fields.Actions)...)
	diags.Append(d.updateExceptionsListFromAPI(ctx, fields.ExceptionsList)...)
	diags.Append(d.updateRiskScoreMappingFromAPI(ctx, fields.RiskScoreMapping)...)
	diags.Append(d.updateInvestigationFieldsFromAPI(ctx, fields.InvestigationFields)...)
	diags.Append(d.updateThreatFromAPI(ctx, &fields.Threat)...)
	diags.Append(d.updateSeverityMappingFromAPI(ctx, &fields.SeverityMapping)...)
	diags.Append(d.updateRelatedIntegrationsFromAPI(ctx, &fields.RelatedIntegrations)...)
	diags.Append(d.updateRequiredFieldsFromAPI(ctx, &fields.RequiredFields)...)
	diags.Append(d.updateAlertSuppressionFromAPI(ctx, fields.AlertSuppression)...)
	diags.Append(d.updateResponseActionsFromAPI(ctx, fields.ResponseActions)...)

	return diags
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

		if apiAction.Uuid != nil {
			action.UUID = types.StringValue(*apiAction.Uuid)
		} else {
			action.UUID = types.StringNull()
		}

		if apiAction.AlertsFilter != nil {
			alertsFilterMap := make(map[string]attr.Value)
			for k, v := range *apiAction.AlertsFilter {
				if v != nil {
					alertsFilterMap[k] = types.StringValue(fmt.Sprintf("%v", v))
				}
			}
			alertsFilterValue, alertsFilterDiags := types.MapValue(types.StringType, alertsFilterMap)
			diags.Append(alertsFilterDiags...)
			action.AlertsFilter = alertsFilterValue
		} else {
			action.AlertsFilter = types.MapNull(types.StringType)
		}

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

// convertExceptionsListToModel converts kbapi.SecurityDetectionsAPIRuleExceptionList slice to Terraform model
func convertExceptionsListToModel(ctx context.Context, apiExceptionsList []kbapi.SecurityDetectionsAPIRuleExceptionList) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(apiExceptionsList) == 0 {
		return types.ListNull(getExceptionsListElementType()), diags
	}

	exceptions := make([]ExceptionsListModel, 0)

	for _, apiException := range apiExceptionsList {
		exception := ExceptionsListModel{
			ID:            types.StringValue(apiException.Id),
			ListID:        types.StringValue(apiException.ListId),
			NamespaceType: types.StringValue(string(apiException.NamespaceType)),
			Type:          types.StringValue(string(apiException.Type)),
		}

		exceptions = append(exceptions, exception)
	}

	listValue, listDiags := types.ListValueFrom(ctx, getExceptionsListElementType(), exceptions)
	diags.Append(listDiags...)
	return listValue, diags
}

// convertRiskScoreMappingToModel converts kbapi.SecurityDetectionsAPIRiskScoreMapping to Terraform model
func convertRiskScoreMappingToModel(ctx context.Context, apiRiskScoreMapping kbapi.SecurityDetectionsAPIRiskScoreMapping) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(apiRiskScoreMapping) == 0 {
		return types.ListNull(getRiskScoreMappingElementType()), diags
	}

	mappings := make([]RiskScoreMappingModel, 0)

	for _, apiMapping := range apiRiskScoreMapping {
		mapping := RiskScoreMappingModel{
			Field:    types.StringValue(apiMapping.Field),
			Operator: types.StringValue(string(apiMapping.Operator)),
			Value:    types.StringValue(apiMapping.Value),
		}

		// Set optional risk score if provided
		if apiMapping.RiskScore != nil {
			mapping.RiskScore = types.Int64Value(int64(*apiMapping.RiskScore))
		} else {
			mapping.RiskScore = types.Int64Null()
		}

		mappings = append(mappings, mapping)
	}

	listValue, listDiags := types.ListValueFrom(ctx, getRiskScoreMappingElementType(), mappings)
	diags.Append(listDiags...)
	return listValue, diags
}

// convertInvestigationFieldsToModel converts kbapi.SecurityDetectionsAPIInvestigationFields to Terraform model
func convertInvestigationFieldsToModel(ctx context.Context, apiInvestigationFields *kbapi.SecurityDetectionsAPIInvestigationFields) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	if apiInvestigationFields == nil || len(apiInvestigationFields.FieldNames) == 0 {
		return types.ListNull(types.StringType), diags
	}

	fieldNames := make([]string, len(apiInvestigationFields.FieldNames))
	copy(fieldNames, apiInvestigationFields.FieldNames)

	return typeutils.SliceToListTypeString(ctx, fieldNames, path.Root("investigation_fields"), &diags), diags
}

// convertRelatedIntegrationsToModel converts kbapi.SecurityDetectionsAPIRelatedIntegrationArray to Terraform model
func convertRelatedIntegrationsToModel(ctx context.Context, apiRelatedIntegrations *kbapi.SecurityDetectionsAPIRelatedIntegrationArray) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	if apiRelatedIntegrations == nil || len(*apiRelatedIntegrations) == 0 {
		return types.ListNull(getRelatedIntegrationElementType()), diags
	}

	integrations := make([]RelatedIntegrationModel, 0)

	for _, apiIntegration := range *apiRelatedIntegrations {
		integration := RelatedIntegrationModel{
			Package: types.StringValue(apiIntegration.Package),
			Version: types.StringValue(apiIntegration.Version),
		}

		// Set optional integration field if provided
		if apiIntegration.Integration != nil {
			integration.Integration = types.StringValue(*apiIntegration.Integration)
		} else {
			integration.Integration = types.StringNull()
		}

		integrations = append(integrations, integration)
	}

	listValue, listDiags := types.ListValueFrom(ctx, getRelatedIntegrationElementType(), integrations)
	diags.Append(listDiags...)
	return listValue, diags
}

// convertRequiredFieldsToModel converts kbapi.SecurityDetectionsAPIRequiredFieldArray to Terraform model
func convertRequiredFieldsToModel(ctx context.Context, apiRequiredFields *kbapi.SecurityDetectionsAPIRequiredFieldArray) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	if apiRequiredFields == nil || len(*apiRequiredFields) == 0 {
		return types.ListNull(getRequiredFieldElementType()), diags
	}

	fields := make([]RequiredFieldModel, 0)

	for _, apiField := range *apiRequiredFields {
		field := RequiredFieldModel{
			Name: types.StringValue(apiField.Name),
			Type: types.StringValue(apiField.Type),
			Ecs:  types.BoolValue(apiField.Ecs),
		}

		fields = append(fields, field)
	}

	listValue, listDiags := types.ListValueFrom(ctx, getRequiredFieldElementType(), fields)
	diags.Append(listDiags...)
	return listValue, diags
}

// convertSeverityMappingToModel converts kbapi.SecurityDetectionsAPISeverityMapping to Terraform model
func convertSeverityMappingToModel(ctx context.Context, apiSeverityMapping *kbapi.SecurityDetectionsAPISeverityMapping) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	if apiSeverityMapping == nil || len(*apiSeverityMapping) == 0 {
		return types.ListNull(getSeverityMappingElementType()), diags
	}

	mappings := make([]SeverityMappingModel, 0)

	for _, apiMapping := range *apiSeverityMapping {
		mapping := SeverityMappingModel{
			Field:    types.StringValue(apiMapping.Field),
			Operator: types.StringValue(string(apiMapping.Operator)),
			Value:    types.StringValue(apiMapping.Value),
			Severity: types.StringValue(string(apiMapping.Severity)),
		}

		mappings = append(mappings, mapping)
	}

	listValue, listDiags := types.ListValueFrom(ctx, getSeverityMappingElementType(), mappings)
	diags.Append(listDiags...)
	return listValue, diags
}

// convertThreatMappingToModel converts kbapi.SecurityDetectionsAPIThreatMapping to the terraform model
func convertThreatMappingToModel(ctx context.Context, apiThreatMappings kbapi.SecurityDetectionsAPIThreatMapping) (types.List, diag.Diagnostics) {
	var threatMappings []TfDataItem

	for _, apiMapping := range apiThreatMappings {
		var entries []TfDataItemEntry

		for _, apiEntry := range apiMapping.Entries {
			entries = append(entries, TfDataItemEntry{
				Field: types.StringValue(apiEntry.Field),
				Type:  types.StringValue(string(apiEntry.Type)),
				Value: types.StringValue(apiEntry.Value),
			})
		}

		entriesListValue, diags := types.ListValueFrom(ctx, getThreatMappingEntryElementType(), entries)
		if diags.HasError() {
			return types.ListNull(getThreatMappingElementType()), diags
		}

		threatMappings = append(threatMappings, TfDataItem{
			Entries: entriesListValue,
		})
	}

	listValue, diags := types.ListValueFrom(ctx, getThreatMappingElementType(), threatMappings)
	return listValue, diags
}

// convertResponseActionsToModel converts kbapi response actions array to the terraform model
func convertResponseActionsToModel(ctx context.Context, apiResponseActions *[]kbapi.SecurityDetectionsAPIResponseAction) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	if apiResponseActions == nil || len(*apiResponseActions) == 0 {
		return types.ListNull(getResponseActionElementType()), diags
	}

	var responseActions []ResponseActionModel

	for _, apiResponseAction := range *apiResponseActions {
		var responseAction ResponseActionModel

		// Use ValueByDiscriminator to get the concrete type
		actionValue, err := apiResponseAction.ValueByDiscriminator()
		if err != nil {
			diags.AddError("Failed to get response action discriminator", fmt.Sprintf("Error: %s", err.Error()))
			continue
		}

		switch concreteAction := actionValue.(type) {
		case kbapi.SecurityDetectionsAPIOsqueryResponseAction:
			convertedAction, convertDiags := convertOsqueryResponseActionToModel(ctx, concreteAction)
			diags.Append(convertDiags...)
			if !convertDiags.HasError() {
				responseAction = convertedAction
			}

		case kbapi.SecurityDetectionsAPIEndpointResponseAction:
			convertedAction, convertDiags := convertEndpointResponseActionToModel(ctx, concreteAction)
			diags.Append(convertDiags...)
			if !convertDiags.HasError() {
				responseAction = convertedAction
			}

		default:
			diags.AddError("Unknown response action type", fmt.Sprintf("Unsupported response action type: %T", concreteAction))
			continue
		}

		responseActions = append(responseActions, responseAction)
	}

	listValue, listDiags := types.ListValueFrom(ctx, getResponseActionElementType(), responseActions)
	if listDiags.HasError() {
		diags.Append(listDiags...)
	}

	return listValue, diags
}

// convertOsqueryResponseActionToModel converts an Osquery response action to the terraform model
func convertOsqueryResponseActionToModel(ctx context.Context, osqueryAction kbapi.SecurityDetectionsAPIOsqueryResponseAction) (ResponseActionModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	var responseAction ResponseActionModel

	responseAction.ActionTypeID = types.StringValue(string(osqueryAction.ActionTypeId))

	// Convert osquery params
	paramsModel := ResponseActionParamsModel{}
	paramsModel.Query = types.StringPointerValue(osqueryAction.Params.Query)
	if osqueryAction.Params.PackId != nil {
		paramsModel.PackID = types.StringPointerValue(osqueryAction.Params.PackId)
	} else {
		paramsModel.PackID = types.StringNull()
	}
	if osqueryAction.Params.SavedQueryId != nil {
		paramsModel.SavedQueryID = types.StringPointerValue(osqueryAction.Params.SavedQueryId)
	} else {
		paramsModel.SavedQueryID = types.StringNull()
	}
	if osqueryAction.Params.Timeout != nil {
		paramsModel.Timeout = types.Int64Value(int64(*osqueryAction.Params.Timeout))
	} else {
		paramsModel.Timeout = types.Int64Null()
	}

	// Convert ECS mapping
	if osqueryAction.Params.EcsMapping != nil {
		ecsMappingAttrs := make(map[string]attr.Value)
		for key, value := range *osqueryAction.Params.EcsMapping {
			if value.Field != nil {
				ecsMappingAttrs[key] = types.StringPointerValue(value.Field)
			} else {
				ecsMappingAttrs[key] = types.StringNull()
			}
		}
		ecsMappingValue, ecsDiags := types.MapValue(types.StringType, ecsMappingAttrs)
		if ecsDiags.HasError() {
			diags.Append(ecsDiags...)
		} else {
			paramsModel.EcsMapping = ecsMappingValue
		}
	} else {
		paramsModel.EcsMapping = types.MapNull(types.StringType)
	}

	// Convert queries array
	if osqueryAction.Params.Queries != nil {
		var queries []OsqueryQueryModel
		for _, apiQuery := range *osqueryAction.Params.Queries {
			query := OsqueryQueryModel{
				ID:    types.StringValue(apiQuery.Id),
				Query: types.StringValue(apiQuery.Query),
			}
			if apiQuery.Platform != nil {
				query.Platform = types.StringPointerValue(apiQuery.Platform)
			} else {
				query.Platform = types.StringNull()
			}
			if apiQuery.Version != nil {
				query.Version = types.StringPointerValue(apiQuery.Version)
			} else {
				query.Version = types.StringNull()
			}
			if apiQuery.Removed != nil {
				query.Removed = types.BoolPointerValue(apiQuery.Removed)
			} else {
				query.Removed = types.BoolNull()
			}
			if apiQuery.Snapshot != nil {
				query.Snapshot = types.BoolPointerValue(apiQuery.Snapshot)
			} else {
				query.Snapshot = types.BoolNull()
			}

			// Convert query ECS mapping
			if apiQuery.EcsMapping != nil {
				queryEcsMappingAttrs := make(map[string]attr.Value)
				for key, value := range *apiQuery.EcsMapping {
					if value.Field != nil {
						queryEcsMappingAttrs[key] = types.StringPointerValue(value.Field)
					} else {
						queryEcsMappingAttrs[key] = types.StringNull()
					}
				}
				queryEcsMappingValue, queryEcsDiags := types.MapValue(types.StringType, queryEcsMappingAttrs)
				if queryEcsDiags.HasError() {
					diags.Append(queryEcsDiags...)
				} else {
					query.EcsMapping = queryEcsMappingValue
				}
			} else {
				query.EcsMapping = types.MapNull(types.StringType)
			}

			queries = append(queries, query)
		}

		queriesListValue, queriesDiags := types.ListValueFrom(ctx, getOsqueryQueryElementType(), queries)
		if queriesDiags.HasError() {
			diags.Append(queriesDiags...)
		} else {
			paramsModel.Queries = queriesListValue
		}
	} else {
		paramsModel.Queries = types.ListNull(getOsqueryQueryElementType())
	}

	// Set remaining fields to null since this is osquery
	paramsModel.Command = types.StringNull()
	paramsModel.Comment = types.StringNull()
	paramsModel.Config = types.ObjectNull(getEndpointProcessConfigType())

	paramsObjectValue, paramsDiags := types.ObjectValueFrom(ctx, getResponseActionParamsType(), paramsModel)
	if paramsDiags.HasError() {
		diags.Append(paramsDiags...)
	} else {
		responseAction.Params = paramsObjectValue
	}

	return responseAction, diags
}

// convertEndpointResponseActionToModel converts an Endpoint response action to the terraform model
func convertEndpointResponseActionToModel(ctx context.Context, endpointAction kbapi.SecurityDetectionsAPIEndpointResponseAction) (ResponseActionModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	var responseAction ResponseActionModel

	responseAction.ActionTypeID = types.StringValue(string(endpointAction.ActionTypeId))

	// Convert endpoint params
	paramsModel := ResponseActionParamsModel{}

	commandParams, err := endpointAction.Params.AsSecurityDetectionsAPIDefaultParams()
	if err == nil {
		switch commandParams.Command {
		case "isolate":
			defaultParams, err := endpointAction.Params.AsSecurityDetectionsAPIDefaultParams()
			if err != nil {
				diags.AddError("Failed to parse endpoint default params", fmt.Sprintf("Error: %s", err.Error()))
			} else {
				paramsModel.Command = types.StringValue(string(defaultParams.Command))
				if defaultParams.Comment != nil {
					paramsModel.Comment = types.StringPointerValue(defaultParams.Comment)
				} else {
					paramsModel.Comment = types.StringNull()
				}
				paramsModel.Config = types.ObjectNull(getEndpointProcessConfigType())
			}
		case "kill-process", "suspend-process":
			processesParams, err := endpointAction.Params.AsSecurityDetectionsAPIProcessesParams()
			if err != nil {
				diags.AddError("Failed to parse endpoint processes params", fmt.Sprintf("Error: %s", err.Error()))
			} else {
				paramsModel.Command = types.StringValue(string(processesParams.Command))
				if processesParams.Comment != nil {
					paramsModel.Comment = types.StringPointerValue(processesParams.Comment)
				} else {
					paramsModel.Comment = types.StringNull()
				}

				// Convert config
				configModel := EndpointProcessConfigModel{
					Field: types.StringValue(processesParams.Config.Field),
				}
				if processesParams.Config.Overwrite != nil {
					configModel.Overwrite = types.BoolPointerValue(processesParams.Config.Overwrite)
				} else {
					configModel.Overwrite = types.BoolNull()
				}

				configObjectValue, configDiags := types.ObjectValueFrom(ctx, getEndpointProcessConfigType(), configModel)
				if configDiags.HasError() {
					diags.Append(configDiags...)
				} else {
					paramsModel.Config = configObjectValue
				}
			}
		}
	} else {
		diags.AddError("Unknown endpoint command", fmt.Sprintf("Unsupported endpoint command: %s. Error: %s", commandParams.Command, err.Error()))
	}

	// Set osquery fields to null since this is endpoint
	paramsModel.Query = types.StringNull()
	paramsModel.PackID = types.StringNull()
	paramsModel.SavedQueryID = types.StringNull()
	paramsModel.Timeout = types.Int64Null()
	paramsModel.EcsMapping = types.MapNull(types.StringType)
	paramsModel.Queries = types.ListNull(getOsqueryQueryElementType())

	paramsObjectValue, paramsDiags := types.ObjectValueFrom(ctx, getResponseActionParamsType(), paramsModel)
	if paramsDiags.HasError() {
		diags.Append(paramsDiags...)
	} else {
		responseAction.Params = paramsObjectValue
	}

	return responseAction, diags
}

// convertThresholdToModel converts kbapi.SecurityDetectionsAPIThreshold to the terraform model
func convertThresholdToModel(ctx context.Context, apiThreshold kbapi.SecurityDetectionsAPIThreshold) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Handle threshold field - can be single string or array
	var fieldList types.List
	if singleField, err := apiThreshold.Field.AsSecurityDetectionsAPIThresholdField0(); err == nil {
		// Single field
		fieldList = typeutils.SliceToListTypeString(ctx, []string{singleField}, path.Root("threshold").AtName("field"), &diags)
	} else if multipleFields, err := apiThreshold.Field.AsSecurityDetectionsAPIThresholdField1(); err == nil {
		// Multiple fields
		fieldStrings := make([]string, len(multipleFields))
		copy(fieldStrings, multipleFields)
		fieldList = typeutils.SliceToListTypeString(ctx, fieldStrings, path.Root("threshold").AtName("field"), &diags)
	} else {
		fieldList = types.ListValueMust(types.StringType, []attr.Value{})
	}

	// Handle cardinality (optional)
	var cardinalityList types.List
	if apiThreshold.Cardinality != nil && len(*apiThreshold.Cardinality) > 0 {
		cardinalityList = typeutils.SliceToListType(ctx, *apiThreshold.Cardinality, getCardinalityType(), path.Root("threshold").AtName("cardinality"), &diags,
			func(item struct {
				Field string `json:"field"`
				Value int    `json:"value"`
			}, _ typeutils.ListMeta) CardinalityModel {
				return CardinalityModel{
					Field: types.StringValue(item.Field),
					Value: types.Int64Value(int64(item.Value)),
				}
			})
	} else {
		cardinalityList = types.ListNull(getCardinalityType())
	}

	thresholdModel := ThresholdModel{
		Field:       fieldList,
		Value:       types.Int64Value(int64(apiThreshold.Value)),
		Cardinality: cardinalityList,
	}

	thresholdObject, objDiags := types.ObjectValueFrom(ctx, getThresholdType(), thresholdModel)
	diags.Append(objDiags...)
	return thresholdObject, diags
}

// convertFiltersFromAPI converts the API filters field back to the Terraform type
func (d *Data) updateFiltersFromAPI(ctx context.Context, apiFilters *kbapi.SecurityDetectionsAPIRuleFilterArray) diag.Diagnostics {
	var diags diag.Diagnostics
	_ = ctx

	if apiFilters == nil || len(*apiFilters) == 0 {
		d.Filters = jsontypes.NewNormalizedNull()
		return diags
	}

	// Marshal the []interface{} to JSON string
	jsonBytes, err := json.Marshal(*apiFilters)
	if err != nil {
		diags.AddError("Failed to marshal filters", err.Error())
		return diags
	}

	// Create a NormalizedValue from the JSON string
	d.Filters = jsontypes.NewNormalizedValue(string(jsonBytes))
	return diags
}

func (d *Data) updateThreatFiltersFromAPI(ctx context.Context, apiThreatFilters *kbapi.SecurityDetectionsAPIThreatFilters) diag.Diagnostics {
	var diags diag.Diagnostics
	_ = ctx

	if apiThreatFilters == nil {
		d.ThreatFilters = types.ListNull(types.StringType)
		return diags
	}

	if len(*apiThreatFilters) == 0 {
		d.ThreatFilters = types.ListValueMust(types.StringType, []attr.Value{})
		return diags
	}

	filters := make([]string, 0, len(*apiThreatFilters))
	for i, filter := range *apiThreatFilters {
		jsonBytes, err := json.Marshal(filter)
		if err != nil {
			diags.AddError("Failed to marshal threat_filters item", fmt.Sprintf("threat_filters[%d]: %s", i, err.Error()))
			continue
		}
		filters = append(filters, string(jsonBytes))
	}

	d.ThreatFilters = typeutils.ListValueFrom(ctx, filters, types.StringType, path.Root("threat_filters"), &diags)
	return diags
}

// Helper function to update severity mapping from API response
func (d *Data) updateSeverityMappingFromAPI(ctx context.Context, severityMapping *kbapi.SecurityDetectionsAPISeverityMapping) diag.Diagnostics {
	var diags diag.Diagnostics

	if severityMapping != nil && len(*severityMapping) > 0 {
		severityMappingValue, severityMappingDiags := convertSeverityMappingToModel(ctx, severityMapping)
		diags.Append(severityMappingDiags...)
		if !severityMappingDiags.HasError() {
			d.SeverityMapping = severityMappingValue
		}
	} else {
		d.SeverityMapping = types.ListNull(getSeverityMappingElementType())
	}

	return diags
}

// Helper function to update index patterns from API response
func (d *Data) updateIndexFromAPI(ctx context.Context, index *[]string) diag.Diagnostics {
	var diags diag.Diagnostics

	if index != nil && len(*index) > 0 {
		d.Index = typeutils.ListValueFrom(ctx, *index, types.StringType, path.Root("index"), &diags)
	} else {
		d.Index = types.ListValueMust(types.StringType, []attr.Value{})
	}

	return diags
}

// Helper function to update author from API response
func (d *Data) updateAuthorFromAPI(ctx context.Context, author []string) diag.Diagnostics {
	var diags diag.Diagnostics

	if len(author) > 0 {
		d.Author = typeutils.ListValueFrom(ctx, author, types.StringType, path.Root("author"), &diags)
	} else {
		d.Author = types.ListValueMust(types.StringType, []attr.Value{})
	}

	return diags
}

// Helper function to update tags from API response
func (d *Data) updateTagsFromAPI(ctx context.Context, tags []string) diag.Diagnostics {
	var diags diag.Diagnostics

	if len(tags) > 0 {
		d.Tags = typeutils.ListValueFrom(ctx, tags, types.StringType, path.Root("tags"), &diags)
	} else {
		d.Tags = types.ListValueMust(types.StringType, []attr.Value{})
	}

	return diags
}

// Helper function to update false positives from API response
func (d *Data) updateFalsePositivesFromAPI(ctx context.Context, falsePositives []string) diag.Diagnostics {
	var diags diag.Diagnostics

	d.FalsePositives = typeutils.ListValueFrom(ctx, falsePositives, types.StringType, path.Root("false_positives"), &diags)

	return diags
}

// Helper function to update references from API response
func (d *Data) updateReferencesFromAPI(ctx context.Context, references []string) diag.Diagnostics {
	var diags diag.Diagnostics

	if len(references) > 0 {
		d.References = typeutils.ListValueFrom(ctx, references, types.StringType, path.Root("references"), &diags)
	} else {
		d.References = types.ListValueMust(types.StringType, []attr.Value{})
	}

	return diags
}

// Helper function to update exceptions list from API response
func (d *Data) updateExceptionsListFromAPI(ctx context.Context, exceptionsList []kbapi.SecurityDetectionsAPIRuleExceptionList) diag.Diagnostics {
	var diags diag.Diagnostics

	if len(exceptionsList) > 0 {
		exceptionsListValue, exceptionsListDiags := convertExceptionsListToModel(ctx, exceptionsList)
		diags.Append(exceptionsListDiags...)
		if !exceptionsListDiags.HasError() {
			d.ExceptionsList = exceptionsListValue
		}
	} else {
		d.ExceptionsList = types.ListNull(getExceptionsListElementType())
	}

	return diags
}

// Helper function to update risk score mapping from API response
func (d *Data) updateRiskScoreMappingFromAPI(ctx context.Context, riskScoreMapping kbapi.SecurityDetectionsAPIRiskScoreMapping) diag.Diagnostics {
	var diags diag.Diagnostics

	if len(riskScoreMapping) > 0 {
		riskScoreMappingValue, riskScoreMappingDiags := convertRiskScoreMappingToModel(ctx, riskScoreMapping)
		diags.Append(riskScoreMappingDiags...)
		if !riskScoreMappingDiags.HasError() {
			d.RiskScoreMapping = riskScoreMappingValue
		}
	} else {
		d.RiskScoreMapping = types.ListNull(getRiskScoreMappingElementType())
	}

	return diags
}

// Helper function to update actions from API response
func (d *Data) updateActionsFromAPI(ctx context.Context, actions []kbapi.SecurityDetectionsAPIRuleAction) diag.Diagnostics {
	var diags diag.Diagnostics

	if len(actions) > 0 {
		actionsListValue, actionDiags := convertActionsToModel(ctx, actions)
		diags.Append(actionDiags...)
		if !actionDiags.HasError() {
			d.Actions = actionsListValue
		}
	} else {
		d.Actions = types.ListNull(getActionElementType())
	}

	return diags
}

func (d *Data) updateAlertSuppressionFromAPI(ctx context.Context, apiSuppression *kbapi.SecurityDetectionsAPIAlertSuppression) diag.Diagnostics {
	var diags diag.Diagnostics

	if apiSuppression == nil {
		d.AlertSuppression = types.ObjectNull(getAlertSuppressionType())
		return diags
	}

	model := AlertSuppressionModel{}

	// Convert group_by (required field according to API)
	if len(apiSuppression.GroupBy) > 0 {
		groupByList := make([]attr.Value, len(apiSuppression.GroupBy))
		for i, field := range apiSuppression.GroupBy {
			groupByList[i] = types.StringValue(field)
		}
		model.GroupBy = types.ListValueMust(types.StringType, groupByList)
	} else {
		model.GroupBy = types.ListNull(types.StringType)
	}

	// Convert duration (optional)
	if apiSuppression.Duration != nil {
		model.Duration = parseDurationFromAPI(*apiSuppression.Duration)
	} else {
		model.Duration = customtypes.NewDurationNull()
	}

	// Convert missing_fields_strategy (optional)
	if apiSuppression.MissingFieldsStrategy != nil {
		model.MissingFieldsStrategy = types.StringValue(string(*apiSuppression.MissingFieldsStrategy))
	} else {
		model.MissingFieldsStrategy = types.StringNull()
	}

	alertSuppressionObj, objDiags := types.ObjectValueFrom(ctx, getAlertSuppressionType(), model)
	diags.Append(objDiags...)

	d.AlertSuppression = alertSuppressionObj

	return diags
}

func (d *Data) updateThresholdAlertSuppressionFromAPI(ctx context.Context, apiSuppression *kbapi.SecurityDetectionsAPIThresholdAlertSuppression) diag.Diagnostics {
	var diags diag.Diagnostics

	if apiSuppression == nil {
		d.AlertSuppression = types.ObjectNull(getAlertSuppressionType())
		return diags
	}

	model := AlertSuppressionModel{}

	// Threshold alert suppression only has duration field, so we set group_by and missing_fields_strategy to null
	model.GroupBy = types.ListNull(types.StringType)
	model.MissingFieldsStrategy = types.StringNull()

	// Convert duration (always present in threshold alert suppression)
	model.Duration = parseDurationFromAPI(apiSuppression.Duration)

	alertSuppressionObj, objDiags := types.ObjectValueFrom(ctx, getAlertSuppressionType(), model)
	diags.Append(objDiags...)

	d.AlertSuppression = alertSuppressionObj

	return diags
}

// updateResponseActionsFromAPI updates the ResponseActions field from API response
func (d *Data) updateResponseActionsFromAPI(ctx context.Context, responseActions *[]kbapi.SecurityDetectionsAPIResponseAction) diag.Diagnostics {
	var diags diag.Diagnostics

	if responseActions != nil && len(*responseActions) > 0 {
		responseActionsValue, responseActionsDiags := convertResponseActionsToModel(ctx, responseActions)
		diags.Append(responseActionsDiags...)
		if !responseActionsDiags.HasError() {
			d.ResponseActions = responseActionsValue
		}
	} else {
		d.ResponseActions = types.ListNull(getResponseActionElementType())
	}

	return diags
}

// Helper function to update investigation fields from API response
func (d *Data) updateInvestigationFieldsFromAPI(ctx context.Context, investigationFields *kbapi.SecurityDetectionsAPIInvestigationFields) diag.Diagnostics {
	var diags diag.Diagnostics

	investigationFieldsValue, investigationFieldsDiags := convertInvestigationFieldsToModel(ctx, investigationFields)
	diags.Append(investigationFieldsDiags...)
	if diags.HasError() {
		return diags
	}
	d.InvestigationFields = investigationFieldsValue

	return diags
}

// Helper function to update related integrations from API response
func (d *Data) updateRelatedIntegrationsFromAPI(ctx context.Context, relatedIntegrations *kbapi.SecurityDetectionsAPIRelatedIntegrationArray) diag.Diagnostics {
	var diags diag.Diagnostics

	if relatedIntegrations != nil && len(*relatedIntegrations) > 0 {
		relatedIntegrationsValue, relatedIntegrationsDiags := convertRelatedIntegrationsToModel(ctx, relatedIntegrations)
		diags.Append(relatedIntegrationsDiags...)
		if !relatedIntegrationsDiags.HasError() {
			d.RelatedIntegrations = relatedIntegrationsValue
		}
	} else {
		d.RelatedIntegrations = types.ListNull(getRelatedIntegrationElementType())
	}

	return diags
}

// Helper function to update required fields from API response
func (d *Data) updateRequiredFieldsFromAPI(ctx context.Context, requiredFields *kbapi.SecurityDetectionsAPIRequiredFieldArray) diag.Diagnostics {
	var diags diag.Diagnostics

	if requiredFields != nil && len(*requiredFields) > 0 {
		requiredFieldsValue, requiredFieldsDiags := convertRequiredFieldsToModel(ctx, requiredFields)
		diags.Append(requiredFieldsDiags...)
		if !requiredFieldsDiags.HasError() {
			d.RequiredFields = requiredFieldsValue
		}
	} else {
		d.RequiredFields = types.ListNull(getRequiredFieldElementType())
	}

	return diags
}

// convertThreatToModel converts kbapi.SecurityDetectionsAPIThreatArray to Terraform model
func convertThreatToModel(ctx context.Context, apiThreats *kbapi.SecurityDetectionsAPIThreatArray) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	if apiThreats == nil || len(*apiThreats) == 0 {
		return types.ListNull(getThreatElementType()), diags
	}

	threats := make([]ThreatModel, 0)

	for _, apiThreat := range *apiThreats {
		threat := ThreatModel{
			Framework: types.StringValue(apiThreat.Framework),
		}

		// Convert tactic
		tacticModel := ThreatTacticModel{
			ID:        types.StringValue(apiThreat.Tactic.Id),
			Name:      types.StringValue(apiThreat.Tactic.Name),
			Reference: types.StringValue(apiThreat.Tactic.Reference),
		}

		tacticObj, tacticDiags := types.ObjectValueFrom(ctx, getThreatTacticType(), tacticModel)
		diags.Append(tacticDiags...)
		if tacticDiags.HasError() {
			continue
		}
		threat.Tactic = tacticObj

		// Convert techniques (optional)
		if apiThreat.Technique != nil && len(*apiThreat.Technique) > 0 {
			techniques := make([]ThreatTechniqueModel, 0)

			for _, apiTechnique := range *apiThreat.Technique {
				technique := ThreatTechniqueModel{
					ID:        types.StringValue(apiTechnique.Id),
					Name:      types.StringValue(apiTechnique.Name),
					Reference: types.StringValue(apiTechnique.Reference),
				}

				// Convert subtechniques (optional)
				if apiTechnique.Subtechnique != nil && len(*apiTechnique.Subtechnique) > 0 {
					subtechniques := make([]ThreatSubtechniqueModel, 0)

					for _, apiSubtechnique := range *apiTechnique.Subtechnique {
						subtechnique := ThreatSubtechniqueModel{
							ID:        types.StringValue(apiSubtechnique.Id),
							Name:      types.StringValue(apiSubtechnique.Name),
							Reference: types.StringValue(apiSubtechnique.Reference),
						}
						subtechniques = append(subtechniques, subtechnique)
					}

					subtechniquesList, subtechniquesListDiags := types.ListValueFrom(ctx, getThreatSubtechniqueElementType(), subtechniques)
					diags.Append(subtechniquesListDiags...)
					if !subtechniquesListDiags.HasError() {
						technique.Subtechnique = subtechniquesList
					}
				} else {
					technique.Subtechnique = types.ListNull(getThreatSubtechniqueElementType())
				}

				techniques = append(techniques, technique)
			}

			techniquesList, techniquesListDiags := types.ListValueFrom(ctx, getThreatTechniqueElementType(), techniques)
			diags.Append(techniquesListDiags...)
			if !techniquesListDiags.HasError() {
				threat.Technique = techniquesList
			}
		} else {
			threat.Technique = types.ListNull(getThreatTechniqueElementType())
		}

		threats = append(threats, threat)
	}

	listValue, listDiags := types.ListValueFrom(ctx, getThreatElementType(), threats)
	diags.Append(listDiags...)
	return listValue, diags
}

// Helper function to update threat from API response
func (d *Data) updateThreatFromAPI(ctx context.Context, threat *kbapi.SecurityDetectionsAPIThreatArray) diag.Diagnostics {
	var diags diag.Diagnostics

	if threat != nil && len(*threat) > 0 {
		threatValue, threatDiags := convertThreatToModel(ctx, threat)
		diags.Append(threatDiags...)
		if !threatDiags.HasError() {
			d.Threat = threatValue
		}
	} else {
		d.Threat = types.ListNull(getThreatElementType())
	}

	return diags
}

// parseDurationFromAPI converts an API duration to customtypes.Duration
func parseDurationFromAPI(apiDuration kbapi.SecurityDetectionsAPIAlertSuppressionDuration) customtypes.Duration {
	// Convert the API's Value + Unit format back to a duration string
	durationStr := strconv.Itoa(apiDuration.Value) + string(apiDuration.Unit)
	return customtypes.NewDurationValue(durationStr)
}
