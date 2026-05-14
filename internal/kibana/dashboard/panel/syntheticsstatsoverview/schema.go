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
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/syntheticscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

const panelType = "synthetics_stats_overview"

// SchemaAttribute returns the synthetics_stats_overview_config SingleNestedAttribute definition.
func SchemaAttribute() schema.Attribute {
	attrs := panelkit.PanelPresentationAttributes()
	attrs["drilldowns"] = panelkit.URLDrilldownListAttribute(
		"Optional list of URL drilldown actions attached to the panel. The API allows up to 100 drilldowns per panel.",
		panelkit.URLDrilldownOptions{},
	)
	attrs["filters"] = syntheticscommon.FilterAttribute(syntheticscommon.FilterAttributeOptions{
		BlockMarkdownDescription: "Optional Synthetics monitor filter constraints. Each filter category " +
			"accepts a list of `{ label, value }` objects. Omit the block or individual categories " +
			"to apply no filtering for those dimensions.",
	})

	return panelkit.PanelConfigBlock(panelkit.PanelConfigBlockOpts{
		Description: "Configuration for a Synthetics stats overview panel. " +
			"All fields are optional; an absent or empty block shows statistics " +
			"for all monitors visible within the space.",
		BlockName:  "synthetics_stats_overview_config",
		PanelType:  panelType,
		Attributes: attrs,
	})
}
