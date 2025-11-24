package securitylist

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *securityListResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SecurityListModel
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
	createReq, diags := plan.toCreateRequest()
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the list
	spaceID := plan.SpaceID.ValueString()
	createResp, diags := kibana_oapi.CreateList(ctx, client, spaceID, *createReq)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if createResp == nil || createResp.JSON200 == nil {
		resp.Diagnostics.AddError("Failed to create security list", "API returned empty response")
		return
	}

	// Read the created list to populate state
	readParams := &kbapi.ReadListParams{
		Id: createResp.JSON200.Id,
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
