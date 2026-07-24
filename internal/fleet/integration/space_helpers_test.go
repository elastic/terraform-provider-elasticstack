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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/stretchr/testify/assert"
)

func TestFleetPackageInstalled(t *testing.T) {
	t.Parallel()

	installedStatus := kbapi.KibanaHTTPAPIsPackageInfoInstallationInfoInstallStatusInstalled
	failedStatus := kbapi.KibanaHTTPAPIsPackageInfoInstallationInfoInstallStatusInstallFailed
	otherSpace := "other-space"
	targetSpace := "target-space"
	installedStatusStr := "installed"

	tests := []struct {
		name  string
		pkg   *kbapi.KibanaHTTPAPIsGetPackageInfo
		scope spaceScope
		want  bool
	}{
		{
			name:  "nil package",
			pkg:   nil,
			scope: spaceScope{},
			want:  false,
		},
		{
			name: "install failed via install status",
			pkg: &kbapi.KibanaHTTPAPIsGetPackageInfo{
				InstallationInfo: &kbapi.KibanaHTTPAPIsPackageInfoInstallationInfo{
					InstallStatus: failedStatus,
				},
			},
			scope: spaceScope{},
			want:  false,
		},
		{
			name: "globally installed default scope",
			pkg: &kbapi.KibanaHTTPAPIsGetPackageInfo{
				InstallationInfo: &kbapi.KibanaHTTPAPIsPackageInfoInstallationInfo{
					InstallStatus: installedStatus,
				},
			},
			scope: spaceScope{},
			want:  true,
		},
		{
			name: "globally installed via status field",
			pkg: &kbapi.KibanaHTTPAPIsGetPackageInfo{
				Status: &installedStatusStr,
			},
			scope: spaceScope{},
			want:  true,
		},
		{
			name: "scoped API path with aware false ignores mismatched space metadata",
			pkg: &kbapi.KibanaHTTPAPIsGetPackageInfo{
				InstallationInfo: &kbapi.KibanaHTTPAPIsPackageInfoInstallationInfo{
					InstallStatus:          installedStatus,
					InstalledKibanaSpaceId: &otherSpace,
				},
			},
			scope: spaceScope{id: targetSpace, aware: false},
			want:  true,
		},
		{
			name: "strict space aware rejects primary install in another space",
			pkg: &kbapi.KibanaHTTPAPIsGetPackageInfo{
				InstallationInfo: &kbapi.KibanaHTTPAPIsPackageInfoInstallationInfo{
					InstallStatus:          installedStatus,
					InstalledKibanaSpaceId: &otherSpace,
				},
			},
			scope: spaceScope{id: targetSpace, aware: true},
			want:  false,
		},
		{
			name: "strict space aware accepts primary install in target space",
			pkg: &kbapi.KibanaHTTPAPIsGetPackageInfo{
				InstallationInfo: &kbapi.KibanaHTTPAPIsPackageInfoInstallationInfo{
					InstallStatus:          installedStatus,
					InstalledKibanaSpaceId: &targetSpace,
				},
			},
			scope: spaceScope{id: targetSpace, aware: true},
			want:  true,
		},
		{
			name: "strict space aware accepts additional space entry",
			pkg: &kbapi.KibanaHTTPAPIsGetPackageInfo{
				InstallationInfo: &kbapi.KibanaHTTPAPIsPackageInfoInstallationInfo{
					InstallStatus: installedStatus,
					AdditionalSpacesInstalledKibana: &map[string][]kbapi.KibanaHTTPAPIsPackageInfoKibanaAssetReference{
						targetSpace: {},
					},
				},
			},
			scope: spaceScope{id: targetSpace, aware: true},
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, fleetPackageInstalled(tt.pkg, tt.scope))
		})
	}
}
