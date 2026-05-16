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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func lensAppByValueModelFromPopulatedVisByValue(t *testing.T, vb *models.VisByValueModel) models.LensDashboardAppByValueModel {
	t.Helper()
	out, ok := lensByValueModelFromChartBlocksAfterRead(&vb.LensByValueChartBlocks)
	require.True(t, ok)
	return out
}

// testMetricByValueFromRoundTrip builds a lens-dashboard-app by_value metric typed model from a vis
// populate path (API union -> populateFromAttributes), stripped to models.MetricChartLensByValueTFModel fields.
func testMetricByValueFromRoundTrip(t *testing.T) models.LensDashboardAppByValueModel {
	t.Helper()
	ctx := context.Background()
	titleM := "M"
	apiChart := kbapi.MetricNoESQL{
		Type:  kbapi.MetricNoESQLTypeMetric,
		Title: &titleM,
		Query: kbapi.FilterSimple{
			Expression: "",
			Language:   new(kbapi.FilterSimpleLanguage("kql")),
		},
		Metrics: []kbapi.MetricNoESQL_Metrics_Item{},
	}
	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	require.NoError(t, attrs.FromMetricNoESQL(apiChart))
	c := lenscommon.ForType(string(kbapi.MetricNoESQLTypeMetric))
	require.NotNil(t, c)
	visBlocks := models.VisByValueModel{}
	require.False(t, c.PopulateFromAttributes(ctx, lensChartResolver(nil), &visBlocks.LensByValueChartBlocks, attrs).HasError())
	return lensAppByValueModelFromPopulatedVisByValue(t, &visBlocks)
}

// testPieByValueConfigBytes is wire JSON for a by-value pie chart (same shape the
// dashboard API stores under lens-dashboard-app config), from the pie converter
// build path, distinct from metric for typed-read fallback.
func testPieByValueConfigBytes(t *testing.T) []byte {
	t.Helper()
	ctx := context.Background()
	title := "P"
	donutHole := kbapi.PieStylingDonutHoleS
	labelPos := kbapi.PieStylingLabelsPositionInside
	visibility := kbapi.PieLegendVisibilityVisible
	nested := true
	truncateLines := float32(3)
	apiChart := kbapi.PieNoESQL{
		Title: &title,
		Styling: kbapi.PieStyling{
			DonutHole: &donutHole,
			Labels: &struct {
				Position *kbapi.PieStylingLabelsPosition `json:"position,omitempty"`
				Visible  *bool                           `json:"visible,omitempty"`
			}{Position: &labelPos},
		},
		Legend: kbapi.PieLegend{
			Size:               kbapi.LegendSizeAuto,
			Nested:             &nested,
			TruncateAfterLines: &truncateLines,
			Visibility:         &visibility,
		},
		DataSource: kbapi.PieNoESQL_DataSource{},
		Query:      kbapi.FilterSimple{Expression: "x", Language: new(kbapi.FilterSimpleLanguageKql)},
		Metrics:    []kbapi.PieNoESQL_Metrics_Item{},
		GroupBy:    new([]kbapi.PieNoESQL_GroupBy_Item{}),
	}
	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	require.NoError(t, attrs.FromPieNoESQL(apiChart))
	c := lenscommon.ForType(string(kbapi.PieNoESQLTypePie))
	require.NotNil(t, c)
	visBlocks := models.VisByValueModel{}
	require.False(t, c.PopulateFromAttributes(ctx, lensChartResolver(nil), &visBlocks.LensByValueChartBlocks, attrs).HasError())
	vis0, d := c.BuildAttributes(&visBlocks.LensByValueChartBlocks, lensChartResolver(nil))
	require.False(t, d.HasError())
	b, err := vis0.MarshalJSON()
	require.NoError(t, err)
	return b
}

// testWaffleByValueModel builds a `waffle_config` by_value source using the same
// data as `buildLensWafflePanelForTest`.
func testWaffleByValueModel(t *testing.T) models.LensDashboardAppByValueModel {
	t.Helper()
	pm := buildLensWafflePanelForTest(t)
	require.NotNil(t, pm.VisConfig)
	require.NotNil(t, pm.VisConfig.ByValue)
	out, ok := lensByValueModelFromChartBlocksAfterRead(&pm.VisConfig.ByValue.LensByValueChartBlocks)
	require.True(t, ok)
	return out
}

// testXyByValueModel builds an `xy_chart_config` by_value source (no-ESQL).
func testXyByValueModel(t *testing.T) models.LensDashboardAppByValueModel {
	t.Helper()
	ctx := context.Background()
	cfg := &models.XYChartConfigModel{
		Title:       types.StringValue("X"),
		Axis:        &models.XYAxisModel{X: &models.XYAxisConfigModel{}, Y: &models.YAxisConfigModel{}},
		Decorations: &models.XYDecorationsModel{},
		Fitting:     &models.XYFittingModel{Type: types.StringValue("none")},
		Layers: []models.XYLayerModel{{
			Type: types.StringValue("area"),
			DataLayer: &models.DataLayerModel{
				DataSourceJSON: jsontypes.NewNormalizedValue(`{"type":"dataView","id":"logs-*"}`),
				Y: []models.YMetricModel{
					{ConfigJSON: jsontypes.NewNormalizedValue(`{"operation":"count","color":"#68BC00","axis":"left"}`)},
				},
			},
		}},
		Legend: &models.XYLegendModel{Visibility: types.StringValue("visible"), Inside: types.BoolValue(false)},
		Query:  &models.FilterSimpleModel{Expression: types.StringValue("*"), Language: types.StringValue("kql")},
	}
	xy, diags := xyChartConfigToAPINoESQL(cfg, nil)
	require.False(t, diags.HasError())
	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	require.NoError(t, attrs.FromXyChartNoESQL(xy))
	c := lenscommon.ForType(string(kbapi.XyChartNoESQLTypeXy))
	require.NotNil(t, c)
	visBlocks := models.VisByValueModel{}
	require.False(t, c.PopulateFromAttributes(ctx, lensChartResolver(nil), &visBlocks.LensByValueChartBlocks, attrs).HasError())
	return lensAppByValueModelFromPopulatedVisByValue(t, &visBlocks)
}

// testPieByValueModel builds a `pie_chart_config` by_value source.
func testPieByValueModel(t *testing.T) models.LensDashboardAppByValueModel {
	t.Helper()
	ctx := context.Background()
	title := "P"
	donutHole := kbapi.PieStylingDonutHoleS
	labelPos := kbapi.PieStylingLabelsPositionInside
	visibility := kbapi.PieLegendVisibilityVisible
	nested := true
	truncateLines := float32(3)
	apiChart := kbapi.PieNoESQL{
		Title: &title,
		Styling: kbapi.PieStyling{
			DonutHole: &donutHole,
			Labels: &struct {
				Position *kbapi.PieStylingLabelsPosition `json:"position,omitempty"`
				Visible  *bool                           `json:"visible,omitempty"`
			}{Position: &labelPos},
		},
		Legend: kbapi.PieLegend{
			Size:               kbapi.LegendSizeAuto,
			Nested:             &nested,
			TruncateAfterLines: &truncateLines,
			Visibility:         &visibility,
		},
		DataSource: kbapi.PieNoESQL_DataSource{},
		Query:      kbapi.FilterSimple{Expression: "x", Language: new(kbapi.FilterSimpleLanguageKql)},
		Metrics:    []kbapi.PieNoESQL_Metrics_Item{},
		GroupBy:    new([]kbapi.PieNoESQL_GroupBy_Item{}),
	}
	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	require.NoError(t, attrs.FromPieNoESQL(apiChart))
	c := lenscommon.ForType(string(kbapi.PieNoESQLTypePie))
	require.NotNil(t, c)
	visBlocks := models.VisByValueModel{}
	require.False(t, c.PopulateFromAttributes(ctx, lensChartResolver(nil), &visBlocks.LensByValueChartBlocks, attrs).HasError())
	return lensAppByValueModelFromPopulatedVisByValue(t, &visBlocks)
}

func Test_visConfig0ToLensAppConfig0_jsonBridge_metric(t *testing.T) {
	t.Parallel()
	titleM := "M"
	apiChart := kbapi.MetricNoESQL{
		Type:  kbapi.MetricNoESQLTypeMetric,
		Title: &titleM,
		Query: kbapi.FilterSimple{
			Expression: "",
			Language:   new(kbapi.FilterSimpleLanguage("kql")),
		},
		Metrics: []kbapi.MetricNoESQL_Metrics_Item{},
	}
	var vis0 kbapi.KbnDashboardPanelTypeVisConfig0
	require.NoError(t, vis0.FromMetricNoESQL(apiChart))

	lens0, err := visConfig0ToLensAppConfig0(vis0)
	require.NoError(t, err)
	metricBack, err := lens0.AsMetricNoESQL()
	require.NoError(t, err)
	assert.Equal(t, kbapi.MetricNoESQLTypeMetric, metricBack.Type)
	assert.Equal(t, "M", *metricBack.Title)
}

func Test_lensByValueToScratchVisPanel_roundTripFields(t *testing.T) {
	t.Parallel()
	by := models.LensDashboardAppByValueModel{
		ConfigJSON:        jsontypes.NewNormalizedNull(),
		MetricChartConfig: &models.MetricChartLensByValueTFModel{},
	}
	pm, ok := lensByValueToScratchVisPanel(by)
	require.True(t, ok)
	require.NotNil(t, pm.VisConfig.ByValue.MetricChartConfig)
}

func Test_lensDashboardAppByValueToAPI_typedMetric_producesLensDashboardAppByValueChart(
	t *testing.T,
) {
	t.Parallel()
	by := testMetricByValueFromRoundTrip(t)
	id := "p1"
	item, diags := lensDashboardAppByValueToAPI(
		by,
		lensDashboardAPIGrid{X: 0, Y: 0, W: float32ptr(24), H: float32ptr(12)},
		&id,
		nil,
	)
	require.False(t, diags.HasError())

	disc, err := item.Discriminator()
	require.NoError(t, err)
	require.Equal(t, panelTypeLensDashboardApp, disc)

	ld, err := item.AsKbnDashboardPanelTypeLensDashboardApp()
	require.NoError(t, err)
	cfg0, err := ld.Config.AsKbnDashboardPanelTypeLensDashboardAppConfig0()
	require.NoError(t, err, "config must be the inline by-value chart union, not a ref shape")
	chart, err := cfg0.AsMetricNoESQL()
	require.NoError(t, err)
	require.Equal(t, kbapi.MetricNoESQLTypeMetric, chart.Type)

	var root map[string]any
	require.NoError(t, json.Unmarshal(mustJSON(t, ld.Config), &root))
	require.Equal(t, "metric", root["type"])
	_, hasAttrs := root["attributes"]
	require.False(t, hasAttrs, "by-value inline chart must not use a vis-style attributes wrapper at config root")
}

// testMetricEsqlByValueModel and peers build `models.LensDashboardAppByValueModel` from ES|QL
// vis0 wire shapes so `lensDashboardAppByValueToAPI` exercises the same adapter as practitioners.
func testMetricEsqlByValueModel(t *testing.T) models.LensDashboardAppByValueModel {
	t.Helper()
	ctx := context.Background()
	metricEsql := kbapi.MetricESQL{
		Type: kbapi.MetricESQLTypeMetric,
		DataSource: kbapi.EsqlDataSource{
			Type:  kbapi.EsqlDataSourceTypeEsql,
			Query: "FROM *",
		},
		Metrics: []kbapi.MetricESQL_Metrics_Item{},
	}
	var vis0 kbapi.KbnDashboardPanelTypeVisConfig0
	require.NoError(t, vis0.FromMetricESQL(metricEsql))
	c := lenscommon.ForType(string(kbapi.MetricNoESQLTypeMetric))
	require.NotNil(t, c)
	visBlocks := models.VisByValueModel{}
	require.False(t, c.PopulateFromAttributes(ctx, lensChartResolver(nil), &visBlocks.LensByValueChartBlocks, vis0).HasError())
	return lensAppByValueModelFromPopulatedVisByValue(t, &visBlocks)
}

func testXyEsqlByValueModel(t *testing.T) models.LensDashboardAppByValueModel {
	t.Helper()
	ctx := context.Background()
	xyEsql := mustUnmarshalXyChartESQL(t, `{
		"type": "xy",
		"title": "E",
		"axis": { "x": {}, "y": {} },
		"filters": [],
		"layers": [
			{
				"type": "line",
				"data_source": {"type": "esql", "query": "FROM logs-* | LIMIT 10"},
				"ignore_global_filters": false,
				"sampling": 1,
				"y": [
					{
						"column": "bytes",
						"format": { "type": "number" }
					}
				]
			}
		],
		"legend": { "visibility": "visible", "inside": false, "size": "auto" },
		"styling": { "line": { "curve": "linear" } },
		"time_range": { "from": "now-7d", "to": "now" }
	}`)
	var vis0 kbapi.KbnDashboardPanelTypeVisConfig0
	require.NoError(t, vis0.FromXyChartESQL(xyEsql))
	c := lenscommon.ForType(string(kbapi.XyChartNoESQLTypeXy))
	require.NotNil(t, c)
	visBlocks := models.VisByValueModel{}
	require.False(t, c.PopulateFromAttributes(ctx, lensChartResolver(nil), &visBlocks.LensByValueChartBlocks, vis0).HasError())
	return lensAppByValueModelFromPopulatedVisByValue(t, &visBlocks)
}

func testPieEsqlByValueModel(t *testing.T) models.LensDashboardAppByValueModel {
	t.Helper()
	ctx := context.Background()
	pieEsql := mustUnmarshalPieESQL(t, `{
		"type": "pie",
		"title": "P",
		"data_source": {"type":"esql","query":"FROM logs-* | LIMIT 10"},
		"legend": {"size":"auto","visibility":"visible"},
		"metrics": [{"operation":"value","column":"bytes","color":{"type":"static","color":"#54B399"},"format":{"type":"number"}}],
		"group_by": [{"operation":"value","column":"h","collapse_by":"avg","color":{"mode":"categorical","palette":"default","mapping":[],"unassignedColor":{"type":"color_code","value":"#D3DAE6"}}}]
	}`)
	var vis0 kbapi.KbnDashboardPanelTypeVisConfig0
	require.NoError(t, vis0.FromPieESQL(pieEsql))
	c := lenscommon.ForType(string(kbapi.PieNoESQLTypePie))
	require.NotNil(t, c)
	visBlocks := models.VisByValueModel{}
	require.False(t, c.PopulateFromAttributes(ctx, lensChartResolver(nil), &visBlocks.LensByValueChartBlocks, vis0).HasError())
	return lensAppByValueModelFromPopulatedVisByValue(t, &visBlocks)
}

func testWaffleEsqlByValueModel(t *testing.T) models.LensDashboardAppByValueModel {
	t.Helper()
	ctx := context.Background()
	vis0 := mustWaffleESQLVis0(t)
	c := lenscommon.ForType(string(kbapi.WaffleNoESQLTypeWaffle))
	require.NotNil(t, c)
	visBlocks := models.VisByValueModel{}
	require.False(t, c.PopulateFromAttributes(ctx, lensChartResolver(nil), &visBlocks.LensByValueChartBlocks, vis0).HasError())
	return lensAppByValueModelFromPopulatedVisByValue(t, &visBlocks)
}

// Test_lensDashboardAppByValueToAPI_typedESQL_adapter_metric_xy_pie_waffle runs
// `lensDashboardAppByValueToAPI` for ES|QL chart data (4.1), complementing no-ESQL
// `Test_lensDashboardAppByValueToAPI_typedNoESQL_adapter_xy_pie_waffle` and union-level
// `Test_visConfig0ToLensAppConfig0_jsonBridge_ESQL_families`.
func Test_lensDashboardAppByValueToAPI_typedESQL_adapter_metric_xy_pie_waffle(t *testing.T) {
	t.Parallel()
	grid := lensDashboardAPIGrid{X: 0, Y: 0, W: float32ptr(24), H: float32ptr(12)}
	cases := []struct {
		name string
		by   func(t *testing.T) models.LensDashboardAppByValueModel
		chk  func(t *testing.T, cfg kbapi.KbnDashboardPanelTypeLensDashboardAppConfig0)
	}{
		{
			"metric",
			testMetricEsqlByValueModel,
			func(t *testing.T, cfg kbapi.KbnDashboardPanelTypeLensDashboardAppConfig0) {
				t.Helper()
				m, err := cfg.AsMetricESQL()
				require.NoError(t, err)
				require.Equal(t, kbapi.MetricESQLTypeMetric, m.Type)
			},
		},
		{
			"xy",
			testXyEsqlByValueModel,
			func(t *testing.T, cfg kbapi.KbnDashboardPanelTypeLensDashboardAppConfig0) {
				t.Helper()
				x, err := cfg.AsXyChartESQL()
				require.NoError(t, err)
				require.Equal(t, kbapi.XyChartESQLTypeXy, x.Type)
			},
		},
		{
			"pie",
			testPieEsqlByValueModel,
			func(t *testing.T, cfg kbapi.KbnDashboardPanelTypeLensDashboardAppConfig0) {
				t.Helper()
				p, err := cfg.AsPieESQL()
				require.NoError(t, err)
				require.Equal(t, kbapi.PieESQLTypePie, p.Type)
			},
		},
		{
			"waffle",
			testWaffleEsqlByValueModel,
			func(t *testing.T, cfg kbapi.KbnDashboardPanelTypeLensDashboardAppConfig0) {
				t.Helper()
				w, err := cfg.AsWaffleESQL()
				require.NoError(t, err)
				require.Equal(t, kbapi.WaffleESQLTypeWaffle, w.Type)
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			by := tc.by(t)
			item, diags := lensDashboardAppByValueToAPI(by, grid, nil, nil)
			require.False(t, diags.HasError())
			ld, err := item.AsKbnDashboardPanelTypeLensDashboardApp()
			require.NoError(t, err)
			cfg0, err := ld.Config.AsKbnDashboardPanelTypeLensDashboardAppConfig0()
			require.NoError(t, err)
			tc.chk(t, cfg0)
		})
	}
}

// Test_lensDashboardAppByValueToAPI_typedNoESQL_adapter_xy_pie_waffle covers the typed
// by_value adapter build path for additional no-ESQL families (metric is above). ES|QL
// build-path coverage is `Test_lensDashboardAppByValueToAPI_typedESQL_adapter_metric_xy_pie_waffle`;
// union JSON bridge: `Test_visConfig0ToLensAppConfig0_jsonBridge_ESQL_families`.
func Test_lensDashboardAppByValueToAPI_typedNoESQL_adapter_xy_pie_waffle(t *testing.T) {
	t.Parallel()
	grid := lensDashboardAPIGrid{X: 0, Y: 0, W: float32ptr(24), H: float32ptr(12)}
	cases := []struct {
		name string
		by   func(t *testing.T) models.LensDashboardAppByValueModel
		chk  func(t *testing.T, cfg kbapi.KbnDashboardPanelTypeLensDashboardAppConfig0)
	}{
		{
			"xy",
			testXyByValueModel,
			func(t *testing.T, cfg kbapi.KbnDashboardPanelTypeLensDashboardAppConfig0) {
				t.Helper()
				x, err := cfg.AsXyChartNoESQL()
				require.NoError(t, err)
				require.Equal(t, kbapi.XyChartNoESQLTypeXy, x.Type)
			},
		},
		{
			"pie",
			testPieByValueModel,
			func(t *testing.T, cfg kbapi.KbnDashboardPanelTypeLensDashboardAppConfig0) {
				t.Helper()
				p, err := cfg.AsPieNoESQL()
				require.NoError(t, err)
				require.Equal(t, kbapi.PieNoESQLTypePie, p.Type)
			},
		},
		{
			"waffle",
			testWaffleByValueModel,
			func(t *testing.T, cfg kbapi.KbnDashboardPanelTypeLensDashboardAppConfig0) {
				t.Helper()
				w, err := cfg.AsWaffleNoESQL()
				require.NoError(t, err)
				require.Equal(t, kbapi.WaffleNoESQLTypeWaffle, w.Type)
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			by := tc.by(t)
			item, diags := lensDashboardAppByValueToAPI(by, grid, nil, nil)
			require.False(t, diags.HasError())
			ld, err := item.AsKbnDashboardPanelTypeLensDashboardApp()
			require.NoError(t, err)
			cfg0, err := ld.Config.AsKbnDashboardPanelTypeLensDashboardAppConfig0()
			require.NoError(t, err)
			tc.chk(t, cfg0)
		})
	}
}

func Test_populateLensDashboardAppByValueFromAPI_typedRead_repopulatesTypedBlock(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	by := testMetricByValueFromRoundTrip(t)
	item, diags := lensDashboardAppByValueToAPI(by, lensDashboardAPIGrid{}, nil, nil)
	require.False(t, diags.HasError())
	ld, err := item.AsKbnDashboardPanelTypeLensDashboardApp()
	require.NoError(t, err)
	configBytes, err := ld.Config.MarshalJSON()
	require.NoError(t, err)

	prior := &models.LensDashboardAppConfigModel{ByValue: &by}
	var pm models.PanelModel
	d := populateLensDashboardAppByValueFromAPI(ctx, nil, prior, configBytes, &pm)
	require.False(t, d.HasError())
	require.NotNil(t, pm.LensDashboardAppConfig)
	require.NotNil(t, pm.LensDashboardAppConfig.ByValue)
	require.NotNil(t, pm.LensDashboardAppConfig.ByValue.MetricChartConfig, "read-back should keep the typed chart block, not only config_json")
	assert.True(
		t,
		pm.LensDashboardAppConfig.ByValue.ConfigJSON.IsNull(),
		"typed by-value read should leave config_json unset (null), not a JSON string",
	)
}

func Test_populateLensDashboardAppByValueFromAPI_typedRead_mismatchedChartFallsBackToConfigJSON(
	t *testing.T,
) {
	t.Parallel()
	ctx := context.Background()
	by := testMetricByValueFromRoundTrip(t)
	pieBytes := testPieByValueConfigBytes(t)
	prior := &models.LensDashboardAppConfigModel{ByValue: &by}
	var pm models.PanelModel
	d := populateLensDashboardAppByValueFromAPI(ctx, nil, prior, pieBytes, &pm)
	require.False(t, d.HasError())
	require.NotNil(t, pm.LensDashboardAppConfig)
	require.NotNil(t, pm.LensDashboardAppConfig.ByValue)
	require.Nil(t, pm.LensDashboardAppConfig.ByValue.MetricChartConfig, "mismatched API chart should drop typed read")
	require.True(t, typeutils.IsKnown(pm.LensDashboardAppConfig.ByValue.ConfigJSON))
	var root map[string]any
	require.NoError(t, json.Unmarshal([]byte(pm.LensDashboardAppConfig.ByValue.ConfigJSON.ValueString()), &root))
	require.NotEqual(t, "metric", root["type"])
}

// Test_typedByValueReadFallback_silent: failed typed mapping must not add error-level
// diagnostics; full populate still succeeds and records config_json.
func Test_typedByValueReadFallback_noErrorDiagnostics(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	by := testMetricByValueFromRoundTrip(t)
	pieBytes := testPieByValueConfigBytes(t)
	prior := &models.LensDashboardAppConfigModel{ByValue: &by}
	var d diag.Diagnostics
	var pm models.PanelModel
	ok := tryPopulateTypedLensByValueFromAPI(ctx, nil, prior, pieBytes, &pm, &d)
	require.False(t, ok)
	require.False(t, d.HasError(), "tryPopulate must not add errors when falling back to config_json read")
	full := populateLensDashboardAppByValueFromAPI(ctx, nil, prior, pieBytes, &pm)
	require.False(t, full.HasError())
}

func Test_lensByValueAdapter_schemaTypedBlocksHaveScratchMapping(t *testing.T) {
	t.Parallel()
	raw := []models.LensByValueChartBlocks{
		{XYChartConfig: &models.XYChartConfigModel{}},
		{TreemapConfig: &models.TreemapConfigModel{}},
		{MosaicConfig: &models.MosaicConfigModel{}},
		{DatatableConfig: &models.DatatableConfigModel{}},
		{TagcloudConfig: &models.TagcloudConfigModel{}},
		{HeatmapConfig: &models.HeatmapConfigModel{}},
		{WaffleConfig: &models.WaffleConfigModel{}},
		{RegionMapConfig: &models.RegionMapConfigModel{}},
		{GaugeConfig: &models.GaugeConfigModel{}},
		{MetricChartConfig: &models.MetricChartConfigModel{}},
		{PieChartConfig: &models.PieChartConfigModel{}},
		{LegacyMetricConfig: &models.LegacyMetricConfigModel{}},
	}
	cases := make([]models.LensDashboardAppByValueModel, 0, len(raw))
	for i := range raw {
		by, ok := lensByValueModelFromChartBlocksAfterRead(&raw[i])
		require.Truef(t, ok, "block set %d should map to a single typed lens by_value model", i)
		cases = append(cases, by)
	}
	want := 0
	for _, name := range lensDashboardAppByValueSourceAttrNames {
		if name != "config_json" {
			want++
		}
	}
	require.Len(t, cases, want, "add a cases entry and a lensByValueToScratchVisPanel arm when adding a by_value chart to schema")
	for i := range cases {
		require.Truef(t, lensByValueModelHasAnyTypedChartBlock(&cases[i]), "case %d should be recognized as a typed by-value source", i)
		pm, ok := lensByValueToScratchVisPanel(cases[i])
		require.Truef(t, ok, "case %d should map to a scratch vis panel", i)
		_, cok := firstLensVisConverterForPanel(pm)
		require.Truef(t, cok, "case %d should match a vis converter", i)
	}
}
