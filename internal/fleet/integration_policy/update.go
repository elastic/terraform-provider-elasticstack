package integration_policy

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	fleetutils "github.com/elastic/terraform-provider-elasticstack/internal/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *integrationPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var planModel integrationPolicyModel
	var stateModel integrationPolicyModel

	diags := req.Plan.Get(ctx, &planModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.State.Get(ctx, &stateModel)
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

	body, diags := planModel.toAPIModel(ctx, true, feat)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	policyID := planModel.PolicyID.ValueString()

	// Read the existing spaces from state to avoid updating in a space where it's not yet visible
	spaceID, diags := fleetutils.GetOperationalSpaceFromState(ctx, req.State)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update using the operational space from STATE
	// The API will handle adding/removing policy from spaces based on space_ids in body
	var policy *kbapi.PackagePolicy
	if spaceID != "" {
		policy, diags = fleet.UpdatePackagePolicyInSpace(ctx, client, policyID, spaceID, body)
	} else {
		policy, diags = fleet.UpdatePackagePolicy(ctx, client, policyID, body)
	}

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = HandleReqRespSecrets(ctx, body, policy, resp.Private)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Remember which agent policy field was originally configured in state
	// so we can preserve it after populateFromAPI
	stateUsedAgentPolicyID := utils.IsKnown(stateModel.AgentPolicyID) && !stateModel.AgentPolicyID.IsNull()
	stateUsedAgentPolicyIDs := utils.IsKnown(stateModel.AgentPolicyIDs) && !stateModel.AgentPolicyIDs.IsNull()

	// Remember the input configuration from state
	stateHadInput := utils.IsKnown(stateModel.Input) && !stateModel.Input.IsNull() && len(stateModel.Input.Elements()) > 0

	diags = planModel.populateFromAPI(ctx, policy)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Restore the agent policy field that was originally configured
	// This prevents populateFromAPI from changing which field is used
	if stateUsedAgentPolicyID && !stateUsedAgentPolicyIDs {
		// Only agent_policy_id was configured, ensure we preserve it
		planModel.AgentPolicyID = types.StringPointerValue(policy.PolicyId)
		planModel.AgentPolicyIDs = types.ListNull(types.StringType)
	} else if stateUsedAgentPolicyIDs && !stateUsedAgentPolicyID {
		// Only agent_policy_ids was configured, ensure we preserve it
		if policy.PolicyIds != nil {
			agentPolicyIDs, d := types.ListValueFrom(ctx, types.StringType, *policy.PolicyIds)
			resp.Diagnostics.Append(d...)
			planModel.AgentPolicyIDs = agentPolicyIDs
		} else {
			planModel.AgentPolicyIDs = types.ListNull(types.StringType)
		}
		planModel.AgentPolicyID = types.StringNull()
	}

	// If state didn't have input configured, ensure we don't add it now
	if !stateHadInput && (planModel.Input.IsNull() || len(planModel.Input.Elements()) == 0) {
		planModel.Input = types.ListNull(getInputTypeV1())
	}

	diags = resp.State.Set(ctx, planModel)
	resp.Diagnostics.Append(diags...)
}
