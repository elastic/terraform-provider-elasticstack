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

package image

import (
	"context"
	_ "embed"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

//go:embed descriptions/image_panel_src.md
var imagePanelSrcDescription string

//go:embed descriptions/image_panel_drilldowns.md
var imagePanelDrilldownsDescription string

//go:embed descriptions/image_panel_url_drilldown_url.md
var imagePanelURLDrilldownURLDescription string

const panelConfigBlock = panelType + "_config"

// SchemaAttribute returns image_config.
func SchemaAttribute() schema.Attribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: panelkit.PanelConfigDescription(
			"Configuration for an `image` panel (`kbn-dashboard-panel-type-image`). Required when `type` is `image`. "+
				"References the Kibana Dashboard API image embeddable `config` shape.",
			panelConfigBlock,
			panelkit.TypedSiblingPanelConfigBlockNames,
		),
		Optional:   true,
		Attributes: nestedAttributes(),
		Validators: []validator.Object{
			objectvalidator.ConflictsWith(
				panelkit.SiblingTypedPanelConfigConflictPathsExcept(panelConfigBlock, panelkit.TypedSiblingPanelConfigBlockNames)...,
			),
			validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelType}),
			validators.RequiredIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelType}),
		},
	}
}

func nestedAttributes() map[string]schema.Attribute {
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
				srcValidator{},
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
			NestedObject: panelkit.ImageDrilldownSchema(panelkit.ImageDrilldownOptions{
				DashboardMarkdownDescriptions: panelkit.ImageDashboardDrilldownDescriptions{},
				URLDrilldownMarkdown: panelkit.ImageURLDrilldownOptions{
					URLMarkdownDescription:          imagePanelURLDrilldownURLDescription,
					EncodeURLMarkdownDescription:    "When true, the URL is percent-encoded. Omit to use the API default; unset stays null when the API echoes the default (REQ-009).",
					OpenInNewTabMarkdownDescription: "When true, opens in a new browser tab. Omit to use the API default; unset stays null when the API echoes the default (REQ-009).",
				},
			}),
		},
	}
}

var _ validator.Object = srcValidator{}

type srcValidator struct{}

func (v srcValidator) Description(_ context.Context) string {
	return "Ensures exactly one of `file` or `url` is set inside `image_config.src`."
}

func (v srcValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (srcValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
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
