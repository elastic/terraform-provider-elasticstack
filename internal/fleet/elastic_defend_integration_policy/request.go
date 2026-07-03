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

const (
	endpointInputType          = "endpoint"
	bootstrapEndpointInputType = "ENDPOINT_INTEGRATION_CONFIG"
)

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

	if !model.Description.IsNull() && !model.Description.IsUnknown() {
		req.Description = model.Description.ValueStringPointer()
	}

	if !model.Force.IsNull() && !model.Force.IsUnknown() {
		req.Force = model.Force.ValueBoolPointer()
	}

	// Build bootstrap input config: _config.value.endpointConfig.preset
	config := map[string]policyshape.TypedVarEntry{}
	if !model.Preset.IsNull() && !model.Preset.IsUnknown() && model.Preset.ValueString() != "" {
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

	if !model.Description.IsNull() && !model.Description.IsUnknown() {
		req.Description = model.Description.ValueStringPointer()
	}

	if !model.Force.IsNull() && !model.Force.IsUnknown() {
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
	if !model.Preset.IsNull() && !model.Preset.IsUnknown() {
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

func buildWindowsPolicyPayload(ctx context.Context, winObj types.Object) (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics
	if winObj.IsNull() || winObj.IsUnknown() {
		return nil, diags
	}

	var wm windowsPolicyModel
	d := winObj.As(ctx, &wm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	win := map[string]any{}

	if !wm.Events.IsNull() && !wm.Events.IsUnknown() {
		var em windowsEventsModel
		d = wm.Events.As(ctx, &em, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		events := map[string]any{}
		typeutils.SetBoolInMap(events, attrProcess, em.Process)
		typeutils.SetBoolInMap(events, "network", em.Network)
		typeutils.SetBoolInMap(events, "file", em.File)
		typeutils.SetBoolInMap(events, "dll_and_driver_load", em.DllAndDriverLoad)
		typeutils.SetBoolInMap(events, "dns", em.DNS)
		typeutils.SetBoolInMap(events, "registry", em.Registry)
		typeutils.SetBoolInMap(events, "security", em.Security)
		typeutils.SetBoolInMap(events, "authentication", em.Authentication)
		win["events"] = events
	}

	if !wm.Malware.IsNull() && !wm.Malware.IsUnknown() {
		var mm malwareFullModel
		d = wm.Malware.As(ctx, &mm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		malware := map[string]any{}
		typeutils.SetStringInMap(malware, "mode", mm.Mode)
		typeutils.SetBoolInMap(malware, "blocklist", mm.Blocklist)
		typeutils.SetBoolInMap(malware, attrOnWriteScan, mm.OnWriteScan)
		typeutils.SetBoolInMap(malware, attrNotifyUser, mm.NotifyUser)
		win["malware"] = malware
	}

	if !wm.Ransomware.IsNull() && !wm.Ransomware.IsUnknown() {
		var rm protectionModeModel
		d = wm.Ransomware.As(ctx, &rm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		ransomware := map[string]any{}
		typeutils.SetStringInMap(ransomware, "mode", rm.Mode)
		typeutils.SetBoolInMap(ransomware, attrSupported, rm.Supported)
		win[attrRansomware] = ransomware
	}

	if !wm.MemoryProtection.IsNull() && !wm.MemoryProtection.IsUnknown() {
		var mm protectionModeModel
		d = wm.MemoryProtection.As(ctx, &mm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		memProt := map[string]any{}
		typeutils.SetStringInMap(memProt, "mode", mm.Mode)
		typeutils.SetBoolInMap(memProt, attrSupported, mm.Supported)
		win["memory_protection"] = memProt
	}

	if !wm.BehaviorProtection.IsNull() && !wm.BehaviorProtection.IsUnknown() {
		var bm behaviorProtectionModel
		d = wm.BehaviorProtection.As(ctx, &bm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		behProt := map[string]any{}
		typeutils.SetStringInMap(behProt, "mode", bm.Mode)
		typeutils.SetBoolInMap(behProt, attrSupported, bm.Supported)
		typeutils.SetBoolInMap(behProt, attrReputationService, bm.ReputationService)
		win["behavior_protection"] = behProt
	}

	if !wm.Popup.IsNull() && !wm.Popup.IsUnknown() {
		var pm windowsPopupModel
		d = wm.Popup.As(ctx, &pm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		popup := map[string]any{}
		setPopupItem(ctx, popup, "malware", pm.Malware, &diags)
		setPopupItem(ctx, popup, attrRansomware, pm.Ransomware, &diags)
		setPopupItem(ctx, popup, "memory_protection", pm.MemoryProtection, &diags)
		setPopupItem(ctx, popup, "behavior_protection", pm.BehaviorProtection, &diags)
		win[attrPopup] = popup
	}

	if !wm.Logging.IsNull() && !wm.Logging.IsUnknown() {
		var lm loggingModel
		d = wm.Logging.As(ctx, &lm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		logging := map[string]any{}
		typeutils.SetStringInMap(logging, "file", lm.File)
		win["logging"] = logging
	}

	if !wm.AntivirusRegistration.IsNull() && !wm.AntivirusRegistration.IsUnknown() {
		var am antivirusRegistrationModel
		d = wm.AntivirusRegistration.As(ctx, &am, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		avr := map[string]any{}
		typeutils.SetStringInMap(avr, "mode", am.Mode)
		typeutils.SetBoolInMap(avr, "enabled", am.Enabled)
		win["antivirus_registration"] = avr
	}

	if !wm.AttackSurfaceReduction.IsNull() && !wm.AttackSurfaceReduction.IsUnknown() {
		var am attackSurfaceReductionModel
		d = wm.AttackSurfaceReduction.As(ctx, &am, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		asr := map[string]any{}
		if !am.CredentialHardening.IsNull() && !am.CredentialHardening.IsUnknown() {
			var cm credentialHardeningModel
			d = am.CredentialHardening.As(ctx, &cm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
			diags.Append(d...)
			if diags.HasError() {
				return nil, diags
			}
			ch := map[string]any{}
			typeutils.SetBoolInMap(ch, "enabled", cm.Enabled)
			asr["credential_hardening"] = ch
		}
		win["attack_surface_reduction"] = asr
	}

	return win, diags
}

func buildMacPolicyPayload(ctx context.Context, macObj types.Object) (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics
	if macObj.IsNull() || macObj.IsUnknown() {
		return nil, diags
	}

	var mm macPolicyModel
	d := macObj.As(ctx, &mm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	mac := map[string]any{}

	if !mm.Events.IsNull() && !mm.Events.IsUnknown() {
		var em macEventsModel
		d = mm.Events.As(ctx, &em, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		events := map[string]any{}
		typeutils.SetBoolInMap(events, attrProcess, em.Process)
		typeutils.SetBoolInMap(events, "network", em.Network)
		typeutils.SetBoolInMap(events, "file", em.File)
		mac["events"] = events
	}

	if !mm.Malware.IsNull() && !mm.Malware.IsUnknown() {
		var malwareModel malwareFullModel
		d = mm.Malware.As(ctx, &malwareModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		malware := map[string]any{}
		typeutils.SetStringInMap(malware, "mode", malwareModel.Mode)
		typeutils.SetBoolInMap(malware, "blocklist", malwareModel.Blocklist)
		typeutils.SetBoolInMap(malware, attrOnWriteScan, malwareModel.OnWriteScan)
		typeutils.SetBoolInMap(malware, attrNotifyUser, malwareModel.NotifyUser)
		mac["malware"] = malware
	}

	if !mm.MemoryProtection.IsNull() && !mm.MemoryProtection.IsUnknown() {
		var pm protectionModeModel
		d = mm.MemoryProtection.As(ctx, &pm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		memProt := map[string]any{}
		typeutils.SetStringInMap(memProt, "mode", pm.Mode)
		typeutils.SetBoolInMap(memProt, attrSupported, pm.Supported)
		mac["memory_protection"] = memProt
	}

	if !mm.BehaviorProtection.IsNull() && !mm.BehaviorProtection.IsUnknown() {
		var bm behaviorProtectionModel
		d = mm.BehaviorProtection.As(ctx, &bm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		behProt := map[string]any{}
		typeutils.SetStringInMap(behProt, "mode", bm.Mode)
		typeutils.SetBoolInMap(behProt, attrSupported, bm.Supported)
		typeutils.SetBoolInMap(behProt, attrReputationService, bm.ReputationService)
		mac["behavior_protection"] = behProt
	}

	if !mm.Popup.IsNull() && !mm.Popup.IsUnknown() {
		var pm macLinuxPopupModel
		d = mm.Popup.As(ctx, &pm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		popup := map[string]any{}
		setPopupItem(ctx, popup, "malware", pm.Malware, &diags)
		setPopupItem(ctx, popup, "memory_protection", pm.MemoryProtection, &diags)
		setPopupItem(ctx, popup, "behavior_protection", pm.BehaviorProtection, &diags)
		mac[attrPopup] = popup
	}

	if !mm.Logging.IsNull() && !mm.Logging.IsUnknown() {
		var lm loggingModel
		d = mm.Logging.As(ctx, &lm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		logging := map[string]any{}
		typeutils.SetStringInMap(logging, "file", lm.File)
		mac["logging"] = logging
	}

	return mac, diags
}

func buildLinuxPolicyPayload(ctx context.Context, linuxObj types.Object) (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics
	if linuxObj.IsNull() || linuxObj.IsUnknown() {
		return nil, diags
	}

	var lm linuxPolicyModel
	d := linuxObj.As(ctx, &lm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	linux := map[string]any{}

	if !lm.Events.IsNull() && !lm.Events.IsUnknown() {
		var em linuxEventsModel
		d = lm.Events.As(ctx, &em, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		events := map[string]any{}
		typeutils.SetBoolInMap(events, attrProcess, em.Process)
		typeutils.SetBoolInMap(events, "network", em.Network)
		typeutils.SetBoolInMap(events, "file", em.File)
		typeutils.SetBoolInMap(events, "session_data", em.SessionData)
		typeutils.SetBoolInMap(events, "tty_io", em.TtyIO)
		linux["events"] = events
	}

	if !lm.Malware.IsNull() && !lm.Malware.IsUnknown() {
		var mm malwareLinuxModel
		d = lm.Malware.As(ctx, &mm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		malware := map[string]any{}
		typeutils.SetStringInMap(malware, "mode", mm.Mode)
		typeutils.SetBoolInMap(malware, "blocklist", mm.Blocklist)
		linux["malware"] = malware
	}

	if !lm.MemoryProtection.IsNull() && !lm.MemoryProtection.IsUnknown() {
		var pm protectionModeModel
		d = lm.MemoryProtection.As(ctx, &pm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		memProt := map[string]any{}
		typeutils.SetStringInMap(memProt, "mode", pm.Mode)
		typeutils.SetBoolInMap(memProt, attrSupported, pm.Supported)
		linux["memory_protection"] = memProt
	}

	if !lm.BehaviorProtection.IsNull() && !lm.BehaviorProtection.IsUnknown() {
		var bm behaviorProtectionModel
		d = lm.BehaviorProtection.As(ctx, &bm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		behProt := map[string]any{}
		typeutils.SetStringInMap(behProt, "mode", bm.Mode)
		typeutils.SetBoolInMap(behProt, attrSupported, bm.Supported)
		typeutils.SetBoolInMap(behProt, attrReputationService, bm.ReputationService)
		linux["behavior_protection"] = behProt
	}

	if !lm.Popup.IsNull() && !lm.Popup.IsUnknown() {
		var pm macLinuxPopupModel
		d = lm.Popup.As(ctx, &pm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		popup := map[string]any{}
		setPopupItem(ctx, popup, "malware", pm.Malware, &diags)
		setPopupItem(ctx, popup, "memory_protection", pm.MemoryProtection, &diags)
		setPopupItem(ctx, popup, "behavior_protection", pm.BehaviorProtection, &diags)
		linux[attrPopup] = popup
	}

	if !lm.Logging.IsNull() && !lm.Logging.IsUnknown() {
		var logm loggingModel
		d = lm.Logging.As(ctx, &logm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		logging := map[string]any{}
		typeutils.SetStringInMap(logging, "file", logm.File)
		linux["logging"] = logging
	}

	return linux, diags
}

// setAgentPoliciesOnRequest populates PolicyIds / PolicyId on a request from the model.
func setAgentPoliciesOnRequest(ctx context.Context, model *elasticDefendIntegrationPolicyModel, req *kbapi.PackagePolicyRequestTypedInputs) diag.Diagnostics {
	var diags diag.Diagnostics
	if !model.AgentPolicyIDs.IsNull() && !model.AgentPolicyIDs.IsUnknown() {
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

// setPopupItem extracts a popup item from a Terraform object and adds it to the map.
func setPopupItem(ctx context.Context, m map[string]any, key string, obj types.Object, diags *diag.Diagnostics) {
	if obj.IsNull() || obj.IsUnknown() {
		return
	}
	var pm popupItemModel
	d := obj.As(ctx, &pm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
	diags.Append(d...)
	item := map[string]any{}
	typeutils.SetStringInMap(item, "message", pm.Message)
	typeutils.SetBoolInMap(item, "enabled", pm.Enabled)
	m[key] = item
}
