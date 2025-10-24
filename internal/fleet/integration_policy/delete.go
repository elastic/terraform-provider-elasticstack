package integration_policy

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *integrationPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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
	force := stateModel.Force.ValueBool()

	// If space_ids is set in state, use space-aware DELETE request
	if !stateModel.SpaceIds.IsNull() && !stateModel.SpaceIds.IsUnknown() {
		var tempDiags diag.Diagnostics
		spaceIDs := utils.ListTypeAs[types.String](ctx, stateModel.SpaceIds, path.Root("space_ids"), &tempDiags)
		if !tempDiags.HasError() && len(spaceIDs) > 0 {
			spaceID := spaceIDs[0].ValueString()
			diags = fleet.DeletePackagePolicyInSpace(ctx, client, policyID, spaceID, force)
		} else {
			diags = fleet.DeletePackagePolicy(ctx, client, policyID, force)
		}
	} else {
		diags = fleet.DeletePackagePolicy(ctx, client, policyID, force)
	}

	resp.Diagnostics.Append(diags...)
}
