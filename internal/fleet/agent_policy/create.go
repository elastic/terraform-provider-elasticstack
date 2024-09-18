package agent_policy

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *agentPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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

	body := planModel.toAPICreateModel()
	policy, diags := fleet.CreateAgentPolicy(ctx, client, body)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	planModel.populateFromAPI(policy)

	resp.State.Set(ctx, planModel)
	resp.Diagnostics.Append(diags...)
}
