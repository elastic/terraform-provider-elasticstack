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

package syntheticsmonitors

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/syntheticscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const panelType = "synthetics_monitors"

// SchemaAttribute returns the synthetics_monitors_config SingleNestedAttribute definition.
func SchemaAttribute() schema.Attribute {
	attrs := panelkit.PanelPresentationAttributes()
	attrs["view"] = schema.StringAttribute{
		MarkdownDescription: "View mode for the panel. Valid values are `cardView` and `compactView`.",
		Optional:            true,
		Validators: []validator.String{
			stringvalidator.OneOf("cardView", "compactView"),
		},
	}
	attrs["filters"] = syntheticscommon.FilterAttribute(syntheticscommon.FilterAttributeOptions{
		BlockMarkdownDescription: "Optional filter configuration for the Synthetics monitors panel. Omit to show all monitors.",
		ProjectsDescription:      "Filter by project. Each entry has a `label` (display name) and a `value` (project ID).",
		TagsDescription:          "Filter by tags. Each entry has a `label` (display name) and a `value` (tag).",
		MonitorIDsDescription:    "Filter by monitor IDs. Each entry has a `label` (display name) and a `value` (monitor ID). The Kibana API accepts up to 5000 items.",
		LocationsDescription:     "Filter by monitor locations. Each entry has a `label` (display name) and a `value` (location ID).",
		MonitorTypesDescription:  "Filter by monitor types. Each entry has a `label` (display name) and a `value` (monitor type, e.g. `browser`, `http`, `tcp`, `icmp`).",
	})

	return panelkit.PanelConfigBlock(panelkit.PanelConfigBlockOpts{
		Description: "Configuration for a Synthetics monitors panel. Displays a table of Elastic Synthetics monitors " +
			"and their current status. All fields are optional — omit the block entirely for a bare panel with no filtering.",
		BlockName:  "synthetics_monitors_config",
		PanelType:  panelType,
		Attributes: attrs,
	})
}
