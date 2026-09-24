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

package elasticdefendintegrationpolicy

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/policyshape"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

const endpointPackageName = "endpoint"

const (
	endpointInputType          = "endpoint"
	bootstrapEndpointInputType = "ENDPOINT_INTEGRATION_CONFIG"
)

// populateModelFromAPI maps a PackagePolicy API response into the
// Terraform state model. It validates that the package name is "endpoint" and
// maps all modelled schema fields. Server-managed fields (artifact_manifest,
// version) are NOT written to the public model; callers must persist them
// separately via savePrivateState.
func populateModelFromAPI(ctx context.Context, model *elasticDefendIntegrationPolicyModel, policy *kbapi.PackagePolicy) diag.Diagnostics {
	var diags diag.Diagnostics

	if policy == nil {
		return diags
	}

	// Validate package identity (REQ-005)
	if policy.Package == nil || policy.Package.Name != endpointPackageName {
		pkgName := "<nil>"
		if policy.Package != nil {
			pkgName = policy.Package.Name
		}
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Not an Elastic Defend policy",
				fmt.Sprintf("Package policy %q belongs to package %q, not %q. "+
					"Only Elastic Defend package policies can be managed by elasticstack_fleet_elastic_defend_integration_policy.",
					policy.Id, pkgName, endpointPackageName),
			),
		}
	}

	policyID := policy.Id
	model.PolicyID = types.StringValue(policyID)
	model.Name = types.StringValue(policy.Name)
	model.Namespace = types.StringPointerValue(policy.Namespace)
	// Kibana retains an existing description when the field is omitted from
	// requests. When the user does not configure description (null), keep null
	// regardless of what the API returns — matching the repo pattern that
	// omitted fields are left unmanaged server-side.
	if !model.Description.IsNull() {
		model.Description = types.StringPointerValue(policy.Description)
	}
	model.Enabled = types.BoolValue(policy.Enabled)

	if policy.Package != nil {
		model.IntegrationVersion = types.StringValue(policy.Package.Version)
	}

	originallyUsedAgentPolicyID := typeutils.IsKnown(model.AgentPolicyID)
	originallyUsedAgentPolicyIDs := typeutils.IsKnown(model.AgentPolicyIDs)

	if originallyUsedAgentPolicyID {
		model.AgentPolicyID = types.StringPointerValue(policy.PolicyId)
	}
	if originallyUsedAgentPolicyIDs {
		if policy.PolicyIds != nil {
			agentPolicyIDs, d := types.ListValueFrom(ctx, types.StringType, *policy.PolicyIds)
			diags.Append(d...)
			model.AgentPolicyIDs = agentPolicyIDs
		} else {
			model.AgentPolicyIDs = types.ListNull(types.StringType)
		}
	}
	if !originallyUsedAgentPolicyID && !originallyUsedAgentPolicyIDs {
		// Default: check API response structure and prefer list form when multiple IDs exist
		if policy.PolicyIds != nil && len(*policy.PolicyIds) > 1 {
			agentPolicyIDs, d := types.ListValueFrom(ctx, types.StringType, *policy.PolicyIds)
			diags.Append(d...)
			model.AgentPolicyIDs = agentPolicyIDs
		} else if policy.PolicyId != nil {
			model.AgentPolicyID = types.StringPointerValue(policy.PolicyId)
		}
	}

	// Populate space_ids — only overwrite when the API actually returns them.
	// If the API omits space_ids, preserve the existing model value so
	// space-aware operations (e.g. update, delete) continue to work correctly.
	originallySetSpaceIDs := typeutils.IsKnown(model.SpaceIDs)
	var operationalSpaceID string
	if policy.SpaceIds != nil {
		spaceIDs, d := types.SetValueFrom(ctx, types.StringType, *policy.SpaceIds)
		diags.Append(d...)
		model.SpaceIDs = spaceIDs
		if len(*policy.SpaceIds) > 0 {
			operationalSpaceID = (*policy.SpaceIds)[0]
		}
	} else if !originallySetSpaceIDs {
		model.SpaceIDs = types.SetNull(types.StringType)
	}

	if operationalSpaceID == "" && originallySetSpaceIDs {
		// Preserve existing space — extract it so the composite ID is correct.
		var existingSpaceIDs []string
		d := model.SpaceIDs.ElementsAs(ctx, &existingSpaceIDs, false)
		diags.Append(d...)
		if len(existingSpaceIDs) > 0 {
			operationalSpaceID = existingSpaceIDs[0]
		}
	}

	// Set composite ID: "<space_id>/<policy_id>" when a space is in use.
	if operationalSpaceID != "" {
		model.ID = types.StringValue(operationalSpaceID + "/" + policyID)
	} else {
		model.ID = types.StringValue(policyID)
	}

	// Extract typed inputs from the union Inputs field
	typedInputs, err := policy.Inputs.AsPackagePolicyTypedInputs()
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Failed to parse policy inputs",
				fmt.Sprintf("Could not decode typed inputs from Defend package policy response: %s", err.Error()),
			),
		}
	}

	// Extract preset and policy from the endpoint input config
	var preset string
	var policyData map[string]any

	for _, input := range typedInputs {
		if input.Type == "endpoint" {
			if input.Config != nil {
				// Extract preset from integration_config.value.endpointConfig.preset
				if icEntry, ok := (*input.Config)["integration_config"]; ok {
					if valMap, ok := icEntry.Value.(map[string]any); ok {
						if ec, ok := valMap["endpointConfig"]; ok {
							if ecMap, ok := ec.(map[string]any); ok {
								if p, ok := ecMap[attrPreset]; ok {
									if pStr, ok := p.(string); ok {
										preset = pStr
									}
								}
							}
						}
					}
				}

				// Extract policy data — the Fleet API returns policy wrapped in a
				// {"value": {...}} envelope, consistent with other config keys.
				if pEntry, ok := (*input.Config)["policy"]; ok {
					if valMap, ok := pEntry.Value.(map[string]any); ok {
						policyData = valMap
					}
				}
			}
			break
		}
	}

	if preset != "" {
		model.Preset = types.StringValue(preset)
	} else {
		model.Preset = types.StringNull()
	}

	// Map policy data to the nested policy attribute
	policyObj, d := mapPolicyFromAPI(ctx, policyData)
	diags.Append(d...)
	model.Policy = policyObj

	originallySetAdvancedSettings := typeutils.IsKnown(model.AdvancedSettings)
	if originallySetAdvancedSettings {
		settings := advancedSettingsFromPolicyData(policyData)
		advancedSettings, d := advancedSettingsMapToTerraform(settings)
		diags.Append(d...)
		model.AdvancedSettings = advancedSettings
	}

	return diags
}

// mapPolicyFromAPI converts the raw Defend policy map from the API response
// into the Terraform policy object.
func mapPolicyFromAPI(ctx context.Context, policyData map[string]any) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if policyData == nil {
		// Return null policy object when there's no data, to avoid spurious plan diffs
		return types.ObjectNull(policyAttrTypes()), diags
	}

	var winData, macData, linuxData map[string]any
	if w, ok := policyData[policyOSWindows]; ok {
		if wMap, ok := w.(map[string]any); ok {
			winData = wMap
		}
	}
	if m, ok := policyData[policyOSMac]; ok {
		if mMap, ok := m.(map[string]any); ok {
			macData = mMap
		}
	}
	if l, ok := policyData[policyOSLinux]; ok {
		if lMap, ok := l.(map[string]any); ok {
			linuxData = lMap
		}
	}

	winObj, d := mapWindowsPolicyFromAPI(ctx, winData)
	diags.Append(d...)

	macObj, d := mapMacPolicyFromAPI(ctx, macData)
	diags.Append(d...)

	linuxObj, d := mapLinuxPolicyFromAPI(ctx, linuxData)
	diags.Append(d...)

	policyObj, d := types.ObjectValueFrom(ctx, policyAttrTypes(), policyModel{
		Windows: winObj,
		Mac:     macObj,
		Linux:   linuxObj,
	})
	diags.Append(d...)
	return policyObj, diags
}

// commonPolicyFields holds the sub-sections shared identically by all three OS
// policy types (Windows, Mac, Linux).
type commonPolicyFields struct {
	MemoryProtection   types.Object
	BehaviorProtection types.Object
	Logging            types.Object
}

// mapCommonPolicyFieldsFromAPI extracts memory_protection, behavior_protection,
// and logging — blocks that are byte-for-byte identical across all three OS
// mapping functions.
func mapCommonPolicyFieldsFromAPI(ctx context.Context, data map[string]any) (commonPolicyFields, diag.Diagnostics) {
	var diags diag.Diagnostics

	memProtObj, d := mapOptionalObject(ctx, data, "memory_protection", protectionModeAttrTypes(), func(m map[string]any) protectionModeModel {
		return protectionModeModel{
			Mode:      typeutils.StringFromMap(m, "mode"),
			Supported: typeutils.BoolFromMap(m, attrSupported),
		}
	})
	diags.Append(d...)

	behProtObj, d := mapOptionalObject(ctx, data, "behavior_protection", behaviorProtectionAttrTypes(), func(m map[string]any) behaviorProtectionModel {
		return behaviorProtectionModel{
			Mode:              typeutils.StringFromMap(m, "mode"),
			Supported:         typeutils.BoolFromMap(m, attrSupported),
			ReputationService: typeutils.BoolFromMap(m, attrReputationService),
		}
	})
	diags.Append(d...)

	loggingObj, d := mapOptionalObject(ctx, data, "logging", loggingAttrTypes(), func(m map[string]any) loggingModel {
		return loggingModel{
			File: typeutils.StringFromMap(m, "file"),
		}
	})
	diags.Append(d...)

	return commonPolicyFields{
		MemoryProtection:   memProtObj,
		BehaviorProtection: behProtObj,
		Logging:            loggingObj,
	}, diags
}

// ---- attr types shared by two or more OS-specific policy types ----

func malwareFullAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrMode:        types.StringType,
		attrBlocklist:   types.BoolType,
		attrOnWriteScan: types.BoolType,
		attrNotifyUser:  types.BoolType,
	}
}

func protectionModeAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrMode:      types.StringType,
		attrSupported: types.BoolType,
	}
}

func behaviorProtectionAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrMode:              types.StringType,
		attrSupported:         types.BoolType,
		attrReputationService: types.BoolType,
	}
}

func loggingAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrFile: types.StringType,
	}
}

func policyAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		policyOSWindows: types.ObjectType{AttrTypes: windowsAttrTypes()},
		policyOSMac:     types.ObjectType{AttrTypes: macAttrTypes()},
		policyOSLinux:   types.ObjectType{AttrTypes: linuxAttrTypes()},
	}
}

// buildBootstrapRequest builds the minimal Defend package policy request used
// for the first create step (bootstrap). Kibana expects the create bootstrap to
// use the special ENDPOINT_INTEGRATION_CONFIG input type with preset mapped
// under config._config.value.endpointConfig.preset.
func buildBootstrapRequest(ctx context.Context, model *elasticDefendIntegrationPolicyModel) (kbapi.PackagePolicyRequestTypedInputs, diag.Diagnostics) {
	var diags diag.Diagnostics

	pkg := kbapi.PackagePolicyRequestPackage{
		Name:    endpointPackageName,
		Version: model.IntegrationVersion.ValueString(),
	}
	req := kbapi.PackagePolicyRequestTypedInputs{
		Name:      &[]string{model.Name.ValueString()}[0],
		Namespace: model.Namespace.ValueStringPointer(),
		Package:   &pkg,
		Enabled:   model.Enabled.ValueBoolPointer(),
	}
	d := setAgentPoliciesOnRequest(ctx, model, &req)
	if d.HasError() {
		return req, d
	}

	if typeutils.IsKnown(model.Description) {
		req.Description = model.Description.ValueStringPointer()
	}

	if typeutils.IsKnown(model.Force) {
		req.Force = model.Force.ValueBoolPointer()
	}

	// Build bootstrap input config: _config.value.endpointConfig.preset
	config := map[string]policyshape.TypedVarEntry{}
	if typeutils.IsKnown(model.Preset) && model.Preset.ValueString() != "" {
		config["_config"] = policyshape.TypedVarEntry{Value: map[string]any{
			"type": endpointPackageName,
			"endpointConfig": map[string]any{
				attrPreset: model.Preset.ValueString(),
			},
		}}
	}

	streams := []kbapi.PackagePolicyRequestTypedInputStream{}
	input := kbapi.PackagePolicyRequestTypedInput{
		Type:    bootstrapEndpointInputType,
		Enabled: true,
		Streams: &streams,
	}
	if len(config) > 0 {
		input.Config = &config
	}
	req.Inputs = &[]kbapi.PackagePolicyRequestTypedInput{input}

	return req, diags
}

// buildFinalizeRequest builds the Defend package policy update request used
// after the bootstrap to apply the user-configured policy settings. It uses
// the typed-inputs format with an "endpoint" input and includes the
// server-managed artifact_manifest and version from the private state.
func buildFinalizeRequest(
	ctx context.Context,
	model *elasticDefendIntegrationPolicyModel,
	priorAdvanced map[string]string,
	ps defendPrivateState,
) (kbapi.PackagePolicyRequestTypedInputs, diag.Diagnostics) {
	var diags diag.Diagnostics

	pkg := kbapi.PackagePolicyRequestPackage{
		Name:    endpointPackageName,
		Version: model.IntegrationVersion.ValueString(),
	}
	req := kbapi.PackagePolicyRequestTypedInputs{
		Name:      &[]string{model.Name.ValueString()}[0],
		Namespace: model.Namespace.ValueStringPointer(),
		Package:   &pkg,
		Enabled:   model.Enabled.ValueBoolPointer(),
	}
	d := setAgentPoliciesOnRequest(ctx, model, &req)
	if d.HasError() {
		// Propagate errors from ElementsAs
		return req, d
	}

	if typeutils.IsKnown(model.Description) {
		req.Description = model.Description.ValueStringPointer()
	}

	if typeutils.IsKnown(model.Force) {
		req.Force = model.Force.ValueBoolPointer()
	}

	// Include the version token for optimistic concurrency control
	if ps.Version != "" {
		req.Version = &ps.Version
	}

	// Build the finalize input config
	config, d := buildFinalizeInputConfig(ctx, model, priorAdvanced, ps)
	diags.Append(d...)
	if diags.HasError() {
		return req, diags
	}

	streams := []kbapi.PackagePolicyRequestTypedInputStream{}
	input := kbapi.PackagePolicyRequestTypedInput{
		Type:    endpointInputType,
		Enabled: true,
		Streams: &streams,
	}
	if len(config) > 0 {
		input.Config = &config
	}
	req.Inputs = &[]kbapi.PackagePolicyRequestTypedInput{input}

	return req, diags
}

// buildFinalizeInputConfig builds the typed config map for the finalize/update
// input. It includes integration_config (with preset), artifact_manifest (from
// private state), and the typed policy payload. Each entry wraps its payload in
// the {value, type, frozen} policyshape.TypedVarEntry envelope the Defend API expects.
func buildFinalizeInputConfig(
	ctx context.Context,
	model *elasticDefendIntegrationPolicyModel,
	priorAdvanced map[string]string,
	ps defendPrivateState,
) (map[string]policyshape.TypedVarEntry, diag.Diagnostics) {
	var diags diag.Diagnostics
	config := map[string]policyshape.TypedVarEntry{}

	// integration_config with preset — only include when preset is set
	preset := ""
	if typeutils.IsKnown(model.Preset) {
		preset = model.Preset.ValueString()
	}
	if preset != "" {
		config["integration_config"] = policyshape.TypedVarEntry{Value: map[string]any{
			"endpointConfig": map[string]any{
				attrPreset: preset,
			},
		}}
	}

	// Kibana requires callers to echo back the opaque artifact_manifest on
	// update/finalize requests. Persist it in private state and round-trip it.
	if ps.ArtifactManifest != nil {
		config["artifact_manifest"] = policyshape.TypedVarEntry{Value: ps.ArtifactManifest}
	}

	// Build the typed policy payload from the Terraform model.
	// The Fleet API expects the policy wrapped in a {value: {...}} envelope,
	// consistent with how other config keys like "integration_config" are structured.
	policyData, d := buildPolicyPayload(ctx, model, priorAdvanced)
	diags.Append(d...)
	if policyData != nil {
		config["policy"] = policyshape.TypedVarEntry{Value: policyData}
	}

	return config, diags
}

// buildPolicyPayload converts the Terraform policy model into the Defend API
// policy map structure.
func buildPolicyPayload(ctx context.Context, model *elasticDefendIntegrationPolicyModel, priorAdvanced map[string]string) (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics

	if model.Policy.IsNull() || model.Policy.IsUnknown() {
		return nil, diags
	}

	var pm policyModel
	d := model.Policy.As(ctx, &pm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	policy := map[string]any{}

	winData, d := buildWindowsPolicyPayload(ctx, pm.Windows)
	diags.Append(d...)
	if winData != nil {
		policy[policyOSWindows] = winData
	}

	macData, d := buildMacPolicyPayload(ctx, pm.Mac)
	diags.Append(d...)
	if macData != nil {
		policy[policyOSMac] = macData
	}

	linuxData, d := buildLinuxPolicyPayload(ctx, pm.Linux)
	diags.Append(d...)
	if linuxData != nil {
		policy[policyOSLinux] = linuxData
	}

	settings, d := advancedSettingsMapFromTerraform(ctx, model.AdvancedSettings)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	mergeAdvancedSettingsIntoPolicy(policy, settings, priorAdvanced)

	return policy, diags
}

// setAgentPoliciesOnRequest populates PolicyIds / PolicyId on a request from the model.
func setAgentPoliciesOnRequest(ctx context.Context, model *elasticDefendIntegrationPolicyModel, req *kbapi.PackagePolicyRequestTypedInputs) diag.Diagnostics {
	var diags diag.Diagnostics
	if typeutils.IsKnown(model.AgentPolicyIDs) {
		var ids []string
		d := model.AgentPolicyIDs.ElementsAs(ctx, &ids, false)
		if d.HasError() {
			diags.Append(d...)
			return diags
		}
		req.PolicyIds = &ids
		if len(ids) > 0 {
			req.PolicyId = &ids[0]
		}
	} else {
		req.PolicyId = model.AgentPolicyID.ValueStringPointer()
	}
	return diags
}
