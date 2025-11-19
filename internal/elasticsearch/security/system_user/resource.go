package system_user

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &systemUserResource{}
var _ resource.ResourceWithConfigure = &systemUserResource{}

func NewSystemUserResource() resource.Resource {
	return &systemUserResource{}
}

type systemUserResource struct {
	client *clients.ApiClient
}

func (r *systemUserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_elasticsearch_security_system_user"
}

func (r *systemUserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}
