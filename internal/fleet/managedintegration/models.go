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

package managedintegration

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/policyshape"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	globalDataTagStringValueAttr = "string_value"
	globalDataTagNumberValueAttr = "number_value"
)

// globalDataTagsItemModel is the element type of the `global_data_tags` map
// (keyed by tag name), matching schema.go's MapNestedAttribute item shape.
type globalDataTagsItemModel struct {
	StringValue types.String  `tfsdk:"string_value"`
	NumberValue types.Float32 `tfsdk:"number_value"`
}

func globalDataTagAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		globalDataTagStringValueAttr: types.StringType,
		globalDataTagNumberValueAttr: types.Float32Type,
	}
}

func globalDataTagsElementType() attr.Type {
	return types.ObjectType{AttrTypes: globalDataTagAttrTypes()}
}

// MinVersion is the minimum Kibana version required for the Fleet
// managed_integrations API. Verified against a 9.5.0-SNAPSHOT build; the
// `-SNAPSHOT` floor allows CI snapshot stacks via ordinary semver comparison.
// User-facing diagnostics name the 9.5.0 release (see minVersionUserFacing and
// policyshape.MinVersionCondition for the condition-support release).
// Enforced only via GetVersionRequirements and the entitycore envelope — there
// is no separate per-request capability gate in this package.
var MinVersion = version.Must(version.NewVersion("9.5.0-SNAPSHOT"))

// minVersionUserFacing is the release version named in practitioner-facing errors.
const minVersionUserFacing = "9.5.0"

// managedIntegrationModel is the Plugin Framework model for the
// elasticstack_fleet_managed_integration resource.
//
// Task 4 of the fleet-managed-integration OpenSpec change
// (openspec/changes/fleet-managed-integration/tasks.md, section "4. Resource:
// schema") adds one field per schema attribute defined in schema.go (see
// openspec/specs/fleet-managed-integration/spec.md, "Schema attributes" requirement).
// CRUD population (populateFromAPI/toAPI*Model conversion functions) is
// Task 5's responsibility; this file only declares the struct shape.
//
// packageModel and cloudConnectorModel back the `package` and `cloud_connector`
// nested attributes; global_data_tags uses globalDataTagsItemModel as the map
// element type (see globalDataTagsElementType). Plain nested objects are
// decoded via types.Object/types.Map helpers, matching internal/fleet/agentpolicy.
// `inputs` and `vars_json` reuse the shared policyshape custom types
// directly (policyshape.InputsValue / policyshape.VarsJSONValue) -- no
// local type duplication.
type managedIntegrationModel struct {
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
	GlobalDataTags                   types.Map                 `tfsdk:"global_data_tags"`                   // > globalDataTagsItemModel
	AdditionalDatastreamsPermissions types.List                `tfsdk:"additional_datastreams_permissions"` // > string
	CreateDatasetTemplates           types.Bool                `tfsdk:"create_dataset_templates"`
	Force                            types.Bool                `tfsdk:"force"`
	ForceDelete                      types.Bool                `tfsdk:"force_delete"`
	SkipTopologyCheck                types.Bool                `tfsdk:"skip_topology_check"`
	CreatedAt                        types.String              `tfsdk:"created_at"`
	UpdatedAt                        types.String              `tfsdk:"updated_at"`
}

// Note: nested attribute element types (packageModel, cloudConnectorModel,
// globalDataTagsItemModel) live in models.go / models_convert.go for conversion
// plumbing only.

func (m managedIntegrationModel) GetID() types.String         { return m.ID }
func (m managedIntegrationModel) GetResourceID() types.String { return m.PolicyID }

// defaultSpaceID is the Kibana space used when space_ids is not configured.
// See openspec/specs/fleet-managed-integration/spec.md, "Resource identity and composite
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
func (m managedIntegrationModel) GetSpaceID() types.String {
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

func (m managedIntegrationModel) GetKibanaConnection() types.List { return m.KibanaConnection }

// GetVersionRequirements enforces the minimum Kibana version for the Fleet
// managed_integrations API (experimental, added in Kibana 9.5.0). See
// openspec/changes/fleet-managed-integration/specs/fleet-managed-integration/
// spec.md, requirement "Version gate for managed_integrations endpoint".
func (m managedIntegrationModel) GetVersionRequirements(_ context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return []entitycore.VersionRequirement{
		{
			MinVersion:   *MinVersion,
			ErrorMessage: fmt.Sprintf("Fleet managed integrations require Elastic Stack v%s or later (experimental API).", minVersionUserFacing),
		},
	}, nil
}
