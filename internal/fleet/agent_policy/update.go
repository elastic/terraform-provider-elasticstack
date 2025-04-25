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

	sVersion, e := r.client.ServerVersion(ctx)
	if e != nil {
		for _, a := range e {
			resp.Diagnostics.AddError(a.Summary, a.Detail)
		}
		return
	}

	body, diags := planModel.toAPIUpdateModel(ctx, sVersion)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	policyID := planModel.PolicyID.ValueString()
	policy, diags := fleet.UpdateAgentPolicy(ctx, client, policyID, body)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	planModel.populateFromAPI(ctx, policy)

	diags = resp.State.Set(ctx, planModel)
	resp.Diagnostics.Append(diags...)
}
