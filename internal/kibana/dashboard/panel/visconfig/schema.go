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

package visconfig

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Allowed values for url.trigger on structured Lens/Vis drilldown items (matches kbapi enums).
var structuredDrilldownURLTriggerEnum = []string{
	"on_click_row",
	"on_click_value",
	"on_open_panel_menu",
	"on_select_range",
}

func structuredDrilldownsAttribute() schema.Attribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "Structured dashboard, Discover, or URL drilldown entries for by-reference panels — " +
			"shared by `vis_config.by_reference` (`vis` panels) and `lens_dashboard_app_config.by_reference` (`lens-dashboard-app` panels). " +
			"Each element must contain exactly one of `dashboard`, `discover`, or `url`; " +
			"the provider sets API `type` and (for dashboard/discover) `trigger` automatically.",
		Optional: true,
		NestedObject: schema.NestedAttributeObject{
			Validators: []validator.Object{
				panelkit.DrilldownItemModeValidator{},
			},
			Attributes: map[string]schema.Attribute{
				"dashboard": schema.SingleNestedAttribute{
					MarkdownDescription: "Open another dashboard (`dashboard_drilldown`). `dashboard_id` and `label` are required; " +
						"remaining fields mirror optional API knobs.",
					Optional: true,
					Attributes: map[string]schema.Attribute{
						"dashboard_id": schema.StringAttribute{
							MarkdownDescription: "Target dashboard ID.",
							Required:            true,
						},
						"label": schema.StringAttribute{
							MarkdownDescription: "Display label.",
							Required:            true,
						},
						"use_filters": schema.BoolAttribute{
							MarkdownDescription: "Pass filters to the target dashboard when set.",
							Optional:            true,
						},
						"use_time_range": schema.BoolAttribute{
							MarkdownDescription: "Pass the current time range to the target dashboard when set.",
							Optional:            true,
						},
						"open_in_new_tab": schema.BoolAttribute{
							MarkdownDescription: "Open in a new browser tab when set.",
							Optional:            true,
						},
					},
				},
				"discover": schema.SingleNestedAttribute{
					MarkdownDescription: "Open in Discover (`discover_drilldown`). Requires `label`.",
					Optional:            true,
					Attributes: map[string]schema.Attribute{
						"label": schema.StringAttribute{
							MarkdownDescription: "Display label.",
							Required:            true,
						},
						"open_in_new_tab": schema.BoolAttribute{
							MarkdownDescription: "Open in a new browser tab when set.",
							Optional:            true,
						},
					},
				},
				"url": schema.SingleNestedAttribute{
					MarkdownDescription: "Custom URL drilldown (`url_drilldown`). Requires `url`, `label`, and `trigger` " +
						"(one of `on_click_row`, `on_click_value`, `on_open_panel_menu`, `on_select_range`). " +
						"The Kibana dashboard API rejects URL drilldowns without `trigger`.",
					Optional: true,
					Attributes: map[string]schema.Attribute{
						"url": schema.StringAttribute{
							MarkdownDescription: "URL template with variables documented in Kibana URL drilldown " +
								"documentation.",
							Required: true,
						},
						"label": schema.StringAttribute{
							MarkdownDescription: "Display label.",
							Required:            true,
						},
						"trigger": schema.StringAttribute{
							MarkdownDescription: "Trigger that activates the drilldown. Required; the Kibana dashboard API rejects URL drilldowns when this field is omitted.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOf(structuredDrilldownURLTriggerEnum...),
							},
						},
						"encode_url": schema.BoolAttribute{
							MarkdownDescription: "Escape the URL via percent-encoding when set.",
							Optional:            true,
						},
						"open_in_new_tab": schema.BoolAttribute{
							MarkdownDescription: "Open in a new browser tab when set.",
							Optional:            true,
						},
					},
				},
			},
		},
	}
}

func lensByReferenceAttributes() map[string]schema.Attribute {
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
		"drilldowns": structuredDrilldownsAttribute(),
		"time_range": schema.SingleNestedAttribute{
			MarkdownDescription: "Required time range for the by-reference panel config " +
				"(used by both `lens_dashboard_app_config.by_reference` and `vis_config.by_reference`).",
			Required:   true,
			Attributes: panelkit.TimeRangeAttributes(),
		},
	}
}

func byValueAttributes() map[string]schema.Attribute {
	out := make(map[string]schema.Attribute)
	for _, c := range lenscommon.All() {
		out[chartBlockKeyOrPanic(c.VizType())] = c.SchemaAttribute()
	}
	return out
}

func visByValueChartAttrNames() []string {
	converters := lenscommon.All()
	out := make([]string, 0, len(converters))
	for _, c := range converters {
		out = append(out, chartBlockKeyOrPanic(c.VizType()))
	}
	return out
}

func chartBlockKeyOrPanic(vizType string) string {
	key := lenscommon.TerraformChartBlockKey(vizType)
	if key == "" {
		panic("visconfig: missing terraform chart block key for VizType " + vizType)
	}
	return key
}

func innerSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"by_value": schema.SingleNestedAttribute{
			MarkdownDescription: "Inline by-value Lens visualization configuration for `type = \"vis\"` panels (`vis_config`). " +
				"Exactly one typed chart kind must be set (no raw JSON here — use panel-level `config_json` for that).",
			Optional:   true,
			Attributes: byValueAttributes(),
			Validators: []validator.Object{
				panelkit.ExactlyOneOfNestedAttrsValidator(panelkit.ExactlyOneOfNestedAttrsOpts{
					AttrNames:     visByValueChartAttrNames(),
					Summary:       "Invalid vis_config.by_value",
					MissingDetail: "Set exactly one supported typed Lens chart block inside `vis_config.by_value`.",
					TooManyDetail: "Set exactly one typed chart block inside `vis_config.by_value` (more than one by-value chart is set).",
					Description:   "Ensures exactly one supported typed Lens chart block is set inside `vis_config.by_value`.",
				}),
			},
		},
		"by_reference": schema.SingleNestedAttribute{
			MarkdownDescription: "By-reference `vis` configuration: structured `drilldowns`, `ref_id`, optional `references_json`, and required `time_range`.",
			Optional:            true,
			Attributes:          lensByReferenceAttributes(),
		},
	}
}

// SchemaAttribute returns the Terraform schema for the vis panel typed configuration block (`vis_config`).
func SchemaAttribute() schema.Attribute {
	return panelkit.PanelConfigBlock(panelkit.PanelConfigBlockOpts{
		Description: "Configuration for a `vis` panel (`type = \"vis\"`). " +
			"Typed alternative to panel-level `config_json`: set exactly one of `by_value` (exactly one of 12 Lens chart kinds) or `by_reference`. " +
			"With `by_reference`, use structured `drilldowns` and required `time_range` like `lens_dashboard_app_config.by_reference`.",
		BlockName:  "vis_config",
		PanelType:  panelType,
		Required:   false,
		Attributes: innerSchemaAttributes(),
		ExtraValidators: []validator.Object{
			panelkit.ExactlyOneOfNestedAttrsValidator(panelkit.ExactlyOneOfNestedAttrsOpts{
				AttrNames:     []string{"by_value", "by_reference"},
				Summary:       "Invalid vis_config",
				MissingDetail: "Exactly one of `by_value` or `by_reference` must be set inside `vis_config`.",
				TooManyDetail: "Exactly one of `by_value` or `by_reference` must be set inside `vis_config`, not both.",
				Description:   "Ensures exactly one of `by_value` or `by_reference` is set inside `vis_config`.",
			}),
		},
	})
}
