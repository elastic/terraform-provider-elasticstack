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

package lenscommon

import (
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/stretchr/testify/require"
)

func TestDetectVizType_chartKindsPerArm(t *testing.T) {
	t.Parallel()

	noESQLHeader := kbapi.DatatableDensity_Height_Header{}
	require.NoError(t, noESQLHeader.FromDatatableDensityHeightHeader0(kbapi.DatatableDensityHeightHeader0{Type: kbapi.DatatableDensityHeightHeader0TypeAuto}))
	noESQLValue := kbapi.DatatableDensity_Height_Value{}
	require.NoError(t, noESQLValue.FromDatatableDensityHeightValue0(kbapi.DatatableDensityHeightValue0{Type: kbapi.DatatableDensityHeightValue0TypeAuto}))
	minDatatableNoESQL := kbapi.DatatableNoESQL{
		Type:    kbapi.DatatableNoESQLTypeDataTable,
		Query:   kbapi.FilterSimple{},
		Styling: kbapi.DatatableStyling{Density: kbapi.DatatableDensity{Mode: new(kbapi.DatatableDensityModeDefault)}},
		Metrics: []kbapi.DatatableNoESQL_Metrics_Item{},
		TimeRange: func() kbapi.KbnEsQueryServerTimeRangeSchema {
			var tr kbapi.KbnEsQueryServerTimeRangeSchema
			require.NoError(t, json.Unmarshal([]byte(`{"from":"now-7d","to":"now"}`), &tr))
			return tr
		}(),
	}
	require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"i"}`), &minDatatableNoESQL.DataSource))
	require.NoError(t, json.Unmarshal([]byte(`{"language":"kql","expression":"*"}`), &minDatatableNoESQL.Query))
	minDatatableNoESQL.Styling.Density.Height = &struct {
		Header *kbapi.DatatableDensity_Height_Header `json:"header,omitempty"`
		Value  *kbapi.DatatableDensity_Height_Value  `json:"value,omitempty"`
	}{Header: &noESQLHeader, Value: &noESQLValue}

	tests := []struct {
		name  string
		build func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0
		want  string
	}{
		{
			name: "empty_union",
			build: func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
				t.Helper()
				return kbapi.KbnDashboardPanelTypeVisConfig0{}
			},
			want: "",
		},
		{
			name: "xy/no_esql",
			build: func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
				t.Helper()
				raw := `{
					"type":"xy","title":"t","axis":{"x":{},"y":{}},
					"layers":[{"type":"line","data_source":{"type":"dataView","id":"l"},"ignore_global_filters":false,"sampling":1,"y":[{"operation":"count"}]}],
					"legend":{"visibility":"visible","inside":false,"size":"auto"},
					"filters":[],"styling":{"line":{"curve":"linear"}},
					"query":{"expression":"*","language":"kql"}
				}`
				var x kbapi.XyChartNoESQL
				require.NoError(t, json.Unmarshal([]byte(raw), &x))
				var v kbapi.KbnDashboardPanelTypeVisConfig0
				require.NoError(t, v.FromXyChartNoESQL(x))
				return v
			},
			want: string(kbapi.XyChartNoESQLTypeXy),
		},
		{
			name: "xy/esql",
			build: func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
				t.Helper()
				raw := `{
					"type":"xy","title":"E","axis":{"x":{},"y":{}},"filters":[],
					"layers":[{"type":"line","data_source":{"type":"esql","query":"FROM logs-* | LIMIT 10"},
						"ignore_global_filters":false,"sampling":1,"y":[{"column":"bytes","format":{"type":"number"}}]}],
					"legend":{"visibility":"visible","inside":false,"size":"auto"},
					"styling":{"line":{"curve":"linear"}},
					"time_range":{"from":"now-7d","to":"now"}
				}`
				var x kbapi.XyChartESQL
				require.NoError(t, json.Unmarshal([]byte(raw), &x))
				var v kbapi.KbnDashboardPanelTypeVisConfig0
				require.NoError(t, v.FromXyChartESQL(x))
				return v
			},
			want: string(kbapi.XyChartNoESQLTypeXy),
		},
		{
			name: "treemap/no_esql",
			build: func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
				t.Helper()
				apiJSON := `{"type":"treemap","title":"t",` +
					`"data_source":{"type":"dataView","id":"m"},` +
					`"query":{"language":"kql","expression":""},` +
					`"legend":{"size":"small"},` +
					`"metrics":[{"operation":"count"}],` +
					`"group_by":[{"operation":"terms","field":"host.name","collapse_by":"avg"}]}`
				var api kbapi.TreemapNoESQL
				require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))
				var v kbapi.KbnDashboardPanelTypeVisConfig0
				require.NoError(t, v.FromTreemapNoESQL(api))
				return v
			},
			want: string(kbapi.TreemapNoESQLTypeTreemap),
		},
		{
			name: "treemap/esql",
			build: func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
				t.Helper()
				apiJSON := `{"type":"treemap","title":"e","description":"","ignore_global_filters":false,"sampling":1,` +
					`"data_source":{"type":"esql","query":"FROM m | LIMIT 1"},` +
					`"legend":{"size":"small"},` +
					`"metrics":[{"column":"bytes","operation":"value","format":{"type":"number"}}],` +
					`"group_by":[{"collapse_by":"avg","column":"host.name","operation":"value"}]}`
				var api kbapi.TreemapESQL
				require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))
				var v kbapi.KbnDashboardPanelTypeVisConfig0
				require.NoError(t, v.FromTreemapESQL(api))
				return v
			},
			want: string(kbapi.TreemapNoESQLTypeTreemap),
		},
		{
			name: "mosaic/no_esql",
			build: func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
				t.Helper()
				const grp = `{"mode":"categorical","palette":"default","mapping":[],"unassigned":{"type":"color_code","value":"#D3DAE6"}}`
				groupBy := `[{"operation":"terms","collapse_by":"avg","fields":["host.name"],"color":` + grp + `}]`
				apiJSON := `{"type":"mosaic","title":"m","data_source":{"type":"dataView","id":"x"},` +
					`"query":{"language":"kql","expression":""},"legend":{"size":"small"},` +
					`"metric":{"operation":"count"},"group_by":` + groupBy + `,"group_breakdown_by":` + groupBy + `}`
				var api kbapi.MosaicNoESQL
				require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))
				var v kbapi.KbnDashboardPanelTypeVisConfig0
				require.NoError(t, v.FromMosaicNoESQL(api))
				return v
			},
			want: string(kbapi.MosaicNoESQLTypeMosaic),
		},
		{
			name: "mosaic/esql",
			build: func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
				t.Helper()
				apiJSON := `{"type":"mosaic","title":"m","data_source":{"type":"esql","query":"FROM m | LIMIT 1"},` +
					`"legend":{"size":"small"},` +
					`"metric":{"column":"bytes","operation":"value","format":{"type":"number"}},` +
					`"group_by":[{"collapse_by":"avg","column":"host.name","operation":"value"}],` +
					`"group_breakdown_by":[{"collapse_by":"avg","column":"s","operation":"value"}]}`
				var api kbapi.MosaicESQL
				require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))
				var v kbapi.KbnDashboardPanelTypeVisConfig0
				require.NoError(t, v.FromMosaicESQL(api))
				return v
			},
			want: string(kbapi.MosaicNoESQLTypeMosaic),
		},
		{
			name: "datatable/no_esql",
			build: func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
				t.Helper()
				var v kbapi.KbnDashboardPanelTypeVisConfig0
				require.NoError(t, v.FromDatatableNoESQL(minDatatableNoESQL))
				return v
			},
			want: string(kbapi.DatatableNoESQLTypeDataTable),
		},
		{
			name: "datatable/esql",
			build: func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
				t.Helper()
				apiJSON := `{"type":"data_table","title":"d","data_source":{"type":"esql","query":"FROM logs-* | LIMIT 5"},` +
					`"filters":[],"metrics":[{"column":"c","operation":"value","format":{"type":"number"}}],` +
					`"rows":[{"column":"r","collapse_by":"avg","format":{"type":"number"}}],` +
					`"styling":{"density":{"mode":"default","height":{"header":{"type":"auto"},"value":{"type":"auto"}}}},` +
					`"time_range":{"from":"now-7d","to":"now"}}`
				var api kbapi.DatatableESQL
				require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))
				var v kbapi.KbnDashboardPanelTypeVisConfig0
				require.NoError(t, v.FromDatatableESQL(api))
				return v
			},
			want: string(kbapi.DatatableNoESQLTypeDataTable),
		},
		{
			name: "tagcloud/no_esql",
			build: func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
				t.Helper()
				api := kbapi.TagcloudNoESQL{
					Type: kbapi.TagcloudNoESQLTypeTagCloud,
				}
				require.NoError(t, json.Unmarshal([]byte(`{"index":"i"}`), &api.DataSource))
				require.NoError(t, json.Unmarshal([]byte(`{"expression":"*","language":"kql"}`), &api.Query))
				require.NoError(t, json.Unmarshal([]byte(`{"operation":{"operation_type":"count"}}`), &api.Metric))
				require.NoError(t, json.Unmarshal([]byte(`{"operation":{"operation_type":"terms"},"field":"t"}`), &api.TagBy))
				require.NoError(t, json.Unmarshal([]byte(`{}`), &api.Styling))
				require.NoError(t, json.Unmarshal([]byte(`[]`), &api.Filters))
				var tr kbapi.KbnEsQueryServerTimeRangeSchema
				require.NoError(t, json.Unmarshal([]byte(`{"from":"now-7d","to":"now"}`), &tr))
				api.TimeRange = tr
				var v kbapi.KbnDashboardPanelTypeVisConfig0
				require.NoError(t, v.FromTagcloudNoESQL(api))
				return v
			},
			want: string(kbapi.TagcloudNoESQLTypeTagCloud),
		},
		{
			name: "tagcloud/esql",
			build: func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
				t.Helper()
				apiJSON := `{"type":"tag_cloud","title":"t","data_source":{"type":"esql","query":"FROM logs-* | STATS c = COUNT() BY h"},` +
					`"filters":[],"metric":{"column":"c","format":{"type":"number"}},` +
					`"tag_by":{"column":"h","format":{"type":"number"}},"styling":{},` +
					`"legend":{"size":"auto"},"time_range":{"from":"now-7d","to":"now"}}`
				var api kbapi.TagcloudESQL
				require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))
				var v kbapi.KbnDashboardPanelTypeVisConfig0
				require.NoError(t, v.FromTagcloudESQL(api))
				return v
			},
			want: string(kbapi.TagcloudNoESQLTypeTagCloud),
		},
		{
			name: "heatmap/no_esql",
			build: func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
				t.Helper()
				heatmap := kbapi.HeatmapNoESQL{
					Type: kbapi.HeatmapNoESQLTypeHeatmap,
					Query: kbapi.FilterSimple{
						Expression: "*",
						Language:   new(kbapi.FilterSimpleLanguage("kql")),
					},
					Axis:    kbapi.HeatmapAxes{X: kbapi.HeatmapXAxis{}, Y: kbapi.HeatmapYAxis{}},
					Styling: kbapi.HeatmapStyling{Cells: kbapi.HeatmapCells{}},
					Legend:  kbapi.HeatmapLegend{Size: kbapi.LegendSizeM},
				}
				require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"m"}`), &heatmap.DataSource))
				require.NoError(t, json.Unmarshal([]byte(`{"operation":"count"}`), &heatmap.Metric))
				require.NoError(t, json.Unmarshal([]byte(`{"operation":"filters","filters":[{"label":"All","filter":{"query":"*","language":"kql"}}]}`), &heatmap.X))
				var v kbapi.KbnDashboardPanelTypeVisConfig0
				require.NoError(t, v.FromHeatmapNoESQL(heatmap))
				return v
			},
			want: string(kbapi.HeatmapNoESQLTypeHeatmap),
		},
		{
			name: "heatmap/esql",
			build: func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
				t.Helper()
				raw := `{"type":"heatmap","title":"h","axis":{"x":{},"y":{}},"styling":{"cells":{}},"legend":{"size":"m"},` +
					`"data_source":{"type":"esql","query":"FROM logs-* | LIMIT 10"},` +
					`"metric":{"operation":"value","column":"bytes","format":{"type":"number"}},` +
					`"x":{"column":"host","format":{"type":"number"},"operation":"value"},` +
					`"y":{"column":"svc","format":{"type":"number"},"operation":"value"}}`
				var api kbapi.HeatmapESQL
				require.NoError(t, json.Unmarshal([]byte(raw), &api))
				var v kbapi.KbnDashboardPanelTypeVisConfig0
				require.NoError(t, v.FromHeatmapESQL(api))
				return v
			},
			want: string(kbapi.HeatmapNoESQLTypeHeatmap),
		},
		{
			name: "region_map/no_esql",
			build: func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
				t.Helper()
				lang := kbapi.FilterSimpleLanguage("kql")
				api := kbapi.RegionMapNoESQL{
					Type: kbapi.RegionMapNoESQLTypeRegionMap,
					Query: kbapi.FilterSimple{
						Language:   &lang,
						Expression: "*",
					},
				}
				require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"m"}`), &api.DataSource))
				require.NoError(t, json.Unmarshal([]byte(`{"operation":"count"}`), &api.Metric))
				require.NoError(t, json.Unmarshal([]byte(`{"operation":"filters","filters":[{"filter":{"query":"*","language":"kql"},"label":"A"}]}`), &api.Region))
				var v kbapi.KbnDashboardPanelTypeVisConfig0
				require.NoError(t, v.FromRegionMapNoESQL(api))
				return v
			},
			want: string(kbapi.RegionMapNoESQLTypeRegionMap),
		},
		{
			name: "region_map/esql",
			build: func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
				t.Helper()
				raw := `{"type":"region_map","title":"r","data_source":{"type":"esql","query":"FROM m | LIMIT 1"},` +
					`"metric":{"operation":"value","column":"v","format":{"type":"number"}},` +
					`"region":{"operation":"value","column":"reg","ems":{"boundaries":"world_countries","join":"name"}}}`
				var api kbapi.RegionMapESQL
				require.NoError(t, json.Unmarshal([]byte(raw), &api))
				var v kbapi.KbnDashboardPanelTypeVisConfig0
				require.NoError(t, v.FromRegionMapESQL(api))
				return v
			},
			want: string(kbapi.RegionMapNoESQLTypeRegionMap),
		},
		{
			name: "legacy_metric/no_esql_only",
			build: func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
				t.Helper()
				raw := `{"type":"legacy_metric","title":"l","data_source":{"type":"data_view_spec","index_pattern":"m"},` +
					`"query":{"language":"kql","query":"*"},"metric":{"operation":"count","format":{"type":"number"}}}`
				var api kbapi.LegacyMetricNoESQL
				require.NoError(t, json.Unmarshal([]byte(raw), &api))
				var v kbapi.KbnDashboardPanelTypeVisConfig0
				require.NoError(t, v.FromLegacyMetricNoESQL(api))
				return v
			},
			want: string(kbapi.LegacyMetric),
		},
		{
			name: "metric/no_esql",
			build: func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
				t.Helper()
				api := kbapi.MetricNoESQL{
					Type: kbapi.MetricNoESQLTypeMetric,
					Query: kbapi.FilterSimple{
						Language:   new(kbapi.FilterSimpleLanguage("kql")),
						Expression: "",
					},
					Metrics: []kbapi.MetricNoESQL_Metrics_Item{},
				}
				var v kbapi.KbnDashboardPanelTypeVisConfig0
				require.NoError(t, v.FromMetricNoESQL(api))
				return v
			},
			want: string(kbapi.MetricNoESQLTypeMetric),
		},
		{
			name: "metric/esql",
			build: func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
				t.Helper()
				api := kbapi.MetricESQL{
					Type: kbapi.MetricESQLTypeMetric,
					DataSource: kbapi.EsqlDataSource{
						Type:  kbapi.EsqlDataSourceTypeEsql,
						Query: "FROM *",
					},
					Metrics: []kbapi.MetricESQL_Metrics_Item{},
				}
				var v kbapi.KbnDashboardPanelTypeVisConfig0
				require.NoError(t, v.FromMetricESQL(api))
				return v
			},
			want: string(kbapi.MetricNoESQLTypeMetric),
		},
		{
			name: "pie/no_esql",
			build: func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
				t.Helper()
				api := kbapi.PieNoESQL{
					Type:    kbapi.PieNoESQLTypePie,
					Query:   kbapi.FilterSimple{Expression: "*", Language: new(kbapi.FilterSimpleLanguageKql)},
					Metrics: []kbapi.PieNoESQL_Metrics_Item{},
					GroupBy: new([]kbapi.PieNoESQL_GroupBy_Item{}),
				}
				require.NoError(t, json.Unmarshal([]byte(`{}`), &api.DataSource))
				var v kbapi.KbnDashboardPanelTypeVisConfig0
				require.NoError(t, v.FromPieNoESQL(api))
				return v
			},
			want: string(kbapi.PieNoESQLTypePie),
		},
		{
			name: "pie/esql",
			build: func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
				t.Helper()
				raw := `{"type":"pie","title":"p","data_source":{"type":"esql","query":"FROM logs-* | LIMIT 10"},` +
					`"legend":{"size":"auto","visibility":"visible"},` +
					`"metrics":[{"operation":"value","column":"bytes","format":{"type":"number"}}],` +
					`"group_by":[{"operation":"value","column":"h","collapse_by":"avg"}]}`
				var api kbapi.PieESQL
				require.NoError(t, json.Unmarshal([]byte(raw), &api))
				var v kbapi.KbnDashboardPanelTypeVisConfig0
				require.NoError(t, v.FromPieESQL(api))
				return v
			},
			want: string(kbapi.PieNoESQLTypePie),
		},
		{
			name: "gauge/no_esql",
			build: func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
				t.Helper()
				api := kbapi.GaugeNoESQL{Type: kbapi.GaugeNoESQLTypeGauge}
				require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"m"}`), &api.DataSource))
				require.NoError(t, json.Unmarshal([]byte(`{"expression":"*","language":"kql"}`), &api.Query))
				require.NoError(t, json.Unmarshal([]byte(`{"operation":"count"}`), &api.Metric))
				var v kbapi.KbnDashboardPanelTypeVisConfig0
				require.NoError(t, v.FromGaugeNoESQL(api))
				return v
			},
			want: string(kbapi.GaugeNoESQLTypeGauge),
		},
		{
			name: "gauge/esql",
			build: func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
				t.Helper()
				api := kbapi.GaugeESQL{Type: kbapi.GaugeESQLTypeGauge}
				require.NoError(t, json.Unmarshal([]byte(`{"type":"esql","query":"FROM *"}`), &api.DataSource))
				require.NoError(t, json.Unmarshal([]byte(`{"type":"number"}`), &api.Metric.Format))
				api.Metric.Column = "c"
				var v kbapi.KbnDashboardPanelTypeVisConfig0
				require.NoError(t, v.FromGaugeESQL(api))
				return v
			},
			want: string(kbapi.GaugeNoESQLTypeGauge),
		},
		{
			name: "waffle/no_esql",
			build: func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
				t.Helper()
				raw := `{"type":"waffle","data_source":{"type":"dataView","id":"m"},"query":{"language":"kql","query":""},"legend":{"size":"medium","visible":"auto"},"metrics":[{"operation":"count"}]}`
				var api kbapi.WaffleNoESQL
				require.NoError(t, json.Unmarshal([]byte(raw), &api))
				var v kbapi.KbnDashboardPanelTypeVisConfig0
				require.NoError(t, v.FromWaffleNoESQL(api))
				return v
			},
			want: string(kbapi.WaffleNoESQLTypeWaffle),
		},
		{
			name: "waffle/esql",
			build: func(t *testing.T) kbapi.KbnDashboardPanelTypeVisConfig0 {
				t.Helper()
				raw := `{"type":"waffle","title":"w","data_source":{"type":"esql","query":"FROM logs-* | LIMIT 5"},"legend":{"size":"s"},"metrics":[{"column":"cnt","format":{"type":"number"}}]}`
				var api kbapi.WaffleESQL
				require.NoError(t, json.Unmarshal([]byte(raw), &api))
				var v kbapi.KbnDashboardPanelTypeVisConfig0
				require.NoError(t, v.FromWaffleESQL(api))
				return v
			},
			want: string(kbapi.WaffleNoESQLTypeWaffle),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			attrs := tc.build(t)
			require.Equal(t, tc.want, DetectVizType(attrs))
		})
	}
}
