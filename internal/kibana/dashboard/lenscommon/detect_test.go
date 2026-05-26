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

	noESQLHeader := kbapi.KibanaHTTPAPIsDatatableDensity_Height_Header{}
	require.NoError(t, noESQLHeader.FromKibanaHTTPAPIsDatatableDensityHeightHeader0(kbapi.KibanaHTTPAPIsDatatableDensityHeightHeader0{Type: kbapi.KibanaHTTPAPIsDatatableDensityHeightHeader0TypeAuto}))
	noESQLValue := kbapi.KibanaHTTPAPIsDatatableDensity_Height_Value{}
	require.NoError(t, noESQLValue.FromKibanaHTTPAPIsDatatableDensityHeightValue0(kbapi.KibanaHTTPAPIsDatatableDensityHeightValue0{Type: kbapi.KibanaHTTPAPIsDatatableDensityHeightValue0TypeAuto}))
	minDatatableNoESQL := kbapi.KibanaHTTPAPIsDatatableNoESQL{
		Type:    kbapi.KibanaHTTPAPIsDatatableNoESQLTypeDataTable,
		Query:   kbapi.KibanaHTTPAPIsFilterSimple{},
		Styling: kbapi.KibanaHTTPAPIsDatatableStyling{Density: kbapi.KibanaHTTPAPIsDatatableDensity{Mode: new(kbapi.KibanaHTTPAPIsDatatableDensityModeDefault)}},
		Metrics: []kbapi.KibanaHTTPAPIsDatatableNoESQL_Metrics_Item{},
		TimeRange: func() kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema {
			var tr kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema
			require.NoError(t, json.Unmarshal([]byte(`{"from":"now-7d","to":"now"}`), &tr))
			return tr
		}(),
	}
	require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"i"}`), &minDatatableNoESQL.DataSource))
	require.NoError(t, json.Unmarshal([]byte(`{"language":"kql","expression":"*"}`), &minDatatableNoESQL.Query))
	minDatatableNoESQL.Styling.Density.Height = &struct {
		Header *kbapi.KibanaHTTPAPIsDatatableDensity_Height_Header `json:"header,omitempty"`
		Value  *kbapi.KibanaHTTPAPIsDatatableDensity_Height_Value  `json:"value,omitempty"`
	}{Header: &noESQLHeader, Value: &noESQLValue}

	tests := []struct {
		name  string
		build func(t *testing.T) VisByValueConfig0
		want  string
	}{
		{
			name: "empty_union",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				return VisByValueConfig0{}
			},
			want: "",
		},
		{
			name: "xy/no_esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				raw := `{
					"type":"xy","title":"t","axis":{"x":{},"y":{}},
					"layers":[{"type":"line","data_source":{"type":"dataView","id":"l"},"ignore_global_filters":false,"sampling":1,"y":[{"operation":"count"}]}],
					"legend":{"visibility":"visible","inside":false,"size":"auto"},
					"filters":[],"styling":{"line":{"curve":"linear"}},
					"query":{"expression":"*","language":"kql"}
				}`
				var x kbapi.KibanaHTTPAPIsXyChartNoESQL
				require.NoError(t, json.Unmarshal([]byte(raw), &x))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsXyChartNoESQL(x))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsXyChartNoESQLTypeXy),
		},
		{
			name: "xy/esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				raw := `{
					"type":"xy","title":"E","axis":{"x":{},"y":{}},"filters":[],
					"layers":[{"type":"line","data_source":{"type":"esql","query":"FROM logs-* | LIMIT 10"},
						"ignore_global_filters":false,"sampling":1,"y":[{"column":"bytes","format":{"type":"number"}}]}],
					"legend":{"visibility":"visible","inside":false,"size":"auto"},
					"styling":{"line":{"curve":"linear"}},
					"time_range":{"from":"now-7d","to":"now"}
				}`
				var x kbapi.KibanaHTTPAPIsXyChartESQL
				require.NoError(t, json.Unmarshal([]byte(raw), &x))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsXyChartESQL(x))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsXyChartNoESQLTypeXy),
		},
		{
			name: "treemap/no_esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				apiJSON := `{"type":"treemap","title":"t",` +
					`"data_source":{"type":"dataView","id":"m"},` +
					`"query":{"language":"kql","expression":""},` +
					`"legend":{"size":"small"},` +
					`"metrics":[{"operation":"count"}],` +
					`"group_by":[{"operation":"terms","field":"host.name","collapse_by":"avg"}]}`
				var api kbapi.KibanaHTTPAPIsTreemapNoESQL
				require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsTreemapNoESQL(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsTreemapNoESQLTypeTreemap),
		},
		{
			name: "treemap/esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				apiJSON := `{"type":"treemap","title":"e","description":"","ignore_global_filters":false,"sampling":1,` +
					`"data_source":{"type":"esql","query":"FROM m | LIMIT 1"},` +
					`"legend":{"size":"small"},` +
					`"metrics":[{"column":"bytes","operation":"value","format":{"type":"number"}}],` +
					`"group_by":[{"collapse_by":"avg","column":"host.name","operation":"value"}]}`
				var api kbapi.KibanaHTTPAPIsTreemapESQL
				require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsTreemapESQL(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsTreemapNoESQLTypeTreemap),
		},
		{
			name: "mosaic/no_esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				const grp = `{"mode":"categorical","palette":"default","mapping":[],"unassigned":{"type":"color_code","value":"#D3DAE6"}}`
				groupBy := `[{"operation":"terms","collapse_by":"avg","fields":["host.name"],"color":` + grp + `}]`
				apiJSON := `{"type":"mosaic","title":"m","data_source":{"type":"dataView","id":"x"},` +
					`"query":{"language":"kql","expression":""},"legend":{"size":"small"},` +
					`"metric":{"operation":"count"},"group_by":` + groupBy + `,"group_breakdown_by":` + groupBy + `}`
				var api kbapi.KibanaHTTPAPIsMosaicNoESQL
				require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsMosaicNoESQL(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsMosaicNoESQLTypeMosaic),
		},
		{
			name: "mosaic/esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				apiJSON := `{"type":"mosaic","title":"m","data_source":{"type":"esql","query":"FROM m | LIMIT 1"},` +
					`"legend":{"size":"small"},` +
					`"metric":{"column":"bytes","operation":"value","format":{"type":"number"}},` +
					`"group_by":[{"collapse_by":"avg","column":"host.name","operation":"value"}],` +
					`"group_breakdown_by":[{"collapse_by":"avg","column":"s","operation":"value"}]}`
				var api kbapi.KibanaHTTPAPIsMosaicESQL
				require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsMosaicESQL(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsMosaicNoESQLTypeMosaic),
		},
		{
			name: "datatable/no_esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsDatatableNoESQL(minDatatableNoESQL))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsDatatableNoESQLTypeDataTable),
		},
		{
			name: "datatable/esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				apiJSON := `{"type":"data_table","title":"d","data_source":{"type":"esql","query":"FROM logs-* | LIMIT 5"},` +
					`"filters":[],"metrics":[{"column":"c","operation":"value","format":{"type":"number"}}],` +
					`"rows":[{"column":"r","collapse_by":"avg","format":{"type":"number"}}],` +
					`"styling":{"density":{"mode":"default","height":{"header":{"type":"auto"},"value":{"type":"auto"}}}},` +
					`"time_range":{"from":"now-7d","to":"now"}}`
				var api kbapi.KibanaHTTPAPIsDatatableESQL
				require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsDatatableESQL(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsDatatableNoESQLTypeDataTable),
		},
		{
			name: "tagcloud/no_esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				api := kbapi.KibanaHTTPAPIsTagcloudNoESQL{
					Type: kbapi.KibanaHTTPAPIsTagcloudNoESQLTypeTagCloud,
				}
				require.NoError(t, json.Unmarshal([]byte(`{"index":"i"}`), &api.DataSource))
				require.NoError(t, json.Unmarshal([]byte(`{"expression":"*","language":"kql"}`), &api.Query))
				require.NoError(t, json.Unmarshal([]byte(`{"operation":{"operation_type":"count"}}`), &api.Metric))
				require.NoError(t, json.Unmarshal([]byte(`{"operation":{"operation_type":"terms"},"field":"t"}`), &api.TagBy))
				require.NoError(t, json.Unmarshal([]byte(`{}`), &api.Styling))
				require.NoError(t, json.Unmarshal([]byte(`[]`), &api.Filters))
				var tr kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema
				require.NoError(t, json.Unmarshal([]byte(`{"from":"now-7d","to":"now"}`), &tr))
				api.TimeRange = tr
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsTagcloudNoESQL(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsTagcloudNoESQLTypeTagCloud),
		},
		{
			name: "tagcloud/esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				apiJSON := `{"type":"tag_cloud","title":"t","data_source":{"type":"esql","query":"FROM logs-* | STATS c = COUNT() BY h"},` +
					`"filters":[],"metric":{"column":"c","format":{"type":"number"}},` +
					`"tag_by":{"column":"h","format":{"type":"number"}},"styling":{},` +
					`"legend":{"size":"auto"},"time_range":{"from":"now-7d","to":"now"}}`
				var api kbapi.KibanaHTTPAPIsTagcloudESQL
				require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsTagcloudESQL(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsTagcloudNoESQLTypeTagCloud),
		},
		{
			name: "heatmap/no_esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				heatmap := kbapi.KibanaHTTPAPIsHeatmapNoESQL{
					Type: kbapi.KibanaHTTPAPIsHeatmapNoESQLTypeHeatmap,
					Query: kbapi.KibanaHTTPAPIsFilterSimple{
						Expression: "*",
						Language:   new(kbapi.KibanaHTTPAPIsFilterSimpleLanguage("kql")),
					},
					Axis:    kbapi.KibanaHTTPAPIsHeatmapAxes{X: kbapi.KibanaHTTPAPIsHeatmapXAxis{}, Y: kbapi.KibanaHTTPAPIsHeatmapYAxis{}},
					Styling: kbapi.KibanaHTTPAPIsHeatmapStyling{Cells: kbapi.KibanaHTTPAPIsHeatmapCells{}},
					Legend:  kbapi.KibanaHTTPAPIsHeatmapLegend{Size: kbapi.KibanaHTTPAPIsLegendSizeM},
				}
				require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"m"}`), &heatmap.DataSource))
				require.NoError(t, json.Unmarshal([]byte(`{"operation":"count"}`), &heatmap.Metric))
				require.NoError(t, json.Unmarshal([]byte(`{"operation":"filters","filters":[{"label":"All","filter":{"query":"*","language":"kql"}}]}`), &heatmap.X))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsHeatmapNoESQL(heatmap))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsHeatmapNoESQLTypeHeatmap),
		},
		{
			name: "heatmap/esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				raw := `{"type":"heatmap","title":"h","axis":{"x":{},"y":{}},"styling":{"cells":{}},"legend":{"size":"m"},` +
					`"data_source":{"type":"esql","query":"FROM logs-* | LIMIT 10"},` +
					`"metric":{"operation":"value","column":"bytes","format":{"type":"number"}},` +
					`"x":{"column":"host","format":{"type":"number"},"operation":"value"},` +
					`"y":{"column":"svc","format":{"type":"number"},"operation":"value"}}`
				var api kbapi.KibanaHTTPAPIsHeatmapESQL
				require.NoError(t, json.Unmarshal([]byte(raw), &api))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsHeatmapESQL(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsHeatmapNoESQLTypeHeatmap),
		},
		{
			name: "region_map/no_esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				lang := kbapi.KibanaHTTPAPIsFilterSimpleLanguage("kql")
				api := kbapi.KibanaHTTPAPIsRegionMapNoESQL{
					Type: kbapi.KibanaHTTPAPIsRegionMapNoESQLTypeRegionMap,
					Query: kbapi.KibanaHTTPAPIsFilterSimple{
						Language:   &lang,
						Expression: "*",
					},
				}
				require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"m"}`), &api.DataSource))
				require.NoError(t, json.Unmarshal([]byte(`{"operation":"count"}`), &api.Metric))
				require.NoError(t, json.Unmarshal([]byte(`{"operation":"filters","filters":[{"filter":{"query":"*","language":"kql"},"label":"A"}]}`), &api.Region))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsRegionMapNoESQL(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsRegionMapNoESQLTypeRegionMap),
		},
		{
			name: "region_map/esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				raw := `{"type":"region_map","title":"r","data_source":{"type":"esql","query":"FROM m | LIMIT 1"},` +
					`"metric":{"operation":"value","column":"v","format":{"type":"number"}},` +
					`"region":{"operation":"value","column":"reg","ems":{"boundaries":"world_countries","join":"name"}}}`
				var api kbapi.KibanaHTTPAPIsRegionMapESQL
				require.NoError(t, json.Unmarshal([]byte(raw), &api))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsRegionMapESQL(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsRegionMapNoESQLTypeRegionMap),
		},
		{
			name: "legacy_metric/no_esql_only",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				raw := `{"type":"legacy_metric","title":"l","data_source":{"type":"data_view_spec","index_pattern":"m"},` +
					`"query":{"language":"kql","query":"*"},"metric":{"operation":"count","format":{"type":"number"}}}`
				var api kbapi.KibanaHTTPAPIsLegacyMetricNoESQL
				require.NoError(t, json.Unmarshal([]byte(raw), &api))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsLegacyMetricNoESQL(api))
				return v
			},
			want: string(kbapi.LegacyMetric),
		},
		{
			name: "metric/no_esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				api := kbapi.KibanaHTTPAPIsMetricNoESQL{
					Type: kbapi.KibanaHTTPAPIsMetricNoESQLTypeMetric,
					Query: kbapi.KibanaHTTPAPIsFilterSimple{
						Language:   new(kbapi.KibanaHTTPAPIsFilterSimpleLanguage("kql")),
						Expression: "",
					},
					Metrics: []kbapi.KibanaHTTPAPIsMetricNoESQL_Metrics_Item{},
				}
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsMetricNoESQL(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsMetricNoESQLTypeMetric),
		},
		{
			name: "metric/esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				api := kbapi.KibanaHTTPAPIsMetricESQL{
					Type: kbapi.KibanaHTTPAPIsMetricESQLTypeMetric,
					DataSource: kbapi.KibanaHTTPAPIsEsqlDataSource{
						Type:  kbapi.KibanaHTTPAPIsEsqlDataSourceTypeEsql,
						Query: "FROM *",
					},
					Metrics: []kbapi.KibanaHTTPAPIsMetricESQL_Metrics_Item{},
				}
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsMetricESQL(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsMetricNoESQLTypeMetric),
		},
		{
			name: "pie/no_esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				api := kbapi.KibanaHTTPAPIsPieNoESQL{
					Type:    kbapi.KibanaHTTPAPIsPieNoESQLTypePie,
					Query:   kbapi.KibanaHTTPAPIsFilterSimple{Expression: "*", Language: new(kbapi.KibanaHTTPAPIsFilterSimpleLanguageKql)},
					Metrics: []kbapi.KibanaHTTPAPIsPieNoESQL_Metrics_Item{},
					GroupBy: new([]kbapi.KibanaHTTPAPIsPieNoESQL_GroupBy_Item{}),
				}
				require.NoError(t, json.Unmarshal([]byte(`{}`), &api.DataSource))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsPieNoESQL(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsPieNoESQLTypePie),
		},
		{
			name: "pie/esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				raw := `{"type":"pie","title":"p","data_source":{"type":"esql","query":"FROM logs-* | LIMIT 10"},` +
					`"legend":{"size":"auto","visibility":"visible"},` +
					`"metrics":[{"operation":"value","column":"bytes","format":{"type":"number"}}],` +
					`"group_by":[{"operation":"value","column":"h","collapse_by":"avg"}]}`
				var api kbapi.KibanaHTTPAPIsPieESQL
				require.NoError(t, json.Unmarshal([]byte(raw), &api))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsPieESQL(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsPieNoESQLTypePie),
		},
		{
			name: "gauge/no_esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				api := kbapi.KibanaHTTPAPIsGaugeNoESQL{Type: kbapi.KibanaHTTPAPIsGaugeNoESQLTypeGauge}
				require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"m"}`), &api.DataSource))
				require.NoError(t, json.Unmarshal([]byte(`{"expression":"*","language":"kql"}`), &api.Query))
				require.NoError(t, json.Unmarshal([]byte(`{"operation":"count"}`), &api.Metric))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsGaugeNoESQL(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsGaugeNoESQLTypeGauge),
		},
		{
			name: "gauge/esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				api := kbapi.KibanaHTTPAPIsGaugeESQL{Type: kbapi.KibanaHTTPAPIsGaugeESQLTypeGauge}
				require.NoError(t, json.Unmarshal([]byte(`{"type":"esql","query":"FROM *"}`), &api.DataSource))
				require.NoError(t, json.Unmarshal([]byte(`{"type":"number"}`), &api.Metric.Format))
				api.Metric.Column = "c"
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsGaugeESQL(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsGaugeNoESQLTypeGauge),
		},
		{
			name: "waffle/no_esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				raw := `{"type":"waffle","data_source":{"type":"dataView","id":"m"},"query":{"language":"kql","query":""},"legend":{"size":"medium","visible":"auto"},"metrics":[{"operation":"count"}]}`
				var api kbapi.KibanaHTTPAPIsWaffleNoESQL
				require.NoError(t, json.Unmarshal([]byte(raw), &api))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsWaffleNoESQL(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsWaffleNoESQLTypeWaffle),
		},
		{
			name: "waffle/esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				raw := `{"type":"waffle","title":"w","data_source":{"type":"esql","query":"FROM logs-* | LIMIT 5"},"legend":{"size":"s"},"metrics":[{"column":"cnt","format":{"type":"number"}}]}`
				var api kbapi.KibanaHTTPAPIsWaffleESQL
				require.NoError(t, json.Unmarshal([]byte(raw), &api))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsWaffleESQL(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsWaffleNoESQLTypeWaffle),
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
