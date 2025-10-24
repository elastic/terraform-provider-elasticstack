package output

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

func (r *outputResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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

	// If space_ids is set in state, use space-aware GET request
	var spaceID string
	if !stateModel.SpaceIds.IsNull() && !stateModel.SpaceIds.IsUnknown() {
		var tempDiags diag.Diagnostics
		spaceIDs := utils.ListTypeAs[types.String](ctx, stateModel.SpaceIds, path.Root("space_ids"), &tempDiags)
		if !tempDiags.HasError() && len(spaceIDs) > 0 {
			spaceID = spaceIDs[0].ValueString()
		}
	}

	var output *kbapi.OutputUnion
	if spaceID != "" && spaceID != "default" {
		output, diags = fleet.GetOutputInSpace(ctx, client, outputID, spaceID)
	} else {
		output, diags = fleet.GetOutput(ctx, client, outputID)
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		resp.State.RemoveResource(ctx)
		return
	}

	if output == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	diags = stateModel.populateFromAPI(ctx, output)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, stateModel)
	resp.Diagnostics.Append(diags...)
}
