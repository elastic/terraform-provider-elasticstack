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
	"github.com/stretchr/testify/require"
)

func TestInstalledInSpace(t *testing.T) {
	t.Parallel()

	statusInstalled := "installed"
	statusNotInstalled := "not_installed"
	desired := "my-space"

	t.Run("nil pkg false", func(t *testing.T) {
		t.Parallel()
		require.False(t, installedInSpace(nil, desired))
	})

	t.Run("status missing false", func(t *testing.T) {
		t.Parallel()
		pkg := &kbapi.PackageInfo{InstallationInfo: &kbapi.PackageInfo_InstallationInfo{}}
		require.False(t, installedInSpace(pkg, desired))
	})

	t.Run("status not installed false", func(t *testing.T) {
		t.Parallel()
		pkg := &kbapi.PackageInfo{Status: &statusNotInstalled, InstallationInfo: &kbapi.PackageInfo_InstallationInfo{}}
		require.False(t, installedInSpace(pkg, desired))
	})

	t.Run("installation info nil false", func(t *testing.T) {
		t.Parallel()
		pkg := &kbapi.PackageInfo{Status: &statusInstalled}
		require.False(t, installedInSpace(pkg, desired))
	})

	t.Run("installed_kibana_space_id matches true", func(t *testing.T) {
		t.Parallel()
		pkg := &kbapi.PackageInfo{
			Status: &statusInstalled,
			InstallationInfo: &kbapi.PackageInfo_InstallationInfo{
				InstalledKibanaSpaceId: &desired,
			},
		}
		require.True(t, installedInSpace(pkg, desired))
	})

	t.Run("additional spaces contains desired true", func(t *testing.T) {
		t.Parallel()
		spaceMap := map[string][]kbapi.PackageInfo_InstallationInfo_AdditionalSpacesInstalledKibana_Item{
			desired: {},
		}
		pkg := &kbapi.PackageInfo{
			Status: &statusInstalled,
			InstallationInfo: &kbapi.PackageInfo_InstallationInfo{
				AdditionalSpacesInstalledKibana: &spaceMap,
			},
		}
		require.True(t, installedInSpace(pkg, desired))
	})

	t.Run("no matches false", func(t *testing.T) {
		t.Parallel()
		other := "other-space"
		spaceMap := map[string][]kbapi.PackageInfo_InstallationInfo_AdditionalSpacesInstalledKibana_Item{
			other: {},
		}
		pkg := &kbapi.PackageInfo{
			Status: &statusInstalled,
			InstallationInfo: &kbapi.PackageInfo_InstallationInfo{
				InstalledKibanaSpaceId:          &other,
				AdditionalSpacesInstalledKibana: &spaceMap,
			},
		}
		require.False(t, installedInSpace(pkg, desired))
	})
}
