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

package panelkit

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// PanelDecodeDiagnostics builds a diagnostic for a failed kbapi union branch decode so callers do
// not silently lose state during read/refresh.
func PanelDecodeDiagnostics(panelType, branch string, err error) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.AddError(
		"Failed to decode "+panelType+" API config",
		"Could not decode the API "+panelType+" "+branch+" config: "+err.Error(),
	)
	return diags
}

// PanelProbeDiagnostics builds a diagnostic for a failure to probe the kbapi union discriminator
// field, distinct from a missing or unexpected discriminator value.
func PanelProbeDiagnostics(panelType string, err error) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.AddError(
		"Failed to decode "+panelType+" API config",
		fmt.Sprintf("Could not determine the %s view_type: %s.", panelType, err.Error()),
	)
	return diags
}
