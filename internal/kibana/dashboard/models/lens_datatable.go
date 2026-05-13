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
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DatatableConfigModel struct {
	NoESQL *DatatableNoESQLConfigModel `tfsdk:"no_esql"`
	ESQL   *DatatableESQLConfigModel   `tfsdk:"esql"`
}

type DatatableNoESQLConfigModel struct {
	LensChartPresentationTFModel
	Title               types.String            `tfsdk:"title"`
	Description         types.String            `tfsdk:"description"`
	DataSourceJSON      jsontypes.Normalized    `tfsdk:"data_source_json"`
	Styling             *DatatableStylingModel  `tfsdk:"styling"`
	IgnoreGlobalFilters types.Bool              `tfsdk:"ignore_global_filters"`
	Sampling            types.Float64           `tfsdk:"sampling"`
	Query               *FilterSimpleModel      `tfsdk:"query"`
	Filters             []ChartFilterJSONModel  `tfsdk:"filters"`
	Metrics             []DatatableMetricModel  `tfsdk:"metrics"`
	Rows                []DatatableRowModel     `tfsdk:"rows"`
	SplitMetricsBy      []DatatableSplitByModel `tfsdk:"split_metrics_by"`
}

type DatatableESQLConfigModel struct {
	LensChartPresentationTFModel
	Title               types.String            `tfsdk:"title"`
	Description         types.String            `tfsdk:"description"`
	DataSourceJSON      jsontypes.Normalized    `tfsdk:"data_source_json"`
	Styling             *DatatableStylingModel  `tfsdk:"styling"`
	IgnoreGlobalFilters types.Bool              `tfsdk:"ignore_global_filters"`
	Sampling            types.Float64           `tfsdk:"sampling"`
	Filters             []ChartFilterJSONModel  `tfsdk:"filters"`
	Metrics             []DatatableMetricModel  `tfsdk:"metrics"`
	Rows                []DatatableRowModel     `tfsdk:"rows"`
	SplitMetricsBy      []DatatableSplitByModel `tfsdk:"split_metrics_by"`
}

type DatatableStylingModel struct {
	Density    *DatatableDensityModel `tfsdk:"density"`
	SortByJSON jsontypes.Normalized   `tfsdk:"sort_by_json"`
	Paging     types.Int64            `tfsdk:"paging"`
}

type DatatableMetricModel struct {
	ConfigJSON jsontypes.Normalized `tfsdk:"config_json"`
}

type DatatableRowModel struct {
	ConfigJSON jsontypes.Normalized `tfsdk:"config_json"`
}

type DatatableSplitByModel struct {
	ConfigJSON jsontypes.Normalized `tfsdk:"config_json"`
}

type DatatableDensityModel struct {
	Mode   types.String                 `tfsdk:"mode"`
	Height *DatatableDensityHeightModel `tfsdk:"height"`
}

type DatatableDensityHeightModel struct {
	Header *DatatableDensityHeightHeaderModel `tfsdk:"header"`
	Value  *DatatableDensityHeightValueModel  `tfsdk:"value"`
}

type DatatableDensityHeightHeaderModel struct {
	Type     types.String  `tfsdk:"type"`
	MaxLines types.Float64 `tfsdk:"max_lines"`
}

type DatatableDensityHeightValueModel struct {
	Type  types.String  `tfsdk:"type"`
	Lines types.Float64 `tfsdk:"lines"`
}
