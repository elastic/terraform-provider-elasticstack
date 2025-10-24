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

	feat, diags := r.buildFeatures(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, diags := planModel.toAPICreateModel(ctx, feat)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sysMonitoring := planModel.SysMonitoring.ValueBool()
	policy, diags := fleet.CreateAgentPolicy(ctx, client, body, sysMonitoring)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The CREATE response may not include all fields (e.g., space_ids can be null in the response
	// even when specified in the request). Read the policy back to get the complete state.
	// Only do this if we got a valid ID from the create response.
	if policy != nil && policy.Id != "" {
		readPolicy, diags := fleet.GetAgentPolicy(ctx, client, policy.Id)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		// Use the read response if available, otherwise fall back to create response
		if readPolicy != nil {
			policy = readPolicy
		}
	}

	diags = planModel.populateFromAPI(ctx, policy)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.Set(ctx, planModel)
	resp.Diagnostics.Append(diags...)
}
