package provider

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
			esKeyName: schema.GetEsFWConnectionBlock(esKeyName, true),
		},
	}
}

func (p *Provider) Configure(ctx context.Context, req fwprovider.ConfigureRequest, res *fwprovider.ConfigureResponse) {
	esConn := []*clients.ElasticSearchConnection{}
	diags := req.Config.GetAttribute(ctx, path.Root(esKeyName), &esConn)
	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}
	apiClient, diags := clients.NewFWEsApiClient(ctx, esConn[0], p.version, true)
	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}
	res.DataSourceData = apiClient
	res.ResourceData = apiClient
}

func (p *Provider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *Provider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}
