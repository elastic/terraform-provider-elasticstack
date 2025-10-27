package server_host

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	fleetutils "github.com/elastic/terraform-provider-elasticstack/internal/fleet"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *serverHostResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var planModel serverHostModel

	diags := req.Plan.Get(ctx, &planModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.GetFleetClient()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	hostID := planModel.HostID.ValueString()
	body, diags := planModel.toAPIUpdateModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract space IDs from PLAN (where user wants changes) and determine operational space
	// Using default-space-first model: always prefer "default" if present
	// API handles adding/removing server host from spaces based on space_ids in body
	planSpaceIDs := fleetutils.ExtractSpaceIDs(ctx, planModel.SpaceIds)
	spaceID := fleetutils.GetOperationalSpace(planSpaceIDs)

	// Update using the operational space
	var host *kbapi.ServerHost
	if spaceID != nil && *spaceID != "" {
		host, diags = fleet.UpdateFleetServerHostInSpace(ctx, client, hostID, *spaceID, body)
	} else {
		host, diags = fleet.UpdateFleetServerHost(ctx, client, hostID, body)
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = planModel.populateFromAPI(ctx, host)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, planModel)
	resp.Diagnostics.Append(diags...)
}
