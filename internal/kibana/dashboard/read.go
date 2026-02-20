package dashboard

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var stateModel dashboardModel

	diags := req.State.Get(ctx, &stateModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	readModel, diags := r.read(ctx, stateModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if readModel == nil {
		// Dashboard not found, remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, *readModel)...)
}

func (r *Resource) read(ctx context.Context, stateModel dashboardModel) (*dashboardModel, diag.Diagnostics) {
	// Parse composite ID
	composite, diags := clients.CompositeIDFromStrFw(stateModel.ID.ValueString())
	if diags.HasError() {
		return nil, diags
	}

	dashboardID := composite.ResourceID
	spaceID := composite.ClusterID

	// Get the Kibana client
	kibanaClient, err := r.client.GetKibanaOapiClient()
	if err != nil {
		diags.AddError("Unable to get Kibana client", err.Error())
		return nil, diags
	}

	// Get the dashboard
	getResp, getDiags := kibanaoapi.GetDashboard(ctx, kibanaClient, spaceID, dashboardID)
	diags.Append(getDiags...)
	if diags.HasError() {
		return nil, diags
	}

	if getResp == nil {
		return nil, diags
	}

	if getResp.JSON200 == nil {
		diags.AddError("Empty response when getting dashboard", "GET dashboard was successful, however contained an empty response")
		return nil, diags
	}

	diags.Append(stateModel.populateFromAPI(ctx, getResp, dashboardID, spaceID)...)
	return &stateModel, diags
}
