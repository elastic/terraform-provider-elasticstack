package agent_policy

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
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

	diags = fleet.DeleteAgentPolicy(ctx, client, policyID)
	resp.Diagnostics.Append(diags...)
}
