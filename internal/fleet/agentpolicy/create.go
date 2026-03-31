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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *agentPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var planModel agentPolicyModel

	diags := req.Plan.Get(ctx, &planModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	planWantsTamperProtection := planModel.IsProtected

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

	body, diags := planModel.toAPICreateModel(ctx, feat)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var spaceID string
	if typeutils.IsKnown(planModel.SpaceIDs) {
		var tempDiags diag.Diagnostics
		spaceIDs := typeutils.SetTypeAs[types.String](ctx, planModel.SpaceIDs, path.Root("space_ids"), &tempDiags)
		if !tempDiags.HasError() && len(spaceIDs) > 0 {
			// Use the first space for the POST request
			spaceID = spaceIDs[0].ValueString()
		}
	}

	sysMonitoring := planModel.SysMonitoring.ValueBool()
	policy, diags := fleet.CreateAgentPolicy(ctx, client, body, sysMonitoring, spaceID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The CREATE response may not include all fields (e.g., space_ids can be null in the response
	// even when specified in the request). Read the policy back to get the complete state.
	// Only do this if we got a valid ID from the create response.
	if policy != nil && policy.Id != "" {
		// If space_ids is set, we need to use a space-aware GET request because the policy
		// exists within that space context, not in the default space.
		var readPolicy *kbapi.AgentPolicy
		var getDiags diag.Diagnostics

		readPolicy, getDiags = fleet.GetAgentPolicy(ctx, client, policy.Id, spaceID)

		resp.Diagnostics.Append(getDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		// Use the read response if available, otherwise fall back to create response
		if readPolicy != nil {
			policy = readPolicy
		}
	}

	// POST /api/fleet/agent_policies may not persist is_protected; a follow-up PUT applies it.
	if policy != nil && typeutils.IsKnown(planWantsTamperProtection) && planWantsTamperProtection.ValueBool() &&
		feat.SupportsTamperProtection && !policy.IsProtected {
		existingFeatures := agentFeaturesFromPolicy(policy)
		updateBody, updateDiags := planModel.toAPIUpdateModel(ctx, feat, existingFeatures)
		resp.Diagnostics.Append(updateDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		updated, updateDiags := fleet.UpdateAgentPolicy(ctx, client, policy.Id, spaceID, updateBody)
		resp.Diagnostics.Append(updateDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		if updated != nil {
			policy = updated
		}
	}

	// Populate from API response
	// With Sets, we don't need order preservation - Terraform handles set comparison automatically
	diags = planModel.populateFromAPI(ctx, policy)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if typeutils.IsKnown(planWantsTamperProtection) && planWantsTamperProtection.ValueBool() &&
		typeutils.IsKnown(planModel.IsProtected) && !planModel.IsProtected.ValueBool() {
		resp.Diagnostics.AddError(
			"Fleet API did not enable tamper protection",
			"The agent policy was saved but is_protected is still false. Ensure Elastic Stack 8.10.0 or later, that your license allows tamper protection, and that the Fleet API accepts is_protected on this deployment.",
		)
		return
	}

	resp.State.Set(ctx, planModel)
	resp.Diagnostics.Append(diags...)
}
