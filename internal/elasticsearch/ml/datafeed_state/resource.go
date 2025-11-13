package datafeed_state

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func NewMLDatafeedStateResource() resource.Resource {
	return &mlDatafeedStateResource{}
}

type mlDatafeedStateResource struct {
	client *clients.ApiClient
}

func (r *mlDatafeedStateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_elasticsearch_ml_datafeed_state"
}

func (r *mlDatafeedStateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func (r *mlDatafeedStateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to datafeed_id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("datafeed_id"), req, resp)
}
