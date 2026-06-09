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

package panelkit

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
)

// StructuredDrilldownURLTriggerEnum lists allowed values for url.trigger on structured
// Lens/Vis drilldown items (matches kbapi enums).
var StructuredDrilldownURLTriggerEnum = []string{
	"on_click_row",
	"on_click_value",
	"on_open_panel_menu",
	"on_select_range",
}

// Default URL drilldown element descriptions (typed panels that fix trigger/type in the model layer).
const (
	urlDrilldownDefaultURLDescription          = "Templated URL for the drilldown."
	urlDrilldownDefaultLabelDescription        = "Display label shown in the drilldown menu."
	urlDrilldownDefaultEncodeURLDescription    = "When true, the URL is percent-encoded. Omit to use the API default."
	urlDrilldownDefaultOpenInNewTabDescription = "When true, the URL opens in a new browser tab. Omit to use the API default."
)

// Terraform schema attribute keys reused across URL and dashboard drilldown
// blocks. Hosting these as constants avoids drift between mutually exclusive
// sibling validators and the attribute definitions themselves.
const (
	attrURL                = "url"
	attrLabel              = "label"
	attrTrigger            = "trigger"
	attrEncodeURL          = "encode_url"
	attrOpenInNewTab       = "open_in_new_tab"
	attrDashboardID        = "dashboard_id"
	attrDashboardDrilldown = "dashboard_drilldown"
	attrURLDrilldown       = "url_drilldown"
)

// Reused markdown descriptions for drilldown attributes shared between the
// dashboard and URL drilldown variants on image and structured panels.
const (
	descDisplayLabel = "Display label."
	descOpenInNewTab = "Open in a new browser tab when set."
)

// URLDrilldownOptions overrides MarkdownDescription on URL drilldown nested object attributes.
// Trigger and type are not schema fields (fixed in the model layer).
// Empty string in a field means use the default for that attribute (see Default URL drilldown constants).
type URLDrilldownOptions struct {
	URLMarkdownDescription          string
	LabelMarkdownDescription        string
	EncodeURLMarkdownDescription    string
	OpenInNewTabMarkdownDescription string
}

// URLDrilldownSchema returns the NestedAttributeObject used inside a ListNestedAttribute `drilldowns`.
func URLDrilldownSchema(opts URLDrilldownOptions) schema.NestedAttributeObject {
	urlDesc := typeutils.NonZero(opts.URLMarkdownDescription, urlDrilldownDefaultURLDescription)
	labelDesc := typeutils.NonZero(opts.LabelMarkdownDescription, urlDrilldownDefaultLabelDescription)
	encodeDesc := typeutils.NonZero(opts.EncodeURLMarkdownDescription, urlDrilldownDefaultEncodeURLDescription)
	openDesc := typeutils.NonZero(opts.OpenInNewTabMarkdownDescription, urlDrilldownDefaultOpenInNewTabDescription)
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			attrURL: schema.StringAttribute{
				MarkdownDescription: urlDesc,
				Required:            true,
			},
			attrLabel: schema.StringAttribute{
				MarkdownDescription: labelDesc,
				Required:            true,
			},
			attrEncodeURL: schema.BoolAttribute{
				MarkdownDescription: encodeDesc,
				Optional:            true,
			},
			attrOpenInNewTab: schema.BoolAttribute{
				MarkdownDescription: openDesc,
				Optional:            true,
			},
		},
	}
}

// ImageDashboardDrilldownDescriptions lifts optional per-attribute MarkdownDescription overrides for `dashboard_drilldown`.
type ImageDashboardDrilldownDescriptions struct {
	DashboardID, Label, Trigger, UseFilters, UseTimeRange, OpenInNewTab string
}

// ImageDrilldownOptions configures MarkdownDescription strings for image panel `drilldowns` entries
// (dashboard vs URL variants). Dashboard attribute descriptions mirror `schema_image_panel.go`.
type ImageDrilldownOptions struct {
	DashboardMarkdownDescriptions ImageDashboardDrilldownDescriptions
	URLDrilldownMarkdown          ImageURLDrilldownOptions
}

// ImageURLDrilldownOptions configures the URL drilldown sub-block inside an image drilldown entry (includes trigger).
type ImageURLDrilldownOptions struct {
	URLMarkdownDescription          string
	LabelMarkdownDescription        string
	EncodeURLMarkdownDescription    string
	OpenInNewTabMarkdownDescription string
	TriggerMarkdownDescription      string
}

const (
	defaultImageURLDrilldownTriggerDescription = "When this drilldown runs. Allowed values: `on_click_image`, `on_open_panel_menu` (Kibana image panel URL drilldown triggers)."
)

// ImageDrilldownSchema returns the NestedAttributeObject for one image panel drilldown list element:
// mutually exclusive `dashboard_drilldown` or `url_drilldown`.
func ImageDrilldownSchema(opts ImageDrilldownOptions) schema.NestedAttributeObject {
	dd := opts.DashboardMarkdownDescriptions
	urlOpts := opts.URLDrilldownMarkdown
	urlDesc := typeutils.NonZero(urlOpts.URLMarkdownDescription, urlDrilldownDefaultURLDescription)
	labelDescURL := typeutils.NonZero(urlOpts.LabelMarkdownDescription, urlDrilldownDefaultLabelDescription)
	triggerDesc := typeutils.NonZero(urlOpts.TriggerMarkdownDescription, defaultImageURLDrilldownTriggerDescription)
	encodeDescURL := typeutils.NonZero(urlOpts.EncodeURLMarkdownDescription, urlDrilldownDefaultEncodeURLDescription)
	openDescURL := typeutils.NonZero(urlOpts.OpenInNewTabMarkdownDescription, urlDrilldownDefaultOpenInNewTabDescription)

	dashboardTriggerDesc := typeutils.NonZero(dd.Trigger, "Dashboard drilldowns on image panels only support `on_click_image` (see Kibana `kbn-dashboard-panel-type-image` drilldown schema).")

	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			attrDashboardDrilldown: schema.SingleNestedAttribute{
				MarkdownDescription: "Open another dashboard when the image is clicked. Mutually exclusive with `url_drilldown` in the same entry.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					attrDashboardID: schema.StringAttribute{
						MarkdownDescription: typeutils.NonZero(dd.DashboardID, "Target dashboard saved object id."),
						Required:            true,
					},
					attrLabel: schema.StringAttribute{
						MarkdownDescription: typeutils.NonZero(dd.Label, "Label shown for this drilldown."),
						Required:            true,
					},
					attrTrigger: schema.StringAttribute{
						MarkdownDescription: dashboardTriggerDesc,
						Required:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("on_click_image"),
						},
					},
					"use_filters": schema.BoolAttribute{
						MarkdownDescription: typeutils.NonZero(dd.UseFilters, "When true, passes the current dashboard filters to the opened dashboard. Omit for API default (typically `false`)."),
						Optional:            true,
					},
					"use_time_range": schema.BoolAttribute{
						MarkdownDescription: typeutils.NonZero(dd.UseTimeRange, "When true, passes the current time range to the opened dashboard. Omit for API default (typically `false`)."),
						Optional:            true,
					},
					attrOpenInNewTab: schema.BoolAttribute{
						MarkdownDescription: typeutils.NonZero(dd.OpenInNewTab, "When true, opens the target dashboard in a new browser tab. Omit for API default (typically `false`)."),
						Optional:            true,
					},
				},
			},
			attrURLDrilldown: schema.SingleNestedAttribute{
				MarkdownDescription: "URL drilldown entry. Mutually exclusive with `dashboard_drilldown` in the same list element.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					attrURL: schema.StringAttribute{
						MarkdownDescription: urlDesc,
						Required:            true,
					},
					attrLabel: schema.StringAttribute{
						MarkdownDescription: labelDescURL,
						Required:            true,
					},
					attrTrigger: schema.StringAttribute{
						MarkdownDescription: triggerDesc,
						Required:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("on_click_image", "on_open_panel_menu"),
						},
					},
					attrEncodeURL: schema.BoolAttribute{
						MarkdownDescription: encodeDescURL,
						Optional:            true,
					},
					attrOpenInNewTab: schema.BoolAttribute{
						MarkdownDescription: openDescURL,
						Optional:            true,
					},
				},
			},
		},
		Validators: []validator.Object{
			ExactlyOneOfNestedAttrsValidator(ExactlyOneOfNestedAttrsOpts{
				AttrNames:     []string{attrDashboardDrilldown, attrURLDrilldown},
				Summary:       "Invalid drilldown entry",
				MissingDetail: "Exactly one of `dashboard_drilldown` or `url_drilldown` must be set.",
				TooManyDetail: "Exactly one of `dashboard_drilldown` or `url_drilldown` must be set, not both.",
				Description:   "Ensures exactly one of `dashboard_drilldown` or `url_drilldown` is set on each `drilldowns` entry.",
			}),
		},
	}
}

// StructuredDrilldownsAttribute returns the ListNestedAttribute used for by-reference
// `drilldowns` entries on `vis_config.by_reference` panels. Each list element must contain
// exactly one of `dashboard`, `discover`, or `url`.
func StructuredDrilldownsAttribute() schema.Attribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "Structured dashboard, Discover, or URL drilldown entries for by-reference panels — " +
			"shared by `vis_config.by_reference` (`vis` panels). " +
			"Each element must contain exactly one of `dashboard`, `discover`, or `url`; " +
			"the provider sets API `type` and (for dashboard/discover) `trigger` automatically.",
		Optional: true,
		NestedObject: schema.NestedAttributeObject{
			Validators: []validator.Object{
				DrilldownItemModeValidator{},
			},
			Attributes: map[string]schema.Attribute{
				"dashboard": schema.SingleNestedAttribute{
					MarkdownDescription: "Open another dashboard (`dashboard_drilldown`). `dashboard_id` and `label` are required; " +
						"remaining fields mirror optional API knobs.",
					Optional: true,
					Attributes: map[string]schema.Attribute{
						attrDashboardID: schema.StringAttribute{
							MarkdownDescription: "Target dashboard ID.",
							Required:            true,
						},
						attrLabel: schema.StringAttribute{
							MarkdownDescription: descDisplayLabel,
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
						attrOpenInNewTab: schema.BoolAttribute{
							MarkdownDescription: descOpenInNewTab,
							Optional:            true,
						},
					},
				},
				"discover": schema.SingleNestedAttribute{
					MarkdownDescription: "Open in Discover (`discover_drilldown`). Requires `label`.",
					Optional:            true,
					Attributes: map[string]schema.Attribute{
						attrLabel: schema.StringAttribute{
							MarkdownDescription: descDisplayLabel,
							Required:            true,
						},
						attrOpenInNewTab: schema.BoolAttribute{
							MarkdownDescription: descOpenInNewTab,
							Optional:            true,
						},
					},
				},
				attrURL: schema.SingleNestedAttribute{
					MarkdownDescription: "Custom URL drilldown (`url_drilldown`). Requires `url`, `label`, and `trigger` " +
						"(one of `on_click_row`, `on_click_value`, `on_open_panel_menu`, `on_select_range`). " +
						"The Kibana dashboard API rejects URL drilldowns without `trigger`.",
					Optional: true,
					Attributes: map[string]schema.Attribute{
						attrURL: schema.StringAttribute{
							MarkdownDescription: "URL template with variables documented in Kibana URL drilldown documentation.",
							Required:            true,
						},
						attrLabel: schema.StringAttribute{
							MarkdownDescription: descDisplayLabel,
							Required:            true,
						},
						attrTrigger: schema.StringAttribute{
							MarkdownDescription: "Trigger that activates the drilldown. Required; the Kibana dashboard API rejects URL drilldowns when this field is omitted.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOf(StructuredDrilldownURLTriggerEnum...),
							},
						},
						attrEncodeURL: schema.BoolAttribute{
							MarkdownDescription: "Escape the URL via percent-encoding when set.",
							Optional:            true,
						},
						attrOpenInNewTab: schema.BoolAttribute{
							MarkdownDescription: descOpenInNewTab,
							Optional:            true,
						},
					},
				},
			},
		},
	}
}

// TimeRangeAttributes returns inner schema attributes for panel/dashboard `time_range` objects:
// required `from` and `to`, optional `mode` (`absolute` | `relative`).
func TimeRangeAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"from": schema.StringAttribute{
			MarkdownDescription: "Start of the time range (e.g., 'now-15m', '2023-01-01T00:00:00Z').",
			Required:            true,
		},
		"to": schema.StringAttribute{
			MarkdownDescription: "End of the time range (e.g., 'now', '2023-12-31T23:59:59Z').",
			Required:            true,
		},
		"mode": schema.StringAttribute{
			MarkdownDescription: "Optional time range mode. When set, must be `absolute` or `relative`.",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("absolute", "relative"),
			},
		},
	}
}

// TimeRangeSchema returns an optional SingleNestedAttribute for panel-level time ranges, using TimeRangeAttributes.
func TimeRangeSchema(markdownDescription string) schema.SingleNestedAttribute {
	markdownDescription = typeutils.NonZero(markdownDescription, "Optional panel time range (`from`, `to`, and optional `mode`).")
	return schema.SingleNestedAttribute{
		MarkdownDescription: markdownDescription,
		Optional:            true,
		Attributes:          TimeRangeAttributes(),
	}
}
