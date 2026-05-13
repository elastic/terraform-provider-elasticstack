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

type XYLayerModel struct {
	Type               types.String             `tfsdk:"type"`
	DataLayer          *DataLayerModel          `tfsdk:"data_layer"`
	ReferenceLineLayer *ReferenceLineLayerModel `tfsdk:"reference_line_layer"`
}

type DataLayerModel struct {
	DataSourceJSON      jsontypes.Normalized `tfsdk:"data_source_json"`
	IgnoreGlobalFilters types.Bool           `tfsdk:"ignore_global_filters"`
	Sampling            types.Float64        `tfsdk:"sampling"`
	XJSON               jsontypes.Normalized `tfsdk:"x_json"`
	Y                   []YMetricModel       `tfsdk:"y"`
	BreakdownByJSON     jsontypes.Normalized `tfsdk:"breakdown_by_json"`
}

type YMetricModel struct {
	ConfigJSON jsontypes.Normalized `tfsdk:"config_json"`
}

type ReferenceLineLayerModel struct {
	DataSourceJSON      jsontypes.Normalized `tfsdk:"data_source_json"`
	IgnoreGlobalFilters types.Bool           `tfsdk:"ignore_global_filters"`
	Sampling            types.Float64        `tfsdk:"sampling"`
	Thresholds          []ThresholdModel     `tfsdk:"thresholds"`
}

type ThresholdModel struct {
	Axis        types.String         `tfsdk:"axis"`
	ColorJSON   jsontypes.Normalized `tfsdk:"color_json"`
	Column      types.String         `tfsdk:"column"`
	ValueJSON   jsontypes.Normalized `tfsdk:"value_json"`
	Fill        types.String         `tfsdk:"fill"`
	Icon        types.String         `tfsdk:"icon"`
	Operation   types.String         `tfsdk:"operation"`
	StrokeDash  types.String         `tfsdk:"stroke_dash"`
	StrokeWidth types.Float64        `tfsdk:"stroke_width"`
	Text        types.String         `tfsdk:"text"`
}

type XYChartConfigModel struct {
	LensChartPresentationTFModel
	Title       types.String           `tfsdk:"title"`
	Description types.String           `tfsdk:"description"`
	Axis        *XYAxisModel           `tfsdk:"axis"`
	Decorations *XYDecorationsModel    `tfsdk:"decorations"`
	Fitting     *XYFittingModel        `tfsdk:"fitting"`
	Layers      []XYLayerModel         `tfsdk:"layers"`
	Legend      *XYLegendModel         `tfsdk:"legend"`
	Query       *FilterSimpleModel     `tfsdk:"query"`
	Filters     []ChartFilterJSONModel `tfsdk:"filters"`
}

type XYAxisModel struct {
	X  *XYAxisConfigModel `tfsdk:"x"`
	Y  *YAxisConfigModel  `tfsdk:"y"`
	Y2 *YAxisConfigModel  `tfsdk:"y2"`
}

type XYAxisConfigModel struct {
	Title            *AxisTitleModel      `tfsdk:"title"`
	Ticks            types.Bool           `tfsdk:"ticks"`
	Grid             types.Bool           `tfsdk:"grid"`
	LabelOrientation types.String         `tfsdk:"label_orientation"`
	Scale            types.String         `tfsdk:"scale"`
	DomainJSON       jsontypes.Normalized `tfsdk:"domain_json"`
}

type YAxisConfigModel struct {
	Title            *AxisTitleModel      `tfsdk:"title"`
	Ticks            types.Bool           `tfsdk:"ticks"`
	Grid             types.Bool           `tfsdk:"grid"`
	LabelOrientation types.String         `tfsdk:"label_orientation"`
	Scale            types.String         `tfsdk:"scale"`
	DomainJSON       jsontypes.Normalized `tfsdk:"domain_json"`
}

type AxisTitleModel struct {
	Value   types.String `tfsdk:"value"`
	Visible types.Bool   `tfsdk:"visible"`
}

type XYDecorationsModel struct {
	ShowEndZones          types.Bool    `tfsdk:"show_end_zones"`
	ShowCurrentTimeMarker types.Bool    `tfsdk:"show_current_time_marker"`
	PointVisibility       types.String  `tfsdk:"point_visibility"`
	LineInterpolation     types.String  `tfsdk:"line_interpolation"`
	MinimumBarHeight      types.Int64   `tfsdk:"minimum_bar_height"`
	ShowValueLabels       types.Bool    `tfsdk:"show_value_labels"`
	FillOpacity           types.Float64 `tfsdk:"fill_opacity"`
}

type XYFittingModel struct {
	Type     types.String `tfsdk:"type"`
	Dotted   types.Bool   `tfsdk:"dotted"`
	EndValue types.String `tfsdk:"end_value"`
}

type XYLegendModel struct {
	Visibility         types.String `tfsdk:"visibility"`
	Statistics         types.List   `tfsdk:"statistics"`
	TruncateAfterLines types.Int64  `tfsdk:"truncate_after_lines"`
	Inside             types.Bool   `tfsdk:"inside"`
	Position           types.String `tfsdk:"position"`
	Size               types.String `tfsdk:"size"`
	Columns            types.Int64  `tfsdk:"columns"`
	Alignment          types.String `tfsdk:"alignment"`
}
