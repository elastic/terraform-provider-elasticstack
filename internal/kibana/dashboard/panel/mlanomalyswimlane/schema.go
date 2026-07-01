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

package mlanomalyswimlane

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	panelType           = "ml_anomaly_swimlane"
	panelConfigBlock    = panelType + "_config"
	swimlaneTypeOverall = "overall"
	swimlaneTypeViewBy  = "viewBy"
)

// SchemaAttribute returns the ml_anomaly_swimlane_config SingleNestedAttribute definition.
func SchemaAttribute() schema.Attribute {
	attrs := panelkit.PanelPresentationAttributes()
	attrs["swimlane_type"] = schema.StringAttribute{
		MarkdownDescription: "Swim lane mode. Use `overall` for a single aggregate lane or `viewBy` to split anomalies by field.",
		Required:            true,
		Validators: []validator.String{
			stringvalidator.OneOf(swimlaneTypeOverall, swimlaneTypeViewBy),
		},
	}
	attrs["job_ids"] = schema.ListAttribute{
		MarkdownDescription: "IDs of anomaly detection jobs or groups whose results appear in the swim lane. At least one entry is required.",
		Required:            true,
		ElementType:         types.StringType,
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
		},
	}
	attrs["view_by"] = schema.StringAttribute{
		MarkdownDescription: "Field name used to split anomalies into a view-by swim lane. Required when `swimlane_type` is `viewBy`; must not be set when `swimlane_type` is `overall`.",
		Optional:            true,
		Validators: []validator.String{
			validators.RequiredIfDependentPathExpressionOneOf(
				path.MatchRelative().AtParent().AtName("swimlane_type"),
				[]string{swimlaneTypeViewBy},
			),
			validators.ForbiddenIfDependentPathExpressionOneOf(
				path.MatchRelative().AtParent().AtName("swimlane_type"),
				[]string{swimlaneTypeOverall},
			),
		},
	}
	attrs["per_page"] = schema.Float32Attribute{
		MarkdownDescription: "Number of rows to display per page in a view-by swim lane. Ignored for overall swim lanes.",
		Optional:            true,
	}
	attrs["time_range"] = panelkit.TimeRangeSchema(
		"Optional panel-level time range (`from`, `to`, and optional `mode`).",
	)

	return panelkit.PanelConfigBlock(panelkit.PanelConfigBlockOpts{
		Description: "Configuration for an ML anomaly swim lane panel (`kbn-dashboard-panel-type-ml_anomaly_swimlane`). " +
			"Required when `type` is `ml_anomaly_swimlane`.",
		BlockName:  panelConfigBlock,
		PanelType:  panelType,
		Required:   true,
		Attributes: attrs,
	})
}
