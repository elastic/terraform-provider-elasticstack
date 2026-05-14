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
	_ "embed"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
	return panelkit.PanelConfigBlock(panelkit.PanelConfigBlockOpts{
		Description: "Configuration for an `image` panel (`kbn-dashboard-panel-type-image`). Required when `type` is `image`. " +
			"References the Kibana Dashboard API image embeddable `config` shape.",
		BlockName:  panelConfigBlock,
		PanelType:  panelType,
		Required:   true,
		Attributes: nestedAttributes(),
	})
}

func nestedAttributes() map[string]schema.Attribute {
	attrs := panelkit.PanelPresentationAttributes()
	attrs["src"] = schema.SingleNestedAttribute{
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
			panelkit.ExactlyOneOfNestedAttrsValidator(panelkit.ExactlyOneOfNestedAttrsOpts{
				AttrNames:     []string{"file", "url"},
				Summary:       "Invalid image_config.src",
				MissingDetail: "Exactly one of `file` or `url` must be set inside `src`.",
				TooManyDetail: "Exactly one of `file` or `url` must be set inside `src`, not both.",
				Description:   "Ensures exactly one of `file` or `url` is set inside `image_config.src`.",
			}),
		},
	}
	attrs["alt_text"] = schema.StringAttribute{
		MarkdownDescription: "Accessible alternate text for the image.",
		Optional:            true,
	}
	attrs["object_fit"] = schema.StringAttribute{
		MarkdownDescription: "How the image is sized within its panel container. Omit to leave unset; the Kibana API defaults to `contain`. " +
			"Unset values stay null in Terraform state when the API echoes that default (REQ-009).",
		Optional: true,
		Validators: []validator.String{
			stringvalidator.OneOf("fill", "contain", "cover", "none"),
		},
	}
	attrs["background_color"] = schema.StringAttribute{
		MarkdownDescription: "Background color behind the image (CSS color string).",
		Optional:            true,
	}
	attrs["drilldowns"] = schema.ListNestedAttribute{
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
	}
	return attrs
}
