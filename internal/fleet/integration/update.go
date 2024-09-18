package integration

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *integrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var planModel integrationModel

	diags := req.Plan.Get(ctx, &planModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.GetFleetClient()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	name := planModel.Name.ValueString()
	version := planModel.Version.ValueString()
	force := planModel.Force.ValueBool()
	diags = fleet.InstallPackage(ctx, client, name, version, force)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	planModel.ID = types.StringValue(getPackageID(name, version))

	diags = resp.State.Set(ctx, planModel)
	resp.Diagnostics.Append(diags...)
}
