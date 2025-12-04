package dashboard

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var stateModel dashboardModel

	diags := req.State.Get(ctx, &stateModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse composite ID
	composite, diags := clients.CompositeIdFromStrFw(stateModel.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	dashboardID := composite.ResourceId
	spaceID := composite.ClusterId

	// Get the Kibana client
	kibanaClient, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError("Unable to get Kibana client", err.Error())
		return
	}

	// Get the dashboard
	getResp, diags := kibana_oapi.GetDashboard(ctx, kibanaClient, spaceID, dashboardID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if getResp == nil {
		// Dashboard not found, remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	// Populate the model from the API response
	if getResp.JSON200 != nil {
		diags = stateModel.populateFromAPI(ctx, getResp, dashboardID, spaceID)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Set state
	diags = resp.State.Set(ctx, stateModel)
	resp.Diagnostics.Append(diags...)
}
