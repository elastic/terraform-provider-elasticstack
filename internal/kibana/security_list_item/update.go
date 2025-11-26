package security_list_item

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *securityListItemResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SecurityListItemModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get Kibana client
	client, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Kibana client", err.Error())
		return
	}

	// Convert plan to API request
	updateReq, diags := plan.toAPIUpdateModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the list item
	updateResp, diags := kibana_oapi.UpdateListItem(ctx, client, plan.SpaceID.ValueString(), *updateReq)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if updateResp == nil || updateResp.JSON200 == nil {
		resp.Diagnostics.AddError("Failed to update security list item", "API returned empty response")
		return
	}

	// Read the updated list item to populate state
	id := kbapi.SecurityListsAPIListId(updateResp.JSON200.Id)
	readParams := &kbapi.ReadListItemParams{
		Id: &id,
	}

	readResp, diags := kibana_oapi.GetListItem(ctx, client, plan.SpaceID.ValueString(), readParams)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if readResp == nil || readResp.JSON200 == nil {
		resp.State.RemoveResource(ctx)
		resp.Diagnostics.AddError("Failed to fetch security list item", "API returned empty response")
		return
	}

	// Unmarshal the response body to get the list item
	var listItem kbapi.SecurityListsAPIListItem
	if err := json.Unmarshal(readResp.Body, &listItem); err != nil {
		resp.Diagnostics.AddError("Failed to parse list item response", err.Error())
		return
	}

	// Update state with read response
	diags = plan.fromAPIModel(ctx, &listItem)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
