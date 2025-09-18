package maintenance_window

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

	isSupported, sdkDiags := r.client.EnforceMinVersion(ctx, version.Must(version.NewVersion("9.1.0")))
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !isSupported {
		resp.Diagnostics.AddError("Unsupported server version", "Maintenance windows are not supported until Elastic Stack v9.0. Upgrade the target server to use this resource")
		return
	}

	client, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	spaceID := planMaintenanceWindow.SpaceID.ValueString()
	createMaintenanceWindowResponse, diags := kibana_oapi.CreateMaintenanceWindow(ctx, client, spaceID, body)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	/*
	* In create/update paths we typically follow the write operation with a read, and then set the state from the read.
	* We want to avoid a dirty plan immediately after an apply.
	 */
	maintenanceWindowID := createMaintenanceWindowResponse.JSON200.Id
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
