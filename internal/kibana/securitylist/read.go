package securitylist

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *securityListResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Kibana client", err.Error())
		return
	}

	// Parse composite ID to get space_id and list_id
	spaceID := state.SpaceID.ValueString()
	listID := state.ListID.ValueString()

	// Try to parse as composite ID from state.ID
	if compID, diags := clients.CompositeIDFromStrFw(state.ID.ValueString()); !diags.HasError() {
		spaceID = compID.ClusterID
		listID = compID.ResourceID
		// Update space_id in state if it was parsed from composite ID
		state.SpaceID = types.StringValue(spaceID)
	}

	params := &kbapi.ReadListParams{
		Id: listID,
	}

	list, diags := kibanaoapi.GetList(ctx, client, spaceID, params)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if list == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Convert API response to model
	diags = state.fromAPI(ctx, list)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
