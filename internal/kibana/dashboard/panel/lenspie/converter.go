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

package lenspie

import (
	"context"
	"maps"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func init() {
	lenscommon.Register(converter{})
}

type converter struct{}

func (converter) VizType() string {
	return string(kbapi.PieNoESQLTypePie)
}

func (converter) HandlesBlocks(blocks *models.LensByValueChartBlocks) bool {
	return blocks != nil && blocks.PieChartConfig != nil
}

func pieChartLegendDefaultObject() types.Object {
	return types.ObjectValueMust(
		map[string]attr.Type{
			"nested":               types.BoolType,
			"size":                 types.StringType,
			"truncate_after_lines": types.Int64Type,
			"visible":              types.StringType,
		},
		map[string]attr.Value{
			"nested":               types.BoolNull(),
			"size":                 types.StringValue("auto"),
			"truncate_after_lines": types.Int64Null(),
			"visible":              types.StringValue("auto"),
		},
	)
}

func (converter) SchemaAttribute() schema.Attribute {
	attrs := maps.Clone(lenscommon.LensChartBaseAttributes())
	attrs["data_source_json"] = schema.StringAttribute{
		MarkdownDescription: "Dataset configuration as JSON. For standard layers, this specifies the data view and query.",
		CustomType:          jsontypes.NormalizedType{},
		Required:            true,
	}
	attrs["donut_hole"] = schema.StringAttribute{
		MarkdownDescription: "Donut hole size: none (pie), s, m, or l.",
		Optional:            true,
		Validators: []validator.String{
			stringvalidator.OneOf("none", "s", "m", "l"),
		},
	}
	attrs["label_position"] = schema.StringAttribute{
		MarkdownDescription: "Position of slice labels: hidden, inside, or outside.",
		Optional:            true,
		Validators: []validator.String{
			stringvalidator.OneOf("hidden", "inside", "outside"),
		},
	}
	attrs["legend"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Optional legend configuration for the pie chart. " +
			"Same shape as treemap and mosaic legends; Terraform `visible` maps to API `visibility`. " +
			"When omitted, the schema default matches typical Kibana legend defaults (size and visibility " +
			"`auto`) so apply/read stay consistent.",
		Optional:   true,
		Computed:   true,
		Default:    objectdefault.StaticValue(pieChartLegendDefaultObject()),
		Attributes: lenscommon.PartitionLegendSchemaAttributes(),
	}
	attrs["query"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Query configuration for filtering data.",
		Optional:            true,
		Attributes:          lenscommon.LensChartFilterSimpleAttributes(),
	}
	attrs["filters"] = schema.ListNestedAttribute{
		MarkdownDescription: "Additional filters to apply to the chart data (maximum 100).",
		Optional:            true,
		NestedObject:        lenscommon.LensChartFilterNestedObject(),
	}
	attrs["metrics"] = schema.ListNestedAttribute{
		MarkdownDescription: "Array of metric configurations (minimum 1).",
		Required:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"config_json": schema.StringAttribute{
					MarkdownDescription: "Metric configuration as JSON.",
					CustomType:          customtypes.NewJSONWithDefaultsType(lenscommon.PopulatePieChartMetricDefaults),
					Required:            true,
				},
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
		},
	}
	attrs["group_by"] = schema.ListNestedAttribute{
		MarkdownDescription: "Array of breakdown dimensions (minimum 1).",
		Optional:            true,
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
		},
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"config_json": schema.StringAttribute{
					MarkdownDescription: "Group by configuration as JSON.",
					CustomType:          customtypes.NewJSONWithDefaultsType(lenscommon.PopulateLensGroupByDefaults),
					Required:            true,
				},
			},
		},
	}
	maps.Copy(attrs, lenscommon.LensChartPresentationAttributes())
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Typed Lens visualization inside `vis_config.by_value`. " +
			"Mutually exclusive with the other chart blocks in the same `by_value` block. " +
			"Shares the attribute shape with `lens_dashboard_app_config.by_value.pie_chart_config`.",
		Optional:   true,
		Attributes: attrs,
	}
}

func (converter) PopulateFromAttributes(ctx context.Context, resolver lenscommon.Resolver, blocks *models.LensByValueChartBlocks, attrs kbapi.KbnDashboardPanelTypeVisConfig0) diag.Diagnostics {
	var prior *models.PieChartConfigModel
	if blocks != nil && blocks.PieChartConfig != nil {
		cpy := *blocks.PieChartConfig
		prior = &cpy
	}
	if blocks == nil {
		var d diag.Diagnostics
		d.AddError("Lens chart blocks missing", "cannot populate pie_chart_config without chart blocks")
		return d
	}
	blocks.PieChartConfig = &models.PieChartConfigModel{}

	if noESQL, err := attrs.AsPieNoESQL(); err == nil && !isPieNoESQLCandidateActuallyESQL(noESQL) {
		return pieChartConfigFromAPINoESQL(ctx, blocks.PieChartConfig, resolver, prior, noESQL)
	}

	esql, err := attrs.AsPieESQL()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return pieChartConfigFromAPIESQL(ctx, blocks.PieChartConfig, resolver, prior, esql)
}

func (converter) BuildAttributes(blocks *models.LensByValueChartBlocks, resolver lenscommon.Resolver) (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	var diags diag.Diagnostics
	if blocks == nil || blocks.PieChartConfig == nil {
		diags.AddError("Pie chart config missing", "pie_chart_config block is required")
		return kbapi.KbnDashboardPanelTypeVisConfig0{}, diags
	}
	configModel := *blocks.PieChartConfig
	return pieChartConfigToAPI(&configModel, resolver)
}

func (converter) AlignStateFromPlan(_ context.Context, plan, state *models.LensByValueChartBlocks) {
	alignPieStateFromPlan(plan, state)
}

func (converter) PopulateJSONDefaults(attrs map[string]any) map[string]any {
	return populatePieLensAttributes(attrs)
}
