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
	"os"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *customIntegrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan customIntegrationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state customIntegrationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	kibanaClient, diags := r.client.GetKibanaClient(ctx, plan.KibanaConnection)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	fleetClient, err := kibanaClient.GetFleetClient()
	if err != nil {
		resp.Diagnostics.AddError("Failed to obtain Fleet client", err.Error())
		return
	}

	path := plan.PackagePath.ValueString()

	newHash, err := sha256File(path)
	if err != nil {
		resp.Diagnostics.AddError("Unable to hash custom integration package", err.Error())
		return
	}

	// Only the mutable query-parameter attributes changed? Re-uploading
	// without a content change is harmless — Fleet accepts a re-upload of
	// the same archive — and keeps the semantics simple: any update
	// re-runs the upload with the current values.
	f, err := os.Open(path)
	if err != nil {
		resp.Diagnostics.AddError("Unable to open custom integration package", err.Error())
		return
	}
	defer f.Close()

	opts := fleet.UploadPackageOptions{
		SpaceID:                   plan.SpaceID.ValueString(),
		ContentType:               detectContentType(path),
		IgnoreMappingUpdateErrors: plan.IgnoreMappingUpdateErrors.ValueBoolPointer(),
		SkipDataStreamRollover:    plan.SkipDataStreamRollover.ValueBoolPointer(),
	}

	result, uploadDiags := fleet.UploadPackage(ctx, fleetClient, f, opts)
	resp.Diagnostics.Append(uploadDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	version := result.PackageVersion
	if version == "" {
		pkgs, listDiags := fleet.GetPackages(ctx, fleetClient, true, plan.SpaceID.ValueString())
		resp.Diagnostics.Append(listDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		version = pickInstalledVersion(pkgs, result.PackageName)
		if version == "" {
			resp.Diagnostics.AddError(
				"Unable to resolve uploaded package version",
				fmt.Sprintf("The Fleet upload for package %q succeeded but no matching installed entry was found in the packages list API.", result.PackageName),
			)
			return
		}
	}

	// If the uploaded package reports a different name than what was in
	// state, the prior package has been superseded. Uninstall the old
	// name+version so it does not linger as an orphan installation. A
	// failure here surfaces as a diagnostic but does not roll back the
	// new upload — the user is told explicitly that they have a hybrid
	// state and can clean it up.
	oldName := state.PackageName.ValueString()
	oldVersion := state.PackageVersion.ValueString()
	if oldName != "" && oldName != result.PackageName {
		uninstallDiags := fleet.Uninstall(ctx, fleetClient, oldName, oldVersion, state.SpaceID.ValueString(), false)
		if uninstallDiags.HasError() {
			resp.Diagnostics.AddWarning(
				fmt.Sprintf("Failed to uninstall superseded package %s/%s", oldName, oldVersion),
				fmt.Sprintf("The new package %s/%s was uploaded successfully but the prior package could not be uninstalled. It may still be present in Fleet and require manual cleanup. Underlying diagnostics: %v", result.PackageName, version, uninstallDiags),
			)
		}
	}

	plan.PackageName = types.StringValue(result.PackageName)
	plan.PackageVersion = types.StringValue(version)
	plan.Checksum = types.StringValue(newHash)
	plan.ID = types.StringValue(getPackageID(result.PackageName, version))

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
