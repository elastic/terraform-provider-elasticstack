package dashboard

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var planModel dashboardModel

	diags := req.Plan.Get(ctx, &planModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the Kibana client
	kibanaClient, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError("Unable to get Kibana client", err.Error())
		return
	}

	spaceID := planModel.SpaceID.ValueString()

	// Convert the plan to an API request
	apiReq := planModel.toAPICreateRequest(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the dashboard
	createResp, diags := kibana_oapi.CreateDashboard(ctx, kibanaClient, spaceID, apiReq)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Populate state from API response
	if createResp.JSON200 != nil {
		diags = planModel.populateFromAPI(ctx, createResp, createResp.JSON200.Id, spaceID)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Set state
	diags = resp.State.Set(ctx, planModel)
	resp.Diagnostics.Append(diags...)
}
