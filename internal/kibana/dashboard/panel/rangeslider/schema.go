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

package rangeslider

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const panelType = "range_slider_control"

// SchemaAttribute returns the dashboard panel range_slider_control_config block.
func SchemaAttribute() schema.Attribute {
	return panelkit.PanelConfigBlock(panelkit.PanelConfigBlockOpts{
		Description: "Configuration for a range slider control panel. Provides a min/max range filter tied to a data view field.",
		BlockName:   "range_slider_control_config",
		PanelType:   panelType,
		Required:    true,
		Attributes: map[string]schema.Attribute{
			"title": schema.StringAttribute{
				MarkdownDescription: "A human-readable title for the control.",
				Optional:            true,
			},
			"data_view_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the data view that the control is tied to.",
				Required:            true,
			},
			"field_name": schema.StringAttribute{
				MarkdownDescription: "The name of the field in the data view that the control is tied to.",
				Required:            true,
			},
			"use_global_filters": schema.BoolAttribute{
				MarkdownDescription: "Whether the control respects dashboard-level filters.",
				Optional:            true,
			},
			"ignore_validations": schema.BoolAttribute{
				MarkdownDescription: "Whether to suppress validation errors during intermediate states.",
				Optional:            true,
			},
			"value": schema.ListAttribute{
				MarkdownDescription: "Initial range as a list of exactly 2 strings: [min, max].",
				ElementType:         types.StringType,
				Optional:            true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(2),
					listvalidator.SizeAtMost(2),
				},
			},
			"step": schema.Float32Attribute{
				MarkdownDescription: "The step size for the range slider. Stored as float32 to match the Kibana API type and avoid refresh drift.",
				Optional:            true,
			},
		},
	})
}
