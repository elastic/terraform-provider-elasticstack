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

package transform

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var settingsRequiredVersions map[string]*version.Version

func init() {
	settingsRequiredVersions = make(map[string]*version.Version)

	// capabilities requiring >= 8.8
	settingsRequiredVersions["destination.aliases"] = version.Must(version.NewVersion("8.8.0"))

	// settings requiring >= 8.1
	settingsRequiredVersions["deduce_mappings"] = version.Must(version.NewVersion("8.1.0"))

	// settings requiring >= 8.4
	settingsRequiredVersions["num_failure_retries"] = version.Must(version.NewVersion("8.4.0"))

	// settings requiring >= 8.5
	settingsRequiredVersions["unattended"] = version.Must(version.NewVersion("8.5.0"))
}

// isSettingAllowed returns true when the given setting is supported by the
// connected Elasticsearch server version. When the setting is not supported,
// it logs a warning and returns false.
func isSettingAllowed(ctx context.Context, settingName string, serverVersion *version.Version) bool {
	if minVersion, ok := settingsRequiredVersions[settingName]; ok {
		if serverVersion.LessThan(minVersion) {
			tflog.Warn(ctx, fmt.Sprintf(
				"Setting [%s] not allowed for Elasticsearch server version %v; min required is %v",
				settingName, *serverVersion, *minVersion,
			))
			return false
		}
	}
	return true
}
