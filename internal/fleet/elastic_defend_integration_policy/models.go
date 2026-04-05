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
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// elasticDefendIntegrationPolicyModel is the Terraform state model for the
// elasticstack_fleet_elastic_defend_integration_policy resource.
type elasticDefendIntegrationPolicyModel struct {
	ID                 types.String `tfsdk:"id"`
	PolicyID           types.String `tfsdk:"policy_id"`
	Name               types.String `tfsdk:"name"`
	Namespace          types.String `tfsdk:"namespace"`
	AgentPolicyID      types.String `tfsdk:"agent_policy_id"`
	Description        types.String `tfsdk:"description"`
	Enabled            types.Bool   `tfsdk:"enabled"`
	Force              types.Bool   `tfsdk:"force"`
	IntegrationVersion types.String `tfsdk:"integration_version"`
	SpaceIDs           types.Set    `tfsdk:"space_ids"`
	Preset             types.String `tfsdk:"preset"`
	Policy             types.Object `tfsdk:"policy"`
}

// policyModel holds the top-level policy nested attribute.
type policyModel struct {
	Windows types.Object `tfsdk:"windows"`
	Mac     types.Object `tfsdk:"mac"`
	Linux   types.Object `tfsdk:"linux"`
}

// windowsPolicyModel holds the Windows-specific policy settings.
type windowsPolicyModel struct {
	Events                 types.Object `tfsdk:"events"`
	Malware                types.Object `tfsdk:"malware"`
	Ransomware             types.Object `tfsdk:"ransomware"`
	MemoryProtection       types.Object `tfsdk:"memory_protection"`
	BehaviorProtection     types.Object `tfsdk:"behavior_protection"`
	Popup                  types.Object `tfsdk:"popup"`
	Logging                types.Object `tfsdk:"logging"`
	AntivirusRegistration  types.Object `tfsdk:"antivirus_registration"`
	AttackSurfaceReduction types.Object `tfsdk:"attack_surface_reduction"`
}

// macPolicyModel holds the macOS-specific policy settings.
type macPolicyModel struct {
	Events             types.Object `tfsdk:"events"`
	Malware            types.Object `tfsdk:"malware"`
	MemoryProtection   types.Object `tfsdk:"memory_protection"`
	BehaviorProtection types.Object `tfsdk:"behavior_protection"`
	Popup              types.Object `tfsdk:"popup"`
	Logging            types.Object `tfsdk:"logging"`
}

// linuxPolicyModel holds the Linux-specific policy settings.
type linuxPolicyModel struct {
	Events             types.Object `tfsdk:"events"`
	Malware            types.Object `tfsdk:"malware"`
	MemoryProtection   types.Object `tfsdk:"memory_protection"`
	BehaviorProtection types.Object `tfsdk:"behavior_protection"`
	Popup              types.Object `tfsdk:"popup"`
	Logging            types.Object `tfsdk:"logging"`
}

// windowsEventsModel holds the Windows event collection flags.
type windowsEventsModel struct {
	Process          types.Bool `tfsdk:"process"`
	Network          types.Bool `tfsdk:"network"`
	File             types.Bool `tfsdk:"file"`
	DllAndDriverLoad types.Bool `tfsdk:"dll_and_driver_load"`
	DNS              types.Bool `tfsdk:"dns"`
	Registry         types.Bool `tfsdk:"registry"`
	Security         types.Bool `tfsdk:"security"`
	Authentication   types.Bool `tfsdk:"authentication"`
}

// macEventsModel holds the macOS event collection flags.
type macEventsModel struct {
	Process types.Bool `tfsdk:"process"`
	Network types.Bool `tfsdk:"network"`
	File    types.Bool `tfsdk:"file"`
}

// linuxEventsModel holds the Linux event collection flags.
type linuxEventsModel struct {
	Process     types.Bool `tfsdk:"process"`
	Network     types.Bool `tfsdk:"network"`
	File        types.Bool `tfsdk:"file"`
	SessionData types.Bool `tfsdk:"session_data"`
	TtyIO       types.Bool `tfsdk:"tty_io"`
}

// malwareModel holds malware protection settings (Windows/Mac have notify_user and on_write_scan).
type malwareFullModel struct {
	Mode        types.String `tfsdk:"mode"`
	Blocklist   types.Bool   `tfsdk:"blocklist"`
	OnWriteScan types.Bool   `tfsdk:"on_write_scan"`
	NotifyUser  types.Bool   `tfsdk:"notify_user"`
}

// malwareLinuxModel holds malware protection settings for Linux (no on_write_scan/notify_user).
type malwareLinuxModel struct {
	Mode      types.String `tfsdk:"mode"`
	Blocklist types.Bool   `tfsdk:"blocklist"`
}

// protectionModeModel holds mode+supported settings for ransomware and memory protection.
type protectionModeModel struct {
	Mode      types.String `tfsdk:"mode"`
	Supported types.Bool   `tfsdk:"supported"`
}

// behaviorProtectionModel holds mode+supported+reputation_service settings.
type behaviorProtectionModel struct {
	Mode              types.String `tfsdk:"mode"`
	Supported         types.Bool   `tfsdk:"supported"`
	ReputationService types.Bool   `tfsdk:"reputation_service"`
}

// popupItemModel holds message+enabled for a single popup entry.
type popupItemModel struct {
	Message types.String `tfsdk:"message"`
	Enabled types.Bool   `tfsdk:"enabled"`
}

// windowsPopupModel holds the Windows popup notification settings.
type windowsPopupModel struct {
	Malware            types.Object `tfsdk:"malware"`
	Ransomware         types.Object `tfsdk:"ransomware"`
	MemoryProtection   types.Object `tfsdk:"memory_protection"`
	BehaviorProtection types.Object `tfsdk:"behavior_protection"`
}

// macLinuxPopupModel holds the Mac/Linux popup notification settings (no ransomware).
type macLinuxPopupModel struct {
	Malware            types.Object `tfsdk:"malware"`
	MemoryProtection   types.Object `tfsdk:"memory_protection"`
	BehaviorProtection types.Object `tfsdk:"behavior_protection"`
}

// loggingModel holds the logging settings.
type loggingModel struct {
	File types.String `tfsdk:"file"`
}

// antivirusRegistrationModel holds the antivirus registration settings.
type antivirusRegistrationModel struct {
	Enabled types.Bool `tfsdk:"enabled"`
}

// credentialHardeningModel holds the credential hardening settings.
type credentialHardeningModel struct {
	Enabled types.Bool `tfsdk:"enabled"`
}

// attackSurfaceReductionModel holds the attack surface reduction settings.
type attackSurfaceReductionModel struct {
	CredentialHardening types.Object `tfsdk:"credential_hardening"`
}

// defendPrivateState is used to store server-managed Defend payloads in
// provider private state. These values are required for updates but should
// not be exposed in the public Terraform schema.
type defendPrivateState struct {
	// ArtifactManifest is the opaque artifact manifest returned by the Defend API.
	ArtifactManifest map[string]any `json:"artifact_manifest,omitempty"`

	// Version is the package policy ES version token used for optimistic
	// concurrency control on update requests.
	Version string `json:"version,omitempty"`
}
