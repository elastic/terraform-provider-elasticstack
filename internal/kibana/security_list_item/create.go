package security_list_item

import (
	"context"
	"encoding/json"

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
	createResp, diags := kibana_oapi.CreateListItem(ctx, client, *createReq)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if createResp == nil || createResp.JSON200 == nil {
		resp.Diagnostics.AddError("Failed to create security list item", "API returned empty response")
		return
	}

	// Read the created list item to populate state
	id := kbapi.SecurityListsAPIListId(createResp.JSON200.Id)
	readParams := &kbapi.ReadListItemParams{
		Id: &id,
	}

	readResp, diags := kibana_oapi.GetListItem(ctx, client, readParams)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if readResp == nil || readResp.JSON200 == nil {
		resp.State.RemoveResource(ctx)
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
