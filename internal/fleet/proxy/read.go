package proxy

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	fleetutils "github.com/elastic/terraform-provider-elasticstack/internal/fleet"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *proxyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var stateModel proxyModel

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

	proxyID := stateModel.ProxyID.ValueString()

	// Read the existing spaces from state to determine where to query
	spaceID, diags := fleetutils.GetOperationalSpaceFromState(ctx, req.State)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Query using the operational space from STATE
	proxy, diags := fleet.GetFleetProxy(ctx, client, proxyID, spaceID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if proxy == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Preserve space_ids from state, as the Fleet API does not return them
	spaceIDs := stateModel.SpaceIds

	diags = stateModel.populateFromAPI(ctx, proxy)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Restore space_ids so they are not lost on refresh
	stateModel.SpaceIds = spaceIDs

	diags = resp.State.Set(ctx, stateModel)
	resp.Diagnostics.Append(diags...)
}
