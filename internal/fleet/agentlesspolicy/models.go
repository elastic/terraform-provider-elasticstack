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
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// MinVersion is the minimum Kibana version required for the Fleet
// agentless policies API. The API is experimental and was added in Kibana
// 9.3.0 (see proposal.md and design.md of the fleet-agentless-policy
// OpenSpec change, and the spec.md "Version gating" requirement).
var MinVersion = version.Must(version.NewVersion("9.3.0"))

// agentlessPolicyModel is the Plugin Framework model for the
// elasticstack_fleet_agentless_policy resource.
//
// This is a skeleton for Task 3 of the fleet-agentless-policy OpenSpec
// change (openspec/changes/fleet-agentless-policy/tasks.md, section "3.
// Resource: skeleton, model, and spike"): it only carries the fields needed
// to satisfy entitycore.KibanaResourceModel and version-requirement wiring.
// The full schema-driven field set (package, inputs, vars_json,
// cloud_connector, global_data_tags, etc. -- see
// specs/fleet-agentless-policy/spec.md, "Schema attributes") is added in
// Task 4, and CRUD population (populateFromAPI/toAPI*Model) in Task 5.
type agentlessPolicyModel struct {
	entitycore.ResourceTimeoutsField
	ID               types.String `tfsdk:"id"`
	KibanaConnection types.List   `tfsdk:"kibana_connection"`
	PolicyID         types.String `tfsdk:"policy_id"`
	SpaceIDs         types.Set    `tfsdk:"space_ids"` // > string
}

func (m agentlessPolicyModel) GetID() types.String         { return m.ID }
func (m agentlessPolicyModel) GetResourceID() types.String { return m.PolicyID }

// GetSpaceID returns the first non-empty space ID from SpaceIDs, or an empty
// string if none is set. Mirrors the pattern used by other Fleet resources
// with a space_ids set attribute (see internal/fleet/serverhost/models.go
// and internal/fleet/output/models.go).
func (m agentlessPolicyModel) GetSpaceID() types.String {
	if m.SpaceIDs.IsNull() || m.SpaceIDs.IsUnknown() {
		return types.StringValue("")
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
	return types.StringValue("")
}

func (m agentlessPolicyModel) GetKibanaConnection() types.List { return m.KibanaConnection }

// GetVersionRequirements enforces the minimum Kibana version for the Fleet
// agentless policies API (experimental, added in Kibana 9.3.0). See
// specs/fleet-agentless-policy/spec.md, requirement "Version gating".
func (m agentlessPolicyModel) GetVersionRequirements(_ context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return []entitycore.VersionRequirement{
		{
			MinVersion:   *MinVersion,
			ErrorMessage: fmt.Sprintf("Fleet agentless policies require Elastic Stack v%s or later (experimental API).", MinVersion),
		},
	}, nil
}
