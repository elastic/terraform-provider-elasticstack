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

package models

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type WaffleConfigModel struct {
	LensChartPresentationTFModel
	LensChartBaseTFModel
	Query        *FilterSimpleModel          `tfsdk:"query"`
	Legend       *WaffleLegendModel          `tfsdk:"legend"`
	ValueDisplay *PartitionValueDisplay      `tfsdk:"value_display"`
	Metrics      []WaffleDSLMetric           `tfsdk:"metrics"`
	GroupBy      []WaffleDSLGroupBy          `tfsdk:"group_by"`
	EsqlMetrics  []WaffleEsqlMetric          `tfsdk:"esql_metrics"`
	EsqlGroupBy  []PartitionEsqlGroupByModel `tfsdk:"esql_group_by"`
}

type WaffleDSLMetric struct {
	Config customtypes.JSONWithDefaultsValue[map[string]any] `tfsdk:"config_json"`
}

type WaffleDSLGroupBy struct {
	Config customtypes.JSONWithDefaultsValue[map[string]any] `tfsdk:"config_json"`
}

type WaffleLegendModel struct {
	Size               types.String `tfsdk:"size"`
	TruncateAfterLines types.Int64  `tfsdk:"truncate_after_lines"`
	Values             types.List   `tfsdk:"values"`
	Visible            types.String `tfsdk:"visible"`
}

type WaffleEsqlMetric struct {
	Column     types.String          `tfsdk:"column"`
	Label      types.String          `tfsdk:"label"`
	FormatJSON jsontypes.Normalized  `tfsdk:"format_json"`
	Color      *LensStaticColorModel `tfsdk:"color"`
}
