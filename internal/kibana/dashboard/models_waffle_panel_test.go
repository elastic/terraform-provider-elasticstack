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
	esql, err := waffleChartJSONUsesESQLDataset([]byte(`{"dataset":{"type":"esql","query":"FROM x"}}`))
	require.NoError(t, err)
	assert.True(t, esql)

	esqlTable, err := waffleChartJSONUsesESQLDataset([]byte(`{"dataset":{"type":"table","table":{}}}`))
	require.NoError(t, err)
	assert.True(t, esqlTable)

	no, err := waffleChartJSONUsesESQLDataset([]byte(`{"dataset":{"type":"dataView","id":"x"},"query":{"query":""}}`))
	require.NoError(t, err)
	assert.False(t, no)
}

func Test_wafflePanelConfigConverter_populateFromAttributes_NoESQL_emptyQueryNoLanguage(t *testing.T) {
	ctx := context.Background()
	// NoESQL with an empty lens query and no language: the old heuristic treated this as ES|QL.
	apiJSON := `{
		"type": "waffle",
		"dataset": {"type":"dataView","id":"metrics-*"},
		"query": {"query":""},
		"legend": {"size":"medium","visible":"auto"},
		"metrics": [{"operation":"count"}]
	}`
	var waffle kbapi.WaffleNoESQL
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &waffle))

	var waffleChart kbapi.WaffleChart
	require.NoError(t, waffleChart.FromWaffleNoESQL(waffle))

	var attrs kbapi.LensApiState
	require.NoError(t, attrs.FromWaffleChart(waffleChart))

	converter := newWafflePanelConfigConverter()
	pm := &panelModel{}
	diags := converter.populateFromAttributes(ctx, pm, attrs)
	require.False(t, diags.HasError(), "%s", diags)
	require.NotNil(t, pm.WaffleConfig)
	assert.False(t, pm.WaffleConfig.usesESQL())
	require.NotNil(t, pm.WaffleConfig.Query)
	assert.True(t, pm.WaffleConfig.Query.Query.IsNull() || pm.WaffleConfig.Query.Query.ValueString() == "")
}

func Test_wafflePanelConfigConverter_populateFromAttributes_buildAttributes_roundTrip_NoESQL(t *testing.T) {
	ctx := context.Background()

	apiJSON := `{
		"type": "waffle",
		"title": "Waffle NoESQL Round-Trip",
		"description": "test",
		"dataset": {"type":"dataView","id":"metrics-*"},
		"query": {"language":"kuery","query":""},
		"legend": {"size":"medium","visible":"auto"},
		"metrics": [{"operation":"count"}],
		"group_by": [{"operation":"terms","field":"host.name","collapse_by":"avg"}]
	}`
	var waffle kbapi.WaffleNoESQL
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &waffle))

	var waffleChart kbapi.WaffleChart
	require.NoError(t, waffleChart.FromWaffleNoESQL(waffle))

	var attrs kbapi.LensApiState
	require.NoError(t, attrs.FromWaffleChart(waffleChart))

	converter := newWafflePanelConfigConverter()
	pm := &panelModel{}
	diags := converter.populateFromAttributes(ctx, pm, attrs)
	require.False(t, diags.HasError(), "%s", diags)
	require.NotNil(t, pm.WaffleConfig)

	attrs2, diags := converter.buildAttributes(*pm)
	require.False(t, diags.HasError(), "%s", diags)

	chart2, err := attrs2.AsWaffleChart()
	require.NoError(t, err)
	noESQL2, err := chart2.AsWaffleNoESQL()
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

	waffle := kbapi.WaffleESQL{
		Type:        kbapi.WaffleESQLTypeWaffle,
		Title:       new("Waffle ESQL Round-Trip"),
		Description: new("esql test"),
		Legend:      kbapi.WaffleLegend{Size: kbapi.LegendSizeSmall},
		Metrics: []struct {
			Color     kbapi.StaticColor                `json:"color"`
			Column    string                           `json:"column"`
			Format    kbapi.FormatType                 `json:"format"`
			Label     *string                          `json:"label,omitempty"`
			Operation kbapi.WaffleESQLMetricsOperation `json:"operation"`
		}{
			{
				Column:    "cnt",
				Operation: kbapi.WaffleESQLMetricsOperationValue,
				Format:    format,
				Color: kbapi.StaticColor{
					Type:  kbapi.Static,
					Color: "#006BB4",
				},
			},
		},
		GroupBy: &[]struct {
			CollapseBy kbapi.CollapseBy                 `json:"collapse_by"`
			Color      kbapi.ColorMapping               `json:"color"`
			Column     string                           `json:"column"`
			Format     kbapi.FormatType                 `json:"format"`
			Label      *string                          `json:"label,omitempty"`
			Operation  kbapi.WaffleESQLGroupByOperation `json:"operation"`
		}{
			{
				Column:     "host",
				Format:     format,
				Operation:  kbapi.WaffleESQLGroupByOperationValue,
				CollapseBy: kbapi.CollapseByAvg,
				Color:      colorMap,
			},
		},
	}
	require.NoError(t, json.Unmarshal([]byte(`{"type":"esql","query":"FROM logs-* | STATS c = COUNT() BY host | LIMIT 10"}`), &waffle.Dataset))

	var waffleChart kbapi.WaffleChart
	require.NoError(t, waffleChart.FromWaffleESQL(waffle))

	var attrs kbapi.LensApiState
	require.NoError(t, attrs.FromWaffleChart(waffleChart))

	converter := newWafflePanelConfigConverter()
	pm := &panelModel{}
	diags := converter.populateFromAttributes(ctx, pm, attrs)
	require.False(t, diags.HasError(), "%s", diags)
	require.NotNil(t, pm.WaffleConfig)
	assert.True(t, pm.WaffleConfig.usesESQL())

	attrs2, diags := converter.buildAttributes(*pm)
	require.False(t, diags.HasError(), "%s", diags)

	chart2, err := attrs2.AsWaffleChart()
	require.NoError(t, err)
	esql2, err := chart2.AsWaffleESQL()
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
		DatasetJSON: jsontypes.NewNormalizedNull(),
		Legend: &waffleLegendModel{
			Size: types.StringValue("medium"),
		},
		Query: &filterSimpleModel{
			Language: types.StringValue("kuery"),
			Query:    types.StringValue(""),
		},
	}
	_, diags := m.toAPI()
	require.True(t, diags.HasError())

	m2 := &waffleConfigModel{
		DatasetJSON: jsontypes.NewNormalizedValue(`{"type":"dataView","id":"x"}`),
		Legend: &waffleLegendModel{
			Size: types.StringValue("medium"),
		},
		Query: &filterSimpleModel{
			Language: types.StringValue("kuery"),
			Query:    types.StringValue(""),
		},
		Metrics: nil,
	}
	_, diags2 := m2.toAPI()
	require.True(t, diags2.HasError())
}

func Test_waffleConfigModel_toAPI_ESQL_errors(t *testing.T) {
	m := &waffleConfigModel{
		DatasetJSON: jsontypes.NewNormalizedValue(`{"type":"esql","query":"FROM x | LIMIT 1"}`),
		Legend: &waffleLegendModel{
			Size: types.StringValue("medium"),
		},
		Query: nil,
	}
	_, diags := m.toAPI()
	require.True(t, diags.HasError())
}
