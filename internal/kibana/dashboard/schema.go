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
	"context"
	"maps"
	"regexp"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/validators"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/float32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	dashboardValueAuto          = "auto"
	dashboardValueAverage       = "average"
	pieChartTypeNumber          = "number"
	pieChartTypePercent         = "percent"
	operationTerms              = "terms"
	panelTypeMarkdown           = "markdown"
	panelTypeLens               = "lens"
	panelTypeTimeSlider         = "time_slider_control"
	panelTypeSloBurnRate        = "slo_burn_rate"
	panelTypeSloErrorBudget     = "slo_error_budget"
	panelTypeEsqlControl        = "esql_control"
	panelTypeOptionsListControl = "options_list_control"
	panelTypeRangeSlider        = "range_slider_control"
)

var sloBurnRateDurationRegex = regexp.MustCompile(`^\d+[mhd]$`)

var panelConfigNames = []string{
	"markdown_config",
	"config_json",
	"xy_chart_config",
	"treemap_config",
	"mosaic_config",
	"tagcloud_config",
	"region_map_config",
	"legacy_metric_config",
	"gauge_config",
	"metric_chart_config",
	"pie_chart_config",
	"datatable_config",
	"heatmap_config",
	"waffle_config",
	"time_slider_control_config",
	"slo_burn_rate_config",
	"slo_overview_config",
	"slo_error_budget_config",
	"esql_control_config",
	"options_list_control_config",
	"range_slider_control_config",
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

func isFieldMetricOperation(operation string) bool {
	switch operation {
	case "count", "unique_count", "min", "max", dashboardValueAverage, "median", "standard_deviation", "sum", "last_value", "percentile", "percentile_rank":
		return true
	default:
		return false
	}
}

// populateTagcloudMetricDefaults populates default values for tagcloud metric configuration
func populateTagcloudMetricDefaults(model map[string]any) map[string]any {
	if model == nil {
		return model
	}
	// Set defaults for all field metric operations
	if operation, ok := model["operation"].(string); ok && isFieldMetricOperation(operation) {
		if _, exists := model["empty_as_null"]; !exists {
			model["empty_as_null"] = false
		}
		if _, exists := model["show_metric_label"]; !exists {
			model["show_metric_label"] = true
		}
	}
	return model
}

// populateLensMetricDefaults populates default values for Lens metric configuration (shared across XY, metric, pie, treemap, datatable, etc.).
func populateLensMetricDefaults(model map[string]any) map[string]any {
	if model == nil {
		return model
	}

	// Set defaults for format
	if format, ok := model["format"].(map[string]any); ok {
		// Kibana has used both `type` and `id` as discriminators for number/percent format across
		// different visualizations/versions. Support both, and support both top-level params as well
		// as nested `params`.
		formatType, _ := format["type"].(string)
		formatID, _ := format["id"].(string)
		isNumberish := formatType == pieChartTypeNumber || formatType == pieChartTypePercent || formatID == pieChartTypeNumber || formatID == pieChartTypePercent

		if isNumberish {
			// If a nested params map exists, prefer setting defaults there.
			if params, ok := format["params"].(map[string]any); ok {
				if _, exists := params["compact"]; !exists {
					params["compact"] = false
				}
				if _, exists := params["decimals"]; !exists {
					params["decimals"] = float64(2)
				}
				format["params"] = params
			} else {
				if _, exists := format["compact"]; !exists {
					format["compact"] = false
				}
				if _, exists := format["decimals"]; !exists {
					format["decimals"] = float64(2)
				}
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

	metricType, _ := model["type"].(string)

	// Primary metrics have value/labels alignment defaults.
	if metricType == "primary" {
		if _, exists := model["value"]; !exists {
			model["value"] = map[string]any{"alignment": "right"}
		} else if v, ok := model["value"].(map[string]any); ok {
			if _, exists := v["alignment"]; !exists {
				v["alignment"] = "right"
			}
		}
		if _, exists := model["labels"]; !exists {
			model["labels"] = map[string]any{"alignment": "left"}
		} else if l, ok := model["labels"].(map[string]any); ok {
			if _, exists := l["alignment"]; !exists {
				l["alignment"] = "left"
			}
		}
	}

	// Secondary metrics have placement and value alignment defaults.
	if metricType == "secondary" {
		if _, exists := model["placement"]; !exists {
			model["placement"] = "before"
		}
		if _, exists := model["value"]; !exists {
			model["value"] = map[string]any{"alignment": "right"}
		} else if v, ok := model["value"].(map[string]any); ok {
			if _, exists := v["alignment"]; !exists {
				v["alignment"] = "right"
			}
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
	if operation, ok := model["operation"].(string); ok && operation == operationTerms {
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

// populatePartitionGroupByDefaults populates default values for partition chart group_by/group_breakdown_by configurations.
// Used by treemap and mosaic. Kibana may add default fields (e.g. rank_by, size) on read, so we normalize both sides.
func populatePartitionGroupByDefaults(model []map[string]any) []map[string]any {
	if model == nil {
		return model
	}

	for _, item := range model {
		if item == nil {
			continue
		}
		operation, _ := item["operation"].(string)
		if operation == "value" {
			continue
		}
		if operation != operationTerms {
			continue
		}
		// termsOperation requires collapse_by and format per API schema.
		if _, exists := item["collapse_by"]; !exists {
			item["collapse_by"] = "avg"
		}
		if _, exists := item["format"]; !exists {
			item["format"] = map[string]any{
				"type":     "number",
				"decimals": float64(2),
			}
		}
		if _, exists := item["rank_by"]; !exists {
			item["rank_by"] = map[string]any{
				"type":      "column",
				"metric":    float64(0),
				"direction": "desc",
			}
		}
		// Treemap defaults to a size of 5 for terms.
		if _, exists := item["size"]; !exists {
			item["size"] = float64(5)
		}
	}

	return model
}

// populatePartitionMetricsDefaults populates default values for partition chart metrics.
// Used by treemap and mosaic. Mirrors the defaulting behavior used by other Lens metric operations.
func populatePartitionMetricsDefaults(model []map[string]any) []map[string]any {
	if model == nil {
		return model
	}

	for i := range model {
		model[i] = populateTagcloudMetricDefaults(model[i])

		// ES|QL treemap metrics may omit format on write, but Kibana may return it as null.
		// Normalize both sides so semantic equality doesn't drift.
		if model[i] == nil {
			continue
		}
		if operation, ok := model[i]["operation"].(string); ok && operation == "value" {
			if _, exists := model[i]["format"]; !exists {
				model[i]["format"] = nil
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
	if operation, ok := model["operation"].(string); ok && isFieldMetricOperation(operation) {
		if _, exists := model["show_array_values"]; !exists {
			model["show_array_values"] = false
		}
		if _, exists := model["empty_as_null"]; !exists {
			model["empty_as_null"] = false
		}
	}

	format, ok := model["format"].(map[string]any)
	if ok {
		if formatType, ok := format["type"].(string); ok {
			switch formatType {
			case pieChartTypeNumber, pieChartTypePercent:
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
	if _, exists := model["title"]; !exists {
		model["title"] = map[string]any{"visible": true}
	}
	if _, exists := model["ticks"]; !exists {
		model["ticks"] = map[string]any{"visible": true, "mode": dashboardValueAuto}
	}

	return model
}

// populateRegionMapMetricDefaults populates default values for region map metric configuration
func populateRegionMapMetricDefaults(model map[string]any) map[string]any {
	if model == nil {
		return model
	}
	if operation, ok := model["operation"].(string); ok && isFieldMetricOperation(operation) {
		if _, exists := model["empty_as_null"]; !exists {
			model["empty_as_null"] = false
		}
		if _, exists := model["show_metric_label"]; !exists {
			model["show_metric_label"] = true
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
				MarkdownDescription: "The Kibana-assigned identifier for the dashboard.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
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
			"time_range": schema.SingleNestedAttribute{
				MarkdownDescription: "Dashboard time selection (`from`, `to`, optional `mode`). Aligns with the Kibana Dashboard API `time_range` object.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"from": schema.StringAttribute{
						MarkdownDescription: "Start of the time range (e.g., 'now-15m', '2023-01-01T00:00:00Z').",
						Required:            true,
					},
					"to": schema.StringAttribute{
						MarkdownDescription: "End of the time range (e.g., 'now', '2023-12-31T23:59:59Z').",
						Required:            true,
					},
					"mode": schema.StringAttribute{
						MarkdownDescription: "Time range mode. Valid values are `absolute` or `relative`. When the GET API omits `mode`, the provider preserves the prior `time_range.mode` from configuration or state.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("absolute", "relative"),
						},
					},
				},
			},
			"refresh_interval": schema.SingleNestedAttribute{
				MarkdownDescription: "Auto-refresh settings for the dashboard. Aligns with the Kibana Dashboard API `refresh_interval` object.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"pause": schema.BoolAttribute{
						MarkdownDescription: "When true, auto-refresh is paused.",
						Required:            true,
					},
					"value": schema.Int64Attribute{
						MarkdownDescription: "Refresh interval in milliseconds when not paused.",
						Required:            true,
					},
				},
			},
			"query": schema.SingleNestedAttribute{
				MarkdownDescription: "Dashboard-level query. Aligns with the Kibana Dashboard API `query` object: `language` plus exactly one of `text` (string branch) or `json` (object branch).",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"language": schema.StringAttribute{
						MarkdownDescription: "Query language (e.g., `kql`, `lucene`, `kuery`).",
						Required:            true,
					},
					"text": schema.StringAttribute{
						MarkdownDescription: "Query string for KQL or Lucene. Exactly one of `text` or `json` must be set.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("json")),
						},
					},
					"json": schema.StringAttribute{
						MarkdownDescription: "Query as normalized JSON for the object branch of the API union. Exactly one of `text` or `json` must be set.",
						CustomType:          jsontypes.NormalizedType{},
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("text")),
						},
					},
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
					"auto_apply_filters": schema.BoolAttribute{
						MarkdownDescription: "When true, control filters are applied automatically.",
						Optional:            true,
					},
					"hide_panel_borders": schema.BoolAttribute{
						MarkdownDescription: "When true, panel borders are hidden in the dashboard layout.",
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
						"uid": schema.StringAttribute{
							MarkdownDescription: "The unique identifier of the section (API `uid`).",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseNonNullStateForUnknown(),
							},
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
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
						Validators: []validator.String{
							stringvalidator.OneOf("write_restricted", "default"),
						},
					},
				},
			},
		},
	}
}

func getPanelSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Validators: []validator.Object{
			panelConfigValidator{},
		},
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of the panel (e.g. 'markdown', 'lens').",
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
			"uid": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the panel (API `uid`).",
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
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelTypeMarkdown}),
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
						Attributes:          getFilterSimple(),
					},
					"filters": schema.ListNestedAttribute{
						MarkdownDescription: "Additional filters to apply to the chart data (maximum 100).",
						Optional:            true,
						NestedObject:        getChartFilter(),
					},
				},
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(
						siblingPanelConfigPathsExcept("xy_chart_config", panelConfigNames)...,
					),
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelTypeLens}),
				},
			},
			"treemap_config": schema.SingleNestedAttribute{
				MarkdownDescription: panelConfigDescription("Configuration for a treemap chart panel.", "treemap_config", panelConfigNames),
				Optional:            true,
				Attributes:          getTreemapSchema(),
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(
						siblingPanelConfigPathsExcept("treemap_config", panelConfigNames)...,
					),
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelTypeLens}),
				},
			},
			"mosaic_config": schema.SingleNestedAttribute{
				MarkdownDescription: panelConfigDescription(
					"Configuration for a mosaic chart panel. Mosaic charts require two slicing dimensions "+
						"(group_by and group_breakdown_by).",
					"mosaic_config", panelConfigNames),
				Optional:   true,
				Attributes: getMosaicSchema(),
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(
						siblingPanelConfigPathsExcept("mosaic_config", panelConfigNames)...,
					),
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelTypeLens}),
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
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelTypeLens}),
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
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelTypeLens}),
				},
			},
			"heatmap_config": schema.SingleNestedAttribute{
				MarkdownDescription: panelConfigDescription("Configuration for a heatmap chart panel.", "heatmap_config", panelConfigNames),
				Optional:            true,
				Attributes:          getHeatmapSchema(),
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(
						siblingPanelConfigPathsExcept("heatmap_config", panelConfigNames)...,
					),
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelTypeLens}),
				},
			},
			"waffle_config": schema.SingleNestedAttribute{
				MarkdownDescription: panelConfigDescription(
					"Configuration for a waffle (grid) chart Lens panel. Omit `query` (or leave `query.expression` and `query.language` unset) for ES|QL mode.",
					"waffle_config",
					panelConfigNames,
				),
				Optional:   true,
				Attributes: getWaffleSchema(),
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(
						siblingPanelConfigPathsExcept("waffle_config", panelConfigNames)...,
					),
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelTypeLens}),
					waffleConfigModeValidator{},
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
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelTypeLens}),
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
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelTypeLens}),
				},
			},
			"metric_chart_config": schema.SingleNestedAttribute{
				MarkdownDescription: panelConfigDescription("Configuration for a metric chart panel. Metric charts display key performance indicators.", "metric_chart_config", panelConfigNames),
				Optional:            true,
				Attributes:          getMetricChart(),
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(
						siblingPanelConfigPathsExcept("metric_chart_config", panelConfigNames)...,
					),
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelTypeLens}),
				},
			},
			"pie_chart_config": schema.SingleNestedAttribute{
				MarkdownDescription: panelConfigDescription("Configuration for a pie chart panel. Use this for pie and donut charts.", "pie_chart_config", panelConfigNames),
				Optional:            true,
				Attributes:          getPieChart(),
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(
						siblingPanelConfigPathsExcept("pie_chart_config", panelConfigNames)...,
					),
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelTypeLens}),
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
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelTypeLens}),
				},
			},
			"time_slider_control_config": schema.SingleNestedAttribute{
				MarkdownDescription: panelConfigDescription(
					"Configuration for a time slider control panel. Controls the visible time window within the dashboard's global time range.",
					"time_slider_control_config",
					panelConfigNames,
				),
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"start_percentage_of_time_range": schema.Float32Attribute{
						MarkdownDescription: "Start of the visible time window as a fraction of the dashboard global range (0.0–1.0). " +
							"Float32 in state matches the Kibana API and avoids refresh drift.",
						Optional: true,
						Validators: []validator.Float32{
							float32validator.Between(0.0, 1.0),
						},
					},
					"end_percentage_of_time_range": schema.Float32Attribute{
						MarkdownDescription: "End of the visible time window as a fraction of the dashboard global range (0.0–1.0). " +
							"Float32 in state matches the Kibana API and avoids refresh drift.",
						Optional: true,
						Validators: []validator.Float32{
							float32validator.Between(0.0, 1.0),
						},
					},
					"is_anchored": schema.BoolAttribute{
						MarkdownDescription: "Whether the start of the time window is anchored (fixed), so only the end slides.",
						Optional:            true,
					},
				},
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(
						siblingPanelConfigPathsExcept("time_slider_control_config", panelConfigNames)...,
					),
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelTypeTimeSlider}),
				},
			},
			"slo_burn_rate_config": schema.SingleNestedAttribute{
				MarkdownDescription: panelConfigDescription(
					"Configuration for an SLO burn rate panel. Use this for panels that visualize the burn rate of an SLO over a configurable look-back window.",
					"slo_burn_rate_config",
					panelConfigNames,
				),
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"slo_id": schema.StringAttribute{
						MarkdownDescription: "The ID of the SLO to display the burn rate for.",
						Required:            true,
					},
					"duration": schema.StringAttribute{
						MarkdownDescription: "Duration for the burn rate chart in the format `[value][unit]`, where unit is `m` (minutes), `h` (hours), or `d` (days). For example: `5m`, `3h`, `6d`.",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.RegexMatches(
								sloBurnRateDurationRegex,
								"must match the pattern `^\\d+[mhd]$` (a positive integer followed by m, h, or d)",
							),
						},
					},
					"slo_instance_id": schema.StringAttribute{
						MarkdownDescription: "ID of the SLO instance. Set when the SLO uses `group_by`; identifies which instance to show. Omit to show all instances (API default `\"*\"`).",
						Optional:            true,
					},
					"title": schema.StringAttribute{
						MarkdownDescription: "Optional panel title.",
						Optional:            true,
					},
					"description": schema.StringAttribute{
						MarkdownDescription: "Optional panel description.",
						Optional:            true,
					},
					"hide_title": schema.BoolAttribute{
						MarkdownDescription: "When true, hides the panel title.",
						Optional:            true,
					},
					"hide_border": schema.BoolAttribute{
						MarkdownDescription: "When true, hides the panel border.",
						Optional:            true,
					},
					"drilldowns": schema.ListNestedAttribute{
						MarkdownDescription: "Optional list of URL drilldowns attached to the panel.",
						Optional:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"url": schema.StringAttribute{
									MarkdownDescription: "Templated URL for the drilldown.",
									Required:            true,
								},
								"label": schema.StringAttribute{
									MarkdownDescription: "Display label shown in the drilldown menu.",
									Required:            true,
								},
								"encode_url": schema.BoolAttribute{
									MarkdownDescription: "When true, the URL is percent-encoded. Omit to use the API default.",
									Optional:            true,
								},
								"open_in_new_tab": schema.BoolAttribute{
									MarkdownDescription: "When true, the URL opens in a new browser tab. Omit to use the API default.",
									Optional:            true,
								},
							},
						},
					},
				},
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(
						siblingPanelConfigPathsExcept("slo_burn_rate_config", panelConfigNames)...,
					),
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelTypeSloBurnRate}),
				},
			},
			"slo_overview_config": schema.SingleNestedAttribute{
				MarkdownDescription: panelConfigDescription(
					"Configuration for an SLO overview panel. Use either `single` (for a single SLO) or `groups` (for grouped SLO overview).",
					"slo_overview_config",
					panelConfigNames,
				),
				Optional:   true,
				Attributes: getSloOverviewSchema(),
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(
						siblingPanelConfigPathsExcept("slo_overview_config", panelConfigNames)...,
					),
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelTypeSloOverview}),
					sloOverviewConfigModeValidator{},
				},
			},
			"slo_error_budget_config": schema.SingleNestedAttribute{
				MarkdownDescription: panelConfigDescription(
					"Configuration for an SLO error budget panel. Displays the burn chart of remaining error budget for a specific SLO.",
					"slo_error_budget_config",
					panelConfigNames,
				),
				Optional:   true,
				Attributes: getSloErrorBudgetSchema(),
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(
						siblingPanelConfigPathsExcept("slo_error_budget_config", panelConfigNames)...,
					),
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelTypeSloErrorBudget}),
				},
			},
			"esql_control_config": schema.SingleNestedAttribute{
				MarkdownDescription: panelConfigDescription(
					"Configuration for an ES|QL control panel. Use this to manage ES|QL variable controls on a dashboard.",
					"esql_control_config",
					panelConfigNames,
				),
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"selected_options": schema.ListAttribute{
						MarkdownDescription: "List of currently selected option values for the control.",
						Required:            true,
						ElementType:         types.StringType,
					},
					"variable_name": schema.StringAttribute{
						MarkdownDescription: "The ES|QL variable name that this control binds to.",
						Required:            true,
					},
					"variable_type": schema.StringAttribute{
						MarkdownDescription: "The type of ES|QL variable. Allowed values: `fields`, `values`, `functions`, `time_literal`, `multi_values`.",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("fields", "values", "functions", "time_literal", "multi_values"),
						},
					},
					"esql_query": schema.StringAttribute{
						MarkdownDescription: "The ES|QL query used to populate the control's options.",
						Required:            true,
					},
					"control_type": schema.StringAttribute{
						MarkdownDescription: "The control type. Allowed values: `STATIC_VALUES`, `VALUES_FROM_QUERY`.",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("STATIC_VALUES", "VALUES_FROM_QUERY"),
						},
					},
					"title": schema.StringAttribute{
						MarkdownDescription: "A human-readable title displayed above the control widget.",
						Optional:            true,
					},
					"single_select": schema.BoolAttribute{
						MarkdownDescription: "When true, restricts the control to single-value selection.",
						Optional:            true,
					},
					"available_options": schema.ListAttribute{
						MarkdownDescription: "Pre-populated list of available options shown before the query executes.",
						Optional:            true,
						ElementType:         types.StringType,
					},
					"display_settings": schema.SingleNestedAttribute{
						MarkdownDescription: "Display configuration for the control widget.",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"placeholder": schema.StringAttribute{
								MarkdownDescription: "Placeholder text shown when no option is selected.",
								Optional:            true,
							},
							"hide_action_bar": schema.BoolAttribute{
								MarkdownDescription: "Whether to hide the action bar on the control.",
								Optional:            true,
							},
							"hide_exclude": schema.BoolAttribute{
								MarkdownDescription: "Whether to hide the exclude option.",
								Optional:            true,
							},
							"hide_exists": schema.BoolAttribute{
								MarkdownDescription: "Whether to hide the exists filter option.",
								Optional:            true,
							},
							"hide_sort": schema.BoolAttribute{
								MarkdownDescription: "Whether to hide the sort option.",
								Optional:            true,
							},
						},
					},
				},
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(
						siblingPanelConfigPathsExcept("esql_control_config", panelConfigNames)...,
					),
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelTypeEsqlControl}),
					validators.RequiredIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelTypeEsqlControl}),
				},
			},
			"options_list_control_config": schema.SingleNestedAttribute{
				MarkdownDescription: panelConfigDescription(
					"Configuration for an options list control panel. Provides a dropdown or multi-select filter based on a field in a data view.",
					"options_list_control_config",
					panelConfigNames,
				),
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"data_view_id": schema.StringAttribute{
						MarkdownDescription: "The ID of the data view that the control is tied to.",
						Required:            true,
					},
					"field_name": schema.StringAttribute{
						MarkdownDescription: "The name of the field in the data view that the control is tied to.",
						Required:            true,
					},
					"title": schema.StringAttribute{
						MarkdownDescription: "Human-readable label displayed above the control.",
						Optional:            true,
					},
					"use_global_filters": schema.BoolAttribute{
						MarkdownDescription: "Whether the control applies the dashboard's global filters to its own query.",
						Optional:            true,
					},
					"ignore_validations": schema.BoolAttribute{
						MarkdownDescription: "Whether the control skips field-level validation against the data view.",
						Optional:            true,
					},
					"single_select": schema.BoolAttribute{
						MarkdownDescription: "When true, only one option may be selected at a time.",
						Optional:            true,
					},
					"exclude": schema.BoolAttribute{
						MarkdownDescription: "When true, selected options are used as an exclusion filter rather than an inclusion filter.",
						Optional:            true,
					},
					"exists_selected": schema.BoolAttribute{
						MarkdownDescription: "When true, the control filters for documents where the field exists.",
						Optional:            true,
					},
					"run_past_timeout": schema.BoolAttribute{
						MarkdownDescription: "When true, the control continues to show results even when the underlying query times out.",
						Optional:            true,
					},
					"search_technique": schema.StringAttribute{
						MarkdownDescription: "The technique used to match suggestions. Must be one of `prefix`, `wildcard`, or `exact` when set.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("prefix", "wildcard", "exact"),
						},
					},
					"selected_options": schema.ListAttribute{
						MarkdownDescription: "The initially or persistently selected option values. All values are represented as strings.",
						Optional:            true,
						ElementType:         types.StringType,
					},
					"display_settings": schema.SingleNestedAttribute{
						MarkdownDescription: "Display preferences for the control widget.",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"placeholder": schema.StringAttribute{
								MarkdownDescription: "Placeholder text shown when no option is selected.",
								Optional:            true,
							},
							"hide_action_bar": schema.BoolAttribute{
								MarkdownDescription: "When true, hides the action bar on the control.",
								Optional:            true,
							},
							"hide_exclude": schema.BoolAttribute{
								MarkdownDescription: "When true, hides the exclude toggle.",
								Optional:            true,
							},
							"hide_exists": schema.BoolAttribute{
								MarkdownDescription: "When true, hides the exists filter option.",
								Optional:            true,
							},
							"hide_sort": schema.BoolAttribute{
								MarkdownDescription: "When true, hides the sort control.",
								Optional:            true,
							},
						},
					},
					"sort": schema.SingleNestedAttribute{
						MarkdownDescription: "Default sort configuration for the suggestion list.",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"by": schema.StringAttribute{
								MarkdownDescription: "The field or criterion to sort by. Must be one of `_count` or `_key`.",
								Required:            true,
								Validators: []validator.String{
									stringvalidator.OneOf("_count", "_key"),
								},
							},
							"direction": schema.StringAttribute{
								MarkdownDescription: "The sort direction. Must be one of `asc` or `desc`.",
								Required:            true,
								Validators: []validator.String{
									stringvalidator.OneOf("asc", "desc"),
								},
							},
						},
					},
				},
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(
						siblingPanelConfigPathsExcept("options_list_control_config", panelConfigNames)...,
					),
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelTypeOptionsListControl}),
					validators.RequiredIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelTypeOptionsListControl}),
				},
			},
			"range_slider_control_config": schema.SingleNestedAttribute{
				MarkdownDescription: panelConfigDescription(
					"Configuration for a range slider control panel. Provides a min/max range filter tied to a data view field.",
					"range_slider_control_config",
					panelConfigNames,
				),
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"title": schema.StringAttribute{
						MarkdownDescription: "A human-readable title for the control.",
						Optional:            true,
					},
					"data_view_id": schema.StringAttribute{
						MarkdownDescription: "The ID of the data view that the control is tied to.",
						Required:            true,
					},
					"field_name": schema.StringAttribute{
						MarkdownDescription: "The name of the field in the data view that the control is tied to.",
						Required:            true,
					},
					"use_global_filters": schema.BoolAttribute{
						MarkdownDescription: "Whether the control respects dashboard-level filters.",
						Optional:            true,
					},
					"ignore_validations": schema.BoolAttribute{
						MarkdownDescription: "Whether to suppress validation errors during intermediate states.",
						Optional:            true,
					},
					"value": schema.ListAttribute{
						MarkdownDescription: "Initial range as a list of exactly 2 strings: [min, max].",
						ElementType:         types.StringType,
						Optional:            true,
						Validators: []validator.List{
							listvalidator.SizeAtLeast(2),
							listvalidator.SizeAtMost(2),
						},
					},
					"step": schema.Float32Attribute{
						MarkdownDescription: "The step size for the range slider. Stored as float32 to match the Kibana API type and avoid refresh drift.",
						Optional:            true,
					},
				},
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(
						siblingPanelConfigPathsExcept("range_slider_control_config", panelConfigNames)...,
					),
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelTypeRangeSlider}),
					validators.RequiredIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelTypeRangeSlider}),
				},
			},
			"config_json": schema.StringAttribute{
				MarkdownDescription: panelConfigDescription("The configuration of the panel as a JSON string.", "config_json", panelConfigNames),
				CustomType:          customtypes.NewJSONWithDefaultsType(populatePanelConfigJSONDefaults),
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(
						siblingPanelConfigPathsExcept("config_json", panelConfigNames)...,
					),
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelTypeLens, panelTypeMarkdown}),
				},
			},
		},
	}
}

// getSloErrorBudgetSchema returns the schema for SLO error budget panel configuration.
func getSloErrorBudgetSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"slo_id": schema.StringAttribute{
			MarkdownDescription: "The ID of the SLO to display the error budget for.",
			Required:            true,
		},
		"slo_instance_id": schema.StringAttribute{
			MarkdownDescription: "ID of the SLO instance. Set when the SLO uses group_by; identifies which instance to show. Defaults to `*` (all instances) when omitted.",
			Optional:            true,
		},
		"title": schema.StringAttribute{
			MarkdownDescription: "The title displayed in the panel header.",
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
		"hide_border": schema.BoolAttribute{
			MarkdownDescription: "Hide the border of the panel.",
			Optional:            true,
		},
		"drilldowns": schema.ListNestedAttribute{
			MarkdownDescription: "URL drilldowns to configure on the panel.",
			Optional:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						MarkdownDescription: "Templated URL. Variables documented at https://www.elastic.co/docs/explore-analyze/dashboards/drilldowns#url-template-variable",
						Required:            true,
					},
					"label": schema.StringAttribute{
						MarkdownDescription: "The label displayed for the drilldown.",
						Required:            true,
					},
					"encode_url": schema.BoolAttribute{
						MarkdownDescription: "When true, the URL is escaped using percent encoding. Defaults to `true` when omitted.",
						Optional:            true,
					},
					"open_in_new_tab": schema.BoolAttribute{
						MarkdownDescription: "When true, the drilldown URL opens in a new browser tab. Defaults to `true` when omitted.",
						Optional:            true,
					},
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
				"scale": schema.StringAttribute{
					MarkdownDescription: "X-axis scale: linear (numeric), ordinal (categorical), or temporal (dates).",
					Optional:            true,
					Validators: []validator.String{
						stringvalidator.OneOf("linear", "ordinal", "temporal"),
					},
				},
				"extent_json": schema.StringAttribute{
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
		"extent_json": schema.StringAttribute{
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
		},
		"position": schema.StringAttribute{
			MarkdownDescription: "Legend position when positioned outside the chart. Valid when 'inside' is false or omitted.",
			Optional:            true,
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

// getFilterSimple returns the schema for simple filter configuration
func getFilterSimple() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"language": schema.StringAttribute{
			MarkdownDescription: "Query language (default: 'kql').",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("kql", "lucene"),
			},
		},
		"expression": schema.StringAttribute{
			MarkdownDescription: "Filter expression string.",
			Required:            true,
		},
	}
}

// getChartFilter returns the schema for a single chart-level filter (API-shaped JSON).
func getChartFilter() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"filter_json": schema.StringAttribute{
				MarkdownDescription: "Chart filter as normalized JSON. Must match the Kibana dashboard API for this chart: " +
					"one of the filter union members (condition, group, DSL, or spatial) described in the dashboards OpenAPI specification.",
				CustomType: jsontypes.NormalizedType{},
				Required:   true,
			},
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
		"dataset_json": schema.StringAttribute{
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
		"dataset_json": schema.StringAttribute{
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

// getTagcloudSchema returns the schema for tagcloud chart configuration
func getTagcloudSchema() map[string]schema.Attribute {
	attrs := lensChartBaseAttributes()
	attrs["dataset_json"] = schema.StringAttribute{
		MarkdownDescription: "Dataset configuration as JSON. For standard layers, this specifies the data view and query.",
		CustomType:          jsontypes.NormalizedType{},
		Required:            true,
	}
	attrs["query"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Query configuration for filtering data.",
		Required:            true,
		Attributes:          getFilterSimple(),
	}
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
	attrs["metric_json"] = schema.StringAttribute{
		MarkdownDescription: tagcloudMetricDescription,
		CustomType:          customtypes.NewJSONWithDefaultsType(populateTagcloudMetricDefaults),
		Required:            true,
	}
	attrs["tag_by_json"] = schema.StringAttribute{
		MarkdownDescription: "Tag grouping configuration as JSON. Can be a date histogram, terms, histogram, range, or filters operation. This determines how tags are grouped and displayed.",
		CustomType:          customtypes.NewJSONWithDefaultsType(populateTagcloudTagByDefaults),
		Required:            true,
	}
	return attrs
}

// getHeatmapSchema returns the schema for heatmap chart configuration
func getHeatmapSchema() map[string]schema.Attribute {
	attrs := lensChartBaseAttributes()
	attrs["dataset_json"] = schema.StringAttribute{
		MarkdownDescription: "Dataset configuration as JSON. For standard heatmaps, this specifies the data view or index; for ES|QL, this specifies the ES|QL query dataset.",
		CustomType:          jsontypes.NormalizedType{},
		Required:            true,
	}
	attrs["query"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Query configuration for filtering data. Required for non-ES|QL heatmaps.",
		Optional:            true,
		Attributes:          getFilterSimple(),
	}
	attrs["axes"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Axis configuration for X and Y axes.",
		Required:            true,
		Attributes:          getHeatmapAxesSchema(),
	}
	attrs["cells"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Cells configuration for the heatmap.",
		Required:            true,
		Attributes:          getHeatmapCellsSchema(),
	}
	attrs["legend"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Legend configuration for the heatmap.",
		Required:            true,
		Attributes:          getHeatmapLegendSchema(),
	}
	attrs["metric_json"] = schema.StringAttribute{
		MarkdownDescription: "Metric configuration as JSON. For non-ES|QL, this can be a field metric, pipeline metric, or formula. For ES|QL, this is the metric column/operation/color configuration.",
		CustomType:          customtypes.NewJSONWithDefaultsType(populateTagcloudMetricDefaults),
		Required:            true,
	}
	attrs["x_axis_json"] = schema.StringAttribute{
		MarkdownDescription: heatmapXAxisDescription,
		CustomType:          jsontypes.NormalizedType{},
		Required:            true,
	}
	attrs["y_axis_json"] = schema.StringAttribute{
		MarkdownDescription: heatmapYAxisDescription,
		CustomType:          jsontypes.NormalizedType{},
		Optional:            true,
	}
	return attrs
}

// lensChartBaseAttributes returns attributes shared by most Lens chart panels:
// title, description, sampling, ignore_global_filters, and filters.
func lensChartBaseAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"title": schema.StringAttribute{
			MarkdownDescription: "The title of the chart displayed in the panel.",
			Optional:            true,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "The description of the chart.",
			Optional:            true,
		},
		"sampling": schema.Float64Attribute{
			MarkdownDescription: "Sampling factor between 0 (no sampling) and 1 (full sampling). Default is 1.",
			Optional:            true,
		},
		"ignore_global_filters": schema.BoolAttribute{
			MarkdownDescription: "If true, ignore global filters when fetching data for this chart. Default is false.",
			Optional:            true,
		},
		"filters": schema.ListNestedAttribute{
			MarkdownDescription: "Additional filters to apply to the chart data (maximum 100).",
			Optional:            true,
			NestedObject:        getChartFilter(),
		},
	}
}

// getPartitionChartBaseSchema returns base attributes shared by partition charts (treemap, mosaic).
func getPartitionChartBaseSchema() map[string]schema.Attribute {
	attrs := lensChartBaseAttributes()
	attrs["dataset_json"] = schema.StringAttribute{
		MarkdownDescription: "Dataset configuration as JSON. For non-ES|QL, this specifies the data view or index; for ES|QL, this specifies the ES|QL query dataset.",
		CustomType:          jsontypes.NormalizedType{},
		Required:            true,
	}
	attrs["query"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Query configuration for filtering data. Required for non-ES|QL partition charts.",
		Optional:            true,
		Attributes:          getFilterSimple(),
	}
	return attrs
}

// getWaffleSchema returns schema for waffle (grid) Lens chart configuration.
func getWaffleSchema() map[string]schema.Attribute {
	attrs := getPartitionChartBaseSchema()
	attrs["legend"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Legend configuration for the waffle chart.",
		Required:            true,
		Attributes:          getWaffleLegendSchema(),
	}
	attrs["value_display"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Configuration for displaying values in chart cells.",
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"mode": schema.StringAttribute{
				MarkdownDescription: "Value display mode in cells: hidden, absolute, or percentage.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("hidden", "absolute", "percentage"),
				},
			},
			"percent_decimals": schema.Float64Attribute{
				MarkdownDescription: "Decimal places for percentage display (0-10).",
				Optional:            true,
			},
		},
	}
	attrs["metrics"] = schema.ListNestedAttribute{
		MarkdownDescription: "Metric configurations for non-ES|QL waffles (minimum 1). Each `config` is a JSON object (e.g. count, sum, or formula) matching the Kibana Lens waffle schema.",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"config": schema.StringAttribute{
					MarkdownDescription: "Metric operation as JSON.",
					CustomType:          customtypes.NewJSONWithDefaultsType(populatePieChartMetricDefaults),
					Required:            true,
				},
			},
		},
	}
	attrs["group_by"] = schema.ListNestedAttribute{
		MarkdownDescription: "Breakdown dimensions for non-ES|QL waffles. Each `config` is a JSON object (terms, date_histogram, etc.) matching the Kibana Lens waffle schema.",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"config": schema.StringAttribute{
					MarkdownDescription: "Group-by operation as JSON.",
					CustomType:          customtypes.NewJSONWithDefaultsType(populateLensGroupByDefaults),
					Required:            true,
				},
			},
		},
	}
	attrs["esql_metrics"] = schema.ListNestedAttribute{
		MarkdownDescription: "Metric columns for ES|QL waffles (minimum 1). Mutually exclusive with `metrics`.",
		Optional:            true,
		NestedObject:        getWaffleESQLMetricSchema(),
	}
	attrs["esql_group_by"] = schema.ListNestedAttribute{
		MarkdownDescription: "Breakdown columns for ES|QL waffles. Mutually exclusive with `group_by`.",
		Optional:            true,
		NestedObject:        getWaffleESQLGroupBySchema(),
	}
	return attrs
}

// getWaffleLegendSchema returns schema for waffle legend (distinct from XY/heatmap legend).
func getWaffleLegendSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"size": schema.StringAttribute{
			MarkdownDescription: "Legend size: auto, s, m, l, or xl.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(dashboardValueAuto, "s", "m", "l", "xl"),
			},
		},
		"truncate_after_lines": schema.Int64Attribute{
			MarkdownDescription: "Maximum lines before truncating legend items (1-10).",
			Optional:            true,
		},
		"values": schema.ListAttribute{
			MarkdownDescription: "Legend value display modes. For example `absolute` shows raw metric values in the legend.",
			ElementType:         types.StringType,
			Optional:            true,
			Validators: []validator.List{
				listvalidator.ValueStringsAre(stringvalidator.OneOf("absolute")),
			},
		},
		"visible": schema.StringAttribute{
			MarkdownDescription: "Legend visibility: auto, visible, or hidden.",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("auto", "visible", "hidden"),
			},
		},
	}
}

func getWaffleESQLMetricSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"column": schema.StringAttribute{
				MarkdownDescription: "ES|QL column name for the metric.",
				Required:            true,
			},
			"label": schema.StringAttribute{
				MarkdownDescription: "Optional label for the metric.",
				Optional:            true,
			},
			"format_json": schema.StringAttribute{
				MarkdownDescription: "Number or other format configuration as JSON (`formatType` union).",
				CustomType:          jsontypes.NormalizedType{},
				Required:            true,
			},
			"color": schema.SingleNestedAttribute{
				MarkdownDescription: "Static color for the metric.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "Color type; use `static` for waffle ES|QL metrics.",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("static"),
						},
					},
					"color": schema.StringAttribute{
						MarkdownDescription: "Color value (e.g. hex).",
						Required:            true,
					},
				},
			},
		},
	}
}

func getWaffleESQLGroupBySchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"column": schema.StringAttribute{
				MarkdownDescription: "ES|QL column for the breakdown.",
				Required:            true,
			},
			"collapse_by": schema.StringAttribute{
				MarkdownDescription: "Collapse function when multiple rows map to the same bucket.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("avg", "max", "min", "sum"),
				},
			},
			"color_json": schema.StringAttribute{
				MarkdownDescription: "Color mapping as JSON (`colorMapping` union).",
				CustomType:          jsontypes.NormalizedType{},
				Required:            true,
			},
			"format_json": schema.StringAttribute{
				MarkdownDescription: "Column format as JSON (e.g. `{\"type\":\"number\"}`). Defaults to numeric format when omitted.",
				CustomType:          jsontypes.NormalizedType{},
				Optional:            true,
			},
			"label": schema.StringAttribute{
				MarkdownDescription: "Optional label for the group-by column.",
				Optional:            true,
			},
		},
	}
}

// getTreemapSchema returns the schema for treemap chart configuration
func getTreemapSchema() map[string]schema.Attribute {
	base := getPartitionChartBaseSchema()
	treemapSpecific := map[string]schema.Attribute{
		"group_by_json": schema.StringAttribute{
			MarkdownDescription: "Array of breakdown dimensions as JSON (minimum 1). " +
				"For non-ES|QL, each item can be date histogram, terms, histogram, range, or filters operations; " +
				"for ES|QL, each item is the column/operation/color configuration.",
			CustomType: customtypes.NewJSONWithDefaultsType(populatePartitionGroupByDefaults),
			Required:   true,
		},
		"metrics_json": schema.StringAttribute{
			MarkdownDescription: "Array of metric configurations as JSON (minimum 1). " +
				"For non-ES|QL, each item can be a field metric, pipeline metric, or formula; " +
				"for ES|QL, each item is the column/operation/color/format configuration.",
			CustomType: customtypes.NewJSONWithDefaultsType(populatePartitionMetricsDefaults),
			Required:   true,
		},
		"legend": schema.SingleNestedAttribute{
			MarkdownDescription: "Legend configuration for the treemap chart.",
			Required:            true,
			Attributes:          getPartitionLegendSchema(),
		},
		"value_display": schema.SingleNestedAttribute{
			MarkdownDescription: "Configuration for displaying values in chart cells.",
			Optional:            true,
			Attributes:          getPartitionValueDisplaySchema(),
		},
	}
	maps.Copy(base, treemapSpecific)
	return base
}

// getMosaicSchema returns the schema for mosaic chart configuration
func getMosaicSchema() map[string]schema.Attribute {
	base := getPartitionChartBaseSchema()
	mosaicSpecific := map[string]schema.Attribute{
		"group_by_json": schema.StringAttribute{
			MarkdownDescription: "Array of primary breakdown dimensions as JSON (minimum 1). " +
				"For non-ES|QL, each item can be date histogram, terms, histogram, range, or filters operations; " +
				"for ES|QL, each item is the column/operation/color configuration.",
			CustomType: customtypes.NewJSONWithDefaultsType(populatePartitionGroupByDefaults),
			Required:   true,
		},
		"group_breakdown_by_json": schema.StringAttribute{
			MarkdownDescription: "Array of secondary breakdown dimensions as JSON (minimum 1). " +
				"Mosaic charts require both group_by and group_breakdown_by. " +
				"For non-ES|QL, each item can be date histogram, terms, histogram, range, or filters operations; " +
				"for ES|QL, each item is the column/operation/color configuration.",
			CustomType: customtypes.NewJSONWithDefaultsType(populatePartitionGroupByDefaults),
			Required:   true,
		},
		"metrics_json": schema.StringAttribute{
			MarkdownDescription: "Array of metric configurations as JSON (exactly 1 required). " +
				"For non-ES|QL, each item can be a field metric, pipeline metric, or formula; " +
				"for ES|QL, each item is the column/operation/color/format configuration.",
			CustomType: customtypes.NewJSONWithDefaultsType(populatePartitionMetricsDefaults),
			Required:   true,
		},
		"legend": schema.SingleNestedAttribute{
			MarkdownDescription: "Legend configuration for the mosaic chart.",
			Required:            true,
			Attributes:          getPartitionLegendSchema(),
		},
		"value_display": schema.SingleNestedAttribute{
			MarkdownDescription: "Configuration for displaying values in chart cells.",
			Optional:            true,
			Attributes:          getPartitionValueDisplaySchema(),
		},
	}
	maps.Copy(base, mosaicSpecific)
	return base
}

func getPartitionLegendSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"nested": schema.BoolAttribute{
			MarkdownDescription: "Show nested legend with hierarchical breakdown levels.",
			Optional:            true,
		},
		"size": schema.StringAttribute{
			MarkdownDescription: "Legend size: auto, s, m, l, or xl.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("auto", "s", "m", "l", "xl"),
			},
		},
		"truncate_after_lines": schema.Float64Attribute{
			MarkdownDescription: "Maximum lines before truncating legend items (1-10).",
			Optional:            true,
		},
		"visible": schema.StringAttribute{
			MarkdownDescription: "Legend visibility: auto, visible, or hidden.",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("auto", "visible", "hidden"),
			},
		},
	}
}

func getPartitionValueDisplaySchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"mode": schema.StringAttribute{
			MarkdownDescription: "Value display mode: hidden, absolute, or percentage.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("hidden", "absolute", "percentage"),
			},
		},
		"percent_decimals": schema.Float64Attribute{
			MarkdownDescription: "Decimal places for percentage display (0-10).",
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
				stringvalidator.OneOf(dashboardValueAuto, "s", "m", "l", "xl"),
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
	attrs := lensChartBaseAttributes()
	attrs["dataset_json"] = schema.StringAttribute{
		MarkdownDescription: "Dataset configuration as JSON. For ES|QL, this specifies the ES|QL query. For standard layers, this specifies the data view and query.",
		CustomType:          jsontypes.NormalizedType{},
		Required:            true,
	}
	attrs["query"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Query configuration for filtering data. Required for non-ES|QL region map configurations.",
		Optional:            true,
		Attributes:          getFilterSimple(),
	}
	attrs["metric_json"] = schema.StringAttribute{
		MarkdownDescription: "Metric configuration as JSON. For ES|QL, this defines the metric column and format. For standard mode, this defines the metric operation or formula.",
		CustomType:          customtypes.NewJSONWithDefaultsType(populateRegionMapMetricDefaults),
		Required:            true,
	}
	attrs["region_json"] = schema.StringAttribute{
		MarkdownDescription: regionMapRegionDescription,
		CustomType:          jsontypes.NormalizedType{},
		Required:            true,
	}
	return attrs
}

// getLegacyMetricSchema returns the schema for legacy metric chart configuration
func getLegacyMetricSchema() map[string]schema.Attribute {
	attrs := lensChartBaseAttributes()
	attrs["dataset_json"] = schema.StringAttribute{
		MarkdownDescription: "Dataset configuration as JSON. Use `dataView` or `index` for standard data sources, and `esql` or `table` for ES|QL sources.",
		CustomType:          jsontypes.NormalizedType{},
		Required:            true,
	}
	attrs["metric_json"] = schema.StringAttribute{
		MarkdownDescription: "Metric configuration as JSON. For standard datasets, use a metric operation or formula. For ES|QL datasets, include format, operation, column, and color configuration.",
		CustomType:          customtypes.NewJSONWithDefaultsType(populateLegacyMetricMetricDefaults),
		Required:            true,
	}
	attrs["query"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Query configuration for filtering data. Required for non-ES|QL datasets.",
		Optional:            true,
		Attributes:          getFilterSimple(),
	}
	return attrs
}

// getGaugeSchema returns the schema for gauge chart configuration
func getGaugeSchema() map[string]schema.Attribute {
	attrs := lensChartBaseAttributes()
	attrs["dataset_json"] = schema.StringAttribute{
		MarkdownDescription: "Dataset configuration as JSON. For standard layers, this specifies the data view and query.",
		CustomType:          jsontypes.NormalizedType{},
		Required:            true,
	}
	attrs["query"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Query configuration for filtering data.",
		Required:            true,
		Attributes:          getFilterSimple(),
	}
	attrs["metric_json"] = schema.StringAttribute{
		MarkdownDescription: gaugeMetricDescription,
		CustomType:          customtypes.NewJSONWithDefaultsType(populateGaugeMetricDefaults),
		Required:            true,
	}
	attrs["shape_json"] = schema.StringAttribute{
		MarkdownDescription: "Gauge shape configuration as JSON. Supports bullet and circular gauges.",
		CustomType:          jsontypes.NormalizedType{},
		Optional:            true,
	}
	return attrs
}

// getMetricChart returns the schema for metric chart configuration
func getMetricChart() map[string]schema.Attribute {
	attrs := lensChartBaseAttributes()
	attrs["dataset_json"] = schema.StringAttribute{
		MarkdownDescription: metricChartDatasetDescription,
		CustomType:          jsontypes.NormalizedType{},
		Required:            true,
	}
	attrs["query"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Query configuration for filtering data. Required for non-ES|QL datasets.",
		Optional:            true,
		Attributes:          getFilterSimple(),
	}
	attrs["metrics"] = schema.ListNestedAttribute{
		MarkdownDescription: metricChartMetricsDescription,
		Required:            true,
		Validators: []validator.List{
			listvalidator.SizeAtMost(2),
		},
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"config_json": schema.StringAttribute{
					MarkdownDescription: metricChartMetricConfigDescription,
					CustomType:          customtypes.NewJSONWithDefaultsType(populateLensMetricDefaults),
					Required:            true,
				},
			},
		},
	}
	attrs["breakdown_by_json"] = schema.StringAttribute{
		MarkdownDescription: metricChartBreakdownByDescription,
		CustomType:          jsontypes.NormalizedType{},
		Optional:            true,
	}
	return attrs
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
	attrs := lensChartBaseAttributes()
	attrs["dataset_json"] = schema.StringAttribute{
		MarkdownDescription: "Dataset configuration as JSON. For standard datatables, this specifies the data view and query.",
		CustomType:          jsontypes.NormalizedType{},
		Required:            true,
	}
	attrs["density"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Density configuration for the datatable.",
		Required:            true,
		Attributes:          getDatatableDensitySchema(),
	}
	attrs["query"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Query configuration for filtering data.",
		Required:            true,
		Attributes:          getFilterSimple(),
	}
	attrs["metrics"] = schema.ListNestedAttribute{
		MarkdownDescription: "Array of metric configurations as JSON. Each entry defines a datatable metric column.",
		Required:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"config_json": schema.StringAttribute{
					MarkdownDescription: "Metric configuration as JSON.",
					CustomType:          jsontypes.NormalizedType{},
					Required:            true,
				},
			},
		},
	}
	attrs["rows"] = schema.ListNestedAttribute{
		MarkdownDescription: "Array of row configurations as JSON. Each entry defines a row split operation.",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"config_json": schema.StringAttribute{
					MarkdownDescription: "Row configuration as JSON.",
					CustomType:          jsontypes.NormalizedType{},
					Required:            true,
				},
			},
		},
	}
	attrs["split_metrics_by"] = schema.ListNestedAttribute{
		MarkdownDescription: "Array of split-metrics configurations as JSON. Each entry defines a split operation for metric columns.",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"config_json": schema.StringAttribute{
					MarkdownDescription: "Split metrics configuration as JSON.",
					CustomType:          jsontypes.NormalizedType{},
					Required:            true,
				},
			},
		},
	}
	attrs["sort_by_json"] = schema.StringAttribute{
		MarkdownDescription: "Sort configuration as JSON. Only one column can be sorted at a time.",
		CustomType:          jsontypes.NormalizedType{},
		Optional:            true,
	}
	attrs["paging"] = schema.Int64Attribute{
		MarkdownDescription: "Enables pagination and sets the number of rows to display per page.",
		Optional:            true,
	}
	return attrs
}

func getDatatableESQLSchema() map[string]schema.Attribute {
	attrs := lensChartBaseAttributes()
	attrs["dataset_json"] = schema.StringAttribute{
		MarkdownDescription: "Dataset configuration as JSON. For ES|QL, this specifies the ES|QL query.",
		CustomType:          jsontypes.NormalizedType{},
		Required:            true,
	}
	attrs["density"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Density configuration for the datatable.",
		Required:            true,
		Attributes:          getDatatableDensitySchema(),
	}
	attrs["metrics"] = schema.ListNestedAttribute{
		MarkdownDescription: "Array of metric configurations as JSON. Each entry defines a datatable metric column.",
		Required:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"config_json": schema.StringAttribute{
					MarkdownDescription: "Metric configuration as JSON.",
					CustomType:          jsontypes.NormalizedType{},
					Required:            true,
				},
			},
		},
	}
	attrs["rows"] = schema.ListNestedAttribute{
		MarkdownDescription: "Array of row configurations as JSON. Each entry defines a row split operation.",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"config_json": schema.StringAttribute{
					MarkdownDescription: "Row configuration as JSON.",
					CustomType:          jsontypes.NormalizedType{},
					Required:            true,
				},
			},
		},
	}
	attrs["split_metrics_by"] = schema.ListNestedAttribute{
		MarkdownDescription: "Array of split-metrics configurations as JSON. Each entry defines a split operation for metric columns.",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"config_json": schema.StringAttribute{
					MarkdownDescription: "Split metrics configuration as JSON.",
					CustomType:          jsontypes.NormalizedType{},
					Required:            true,
				},
			},
		},
	}
	attrs["sort_by_json"] = schema.StringAttribute{
		MarkdownDescription: "Sort configuration as JSON. Only one column can be sorted at a time.",
		CustomType:          jsontypes.NormalizedType{},
		Optional:            true,
	}
	attrs["paging"] = schema.Int64Attribute{
		MarkdownDescription: "Enables pagination and sets the number of rows to display per page.",
		Optional:            true,
	}
	return attrs
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
								stringvalidator.OneOf(dashboardValueAuto, "custom"),
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
								stringvalidator.OneOf(dashboardValueAuto, "custom"),
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

// populatePieChartMetricDefaults populates default values for pie chart metric configuration
func populatePieChartMetricDefaults(model map[string]any) map[string]any {
	if model == nil {
		return model
	}

	if _, exists := model["empty_as_null"]; !exists {
		model["empty_as_null"] = false
	}

	// Set defaults for format
	if format, ok := model["format"].(map[string]any); ok {
		if format["type"] == pieChartTypeNumber {
			if _, exists := format["compact"]; !exists {
				format["compact"] = false
			}
			if _, exists := format["decimals"]; !exists {
				format["decimals"] = float64(2)
			}
		}
	}

	return model
}

// populateLensGroupByDefaults populates default values for Lens dimension/group-by configuration (shared across pie, treemap, datatable, etc.).
func populateLensGroupByDefaults(model map[string]any) map[string]any {
	if model == nil {
		return model
	}

	if operation, ok := model["operation"].(string); ok && operation == operationTerms {
		if _, exists := model["size"]; !exists {
			model["size"] = float64(5)
		}
		if _, exists := model["rank_by"]; !exists {
			model["rank_by"] = map[string]any{
				"direction": "desc",
				"metric":    float64(0),
				"type":      "column",
			}
		}
	}

	return model
}

// pieChartLegendDefaultObject is the schema default when the legend block is omitted from config,
// aligned with typical Kibana read-back so apply and refresh stay consistent.
func pieChartLegendDefaultObject() types.Object {
	return types.ObjectValueMust(
		map[string]attr.Type{
			"nested":               types.BoolType,
			"size":                 types.StringType,
			"truncate_after_lines": types.Float64Type,
			"visible":              types.StringType,
		},
		map[string]attr.Value{
			"nested":               types.BoolNull(),
			"size":                 types.StringValue("auto"),
			"truncate_after_lines": types.Float64Null(),
			"visible":              types.StringValue("auto"),
		},
	)
}

// getPieChart returns the schema for pie chart configuration
func getPieChart() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"title": schema.StringAttribute{
			MarkdownDescription: "The title of the chart displayed in the panel.",
			Optional:            true,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "The description of the chart.",
			Optional:            true,
		},
		"dataset_json": schema.StringAttribute{
			MarkdownDescription: "Dataset configuration as JSON. For standard layers, this specifies the data view and query.",
			CustomType:          jsontypes.NormalizedType{},
			Optional:            true,
		},
		"ignore_global_filters": schema.BoolAttribute{
			MarkdownDescription: "If true, ignore global filters when fetching data for this layer. Default is false.",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
		},
		"sampling": schema.Float64Attribute{
			MarkdownDescription: "Sampling factor between 0 (no sampling) and 1 (full sampling). Default is 1.",
			Optional:            true,
			Computed:            true,
			Default:             float64default.StaticFloat64(1.0),
		},
		"donut_hole": schema.StringAttribute{
			MarkdownDescription: "Donut hole size: none (pie), s, m, or l.",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("none", "s", "m", "l"),
			},
		},
		"label_position": schema.StringAttribute{
			MarkdownDescription: "Position of slice labels: hidden, inside, or outside.",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("hidden", "inside", "outside"),
			},
		},
		"legend": schema.SingleNestedAttribute{
			MarkdownDescription: "Optional legend configuration for the pie chart. " +
				"Same shape as treemap and mosaic legends; Terraform `visible` maps to API `visibility`. " +
				"When omitted, the schema default matches typical Kibana legend defaults (size and visibility " +
				"`auto`) so apply/read stay consistent.",
			Optional:   true,
			Computed:   true,
			Default:    objectdefault.StaticValue(pieChartLegendDefaultObject()),
			Attributes: getPartitionLegendSchema(),
		},
		"query": schema.SingleNestedAttribute{
			MarkdownDescription: "Query configuration for filtering data.",
			Optional:            true,
			Attributes:          getFilterSimple(),
		},
		"filters": schema.ListNestedAttribute{
			MarkdownDescription: "Additional filters to apply to the chart data (maximum 100).",
			Optional:            true,
			NestedObject:        getChartFilter(),
		},
		"metrics": schema.ListNestedAttribute{
			MarkdownDescription: "Array of metric configurations (minimum 1).",
			Required:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"config": schema.StringAttribute{
						MarkdownDescription: "Metric configuration as JSON.",
						CustomType:          customtypes.NewJSONWithDefaultsType(populatePieChartMetricDefaults),
						Required:            true,
					},
				},
			},
			Validators: []validator.List{
				listvalidator.SizeAtLeast(1),
			},
		},
		"group_by": schema.ListNestedAttribute{
			MarkdownDescription: "Array of breakdown dimensions (minimum 1).",
			Optional:            true,
			Validators: []validator.List{
				listvalidator.SizeAtLeast(1),
			},
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"config": schema.StringAttribute{
						MarkdownDescription: "Group by configuration as JSON.",
						CustomType:          customtypes.NewJSONWithDefaultsType(populateLensGroupByDefaults),
						Required:            true,
					},
				},
			},
		},
	}
}

// getSloOverviewSchema returns the schema for the slo_overview_config block.
func getSloOverviewSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"single": schema.SingleNestedAttribute{
			MarkdownDescription: "Configuration for a single-SLO overview panel. Mutually exclusive with `groups`.",
			Optional:            true,
			Attributes:          getSloSingleSchema(),
			Validators: []validator.Object{
				objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("groups")),
			},
		},
		"groups": schema.SingleNestedAttribute{
			MarkdownDescription: "Configuration for a grouped SLO overview panel. Mutually exclusive with `single`.",
			Optional:            true,
			Attributes:          getSloGroupsSchema(),
			Validators: []validator.Object{
				objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("single")),
			},
		},
	}
}

// getSloSharedDisplaySchema returns display attributes shared by both single and groups modes.
func getSloSharedDisplaySchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"title": schema.StringAttribute{
			MarkdownDescription: "The title displayed on the panel.",
			Optional:            true,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "The description displayed on the panel.",
			Optional:            true,
		},
		"hide_title": schema.BoolAttribute{
			MarkdownDescription: "When true, the panel title is hidden.",
			Optional:            true,
		},
		"hide_border": schema.BoolAttribute{
			MarkdownDescription: "When true, the panel border is hidden.",
			Optional:            true,
		},
		"drilldowns": schema.ListNestedAttribute{
			MarkdownDescription: "URL drilldowns attached to the panel. The trigger (`on_open_panel_menu`) and type (`url_drilldown`) are set automatically.",
			Optional:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						MarkdownDescription: "The URL template for the drilldown. Variables are documented at https://www.elastic.co/docs/explore-analyze/dashboards/drilldowns#url-template-variable.",
						Required:            true,
					},
					"label": schema.StringAttribute{
						MarkdownDescription: "The display label for the drilldown link.",
						Required:            true,
					},
					"encode_url": schema.BoolAttribute{
						MarkdownDescription: "When true, the URL is percent-encoded.",
						Optional:            true,
					},
					"open_in_new_tab": schema.BoolAttribute{
						MarkdownDescription: "When true, the drilldown URL opens in a new browser tab.",
						Optional:            true,
					},
				},
			},
		},
	}
}

// getSloSingleSchema returns the attributes for the single sub-block.
func getSloSingleSchema() map[string]schema.Attribute {
	attrs := getSloSharedDisplaySchema()
	attrs["slo_id"] = schema.StringAttribute{
		MarkdownDescription: "The unique identifier of the SLO to display.",
		Required:            true,
	}
	attrs["slo_instance_id"] = schema.StringAttribute{
		MarkdownDescription: "The SLO instance ID. Set when the SLO uses group_by; identifies which instance to display. Defaults to `*` (all instances) when omitted.",
		Optional:            true,
	}
	attrs["remote_name"] = schema.StringAttribute{
		MarkdownDescription: "The name of the remote cluster where the SLO is defined.",
		Optional:            true,
	}
	return attrs
}

// getSloGroupsSchema returns the attributes for the groups sub-block.
func getSloGroupsSchema() map[string]schema.Attribute {
	attrs := getSloSharedDisplaySchema()
	attrs["group_filters"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Optional filters for grouped SLO overview mode.",
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"group_by": schema.StringAttribute{
				MarkdownDescription: "Group SLOs by this field. Valid values are `slo.tags`, `status`, `slo.indicator.type`, `_index`.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("slo.tags", "status", "slo.indicator.type", "_index"),
				},
			},
			"groups": schema.ListAttribute{
				MarkdownDescription: "List of group values to include (maximum 100).",
				Optional:            true,
				ElementType:         types.StringType,
				Validators: []validator.List{
					listvalidator.SizeAtMost(100),
				},
			},
			"kql_query": schema.StringAttribute{
				MarkdownDescription: "KQL query string to filter the SLOs shown in the group overview.",
				Optional:            true,
			},
			"filters_json": schema.StringAttribute{
				MarkdownDescription: "AS-code filter array as a JSON string. Accepts the polymorphic filter schema (condition, group, DSL, spatial).",
				CustomType:          jsontypes.NormalizedType{},
				Optional:            true,
			},
		},
	}
	return attrs
}

// sloOverviewConfigModeValidator ensures exactly one of single or groups is set.
var _ validator.Object = sloOverviewConfigModeValidator{}

type sloOverviewConfigModeValidator struct{}

func (v sloOverviewConfigModeValidator) Description(_ context.Context) string {
	return "Ensures exactly one of `single` or `groups` is configured inside `slo_overview_config`."
}

func (v sloOverviewConfigModeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v sloOverviewConfigModeValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	attrs := req.ConfigValue.Attributes()
	singleVal := attrs["single"]
	groupsVal := attrs["groups"]

	singleSet := singleVal != nil && !singleVal.IsNull() && !singleVal.IsUnknown()
	groupsSet := groupsVal != nil && !groupsVal.IsNull() && !groupsVal.IsUnknown()

	if singleSet && groupsSet {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid slo_overview_config", "Exactly one of `single` or `groups` must be configured inside `slo_overview_config`, not both.")
		return
	}
	if !singleSet && !groupsSet {
		// Both unknown is acceptable (during planning with computed resources).
		singleUnknown := singleVal != nil && singleVal.IsUnknown()
		groupsUnknown := groupsVal != nil && groupsVal.IsUnknown()
		if singleUnknown || groupsUnknown {
			return
		}
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid slo_overview_config", "Exactly one of `single` or `groups` must be configured inside `slo_overview_config`.")
	}
}
