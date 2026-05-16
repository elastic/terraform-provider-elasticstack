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

package lenswaffle

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

// waffleModeListState describes list lengths during waffle mode validation at apply time.
type waffleModeListState struct {
	Count   int
	Unknown bool
}

func waffleModeListStateFromSlice(n int) waffleModeListState {
	return waffleModeListState{Count: n}
}

// waffleConfigModeValidateDiags returns ES|QL vs non-ES|QL waffle field consistency diagnostics.
func waffleConfigModeValidateDiags(esqlMode bool, metrics, groupBy, esqlMetrics, esqlGroupBy waffleModeListState, attrPath *path.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	add := func(summary, detail string) {
		if attrPath != nil {
			diags.AddAttributeError(*attrPath, summary, detail)
		} else {
			diags.AddError(summary, detail)
		}
	}
	if esqlMode {
		if (!metrics.Unknown && metrics.Count > 0) || (!groupBy.Unknown && groupBy.Count > 0) {
			add(
				"Invalid waffle_config for ES|QL mode",
				"Do not set `metrics` or `group_by` when using ES|QL mode (omit `query` or leave `query.expression` and `query.language` unset). Use `esql_metrics` instead.",
			)
		}
		if !esqlMetrics.Unknown && esqlMetrics.Count < 1 {
			add("Missing esql_metrics", "ES|QL waffles require at least one `esql_metrics` entry.")
		}
		return diags
	}

	if (!esqlMetrics.Unknown && esqlMetrics.Count > 0) || (!esqlGroupBy.Unknown && esqlGroupBy.Count > 0) {
		add("Invalid waffle_config for non-ES|QL mode", "Do not set `esql_metrics` or `esql_group_by` when using a non-ES|QL waffle. Set `query` (and use `metrics` / optional `group_by`) instead.")
	}
	if !metrics.Unknown && metrics.Count < 1 {
		add("Missing metrics", "Non-ES|QL waffles require at least one `metrics` entry.")
	}
	return diags
}
