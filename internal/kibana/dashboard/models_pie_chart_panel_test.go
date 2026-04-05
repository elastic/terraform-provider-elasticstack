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
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_pieChartPanelConfigConverter_populateFromAttributes_buildAttributes_roundTrip(t *testing.T) {
	ctx := context.Background()

	title := "Pie Round-Trip"
	desc := "Converter test"
	donutHole := kbapi.PieNoESQLDonutHoleS
	labelPos := kbapi.PieNoESQLLabelsPositionInside
	visibility := kbapi.PieLegendVisibilityVisible

	apiChart := kbapi.PieNoESQL{
		Title:       &title,
		Description: &desc,
		DonutHole:   &donutHole,
		Labels: &struct {
			Position *kbapi.PieNoESQLLabelsPosition `json:"position,omitempty"`
			Visible  *bool                          `json:"visible,omitempty"`
		}{Position: &labelPos},
		Legend:  kbapi.PieLegend{Visibility: &visibility},
		Dataset: kbapi.PieNoESQL_Dataset{},
		Query:   kbapi.FilterSimple{Expression: "response:200", Language: new(kbapi.FilterSimpleLanguageKql)},
		Metrics: []kbapi.PieNoESQL_Metrics_Item{},
		GroupBy: new([]kbapi.PieNoESQL_GroupBy_Item{}),
	}

	var pieChart kbapi.PieChart
	require.NoError(t, pieChart.FromPieNoESQL(apiChart))

	var attrs kbapi.LensApiState
	require.NoError(t, attrs.FromPieChart(pieChart))

	converter := newPieChartPanelConfigConverter()
	pm := &panelModel{}
	diags := converter.populateFromAttributes(ctx, pm, attrs)
	require.False(t, diags.HasError())
	require.NotNil(t, pm.PieChartConfig)

	attrs2, diags := converter.buildAttributes(*pm)
	require.False(t, diags.HasError())

	chart2, err := attrs2.AsPieChart()
	require.NoError(t, err)
	noESQL2, err := chart2.AsPieNoESQL()
	require.NoError(t, err)
	assert.Equal(t, "Pie Round-Trip", *noESQL2.Title)
	assert.Equal(t, "response:200", noESQL2.Query.Expression)
}

func Test_pieChartConfigModel_fromAPI_toAPI_PieNoESQL(t *testing.T) {
	// Setup test data
	title := "My Pie Chart"
	desc := "A delicious pie chart"
	donutHole := kbapi.PieNoESQLDonutHoleS
	labelPos := kbapi.PieNoESQLLabelsPositionInside

	// Create a dummy dataset
	dataset := kbapi.PieNoESQL_Dataset{}

	visibility := kbapi.PieLegendVisibilityVisible
	legend := kbapi.PieLegend{
		Visibility: &visibility,
	}

	query := kbapi.FilterSimple{
		Expression: "response:200",
		Language:   new(kbapi.FilterSimpleLanguageKql),
	}

	apiChart := kbapi.PieNoESQL{
		Title:       &title,
		Description: &desc,
		DonutHole:   &donutHole,
		Labels: &struct {
			Position *kbapi.PieNoESQLLabelsPosition `json:"position,omitempty"`
			Visible  *bool                          `json:"visible,omitempty"`
		}{Position: &labelPos},
		Legend:  legend,
		Dataset: dataset,
		Query:   query,
		Metrics: []kbapi.PieNoESQL_Metrics_Item{}, // Empty for simplicity
		GroupBy: new([]kbapi.PieNoESQL_GroupBy_Item{}),
	}

	// Wrap in PieChart
	var apiSchema kbapi.PieChart
	err := apiSchema.FromPieNoESQL(apiChart)
	require.NoError(t, err)

	// Test fromAPI
	ctx := context.Background()
	model := &pieChartConfigModel{}
	diags := model.fromAPI(ctx, apiSchema)
	require.False(t, diags.HasError(), "fromAPI should not have errors")

	// Verify fields
	assert.Equal(t, title, model.Title.ValueString())
	assert.Equal(t, desc, model.Description.ValueString())
	assert.Equal(t, string(donutHole), model.DonutHole.ValueString())
	assert.Equal(t, string(labelPos), model.LabelPosition.ValueString())
	assert.Equal(t, "response:200", model.Query.Expression.ValueString())

	// Test toAPI
	resultSchema, diags := model.toAPI()
	require.False(t, diags.HasError(), "toAPI should not have errors")

	// Verify we can convert back to PieNoESQL
	resultNoESQL, err := resultSchema.AsPieNoESQL()
	require.NoError(t, err)

	assert.Equal(t, title, *resultNoESQL.Title)
	assert.Equal(t, desc, *resultNoESQL.Description)
}

func Test_pieChartConfigModel_fromAPI_toAPI_PieESQL(t *testing.T) {
	ctx := context.Background()

	apiJSON := `{
		"type": "pie",
		"title": "ESQL Pie Chart",
		"description": "ESQL pie description",
		"dataset": {"type":"esql","query":"FROM logs-* | LIMIT 10"},
		"sampling": 0.5,
		"ignore_global_filters": true,
		"legend": {"visible": "show"},
		"metrics": [{"operation":"value","column":"bytes","color":{"type":"static","color":"#54B399"},"format":{"type":"number"}}],
		"group_by": [{"operation":"value","column":"host.name","collapse_by":"avg","color":{"mode":"categorical","palette":"default","mapping":[],"unassignedColor":{"type":"color_code","value":"#D3DAE6"}}}]
	}`
	var apiESQL kbapi.PieESQL
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &apiESQL))

	var pieChart kbapi.PieChart
	require.NoError(t, pieChart.FromPieESQL(apiESQL))

	model := &pieChartConfigModel{}
	diags := model.fromAPI(ctx, pieChart)
	require.False(t, diags.HasError())

	assert.Equal(t, "ESQL Pie Chart", model.Title.ValueString())
	assert.Equal(t, "ESQL pie description", model.Description.ValueString())
	assert.Len(t, model.Metrics, 1)
	assert.Len(t, model.GroupBy, 1)

	resultChart, diags := model.toAPI()
	require.False(t, diags.HasError())

	esql2, err := resultChart.AsPieESQL()
	require.NoError(t, err)
	assert.Equal(t, "ESQL Pie Chart", *esql2.Title)
	assert.Equal(t, kbapi.PieESQLType("pie"), esql2.Type)
	assert.Len(t, esql2.Metrics, 1)
	assert.Equal(t, "bytes", esql2.Metrics[0].Column)
}

func Test_pieChartConfigModel_toAPI_withMetrics(t *testing.T) {
	model := &pieChartConfigModel{
		Title:       types.StringValue("Pie with metrics"),
		Description: types.StringValue("Test"),
		DatasetJSON: jsontypes.NewNormalizedValue(`{"type":"dataView","id":"logs-*"}`),
		LegendJSON:  jsontypes.NewNormalizedValue(`{"visible":"show"}`),
		Query:       &filterSimpleModel{Expression: types.StringValue("*"), Language: types.StringValue("kql")},
		Metrics: []pieMetricModel{
			{Config: customtypes.NewJSONWithDefaultsValue[map[string]any](`{"operation":"count"}`, populatePieChartMetricDefaults)},
			{Config: customtypes.NewJSONWithDefaultsValue[map[string]any](`{"operation":"sum","field":"bytes"}`, populatePieChartMetricDefaults)},
		},
		GroupBy: []pieGroupByModel{
			{Config: customtypes.NewJSONWithDefaultsValue(`{"operation":"terms","field":"host.name"}`, populateLensGroupByDefaults)},
		},
	}

	chart, diags := model.toAPI()
	require.False(t, diags.HasError())

	noESQL, err := chart.AsPieNoESQL()
	require.NoError(t, err)
	assert.Len(t, noESQL.Metrics, 2)
	assert.NotNil(t, noESQL.GroupBy)
	assert.Len(t, *noESQL.GroupBy, 1)
}

func Test_pieChartConfigModel_toAPI_withGroupBy(t *testing.T) {
	model := &pieChartConfigModel{
		Title:   types.StringValue("Pie with groupBy"),
		DatasetJSON: jsontypes.NewNormalizedValue(`{"type":"dataView","id":"logs-*"}`),
		LegendJSON:  jsontypes.NewNormalizedValue(`{"visible":"show"}`),
		Query:   &filterSimpleModel{Expression: types.StringValue("*"), Language: types.StringValue("kql")},
		Metrics: []pieMetricModel{
			{Config: customtypes.NewJSONWithDefaultsValue[map[string]any](`{"operation":"count"}`, populatePieChartMetricDefaults)},
		},
		GroupBy: []pieGroupByModel{
			{Config: customtypes.NewJSONWithDefaultsValue(`{"operation":"terms","field":"host.name","size":10}`, populateLensGroupByDefaults)},
			{Config: customtypes.NewJSONWithDefaultsValue(`{"operation":"terms","field":"service.name"}`, populateLensGroupByDefaults)},
		},
	}

	chart, diags := model.toAPI()
	require.False(t, diags.HasError())

	noESQL, err := chart.AsPieNoESQL()
	require.NoError(t, err)
	require.NotNil(t, noESQL.GroupBy)
	assert.Len(t, *noESQL.GroupBy, 2)
}
