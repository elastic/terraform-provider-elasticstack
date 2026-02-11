package output

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (d *outputDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model outputDataSourceModel

	diags := req.Config.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := d.client.GetFleetClient()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	outputID := model.OutputID.ValueString()
	spaceID := model.SpaceID.ValueString()

	output, diags := fleet.GetOutput(ctx, client, outputID, spaceID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if output == nil {
		resp.Diagnostics.AddError("Output not found", fmt.Sprintf("Fleet output %q was not found.", outputID))
		return
	}

	discriminator, err := output.Discriminator()
	if err != nil {
		resp.Diagnostics.AddError("Failed to read output type", err.Error())
		return
	}

	switch discriminator {
	case "elasticsearch", "logstash", "kafka":
	default:
		resp.Diagnostics.AddError(
			"Unsupported Fleet output type",
			fmt.Sprintf("Fleet output %q has unsupported type %q. Supported types are elasticsearch, logstash, kafka.", outputID, discriminator),
		)
		return
	}

	diags = model.populateFromAPI(ctx, output)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if model.ID.IsNull() || model.ID.IsUnknown() {
		model.ID = types.StringValue(outputID)
	}

	diags = resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
}
