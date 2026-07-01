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

package apmservicemap

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const panelType = "apm_service_map"

// SchemaAttribute returns the apm_service_map_config SingleNestedAttribute definition.
func SchemaAttribute() schema.Attribute {
	attrs := panelkit.PanelPresentationAttributes()
	attrs["environment"] = schema.StringAttribute{
		MarkdownDescription: "APM service environment (for example, `production`).",
		Optional:            true,
	}
	attrs["service_name"] = schema.StringAttribute{
		MarkdownDescription: "Focus the service map on a specific APM service.",
		Optional:            true,
	}
	attrs["service_group_id"] = schema.StringAttribute{
		MarkdownDescription: "Opaque identifier of a saved APM service group.",
		Optional:            true,
	}
	attrs["kuery"] = schema.StringAttribute{
		MarkdownDescription: "KQL query string applied to the service map.",
		Optional:            true,
	}
	attrs["map_orientation"] = schema.StringAttribute{
		MarkdownDescription: "Layout orientation of the service map.",
		Optional:            true,
		Validators: []validator.String{
			stringvalidator.OneOf("horizontal", "vertical"),
		},
	}
	attrs["sync_with_dashboard_filters"] = schema.BoolAttribute{
		MarkdownDescription: "When set, the panel follows dashboard-level filters.",
		Optional:            true,
	}
	attrs["alert_status_filter"] = schema.SetAttribute{
		MarkdownDescription: "Filter services by alert status.",
		Optional:            true,
		ElementType:         types.StringType,
		Validators: []validator.Set{
			setvalidator.ValueStringsAre(
				stringvalidator.OneOf("active", "delayed", "recovered", "untracked"),
			),
		},
	}
	attrs["anomaly_severity_filter"] = schema.SetAttribute{
		MarkdownDescription: "Filter services by anomaly severity.",
		Optional:            true,
		ElementType:         types.StringType,
		Validators: []validator.Set{
			setvalidator.ValueStringsAre(
				stringvalidator.OneOf("low", "warning", "minor", "major", "critical", "unknown"),
			),
		},
	}
	attrs["connection_filter"] = schema.SetAttribute{
		MarkdownDescription: "Filter services by connection state.",
		Optional:            true,
		ElementType:         types.StringType,
		Validators: []validator.Set{
			setvalidator.ValueStringsAre(
				stringvalidator.OneOf("connected", "orphaned"),
			),
		},
	}
	attrs["slo_status_filter"] = schema.SetAttribute{
		MarkdownDescription: "Filter services by SLO status.",
		Optional:            true,
		ElementType:         types.StringType,
		Validators: []validator.Set{
			setvalidator.ValueStringsAre(
				stringvalidator.OneOf("degrading", "healthy", "noData", "violated"),
			),
		},
	}
	attrs["time_range"] = panelkit.TimeRangeSchema(
		"Optional panel time range (`from`, `to`, and optional `mode`).",
	)

	return panelkit.PanelConfigBlock(panelkit.PanelConfigBlockOpts{
		Description: "Configuration for an APM service map panel. All fields are optional.",
		BlockName:   "apm_service_map_config",
		PanelType:   panelType,
		Attributes:  attrs,
	})
}
