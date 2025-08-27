package maintenance_window

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *MaintenanceWindowResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var planMaintenanceWindow MaintenanceWindowModel

	diags := req.Plan.Get(ctx, &planMaintenanceWindow)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	body, diags := planMaintenanceWindow.toAPICreateRequest(ctx)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverVersion, sdkDiags := r.client.ServerVersion(ctx)
	if sdkDiags.HasError() {
		return
	}

	serverFlavor, sdkDiags := r.client.ServerFlavor(ctx)
	if sdkDiags.HasError() {
		return
	}

	diags = validateMaintenanceWindowServer(serverVersion, serverFlavor)
	if diags.HasError() {
		return
	}

	client, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	spaceID := planMaintenanceWindow.SpaceID.ValueString()
	maintenanceWindowAPIResponse, diags := kibana_oapi.CreateMaintenanceWindow(ctx, client, spaceID, body)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = planMaintenanceWindow.fromAPICreateResponse(ctx, maintenanceWindowAPIResponse)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, planMaintenanceWindow)
	resp.Diagnostics.Append(diags...)
}
