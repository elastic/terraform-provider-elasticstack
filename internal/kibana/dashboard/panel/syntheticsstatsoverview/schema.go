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

package syntheticsstatsoverview

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const panelType = "synthetics_stats_overview"

// SchemaAttribute returns the synthetics_stats_overview_config SingleNestedAttribute definition.
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
			"Configuration for a Synthetics stats overview panel. "+
				"All fields are optional; an absent or empty block shows statistics "+
				"for all monitors visible within the space.",
			"synthetics_stats_overview_config",
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
			"drilldowns": schema.ListNestedAttribute{
				MarkdownDescription: "Optional list of URL drilldown actions attached to the panel. The API allows up to 100 drilldowns per panel.",
				Optional:            true,
				NestedObject:        panelkit.URLDrilldownSchema(panelkit.URLDrilldownOptions{}),
			},
			"filters": schema.SingleNestedAttribute{
				MarkdownDescription: "Optional Synthetics monitor filter constraints. Each filter category " +
					"accepts a list of `{ label, value }` objects. Omit the block or individual categories " +
					"to apply no filtering for those dimensions.",
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"projects": schema.ListNestedAttribute{
						MarkdownDescription: "Filter by Synthetics project.",
						Optional:            true,
						NestedObject:        filterItem,
					},
					"tags": schema.ListNestedAttribute{
						MarkdownDescription: "Filter by monitor tag.",
						Optional:            true,
						NestedObject:        filterItem,
					},
					"monitor_ids": schema.ListNestedAttribute{
						MarkdownDescription: "Filter by monitor ID. The API accepts up to 5000 entries.",
						Optional:            true,
						NestedObject:        filterItem,
					},
					"locations": schema.ListNestedAttribute{
						MarkdownDescription: "Filter by monitor location.",
						Optional:            true,
						NestedObject:        filterItem,
					},
					"monitor_types": schema.ListNestedAttribute{
						MarkdownDescription: "Filter by monitor type (e.g. `browser`, `http`).",
						Optional:            true,
						NestedObject:        filterItem,
					},
				},
			},
		},
		Validators: []validator.Object{
			objectvalidator.ConflictsWith(
				panelkit.SiblingTypedPanelConfigConflictPathsExcept("synthetics_stats_overview_config", panelkit.TypedSiblingPanelConfigBlockNames)...,
			),
			validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelType}),
		},
	}
}
