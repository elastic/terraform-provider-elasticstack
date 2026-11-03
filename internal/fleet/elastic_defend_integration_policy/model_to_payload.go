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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/policyshape"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// commonPolicyPayloadFields holds the memory_protection, behavior_protection,
// and logging model objects shared identically by all three OS payload
// builders (buildLinuxPolicyPayload, buildMacPolicyPayload,
// buildWindowsPolicyPayload).
type commonPolicyPayloadFields struct {
	MemoryProtection   types.Object
	BehaviorProtection types.Object
	Logging            types.Object
}

// buildCommonPolicyPayloadFields decodes memory_protection,
// behavior_protection, and logging — blocks that are byte-for-byte identical
// across all three OS build functions — and writes the resulting maps into
// target under their respective attribute keys when present.
func buildCommonPolicyPayloadFields(ctx context.Context, target map[string]any, fields commonPolicyPayloadFields, diags *diag.Diagnostics) {
	if pm, ok := decodeObjectField[memoryProtectionModel](ctx, fields.MemoryProtection, diags); ok {
		memProt := map[string]any{}
		typeutils.SetStringInMap(memProt, attrMode, pm.Mode)
		typeutils.SetBoolInMap(memProt, attrSupported, pm.Supported)
		typeutils.SetBoolInMap(memProt, attrCustomYara, pm.CustomYaraSignatures)
		target[attrMemoryProtection] = memProt
	}

	if bm, ok := decodeObjectField[behaviorProtectionModel](ctx, fields.BehaviorProtection, diags); ok {
		behProt := map[string]any{}
		typeutils.SetStringInMap(behProt, attrMode, bm.Mode)
		typeutils.SetBoolInMap(behProt, attrSupported, bm.Supported)
		typeutils.SetBoolInMap(behProt, attrReputationService, bm.ReputationService)
		target[attrBehaviorProtection] = behProt
	}

	if logm, ok := decodeObjectField[loggingModel](ctx, fields.Logging, diags); ok {
		logging := map[string]any{}
		typeutils.SetStringInMap(logging, attrFile, logm.File)
		target[attrLogging] = logging
	}
}

// buildMalwareAndRansomwarePayloadFields decodes malware and ransomware —
// blocks that are byte-for-byte identical across buildMacPolicyPayload and
// buildWindowsPolicyPayload — and writes the resulting maps into target under
// their respective attribute keys when present.
func buildMalwareAndRansomwarePayloadFields(ctx context.Context, target map[string]any, fields macWindowsMalwareFields, diags *diag.Diagnostics) {
	if mm, ok := decodeObjectField[malwareFullModel](ctx, fields.Malware, diags); ok {
		malware := map[string]any{}
		typeutils.SetStringInMap(malware, attrMode, mm.Mode)
		typeutils.SetBoolInMap(malware, attrBlocklist, mm.Blocklist)
		typeutils.SetBoolInMap(malware, attrOnWriteScan, mm.OnWriteScan)
		typeutils.SetBoolInMap(malware, attrNotifyUser, mm.NotifyUser)
		target[attrMalware] = malware
	}

	if rm, ok := decodeObjectField[protectionModeModel](ctx, fields.Ransomware, diags); ok {
		ransomware := map[string]any{}
		typeutils.SetStringInMap(ransomware, attrMode, rm.Mode)
		typeutils.SetBoolInMap(ransomware, attrSupported, rm.Supported)
		target[attrRansomware] = ransomware
	}
}

// buildDeviceControlPayloadField decodes device_control — a block that is
// byte-for-byte identical across buildMacPolicyPayload and
// buildWindowsPolicyPayload — and writes it into target when present.
func buildDeviceControlPayloadField(ctx context.Context, target map[string]any, deviceControl types.Object, diags *diag.Diagnostics) {
	if dm, ok := decodeObjectField[deviceControlModel](ctx, deviceControl, diags); ok {
		dc := map[string]any{}
		typeutils.SetBoolInMap(dc, attrEnabled, dm.Enabled)
		typeutils.SetStringInMap(dc, attrUsbStorage, dm.UsbStorage)
		target[attrDeviceControl] = dc
	}
}

// basePackagePolicyFields holds the request fields that are computed
// identically for both the bootstrap (create) and finalize (update) package
// policy requests: agent policy id(s), additional datastream permissions,
// description, and force.
type basePackagePolicyFields struct {
	PolicyID                         *string
	PolicyIDs                        *[]string
	AdditionalDatastreamsPermissions *[]string
	Description                      *string
	Force                            *bool
}

// buildBasePackagePolicyFields computes the request fields shared by
// buildBootstrapRequest and buildFinalizeRequest, so the two request builders
// don't duplicate the field-population logic for fields they have in common.
func buildBasePackagePolicyFields(ctx context.Context, model *elasticDefendIntegrationPolicyModel) (basePackagePolicyFields, diag.Diagnostics) {
	var diags diag.Diagnostics
	var fields basePackagePolicyFields

	policyID, policyIDs, d := agentPoliciesFromModel(ctx, model)
	diags.Append(d...)
	if diags.HasError() {
		return fields, diags
	}
	fields.PolicyID, fields.PolicyIDs = policyID, policyIDs

	perms, permsDiags := additionalDatastreamsPermissionsFromModel(ctx, model)
	diags.Append(permsDiags...)
	if diags.HasError() {
		return fields, diags
	}
	fields.AdditionalDatastreamsPermissions = perms

	if typeutils.IsKnown(model.Description) {
		fields.Description = model.Description.ValueStringPointer()
	}

	if typeutils.IsKnown(model.Force) {
		fields.Force = model.Force.ValueBoolPointer()
	}

	return fields, diags
}

// buildBootstrapRequest builds the minimal Defend package policy request used
// for the first create step (bootstrap). Kibana expects the create bootstrap to
// use the special ENDPOINT_INTEGRATION_CONFIG input type with preset mapped
// under config._config.value.endpointConfig.preset.
func buildBootstrapRequest(ctx context.Context, model *elasticDefendIntegrationPolicyModel) (kbapi.PackagePolicyCreateRequestTypedInputs, diag.Diagnostics) {
	var diags diag.Diagnostics

	pkg := kbapi.PackagePolicyRequestPackage{
		Name:    endpointPackageName,
		Version: model.IntegrationVersion.ValueString(),
	}
	req := kbapi.PackagePolicyCreateRequestTypedInputs{
		Id:        typeutils.OptionalString(model.PolicyID),
		Name:      model.Name.ValueString(),
		Namespace: model.Namespace.ValueStringPointer(),
		Package:   &pkg,
		Enabled:   model.Enabled.ValueBoolPointer(),
	}

	base, d := buildBasePackagePolicyFields(ctx, model)
	diags.Append(d...)
	if diags.HasError() {
		return req, diags
	}
	req.PolicyId, req.PolicyIds = base.PolicyID, base.PolicyIDs
	req.AdditionalDatastreamsPermissions = base.AdditionalDatastreamsPermissions
	req.Description = base.Description
	req.Force = base.Force

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
	req.Inputs = []kbapi.PackagePolicyRequestTypedInput{input}

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
) (kbapi.PackagePolicyUpdateRequestTypedInputs, diag.Diagnostics) {
	var diags diag.Diagnostics

	pkg := kbapi.PackagePolicyRequestPackage{
		Name:    endpointPackageName,
		Version: model.IntegrationVersion.ValueString(),
	}
	name := model.Name.ValueString()
	req := kbapi.PackagePolicyUpdateRequestTypedInputs{
		Name:      &name,
		Namespace: model.Namespace.ValueStringPointer(),
		Package:   &pkg,
		Enabled:   model.Enabled.ValueBoolPointer(),
	}

	base, d := buildBasePackagePolicyFields(ctx, model)
	diags.Append(d...)
	if diags.HasError() {
		// Propagate errors from ElementsAs
		return req, diags
	}
	req.PolicyId, req.PolicyIds = base.PolicyID, base.PolicyIDs
	req.AdditionalDatastreamsPermissions = base.AdditionalDatastreamsPermissions
	req.Description = base.Description
	req.Force = base.Force

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

// agentPoliciesFromModel returns the PolicyId / PolicyIds request values for the model.
func agentPoliciesFromModel(ctx context.Context, model *elasticDefendIntegrationPolicyModel) (*string, *[]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if typeutils.IsKnown(model.AgentPolicyIDs) {
		var ids []string
		d := model.AgentPolicyIDs.ElementsAs(ctx, &ids, false)
		if d.HasError() {
			diags.Append(d...)
			return nil, nil, diags
		}
		var policyID *string
		if len(ids) > 0 {
			policyID = &ids[0]
		}
		return policyID, &ids, diags
	}
	return model.AgentPolicyID.ValueStringPointer(), nil, diags
}

// additionalDatastreamsPermissionsFromModel returns the request value for
// additional_datastreams_permissions: a known list (including an empty one) is
// sent, while null or unknown is omitted so the server value is retained.
func additionalDatastreamsPermissionsFromModel(ctx context.Context, model *elasticDefendIntegrationPolicyModel) (*[]string, diag.Diagnostics) {
	if !typeutils.IsKnown(model.AdditionalDatastreamsPermissions) {
		return nil, nil
	}
	perms := []string{}
	diags := model.AdditionalDatastreamsPermissions.ElementsAs(ctx, &perms, false)
	if diags.HasError() {
		return nil, diags
	}
	return &perms, diags
}
