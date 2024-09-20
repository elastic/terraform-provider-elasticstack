package enrollment_tokens

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var (
	_ datasource.DataSource              = &enrollmentTokensDataSource{}
	_ datasource.DataSourceWithConfigure = &enrollmentTokensDataSource{}
)

// NewDataSource is a helper function to simplify the provider implementation.
func NewDataSource() datasource.DataSource {
	return &enrollmentTokensDataSource{}
}

type enrollmentTokensDataSource struct {
	client *clients.ApiClient
}

func (d *enrollmentTokensDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_%s", req.ProviderTypeName, "fleet_enrollment_tokens")
}

func (d *enrollmentTokensDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	d.client = client
}
