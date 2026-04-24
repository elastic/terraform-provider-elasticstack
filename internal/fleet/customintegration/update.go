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
	"fmt"
	"strings"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/asyncutils"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
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

	updateTimeout, fwDiags := plan.Timeouts.Update(ctx, 20*time.Minute)
	resp.Diagnostics.Append(fwDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	apiClient, diags := r.Client().GetKibanaClient(ctx, plan.KibanaConnection)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	fleetClient, err := apiClient.GetFleetClient()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	meetsMinVersion, verDiags := apiClient.EnforceMinVersion(ctx, minVersionCustomPackageGet)
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(verDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !meetsMinVersion {
		resp.Diagnostics.AddError(
			"Kibana version not supported",
			"elasticstack_fleet_custom_integration requires Kibana 8.2.0 or later.",
		)
		return
	}

	// Determine whether we need to re-upload. Re-upload is required if:
	//   - The checksum is unknown (file content changed, signalled by ModifyPlan), or
	//   - Query parameters changed (ignore_mapping_update_errors or skip_data_stream_rollover).
	checksumChanged := plan.Checksum.IsUnknown()
	queryParamsChanged := plan.IgnoreMappingUpdateErrors.ValueBool() != state.IgnoreMappingUpdateErrors.ValueBool() ||
		plan.SkipDataStreamRollover.ValueBool() != state.SkipDataStreamRollover.ValueBool()

	if checksumChanged || queryParamsChanged {
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

		result, diags = fleet.UploadPackage(ctx, fleetClient, uploadOpts)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// If the package name or version changed, uninstall the old package from the
		// same space now that the new one is successfully installed.
		if (result.PackageName != state.PackageName.ValueString() || result.PackageVersion != state.PackageVersion.ValueString()) && state.PackageName.ValueString() != "" {
			diags = fleet.Uninstall(ctx, fleetClient, state.PackageName.ValueString(), state.PackageVersion.ValueString(), state.SpaceID.ValueString(), false)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			diags = waitForInstalledCustomIntegration(ctx, fleetClient, result.PackageName, result.PackageVersion, plan.SpaceID.ValueString())
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
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

func waitForInstalledCustomIntegration(ctx context.Context, fleetClient *fleet.Client, packageName, packageVersion, spaceID string) diag.Diagnostics {
	waitCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	waitErr := asyncutils.WaitForStateTransition(waitCtx, "fleet custom integration", getPackageID(packageName, packageVersion), func(ctx context.Context) (bool, error) {
		pkg, diags := fleet.GetPackage(ctx, fleetClient, packageName, packageVersion, spaceID)
		if diags.HasError() {
			return false, fmt.Errorf("failed to read package installation status: %s", diags[0].Summary())
		}
		if pkg != nil && pkg.Status != nil && strings.EqualFold(*pkg.Status, "installed") {
			return true, nil
		}
		if pkg != nil && pkg.Status != nil && strings.EqualFold(*pkg.Status, "install_failed") {
			return false, fmt.Errorf("package %s/%s installation failed", packageName, packageVersion)
		}

		packages, diags := fleet.GetPackages(ctx, fleetClient, true, spaceID)
		if diags.HasError() {
			return false, fmt.Errorf("failed to list packages during verification: %s", diags[0].Summary())
		}
		for _, candidate := range packages {
			if candidate.Name != packageName || candidate.Version != packageVersion {
				continue
			}
			if candidate.Status != nil && strings.EqualFold(*candidate.Status, "installed") {
				return true, nil
			}
			if candidate.Status != nil && strings.EqualFold(*candidate.Status, "install_failed") {
				return false, fmt.Errorf("package %s/%s installation failed", packageName, packageVersion)
			}
		}

		return false, nil
	})
	if waitErr != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Package not ready after update",
				fmt.Sprintf("Package %s/%s did not become readable as installed after update: %s", packageName, packageVersion, waitErr.Error()),
			),
		}
	}
	return nil
}
