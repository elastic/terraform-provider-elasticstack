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

	noESQLHeader := kbapi.KibanaHTTPAPIsVisDatatableDensity_Height_Header{}
	require.NoError(t, noESQLHeader.FromKibanaHTTPAPIsVisDatatableDensityHeightHeader0(kbapi.KibanaHTTPAPIsVisDatatableDensityHeightHeader0{Type: kbapi.KibanaHTTPAPIsVisDatatableDensityHeightHeader0TypeAuto}))
	noESQLValue := kbapi.KibanaHTTPAPIsVisDatatableDensity_Height_Value{}
	require.NoError(t, noESQLValue.FromKibanaHTTPAPIsVisDatatableDensityHeightValue0(kbapi.KibanaHTTPAPIsVisDatatableDensityHeightValue0{Type: kbapi.KibanaHTTPAPIsVisDatatableDensityHeightValue0TypeAuto}))
	minDatatableNoESQL := kbapi.KibanaHTTPAPIsVisDatatableNoESQLByValuePanel{
		Type:    kbapi.KibanaHTTPAPIsVisDatatableNoESQLByValuePanelTypeDataTable,
		Query:   &kbapi.KibanaHTTPAPIsVisFilterSimple{},
		Styling: &kbapi.KibanaHTTPAPIsVisDatatableStyling{Density: &kbapi.KibanaHTTPAPIsVisDatatableDensity{Mode: new(kbapi.KibanaHTTPAPIsVisDatatableDensityModeDefault)}},
		Metrics: nil,
		TimeRange: func() *kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema {
			var tr kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema
			require.NoError(t, json.Unmarshal([]byte(`{"from":"now-7d","to":"now"}`), &tr))
			return &tr
		}(),
	}
	require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"i"}`), &minDatatableNoESQL.DataSource))
	require.NoError(t, json.Unmarshal([]byte(`{"language":"kql","expression":"*"}`), &minDatatableNoESQL.Query))
	minDatatableNoESQL.Styling.Density.Height = &struct {
		Header *kbapi.KibanaHTTPAPIsVisDatatableDensity_Height_Header `json:"header,omitempty"`
		Value  *kbapi.KibanaHTTPAPIsVisDatatableDensity_Height_Value  `json:"value,omitempty"`
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
				var x kbapi.KibanaHTTPAPIsVisXyChartNoESQLByValuePanel
				require.NoError(t, json.Unmarshal([]byte(raw), &x))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsVisXyChartNoESQLByValuePanel(x))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsVisXyChartNoESQLByValuePanelTypeXy),
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
				var x kbapi.KibanaHTTPAPIsVisXyChartESQLByValuePanel
				require.NoError(t, json.Unmarshal([]byte(raw), &x))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsVisXyChartESQLByValuePanel(x))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsVisXyChartNoESQLByValuePanelTypeXy),
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
				var api kbapi.KibanaHTTPAPIsVisTreemapNoESQLByValuePanel
				require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsVisTreemapNoESQLByValuePanel(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsVisTreemapNoESQLByValuePanelTypeTreemap),
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
				var api kbapi.KibanaHTTPAPIsVisTreemapESQLByValuePanel
				require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsVisTreemapESQLByValuePanel(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsVisTreemapNoESQLByValuePanelTypeTreemap),
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
				var api kbapi.KibanaHTTPAPIsVisMosaicNoESQLByValuePanel
				require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsVisMosaicNoESQLByValuePanel(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsVisMosaicNoESQLByValuePanelTypeMosaic),
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
				var api kbapi.KibanaHTTPAPIsVisMosaicESQLByValuePanel
				require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsVisMosaicESQLByValuePanel(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsVisMosaicNoESQLByValuePanelTypeMosaic),
		},
		{
			name: "datatable/no_esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsVisDatatableNoESQLByValuePanel(minDatatableNoESQL))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsVisDatatableNoESQLByValuePanelTypeDataTable),
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
				var api kbapi.KibanaHTTPAPIsVisDatatableESQLByValuePanel
				require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsVisDatatableESQLByValuePanel(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsVisDatatableNoESQLByValuePanelTypeDataTable),
		},
		{
			name: "tagcloud/no_esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				api := kbapi.KibanaHTTPAPIsVisTagcloudNoESQLByValuePanel{
					Type: kbapi.KibanaHTTPAPIsVisTagcloudNoESQLByValuePanelTypeTagCloud,
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
				require.NoError(t, v.FromKibanaHTTPAPIsVisTagcloudNoESQLByValuePanel(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsVisTagcloudNoESQLByValuePanelTypeTagCloud),
		},
		{
			name: "tagcloud/esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				apiJSON := `{"type":"tag_cloud","title":"t","data_source":{"type":"esql","query":"FROM logs-* | STATS c = COUNT() BY h"},` +
					`"filters":[],"metric":{"column":"c","format":{"type":"number"}},` +
					`"tag_by":{"column":"h","format":{"type":"number"}},"styling":{},` +
					`"legend":{"size":"auto"},"time_range":{"from":"now-7d","to":"now"}}`
				var api kbapi.KibanaHTTPAPIsVisTagcloudESQLByValuePanel
				require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsVisTagcloudESQLByValuePanel(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsVisTagcloudNoESQLByValuePanelTypeTagCloud),
		},
		{
			name: "heatmap/no_esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				heatmap := kbapi.KibanaHTTPAPIsVisHeatmapNoESQLByValuePanel{
					Type: kbapi.KibanaHTTPAPIsVisHeatmapNoESQLByValuePanelTypeHeatmap,
					Query: &kbapi.KibanaHTTPAPIsVisFilterSimple{
						Expression: "*",
						Language:   new(kbapi.KibanaHTTPAPIsVisFilterSimpleLanguage("kql")),
					},
					Axis:    &kbapi.KibanaHTTPAPIsVisHeatmapAxes{X: &kbapi.KibanaHTTPAPIsVisHeatmapXAxis{}, Y: &kbapi.KibanaHTTPAPIsVisHeatmapYAxis{}},
					Styling: &kbapi.KibanaHTTPAPIsVisHeatmapStyling{Cells: &kbapi.KibanaHTTPAPIsVisHeatmapCells{}},
					Legend:  &kbapi.KibanaHTTPAPIsVisHeatmapLegend{Size: new(kbapi.KibanaHTTPAPIsVisLegendSizeM)},
				}
				require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"m"}`), &heatmap.DataSource))
				require.NoError(t, json.Unmarshal([]byte(`{"operation":"count"}`), &heatmap.Metric))
				require.NoError(t, json.Unmarshal([]byte(`{"operation":"filters","filters":[{"label":"All","filter":{"query":"*","language":"kql"}}]}`), &heatmap.X))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsVisHeatmapNoESQLByValuePanel(heatmap))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsVisHeatmapNoESQLByValuePanelTypeHeatmap),
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
				var api kbapi.KibanaHTTPAPIsVisHeatmapESQLByValuePanel
				require.NoError(t, json.Unmarshal([]byte(raw), &api))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsVisHeatmapESQLByValuePanel(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsVisHeatmapNoESQLByValuePanelTypeHeatmap),
		},
		{
			name: "region_map/no_esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				lang := kbapi.KibanaHTTPAPIsVisFilterSimpleLanguage("kql")
				api := kbapi.KibanaHTTPAPIsVisRegionMapNoESQLByValuePanel{
					Type: kbapi.KibanaHTTPAPIsVisRegionMapNoESQLByValuePanelTypeRegionMap,
					Query: &kbapi.KibanaHTTPAPIsVisFilterSimple{
						Language:   &lang,
						Expression: "*",
					},
				}
				require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"m"}`), &api.DataSource))
				require.NoError(t, json.Unmarshal([]byte(`{"operation":"count"}`), &api.Metric))
				require.NoError(t, json.Unmarshal([]byte(`{"operation":"filters","filters":[{"filter":{"query":"*","language":"kql"},"label":"A"}]}`), &api.Region))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsVisRegionMapNoESQLByValuePanel(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsVisRegionMapNoESQLByValuePanelTypeRegionMap),
		},
		{
			name: "region_map/esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				raw := `{"type":"region_map","title":"r","data_source":{"type":"esql","query":"FROM m | LIMIT 1"},` +
					`"metric":{"operation":"value","column":"v","format":{"type":"number"}},` +
					`"region":{"operation":"value","column":"reg","ems":{"boundaries":"world_countries","join":"name"}}}`
				var api kbapi.KibanaHTTPAPIsVisRegionMapESQLByValuePanel
				require.NoError(t, json.Unmarshal([]byte(raw), &api))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsVisRegionMapESQLByValuePanel(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsVisRegionMapNoESQLByValuePanelTypeRegionMap),
		},
		{
			name: "legacy_metric/no_esql_only",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				raw := `{"type":"legacy_metric","title":"l","data_source":{"type":"data_view_spec","index_pattern":"m"},` +
					`"query":{"language":"kql","query":"*"},"metric":{"operation":"count","format":{"type":"number"}}}`
				var api kbapi.KibanaHTTPAPIsVisLegacyMetricNoESQLByValuePanel
				require.NoError(t, json.Unmarshal([]byte(raw), &api))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsVisLegacyMetricNoESQLByValuePanel(api))
				return v
			},
			want: string(kbapi.LegacyMetric),
		},
		{
			name: "metric/no_esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				api := kbapi.KibanaHTTPAPIsVisMetricNoESQLByValuePanel{
					Type: kbapi.KibanaHTTPAPIsVisMetricNoESQLByValuePanelTypeMetric,
					Query: &kbapi.KibanaHTTPAPIsVisFilterSimple{
						Language:   new(kbapi.KibanaHTTPAPIsVisFilterSimpleLanguage("kql")),
						Expression: "",
					},
					Metrics: []kbapi.KibanaHTTPAPIsVisMetricNoESQLByValuePanel_Metrics_Item{},
				}
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsVisMetricNoESQLByValuePanel(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsVisMetricNoESQLByValuePanelTypeMetric),
		},
		{
			name: "metric/esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				api := kbapi.KibanaHTTPAPIsVisMetricESQLByValuePanel{
					Type: kbapi.KibanaHTTPAPIsVisMetricESQLByValuePanelTypeMetric,
					DataSource: kbapi.KibanaHTTPAPIsEsqlDataSource{
						Type:  kbapi.KibanaHTTPAPIsEsqlDataSourceTypeEsql,
						Query: "FROM *",
					},
					Metrics: []kbapi.KibanaHTTPAPIsVisMetricESQLByValuePanel_Metrics_Item{},
				}
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsVisMetricESQLByValuePanel(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsVisMetricNoESQLByValuePanelTypeMetric),
		},
		{
			name: "pie/no_esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				api := kbapi.KibanaHTTPAPIsVisPieNoESQLByValuePanel{
					Type:    kbapi.KibanaHTTPAPIsVisPieNoESQLByValuePanelTypePie,
					Query:   &kbapi.KibanaHTTPAPIsVisFilterSimple{Expression: "*", Language: new(kbapi.KibanaHTTPAPIsVisFilterSimpleLanguageKql)},
					Styling: &kbapi.KibanaHTTPAPIsVisPieStyling{},
					Metrics: nil,
				}
				require.NoError(t, json.Unmarshal([]byte(`{}`), &api.DataSource))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsVisPieNoESQLByValuePanel(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsVisPieNoESQLByValuePanelTypePie),
		},
		{
			name: "pie/esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				raw := `{"type":"pie","title":"p","data_source":{"type":"esql","query":"FROM logs-* | LIMIT 10"},` +
					`"legend":{"size":"auto","visibility":"visible"},` +
					`"metrics":[{"operation":"value","column":"bytes","format":{"type":"number"}}],` +
					`"group_by":[{"operation":"value","column":"h","collapse_by":"avg"}]}`
				var api kbapi.KibanaHTTPAPIsVisPieESQLByValuePanel
				require.NoError(t, json.Unmarshal([]byte(raw), &api))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsVisPieESQLByValuePanel(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsVisPieNoESQLByValuePanelTypePie),
		},
		{
			name: "gauge/no_esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				api := kbapi.KibanaHTTPAPIsVisGaugeNoESQLByValuePanel{Type: kbapi.KibanaHTTPAPIsVisGaugeNoESQLByValuePanelTypeGauge}
				require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"m"}`), &api.DataSource))
				require.NoError(t, json.Unmarshal([]byte(`{"expression":"*","language":"kql"}`), &api.Query))
				require.NoError(t, json.Unmarshal([]byte(`{"operation":"count"}`), &api.Metric))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsVisGaugeNoESQLByValuePanel(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsVisGaugeNoESQLByValuePanelTypeGauge),
		},
		{
			name: "gauge/esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				api := kbapi.KibanaHTTPAPIsVisGaugeESQLByValuePanel{Type: kbapi.KibanaHTTPAPIsVisGaugeESQLByValuePanelTypeGauge}
				require.NoError(t, json.Unmarshal([]byte(`{"type":"esql","query":"FROM *"}`), &api.DataSource))
				require.NoError(t, json.Unmarshal([]byte(`{"type":"number"}`), &api.Metric.Format))
				api.Metric.Column = "c"
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsVisGaugeESQLByValuePanel(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsVisGaugeNoESQLByValuePanelTypeGauge),
		},
		{
			name: "waffle/no_esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				raw := `{"type":"waffle","data_source":{"type":"dataView","id":"m"},` +
					`"query":{"language":"kql","query":""},` +
					`"legend":{"size":"medium","visible":"auto"},"styling":{"values":{}},` +
					`"metrics":[{"operation":"count"}]}`
				var api kbapi.KibanaHTTPAPIsVisWaffleNoESQLByValuePanel
				require.NoError(t, json.Unmarshal([]byte(raw), &api))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsVisWaffleNoESQLByValuePanel(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsVisWaffleNoESQLByValuePanelTypeWaffle),
		},
		{
			name: "waffle/esql",
			build: func(t *testing.T) VisByValueConfig0 {
				t.Helper()
				raw := `{"type":"waffle","title":"w","data_source":{"type":"esql","query":"FROM logs-* | LIMIT 5"},"legend":{"size":"s"},"metrics":[{"column":"cnt","format":{"type":"number"}}]}`
				var api kbapi.KibanaHTTPAPIsVisWaffleESQLByValuePanel
				require.NoError(t, json.Unmarshal([]byte(raw), &api))
				var v VisByValueConfig0
				require.NoError(t, v.FromKibanaHTTPAPIsVisWaffleESQLByValuePanel(api))
				return v
			},
			want: string(kbapi.KibanaHTTPAPIsVisWaffleNoESQLByValuePanelTypeWaffle),
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
