package provider

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/config"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/indices"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/agent_policy"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/enrollment_tokens"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/integration"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/integration_ds"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/server_host"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/data_view"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/import_saved_objects"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/spaces"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics/private_location"
	"github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	fwschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ fwprovider.Provider = &Provider{}
)

type Provider struct {
	version string
}

// NewFrameworkProvider instantiates plugin framework's provider
func NewFrameworkProvider(version string) fwprovider.Provider {
	return &Provider{
		version: version,
	}
}

func (p *Provider) Metadata(_ context.Context, _ fwprovider.MetadataRequest, res *fwprovider.MetadataResponse) {
	res.TypeName = "elasticstack"
	res.Version = p.version
}

func (p *Provider) Schema(ctx context.Context, req fwprovider.SchemaRequest, res *fwprovider.SchemaResponse) {
	res.Schema = fwschema.Schema{
		Blocks: map[string]fwschema.Block{
			esKeyName:    schema.GetEsFWConnectionBlock(esKeyName, true),
			kbKeyName:    schema.GetKbFWConnectionBlock(),
			fleetKeyName: schema.GetFleetFWConnectionBlock(),
		},
	}
}

func (p *Provider) Configure(ctx context.Context, req fwprovider.ConfigureRequest, res *fwprovider.ConfigureResponse) {
	var cfg config.ProviderConfiguration

	res.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if res.Diagnostics.HasError() {
		return
	}

	client, diags := clients.NewApiClientFromFramework(ctx, cfg, p.version)
	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}

	res.DataSourceData = client
	res.ResourceData = client
}

func (p *Provider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		indices.NewDataSource,
		spaces.NewDataSource,
		enrollment_tokens.NewDataSource,
		integration_ds.NewDataSource,
	}
}

func (p *Provider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		func() resource.Resource { return &import_saved_objects.Resource{} },
		func() resource.Resource { return &data_view.Resource{} },
		func() resource.Resource { return &private_location.Resource{} },
		func() resource.Resource { return &index.Resource{} },
		func() resource.Resource { return &synthetics.Resource{} },
		agent_policy.NewResource,
		integration.NewResource,
		server_host.NewResource,
	}
}
