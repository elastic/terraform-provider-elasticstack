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

import "encoding/json"

// EsqlValuesSourceUserValue is the only value the Terraform-facing `by_esql.values_source`
// attribute accepts on control panels that support a by_field/by_esql union (currently
// options_list_control and range_slider_control). It is translated to/from each control's wire
// enum, whose only legal value is "esql" (not "esql_query") — see, e.g.,
// kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsqlValuesSourceEsql —
// "esql_query" reads more clearly next to the `esql_query` attribute and matches the value
// documented in the originating feature request.
const EsqlValuesSourceUserValue = "esql_query"

// IsEsqlBranch reports whether raw (an API control config's raw JSON) discriminates to the ES|QL
// branch of a by_field/by_esql union, identified by the presence of the `esql_query` key, which
// only the ES|QL branch schema defines.
func IsEsqlBranch(raw []byte) bool {
	var probe struct {
		EsqlQuery *string `json:"esql_query"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return false
	}
	return probe.EsqlQuery != nil
}
