package securitylist

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *securityListResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SecurityListModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state SecurityListModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve version_id from state for optimistic locking
	if state.VersionID.ValueString() != "" {
		plan.VersionID = state.VersionID
	}

	// Convert plan to API request
	updateReq, diags := plan.toUpdateRequest()
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get Kibana client
	client, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Kibana client", err.Error())
		return
	}

	// Update the list
	spaceID := plan.SpaceID.ValueString()
	updateResp, diags := kibana_oapi.UpdateList(ctx, client, spaceID, *updateReq)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if updateResp == nil || updateResp.JSON200 == nil {
		resp.Diagnostics.AddError("Failed to update security list", "API returned empty response")
		return
	}

	// Read the updated list to populate state
	readParams := &kbapi.ReadListParams{
		Id: updateResp.JSON200.Id,
	}

	readResp, diags := kibana_oapi.GetList(ctx, client, spaceID, readParams)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if readResp == nil || readResp.JSON200 == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Update state with read response
	diags = plan.fromAPI(ctx, readResp.JSON200)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
