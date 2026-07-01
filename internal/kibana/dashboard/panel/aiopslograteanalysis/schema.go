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

package aiopslograteanalysis

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

const panelType = "aiops_log_rate_analysis"

// SchemaAttribute returns the aiops_log_rate_analysis_config SingleNestedAttribute definition.
func SchemaAttribute() schema.Attribute {
	attrs := panelkit.PanelPresentationAttributes()
	attrs["data_view_id"] = schema.StringAttribute{
		MarkdownDescription: "The data view ID used to run log rate analysis.",
		Required:            true,
	}
	attrs["time_range"] = panelkit.TimeRangeSchema(
		"Optional panel time range (`from`, `to`, optional `mode`). When omitted, the panel inherits the dashboard `time_range` and this attribute stays null in state (REQ-009).",
	)

	return panelkit.PanelConfigBlock(panelkit.PanelConfigBlockOpts{
		Description: "Configuration for an AIOps log rate analysis panel. Anchored to a data view; " +
			"the remaining fields are the standard optional panel presentation passthroughs.",
		BlockName:  "aiops_log_rate_analysis_config",
		PanelType:  panelType,
		Required:   true,
		Attributes: attrs,
	})
}
