// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package integration

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
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
	needsVersionCheck := typeutils.IsKnown(planModel.IgnoreMappingUpdateErrors) || typeutils.IsKnown(planModel.SkipDataStreamRollover)
	if needsVersionCheck {
		serverVersion, versionDiags := r.client.ServerVersion(ctx)
		respDiags.Append(diagutil.FrameworkDiagsFromSDK(versionDiags)...)
		if respDiags.HasError() {
			return
		}

		// Validate ignore_mapping_update_errors
		if typeutils.IsKnown(planModel.IgnoreMappingUpdateErrors) {
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
		if typeutils.IsKnown(planModel.SkipDataStreamRollover) {
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
	if typeutils.IsKnown(planModel.SpaceID) {
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
