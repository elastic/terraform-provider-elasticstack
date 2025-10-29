package agent_policy

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
		// If space_ids is set, we need to use a space-aware GET request because the policy
		// exists within that space context, not in the default space.
		var readPolicy *kbapi.AgentPolicy
		var getDiags diag.Diagnostics

		if !planModel.SpaceIds.IsNull() && !planModel.SpaceIds.IsUnknown() {
			var tempDiags diag.Diagnostics
			spaceIDs := utils.SetTypeAs[types.String](ctx, planModel.SpaceIds, path.Root("space_ids"), &tempDiags)
			if !tempDiags.HasError() && len(spaceIDs) > 0 {
				// Use the first space for the GET request
				spaceID := spaceIDs[0].ValueString()
				readPolicy, getDiags = fleet.GetAgentPolicyInSpace(ctx, client, policy.Id, spaceID)
			} else {
				// Fall back to standard GET if we couldn't extract space IDs
				readPolicy, getDiags = fleet.GetAgentPolicy(ctx, client, policy.Id)
			}
		} else {
			// No space_ids, use standard GET
			readPolicy, getDiags = fleet.GetAgentPolicy(ctx, client, policy.Id)
		}

		resp.Diagnostics.Append(getDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		// Use the read response if available, otherwise fall back to create response
		if readPolicy != nil {
			policy = readPolicy
		}
	}

	// Populate from API response
	// With Sets, we don't need order preservation - Terraform handles set comparison automatically
	diags = planModel.populateFromAPI(ctx, policy)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.Set(ctx, planModel)
	resp.Diagnostics.Append(diags...)
}
