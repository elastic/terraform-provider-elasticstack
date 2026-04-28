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

package integrationpolicy

import (
	"context"
	"fmt"
	"sync"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	fleetpkg "github.com/elastic/terraform-provider-elasticstack/internal/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/resourcecore"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                 = newIntegrationPolicyResource()
	_ resource.ResourceWithConfigure    = newIntegrationPolicyResource()
	_ resource.ResourceWithImportState  = newIntegrationPolicyResource()
	_ resource.ResourceWithUpgradeState = newIntegrationPolicyResource()
)

var (
	MinVersionPolicyIDs = version.Must(version.NewVersion("8.15.0"))
	MinVersionOutputID  = version.Must(version.NewVersion("8.16.0"))
)

type integrationPolicyResource struct {
	*resourcecore.Core
	*fleetpkg.SpaceImporter
}

func newIntegrationPolicyResource() *integrationPolicyResource {
	return &integrationPolicyResource{
		Core:          resourcecore.New(resourcecore.ComponentFleet, "integration_policy"),
		SpaceImporter: fleetpkg.NewSpaceImporter(path.Root("policy_id")),
	}
}

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return newIntegrationPolicyResource()
}

func (r *integrationPolicyResource) UpgradeState(context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {PriorSchema: getSchemaV0(), StateUpgrader: upgradeV0ToV2},
		1: {PriorSchema: getSchemaV1(), StateUpgrader: upgradeV1ToV2},
	}
}

func (r *integrationPolicyResource) buildFeatures(ctx context.Context, apiClient *clients.KibanaScopedClient) (features, diag.Diagnostics) {
	supportsPolicyIDs, diags := apiClient.EnforceMinVersion(ctx, MinVersionPolicyIDs)
	if diags.HasError() {
		return features{}, diagutil.FrameworkDiagsFromSDK(diags)
	}

	supportsOutputID, outputIDDiags := apiClient.EnforceMinVersion(ctx, MinVersionOutputID)
	if outputIDDiags.HasError() {
		return features{}, diagutil.FrameworkDiagsFromSDK(outputIDDiags)
	}

	return features{
		SupportsPolicyIDs: supportsPolicyIDs,
		SupportsOutputID:  supportsOutputID,
	}, nil
}

var knownPackages sync.Map

func getPackageCacheKey(name, version string) string {
	return fmt.Sprintf("%s-%s", name, version)
}

func getPackageInfo(ctx context.Context, client *fleet.Client, name, version, spaceID string) (*kbapi.PackageInfo, diag.Diagnostics) {
	var diags diag.Diagnostics

	if pkg, ok := getCachedPackageInfo(name, version); ok {
		return &pkg, diags
	}

	// Try the exact version first; fall back to no version (returns the installed
	// package) when the requested version has been removed from the registry.
	pkg, diags := fleet.GetPackage(ctx, client, name, version, spaceID)
	if diags.HasError() {
		return nil, diags
	}
	if pkg == nil {
		diags.AddWarning(
			"Package version not found",
			fmt.Sprintf("Package '%s' version '%s' was not found in the registry. "+
				"Using the installed package version instead. Input defaults may differ. "+
				"Consider updating integration_version to an available version.", name, version),
		)
		var fallbackDiags diag.Diagnostics
		pkg, fallbackDiags = fleet.GetPackage(ctx, client, name, "", spaceID)
		diags.Append(fallbackDiags...)
		if diags.HasError() {
			return nil, diags
		}
	}
	if pkg == nil {
		diags.AddWarning(
			"Package not found",
			fmt.Sprintf("Package '%s' was not found in the registry. Input defaults may be unavailable. Consider updating integration_version to an available version.", name),
		)
		return nil, diags
	}
	knownPackages.Store(getPackageCacheKey(name, version), *pkg)
	return pkg, diags
}

func getCachedPackageInfo(name, version string) (kbapi.PackageInfo, bool) {
	value, ok := knownPackages.Load(getPackageCacheKey(name, version))
	if !ok {
		return kbapi.PackageInfo{}, false
	}
	pkg, ok := value.(kbapi.PackageInfo)
	if !ok {
		return kbapi.PackageInfo{}, false
	}
	return pkg, true
}
