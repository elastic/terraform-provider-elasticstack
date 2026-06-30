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

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/lenswaffle"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ── Gauge ────────────────────────────────────────────────────────────────────

var _ validator.Object = gaugeConfigModeValidator{}

// gaugeConfigModeValidator enforces consistency between non-ES|QL and ES|QL gauge fields,
// matching heatmap-style ES|QL detection (omit `query` or leave both `query.expression` and `query.language` unset).
type gaugeConfigModeValidator struct{}

func (gaugeConfigModeValidator) Description(_ context.Context) string {
	return "Ensures gauge_config uses `metric_json` for non-ES|QL mode and `esql_metric` for ES|QL mode."
}

func (v gaugeConfigModeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// gaugeConfigModeValidateDiags returns ES|QL vs non-ES|QL gauge field consistency diagnostics.
func gaugeConfigModeValidateDiags(esqlMode bool, metricJSON customtypes.JSONWithDefaultsValue[map[string]any], esqlMetricObj types.Object, attrPath *path.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	add := func(summary, detail string) {
		if attrPath != nil {
			diags.AddAttributeError(*attrPath, summary, detail)
		} else {
			diags.AddError(summary, detail)
		}
	}

	hasMetricJSON := typeutils.IsKnown(metricJSON) && !metricJSON.IsNull()
	hasEsqlMetric := typeutils.IsKnown(esqlMetricObj) && !esqlMetricObj.IsNull()

	if esqlMode {
		if hasMetricJSON {
			add(
				"Invalid gauge_config for ES|QL mode",
				"Do not set `metric_json` when using ES|QL mode (omit `query` or leave `query.expression` and `query.language` unset). Use `esql_metric` instead.",
			)
		}
		if !hasEsqlMetric {
			add("Missing esql_metric", "ES|QL gauges require `esql_metric`.")
		}
		return diags
	}

	if hasEsqlMetric {
		add(
			"Invalid gauge_config for non-ES|QL mode",
			"Do not set `esql_metric` when using a non-ES|QL gauge (`query` with both `expression` and `language` set). Use `metric_json` instead, or omit `query` for ES|QL mode.",
		)
	}
	if !hasMetricJSON {
		add(
			"Missing metric_json",
			"Non-ES|QL gauges require `metric_json`. Set it, or omit `query` and provide `esql_metric` for ES|QL mode.",
		)
	}
	return diags
}

func (v gaugeConfigModeValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	esqlMode, ok := lensQueryESQLMode(ctx, req.Config, req.Path, &resp.Diagnostics)
	if !ok {
		return
	}

	var metricJSON customtypes.JSONWithDefaultsValue[map[string]any]
	var esqlMetric types.Object
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.AtName("metric_json"), &metricJSON)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.AtName("esql_metric"), &esqlMetric)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(gaugeConfigModeValidateDiags(esqlMode, metricJSON, esqlMetric, &req.Path)...)
}

// ── Tagcloud ─────────────────────────────────────────────────────────────────

var _ validator.Object = tagcloudConfigModeValidator{}

// tagcloudConfigModeValidator enforces consistency between non-ES|QL and ES|QL tagcloud fields,
// matching heatmap-style ES|QL detection (omit `query` or leave both `query.expression` and `query.language` unset).
type tagcloudConfigModeValidator struct{}

func (tagcloudConfigModeValidator) Description(_ context.Context) string {
	return "Ensures tagcloud_config uses `metric_json`/`tag_by_json` for non-ES|QL mode and `esql_metric`/`esql_tag_by` for ES|QL mode."
}

func (v tagcloudConfigModeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// tagcloudConfigModeValidateDiags returns ES|QL vs non-ES|QL tagcloud field consistency diagnostics.
func tagcloudConfigModeValidateDiags(esqlMode bool, metricJSON, tagByJSON customtypes.JSONWithDefaultsValue[map[string]any], esqlMetric, esqlTagBy types.Object, attrPath *path.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	add := func(summary, detail string) {
		if attrPath != nil {
			diags.AddAttributeError(*attrPath, summary, detail)
		} else {
			diags.AddError(summary, detail)
		}
	}

	hasMetricJSON := typeutils.IsKnown(metricJSON) && !metricJSON.IsNull()
	hasTagByJSON := typeutils.IsKnown(tagByJSON) && !tagByJSON.IsNull()
	hasEsqlMetric := typeutils.IsKnown(esqlMetric) && !esqlMetric.IsNull()
	hasEsqlTagBy := typeutils.IsKnown(esqlTagBy) && !esqlTagBy.IsNull()

	if esqlMode {
		if hasMetricJSON {
			add(
				"Invalid tagcloud_config for ES|QL mode",
				"Do not set `metric_json` when using ES|QL mode (omit `query` or leave `query.expression` and `query.language` unset). Use `esql_metric` instead.",
			)
		}
		if hasTagByJSON {
			add(
				"Invalid tagcloud_config for ES|QL mode",
				"Do not set `tag_by_json` when using ES|QL mode (omit `query` or leave `query.expression` and `query.language` unset). Use `esql_tag_by` instead.",
			)
		}
		if hasEsqlTagBy && !hasEsqlMetric {
			add(
				"Invalid ES|QL tagcloud configuration",
				"`esql_tag_by` requires `esql_metric`. Set both typed ES|QL blocks together.",
			)
			return diags
		}
		if !hasEsqlMetric {
			add("Missing esql_metric", "ES|QL tagclouds require `esql_metric`.")
			return diags
		}
		if !hasEsqlTagBy {
			add("Missing esql_tag_by", "ES|QL tagclouds require `esql_tag_by`.")
		}
		return diags
	}

	if hasEsqlMetric || hasEsqlTagBy {
		add(
			"Invalid tagcloud_config for non-ES|QL mode",
			"Do not set `esql_metric` or `esql_tag_by` when using a non-ES|QL tagcloud "+
				"(`query` with both `expression` and `language` set). "+
				"Use `metric_json` and `tag_by_json` instead, or omit `query` for ES|QL mode.",
		)
	}
	if !hasMetricJSON {
		add(
			"Missing metric_json",
			"Non-ES|QL tagclouds require `metric_json`. Set it, or omit `query` and provide `esql_metric` for ES|QL mode.",
		)
	}
	if !hasTagByJSON {
		add(
			"Missing tag_by_json",
			"Non-ES|QL tagclouds require `tag_by_json`. Set it, or omit `query` and provide `esql_tag_by` for ES|QL mode.",
		)
	}
	return diags
}

func (v tagcloudConfigModeValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	esqlMode, ok := lensQueryESQLMode(ctx, req.Config, req.Path, &resp.Diagnostics)
	if !ok {
		return
	}

	var metricJSON, tagByJSON customtypes.JSONWithDefaultsValue[map[string]any]
	var esqlMetric, esqlTagBy types.Object
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.AtName("metric_json"), &metricJSON)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.AtName("tag_by_json"), &tagByJSON)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.AtName("esql_metric"), &esqlMetric)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.AtName("esql_tag_by"), &esqlTagBy)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(tagcloudConfigModeValidateDiags(esqlMode, metricJSON, tagByJSON, esqlMetric, esqlTagBy, &req.Path)...)
}

// ── Waffle ───────────────────────────────────────────────────────────────────

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

func waffleModeListStateFromTF(list types.List) lenswaffle.WaffleModeListState {
	if list.IsUnknown() {
		return lenswaffle.WaffleModeListState{Unknown: true}
	}
	if list.IsNull() {
		return lenswaffle.WaffleModeListState{Count: 0}
	}
	return lenswaffle.WaffleModeListState{Count: len(list.Elements())}
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

	resp.Diagnostics.Append(lenswaffle.WaffleConfigModeValidateDiags(esqlMode,
		waffleModeListStateFromTF(metrics),
		waffleModeListStateFromTF(groupBy),
		waffleModeListStateFromTF(esqlMetrics),
		waffleModeListStateFromTF(esqlGroupBy),
	)...)
}

// ── Drilldown ────────────────────────────────────────────────────────────────

// drilldownItemModeValidator is an alias for panelkit.DrilldownItemModeValidator kept for
// backward compatibility with internal usages and the existing unit test.
type drilldownItemModeValidator = panelkit.DrilldownItemModeValidator
