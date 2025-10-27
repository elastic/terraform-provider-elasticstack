package output

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	fleetutils "github.com/elastic/terraform-provider-elasticstack/internal/fleet"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *outputResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var planModel outputModel
	var stateModel outputModel

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

	body, diags := planModel.toAPIUpdateModel(ctx, r.client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	outputID := planModel.OutputID.ValueString()

	// Extract space IDs from PLAN (where user wants changes) and determine operational space
	// Using default-space-first model: always prefer "default" if present
	// API handles adding/removing output from spaces based on space_ids in body
	planSpaceIDs := fleetutils.ExtractSpaceIDs(ctx, planModel.SpaceIds)
	spaceID := fleetutils.GetOperationalSpace(planSpaceIDs)

	// Update using the operational space
	var output *kbapi.OutputUnion
	if spaceID != nil && *spaceID != "" {
		output, diags = fleet.UpdateOutputInSpace(ctx, client, outputID, *spaceID, body)
	} else {
		output, diags = fleet.UpdateOutput(ctx, client, outputID, body)
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = planModel.populateFromAPI(ctx, output)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve space_ids from state after populateFromAPI
	// The API doesn't return space_ids, so we need to restore it from state
	if planModel.SpaceIds.IsNull() || planModel.SpaceIds.IsUnknown() {
		planModel.SpaceIds = stateModel.SpaceIds
	}

	diags = resp.State.Set(ctx, planModel)
	resp.Diagnostics.Append(diags...)
}
