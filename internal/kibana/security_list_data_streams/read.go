package security_list_data_streams

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *securityListDataStreamsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SecurityListDataStreamsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// During import, space_id might not be set yet, derive it from ID
	if state.SpaceID.IsNull() || state.SpaceID.IsUnknown() {
		if !state.ID.IsNull() && !state.ID.IsUnknown() {
			state.SpaceID = state.ID
		}
	}

	spaceID := state.SpaceID.ValueString()

	// Get Kibana client
	client, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Kibana client", err.Error())
		return
	}

	// Check if the data streams exist
	exists, diags := kibana_oapi.ReadListIndex(ctx, client, spaceID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !exists {
		// Data streams don't exist, remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	// Data streams exist, update state
	state.ID = types.StringValue(spaceID)
	state.Acknowledged = types.BoolValue(true)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
