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

package timeslider

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework-validators/float32validator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const panelType = "time_slider_control"

// SchemaAttribute returns the time_slider_control_config SingleNestedAttribute for dashboard panels.
func SchemaAttribute() schema.Attribute {
	return panelkit.PanelConfigBlock(panelkit.PanelConfigBlockOpts{
		Description: "Configuration for a time slider control panel. Controls the visible time window within the dashboard's global time range.",
		BlockName:   "time_slider_control_config",
		PanelType:   panelType,
		Attributes: map[string]schema.Attribute{
			"start_percentage_of_time_range": schema.Float32Attribute{
				MarkdownDescription: "Start of the visible time window as a fraction of the dashboard global range (0.0–1.0). " +
					"Float32 in state matches the Kibana API and avoids refresh drift.",
				Optional: true,
				Validators: []validator.Float32{
					float32validator.Between(0.0, 1.0),
				},
			},
			"end_percentage_of_time_range": schema.Float32Attribute{
				MarkdownDescription: "End of the visible time window as a fraction of the dashboard global range (0.0–1.0). " +
					"Float32 in state matches the Kibana API and avoids refresh drift.",
				Optional: true,
				Validators: []validator.Float32{
					float32validator.Between(0.0, 1.0),
				},
			},
			"is_anchored": schema.BoolAttribute{
				MarkdownDescription: "Whether the start of the time window is anchored (fixed), so only the end slides.",
				Optional:            true,
			},
		},
	})
}
