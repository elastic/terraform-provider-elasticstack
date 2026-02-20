package securitydetectionrule

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Utilities to convert various API types to Terraform model types

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

		// Convert params
		if apiAction.Params != nil {
			paramsMap := make(map[string]attr.Value)
			for k, v := range apiAction.Params {
				if v != nil {
					paramsMap[k] = types.StringValue(fmt.Sprintf("%v", v))
				}
			}
			paramsValue, paramsDiags := types.MapValue(types.StringType, paramsMap)
			diags.Append(paramsDiags...)
			action.Params = paramsValue
		} else {
			action.Params = types.MapNull(types.StringType)
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

// Helper function to update data view ID from API response
func (d *Data) updateDataViewIDFromAPI(ctx context.Context, dataViewID *kbapi.SecurityDetectionsAPIDataViewId) diag.Diagnostics {
	diags := diag.Diagnostics{}
	_ = ctx

	if dataViewID != nil {
		d.DataViewID = types.StringValue(*dataViewID)
	} else {
		d.DataViewID = types.StringNull()
	}

	return diags
}

// Helper function to update timeline ID from API response
func (d *Data) updateTimelineIDFromAPI(ctx context.Context, timelineID *kbapi.SecurityDetectionsAPITimelineTemplateId) diag.Diagnostics {
	diags := diag.Diagnostics{}
	_ = ctx

	if timelineID != nil {
		d.TimelineID = types.StringValue(*timelineID)
	} else {
		d.TimelineID = types.StringNull()
	}

	return diags
}

// Helper function to update timeline title from API response
func (d *Data) updateTimelineTitleFromAPI(ctx context.Context, timelineTitle *kbapi.SecurityDetectionsAPITimelineTemplateTitle) diag.Diagnostics {
	diags := diag.Diagnostics{}
	_ = ctx

	if timelineTitle != nil {
		d.TimelineTitle = types.StringValue(*timelineTitle)
	} else {
		d.TimelineTitle = types.StringNull()
	}

	return diags
}

// Helper function to update namespace from API response
func (d *Data) updateNamespaceFromAPI(ctx context.Context, namespace *kbapi.SecurityDetectionsAPIAlertsIndexNamespace) diag.Diagnostics {
	diags := diag.Diagnostics{}
	_ = ctx

	if namespace != nil {
		d.Namespace = types.StringValue(*namespace)
	} else {
		d.Namespace = types.StringNull()
	}

	return diags
}

// Helper function to update rule name override from API response
func (d *Data) updateRuleNameOverrideFromAPI(ctx context.Context, ruleNameOverride *kbapi.SecurityDetectionsAPIRuleNameOverride) diag.Diagnostics {
	diags := diag.Diagnostics{}
	_ = ctx

	if ruleNameOverride != nil {
		d.RuleNameOverride = types.StringValue(*ruleNameOverride)
	} else {
		d.RuleNameOverride = types.StringNull()
	}

	return diags
}

// Helper function to update timestamp override from API response
func (d *Data) updateTimestampOverrideFromAPI(ctx context.Context, timestampOverride *kbapi.SecurityDetectionsAPITimestampOverride) diag.Diagnostics {
	diags := diag.Diagnostics{}
	_ = ctx

	if timestampOverride != nil {
		d.TimestampOverride = types.StringValue(*timestampOverride)
	} else {
		d.TimestampOverride = types.StringNull()
	}

	return diags
}

// Helper function to update timestamp override fallback disabled from API response
func (d *Data) updateTimestampOverrideFallbackDisabledFromAPI(
	ctx context.Context,
	timestampOverrideFallbackDisabled *kbapi.SecurityDetectionsAPITimestampOverrideFallbackDisabled,
) diag.Diagnostics {
	diags := diag.Diagnostics{}
	_ = ctx

	if timestampOverrideFallbackDisabled != nil {
		d.TimestampOverrideFallbackDisabled = types.BoolValue(*timestampOverrideFallbackDisabled)
	} else {
		d.TimestampOverrideFallbackDisabled = types.BoolNull()
	}

	return diags
}

// Helper function to update building block type from API response
func (d *Data) updateBuildingBlockTypeFromAPI(ctx context.Context, buildingBlockType *kbapi.SecurityDetectionsAPIBuildingBlockType) diag.Diagnostics {
	diags := diag.Diagnostics{}
	_ = ctx

	if buildingBlockType != nil {
		d.BuildingBlockType = types.StringValue(*buildingBlockType)
	} else {
		d.BuildingBlockType = types.StringNull()
	}

	return diags
}

// Helper function to update license from API response
func (d *Data) updateLicenseFromAPI(ctx context.Context, license *kbapi.SecurityDetectionsAPIRuleLicense) diag.Diagnostics {
	diags := diag.Diagnostics{}
	_ = ctx

	if license != nil {
		d.License = types.StringValue(*license)
	} else {
		d.License = types.StringNull()
	}

	return diags
}

// Helper function to update note from API response
func (d *Data) updateNoteFromAPI(ctx context.Context, note *kbapi.SecurityDetectionsAPIInvestigationGuide) diag.Diagnostics {
	diags := diag.Diagnostics{}
	_ = ctx

	if note != nil {
		d.Note = types.StringValue(*note)
	} else {
		d.Note = types.StringNull()
	}

	return diags
}

// Helper function to update setup from API response
func (d *Data) updateSetupFromAPI(ctx context.Context, setup kbapi.SecurityDetectionsAPISetupGuide) diag.Diagnostics {
	diags := diag.Diagnostics{}
	_ = ctx

	// Handle setup field - if empty, set to null to maintain consistency with optional schema
	if setup != "" {
		d.Setup = types.StringValue(setup)
	} else {
		d.Setup = types.StringNull()
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
