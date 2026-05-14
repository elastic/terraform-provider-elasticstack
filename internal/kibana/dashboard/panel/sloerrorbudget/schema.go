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

package sloerrorbudget

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

const panelType = "slo_error_budget"

// SchemaAttribute returns the slo_error_budget_config SingleNestedAttribute definition.
func SchemaAttribute() schema.Attribute {
	attrs := panelkit.PanelPresentationAttributes()
	attrs["slo_id"] = schema.StringAttribute{
		MarkdownDescription: "The ID of the SLO to display the error budget for.",
		Required:            true,
	}
	attrs["slo_instance_id"] = schema.StringAttribute{
		MarkdownDescription: "ID of the SLO instance. Set when the SLO uses group_by; identifies which instance to show. Defaults to `*` (all instances) when omitted.",
		Optional:            true,
	}
	attrs["drilldowns"] = panelkit.URLDrilldownListAttribute(
		"URL drilldowns to configure on the panel.",
		panelkit.URLDrilldownOptions{
			URLMarkdownDescription:          "Templated URL. Variables documented at https://www.elastic.co/docs/explore-analyze/dashboards/drilldowns#url-template-variable",
			LabelMarkdownDescription:        "The label displayed for the drilldown.",
			EncodeURLMarkdownDescription:    "When true, the URL is escaped using percent encoding. Defaults to `true` when omitted.",
			OpenInNewTabMarkdownDescription: "When true, the drilldown URL opens in a new browser tab. Defaults to `true` when omitted.",
		},
	)

	return panelkit.PanelConfigBlock(panelkit.PanelConfigBlockOpts{
		Description: "Configuration for an SLO error budget panel. Displays the burn chart of remaining error budget for a specific SLO.",
		BlockName:   "slo_error_budget_config",
		PanelType:   panelType,
		Attributes:  attrs,
	})
}
