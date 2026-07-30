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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestResolveSpaceID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		spaceID types.String
		want    string
	}{
		{name: "null", spaceID: types.StringNull(), want: ""},
		{name: "unknown", spaceID: types.StringUnknown(), want: ""},
		{name: "empty", spaceID: types.StringValue(""), want: ""},
		{name: "default", spaceID: types.StringValue("default"), want: "default"},
		{name: "custom", spaceID: types.StringValue("target-space"), want: "target-space"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, resolveSpaceID(tt.spaceID))
		})
	}
}

func TestFleetPackageInstalled(t *testing.T) {
	t.Parallel()

	installedStatus := kbapi.KibanaHTTPAPIsPackageInfoInstallationInfoInstallStatusInstalled
	failedStatus := kbapi.KibanaHTTPAPIsPackageInfoInstallationInfoInstallStatusInstallFailed
	otherSpace := "other-space"
	targetSpace := "target-space"
	installedStatusStr := "installed"

	tests := []struct {
		name        string
		pkg         *kbapi.KibanaHTTPAPIsGetPackageInfo
		spaceID     string
		wantGlobal  bool
		wantInSpace bool
	}{
		{
			name: "nil package",
		},
		{
			name: "install failed via install status",
			pkg: &kbapi.KibanaHTTPAPIsGetPackageInfo{
				InstallationInfo: &kbapi.KibanaHTTPAPIsPackageInfoInstallationInfo{
					InstallStatus: failedStatus,
				},
			},
		},
		{
			name: "globally installed default scope",
			pkg: &kbapi.KibanaHTTPAPIsGetPackageInfo{
				InstallationInfo: &kbapi.KibanaHTTPAPIsPackageInfoInstallationInfo{
					InstallStatus: installedStatus,
				},
			},
			wantGlobal: true,
		},
		{
			name: "globally installed via status field",
			pkg: &kbapi.KibanaHTTPAPIsGetPackageInfo{
				Status: &installedStatusStr,
			},
			wantGlobal: true,
		},
		{
			name: "space check rejects primary install in another space",
			pkg: &kbapi.KibanaHTTPAPIsGetPackageInfo{
				InstallationInfo: &kbapi.KibanaHTTPAPIsPackageInfoInstallationInfo{
					InstallStatus:          installedStatus,
					InstalledKibanaSpaceId: &otherSpace,
				},
			},
			spaceID:    targetSpace,
			wantGlobal: true,
		},
		{
			name: "space check accepts primary install in target space",
			pkg: &kbapi.KibanaHTTPAPIsGetPackageInfo{
				InstallationInfo: &kbapi.KibanaHTTPAPIsPackageInfoInstallationInfo{
					InstallStatus:          installedStatus,
					InstalledKibanaSpaceId: &targetSpace,
				},
			},
			spaceID:     targetSpace,
			wantGlobal:  true,
			wantInSpace: true,
		},
		{
			name: "space check accepts additional space entry",
			pkg: &kbapi.KibanaHTTPAPIsGetPackageInfo{
				InstallationInfo: &kbapi.KibanaHTTPAPIsPackageInfoInstallationInfo{
					InstallStatus: installedStatus,
					AdditionalSpacesInstalledKibana: &map[string][]kbapi.KibanaHTTPAPIsPackageInfoKibanaAssetReference{
						targetSpace: {},
					},
				},
			},
			spaceID:     targetSpace,
			wantGlobal:  true,
			wantInSpace: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.wantGlobal, fleetPackageInstalledGlobally(tt.pkg))
			assert.Equal(t, tt.wantInSpace, fleetPackageInstalledInSpace(tt.pkg, tt.spaceID))
		})
	}
}
