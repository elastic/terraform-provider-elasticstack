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

package entity

import (
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	MinVersion = version.Must(version.NewVersion("9.4.0"))
)

type tfModel struct {
	ID               types.String         `tfsdk:"id"`
	KibanaConnection types.List           `tfsdk:"kibana_connection"`
	SpaceID          types.String         `tfsdk:"space_id"`
	EntityType       types.String         `tfsdk:"entity_type"`
	EntityID         types.String         `tfsdk:"entity_id"`
	Timestamp        types.String         `tfsdk:"timestamp"`
	Entity           types.Object         `tfsdk:"entity"`
	Host             types.Object         `tfsdk:"host"`
	User             types.Object         `tfsdk:"user"`
	Service          types.Object         `tfsdk:"service"`
	Cloud            types.Object         `tfsdk:"cloud"`
	Asset            types.Object         `tfsdk:"asset"`
	Orchestrator     types.Object         `tfsdk:"orchestrator"`
	Event            types.Object         `tfsdk:"event"`
	Labels           types.Map            `tfsdk:"labels"`
	Tags             types.Set            `tfsdk:"tags"`
	EntityJSON       jsontypes.Normalized `tfsdk:"entity_json"`
	HostJSON         jsontypes.Normalized `tfsdk:"host_json"`
	UserJSON         jsontypes.Normalized `tfsdk:"user_json"`
	ServiceJSON      jsontypes.Normalized `tfsdk:"service_json"`
	CloudJSON        jsontypes.Normalized `tfsdk:"cloud_json"`
	AssetJSON        jsontypes.Normalized `tfsdk:"asset_json"`
	OrchestratorJSON jsontypes.Normalized `tfsdk:"orchestrator_json"`
	EventJSON        jsontypes.Normalized `tfsdk:"event_json"`
	LabelsJSON       jsontypes.Normalized `tfsdk:"labels_json"`
	Force            types.Bool           `tfsdk:"force"`
	DocumentJSON     jsontypes.Normalized `tfsdk:"document_json"`
	ResponseJSON     jsontypes.Normalized `tfsdk:"response_json"`
}

var _ entitycore.KibanaResourceModel = tfModel{}
var _ entitycore.WithVersionRequirements = (*tfModel)(nil)

func (model tfModel) GetID() types.String             { return model.ID }
func (model tfModel) GetSpaceID() types.String        { return model.SpaceID }
func (model tfModel) GetKibanaConnection() types.List { return model.KibanaConnection }
func (model tfModel) GetResourceID() types.String     { return types.StringValue("") }

func (*tfModel) GetVersionRequirements() ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return []entitycore.VersionRequirement{{
		MinVersion:   *MinVersion,
		ErrorMessage: fmt.Sprintf("elasticstack_kibana_security_entity_store_entity is supported only for Kibana v%s and above", MinVersion.String()),
	}}, nil
}

// entityBlockModel represents the typed "entity" block.
type entityBlockModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Type          types.String `tfsdk:"type"`
	SubType       types.String `tfsdk:"sub_type"`
	Source        types.String `tfsdk:"source"`
	Attributes    types.Object `tfsdk:"attributes"`
	Behaviors     types.Object `tfsdk:"behaviors"`
	Lifecycle     types.Object `tfsdk:"lifecycle"`
	Risk          types.Object `tfsdk:"risk"`
	Relationships types.Object `tfsdk:"relationships"`
}

// entityAttributesBlockModel represents the nested "attributes" block inside entity.
type entityAttributesBlockModel struct {
	Asset      types.Bool `tfsdk:"asset"`
	Managed    types.Bool `tfsdk:"managed"`
	Privileged types.Bool `tfsdk:"privileged"`
	MfaEnabled types.Bool `tfsdk:"mfa_enabled"`
}

// entityBehaviorsBlockModel represents the nested "behaviors" block inside entity.
type entityBehaviorsBlockModel struct {
	BruteForceVictim types.Bool `tfsdk:"brute_force_victim"`
	NewCountryLogin  types.Bool `tfsdk:"new_country_login"`
	UsedUsbDevice    types.Bool `tfsdk:"used_usb_device"`
}

// entityLifecycleBlockModel represents the nested "lifecycle" block inside entity.
type entityLifecycleBlockModel struct {
	FirstSeen    types.String `tfsdk:"first_seen"`
	LastSeen     types.String `tfsdk:"last_seen"`
	LastActivity types.String `tfsdk:"last_activity"`
}

// entityRiskBlockModel represents the nested "risk" block (used inside entity, host, user, service).
type entityRiskBlockModel struct {
	CalculatedLevel     types.String  `tfsdk:"calculated_level"`
	CalculatedScore     types.Float64 `tfsdk:"calculated_score"`
	CalculatedScoreNorm types.Float64 `tfsdk:"calculated_score_norm"`
}

// entityRelationshipsBlockModel represents the nested "relationships" block inside entity.
type entityRelationshipsBlockModel struct {
	OwnedBy              types.Set `tfsdk:"owned_by"`
	Owns                 types.Set `tfsdk:"owns"`
	SupervisedBy         types.Set `tfsdk:"supervised_by"`
	Supervises           types.Set `tfsdk:"supervises"`
	DependsOn            types.Set `tfsdk:"depends_on"`
	DependentOf          types.Set `tfsdk:"dependent_of"`
	CommunicatesWith     types.Set `tfsdk:"communicates_with"`
	AccessesFrequently   types.Set `tfsdk:"accesses_frequently"`
	AccessedFrequentlyBy types.Set `tfsdk:"accessed_frequently_by"`
	AccessesInfrequently types.Set `tfsdk:"accesses_infrequently"`
}

// hostBlockModel represents the typed "host" block.
type hostBlockModel struct {
	Name         types.String `tfsdk:"name"`
	Domain       types.Set    `tfsdk:"domain"`
	Hostname     types.Set    `tfsdk:"hostname"`
	ID           types.Set    `tfsdk:"id"`
	IP           types.Set    `tfsdk:"ip"`
	Mac          types.Set    `tfsdk:"mac"`
	Type         types.Set    `tfsdk:"type"`
	Architecture types.Set    `tfsdk:"architecture"`
	Os           types.Object `tfsdk:"os"`
	Risk         types.Object `tfsdk:"risk"`
}

// hostOsBlockModel represents the nested "os" block inside host.
type hostOsBlockModel struct {
	Family   types.String `tfsdk:"family"`
	Full     types.String `tfsdk:"full"`
	Kernel   types.String `tfsdk:"kernel"`
	Name     types.String `tfsdk:"name"`
	Platform types.String `tfsdk:"platform"`
	Type     types.String `tfsdk:"type"`
	Version  types.String `tfsdk:"version"`
}

// userBlockModel represents the typed "user" block.
type userBlockModel struct {
	Name     types.String `tfsdk:"name"`
	Domain   types.Set    `tfsdk:"domain"`
	Email    types.Set    `tfsdk:"email"`
	FullName types.Set    `tfsdk:"full_name"`
	Hash     types.Set    `tfsdk:"hash"`
	ID       types.Set    `tfsdk:"id"`
	Roles    types.Set    `tfsdk:"roles"`
	Risk     types.Object `tfsdk:"risk"`
}

// serviceBlockModel represents the typed "service" block.
type serviceBlockModel struct {
	Name types.String `tfsdk:"name"`
	Risk types.Object `tfsdk:"risk"`
}

// orchestratorBlockModel represents the typed "orchestrator" block.
type orchestratorBlockModel struct {
	Name           types.String `tfsdk:"name"`
	Type           types.String `tfsdk:"type"`
	Namespace      types.String `tfsdk:"namespace"`
	ClusterID      types.String `tfsdk:"cluster_id"`
	ClusterName    types.String `tfsdk:"cluster_name"`
	ClusterVersion types.String `tfsdk:"cluster_version"`
	ResourceID     types.String `tfsdk:"resource_id"`
	ResourceName   types.String `tfsdk:"resource_name"`
	ResourceType   types.String `tfsdk:"resource_type"`
}

// cloudBlockModel represents the typed "cloud" block.
type cloudBlockModel struct {
	Provider    types.String `tfsdk:"provider"`
	Region      types.String `tfsdk:"region"`
	AccountID   types.String `tfsdk:"account_id"`
	AccountName types.String `tfsdk:"account_name"`
	ProjectID   types.String `tfsdk:"project_id"`
	ProjectName types.String `tfsdk:"project_name"`
	ServiceName types.String `tfsdk:"service_name"`
}

// eventBlockModel represents the typed "event" block.
type eventBlockModel struct {
	Category  types.String `tfsdk:"category"`
	Type      types.String `tfsdk:"type"`
	Dataset   types.String `tfsdk:"dataset"`
	Kind      types.String `tfsdk:"kind"`
	Outcome   types.String `tfsdk:"outcome"`
	Provider  types.String `tfsdk:"provider"`
	Action    types.String `tfsdk:"action"`
	Code      types.String `tfsdk:"code"`
	Reference types.String `tfsdk:"reference"`
	Reason    types.String `tfsdk:"reason"`
	Severity  types.String `tfsdk:"severity"`
	Timezone  types.String `tfsdk:"timezone"`
	URL       types.String `tfsdk:"url"`
	Ingested  types.String `tfsdk:"ingested"`
}

// assetBlockModel represents the typed "asset" block.
type assetBlockModel struct {
	Criticality         types.String  `tfsdk:"criticality"`
	CriticalityFeedback types.Object  `tfsdk:"criticality_feedback"`
	Owner               types.Object  `tfsdk:"owner"`
	Value               types.Float64 `tfsdk:"value"`
}

// assetCriticalityFeedbackBlockModel represents the nested "criticality_feedback" block inside asset.
type assetCriticalityFeedbackBlockModel struct {
	Notes  types.String `tfsdk:"notes"`
	Reason types.String `tfsdk:"reason"`
}

// assetOwnerBlockModel represents the nested "owner" block inside asset.
type assetOwnerBlockModel struct {
	Name       types.String `tfsdk:"name"`
	Department types.String `tfsdk:"department"`
	Email      types.String `tfsdk:"email"`
	Ext        types.String `tfsdk:"ext"`
}
