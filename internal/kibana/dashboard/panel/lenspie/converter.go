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
	"encoding/json"
	"maps"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
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
	return string(kbapi.KibanaHTTPAPIsPieNoESQLTypePie)
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
	attrs := lenscommon.LensChartBaseAttributes()
	attrs["data_source_json"] = lenscommon.DataSourceJSONAttribute(
		"Dataset configuration as JSON. For standard layers, this specifies the data view and query.",
	)
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
	attrs["query"] = lenscommon.QueryAttribute("Query configuration for filtering data.")
	attrs["metrics"] = lenscommon.JSONConfigItemList(
		"Array of metric configurations (minimum 1).",
		"Metric configuration as JSON.",
		lenscommon.PopulatePieChartMetricDefaults, true,
		listvalidator.SizeAtLeast(1),
	)
	attrs["group_by"] = lenscommon.JSONConfigItemList(
		"Array of breakdown dimensions (minimum 1).",
		"Group by configuration as JSON.",
		lenscommon.PopulateLensGroupByDefaults, false,
		listvalidator.SizeAtLeast(1),
	)
	maps.Copy(attrs, lenscommon.LensChartPresentationAttributes())
	return lenscommon.ByValueChartNestedAttribute("pie_chart_config", attrs)
}

func (converter) PopulateFromAttributes(ctx context.Context, blocks *models.LensByValueChartBlocks, attrs lenscommon.LensByValueConfig) diag.Diagnostics {
	if diags := lenscommon.ValidateLensBlocks(blocks, "pie_chart_config"); diags.HasError() {
		return diags
	}
	prior := lenscommon.SnapshotAndResetBlock(&blocks.PieChartConfig)
	diags := lenscommon.PopulateFromNoESQLOrESQL(
		ctx, blocks.PieChartConfig, prior,
		func() (kbapi.KibanaHTTPAPIsPieNoESQL, error) {
			var chart kbapi.KibanaHTTPAPIsPieNoESQL
			return chart, pieChartFromLensAPI(attrs.Chart, &chart)
		},
		func() (kbapi.KibanaHTTPAPIsPieESQL, error) {
			var chart kbapi.KibanaHTTPAPIsPieESQL
			return chart, pieChartFromLensAPI(attrs.Chart, &chart)
		},
		func(v kbapi.KibanaHTTPAPIsPieNoESQL) bool {
			return !lenscommon.IsNoESQLCandidateActuallyESQL(v.DataSource)
		},
		pieChartConfigFromAPINoESQL,
		pieChartConfigFromAPIESQL,
	)
	if diags.HasError() {
		return diags
	}
	if !lenscommon.PopulateLensChartPresentation(
		ctx, &blocks.PieChartConfig.LensChartPresentationTFModel, prior,
		attrs.Presentation.TimeRange, attrs.Presentation.HideTitle, attrs.Presentation.HideBorder,
		attrs.Presentation.References, attrs.Presentation.Drilldowns, &diags,
	) {
		return diags
	}
	diags.Append(populatePieRawDimensions(blocks.PieChartConfig, attrs.Chart)...)
	return diags
}

func (converter) BuildAttributes(blocks *models.LensByValueChartBlocks) (lenscommon.LensByValueConfig, diag.Diagnostics) {
	var attrs lenscommon.LensByValueConfig
	var diags diag.Diagnostics
	if blocks == nil {
		return attrs, diags
	}
	return pieChartConfigToAPI(blocks.PieChartConfig)
}

func (converter) AlignStateFromPlan(ctx context.Context, plan, state *models.LensByValueChartBlocks) {
	alignPieStateFromPlan(ctx, plan, state)
}

func (converter) PopulateJSONDefaults(attrs map[string]any) map[string]any {
	return populatePieLensAttributes(attrs)
}

func populatePieRawDimensions(m *models.PieChartConfigModel, chart kbapi.KibanaHTTPAPIsLensApiConfig) diag.Diagnostics {
	var diags diag.Diagnostics
	raw, err := json.Marshal(chart)
	if err != nil {
		diags.AddError("Failed to marshal pie chart", err.Error())
		return diags
	}
	var payload struct {
		Metrics []json.RawMessage `json:"metrics"`
		GroupBy []json.RawMessage `json:"group_by"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		diags.AddError("Failed to decode pie chart dimensions", err.Error())
		return diags
	}
	if payload.Metrics != nil {
		m.Metrics = make([]models.PieMetricModel, len(payload.Metrics))
		for i, metric := range payload.Metrics {
			m.Metrics[i].Config = customtypes.NewJSONWithDefaultsValue(string(metric), lenscommon.PopulatePieChartMetricDefaults)
		}
	}
	if payload.GroupBy != nil {
		m.GroupBy = make([]models.PieGroupByModel, len(payload.GroupBy))
		for i, groupBy := range payload.GroupBy {
			m.GroupBy[i].Config = customtypes.NewJSONWithDefaultsValue(string(groupBy), lenscommon.PopulateLensGroupByDefaults)
		}
	}
	return diags
}
