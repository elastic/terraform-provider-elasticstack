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
	return string(kbapi.KibanaHTTPAPIsMetricNoESQLTypeMetric)
}

func (converter) HandlesBlocks(blocks *models.LensByValueChartBlocks) bool {
	return blocks != nil && blocks.MetricChartConfig != nil
}

func (converter) SchemaAttribute() schema.Attribute {
	return lenscommon.ByValueChartNestedAttribute("metric_chart_config", metricChartSchemaAttrs(true))
}

func (converter) PopulateFromAttributes(ctx context.Context, blocks *models.LensByValueChartBlocks, attrs lenscommon.LensByValueConfig) diag.Diagnostics {
	if diags := lenscommon.ValidateLensBlocks(blocks, "metric_chart_config"); diags.HasError() {
		return diags
	}
	priorConfig := lenscommon.SnapshotAndResetBlock(&blocks.MetricChartConfig)
	if priorConfig != nil {
		blocks.MetricChartConfig.Metrics = priorConfig.Metrics
	}

	return lenscommon.PopulateFromNoESQLOrESQL(
		ctx, blocks.MetricChartConfig, priorConfig,
		func() (kbapi.KibanaHTTPAPIsMetricNoESQL, error) {
			chart, err := attrs.Chart.AsKibanaHTTPAPIsMetricChart()
			if err != nil {
				return kbapi.KibanaHTTPAPIsMetricNoESQL{}, err
			}
			return chart.AsKibanaHTTPAPIsMetricNoESQL()
		},
		func() (kbapi.KibanaHTTPAPIsMetricESQL, error) {
			chart, err := attrs.Chart.AsKibanaHTTPAPIsMetricChart()
			if err != nil {
				return kbapi.KibanaHTTPAPIsMetricESQL{}, err
			}
			return chart.AsKibanaHTTPAPIsMetricESQL()
		},
		func(v kbapi.KibanaHTTPAPIsMetricNoESQL) bool {
			return !lenscommon.IsNoESQLCandidateActuallyESQL(v.DataSource)
		},
		func(ctx context.Context, m *models.MetricChartConfigModel, prior *models.MetricChartConfigModel, api kbapi.KibanaHTTPAPIsMetricNoESQL) diag.Diagnostics {
			return metricChartConfigFromAPIVariant0(ctx, m, prior, api, attrs.Presentation)
		},
		func(ctx context.Context, m *models.MetricChartConfigModel, prior *models.MetricChartConfigModel, api kbapi.KibanaHTTPAPIsMetricESQL) diag.Diagnostics {
			return metricChartConfigFromAPIVariant1(ctx, m, prior, api, attrs.Presentation)
		},
	)
}

func (converter) BuildAttributes(blocks *models.LensByValueChartBlocks) (lenscommon.LensByValueConfig, diag.Diagnostics) {
	var attrs lenscommon.LensByValueConfig
	var diags diag.Diagnostics
	if blocks == nil {
		return attrs, diags
	}
	attrs, metricDiags := metricChartConfigToAPI(blocks.MetricChartConfig)
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
