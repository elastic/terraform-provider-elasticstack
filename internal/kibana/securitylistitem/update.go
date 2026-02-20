package securitylistitem

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *securityListItemResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan Model
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

	// Parse composite ID to get space_id
	compID, compIDDiags := clients.CompositeIDFromStrFw(plan.ID.ValueString())
	resp.Diagnostics.Append(compIDDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert plan to API request
	updateReq, diags := plan.toAPIUpdateModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the list item
	updatedListItem, diags := kibanaoapi.UpdateListItem(ctx, client, compID.ClusterID, *updateReq)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if updatedListItem == nil {
		resp.Diagnostics.AddError("Failed to update security list item", "API returned empty response")
		return
	}

	// Read the updated list item to populate state
	id := updatedListItem.Id
	readParams := &kbapi.ReadListItemParams{
		Id: &id,
	}

	listItem, diags := kibanaoapi.GetListItem(ctx, client, compID.ClusterID, readParams)
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
