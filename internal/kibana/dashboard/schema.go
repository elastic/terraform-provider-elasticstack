package dashboard

import (
	"context"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
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

var panelConfigNames = []string{
	"markdown_config",
	"config_json",
	"xy_chart_config",
	"tagcloud_config",
	"region_map_config",
	"legacy_metric_config",
	"gauge_config",
	"metric_chart_config",
	"datatable_config",
	"heatmap_config",
}

func panelConfigPaths(names []string) []path.Expression {
	paths := make([]path.Expression, 0, len(names))
	for _, name := range names {
		paths = append(paths, path.MatchRelative().AtName(name))
	}
	return paths
}

func siblingPanelConfigPathsExcept(name string, names []string) []path.Expression {
	paths := make([]path.Expression, 0, len(names)-1)
	for _, n := range names {
		if n == name {
			continue
		}
		paths = append(paths, path.MatchRelative().AtParent().AtName(n))
	}
	return paths
}

func panelConfigDescription(base, self string, names []string) string {
	others := make([]string, 0, len(names)-1)
	for _, name := range names {
		if name == self {
			continue
		}
		others = append(others, "`"+name+"`")
	}
	if len(others) == 0 {
		return base
	}
	return base + " Mutually exclusive with " + strings.Join(others, ", ") + "."
}

// populateTagcloudMetricDefaults populates default values for tagcloud metric configuration
func populateTagcloudMetricDefaults(model map[string]any) map[string]any {
	if model == nil {
		return model
	}
	// Set defaults for all field metric operations
	if operation, ok := model["operation"].(string); ok {
		// These operations all have the same defaults
		switch operation {
		case "count", "unique_count", "min", "max", "average", "median", "standard_deviation", "sum", "last_value", "percentile", "percentile_rank":
			if _, exists := model["empty_as_null"]; !exists {
				model["empty_as_null"] = false
			}
			if _, exists := model["show_metric_label"]; !exists {
				model["show_metric_label"] = true
			}
		}
	}
	return model
}

// populateMetricChartMetricDefaults populates default values for metric chart metric configuration
func populateMetricChartMetricDefaults(model map[string]any) map[string]any {
	if model == nil {
		return model
	}

	// Set defaults for format
	if format, ok := model["format"].(map[string]any); ok {
		if format["type"] == "number" || format["type"] == "percent" {
			if _, exists := format["compact"]; !exists {
				format["compact"] = false
			}
			if _, exists := format["decimals"]; !exists {
				format["decimals"] = float64(2)
			}
		}
	}

	// Set defaults for all metric types
	if _, exists := model["empty_as_null"]; !exists {
		model["empty_as_null"] = false
	}
	if _, exists := model["fit"]; !exists {
		model["fit"] = false
	}

	// Set defaults for icon alignment if icon exists
	if icon, ok := model["icon"].(map[string]any); ok {
		if _, exists := icon["align"]; !exists {
			icon["align"] = "left"
		}
	}

	// Set defaults for alignments if present
	if alignments, ok := model["alignments"].(map[string]any); ok {
		if _, exists := alignments["value"]; !exists {
			alignments["value"] = "right"
		}
	}

	return model
}

// populateTagcloudTagByDefaults populates default values for tagcloud tag_by configuration
func populateTagcloudTagByDefaults(model map[string]any) map[string]any {
	if model == nil {
		return model
	}
	// Set defaults for terms operation
	if operation, ok := model["operation"].(string); ok && operation == "terms" {
		if _, exists := model["rank_by"]; !exists {
			model["rank_by"] = map[string]any{
				"type":      "column",
				"metric":    0,
				"direction": "desc",
			}
		}
	}
	return model
}

// populateLegacyMetricMetricDefaults populates default values for legacy metric operations
func populateLegacyMetricMetricDefaults(model map[string]any) map[string]any {
	if model == nil {
		return model
	}
	if operation, ok := model["operation"].(string); ok {
		switch operation {
		case "count", "unique_count", "min", "max", "average", "median", "standard_deviation", "sum", "last_value", "percentile", "percentile_rank":
			if _, exists := model["empty_as_null"]; !exists {
				model["empty_as_null"] = false
			}
		}
	}

	format, ok := model["format"].(map[string]any)
	if ok {
		if formatType, ok := format["type"].(string); ok {
			switch formatType {
			case "number", "percent":
				if _, exists := format["decimals"]; !exists {
					format["decimals"] = float64(2)
				}
				if _, exists := format["compact"]; !exists {
					format["compact"] = false
				}
			case "bytes", "bits":
				if _, exists := format["decimals"]; !exists {
					format["decimals"] = float64(2)
				}
			}
		}
		model["format"] = format
	}

	return model
}

// populateGaugeMetricDefaults populates default values for gauge metric configuration
func populateGaugeMetricDefaults(model map[string]any) map[string]any {
	if model == nil {
		return model
	}

	if _, exists := model["empty_as_null"]; !exists {
		model["empty_as_null"] = false
	}
	if _, exists := model["hide_title"]; !exists {
		model["hide_title"] = false
	}
	if _, exists := model["ticks"]; !exists {
		model["ticks"] = "auto"
	}

	return model
}

// populateRegionMapMetricDefaults populates default values for region map metric configuration
func populateRegionMapMetricDefaults(model map[string]any) map[string]any {
	if model == nil {
		return model
	}
	if operation, ok := model["operation"].(string); ok {
		switch operation {
		case "count", "unique_count", "min", "max", "average", "median", "standard_deviation", "sum", "last_value", "percentile", "percentile_rank":
			if _, exists := model["empty_as_null"]; !exists {
				model["empty_as_null"] = false
			}
			if _, exists := model["show_metric_label"]; !exists {
				model["show_metric_label"] = true
			}
		}
	}
	return model
}

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
		Validators: []validator.Object{
			objectvalidator.AtLeastOneOf(
				panelConfigPaths(panelConfigNames)...,
			),
		},
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
				MarkdownDescription: panelConfigDescription("The configuration of a markdown panel.", "markdown_config", panelConfigNames),
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
					"hide_title": schema.BoolAttribute{
						MarkdownDescription: "Hide the title of the panel.",
						Optional:            true,
					},
					"title": schema.StringAttribute{
						MarkdownDescription: "The title of the panel.",
						Optional:            true,
					},
				},
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(
						siblingPanelConfigPathsExcept("markdown_config", panelConfigNames)...,
					),
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{"DASHBOARD_MARKDOWN"}),
				},
			},
			"xy_chart_config": schema.SingleNestedAttribute{
				MarkdownDescription: panelConfigDescription("Configuration for an XY chart panel. Use this for line, area, and bar charts.", "xy_chart_config", panelConfigNames),
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
						siblingPanelConfigPathsExcept("xy_chart_config", panelConfigNames)...,
					),
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{"lens"}),
				},
			},
			"datatable_config": schema.SingleNestedAttribute{
				MarkdownDescription: panelConfigDescription("Configuration for a datatable chart panel.", "datatable_config", panelConfigNames),
				Optional:            true,
				Attributes:          getDatatableSchema(),
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(
						siblingPanelConfigPathsExcept("datatable_config", panelConfigNames)...,
					),
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{"lens"}),
				},
			},
			"tagcloud_config": schema.SingleNestedAttribute{
				MarkdownDescription: panelConfigDescription("Configuration for a tagcloud chart panel. Tag clouds visualize word frequency.", "tagcloud_config", panelConfigNames),
				Optional:            true,
				Attributes:          getTagcloudSchema(),
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(
						siblingPanelConfigPathsExcept("tagcloud_config", panelConfigNames)...,
					),
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{"lens"}),
				},
			},
			"heatmap_config": schema.SingleNestedAttribute{
				MarkdownDescription: panelConfigDescription("Configuration for a heatmap chart panel.", "markdown_config", panelConfigNames),
				Optional:            true,
				Attributes:          getHeatmapSchema(),
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(
						siblingPanelConfigPathsExcept("heatmap_config", panelConfigNames)...,
					),
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{"lens"}),
				},
			},
			"region_map_config": schema.SingleNestedAttribute{
				MarkdownDescription: panelConfigDescription("Configuration for a region map chart panel. Use this for geographic region maps.", "region_map_config", panelConfigNames),
				Optional:            true,
				Attributes:          getRegionMapSchema(),
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(
						siblingPanelConfigPathsExcept("region_map_config", panelConfigNames)...,
					),
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{"lens"}),
				},
			},
			"gauge_config": schema.SingleNestedAttribute{
				MarkdownDescription: panelConfigDescription("Configuration for a gauge chart panel.", "gauge_config", panelConfigNames),
				Optional:            true,
				Attributes:          getGaugeSchema(),
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(
						siblingPanelConfigPathsExcept("gauge_config", panelConfigNames)...,
					),
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{"lens"}),
				},
			},
			"metric_chart_config": schema.SingleNestedAttribute{
				MarkdownDescription: panelConfigDescription("Configuration for a metric chart panel. Metric charts display key performance indicators.", "metric_chart_config", panelConfigNames),
				Optional:            true,
				Attributes:          getMetricChartSchema(),
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(
						siblingPanelConfigPathsExcept("metric_chart_config", panelConfigNames)...,
					),
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{"lens"}),
				},
			},
			"legacy_metric_config": schema.SingleNestedAttribute{
				MarkdownDescription: panelConfigDescription("Configuration for a legacy metric chart panel. Use this for legacy single-value metric visualizations.", "legacy_metric_config", panelConfigNames),
				Optional:            true,
				Attributes:          getLegacyMetricSchema(),
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(
						siblingPanelConfigPathsExcept("legacy_metric_config", panelConfigNames)...,
					),
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{"lens"}),
				},
			},
			"config_json": schema.StringAttribute{
				MarkdownDescription: panelConfigDescription("The configuration of the panel as a JSON string.", "config_json", panelConfigNames),
				CustomType:          jsontypes.NormalizedType{},
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(
						siblingPanelConfigPathsExcept("config_json", panelConfigNames)...,
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
				stringvalidator.OneOf("auto", "always", "never"),
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
		"visibility": schema.StringAttribute{
			MarkdownDescription: "Legend visibility (auto, visible, hidden).",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("auto", "visible", "hidden"),
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
				MarkdownDescription: "Query language. Defaults to `kuery` if not specified.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("kuery"),
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
				MarkdownDescription: "The type of layer. Valid values: 'area', 'line', 'bar', 'horizontal_bar', 'referenceLines' for NoESQL layers; 'area_chart', 'line_chart', 'bar_chart', 'horizontal_bar_chart', 'referenceLines' for ESQL layers.",
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
							stringvalidator.OneOf("bottom", "left", "right"),
						},
					},
					"color": schema.StringAttribute{
						MarkdownDescription: "Color for the reference line. Can be a static color string or dynamic color configuration as JSON.",
						CustomType:          jsontypes.NormalizedType{},
						Optional:            true,
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
						MarkdownDescription: "Text display option for the reference line. Valid values include: 'auto', 'name', 'none', 'label'.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("auto", "name", "none", "label"),
						},
					},
				},
			},
		},
	}
}

// getTagcloudSchema returns the schema for tagcloud chart configuration
func getTagcloudSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"title": schema.StringAttribute{
			MarkdownDescription: "The title of the chart displayed in the panel.",
			Optional:            true,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "The description of the chart.",
			Optional:            true,
		},
		"dataset": schema.StringAttribute{
			MarkdownDescription: "Dataset configuration as JSON. For standard layers, this specifies the data view and query.",
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
		"orientation": schema.StringAttribute{
			MarkdownDescription: "Orientation of the tagcloud. Valid values: 'horizontal', 'vertical', 'angled'.",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("horizontal", "vertical", "angled"),
			},
		},
		"font_size": schema.SingleNestedAttribute{
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
		},
		"metric": schema.StringAttribute{
			MarkdownDescription: "Metric configuration as JSON. Can be a field metric operation (count, unique count, min, max, avg, median, std dev, sum, last value, percentile, percentile ranks), a pipeline operation (differences, moving average, cumulative sum, counter rate), or a formula operation.",
			CustomType:          customtypes.NewJSONWithDefaultsType(populateTagcloudMetricDefaults),
			Required:            true,
		},
		"tag_by": schema.StringAttribute{
			MarkdownDescription: "Tag grouping configuration as JSON. Can be a date histogram, terms, histogram, range, or filters operation. This determines how tags are grouped and displayed.",
			CustomType:          customtypes.NewJSONWithDefaultsType(populateTagcloudTagByDefaults),
			Required:            true,
		},
	}
}

// getHeatmapSchema returns the schema for heatmap chart configuration
func getHeatmapSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"title": schema.StringAttribute{
			MarkdownDescription: "The title of the chart displayed in the panel.",
			Optional:            true,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "The description of the chart.",
			Optional:            true,
		},
		"dataset": schema.StringAttribute{
			MarkdownDescription: "Dataset configuration as JSON. For standard heatmaps, this specifies the data view or index; for ES|QL, this specifies the ES|QL query dataset.",
			CustomType:          jsontypes.NormalizedType{},
			Required:            true,
		},
		"ignore_global_filters": schema.BoolAttribute{
			MarkdownDescription: "If true, ignore global filters when fetching data for this chart. Default is false.",
			Optional:            true,
		},
		"sampling": schema.Float64Attribute{
			MarkdownDescription: "Sampling factor between 0 (no sampling) and 1 (full sampling). Default is 1.",
			Optional:            true,
		},
		"query": schema.SingleNestedAttribute{
			MarkdownDescription: "Query configuration for filtering data. Required for non-ES|QL heatmaps.",
			Optional:            true,
			Attributes:          getFilterSimpleSchema(),
		},
		"filters": schema.ListNestedAttribute{
			MarkdownDescription: "Additional filters to apply to the chart data (maximum 100).",
			Optional:            true,
			NestedObject:        getSearchFilterSchema(),
		},
		"axes": schema.SingleNestedAttribute{
			MarkdownDescription: "Axis configuration for X and Y axes.",
			Required:            true,
			Attributes:          getHeatmapAxesSchema(),
		},
		"cells": schema.SingleNestedAttribute{
			MarkdownDescription: "Cells configuration for the heatmap.",
			Required:            true,
			Attributes:          getHeatmapCellsSchema(),
		},
		"legend": schema.SingleNestedAttribute{
			MarkdownDescription: "Legend configuration for the heatmap.",
			Required:            true,
			Attributes:          getHeatmapLegendSchema(),
		},
		"metric": schema.StringAttribute{
			MarkdownDescription: "Metric configuration as JSON. For non-ES|QL, this can be a field metric, pipeline metric, or formula. For ES|QL, this is the metric column/operation/color configuration.",
			CustomType:          customtypes.NewJSONWithDefaultsType(populateTagcloudMetricDefaults),
			Required:            true,
		},
		"x_axis": schema.StringAttribute{
			MarkdownDescription: "X-axis operation configuration as JSON. For non-ES|QL, this can be date histogram, terms, histogram, range, or filters operations; for ES|QL, this is the column/operation configuration.",
			CustomType:          jsontypes.NormalizedType{},
			Required:            true,
		},
		"y_axis": schema.StringAttribute{
			MarkdownDescription: "Y-axis operation configuration as JSON. For non-ES|QL, this can be date histogram, terms, histogram, range, or filters operations; for ES|QL, this is the column/operation configuration.",
			CustomType:          jsontypes.NormalizedType{},
			Optional:            true,
		},
	}
}

// getHeatmapAxesSchema returns schema for heatmap axes configuration
func getHeatmapAxesSchema() map[string]schema.Attribute {
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
				"title": schema.SingleNestedAttribute{
					MarkdownDescription: "X-axis title configuration.",
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
				"title": schema.SingleNestedAttribute{
					MarkdownDescription: "Y-axis title configuration.",
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
			},
		},
	}
}

// getHeatmapCellsSchema returns schema for heatmap cells configuration
func getHeatmapCellsSchema() map[string]schema.Attribute {
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
func getHeatmapLegendSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"visible": schema.BoolAttribute{
			MarkdownDescription: "Whether to show the legend.",
			Optional:            true,
		},
		"position": schema.StringAttribute{
			MarkdownDescription: "Legend position.",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("top", "bottom", "left", "right"),
			},
		},
		"size": schema.StringAttribute{
			MarkdownDescription: "Legend size: auto, small, medium, large, or xlarge.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("auto", "small", "medium", "large", "xlarge"),
			},
		},
		"truncate_after_lines": schema.Int64Attribute{
			MarkdownDescription: "Maximum lines before truncating legend items (1-10).",
			Optional:            true,
		},
	}
}

// getRegionMapSchema returns the schema for region map chart configuration
func getRegionMapSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"title": schema.StringAttribute{
			MarkdownDescription: "The title of the chart displayed in the panel.",
			Optional:            true,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "The description of the chart.",
			Optional:            true,
		},
		"dataset": schema.StringAttribute{
			MarkdownDescription: "Dataset configuration as JSON. For ES|QL, this specifies the ES|QL query. For standard layers, this specifies the data view and query.",
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
		"query": schema.SingleNestedAttribute{
			MarkdownDescription: "Query configuration for filtering data. Required for non-ES|QL region map configurations.",
			Optional:            true,
			Attributes:          getFilterSimpleSchema(),
		},
		"filters": schema.ListNestedAttribute{
			MarkdownDescription: "Additional filters to apply to the chart data (maximum 100).",
			Optional:            true,
			NestedObject:        getSearchFilterSchema(),
		},
		"metric": schema.StringAttribute{
			MarkdownDescription: "Metric configuration as JSON. For ES|QL, this defines the metric column and format. For standard mode, this defines the metric operation or formula.",
			CustomType:          customtypes.NewJSONWithDefaultsType(populateRegionMapMetricDefaults),
			Required:            true,
		},
		"region": schema.StringAttribute{
			MarkdownDescription: "Region configuration as JSON. For ES|QL, this defines the region column and EMS join. For standard mode, this defines the bucket operation (terms, histogram, range, filters) and optional EMS settings.",
			CustomType:          jsontypes.NormalizedType{},
			Required:            true,
		},
	}
}

// getLegacyMetricSchema returns the schema for legacy metric chart configuration
func getLegacyMetricSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"title": schema.StringAttribute{
			MarkdownDescription: "The title of the chart displayed in the panel.",
			Optional:            true,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "The description of the chart.",
			Optional:            true,
		},
		"dataset": schema.StringAttribute{
			MarkdownDescription: "Dataset configuration as JSON. Use `dataView` or `index` for standard data sources, and `esql` or `table` for ES|QL sources.",
			CustomType:          jsontypes.NormalizedType{},
			Required:            true,
		},
		"metric": schema.StringAttribute{
			MarkdownDescription: "Metric configuration as JSON. For standard datasets, use a metric operation or formula. For ES|QL datasets, include format, operation, column, and color configuration.",
			CustomType:          customtypes.NewJSONWithDefaultsType(populateLegacyMetricMetricDefaults),
			Required:            true,
		},
		"query": schema.SingleNestedAttribute{
			MarkdownDescription: "Query configuration for filtering data. Required for non-ES|QL datasets.",
			Optional:            true,
			Attributes:          getFilterSimpleSchema(),
		},
		"filters": schema.ListNestedAttribute{
			MarkdownDescription: "Additional filters to apply to the chart data (maximum 100).",
			Optional:            true,
			NestedObject:        getSearchFilterSchema(),
		},
		"sampling": schema.Float64Attribute{
			MarkdownDescription: "Sampling factor between 0 (no sampling) and 1 (full sampling). Default is 1.",
			Optional:            true,
		},
		"ignore_global_filters": schema.BoolAttribute{
			MarkdownDescription: "If true, ignore global filters when fetching data for this panel. Default is false.",
			Optional:            true,
		},
	}
}

// getGaugeSchema returns the schema for gauge chart configuration
func getGaugeSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"title": schema.StringAttribute{
			MarkdownDescription: "The title of the chart displayed in the panel.",
			Optional:            true,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "The description of the chart.",
			Optional:            true,
		},
		"dataset": schema.StringAttribute{
			MarkdownDescription: "Dataset configuration as JSON. For standard layers, this specifies the data view and query.",
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
		"metric": schema.StringAttribute{
			MarkdownDescription: "Metric configuration as JSON. Supports metric operations such as count, unique count, min, max, average, median, standard deviation, sum, last value, percentile, percentile ranks, or formula.",
			CustomType:          customtypes.NewJSONWithDefaultsType(populateGaugeMetricDefaults),
			Required:            true,
		},
		"shape": schema.StringAttribute{
			MarkdownDescription: "Gauge shape configuration as JSON. Supports bullet and circular gauges.",
			CustomType:          jsontypes.NormalizedType{},
			Optional:            true,
		},
	}
}

// getMetricChartSchema returns the schema for metric chart configuration
func getMetricChartSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"title": schema.StringAttribute{
			MarkdownDescription: "The title of the chart displayed in the panel.",
			Optional:            true,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "The description of the chart.",
			Optional:            true,
		},
		"dataset": schema.StringAttribute{
			MarkdownDescription: "Dataset configuration as JSON. Can be a data view dataset (`type: 'dataview'`), index dataset (`type: 'index'`), ES|QL dataset (`type: 'esql'`), or table ES|QL dataset (`type: 'tableESQLDatasetType'`).",
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
		"query": schema.SingleNestedAttribute{
			MarkdownDescription: "Query configuration for filtering data. Required for non-ES|QL datasets.",
			Optional:            true,
			Attributes:          getFilterSimpleSchema(),
		},
		"filters": schema.ListNestedAttribute{
			MarkdownDescription: "Additional filters to apply to the chart data (maximum 100).",
			Optional:            true,
			NestedObject:        getSearchFilterSchema(),
		},
		"metrics": schema.ListNestedAttribute{
			MarkdownDescription: "Array of metrics to display (1-2 items). Each metric can be a primary metric (displays prominently) or secondary metric (displays as comparison). Metrics can use field operations (count, unique count, min, max, avg, median, std dev, sum, last value, percentile, percentile ranks), pipeline operations (differences, moving average, cumulative sum, counter rate), formula operations, or for ES|QL datasets, column-based value operations.",
			Required:            true,
			Validators: []validator.List{
				listvalidator.SizeAtMost(2),
			},
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"config": schema.StringAttribute{
						MarkdownDescription: "Metric configuration as JSON. For primary metrics: includes type ('primary'), operation, format, alignments, icon, and optional fields like sub_label, fit, color, apply_color_to, and background_chart. For secondary metrics: includes type ('secondary'), operation, format, and optional fields like label, prefix, compare, and color.",
						CustomType:          customtypes.NewJSONWithDefaultsType(populateMetricChartMetricDefaults),
						Required:            true,
					},
				},
			},
		},
		"breakdown_by": schema.StringAttribute{
			MarkdownDescription: "Breakdown configuration as JSON. Groups metrics by a dimension. Can use operations like date histogram, terms, histogram, range, filters, or for ES|QL datasets, value operations with columns. Includes optional columns count and collapse_by configuration.",
			CustomType:          jsontypes.NormalizedType{},
			Optional:            true,
		},
	}
}

// getDatatableSchema returns the schema for datatable chart configuration
func getDatatableSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"no_esql": schema.SingleNestedAttribute{
			MarkdownDescription: "Datatable configuration for standard (non-ES|QL) queries.",
			Optional:            true,
			Attributes:          getDatatableNoESQLSchema(),
			Validators: []validator.Object{
				objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("esql")),
			},
		},
		"esql": schema.SingleNestedAttribute{
			MarkdownDescription: "Datatable configuration for ES|QL queries.",
			Optional:            true,
			Attributes:          getDatatableESQLSchema(),
			Validators: []validator.Object{
				objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("no_esql")),
			},
		},
	}
}

func getDatatableNoESQLSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"title": schema.StringAttribute{
			MarkdownDescription: "The title of the chart displayed in the panel.",
			Optional:            true,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "The description of the chart.",
			Optional:            true,
		},
		"dataset": schema.StringAttribute{
			MarkdownDescription: "Dataset configuration as JSON. For standard datatables, this specifies the data view and query.",
			CustomType:          jsontypes.NormalizedType{},
			Required:            true,
		},
		"density": schema.SingleNestedAttribute{
			MarkdownDescription: "Density configuration for the datatable.",
			Required:            true,
			Attributes:          getDatatableDensitySchema(),
		},
		"ignore_global_filters": schema.BoolAttribute{
			MarkdownDescription: "If true, ignore global filters when fetching data for this datatable.",
			Optional:            true,
		},
		"sampling": schema.Float64Attribute{
			MarkdownDescription: "Sampling factor between 0 (no sampling) and 1 (full sampling). Default is 1.",
			Optional:            true,
		},
		"query": schema.SingleNestedAttribute{
			MarkdownDescription: "Query configuration for filtering data.",
			Required:            true,
			Attributes:          getFilterSimpleSchema(),
		},
		"filters": schema.ListNestedAttribute{
			MarkdownDescription: "Additional filters to apply to the datatable data (maximum 100).",
			Optional:            true,
			NestedObject:        getSearchFilterSchema(),
		},
		"metrics": schema.ListNestedAttribute{
			MarkdownDescription: "Array of metric configurations as JSON. Each entry defines a datatable metric column.",
			Required:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"config": schema.StringAttribute{
						MarkdownDescription: "Metric configuration as JSON.",
						CustomType:          jsontypes.NormalizedType{},
						Required:            true,
					},
				},
			},
		},
		"rows": schema.ListNestedAttribute{
			MarkdownDescription: "Array of row configurations as JSON. Each entry defines a row split operation.",
			Optional:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"config": schema.StringAttribute{
						MarkdownDescription: "Row configuration as JSON.",
						CustomType:          jsontypes.NormalizedType{},
						Required:            true,
					},
				},
			},
		},
		"split_metrics_by": schema.ListNestedAttribute{
			MarkdownDescription: "Array of split-metrics configurations as JSON. Each entry defines a split operation for metric columns.",
			Optional:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"config": schema.StringAttribute{
						MarkdownDescription: "Split metrics configuration as JSON.",
						CustomType:          jsontypes.NormalizedType{},
						Required:            true,
					},
				},
			},
		},
		"sort_by": schema.StringAttribute{
			MarkdownDescription: "Sort configuration as JSON. Only one column can be sorted at a time.",
			CustomType:          jsontypes.NormalizedType{},
			Optional:            true,
		},
		"paging": schema.Int64Attribute{
			MarkdownDescription: "Enables pagination and sets the number of rows to display per page.",
			Optional:            true,
		},
	}
}

func getDatatableESQLSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"title": schema.StringAttribute{
			MarkdownDescription: "The title of the chart displayed in the panel.",
			Optional:            true,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "The description of the chart.",
			Optional:            true,
		},
		"dataset": schema.StringAttribute{
			MarkdownDescription: "Dataset configuration as JSON. For ES|QL, this specifies the ES|QL query.",
			CustomType:          jsontypes.NormalizedType{},
			Required:            true,
		},
		"density": schema.SingleNestedAttribute{
			MarkdownDescription: "Density configuration for the datatable.",
			Required:            true,
			Attributes:          getDatatableDensitySchema(),
		},
		"ignore_global_filters": schema.BoolAttribute{
			MarkdownDescription: "If true, ignore global filters when fetching data for this datatable.",
			Optional:            true,
		},
		"sampling": schema.Float64Attribute{
			MarkdownDescription: "Sampling factor between 0 (no sampling) and 1 (full sampling). Default is 1.",
			Optional:            true,
		},
		"filters": schema.ListNestedAttribute{
			MarkdownDescription: "Additional filters to apply to the datatable data (maximum 100).",
			Optional:            true,
			NestedObject:        getSearchFilterSchema(),
		},
		"metrics": schema.ListNestedAttribute{
			MarkdownDescription: "Array of metric configurations as JSON. Each entry defines a datatable metric column.",
			Required:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"config": schema.StringAttribute{
						MarkdownDescription: "Metric configuration as JSON.",
						CustomType:          jsontypes.NormalizedType{},
						Required:            true,
					},
				},
			},
		},
		"rows": schema.ListNestedAttribute{
			MarkdownDescription: "Array of row configurations as JSON. Each entry defines a row split operation.",
			Optional:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"config": schema.StringAttribute{
						MarkdownDescription: "Row configuration as JSON.",
						CustomType:          jsontypes.NormalizedType{},
						Required:            true,
					},
				},
			},
		},
		"split_metrics_by": schema.ListNestedAttribute{
			MarkdownDescription: "Array of split-metrics configurations as JSON. Each entry defines a split operation for metric columns.",
			Optional:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"config": schema.StringAttribute{
						MarkdownDescription: "Split metrics configuration as JSON.",
						CustomType:          jsontypes.NormalizedType{},
						Required:            true,
					},
				},
			},
		},
		"sort_by": schema.StringAttribute{
			MarkdownDescription: "Sort configuration as JSON. Only one column can be sorted at a time.",
			CustomType:          jsontypes.NormalizedType{},
			Optional:            true,
		},
		"paging": schema.Int64Attribute{
			MarkdownDescription: "Enables pagination and sets the number of rows to display per page.",
			Optional:            true,
		},
	}
}

func getDatatableDensitySchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"mode": schema.StringAttribute{
			MarkdownDescription: "Density mode. Valid values: 'compact', 'default', 'expanded'.",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("compact", "default", "expanded"),
			},
		},
		"height": schema.SingleNestedAttribute{
			MarkdownDescription: "Header and value height configuration.",
			Optional:            true,
			Attributes: map[string]schema.Attribute{
				"header": schema.SingleNestedAttribute{
					MarkdownDescription: "Header height configuration.",
					Optional:            true,
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							MarkdownDescription: "Header height type. Valid values: 'auto', 'custom'.",
							Optional:            true,
							Validators: []validator.String{
								stringvalidator.OneOf("auto", "custom"),
							},
						},
						"max_lines": schema.Float64Attribute{
							MarkdownDescription: "Maximum number of lines to use before header is truncated (for custom header height).",
							Optional:            true,
						},
					},
				},
				"value": schema.SingleNestedAttribute{
					MarkdownDescription: "Value height configuration.",
					Optional:            true,
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							MarkdownDescription: "Value height type. Valid values: 'auto', 'custom'.",
							Optional:            true,
							Validators: []validator.String{
								stringvalidator.OneOf("auto", "custom"),
							},
						},
						"lines": schema.Float64Attribute{
							MarkdownDescription: "Number of lines to display per table body cell (for custom value height).",
							Optional:            true,
						},
					},
				},
			},
		},
	}
}
