package security_list_item

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *securityListItemResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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
	createReq, diags := plan.toAPICreateModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the list item
	createdListItem, diags := kibana_oapi.CreateListItem(ctx, client, plan.SpaceID.ValueString(), *createReq)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if createdListItem == nil {
		resp.Diagnostics.AddError("Failed to create security list item", "API returned empty response")
		return
	}

	// Read the created list item to populate state
	id := kbapi.SecurityListsAPIListId(createdListItem.Id)
	readParams := &kbapi.ReadListItemParams{
		Id: &id,
	}

	listItem, diags := kibana_oapi.GetListItem(ctx, client, plan.SpaceID.ValueString(), readParams)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if listItem == nil {
		resp.State.RemoveResource(ctx)
		resp.Diagnostics.AddError("Failed to fetch security list item", "API returned empty response")
		return
	}

	// Update state with read response
	diags = plan.fromAPIModel(ctx, listItem)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
