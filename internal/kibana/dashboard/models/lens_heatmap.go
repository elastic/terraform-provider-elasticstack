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

type HeatmapConfigModel struct {
	LensChartPresentationTFModel
	Title               types.String                                      `tfsdk:"title"`
	Description         types.String                                      `tfsdk:"description"`
	DataSourceJSON      jsontypes.Normalized                              `tfsdk:"data_source_json"`
	IgnoreGlobalFilters types.Bool                                        `tfsdk:"ignore_global_filters"`
	Sampling            types.Float64                                     `tfsdk:"sampling"`
	Query               *FilterSimpleModel                                `tfsdk:"query"`
	Filters             []ChartFilterJSONModel                            `tfsdk:"filters"`
	Axis                *HeatmapAxesModel                                 `tfsdk:"axis"`
	Styling             *HeatmapStylingModel                              `tfsdk:"styling"`
	Legend              *HeatmapLegendModel                               `tfsdk:"legend"`
	MetricJSON          customtypes.JSONWithDefaultsValue[map[string]any] `tfsdk:"metric_json"`
	XAxisJSON           jsontypes.Normalized                              `tfsdk:"x_axis_json"`
	YAxisJSON           jsontypes.Normalized                              `tfsdk:"y_axis_json"`
}

type HeatmapStylingModel struct {
	Cells *HeatmapCellsModel `tfsdk:"cells"`
}

type HeatmapAxesModel struct {
	X *HeatmapXAxisModel `tfsdk:"x"`
	Y *HeatmapYAxisModel `tfsdk:"y"`
}

type HeatmapXAxisModel struct {
	Labels *HeatmapXAxisLabelsModel `tfsdk:"labels"`
	Title  *AxisTitleModel          `tfsdk:"title"`
}

type HeatmapXAxisLabelsModel struct {
	Orientation types.String `tfsdk:"orientation"`
	Visible     types.Bool   `tfsdk:"visible"`
}

type HeatmapYAxisModel struct {
	Labels *HeatmapYAxisLabelsModel `tfsdk:"labels"`
	Title  *AxisTitleModel          `tfsdk:"title"`
}

type HeatmapYAxisLabelsModel struct {
	Visible types.Bool `tfsdk:"visible"`
}

type HeatmapCellsModel struct {
	Labels *HeatmapCellsLabelsModel `tfsdk:"labels"`
}

type HeatmapCellsLabelsModel struct {
	Visible types.Bool `tfsdk:"visible"`
}

type HeatmapLegendModel struct {
	Visibility         types.String `tfsdk:"visibility"`
	Size               types.String `tfsdk:"size"`
	TruncateAfterLines types.Int64  `tfsdk:"truncate_after_lines"`
}
