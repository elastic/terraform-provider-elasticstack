package maintenance_window

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource              = &MaintenanceWindowResource{}
	_ resource.ResourceWithConfigure = &MaintenanceWindowResource{}
	// _ resource.ResourceWithImportState = &MaintenanceWindowResource{}
)

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &MaintenanceWindowResource{}
}

type MaintenanceWindowResource struct {
	client *clients.ApiClient
}

func (r *MaintenanceWindowResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

// Metadata returns the provider type name.
func (r *MaintenanceWindowResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_%s", req.ProviderTypeName, "kibana_maintenance_window_2")
}

// func (r *MaintenanceWindowResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
// 	composite, diags := clients.CompositeIdFromStrFw(req.ID)
// 	resp.Diagnostics.Append(diags...)
// 	if diags.HasError() {
// 		return
// 	}

// 	stateModel := maintenanceWindowModel{
// 		ID:                types.StringValue(req.ID),
// 		SpaceID:           types.StringValue(composite.ClusterId),
// 		Override:          types.BoolValue(false),
// 		MaintenanceWindow: types.ObjectUnknown(getMaintenanceWindowAttrTypes()),
// 	}

// 	diags = resp.State.Set(ctx, stateModel)
// 	resp.Diagnostics.Append(diags...)
// }
