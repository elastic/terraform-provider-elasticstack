package exception_container

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &exceptionContainerResource{}
var _ resource.ResourceWithConfigure = &exceptionContainerResource{}
var _ resource.ResourceWithImportState = &exceptionContainerResource{}

func NewExceptionContainerResource() resource.Resource {
	return &exceptionContainerResource{}
}

type exceptionContainerResource struct {
	client *clients.ApiClient
}

func (r *exceptionContainerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kibana_security_exception_container"
}

func (r *exceptionContainerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func (r *exceptionContainerResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}
