package output

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *outputResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var stateModel outputModel

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

	outputID := stateModel.OutputID.ValueString()

	// If space_ids is set, use space-aware DELETE request
	var spaceID string
	if !stateModel.SpaceIds.IsNull() && !stateModel.SpaceIds.IsUnknown() {
		var tempDiags diag.Diagnostics
		spaceIDs := utils.ListTypeAs[types.String](ctx, stateModel.SpaceIds, path.Root("space_ids"), &tempDiags)
		if !tempDiags.HasError() && len(spaceIDs) > 0 {
			spaceID = spaceIDs[0].ValueString()
		}
	}

	if spaceID != "" && spaceID != "default" {
		diags = fleet.DeleteOutputInSpace(ctx, client, outputID, spaceID)
	} else {
		diags = fleet.DeleteOutput(ctx, client, outputID)
	}
	resp.Diagnostics.Append(diags...)
}
