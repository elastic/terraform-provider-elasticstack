package output

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	fleetutils "github.com/elastic/terraform-provider-elasticstack/internal/fleet"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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

	// Extract space IDs from state and determine operational space
	// Using default-space-first model: always prefer "default" if present
	// This prevents resource orphaning when space_ids is reordered
	spaceIDs := fleetutils.ExtractSpaceIDs(ctx, stateModel.SpaceIds)
	spaceID := fleetutils.GetOperationalSpace(spaceIDs)

	// Query using the operational space
	var output *kbapi.OutputUnion
	if spaceID != nil && *spaceID != "" {
		output, diags = fleet.GetOutputInSpace(ctx, client, outputID, *spaceID)
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
