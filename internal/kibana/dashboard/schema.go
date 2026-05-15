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
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/validators"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	dashboardValueAuto    = "auto"
	dashboardValueAverage = "average"
	panelTypeImage        = "image"
	panelTypeMarkdown     = "markdown"
	// panelTypeVis is Kibana's dashboard panel API discriminator for Lens-backed visualizations; Terraform typed configuration uses vis_config.
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
	panelTypeSloOverview             = "slo_overview"
)

// panelConfigNames returns the full set of panel-level typed config attribute names used for mutual-exclusion
// validators and documentation. Populated from handler registration via panelkit wiring.
func panelConfigNames() []string {
	return panelkit.TypedSiblingPanelConfigBlockNames()
}

// populateLensMetricDefaults populates default values for Lens metric configuration (shared across XY, metric, pie, treemap, datatable, etc.).
func populateLensMetricDefaults(model map[string]any) map[string]any {
	return lenscommon.PopulateLensMetricDefaults(model)
}

func populateMetricChartMetricDefaults(model map[string]any) map[string]any {
	return lenscommon.PopulateMetricChartMetricDefaults(model)
}

// populatePartitionGroupByDefaults populates default values for partition chart group_by/group_breakdown_by configurations.
// Used by treemap and mosaic. Kibana may add default fields (e.g. rank_by, size) on read, so we normalize both sides.
func populatePartitionGroupByDefaults(model []map[string]any) []map[string]any {
	return lenscommon.PopulatePartitionGroupByDefaults(model)
}

// populatePartitionMetricsDefaults populates default values for partition chart metrics.
// Used by treemap and mosaic. Mirrors the defaulting behavior used by other Lens metric operations.
func populatePartitionMetricsDefaults(model []map[string]any) []map[string]any {
	return lenscommon.PopulatePartitionMetricsDefaults(model)
}

// populateLegacyMetricMetricDefaults populates default values for legacy metric operations
func populateLegacyMetricMetricDefaults(model map[string]any) map[string]any {
	return lenscommon.PopulateLegacyMetricMetricDefaults(model)
}

// populateGaugeMetricDefaults populates default values for gauge metric configuration
func populateGaugeMetricDefaults(model map[string]any) map[string]any {
	return lenscommon.PopulateGaugeMetricDefaults(model)
}

// populateRegionMapMetricDefaults populates default values for region map metric configuration
func populateRegionMapMetricDefaults(model map[string]any) map[string]any {
	return lenscommon.PopulateRegionMapMetricDefaults(model)
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
	names := panelConfigNames()
	attrs := map[string]schema.Attribute{
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
	}
	for _, h := range AllHandlers() {
		attrs[h.PanelType()+"_config"] = h.SchemaAttribute()
	}
	attrs["config_json"] = schema.StringAttribute{
		MarkdownDescription: panelkit.PanelConfigDescription(
			"The configuration of the panel as a JSON string. "+
				"Practitioner-authored panel-level `config_json` is valid only when `type` is `markdown` or `vis`. "+
				"Typed panel kinds such as `lens-dashboard-app`, `image`, `slo_alerts`, and `discover_session` use their dedicated blocks "+
				"(`lens_dashboard_app_config`, `image_config`, `slo_alerts_config`, `discover_session_config`), not panel-level `config_json`.",
			"config_json",
			names,
		),
		CustomType: customtypes.NewJSONWithDefaultsType(populatePanelConfigJSONDefaults),
		Optional:   true,
		Computed:   true,
		Validators: []validator.String{
			stringvalidator.ConflictsWith(
				panelkit.SiblingTypedPanelConfigConflictPathsExcept("config_json", names)...,
			),
			validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelTypeVis, panelTypeMarkdown}),
		},
	}
	return schema.NestedAttributeObject{
		Validators: []validator.Object{
			panelConfigValidator{},
		},
		Attributes: attrs,
	}
}

func getFilterSimple() map[string]schema.Attribute {
	return lenscommon.LensChartFilterSimpleAttributes()
}

// getChartFilter returns the schema for a single chart-level filter (API-shaped JSON).
func getChartFilter() schema.NestedAttributeObject {
	return lenscommon.LensChartFilterNestedObject()
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
	attrs["metric_json"] = schema.StringAttribute{
		MarkdownDescription: "Metric configuration as JSON. For non-ES|QL, this can be a field metric, pipeline metric, or formula. For ES|QL, this is the metric column/operation/color configuration.",
		CustomType:          customtypes.NewJSONWithDefaultsType(populateTagcloudMetricDefaults),
		Required:            true,
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

// lensChartPresentationAttributes returns optional chart-root presentation fields shared by all typed Lens chart blocks:
// `time_range` (inherits dashboard-level when null — see REQ-038), `hide_title`, `hide_border`, `references_json`, and `drilldowns`.
func lensChartPresentationAttributes() map[string]schema.Attribute {
	return lenscommon.LensChartPresentationAttributes()
}

// lensChartDrilldownListItemAttributes forwards to lenscommon for tests that assert nested drilldown schema wiring.
func lensChartDrilldownListItemAttributes() map[string]schema.Attribute {
	return lenscommon.LensChartDrilldownListItemAttributes()
}

// lensChartBaseAttributes returns attributes shared by most Lens chart panels:
// title, description, sampling, ignore_global_filters, and filters.
func lensChartBaseAttributes() map[string]schema.Attribute {
	return lenscommon.LensChartBaseAttributes()
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
		Attributes:          getPartitionValueDisplaySchema(),
	}
	attrs["metrics"] = schema.ListNestedAttribute{
		MarkdownDescription: "Metric configurations for non-ES|QL waffles (minimum 1). Each `config_json` is a JSON object (e.g. count, sum, or formula) matching the Kibana Lens waffle schema.",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"config_json": schema.StringAttribute{
					MarkdownDescription: "Metric operation as JSON.",
					CustomType:          customtypes.NewJSONWithDefaultsType(populatePieChartMetricDefaults),
					Required:            true,
				},
			},
		},
	}
	attrs["group_by"] = schema.ListNestedAttribute{
		MarkdownDescription: "Breakdown dimensions for non-ES|QL waffles. Each `config_json` is a JSON object (terms, date_histogram, etc.) matching the Kibana Lens waffle schema.",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"config_json": schema.StringAttribute{
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
		NestedObject:        getPartitionESQLMetricSchema(),
	}
	attrs["esql_group_by"] = schema.ListNestedAttribute{
		MarkdownDescription: "Breakdown columns for ES|QL waffles. Mutually exclusive with `group_by`.",
		Optional:            true,
		NestedObject:        getPartitionESQLGroupBySchema(),
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

// getPartitionESQLMetricSchema returns the shared ES|QL metric schema used by waffle,
// treemap, and mosaic.
func getPartitionESQLMetricSchema() schema.NestedAttributeObject {
	return lenscommon.PartitionESQLMetricNestedObject()
}

// getPartitionESQLGroupBySchema returns the shared ES|QL group-by schema used by waffle,
// treemap, and mosaic.
func getPartitionESQLGroupBySchema() schema.NestedAttributeObject {
	return lenscommon.PartitionESQLGroupByNestedObject()
}

// getMosaicESQLMetricSchema returns the ES|QL metric schema for mosaic.
// Mosaic ES|QL uses a single metric without color, so this omits the color
// block present in waffle/treemap.
func getMosaicESQLMetricSchema() schema.NestedAttributeObject {
	return lenscommon.MosaicESQLMetricNestedObject()
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
			Optional:   true,
			Validators: []validator.String{
				stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("esql_group_by")),
			},
		},
		"metrics_json": schema.StringAttribute{
			MarkdownDescription: "Array of metric configurations as JSON (minimum 1). " +
				"For non-ES|QL, each item can be a field metric, pipeline metric, or formula; " +
				"for ES|QL, each item is the column/operation/color/format configuration.",
			CustomType: customtypes.NewJSONWithDefaultsType(populatePartitionMetricsDefaults),
			Optional:   true,
			Validators: []validator.String{
				stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("esql_metrics")),
			},
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
		"esql_metrics": schema.ListNestedAttribute{
			MarkdownDescription: "Metric columns for ES|QL treemaps. Mutually exclusive with `metrics_json`.",
			Optional:            true,
			NestedObject:        getPartitionESQLMetricSchema(),
			Validators: []validator.List{
				listvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("metrics_json")),
			},
		},
		"esql_group_by": schema.ListNestedAttribute{
			MarkdownDescription: "Breakdown columns for ES|QL treemaps. Mutually exclusive with `group_by_json`.",
			Optional:            true,
			NestedObject:        getPartitionESQLGroupBySchema(),
			Validators: []validator.List{
				listvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("group_by_json")),
			},
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
			Optional:   true,
			Validators: []validator.String{
				stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("esql_group_by")),
			},
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
			Optional:   true,
			Validators: []validator.String{
				stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("esql_metrics")),
			},
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
		"esql_metrics": schema.ListNestedAttribute{
			MarkdownDescription: "Metric columns for ES|QL mosaics (exactly 1). Mutually exclusive with `metrics_json`.",
			Optional:            true,
			NestedObject:        getMosaicESQLMetricSchema(),
			Validators: []validator.List{
				listvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("metrics_json")),
			},
		},
		"esql_group_by": schema.ListNestedAttribute{
			MarkdownDescription: "Breakdown columns for ES|QL mosaics. Mutually exclusive with `group_by_json`.",
			Optional:            true,
			NestedObject:        getPartitionESQLGroupBySchema(),
			Validators: []validator.List{
				listvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("group_by_json")),
			},
		},
	}
	maps.Copy(base, mosaicSpecific)
	return base
}

func getPartitionLegendSchema() map[string]schema.Attribute {
	return lenscommon.PartitionLegendSchemaAttributes()
}

func getPartitionValueDisplaySchema() map[string]schema.Attribute {
	return lenscommon.PartitionValueDisplaySchemaAttributes()
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
		MarkdownDescription: "Query configuration for filtering data. Required for non-ES|QL gauges; omit for ES|QL mode.",
		Optional:            true,
		Attributes:          getFilterSimple(),
	}
	attrs["metric_json"] = schema.StringAttribute{
		MarkdownDescription: gaugeMetricDescription + " Required for non-ES|QL gauges; mutually exclusive with `esql_metric`.",
		CustomType:          customtypes.NewJSONWithDefaultsType(populateGaugeMetricDefaults),
		Optional:            true,
		Validators: []validator.String{
			stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("esql_metric")),
		},
	}
	attrs["esql_metric"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Typed metric column for ES|QL gauges. Mutually exclusive with `metric_json`.",
		Optional:            true,
		Attributes:          getGaugeESQLMetricSchema(),
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

// getGaugeESQLMetricSchema returns the typed attribute schema for the
// `esql_metric` block on gauge ES|QL charts.
func getGaugeESQLMetricSchema() map[string]schema.Attribute {
	gaugeRefSchema := func(desc string) schema.SingleNestedAttribute {
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
	return map[string]schema.Attribute{
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
		"goal": gaugeRefSchema("Goal column reference."),
		"max":  gaugeRefSchema("Max column reference."),
		"min":  gaugeRefSchema("Min column reference."),
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
	}
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
	return lenscommon.PopulatePieChartMetricDefaults(model)
}

// populateLensGroupByDefaults populates default values for Lens dimension/group-by configuration (shared across pie, treemap, datatable, etc.).
func populateLensGroupByDefaults(model map[string]any) map[string]any {
	return lenscommon.PopulateLensGroupByDefaults(model)
}

// pieChartLegendDefaultObject is the schema default when the legend block is omitted from config,
// aligned with typical Kibana read-back so apply and refresh stay consistent.
func pieChartLegendDefaultObject() types.Object {
	return types.ObjectValueMust(
		map[string]attr.Type{
			"nested":               types.BoolType,
			"size":                 types.StringType,
			"truncate_after_lines": types.Int64Type,
			"visible":              types.StringType,
		},
		map[string]attr.Value{
			"nested":               types.BoolNull(),
			"size":                 types.StringValue("auto"),
			"truncate_after_lines": types.Int64Null(),
			"visible":              types.StringValue("auto"),
		},
	)
}

// getPieChart returns the schema for pie chart configuration.
func getPieChart(includePresentation bool) map[string]schema.Attribute {
	attrs := lensChartBaseAttributes()
	attrs["data_source_json"] = schema.StringAttribute{
		MarkdownDescription: "Dataset configuration as JSON. For standard layers, this specifies the data view and query.",
		CustomType:          jsontypes.NormalizedType{},
		Required:            true,
	}
	attrs["donut_hole"] = schema.StringAttribute{
		MarkdownDescription: "Donut hole size: none (pie), s, m, or l.",
		Optional:            true,
		Validators: []validator.String{
			stringvalidator.OneOf("none", "s", "m", "l"),
		},
	}
	attrs["label_position"] = schema.StringAttribute{
		MarkdownDescription: "Position of slice labels: hidden, inside, or outside.",
		Optional:            true,
		Validators: []validator.String{
			stringvalidator.OneOf("hidden", "inside", "outside"),
		},
	}
	attrs["legend"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Optional legend configuration for the pie chart. " +
			"Same shape as treemap and mosaic legends; Terraform `visible` maps to API `visibility`. " +
			"When omitted, the schema default matches typical Kibana legend defaults (size and visibility " +
			"`auto`) so apply/read stay consistent.",
		Optional:   true,
		Computed:   true,
		Default:    objectdefault.StaticValue(pieChartLegendDefaultObject()),
		Attributes: getPartitionLegendSchema(),
	}
	attrs["query"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Query configuration for filtering data.",
		Optional:            true,
		Attributes:          getFilterSimple(),
	}
	attrs["filters"] = schema.ListNestedAttribute{
		MarkdownDescription: "Additional filters to apply to the chart data (maximum 100).",
		Optional:            true,
		NestedObject:        getChartFilter(),
	}
	attrs["metrics"] = schema.ListNestedAttribute{
		MarkdownDescription: "Array of metric configurations (minimum 1).",
		Required:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"config_json": schema.StringAttribute{
					MarkdownDescription: "Metric configuration as JSON.",
					CustomType:          customtypes.NewJSONWithDefaultsType(populatePieChartMetricDefaults),
					Required:            true,
				},
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
		},
	}
	attrs["group_by"] = schema.ListNestedAttribute{
		MarkdownDescription: "Array of breakdown dimensions (minimum 1).",
		Optional:            true,
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
		},
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"config_json": schema.StringAttribute{
					MarkdownDescription: "Group by configuration as JSON.",
					CustomType:          customtypes.NewJSONWithDefaultsType(populateLensGroupByDefaults),
					Required:            true,
				},
			},
		},
	}
	if includePresentation {
		maps.Copy(attrs, lensChartPresentationAttributes())
	}
	return attrs
}
