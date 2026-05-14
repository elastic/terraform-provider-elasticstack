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

package optionslist

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const panelType = "options_list_control"

// SchemaAttribute returns the dashboard panel options_list_control_config block.
func SchemaAttribute() schema.Attribute {
	return panelkit.PanelConfigBlock(panelkit.PanelConfigBlockOpts{
		Description: "Configuration for an options list control panel. Provides a dropdown or multi-select filter based on a field in a data view.",
		BlockName:   "options_list_control_config",
		PanelType:   panelType,
		Required:    true,
		Attributes: map[string]schema.Attribute{
			"data_view_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the data view that the control is tied to.",
				Required:            true,
			},
			"field_name": schema.StringAttribute{
				MarkdownDescription: "The name of the field in the data view that the control is tied to.",
				Required:            true,
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "Human-readable label displayed above the control.",
				Optional:            true,
			},
			"use_global_filters": schema.BoolAttribute{
				MarkdownDescription: "Whether the control applies the dashboard's global filters to its own query.",
				Optional:            true,
			},
			"ignore_validations": schema.BoolAttribute{
				MarkdownDescription: "Whether the control skips field-level validation against the data view.",
				Optional:            true,
			},
			"single_select": schema.BoolAttribute{
				MarkdownDescription: "When true, only one option may be selected at a time.",
				Optional:            true,
			},
			"exclude": schema.BoolAttribute{
				MarkdownDescription: "When true, selected options are used as an exclusion filter rather than an inclusion filter.",
				Optional:            true,
			},
			"exists_selected": schema.BoolAttribute{
				MarkdownDescription: "When true, the control filters for documents where the field exists.",
				Optional:            true,
			},
			"run_past_timeout": schema.BoolAttribute{
				MarkdownDescription: "When true, the control continues to show results even when the underlying query times out.",
				Optional:            true,
			},
			"search_technique": schema.StringAttribute{
				MarkdownDescription: "The technique used to match suggestions. Must be one of `prefix`, `wildcard`, or `exact` when set.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("prefix", "wildcard", "exact"),
				},
			},
			"selected_options": schema.ListAttribute{
				MarkdownDescription: "The initially or persistently selected option values. All values are represented as strings.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"display_settings": schema.SingleNestedAttribute{
				MarkdownDescription: "Display preferences for the control widget.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"placeholder": schema.StringAttribute{
						MarkdownDescription: "Placeholder text shown when no option is selected.",
						Optional:            true,
					},
					"hide_action_bar": schema.BoolAttribute{
						MarkdownDescription: "When true, hides the action bar on the control.",
						Optional:            true,
					},
					"hide_exclude": schema.BoolAttribute{
						MarkdownDescription: "When true, hides the exclude toggle.",
						Optional:            true,
					},
					"hide_exists": schema.BoolAttribute{
						MarkdownDescription: "When true, hides the exists filter option.",
						Optional:            true,
					},
					"hide_sort": schema.BoolAttribute{
						MarkdownDescription: "When true, hides the sort control.",
						Optional:            true,
					},
				},
			},
			"sort": schema.SingleNestedAttribute{
				MarkdownDescription: "Default sort configuration for the suggestion list.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"by": schema.StringAttribute{
						MarkdownDescription: "The field or criterion to sort by. Must be one of `_count` or `_key`.",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("_count", "_key"),
						},
					},
					"direction": schema.StringAttribute{
						MarkdownDescription: "The sort direction. Must be one of `asc` or `desc`.",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("asc", "desc"),
						},
					},
				},
			},
		},
	})
}
