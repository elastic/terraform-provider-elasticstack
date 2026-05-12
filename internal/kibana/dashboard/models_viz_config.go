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

package dashboard

// vizByReferenceModel duplicates lensDashboardAppByReferenceModel — both branches use getLensByReferenceAttributes()
// in schema and identical API saved-object linkage fields on read/write (design D3).
type vizByReferenceModel = lensDashboardAppByReferenceModel

// vizByValueModel is Terraform model for viz_config.by_value (12 Lens chart kinds, no nested config_json; design D4).
// Chart conversion helpers under models_<kind>_panel.go still use zombie panel-level panelModel pointers until task 6; field shapes align with this struct for rewiring (Option B in task 5).
type vizByValueModel struct {
	XYChartConfig      *xyChartConfigModel      `tfsdk:"xy_chart_config"`
	TreemapConfig      *treemapConfigModel      `tfsdk:"treemap_config"`
	MosaicConfig       *mosaicConfigModel       `tfsdk:"mosaic_config"`
	DatatableConfig    *datatableConfigModel    `tfsdk:"datatable_config"`
	TagcloudConfig     *tagcloudConfigModel     `tfsdk:"tagcloud_config"`
	HeatmapConfig      *heatmapConfigModel      `tfsdk:"heatmap_config"`
	WaffleConfig       *waffleConfigModel       `tfsdk:"waffle_config"`
	RegionMapConfig    *regionMapConfigModel    `tfsdk:"region_map_config"`
	GaugeConfig        *gaugeConfigModel        `tfsdk:"gauge_config"`
	MetricChartConfig  *metricChartConfigModel  `tfsdk:"metric_chart_config"`
	PieChartConfig     *pieChartConfigModel     `tfsdk:"pie_chart_config"`
	LegacyMetricConfig *legacyMetricConfigModel `tfsdk:"legacy_metric_config"`
}

// vizConfigModel is nested `viz_config` on panels with type vis (design D10; mapPanelFromAPI / toAPI classification in task 6).
type vizConfigModel struct {
	ByValue     *vizByValueModel     `tfsdk:"by_value"`
	ByReference *vizByReferenceModel `tfsdk:"by_reference"`
}
