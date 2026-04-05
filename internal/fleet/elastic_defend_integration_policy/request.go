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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

const (
	bootstrapInputType  = "ENDPOINT_INTEGRATION_CONFIG"
	finalizeInputType   = "endpoint"
)

// buildBootstrapRequest builds the minimal Defend package policy request used
// for the first create step (bootstrap). It uses the ENDPOINT_INTEGRATION_CONFIG
// input type with preset mapped under config._config.value.endpointConfig.preset.
func buildBootstrapRequest(model *elasticDefendIntegrationPolicyModel) kbapi.DefendPackagePolicyRequest {
	req := kbapi.DefendPackagePolicyRequest{
		Name:      model.Name.ValueString(),
		Namespace: model.Namespace.ValueStringPointer(),
		Package: kbapi.PackagePolicyRequestPackage{
			Name:    endpointPackageName,
			Version: model.IntegrationVersion.ValueString(),
		},
		PolicyId: model.AgentPolicyID.ValueStringPointer(),
		Enabled:  model.Enabled.ValueBoolPointer(),
	}

	if !model.Description.IsNull() && !model.Description.IsUnknown() {
		req.Description = model.Description.ValueStringPointer()
	}

	if !model.Force.IsNull() && !model.Force.IsUnknown() {
		req.Force = model.Force.ValueBoolPointer()
	}

	// Build bootstrap input config: _config.value.endpointConfig.preset
	inputConfig := map[string]any{}
	if !model.Preset.IsNull() && !model.Preset.IsUnknown() && model.Preset.ValueString() != "" {
		inputConfig["_config"] = map[string]any{
			"value": map[string]any{
				"endpointConfig": map[string]any{
					"preset": model.Preset.ValueString(),
				},
			},
		}
	}

	req.Inputs = []kbapi.DefendPackagePolicyRequestInput{
		{
			Type:    bootstrapInputType,
			Enabled: true,
			Streams: []any{},
			Config:  inputConfig,
		},
	}

	return req
}

// buildFinalizeRequest builds the Defend package policy update request used
// after the bootstrap to apply the user-configured policy settings. It uses
// the "endpoint" input type and includes the server-managed artifact_manifest
// and version from the private state.
func buildFinalizeRequest(ctx context.Context, model *elasticDefendIntegrationPolicyModel, ps defendPrivateState) (kbapi.DefendPackagePolicyRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	req := kbapi.DefendPackagePolicyRequest{
		Name:      model.Name.ValueString(),
		Namespace: model.Namespace.ValueStringPointer(),
		Package: kbapi.PackagePolicyRequestPackage{
			Name:    endpointPackageName,
			Version: model.IntegrationVersion.ValueString(),
		},
		PolicyId: model.AgentPolicyID.ValueStringPointer(),
		Enabled:  model.Enabled.ValueBoolPointer(),
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
	inputConfig, d := buildFinalizeInputConfig(ctx, model, ps)
	diags.Append(d...)
	if diags.HasError() {
		return req, diags
	}

	req.Inputs = []kbapi.DefendPackagePolicyRequestInput{
		{
			Type:    finalizeInputType,
			Enabled: true,
			Streams: []any{},
			Config:  inputConfig,
		},
	}

	return req, diags
}

// buildFinalizeInputConfig builds the config map for the finalize/update input.
// It includes integration_config (with preset), artifact_manifest (from private
// state), and the typed policy payload.
func buildFinalizeInputConfig(ctx context.Context, model *elasticDefendIntegrationPolicyModel, ps defendPrivateState) (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics
	config := map[string]any{}

	// integration_config with preset — only include when preset is set
	preset := ""
	if !model.Preset.IsNull() && !model.Preset.IsUnknown() {
		preset = model.Preset.ValueString()
	}
	if preset != "" {
		config["integration_config"] = map[string]any{
			"value": map[string]any{
				"endpointConfig": map[string]any{
					"preset": preset,
				},
			},
		}
	}

	// Preserve artifact_manifest from private state
	if ps.ArtifactManifest != nil {
		config["artifact_manifest"] = ps.ArtifactManifest
	}

	// Build the typed policy payload from the Terraform model
	policyData, d := buildPolicyPayload(ctx, model)
	diags.Append(d...)
	if policyData != nil {
		config["policy"] = policyData
	}

	return config, diags
}

// buildPolicyPayload converts the Terraform policy model into the Defend API
// policy map structure.
func buildPolicyPayload(ctx context.Context, model *elasticDefendIntegrationPolicyModel) (map[string]any, diag.Diagnostics) {
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
		policy["windows"] = winData
	}

	macData, d := buildMacPolicyPayload(ctx, pm.Mac)
	diags.Append(d...)
	if macData != nil {
		policy["mac"] = macData
	}

	linuxData, d := buildLinuxPolicyPayload(ctx, pm.Linux)
	diags.Append(d...)
	if linuxData != nil {
		policy["linux"] = linuxData
	}

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
		setBoolField(events, "process", em.Process)
		setBoolField(events, "network", em.Network)
		setBoolField(events, "file", em.File)
		setBoolField(events, "dll_and_driver_load", em.DllAndDriverLoad)
		setBoolField(events, "dns", em.DNS)
		setBoolField(events, "registry", em.Registry)
		setBoolField(events, "security", em.Security)
		setBoolField(events, "authentication", em.Authentication)
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
		setStringField(malware, "mode", mm.Mode)
		setBoolField(malware, "blocklist", mm.Blocklist)
		setBoolField(malware, "on_write_scan", mm.OnWriteScan)
		setBoolField(malware, "notify_user", mm.NotifyUser)
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
		setStringField(ransomware, "mode", rm.Mode)
		setBoolField(ransomware, "supported", rm.Supported)
		win["ransomware"] = ransomware
	}

	if !wm.MemoryProtection.IsNull() && !wm.MemoryProtection.IsUnknown() {
		var mm protectionModeModel
		d = wm.MemoryProtection.As(ctx, &mm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		memProt := map[string]any{}
		setStringField(memProt, "mode", mm.Mode)
		setBoolField(memProt, "supported", mm.Supported)
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
		setStringField(behProt, "mode", bm.Mode)
		setBoolField(behProt, "supported", bm.Supported)
		setBoolField(behProt, "reputation_service", bm.ReputationService)
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
		setPopupItem(ctx, popup, "ransomware", pm.Ransomware, &diags)
		setPopupItem(ctx, popup, "memory_protection", pm.MemoryProtection, &diags)
		setPopupItem(ctx, popup, "behavior_protection", pm.BehaviorProtection, &diags)
		win["popup"] = popup
	}

	if !wm.Logging.IsNull() && !wm.Logging.IsUnknown() {
		var lm loggingModel
		d = wm.Logging.As(ctx, &lm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		logging := map[string]any{}
		setStringField(logging, "file", lm.File)
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
		setBoolField(avr, "enabled", am.Enabled)
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
			setBoolField(ch, "enabled", cm.Enabled)
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
		setBoolField(events, "process", em.Process)
		setBoolField(events, "network", em.Network)
		setBoolField(events, "file", em.File)
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
		setStringField(malware, "mode", malwareModel.Mode)
		setBoolField(malware, "blocklist", malwareModel.Blocklist)
		setBoolField(malware, "on_write_scan", malwareModel.OnWriteScan)
		setBoolField(malware, "notify_user", malwareModel.NotifyUser)
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
		setStringField(memProt, "mode", pm.Mode)
		setBoolField(memProt, "supported", pm.Supported)
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
		setStringField(behProt, "mode", bm.Mode)
		setBoolField(behProt, "supported", bm.Supported)
		setBoolField(behProt, "reputation_service", bm.ReputationService)
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
		mac["popup"] = popup
	}

	if !mm.Logging.IsNull() && !mm.Logging.IsUnknown() {
		var lm loggingModel
		d = mm.Logging.As(ctx, &lm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		logging := map[string]any{}
		setStringField(logging, "file", lm.File)
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
		setBoolField(events, "process", em.Process)
		setBoolField(events, "network", em.Network)
		setBoolField(events, "file", em.File)
		setBoolField(events, "session_data", em.SessionData)
		setBoolField(events, "tty_io", em.TtyIO)
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
		setStringField(malware, "mode", mm.Mode)
		setBoolField(malware, "blocklist", mm.Blocklist)
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
		setStringField(memProt, "mode", pm.Mode)
		setBoolField(memProt, "supported", pm.Supported)
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
		setStringField(behProt, "mode", bm.Mode)
		setBoolField(behProt, "supported", bm.Supported)
		setBoolField(behProt, "reputation_service", bm.ReputationService)
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
		linux["popup"] = popup
	}

	if !lm.Logging.IsNull() && !lm.Logging.IsUnknown() {
		var logm loggingModel
		d = lm.Logging.As(ctx, &logm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		logging := map[string]any{}
		setStringField(logging, "file", logm.File)
		linux["logging"] = logging
	}

	return linux, diags
}

// setBoolField sets a bool field in the map if the value is known and non-null.
func setBoolField(m map[string]any, key string, val types.Bool) {
	if !val.IsNull() && !val.IsUnknown() {
		m[key] = val.ValueBool()
	}
}

// setStringField sets a string field in the map if the value is known and non-null.
func setStringField(m map[string]any, key string, val types.String) {
	if !val.IsNull() && !val.IsUnknown() {
		m[key] = val.ValueString()
	}
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
	setStringField(item, "message", pm.Message)
	setBoolField(item, "enabled", pm.Enabled)
	m[key] = item
}
