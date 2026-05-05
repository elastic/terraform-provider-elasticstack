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
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// getXYAxisSchema returns the schema for XY chart axis configuration
func getXYAxisSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"x": schema.SingleNestedAttribute{
			MarkdownDescription: "X-axis (horizontal) configuration.",
			Optional:            true,
			Attributes: map[string]schema.Attribute{
				"title": schema.SingleNestedAttribute{
					MarkdownDescription: "Axis title configuration.",
					Optional:            true,
					Attributes: map[string]schema.Attribute{
						"value": schema.StringAttribute{
							MarkdownDescription: "Axis title text.",
							Optional:            true,
						},
						"visible": schema.BoolAttribute{
							MarkdownDescription: "Whether to show the title.",
							Optional:            true,
							Computed:            true,
						},
					},
				},
				"ticks": schema.BoolAttribute{
					MarkdownDescription: "Whether to show tick marks on the axis.",
					Optional:            true,
					Computed:            true,
				},
				"grid": schema.BoolAttribute{
					MarkdownDescription: "Whether to show grid lines for this axis.",
					Optional:            true,
					Computed:            true,
				},
				"label_orientation": schema.StringAttribute{
					MarkdownDescription: "Orientation of the axis labels.",
					Optional:            true,
					Computed:            true,
					Validators: []validator.String{
						stringvalidator.OneOf("horizontal", "vertical", "angled"),
					},
				},
				"scale": schema.StringAttribute{
					MarkdownDescription: "X-axis scale: linear (numeric), ordinal (categorical), or temporal (dates).",
					Optional:            true,
					Validators: []validator.String{
						stringvalidator.OneOf("linear", "ordinal", "temporal"),
					},
				},
				"domain_json": schema.StringAttribute{
					MarkdownDescription: "Axis domain configuration as JSON. Can be 'fit' mode or 'custom' mode with min, max, and optional fit flags.",
					CustomType:          jsontypes.NormalizedType{},
					Optional:            true,
					Computed:            true,
				},
			},
		},
		"y": schema.SingleNestedAttribute{
			MarkdownDescription: "Primary Y-axis configuration with scale and bounds.",
			Optional:            true,
			Attributes:          getYAxisAttributes(),
		},
		"y2": schema.SingleNestedAttribute{
			MarkdownDescription: "Secondary Y-axis configuration with scale and bounds.",
			Optional:            true,
			Attributes:          getYAxisAttributes(),
		},
	}
}

// getYAxisAttributes returns common Y-axis attributes
func getYAxisAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"title": schema.SingleNestedAttribute{
			MarkdownDescription: "Axis title configuration.",
			Optional:            true,
			Attributes: map[string]schema.Attribute{
				"value": schema.StringAttribute{
					MarkdownDescription: "Axis title text.",
					Optional:            true,
				},
				"visible": schema.BoolAttribute{
					MarkdownDescription: "Whether to show the title.",
					Optional:            true,
					Computed:            true,
				},
			},
		},
		"ticks": schema.BoolAttribute{
			MarkdownDescription: "Whether to show tick marks on the axis.",
			Optional:            true,
		},
		"grid": schema.BoolAttribute{
			MarkdownDescription: "Whether to show grid lines for this axis.",
			Optional:            true,
		},
		"label_orientation": schema.StringAttribute{
			MarkdownDescription: "Orientation of the axis labels.",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("horizontal", "vertical", "angled"),
			},
		},
		"scale": schema.StringAttribute{
			MarkdownDescription: "Y-axis scale type for data transformation.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("time", "linear", "log", "sqrt"),
			},
		},
		"domain_json": schema.StringAttribute{
			MarkdownDescription: "Y-axis domain configuration as JSON. Can be 'fit' mode or 'custom' mode with min, max, and optional fit flags.",
			CustomType:          jsontypes.NormalizedType{},
			Required:            true,
		},
	}
}

// getXYDecorationsSchema returns the schema for XY chart decorations
func getXYDecorationsSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"show_end_zones": schema.BoolAttribute{
			MarkdownDescription: "Show end zones for partial buckets.",
			Optional:            true,
		},
		"show_current_time_marker": schema.BoolAttribute{
			MarkdownDescription: "Show current time marker line.",
			Optional:            true,
		},
		"point_visibility": schema.StringAttribute{
			MarkdownDescription: "Show data points on lines. Valid values are: auto, always, never.",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(dashboardValueAuto, "always", "never"),
			},
		},
		"line_interpolation": schema.StringAttribute{
			MarkdownDescription: "Line interpolation method.",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("linear", "smooth", "stepped"),
			},
		},
		"minimum_bar_height": schema.Int64Attribute{
			MarkdownDescription: "Minimum bar height in pixels.",
			Optional:            true,
		},
		"show_value_labels": schema.BoolAttribute{
			MarkdownDescription: "Display value labels on data points.",
			Optional:            true,
		},
		"fill_opacity": schema.Float64Attribute{
			MarkdownDescription: "Area chart fill opacity (0-1 typical, max 2 for legacy).",
			Optional:            true,
			Computed:            true,
		},
	}
}

// getXYFittingSchema returns the schema for XY chart fitting configuration
func getXYFittingSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"type": schema.StringAttribute{
			MarkdownDescription: "Fitting function type for missing data.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("none", "zero", "linear", "carry", "lookahead", dashboardValueAverage, "nearest"),
			},
		},
		"dotted": schema.BoolAttribute{
			MarkdownDescription: "Show fitted values as dotted lines.",
			Optional:            true,
		},
		"end_value": schema.StringAttribute{
			MarkdownDescription: "How to handle the end value for fitting.",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("none", "zero", "nearest"),
			},
		},
	}
}

// getXYLegendSchema returns the schema for XY chart legend configuration
func getXYLegendSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"visibility": schema.StringAttribute{
			MarkdownDescription: "Legend visibility (auto, visible, hidden).",
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(dashboardValueAuto, "visible", "hidden"),
			},
		},
		"statistics": schema.ListAttribute{
			MarkdownDescription: "Statistics to display in legend (maximum 17).",
			ElementType:         types.StringType,
			Optional:            true,
		},
		"truncate_after_lines": schema.Int64Attribute{
			MarkdownDescription: "Maximum lines before truncating legend items (1-10).",
			Optional:            true,
		},
		"inside": schema.BoolAttribute{
			MarkdownDescription: "Position legend inside the chart. When true, use 'columns' and 'alignment'. When false or omitted, use 'position' and 'size'.",
			Optional:            true,
			Computed:            true,
		},
		"position": schema.StringAttribute{
			MarkdownDescription: "Legend position when positioned outside the chart. Valid when 'inside' is false or omitted.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("top", "bottom", "left", "right"),
			},
		},
		"size": schema.StringAttribute{
			MarkdownDescription: "Legend size when positioned outside the chart. Valid for left/right outside legends. Values use the Kibana API enum: auto, s, m, l, xl.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(dashboardValueAuto, "s", "m", "l", "xl"),
			},
		},
		"columns": schema.Int64Attribute{
			MarkdownDescription: "Number of legend columns when positioned inside the chart (1-5). Valid when 'inside' is true.",
			Optional:            true,
		},
		"alignment": schema.StringAttribute{
			MarkdownDescription: "Legend alignment when positioned inside the chart. Valid when 'inside' is true.",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("top_right", "bottom_right", "top_left", "bottom_left"),
			},
		},
	}
}

// getXYChartConfigAttributes returns attributes for an `xy_chart_config` block (vis panels and lens-dashboard-app by_value).
func getXYChartConfigAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"title": schema.StringAttribute{
			MarkdownDescription: "The title of the chart displayed in the panel.",
			Optional:            true,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "The description of the chart.",
			Optional:            true,
		},
		"axis": schema.SingleNestedAttribute{
			MarkdownDescription: "Axis configuration for X, Y, and secondary Y axes.",
			Required:            true,
			Attributes:          getXYAxisSchema(),
		},
		"decorations": schema.SingleNestedAttribute{
			MarkdownDescription: "Visual enhancements and styling options for the chart.",
			Required:            true,
			Attributes:          getXYDecorationsSchema(),
		},
		"fitting": schema.SingleNestedAttribute{
			MarkdownDescription: "Missing data interpolation configuration. Only valid fitting types are applied per chart type.",
			Required:            true,
			Attributes:          getXYFittingSchema(),
		},
		"layers": schema.ListNestedAttribute{
			MarkdownDescription: "Chart layers configuration. Minimum 1 layer required. Each layer can be a data layer or reference line layer.",
			Required:            true,
			NestedObject:        getXYLayerSchema(),
			Validators: []validator.List{
				listvalidator.SizeAtLeast(1),
			},
		},
		"legend": schema.SingleNestedAttribute{
			MarkdownDescription: "Legend configuration for the XY chart.",
			Required:            true,
			Attributes:          getXYLegendSchema(),
		},
		"query": schema.SingleNestedAttribute{
			MarkdownDescription: "Query configuration for filtering data.",
			Required:            true,
			Attributes:          getFilterSimple(),
		},
		"filters": schema.ListNestedAttribute{
			MarkdownDescription: "Additional filters to apply to the chart data (maximum 100).",
			Optional:            true,
			NestedObject:        getChartFilter(),
		},
	}
}

// getXYLayerSchema returns the schema for XY chart layers
func getXYLayerSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				MarkdownDescription: xyLayerTypeDescription,
				Required:            true,
			},
			"data_layer": schema.SingleNestedAttribute{
				MarkdownDescription: "Configuration for data layers (area, line, bar charts). Mutually exclusive with `reference_line_layer`.",
				Optional:            true,
				Attributes:          getDataLayerAttributes(),
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("reference_line_layer")),
				},
			},
			"reference_line_layer": schema.SingleNestedAttribute{
				MarkdownDescription: "Configuration for reference line layers. Mutually exclusive with `data_layer`.",
				Optional:            true,
				Attributes:          getReferenceLineLayerAttributes(),
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("data_layer")),
				},
			},
		},
	}
}

// getDataLayerAttributes returns attributes for data layers (standard and ES|QL)
func getDataLayerAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"data_source_json": schema.StringAttribute{
			MarkdownDescription: "Dataset configuration as JSON. For ES|QL layers, this specifies the ES|QL query. For standard layers, this specifies the data view and query.",
			CustomType:          jsontypes.NormalizedType{},
			Required:            true,
		},
		"ignore_global_filters": schema.BoolAttribute{
			MarkdownDescription: "If true, ignore global filters when fetching data for this layer. Default is false.",
			Optional:            true,
			Computed:            true,
		},
		"sampling": schema.Float64Attribute{
			MarkdownDescription: "Sampling factor between 0 (no sampling) and 1 (full sampling). Default is 1.",
			Optional:            true,
			Computed:            true,
		},
		"x_json": schema.StringAttribute{
			MarkdownDescription: "X-axis configuration as JSON. For ES|QL: column and operation. For standard: field, operation, and optional parameters.",
			CustomType:          jsontypes.NormalizedType{},
			Optional:            true,
		},
		"y": schema.ListNestedAttribute{
			MarkdownDescription: "Array of Y-axis metrics. Each entry defines a metric to display on the Y-axis.",
			Required:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"config_json": schema.StringAttribute{
						MarkdownDescription: "Y-axis metric configuration as JSON. For ES|QL: axis, color, column, and operation. For standard: axis, color, and metric definition.",
						CustomType:          jsontypes.NormalizedType{},
						Required:            true,
					},
				},
			},
		},
		"breakdown_by_json": schema.StringAttribute{
			MarkdownDescription: "Split series configuration as JSON. For ES|QL: column, operation, optional collapse_by, and color mapping. For standard: field, operation, and optional parameters.",
			CustomType:          jsontypes.NormalizedType{},
			Optional:            true,
		},
	}
}

// getReferenceLineLayerAttributes returns attributes for reference line layers
func getReferenceLineLayerAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"data_source_json": schema.StringAttribute{
			MarkdownDescription: "Dataset configuration as JSON. For ES|QL layers, this specifies the ES|QL query. For standard layers, this specifies the data view and query.",
			CustomType:          jsontypes.NormalizedType{},
			Required:            true,
		},
		"ignore_global_filters": schema.BoolAttribute{
			MarkdownDescription: "If true, ignore global filters when fetching data for this layer. Default is false.",
			Optional:            true,
			Computed:            true,
		},
		"sampling": schema.Float64Attribute{
			MarkdownDescription: "Sampling factor between 0 (no sampling) and 1 (full sampling). Default is 1.",
			Optional:            true,
			Computed:            true,
		},
		"thresholds": schema.ListNestedAttribute{
			MarkdownDescription: "Array of reference line thresholds.",
			Required:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"axis": schema.StringAttribute{
						MarkdownDescription: "Which axis the reference line applies to. Valid values: 'left', 'right'.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("bottom", "left", "right"),
						},
					},
					"color_json": schema.StringAttribute{
						MarkdownDescription: "Color for the reference line. Can be a static color string or dynamic color configuration as JSON.",
						CustomType:          jsontypes.NormalizedType{},
						Optional:            true,
					},
					"column": schema.StringAttribute{
						MarkdownDescription: "Column to use (for ES|QL layers).",
						Optional:            true,
					},
					"value_json": schema.StringAttribute{
						MarkdownDescription: "Metric configuration as JSON (for standard layers). Defines the calculation for the threshold value.",
						CustomType:          jsontypes.NormalizedType{},
						Optional:            true,
					},
					"fill": schema.StringAttribute{
						MarkdownDescription: "Fill direction for reference line. Valid values: 'none', 'above', 'below'.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("none", "above", "below"),
						},
					},
					"icon": schema.StringAttribute{
						MarkdownDescription: referenceLineIconDescription,
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("alert", "asterisk", "bell", "bolt", "bug", "circle", "editorComment", "flag", "heart", "mapMarker", "pinFilled", "starEmpty", "starFilled", "tag", "triangle"),
						},
					},
					"operation": schema.StringAttribute{
						MarkdownDescription: "Operation to apply (for ES|QL: aggregation function; for standard: metric calculation type).",
						Optional:            true,
					},
					"stroke_dash": schema.StringAttribute{
						MarkdownDescription: "Line style. Valid values: 'solid', 'dashed', 'dotted'.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("solid", "dashed", "dotted"),
						},
					},
					"stroke_width": schema.Float64Attribute{
						MarkdownDescription: "Line width in pixels.",
						Optional:            true,
					},
					"text": schema.StringAttribute{
						MarkdownDescription: "Text display option for the reference line. Valid values include: 'auto', 'name', 'none', 'label'.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.OneOf(dashboardValueAuto, "name", "none", "label"),
						},
					},
				},
			},
		},
	}
}
