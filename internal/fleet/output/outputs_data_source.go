package output

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var (
	_ datasource.DataSource              = &outputsDataSource{}
	_ datasource.DataSourceWithConfigure = &outputsDataSource{}
)

// NewOutputsDataSource is a helper function to simplify the provider implementation.
func NewOutputsDataSource() datasource.DataSource {
	return &outputsDataSource{}
}

type outputsDataSource struct {
	client *clients.ApiClient
}

func (d *outputsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_%s", req.ProviderTypeName, "fleet_outputs")
}

func (d *outputsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	d.client = client
}

func (d *outputsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = getOutputsDataSourceSchema()
}
