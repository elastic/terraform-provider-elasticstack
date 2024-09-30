package agent_policy

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
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

	body := planModel.toAPIUpdateModel()

	policyID := planModel.PolicyID.ValueString()
	policy, diags := fleet.UpdateAgentPolicy(ctx, client, policyID, body)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	planModel.populateFromAPI(policy)

	diags = resp.State.Set(ctx, planModel)
	resp.Diagnostics.Append(diags...)
}
