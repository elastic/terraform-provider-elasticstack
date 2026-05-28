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
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func supportsSpaceAwareIntegration(ctx context.Context, client clients.MinVersionEnforceable, spaceID string) (bool, diag.Diagnostics) {
	if spaceID == "" {
		return false, nil
	}

	return client.EnforceMinVersion(ctx, MinVersionSpaceAwareIntegration)
}

// fleetPackageInstalled determines whether Fleet reports a package as fully installed.
// Newer Kibana versions may populate InstallationInfo.install_status instead of (or in addition to) status,
// and status casing can vary.
func fleetPackageInstalled(pkg *kbapi.KibanaHTTPAPIsGetPackageInfo, spaceID string, spaceAware bool) bool {
	if pkg == nil {
		return false
	}

	globalInstalled := false
	if pkg.InstallationInfo != nil {
		switch pkg.InstallationInfo.InstallStatus {
		case kbapi.KibanaHTTPAPIsInstallationInfoInstallStatusInstalled:
			globalInstalled = true
		case kbapi.KibanaHTTPAPIsInstallationInfoInstallStatusInstallFailed:
			return false
		}
	}
	if !globalInstalled && pkg.Status != nil {
		globalInstalled = strings.EqualFold(*pkg.Status, "installed")
	}
	if !globalInstalled {
		return false
	}

	if !spaceAware || spaceID == "" {
		return true
	}

	return packageInstalledInKibanaSpace(pkg.InstallationInfo, spaceID)
}

func packageInstalledInKibanaSpace(info *kbapi.KibanaHTTPAPIsInstallationInfo, spaceID string) bool {
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
