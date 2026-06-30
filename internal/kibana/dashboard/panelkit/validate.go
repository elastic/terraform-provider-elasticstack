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

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// RejectConfigJSON returns an error diagnostic when pm.ConfigJSON is set (known and non-null),
// which is unsupported for panel types that require a typed config block instead.
// panelType is used solely for the human-readable error message (e.g. "discover_session").
func RejectConfigJSON(pm models.PanelModel, panelType string) diag.Diagnostics {
	if !typeutils.IsKnown(pm.ConfigJSON) || pm.ConfigJSON.IsNull() {
		return nil
	}
	var diags diag.Diagnostics
	diags.AddError(
		"Unsupported panel type for config_json",
		fmt.Sprintf(
			"Panel-level `config_json` is not supported for `%s` panels. Use `%s_config` instead.",
			panelType, panelType,
		),
	)
	return diags
}
