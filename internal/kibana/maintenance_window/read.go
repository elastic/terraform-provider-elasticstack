package maintenance_window

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *MaintenanceWindowResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var stateModel MaintenanceWindowModel

	diags := req.State.Get(ctx, &stateModel)
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

	viewID, spaceID := stateModel.getMaintenanceWindowIDAndSpaceID()
	maintenanceWindow, diags := kibana_oapi.GetMaintenanceWindow(ctx, client, spaceID, viewID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if maintenanceWindow == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	diags = stateModel.fromAPIReadResponse(ctx, maintenanceWindow)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, stateModel)
	resp.Diagnostics.Append(diags...)
}
