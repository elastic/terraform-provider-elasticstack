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

package sloalerts

import (
	_ "embed"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

//go:embed descriptions/slo_alerts_panel_slos.md
var sloAlertsPanelSlosDescription string

//go:embed descriptions/slo_alerts_panel_drilldowns.md
var sloAlertsPanelDrilldownsDescription string

const panelType = "slo_alerts"

const panelConfigBlock = panelType + "_config"

// SchemaAttribute returns slo_alerts_config.
func SchemaAttribute() schema.Attribute {
	return panelkit.PanelConfigBlock(panelkit.PanelConfigBlockOpts{
		Description: "Configuration for an `slo_alerts` panel (`kbn-dashboard-panel-type-slo_alerts`). " +
			"Required when `type` is `slo_alerts`.",
		BlockName:  panelConfigBlock,
		PanelType:  panelType,
		Required:   true,
		Attributes: nestedAttributes(),
	})
}

func nestedAttributes() map[string]schema.Attribute {
	attrs := panelkit.PanelPresentationAttributes()
	attrs["slos"] = schema.ListNestedAttribute{
		MarkdownDescription: sloAlertsPanelSlosDescription,
		Required:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"slo_id": schema.StringAttribute{
					MarkdownDescription: "Identifier of the SLO to include.",
					Required:            true,
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
					},
				},
				"slo_instance_id": schema.StringAttribute{
					MarkdownDescription: "SLO instance ID when the SLO uses grouping. Omit for all instances (API default `\"*\"`). Unset values stay null when the API echoes that default (REQ-009).",
					Optional:            true,
				},
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
			listvalidator.SizeAtMost(100),
		},
	}
	attrs["drilldowns"] = panelkit.URLDrilldownListAttribute(
		sloAlertsPanelDrilldownsDescription,
		panelkit.URLDrilldownOptions{
			URLMarkdownDescription:          "Templated URL for the drilldown.",
			LabelMarkdownDescription:        "Display label shown in the drilldown menu.",
			EncodeURLMarkdownDescription:    "When true, the URL is percent-encoded. Omit to use the API default.",
			OpenInNewTabMarkdownDescription: "When true, the URL opens in a new browser tab. Omit to use the API default.",
		},
	)
	return attrs
}
