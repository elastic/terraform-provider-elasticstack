package dashboard

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/validators"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = getSchema()
}

func getSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages Kibana [dashboards](https://www.elastic.co/docs/api/doc/kibana). This functionality is in technical preview and may be changed or removed in a future release.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Generated composite identifier for the dashboard.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				MarkdownDescription: "An identifier for the space. If space_id is not provided, the default space is used.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("default"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"dashboard_id": schema.StringAttribute{
				MarkdownDescription: "A unique identifier for the dashboard. If not provided, one will be generated.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "A human-readable title for the dashboard.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A short description of the dashboard.",
				Optional:            true,
			},
			"time_from": schema.StringAttribute{
				MarkdownDescription: "The start time for the dashboard's time range (e.g., 'now-15m', '2023-01-01T00:00:00Z').",
				Required:            true,
			},
			"time_to": schema.StringAttribute{
				MarkdownDescription: "The end time for the dashboard's time range (e.g., 'now', '2023-12-31T23:59:59Z').",
				Required:            true,
			},
			"time_range_mode": schema.StringAttribute{
				MarkdownDescription: "The time range mode. Valid values are 'absolute' or 'relative'.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("absolute", "relative"),
				},
			},
			"refresh_interval_pause": schema.BoolAttribute{
				MarkdownDescription: "Set to false to auto-refresh data on an interval.",
				Required:            true,
			},
			"refresh_interval_value": schema.Int64Attribute{
				MarkdownDescription: "A numeric value indicating refresh frequency in milliseconds.",
				Required:            true,
			},
			"query_language": schema.StringAttribute{
				MarkdownDescription: "The query language (e.g., 'kuery', 'lucene').",
				Required:            true,
			},
			"query_text": schema.StringAttribute{
				MarkdownDescription: "The query text for text-based queries such as Kibana Query Language (KQL) or Lucene query language. Mutually exclusive with `query_json`.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("query_json")),
				},
			},
			"query_json": schema.StringAttribute{
				MarkdownDescription: "The query as a JSON object for structured queries. Mutually exclusive with `query_text`.",
				CustomType:          jsontypes.NormalizedType{},
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("query_text")),
				},
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "An array of tag IDs applied to this dashboard.",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"options": schema.SingleNestedAttribute{
				MarkdownDescription: "Display options for the dashboard.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"hide_panel_titles": schema.BoolAttribute{
						MarkdownDescription: "Hide the panel titles in the dashboard.",
						Optional:            true,
					},
					"use_margins": schema.BoolAttribute{
						MarkdownDescription: "Show margins between panels in the dashboard layout.",
						Optional:            true,
					},
					"sync_colors": schema.BoolAttribute{
						MarkdownDescription: "Synchronize colors between related panels in the dashboard.",
						Optional:            true,
					},
					"sync_tooltips": schema.BoolAttribute{
						MarkdownDescription: "Synchronize tooltips between related panels in the dashboard.",
						Optional:            true,
					},
					"sync_cursor": schema.BoolAttribute{
						MarkdownDescription: "Synchronize cursor position between related panels in the dashboard.",
						Optional:            true,
					},
				},
			},
			"panels": schema.ListNestedAttribute{
				MarkdownDescription: "The panels to display in the dashboard.",
				Optional:            true,
				NestedObject:        getPanelSchema(),
			},
			"sections": schema.ListNestedAttribute{
				MarkdownDescription: "Sections organize panels into collapsible groups. This is a technical preview feature.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"title": schema.StringAttribute{
							MarkdownDescription: "The title of the section.",
							Required:            true,
						},
						"id": schema.StringAttribute{
							MarkdownDescription: "The unique identifier of the section.",
							Optional:            true,
							Computed:            true,
						},
						"collapsed": schema.BoolAttribute{
							MarkdownDescription: "The collapsed state of the section.",
							Optional:            true,
						},
						"grid": schema.SingleNestedAttribute{
							MarkdownDescription: "The grid coordinates of the section.",
							Required:            true,
							Attributes: map[string]schema.Attribute{
								"y": schema.Int64Attribute{
									MarkdownDescription: "The Y coordinate.",
									Required:            true,
								},
							},
						},
						"panels": schema.ListNestedAttribute{
							MarkdownDescription: "The panels to display in the section.",
							Optional:            true,
							NestedObject:        getPanelSchema(),
						},
					},
				},
			},
			"access_control": schema.SingleNestedAttribute{
				MarkdownDescription: "Access control parameters for the dashboard.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"access_mode": schema.StringAttribute{
						MarkdownDescription: "The access mode for the dashboard (e.g., 'write_restricted', 'default').",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("write_restricted", "default"),
						},
					},
					"owner": schema.StringAttribute{
						MarkdownDescription: "The owner of the dashboard.",
						Optional:            true,
					},
				},
			},
		},
	}
}

func getPanelSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of the panel (e.g. 'DASHBOARD_MARKDOWN', 'lens').",
				Required:            true,
			},
			"grid": schema.SingleNestedAttribute{
				MarkdownDescription: "The grid coordinates and dimensions of the panel.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"x": schema.Int64Attribute{
						MarkdownDescription: "The X coordinate.",
						Required:            true,
					},
					"y": schema.Int64Attribute{
						MarkdownDescription: "The Y coordinate.",
						Required:            true,
					},
					"w": schema.Int64Attribute{
						MarkdownDescription: "The width.",
						Optional:            true,
					},
					"h": schema.Int64Attribute{
						MarkdownDescription: "The height.",
						Optional:            true,
					},
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the panel.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseNonNullStateForUnknown(),
				},
			},
			"markdown_config": schema.SingleNestedAttribute{
				MarkdownDescription: "The configuration of a markdown panel. Mutually exclusive with `config_json` and `xy_chart_config`.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"content": schema.StringAttribute{
						MarkdownDescription: "The content of the panel.",
						Optional:            true,
					},
					"description": schema.StringAttribute{
						MarkdownDescription: "The description of the panel.",
						Optional:            true,
					},
					"hide_panel_titles": schema.BoolAttribute{
						MarkdownDescription: "Hide the panel titles.",
						Optional:            true,
					},
					"title": schema.StringAttribute{
						MarkdownDescription: "The title of the panel.",
						Optional:            true,
					},
				},
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(
						path.MatchRelative().AtParent().AtName("config_json"),
						path.MatchRelative().AtParent().AtName("xy_chart_config"),
					),
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{"DASHBOARD_MARKDOWN"}),
				},
			},
			"xy_chart_config": schema.SingleNestedAttribute{
				MarkdownDescription: "Configuration for an XY chart panel. Mutually exclusive with `markdown_config` and `config_json`. Use this for line, area, and bar charts.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"title": schema.StringAttribute{
						MarkdownDescription: "The title of the chart displayed in the panel.",
						Optional:            true,
					},
					"description": schema.StringAttribute{
						MarkdownDescription: "The description of the chart.",
						Optional:            true,
					},
					"axis": schema.SingleNestedAttribute{
						MarkdownDescription: "Axis configuration for X, left Y, and right Y axes.",
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
						Attributes:          getFilterSimpleSchema(),
					},
					"filters": schema.ListNestedAttribute{
						MarkdownDescription: "Additional filters to apply to the chart data (maximum 100).",
						Optional:            true,
						NestedObject:        getSearchFilterSchema(),
					},
				},
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(
						path.MatchRelative().AtParent().AtName("markdown_config"),
						path.MatchRelative().AtParent().AtName("config_json"),
					),
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{"lens"}),
				},
			},
			"config_json": schema.StringAttribute{
				MarkdownDescription: "The configuration of the panel as a JSON string. Mutually exclusive with `markdown_config` and `xy_chart_config`.",
				CustomType:          jsontypes.NormalizedType{},
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(
						path.MatchRelative().AtParent().AtName("markdown_config"),
						path.MatchRelative().AtParent().AtName("xy_chart_config"),
					),
				},
			},
		},
	}
}

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
				"extent": schema.StringAttribute{
					MarkdownDescription: "Axis extent configuration as JSON. Can be 'full' mode with optional integer_rounding, or 'custom' mode with start, end, and optional integer_rounding.",
					CustomType:          jsontypes.NormalizedType{},
					Optional:            true,
				},
			},
		},
		"left": schema.SingleNestedAttribute{
			MarkdownDescription: "Left Y-axis configuration with scale and bounds.",
			Optional:            true,
			Attributes:          getYAxisAttributes(),
		},
		"right": schema.SingleNestedAttribute{
			MarkdownDescription: "Right Y-axis configuration with scale and bounds.",
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
			Validators: []validator.String{
				stringvalidator.OneOf("time", "linear", "log", "sqrt"),
			},
		},
		"extent": schema.StringAttribute{
			MarkdownDescription: "Y-axis extent configuration as JSON. Can be 'full' or 'focus' mode, or 'custom' mode with start and end values.",
			CustomType:          jsontypes.NormalizedType{},
			Optional:            true,
		},
	}
}

// getXYDecorationsSchema returns the schema for XY chart decorations
func getXYDecorationsSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"end_zones": schema.BoolAttribute{
			MarkdownDescription: "Show end zones for partial buckets.",
			Optional:            true,
		},
		"current_time_marker": schema.BoolAttribute{
			MarkdownDescription: "Show current time marker line.",
			Optional:            true,
		},
		"point_visibility": schema.BoolAttribute{
			MarkdownDescription: "Show data points on lines.",
			Optional:            true,
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
		},
		"value_labels": schema.BoolAttribute{
			MarkdownDescription: "Show value labels (alternative property).",
			Optional:            true,
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
				stringvalidator.OneOf("none", "zero", "linear", "carry", "lookahead", "average", "nearest"),
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
		"visible": schema.BoolAttribute{
			MarkdownDescription: "Whether to show the legend.",
			Optional:            true,
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
		},
		"position": schema.StringAttribute{
			MarkdownDescription: "Legend position when positioned outside the chart. Valid when 'inside' is false or omitted.",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("top", "bottom", "left", "right"),
			},
		},
		"size": schema.StringAttribute{
			MarkdownDescription: "Legend size when positioned outside the chart. Valid when 'inside' is false or omitted.",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("small", "medium", "large", "xlarge"),
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

// getFilterSimpleSchema returns the schema for simple filter configuration
func getFilterSimpleSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"language": schema.StringAttribute{
			MarkdownDescription: "Query language (default: 'kuery').",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("kuery", "lucene"),
			},
		},
		"query": schema.StringAttribute{
			MarkdownDescription: "Filter query string.",
			Required:            true,
		},
	}
}

// getSearchFilterSchema returns the schema for search filter configuration
func getSearchFilterSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"query": schema.StringAttribute{
				MarkdownDescription: "Filter query string or JSON object.",
				Optional:            true,
			},
			"meta": schema.StringAttribute{
				MarkdownDescription: "Filter metadata as JSON.",
				CustomType:          jsontypes.NormalizedType{},
				Optional:            true,
			},
			"language": schema.StringAttribute{
				MarkdownDescription: "Query language.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("kuery", "lucene"),
				},
			},
		},
	}
}

// getXYLayerSchema returns the schema for XY chart layers
func getXYLayerSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of layer. Valid values: 'area', 'line', 'bar', 'horizontal_bar', 'reference_line' for NoESQL layers; 'area_chart', 'line_chart', 'bar_chart', 'horizontal_bar_chart', 'reference_line' for ESQL layers.",
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
		"dataset": schema.StringAttribute{
			MarkdownDescription: "Dataset configuration as JSON. For ES|QL layers, this specifies the ES|QL query. For standard layers, this specifies the data view and query.",
			CustomType:          jsontypes.NormalizedType{},
			Required:            true,
		},
		"ignore_global_filters": schema.BoolAttribute{
			MarkdownDescription: "If true, ignore global filters when fetching data for this layer. Default is false.",
			Optional:            true,
		},
		"sampling": schema.Float64Attribute{
			MarkdownDescription: "Sampling factor between 0 (no sampling) and 1 (full sampling). Default is 1.",
			Optional:            true,
		},
		"x": schema.StringAttribute{
			MarkdownDescription: "X-axis configuration as JSON. For ES|QL: column and operation. For standard: field, operation, and optional parameters.",
			CustomType:          jsontypes.NormalizedType{},
			Optional:            true,
		},
		"y": schema.ListNestedAttribute{
			MarkdownDescription: "Array of Y-axis metrics. Each entry defines a metric to display on the Y-axis.",
			Required:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"config": schema.StringAttribute{
						MarkdownDescription: "Y-axis metric configuration as JSON. For ES|QL: axis, color, column, and operation. For standard: axis, color, and metric definition.",
						CustomType:          jsontypes.NormalizedType{},
						Required:            true,
					},
				},
			},
		},
		"breakdown_by": schema.StringAttribute{
			MarkdownDescription: "Split series configuration as JSON. For ES|QL: column and operation. For standard: field, operation, and optional parameters.",
			CustomType:          jsontypes.NormalizedType{},
			Optional:            true,
		},
	}
}

// getReferenceLineLayerAttributes returns attributes for reference line layers
func getReferenceLineLayerAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"dataset": schema.StringAttribute{
			MarkdownDescription: "Dataset configuration as JSON. For ES|QL layers, this specifies the ES|QL query. For standard layers, this specifies the data view and query.",
			CustomType:          jsontypes.NormalizedType{},
			Required:            true,
		},
		"ignore_global_filters": schema.BoolAttribute{
			MarkdownDescription: "If true, ignore global filters when fetching data for this layer. Default is false.",
			Optional:            true,
		},
		"sampling": schema.Float64Attribute{
			MarkdownDescription: "Sampling factor between 0 (no sampling) and 1 (full sampling). Default is 1.",
			Optional:            true,
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
							stringvalidator.OneOf("left", "right"),
						},
					},
					"color": schema.StringAttribute{
						MarkdownDescription: "Color for the reference line. Can be a static color string or dynamic color configuration as JSON.",
						CustomType:          jsontypes.NormalizedType{},
						Required:            true,
					},
					"column": schema.StringAttribute{
						MarkdownDescription: "Column to use (for ES|QL layers).",
						Optional:            true,
					},
					"value": schema.StringAttribute{
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
						MarkdownDescription: "Icon to display on the reference line. Valid values: 'asterisk', 'bell', 'bolt', 'bug', 'circle', 'dot', 'empty', 'flag', 'heart', 'partial', 'square', 'star', 'tag', 'triangle'.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("asterisk", "bell", "bolt", "bug", "circle", "dot", "empty", "flag", "heart", "partial", "square", "star", "tag", "triangle"),
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
						MarkdownDescription: "Text display option for the reference line. Valid values: 'auto', 'name'.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("auto", "name"),
						},
					},
				},
			},
		},
	}
}
