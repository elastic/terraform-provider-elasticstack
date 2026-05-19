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

package lensmetric

import (
	"context"

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
	return string(kbapi.MetricNoESQLTypeMetric)
}

func (converter) HandlesBlocks(blocks *models.LensByValueChartBlocks) bool {
	return blocks != nil && blocks.MetricChartConfig != nil
}

func (converter) SchemaAttribute() schema.Attribute {
	return lenscommon.ByValueChartNestedAttribute("metric_chart_config", metricChartSchemaAttrs(true))
}

func (converter) PopulateFromAttributes(ctx context.Context, resolver lenscommon.Resolver, blocks *models.LensByValueChartBlocks, attrs kbapi.KbnDashboardPanelTypeVisConfig0) diag.Diagnostics {
	var priorConfig *models.MetricChartConfigModel
	if blocks != nil && blocks.MetricChartConfig != nil {
		cpy := *blocks.MetricChartConfig
		priorConfig = &cpy
	}
	if blocks == nil {
		var d diag.Diagnostics
		d.AddError("Lens chart blocks missing", "cannot populate metric_chart_config without chart blocks")
		return d
	}
	blocks.MetricChartConfig = &models.MetricChartConfigModel{}
	if priorConfig != nil {
		blocks.MetricChartConfig.Metrics = priorConfig.Metrics
	}

	if variant0, err := attrs.AsMetricNoESQL(); err == nil && !isMetricNoESQLCandidateActuallyESQL(variant0) {
		return metricChartConfigFromAPIVariant0(ctx, blocks.MetricChartConfig, resolver, priorConfig, variant0)
	}
	variant1, err := attrs.AsMetricESQL()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return metricChartConfigFromAPIVariant1(ctx, blocks.MetricChartConfig, resolver, priorConfig, variant1)
}

func (converter) BuildAttributes(blocks *models.LensByValueChartBlocks, resolver lenscommon.Resolver) (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	var diags diag.Diagnostics
	if blocks == nil {
		return attrs, diags
	}
	attrs, metricDiags := metricChartConfigToAPI(blocks.MetricChartConfig, resolver)
	diags.Append(metricDiags...)
	return attrs, diags
}

func (converter) AlignStateFromPlan(_ context.Context, plan, state *models.LensByValueChartBlocks) {
	if plan == nil || state == nil {
		return
	}
	if plan.MetricChartConfig == nil || state.MetricChartConfig == nil {
		return
	}
	alignMetricStateFromPlan(plan.MetricChartConfig, state.MetricChartConfig)
}

func (converter) PopulateJSONDefaults(attrs map[string]any) map[string]any {
	return populateMetricChartLensAttributes(attrs)
}
