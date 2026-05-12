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
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test_eachExposedByValueSource_visAndLensUnionsJSONBridge checks that for every
// non-config_json by_value schema block there is a representative Kibana inline
// chart (vis and lens-dashboard-app share the same generated variant structs).
func Test_eachExposedByValueSource_visAndLensUnionsJSONBridge(t *testing.T) {
	t.Parallel()
	cases := []struct {
		attr  string
		build func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0
		want  string // root JSON "type" string
	}{
		{
			"xy_chart_config",
			func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
				t.Helper()
				xy, diags := (&xyChartConfigModel{
					Title:       types.StringValue("Parity"),
					Axis:        &xyAxisModel{X: &xyAxisConfigModel{}, Y: &yAxisConfigModel{}},
					Decorations: &xyDecorationsModel{},
					Fitting:     &xyFittingModel{Type: types.StringValue("none")},
					Layers: []xyLayerModel{{
						Type: types.StringValue("area"),
						DataLayer: &dataLayerModel{
							DataSourceJSON: jsontypes.NewNormalizedValue(`{"type":"dataView","id":"logs-*"}`),
							Y: []yMetricModel{
								{ConfigJSON: jsontypes.NewNormalizedValue(`{"operation":"count","color":"#68BC00","axis":"left"}`)},
							},
						},
					}},
					Legend: &xyLegendModel{Visibility: types.StringValue("visible"), Inside: types.BoolValue(false)},
					Query:  &filterSimpleModel{Expression: types.StringValue("*"), Language: types.StringValue("kql")},
				}).toAPINoESQL(nil)
				require.False(t, diags.HasError())
				var vis kbapi.KbnDashboardPanelTypeVisConfig0
				require.NoError(t, vis.FromXyChartNoESQL(xy))
				return vis
			},
			"xy",
		},
		{
			"treemap_config",
			func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
				t.Helper()
				apiJSON := `{
					"type": "treemap",
					"title": "T",
					"data_source": {"type":"dataView","id":"metrics-*"},
					"query": {"language":"kql","expression":""},
					"legend": {"size": "small"},
					"metrics": [{"operation":"count"}],
					"group_by": [{"operation":"terms","field":"host.name","collapse_by":"avg"}]
				}`
				var api kbapi.TreemapNoESQL
				require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))
				var vis kbapi.KbnDashboardPanelTypeVisConfig0
				require.NoError(t, vis.FromTreemapNoESQL(api))
				return vis
			},
			"treemap",
		},
		{
			"mosaic_config",
			func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
				t.Helper()
				const grp = `{"mode":"categorical","palette":"default","mapping":[],"unassigned":{"type":"color_code","value":"#D3DAE6"}}`
				groupBy := `[{"operation":"terms","collapse_by":"avg","fields":["host.name"],"color":` + grp + `}]`
				groupBreakdownBy := `[{"operation":"terms","collapse_by":"avg","fields":["service.name"],"color":` + grp + `}]`
				apiJSON := `{
					"type": "mosaic",
					"title": "M",
					"data_source": {"type":"dataView","id":"metrics-*"},
					"query": {"language":"kql","expression":""},
					"legend": {"size":"small"},
					"metric": {"operation":"count"},
					"group_by": ` + groupBy + `,
					"group_breakdown_by": ` + groupBreakdownBy + `
				}`
				var api kbapi.MosaicNoESQL
				require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))
				var vis kbapi.KbnDashboardPanelTypeVisConfig0
				require.NoError(t, vis.FromMosaicNoESQL(api))
				return vis
			},
			"mosaic",
		},
		{
			"datatable_config",
			func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
				t.Helper()
				header := kbapi.DatatableDensity_Height_Header{}
				require.NoError(t, header.FromDatatableDensityHeightHeader0(kbapi.DatatableDensityHeightHeader0{Type: kbapi.DatatableDensityHeightHeader0TypeAuto}))
				value := kbapi.DatatableDensity_Height_Value{}
				require.NoError(t, value.FromDatatableDensityHeightValue0(kbapi.DatatableDensityHeightValue0{Type: kbapi.DatatableDensityHeightValue0TypeAuto}))
				api := kbapi.DatatableNoESQL{
					Type:                kbapi.DatatableNoESQLTypeDataTable,
					Title:               new("Datatable NoESQL Round-Trip"),
					Description:         new("Converter test"),
					IgnoreGlobalFilters: new(true),
					Sampling:            new(float32(0.5)),
					Styling: kbapi.DatatableStyling{
						Density: kbapi.DatatableDensity{
							Mode: new(kbapi.DatatableDensityModeDefault),
							Height: &struct {
								Header *kbapi.DatatableDensity_Height_Header `json:"header,omitempty"`
								Value  *kbapi.DatatableDensity_Height_Value  `json:"value,omitempty"`
							}{Header: &header, Value: &value},
						},
					},
					Query:   kbapi.FilterSimple{},
					Metrics: []kbapi.DatatableNoESQL_Metrics_Item{},
				}
				require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"metrics-*"}`), &api.DataSource))
				require.NoError(t, json.Unmarshal([]byte(`{"language":"kql","expression":"*"}`), &api.Query))
				var vis kbapi.KbnDashboardPanelTypeVisConfig0
				require.NoError(t, vis.FromDatatableNoESQL(api))
				return vis
			},
			"data_table",
		},
		{
			"tagcloud_config",
			func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
				t.Helper()
				api := kbapi.TagcloudNoESQL{
					Type:        kbapi.TagcloudNoESQLTypeTagCloud,
					Title:       new("T"),
					Description: new("d"),
				}
				_ = json.Unmarshal([]byte(`{"index":"i"}`), &api.DataSource)
				_ = json.Unmarshal([]byte(`{"expression":"*","language":"kql"}`), &api.Query)
				_ = json.Unmarshal([]byte(`{"operation":{"operation_type":"count"}}`), &api.Metric)
				_ = json.Unmarshal([]byte(`{"operation":{"operation_type":"terms"},"field":"t"}`), &api.TagBy)
				_ = json.Unmarshal([]byte(`{}`), &api.Styling)
				_ = json.Unmarshal([]byte(`[]`), &api.Filters)
				var tr kbapi.KbnEsQueryServerTimeRangeSchema
				_ = json.Unmarshal([]byte(`{"from":"now-7d","to":"now"}`), &tr)
				api.TimeRange = tr
				var vis kbapi.KbnDashboardPanelTypeVisConfig0
				require.NoError(t, vis.FromTagcloudNoESQL(api))
				return vis
			},
			"tag_cloud",
		},
		{
			"heatmap_config",
			func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
				t.Helper()
				heatmap := kbapi.HeatmapNoESQL{
					Type:                kbapi.HeatmapNoESQLTypeHeatmap,
					Title:               new("H"),
					Description:         new("d"),
					IgnoreGlobalFilters: new(true),
					Sampling:            new(float32(0.5)),
					Query: kbapi.FilterSimple{
						Expression: "status:200",
						Language:   new(kbapi.FilterSimpleLanguage("kql")),
					},
					Axis: kbapi.HeatmapAxes{
						X: kbapi.HeatmapXAxis{},
						Y: kbapi.HeatmapYAxis{},
					},
					Styling: kbapi.HeatmapStyling{Cells: kbapi.HeatmapCells{}},
					Legend:  kbapi.HeatmapLegend{Size: kbapi.LegendSizeM},
				}
				require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"metrics-*"}`), &heatmap.DataSource))
				require.NoError(t, json.Unmarshal([]byte(`{"operation":"count"}`), &heatmap.Metric))
				require.NoError(t, json.Unmarshal([]byte(`{"operation":"filters","filters":[{"label":"All","filter":{"query":"*","language":"kql"}}]}`), &heatmap.X))
				var vis kbapi.KbnDashboardPanelTypeVisConfig0
				require.NoError(t, vis.FromHeatmapNoESQL(heatmap))
				return vis
			},
			"heatmap",
		},
		{
			"waffle_config",
			func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
				t.Helper()
				pm := buildLensWafflePanelForTest(t)
				converter := newWafflePanelConfigConverter()
				blocks := lensByValueChartBlocksFromPanel(&pm)
				require.NotNil(t, blocks)
				vis0, d := converter.buildAttributes(blocks, nil)
				require.False(t, d.HasError())
				return vis0
			},
			"waffle",
		},
		{
			"region_map_config",
			func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
				t.Helper()
				lang := kbapi.FilterSimpleLanguage("kql")
				api := kbapi.RegionMapNoESQL{
					Type: kbapi.RegionMapNoESQLTypeRegionMap,
					Query: kbapi.FilterSimple{
						Language:   &lang,
						Expression: "*",
					},
				}
				_ = json.Unmarshal([]byte(`{"type":"dataView","id":"metrics-*"}`), &api.DataSource)
				_ = json.Unmarshal([]byte(`{"operation":"count"}`), &api.Metric)
				_ = json.Unmarshal([]byte(`{"operation":"filters","filters":[{"filter":{"query":"*","language":"kql"},"label":"A"}]}`), &api.Region)
				var vis kbapi.KbnDashboardPanelTypeVisConfig0
				require.NoError(t, vis.FromRegionMapNoESQL(api))
				return vis
			},
			"region_map",
		},
		{
			"gauge_config",
			func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
				t.Helper()
				api := kbapi.GaugeNoESQL{
					Type:                kbapi.GaugeNoESQLTypeGauge,
					Title:               new("G"),
					Description:         new("d"),
					IgnoreGlobalFilters: new(true),
					Sampling:            new(float32(0.5)),
				}
				require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"metrics-*"}`), &api.DataSource))
				require.NoError(t, json.Unmarshal([]byte(`{"expression":"*","language":"kql"}`), &api.Query))
				require.NoError(t, json.Unmarshal([]byte(`{"operation":"count"}`), &api.Metric))
				var vis kbapi.KbnDashboardPanelTypeVisConfig0
				require.NoError(t, vis.FromGaugeNoESQL(api))
				return vis
			},
			"gauge",
		},
		{
			"metric_chart_config",
			func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
				t.Helper()
				m := testMetricByValueFromRoundTrip(t)
				return m.metricsTypedVis0(t)
			},
			"metric",
		},
		{
			"pie_chart_config",
			func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
				t.Helper()
				return testPieByValueConfigVis0(t)
			},
			"pie",
		},
		{
			"legacy_metric_config",
			func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
				t.Helper()
				const apiJSON = `{
					"type": "legacy_metric",
					"title": "Legacy Metric Round-Trip",
					"description": "Converter test",
					"data_source": {"type": "data_view_spec", "index_pattern": "metrics-*"},
					"query": {"language": "kql", "query": "*"},
					"sampling": 0.5,
					"ignore_global_filters": true,
					"metric": {"operation": "count", "format": {"type": "number"}}
				}`
				var api kbapi.LegacyMetricNoESQL
				require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))
				var vis kbapi.KbnDashboardPanelTypeVisConfig0
				require.NoError(t, vis.FromLegacyMetricNoESQL(api))
				return vis
			},
			"legacy_metric",
		},
	}

	want := 0
	for _, n := range lensDashboardAppByValueSourceAttrNames {
		if n != "config_json" {
			want++
		}
	}
	require.Len(t, cases, want, "add a case when adding a by_value chart block in schema")
	for _, tc := range cases {
		t.Run(tc.attr, func(t *testing.T) {
			t.Parallel()
			vis0 := tc.build(t)
			visWire, err := vis0.MarshalJSON()
			require.NoError(t, err)
			var visRoot map[string]any
			require.NoError(t, json.Unmarshal(visWire, &visRoot))
			require.Equal(t, tc.want, visRoot["type"], "vis union wire type")

			lens0, err := visConfig0ToLensAppConfig0(vis0)
			require.NoError(t, err)
			lensWire, err := json.Marshal(lens0)
			require.NoError(t, err)
			var lensRoot map[string]any
			require.NoError(t, json.Unmarshal(lensWire, &lensRoot))
			assert.Equal(t, visRoot["type"], lensRoot["type"], "lens-dashboard-app inline config union should match vis chart type")
		})
	}
}

func (m lensDashboardAppByValueModel) metricsTypedVis0(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
	t.Helper()
	pm, ok := lensByValueToScratchVisPanel(m)
	require.True(t, ok)
	conv, okc := firstLensVizConverterForPanel(pm)
	require.True(t, okc)
	blocks := lensByValueChartBlocksFromPanel(&pm)
	require.NotNil(t, blocks)
	vis, d := conv.buildAttributes(blocks, nil)
	require.False(t, d.HasError())
	return vis
}

// testPieByValueConfigVis0 returns vis0 JSON built by the pie chart converter.
func testPieByValueConfigVis0(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
	t.Helper()
	b := testPieByValueConfigBytes(t)
	var vis0 kbapi.KbnDashboardPanelTypeVisConfig0
	require.NoError(t, json.Unmarshal(b, &vis0))
	return vis0
}

// Test_visConfig0ToLensAppConfig0_jsonBridge_ESQL_families: representative ES|QL chart families
// used by the adapter JSON bridge (metric, xy, pie, waffle).
func Test_visConfig0ToLensAppConfig0_jsonBridge_ESQL_families(t *testing.T) {
	t.Parallel()

	metricEsql := kbapi.MetricESQL{
		Type: kbapi.MetricESQLTypeMetric,
		DataSource: kbapi.EsqlDataSource{
			Type:  kbapi.EsqlDataSourceTypeEsql,
			Query: "FROM *",
		},
		Metrics: []kbapi.MetricESQL_Metrics_Item{},
	}
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
	pieEsql := mustUnmarshalPieESQL(t, `{
		"type": "pie",
		"title": "P",
		"data_source": {"type":"esql","query":"FROM logs-* | LIMIT 10"},
		"legend": {"size":"auto","visibility":"visible"},
		"metrics": [{"operation":"value","column":"bytes","color":{"type":"static","color":"#54B399"},"format":{"type":"number"}}],
		"group_by": [{"operation":"value","column":"h","collapse_by":"avg","color":{"mode":"categorical","palette":"default","mapping":[],"unassignedColor":{"type":"color_code","value":"#D3DAE6"}}}]
	}`)
	waffleVis := mustWaffleESQLVis0(t)

	for _, tc := range []struct {
		name string
		arm  func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0
	}{
		{
			"metric",
			func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
				t.Helper()
				var v kbapi.KbnDashboardPanelTypeVisConfig0
				require.NoError(t, v.FromMetricESQL(metricEsql))
				return v
			},
		},
		{
			"xy",
			func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
				t.Helper()
				var v kbapi.KbnDashboardPanelTypeVisConfig0
				require.NoError(t, v.FromXyChartESQL(xyEsql))
				return v
			},
		},
		{
			"pie",
			func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
				t.Helper()
				var v kbapi.KbnDashboardPanelTypeVisConfig0
				require.NoError(t, v.FromPieESQL(pieEsql))
				return v
			},
		},
		{
			"waffle",
			func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
				t.Helper()
				return waffleVis
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			vis0 := tc.arm(t)
			lens0, err := visConfig0ToLensAppConfig0(vis0)
			require.NoError(t, err)
			visB, _ := vis0.MarshalJSON()
			lensB, _ := json.Marshal(lens0)
			var a, b map[string]any
			require.NoError(t, json.Unmarshal(visB, &a))
			require.NoError(t, json.Unmarshal(lensB, &b))
			require.Equal(t, a["type"], b["type"], "esql vis/lens jsonBridge type")
		})
	}
}

func mustUnmarshalXyChartESQL(t *testing.T, s string) kbapi.XyChartESQL {
	t.Helper()
	var v kbapi.XyChartESQL
	require.NoError(t, json.Unmarshal([]byte(s), &v))
	return v
}

func mustUnmarshalPieESQL(t *testing.T, s string) kbapi.PieESQL {
	t.Helper()
	var v kbapi.PieESQL
	require.NoError(t, json.Unmarshal([]byte(s), &v))
	return v
}

// mustWaffleESQLVis0 is copied from Test_wafflePanelConfigConverter_populateFromAttributes_buildAttributes_roundTrip_ESQL.
func mustWaffleESQLVis0(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
	t.Helper()
	var format kbapi.FormatType
	require.NoError(t, json.Unmarshal([]byte(`{"type":"number"}`), &format))

	var colorMap kbapi.ColorMapping
	require.NoError(t, json.Unmarshal([]byte(`{"mode":"categorical","palette":"default","mapping":[],"unassignedColor":{"type":"color_code","value":"#D3DAE6"}}`), &colorMap))

	staticColorUnion := kbapi.WaffleESQL_Metrics_Color{}
	require.NoError(t, staticColorUnion.FromStaticColor(kbapi.StaticColor{
		Type:  kbapi.Static,
		Color: "#006BB4",
	}))

	waffle := kbapi.WaffleESQL{
		Type:        kbapi.WaffleESQLTypeWaffle,
		Title:       new("Waffle ESQL Round-Trip"),
		Description: new("esql test"),
		Legend:      kbapi.WaffleLegend{Size: kbapi.LegendSizeS},
		Metrics: []struct {
			Color  *kbapi.WaffleESQL_Metrics_Color `json:"color,omitempty"`
			Column string                          `json:"column"`
			Format kbapi.FormatType                `json:"format"`
			Label  *string                         `json:"label,omitempty"`
		}{
			{Column: "cnt", Format: format, Color: &staticColorUnion},
		},
		GroupBy: &[]struct {
			CollapseBy kbapi.CollapseBy   `json:"collapse_by"`
			Color      kbapi.ColorMapping `json:"color"`
			Column     string             `json:"column"`
			Format     kbapi.FormatType   `json:"format"`
			Label      *string            `json:"label,omitempty"`
		}{
			{Column: "host", Format: format, CollapseBy: kbapi.CollapseByAvg, Color: colorMap},
		},
	}
	require.NoError(t, json.Unmarshal([]byte(`{"type":"esql","query":"FROM logs-* | STATS c = COUNT() BY host | LIMIT 10"}`), &waffle.DataSource))
	var vis kbapi.KbnDashboardPanelTypeVisConfig0
	require.NoError(t, vis.FromWaffleESQL(waffle))
	return vis
}
