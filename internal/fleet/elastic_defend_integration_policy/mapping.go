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
					typeutils.Deref(policy.Id), pkgName, endpointPackageName),
			),
		}
	}

	policyID := typeutils.Deref(policy.Id)
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
	if w, ok := policyData["windows"]; ok {
		if wMap, ok := w.(map[string]any); ok {
			winData = wMap
		}
	}
	if m, ok := policyData["mac"]; ok {
		if mMap, ok := m.(map[string]any); ok {
			macData = mMap
		}
	}
	if l, ok := policyData["linux"]; ok {
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

func mapWindowsPolicyFromAPI(ctx context.Context, data map[string]any) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if data == nil {
		return types.ObjectNull(windowsAttrTypes()), diags
	}

	eventsData := getMap(data, "events")
	var eventsObj types.Object
	if eventsData != nil {
		var d diag.Diagnostics
		eventsObj, d = types.ObjectValueFrom(ctx, windowsEventsAttrTypes(), windowsEventsModel{
			Process:          typeutils.BoolFromMap(eventsData, attrProcess),
			Network:          typeutils.BoolFromMap(eventsData, "network"),
			File:             typeutils.BoolFromMap(eventsData, "file"),
			DllAndDriverLoad: typeutils.BoolFromMap(eventsData, "dll_and_driver_load"),
			DNS:              typeutils.BoolFromMap(eventsData, "dns"),
			Registry:         typeutils.BoolFromMap(eventsData, "registry"),
			Security:         typeutils.BoolFromMap(eventsData, "security"),
			Authentication:   typeutils.BoolFromMap(eventsData, "authentication"),
		})
		diags.Append(d...)
	} else {
		eventsObj = types.ObjectNull(windowsEventsAttrTypes())
	}

	malwareData := getMap(data, "malware")
	var malwareObj types.Object
	if malwareData != nil {
		var d diag.Diagnostics
		malwareObj, d = types.ObjectValueFrom(ctx, malwareFullAttrTypes(), malwareFullModel{
			Mode:        typeutils.StringFromMap(malwareData, "mode"),
			Blocklist:   typeutils.BoolFromMap(malwareData, "blocklist"),
			OnWriteScan: typeutils.BoolFromMap(malwareData, attrOnWriteScan),
			NotifyUser:  typeutils.BoolFromMap(malwareData, attrNotifyUser),
		})
		diags.Append(d...)
	} else {
		malwareObj = types.ObjectNull(malwareFullAttrTypes())
	}

	ransomwareData := getMap(data, attrRansomware)
	var ransomwareObj types.Object
	if ransomwareData != nil {
		var d diag.Diagnostics
		ransomwareObj, d = types.ObjectValueFrom(ctx, protectionModeAttrTypes(), protectionModeModel{
			Mode:      typeutils.StringFromMap(ransomwareData, "mode"),
			Supported: typeutils.BoolFromMap(ransomwareData, attrSupported),
		})
		diags.Append(d...)
	} else {
		ransomwareObj = types.ObjectNull(protectionModeAttrTypes())
	}

	memProtData := getMap(data, "memory_protection")
	var memProtObj types.Object
	if memProtData != nil {
		var d diag.Diagnostics
		memProtObj, d = types.ObjectValueFrom(ctx, protectionModeAttrTypes(), protectionModeModel{
			Mode:      typeutils.StringFromMap(memProtData, "mode"),
			Supported: typeutils.BoolFromMap(memProtData, attrSupported),
		})
		diags.Append(d...)
	} else {
		memProtObj = types.ObjectNull(protectionModeAttrTypes())
	}

	behProtData := getMap(data, "behavior_protection")
	var behProtObj types.Object
	if behProtData != nil {
		var d diag.Diagnostics
		behProtObj, d = types.ObjectValueFrom(ctx, behaviorProtectionAttrTypes(), behaviorProtectionModel{
			Mode:              typeutils.StringFromMap(behProtData, "mode"),
			Supported:         typeutils.BoolFromMap(behProtData, attrSupported),
			ReputationService: typeutils.BoolFromMap(behProtData, attrReputationService),
		})
		diags.Append(d...)
	} else {
		behProtObj = types.ObjectNull(behaviorProtectionAttrTypes())
	}

	popupData := getMap(data, attrPopup)
	popupObj, d := mapWindowsPopupFromAPI(ctx, popupData)
	diags.Append(d...)

	loggingData := getMap(data, "logging")
	var loggingObj types.Object
	if loggingData != nil {
		var d diag.Diagnostics
		loggingObj, d = types.ObjectValueFrom(ctx, loggingAttrTypes(), loggingModel{
			File: typeutils.StringFromMap(loggingData, "file"),
		})
		diags.Append(d...)
	} else {
		loggingObj = types.ObjectNull(loggingAttrTypes())
	}

	avrData := getMap(data, "antivirus_registration")
	var avrObj types.Object
	if avrData != nil {
		var d diag.Diagnostics
		avrObj, d = types.ObjectValueFrom(ctx, antivirusRegistrationAttrTypes(), antivirusRegistrationModel{
			Mode:    typeutils.StringFromMap(avrData, "mode"),
			Enabled: typeutils.BoolFromMap(avrData, "enabled"),
		})
		diags.Append(d...)
	} else {
		avrObj = types.ObjectNull(antivirusRegistrationAttrTypes())
	}

	asrData := getMap(data, "attack_surface_reduction")
	var asrObj types.Object
	if asrData != nil {
		chData := getMap(asrData, "credential_hardening")
		var chObj types.Object
		if chData != nil {
			var d diag.Diagnostics
			chObj, d = types.ObjectValueFrom(ctx, credentialHardeningAttrTypes(), credentialHardeningModel{
				Enabled: typeutils.BoolFromMap(chData, "enabled"),
			})
			diags.Append(d...)
		} else {
			chObj = types.ObjectNull(credentialHardeningAttrTypes())
		}
		var d diag.Diagnostics
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
		MemoryProtection:       memProtObj,
		BehaviorProtection:     behProtObj,
		Popup:                  popupObj,
		Logging:                loggingObj,
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

	eventsData := getMap(data, "events")
	var eventsObj types.Object
	if eventsData != nil {
		var d diag.Diagnostics
		eventsObj, d = types.ObjectValueFrom(ctx, macEventsAttrTypes(), macEventsModel{
			Process: typeutils.BoolFromMap(eventsData, attrProcess),
			Network: typeutils.BoolFromMap(eventsData, "network"),
			File:    typeutils.BoolFromMap(eventsData, "file"),
		})
		diags.Append(d...)
	} else {
		eventsObj = types.ObjectNull(macEventsAttrTypes())
	}

	malwareData := getMap(data, "malware")
	var malwareObj types.Object
	if malwareData != nil {
		var d diag.Diagnostics
		malwareObj, d = types.ObjectValueFrom(ctx, malwareFullAttrTypes(), malwareFullModel{
			Mode:        typeutils.StringFromMap(malwareData, "mode"),
			Blocklist:   typeutils.BoolFromMap(malwareData, "blocklist"),
			OnWriteScan: typeutils.BoolFromMap(malwareData, attrOnWriteScan),
			NotifyUser:  typeutils.BoolFromMap(malwareData, attrNotifyUser),
		})
		diags.Append(d...)
	} else {
		malwareObj = types.ObjectNull(malwareFullAttrTypes())
	}

	memProtData := getMap(data, "memory_protection")
	var memProtObj types.Object
	if memProtData != nil {
		var d diag.Diagnostics
		memProtObj, d = types.ObjectValueFrom(ctx, protectionModeAttrTypes(), protectionModeModel{
			Mode:      typeutils.StringFromMap(memProtData, "mode"),
			Supported: typeutils.BoolFromMap(memProtData, attrSupported),
		})
		diags.Append(d...)
	} else {
		memProtObj = types.ObjectNull(protectionModeAttrTypes())
	}

	behProtData := getMap(data, "behavior_protection")
	var behProtObj types.Object
	if behProtData != nil {
		var d diag.Diagnostics
		behProtObj, d = types.ObjectValueFrom(ctx, behaviorProtectionAttrTypes(), behaviorProtectionModel{
			Mode:              typeutils.StringFromMap(behProtData, "mode"),
			Supported:         typeutils.BoolFromMap(behProtData, attrSupported),
			ReputationService: typeutils.BoolFromMap(behProtData, attrReputationService),
		})
		diags.Append(d...)
	} else {
		behProtObj = types.ObjectNull(behaviorProtectionAttrTypes())
	}

	popupData := getMap(data, attrPopup)
	popupObj, d := mapMacLinuxPopupFromAPI(ctx, popupData)
	diags.Append(d...)

	loggingData := getMap(data, "logging")
	var loggingObj types.Object
	if loggingData != nil {
		var d diag.Diagnostics
		loggingObj, d = types.ObjectValueFrom(ctx, loggingAttrTypes(), loggingModel{
			File: typeutils.StringFromMap(loggingData, "file"),
		})
		diags.Append(d...)
	} else {
		loggingObj = types.ObjectNull(loggingAttrTypes())
	}

	macObj, d := types.ObjectValueFrom(ctx, macAttrTypes(), macPolicyModel{
		Events:             eventsObj,
		Malware:            malwareObj,
		MemoryProtection:   memProtObj,
		BehaviorProtection: behProtObj,
		Popup:              popupObj,
		Logging:            loggingObj,
	})
	diags.Append(d...)
	return macObj, diags
}

func mapLinuxPolicyFromAPI(ctx context.Context, data map[string]any) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if data == nil {
		return types.ObjectNull(linuxAttrTypes()), diags
	}

	eventsData := getMap(data, "events")
	var eventsObj types.Object
	if eventsData != nil {
		var d diag.Diagnostics
		eventsObj, d = types.ObjectValueFrom(ctx, linuxEventsAttrTypes(), linuxEventsModel{
			Process:     typeutils.BoolFromMap(eventsData, attrProcess),
			Network:     typeutils.BoolFromMap(eventsData, "network"),
			File:        typeutils.BoolFromMap(eventsData, "file"),
			SessionData: typeutils.BoolFromMap(eventsData, "session_data"),
			TtyIO:       typeutils.BoolFromMap(eventsData, "tty_io"),
		})
		diags.Append(d...)
	} else {
		eventsObj = types.ObjectNull(linuxEventsAttrTypes())
	}

	malwareData := getMap(data, "malware")
	var malwareObj types.Object
	if malwareData != nil {
		var d diag.Diagnostics
		malwareObj, d = types.ObjectValueFrom(ctx, malwareLinuxAttrTypes(), malwareLinuxModel{
			Mode:      typeutils.StringFromMap(malwareData, "mode"),
			Blocklist: typeutils.BoolFromMap(malwareData, "blocklist"),
		})
		diags.Append(d...)
	} else {
		malwareObj = types.ObjectNull(malwareLinuxAttrTypes())
	}

	memProtData := getMap(data, "memory_protection")
	var memProtObj types.Object
	if memProtData != nil {
		var d diag.Diagnostics
		memProtObj, d = types.ObjectValueFrom(ctx, protectionModeAttrTypes(), protectionModeModel{
			Mode:      typeutils.StringFromMap(memProtData, "mode"),
			Supported: typeutils.BoolFromMap(memProtData, attrSupported),
		})
		diags.Append(d...)
	} else {
		memProtObj = types.ObjectNull(protectionModeAttrTypes())
	}

	behProtData := getMap(data, "behavior_protection")
	var behProtObj types.Object
	if behProtData != nil {
		var d diag.Diagnostics
		behProtObj, d = types.ObjectValueFrom(ctx, behaviorProtectionAttrTypes(), behaviorProtectionModel{
			Mode:              typeutils.StringFromMap(behProtData, "mode"),
			Supported:         typeutils.BoolFromMap(behProtData, attrSupported),
			ReputationService: typeutils.BoolFromMap(behProtData, attrReputationService),
		})
		diags.Append(d...)
	} else {
		behProtObj = types.ObjectNull(behaviorProtectionAttrTypes())
	}

	popupData := getMap(data, attrPopup)
	popupObj, d := mapMacLinuxPopupFromAPI(ctx, popupData)
	diags.Append(d...)

	loggingData := getMap(data, "logging")
	var loggingObj types.Object
	if loggingData != nil {
		var d diag.Diagnostics
		loggingObj, d = types.ObjectValueFrom(ctx, loggingAttrTypes(), loggingModel{
			File: typeutils.StringFromMap(loggingData, "file"),
		})
		diags.Append(d...)
	} else {
		loggingObj = types.ObjectNull(loggingAttrTypes())
	}

	linuxObj, d := types.ObjectValueFrom(ctx, linuxAttrTypes(), linuxPolicyModel{
		Events:             eventsObj,
		Malware:            malwareObj,
		MemoryProtection:   memProtObj,
		BehaviorProtection: behProtObj,
		Popup:              popupObj,
		Logging:            loggingObj,
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
		"windows": types.ObjectType{AttrTypes: windowsAttrTypes()},
		"mac":     types.ObjectType{AttrTypes: macAttrTypes()},
		"linux":   types.ObjectType{AttrTypes: linuxAttrTypes()},
	}
}
