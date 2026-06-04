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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func deleteIntegration(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	resourceID string,
	spaceID string,
	model integrationModel,
) diag.Diagnostics {
	var diags diag.Diagnostics

	fleetClient := client.GetFleetClient()

	name := model.Name.ValueString()
	version := model.Version.ValueString()
	force := model.Force.ValueBool()
	if model.SkipDestroy.ValueBool() {
		tflog.Debug(ctx, "Skipping uninstall of integration package", map[string]any{attrName: name, attrVersion: version})
		return diags
	}

	spaceAware := false
	if typeutils.IsKnown(model.SpaceID) {
		supported, versionDiags := supportsSpaceAwareIntegration(ctx, client, spaceID)
		diags.Append(versionDiags...)
		if diags.HasError() {
			return diags
		}
		spaceAware = supported
	}

	if spaceAware {
		pkg, getDiags := fleet.GetPackage(ctx, fleetClient, name, version, spaceID)
		diags.Append(getDiags...)
		if diags.HasError() {
			return diags
		}

		if isInstalledInMultipleSpaces(pkg, spaceID) {
			deleteDiags := fleet.DeleteKibanaAssets(ctx, fleetClient, name, version, spaceID, force)
			diags.Append(deleteDiags...)
			return diags
		}
	}

	uninstallDiags := fleet.Uninstall(ctx, fleetClient, name, version, spaceID, force)
	diags.Append(uninstallDiags...)
	return diags
}

func isInstalledInMultipleSpaces(pkg *kbapi.PackageInfo, spaceID string) bool {
	if pkg == nil || pkg.InstallationInfo == nil {
		return false
	}

	if !packageInstalledInKibanaSpace(pkg.InstallationInfo, spaceID) {
		return false
	}

	otherSpaces := 0
	if pkg.InstallationInfo.AdditionalSpacesInstalledKibana != nil {
		otherSpaces = len(*pkg.InstallationInfo.AdditionalSpacesInstalledKibana)
	}
	isPrimary := pkg.InstallationInfo.InstalledKibanaSpaceId != nil &&
		*pkg.InstallationInfo.InstalledKibanaSpaceId == spaceID
	if isPrimary {
		return otherSpaces > 0
	}
	// Target is in additional spaces: primary + (additional minus self) = multi.
	return otherSpaces > 1 || pkg.InstallationInfo.InstalledKibanaSpaceId != nil
}
