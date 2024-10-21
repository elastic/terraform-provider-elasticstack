package integration

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *integrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	r.create(ctx, req.Plan, &resp.State, &resp.Diagnostics)
}

func (r integrationResource) create(ctx context.Context, plan tfsdk.Plan, state *tfsdk.State, respDiags *diag.Diagnostics) {
	var planModel integrationModel

	diags := plan.Get(ctx, &planModel)
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return
	}

	client, err := r.client.GetFleetClient()
	if err != nil {
		respDiags.AddError(err.Error(), "")
		return
	}

	name := planModel.Name.ValueString()
	version := planModel.Version.ValueString()
	force := planModel.Force.ValueBool()
	diags = fleet.InstallPackage(ctx, client, name, version, force)
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return
	}

	planModel.ID = types.StringValue(getPackageID(name, version))

	diags = state.Set(ctx, planModel)
	respDiags.Append(diags...)
}
