package spaces

import (
	"context"

	"github.com/disaster37/go-kibana-rest/v8/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &dataSource{}
	_ datasource.DataSource              = &dataSource{}
	_ datasource.DataSourceWithConfigure = &dataSource{}
)

// NewDataSource is a helper function to simplify the provider implementation.
func NewDataSource() datasource.DataSource {
	return &dataSource{}
}

// dataSource is the data source implementation.
type dataSource struct {
	client *kbapi.KibanaSpacesAPI
}

// Metadata returns the data source type name.
func (d *dataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kibana_spaces"
}

// Configure adds the provider configured client to the data source.
func (d *dataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	kibanaClient, err := client.GetKibanaClient()
	if err != nil {
		resp.Diagnostics.AddError("unable to get Kibana client", err.Error())
		return
	}

	d.client = kibanaClient.KibanaSpaces
}
