package agent_policy

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	fleetutils "github.com/elastic/terraform-provider-elasticstack/internal/fleet"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *agentPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var stateModel agentPolicyModel

	diags := req.State.Get(ctx, &stateModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.GetFleetClient()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	policyID := stateModel.PolicyID.ValueString()

	// Extract space IDs from state and determine operational space
	// Using default-space-first model: always prefer "default" if present
	// This prevents resource orphaning when space_ids is reordered
	spaceIDs := fleetutils.ExtractSpaceIDs(ctx, stateModel.SpaceIds)
	spaceID := fleetutils.GetOperationalSpace(spaceIDs)

	// Query using the operational space
	var policy *kbapi.AgentPolicy
	if spaceID != nil && *spaceID != "" {
		policy, diags = fleet.GetAgentPolicyInSpace(ctx, client, policyID, *spaceID)
	} else {
		policy, diags = fleet.GetAgentPolicy(ctx, client, policyID)
	}

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if policy == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	diags = stateModel.populateFromAPI(ctx, policy)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.Set(ctx, stateModel)
	resp.Diagnostics.Append(diags...)
}
