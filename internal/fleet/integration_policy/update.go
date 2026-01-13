package integration_policy

import (
	"context"

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

	// Preserve computed fields from state before building the API request
	// This ensures fields like agent_policy_id are included in the update request
	// even when they're not explicitly changed by the user
	if utils.IsKnown(stateModel.ID) {
		planModel.ID = stateModel.ID
	}
	if utils.IsKnown(stateModel.PolicyID) {
		planModel.PolicyID = stateModel.PolicyID
	}
	// Only preserve optional fields when plan doesn't have a value (user didn't change them)
	// This prevents overwriting user changes while still fixing the null bug
	if !utils.IsKnown(planModel.AgentPolicyID) && utils.IsKnown(stateModel.AgentPolicyID) {
		planModel.AgentPolicyID = stateModel.AgentPolicyID
	}
	if !utils.IsKnown(planModel.AgentPolicyIDs) && utils.IsKnown(stateModel.AgentPolicyIDs) {
		planModel.AgentPolicyIDs = stateModel.AgentPolicyIDs
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

	body, diags := planModel.toAPIModel(ctx, feat)
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
	stateUsedAgentPolicyID := utils.IsKnown(stateModel.AgentPolicyID)
	stateUsedAgentPolicyIDs := utils.IsKnown(stateModel.AgentPolicyIDs)

	// Remember the input configuration from state
	stateHadInput := utils.IsKnown(stateModel.Inputs) && !stateModel.Inputs.IsNull() && len(stateModel.Inputs.Elements()) > 0

	pkg, diags := getPackageInfo(ctx, client, policy.Package.Name, policy.Package.Version)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = planModel.populateFromAPI(ctx, pkg, policy)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Restore the agent policy field that was originally configured
	// This prevents populateFromAPI from changing which field is used
	// IMPORTANT: Use state values, not API response, to avoid null values causing inconsistent state
	if stateUsedAgentPolicyID && !stateUsedAgentPolicyIDs {
		// Only agent_policy_id was configured, ensure we preserve it from state
		planModel.AgentPolicyID = stateModel.AgentPolicyID
		planModel.AgentPolicyIDs = types.ListNull(types.StringType)
	} else if stateUsedAgentPolicyIDs && !stateUsedAgentPolicyID {
		// Only agent_policy_ids was configured, ensure we preserve it from state
		planModel.AgentPolicyIDs = stateModel.AgentPolicyIDs
		planModel.AgentPolicyID = types.StringNull()
	}

	// If state didn't have input configured, ensure we don't add it now
	// IMPORTANT: Always set to null if state didn't have it, even if populateFromAPI added it
	if !stateHadInput {
		planModel.Inputs = NewInputsNull(getInputsElementType())
	}

	diags = resp.State.Set(ctx, planModel)
	resp.Diagnostics.Append(diags...)
}
