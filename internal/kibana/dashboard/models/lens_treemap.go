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

type TreemapConfigModel struct {
	LensChartPresentationTFModel
	Title               types.String                                        `tfsdk:"title"`
	Description         types.String                                        `tfsdk:"description"`
	DataSourceJSON      jsontypes.Normalized                                `tfsdk:"data_source_json"`
	IgnoreGlobalFilters types.Bool                                          `tfsdk:"ignore_global_filters"`
	Sampling            types.Float64                                       `tfsdk:"sampling"`
	Query               *FilterSimpleModel                                  `tfsdk:"query"`
	Filters             []ChartFilterJSONModel                              `tfsdk:"filters"`
	GroupBy             customtypes.JSONWithDefaultsValue[[]map[string]any] `tfsdk:"group_by_json"`
	Metrics             customtypes.JSONWithDefaultsValue[[]map[string]any] `tfsdk:"metrics_json"`
	Legend              *PartitionLegendModel                               `tfsdk:"legend"`
	ValueDisplay        *PartitionValueDisplay                              `tfsdk:"value_display"`
	EsqlMetrics         []TreemapEsqlMetric                                 `tfsdk:"esql_metrics"`
	EsqlGroupBy         []TreemapEsqlGroupBy                                `tfsdk:"esql_group_by"`
}

type TreemapEsqlMetric struct {
	Column     types.String            `tfsdk:"column"`
	Label      types.String            `tfsdk:"label"`
	FormatJSON jsontypes.Normalized    `tfsdk:"format_json"`
	Color      *TreemapEsqlMetricColor `tfsdk:"color"`
}

type TreemapEsqlMetricColor struct {
	Type  types.String `tfsdk:"type"`
	Color types.String `tfsdk:"color"`
}

type TreemapEsqlGroupBy struct {
	Column     types.String         `tfsdk:"column"`
	CollapseBy types.String         `tfsdk:"collapse_by"`
	ColorJSON  jsontypes.Normalized `tfsdk:"color_json"`
	FormatJSON jsontypes.Normalized `tfsdk:"format_json"`
	Label      types.String         `tfsdk:"label"`
}
