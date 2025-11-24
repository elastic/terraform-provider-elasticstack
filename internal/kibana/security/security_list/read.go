package securitylist

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *securityListResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SecurityListModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Kibana client", err.Error())
		return
	}

	spaceID := state.SpaceID.ValueString()
	listID := state.ListID.ValueString()

	params := &kbapi.ReadListParams{
		Id: kbapi.SecurityListsAPIListId(listID),
	}

	readResp, diags := kibana_oapi.GetList(ctx, client, spaceID, params)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if readResp == nil || readResp.JSON200 == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Convert API response to model
	diags = state.fromAPI(ctx, readResp.JSON200)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
