package security_list_data_streams

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *securityListDataStreamsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SecurityListDataStreamsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// During import, space_id might not be set yet, derive it from ID
	if !utils.IsKnown(state.SpaceID) && utils.IsKnown(state.ID) {
		state.SpaceID = state.ID
	}

	spaceID := state.SpaceID.ValueString()

	// Get Kibana client
	client, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Kibana client", err.Error())
		return
	}

	// Check if the data streams exist
	listIndex, listItemIndex, diags := kibana_oapi.ReadListIndex(ctx, client, spaceID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !listIndex || !listItemIndex {
		// Data streams don't exist, remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	// Data streams exist, update state using the fromAPIResponse helper method
	state.fromAPIResponse(spaceID, listIndex, listItemIndex)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
