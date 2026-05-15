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

package lenstagcloud

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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func init() {
	lenscommon.Register(converter{})
}

type converter struct{}

func (converter) VizType() string {
	return string(kbapi.TagcloudNoESQLTypeTagCloud)
}

func (converter) HandlesBlocks(blocks *models.LensByValueChartBlocks) bool {
	return blocks != nil && blocks.TagcloudConfig != nil
}

const tagcloudMetricMarkdown = "Metric configuration as JSON. Can be a field metric operation (count, unique count, min, max, " +
	"avg, median, std dev, sum, last value, percentile, percentile ranks), a pipeline operation (differences, moving average, " +
	"cumulative sum, counter rate), or a formula operation."

func (converter) SchemaAttribute() schema.Attribute {
	attrs := lenscommon.LensChartBaseAttributes()
	attrs["data_source_json"] = lenscommon.DataSourceJSONAttribute(
		"Dataset configuration as JSON. For standard layers, this specifies the data view and query.",
	)
	attrs["query"] = lenscommon.QueryAttribute(
		"Query configuration for filtering data. Required for non-ES|QL tagclouds; omit for ES|QL mode.",
	)
	attrs["orientation"] = schema.StringAttribute{
		MarkdownDescription: "Orientation of the tagcloud. Valid values: 'horizontal', 'vertical', 'angled'.",
		Optional:            true,
		Validators: []validator.String{
			stringvalidator.OneOf("horizontal", "vertical", "angled"),
		},
	}
	attrs["font_size"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Minimum and maximum font size for the tags.",
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"min": schema.Float64Attribute{
				MarkdownDescription: "Minimum font size (default: 18, minimum: 1).",
				Optional:            true,
			},
			"max": schema.Float64Attribute{
				MarkdownDescription: "Maximum font size (default: 72, maximum: 120).",
				Optional:            true,
			},
		},
	}
	attrs["metric_json"] = lenscommon.MetricJSONAttribute(
		tagcloudMetricMarkdown+" Required for non-ES|QL tagclouds; mutually exclusive with `esql_metric`.",
		lenscommon.PopulateTagcloudMetricDefaults, false, "esql_metric",
	)
	attrs["tag_by_json"] = schema.StringAttribute{
		MarkdownDescription: "Tag grouping configuration as JSON. Can be a date histogram, terms, histogram, range, or filters operation. " +
			"This determines how tags are grouped and displayed. Required for non-ES|QL tagclouds; mutually exclusive with `esql_tag_by`.",
		CustomType: customtypes.NewJSONWithDefaultsType(lenscommon.PopulateTagcloudTagByDefaults),
		Optional:   true,
		Validators: lenscommon.MutuallyExclusiveStringValidator("esql_tag_by"),
	}
	attrs["esql_metric"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Typed metric column for ES|QL tagclouds. Mutually exclusive with `metric_json`.",
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
		},
	}
	attrs["esql_tag_by"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Typed tag-by column for ES|QL tagclouds. Mutually exclusive with `tag_by_json`.",
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"column": schema.StringAttribute{
				MarkdownDescription: "ES|QL column for the tag dimension.",
				Required:            true,
			},
			"format_json": schema.StringAttribute{
				MarkdownDescription: "Column format as JSON (`formatType` union).",
				CustomType:          jsontypes.NormalizedType{},
				Required:            true,
			},
			"color_json": schema.StringAttribute{
				MarkdownDescription: "Color mapping as JSON (`colorMapping` union).",
				CustomType:          jsontypes.NormalizedType{},
				Required:            true,
			},
			"label": schema.StringAttribute{
				MarkdownDescription: "Optional label for the tag-by column.",
				Optional:            true,
			},
		},
	}
	maps.Copy(attrs, lenscommon.LensChartPresentationAttributes())
	return lenscommon.ByValueChartNestedAttribute("tagcloud_config", attrs)
}

func (converter) PopulateFromAttributes(ctx context.Context, resolver lenscommon.Resolver, blocks *models.LensByValueChartBlocks, attrs kbapi.KbnDashboardPanelTypeVisConfig0) diag.Diagnostics {
	var prior *models.TagcloudConfigModel
	if blocks != nil && blocks.TagcloudConfig != nil {
		cpy := *blocks.TagcloudConfig
		prior = &cpy
	}
	if blocks == nil {
		var d diag.Diagnostics
		d.AddError("Lens chart blocks missing", "cannot populate tagcloud_config without chart blocks")
		return d
	}
	blocks.TagcloudConfig = &models.TagcloudConfigModel{}

	if noESQL, err := attrs.AsTagcloudNoESQL(); err == nil && !isTagcloudNoESQLCandidateActuallyESQL(noESQL) {
		return tagcloudConfigFromAPI(ctx, blocks.TagcloudConfig, resolver, prior, noESQL)
	}

	esql, err := attrs.AsTagcloudESQL()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return tagcloudConfigFromAPIESQL(ctx, blocks.TagcloudConfig, resolver, prior, esql)
}

func (converter) BuildAttributes(blocks *models.LensByValueChartBlocks, resolver lenscommon.Resolver) (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	var diags diag.Diagnostics
	if blocks == nil {
		return attrs, diags
	}
	return tagcloudConfigToAPI(blocks.TagcloudConfig, resolver)
}

func (converter) AlignStateFromPlan(ctx context.Context, plan, state *models.LensByValueChartBlocks) {
	alignTagcloudStateFromPlan(ctx, plan, state)
}

func (converter) PopulateJSONDefaults(attrs map[string]any) map[string]any {
	return populateTagcloudLensAttributes(attrs)
}
