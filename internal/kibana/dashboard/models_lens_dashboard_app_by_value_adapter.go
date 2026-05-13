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

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// lensByValueModelHasAnyTypedChartBlock is true when the by_value model was authored
// with a typed chart block (not raw config_json) as the single source.
func lensByValueModelHasAnyTypedChartBlock(m *models.LensDashboardAppByValueModel) bool {
	if m == nil {
		return false
	}
	return m.XYChartConfig != nil ||
		m.TreemapConfig != nil ||
		m.MosaicConfig != nil ||
		m.DatatableConfig != nil ||
		m.TagcloudConfig != nil ||
		m.HeatmapConfig != nil ||
		m.WaffleConfig != nil ||
		m.RegionMapConfig != nil ||
		m.GaugeConfig != nil ||
		m.MetricChartConfig != nil ||
		m.PieChartConfig != nil ||
		m.LegacyMetricConfig != nil
}

// lensByValueChartBlocksForTypedLensApp materializes shared `models.LensByValueChartBlocks` from
// `lens_dashboard_app_config.by_value` fields so vis converters can run (metric expands for the duration).
func lensByValueChartBlocksForTypedLensApp(byValue models.LensDashboardAppByValueModel) (*models.LensByValueChartBlocks, bool) {
	if !lensByValueModelHasAnyTypedChartBlock(&byValue) {
		return nil, false
	}
	var blocks models.LensByValueChartBlocks
	set := 0
	if byValue.XYChartConfig != nil {
		set++
		blocks.XYChartConfig = byValue.XYChartConfig
	}
	if byValue.TreemapConfig != nil {
		set++
		blocks.TreemapConfig = byValue.TreemapConfig
	}
	if byValue.MosaicConfig != nil {
		set++
		blocks.MosaicConfig = byValue.MosaicConfig
	}
	if byValue.DatatableConfig != nil {
		set++
		blocks.DatatableConfig = byValue.DatatableConfig
	}
	if byValue.TagcloudConfig != nil {
		set++
		blocks.TagcloudConfig = byValue.TagcloudConfig
	}
	if byValue.HeatmapConfig != nil {
		set++
		blocks.HeatmapConfig = byValue.HeatmapConfig
	}
	if byValue.WaffleConfig != nil {
		set++
		blocks.WaffleConfig = byValue.WaffleConfig
	}
	if byValue.RegionMapConfig != nil {
		set++
		blocks.RegionMapConfig = byValue.RegionMapConfig
	}
	if byValue.GaugeConfig != nil {
		set++
		blocks.GaugeConfig = byValue.GaugeConfig
	}
	if byValue.MetricChartConfig != nil {
		set++
		blocks.MetricChartConfig = metricChartLensByValueTFExpandToVisMetricChart(byValue.MetricChartConfig)
	}
	if byValue.PieChartConfig != nil {
		set++
		blocks.PieChartConfig = byValue.PieChartConfig
	}
	if byValue.LegacyMetricConfig != nil {
		set++
		blocks.LegacyMetricConfig = byValue.LegacyMetricConfig
	}
	if set != 1 {
		return nil, false
	}
	return &blocks, true
}

// lensByValueModelFromChartBlocksAfterRead maps chart blocks populated by a vis converter into
// lens_dashboard_app_config.by_value (typed block only; config_json left unset).
func lensByValueModelFromChartBlocksAfterRead(blocks *models.LensByValueChartBlocks) (models.LensDashboardAppByValueModel, bool) {
	if blocks == nil {
		return models.LensDashboardAppByValueModel{}, false
	}
	out := models.LensDashboardAppByValueModel{
		ConfigJSON: jsontypes.NewNormalizedNull(),
	}
	set := 0
	if blocks.XYChartConfig != nil {
		set++
		out.XYChartConfig = blocks.XYChartConfig
	}
	if blocks.TreemapConfig != nil {
		set++
		out.TreemapConfig = blocks.TreemapConfig
	}
	if blocks.MosaicConfig != nil {
		set++
		out.MosaicConfig = blocks.MosaicConfig
	}
	if blocks.DatatableConfig != nil {
		set++
		out.DatatableConfig = blocks.DatatableConfig
	}
	if blocks.TagcloudConfig != nil {
		set++
		out.TagcloudConfig = blocks.TagcloudConfig
	}
	if blocks.HeatmapConfig != nil {
		set++
		out.HeatmapConfig = blocks.HeatmapConfig
	}
	if blocks.WaffleConfig != nil {
		set++
		out.WaffleConfig = blocks.WaffleConfig
	}
	if blocks.RegionMapConfig != nil {
		set++
		out.RegionMapConfig = blocks.RegionMapConfig
	}
	if blocks.GaugeConfig != nil {
		set++
		out.GaugeConfig = blocks.GaugeConfig
	}
	if blocks.MetricChartConfig != nil {
		set++
		out.MetricChartConfig = metricLensByValueFromVisFull(blocks.MetricChartConfig)
	}
	if blocks.PieChartConfig != nil {
		set++
		out.PieChartConfig = blocks.PieChartConfig
	}
	if blocks.LegacyMetricConfig != nil {
		set++
		out.LegacyMetricConfig = blocks.LegacyMetricConfig
	}
	if set != 1 {
		return models.LensDashboardAppByValueModel{}, false
	}
	return out, true
}

// visConfig0ToLensAppConfig0 bridges KbnDashboardPanelTypeVisConfig0 to
// KbnDashboardPanelTypeLensDashboardAppConfig0 by JSON because the two unions
// share the same generated inline chart struct variants.
func visConfig0ToLensAppConfig0(vis kbapi.KbnDashboardPanelTypeVisConfig0) (kbapi.KbnDashboardPanelTypeLensDashboardAppConfig0, error) {
	raw, err := json.Marshal(vis)
	if err != nil {
		return kbapi.KbnDashboardPanelTypeLensDashboardAppConfig0{}, err
	}
	var out kbapi.KbnDashboardPanelTypeLensDashboardAppConfig0
	if err := json.Unmarshal(raw, &out); err != nil {
		return kbapi.KbnDashboardPanelTypeLensDashboardAppConfig0{}, err
	}
	return out, nil
}

// lensByValueConfigFromVisConfig0 is the build path: typed by_value chart -> vis
// config union -> JSON -> lens-dashboard-app inline chart union.
func lensByValueConfigFromVisConfig0(vis kbapi.KbnDashboardPanelTypeVisConfig0) (kbapi.KbnDashboardPanelTypeLensDashboardApp_Config, error) {
	lens0, err := visConfig0ToLensAppConfig0(vis)
	if err != nil {
		return kbapi.KbnDashboardPanelTypeLensDashboardApp_Config{}, err
	}
	var out kbapi.KbnDashboardPanelTypeLensDashboardApp_Config
	if err := out.FromKbnDashboardPanelTypeLensDashboardAppConfig0(lens0); err != nil {
		return kbapi.KbnDashboardPanelTypeLensDashboardApp_Config{}, err
	}
	return out, nil
}

// tryPopulateTypedLensByValueFromAPI fills `pm.LensDashboardAppConfig` from API
// config bytes when prior state used a matching typed by-value chart. Returns true
// if state was set; otherwise the caller should fall back to by_value.config_json.
func tryPopulateTypedLensByValueFromAPI(
	ctx context.Context,
	dashboard *models.DashboardModel,
	prior *models.LensDashboardAppConfigModel,
	configBytes []byte,
	pm *models.PanelModel,
	diags *diag.Diagnostics,
) bool {
	if prior == nil || prior.ByValue == nil {
		return false
	}
	priorBlocks, hasTyped := lensByValueChartBlocksForTypedLensApp(*prior.ByValue)
	if !hasTyped {
		return false
	}
	conv, ok := firstLensVisConverterForChartBlocks(priorBlocks)
	if !ok {
		return false
	}
	var vis0 kbapi.KbnDashboardPanelTypeVisConfig0
	if err := json.Unmarshal(configBytes, &vis0); err != nil {
		return false
	}
	if conv.visType() != detectLensVisType(vis0) {
		return false
	}
	var scratch models.LensByValueChartBlocks
	var priorPanel *models.PanelModel
	if prior != nil && prior.ByValue != nil {
		v := *prior.ByValue
		priorPanel = &models.PanelModel{
			LensDashboardAppConfig: &models.LensDashboardAppConfigModel{
				ByValue: &v,
			},
		}
	}
	d := conv.populateFromAttributes(ctx, dashboard, priorPanel, &scratch, vis0)
	if d.HasError() {
		// Intentional: no error diagnostic here; caller falls back to by_value.config_json
		// (REQ-035) without treating typed read as a user failure.
		return false
	}
	if diags != nil {
		diags.Append(d...)
	}
	by, ok2 := lensByValueModelFromChartBlocksAfterRead(&scratch)
	if !ok2 {
		return false
	}
	pm.LensDashboardAppConfig = &models.LensDashboardAppConfigModel{
		ByValue: &by,
	}
	return true
}

// lensByValueToScratchVisPanel maps a typed lens_dashboard_app.by_value chart to a synthetic vis
// panel Model that shares models.LensByValueChartBlocks (tests and parity checks only).
func lensByValueToScratchVisPanel(by models.LensDashboardAppByValueModel) (models.PanelModel, bool) {
	blocks, ok := lensByValueChartBlocksForTypedLensApp(by)
	if !ok {
		return models.PanelModel{}, false
	}
	return models.PanelModel{
		VisConfig: &models.VisConfigModel{
			ByValue: &models.VisByValueModel{LensByValueChartBlocks: *blocks},
		},
	}, true
}

// firstLensVisConverterForPanel resolves the Lens converter for whichever typed chart sits under
// vis_config.by_value or lens_dashboard_app_config.by_value on the panel.
func firstLensVisConverterForPanel(pm models.PanelModel) (lensVisualizationConverter, bool) {
	blocks := lensByValueChartBlocksFromPanel(&pm)
	if blocks == nil {
		return nil, false
	}
	return firstLensVisConverterForChartBlocks(blocks)
}
