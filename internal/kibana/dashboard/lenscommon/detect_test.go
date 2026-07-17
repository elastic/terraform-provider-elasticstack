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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHasLensByReferenceShapeAtRoot_refIDOnly(t *testing.T) {
	t.Parallel()
	assert.True(t, HasLensByReferenceShapeAtRoot(map[string]any{"ref_id": "panel_0"}))
	assert.False(t, HasLensByReferenceShapeAtRoot(map[string]any{"ref_id": ""}))
	assert.False(t, HasLensByReferenceShapeAtRoot(map[string]any{"time_range": map[string]any{"from": "now-7d", "to": "now"}}))
}

func TestDetectVizType_chartKindsPerArm(t *testing.T) {
	t.Parallel()

	noESQLHeader := kbapi.KibanaHTTPAPIsDatatableDensity_Height_Header{}
	require.NoError(t, noESQLHeader.FromKibanaHTTPAPIsDatatableDensityHeightHeader0(kbapi.KibanaHTTPAPIsDatatableDensityHeightHeader0{Type: kbapi.KibanaHTTPAPIsDatatableDensityHeightHeader0TypeAuto}))
	noESQLValue := kbapi.KibanaHTTPAPIsDatatableDensity_Height_Value{}
	require.NoError(t, noESQLValue.FromKibanaHTTPAPIsDatatableDensityHeightValue0(kbapi.KibanaHTTPAPIsDatatableDensityHeightValue0{Type: kbapi.KibanaHTTPAPIsDatatableDensityHeightValue0TypeAuto}))
	minDatatableNoESQL := kbapi.KibanaHTTPAPIsDatatableNoESQLByValuePanel{
		Type:    kbapi.KibanaHTTPAPIsDatatableNoESQLByValuePanelTypeDataTable,
		Query:   &kbapi.KibanaHTTPAPIsFilterSimple{},
		Styling: &kbapi.KibanaHTTPAPIsDatatableStyling{Density: &kbapi.KibanaHTTPAPIsDatatableDensity{Mode: new(kbapi.KibanaHTTPAPIsDatatableDensityModeDefault)}},
		Metrics: []kbapi.KibanaHTTPAPIsDatatableNoESQLByValuePanel_Metrics_Item{},
		TimeRange: func() *kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema {
			var tr kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema
			require.NoError(t, json.Unmarshal([]byte(`{"from":"now-7d","to":"now"}`), &tr))
			return &tr
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
				var x kbapi.KibanaHTTPAPIsXyChartNoESQLByValuePanel
				require.NoError(t, json.Unmarshal([]byte(raw), &x))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsXyChartNoESQLByValuePanel(x))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsXyChartNoESQLByValuePanelTypeXy),
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
				var x kbapi.KibanaHTTPAPIsXyChartESQLByValuePanel
				require.NoError(t, json.Unmarshal([]byte(raw), &x))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsXyChartESQLByValuePanel(x))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsXyChartNoESQLByValuePanelTypeXy),
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
				var api kbapi.KibanaHTTPAPIsTreemapNoESQLByValuePanel
				require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsTreemapNoESQLByValuePanel(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsTreemapNoESQLByValuePanelTypeTreemap),
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
				var api kbapi.KibanaHTTPAPIsTreemapESQLByValuePanel
				require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsTreemapESQLByValuePanel(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsTreemapNoESQLByValuePanelTypeTreemap),
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
				var api kbapi.KibanaHTTPAPIsMosaicNoESQLByValuePanel
				require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsMosaicNoESQLByValuePanel(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsMosaicNoESQLByValuePanelTypeMosaic),
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
				var api kbapi.KibanaHTTPAPIsMosaicESQLByValuePanel
				require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsMosaicESQLByValuePanel(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsMosaicNoESQLByValuePanelTypeMosaic),
		},
		{
			name: "datatable/no_esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsDatatableNoESQLByValuePanel(minDatatableNoESQL))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsDatatableNoESQLByValuePanelTypeDataTable),
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
				var api kbapi.KibanaHTTPAPIsDatatableESQLByValuePanel
				require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsDatatableESQLByValuePanel(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsDatatableNoESQLByValuePanelTypeDataTable),
		},
		{
			name: "tagcloud/no_esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				api := kbapi.KibanaHTTPAPIsTagcloudNoESQLByValuePanel{
					Type: kbapi.KibanaHTTPAPIsTagcloudNoESQLByValuePanelTypeTagCloud,
				}
				require.NoError(t, json.Unmarshal([]byte(`{"index":"i"}`), &api.DataSource))
				require.NoError(t, json.Unmarshal([]byte(`{"expression":"*","language":"kql"}`), &api.Query))
				require.NoError(t, json.Unmarshal([]byte(`{"operation":{"operation_type":"count"}}`), &api.Metric))
				require.NoError(t, json.Unmarshal([]byte(`{"operation":{"operation_type":"terms"},"field":"t"}`), &api.TagBy))
				require.NoError(t, json.Unmarshal([]byte(`{}`), &api.Styling))
				require.NoError(t, json.Unmarshal([]byte(`[]`), &api.Filters))
				var tr kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema
				require.NoError(t, json.Unmarshal([]byte(`{"from":"now-7d","to":"now"}`), &tr))
				api.TimeRange = &tr
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsTagcloudNoESQLByValuePanel(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsTagcloudNoESQLByValuePanelTypeTagCloud),
		},
		{
			name: "tagcloud/esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				apiJSON := `{"type":"tag_cloud","title":"t","data_source":{"type":"esql","query":"FROM logs-* | STATS c = COUNT() BY h"},` +
					`"filters":[],"metric":{"column":"c","format":{"type":"number"}},` +
					`"tag_by":{"column":"h","format":{"type":"number"}},"styling":{},` +
					`"legend":{"size":"auto"},"time_range":{"from":"now-7d","to":"now"}}`
				var api kbapi.KibanaHTTPAPIsTagcloudESQLByValuePanel
				require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsTagcloudESQLByValuePanel(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsTagcloudNoESQLByValuePanelTypeTagCloud),
		},
		{
			name: "heatmap/no_esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				heatmap := kbapi.KibanaHTTPAPIsHeatmapNoESQLByValuePanel{
					Type: kbapi.KibanaHTTPAPIsHeatmapNoESQLByValuePanelTypeHeatmap,
					Query: &kbapi.KibanaHTTPAPIsFilterSimple{
						Expression: "*",
						Language:   new(kbapi.KibanaHTTPAPIsFilterSimpleLanguage("kql")),
					},
					Axis:    &kbapi.KibanaHTTPAPIsHeatmapAxes{X: &kbapi.KibanaHTTPAPIsHeatmapXAxis{}, Y: &kbapi.KibanaHTTPAPIsHeatmapYAxis{}},
					Styling: &kbapi.KibanaHTTPAPIsHeatmapStyling{Cells: &kbapi.KibanaHTTPAPIsHeatmapCells{}},
					Legend:  &kbapi.KibanaHTTPAPIsHeatmapLegend{Size: new(kbapi.KibanaHTTPAPIsLegendSizeM)},
				}
				require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"m"}`), &heatmap.DataSource))
				require.NoError(t, json.Unmarshal([]byte(`{"operation":"count"}`), &heatmap.Metric))
				require.NoError(t, json.Unmarshal([]byte(`{"operation":"filters","filters":[{"label":"All","filter":{"query":"*","language":"kql"}}]}`), &heatmap.X))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsHeatmapNoESQLByValuePanel(heatmap))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsHeatmapNoESQLByValuePanelTypeHeatmap),
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
				var api kbapi.KibanaHTTPAPIsHeatmapESQLByValuePanel
				require.NoError(t, json.Unmarshal([]byte(raw), &api))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsHeatmapESQLByValuePanel(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsHeatmapNoESQLByValuePanelTypeHeatmap),
		},
		{
			name: "region_map/no_esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				lang := kbapi.KibanaHTTPAPIsFilterSimpleLanguage("kql")
				api := kbapi.KibanaHTTPAPIsRegionMapNoESQLByValuePanel{
					Type: kbapi.KibanaHTTPAPIsRegionMapNoESQLByValuePanelTypeRegionMap,
					Query: &kbapi.KibanaHTTPAPIsFilterSimple{
						Language:   &lang,
						Expression: "*",
					},
				}
				require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"m"}`), &api.DataSource))
				require.NoError(t, json.Unmarshal([]byte(`{"operation":"count"}`), &api.Metric))
				require.NoError(t, json.Unmarshal([]byte(`{"operation":"filters","filters":[{"filter":{"query":"*","language":"kql"},"label":"A"}]}`), &api.Region))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsRegionMapNoESQLByValuePanel(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsRegionMapNoESQLByValuePanelTypeRegionMap),
		},
		{
			name: "region_map/esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				raw := `{"type":"region_map","title":"r","data_source":{"type":"esql","query":"FROM m | LIMIT 1"},` +
					`"metric":{"operation":"value","column":"v","format":{"type":"number"}},` +
					`"region":{"operation":"value","column":"reg","ems":{"boundaries":"world_countries","join":"name"}}}`
				var api kbapi.KibanaHTTPAPIsRegionMapESQLByValuePanel
				require.NoError(t, json.Unmarshal([]byte(raw), &api))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsRegionMapESQLByValuePanel(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsRegionMapNoESQLByValuePanelTypeRegionMap),
		},
		{
			name: "legacy_metric/no_esql_only",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				raw := `{"type":"legacy_metric","title":"l","data_source":{"type":"data_view_spec","index_pattern":"m"},` +
					`"query":{"language":"kql","query":"*"},"metric":{"operation":"count","format":{"type":"number"}}}`
				var api kbapi.KibanaHTTPAPIsLegacyMetricNoESQLByValuePanel
				require.NoError(t, json.Unmarshal([]byte(raw), &api))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsLegacyMetricNoESQLByValuePanel(api))
				return v
			},
			want: string(kbapi.LegacyMetric),
		},
		{
			name: "metric/no_esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				api := kbapi.KibanaHTTPAPIsMetricNoESQLByValuePanel{
					Type: kbapi.KibanaHTTPAPIsMetricNoESQLByValuePanelTypeMetric,
					Query: &kbapi.KibanaHTTPAPIsFilterSimple{
						Language:   new(kbapi.KibanaHTTPAPIsFilterSimpleLanguage("kql")),
						Expression: "",
					},
					Metrics: []kbapi.KibanaHTTPAPIsMetricNoESQLByValuePanel_Metrics_Item{},
				}
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsMetricNoESQLByValuePanel(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsMetricNoESQLByValuePanelTypeMetric),
		},
		{
			name: "metric/esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				api := kbapi.KibanaHTTPAPIsMetricESQLByValuePanel{
					Type: kbapi.KibanaHTTPAPIsMetricESQLByValuePanelTypeMetric,
					DataSource: kbapi.KibanaHTTPAPIsEsqlDataSource{
						Type:  kbapi.KibanaHTTPAPIsEsqlDataSourceTypeEsql,
						Query: "FROM *",
					},
					Metrics: []kbapi.KibanaHTTPAPIsMetricESQLByValuePanel_Metrics_Item{},
				}
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsMetricESQLByValuePanel(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsMetricNoESQLByValuePanelTypeMetric),
		},
		{
			name: "pie/no_esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				api := kbapi.KibanaHTTPAPIsPieNoESQLByValuePanel{
					Type:    kbapi.KibanaHTTPAPIsPieNoESQLByValuePanelTypePie,
					Query:   &kbapi.KibanaHTTPAPIsFilterSimple{Expression: "*", Language: new(kbapi.KibanaHTTPAPIsFilterSimpleLanguageKql)},
					Styling: &kbapi.KibanaHTTPAPIsPieStyling{},
					Metrics: []kbapi.KibanaHTTPAPIsPieNoESQLByValuePanel_Metrics_Item{},
					GroupBy: new([]kbapi.KibanaHTTPAPIsPieNoESQLByValuePanel_GroupBy_Item{}),
				}
				require.NoError(t, json.Unmarshal([]byte(`{}`), &api.DataSource))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsPieNoESQLByValuePanel(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsPieNoESQLByValuePanelTypePie),
		},
		{
			name: "pie/esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				raw := `{"type":"pie","title":"p","data_source":{"type":"esql","query":"FROM logs-* | LIMIT 10"},` +
					`"legend":{"size":"auto","visibility":"visible"},` +
					`"metrics":[{"operation":"value","column":"bytes","format":{"type":"number"}}],` +
					`"group_by":[{"operation":"value","column":"h","collapse_by":"avg"}]}`
				var api kbapi.KibanaHTTPAPIsPieESQLByValuePanel
				require.NoError(t, json.Unmarshal([]byte(raw), &api))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsPieESQLByValuePanel(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsPieNoESQLByValuePanelTypePie),
		},
		{
			name: "gauge/no_esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				api := kbapi.KibanaHTTPAPIsGaugeNoESQLByValuePanel{Type: kbapi.KibanaHTTPAPIsGaugeNoESQLByValuePanelTypeGauge}
				require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"m"}`), &api.DataSource))
				require.NoError(t, json.Unmarshal([]byte(`{"expression":"*","language":"kql"}`), &api.Query))
				require.NoError(t, json.Unmarshal([]byte(`{"operation":"count"}`), &api.Metric))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsGaugeNoESQLByValuePanel(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsGaugeNoESQLByValuePanelTypeGauge),
		},
		{
			name: "gauge/esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				api := kbapi.KibanaHTTPAPIsGaugeESQLByValuePanel{Type: kbapi.KibanaHTTPAPIsGaugeESQLByValuePanelTypeGauge}
				require.NoError(t, json.Unmarshal([]byte(`{"type":"esql","query":"FROM *"}`), &api.DataSource))
				require.NoError(t, json.Unmarshal([]byte(`{"type":"number"}`), &api.Metric.Format))
				api.Metric.Column = "c"
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsGaugeESQLByValuePanel(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsGaugeNoESQLByValuePanelTypeGauge),
		},
		{
			name: "waffle/no_esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				raw := `{"type":"waffle","data_source":{"type":"dataView","id":"m"},` +
					`"query":{"language":"kql","query":""},` +
					`"legend":{"size":"medium","visible":"auto"},"styling":{"values":{}},` +
					`"metrics":[{"operation":"count"}]}`
				var api kbapi.KibanaHTTPAPIsWaffleNoESQLByValuePanel
				require.NoError(t, json.Unmarshal([]byte(raw), &api))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsWaffleNoESQLByValuePanel(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsWaffleNoESQLByValuePanelTypeWaffle),
		},
		{
			name: "waffle/esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				raw := `{"type":"waffle","title":"w","data_source":{"type":"esql","query":"FROM logs-* | LIMIT 5"},"legend":{"size":"s"},"metrics":[{"column":"cnt","format":{"type":"number"}}]}`
				var api kbapi.KibanaHTTPAPIsWaffleESQLByValuePanel
				require.NoError(t, json.Unmarshal([]byte(raw), &api))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsWaffleESQLByValuePanel(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsWaffleNoESQLByValuePanelTypeWaffle),
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
