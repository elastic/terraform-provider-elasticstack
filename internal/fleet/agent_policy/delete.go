package agent_policy

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	fleetutils "github.com/elastic/terraform-provider-elasticstack/internal/fleet"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *agentPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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
	skipDestroy := stateModel.SkipDestroy.ValueBool()
	if skipDestroy {
		tflog.Debug(ctx, "Skipping destroy of Agent Policy", map[string]any{"policy_id": policyID})
		return
	}

	// Read the existing spaces from state to determine where to delete
	// NOTE: DELETE removes the policy from ALL spaces (global delete)
	// To remove from specific spaces only, UPDATE space_ids instead of deleting
	spaceID, diags := fleetutils.GetOperationalSpaceFromState(ctx, req.State)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete using the operational space from STATE
	if spaceID != "" {
		diags = fleet.DeleteAgentPolicyInSpace(ctx, client, policyID, spaceID)
	} else {
		diags = fleet.DeleteAgentPolicy(ctx, client, policyID)
	}

	resp.Diagnostics.Append(diags...)
}
