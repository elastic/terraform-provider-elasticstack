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

package agentpolicy

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	fleetutils "github.com/elastic/terraform-provider-elasticstack/internal/fleet"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *agentPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var planModel agentPolicyModel

	diags := req.Plan.Get(ctx, &planModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.GetFleetClient()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	feat, diags := r.buildFeatures(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	policyID := planModel.PolicyID.ValueString()

	// Read the existing spaces from state to avoid updating the policy
	// in a space where it's not yet visible.
	// This prevents errors when prepending a new space to space_ids:
	// e.g., ["space-a"] → ["space-b", "space-a"] would fail if we queried "space-b"
	// because the policy doesn't exist there yet.
	spaceID, diags := fleetutils.GetOperationalSpaceFromState(ctx, req.State)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read current policy to get existing AgentFeatures (so we can preserve other features)
	currentPolicy, diags := fleet.GetAgentPolicy(ctx, client, policyID, spaceID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var existingFeatures []apiAgentFeature
	if currentPolicy != nil && currentPolicy.AgentFeatures != nil {
		existingFeatures = *currentPolicy.AgentFeatures
	}

	body, diags := planModel.toAPIUpdateModel(ctx, feat, existingFeatures)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update using the operational space from STATE
	// The API will handle adding/removing the policy from spaces based on space_ids in body
	policy, diags := fleet.UpdateAgentPolicy(ctx, client, policyID, spaceID, body)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Populate from API response
	// With Sets, we don't need order preservation - Terraform handles set comparison automatically
	planModel.populateFromAPI(ctx, policy)

	diags = resp.State.Set(ctx, planModel)
	resp.Diagnostics.Append(diags...)
}
