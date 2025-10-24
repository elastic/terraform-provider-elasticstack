package integration

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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

	// If space_ids is set, use space-aware installation
	var spaceID string
	if !planModel.SpaceIds.IsNull() && !planModel.SpaceIds.IsUnknown() {
		var tempDiags diag.Diagnostics
		spaceIDs := utils.ListTypeAs[types.String](ctx, planModel.SpaceIds, path.Root("space_ids"), &tempDiags)
		if !tempDiags.HasError() && len(spaceIDs) > 0 {
			spaceID = spaceIDs[0].ValueString()
		}
	}

	if spaceID != "" && spaceID != "default" {
		diags = fleet.InstallPackageInSpace(ctx, client, name, version, spaceID, force)
	} else {
		diags = fleet.InstallPackage(ctx, client, name, version, force)
	}
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return
	}

	planModel.ID = types.StringValue(getPackageID(name, version))

	// Populate space_ids in state
	if planModel.SpaceIds.IsNull() {
		planModel.SpaceIds = types.ListNull(types.StringType)
	}

	diags = state.Set(ctx, planModel)
	respDiags.Append(diags...)
}
