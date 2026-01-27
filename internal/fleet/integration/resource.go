package integration

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                 = &integrationResource{}
	_ resource.ResourceWithConfigure    = &integrationResource{}
	_ resource.ResourceWithUpgradeState = &integrationResource{}

	// MinVersionIgnoreMappingUpdateErrors is the minimum version that supports the ignore_mapping_update_errors parameter
	MinVersionIgnoreMappingUpdateErrors = version.Must(version.NewVersion("8.11.0"))
	// MinVersionSkipDataStreamRollover is the minimum version that supports the skip_data_stream_rollover parameter
	MinVersionSkipDataStreamRollover = MinVersionIgnoreMappingUpdateErrors
)

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &integrationResource{}
}

type integrationResource struct {
	client *clients.ApiClient
}

func (r *integrationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func (r *integrationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_%s", req.ProviderTypeName, "fleet_integration")
}
