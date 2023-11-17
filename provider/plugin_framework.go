package provider

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/config"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/data_view"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/import_saved_objects"
	"github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	fwschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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
			esKeyName:    schema.GetEsFWConnectionBlock(esKeyName),
			kbKeyName:    schema.GetKbFWConnectionBlock(),
			fleetKeyName: schema.GetFleetFWConnectionBlock(),
		},
	}
}

func (p *Provider) Configure(ctx context.Context, req fwprovider.ConfigureRequest, res *fwprovider.ConfigureResponse) {
	var config config.ProviderConfiguration

	res.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if res.Diagnostics.HasError() {
		return
	}

	client, diags := clients.NewApiClientFromFramework(ctx, config, p.version)
	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}

	res.DataSourceData = client
	res.ResourceData = client
}

func (p *Provider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *Provider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		func() resource.Resource { return &import_saved_objects.Resource{} },
		func() resource.Resource { return &data_view.Resource{} },
	}
}
