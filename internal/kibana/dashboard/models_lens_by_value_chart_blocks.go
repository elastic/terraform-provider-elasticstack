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

// lensByValueChartBlocks holds the shared typed Lens chart pointers for
// vis_config.by_value and lens_dashboard_app_config.by_value (design D3/D10).
// Embedded in both Terraform models so attributes stay at the by_value object root.
type lensByValueChartBlocks struct {
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

func lensByValueChartBlocksFromPanel(pm *panelModel) *lensByValueChartBlocks {
	if pm == nil {
		return nil
	}
	if pm.VisConfig != nil && pm.VisConfig.ByValue != nil {
		return &pm.VisConfig.ByValue.lensByValueChartBlocks
	}
	if pm.LensDashboardAppConfig != nil && pm.LensDashboardAppConfig.ByValue != nil {
		blocks, ok := lensByValueChartBlocksForTypedLensApp(*pm.LensDashboardAppConfig.ByValue)
		if !ok {
			return nil
		}
		return blocks
	}
	return nil
}

func firstLensVisConverterForChartBlocks(blocks *lensByValueChartBlocks) (lensVisualizationConverter, bool) {
	for _, c := range lensVisConverters {
		if c.handlesTFConfigBlocks(blocks) {
			return c, true
		}
	}
	return nil, false
}

// seedWaffleLensByValueChartFromPriorPanel assigns the waffle chart pointer from practitioner plan/state
// into dest before vis read-mapping replaces blocks.WaffleConfig. The waffle converter keeps that pointer as
// `seed` across `populateFromAttributes` so mergeWaffleConfigFromPlanSeed can reconcile Kibana read omissions.
func seedWaffleLensByValueChartFromPriorPanel(dest *lensByValueChartBlocks, prior *panelModel) {
	if dest == nil || prior == nil || prior.VisConfig == nil || prior.VisConfig.ByValue == nil {
		return
	}
	src := &prior.VisConfig.ByValue.lensByValueChartBlocks
	if src.WaffleConfig != nil {
		dest.WaffleConfig = src.WaffleConfig
	}
}
