package integration_ds

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (d *integrationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model integrationDataSourceModel

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

	name := model.Name.ValueString()
	prerelease := model.Prerelease.ValueBool()
	packages, diags := fleet.GetPackages(ctx, client, prerelease)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if model.ID.ValueString() == "" {
		hash, err := utils.StringToHash(name)
		if err != nil {
			resp.Diagnostics.AddError(err.Error(), "")
			return
		}
		model.ID = types.StringPointerValue(hash)
	}

	model.populateFromAPI(name, packages)

	diags = resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
}
