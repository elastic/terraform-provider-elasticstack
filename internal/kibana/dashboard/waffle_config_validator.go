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

package dashboard

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.Object = waffleConfigModeValidator{}

// waffleConfigModeValidator enforces consistency between non-ES|QL and ES|QL waffle fields,
// matching heatmap-style ES|QL detection (omit `query` or leave both `query.expression` and `query.language` unset).
type waffleConfigModeValidator struct{}

func (waffleConfigModeValidator) Description(_ context.Context) string {
	return "Ensures waffle_config uses `metrics`/`group_by` for non-ES|QL mode and `esql_metrics`/`esql_group_by` for ES|QL mode."
}

func (v waffleConfigModeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// waffleModeListState describes a Terraform list (or slice at apply time) for waffle mode validation.
// When Unknown is true, count-based rules that depend on that list are skipped (deferred until values are known).
type waffleModeListState struct {
	Count   int
	Unknown bool
}

func waffleModeListStateFromTF(list types.List) waffleModeListState {
	if list.IsUnknown() {
		return waffleModeListState{Unknown: true}
	}
	if list.IsNull() {
		return waffleModeListState{Count: 0}
	}
	return waffleModeListState{Count: len(list.Elements())}
}

// waffleModeListStateFromSlice is used when converting from a parsed model at apply time (lengths are always known).
func waffleModeListStateFromSlice(n int) waffleModeListState {
	return waffleModeListState{Count: n}
}

// waffleConfigModeValidateDiags returns ES|QL vs non-ES|QL waffle field consistency diagnostics.
// If attrPath is non-nil, errors are attribute-scoped (plan-time); if nil, plain errors (e.g. apply-time model conversion).
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

func (v waffleConfigModeValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var queryObj types.Object
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.AtName("query"), &queryObj)...)
	if resp.Diagnostics.HasError() {
		return
	}

	esqlMode := queryObj.IsNull()
	if !esqlMode {
		var lang, qStr types.String
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.AtName("query").AtName("language"), &lang)...)
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.AtName("query").AtName("expression"), &qStr)...)
		if resp.Diagnostics.HasError() {
			return
		}
		esqlMode = lang.IsNull() && qStr.IsNull()
	}

	var metrics, groupBy, esqlMetrics, esqlGroupBy types.List
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.AtName("metrics"), &metrics)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.AtName("group_by"), &groupBy)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.AtName("esql_metrics"), &esqlMetrics)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.AtName("esql_group_by"), &esqlGroupBy)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(waffleConfigModeValidateDiags(esqlMode,
		waffleModeListStateFromTF(metrics),
		waffleModeListStateFromTF(groupBy),
		waffleModeListStateFromTF(esqlMetrics),
		waffleModeListStateFromTF(esqlGroupBy),
		&req.Path,
	)...)
}
