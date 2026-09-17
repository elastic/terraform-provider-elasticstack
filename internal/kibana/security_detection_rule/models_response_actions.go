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
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

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
	osqueryAction.Params.EcsMapping = buildEcsMappingFromModel(ctx, params.EcsMapping, &diags)
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
				apiQuery.EcsMapping = buildEcsMappingFromModel(ctx, query.EcsMapping, &diags)
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
		case endpointCommandIsolate:
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

		case endpointCommandKillProcess:
			killParams := kbapi.SecurityDetectionsAPIKillProcessParams{
				Command: kbapi.KillProcess,
			}
			if typeutils.IsKnown(params.Comment) {
				killParams.Comment = params.Comment.ValueStringPointer()
			}
			if typeutils.IsKnown(params.Config) {
				config := typeutils.ObjectTypeToStruct(ctx, params.Config, path.Root("response_actions").AtName("params").AtName("config"), &diags,
					func(item EndpointProcessConfigModel, _ typeutils.ObjectMeta) EndpointProcessConfigModel {
						return item
					})
				killParams.Config.Field = config.Field.ValueString()
				if typeutils.IsKnown(config.Overwrite) {
					killParams.Config.Overwrite = config.Overwrite.ValueBoolPointer()
				}
			}
			var processesParams kbapi.SecurityDetectionsAPIProcessesParams
			if err := processesParams.FromSecurityDetectionsAPIKillProcessParams(killParams); err != nil {
				diags.AddError("Error setting endpoint kill-process params", err.Error())
				return kbapi.SecurityDetectionsAPIResponseAction{}, diags
			}
			if err := endpointAction.Params.FromSecurityDetectionsAPIProcessesParams(processesParams); err != nil {
				diags.AddError("Error setting endpoint processes params", err.Error())
				return kbapi.SecurityDetectionsAPIResponseAction{}, diags
			}

		case endpointCommandSuspendProcess:
			suspendParams := kbapi.SecurityDetectionsAPISuspendProcessParams{
				Command: kbapi.SuspendProcess,
			}
			if typeutils.IsKnown(params.Comment) {
				suspendParams.Comment = params.Comment.ValueStringPointer()
			}
			if typeutils.IsKnown(params.Config) {
				config := typeutils.ObjectTypeToStruct(ctx, params.Config, path.Root("response_actions").AtName("params").AtName("config"), &diags,
					func(item EndpointProcessConfigModel, _ typeutils.ObjectMeta) EndpointProcessConfigModel {
						return item
					})
				suspendParams.Config.Field = config.Field.ValueString()
				if typeutils.IsKnown(config.Overwrite) {
					suspendParams.Config.Overwrite = config.Overwrite.ValueBoolPointer()
				}
			}
			var processesParams kbapi.SecurityDetectionsAPIProcessesParams
			if err := processesParams.FromSecurityDetectionsAPISuspendProcessParams(suspendParams); err != nil {
				diags.AddError("Error setting endpoint suspend-process params", err.Error())
				return kbapi.SecurityDetectionsAPIResponseAction{}, diags
			}
			if err := endpointAction.Params.FromSecurityDetectionsAPIProcessesParams(processesParams); err != nil {
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

// Helper function to process response actions configuration for all rule types
func (d Data) responseActionsToAPI(ctx context.Context) ([]kbapi.SecurityDetectionsAPIResponseAction, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !typeutils.IsKnown(d.ResponseActions) || len(d.ResponseActions.Elements()) == 0 {
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

// buildEcsMappingFromModel converts a types.Map to a kbapi ECS mapping pointer.
func buildEcsMappingFromModel(ctx context.Context, value types.Map, diags *diag.Diagnostics) *kbapi.SecurityDetectionsAPIEcsMapping {
	if !typeutils.IsKnown(value) {
		return nil
	}
	elems := make(map[string]basetypes.StringValue)
	if d := value.ElementsAs(ctx, &elems, false); d.HasError() {
		diags.Append(d...)
		return nil
	}
	result := make(kbapi.SecurityDetectionsAPIEcsMapping, len(elems))
	for key, v := range elems {
		if typeutils.IsKnown(v) {
			result[key] = struct {
				Field *string                                      `json:"field,omitempty"`
				Value *kbapi.SecurityDetectionsAPIEcsMapping_Value `json:"value,omitempty"`
			}{Field: v.ValueStringPointer()}
		}
	}
	return &result
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
	paramsModel := ResponseActionParamsModel{
		Query: types.StringPointerValue(osqueryAction.Params.Query)}
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
	paramsModel.EcsMapping = convertEcsMappingToModel(osqueryAction.Params.EcsMapping)

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
			query.Removed = types.BoolPointerValue(apiQuery.Removed)
			query.Snapshot = types.BoolPointerValue(apiQuery.Snapshot)

			// Convert query ECS mapping
			query.EcsMapping = convertEcsMappingToModel(apiQuery.EcsMapping)

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
		case endpointCommandIsolate:
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
		case endpointCommandKillProcess, endpointCommandSuspendProcess:
			processesParams, err := endpointAction.Params.AsSecurityDetectionsAPIProcessesParams()
			if err != nil {
				diags.AddError("Failed to parse endpoint processes params", fmt.Sprintf("Error: %s", err.Error()))
				break
			}
			var (
				command   string
				comment   *string
				field     string
				overwrite *bool
			)
			if commandParams.Command == endpointCommandKillProcess {
				killParams, killErr := processesParams.AsSecurityDetectionsAPIKillProcessParams()
				if killErr != nil {
					diags.AddError("Failed to parse endpoint kill-process params", fmt.Sprintf("Error: %s", killErr.Error()))
					break
				}
				command = string(killParams.Command)
				comment = killParams.Comment
				field = killParams.Config.Field
				overwrite = killParams.Config.Overwrite
			} else {
				suspendParams, suspendErr := processesParams.AsSecurityDetectionsAPISuspendProcessParams()
				if suspendErr != nil {
					diags.AddError("Failed to parse endpoint suspend-process params", fmt.Sprintf("Error: %s", suspendErr.Error()))
					break
				}
				command = string(suspendParams.Command)
				comment = suspendParams.Comment
				field = suspendParams.Config.Field
				overwrite = suspendParams.Config.Overwrite
			}
			paramsModel.Command = types.StringValue(command)
			if comment != nil {
				paramsModel.Comment = types.StringPointerValue(comment)
			} else {
				paramsModel.Comment = types.StringNull()
			}

			configModel := EndpointProcessConfigModel{
				Field: types.StringValue(field),
			}
			if overwrite != nil {
				configModel.Overwrite = types.BoolPointerValue(overwrite)
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

// convertEcsMappingToModel converts a kbapi ECS mapping pointer to a types.Map.
func convertEcsMappingToModel(ecsMapping *kbapi.SecurityDetectionsAPIEcsMapping) types.Map {
	if ecsMapping == nil {
		return types.MapNull(types.StringType)
	}
	attrs := make(map[string]attr.Value, len(*ecsMapping))
	for key, value := range *ecsMapping {
		if value.Field != nil {
			attrs[key] = types.StringPointerValue(value.Field)
		} else {
			attrs[key] = types.StringNull()
		}
	}
	result, _ := types.MapValue(types.StringType, attrs)
	return result
}

func (d *Data) updateResponseActionsFromAPI(ctx context.Context, responseActions *[]kbapi.SecurityDetectionsAPIResponseAction) diag.Diagnostics {
	var diags diag.Diagnostics
	if responseActions != nil && len(*responseActions) > 0 {
		d.ResponseActions, diags = convertResponseActionsToModel(ctx, responseActions)
	} else {
		d.ResponseActions = types.ListNull(getResponseActionElementType())
	}
	return diags
}
