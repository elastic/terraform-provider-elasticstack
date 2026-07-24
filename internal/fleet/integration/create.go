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
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/asyncutils"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func createIntegration(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	req entitycore.KibanaWriteRequest[integrationModel],
) (entitycore.KibanaWriteResult[integrationModel], diag.Diagnostics) {
	return writeIntegration(ctx, client, req)
}

func writeIntegration(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	req entitycore.KibanaWriteRequest[integrationModel],
) (entitycore.KibanaWriteResult[integrationModel], diag.Diagnostics) {
	var diags diag.Diagnostics

	planModel := req.Plan
	fleetClient := client.GetFleetClient()

	name := planModel.Name.ValueString()
	version := planModel.Version.ValueString()

	scope := resolveSpaceScope(ctx, client, planModel.SpaceID, &diags)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[integrationModel]{}, diags
	}

	installOptions := fleet.InstallPackageOptions{
		Force:             planModel.Force.ValueBool(),
		Prerelease:        planModel.Prerelease.ValueBool(),
		IgnoreConstraints: planModel.IgnoreConstraints.ValueBool(),
	}

	if typeutils.IsKnown(planModel.IgnoreMappingUpdateErrors) {
		installOptions.IgnoreMappingUpdateErrors = planModel.IgnoreMappingUpdateErrors.ValueBoolPointer()
	}

	if typeutils.IsKnown(planModel.SkipDataStreamRollover) {
		installOptions.SkipDataStreamRollover = planModel.SkipDataStreamRollover.ValueBoolPointer()
	}

	if scope.id != "" {
		installOptions.SpaceID = scope.id
	}

	installDiags := fleet.InstallPackage(ctx, fleetClient, name, version, installOptions)
	diags.Append(installDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[integrationModel]{}, diags
	}

	waitErr := waitForFleetIntegrationInstalled(ctx, fleetClient, name, version, scope)
	if waitErr != nil {
		diags.AddError(
			"Failed to install Fleet integration package",
			fmt.Sprintf("Package %s/%s did not reach an installed state: %s", name, version, waitErr.Error()),
		)
		return entitycore.KibanaWriteResult[integrationModel]{}, diags
	}

	pkg, getDiags := fleet.GetPackage(ctx, fleetClient, name, version, scope.id)
	diags.Append(getDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[integrationModel]{}, diags
	}

	globallyInstalled := fleetPackageInstalled(pkg, spaceScope{})
	installedInTargetSpace := fleetPackageInstalled(pkg, spaceScope{id: scope.id, aware: true})
	installedElsewhere := globallyInstalled && scope.id != "" && !installedInTargetSpace

	if installedElsewhere {
		spaceDiags := installInSpace(ctx, client, fleetClient, name, version, scope.id, planModel.Force.ValueBool())
		diags.Append(spaceDiags...)
		if diags.HasError() {
			return entitycore.KibanaWriteResult[integrationModel]{}, diags
		}
	}

	planModel.ID = types.StringValue(getPackageID(name, version))

	if planModel.SpaceID.IsUnknown() {
		planModel.SpaceID = installedKibanaSpaceID(pkg)
	}

	return entitycore.KibanaWriteResult[integrationModel]{Model: planModel}, diags
}

func installedKibanaSpaceID(pkg *kbapi.KibanaHTTPAPIsGetPackageInfo) types.String {
	if pkg == nil || pkg.InstallationInfo == nil {
		return types.StringNull()
	}

	return typeutils.StringishPointerValue(pkg.InstallationInfo.InstalledKibanaSpaceId)
}

func installInSpace(ctx context.Context, client clients.MinVersionEnforceable, fleetClient *fleet.Client, name, version, spaceID string, force bool) diag.Diagnostics {
	var diags diag.Diagnostics

	spaceAware, versionDiags := supportsSpaceAwareIntegration(ctx, client, spaceID)
	diags.Append(versionDiags...)
	if diags.HasError() {
		return diags
	}

	if !spaceAware {
		diags.AddWarning(
			"Package already installed in a different space",
			fmt.Sprintf("Package %s/%s is already installed in a different space. Kibana assets may not be available in space %s "+
				"because the server does not support space-aware asset installation.", name, version, spaceID),
		)
		return diags
	}

	installDiags := fleet.InstallKibanaAssets(ctx, fleetClient, name, version, spaceID, force)
	diags.Append(installDiags...)
	if diags.HasError() {
		return diags
	}

	waitErr := waitForFleetIntegrationInstalled(ctx, fleetClient, name, version, spaceScope{id: spaceID, aware: true})
	if waitErr != nil {
		diags.AddError(
			"Failed to install Fleet integration package",
			fmt.Sprintf("Package %s/%s did not reach an installed state in space %s: %s", name, version, spaceID, waitErr.Error()),
		)
	}

	return diags
}

func waitForFleetIntegrationInstalled(ctx context.Context, fleetClient *fleet.Client, name, version string, scope spaceScope) error {
	return asyncutils.WaitForStateTransition(ctx, "fleet integration", getPackageID(name, version), func(ctx context.Context) (bool, error) {
		pkg, getDiags := fleet.GetPackage(ctx, fleetClient, name, version, scope.id)
		if getDiags.HasError() {
			return false, fmt.Errorf("failed to read package installation status: %v", getDiags)
		}
		if pkg == nil {
			return false, nil
		}

		if fleetPackageInstalled(pkg, scope) {
			return true, nil
		}

		if pkg.InstallationInfo != nil && pkg.InstallationInfo.InstallStatus == kbapi.KibanaHTTPAPIsPackageInfoInstallationInfoInstallStatusInstallFailed {
			return false, fmt.Errorf("package %s/%s installation failed", name, version)
		}
		if pkg.Status != nil && strings.EqualFold(*pkg.Status, "install_failed") {
			return false, fmt.Errorf("package %s/%s installation failed", name, version)
		}

		return false, nil
	})
}
