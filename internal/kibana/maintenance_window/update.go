package maintenance_window

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *MaintenanceWindowResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var planMaintenanceWindow MaintenanceWindowModel

	diags := req.Plan.Get(ctx, &planMaintenanceWindow)
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

	body, diags := planMaintenanceWindow.toAPIUpdateRequest(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	viewID, spaceID := planMaintenanceWindow.getMaintenanceWindowIDAndSpaceID()
	maintenanceWindow, diags := kibana_oapi.UpdateMaintenanceWindow(ctx, client, spaceID, viewID, body)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = planMaintenanceWindow.fromAPIUpdateResponse(ctx, maintenanceWindow)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, planMaintenanceWindow)
	resp.Diagnostics.Append(diags...)
}
