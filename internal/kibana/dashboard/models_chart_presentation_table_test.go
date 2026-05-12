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
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func lensPresentationTestDashboard() *dashboardModel {
	return &dashboardModel{
		TimeRange: &timeRangeModel{
			From: types.StringValue("now-7d"),
			To:   types.StringValue("now"),
		},
	}
}

func sampleDatatableNoESQLModel(t *testing.T) *datatableNoESQLConfigModel {
	t.Helper()

	header := kbapi.DatatableDensity_Height_Header{}
	require.NoError(t, header.FromDatatableDensityHeightHeader0(kbapi.DatatableDensityHeightHeader0{
		Type: kbapi.DatatableDensityHeightHeader0TypeAuto,
	}))
	value := kbapi.DatatableDensity_Height_Value{}
	require.NoError(t, value.FromDatatableDensityHeightValue0(kbapi.DatatableDensityHeightValue0{
		Type: kbapi.DatatableDensityHeightValue0TypeAuto,
	}))
	density := kbapi.DatatableDensity{
		Mode: new(kbapi.DatatableDensityModeDefault),
		Height: &struct {
			Header *kbapi.DatatableDensity_Height_Header `json:"header,omitempty"`
			Value  *kbapi.DatatableDensity_Height_Value  `json:"value,omitempty"`
		}{
			Header: &header,
			Value:  &value,
		},
	}

	api := kbapi.DatatableNoESQL{
		Type:                kbapi.DatatableNoESQLTypeDataTable,
		Title:               new("Lens presentation datatable"),
		IgnoreGlobalFilters: new(false),
		Styling:             kbapi.DatatableStyling{Density: density},
		Query:               kbapi.FilterSimple{},
		Metrics:             []kbapi.DatatableNoESQL_Metrics_Item{},
	}
	require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"metrics-*"}`), &api.DataSource))
	require.NoError(t, json.Unmarshal([]byte(`{"language":"kql","expression":"*"}`), &api.Query))

	metric := kbapi.DatatableNoESQL_Metrics_Item{}
	require.NoError(t, json.Unmarshal([]byte(`{"operation":"count"}`), &metric))
	api.Metrics = append(api.Metrics, metric)

	m := &datatableNoESQLConfigModel{}
	diags := m.fromAPI(context.Background(), nil, nil, api)
	require.False(t, diags.HasError())
	return m
}

func sampleMetricNoESQLModel(t *testing.T) *metricChartConfigModel {
	t.Helper()

	apiChart := kbapi.MetricNoESQL{
		Type:                kbapi.MetricNoESQLTypeMetric,
		Title:               new("Lens presentation metric"),
		Description:         new("desc"),
		IgnoreGlobalFilters: new(false),
		Sampling:            new(float32(1.0)),
		Query: kbapi.FilterSimple{
			Language:   new(kbapi.FilterSimpleLanguage("kql")),
			Expression: "",
		},
		Metrics: []kbapi.MetricNoESQL_Metrics_Item{},
	}
	require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"metrics-*"}`), &apiChart.DataSource))

	metric := kbapi.MetricNoESQL_Metrics_Item{}
	require.NoError(t, json.Unmarshal([]byte(`{"operation":"count"}`), &metric))
	apiChart.Metrics = append(apiChart.Metrics, metric)

	m := &metricChartConfigModel{}
	diags := m.fromAPIVariant0(context.Background(), nil, nil, apiChart)
	require.False(t, diags.HasError())
	return m
}

func sampleGaugeNoESQLModel(t *testing.T) *gaugeConfigModel {
	t.Helper()

	api := kbapi.GaugeNoESQL{
		Type: kbapi.GaugeNoESQLTypeGauge,
	}
	require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"metrics-*"}`), &api.DataSource))
	require.NoError(t, json.Unmarshal([]byte(`{"expression":"*","language":"kql"}`), &api.Query))
	require.NoError(t, json.Unmarshal([]byte(`{"operation":"count"}`), &api.Metric))

	m := &gaugeConfigModel{}
	diags := m.fromAPI(context.Background(), nil, nil, api)
	require.False(t, diags.HasError())
	return m
}

func samplePieNoESQLModel(t *testing.T) *pieChartConfigModel {
	t.Helper()

	mode := kbapi.ValueDisplayModePercentage
	title := "Lens presentation pie"
	apiChart := kbapi.PieNoESQL{
		Title:      &title,
		Styling:    kbapi.PieStyling{Values: kbapi.ValueDisplay{Mode: &mode}},
		DataSource: kbapi.PieNoESQL_DataSource{},
		Query:      kbapi.FilterSimple{Expression: "*", Language: new(kbapi.FilterSimpleLanguageKql)},
		Metrics:    []kbapi.PieNoESQL_Metrics_Item{},
		Legend:     kbapi.PieLegend{Size: kbapi.LegendSizeAuto},
		GroupBy:    new([]kbapi.PieNoESQL_GroupBy_Item),
	}
	require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"metrics-*"}`), &apiChart.DataSource))

	metric := kbapi.PieNoESQL_Metrics_Item{}
	require.NoError(t, json.Unmarshal([]byte(`{"operation":"count"}`), &metric))
	apiChart.Metrics = append(apiChart.Metrics, metric)

	m := &pieChartConfigModel{}
	diags := m.fromAPINoESQL(context.Background(), nil, nil, apiChart)
	require.False(t, diags.HasError())
	return m
}

func urlDrilldownItem() lensDrilldownItemTFModel {
	return lensDrilldownItemTFModel{
		URLDrilldown: &lensURLDrilldownTFModel{
			URL:     types.StringValue("https://example.test/{{event.url}}"),
			Label:   types.StringValue("Open"),
			Trigger: types.StringValue("on_click_row"),
		},
	}
}

func runDatatableNoESQLLensChartPresentationComprehensive(t *testing.T) {
	ctx := context.Background()
	dash := lensPresentationTestDashboard()
	base := sampleDatatableNoESQLModel(t)

	t.Run("time_range null preserved when API echoes dashboard", func(t *testing.T) {
		m := *base
		m.TimeRange = nil

		api, diags := m.toAPI(dash)
		require.False(t, diags.HasError())
		api.TimeRange = timeRangeModelToAPI(dash.TimeRange)

		out := &datatableNoESQLConfigModel{}
		diags = out.fromAPI(ctx, dash, &m, api)
		require.False(t, diags.HasError())
		assert.Nil(t, out.TimeRange)
	})

	t.Run("hide_title round trip", func(t *testing.T) {
		m := *base
		m.HideTitle = types.BoolValue(true)

		api, diags := m.toAPI(dash)
		require.False(t, diags.HasError())
		require.NotNil(t, api.HideTitle)
		assert.True(t, *api.HideTitle)

		out := &datatableNoESQLConfigModel{}
		diags = out.fromAPI(ctx, dash, &m, api)
		require.False(t, diags.HasError())
		assert.Equal(t, types.BoolValue(true), out.HideTitle)
	})

	t.Run("drilldown url round trip", func(t *testing.T) {
		m := *base
		m.Drilldowns = []lensDrilldownItemTFModel{urlDrilldownItem()}

		api, diags := m.toAPI(dash)
		require.False(t, diags.HasError())
		require.NotNil(t, api.Drilldowns)
		assert.GreaterOrEqual(t, len(*api.Drilldowns), 1)

		out := &datatableNoESQLConfigModel{}
		diags = out.fromAPI(ctx, dash, &m, api)
		require.False(t, diags.HasError())
		require.Len(t, out.Drilldowns, 1)
		require.NotNil(t, out.Drilldowns[0].URLDrilldown)
		assert.Equal(t, "https://example.test/{{event.url}}", out.Drilldowns[0].URLDrilldown.URL.ValueString())
	})
}

func runMetricNoESQLLensChartPresentationComprehensive(t *testing.T) {
	ctx := context.Background()
	dash := lensPresentationTestDashboard()
	base := sampleMetricNoESQLModel(t)

	t.Run("time_range null preserved when API echoes dashboard", func(t *testing.T) {
		m := *base
		m.TimeRange = nil

		attrs, diags := m.toAPI(dash)
		require.False(t, diags.HasError())
		api, err := attrs.AsMetricNoESQL()
		require.NoError(t, err)
		api.TimeRange = timeRangeModelToAPI(dash.TimeRange)

		out := &metricChartConfigModel{}
		diags = out.fromAPIVariant0(ctx, dash, &m, api)
		require.False(t, diags.HasError())
		assert.Nil(t, out.TimeRange)
	})

	t.Run("hide_border round trip", func(t *testing.T) {
		m := *base
		m.HideBorder = types.BoolValue(true)

		attrs, diags := m.toAPI(dash)
		require.False(t, diags.HasError())
		api, err := attrs.AsMetricNoESQL()
		require.NoError(t, err)
		require.NotNil(t, api.HideBorder)
		assert.True(t, *api.HideBorder)

		out := &metricChartConfigModel{}
		diags = out.fromAPIVariant0(ctx, dash, &m, api)
		require.False(t, diags.HasError())
		assert.Equal(t, types.BoolValue(true), out.HideBorder)
	})
}

func runGaugeNoESQLLensChartPresentationComprehensive(t *testing.T) {
	ctx := context.Background()
	dash := lensPresentationTestDashboard()
	base := sampleGaugeNoESQLModel(t)

	t.Run("time_range null preserved when API echoes dashboard", func(t *testing.T) {
		m := *base
		m.TimeRange = nil

		api, diags := m.toAPI(dash)
		require.False(t, diags.HasError())
		api.TimeRange = timeRangeModelToAPI(dash.TimeRange)

		out := &gaugeConfigModel{}
		diags = out.fromAPI(ctx, dash, &m, api)
		require.False(t, diags.HasError())
		assert.Nil(t, out.TimeRange)
	})

	t.Run("references_json round trip", func(t *testing.T) {
		raw := `[{"id":"dash1","name":"N","type":"dashboard"}]`
		m := *base
		m.ReferencesJSON = jsontypes.NewNormalizedValue(raw)

		api, diags := m.toAPI(dash)
		require.False(t, diags.HasError())
		require.NotNil(t, api.References)

		out := &gaugeConfigModel{}
		diags = out.fromAPI(ctx, dash, &m, api)
		require.False(t, diags.HasError())
		assert.JSONEq(t, raw, out.ReferencesJSON.ValueString())
	})
}

func runPieNoESQLLensChartPresentationComprehensive(t *testing.T) {
	ctx := context.Background()
	dash := lensPresentationTestDashboard()
	base := samplePieNoESQLModel(t)

	t.Run("discover drilldown round trip", func(t *testing.T) {
		m := *base
		m.Drilldowns = []lensDrilldownItemTFModel{
			{
				DiscoverDrilldown: &lensDiscoverDrilldownTFModel{
					Label: types.StringValue("Discover drilldown"),
				},
			},
		}

		attrs, diags := m.toAPI(dash)
		require.False(t, diags.HasError())
		api, err := attrs.AsPieNoESQL()
		require.NoError(t, err)
		require.NotNil(t, api.Drilldowns)

		out := &pieChartConfigModel{}
		diags = out.fromAPINoESQL(ctx, dash, &m, api)
		require.False(t, diags.HasError())
		require.Len(t, out.Drilldowns, 1)
		require.NotNil(t, out.Drilldowns[0].DiscoverDrilldown)
		assert.Equal(t, lensDrilldownTriggerOnApplyFilter, out.Drilldowns[0].DiscoverDrilldown.Trigger.ValueString())
	})
}
