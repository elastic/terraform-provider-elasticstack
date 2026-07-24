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

package integrationpolicy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/policyshape"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type integrationPolicyModel struct {
	ID                 types.String  `tfsdk:"id"`
	KibanaConnection   types.List    `tfsdk:"kibana_connection"`
	PolicyID           types.String  `tfsdk:"policy_id"`
	Name               types.String  `tfsdk:"name"`
	Namespace          types.String  `tfsdk:"namespace"`
	AgentPolicyID      types.String  `tfsdk:"agent_policy_id"`
	AgentPolicyIDs     types.List    `tfsdk:"agent_policy_ids"`
	Description        types.String  `tfsdk:"description"`
	Force              types.Bool    `tfsdk:"force"`
	IntegrationName    types.String  `tfsdk:"integration_name"`
	IntegrationVersion types.String  `tfsdk:"integration_version"`
	OutputID           types.String  `tfsdk:"output_id"`
	Inputs             InputsValue   `tfsdk:"inputs"` // > integrationPolicyInputsModel
	VarsJSON           VarsJSONValue `tfsdk:"vars_json"`
	SpaceIDs           types.Set     `tfsdk:"space_ids"`
}

// integrationPolicyInputsModel and integrationPolicyInputStreamModel are
// aliases of the shared policyshape.InputModel/InputStreamModel types; see
// policyshape_aliases.go.

func (model *integrationPolicyModel) populateFromAPI(ctx context.Context, pkg *kbapi.KibanaHTTPAPIsGetPackageInfo, data *kbapi.PackagePolicy) diag.Diagnostics {
	if data == nil {
		return nil
	}

	var diags diag.Diagnostics

	dataID := data.Id
	model.ID = types.StringValue(dataID)
	model.PolicyID = types.StringValue(dataID)
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
	model.IntegrationName = types.StringValue(data.Package.Name)
	model.IntegrationVersion = types.StringValue(data.Package.Version)
	model.OutputID = types.StringPointerValue(data.OutputId)

	varsMap := varsAnyToMap(data.Vars)
	if len(varsMap) == 0 {
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
	// Extract mapped inputs from the union Inputs field (simplified format returns mapped inputs).
	// The union field may be empty (nil JSON) when inputs are not present in the response.
	mappedInputs, err := data.Inputs.AsPackagePolicyMappedInputs()
	if err != nil {
		// If the union is empty/nil, treat as no inputs rather than an error
		mappedInputs = kbapi.PackagePolicyMappedInputs{}
	}
	model.populateInputsFromAPI(ctx, pkg, mappedInputs, &diags)

	return diags
}

func (model *integrationPolicyModel) populateInputsFromAPI(ctx context.Context, pkg *kbapi.KibanaHTTPAPIsGetPackageInfo, inputs kbapi.PackagePolicyMappedInputs, diags *diag.Diagnostics) {
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
			Enabled:   types.BoolPointerValue(inputData.Enabled),
			Condition: types.StringPointerValue(inputData.Condition),
			Vars:      typeutils.MarshalToNormalized(typeutils.Deref(inputData.Vars), path.Root("inputs").AtMapKey(inputID).AtName("vars"), diags),
		}

		// Populate streams
		if inputData.Streams != nil && len(*inputData.Streams) > 0 {
			streams := make(map[string]integrationPolicyInputStreamModel)
			for streamID, streamData := range *inputData.Streams {
				streamModel := integrationPolicyInputStreamModel{
					Enabled:   types.BoolPointerValue(streamData.Enabled),
					Condition: types.StringPointerValue(streamData.Condition),
					Vars:      typeutils.MarshalToNormalized(typeutils.Deref(streamData.Vars), path.Root("inputs").AtMapKey(inputID).AtName("streams").AtMapKey(streamID).AtName("vars"), diags),
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

func (model integrationPolicyModel) toAPIModel(ctx context.Context, feat integrationPolicyFeatures) (kbapi.PackagePolicyRequest, diag.Diagnostics) {
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

	// Decode the 'inputs' attribute (including each input's nested 'streams'
	// map) exactly once: validateConditionSupport and
	// toAPIInputsFromInputsAttribute below both need it, and the decode is
	// reflection-based (typeutils.MapTypeAs) over the same structure, so
	// decoding it independently in each would do the same work twice.
	decodedInputs := model.decodeInputs(ctx, &diags)

	// Check if any input/stream condition is configured and version supports it
	if condDiags := model.validateConditionSupport(feat, decodedInputs); condDiags.HasError() {
		diags.Append(condDiags...)
		return kbapi.PackagePolicyRequest{}, diags
	}

	mappedBody := kbapi.PackagePolicyRequestMappedInputs{
		Description: model.Description.ValueStringPointer(),
		Force:       model.Force.ValueBoolPointer(),
		Name:        model.Name.ValueString(),
		Namespace:   model.Namespace.ValueStringPointer(),
		Package: kbapi.PackagePolicyRequestPackage{
			Name:    model.IntegrationName.ValueString(),
			Version: model.IntegrationVersion.ValueString(),
		},
		Vars: func() *map[string]*kbapi.KibanaHTTPAPIsSimplifiedCreatePackagePolicyRequest_Vars_AdditionalProperties {
			// Use SanitizedValue to strip internal metadata like __tf_provider_context
			// before sending to the Kibana API. This prevents HTTP 400 errors when Kibana
			// doesn't recognize the internal __tf_provider_context variable.
			if !typeutils.IsKnown(model.VarsJSON) {
				return nil
			}
			sanitizedVars, sanitizeDiags := model.VarsJSON.SanitizedValue()
			diags.Append(sanitizeDiags...)
			if diags.HasError() {
				return nil
			}
			m := typeutils.NormalizedTypeToMap[any](jsontypes.NewNormalizedValue(sanitizedVars), path.Root("vars_json"), &diags)
			return varsMapToTypedMap[kbapi.KibanaHTTPAPIsSimplifiedCreatePackagePolicyRequest_Vars_AdditionalProperties](m)
		}(),
	}

	if typeutils.IsKnown(model.PolicyID) {
		mappedBody.Id = model.PolicyID.ValueStringPointer()
	}

	if typeutils.IsKnown(model.ID) {
		mappedBody.Id = model.ID.ValueStringPointer()
	}

	mappedBody.Inputs = model.toAPIInputsFromInputsAttribute(decodedInputs, &diags)
	// Note: space_ids is not included in the request body; the Fleet API manages space assignment

	// output_id / policy_id / policy_ids are declared on the simplified create
	// body (Kibana_HTTP_APIs_simplified_create_package_policy_request) again, so
	// set them directly on the typed mapped body and populate the union via its
	// generated accessor instead of a JSON wrapper round-trip.
	mappedBody.OutputId = model.OutputID.ValueStringPointer()
	mappedBody.PolicyId = model.AgentPolicyID.ValueStringPointer()
	mappedBody.PolicyIds = func() *[]string {
		if !model.AgentPolicyIDs.IsNull() && !model.AgentPolicyIDs.IsUnknown() {
			var policyIDs []string
			d := model.AgentPolicyIDs.ElementsAs(ctx, &policyIDs, false)
			diags.Append(d...)
			return &policyIDs
		}
		// 8.15+ accepts an empty array to clear any existing associations.
		if feat.SupportsPolicyIDs {
			emptyArray := []string{}
			return &emptyArray
		}
		return nil
	}()

	var body kbapi.PackagePolicyRequest
	if err := body.FromPackagePolicyRequestMappedInputs(mappedBody); err != nil {
		diags.AddError("Failed to build package policy request", err.Error())
		return kbapi.PackagePolicyRequest{}, diags
	}
	return body, diags
}

// decodedInput is the once-decoded form of a single `inputs` map element,
// with its nested `streams` map (if any) already decoded too. decodeInputs
// produces this so that toAPIModel's two downstream consumers
// (validateConditionSupport, toAPIInputsFromInputsAttribute) don't each
// independently re-run the reflection-based typeutils.MapTypeAs decode over
// the same inputs+streams structure.
type decodedInput = policyshape.DecodedInput[integrationPolicyInputsModel]

// decodeInputs decodes the 'inputs' attribute, and each input's nested
// 'streams' map, exactly once. Returns nil if 'inputs' itself is null/unknown
// or fails to decode.
func (model integrationPolicyModel) decodeInputs(ctx context.Context, diags *diag.Diagnostics) map[string]decodedInput {
	return policyshape.DecodeInputs[integrationPolicyInputsModel](ctx, model.Inputs, path.Root("inputs"), diags)
}

// validateConditionSupport returns an attribute-scoped error diagnostic for
// every input/stream `condition` value that is set when the connected Kibana
// version does not support the `condition` field on package-policy
// inputs/streams (added in Kibana 9.5.0; see MinVersionCondition). It is a
// no-op when the version supports condition or when no inputs are configured.
func (model integrationPolicyModel) validateConditionSupport(feat integrationPolicyFeatures, decoded map[string]decodedInput) diag.Diagnostics {
	var diags diag.Diagnostics

	if feat.SupportsCondition {
		return diags
	}

	for inputID, di := range decoded {
		inputPath := path.Root("inputs").AtMapKey(inputID)

		if typeutils.IsKnown(di.Model.Condition) {
			diags.AddAttributeError(
				inputPath.AtName("condition"),
				"Unsupported Elasticsearch version",
				fmt.Sprintf("Input condition is only supported in Elastic Stack %s and above", MinVersionCondition),
			)
		}

		for streamID, streamModel := range di.Streams {
			if typeutils.IsKnown(streamModel.Condition) {
				diags.AddAttributeError(
					inputPath.AtName("streams").AtMapKey(streamID).AtName("condition"),
					"Unsupported Elasticsearch version",
					fmt.Sprintf("Stream condition is only supported in Elastic Stack %s and above", MinVersionCondition),
				)
			}
		}
	}

	return diags
}

// toAPIInputsFromInputsAttribute converts the already-decoded 'inputs' map
// (see decodeInputs) to the API model format.
func (model integrationPolicyModel) toAPIInputsFromInputsAttribute(decoded map[string]decodedInput, diags *diag.Diagnostics) *map[string]kbapi.PackagePolicyRequestMappedInput {
	result := make(map[string]kbapi.PackagePolicyRequestMappedInput, len(decoded))

	for inputID, di := range decoded {
		inputPath := path.Root("inputs").AtMapKey(inputID)

		apiInput := kbapi.PackagePolicyRequestMappedInput{
			Enabled:   di.Model.Enabled.ValueBoolPointer(),
			Condition: di.Model.Condition.ValueStringPointer(),
			Vars:      varsMapToTypedMap[kbapi.PackagePolicyRequestMappedInput_Vars_AdditionalProperties](typeutils.NormalizedTypeToMap[any](di.Model.Vars, inputPath.AtName("vars"), diags)),
		}

		// Convert streams if present
		if len(di.Streams) > 0 {
			streams := make(map[string]kbapi.PackagePolicyRequestMappedInputStream, len(di.Streams))
			for streamID, streamModel := range di.Streams {
				streamVarsPath := inputPath.AtName("streams").AtMapKey(streamID).AtName("vars")
				streamVars := typeutils.NormalizedTypeToMap[any](streamModel.Vars, streamVarsPath, diags)
				streams[streamID] = kbapi.PackagePolicyRequestMappedInputStream{
					Enabled:   streamModel.Enabled.ValueBoolPointer(),
					Condition: streamModel.Condition.ValueStringPointer(),
					Vars:      varsMapToTypedMap[kbapi.PackagePolicyRequestMappedInputStream_Vars_AdditionalProperties](streamVars),
				}
			}
			apiInput.Streams = &streams
		}

		result[inputID] = apiInput
	}

	return &result
}
