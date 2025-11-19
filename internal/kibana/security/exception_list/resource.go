package exception_list

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &exceptionListResource{}
var _ resource.ResourceWithConfigure = &exceptionListResource{}

func NewExceptionListResource() resource.Resource {
	return &exceptionListResource{}
}

type exceptionListResource struct {
	client *clients.ApiClient
}

func (r *exceptionListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kibana_security_exception_list"
}

func (r *exceptionListResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}
