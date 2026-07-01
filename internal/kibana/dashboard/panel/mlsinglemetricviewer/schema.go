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

package mlsinglemetricviewer

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	panelType        = "ml_single_metric_viewer"
	panelConfigBlock = panelType + "_config"

	entityAttrStringValue  = "string_value"
	entityAttrNumericValue = "numeric_value"
)

// SchemaAttribute returns the ml_single_metric_viewer_config SingleNestedAttribute definition.
func SchemaAttribute() schema.Attribute {
	attrs := panelkit.PanelPresentationAttributes()
	attrs["job_ids"] = schema.ListAttribute{
		MarkdownDescription: "Anomaly detection job ID whose results appear in the single metric viewer. Exactly one entry is required.",
		Required:            true,
		ElementType:         types.StringType,
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
			listvalidator.SizeAtMost(1),
		},
	}
	attrs["selected_detector_index"] = schema.Float32Attribute{
		MarkdownDescription: "Zero-based index of the detector within the job whose results are shown.",
		Optional:            true,
	}
	attrs["forecast_id"] = schema.StringAttribute{
		MarkdownDescription: "Forecast identifier to overlay on the chart.",
		Optional:            true,
	}
	attrs["function_description"] = schema.StringAttribute{
		MarkdownDescription: "For `metric` detectors, selects which value to plot: `min`, `max`, or `mean`. Ignored for other detector functions.",
		Optional:            true,
		Validators: []validator.String{
			stringvalidator.OneOf("min", "max", "mean"),
		},
	}
	attrs["selected_entities"] = schema.MapNestedAttribute{
		MarkdownDescription: "Values of partition, by, or over fields that identify the single time series to display. Each map entry must set exactly one of `string_value` or `numeric_value`.",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				entityAttrStringValue: schema.StringAttribute{
					MarkdownDescription: "String entity value for the field.",
					Optional:            true,
				},
				entityAttrNumericValue: schema.NumberAttribute{
					MarkdownDescription: "Numeric entity value for the field.",
					Optional:            true,
				},
			},
			Validators: []validator.Object{
				validators.ExactlyOneOfNestedAttrsValidator(validators.ExactlyOneOfNestedAttrsOpts{
					AttrNames:     []string{entityAttrStringValue, entityAttrNumericValue},
					Summary:       "Invalid selected_entities entry",
					MissingDetail: "Exactly one of `string_value` or `numeric_value` must be set for each `selected_entities` entry.",
					TooManyDetail: "Exactly one of `string_value` or `numeric_value` must be set for each `selected_entities` entry, not both.",
					Description:   "Ensures exactly one of `string_value` or `numeric_value` is set for each `selected_entities` entry.",
				}),
			},
		},
	}
	attrs["time_range"] = panelkit.TimeRangeSchema(
		"Optional panel-level time range (`from`, `to`, and optional `mode`).",
	)

	return panelkit.PanelConfigBlock(panelkit.PanelConfigBlockOpts{
		Description: "Configuration for an ML single metric viewer panel (`kbn-dashboard-panel-type-ml_single_metric_viewer`). " +
			"Required when `type` is `ml_single_metric_viewer`.",
		BlockName:  panelConfigBlock,
		PanelType:  panelType,
		Required:   true,
		Attributes: attrs,
	})
}
