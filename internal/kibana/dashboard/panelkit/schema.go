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
)

// Default URL drilldown element descriptions (typed panels that fix trigger/type in the model layer).
const (
	urlDrilldownDefaultURLDescription          = "Templated URL for the drilldown."
	urlDrilldownDefaultLabelDescription        = "Display label shown in the drilldown menu."
	urlDrilldownDefaultEncodeURLDescription    = "When true, the URL is percent-encoded. Omit to use the API default."
	urlDrilldownDefaultOpenInNewTabDescription = "When true, the URL opens in a new browser tab. Omit to use the API default."
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
	urlDesc := opts.URLMarkdownDescription
	if urlDesc == "" {
		urlDesc = urlDrilldownDefaultURLDescription
	}
	labelDesc := opts.LabelMarkdownDescription
	if labelDesc == "" {
		labelDesc = urlDrilldownDefaultLabelDescription
	}
	encodeDesc := opts.EncodeURLMarkdownDescription
	if encodeDesc == "" {
		encodeDesc = urlDrilldownDefaultEncodeURLDescription
	}
	openDesc := opts.OpenInNewTabMarkdownDescription
	if openDesc == "" {
		openDesc = urlDrilldownDefaultOpenInNewTabDescription
	}
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				MarkdownDescription: urlDesc,
				Required:            true,
			},
			"label": schema.StringAttribute{
				MarkdownDescription: labelDesc,
				Required:            true,
			},
			"encode_url": schema.BoolAttribute{
				MarkdownDescription: encodeDesc,
				Optional:            true,
			},
			"open_in_new_tab": schema.BoolAttribute{
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
	urlDesc := urlOpts.URLMarkdownDescription
	if urlDesc == "" {
		urlDesc = urlDrilldownDefaultURLDescription
	}
	labelDescURL := urlOpts.LabelMarkdownDescription
	if labelDescURL == "" {
		labelDescURL = urlDrilldownDefaultLabelDescription
	}
	triggerDesc := urlOpts.TriggerMarkdownDescription
	if triggerDesc == "" {
		triggerDesc = defaultImageURLDrilldownTriggerDescription
	}
	encodeDescURL := urlOpts.EncodeURLMarkdownDescription
	if encodeDescURL == "" {
		encodeDescURL = urlDrilldownDefaultEncodeURLDescription
	}
	openDescURL := urlOpts.OpenInNewTabMarkdownDescription
	if openDescURL == "" {
		openDescURL = urlDrilldownDefaultOpenInNewTabDescription
	}

	dashboardTriggerDesc := dd.Trigger
	if dashboardTriggerDesc == "" {
		dashboardTriggerDesc = "Dashboard drilldowns on image panels only support `on_click_image` (see Kibana `kbn-dashboard-panel-type-image` drilldown schema)."
	}

	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"dashboard_drilldown": schema.SingleNestedAttribute{
				MarkdownDescription: "Open another dashboard when the image is clicked. Mutually exclusive with `url_drilldown` in the same entry.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"dashboard_id": schema.StringAttribute{
						MarkdownDescription: nz(dd.DashboardID, "Target dashboard saved object id."),
						Required:            true,
					},
					"label": schema.StringAttribute{
						MarkdownDescription: nz(dd.Label, "Label shown for this drilldown."),
						Required:            true,
					},
					"trigger": schema.StringAttribute{
						MarkdownDescription: dashboardTriggerDesc,
						Required:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("on_click_image"),
						},
					},
					"use_filters": schema.BoolAttribute{
						MarkdownDescription: nz(dd.UseFilters, "When true, passes the current dashboard filters to the opened dashboard. Omit for API default (typically `false`)."),
						Optional:            true,
					},
					"use_time_range": schema.BoolAttribute{
						MarkdownDescription: nz(dd.UseTimeRange, "When true, passes the current time range to the opened dashboard. Omit for API default (typically `false`)."),
						Optional:            true,
					},
					"open_in_new_tab": schema.BoolAttribute{
						MarkdownDescription: nz(dd.OpenInNewTab, "When true, opens the target dashboard in a new browser tab. Omit for API default (typically `false`)."),
						Optional:            true,
					},
				},
			},
			"url_drilldown": schema.SingleNestedAttribute{
				MarkdownDescription: "URL drilldown entry. Mutually exclusive with `dashboard_drilldown` in the same list element.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						MarkdownDescription: urlDesc,
						Required:            true,
					},
					"label": schema.StringAttribute{
						MarkdownDescription: labelDescURL,
						Required:            true,
					},
					"trigger": schema.StringAttribute{
						MarkdownDescription: triggerDesc,
						Required:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("on_click_image", "on_open_panel_menu"),
						},
					},
					"encode_url": schema.BoolAttribute{
						MarkdownDescription: encodeDescURL,
						Optional:            true,
					},
					"open_in_new_tab": schema.BoolAttribute{
						MarkdownDescription: openDescURL,
						Optional:            true,
					},
				},
			},
		},
		Validators: []validator.Object{
			ExactlyOneOfNestedAttrsValidator(ExactlyOneOfNestedAttrsOpts{
				AttrNames:     []string{"dashboard_drilldown", "url_drilldown"},
				Summary:       "Invalid drilldown entry",
				MissingDetail: "Exactly one of `dashboard_drilldown` or `url_drilldown` must be set.",
				TooManyDetail: "Exactly one of `dashboard_drilldown` or `url_drilldown` must be set, not both.",
				Description:   "Ensures exactly one of `dashboard_drilldown` or `url_drilldown` is set on each `drilldowns` entry.",
			}),
		},
	}
}

func nz(s, def string) string {
	if s != "" {
		return s
	}
	return def
}
