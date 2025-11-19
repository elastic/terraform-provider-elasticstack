package exception_item

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &exceptionItemResource{}
var _ resource.ResourceWithConfigure = &exceptionItemResource{}
var _ resource.ResourceWithImportState = &exceptionItemResource{}

func NewExceptionItemResource() resource.Resource {
	return &exceptionItemResource{}
}

type exceptionItemResource struct {
	client *clients.ApiClient
}

func (r *exceptionItemResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kibana_security_exception_item"
}

func (r *exceptionItemResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func (r *exceptionItemResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}
