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

package agentlesspolicy

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/policyshape"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// MinVersion is the minimum Kibana version required for the Fleet
// managed_integrations API. Verified against a 9.5.0-SNAPSHOT build; same
// floor as policyshape.MinVersionCondition (see openspec/changes/
// fleet-managed-integration/design.md Decision 8 and spec.md "Version gate
// for managed_integrations endpoint").
var MinVersion = version.Must(version.NewVersion("9.5.0"))

// agentlessPolicyModel is the Plugin Framework model for the
// elasticstack_fleet_agentless_policy resource.
//
// Task 4 of the fleet-agentless-policy OpenSpec change
// (openspec/changes/fleet-agentless-policy/tasks.md, section "4. Resource:
// schema") adds one field per schema attribute defined in schema.go (see
// specs/fleet-agentless-policy/spec.md, "Schema attributes" requirement).
// CRUD population (populateFromAPI/toAPI*Model conversion functions) is
// Task 5's responsibility; this file only declares the struct shape.
//
// packageModel, cloudConnectorModel, and globalDataTagModel back the
// `package`, `cloud_connector`, and `global_data_tags` nested attributes
// respectively; they are plain (non-custom-type) nested objects decoded via
// types.Object/types.List .As(), matching the convention used by
// internal/fleet/agentpolicy's advanced_settings/global_data_tags fields.
// `inputs` and `vars_json` reuse the shared policyshape custom types
// directly (policyshape.InputsValue / policyshape.VarsJSONValue) -- no
// local type duplication.
type agentlessPolicyModel struct {
	entitycore.ResourceTimeoutsField
	ID                               types.String              `tfsdk:"id"`
	KibanaConnection                 types.List                `tfsdk:"kibana_connection"`
	PolicyID                         types.String              `tfsdk:"policy_id"`
	Name                             types.String              `tfsdk:"name"`
	Description                      types.String              `tfsdk:"description"`
	Namespace                        types.String              `tfsdk:"namespace"`
	SpaceIDs                         types.Set                 `tfsdk:"space_ids"` // > string
	Package                          types.Object              `tfsdk:"package"`   // > packageModel
	PolicyTemplate                   types.String              `tfsdk:"policy_template"`
	VarsJSON                         policyshape.VarsJSONValue `tfsdk:"vars_json"`
	VarGroupSelections               types.Map                 `tfsdk:"var_group_selections"`               // > string
	Inputs                           policyshape.InputsValue   `tfsdk:"inputs"`                             // > policyshape.InputModel
	CloudConnector                   types.Object              `tfsdk:"cloud_connector"`                    // > cloudConnectorModel
	GlobalDataTags                   types.List                `tfsdk:"global_data_tags"`                   // > globalDataTagModel
	AdditionalDatastreamsPermissions types.List                `tfsdk:"additional_datastreams_permissions"` // > string
	CreateDatasetTemplates           types.Bool                `tfsdk:"create_dataset_templates"`
	Force                            types.Bool                `tfsdk:"force"`
	ForceDelete                      types.Bool                `tfsdk:"force_delete"`
	SkipTopologyCheck                types.Bool                `tfsdk:"skip_topology_check"`
	CreatedAt                        types.String              `tfsdk:"created_at"`
	UpdatedAt                        types.String              `tfsdk:"updated_at"`
}

// Note: the Go representations of the `package`, `cloud_connector`, and
// `global_data_tags` nested attributes (e.g. a packageModel with
// Name/Version/Title fields) are intentionally not declared here. They are
// pure conversion-layer plumbing consumed only by populateFromAPI/toAPI
// model functions, which is Task 5's responsibility (see
// openspec/changes/fleet-agentless-policy/tasks.md, section "5. Resource:
// CRUD + import"); declaring them now with no caller would be dead code.

func (m agentlessPolicyModel) GetID() types.String         { return m.ID }
func (m agentlessPolicyModel) GetResourceID() types.String { return m.PolicyID }

// defaultSpaceID is the Kibana space used when space_ids is not configured.
// See specs/fleet-agentless-policy/spec.md, "Resource identity and composite
// ID": "space_ids SHALL be Optional+Computed defaulting to [\"default\"]".
const defaultSpaceID = "default"

// GetSpaceID returns the first non-empty space ID from SpaceIDs, or
// defaultSpaceID if none is set (SpaceIDs is null, unknown, or contains only
// empty/unknown elements).
//
// Task 5 note: this resource is genuinely space-scoped (unlike
// internal/fleet/output and internal/fleet/serverhost, which opt out of the
// space-required check entirely via KibanaUnscopedSpace/IsUnscopedSpace), so
// entitycore's Create/Update path calls validateSpaceID, which errors when
// GetSpaceID() returns an empty string for a non-unscoped model. Since
// space_ids has no schema-level Default (Computed with no Default plan
// modifier -- consistent with the same repo pattern in
// internal/fleet/output/schema.go and internal/fleet/serverhost/schema.go),
// omitting space_ids from config leaves it Unknown during Create; defaulting
// here (rather than returning "") is what makes the spec's "Create with
// auto-assigned policy_id" scenario -- which never mentions space_ids --
// actually succeed. This is a deliberate behavior change from the Task 3
// skeleton (which returned "" for null/unknown), made in Task 5 because
// Create is the first caller that actually exercises this path end-to-end;
// see the corresponding test update in schema_test.go/entitycore_contract_test.go.
func (m agentlessPolicyModel) GetSpaceID() types.String {
	if m.SpaceIDs.IsNull() || m.SpaceIDs.IsUnknown() {
		return types.StringValue(defaultSpaceID)
	}
	for _, elem := range m.SpaceIDs.Elements() {
		s, ok := elem.(types.String)
		if !ok || s.IsNull() || s.IsUnknown() {
			continue
		}
		if v := s.ValueString(); v != "" {
			return s
		}
	}
	return types.StringValue(defaultSpaceID)
}

func (m agentlessPolicyModel) GetKibanaConnection() types.List { return m.KibanaConnection }

// GetVersionRequirements enforces the minimum Kibana version for the Fleet
// managed_integrations API (experimental, added in Kibana 9.5.0). See
// openspec/changes/fleet-managed-integration/specs/fleet-managed-integration/
// spec.md, requirement "Version gate for managed_integrations endpoint".
func (m agentlessPolicyModel) GetVersionRequirements(_ context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return []entitycore.VersionRequirement{
		{
			MinVersion:   *MinVersion,
			ErrorMessage: fmt.Sprintf("Fleet managed integrations require Elastic Stack v%s or later (experimental API).", MinVersion),
		},
	}, nil
}
