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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

	// An empty or absent server value imports as null; only a known empty list in the model round-trips as [].
	switch {
	case policy.AdditionalDatastreamsPermissions != nil && len(*policy.AdditionalDatastreamsPermissions) > 0:
		perms, d := types.ListValueFrom(ctx, types.StringType, *policy.AdditionalDatastreamsPermissions)
		diags.Append(d...)
		model.AdditionalDatastreamsPermissions = perms
	case typeutils.IsKnown(model.AdditionalDatastreamsPermissions) && len(model.AdditionalDatastreamsPermissions.Elements()) == 0:
		// Preserve an explicit empty list so `[]` round-trips.
	default:
		model.AdditionalDatastreamsPermissions = types.ListNull(types.StringType)
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

	memProtObj, d := mapOptionalObject(ctx, data, attrMemoryProtection, memoryProtectionAttrTypes(), func(m map[string]any) memoryProtectionModel {
		return memoryProtectionModel{
			Mode:                 typeutils.StringFromMap(m, attrMode),
			Supported:            typeutils.BoolFromMap(m, attrSupported),
			CustomYaraSignatures: typeutils.BoolFromMap(m, attrCustomYara),
		}
	})
	diags.Append(d...)

	behProtObj, d := mapOptionalObject(ctx, data, attrBehaviorProtection, behaviorProtectionAttrTypes(), func(m map[string]any) behaviorProtectionModel {
		return behaviorProtectionModel{
			Mode:              typeutils.StringFromMap(m, attrMode),
			Supported:         typeutils.BoolFromMap(m, attrSupported),
			ReputationService: typeutils.BoolFromMap(m, attrReputationService),
		}
	})
	diags.Append(d...)

	loggingObj, d := mapOptionalObject(ctx, data, attrLogging, loggingAttrTypes(), func(m map[string]any) loggingModel {
		return loggingModel{
			File: typeutils.StringFromMap(m, attrFile),
		}
	})
	diags.Append(d...)

	return commonPolicyFields{
		MemoryProtection:   memProtObj,
		BehaviorProtection: behProtObj,
		Logging:            loggingObj,
	}, diags
}

// macWindowsMalwareFields holds the malware and ransomware sub-objects, which
// are byte-for-byte identical between mapMacPolicyFromAPI and
// mapWindowsPolicyFromAPI. Linux is excluded: its malware shape lacks
// notify_user and it has no ransomware field, so it is not covered by
// commonPolicyFields.
type macWindowsMalwareFields struct {
	Malware    types.Object
	Ransomware types.Object
}

// mapMalwareAndRansomwareFromAPI extracts malware and ransomware — blocks
// that are byte-for-byte identical across mapMacPolicyFromAPI and
// mapWindowsPolicyFromAPI.
func mapMalwareAndRansomwareFromAPI(ctx context.Context, data map[string]any) (macWindowsMalwareFields, diag.Diagnostics) {
	var diags diag.Diagnostics

	malwareObj, d := mapOptionalObject(ctx, data, attrMalware, malwareFullAttrTypes(), func(m map[string]any) malwareFullModel {
		return malwareFullModel{
			Mode:        typeutils.StringFromMap(m, attrMode),
			Blocklist:   typeutils.BoolFromMap(m, attrBlocklist),
			OnWriteScan: typeutils.BoolFromMap(m, attrOnWriteScan),
			NotifyUser:  typeutils.BoolFromMap(m, attrNotifyUser),
		}
	})
	diags.Append(d...)

	ransomwareObj, d := mapOptionalObject(ctx, data, attrRansomware, protectionModeAttrTypes(), func(m map[string]any) protectionModeModel {
		return protectionModeModel{
			Mode:      typeutils.StringFromMap(m, attrMode),
			Supported: typeutils.BoolFromMap(m, attrSupported),
		}
	})
	diags.Append(d...)

	return macWindowsMalwareFields{
		Malware:    malwareObj,
		Ransomware: ransomwareObj,
	}, diags
}
