package securitylistitem

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.Resource                = &securityListItemResource{}
	_ resource.ResourceWithConfigure   = &securityListItemResource{}
	_ resource.ResourceWithImportState = &securityListItemResource{}
)

func NewResource() resource.Resource {
	return &securityListItemResource{}
}

type securityListItemResource struct {
	client *clients.ApiClient
}

func (r *securityListItemResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kibana_security_list_item"
}

func (r *securityListItemResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func (r *securityListItemResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}
