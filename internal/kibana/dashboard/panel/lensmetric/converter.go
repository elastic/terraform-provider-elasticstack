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
	return string(kbapi.KibanaHTTPAPIsMetricNoESQLByValuePanelTypeMetric)
}

func (converter) HandlesBlocks(blocks *models.LensByValueChartBlocks) bool {
	return blocks != nil && blocks.MetricChartConfig != nil
}

func (converter) SchemaAttribute() schema.Attribute {
	return lenscommon.ByValueChartNestedAttribute("metric_chart_config", metricChartSchemaAttrs(true))
}

func (converter) PopulateFromAttributes(ctx context.Context, blocks *models.LensByValueChartBlocks, attrs lenscommon.VisByValueConfig0) diag.Diagnostics {
	if diags := lenscommon.ValidateLensBlocks(blocks, "metric_chart_config"); diags.HasError() {
		return diags
	}
	priorConfig := lenscommon.SnapshotAndResetBlock(&blocks.MetricChartConfig)
	if priorConfig != nil {
		blocks.MetricChartConfig.Metrics = priorConfig.Metrics
	}

	return lenscommon.PopulateFromNoESQLOrESQL(
		ctx, blocks.MetricChartConfig, priorConfig,
		attrs.AsKibanaHTTPAPIsMetricNoESQLByValuePanel,
		attrs.AsKibanaHTTPAPIsMetricESQLByValuePanel,
		func(v kbapi.KibanaHTTPAPIsMetricNoESQLByValuePanel) bool {
			return !lenscommon.IsNoESQLCandidateActuallyESQL(v.DataSource)
		},
		metricChartConfigFromAPIVariant0,
		metricChartConfigFromAPIVariant1,
	)
}

func (converter) BuildAttributes(blocks *models.LensByValueChartBlocks) (lenscommon.VisByValueConfig0, diag.Diagnostics) {
	var attrs lenscommon.VisByValueConfig0
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
