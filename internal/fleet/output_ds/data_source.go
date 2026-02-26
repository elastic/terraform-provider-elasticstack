package output_ds

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var (
	_ datasource.DataSource              = &outputDataSource{}
	_ datasource.DataSourceWithConfigure = &outputDataSource{}
)

func NewDataSource() datasource.DataSource {
	return &outputDataSource{}
}

type outputDataSource struct {
	client *clients.APIClient
}

func (d *outputDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fleet_output"
}

func (d *outputDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	d.client = client
}
