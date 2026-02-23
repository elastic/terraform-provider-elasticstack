package security_enable_rule

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource              = &EnableRuleResource{}
	_ resource.ResourceWithConfigure = &EnableRuleResource{}

	minSupportedVersion = version.Must(version.NewVersion("8.11.0"))
)

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &EnableRuleResource{}
}

type EnableRuleResource struct {
	client *clients.APIClient
}

func (r *EnableRuleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func (r *EnableRuleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_%s", req.ProviderTypeName, "kibana_security_enable_rule")
}
