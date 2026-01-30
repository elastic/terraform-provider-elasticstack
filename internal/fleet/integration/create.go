package integration

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
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
	installOptions := fleet.InstallPackageOptions{
		Force:             planModel.Force.ValueBool(),
		Prerelease:        planModel.Prerelease.ValueBool(),
		IgnoreConstraints: planModel.IgnoreConstraints.ValueBool(),
	}

	// Check if version-dependent parameters are set and validate version support
	needsVersionCheck := utils.IsKnown(planModel.IgnoreMappingUpdateErrors) || utils.IsKnown(planModel.SkipDataStreamRollover)
	if needsVersionCheck {
		serverVersion, versionDiags := r.client.ServerVersion(ctx)
		respDiags.Append(diagutil.FrameworkDiagsFromSDK(versionDiags)...)
		if respDiags.HasError() {
			return
		}

		// Validate ignore_mapping_update_errors
		if utils.IsKnown(planModel.IgnoreMappingUpdateErrors) {
			if serverVersion.LessThan(MinVersionIgnoreMappingUpdateErrors) {
				respDiags.AddError(
					"Unsupported parameter for server version",
					fmt.Sprintf("The 'ignore_mapping_update_errors' parameter requires server version %s or higher. Current version: %s",
						MinVersionIgnoreMappingUpdateErrors.String(), serverVersion.String()),
				)
				return
			}
			installOptions.IgnoreMappingUpdateErrors = planModel.IgnoreMappingUpdateErrors.ValueBoolPointer()
		}

		// Validate skip_data_stream_rollover
		if utils.IsKnown(planModel.SkipDataStreamRollover) {
			if serverVersion.LessThan(MinVersionSkipDataStreamRollover) {
				respDiags.AddError(
					"Unsupported parameter for server version",
					fmt.Sprintf("The 'skip_data_stream_rollover' parameter requires server version %s or higher. Current version: %s",
						MinVersionSkipDataStreamRollover.String(), serverVersion.String()),
				)
				return
			}
			installOptions.SkipDataStreamRollover = planModel.SkipDataStreamRollover.ValueBoolPointer()
		}
	}

	// If space_id is set, use space-aware installation
	if utils.IsKnown(planModel.SpaceID) {
		installOptions.SpaceID = planModel.SpaceID.ValueString()
	}

	diags = fleet.InstallPackage(ctx, client, name, version, installOptions)
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return
	}

	planModel.ID = types.StringValue(getPackageID(name, version))

	// Populate space_id in state
	// If space_id is unknown (not provided by user), set to null to satisfy Terraform's requirement
	if planModel.SpaceID.IsUnknown() {
		planModel.SpaceID = types.StringNull()
	}

	diags = state.Set(ctx, planModel)
	respDiags.Append(diags...)
}
