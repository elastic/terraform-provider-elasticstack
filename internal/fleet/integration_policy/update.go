package integration_policy

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	fleetutils "github.com/elastic/terraform-provider-elasticstack/internal/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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

	// Remember if the user had inputs configured in state before calling populateFromAPI
	// This prevents populateFromAPI from adding inputs when the user didn't configure them
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

	// If state didn't have inputs configured, ensure we don't add them now
	// This prevents "Provider produced inconsistent result" errors where
	// inputs block count changes from 0 to 1 unexpectedly
	if !stateHadInput {
		planModel.Inputs = NewInputsNull(getInputsElementType())
	}

	diags = resp.State.Set(ctx, planModel)
	resp.Diagnostics.Append(diags...)
}
