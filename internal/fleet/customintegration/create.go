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

func (r *customIntegrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan customIntegrationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
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

	checksum, err := sha256File(path)
	if err != nil {
		resp.Diagnostics.AddError("Unable to hash custom integration package", err.Error())
		return
	}

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

	result, diags := fleet.UploadPackage(ctx, fleetClient, f, opts)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	version := result.PackageVersion
	if version == "" {
		// The upload response did not carry a version. Fall back to the
		// packages list API, filtering by the name we just uploaded, and
		// picking the highest-ranked installed entry. This matches the
		// spec's version-resolution fallback.
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

	plan.PackageName = types.StringValue(result.PackageName)
	plan.PackageVersion = types.StringValue(version)
	plan.Checksum = types.StringValue(checksum)
	plan.ID = types.StringValue(getPackageID(result.PackageName, version))

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
