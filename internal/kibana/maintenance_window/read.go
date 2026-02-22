package maintenancewindow

import (
	"context"

	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var stateModel Model

	req.State.GetAttribute(ctx, path.Root("id"), &stateModel.ID)
	req.State.GetAttribute(ctx, path.Root("space_id"), &stateModel.SpaceID)

	serverVersion, sdkDiags := r.client.ServerVersion(ctx)
	if sdkDiags.HasError() {
		return
	}

	serverFlavor, sdkDiags := r.client.ServerFlavor(ctx)
	if sdkDiags.HasError() {
		return
	}

	diags := validateMaintenanceWindowServer(serverVersion, serverFlavor)
	if diags.HasError() {
		return
	}

	client, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	maintenanceWindowID, spaceID := stateModel.getMaintenanceWindowIDAndSpaceID()
	maintenanceWindow, diags := kibanaoapi.GetMaintenanceWindow(ctx, client, spaceID, maintenanceWindowID)
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

	stateModel.ID = types.StringValue(maintenanceWindowID)
	stateModel.SpaceID = types.StringValue(spaceID)

	diags = resp.State.Set(ctx, stateModel)
	resp.Diagnostics.Append(diags...)
}
