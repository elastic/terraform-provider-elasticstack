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
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// lensByValueModelHasAnyTypedChartBlock is true when the by_value model was authored
// with a typed chart block (not raw config_json) as the single source.
func lensByValueModelHasAnyTypedChartBlock(m *lensDashboardAppByValueModel) bool {
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

// lensByValueToScratchVisPanel copies the non-nil typed chart from by_value onto a
// scratch top-level `panelModel` so existing vis `lensVisualizationConverter` code can run.
func lensByValueToScratchVisPanel(byValue lensDashboardAppByValueModel) (panelModel, bool) {
	var pm panelModel
	switch {
	case byValue.XYChartConfig != nil:
		pm.XYChartConfig = byValue.XYChartConfig
	case byValue.TreemapConfig != nil:
		pm.TreemapConfig = byValue.TreemapConfig
	case byValue.MosaicConfig != nil:
		pm.MosaicConfig = byValue.MosaicConfig
	case byValue.DatatableConfig != nil:
		pm.DatatableConfig = byValue.DatatableConfig
	case byValue.TagcloudConfig != nil:
		pm.TagcloudConfig = byValue.TagcloudConfig
	case byValue.HeatmapConfig != nil:
		pm.HeatmapConfig = byValue.HeatmapConfig
	case byValue.WaffleConfig != nil:
		pm.WaffleConfig = byValue.WaffleConfig
	case byValue.RegionMapConfig != nil:
		pm.RegionMapConfig = byValue.RegionMapConfig
	case byValue.GaugeConfig != nil:
		pm.GaugeConfig = byValue.GaugeConfig
	case byValue.MetricChartConfig != nil:
		pm.MetricChartConfig = byValue.MetricChartConfig.expandToVisMetricChart()
	case byValue.PieChartConfig != nil:
		pm.PieChartConfig = byValue.PieChartConfig
	case byValue.LegacyMetricConfig != nil:
		pm.LegacyMetricConfig = byValue.LegacyMetricConfig
	default:
		return panelModel{}, false
	}
	return pm, true
}

// firstLensVizConverterForPanel returns the first converter whose handlesTFConfig
// matches a scratch `panelModel` (same rule as `panelModel.toAPI` for vis).
func firstLensVizConverterForPanel(pm panelModel) (lensVisualizationConverter, bool) {
	for _, c := range lensVizConverters {
		if c.handlesTFConfig(pm) {
			return c, true
		}
	}
	return nil, false
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

// lensByValueModelFromVizPanelAfterRead maps a `panelModel` that has been populated
// by the vis converter back into `lensDashboardAppByValueModel` (typed block only;
// `config_json` is left unset for optional JSON).
func lensByValueModelFromVizPanelAfterRead(pm *panelModel) (lensDashboardAppByValueModel, bool) {
	if pm == nil {
		return lensDashboardAppByValueModel{}, false
	}
	out := lensDashboardAppByValueModel{
		ConfigJSON: jsontypes.NewNormalizedNull(),
	}
	set := 0
	if pm.XYChartConfig != nil {
		set++
		out.XYChartConfig = pm.XYChartConfig
	}
	if pm.TreemapConfig != nil {
		set++
		out.TreemapConfig = pm.TreemapConfig
	}
	if pm.MosaicConfig != nil {
		set++
		out.MosaicConfig = pm.MosaicConfig
	}
	if pm.DatatableConfig != nil {
		set++
		out.DatatableConfig = pm.DatatableConfig
	}
	if pm.TagcloudConfig != nil {
		set++
		out.TagcloudConfig = pm.TagcloudConfig
	}
	if pm.HeatmapConfig != nil {
		set++
		out.HeatmapConfig = pm.HeatmapConfig
	}
	if pm.WaffleConfig != nil {
		set++
		out.WaffleConfig = pm.WaffleConfig
	}
	if pm.RegionMapConfig != nil {
		set++
		out.RegionMapConfig = pm.RegionMapConfig
	}
	if pm.GaugeConfig != nil {
		set++
		out.GaugeConfig = pm.GaugeConfig
	}
	if pm.MetricChartConfig != nil {
		set++
		out.MetricChartConfig = metricLensByValueFromVisFull(pm.MetricChartConfig)
	}
	if pm.PieChartConfig != nil {
		set++
		out.PieChartConfig = pm.PieChartConfig
	}
	if pm.LegacyMetricConfig != nil {
		set++
		out.LegacyMetricConfig = pm.LegacyMetricConfig
	}
	if set != 1 {
		return lensDashboardAppByValueModel{}, false
	}
	return out, true
}

// tryPopulateTypedLensByValueFromAPI fills `pm.LensDashboardAppConfig` from API
// config bytes when prior state used a matching typed by-value chart. Returns true
// if state was set; otherwise the caller should fall back to by_value.config_json.
func tryPopulateTypedLensByValueFromAPI(
	ctx context.Context,
	dashboard *dashboardModel,
	prior *lensDashboardAppConfigModel,
	configBytes []byte,
	pm *panelModel,
	diags *diag.Diagnostics,
) bool {
	if prior == nil || prior.ByValue == nil {
		return false
	}
	scratchPM, hasTyped := lensByValueToScratchVisPanel(*prior.ByValue)
	if !hasTyped {
		return false
	}
	conv, ok := firstLensVizConverterForPanel(scratchPM)
	if !ok {
		return false
	}
	var vis0 kbapi.KbnDashboardPanelTypeVisConfig0
	if err := json.Unmarshal(configBytes, &vis0); err != nil {
		return false
	}
	if conv.vizType() != detectLensVizType(vis0) {
		return false
	}
	d := conv.populateFromAttributes(ctx, dashboard, &scratchPM, vis0)
	if d.HasError() {
		// Intentional: no error diagnostic here; caller falls back to by_value.config_json
		// (REQ-035) without treating typed read as a user failure.
		return false
	}
	if diags != nil {
		diags.Append(d...)
	}
	by, ok2 := lensByValueModelFromVizPanelAfterRead(&scratchPM)
	if !ok2 {
		return false
	}
	pm.LensDashboardAppConfig = &lensDashboardAppConfigModel{
		ByValue: &by,
	}
	return true
}
