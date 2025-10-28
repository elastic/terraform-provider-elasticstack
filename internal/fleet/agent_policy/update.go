package agent_policy

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
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

	body, diags := planModel.toAPIUpdateModel(ctx, feat)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	policyID := planModel.PolicyID.ValueString()

	// Extract space IDs from plan and determine operational space
	// Using default-space-first model for stable multi-space updates
	planSpaceIDs := fleetutils.ExtractSpaceIDs(ctx, planModel.SpaceIds)
	spaceID := fleetutils.GetOperationalSpace(planSpaceIDs)

	// Update using the operational space
	// The API will handle adding/removing the policy from spaces based on space_ids in body
	var policy *kbapi.AgentPolicy
	if spaceID != nil && *spaceID != "" {
		policy, diags = fleet.UpdateAgentPolicyInSpace(ctx, client, policyID, *spaceID, body)
	} else {
		policy, diags = fleet.UpdateAgentPolicy(ctx, client, policyID, body)
	}

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve the space_ids order from plan before populating from API
	// The Kibana API may return space_ids in a different order (sorted),
	// but we need to maintain the user's configured order to avoid false drift detection
	originalSpaceIds := planModel.SpaceIds

	planModel.populateFromAPI(ctx, policy)

	// Restore the original space_ids order from plan if the API returned spaces
	// We only need to verify that the set of spaces matches, not the order
	if policy.SpaceIds != nil && !originalSpaceIds.IsNull() {
		planModel.SpaceIds = originalSpaceIds
	}

	diags = resp.State.Set(ctx, planModel)
	resp.Diagnostics.Append(diags...)
}
