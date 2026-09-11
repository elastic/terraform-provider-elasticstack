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

package iface

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// UnionDecodeDiagnostics builds a diagnostic for a failed kbapi union branch decode, so callers
// do not silently lose state during read/refresh. entityName identifies the panel type (for
// example "field_stats_table" or "discover_session") and branch identifies which union variant
// failed to decode (for example "by_esql" or "by-reference").
func UnionDecodeDiagnostics(entityName string, err error, branch string) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.AddError(
		"Failed to decode "+entityName+" API config",
		"Could not decode the API "+entityName+" "+branch+" config: "+err.Error(),
	)
	return diags
}
