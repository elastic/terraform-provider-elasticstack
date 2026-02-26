package output_ds

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func (d *outputDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model outputModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := d.client.GetFleetClient()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	outputName := model.Name.ValueString()
	spaceID := model.SpaceID.ValueString()
	outputs, diags := fleet.GetOutputs(ctx, client, spaceID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	matched := false
	for _, union := range outputs {
		diags = model.populateFromAPI(ctx, &union)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		if model.Name.ValueString() == outputName {
			matched = true
			break
		}
	}

	if !matched {
		if spaceID == "" {
			resp.Diagnostics.AddError("Output not found", fmt.Sprintf("Output '%s' not found", outputName))
		} else {
			resp.Diagnostics.AddError("Output not found", fmt.Sprintf("Output '%s' not found in space '%s'", outputName, spaceID))
		}
		return
	}

	diags = resp.State.Set(ctx, &model)
	resp.Diagnostics.Append(diags...)
}
