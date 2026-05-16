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

package esqlcontrol

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const panelType = "esql_control"

// SchemaAttribute returns the dashboard panel esql_control_config block.
func SchemaAttribute() schema.Attribute {
	return panelkit.PanelConfigBlock(panelkit.PanelConfigBlockOpts{
		Description: "Configuration for an ES|QL control panel. Use this to manage ES|QL variable controls on a dashboard.",
		BlockName:   "esql_control_config",
		PanelType:   panelType,
		Required:    true,
		Attributes: map[string]schema.Attribute{
			"selected_options": schema.ListAttribute{
				MarkdownDescription: "List of currently selected option values for the control.",
				Required:            true,
				ElementType:         types.StringType,
			},
			"variable_name": schema.StringAttribute{
				MarkdownDescription: "The ES|QL variable name that this control binds to.",
				Required:            true,
			},
			"variable_type": schema.StringAttribute{
				MarkdownDescription: "The type of ES|QL variable. Allowed values: `fields`, `values`, `functions`, `time_literal`, `multi_values`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("fields", "values", "functions", "time_literal", "multi_values"),
				},
			},
			"esql_query": schema.StringAttribute{
				MarkdownDescription: "The ES|QL query used to populate the control's options.",
				Required:            true,
			},
			"control_type": schema.StringAttribute{
				MarkdownDescription: "The control type. Allowed values: `STATIC_VALUES`, `VALUES_FROM_QUERY`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("STATIC_VALUES", "VALUES_FROM_QUERY"),
				},
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "A human-readable title displayed above the control widget.",
				Optional:            true,
			},
			"single_select": schema.BoolAttribute{
				MarkdownDescription: "When true, restricts the control to single-value selection.",
				Optional:            true,
			},
			"available_options": schema.ListAttribute{
				MarkdownDescription: "Pre-populated list of available options shown before the query executes.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"display_settings": schema.SingleNestedAttribute{
				MarkdownDescription: "Display configuration for the control widget.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"placeholder": schema.StringAttribute{
						MarkdownDescription: "Placeholder text shown when no option is selected.",
						Optional:            true,
					},
					"hide_action_bar": schema.BoolAttribute{
						MarkdownDescription: "Whether to hide the action bar on the control.",
						Optional:            true,
					},
					"hide_exclude": schema.BoolAttribute{
						MarkdownDescription: "Whether to hide the exclude option.",
						Optional:            true,
					},
					"hide_exists": schema.BoolAttribute{
						MarkdownDescription: "Whether to hide the exists filter option.",
						Optional:            true,
					},
					"hide_sort": schema.BoolAttribute{
						MarkdownDescription: "Whether to hide the sort option.",
						Optional:            true,
					},
				},
			},
		},
	})
}
