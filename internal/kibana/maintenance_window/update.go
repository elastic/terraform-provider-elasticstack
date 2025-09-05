package maintenance_window

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

	maintenanceWindowID, spaceID := planMaintenanceWindow.getMaintenanceWindowIDAndSpaceID()
	diags = kibana_oapi.UpdateMaintenanceWindow(ctx, client, spaceID, maintenanceWindowID, body)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	/*
	* In create/update paths we typically follow the write operation with a read, and then set the state from the read.
	* We want to avoid a dirty plan immediately after an apply.
	 */
	readMaintenanceWindowResponse, diags := kibana_oapi.GetMaintenanceWindow(ctx, client, spaceID, maintenanceWindowID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if readMaintenanceWindowResponse == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	diags = planMaintenanceWindow.fromAPIReadResponse(ctx, readMaintenanceWindowResponse)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	planMaintenanceWindow.ID = types.StringValue(maintenanceWindowID)
	planMaintenanceWindow.SpaceID = types.StringValue(spaceID)

	diags = resp.State.Set(ctx, planMaintenanceWindow)
	resp.Diagnostics.Append(diags...)
}
