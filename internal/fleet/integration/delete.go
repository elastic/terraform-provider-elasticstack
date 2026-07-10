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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// deleteKibanaAssetsWithFallback deletes the Kibana assets for name/version in
// spaceID. When Fleet rejects the call because spaceID is the package's
// install space and the package is also installed in other spaces, the only
// remaining way to clear spaceID is a full package uninstall — but that
// uninstalls the package from EVERY space, not just spaceID. That is only
// acceptable when the caller has explicitly opted in via force; otherwise an
// actionable error is returned instead of silently affecting other,
// independently managed spaces.
func deleteKibanaAssetsWithFallback(
	ctx context.Context,
	fleetClient *fleet.Client,
	name, version, spaceID string,
	force bool,
) diag.Diagnostics {
	deleteDiags := fleet.DeleteKibanaAssets(ctx, fleetClient, name, version, spaceID, force)
	if !deleteDiags.HasError() {
		return deleteDiags
	}
	if !fleet.ContainsInstallSpaceDeleteRejection(deleteDiags) {
		return deleteDiags
	}

	if !force {
		return diag.Diagnostics{diag.NewErrorDiagnostic(
			"Cannot remove Kibana assets from the package's install space",
			fmt.Sprintf(
				"Fleet rejected removing Kibana assets for package %q version %q from space %q "+
					"because that space is the package's install space (the space where it was "+
					"originally installed), and the package is also installed in one or more other "+
					"spaces. Fleet does not support removing Kibana assets from only the install "+
					"space while the package remains installed elsewhere; the only way to clear "+
					"this space is to fully uninstall the package, which would also remove it from "+
					"every other space where it is installed. To destroy this resource without "+
					"affecting other spaces, first destroy the elasticstack_fleet_integration "+
					"resource(s) managing those other space(s). Alternatively, set force = true on "+
					"this resource to acknowledge that destroying it will uninstall the package "+
					"from ALL spaces where it is installed.",
				name, version, spaceID,
			),
		)}
	}

	tflog.Debug(ctx, "DeleteKibanaAssets rejected by Fleet (install space); force=true, falling back to global Uninstall", map[string]any{attrName: name, attrVersion: version, "space_id": spaceID})
	return fleet.Uninstall(ctx, fleetClient, name, version, spaceID, force)
}

func deleteIntegration(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	_ string,
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

	spaceAware := resolveSpaceAware(ctx, client, model.SpaceID, &diags)
	if diags.HasError() {
		return diags
	}

	if spaceAware {
		pkg, getDiags := fleet.GetPackage(ctx, fleetClient, name, version, spaceID)
		diags.Append(getDiags...)
		if diags.HasError() {
			return diags
		}

		if isInstalledInMultipleSpaces(pkg, spaceID) {
			return deleteKibanaAssetsWithFallback(ctx, fleetClient, name, version, spaceID, force)
		}
	}

	uninstallDiags := fleet.Uninstall(ctx, fleetClient, name, version, spaceID, force)
	diags.Append(uninstallDiags...)
	return diags
}

func isInstalledInMultipleSpaces(pkg *kbapi.KibanaHTTPAPIsGetPackageInfo, spaceID string) bool {
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
