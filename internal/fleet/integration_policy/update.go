package integration_policy

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	fleetutils "github.com/elastic/terraform-provider-elasticstack/internal/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/integration_policy/models"
	v2 "github.com/elastic/terraform-provider-elasticstack/internal/fleet/integration_policy/models/v2"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *integrationPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var planModel v2.IntegrationPolicyModel
	var stateModel v2.IntegrationPolicyModel

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

	feat, diags := models.NewFeatures(ctx, r.client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, diags := planModel.ToAPIModel(ctx, true, feat)
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
	policy, diags := fleet.UpdatePackagePolicy(ctx, client, policyID, spaceID, body)

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
	stateHadInput := utils.IsKnown(stateModel.Inputs) && !stateModel.Inputs.IsNull() && len(stateModel.Inputs.Elements()) > 0

	diags = planModel.PopulateFromAPI(ctx, policy)
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
	if !stateHadInput && (planModel.Inputs.IsNull() || len(planModel.Inputs.Elements()) == 0) {
		planModel.Inputs = v2.NewInputsNull(v2.GetInputsElementType())
	}

	diags = resp.State.Set(ctx, planModel)
	resp.Diagnostics.Append(diags...)
}
