package securitylistitem

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *securityListItemResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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

	// Delete by ID
	id := kbapi.SecurityListsAPIListItemId(state.ID.ValueString())
	params := &kbapi.DeleteListItemParams{
		Id: &id,
	}

	diags := kibana_oapi.DeleteListItem(ctx, client, params)
	resp.Diagnostics.Append(diags...)
}
