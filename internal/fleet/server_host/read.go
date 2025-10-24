package server_host

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

func (r *serverHostResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var stateModel serverHostModel

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

	hostID := stateModel.HostID.ValueString()

	// If space_ids is set in state, use space-aware GET request
	var spaceID string
	if !stateModel.SpaceIds.IsNull() && !stateModel.SpaceIds.IsUnknown() {
		var tempDiags diag.Diagnostics
		spaceIDs := utils.ListTypeAs[types.String](ctx, stateModel.SpaceIds, path.Root("space_ids"), &tempDiags)
		if !tempDiags.HasError() && len(spaceIDs) > 0 {
			spaceID = spaceIDs[0].ValueString()
		}
	}

	var host *kbapi.ServerHost
	if spaceID != "" && spaceID != "default" {
		host, diags = fleet.GetFleetServerHostInSpace(ctx, client, hostID, spaceID)
	} else {
		host, diags = fleet.GetFleetServerHost(ctx, client, hostID)
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if host == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	diags = stateModel.populateFromAPI(ctx, host)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, stateModel)
	resp.Diagnostics.Append(diags...)
}
