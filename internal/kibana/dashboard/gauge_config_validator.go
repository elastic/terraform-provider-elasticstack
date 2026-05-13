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

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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

	hasNonEmptyQueryMode := !esqlMode
	hasEsqlMetric := typeutils.IsKnown(esqlMetricObj) && !esqlMetricObj.IsNull()

	if hasNonEmptyQueryMode && hasEsqlMetric {
		add(
			"Invalid gauge_config fields",
			"Do not set `esql_metric` when using a non-ES|QL gauge (`query` with both `expression` and `language` set). Omit `esql_metric` and use `metric_json`, or omit `query` for ES|QL mode.",
		)
	}

	if esqlMode {
		if typeutils.IsKnown(metricJSON) && !metricJSON.IsNull() {
			add(
				"Invalid gauge_config for ES|QL mode",
				"Do not set `metric_json` when using ES|QL mode (omit `query` or leave `query.expression` and `query.language` unset). Use `esql_metric` instead.",
			)
		}
		if !typeutils.IsKnown(esqlMetricObj) || esqlMetricObj.IsNull() {
			add("Missing esql_metric", "ES|QL gauges require `esql_metric`.")
		}
		return diags
	}

	if hasEsqlMetric {
		add(
			"Invalid gauge_config for non-ES|QL mode",
			"Do not set `esql_metric` when using a non-ES|QL gauge. Set `query` and use `metric_json` instead.",
		)
	}
	return diags
}

func (v gaugeConfigModeValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var queryObj types.Object
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.AtName("query"), &queryObj)...)
	if resp.Diagnostics.HasError() {
		return
	}

	esqlMode := queryObj.IsNull()
	if queryObj.IsUnknown() {
		return
	}
	if !esqlMode {
		var lang, qStr types.String
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.AtName("query").AtName("language"), &lang)...)
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.AtName("query").AtName("expression"), &qStr)...)
		if resp.Diagnostics.HasError() {
			return
		}
		esqlMode = lang.IsNull() && qStr.IsNull()
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
