package agent_policy

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
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

	sVersion, e := r.client.ServerVersion(ctx)
	if e != nil {
		return
	}

	policyID := stateModel.PolicyID.ValueString()
	policy, diags := fleet.GetAgentPolicy(ctx, client, policyID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if policy == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	stateModel.populateFromAPI(ctx, policy, sVersion)

	resp.State.Set(ctx, stateModel)
	resp.Diagnostics.Append(diags...)
}
