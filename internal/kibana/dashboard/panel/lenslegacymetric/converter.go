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

package lenslegacymetric

import (
	"context"
	"maps"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func init() {
	lenscommon.Register(converter{})
}

type converter struct{}

func (converter) VizType() string {
	return string(kbapi.LegacyMetric)
}

func (converter) HandlesBlocks(blocks *models.LensByValueChartBlocks) bool {
	return blocks != nil && blocks.LegacyMetricConfig != nil
}

func (converter) SchemaAttribute() schema.Attribute {
	attrs := lenscommon.LensChartBaseAttributes()
	attrs["data_source_json"] = lenscommon.DataSourceJSONAttribute(
		"Dataset configuration as JSON. Use `dataView` or `index` for standard data sources, and `esql` or `table` for ES|QL sources.",
	)
	attrs["metric_json"] = lenscommon.MetricJSONAttribute(
		"Metric configuration as JSON. For standard datasets, use a metric operation or formula. For ES|QL datasets, include format, operation, column, and color configuration.",
		lenscommon.PopulateLegacyMetricMetricDefaults, true, "",
	)
	attrs["query"] = lenscommon.QueryAttribute(
		"Query configuration for filtering data. Required for non-ES|QL datasets.",
	)
	maps.Copy(attrs, lenscommon.LensChartPresentationAttributes())
	return lenscommon.ByValueChartNestedAttribute("legacy_metric_config", attrs)
}

func (converter) PopulateFromAttributes(ctx context.Context, resolver lenscommon.Resolver, blocks *models.LensByValueChartBlocks, attrs kbapi.KbnDashboardPanelTypeVisConfig0) diag.Diagnostics {
	legacyMetric, err := attrs.AsLegacyMetricNoESQL()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	var prior *models.LegacyMetricConfigModel
	if blocks != nil && blocks.LegacyMetricConfig != nil {
		cpy := *blocks.LegacyMetricConfig
		prior = &cpy
	}
	if blocks == nil {
		var d diag.Diagnostics
		d.AddError("Lens chart blocks missing", "cannot populate legacy_metric_config without chart blocks")
		return d
	}
	blocks.LegacyMetricConfig = &models.LegacyMetricConfigModel{}
	return legacyMetricConfigFromAPINoESQL(ctx, blocks.LegacyMetricConfig, resolver, prior, legacyMetric)
}

func (converter) BuildAttributes(blocks *models.LensByValueChartBlocks, resolver lenscommon.Resolver) (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	var diags diag.Diagnostics
	if blocks == nil {
		return attrs, diags
	}
	return legacyMetricConfigToAPI(blocks.LegacyMetricConfig, resolver)
}

func (converter) AlignStateFromPlan(ctx context.Context, plan, state *models.LensByValueChartBlocks) {
	alignLegacyMetricStateFromPlan(ctx, plan, state)
}

func (converter) PopulateJSONDefaults(attrs map[string]any) map[string]any {
	return populateLegacyMetricLensAttributes(attrs)
}
