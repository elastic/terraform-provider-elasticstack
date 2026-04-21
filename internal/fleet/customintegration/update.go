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

package customintegration

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *customIntegrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan customIntegrationModel
	var state customIntegrationModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiClient, diags := r.client.GetKibanaClient(ctx, plan.KibanaConnection)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	fleetClient, err := apiClient.GetFleetClient()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	// Determine whether we need to re-upload. Re-upload is required if:
	//   - The checksum is unknown (file content changed, signalled by ModifyPlan), or
	//   - Query parameters changed (ignore_mapping_update_errors or skip_data_stream_rollover), or
	//   - The target space changed (package must be installed in the new space).
	checksumChanged := plan.Checksum.IsUnknown()
	queryParamsChanged := plan.IgnoreMappingUpdateErrors.ValueBool() != state.IgnoreMappingUpdateErrors.ValueBool() ||
		plan.SkipDataStreamRollover.ValueBool() != state.SkipDataStreamRollover.ValueBool()
	spaceIDChanged := plan.SpaceID.ValueString() != state.SpaceID.ValueString()

	if checksumChanged || queryParamsChanged || spaceIDChanged {
		filePath := plan.PackagePath.ValueString()
		contentType := detectContentType(filePath)

		uploadOpts := fleet.UploadPackageOptions{
			PackagePath:               filePath,
			ContentType:               contentType,
			IgnoreMappingUpdateErrors: plan.IgnoreMappingUpdateErrors.ValueBool(),
			SkipDataStreamRollover:    plan.SkipDataStreamRollover.ValueBool(),
			SpaceID:                   plan.SpaceID.ValueString(),
		}

		var result *fleet.UploadPackageResult

		if spaceIDChanged {
			// The target space changed. Upload to the new space first (the new space
			// does not have this package so there is no "already installed" conflict),
			// then uninstall from the old space once the upload succeeds.
			var uploadDiags diag.Diagnostics
			result, uploadDiags = fleet.UploadPackage(ctx, fleetClient, uploadOpts)
			resp.Diagnostics.Append(uploadDiags...)
			if resp.Diagnostics.HasError() {
				return
			}

			diags = fleet.Uninstall(ctx, fleetClient, state.PackageName.ValueString(), state.PackageVersion.ValueString(), state.SpaceID.ValueString(), false)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
		} else {
			// Same space: attempt upload first (spec-mandated ordering). If Fleet rejects
			// because the same-name package is already installed (Kibana 8.0.x), uninstall
			// the existing package and retry once.
			var uploadDiags diag.Diagnostics
			result, uploadDiags = fleet.UploadPackage(ctx, fleetClient, uploadOpts)
			resp.Diagnostics.Append(uploadDiags...)
			if resp.Diagnostics.HasError() {
				return
			}

			if result.AlreadyInstalled {
				diags = fleet.Uninstall(ctx, fleetClient, state.PackageName.ValueString(), state.PackageVersion.ValueString(), state.SpaceID.ValueString(), false)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}

				result, uploadDiags = fleet.UploadPackage(ctx, fleetClient, uploadOpts)
				resp.Diagnostics.Append(uploadDiags...)
				if resp.Diagnostics.HasError() {
					return
				}
			}

			// If the package name changed, uninstall the old package from the same space
			// now that the new one is successfully installed.
			if result.PackageName != state.PackageName.ValueString() && state.PackageName.ValueString() != "" {
				diags = fleet.Uninstall(ctx, fleetClient, state.PackageName.ValueString(), state.PackageVersion.ValueString(), state.SpaceID.ValueString(), false)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
			}
		}

		checksum, err := computeSHA256(filePath)
		if err != nil {
			resp.Diagnostics.AddError("Failed to compute checksum", err.Error())
			return
		}

		plan.PackageName = types.StringValue(result.PackageName)
		plan.PackageVersion = types.StringValue(result.PackageVersion)
		plan.Checksum = types.StringValue(checksum)
		plan.ID = types.StringValue(getPackageID(result.PackageName, result.PackageVersion))
	} else {
		// No re-upload needed — carry forward computed fields from state.
		plan.PackageName = state.PackageName
		plan.PackageVersion = state.PackageVersion
		plan.Checksum = state.Checksum
		plan.ID = state.ID
	}

	if plan.SpaceID.IsUnknown() {
		plan.SpaceID = types.StringNull()
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}
