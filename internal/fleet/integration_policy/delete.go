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

	// Read the existing spaces from state to determine where to delete
	// NOTE: DELETE removes the policy from ALL spaces (global delete)
	// To remove from specific spaces only, UPDATE space_ids instead
	spaceID, diags := fleetutils.GetOperationalSpaceFromState(ctx, req.State)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete using the operational space from STATE
	if spaceID != "" {
		diags = fleet.DeletePackagePolicyInSpace(ctx, client, policyID, spaceID, force)
	} else {
		diags = fleet.DeletePackagePolicy(ctx, client, policyID, force)
	}

	resp.Diagnostics.Append(diags...)
}
