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

	model.PolicyID = types.StringValue(policy.Id)
	model.Name = types.StringValue(policy.Name)
	model.Namespace = types.StringPointerValue(policy.Namespace)
	// Treat empty string description as null (Kibana may return "" when not set)
	if policy.Description != nil && *policy.Description == "" {
		model.Description = types.StringNull()
	} else {
		model.Description = types.StringPointerValue(policy.Description)
	}
	model.Enabled = types.BoolValue(policy.Enabled)

	if policy.Package != nil {
		model.IntegrationVersion = types.StringValue(policy.Package.Version)
	}

	model.AgentPolicyID = types.StringPointerValue(policy.PolicyId)

	// Populate space_ids — only overwrite when the API actually returns them.
	// If the API omits space_ids, preserve the existing model value so
	// space-aware operations (e.g. update, delete) continue to work correctly.
	var operationalSpaceID string
	switch {
	case policy.SpaceIds != nil && len(*policy.SpaceIds) > 0:
		spaceIDs, d := types.SetValueFrom(ctx, types.StringType, *policy.SpaceIds)
		diags.Append(d...)
		model.SpaceIDs = spaceIDs
		operationalSpaceID = (*policy.SpaceIds)[0]
	case model.SpaceIDs.IsNull() || model.SpaceIDs.IsUnknown():
		model.SpaceIDs = types.SetNull(types.StringType)
	default:
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
		model.ID = types.StringValue(operationalSpaceID + "/" + policy.Id)
	} else {
		model.ID = types.StringValue(policy.Id)
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
								if p, ok := ecMap["preset"]; ok {
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

// Helper to extract bool from nested map.
func getBool(m map[string]any, key string) types.Bool {
	if m == nil {
		return types.BoolNull()
	}
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return types.BoolValue(b)
		}
	}
	return types.BoolNull()
}

// Helper to extract string from nested map.
func getString(m map[string]any, key string) types.String {
	if m == nil {
		return types.StringNull()
	}
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return types.StringValue(s)
		}
	}
	return types.StringNull()
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
			Process:          getBool(eventsData, "process"),
			Network:          getBool(eventsData, "network"),
			File:             getBool(eventsData, "file"),
			DllAndDriverLoad: getBool(eventsData, "dll_and_driver_load"),
			DNS:              getBool(eventsData, "dns"),
			Registry:         getBool(eventsData, "registry"),
			Security:         getBool(eventsData, "security"),
			Authentication:   getBool(eventsData, "authentication"),
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
			Mode:        getString(malwareData, "mode"),
			Blocklist:   getBool(malwareData, "blocklist"),
			OnWriteScan: getBool(malwareData, "on_write_scan"),
			NotifyUser:  getBool(malwareData, "notify_user"),
		})
		diags.Append(d...)
	} else {
		malwareObj = types.ObjectNull(malwareFullAttrTypes())
	}

	ransomwareData := getMap(data, "ransomware")
	var ransomwareObj types.Object
	if ransomwareData != nil {
		var d diag.Diagnostics
		ransomwareObj, d = types.ObjectValueFrom(ctx, protectionModeAttrTypes(), protectionModeModel{
			Mode:      getString(ransomwareData, "mode"),
			Supported: getBool(ransomwareData, "supported"),
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
			Mode:      getString(memProtData, "mode"),
			Supported: getBool(memProtData, "supported"),
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
			Mode:              getString(behProtData, "mode"),
			Supported:         getBool(behProtData, "supported"),
			ReputationService: getBool(behProtData, "reputation_service"),
		})
		diags.Append(d...)
	} else {
		behProtObj = types.ObjectNull(behaviorProtectionAttrTypes())
	}

	popupData := getMap(data, "popup")
	popupObj, d := mapWindowsPopupFromAPI(ctx, popupData)
	diags.Append(d...)

	loggingData := getMap(data, "logging")
	var loggingObj types.Object
	if loggingData != nil {
		var d diag.Diagnostics
		loggingObj, d = types.ObjectValueFrom(ctx, loggingAttrTypes(), loggingModel{
			File: getString(loggingData, "file"),
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
			Mode:    getString(avrData, "mode"),
			Enabled: getBool(avrData, "enabled"),
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
				Enabled: getBool(chData, "enabled"),
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
	ransomwareData := getMap(data, "ransomware")
	memProtData := getMap(data, "memory_protection")
	behProtData := getMap(data, "behavior_protection")

	malwareObj, d := types.ObjectValueFrom(ctx, popupItemAttrTypes(), popupItemModel{
		Message: getString(malwareData, "message"),
		Enabled: getBool(malwareData, "enabled"),
	})
	diags.Append(d...)

	ransomwareObj, d := types.ObjectValueFrom(ctx, popupItemAttrTypes(), popupItemModel{
		Message: getString(ransomwareData, "message"),
		Enabled: getBool(ransomwareData, "enabled"),
	})
	diags.Append(d...)

	memProtObj, d := types.ObjectValueFrom(ctx, popupItemAttrTypes(), popupItemModel{
		Message: getString(memProtData, "message"),
		Enabled: getBool(memProtData, "enabled"),
	})
	diags.Append(d...)

	behProtObj, d := types.ObjectValueFrom(ctx, popupItemAttrTypes(), popupItemModel{
		Message: getString(behProtData, "message"),
		Enabled: getBool(behProtData, "enabled"),
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
			Process: getBool(eventsData, "process"),
			Network: getBool(eventsData, "network"),
			File:    getBool(eventsData, "file"),
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
			Mode:        getString(malwareData, "mode"),
			Blocklist:   getBool(malwareData, "blocklist"),
			OnWriteScan: getBool(malwareData, "on_write_scan"),
			NotifyUser:  getBool(malwareData, "notify_user"),
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
			Mode:      getString(memProtData, "mode"),
			Supported: getBool(memProtData, "supported"),
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
			Mode:              getString(behProtData, "mode"),
			Supported:         getBool(behProtData, "supported"),
			ReputationService: getBool(behProtData, "reputation_service"),
		})
		diags.Append(d...)
	} else {
		behProtObj = types.ObjectNull(behaviorProtectionAttrTypes())
	}

	popupData := getMap(data, "popup")
	popupObj, d := mapMacLinuxPopupFromAPI(ctx, popupData)
	diags.Append(d...)

	loggingData := getMap(data, "logging")
	var loggingObj types.Object
	if loggingData != nil {
		var d diag.Diagnostics
		loggingObj, d = types.ObjectValueFrom(ctx, loggingAttrTypes(), loggingModel{
			File: getString(loggingData, "file"),
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
			Process:     getBool(eventsData, "process"),
			Network:     getBool(eventsData, "network"),
			File:        getBool(eventsData, "file"),
			SessionData: getBool(eventsData, "session_data"),
			TtyIO:       getBool(eventsData, "tty_io"),
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
			Mode:      getString(malwareData, "mode"),
			Blocklist: getBool(malwareData, "blocklist"),
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
			Mode:      getString(memProtData, "mode"),
			Supported: getBool(memProtData, "supported"),
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
			Mode:              getString(behProtData, "mode"),
			Supported:         getBool(behProtData, "supported"),
			ReputationService: getBool(behProtData, "reputation_service"),
		})
		diags.Append(d...)
	} else {
		behProtObj = types.ObjectNull(behaviorProtectionAttrTypes())
	}

	popupData := getMap(data, "popup")
	popupObj, d := mapMacLinuxPopupFromAPI(ctx, popupData)
	diags.Append(d...)

	loggingData := getMap(data, "logging")
	var loggingObj types.Object
	if loggingData != nil {
		var d diag.Diagnostics
		loggingObj, d = types.ObjectValueFrom(ctx, loggingAttrTypes(), loggingModel{
			File: getString(loggingData, "file"),
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
		Message: getString(malwareData, "message"),
		Enabled: getBool(malwareData, "enabled"),
	})
	diags.Append(d...)

	memProtObj, d := types.ObjectValueFrom(ctx, popupItemAttrTypes(), popupItemModel{
		Message: getString(memProtData, "message"),
		Enabled: getBool(memProtData, "enabled"),
	})
	diags.Append(d...)

	behProtObj, d := types.ObjectValueFrom(ctx, popupItemAttrTypes(), popupItemModel{
		Message: getString(behProtData, "message"),
		Enabled: getBool(behProtData, "enabled"),
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
		"message": types.StringType,
		"enabled": types.BoolType,
	}
}

func windowsEventsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"process":             types.BoolType,
		"network":             types.BoolType,
		"file":                types.BoolType,
		"dll_and_driver_load": types.BoolType,
		"dns":                 types.BoolType,
		"registry":            types.BoolType,
		"security":            types.BoolType,
		"authentication":      types.BoolType,
	}
}

func macEventsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"process": types.BoolType,
		"network": types.BoolType,
		"file":    types.BoolType,
	}
}

func linuxEventsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"process":      types.BoolType,
		"network":      types.BoolType,
		"file":         types.BoolType,
		"session_data": types.BoolType,
		"tty_io":       types.BoolType,
	}
}

func malwareFullAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"mode":          types.StringType,
		"blocklist":     types.BoolType,
		"on_write_scan": types.BoolType,
		"notify_user":   types.BoolType,
	}
}

func malwareLinuxAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"mode":      types.StringType,
		"blocklist": types.BoolType,
	}
}

func protectionModeAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"mode":      types.StringType,
		"supported": types.BoolType,
	}
}

func behaviorProtectionAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"mode":               types.StringType,
		"supported":          types.BoolType,
		"reputation_service": types.BoolType,
	}
}

func loggingAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"file": types.StringType,
	}
}

func antivirusRegistrationAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"mode":    types.StringType,
		"enabled": types.BoolType,
	}
}

func credentialHardeningAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"enabled": types.BoolType,
	}
}

func attackSurfaceReductionAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"credential_hardening": types.ObjectType{AttrTypes: credentialHardeningAttrTypes()},
	}
}

func windowsPopupAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"malware":             types.ObjectType{AttrTypes: popupItemAttrTypes()},
		"ransomware":          types.ObjectType{AttrTypes: popupItemAttrTypes()},
		"memory_protection":   types.ObjectType{AttrTypes: popupItemAttrTypes()},
		"behavior_protection": types.ObjectType{AttrTypes: popupItemAttrTypes()},
	}
}

func macLinuxPopupAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"malware":             types.ObjectType{AttrTypes: popupItemAttrTypes()},
		"memory_protection":   types.ObjectType{AttrTypes: popupItemAttrTypes()},
		"behavior_protection": types.ObjectType{AttrTypes: popupItemAttrTypes()},
	}
}

func windowsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"events":                   types.ObjectType{AttrTypes: windowsEventsAttrTypes()},
		"malware":                  types.ObjectType{AttrTypes: malwareFullAttrTypes()},
		"ransomware":               types.ObjectType{AttrTypes: protectionModeAttrTypes()},
		"memory_protection":        types.ObjectType{AttrTypes: protectionModeAttrTypes()},
		"behavior_protection":      types.ObjectType{AttrTypes: behaviorProtectionAttrTypes()},
		"popup":                    types.ObjectType{AttrTypes: windowsPopupAttrTypes()},
		"logging":                  types.ObjectType{AttrTypes: loggingAttrTypes()},
		"antivirus_registration":   types.ObjectType{AttrTypes: antivirusRegistrationAttrTypes()},
		"attack_surface_reduction": types.ObjectType{AttrTypes: attackSurfaceReductionAttrTypes()},
	}
}

func macAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"events":              types.ObjectType{AttrTypes: macEventsAttrTypes()},
		"malware":             types.ObjectType{AttrTypes: malwareFullAttrTypes()},
		"memory_protection":   types.ObjectType{AttrTypes: protectionModeAttrTypes()},
		"behavior_protection": types.ObjectType{AttrTypes: behaviorProtectionAttrTypes()},
		"popup":               types.ObjectType{AttrTypes: macLinuxPopupAttrTypes()},
		"logging":             types.ObjectType{AttrTypes: loggingAttrTypes()},
	}
}

func linuxAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"events":              types.ObjectType{AttrTypes: linuxEventsAttrTypes()},
		"malware":             types.ObjectType{AttrTypes: malwareLinuxAttrTypes()},
		"memory_protection":   types.ObjectType{AttrTypes: protectionModeAttrTypes()},
		"behavior_protection": types.ObjectType{AttrTypes: behaviorProtectionAttrTypes()},
		"popup":               types.ObjectType{AttrTypes: macLinuxPopupAttrTypes()},
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
