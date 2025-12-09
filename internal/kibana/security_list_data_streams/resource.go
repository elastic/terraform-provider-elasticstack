package security_list_data_streams

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.Resource                = &securityListDataStreamsResource{}
	_ resource.ResourceWithConfigure   = &securityListDataStreamsResource{}
	_ resource.ResourceWithImportState = &securityListDataStreamsResource{}
)

func NewResource() resource.Resource {
	return &securityListDataStreamsResource{}
}

type securityListDataStreamsResource struct {
	client *clients.ApiClient
}

func (r *securityListDataStreamsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kibana_security_list_data_streams"
}

func (r *securityListDataStreamsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func (r *securityListDataStreamsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
