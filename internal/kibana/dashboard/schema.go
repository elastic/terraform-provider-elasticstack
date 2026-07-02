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
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/kbschema"
	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/validators"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
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
	panelTypeSloOverview             = "slo_overview"
)

// panelConfigNames returns the full set of panel-level typed config attribute names used for mutual-exclusion
// validators and documentation. Populated from handler registration via panelkit wiring.
func panelConfigNames() []string {
	return panelkit.TypedSiblingPanelConfigBlockNames()
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
			"space_id": kbschema.ResourceSpaceIDAttributeRequiresReplaceOnly(),
			attrDashboardID: schema.StringAttribute{
				MarkdownDescription: "Optional dashboard identifier. When set, create uses PUT upsert semantics; changing this value forces replacement. When omitted, Kibana assigns a UUID.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			attrTitle: schema.StringAttribute{
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
					attrValue: schema.Int64Attribute{
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
						attrTitle: schema.StringAttribute{
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
						attrPanelGrid: schema.SingleNestedAttribute{
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
		attrPanelType: schema.StringAttribute{
			MarkdownDescription: "The type of the panel (e.g. 'markdown', 'vis').",
			Required:            true,
		},
		attrPanelGrid: schema.SingleNestedAttribute{
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
				"Typed panel kinds such as `image`, `slo_alerts`, `discover_session`, `ml_anomaly_swimlane`, `ml_anomaly_charts`, and `ml_single_metric_viewer` use their dedicated blocks "+
				"(`image_config`, `slo_alerts_config`, `discover_session_config`, `ml_anomaly_swimlane_config`, `ml_anomaly_charts_config`, `ml_single_metric_viewer_config`), not panel-level `config_json`.",
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
			validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName(attrPanelType), []string{panelTypeVis, panelTypeMarkdown}, validators.AllowedIfOptions{}),
		},
	}
	return schema.NestedAttributeObject{
		Validators: []validator.Object{
			panelConfigValidator{},
		},
		Attributes: attrs,
	}
}

// getDashboardRootSavedFiltersNestedObject returns the nested object schema for one dashboard-level saved filter.
// Uses jsontypes.NormalizedType for API-shaped filter JSON (consistent with chart-level saved filters).
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
