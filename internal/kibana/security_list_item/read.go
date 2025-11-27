package security_list_item

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *securityListItemResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SecurityListItemModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get Kibana client
	client, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Kibana client", err.Error())
		return
	}

	// Parse composite ID to get space_id and resource id
	compId, compIdDiags := clients.CompositeIdFromStrFw(state.ID.ValueString())

	if !compIdDiags.HasError() {
		state.SpaceID = types.StringValue(compId.ClusterId)
	}

	resp.Diagnostics.Append(compIdDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read by resource ID from composite ID
	id := kbapi.SecurityListsAPIListId(compId.ResourceId)
	params := &kbapi.ReadListItemParams{
		Id: &id,
	}

	readResp, diags := kibana_oapi.GetListItem(ctx, client, compId.ClusterId, params)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if readResp == nil || readResp.JSON200 == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// The response can be a single item or an array, so we need to unmarshal from the body
	// When querying by ID, we expect a single item
	var listItem kbapi.SecurityListsAPIListItem
	if err := json.Unmarshal(readResp.Body, &listItem); err != nil {
		resp.Diagnostics.AddError("Failed to parse list item response", err.Error())
		return
	}

	// Update state with response
	diags = state.fromAPIModel(ctx, &listItem)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
