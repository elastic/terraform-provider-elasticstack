package output

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	fleetutils "github.com/elastic/terraform-provider-elasticstack/internal/fleet"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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

	// Extract space IDs from STATE and determine operational space
	// Using default-space-first model: always prefer "default" if present
	stateSpaceIDs := fleetutils.ExtractSpaceIDs(ctx, stateModel.SpaceIds)
	spaceID := fleetutils.GetOperationalSpace(stateSpaceIDs)

	// NOTE: DELETE removes the output from ALL spaces (global delete)
	// To remove from specific spaces only, UPDATE space_ids instead
	if spaceID != nil && *spaceID != "" {
		diags = fleet.DeleteOutputInSpace(ctx, client, outputID, *spaceID)
	} else {
		diags = fleet.DeleteOutput(ctx, client, outputID)
	}
	resp.Diagnostics.Append(diags...)
}
