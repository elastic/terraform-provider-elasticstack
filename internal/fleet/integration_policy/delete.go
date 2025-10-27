package integration_policy

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	fleetutils "github.com/elastic/terraform-provider-elasticstack/internal/fleet"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *integrationPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var stateModel integrationPolicyModel

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
	force := stateModel.Force.ValueBool()

	// Extract space IDs from STATE and determine operational space
	// Using default-space-first model: always prefer "default" if present
	stateSpaceIDs := fleetutils.ExtractSpaceIDs(ctx, stateModel.SpaceIds)
	spaceID := fleetutils.GetOperationalSpace(stateSpaceIDs)

	// NOTE: DELETE removes the policy from ALL spaces (global delete)
	// To remove from specific spaces only, UPDATE space_ids instead
	if spaceID != nil && *spaceID != "" {
		diags = fleet.DeletePackagePolicyInSpace(ctx, client, policyID, *spaceID, force)
	} else {
		diags = fleet.DeletePackagePolicy(ctx, client, policyID, force)
	}

	resp.Diagnostics.Append(diags...)
}
