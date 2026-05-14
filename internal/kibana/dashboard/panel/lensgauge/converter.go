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

package lensgauge

import (
	"context"
	"maps"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func init() {
	lenscommon.Register(converter{})
}

type converter struct{}

func (converter) VizType() string {
	return string(kbapi.GaugeNoESQLTypeGauge)
}

func (converter) HandlesBlocks(blocks *models.LensByValueChartBlocks) bool {
	return blocks != nil && blocks.GaugeConfig != nil
}

const gaugeMetricMarkdown = "Metric configuration as JSON. Supports metric operations such as count, unique count, " +
	"min, max, average, median, standard deviation, sum, last value, percentile, percentile ranks, or formula."

func gaugeRefAttr(desc string) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: desc,
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"column": schema.StringAttribute{
				MarkdownDescription: "ES|QL column name.",
				Required:            true,
			},
			"label": schema.StringAttribute{
				MarkdownDescription: "Optional label for the operation.",
				Optional:            true,
			},
		},
	}
}

func (converter) SchemaAttribute() schema.Attribute {
	attrs := maps.Clone(lenscommon.LensChartBaseAttributes())
	attrs["data_source_json"] = schema.StringAttribute{
		MarkdownDescription: "Dataset configuration as JSON. For standard layers, this specifies the data view and query.",
		CustomType:          jsontypes.NormalizedType{},
		Required:            true,
	}
	attrs["query"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Query configuration for filtering data. Required for non-ES|QL gauges; omit for ES|QL mode.",
		Optional:            true,
		Attributes:          lenscommon.LensChartFilterSimpleAttributes(),
	}
	attrs["metric_json"] = schema.StringAttribute{
		MarkdownDescription: gaugeMetricMarkdown + " Required for non-ES|QL gauges; mutually exclusive with `esql_metric`.",
		CustomType:          customtypes.NewJSONWithDefaultsType(lenscommon.PopulateGaugeMetricDefaults),
		Optional:            true,
		Validators: []validator.String{
			stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("esql_metric")),
		},
	}
	attrs["esql_metric"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Typed metric column for ES|QL gauges. Mutually exclusive with `metric_json`.",
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"column": schema.StringAttribute{
				MarkdownDescription: "ES|QL column name for the metric.",
				Required:            true,
			},
			"format_json": schema.StringAttribute{
				MarkdownDescription: "Number or other format configuration as JSON (`formatType` union).",
				CustomType:          jsontypes.NormalizedType{},
				Required:            true,
			},
			"label": schema.StringAttribute{
				MarkdownDescription: "Optional label for the metric.",
				Optional:            true,
			},
			"color_json": schema.StringAttribute{
				MarkdownDescription: "Gauge fill color configuration as JSON (`colorByValue`, `noColor`, or `autoColor` union).",
				CustomType:          jsontypes.NormalizedType{},
				Optional:            true,
			},
			"subtitle": schema.StringAttribute{
				MarkdownDescription: "Subtitle text rendered below the gauge value.",
				Optional:            true,
			},
			"goal": gaugeRefAttr("Goal column reference."),
			"max":  gaugeRefAttr("Max column reference."),
			"min":  gaugeRefAttr("Min column reference."),
			"ticks": schema.SingleNestedAttribute{
				MarkdownDescription: "Tick configuration.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"mode": schema.StringAttribute{
						MarkdownDescription: "Tick placement mode.",
						Optional:            true,
					},
					"visible": schema.BoolAttribute{
						MarkdownDescription: "Whether tick marks are displayed.",
						Optional:            true,
					},
				},
			},
			"title": schema.SingleNestedAttribute{
				MarkdownDescription: "Title configuration.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"text": schema.StringAttribute{
						MarkdownDescription: "Title text.",
						Optional:            true,
					},
					"visible": schema.BoolAttribute{
						MarkdownDescription: "Whether the title is displayed.",
						Optional:            true,
					},
				},
			},
		},
	}
	attrs["styling"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Gauge styling configuration.",
		Required:            true,
		Attributes: map[string]schema.Attribute{
			"shape_json": schema.StringAttribute{
				MarkdownDescription: "Gauge shape configuration as JSON. Supports bullet and circular gauges.",
				CustomType:          jsontypes.NormalizedType{},
				Optional:            true,
			},
		},
	}
	maps.Copy(attrs, lenscommon.LensChartPresentationAttributes())
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Typed Lens visualization inside `vis_config.by_value`. " +
			"Mutually exclusive with the other chart blocks in the same `by_value` block. " +
			"Shares the attribute shape with `lens_dashboard_app_config.by_value.gauge_config`.",
		Optional:   true,
		Attributes: attrs,
	}
}

func (converter) PopulateFromAttributes(ctx context.Context, resolver lenscommon.Resolver, blocks *models.LensByValueChartBlocks, attrs kbapi.KbnDashboardPanelTypeVisConfig0) diag.Diagnostics {
	var prior *models.GaugeConfigModel
	if blocks != nil && blocks.GaugeConfig != nil {
		cpy := *blocks.GaugeConfig
		prior = &cpy
	}
	if blocks == nil {
		var d diag.Diagnostics
		d.AddError("Lens chart blocks missing", "cannot populate gauge_config without chart blocks")
		return d
	}
	blocks.GaugeConfig = &models.GaugeConfigModel{}

	if noESQL, err := attrs.AsGaugeNoESQL(); err == nil && !isGaugeNoESQLCandidateActuallyESQL(noESQL) {
		return gaugeConfigFromAPI(ctx, blocks.GaugeConfig, resolver, prior, noESQL)
	}

	esql, err := attrs.AsGaugeESQL()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return gaugeConfigFromAPIESQL(ctx, blocks.GaugeConfig, resolver, prior, esql)
}

func (converter) BuildAttributes(blocks *models.LensByValueChartBlocks, resolver lenscommon.Resolver) (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	var diags diag.Diagnostics
	if blocks == nil {
		return attrs, diags
	}
	return gaugeConfigToAPI(blocks.GaugeConfig, resolver)
}

func (converter) AlignStateFromPlan(ctx context.Context, plan, state *models.LensByValueChartBlocks) {
	alignGaugeStateFromPlan(ctx, plan, state)
}

func (converter) PopulateJSONDefaults(attrs map[string]any) map[string]any {
	return populateGaugeLensAttributes(attrs)
}
