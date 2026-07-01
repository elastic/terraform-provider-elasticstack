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

package aiopspatternanalysis

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework-validators/float32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const panelType = "aiops_pattern_analysis"

// SchemaAttribute returns the aiops_pattern_analysis_config SingleNestedAttribute definition.
func SchemaAttribute() schema.Attribute {
	attrs := panelkit.PanelPresentationAttributes()
	attrs["data_view_id"] = schema.StringAttribute{
		MarkdownDescription: "The data view ID used for pattern analysis.",
		Required:            true,
	}
	attrs["field_name"] = schema.StringAttribute{
		MarkdownDescription: "The text field on which to run pattern analysis.",
		Required:            true,
	}
	attrs["minimum_time_range"] = schema.StringAttribute{
		MarkdownDescription: "Minimum time range for pattern analysis. One of `no_minimum`, `1_week`, `1_month`, `3_months`, `6_months`.",
		Optional:            true,
		Validators: []validator.String{
			stringvalidator.OneOf("no_minimum", "1_week", "1_month", "3_months", "6_months"),
		},
	}
	attrs["random_sampler_mode"] = schema.StringAttribute{
		MarkdownDescription: "The random sampler mode. One of `off`, `on_automatic`, `on_manual`.",
		Optional:            true,
		Validators: []validator.String{
			stringvalidator.OneOf("off", "on_automatic", "on_manual"),
		},
	}
	attrs["random_sampler_probability"] = schema.Float32Attribute{
		MarkdownDescription: "Sampling probability, only meaningful when `random_sampler_mode = on_manual`. " +
			"Must be between `0.00001` and `0.5`. Float32 in state matches the Kibana API and avoids refresh drift.",
		Optional: true,
		Validators: []validator.Float32{
			float32validator.Between(0.00001, 0.5),
		},
	}
	attrs["time_range"] = panelkit.TimeRangeSchema(
		"Optional panel time range (`from`, `to`, optional `mode`). When omitted, the panel inherits the dashboard `time_range` and this attribute stays null in state (REQ-009).",
	)

	return panelkit.PanelConfigBlock(panelkit.PanelConfigBlockOpts{
		Description: "Configuration for an AIOps pattern analysis panel. Anchored to a data view and text field; " +
			"optional sampling and time-range controls follow the API-documented bounds.",
		BlockName:  "aiops_pattern_analysis_config",
		PanelType:  panelType,
		Attributes: attrs,
	})
}
