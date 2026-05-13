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

type URLDrilldownModel struct {
	URL          types.String `tfsdk:"url"`
	Label        types.String `tfsdk:"label"`
	EncodeURL    types.Bool   `tfsdk:"encode_url"`
	OpenInNewTab types.Bool   `tfsdk:"open_in_new_tab"`
}

type DrilldownsModel []DrilldownItemModel

type DrilldownItemModel struct {
	Dashboard *DrilldownDashboardBlockModel `tfsdk:"dashboard"`
	Discover  *DrilldownDiscoverBlockModel  `tfsdk:"discover"`
	URL       *DrilldownURLBlockModel       `tfsdk:"url"`
}

type DrilldownDashboardBlockModel struct {
	DashboardID  types.String `tfsdk:"dashboard_id"`
	Label        types.String `tfsdk:"label"`
	UseFilters   types.Bool   `tfsdk:"use_filters"`
	UseTimeRange types.Bool   `tfsdk:"use_time_range"`
	OpenInNewTab types.Bool   `tfsdk:"open_in_new_tab"`
}

type DrilldownDiscoverBlockModel struct {
	Label        types.String `tfsdk:"label"`
	OpenInNewTab types.Bool   `tfsdk:"open_in_new_tab"`
}

type DrilldownURLBlockModel struct {
	URL          types.String `tfsdk:"url"`
	Label        types.String `tfsdk:"label"`
	Trigger      types.String `tfsdk:"trigger"`
	EncodeURL    types.Bool   `tfsdk:"encode_url"`
	OpenInNewTab types.Bool   `tfsdk:"open_in_new_tab"`
}

type LensChartPresentationTFModel struct {
	TimeRange      *TimeRangeModel            `tfsdk:"time_range"`
	HideTitle      types.Bool                 `tfsdk:"hide_title"`
	HideBorder     types.Bool                 `tfsdk:"hide_border"`
	ReferencesJSON jsontypes.Normalized       `tfsdk:"references_json"`
	Drilldowns     []LensDrilldownItemTFModel `tfsdk:"drilldowns"`
}

type LensDrilldownItemTFModel struct {
	DashboardDrilldown *LensDashboardDrilldownTFModel `tfsdk:"dashboard_drilldown"`
	DiscoverDrilldown  *LensDiscoverDrilldownTFModel  `tfsdk:"discover_drilldown"`
	URLDrilldown       *LensURLDrilldownTFModel       `tfsdk:"url_drilldown"`
}

type LensDashboardDrilldownTFModel struct {
	DashboardID  types.String `tfsdk:"dashboard_id"`
	Label        types.String `tfsdk:"label"`
	Trigger      types.String `tfsdk:"trigger"`
	UseFilters   types.Bool   `tfsdk:"use_filters"`
	UseTimeRange types.Bool   `tfsdk:"use_time_range"`
	OpenInNewTab types.Bool   `tfsdk:"open_in_new_tab"`
}

type LensDiscoverDrilldownTFModel struct {
	Label        types.String `tfsdk:"label"`
	Trigger      types.String `tfsdk:"trigger"`
	OpenInNewTab types.Bool   `tfsdk:"open_in_new_tab"`
}

type LensURLDrilldownTFModel struct {
	URL          types.String `tfsdk:"url"`
	Label        types.String `tfsdk:"label"`
	Trigger      types.String `tfsdk:"trigger"`
	EncodeURL    types.Bool   `tfsdk:"encode_url"`
	OpenInNewTab types.Bool   `tfsdk:"open_in_new_tab"`
}

type LensByValueChartBlocks struct {
	XYChartConfig      *XYChartConfigModel      `tfsdk:"xy_chart_config"`
	TreemapConfig      *TreemapConfigModel      `tfsdk:"treemap_config"`
	MosaicConfig       *MosaicConfigModel       `tfsdk:"mosaic_config"`
	DatatableConfig    *DatatableConfigModel    `tfsdk:"datatable_config"`
	TagcloudConfig     *TagcloudConfigModel     `tfsdk:"tagcloud_config"`
	HeatmapConfig      *HeatmapConfigModel      `tfsdk:"heatmap_config"`
	WaffleConfig       *WaffleConfigModel       `tfsdk:"waffle_config"`
	RegionMapConfig    *RegionMapConfigModel    `tfsdk:"region_map_config"`
	GaugeConfig        *GaugeConfigModel        `tfsdk:"gauge_config"`
	MetricChartConfig  *MetricChartConfigModel  `tfsdk:"metric_chart_config"`
	PieChartConfig     *PieChartConfigModel     `tfsdk:"pie_chart_config"`
	LegacyMetricConfig *LegacyMetricConfigModel `tfsdk:"legacy_metric_config"`
}

type LensDashboardAppConfigModel struct {
	ByValue     *LensDashboardAppByValueModel     `tfsdk:"by_value"`
	ByReference *LensDashboardAppByReferenceModel `tfsdk:"by_reference"`
}

type LensDashboardAppByValueModel struct {
	ConfigJSON         jsontypes.Normalized           `tfsdk:"config_json"`
	XYChartConfig      *XYChartConfigModel            `tfsdk:"xy_chart_config"`
	TreemapConfig      *TreemapConfigModel            `tfsdk:"treemap_config"`
	MosaicConfig       *MosaicConfigModel             `tfsdk:"mosaic_config"`
	DatatableConfig    *DatatableConfigModel          `tfsdk:"datatable_config"`
	TagcloudConfig     *TagcloudConfigModel           `tfsdk:"tagcloud_config"`
	HeatmapConfig      *HeatmapConfigModel            `tfsdk:"heatmap_config"`
	WaffleConfig       *WaffleConfigModel             `tfsdk:"waffle_config"`
	RegionMapConfig    *RegionMapConfigModel          `tfsdk:"region_map_config"`
	GaugeConfig        *GaugeConfigModel              `tfsdk:"gauge_config"`
	MetricChartConfig  *MetricChartLensByValueTFModel `tfsdk:"metric_chart_config"`
	PieChartConfig     *PieChartConfigModel           `tfsdk:"pie_chart_config"`
	LegacyMetricConfig *LegacyMetricConfigModel       `tfsdk:"legacy_metric_config"`
}

type LensDashboardAppByReferenceModel struct {
	RefID          types.String                   `tfsdk:"ref_id"`
	ReferencesJSON jsontypes.Normalized           `tfsdk:"references_json"`
	Title          types.String                   `tfsdk:"title"`
	Description    types.String                   `tfsdk:"description"`
	HideTitle      types.Bool                     `tfsdk:"hide_title"`
	HideBorder     types.Bool                     `tfsdk:"hide_border"`
	Drilldowns     DrilldownsModel                `tfsdk:"drilldowns"`
	TimeRange      LensDashboardAppTimeRangeModel `tfsdk:"time_range"`
}

type LensDashboardAppTimeRangeModel struct {
	From types.String `tfsdk:"from"`
	To   types.String `tfsdk:"to"`
	Mode types.String `tfsdk:"mode"`
}

type VisByReferenceModel = LensDashboardAppByReferenceModel

type VisConfigModel struct {
	ByValue     *VisByValueModel     `tfsdk:"by_value"`
	ByReference *VisByReferenceModel `tfsdk:"by_reference"`
}

type VisByValueModel struct {
	LensByValueChartBlocks
}
