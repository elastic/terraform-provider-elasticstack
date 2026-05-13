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

	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/validators"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
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
	dashboardValueAuto               = "auto"
	dashboardValueAverage            = "average"
	pieChartTypeNumber               = "number"
	pieChartTypePercent              = "percent"
	operationTerms                   = "terms"
	panelTypeImage                   = "image"
	panelTypeMarkdown                = "markdown"
	panelTypeVis                     = "vis"
	panelTypeTimeSlider              = "time_slider_control"
	panelTypeSloAlerts               = "slo_alerts"
	panelTypeDiscoverSession         = "discover_session"
	panelTypeSloBurnRate             = "slo_burn_rate"
	panelTypeSloErrorBudget          = "slo_error_budget"
	panelTypeEsqlControl             = "esql_control"
	panelTypeOptionsListControl      = "options_list_control"
	panelTypeRangeSlider             = "range_slider_control"
	panelTypeSyntheticsStatsOverview = "synthetics_stats_overview"
	panelTypeSyntheticsMonitors      = "synthetics_monitors"
	panelTypeLensDashboardApp        = "lens-dashboard-app"
)

var sloBurnRateDurationRegex = regexp.MustCompile(`^\d+[mhd]$`)

var panelConfigNames = []string{
	"config_json",
	"markdown_config",
	"vis_config",
	"lens_dashboard_app_config",
	"esql_control_config",
	"options_list_control_config",
	"range_slider_control_config",
	"time_slider_control_config",
	"slo_alerts_config",
	"slo_burn_rate_config",
	"slo_overview_config",
	"slo_error_budget_config",
	"synthetics_monitors_config",
	"synthetics_stats_overview_config",
	"image_config",
	"discover_session_config",
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
	if _, exists := model["color"]; !exists {
		model["color"] = map[string]any{"type": "auto"}
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

func populateMetricChartMetricDefaults(model map[string]any) map[string]any {
	_, hadColor := model["color"]
	model = populateLensMetricDefaults(model)
	if model == nil {
		return model
	}

	if metricType, _ := model["type"].(string); metricType == "secondary" && !hadColor {
		model["color"] = map[string]any{"type": "none"}
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
		model["ticks"] = map[string]any{"visible": true, "mode": "bands"}
	}
	if _, exists := model["color"]; !exists {
		model["color"] = map[string]any{"type": "auto"}
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
		if _, exists := model["color"]; !exists {
			model["color"] = map[string]any{"type": "auto"}
		}
	}
	return model
}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = getSchema()
}

func getSchema() schema.Schema {
	dashboardNotes := "### Notes\n\n" +
		"- **Image `file_id`**: `image_config.src.file.file_id` is an opaque Kibana file asset id. " +
		"Uploading or lifecycle-managing that file is outside this resource for now; prepare the id outside Terraform " +
		"(for example via Kibana UI or HTTP upload). A future `elasticstack_kibana_file` resource may cover uploads.\n" +
		"- **`discover_session` `data_source_json`**: Must be JSON matching the Kibana Dashboard API tab payload — " +
		"the polymorphic data source for DSL tabs (`data_view_reference`, `data_view_spec`, etc.) and the ES|QL branch " +
		"for `tab.esql`. Follow the OpenAPI shapes published with the [Kibana REST API]" +
		"(https://www.elastic.co/docs/api/doc/kibana) (`kbn-dashboard-panel-type-discover_session`). " +
		"For `data_view_reference`, use **`ref_id`** (not `id`) for the linked data view.\n" +
		"- **Single Discover tab**: `discover_session_config.by_value.tab` is one object because the API currently allows " +
		"a single tab entry; a future `tabs` list could be added without breaking existing configs if Kibana lifts the limit."

	return schema.Schema{
		MarkdownDescription: "Manages Kibana [dashboards](https://www.elastic.co/docs/api/doc/kibana). " +
			"This functionality is in technical preview and may be changed or removed in a future release.\n\n" +
			dashboardNotes,
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
			"time_range": timeRangeSingleNestedAttribute(
				"Dashboard time selection (`from`, `to`, optional `mode`). Aligns with the Kibana Dashboard API `time_range` object.",
				true,
			),
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
						MarkdownDescription: "Query language (`kql` or `lucene`).",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("kql", "lucene"),
						},
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
			"filters": schema.ListNestedAttribute{
				MarkdownDescription: dashboardFiltersDescription,
				Optional:            true,
				NestedObject:        getDashboardRootSavedFiltersNestedObject(),
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
			"pinned_panels": schema.ListNestedAttribute{
				MarkdownDescription: strings.TrimSpace(pinnedPanelsDescription),
				Optional:            true,
				NestedObject:        pinnedPanelsNestedObject(),
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
							MarkdownDescription: "The identifier of the section (API `id`).",
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

		Blocks: map[string]schema.Block{
			"kibana_connection": providerschema.GetKbFWConnectionBlock(),
		}}
}

func getPanelSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Validators: []validator.Object{
			panelConfigValidator{},
		},
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of the panel (e.g. 'markdown', 'vis').",
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
				MarkdownDescription: "The identifier of the panel (API `id`).",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseNonNullStateForUnknown(),
				},
			},
			"markdown_config": schema.SingleNestedAttribute{
				MarkdownDescription: panelConfigDescription(
					"Configuration for a `markdown` panel (the Kibana Dashboard API `kbn-dashboard-panel-type-markdown` shape). "+
						"Set exactly one of `by_value` (inline `content` with required nested `settings`) or `by_reference` (existing library item via `ref_id`). "+
						"Presentation fields (`description`, `hide_title`, `title`, `hide_border`) are supported in both branches.",
					"markdown_config",
					panelConfigNames,
				),
				Optional:   true,
				Attributes: getMarkdownConfigSchema(),
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(
						siblingPanelConfigPathsExcept("markdown_config", panelConfigNames)...,
					),
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelTypeMarkdown}),
					markdownConfigModeValidator{},
				},
			},
			"time_slider_control_config": panelTimeSliderControlConfigSchema(),
			"slo_alerts_config": schema.SingleNestedAttribute{
				MarkdownDescription: panelConfigDescription(
					"Configuration for an `slo_alerts` panel (`kbn-dashboard-panel-type-slo_alerts`). "+
						"Required when `type` is `slo_alerts`.",
					"slo_alerts_config",
					panelConfigNames,
				),
				Optional:   true,
				Attributes: getSloAlertsPanelConfigAttributes(),
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(
						siblingPanelConfigPathsExcept("slo_alerts_config", panelConfigNames)...,
					),
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelTypeSloAlerts}),
					validators.RequiredIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelTypeSloAlerts}),
				},
			},
			"discover_session_config": schema.SingleNestedAttribute{
				MarkdownDescription: panelConfigDescription(
					"Configuration for a `discover_session` panel (`kbn-dashboard-panel-type-discover_session`). "+
						"Required when `type` is `discover_session`. Set exactly one of `by_value` or `by_reference`.",
					"discover_session_config",
					panelConfigNames,
				),
				Optional:   true,
				Attributes: getDiscoverSessionPanelConfigAttributes(),
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(
						siblingPanelConfigPathsExcept("discover_session_config", panelConfigNames)...,
					),
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelTypeDiscoverSession}),
					validators.RequiredIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelTypeDiscoverSession}),
					discoverSessionConfigModeValidator{},
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
						NestedObject: urlDrilldownNestedAttributeObject(URLDrilldownNestedOpts{
							AllowedTriggers:                 []string{"on_open_panel_menu"},
							URLMarkdownDescription:          "Templated URL for the drilldown.",
							LabelMarkdownDescription:        "Display label shown in the drilldown menu.",
							EncodeURLMarkdownDescription:    "When true, the URL is percent-encoded. Omit to use the API default.",
							OpenInNewTabMarkdownDescription: "When true, the URL opens in a new browser tab. Omit to use the API default.",
						}),
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
			"esql_control_config":         panelEsqlControlConfigSchema(),
			"options_list_control_config": panelOptionsListControlConfigSchema(),
			"range_slider_control_config": panelRangeSliderControlConfigSchema(),
			"synthetics_stats_overview_config": schema.SingleNestedAttribute{
				MarkdownDescription: panelConfigDescription(
					"Configuration for a Synthetics stats overview panel. "+
						"All fields are optional; an absent or empty block shows statistics "+
						"for all monitors visible within the space.",
					"synthetics_stats_overview_config",
					panelConfigNames,
				),
				Optional:   true,
				Attributes: getSyntheticsStatsOverviewSchema(),
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(
						siblingPanelConfigPathsExcept("synthetics_stats_overview_config", panelConfigNames)...,
					),
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelTypeSyntheticsStatsOverview}),
				},
			},
			"synthetics_monitors_config": schema.SingleNestedAttribute{
				MarkdownDescription: panelConfigDescription(
					"Configuration for a Synthetics monitors panel. Displays a table of Elastic Synthetics monitors "+
						"and their current status. All fields are optional — omit the block entirely for a bare panel with no filtering.",
					"synthetics_monitors_config",
					panelConfigNames,
				),
				Optional:   true,
				Attributes: getSyntheticsMonitorsSchema(),
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(
						siblingPanelConfigPathsExcept("synthetics_monitors_config", panelConfigNames)...,
					),
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelTypeSyntheticsMonitors}),
				},
			},
			"image_config": schema.SingleNestedAttribute{
				MarkdownDescription: panelConfigDescription(
					"Configuration for an `image` panel (`kbn-dashboard-panel-type-image`). Required when `type` is `image`. "+
						"References the Kibana Dashboard API image embeddable `config` shape.",
					"image_config",
					panelConfigNames,
				),
				Optional:   true,
				Attributes: getImagePanelConfigAttributes(),
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(
						siblingPanelConfigPathsExcept("image_config", panelConfigNames)...,
					),
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelTypeImage}),
					validators.RequiredIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelTypeImage}),
				},
			},
			"lens_dashboard_app_config": schema.SingleNestedAttribute{
				MarkdownDescription: panelConfigDescription(
					"Configuration for a `lens-dashboard-app` panel (the Kibana Dashboard API `lens-dashboard-app` panel type). "+
						"Required when `type` is `lens-dashboard-app`. "+
						"Set exactly one of `by_value` or `by_reference`. "+
						"With `by_value`, set exactly one of `config_json` or one supported typed Lens chart block. "+
						"With `by_reference`, use `ref_id` and `references_json` to map the API `references` list. "+
						"Supported typed by-value blocks are sent as the `lens-dashboard-app` API `config` and do not use `type = \"vis\"` panels.",
					"lens_dashboard_app_config",
					panelConfigNames,
				),
				Optional:   true,
				Attributes: getLensDashboardAppConfigSchema(),
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(
						siblingPanelConfigPathsExcept("lens_dashboard_app_config", panelConfigNames)...,
					),
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelTypeLensDashboardApp}),
					validators.RequiredIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelTypeLensDashboardApp}),
					lensDashboardAppConfigModeValidator{},
				},
			},
			"vis_config": schema.SingleNestedAttribute{
				MarkdownDescription: panelConfigDescription(
					"Configuration for a `vis` panel (`type = \"vis\"`). "+
						"Typed alternative to `config_json`: set exactly one of `by_value` (exactly one of 12 Lens chart kinds) or `by_reference`. "+
						"With `by_reference`, use structured `drilldowns` and required `time_range` like `lens_dashboard_app_config.by_reference`.",
					"vis_config",
					panelConfigNames,
				),
				Optional:   true,
				Attributes: getVisConfigSchema(),
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(
						siblingPanelConfigPathsExcept("vis_config", panelConfigNames)...,
					),
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelTypeVis}),
					visConfigModeValidator{},
				},
			},
			"config_json": schema.StringAttribute{
				MarkdownDescription: panelConfigDescription(
					"The configuration of the panel as a JSON string. "+
						"Practitioner-authored panel-level `config_json` is valid only when `type` is `markdown` or `vis`. "+
						"Typed panel kinds such as `lens-dashboard-app`, `image`, `slo_alerts`, and `discover_session` use their dedicated blocks "+
						"(`lens_dashboard_app_config`, `image_config`, `slo_alerts_config`, `discover_session_config`), not panel-level `config_json`.",
					"config_json",
					panelConfigNames,
				),
				CustomType: customtypes.NewJSONWithDefaultsType(populatePanelConfigJSONDefaults),
				Optional:   true,
				Computed:   true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(
						siblingPanelConfigPathsExcept("config_json", panelConfigNames)...,
					),
					validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelTypeVis, panelTypeMarkdown}),
				},
			},
		},
	}
}

// lensByValueVisMirrorDescription documents a typed by-value block whose shape matches the same-named block under `vis_config.by_value`.
func lensByValueVisMirrorDescription(visBlockName string) string {
	return "Typed Lens chart for a `lens-dashboard-app` by-value panel. The chart is sent as the Kibana `lens-dashboard-app` API `config` and does not create a `type = \"vis\"` panel. " +
		"Attribute shape matches `vis_config.by_value." + visBlockName + "` for `type = \"vis\"` panels."
}

// visConfigByValueBlockDescription documents a typed `vis_config.by_value` chart block.
func visConfigByValueBlockDescription(visSiblingName string) string {
	return "Typed Lens visualization inside `vis_config.by_value`. " +
		"Mutually exclusive with the other chart blocks in the same `by_value` block. " +
		"Shares the attribute shape with `lens_dashboard_app_config.by_value." + visSiblingName + "`."
}

// visByValueSourceAttrNames lists mutually exclusive typed chart kinds under `vis_config.by_value`.
// Keep in sync with getVisByValueAttributes().
var visByValueSourceAttrNames = []string{
	"xy_chart_config",
	"metric_chart_config",
	"legacy_metric_config",
	"gauge_config",
	"heatmap_config",
	"tagcloud_config",
	"region_map_config",
	"datatable_config",
	"pie_chart_config",
	"mosaic_config",
	"treemap_config",
	"waffle_config",
}

// getVisByValueAttributes returns typed chart attributes for `vis_config.by_value` (inline `vis` config).
func getVisByValueAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"xy_chart_config": schema.SingleNestedAttribute{
			MarkdownDescription: visConfigByValueBlockDescription("xy_chart_config"),
			Optional:            true,
			Attributes:          getXYChartConfigAttributes(true),
		},
		"metric_chart_config": schema.SingleNestedAttribute{
			MarkdownDescription: visConfigByValueBlockDescription("metric_chart_config"),
			Optional:            true,
			Attributes:          getMetricChart(true),
		},
		"legacy_metric_config": schema.SingleNestedAttribute{
			MarkdownDescription: visConfigByValueBlockDescription("legacy_metric_config"),
			Optional:            true,
			Attributes:          getLegacyMetricSchema(true),
		},
		"gauge_config": schema.SingleNestedAttribute{
			MarkdownDescription: visConfigByValueBlockDescription("gauge_config"),
			Optional:            true,
			Attributes:          getGaugeSchema(true),
		},
		"heatmap_config": schema.SingleNestedAttribute{
			MarkdownDescription: visConfigByValueBlockDescription("heatmap_config"),
			Optional:            true,
			Attributes:          getHeatmapSchema(true),
		},
		"tagcloud_config": schema.SingleNestedAttribute{
			MarkdownDescription: visConfigByValueBlockDescription("tagcloud_config"),
			Optional:            true,
			Attributes:          getTagcloudSchema(true),
		},
		"region_map_config": schema.SingleNestedAttribute{
			MarkdownDescription: visConfigByValueBlockDescription("region_map_config"),
			Optional:            true,
			Attributes:          getRegionMapSchema(true),
		},
		"datatable_config": schema.SingleNestedAttribute{
			MarkdownDescription: visConfigByValueBlockDescription("datatable_config"),
			Optional:            true,
			Attributes:          getDatatableSchema(true),
		},
		"pie_chart_config": schema.SingleNestedAttribute{
			MarkdownDescription: visConfigByValueBlockDescription("pie_chart_config"),
			Optional:            true,
			Attributes:          getPieChart(true),
		},
		"mosaic_config": schema.SingleNestedAttribute{
			MarkdownDescription: visConfigByValueBlockDescription("mosaic_config"),
			Optional:            true,
			Attributes:          getMosaicSchema(true),
		},
		"treemap_config": schema.SingleNestedAttribute{
			MarkdownDescription: visConfigByValueBlockDescription("treemap_config"),
			Optional:            true,
			Attributes:          getTreemapSchema(true),
		},
		"waffle_config": schema.SingleNestedAttribute{
			MarkdownDescription: visConfigByValueBlockDescription("waffle_config"),
			Optional:            true,
			Attributes:          getWaffleSchema(true),
			Validators: []validator.Object{
				waffleConfigModeValidator{},
			},
		},
	}
}

// lensDashboardAppByValueSourceAttrNames lists mutually exclusive by-value content attributes.
// Keep in sync with getLensDashboardAppByValueNestedAttributes.
var lensDashboardAppByValueSourceAttrNames = []string{
	"config_json",
	"xy_chart_config",
	"treemap_config",
	"mosaic_config",
	"datatable_config",
	"tagcloud_config",
	"heatmap_config",
	"waffle_config",
	"region_map_config",
	"gauge_config",
	"metric_chart_config",
	"pie_chart_config",
	"legacy_metric_config",
}

// getLensDashboardAppByValueNestedAttributes returns attributes for `lens_dashboard_app_config.by_value`.
func getLensDashboardAppByValueNestedAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"config_json": schema.StringAttribute{
			MarkdownDescription: "Optional raw normalized JSON for the by-value Lens chart `config` (full API shape, including chart `type` and `time_range` where the API requires them). " +
				"Use as the single `by_value` source, or use one supported typed chart block instead (not both). " +
				"Distinct from panel-level `config_json` on the panel.",
			Optional:   true,
			CustomType: jsontypes.NormalizedType{},
		},
		"xy_chart_config": schema.SingleNestedAttribute{
			MarkdownDescription: lensByValueVisMirrorDescription("xy_chart_config"),
			Optional:            true,
			Attributes:          getXYChartConfigAttributes(false),
		},
		"treemap_config": schema.SingleNestedAttribute{
			MarkdownDescription: lensByValueVisMirrorDescription("treemap_config"),
			Optional:            true,
			Attributes:          getTreemapSchema(false),
		},
		"mosaic_config": schema.SingleNestedAttribute{
			MarkdownDescription: lensByValueVisMirrorDescription("mosaic_config"),
			Optional:            true,
			Attributes:          getMosaicSchema(false),
		},
		"datatable_config": schema.SingleNestedAttribute{
			MarkdownDescription: lensByValueVisMirrorDescription("datatable_config"),
			Optional:            true,
			Attributes:          getDatatableSchema(false),
		},
		"tagcloud_config": schema.SingleNestedAttribute{
			MarkdownDescription: lensByValueVisMirrorDescription("tagcloud_config"),
			Optional:            true,
			Attributes:          getTagcloudSchema(false),
		},
		"heatmap_config": schema.SingleNestedAttribute{
			MarkdownDescription: lensByValueVisMirrorDescription("heatmap_config"),
			Optional:            true,
			Attributes:          getHeatmapSchema(false),
		},
		"waffle_config": schema.SingleNestedAttribute{
			MarkdownDescription: lensByValueVisMirrorDescription("waffle_config"),
			Optional:            true,
			Attributes:          getWaffleSchema(false),
			Validators: []validator.Object{
				waffleConfigModeValidator{},
			},
		},
		"region_map_config": schema.SingleNestedAttribute{
			MarkdownDescription: lensByValueVisMirrorDescription("region_map_config"),
			Optional:            true,
			Attributes:          getRegionMapSchema(false),
		},
		"gauge_config": schema.SingleNestedAttribute{
			MarkdownDescription: lensByValueVisMirrorDescription("gauge_config"),
			Optional:            true,
			Attributes:          getGaugeSchema(false),
		},
		"metric_chart_config": schema.SingleNestedAttribute{
			MarkdownDescription: lensByValueVisMirrorDescription("metric_chart_config"),
			Optional:            true,
			Attributes:          getMetricChart(false),
		},
		"pie_chart_config": schema.SingleNestedAttribute{
			MarkdownDescription: lensByValueVisMirrorDescription("pie_chart_config"),
			Optional:            true,
			Attributes:          getPieChart(false),
		},
		"legacy_metric_config": schema.SingleNestedAttribute{
			MarkdownDescription: lensByValueVisMirrorDescription("legacy_metric_config"),
			Optional:            true,
			Attributes:          getLegacyMetricSchema(false),
		},
	}
}

// getLensByReferenceAttributes returns the by-reference attribute map shared by `lens_dashboard_app_config.by_reference`
// and `vis_config.by_reference`. Drilldowns are authored exclusively via the structured `drilldowns` block.
func getLensByReferenceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"ref_id": schema.StringAttribute{
			MarkdownDescription: "Reference name in the API `ref_id` field. When `references_json` is set, `ref_id` typically should match a `name` in that list so the link resolves as expected.",
			Required:            true,
		},
		"references_json": schema.StringAttribute{
			MarkdownDescription: "Optional normalized JSON array of `{ id, name, type }` saved-object references, matching the API `references` list (for example wiring a `lens` saved object to `ref_id`).",
			Optional:            true,
			CustomType:          jsontypes.NormalizedType{},
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
			MarkdownDescription: "When true, suppresses the panel title.",
			Optional:            true,
		},
		"hide_border": schema.BoolAttribute{
			MarkdownDescription: "When true, suppresses the panel border.",
			Optional:            true,
		},
		"drilldowns": getStructuredDrilldownsAttribute(),
		"time_range": schema.SingleNestedAttribute{
			MarkdownDescription: "Required time range for the by-reference panel config " +
				"(used by both `lens_dashboard_app_config.by_reference` and `vis_config.by_reference`).",
			Required: true,
			Attributes: map[string]schema.Attribute{
				"from": schema.StringAttribute{
					MarkdownDescription: "Range start, matching the Kibana time range `from` field.",
					Required:            true,
				},
				"to": schema.StringAttribute{
					MarkdownDescription: "Range end, matching the Kibana time range `to` field.",
					Required:            true,
				},
				"mode": schema.StringAttribute{
					MarkdownDescription: "Optional time range mode. When set, must be `absolute` or `relative`.",
					Optional:            true,
					Validators: []validator.String{
						stringvalidator.OneOf("absolute", "relative"),
					},
				},
			},
		},
	}
}

// getLensDashboardAppConfigSchema returns attributes for the lens_dashboard_app_config block.
func getLensDashboardAppConfigSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"by_value": schema.SingleNestedAttribute{
			MarkdownDescription: "Inline by-value `lens-dashboard-app` configuration. " +
				"Set exactly one of `config_json` (raw JSON) or one supported typed Lens chart block, not both. " +
				"Typed by-value blocks send the chart as the Kibana `lens-dashboard-app` API `config` and do not create a `type = \"vis\"` panel. " +
				"On read, when state used a typed chart block and the API `config` can be round-tripped into that same block, the provider repopulates the typed block; " +
				"otherwise the response is reflected in `config_json` instead. " +
				"Distinct from panel-level `config_json` on the panel.",
			Optional:   true,
			Attributes: getLensDashboardAppByValueNestedAttributes(),
			Validators: []validator.Object{
				lensDashboardAppByValueSourceValidator{},
			},
		},
		"by_reference": schema.SingleNestedAttribute{
			MarkdownDescription: "By-reference `lens-dashboard-app` configuration: link a saved Lens visualization via `ref_id`, optional `references_json`, " +
				"optional structured `drilldowns`, and required `time_range`.",
			Optional:   true,
			Attributes: getLensByReferenceAttributes(),
		},
	}
}

// getVisConfigSchema returns attributes for `vis_config` on `vis` panels (symmetric with `getLensDashboardAppConfigSchema`, minus `by_value.config_json`).
func getVisConfigSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"by_value": schema.SingleNestedAttribute{
			MarkdownDescription: "Inline by-value Lens visualization configuration for `type = \"vis\"` panels (`vis_config`). " +
				"Exactly one typed chart kind must be set (no raw JSON here — use panel-level `config_json` for that).",
			Optional:   true,
			Attributes: getVisByValueAttributes(),
			Validators: []validator.Object{
				visByValueSourceValidator{},
			},
		},
		"by_reference": schema.SingleNestedAttribute{
			MarkdownDescription: "By-reference `vis` configuration: structured `drilldowns`, `ref_id`, optional `references_json`, and required `time_range`. " +
				"Shares the attribute shape with `lens_dashboard_app_config.by_reference` via `getLensByReferenceAttributes()`.",
			Optional:   true,
			Attributes: getLensByReferenceAttributes(),
		},
	}
}

// validateExactlyOneNestedAttr counts the named child attributes of a nested object
// that are concretely set (not null, not unknown). It writes diagnostics scoped to req.Path
// when the count is wrong, deferring (no diagnostic) if any candidate is unknown so plan
// re-runs after refinement get a final verdict.
//
// Used by every "exactly one of …" object validator on the dashboard schema (mode validators
// that gate `by_value` vs `by_reference`, source validators that pick a chart kind under
// `by_value`, etc.). It assumes req.ConfigValue is non-null and non-known-unknown — the
// caller short-circuits on those because the framework asks each validator separately.
func validateExactlyOneNestedAttr(
	req validator.ObjectRequest,
	resp *validator.ObjectResponse,
	blockLabel string,
	attrNames []string,
	missingDetail string,
	tooManyDetail string,
) {
	attrs := req.ConfigValue.Attributes()
	count := 0
	hasUnknown := false
	for _, name := range attrNames {
		av, ok := attrs[name]
		if !ok || av == nil {
			continue
		}
		switch {
		case av.IsUnknown():
			hasUnknown = true
		case av.IsNull():
			// not set
		default:
			count++
		}
	}
	if count > 1 {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid "+blockLabel, tooManyDetail)
		return
	}
	if hasUnknown {
		return
	}
	if count == 0 {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid "+blockLabel, missingDetail)
	}
}

// modeAttrNames lists the two children of every block that uses the by_value / by_reference union.
var modeAttrNames = []string{"by_value", "by_reference"}

// lensDashboardAppConfigModeValidator enforces that exactly one of by_value or by_reference is set.
var _ validator.Object = lensDashboardAppConfigModeValidator{}

type lensDashboardAppConfigModeValidator struct{}

func (v lensDashboardAppConfigModeValidator) Description(_ context.Context) string {
	return "Ensures exactly one of `by_value` or `by_reference` is set inside `lens_dashboard_app_config`."
}

func (v lensDashboardAppConfigModeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v lensDashboardAppConfigModeValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	validateExactlyOneNestedAttr(
		req, resp,
		"lens_dashboard_app_config",
		modeAttrNames,
		"Exactly one of `by_value` or `by_reference` must be set inside `lens_dashboard_app_config`.",
		"Exactly one of `by_value` or `by_reference` must be set inside `lens_dashboard_app_config`, not both.",
	)
}

// lensDashboardAppByValueSourceValidator enforces exactly one of config_json or a typed chart block inside by_value.
var _ validator.Object = lensDashboardAppByValueSourceValidator{}

type lensDashboardAppByValueSourceValidator struct{}

func (lensDashboardAppByValueSourceValidator) Description(_ context.Context) string {
	return "Ensures exactly one of `config_json` or one supported typed Lens chart block is set inside `by_value`."
}

func (v lensDashboardAppByValueSourceValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (lensDashboardAppByValueSourceValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	validateExactlyOneNestedAttr(
		req, resp,
		"lens_dashboard_app_config.by_value",
		lensDashboardAppByValueSourceAttrNames,
		"Set exactly one of `config_json` or one supported typed Lens chart block inside `by_value`.",
		"Set exactly one of `config_json` or one supported typed Lens chart block inside `by_value` (more than one by-value source is set).",
	)
}

// getMarkdownConfigSchema returns attributes for the markdown_config block.
func getMarkdownConfigSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"by_value": schema.SingleNestedAttribute{
			MarkdownDescription: "Inline markdown: required `content` and nested `settings` (API `settings` object). " +
				"Optional `description`, `hide_title`, `title`, and `hide_border`.",
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"content": schema.StringAttribute{
					MarkdownDescription: "Markdown source for the panel body (API `content`).",
					Required:            true,
				},
				"settings": schema.SingleNestedAttribute{
					MarkdownDescription: "Required settings object for by-value markdown. " +
						"`open_links_in_new_tab` is optional; when unset, Kibana applies its default (`true`).",
					Required: true,
					Attributes: map[string]schema.Attribute{
						"open_links_in_new_tab": schema.BoolAttribute{
							MarkdownDescription: "When true, links in the markdown open in a new tab. When omitted, Kibana defaults to true.",
							Optional:            true,
						},
					},
				},
				"description": schema.StringAttribute{
					MarkdownDescription: "Optional panel description.",
					Optional:            true,
				},
				"hide_title": schema.BoolAttribute{
					MarkdownDescription: "When true, suppresses the panel title.",
					Optional:            true,
				},
				"title": schema.StringAttribute{
					MarkdownDescription: "Optional panel title.",
					Optional:            true,
				},
				"hide_border": schema.BoolAttribute{
					MarkdownDescription: "When true, suppresses the panel border.",
					Optional:            true,
				},
			},
		},
		"by_reference": schema.SingleNestedAttribute{
			MarkdownDescription: "Reference an existing markdown library item via `ref_id`. " +
				"Optional `description`, `hide_title`, `title`, and `hide_border`.",
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"ref_id": schema.StringAttribute{
					MarkdownDescription: "Unique identifier of the markdown library item (API `ref_id`). The provider does not verify the item exists at plan time.",
					Required:            true,
				},
				"description": schema.StringAttribute{
					MarkdownDescription: "Optional panel description.",
					Optional:            true,
				},
				"hide_title": schema.BoolAttribute{
					MarkdownDescription: "When true, suppresses the panel title.",
					Optional:            true,
				},
				"title": schema.StringAttribute{
					MarkdownDescription: "Optional panel title.",
					Optional:            true,
				},
				"hide_border": schema.BoolAttribute{
					MarkdownDescription: "When true, suppresses the panel border.",
					Optional:            true,
				},
			},
		},
	}
}

// markdownConfigModeValidator enforces that exactly one of by_value or by_reference is set.
var _ validator.Object = markdownConfigModeValidator{}

type markdownConfigModeValidator struct{}

func (v markdownConfigModeValidator) Description(_ context.Context) string {
	return "Ensures exactly one of `by_value` or `by_reference` is set inside `markdown_config`."
}

func (v markdownConfigModeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v markdownConfigModeValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	validateExactlyOneNestedAttr(
		req, resp,
		"markdown_config",
		modeAttrNames,
		"Exactly one of `by_value` or `by_reference` must be set inside `markdown_config`.",
		"Exactly one of `by_value` or `by_reference` must be set inside `markdown_config`, not both.",
	)
}

// visConfigModeValidator enforces that exactly one of by_value or by_reference is set under `vis_config`.
var _ validator.Object = visConfigModeValidator{}

type visConfigModeValidator struct{}

func (visConfigModeValidator) Description(_ context.Context) string {
	return "Ensures exactly one of `by_value` or `by_reference` is set inside `vis_config`."
}

func (v visConfigModeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (visConfigModeValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	validateExactlyOneNestedAttr(
		req, resp,
		"vis_config",
		modeAttrNames,
		"Exactly one of `by_value` or `by_reference` must be set inside `vis_config`.",
		"Exactly one of `by_value` or `by_reference` must be set inside `vis_config`, not both.",
	)
}

// visByValueSourceValidator enforces exactly one typed chart kind inside `vis_config.by_value`.
var _ validator.Object = visByValueSourceValidator{}

type visByValueSourceValidator struct{}

func (visByValueSourceValidator) Description(_ context.Context) string {
	return "Ensures exactly one supported typed Lens chart block is set inside `vis_config.by_value`."
}

func (v visByValueSourceValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (visByValueSourceValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	validateExactlyOneNestedAttr(
		req, resp,
		"vis_config.by_value",
		visByValueSourceAttrNames,
		"Set exactly one supported typed Lens chart block inside `vis_config.by_value`.",
		"Set exactly one typed chart block inside `vis_config.by_value` (more than one by-value chart is set).",
	)
}

// getSyntheticsStatsOverviewSchema returns the schema attributes for the synthetics_stats_overview_config block.
func getSyntheticsStatsOverviewSchema() map[string]schema.Attribute {
	filterItemSchema := schema.ListNestedAttribute{
		Optional: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"label": schema.StringAttribute{
					MarkdownDescription: "Human-readable display label for the filter option.",
					Required:            true,
				},
				"value": schema.StringAttribute{
					MarkdownDescription: "Machine-readable value used for actual filtering.",
					Required:            true,
				},
			},
		},
	}

	return map[string]schema.Attribute{
		"title": schema.StringAttribute{
			MarkdownDescription: "Display title shown in the panel header.",
			Optional:            true,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "Descriptive text for the panel.",
			Optional:            true,
		},
		"hide_title": schema.BoolAttribute{
			MarkdownDescription: "When true, suppresses the panel title in the dashboard.",
			Optional:            true,
		},
		"hide_border": schema.BoolAttribute{
			MarkdownDescription: "When true, suppresses the panel border in the dashboard.",
			Optional:            true,
		},
		"drilldowns": schema.ListNestedAttribute{
			MarkdownDescription: "Optional list of URL drilldown actions attached to the panel. The API allows up to 100 drilldowns per panel.",
			Optional:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						MarkdownDescription: "Templated URL for the drilldown action. Variables are documented at https://www.elastic.co/docs/explore-analyze/dashboards/drilldowns#url-template-variable.",
						Required:            true,
					},
					"label": schema.StringAttribute{
						MarkdownDescription: "Human-readable label shown in the drilldown menu.",
						Required:            true,
					},
					"encode_url": schema.BoolAttribute{
						MarkdownDescription: "When true, the URL is percent-encoded. Omit to use the API default (`true`).",
						Optional:            true,
					},
					"open_in_new_tab": schema.BoolAttribute{
						MarkdownDescription: "When true, the drilldown opens in a new browser tab. Omit to use the API default (`true`).",
						Optional:            true,
					},
				},
			},
		},
		"filters": schema.SingleNestedAttribute{
			MarkdownDescription: "Optional Synthetics monitor filter constraints. Each filter category " +
				"accepts a list of `{ label, value }` objects. Omit the block or individual categories " +
				"to apply no filtering for those dimensions.",
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"projects": schema.ListNestedAttribute{
					MarkdownDescription: "Filter by Synthetics project.",
					Optional:            true,
					NestedObject:        filterItemSchema.NestedObject,
				},
				"tags": schema.ListNestedAttribute{
					MarkdownDescription: "Filter by monitor tag.",
					Optional:            true,
					NestedObject:        filterItemSchema.NestedObject,
				},
				"monitor_ids": schema.ListNestedAttribute{
					MarkdownDescription: "Filter by monitor ID. The API accepts up to 5000 entries.",
					Optional:            true,
					NestedObject:        filterItemSchema.NestedObject,
				},
				"locations": schema.ListNestedAttribute{
					MarkdownDescription: "Filter by monitor location.",
					Optional:            true,
					NestedObject:        filterItemSchema.NestedObject,
				},
				"monitor_types": schema.ListNestedAttribute{
					MarkdownDescription: "Filter by monitor type (e.g. `browser`, `http`).",
					Optional:            true,
					NestedObject:        filterItemSchema.NestedObject,
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

// getFilterSimple returns the schema for simple filter configuration
func getFilterSimple() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"language": schema.StringAttribute{
			MarkdownDescription: "Query language (default: 'kql').",
			Optional:            true,
			Computed:            true,
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

// getDashboardRootSavedFiltersNestedObject returns the nested object schema for one dashboard-level saved filter.
// Shape matches getChartFilter (filter_json with jsontypes.NormalizedType) for consistency with chart blocks.
func getDashboardRootSavedFiltersNestedObject() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"filter_json": schema.StringAttribute{
				MarkdownDescription: dashboardFilterJSONDescription,
				CustomType:          jsontypes.NormalizedType{},
				Required:            true,
			},
		},
	}
}

// getHeatmapSchema returns the schema for heatmap chart configuration.
// includePresentation merges REQ-037 fields for vis panels only.
func getHeatmapSchema(includePresentation bool) map[string]schema.Attribute {
	attrs := lensChartBaseAttributes()
	attrs["data_source_json"] = schema.StringAttribute{
		MarkdownDescription: "Dataset configuration as JSON. For standard heatmaps, this specifies the data view or index; for ES|QL, this specifies the ES|QL query dataset.",
		CustomType:          jsontypes.NormalizedType{},
		Required:            true,
	}
	attrs["query"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Query configuration for filtering data. Required for non-ES|QL heatmaps.",
		Optional:            true,
		Attributes:          getFilterSimple(),
	}
	attrs["axis"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Axis configuration for X and Y axes.",
		Required:            true,
		Attributes:          getHeatmapAxesSchema(),
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
	attrs["styling"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Heatmap styling configuration.",
		Required:            true,
		Attributes: map[string]schema.Attribute{
			"cells": schema.SingleNestedAttribute{
				MarkdownDescription: "Cells configuration for the heatmap.",
				Required:            true,
				Attributes:          getHeatmapCellsSchema(),
			},
		},
	}
	if includePresentation {
		maps.Copy(attrs, lensChartPresentationAttributes())
	}
	return attrs
}

var (
	_ validator.Object = drilldownListItemVariantsValidator{}

	exprDrilldownDashboardDrilldown = path.MatchRelative().AtParent().AtName("dashboard_drilldown")
	exprDrilldownDiscoverDrilldown  = path.MatchRelative().AtParent().AtName("discover_drilldown")
	exprDrilldownURLDrilldown       = path.MatchRelative().AtParent().AtName("url_drilldown")
)

// drilldownListItemVariantsValidator rejects drilldown list items where none of the three variant blocks are set.
// Pairwise mutual exclusion when multiple variants are set is enforced via
// validators.ForbiddenIfDrilldownVariantSiblingNestedPresent on each variant block (REQ-039).
type drilldownListItemVariantsValidator struct{}

func (drilldownListItemVariantsValidator) Description(_ context.Context) string {
	return "Requires at least one drilldown variant (`dashboard_drilldown`, `discover_drilldown`, or `url_drilldown`); multiple variants are rejected by sibling object validators."
}

func (v drilldownListItemVariantsValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (drilldownListItemVariantsValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsUnknown() {
		return
	}
	if req.ConfigValue.IsNull() {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid drilldown",
			"Set exactly one of `dashboard_drilldown`, `discover_drilldown`, or `url_drilldown`.",
		)
		return
	}
	attrs := req.ConfigValue.Attributes()
	count := 0
	for _, key := range []string{"dashboard_drilldown", "discover_drilldown", "url_drilldown"} {
		val, okAttr := attrs[key]
		if !okAttr || val.IsNull() || val.IsUnknown() {
			continue
		}
		count++
	}
	if count == 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid drilldown",
			"Set exactly one of `dashboard_drilldown`, `discover_drilldown`, or `url_drilldown`.",
		)
	}
}

// lensChartPresentationAttributes returns optional chart-root presentation fields shared by all typed Lens chart blocks:
// `time_range` (inherits dashboard-level when null — see REQ-038), `hide_title`, `hide_border`, `references_json`, and `drilldowns`.
func lensChartPresentationAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"time_range": schema.SingleNestedAttribute{
			MarkdownDescription: "Chart-level time selection (`from`, `to`, optional `mode`), same shape as the dashboard root `time_range`. " +
				"When omitted (null), the provider inherits the dashboard-level `time_range` on write and preserves null in state when the API echoes the inherited value on read.",
			Optional:   true,
			Attributes: lensChartPresentationTimeRangeAttributes(),
		},
		"hide_title": schema.BoolAttribute{
			MarkdownDescription: "When true, suppresses the chart title.",
			Optional:            true,
		},
		"hide_border": schema.BoolAttribute{
			MarkdownDescription: "When true, suppresses the chart panel border.",
			Optional:            true,
		},
		"references_json": schema.StringAttribute{
			MarkdownDescription: "Optional normalized JSON array of `{ id, name, type }` saved-object references, matching the chart root API `references` list.",
			Optional:            true,
			CustomType:          jsontypes.NormalizedType{},
		},
		"drilldowns": schema.ListNestedAttribute{
			MarkdownDescription: "Optional drilldowns for this chart (max 100 per Kibana API). Each entry sets exactly one of `dashboard_drilldown`, `discover_drilldown`, or `url_drilldown`.",
			Optional:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: lensChartDrilldownListItemAttributes(),
				Validators: []validator.Object{
					drilldownListItemVariantsValidator{},
				},
			},
			Validators: []validator.List{
				listvalidator.SizeAtMost(100),
			},
		},
	}
}

func lensChartPresentationTimeRangeAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"from": schema.StringAttribute{
			MarkdownDescription: "Start of the chart time range.",
			Required:            true,
		},
		"to": schema.StringAttribute{
			MarkdownDescription: "End of the chart time range.",
			Required:            true,
		},
		"mode": schema.StringAttribute{
			MarkdownDescription: "Optional time range mode. Valid values are `absolute` or `relative`. " +
				"When the GET API omits `mode`, the provider preserves the prior chart `time_range.mode` from configuration or state " +
				"(same pattern as REQ-009 on the dashboard `time_range`).",
			Optional: true,
			Validators: []validator.String{
				stringvalidator.OneOf("absolute", "relative"),
			},
		},
	}
}

func lensChartDrilldownListItemAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"dashboard_drilldown": schema.SingleNestedAttribute{
			MarkdownDescription: "Navigate to another dashboard using current filters/time range.",
			Optional:            true,
			Attributes: map[string]schema.Attribute{
				"dashboard_id": schema.StringAttribute{
					MarkdownDescription: "Target dashboard id.",
					Required:            true,
				},
				"label": schema.StringAttribute{
					MarkdownDescription: "Human-readable drilldown label.",
					Required:            true,
				},
				"trigger": schema.StringAttribute{
					MarkdownDescription: "**Computed** — Kibana fixes this to `on_apply_filter`; reflected in state after apply. Do not set in configuration.",
					Computed:            true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"use_filters": schema.BoolAttribute{
					MarkdownDescription: "When true, forwards filter context.",
					Optional:            true,
					Computed:            true,
					Default:             booldefault.StaticBool(true),
				},
				"use_time_range": schema.BoolAttribute{
					MarkdownDescription: "When true, forwards the time range.",
					Optional:            true,
					Computed:            true,
					Default:             booldefault.StaticBool(true),
				},
				"open_in_new_tab": schema.BoolAttribute{
					MarkdownDescription: "When true, opens the target dashboard in a new browser tab.",
					Optional:            true,
					Computed:            true,
					Default:             booldefault.StaticBool(false),
				},
			},
			Validators: []validator.Object{
				validators.ForbiddenIfDrilldownVariantSiblingNestedPresent(exprDrilldownDiscoverDrilldown),
				validators.ForbiddenIfDrilldownVariantSiblingNestedPresent(exprDrilldownURLDrilldown),
			},
		},
		"discover_drilldown": schema.SingleNestedAttribute{
			MarkdownDescription: "Open Discover with contextual filters.",
			Optional:            true,
			Attributes: map[string]schema.Attribute{
				"label": schema.StringAttribute{
					MarkdownDescription: "Human-readable drilldown label.",
					Required:            true,
				},
				"trigger": schema.StringAttribute{
					MarkdownDescription: "**Computed** — Kibana fixes this to `on_apply_filter`; reflected in state after apply. Do not set in configuration.",
					Computed:            true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"open_in_new_tab": schema.BoolAttribute{
					MarkdownDescription: "When true, opens Discover in a new browser tab.",
					Optional:            true,
					Computed:            true,
					Default:             booldefault.StaticBool(true),
				},
			},
			Validators: []validator.Object{
				validators.ForbiddenIfDrilldownVariantSiblingNestedPresent(exprDrilldownDashboardDrilldown),
				validators.ForbiddenIfDrilldownVariantSiblingNestedPresent(exprDrilldownURLDrilldown),
			},
		},
		"url_drilldown": schema.SingleNestedAttribute{
			MarkdownDescription: "Open a URL drilldown configured with explicit trigger semantics.",
			Optional:            true,
			Attributes: map[string]schema.Attribute{
				"url": schema.StringAttribute{
					MarkdownDescription: "Destination URL.",
					Required:            true,
				},
				"label": schema.StringAttribute{
					MarkdownDescription: "Human-readable drilldown label.",
					Required:            true,
				},
				"trigger": schema.StringAttribute{
					MarkdownDescription: "Trigger that fires this drilldown.",
					Required:            true,
					Validators: []validator.String{
						stringvalidator.OneOf(
							"on_click_row",
							"on_click_value",
							"on_open_panel_menu",
							"on_select_range",
						),
					},
				},
				"encode_url": schema.BoolAttribute{
					MarkdownDescription: "When true, encodes interpolated URL parameters.",
					Optional:            true,
					Computed:            true,
					Default:             booldefault.StaticBool(true),
				},
				"open_in_new_tab": schema.BoolAttribute{
					MarkdownDescription: "When true, opens the URL in a new browser tab.",
					Optional:            true,
					Computed:            true,
					Default:             booldefault.StaticBool(true),
				},
			},
			Validators: []validator.Object{
				validators.ForbiddenIfDrilldownVariantSiblingNestedPresent(exprDrilldownDashboardDrilldown),
				validators.ForbiddenIfDrilldownVariantSiblingNestedPresent(exprDrilldownDiscoverDrilldown),
			},
		},
	}
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
			Computed:            true,
		},
		"ignore_global_filters": schema.BoolAttribute{
			MarkdownDescription: "If true, ignore global filters when fetching data for this chart. Default is false.",
			Optional:            true,
			Computed:            true,
		},
		"filters": schema.ListNestedAttribute{
			MarkdownDescription: "Additional filters to apply to the chart data (maximum 100).",
			Optional:            true,
			NestedObject:        getChartFilter(),
		},
	}
}

// getPartitionChartBaseSchema returns base attributes shared by partition charts (treemap, mosaic).
// includePresentation merges REQ-037 chart-root attributes for typed `vis` panels only (`lens-dashboard-app` by_value passes false).
func getPartitionChartBaseSchema(includePresentation bool) map[string]schema.Attribute {
	attrs := lensChartBaseAttributes()
	attrs["data_source_json"] = schema.StringAttribute{
		MarkdownDescription: "Dataset configuration as JSON. For non-ES|QL, this specifies the data view or index; for ES|QL, this specifies the ES|QL query dataset.",
		CustomType:          jsontypes.NormalizedType{},
		Required:            true,
	}
	attrs["query"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Query configuration for filtering data. Required for non-ES|QL partition charts.",
		Optional:            true,
		Attributes:          getFilterSimple(),
	}
	if includePresentation {
		maps.Copy(attrs, lensChartPresentationAttributes())
	}
	return attrs
}

// getWaffleSchema returns schema for waffle (grid) Lens chart configuration.
func getWaffleSchema(includePresentation bool) map[string]schema.Attribute {
	attrs := getPartitionChartBaseSchema(includePresentation)
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

// getTreemapSchema returns the schema for treemap chart configuration.
func getTreemapSchema(includePresentation bool) map[string]schema.Attribute {
	base := getPartitionChartBaseSchema(includePresentation)
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

// getMosaicSchema returns the schema for mosaic chart configuration.
func getMosaicSchema(includePresentation bool) map[string]schema.Attribute {
	base := getPartitionChartBaseSchema(includePresentation)
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

// getRegionMapSchema returns the schema for region map chart configuration.
func getRegionMapSchema(includePresentation bool) map[string]schema.Attribute {
	attrs := lensChartBaseAttributes()
	attrs["data_source_json"] = schema.StringAttribute{
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
	if includePresentation {
		maps.Copy(attrs, lensChartPresentationAttributes())
	}
	return attrs
}

// getLegacyMetricSchema returns the schema for legacy metric chart configuration.
func getLegacyMetricSchema(includePresentation bool) map[string]schema.Attribute {
	attrs := lensChartBaseAttributes()
	attrs["data_source_json"] = schema.StringAttribute{
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
	if includePresentation {
		maps.Copy(attrs, lensChartPresentationAttributes())
	}
	return attrs
}

// getGaugeSchema returns the schema for gauge chart configuration.
func getGaugeSchema(includePresentation bool) map[string]schema.Attribute {
	attrs := lensChartBaseAttributes()
	attrs["data_source_json"] = schema.StringAttribute{
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
	if includePresentation {
		maps.Copy(attrs, lensChartPresentationAttributes())
	}
	return attrs
}

// getMetricChart returns the schema for metric chart configuration.
func getMetricChart(includePresentation bool) map[string]schema.Attribute {
	attrs := lensChartBaseAttributes()
	attrs["data_source_json"] = schema.StringAttribute{
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
					CustomType:          customtypes.NewJSONWithDefaultsType(populateMetricChartMetricDefaults),
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
	if includePresentation {
		maps.Copy(attrs, lensChartPresentationAttributes())
	}
	return attrs
}

// populatePieChartMetricDefaults populates default values for pie chart metric configuration
func populatePieChartMetricDefaults(model map[string]any) map[string]any {
	if model == nil {
		return model
	}

	if _, exists := model["empty_as_null"]; !exists {
		model["empty_as_null"] = false
	}
	if _, exists := model["color"]; !exists {
		model["color"] = map[string]any{"type": "auto"}
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

// getPieChart returns the schema for pie chart configuration.
func getPieChart(includePresentation bool) map[string]schema.Attribute {
	attrs := map[string]schema.Attribute{
		"title": schema.StringAttribute{
			MarkdownDescription: "The title of the chart displayed in the panel.",
			Optional:            true,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "The description of the chart.",
			Optional:            true,
		},
		"data_source_json": schema.StringAttribute{
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
	if includePresentation {
		maps.Copy(attrs, lensChartPresentationAttributes())
	}
	return attrs
}

// getSyntheticsMonitorsSchema returns the schema for the synthetics_monitors_config block.
// All fields are optional — the block itself may be omitted for a bare panel.
func getSyntheticsMonitorsSchema() map[string]schema.Attribute {
	filterItemSchema := schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"label": schema.StringAttribute{
				MarkdownDescription: "Display label for the filter option.",
				Required:            true,
			},
			"value": schema.StringAttribute{
				MarkdownDescription: "Value for the filter option.",
				Required:            true,
			},
		},
	}
	return map[string]schema.Attribute{
		"title": schema.StringAttribute{
			MarkdownDescription: "Display title shown in the panel header.",
			Optional:            true,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "Descriptive text for the panel.",
			Optional:            true,
		},
		"hide_title": schema.BoolAttribute{
			MarkdownDescription: "When true, suppresses the panel title in the dashboard.",
			Optional:            true,
		},
		"hide_border": schema.BoolAttribute{
			MarkdownDescription: "When true, suppresses the panel border in the dashboard.",
			Optional:            true,
		},
		"view": schema.StringAttribute{
			MarkdownDescription: "View mode for the panel. Valid values are `cardView` and `compactView`.",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("cardView", "compactView"),
			},
		},
		"filters": schema.SingleNestedAttribute{
			MarkdownDescription: "Optional filter configuration for the Synthetics monitors panel. Omit to show all monitors.",
			Optional:            true,
			Attributes: map[string]schema.Attribute{
				"projects": schema.ListNestedAttribute{
					MarkdownDescription: "Filter by project. Each entry has a `label` (display name) and a `value` (project ID).",
					Optional:            true,
					NestedObject:        filterItemSchema,
				},
				"tags": schema.ListNestedAttribute{
					MarkdownDescription: "Filter by tags. Each entry has a `label` (display name) and a `value` (tag).",
					Optional:            true,
					NestedObject:        filterItemSchema,
				},
				"monitor_ids": schema.ListNestedAttribute{
					MarkdownDescription: "Filter by monitor IDs. Each entry has a `label` (display name) and a `value` (monitor ID). The Kibana API accepts up to 5000 items.",
					Optional:            true,
					NestedObject:        filterItemSchema,
				},
				"locations": schema.ListNestedAttribute{
					MarkdownDescription: "Filter by monitor locations. Each entry has a `label` (display name) and a `value` (location ID).",
					Optional:            true,
					NestedObject:        filterItemSchema,
				},
				"monitor_types": schema.ListNestedAttribute{
					MarkdownDescription: "Filter by monitor types. Each entry has a `label` (display name) and a `value` (monitor type, e.g. `browser`, `http`, `tcp`, `icmp`).",
					Optional:            true,
					NestedObject:        filterItemSchema,
				},
			},
		},
	}
}
