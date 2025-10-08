package integration

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *integrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var stateModel integrationModel

	diags := req.State.Get(ctx, &stateModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.GetFleetClient()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	name := stateModel.Name.ValueString()
	pkg, diags := fleet.GetPackage(ctx, client, name, "")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if pkg.Status != nil && *pkg.Status != "installed" {
		resp.State.RemoveResource(ctx)
		return
	}

	stateModel.ID = types.StringValue(getPackageID(name))
	stateModel.Version = types.StringValue(pkg.Version)

	diags = resp.State.Set(ctx, stateModel)
	resp.Diagnostics.Append(diags...)
}
