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

package maintenancewindow

import (
	"fmt"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func validateMaintenanceWindowServer(serverVersion *version.Version, serverFlavor string) diag.Diagnostics {
	var serverlessFlavor = "serverless"
	var maintenanceWindowPublicAPIMinSupportedVersion = version.Must(version.NewVersion("9.1.0"))
	var diags diag.Diagnostics

	if serverVersion.LessThan(maintenanceWindowPublicAPIMinSupportedVersion) && serverFlavor != serverlessFlavor {
		diags.AddError(
			"Maintenance window API not supported",
			fmt.Sprintf(
				`The maintenance Window public API feature requires a minimum Elasticsearch version of "%s" or a serverless Kibana instance.`,
				maintenanceWindowPublicAPIMinSupportedVersion,
			),
		)
		return diags
	}

	return nil
}
