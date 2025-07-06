package agent_configuration

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &resourceAgentConfiguration{}

func NewAgentConfigurationResource() resource.Resource {
	return &resourceAgentConfiguration{}
}

type resourceAgentConfiguration struct {
	client *clients.ApiClient
}

func (r *resourceAgentConfiguration) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_apm_agent_configuration"
}

func (r *resourceAgentConfiguration) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}
