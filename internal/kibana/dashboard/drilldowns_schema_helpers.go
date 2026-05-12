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
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Allowed values for `url.trigger` on structured Lens/Vis drilldown items (matches kbapi enums).
var structuredDrilldownURLTriggerEnum = []string{
	"on_click_row",
	"on_click_value",
	"on_open_panel_menu",
	"on_select_range",
}

func getStructuredDrilldownsAttribute() schema.Attribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "Structured dashboard, Discover, or URL drilldown entries for by-reference panels — " +
			"shared by `viz_config.by_reference` (`vis` panels) and `lens_dashboard_app_config.by_reference` (`lens-dashboard-app` panels). " +
			"Each element must contain exactly one of `dashboard`, `discover`, or `url`; " +
			"the provider sets API `type` and (for dashboard/discover) `trigger` automatically.",
		Optional: true,
		NestedObject: schema.NestedAttributeObject{
			Validators: []validator.Object{
				drilldownItemModeValidator{},
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
