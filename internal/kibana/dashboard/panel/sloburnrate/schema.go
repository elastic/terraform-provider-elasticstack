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

package sloburnrate

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// SchemaAttribute returns the slo_burn_rate_config SingleNestedAttribute definition.
func SchemaAttribute() schema.Attribute {
	attrs := panelkit.PanelPresentationAttributes()
	attrs["slo_id"] = schema.StringAttribute{
		MarkdownDescription: "The ID of the SLO to display the burn rate for.",
		Required:            true,
	}
	attrs["duration"] = schema.StringAttribute{
		MarkdownDescription: "Duration for the burn rate chart in the format `[value][unit]`, where unit is `m` (minutes), `h` (hours), or `d` (days). For example: `5m`, `3h`, `6d`.",
		Required:            true,
		Validators: []validator.String{
			stringvalidator.RegexMatches(
				sloBurnRateDurationRegexp,
				"must match the pattern `^\\d+[mhd]$` (a positive integer followed by m, h, or d)",
			),
		},
	}
	attrs["slo_instance_id"] = schema.StringAttribute{
		MarkdownDescription: "ID of the SLO instance. Set when the SLO uses `group_by`; identifies which instance to show. Omit to show all instances (API default `\"*\"`).",
		Optional:            true,
	}
	attrs["drilldowns"] = panelkit.URLDrilldownListAttribute(
		"Optional list of URL drilldowns attached to the panel.",
		panelkit.URLDrilldownOptions{},
	)

	return panelkit.PanelConfigBlock(panelkit.PanelConfigBlockOpts{
		Description: "Configuration for an SLO burn rate panel. Use this for panels that visualize the burn rate of an SLO over a configurable look-back window.",
		BlockName:   "slo_burn_rate_config",
		PanelType:   "slo_burn_rate",
		Attributes:  attrs,
	})
}
