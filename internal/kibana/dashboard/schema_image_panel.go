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

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// getImagePanelConfigAttributes returns schema attributes for `image_config`.
func getImagePanelConfigAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"src": schema.SingleNestedAttribute{
			MarkdownDescription: imagePanelSrcDescription,
			Required:            true,
			Attributes: map[string]schema.Attribute{
				"file": schema.SingleNestedAttribute{
					MarkdownDescription: "Use an uploaded file as the image source. Mutually exclusive with `url` inside `src`.",
					Optional:            true,
					Attributes: map[string]schema.Attribute{
						"file_id": schema.StringAttribute{
							MarkdownDescription: "Kibana file identifier for the uploaded image.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
					},
				},
				"url": schema.SingleNestedAttribute{
					MarkdownDescription: "Use an external URL as the image source. Mutually exclusive with `file` inside `src`.",
					Optional:            true,
					Attributes: map[string]schema.Attribute{
						"url": schema.StringAttribute{
							MarkdownDescription: "HTTPS or HTTP URL of the image.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
					},
				},
			},
			Validators: []validator.Object{
				imagePanelSrcValidator{},
			},
		},
		"alt_text": schema.StringAttribute{
			MarkdownDescription: "Accessible alternate text for the image.",
			Optional:            true,
		},
		"object_fit": schema.StringAttribute{
			MarkdownDescription: "How the image is sized within its panel container. Omit to leave unset; the Kibana API defaults to `contain`. " +
				"Unset values stay null in Terraform state when the API echoes that default (REQ-009).",
			Optional: true,
			Validators: []validator.String{
				stringvalidator.OneOf("fill", "contain", "cover", "none"),
			},
		},
		"background_color": schema.StringAttribute{
			MarkdownDescription: "Background color behind the image (CSS color string).",
			Optional:            true,
		},
		"title": schema.StringAttribute{
			MarkdownDescription: "Panel title shown by Kibana.",
			Optional:            true,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "Panel description shown by Kibana.",
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
			MarkdownDescription: imagePanelDrilldownsDescription,
			Optional:            true,
			Validators: []validator.List{
				listvalidator.SizeAtMost(100),
			},
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"dashboard_drilldown": schema.SingleNestedAttribute{
						MarkdownDescription: "Open another dashboard when the image is clicked. Mutually exclusive with `url_drilldown` in the same entry.",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"dashboard_id": schema.StringAttribute{
								MarkdownDescription: "Target dashboard saved object id.",
								Required:            true,
							},
							"label": schema.StringAttribute{
								MarkdownDescription: "Label shown for this drilldown.",
								Required:            true,
							},
							"trigger": schema.StringAttribute{
								MarkdownDescription: "Dashboard drilldowns on image panels only support `on_click_image` (see Kibana `kbn-dashboard-panel-type-image` drilldown schema).",
								Required:            true,
								Validators: []validator.String{
									stringvalidator.OneOf("on_click_image"),
								},
							},
							"use_filters": schema.BoolAttribute{
								MarkdownDescription: "When true, passes the current dashboard filters to the opened dashboard. Omit for API default (typically `false`).",
								Optional:            true,
							},
							"use_time_range": schema.BoolAttribute{
								MarkdownDescription: "When true, passes the current time range to the opened dashboard. Omit for API default (typically `false`).",
								Optional:            true,
							},
							"open_in_new_tab": schema.BoolAttribute{
								MarkdownDescription: "When true, opens the target dashboard in a new browser tab. Omit for API default (typically `false`).",
								Optional:            true,
							},
						},
					},
					"url_drilldown": schema.SingleNestedAttribute{
						MarkdownDescription: "URL drilldown entry. Mutually exclusive with `dashboard_drilldown` in the same list element.",
						Optional:            true,
						Attributes:          urlDrilldownNestedAttributeObject(imageURLDrilldownOpts()).Attributes,
					},
				},
				Validators: []validator.Object{
					imagePanelDrilldownEntryValidator{},
				},
			},
		},
	}
}

func imageURLDrilldownOpts() URLDrilldownNestedOpts {
	return URLDrilldownNestedOpts{
		AllowedTriggers:                 []string{"on_click_image", "on_open_panel_menu"},
		TriggerMarkdownDescription:      "When this drilldown runs. Allowed values: `on_click_image`, `on_open_panel_menu` (Kibana image panel URL drilldown triggers).",
		URLMarkdownDescription:          imagePanelURLDrilldownURLDescription,
		LabelMarkdownDescription:        "Display label for the drilldown.",
		EncodeURLMarkdownDescription:    "When true, the URL is percent-encoded. Omit to use the API default; unset stays null when the API echoes the default (REQ-009).",
		OpenInNewTabMarkdownDescription: "When true, opens in a new browser tab. Omit to use the API default; unset stays null when the API echoes the default (REQ-009).",
	}
}

// imagePanelSrcValidator ensures exactly one of `file` or `url` is set under `src`.
var _ validator.Object = imagePanelSrcValidator{}

type imagePanelSrcValidator struct{}

func (v imagePanelSrcValidator) Description(_ context.Context) string {
	return "Ensures exactly one of `file` or `url` is set inside `image_config.src`."
}

func (v imagePanelSrcValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v imagePanelSrcValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	attrs := req.ConfigValue.Attributes()
	fileVal := attrs["file"]
	urlVal := attrs["url"]

	fileSet := attrObjectSet(fileVal)
	urlSet := attrObjectSet(urlVal)

	if fileSet && urlSet {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid image_config.src", "Exactly one of `file` or `url` must be set inside `src`, not both.")
		return
	}
	if !fileSet && !urlSet {
		fileUnknown := fileVal != nil && fileVal.IsUnknown()
		urlUnknown := urlVal != nil && urlVal.IsUnknown()
		if fileUnknown || urlUnknown {
			return
		}
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid image_config.src", "Exactly one of `file` or `url` must be set inside `src`.")
	}
}

func attrObjectSet(v attr.Value) bool {
	return v != nil && !v.IsNull() && !v.IsUnknown()
}

// imagePanelDrilldownEntryValidator ensures exactly one drilldown sub-block per list element.
var _ validator.Object = imagePanelDrilldownEntryValidator{}

type imagePanelDrilldownEntryValidator struct{}

func (v imagePanelDrilldownEntryValidator) Description(_ context.Context) string {
	return "Ensures exactly one of `dashboard_drilldown` or `url_drilldown` is set on each `drilldowns` entry."
}

func (v imagePanelDrilldownEntryValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v imagePanelDrilldownEntryValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	attrs := req.ConfigValue.Attributes()
	dashVal := attrs["dashboard_drilldown"]
	urlVal := attrs["url_drilldown"]

	dashSet := attrObjectSet(dashVal)
	urlSet := attrObjectSet(urlVal)

	if dashSet && urlSet {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid drilldown entry", "Exactly one of `dashboard_drilldown` or `url_drilldown` must be set, not both.")
		return
	}
	if !dashSet && !urlSet {
		dashUnknown := dashVal != nil && dashVal.IsUnknown()
		urlUnknown := urlVal != nil && urlVal.IsUnknown()
		if dashUnknown || urlUnknown {
			return
		}
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid drilldown entry", "Exactly one of `dashboard_drilldown` or `url_drilldown` must be set.")
	}
}
