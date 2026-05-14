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

package lenswaffle

import (
	"maps"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func partitionChartBaseAttrs(includePresentation bool) map[string]schema.Attribute {
	attrs := lenscommon.LensChartBaseAttributes()
	attrs["data_source_json"] = schema.StringAttribute{
		MarkdownDescription: "Dataset configuration as JSON. For non-ES|QL, this specifies the data view or index; for ES|QL, this specifies the ES|QL query dataset.",
		CustomType:          jsontypes.NormalizedType{},
		Required:            true,
	}
	attrs["query"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Query configuration for filtering data. Required for non-ES|QL partition charts.",
		Optional:            true,
		Attributes:          lenscommon.LensChartFilterSimpleAttributes(),
	}
	if includePresentation {
		maps.Copy(attrs, lenscommon.LensChartPresentationAttributes())
	}
	return attrs
}

// getWaffleSchema returns schema for waffle (grid) Lens chart configuration.
func waffleSchemaAttrs(includePresentation bool) map[string]schema.Attribute {
	attrs := partitionChartBaseAttrs(includePresentation)
	attrs["legend"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Legend configuration for the waffle chart.",
		Required:            true,
		Attributes:          waffleLegendSchemaAttrs(),
	}
	attrs["value_display"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Configuration for displaying values in chart cells.",
		Optional:            true,
		Attributes:          lenscommon.PartitionValueDisplaySchemaAttributes(),
	}
	attrs["metrics"] = schema.ListNestedAttribute{
		MarkdownDescription: "Metric configurations for non-ES|QL waffles (minimum 1). Each `config_json` is a JSON object (e.g. count, sum, or formula) matching the Kibana Lens waffle schema.",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"config_json": schema.StringAttribute{
					MarkdownDescription: "Metric operation as JSON.",
					CustomType:          customtypes.NewJSONWithDefaultsType(lenscommon.PopulatePieChartMetricDefaults),
					Required:            true,
				},
			},
		},
	}
	attrs["group_by"] = schema.ListNestedAttribute{
		MarkdownDescription: "Breakdown dimensions for non-ES|QL waffles. Each `config_json` is a JSON object (terms, date_histogram, etc.) matching the Kibana Lens waffle schema.",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"config_json": schema.StringAttribute{
					MarkdownDescription: "Group-by operation as JSON.",
					CustomType:          customtypes.NewJSONWithDefaultsType(lenscommon.PopulateLensGroupByDefaults),
					Required:            true,
				},
			},
		},
	}
	attrs["esql_metrics"] = schema.ListNestedAttribute{
		MarkdownDescription: "Metric columns for ES|QL waffles (minimum 1). Mutually exclusive with `metrics`.",
		Optional:            true,
		NestedObject:        partitionESQLMetricNested(),
	}
	attrs["esql_group_by"] = schema.ListNestedAttribute{
		MarkdownDescription: "Breakdown columns for ES|QL waffles. Mutually exclusive with `group_by`.",
		Optional:            true,
		NestedObject:        partitionESQLGroupByNested(),
	}
	return attrs
}

// getWaffleLegendSchema returns schema for waffle legend (distinct from XY/heatmap legend).
func waffleLegendSchemaAttrs() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"size": schema.StringAttribute{
			MarkdownDescription: "Legend size: auto, s, m, l, or xl.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("auto", "s", "m", "l", "xl"),
			},
		},
		"truncate_after_lines": schema.Int64Attribute{
			MarkdownDescription: "Maximum lines before truncating legend items (1-10).",
			Optional:            true,
		},
		"values": schema.ListAttribute{
			MarkdownDescription: "Legend value display modes. For example `absolute` shows raw metric values in the legend.",
			ElementType:         types.StringType,
			Optional:            true,
			Validators: []validator.List{
				listvalidator.ValueStringsAre(stringvalidator.OneOf("absolute")),
			},
		},
		"visible": schema.StringAttribute{
			MarkdownDescription: "Legend visibility: auto, visible, or hidden.",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("auto", "visible", "hidden"),
			},
		},
	}
}

// getPartitionESQLMetricSchema returns the shared ES|QL metric schema used by waffle,
// treemap, and mosaic.
func partitionESQLMetricNested() schema.NestedAttributeObject {
	return lenscommon.PartitionESQLMetricNestedObject()
}

// getPartitionESQLGroupBySchema returns the shared ES|QL group-by schema used by waffle,
// treemap, and mosaic.
func partitionESQLGroupByNested() schema.NestedAttributeObject {
	return lenscommon.PartitionESQLGroupByNestedObject()
}
