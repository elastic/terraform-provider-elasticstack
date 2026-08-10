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
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// spaceScope is the Fleet API routing context for one CRUD operation.
type spaceScope struct {
	id string
}

func supportsSpaceAwareIntegration(ctx context.Context, client clients.MinVersionEnforceable, spaceID string) (bool, diag.Diagnostics) {
	if spaceID == "" {
		return false, nil
	}

	return client.EnforceMinVersion(ctx, MinVersionSpaceAwareIntegration)
}

// resolveSpaceID preserves literal "default": it shares the unscoped API path
// with "", but remains a configured space ID for capability and metadata checks.
func resolveSpaceID(spaceID types.String) string {
	if !typeutils.IsKnown(spaceID) {
		return ""
	}

	return spaceID.ValueString()
}

func resolveSpaceScope(spaceID types.String) spaceScope {
	return spaceScope{id: resolveSpaceID(spaceID)}
}

// fleetPackageInstalledGlobally determines whether Fleet reports a package as fully installed.
// Newer Kibana versions may populate InstallationInfo.install_status instead of (or in addition to) status,
// and status casing can vary.
func fleetPackageInstalledGlobally(pkg *kbapi.KibanaHTTPAPIsGetPackageInfo) bool {
	return fleet.IsPackageInstalled(pkg)
}

func fleetPackageInstalledInSpace(pkg *kbapi.KibanaHTTPAPIsGetPackageInfo, spaceID string) bool {
	return fleetPackageInstalledGlobally(pkg) && packageInstalledInKibanaSpace(pkg.InstallationInfo, spaceID)
}

func packageInstalledInKibanaSpace(info *kbapi.KibanaHTTPAPIsPackageInfoInstallationInfo, spaceID string) bool {
	if info == nil {
		return false
	}
	if info.InstalledKibanaSpaceId != nil && *info.InstalledKibanaSpaceId == spaceID {
		return true
	}
	if info.AdditionalSpacesInstalledKibana != nil {
		if _, ok := (*info.AdditionalSpacesInstalledKibana)[spaceID]; ok {
			return true
		}
	}

	return false
}
