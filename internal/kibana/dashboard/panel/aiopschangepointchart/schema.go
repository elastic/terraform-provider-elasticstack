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

package aiopschangepointchart

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const panelType = "aiops_change_point_chart"

// SchemaAttribute returns the aiops_change_point_chart_config SingleNestedAttribute definition.
func SchemaAttribute() schema.Attribute {
	attrs := panelkit.PanelPresentationAttributes()
	attrs["data_view_id"] = schema.StringAttribute{
		MarkdownDescription: "The data view ID used for change point detection.",
		Required:            true,
	}
	attrs["metric_field"] = schema.StringAttribute{
		MarkdownDescription: "The metric field used by the aggregation function.",
		Required:            true,
	}
	attrs["aggregation_function"] = schema.StringAttribute{
		MarkdownDescription: "The aggregation function used to calculate the metric values. One of `avg`, `max`, `min`, `sum`.",
		Optional:            true,
		Validators: []validator.String{
			stringvalidator.OneOf("avg", "max", "min", "sum"),
		},
	}
	attrs["split_field"] = schema.StringAttribute{
		MarkdownDescription: "The optional field used to split change-point results.",
		Optional:            true,
	}
	attrs["partitions"] = schema.SetAttribute{
		MarkdownDescription: "Optional split field values to include in the panel. Modelled as a set to prevent " +
			"plan drift from API-returned ordering; duplicate entries are silently deduplicated.",
		Optional:    true,
		ElementType: types.StringType,
	}
	attrs["max_series_to_plot"] = schema.Float32Attribute{
		MarkdownDescription: "Maximum number of change points to visualise. Kibana default is `6`. Float32 in state matches the Kibana API and avoids refresh drift.",
		Optional:            true,
	}
	attrs["view_type"] = schema.StringAttribute{
		MarkdownDescription: "The type of change point detection view to display. One of `charts`, `table`.",
		Optional:            true,
		Validators: []validator.String{
			stringvalidator.OneOf("charts", "table"),
		},
	}
	attrs["time_range"] = panelkit.TimeRangeSchema(
		"Optional panel time range (`from`, `to`, optional `mode`). When omitted, the panel inherits the dashboard `time_range` and this attribute stays null in state (REQ-009).",
	)

	return panelkit.PanelConfigBlock(panelkit.PanelConfigBlockOpts{
		Description: "Configuration for an AIOps change point chart panel. Anchored to a data view and metric field; " +
			"optional aggregation, split, partitions, and view controls follow the API-documented enums.",
		BlockName:  "aiops_change_point_chart_config",
		PanelType:  panelType,
		Required:   true,
		Attributes: attrs,
	})
}
