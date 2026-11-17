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

package advancedsettings

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

const (
	settingsRoute    = "/api/kibana/settings"
	jsonSettingsHint = "Settings with non-string values, such as a boolean or a number, belong in json_settings: string_settings always sends its values to Kibana as strings."
)

func sendsStringSetting(plan advancedSettingsModel, changes map[string]any) bool {
	for key := range plan.StringSettings.Elements() {
		if _, sent := changes[key]; sent {
			return true
		}
	}
	return false
}

func withJSONSettingsHint(diags diag.Diagnostics) diag.Diagnostics {
	hinted := make(diag.Diagnostics, 0, len(diags))
	for _, d := range diags {
		if d.Severity() == diag.SeverityError {
			d = diag.NewErrorDiagnostic(d.Summary(), d.Detail()+"\n\n"+jsonSettingsHint)
		}
		hinted = append(hinted, d)
	}
	return hinted
}

func missingSettingsRouteDiagnostic(method string) diag.Diagnostic {
	return diag.NewErrorDiagnostic(
		"Kibana advanced settings route not found",
		fmt.Sprintf("Kibana responded with HTTP 404 to %s %s, so this Kibana deployment does not serve the advanced settings route that this resource needs.", method, settingsRoute),
	)
}
