package integration_policy

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	fleetutils "github.com/elastic/terraform-provider-elasticstack/internal/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *integrationPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var stateModel integrationPolicyModel

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

	policyID := stateModel.PolicyID.ValueString()

	// Read the existing spaces from state to determine where to query
	spaceID, diags := fleetutils.GetOperationalSpaceFromState(ctx, req.State)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Query using the operational space from STATE
	policy, diags := fleet.GetPackagePolicy(ctx, client, policyID, spaceID)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if policy == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	diags = HandleRespSecrets(ctx, policy, resp.Private)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Remember if the state had input configured
	stateHadInput := utils.IsKnown(stateModel.Inputs) && !stateModel.Inputs.IsNull() && len(stateModel.Inputs.Elements()) > 0

	// Check if this is an import operation (PolicyID is the only field set)
	isImport := stateModel.PolicyID.ValueString() != "" &&
		(stateModel.Name.IsNull() || stateModel.Name.IsUnknown())

	pkg, diags := getPackageInfo(ctx, client, policy.Package.Name, policy.Package.Version)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = stateModel.populateFromAPI(ctx, pkg, policy)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If state didn't have input configured and this is not an import, ensure we don't add it now
	// This prevents "Provider produced inconsistent result" errors during refresh
	// However, during import we should always populate inputs from the API
	if !stateHadInput && !isImport {
		stateModel.Inputs = NewInputsNull(getInputsElementType())
	}

	diags = resp.State.Set(ctx, stateModel)
	resp.Diagnostics.Append(diags...)
}
