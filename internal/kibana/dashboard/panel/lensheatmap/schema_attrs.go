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

package lensheatmap

import (
	"maps"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func heatmapSchemaAttrs(includePresentation bool) map[string]schema.Attribute {
	attrs := lenscommon.LensChartBaseAttributes()
	attrs["data_source_json"] = lenscommon.DataSourceJSONAttribute(
		"Dataset configuration as JSON. For standard heatmaps, this specifies the data view or index; for ES|QL, this specifies the ES|QL query dataset.",
	)
	attrs["query"] = lenscommon.QueryAttribute(
		"Query configuration for filtering data. Required for non-ES|QL heatmaps.",
	)
	attrs["axis"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Axis configuration for X and Y axes.",
		Required:            true,
		Attributes:          heatmapAxesSchemaAttrs(),
	}
	attrs["legend"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Legend configuration for the heatmap.",
		Required:            true,
		Attributes:          heatmapLegendSchemaAttrs(),
	}
	attrs["x_axis_json"] = schema.StringAttribute{
		MarkdownDescription: "Breakdown dimension configuration for the X axis as JSON. This specifies the operation (e.g., `terms`, `date_histogram`, `histogram`, `range`, `filters`) and its parameters.",
		CustomType:          jsontypes.NormalizedType{},
		Required:            true,
	}
	attrs["y_axis_json"] = schema.StringAttribute{
		MarkdownDescription: "Breakdown dimension configuration for the Y axis as JSON. When omitted, the heatmap renders without a Y breakdown.",
		CustomType:          jsontypes.NormalizedType{},
		Optional:            true,
	}
	attrs["metric_json"] = lenscommon.MetricJSONAttribute(
		"Metric configuration as JSON. For non-ES|QL, this can be a field metric, pipeline metric, or formula. For ES|QL, this is the metric column/operation/color configuration.",
		lenscommon.PopulateTagcloudMetricDefaults, true, "",
	)
	attrs["styling"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Heatmap styling configuration.",
		Required:            true,
		Attributes: map[string]schema.Attribute{
			"cells": schema.SingleNestedAttribute{
				MarkdownDescription: "Cells configuration for the heatmap.",
				Required:            true,
				Attributes:          heatmapCellsSchemaAttrs(),
			},
		},
	}
	if includePresentation {
		maps.Copy(attrs, lenscommon.LensChartPresentationAttributes())
	}
	return attrs
}

func heatmapAxesSchemaAttrs() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"x": schema.SingleNestedAttribute{
			MarkdownDescription: "X-axis configuration.",
			Optional:            true,
			Attributes: map[string]schema.Attribute{
				"labels": schema.SingleNestedAttribute{
					MarkdownDescription: "X-axis label configuration.",
					Optional:            true,
					Attributes: map[string]schema.Attribute{
						"orientation": schema.StringAttribute{
							MarkdownDescription: "Orientation of the axis labels.",
							Optional:            true,
							Validators: []validator.String{
								stringvalidator.OneOf("horizontal", "vertical", "angled"),
							},
						},
						"visible": schema.BoolAttribute{
							MarkdownDescription: "Whether to show axis labels.",
							Optional:            true,
						},
					},
				},
				"title": lenscommon.AxisTitleAttribute(false),
			},
		},
		"y": schema.SingleNestedAttribute{
			MarkdownDescription: "Y-axis configuration.",
			Optional:            true,
			Attributes: map[string]schema.Attribute{
				"labels": schema.SingleNestedAttribute{
					MarkdownDescription: "Y-axis label configuration.",
					Optional:            true,
					Attributes: map[string]schema.Attribute{
						"visible": schema.BoolAttribute{
							MarkdownDescription: "Whether to show axis labels.",
							Optional:            true,
						},
					},
				},
				"title": lenscommon.AxisTitleAttribute(false),
			},
		},
	}
}

// getHeatmapCellsSchema returns schema for heatmap cells configuration
func heatmapCellsSchemaAttrs() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"labels": schema.SingleNestedAttribute{
			MarkdownDescription: "Cell label configuration.",
			Optional:            true,
			Attributes: map[string]schema.Attribute{
				"visible": schema.BoolAttribute{
					MarkdownDescription: "Whether to show cell labels.",
					Optional:            true,
				},
			},
		},
	}
}

// getHeatmapLegendSchema returns schema for heatmap legend configuration
func heatmapLegendSchemaAttrs() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"visibility": schema.StringAttribute{
			MarkdownDescription: "Legend visibility. Valid values are `visible` or `hidden`.",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("visible", "hidden"),
			},
		},
		"size": schema.StringAttribute{
			MarkdownDescription: "Legend size: auto, s, m, l, or xl.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("auto", "s", "m", "l", "xl"),
			},
		},
		"truncate_after_lines": schema.Int64Attribute{
			MarkdownDescription: "Maximum lines before truncating legend items (1-10).",
			Optional:            true,
		},
	}
}
