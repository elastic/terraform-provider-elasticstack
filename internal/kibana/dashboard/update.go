package dashboard

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var planModel dashboardModel

	diags := req.Plan.Get(ctx, &planModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse composite ID
	composite, diags := clients.CompositeIdFromStrFw(planModel.ID.ValueString())
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

	// Convert the plan to an API request
	apiReq := planModel.toAPIUpdateRequest(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the dashboard
	_, diags = kibana_oapi.UpdateDashboard(ctx, kibanaClient, spaceID, dashboardID, apiReq)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	readModel, diags := r.read(ctx, planModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if readModel == nil {
		resp.Diagnostics.AddError("Error reading dashboard after update", "The dashboard was updated but could not be read.")
		return
	}

	// Set state
	diags = resp.State.Set(ctx, *readModel)
	resp.Diagnostics.Append(diags...)
}
