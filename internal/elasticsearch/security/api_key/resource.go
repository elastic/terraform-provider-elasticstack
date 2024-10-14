package api_key

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &Resource{}
var _ resource.ResourceWithConfigure = &Resource{}
var (
	MinVersion                         = version.Must(version.NewVersion("8.0.0")) // Enabled in 8.0
	MinVersionReturningRoleDescriptors = version.Must(version.NewVersion("8.5.0"))
	MinVersionWithRestriction          = version.Must(version.NewVersion("8.9.0")) // Enabled in 8.0
)

type Resource struct {
	client *clients.ApiClient
}

var configuredResources = []*Resource{}

func (r *Resource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(request.ProviderData)
	response.Diagnostics.Append(diags...)
	r.client = client
	configuredResources = append(configuredResources, r)
}

func (r *Resource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_elasticsearch_security_api_key"
}

func (r *Resource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	response.Diagnostics.AddError("Update not supported", "Update not supported")
}
