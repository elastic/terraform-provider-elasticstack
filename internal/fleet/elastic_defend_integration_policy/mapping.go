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
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const endpointPackageName = "endpoint"

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

// Helper to extract sub-map from a map.
func getMap(m map[string]any, key string) map[string]any {
	if m == nil {
		return nil
	}
	if v, ok := m[key]; ok {
		if sm, ok := v.(map[string]any); ok {
			return sm
		}
	}
	return nil
}

// mapOptionalObject maps a sub-section from data using getMap; returns
// ObjectValueFrom when present, ObjectNull otherwise.
func mapOptionalObject[M any](ctx context.Context, data map[string]any, key string, attrTypes map[string]attr.Type, build func(map[string]any) M) (types.Object, diag.Diagnostics) {
	sub := getMap(data, key)
	if sub == nil {
		return types.ObjectNull(attrTypes), nil
	}
	return types.ObjectValueFrom(ctx, attrTypes, build(sub))
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

func mapWindowsPolicyFromAPI(ctx context.Context, data map[string]any) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if data == nil {
		return types.ObjectNull(windowsAttrTypes()), diags
	}

	eventsObj, d := mapOptionalObject(ctx, data, "events", windowsEventsAttrTypes(), func(m map[string]any) windowsEventsModel {
		return windowsEventsModel{
			Process:          typeutils.BoolFromMap(m, attrProcess),
			Network:          typeutils.BoolFromMap(m, "network"),
			File:             typeutils.BoolFromMap(m, "file"),
			DllAndDriverLoad: typeutils.BoolFromMap(m, "dll_and_driver_load"),
			DNS:              typeutils.BoolFromMap(m, "dns"),
			Registry:         typeutils.BoolFromMap(m, "registry"),
			Security:         typeutils.BoolFromMap(m, "security"),
			Authentication:   typeutils.BoolFromMap(m, "authentication"),
		}
	})
	diags.Append(d...)

	malwareObj, d := mapOptionalObject(ctx, data, "malware", malwareFullAttrTypes(), func(m map[string]any) malwareFullModel {
		return malwareFullModel{
			Mode:        typeutils.StringFromMap(m, "mode"),
			Blocklist:   typeutils.BoolFromMap(m, "blocklist"),
			OnWriteScan: typeutils.BoolFromMap(m, attrOnWriteScan),
			NotifyUser:  typeutils.BoolFromMap(m, attrNotifyUser),
		}
	})
	diags.Append(d...)

	ransomwareObj, d := mapOptionalObject(ctx, data, attrRansomware, protectionModeAttrTypes(), func(m map[string]any) protectionModeModel {
		return protectionModeModel{
			Mode:      typeutils.StringFromMap(m, "mode"),
			Supported: typeutils.BoolFromMap(m, attrSupported),
		}
	})
	diags.Append(d...)

	common, d := mapCommonPolicyFieldsFromAPI(ctx, data)
	diags.Append(d...)

	popupData := getMap(data, attrPopup)
	popupObj, d := mapWindowsPopupFromAPI(ctx, popupData)
	diags.Append(d...)

	avrObj, d := mapOptionalObject(ctx, data, "antivirus_registration", antivirusRegistrationAttrTypes(), func(m map[string]any) antivirusRegistrationModel {
		return antivirusRegistrationModel{
			Mode:    typeutils.StringFromMap(m, "mode"),
			Enabled: typeutils.BoolFromMap(m, "enabled"),
		}
	})
	diags.Append(d...)

	// attack_surface_reduction contains a nested credential_hardening object,
	// so it requires two levels of mapOptionalObject.
	asrData := getMap(data, "attack_surface_reduction")
	var asrObj types.Object
	if asrData != nil {
		chObj, d := mapOptionalObject(ctx, asrData, "credential_hardening", credentialHardeningAttrTypes(), func(m map[string]any) credentialHardeningModel {
			return credentialHardeningModel{
				Enabled: typeutils.BoolFromMap(m, "enabled"),
			}
		})
		diags.Append(d...)
		asrObj, d = types.ObjectValueFrom(ctx, attackSurfaceReductionAttrTypes(), attackSurfaceReductionModel{
			CredentialHardening: chObj,
		})
		diags.Append(d...)
	} else {
		asrObj = types.ObjectNull(attackSurfaceReductionAttrTypes())
	}

	winObj, d := types.ObjectValueFrom(ctx, windowsAttrTypes(), windowsPolicyModel{
		Events:                 eventsObj,
		Malware:                malwareObj,
		Ransomware:             ransomwareObj,
		MemoryProtection:       common.MemoryProtection,
		BehaviorProtection:     common.BehaviorProtection,
		Popup:                  popupObj,
		Logging:                common.Logging,
		AntivirusRegistration:  avrObj,
		AttackSurfaceReduction: asrObj,
	})
	diags.Append(d...)
	return winObj, diags
}

func mapWindowsPopupFromAPI(ctx context.Context, data map[string]any) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(data) == 0 {
		return types.ObjectNull(windowsPopupAttrTypes()), diags
	}
	malwareData := getMap(data, "malware")
	ransomwareData := getMap(data, attrRansomware)
	memProtData := getMap(data, "memory_protection")
	behProtData := getMap(data, "behavior_protection")

	malwareObj, d := types.ObjectValueFrom(ctx, popupItemAttrTypes(), popupItemModel{
		Message: typeutils.StringFromMap(malwareData, "message"),
		Enabled: typeutils.BoolFromMap(malwareData, "enabled"),
	})
	diags.Append(d...)

	ransomwareObj, d := types.ObjectValueFrom(ctx, popupItemAttrTypes(), popupItemModel{
		Message: typeutils.StringFromMap(ransomwareData, "message"),
		Enabled: typeutils.BoolFromMap(ransomwareData, "enabled"),
	})
	diags.Append(d...)

	memProtObj, d := types.ObjectValueFrom(ctx, popupItemAttrTypes(), popupItemModel{
		Message: typeutils.StringFromMap(memProtData, "message"),
		Enabled: typeutils.BoolFromMap(memProtData, "enabled"),
	})
	diags.Append(d...)

	behProtObj, d := types.ObjectValueFrom(ctx, popupItemAttrTypes(), popupItemModel{
		Message: typeutils.StringFromMap(behProtData, "message"),
		Enabled: typeutils.BoolFromMap(behProtData, "enabled"),
	})
	diags.Append(d...)

	obj, d := types.ObjectValueFrom(ctx, windowsPopupAttrTypes(), windowsPopupModel{
		Malware:            malwareObj,
		Ransomware:         ransomwareObj,
		MemoryProtection:   memProtObj,
		BehaviorProtection: behProtObj,
	})
	diags.Append(d...)
	return obj, diags
}

func mapMacPolicyFromAPI(ctx context.Context, data map[string]any) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if data == nil {
		return types.ObjectNull(macAttrTypes()), diags
	}

	eventsObj, d := mapOptionalObject(ctx, data, "events", macEventsAttrTypes(), func(m map[string]any) macEventsModel {
		return macEventsModel{
			Process: typeutils.BoolFromMap(m, attrProcess),
			Network: typeutils.BoolFromMap(m, "network"),
			File:    typeutils.BoolFromMap(m, "file"),
		}
	})
	diags.Append(d...)

	malwareObj, d := mapOptionalObject(ctx, data, "malware", malwareFullAttrTypes(), func(m map[string]any) malwareFullModel {
		return malwareFullModel{
			Mode:        typeutils.StringFromMap(m, "mode"),
			Blocklist:   typeutils.BoolFromMap(m, "blocklist"),
			OnWriteScan: typeutils.BoolFromMap(m, attrOnWriteScan),
			NotifyUser:  typeutils.BoolFromMap(m, attrNotifyUser),
		}
	})
	diags.Append(d...)

	common, d := mapCommonPolicyFieldsFromAPI(ctx, data)
	diags.Append(d...)

	popupData := getMap(data, attrPopup)
	popupObj, d := mapMacLinuxPopupFromAPI(ctx, popupData)
	diags.Append(d...)

	macObj, d := types.ObjectValueFrom(ctx, macAttrTypes(), macPolicyModel{
		Events:             eventsObj,
		Malware:            malwareObj,
		MemoryProtection:   common.MemoryProtection,
		BehaviorProtection: common.BehaviorProtection,
		Popup:              popupObj,
		Logging:            common.Logging,
	})
	diags.Append(d...)
	return macObj, diags
}

func mapLinuxPolicyFromAPI(ctx context.Context, data map[string]any) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if data == nil {
		return types.ObjectNull(linuxAttrTypes()), diags
	}

	eventsObj, d := mapOptionalObject(ctx, data, "events", linuxEventsAttrTypes(), func(m map[string]any) linuxEventsModel {
		return linuxEventsModel{
			Process:     typeutils.BoolFromMap(m, attrProcess),
			Network:     typeutils.BoolFromMap(m, "network"),
			File:        typeutils.BoolFromMap(m, "file"),
			SessionData: typeutils.BoolFromMap(m, "session_data"),
			TtyIO:       typeutils.BoolFromMap(m, "tty_io"),
		}
	})
	diags.Append(d...)

	malwareObj, d := mapOptionalObject(ctx, data, "malware", malwareLinuxAttrTypes(), func(m map[string]any) malwareLinuxModel {
		return malwareLinuxModel{
			Mode:      typeutils.StringFromMap(m, "mode"),
			Blocklist: typeutils.BoolFromMap(m, "blocklist"),
		}
	})
	diags.Append(d...)

	common, d := mapCommonPolicyFieldsFromAPI(ctx, data)
	diags.Append(d...)

	popupData := getMap(data, attrPopup)
	popupObj, d := mapMacLinuxPopupFromAPI(ctx, popupData)
	diags.Append(d...)

	linuxObj, d := types.ObjectValueFrom(ctx, linuxAttrTypes(), linuxPolicyModel{
		Events:             eventsObj,
		Malware:            malwareObj,
		MemoryProtection:   common.MemoryProtection,
		BehaviorProtection: common.BehaviorProtection,
		Popup:              popupObj,
		Logging:            common.Logging,
	})
	diags.Append(d...)
	return linuxObj, diags
}

func mapMacLinuxPopupFromAPI(ctx context.Context, data map[string]any) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(data) == 0 {
		return types.ObjectNull(macLinuxPopupAttrTypes()), diags
	}
	malwareData := getMap(data, "malware")
	memProtData := getMap(data, "memory_protection")
	behProtData := getMap(data, "behavior_protection")

	malwareObj, d := types.ObjectValueFrom(ctx, popupItemAttrTypes(), popupItemModel{
		Message: typeutils.StringFromMap(malwareData, "message"),
		Enabled: typeutils.BoolFromMap(malwareData, "enabled"),
	})
	diags.Append(d...)

	memProtObj, d := types.ObjectValueFrom(ctx, popupItemAttrTypes(), popupItemModel{
		Message: typeutils.StringFromMap(memProtData, "message"),
		Enabled: typeutils.BoolFromMap(memProtData, "enabled"),
	})
	diags.Append(d...)

	behProtObj, d := types.ObjectValueFrom(ctx, popupItemAttrTypes(), popupItemModel{
		Message: typeutils.StringFromMap(behProtData, "message"),
		Enabled: typeutils.BoolFromMap(behProtData, "enabled"),
	})
	diags.Append(d...)

	obj, d := types.ObjectValueFrom(ctx, macLinuxPopupAttrTypes(), macLinuxPopupModel{
		Malware:            malwareObj,
		MemoryProtection:   memProtObj,
		BehaviorProtection: behProtObj,
	})
	diags.Append(d...)
	return obj, diags
}

// ---- attr types helpers ----

func popupItemAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrMessage: types.StringType,
		attrEnabled: types.BoolType,
	}
}

func windowsEventsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrProcess:           types.BoolType,
		attrNetwork:           types.BoolType,
		attrFile:              types.BoolType,
		"dll_and_driver_load": types.BoolType,
		"dns":                 types.BoolType,
		"registry":            types.BoolType,
		"security":            types.BoolType,
		"authentication":      types.BoolType,
	}
}

func macEventsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrProcess: types.BoolType,
		attrNetwork: types.BoolType,
		attrFile:    types.BoolType,
	}
}

func linuxEventsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrProcess:    types.BoolType,
		attrNetwork:    types.BoolType,
		attrFile:       types.BoolType,
		"session_data": types.BoolType,
		"tty_io":       types.BoolType,
	}
}

func malwareFullAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrMode:        types.StringType,
		attrBlocklist:   types.BoolType,
		attrOnWriteScan: types.BoolType,
		attrNotifyUser:  types.BoolType,
	}
}

func malwareLinuxAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrMode:      types.StringType,
		attrBlocklist: types.BoolType,
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

func antivirusRegistrationAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrMode:    types.StringType,
		attrEnabled: types.BoolType,
	}
}

func credentialHardeningAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrEnabled: types.BoolType,
	}
}

func attackSurfaceReductionAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrCredentialHardening: types.ObjectType{AttrTypes: credentialHardeningAttrTypes()},
	}
}

func windowsPopupAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrMalware:            types.ObjectType{AttrTypes: popupItemAttrTypes()},
		attrRansomware:         types.ObjectType{AttrTypes: popupItemAttrTypes()},
		attrMemoryProtection:   types.ObjectType{AttrTypes: popupItemAttrTypes()},
		attrBehaviorProtection: types.ObjectType{AttrTypes: popupItemAttrTypes()},
	}
}

func macLinuxPopupAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrMalware:            types.ObjectType{AttrTypes: popupItemAttrTypes()},
		attrMemoryProtection:   types.ObjectType{AttrTypes: popupItemAttrTypes()},
		attrBehaviorProtection: types.ObjectType{AttrTypes: popupItemAttrTypes()},
	}
}

func windowsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrEvents:                 types.ObjectType{AttrTypes: windowsEventsAttrTypes()},
		attrMalware:                types.ObjectType{AttrTypes: malwareFullAttrTypes()},
		attrRansomware:             types.ObjectType{AttrTypes: protectionModeAttrTypes()},
		attrMemoryProtection:       types.ObjectType{AttrTypes: protectionModeAttrTypes()},
		attrBehaviorProtection:     types.ObjectType{AttrTypes: behaviorProtectionAttrTypes()},
		attrPopup:                  types.ObjectType{AttrTypes: windowsPopupAttrTypes()},
		attrLogging:                types.ObjectType{AttrTypes: loggingAttrTypes()},
		"antivirus_registration":   types.ObjectType{AttrTypes: antivirusRegistrationAttrTypes()},
		"attack_surface_reduction": types.ObjectType{AttrTypes: attackSurfaceReductionAttrTypes()},
	}
}

func macAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrEvents:             types.ObjectType{AttrTypes: macEventsAttrTypes()},
		attrMalware:            types.ObjectType{AttrTypes: malwareFullAttrTypes()},
		attrMemoryProtection:   types.ObjectType{AttrTypes: protectionModeAttrTypes()},
		attrBehaviorProtection: types.ObjectType{AttrTypes: behaviorProtectionAttrTypes()},
		attrPopup:              types.ObjectType{AttrTypes: macLinuxPopupAttrTypes()},
		attrLogging:            types.ObjectType{AttrTypes: loggingAttrTypes()},
	}
}

func linuxAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"events":              types.ObjectType{AttrTypes: linuxEventsAttrTypes()},
		"malware":             types.ObjectType{AttrTypes: malwareLinuxAttrTypes()},
		"memory_protection":   types.ObjectType{AttrTypes: protectionModeAttrTypes()},
		"behavior_protection": types.ObjectType{AttrTypes: behaviorProtectionAttrTypes()},
		attrPopup:             types.ObjectType{AttrTypes: macLinuxPopupAttrTypes()},
		"logging":             types.ObjectType{AttrTypes: loggingAttrTypes()},
	}
}

func policyAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		policyOSWindows: types.ObjectType{AttrTypes: windowsAttrTypes()},
		policyOSMac:     types.ObjectType{AttrTypes: macAttrTypes()},
		policyOSLinux:   types.ObjectType{AttrTypes: linuxAttrTypes()},
	}
}
