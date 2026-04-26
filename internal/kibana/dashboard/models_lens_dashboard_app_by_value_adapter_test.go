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
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testMetricByValueFromRoundTrip is a real metricChartConfigModel produced via the same
// path as the vis round-trip test (API union -> populateFromAttributes).
func testMetricByValueFromRoundTrip(t *testing.T) lensDashboardAppByValueModel {
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
	converter := newMetricChartPanelConfigConverter()
	pm := &panelModel{}
	require.False(t, converter.populateFromAttributes(ctx, pm, attrs).HasError())
	return lensDashboardAppByValueModel{MetricChartConfig: pm.MetricChartConfig}
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
	converter := newPieChartPanelConfigConverter()
	pm := &panelModel{}
	require.False(t, converter.populateFromAttributes(ctx, pm, attrs).HasError())
	vis0, d := converter.buildAttributes(*pm)
	require.False(t, d.HasError())
	b, err := vis0.MarshalJSON()
	require.NoError(t, err)
	return b
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
	by := lensDashboardAppByValueModel{MetricChartConfig: &metricChartConfigModel{}}
	pm, ok := lensByValueToScratchVisPanel(by)
	require.True(t, ok)
	require.NotNil(t, pm.MetricChartConfig)
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

func Test_populateLensDashboardAppByValueFromAPI_typedRead_repopulatesTypedBlock(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	by := testMetricByValueFromRoundTrip(t)
	item, diags := lensDashboardAppByValueToAPI(by, lensDashboardAPIGrid{}, nil)
	require.False(t, diags.HasError())
	ld, err := item.AsKbnDashboardPanelTypeLensDashboardApp()
	require.NoError(t, err)
	configBytes, err := ld.Config.MarshalJSON()
	require.NoError(t, err)

	prior := &lensDashboardAppConfigModel{ByValue: &by}
	var pm panelModel
	d := populateLensDashboardAppByValueFromAPI(ctx, prior, configBytes, &pm)
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
	prior := &lensDashboardAppConfigModel{ByValue: &by}
	var pm panelModel
	d := populateLensDashboardAppByValueFromAPI(ctx, prior, pieBytes, &pm)
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
	prior := &lensDashboardAppConfigModel{ByValue: &by}
	var d diag.Diagnostics
	var pm panelModel
	ok := tryPopulateTypedLensByValueFromAPI(ctx, prior, pieBytes, &pm, &d)
	require.False(t, ok)
	require.False(t, d.HasError(), "tryPopulate must not add errors when falling back to config_json read")
	full := populateLensDashboardAppByValueFromAPI(ctx, prior, pieBytes, &pm)
	require.False(t, full.HasError())
}

func Test_lensByValueAdapter_schemaTypedBlocksHaveScratchMapping(t *testing.T) {
	t.Parallel()
	cases := []lensDashboardAppByValueModel{
		{XYChartConfig: &xyChartConfigModel{}},
		{TreemapConfig: &treemapConfigModel{}},
		{MosaicConfig: &mosaicConfigModel{}},
		{DatatableConfig: &datatableConfigModel{}},
		{TagcloudConfig: &tagcloudConfigModel{}},
		{HeatmapConfig: &heatmapConfigModel{}},
		{WaffleConfig: &waffleConfigModel{}},
		{RegionMapConfig: &regionMapConfigModel{}},
		{GaugeConfig: &gaugeConfigModel{}},
		{MetricChartConfig: &metricChartConfigModel{}},
		{PieChartConfig: &pieChartConfigModel{}},
		{LegacyMetricConfig: &legacyMetricConfigModel{}},
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
		_, cok := firstLensVizConverterForPanel(pm)
		require.Truef(t, cok, "case %d should match a vis converter", i)
	}
}
