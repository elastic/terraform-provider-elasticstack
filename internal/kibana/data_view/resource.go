package data_view

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &DataViewResource{}
	_ resource.ResourceWithConfigure   = &DataViewResource{}
	_ resource.ResourceWithImportState = &DataViewResource{}
)

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &DataViewResource{}
}

type DataViewResource struct {
	client *clients.ApiClient
}

func (r *DataViewResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func (r *DataViewResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_%s", req.ProviderTypeName, "kibana_data_view")
}

func (r *DataViewResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	composite, diags := clients.CompositeIdFromStrFw(req.ID)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	stateModel := dataViewModel{
		ID:       types.StringValue(req.ID),
		SpaceID:  types.StringValue(composite.ClusterId),
		Override: types.BoolValue(false),
		DataView: types.ObjectUnknown(getDataViewAttrTypes()),
	}

	diags = resp.State.Set(ctx, stateModel)
	resp.Diagnostics.Append(diags...)
}
