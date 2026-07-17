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

package visconfig_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/iface"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/visconfig"
	"github.com/stretchr/testify/require"
)

// minimalVisConfig0ForChartKind returns a minimal valid VisConfig0 union member for each Lens chart kind.
// Fixtures are adapted from internal/kibana/dashboard/lenscommon/detect_test.go (chartKindsPerArm table).
func minimalVisConfig0ForChartKind(t *testing.T, vizType string) lenscommon.VisByValueConfig0 {
	t.Helper()
	switch vizType {
	case string(kbapi.KibanaHTTPAPIsXyChartNoESQLByValuePanelTypeXy):
		raw := `{
					"type":"xy","title":"t","axis":{"x":{},"y":{}},
					"layers":[{"type":"line","data_source":{"type":"dataView","id":"l"},"ignore_global_filters":false,"sampling":1,"y":[{"operation":"count"}]}],
					"legend":{"visibility":"visible","inside":false,"size":"auto"},
					"filters":[],"styling":{"line":{"curve":"linear"}},
					"query":{"expression":"*","language":"kql"}
				}`
		var x kbapi.KibanaHTTPAPIsXyChartNoESQLByValuePanel
		require.NoError(t, json.Unmarshal([]byte(raw), &x))
		var v lenscommon.VisByValueConfig0
		require.NoError(t, v.FromKibanaHTTPAPIsXyChartNoESQLByValuePanel(x))
		return v

	case string(kbapi.KibanaHTTPAPIsTreemapNoESQLByValuePanelTypeTreemap):
		apiJSON := `{"type":"treemap","title":"t",` +
			`"data_source":{"type":"dataView","id":"m"},` +
			`"query":{"language":"kql","expression":""},` +
			`"legend":{"size":"small"},` +
			`"metrics":[{"operation":"count"}],` +
			`"group_by":[{"operation":"terms","field":"host.name","collapse_by":"avg"}]}`
		var api kbapi.KibanaHTTPAPIsTreemapNoESQLByValuePanel
		require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))
		var v lenscommon.VisByValueConfig0
		require.NoError(t, v.FromKibanaHTTPAPIsTreemapNoESQLByValuePanel(api))
		return v

	case string(kbapi.KibanaHTTPAPIsMosaicNoESQLByValuePanelTypeMosaic):
		const grp = `{"mode":"categorical","palette":"default","mapping":[],"unassigned":{"type":"color_code","value":"#D3DAE6"}}`
		groupBy := `[{"operation":"terms","collapse_by":"avg","fields":["host.name"],"color":` + grp + `}]`
		apiJSON := `{"type":"mosaic","title":"m","data_source":{"type":"dataView","id":"x"},` +
			`"query":{"language":"kql","expression":""},"legend":{"size":"small"},` +
			`"metric":{"operation":"count"},"group_by":` + groupBy + `,"group_breakdown_by":` + groupBy + `}`
		var api kbapi.KibanaHTTPAPIsMosaicNoESQLByValuePanel
		require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))
		var v lenscommon.VisByValueConfig0
		require.NoError(t, v.FromKibanaHTTPAPIsMosaicNoESQLByValuePanel(api))
		return v

	case string(kbapi.KibanaHTTPAPIsDatatableNoESQLByValuePanelTypeDataTable):
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
		var v lenscommon.VisByValueConfig0
		require.NoError(t, v.FromKibanaHTTPAPIsDatatableNoESQLByValuePanel(minDatatableNoESQL))
		return v

	case string(kbapi.KibanaHTTPAPIsTagcloudNoESQLByValuePanelTypeTagCloud):
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
		var v lenscommon.VisByValueConfig0
		require.NoError(t, v.FromKibanaHTTPAPIsTagcloudNoESQLByValuePanel(api))
		return v

	case string(kbapi.KibanaHTTPAPIsHeatmapNoESQLByValuePanelTypeHeatmap):
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
		var v lenscommon.VisByValueConfig0
		require.NoError(t, v.FromKibanaHTTPAPIsHeatmapNoESQLByValuePanel(heatmap))
		return v

	case string(kbapi.KibanaHTTPAPIsRegionMapNoESQLByValuePanelTypeRegionMap):
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
		var v lenscommon.VisByValueConfig0
		require.NoError(t, v.FromKibanaHTTPAPIsRegionMapNoESQLByValuePanel(api))
		return v

	case string(kbapi.LegacyMetric):
		raw := `{"type":"legacy_metric","title":"l","data_source":{"type":"data_view_spec","index_pattern":"m"},` +
			`"query":{"language":"kql","query":"*"},"metric":{"operation":"count","format":{"type":"number"}}}`
		var api kbapi.KibanaHTTPAPIsLegacyMetricNoESQLByValuePanel
		require.NoError(t, json.Unmarshal([]byte(raw), &api))
		var v lenscommon.VisByValueConfig0
		require.NoError(t, v.FromKibanaHTTPAPIsLegacyMetricNoESQLByValuePanel(api))
		return v

	case string(kbapi.KibanaHTTPAPIsMetricNoESQLByValuePanelTypeMetric):
		const inner = `{
		"type": "metric",
		"title": "M",
		"query": { "expression": "*", "language": "kql" },
		"metrics": []
	}`
		var v lenscommon.VisByValueConfig0
		require.NoError(t, json.Unmarshal([]byte(inner), &v))
		return v

	case string(kbapi.KibanaHTTPAPIsPieNoESQLByValuePanelTypePie):
		api := kbapi.KibanaHTTPAPIsPieNoESQLByValuePanel{
			Type:    kbapi.KibanaHTTPAPIsPieNoESQLByValuePanelTypePie,
			Query:   &kbapi.KibanaHTTPAPIsFilterSimple{Expression: "*", Language: new(kbapi.KibanaHTTPAPIsFilterSimpleLanguageKql)},
			Styling: &kbapi.KibanaHTTPAPIsPieStyling{},
			Metrics: []kbapi.KibanaHTTPAPIsPieNoESQLByValuePanel_Metrics_Item{},
			GroupBy: new([]kbapi.KibanaHTTPAPIsPieNoESQLByValuePanel_GroupBy_Item{}),
		}
		require.NoError(t, json.Unmarshal([]byte(`{}`), &api.DataSource))
		var v lenscommon.VisByValueConfig0
		require.NoError(t, v.FromKibanaHTTPAPIsPieNoESQLByValuePanel(api))
		return v

	case string(kbapi.KibanaHTTPAPIsGaugeNoESQLByValuePanelTypeGauge):
		api := kbapi.KibanaHTTPAPIsGaugeNoESQLByValuePanel{Type: kbapi.KibanaHTTPAPIsGaugeNoESQLByValuePanelTypeGauge}
		require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"m"}`), &api.DataSource))
		require.NoError(t, json.Unmarshal([]byte(`{"expression":"*","language":"kql"}`), &api.Query))
		require.NoError(t, json.Unmarshal([]byte(`{"operation":"count"}`), &api.Metric))
		var v lenscommon.VisByValueConfig0
		require.NoError(t, v.FromKibanaHTTPAPIsGaugeNoESQLByValuePanel(api))
		return v

	case string(kbapi.KibanaHTTPAPIsWaffleNoESQLByValuePanelTypeWaffle):
		raw := `{"type":"waffle","data_source":{"type":"dataView","id":"m"},` +
			`"query":{"language":"kql","query":""},` +
			`"legend":{"size":"medium","visible":"auto"},"styling":{"values":{}},` +
			`"metrics":[{"operation":"count"}]}`
		var api kbapi.KibanaHTTPAPIsWaffleNoESQLByValuePanel
		require.NoError(t, json.Unmarshal([]byte(raw), &api))
		var v lenscommon.VisByValueConfig0
		require.NoError(t, v.FromKibanaHTTPAPIsWaffleNoESQLByValuePanel(api))
		return v

	default:
		t.Fatalf("no minimal fixture for viz type %q — add one adapted from lenscommon/detect_test.go", vizType)
		return lenscommon.VisByValueConfig0{}
	}
}

func assertExactlyOneLensChartBlock(t *testing.T, want lenscommon.VizConverter, blocks *models.LensByValueChartBlocks) {
	t.Helper()
	require.True(t, want.HandlesBlocks(blocks), "converter %q should recognize its chart block", want.VizType())
	for _, c := range lenscommon.All() {
		if c.VizType() == want.VizType() {
			continue
		}
		require.False(t, c.HandlesBlocks(blocks), "unexpected secondary chart match from converter %q while testing %q",
			c.VizType(), want.VizType())
	}
}

func TestHandler_FromAPI_byValue_allRegisteredLensCharts(t *testing.T) {
	ctx := iface.WithEnclosingDashboard(context.Background(), &models.DashboardModel{})
	for _, c := range lenscommon.All() {
		t.Run(c.VizType(), func(t *testing.T) {
			cfg0 := minimalVisConfig0ForChartKind(t, c.VizType())
			item := mustVisPanelItem(t, cfg0)
			var pm models.PanelModel
			diags := visconfig.Handler{}.FromAPI(ctx, &pm, nil, item)
			require.False(t, diags.HasError(), "%s", diags)
			require.NotNil(t, pm.VisConfig)
			require.NotNil(t, pm.VisConfig.ByValue)
			assertExactlyOneLensChartBlock(t, c, &pm.VisConfig.ByValue.LensByValueChartBlocks)
		})
	}
}
