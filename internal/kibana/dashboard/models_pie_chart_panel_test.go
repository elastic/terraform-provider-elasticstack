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
	nested := true
	truncateLines := float32(3)

	apiChart := kbapi.PieNoESQL{
		Title:       &title,
		Description: &desc,
		DonutHole:   &donutHole,
		Labels: &struct {
			Position *kbapi.PieNoESQLLabelsPosition `json:"position,omitempty"`
			Visible  *bool                          `json:"visible,omitempty"`
		}{Position: &labelPos},
		Legend: kbapi.PieLegend{
			Size:               kbapi.LegendSizeAuto,
			Nested:             &nested,
			TruncateAfterLines: &truncateLines,
			Visibility:         &visibility,
		},
		DataSource: kbapi.PieNoESQL_DataSource{},
		Query:      kbapi.FilterSimple{Expression: "response:200", Language: new(kbapi.FilterSimpleLanguageKql)},
		Metrics:    []kbapi.PieNoESQL_Metrics_Item{},
		GroupBy:    new([]kbapi.PieNoESQL_GroupBy_Item{}),
	}

	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	require.NoError(t, attrs.FromPieNoESQL(apiChart))

	converter := newPieChartPanelConfigConverter()
	pm := &panelModel{}
	diags := converter.populateFromAttributes(ctx, pm, attrs)
	require.False(t, diags.HasError())
	require.NotNil(t, pm.PieChartConfig)
	require.NotNil(t, pm.PieChartConfig.Legend)
	assert.Equal(t, "auto", pm.PieChartConfig.Legend.Size.ValueString())
	assert.True(t, pm.PieChartConfig.Legend.Nested.ValueBool())
	assert.InEpsilon(t, float64(3), pm.PieChartConfig.Legend.TruncateAfterLine.ValueFloat64(), 0.001)
	assert.Equal(t, string(visibility), pm.PieChartConfig.Legend.Visible.ValueString())

	attrs2, diags := converter.buildAttributes(*pm)
	require.False(t, diags.HasError())

	noESQL2, err := attrs2.AsPieNoESQL()
	require.NoError(t, err)
	assert.Equal(t, "Pie Round-Trip", *noESQL2.Title)
	assert.Equal(t, "response:200", noESQL2.Query.Expression)
	assert.Equal(t, kbapi.LegendSizeAuto, noESQL2.Legend.Size)
	require.NotNil(t, noESQL2.Legend.Nested)
	assert.True(t, *noESQL2.Legend.Nested)
	require.NotNil(t, noESQL2.Legend.TruncateAfterLines)
	assert.InEpsilon(t, float64(3), float64(*noESQL2.Legend.TruncateAfterLines), 0.001)
	require.NotNil(t, noESQL2.Legend.Visibility)
	assert.Equal(t, visibility, *noESQL2.Legend.Visibility)
}

func Test_pieChartConfigModel_fromAPI_toAPI_PieNoESQL(t *testing.T) {
	// Setup test data
	title := "My Pie Chart"
	desc := "A delicious pie chart"
	donutHole := kbapi.PieNoESQLDonutHoleS
	labelPos := kbapi.PieNoESQLLabelsPositionInside

	// Create a dummy dataset
	dataset := kbapi.PieNoESQL_DataSource{}

	visibility := kbapi.PieLegendVisibilityVisible
	nested := true
	truncate := float32(4)
	legend := kbapi.PieLegend{
		Size:               kbapi.LegendSizeAuto,
		Nested:             &nested,
		TruncateAfterLines: &truncate,
		Visibility:         &visibility,
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
		Legend:     legend,
		DataSource: dataset,
		Query:      query,
		Metrics:    []kbapi.PieNoESQL_Metrics_Item{}, // Empty for simplicity
		GroupBy:    new([]kbapi.PieNoESQL_GroupBy_Item{}),
	}

	model := &pieChartConfigModel{}
	diags := model.fromAPINoESQL(apiChart)
	require.False(t, diags.HasError(), "fromAPINoESQL should not have errors")

	// Verify fields
	assert.Equal(t, title, model.Title.ValueString())
	assert.Equal(t, desc, model.Description.ValueString())
	assert.Equal(t, string(donutHole), model.DonutHole.ValueString())
	assert.Equal(t, string(labelPos), model.LabelPosition.ValueString())
	assert.Equal(t, "response:200", model.Query.Expression.ValueString())
	require.NotNil(t, model.Legend)
	assert.Equal(t, "auto", model.Legend.Size.ValueString())
	assert.True(t, model.Legend.Nested.ValueBool())
	assert.InEpsilon(t, float64(4), model.Legend.TruncateAfterLine.ValueFloat64(), 0.001)
	assert.Equal(t, string(visibility), model.Legend.Visible.ValueString())

	// Test toAPI
	resultAttrs, diags := model.toAPI()
	require.False(t, diags.HasError(), "toAPI should not have errors")

	resultNoESQL, err := resultAttrs.AsPieNoESQL()
	require.NoError(t, err)

	assert.Equal(t, title, *resultNoESQL.Title)
	assert.Equal(t, desc, *resultNoESQL.Description)
	assert.Equal(t, kbapi.LegendSizeAuto, resultNoESQL.Legend.Size)
	require.NotNil(t, resultNoESQL.Legend.Nested)
	assert.True(t, *resultNoESQL.Legend.Nested)
	require.NotNil(t, resultNoESQL.Legend.TruncateAfterLines)
	assert.InEpsilon(t, float64(4), float64(*resultNoESQL.Legend.TruncateAfterLines), 0.001)
	require.NotNil(t, resultNoESQL.Legend.Visibility)
	assert.Equal(t, visibility, *resultNoESQL.Legend.Visibility)
}

func Test_pieChartConfigModel_fromAPI_toAPI_PieESQL(t *testing.T) {
	apiJSON := `{
		"type": "pie",
		"title": "ESQL Pie Chart",
		"description": "ESQL pie description",
		"data_source": {"type":"esql","query":"FROM logs-* | LIMIT 10"},
		"sampling": 0.5,
		"ignore_global_filters": true,
		"legend": {"size":"auto","visibility":"visible"},
		"metrics": [{"operation":"value","column":"bytes","color":{"type":"static","color":"#54B399"},"format":{"type":"number"}}],
		"group_by": [{"operation":"value","column":"host.name","collapse_by":"avg","color":{"mode":"categorical","palette":"default","mapping":[],"unassignedColor":{"type":"color_code","value":"#D3DAE6"}}}]
	}`
	var apiESQL kbapi.PieESQL
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &apiESQL))

	model := &pieChartConfigModel{}
	diags := model.fromAPIESQL(apiESQL)
	require.False(t, diags.HasError())

	assert.Equal(t, "ESQL Pie Chart", model.Title.ValueString())
	assert.Equal(t, "ESQL pie description", model.Description.ValueString())
	assert.Len(t, model.Metrics, 1)
	assert.Len(t, model.GroupBy, 1)
	require.NotNil(t, model.Legend)
	assert.Equal(t, "auto", model.Legend.Size.ValueString())
	assert.Equal(t, string(kbapi.PieLegendVisibilityVisible), model.Legend.Visible.ValueString())

	resultAttrs, diags := model.toAPI()
	require.False(t, diags.HasError())

	esql2, err := resultAttrs.AsPieESQL()
	require.NoError(t, err)
	assert.Equal(t, "ESQL Pie Chart", *esql2.Title)
	assert.Equal(t, kbapi.PieESQLType("pie"), esql2.Type)
	assert.Len(t, esql2.Metrics, 1)
	assert.Equal(t, "bytes", esql2.Metrics[0].Column)
	assert.Equal(t, kbapi.LegendSizeAuto, esql2.Legend.Size)
	require.NotNil(t, esql2.Legend.Visibility)
	assert.Equal(t, kbapi.PieLegendVisibilityVisible, *esql2.Legend.Visibility)
}

func Test_pieChartConfigModel_toAPI_withMetrics(t *testing.T) {
	model := &pieChartConfigModel{
		Title:          types.StringValue("Pie with metrics"),
		Description:    types.StringValue("Test"),
		DataSourceJSON: jsontypes.NewNormalizedValue(`{"type":"dataView","id":"logs-*"}`),
		Legend: &partitionLegendModel{
			Size:    types.StringValue("auto"),
			Visible: types.StringValue("visible"),
		},
		Query: &filterSimpleModel{Expression: types.StringValue("*"), Language: types.StringValue("kql")},
		Metrics: []pieMetricModel{
			{Config: customtypes.NewJSONWithDefaultsValue[map[string]any](`{"operation":"count"}`, populatePieChartMetricDefaults)},
			{Config: customtypes.NewJSONWithDefaultsValue[map[string]any](`{"operation":"sum","field":"bytes"}`, populatePieChartMetricDefaults)},
		},
		GroupBy: []pieGroupByModel{
			{Config: customtypes.NewJSONWithDefaultsValue(`{"operation":"terms","field":"host.name"}`, populateLensGroupByDefaults)},
		},
	}

	attrs, diags := model.toAPI()
	require.False(t, diags.HasError())

	noESQL, err := attrs.AsPieNoESQL()
	require.NoError(t, err)
	assert.Len(t, noESQL.Metrics, 2)
	assert.NotNil(t, noESQL.GroupBy)
	assert.Len(t, *noESQL.GroupBy, 1)
}

func Test_pieChartConfigModel_toAPI_withGroupBy(t *testing.T) {
	model := &pieChartConfigModel{
		Title:          types.StringValue("Pie with groupBy"),
		DataSourceJSON: jsontypes.NewNormalizedValue(`{"type":"dataView","id":"logs-*"}`),
		Legend: &partitionLegendModel{
			Size:    types.StringValue("auto"),
			Visible: types.StringValue("visible"),
		},
		Query: &filterSimpleModel{Expression: types.StringValue("*"), Language: types.StringValue("kql")},
		Metrics: []pieMetricModel{
			{Config: customtypes.NewJSONWithDefaultsValue[map[string]any](`{"operation":"count"}`, populatePieChartMetricDefaults)},
		},
		GroupBy: []pieGroupByModel{
			{Config: customtypes.NewJSONWithDefaultsValue(`{"operation":"terms","field":"host.name","size":10}`, populateLensGroupByDefaults)},
			{Config: customtypes.NewJSONWithDefaultsValue(`{"operation":"terms","field":"service.name"}`, populateLensGroupByDefaults)},
		},
	}

	attrs, diags := model.toAPI()
	require.False(t, diags.HasError())

	noESQL, err := attrs.AsPieNoESQL()
	require.NoError(t, err)
	require.NotNil(t, noESQL.GroupBy)
	assert.Len(t, *noESQL.GroupBy, 2)
}

func Test_pieChartConfigModel_toAPI_legendOmitted(t *testing.T) {
	model := &pieChartConfigModel{
		Title:          types.StringValue("Pie default legend"),
		DataSourceJSON: jsontypes.NewNormalizedValue(`{"type":"dataView","id":"logs-*"}`),
		Query:          &filterSimpleModel{Expression: types.StringValue("*"), Language: types.StringValue("kql")},
		Metrics: []pieMetricModel{
			{Config: customtypes.NewJSONWithDefaultsValue[map[string]any](`{"operation":"count"}`, populatePieChartMetricDefaults)},
		},
		Legend: nil,
	}

	attrs, diags := model.toAPI()
	require.False(t, diags.HasError())

	noESQL, err := attrs.AsPieNoESQL()
	require.NoError(t, err)
	assert.Equal(t, kbapi.LegendSizeAuto, noESQL.Legend.Size)
}

func Test_pieChartConfigModel_toAPI_legendOmitted_PieESQL(t *testing.T) {
	model := &pieChartConfigModel{
		Title: types.StringValue("ESQL pie default legend"),
		DataSourceJSON: jsontypes.NewNormalizedValue(
			`{"type":"esql","query":"FROM logs-* | LIMIT 10"}`,
		),
		Query: nil,
		Metrics: []pieMetricModel{
			{
				Config: customtypes.NewJSONWithDefaultsValue[map[string]any](
					`{"operation":"value","column":"bytes","color":{"type":"static","color":"#54B399"},"format":{"type":"number"}}`,
					populatePieChartMetricDefaults,
				),
			},
		},
		GroupBy: []pieGroupByModel{
			{
				Config: customtypes.NewJSONWithDefaultsValue(
					`{"operation":"value","column":"host.name","collapse_by":"avg","color":{"mode":"categorical","palette":"default","mapping":[],"unassignedColor":{"type":"color_code","value":"#D3DAE6"}}}`,
					populateLensGroupByDefaults,
				),
			},
		},
		Legend: nil,
	}

	attrs, diags := model.toAPI()
	require.False(t, diags.HasError())

	esql, err := attrs.AsPieESQL()
	require.NoError(t, err)
	assert.Equal(t, kbapi.LegendSizeAuto, esql.Legend.Size)
}
