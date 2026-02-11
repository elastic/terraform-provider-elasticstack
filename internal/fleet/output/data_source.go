package output

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &outputDataSource{}
	_ datasource.DataSourceWithConfigure = &outputDataSource{}
)

func NewDataSource() datasource.DataSource {
	return &outputDataSource{}
}

type outputDataSource struct {
	client *clients.ApiClient
}

func (d *outputDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_%s", req.ProviderTypeName, "fleet_output")
}

func (d *outputDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	d.client = client
}

func (d *outputDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = getDataSourceSchema()
}

func (d *outputDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config outputDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := d.client.GetFleetClient()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	outputID := config.OutputID.ValueString()
	requestSpaceID := ""
	if utils.IsKnown(config.SpaceID) {
		requestSpaceID = config.SpaceID.ValueString()
	}

	output, diags := fleet.GetOutput(ctx, client, outputID, requestSpaceID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	compositeSpaceID := requestSpaceID
	if compositeSpaceID == "" {
		compositeSpaceID = "default"
	}

	if output == nil {
		resp.Diagnostics.AddError(
			"Output not found",
			fmt.Sprintf("Fleet output %q was not found in space %q", outputID, compositeSpaceID),
		)
		return
	}

	var state outputDataSourceModel
	if utils.IsKnown(config.SpaceID) {
		state.SpaceID = config.SpaceID
	} else {
		state.SpaceID = types.StringNull()
	}
	state.OutputID = types.StringValue(outputID)

	resp.Diagnostics.Append(state.populateFromAPI(ctx, output)...)
	if resp.Diagnostics.HasError() {
		return
	}

	compositeID := clients.CompositeId{ClusterId: compositeSpaceID, ResourceId: outputID}
	state.ID = types.StringValue(compositeID.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
