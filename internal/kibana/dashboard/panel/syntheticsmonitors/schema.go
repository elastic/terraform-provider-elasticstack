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

package syntheticsmonitors

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const panelType = "synthetics_monitors"

// SchemaAttribute returns the synthetics_monitors_config SingleNestedAttribute definition.
func SchemaAttribute() schema.Attribute {
	filterItem := schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"label": schema.StringAttribute{
				MarkdownDescription: "Display label for the filter option.",
				Required:            true,
			},
			"value": schema.StringAttribute{
				MarkdownDescription: "Value for the filter option.",
				Required:            true,
			},
		},
	}

	return schema.SingleNestedAttribute{
		MarkdownDescription: panelkit.PanelConfigDescription(
			"Configuration for a Synthetics monitors panel. Displays a table of Elastic Synthetics monitors "+
				"and their current status. All fields are optional — omit the block entirely for a bare panel with no filtering.",
			"synthetics_monitors_config",
			panelkit.TypedSiblingPanelConfigBlockNames,
		),
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"title": schema.StringAttribute{
				MarkdownDescription: "Display title shown in the panel header.",
				Optional:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Descriptive text for the panel.",
				Optional:            true,
			},
			"hide_title": schema.BoolAttribute{
				MarkdownDescription: "When true, suppresses the panel title in the dashboard.",
				Optional:            true,
			},
			"hide_border": schema.BoolAttribute{
				MarkdownDescription: "When true, suppresses the panel border in the dashboard.",
				Optional:            true,
			},
			"view": schema.StringAttribute{
				MarkdownDescription: "View mode for the panel. Valid values are `cardView` and `compactView`.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("cardView", "compactView"),
				},
			},
			"filters": schema.SingleNestedAttribute{
				MarkdownDescription: "Optional filter configuration for the Synthetics monitors panel. Omit to show all monitors.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"projects": schema.ListNestedAttribute{
						MarkdownDescription: "Filter by project. Each entry has a `label` (display name) and a `value` (project ID).",
						Optional:            true,
						NestedObject:        filterItem,
					},
					"tags": schema.ListNestedAttribute{
						MarkdownDescription: "Filter by tags. Each entry has a `label` (display name) and a `value` (tag).",
						Optional:            true,
						NestedObject:        filterItem,
					},
					"monitor_ids": schema.ListNestedAttribute{
						MarkdownDescription: "Filter by monitor IDs. Each entry has a `label` (display name) and a `value` (monitor ID). The Kibana API accepts up to 5000 items.",
						Optional:            true,
						NestedObject:        filterItem,
					},
					"locations": schema.ListNestedAttribute{
						MarkdownDescription: "Filter by monitor locations. Each entry has a `label` (display name) and a `value` (location ID).",
						Optional:            true,
						NestedObject:        filterItem,
					},
					"monitor_types": schema.ListNestedAttribute{
						MarkdownDescription: "Filter by monitor types. Each entry has a `label` (display name) and a `value` (monitor type, e.g. `browser`, `http`, `tcp`, `icmp`).",
						Optional:            true,
						NestedObject:        filterItem,
					},
				},
			},
		},
		Validators: []validator.Object{
			objectvalidator.ConflictsWith(
				panelkit.SiblingTypedPanelConfigConflictPathsExcept("synthetics_monitors_config", panelkit.TypedSiblingPanelConfigBlockNames)...,
			),
			validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelType}),
		},
	}
}
