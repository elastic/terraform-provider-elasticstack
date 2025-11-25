package security_list

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *securityListResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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

	params := &kbapi.DeleteListParams{
		Id: kbapi.SecurityListsAPIListId(listID),
	}

	diags := kibana_oapi.DeleteList(ctx, client, spaceID, params)
	resp.Diagnostics.Append(diags...)
}
