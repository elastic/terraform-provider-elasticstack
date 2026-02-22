package integrationpolicy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type features struct {
	SupportsPolicyIDs bool
	SupportsOutputID  bool
}

type integrationPolicyModel struct {
	ID                 types.String  `tfsdk:"id"`
	PolicyID           types.String  `tfsdk:"policy_id"`
	Name               types.String  `tfsdk:"name"`
	Namespace          types.String  `tfsdk:"namespace"`
	AgentPolicyID      types.String  `tfsdk:"agent_policy_id"`
	AgentPolicyIDs     types.List    `tfsdk:"agent_policy_ids"`
	Description        types.String  `tfsdk:"description"`
	Enabled            types.Bool    `tfsdk:"enabled"`
	Force              types.Bool    `tfsdk:"force"`
	IntegrationName    types.String  `tfsdk:"integration_name"`
	IntegrationVersion types.String  `tfsdk:"integration_version"`
	OutputID           types.String  `tfsdk:"output_id"`
	Inputs             InputsValue   `tfsdk:"inputs"` // > integrationPolicyInputsModel
	VarsJSON           VarsJSONValue `tfsdk:"vars_json"`
	SpaceIDs           types.Set     `tfsdk:"space_ids"`
}

type integrationPolicyInputsModel struct {
	Enabled  types.Bool           `tfsdk:"enabled"`
	Vars     jsontypes.Normalized `tfsdk:"vars"`
	Defaults types.Object         `tfsdk:"defaults"` // > inputDefaultsModel
	Streams  types.Map            `tfsdk:"streams"`  // > integrationPolicyInputStreamModel
}

type integrationPolicyInputStreamModel struct {
	Enabled types.Bool           `tfsdk:"enabled"`
	Vars    jsontypes.Normalized `tfsdk:"vars"`
}

func (model *integrationPolicyModel) populateFromAPI(ctx context.Context, pkg *kbapi.PackageInfo, data *kbapi.PackagePolicy) diag.Diagnostics {
	if data == nil {
		return nil
	}

	var diags diag.Diagnostics

	model.ID = types.StringValue(data.Id)
	model.PolicyID = types.StringValue(data.Id)
	model.Name = types.StringValue(data.Name)
	model.Namespace = types.StringPointerValue(data.Namespace)

	// Only populate the agent policy field that was originally configured
	// to avoid Terraform detecting inconsistent state
	originallyUsedAgentPolicyID := typeutils.IsKnown(model.AgentPolicyID)
	originallyUsedAgentPolicyIDs := typeutils.IsKnown(model.AgentPolicyIDs)

	if originallyUsedAgentPolicyID {
		model.AgentPolicyID = types.StringPointerValue(data.PolicyId)
	}
	if originallyUsedAgentPolicyIDs {
		if data.PolicyIds != nil {
			agentPolicyIDs, d := types.ListValueFrom(ctx, types.StringType, *data.PolicyIds)
			diags.Append(d...)
			model.AgentPolicyIDs = agentPolicyIDs
		} else {
			model.AgentPolicyIDs = types.ListNull(types.StringType)
		}
	}

	if !originallyUsedAgentPolicyID && !originallyUsedAgentPolicyIDs {
		// Handle edge cases: both fields configured or neither configured
		// Default to the behavior based on API response structure
		if data.PolicyIds != nil && len(*data.PolicyIds) > 1 {
			// Multiple policy IDs, use agent_policy_ids
			agentPolicyIDs, d := types.ListValueFrom(ctx, types.StringType, *data.PolicyIds)
			diags.Append(d...)
			model.AgentPolicyIDs = agentPolicyIDs
		} else if data.PolicyId != nil {
			// Single policy ID, use agent_policy_id
			model.AgentPolicyID = types.StringPointerValue(data.PolicyId)
		}
	}

	model.Description = types.StringPointerValue(data.Description)
	model.Enabled = types.BoolValue(data.Enabled)
	model.IntegrationName = types.StringValue(data.Package.Name)
	model.IntegrationVersion = types.StringValue(data.Package.Version)
	model.OutputID = types.StringPointerValue(data.OutputId)

	varsMap := schemautil.Deref(data.Vars)
	if varsMap == nil {
		model.VarsJSON = NewVarsJSONNull()
	} else {
		jsonBytes, err := json.Marshal(varsMap)
		if err != nil {
			diags.AddError("Failed to marshal vars", err.Error())
		} else {
			var d diag.Diagnostics
			model.VarsJSON, d = NewVarsJSONWithIntegration(string(jsonBytes), data.Package.Name, data.Package.Version)
			diags.Append(d...)
		}
	}

	// Preserve space_ids if it was originally set in the plan/state
	// The API response may not include space_ids, so we keep the original value
	originallySetSpaceIDs := typeutils.IsKnown(model.SpaceIDs)
	if data.SpaceIds != nil {
		spaceIDs, d := types.SetValueFrom(ctx, types.StringType, *data.SpaceIds)
		diags.Append(d...)
		model.SpaceIDs = spaceIDs
	} else if !originallySetSpaceIDs {
		// Only set to null if it wasn't originally set
		model.SpaceIDs = types.SetNull(types.StringType)
	}
	// If originally set but API didn't return it, keep the original value
	model.populateInputsFromAPI(ctx, pkg, data.Inputs, &diags)

	return diags
}

func (model *integrationPolicyModel) populateInputsFromAPI(ctx context.Context, pkg *kbapi.PackageInfo, inputs map[string]kbapi.PackagePolicyInput, diags *diag.Diagnostics) {
	// Handle input population based on context:
	// 1. If model.Inputs is unknown: we're importing or reading fresh state → populate from API
	// 2. If model.Inputs is known and null/empty: user explicitly didn't configure inputs → don't populate (avoid inconsistent state)
	// 3. If model.Inputs is known and has values: user configured inputs → populate from API

	isInputKnown := typeutils.IsKnown(model.Inputs)
	isInputNullOrEmpty := model.Inputs.IsNull() || (isInputKnown && len(model.Inputs.Elements()) == 0)

	// Case 2: Known and null/empty - user explicitly didn't configure inputs
	if isInputNullOrEmpty && isInputKnown {
		// Don't populate to avoid "Provider produced inconsistent result" error
		model.Inputs = NewInputsNull(getInputsElementType())
		return
	}
	// Case 1 & 3: Unknown (import/fresh read) or known with values - populate from API

	// Fetch package info to get defaults
	inputDefaults, defaultsDiags := packageInfoToDefaults(pkg)
	diags.Append(defaultsDiags...)
	if diags.HasError() {
		return
	}

	if inputDefaults == nil {
		inputDefaults = make(map[string]inputDefaultsModel)
	}

	newInputs := make(map[string]integrationPolicyInputsModel)
	for inputID, inputData := range inputs {
		inputModel := integrationPolicyInputsModel{
			Enabled: types.BoolPointerValue(inputData.Enabled),
			Vars:    typeutils.MapToNormalizedType(schemautil.Deref(inputData.Vars), path.Root("inputs").AtMapKey(inputID).AtName("vars"), diags),
		}

		// Populate streams
		if inputData.Streams != nil && len(*inputData.Streams) > 0 {
			streams := make(map[string]integrationPolicyInputStreamModel)
			for streamID, streamData := range *inputData.Streams {
				streamModel := integrationPolicyInputStreamModel{
					Enabled: types.BoolPointerValue(streamData.Enabled),
					Vars:    typeutils.MapToNormalizedType(schemautil.Deref(streamData.Vars), path.Root("inputs").AtMapKey(inputID).AtName("streams").AtMapKey(streamID).AtName("vars"), diags),
				}

				streams[streamID] = streamModel
			}

			streamsMap, d := types.MapValueFrom(ctx, getInputStreamType(), streams)
			diags.Append(d...)
			inputModel.Streams = streamsMap
		} else {
			inputModel.Streams = types.MapNull(getInputStreamType())
		}

		// Populate defaults if available
		if defaults, ok := inputDefaults[inputID]; ok {
			defaultsObj, d := types.ObjectValueFrom(ctx, getInputDefaultsAttrTypes(), defaults)
			diags.Append(d...)
			inputModel.Defaults = defaultsObj
		} else {
			inputModel.Defaults = types.ObjectNull(getInputDefaultsAttrTypes())
		}

		newInputs[inputID] = inputModel
	}

	inputsMap, d := NewInputsValueFrom(ctx, getInputsElementType(), newInputs)
	diags.Append(d...)
	model.Inputs = inputsMap
}

func (model integrationPolicyModel) toAPIModel(ctx context.Context, feat features) (kbapi.PackagePolicyRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Check if agent_policy_ids is configured and version supports it
	if typeutils.IsKnown(model.AgentPolicyIDs) {
		if !feat.SupportsPolicyIDs {
			return kbapi.PackagePolicyRequest{}, diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("agent_policy_ids"),
					"Unsupported Elasticsearch version",
					fmt.Sprintf("Agent policy IDs are only supported in Elastic Stack %s and above", MinVersionPolicyIDs),
				),
			}
		}
	}

	// Check if output_id is configured and version supports it
	if typeutils.IsKnown(model.OutputID) {
		if !feat.SupportsOutputID {
			return kbapi.PackagePolicyRequest{}, diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("output_id"),
					"Unsupported Elasticsearch version",
					fmt.Sprintf("Output ID is only supported in Elastic Stack %s and above", MinVersionOutputID),
				),
			}
		}
	}

	body := kbapi.PackagePolicyRequest{
		Description: model.Description.ValueStringPointer(),
		Force:       model.Force.ValueBoolPointer(),
		Name:        model.Name.ValueString(),
		Namespace:   model.Namespace.ValueStringPointer(),
		OutputId:    model.OutputID.ValueStringPointer(),
		Package: kbapi.PackagePolicyRequestPackage{
			Name:    model.IntegrationName.ValueString(),
			Version: model.IntegrationVersion.ValueString(),
		},
		PolicyId: model.AgentPolicyID.ValueStringPointer(),
		PolicyIds: func() *[]string {
			if !model.AgentPolicyIDs.IsNull() && !model.AgentPolicyIDs.IsUnknown() {
				var policyIDs []string
				d := model.AgentPolicyIDs.ElementsAs(ctx, &policyIDs, false)
				diags.Append(d...)
				return &policyIDs
			}
			// Only return empty array for 8.15+ when agent_policy_ids is not defined
			if feat.SupportsPolicyIDs {
				emptyArray := []string{}
				return &emptyArray
			}
			return nil
		}(),
		Vars: schemautil.MapRef(typeutils.NormalizedTypeToMap[any](model.VarsJSON.Normalized, path.Root("vars_json"), &diags)),
	}

	if typeutils.IsKnown(model.PolicyID) {
		body.Id = model.PolicyID.ValueStringPointer()
	}

	if typeutils.IsKnown(model.ID) {
		body.Id = model.ID.ValueStringPointer()
	}

	body.Inputs = model.toAPIInputsFromInputsAttribute(ctx, &diags)
	// Note: space_ids is read-only for integration policies and inherited from the agent policy

	return body, diags
}

// toAPIInputsFromInputsAttribute converts the 'inputs' attribute to the API model format
func (model integrationPolicyModel) toAPIInputsFromInputsAttribute(ctx context.Context, diags *diag.Diagnostics) *map[string]kbapi.PackagePolicyRequestInput {
	result := make(map[string]kbapi.PackagePolicyRequestInput, len(model.Inputs.Elements()))
	if !typeutils.IsKnown(model.Inputs.MapValue) {
		return &result
	}

	inputsMap := typeutils.MapTypeAs[integrationPolicyInputsModel](ctx, model.Inputs.MapValue, path.Root("inputs"), diags)
	if inputsMap == nil {
		return &result
	}

	for inputID, inputModel := range inputsMap {
		inputPath := path.Root("inputs").AtMapKey(inputID)

		apiInput := kbapi.PackagePolicyRequestInput{
			Enabled: inputModel.Enabled.ValueBoolPointer(),
			Vars:    schemautil.MapRef(typeutils.NormalizedTypeToMap[any](inputModel.Vars, inputPath.AtName("vars"), diags)),
		}

		// Convert streams if present
		if typeutils.IsKnown(inputModel.Streams) && len(inputModel.Streams.Elements()) > 0 {
			streamsMap := typeutils.MapTypeAs[integrationPolicyInputStreamModel](ctx, inputModel.Streams, inputPath.AtName("streams"), diags)
			if streamsMap != nil {
				streams := make(map[string]kbapi.PackagePolicyRequestInputStream, len(streamsMap))
				for streamID, streamModel := range streamsMap {
					streams[streamID] = kbapi.PackagePolicyRequestInputStream{
						Enabled: streamModel.Enabled.ValueBoolPointer(),
						Vars:    schemautil.MapRef(typeutils.NormalizedTypeToMap[any](streamModel.Vars, inputPath.AtName("streams").AtMapKey(streamID).AtName("vars"), diags)),
					}
				}
				apiInput.Streams = &streams
			}
		}

		result[inputID] = apiInput
	}

	return &result
}
