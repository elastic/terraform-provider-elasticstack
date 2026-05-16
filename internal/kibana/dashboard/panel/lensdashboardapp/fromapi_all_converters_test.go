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

package lensdashboardapp_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	_ "github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard" // register lens converters
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/iface"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/lensdashboardapp"
	"github.com/stretchr/testify/require"
)

// minimalVisConfig0ForChartKind returns a minimal valid VisConfig0 union member for each Lens chart kind.
// Fixtures are adapted from internal/kibana/dashboard/lenscommon/detect_test.go (chartKindsPerArm table).
func minimalVisConfig0ForChartKind(t *testing.T, vizType string) kbapi.KbnDashboardPanelTypeVisConfig0 {
	t.Helper()
	switch vizType {
	case string(kbapi.XyChartNoESQLTypeXy):
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

	case string(kbapi.TreemapNoESQLTypeTreemap):
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

	case string(kbapi.MosaicNoESQLTypeMosaic):
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

	case string(kbapi.DatatableNoESQLTypeDataTable):
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
		var v kbapi.KbnDashboardPanelTypeVisConfig0
		require.NoError(t, v.FromDatatableNoESQL(minDatatableNoESQL))
		return v

	case string(kbapi.TagcloudNoESQLTypeTagCloud):
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

	case string(kbapi.HeatmapNoESQLTypeHeatmap):
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

	case string(kbapi.RegionMapNoESQLTypeRegionMap):
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

	case string(kbapi.LegacyMetric):
		raw := `{"type":"legacy_metric","title":"l","data_source":{"type":"data_view_spec","index_pattern":"m"},` +
			`"query":{"language":"kql","query":"*"},"metric":{"operation":"count","format":{"type":"number"}}}`
		var api kbapi.LegacyMetricNoESQL
		require.NoError(t, json.Unmarshal([]byte(raw), &api))
		var v kbapi.KbnDashboardPanelTypeVisConfig0
		require.NoError(t, v.FromLegacyMetricNoESQL(api))
		return v

	case string(kbapi.MetricNoESQLTypeMetric):
		const inner = `{
		"type": "metric",
		"title": "M",
		"query": { "expression": "*", "language": "kql" },
		"metrics": []
	}`
		var v kbapi.KbnDashboardPanelTypeVisConfig0
		require.NoError(t, json.Unmarshal([]byte(inner), &v))
		return v

	case string(kbapi.PieNoESQLTypePie):
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

	case string(kbapi.GaugeNoESQLTypeGauge):
		api := kbapi.GaugeNoESQL{Type: kbapi.GaugeNoESQLTypeGauge}
		require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"m"}`), &api.DataSource))
		require.NoError(t, json.Unmarshal([]byte(`{"expression":"*","language":"kql"}`), &api.Query))
		require.NoError(t, json.Unmarshal([]byte(`{"operation":"count"}`), &api.Metric))
		var v kbapi.KbnDashboardPanelTypeVisConfig0
		require.NoError(t, v.FromGaugeNoESQL(api))
		return v

	case string(kbapi.WaffleNoESQLTypeWaffle):
		raw := `{"type":"waffle","data_source":{"type":"dataView","id":"m"},"query":{"language":"kql","query":""},"legend":{"size":"medium","visible":"auto"},"metrics":[{"operation":"count"}]}`
		var api kbapi.WaffleNoESQL
		require.NoError(t, json.Unmarshal([]byte(raw), &api))
		var v kbapi.KbnDashboardPanelTypeVisConfig0
		require.NoError(t, v.FromWaffleNoESQL(api))
		return v

	default:
		t.Fatalf("no minimal fixture for viz type %q — add one adapted from lenscommon/detect_test.go", vizType)
		return kbapi.KbnDashboardPanelTypeVisConfig0{}
	}
}

func TestHandler_FromAPI_byValue_allRegisteredLensCharts(t *testing.T) {
	ctx := iface.WithEnclosingDashboard(context.Background(), &models.DashboardModel{})
	for _, c := range lenscommon.All() {
		t.Run(c.VizType(), func(t *testing.T) {
			cfg0 := minimalVisConfig0ForChartKind(t, c.VizType())
			item := mustLensDashboardAppPanelItem(t, cfg0)
			var pm models.PanelModel
			diags := lensdashboardapp.Handler{}.FromAPI(ctx, &pm, nil, item)
			require.False(t, diags.HasError(), "%s", diags)
			require.NotNil(t, pm.LensDashboardAppConfig)
			require.NotNil(t, pm.LensDashboardAppConfig.ByValue)
			require.False(t, pm.LensDashboardAppConfig.ByValue.ConfigJSON.IsNull())
			var root map[string]any
			require.NoError(t, json.Unmarshal([]byte(pm.LensDashboardAppConfig.ByValue.ConfigJSON.ValueString()), &root))
			require.Equal(t, c.VizType(), root["type"])
		})
	}
}
