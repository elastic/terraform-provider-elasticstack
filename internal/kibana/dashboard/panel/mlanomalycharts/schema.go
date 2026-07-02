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

package mlanomalycharts

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
	panelType        = "ml_anomaly_charts"
	panelConfigBlock = panelType + "_config"

	severityLow      = "low"
	severityWarning  = "warning"
	severityMinor    = "minor"
	severityMajor    = "major"
	severityCritical = "critical"
)

var severityEnumValues = []string{
	severityLow,
	severityWarning,
	severityMinor,
	severityMajor,
	severityCritical,
}

// SchemaAttribute returns the ml_anomaly_charts_config SingleNestedAttribute definition.
func SchemaAttribute() schema.Attribute {
	attrs := panelkit.PanelPresentationAttributes()
	attrs["job_ids"] = schema.ListAttribute{
		MarkdownDescription: "Anomaly detection job IDs or group IDs whose results appear in the charts. At least one entry is required.",
		Required:            true,
		ElementType:         types.StringType,
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
		},
	}
	attrs["max_series_to_plot"] = schema.Int64Attribute{
		MarkdownDescription: "Maximum number of anomaly series to plot.",
		Optional:            true,
	}
	attrs["severity_threshold"] = schema.ListNestedAttribute{
		MarkdownDescription: "Severity bands to display. Each item sets either a named `severity` shortcut or a raw numeric `min`/`max` range, never both.",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"severity": schema.StringAttribute{
					MarkdownDescription: "Named severity shortcut (`low`, `warning`, `minor`, `major`, `critical`).",
					Optional:            true,
					Validators: []validator.String{
						stringvalidator.OneOf(severityEnumValues...),
					},
				},
				"min": schema.Int64Attribute{
					MarkdownDescription: "Lower bound of a raw severity range. Required when `severity` is omitted.",
					Optional:            true,
				},
				"max": schema.Int64Attribute{
					MarkdownDescription: "Upper bound of a raw severity range. Valid only with `min` when `severity` is unset.",
					Optional:            true,
					Validators: []validator.Int64{
						validators.ForbiddenIfDependentPathExpressionOneOf(
							path.MatchRelative().AtParent().AtName("severity"),
							severityEnumValues,
						),
					},
				},
			},
			Validators: []validator.Object{
				validators.ExactlyOneOfNestedAttrsValidator(validators.ExactlyOneOfNestedAttrsOpts{
					AttrNames:     []string{"severity", "min"},
					Summary:       "Invalid severity_threshold entry",
					MissingDetail: "Exactly one of `severity` or `min` must be set for each `severity_threshold` entry.",
					TooManyDetail: "Exactly one of `severity` or `min` must be set for each `severity_threshold` entry, not both.",
					Description:   "Ensures exactly one of `severity` or `min` is set for each `severity_threshold` entry.",
				}),
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
		},
	}
	attrs["time_range"] = panelkit.TimeRangeSchema(
		"Optional panel-level time range (`from`, `to`, and optional `mode`).",
	)

	return panelkit.PanelConfigBlock(panelkit.PanelConfigBlockOpts{
		Description: "Configuration for an ML anomaly charts panel (`kbn-dashboard-panel-type-ml_anomaly_charts`). " +
			"Required when `type` is `ml_anomaly_charts`.",
		BlockName:  panelConfigBlock,
		PanelType:  panelType,
		Required:   true,
		Attributes: attrs,
	})
}
