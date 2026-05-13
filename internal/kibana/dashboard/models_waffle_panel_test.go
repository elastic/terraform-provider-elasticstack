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

func Test_waffleChartJSONUsesESQLDataset(t *testing.T) {
	t.Parallel()
	esql, err := waffleChartJSONUsesESQLDataset([]byte(`{"data_source":{"type":"esql","query":"FROM x"}}`))
	require.NoError(t, err)
	assert.True(t, esql)

	esqlTable, err := waffleChartJSONUsesESQLDataset([]byte(`{"data_source":{"type":"table","table":{}}}`))
	require.NoError(t, err)
	assert.True(t, esqlTable)

	no, err := waffleChartJSONUsesESQLDataset([]byte(`{"data_source":{"type":"dataView","id":"x"},"query":{"query":""}}`))
	require.NoError(t, err)
	assert.False(t, no)
}

func Test_wafflePanelConfigConverter_populateFromAttributes_NoESQL_emptyQueryNoLanguage(t *testing.T) {
	ctx := context.Background()
	// NoESQL with an empty lens query and no language: the old heuristic treated this as ES|QL.
	apiJSON := `{
		"type": "waffle",
		"data_source": {"type":"dataView","id":"metrics-*"},
		"query": {"query":""},
		"legend": {"size":"medium","visible":"auto"},
		"metrics": [{"operation":"count"}]
	}`
	var waffle kbapi.WaffleNoESQL
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &waffle))

	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	require.NoError(t, attrs.FromWaffleNoESQL(waffle))

	converter := newWafflePanelConfigConverter()
	visBv := visByValueModel{}
	diags := converter.populateFromAttributes(ctx, nil, nil, &visBv.lensByValueChartBlocks, attrs)
	require.False(t, diags.HasError(), "%s", diags)
	require.NotNil(t, visBv.WaffleConfig)
	assert.False(t, visBv.WaffleConfig.usesESQL())
	require.NotNil(t, visBv.WaffleConfig.Query)
	assert.True(t, visBv.WaffleConfig.Query.Expression.IsNull() || visBv.WaffleConfig.Query.Expression.ValueString() == "")
}

func Test_wafflePanelConfigConverter_populateFromAttributes_buildAttributes_roundTrip_NoESQL(t *testing.T) {
	ctx := context.Background()

	apiJSON := `{
		"type": "waffle",
		"title": "Waffle NoESQL Round-Trip",
		"description": "test",
		"data_source": {"type":"dataView","id":"metrics-*"},
		"query": {"language":"kql","query":""},
		"legend": {"size":"medium","visible":"auto"},
		"metrics": [{"operation":"count"}],
		"group_by": [{"operation":"terms","field":"host.name","collapse_by":"avg"}]
	}`
	var waffle kbapi.WaffleNoESQL
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &waffle))

	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	require.NoError(t, attrs.FromWaffleNoESQL(waffle))

	converter := newWafflePanelConfigConverter()
	visBv := visByValueModel{}
	diags := converter.populateFromAttributes(ctx, nil, nil, &visBv.lensByValueChartBlocks, attrs)
	require.False(t, diags.HasError(), "%s", diags)
	require.NotNil(t, visBv.WaffleConfig)

	attrs2, diags := converter.buildAttributes(&visBv.lensByValueChartBlocks, nil)
	require.False(t, diags.HasError(), "%s", diags)

	noESQL2, err := attrs2.AsWaffleNoESQL()
	require.NoError(t, err)
	assert.Equal(t, "Waffle NoESQL Round-Trip", *noESQL2.Title)
	assert.Equal(t, kbapi.WaffleNoESQLTypeWaffle, noESQL2.Type)
	require.Len(t, noESQL2.Metrics, 1)
}

func Test_wafflePanelConfigConverter_populateFromAttributes_buildAttributes_roundTrip_ESQL(t *testing.T) {
	ctx := context.Background()

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
			{
				Column: "cnt",
				Format: format,
				Color:  &staticColorUnion,
			},
		},
		GroupBy: &[]struct {
			CollapseBy kbapi.CollapseBy   `json:"collapse_by"`
			Color      kbapi.ColorMapping `json:"color"`
			Column     string             `json:"column"`
			Format     kbapi.FormatType   `json:"format"`
			Label      *string            `json:"label,omitempty"`
		}{
			{
				Column:     "host",
				Format:     format,
				CollapseBy: kbapi.CollapseByAvg,
				Color:      colorMap,
			},
		},
	}
	require.NoError(t, json.Unmarshal([]byte(`{"type":"esql","query":"FROM logs-* | STATS c = COUNT() BY host | LIMIT 10"}`), &waffle.DataSource))

	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	require.NoError(t, attrs.FromWaffleESQL(waffle))

	converter := newWafflePanelConfigConverter()
	visBv := visByValueModel{}
	diags := converter.populateFromAttributes(ctx, nil, nil, &visBv.lensByValueChartBlocks, attrs)
	require.False(t, diags.HasError(), "%s", diags)
	require.NotNil(t, visBv.WaffleConfig)
	assert.True(t, visBv.WaffleConfig.usesESQL())

	attrs2, diags := converter.buildAttributes(&visBv.lensByValueChartBlocks, nil)
	require.False(t, diags.HasError(), "%s", diags)

	esql2, err := attrs2.AsWaffleESQL()
	require.NoError(t, err)
	assert.Equal(t, "Waffle ESQL Round-Trip", *esql2.Title)
	assert.Equal(t, kbapi.WaffleESQLTypeWaffle, esql2.Type)
	require.Len(t, esql2.Metrics, 1)
	assert.Equal(t, "cnt", esql2.Metrics[0].Column)
	require.NotNil(t, esql2.GroupBy)
	require.Len(t, *esql2.GroupBy, 1)
}

func Test_waffleConfigModel_toAPI_NoESQL_errors(t *testing.T) {
	m := &waffleConfigModel{
		DataSourceJSON: jsontypes.NewNormalizedNull(),
		Legend: &waffleLegendModel{
			Size: types.StringValue("medium"),
		},
		Query: &filterSimpleModel{
			Language:   types.StringValue("kql"),
			Expression: types.StringValue(""),
		},
	}
	_, diags := m.toAPI(nil)
	require.True(t, diags.HasError())

	m2 := &waffleConfigModel{
		DataSourceJSON: jsontypes.NewNormalizedValue(`{"type":"dataView","id":"x"}`),
		Legend: &waffleLegendModel{
			Size: types.StringValue("medium"),
		},
		Query: &filterSimpleModel{
			Language:   types.StringValue("kql"),
			Expression: types.StringValue(""),
		},
		Metrics: nil,
	}
	_, diags2 := m2.toAPI(nil)
	require.True(t, diags2.HasError())
}

func Test_waffleConfigModel_toAPI_ESQL_errors(t *testing.T) {
	m := &waffleConfigModel{
		DataSourceJSON: jsontypes.NewNormalizedValue(`{"type":"esql","query":"FROM x | LIMIT 1"}`),
		Legend: &waffleLegendModel{
			Size: types.StringValue("medium"),
		},
		Query: nil,
	}
	_, diags := m.toAPI(nil)
	require.True(t, diags.HasError())
}

func Test_waffleConfigModel_config_json_metricRoundTrip(t *testing.T) {
	// Verifies that waffle metrics use config_json (not config) for round-trip.
	// The struct field Config has tfsdk tag "config_json".
	ctx := context.Background()

	apiJSON := `{
		"type": "waffle",
		"data_source": {"type":"dataView","id":"metrics-*"},
		"query": {"language":"kql","query":"status:200"},
		"legend": {"size":"medium"},
		"metrics": [{"operation":"count"},{"operation":"sum","field":"bytes"}],
		"group_by": [{"operation":"terms","field":"host.name","collapse_by":"avg"}]
	}`
	var waffle kbapi.WaffleNoESQL
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &waffle))

	model := &waffleConfigModel{}
	diags := model.fromAPINoESQL(ctx, nil, nil, waffle)
	require.False(t, diags.HasError(), "%s", diags)

	// Metrics should use config_json (the tfsdk tag on waffleDSLMetric.Config)
	require.Len(t, model.Metrics, 2)
	assert.False(t, model.Metrics[0].Config.IsNull())
	assert.False(t, model.Metrics[1].Config.IsNull())
	assert.Contains(t, model.Metrics[0].Config.ValueString(), "count")
	assert.Contains(t, model.Metrics[1].Config.ValueString(), "sum")

	// GroupBy should also use config_json (the tfsdk tag on waffleDSLGroupBy.Config)
	require.Len(t, model.GroupBy, 1)
	assert.False(t, model.GroupBy[0].Config.IsNull())
	assert.Contains(t, model.GroupBy[0].Config.ValueString(), "terms")

	// Round-trip back to API
	attrs, diags := model.toAPI(nil)
	require.False(t, diags.HasError())
	noESQL, err := attrs.AsWaffleNoESQL()
	require.NoError(t, err)
	require.Len(t, noESQL.Metrics, 2)
	require.NotNil(t, noESQL.GroupBy)
	require.Len(t, *noESQL.GroupBy, 1)
}

func Test_waffleConfig_lensChartPresentation_hideTitleRoundTrip(t *testing.T) {
	ctx := context.Background()
	dash := lensPresentationTestDashboard()
	pm := buildLensWafflePanelForTest(t)

	require.NotNil(t, pm.VisConfig)
	require.NotNil(t, pm.VisConfig.ByValue)
	require.NotNil(t, pm.VisConfig.ByValue.WaffleConfig)
	m := *pm.VisConfig.ByValue.WaffleConfig
	m.HideTitle = types.BoolValue(true)

	attrs, diags := m.toAPI(dash)
	require.False(t, diags.HasError())
	api, err := attrs.AsWaffleNoESQL()
	require.NoError(t, err)

	got := &waffleConfigModel{}
	require.False(t, got.fromAPINoESQL(ctx, dash, &m, api).HasError())
	assert.Equal(t, types.BoolValue(true), got.HideTitle)
}
