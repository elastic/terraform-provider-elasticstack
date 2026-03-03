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

func Test_pieChartConfigModel_fromAPI_toAPI_PieNoESQL(t *testing.T) {
	// Setup test data
	title := "My Pie Chart"
	desc := "A delicious pie chart"
	donutHole := kbapi.PieNoESQLDonutHoleSmall
	labelPos := kbapi.PieNoESQLLabelPositionInside

	// Create a dummy dataset
	dataset := kbapi.PieNoESQL_Dataset{}

	visible := kbapi.PieLegendVisibleShow
	legend := kbapi.PieLegend{
		Visible: &visible,
	}

	query := kbapi.FilterSimple{
		Query:    "response:200",
		Language: new(kbapi.FilterSimpleLanguageKuery),
	}

	apiChart := kbapi.PieNoESQL{
		Title:         &title,
		Description:   &desc,
		DonutHole:     &donutHole,
		LabelPosition: &labelPos,
		Legend:        legend,
		Dataset:       dataset,
		Query:         query,
		Metrics:       []kbapi.PieNoESQL_Metrics_Item{}, // Empty for simplicity
		GroupBy:       new([]kbapi.PieNoESQL_GroupBy_Item{}),
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
	assert.Equal(t, "response:200", model.Query.Query.ValueString())

	// Test toAPI
	resultSchema, diags := model.toAPI()
	require.False(t, diags.HasError(), "toAPI should not have errors")

	// Verify we can convert back to PieNoESQL
	resultNoESQL, err := resultSchema.AsPieNoESQL()
	require.NoError(t, err)

	assert.Equal(t, title, *resultNoESQL.Title)
	assert.Equal(t, desc, *resultNoESQL.Description)
}

func Test_pieChartPanelConfigConverter_roundTrip(t *testing.T) {
	marshalConfig := func(t *testing.T, cfg kbapi.DashboardPanelItem_Config) map[string]any {
		t.Helper()
		b, err := cfg.MarshalJSON()
		require.NoError(t, err)
		var m map[string]any
		require.NoError(t, json.Unmarshal(b, &m))
		return m
	}

	t.Run("NoESQL", func(t *testing.T) {
		converter := newPieChartPanelConfigConverter()

		legendVisible := kbapi.PieLegendVisibleShow
		legend := kbapi.PieLegend{
			Size:    kbapi.LegendSizeAuto,
			Visible: &legendVisible,
		}
		legendJSON, err := json.Marshal(legend)
		require.NoError(t, err)

		configModel := &pieChartConfigModel{
			Title:               types.StringValue("Round Trip Pie (NoESQL)"),
			Description:         types.StringValue("NoESQL round-trip test"),
			Dataset:             jsontypes.NewNormalizedValue(`{"type":"dataView","id":"metrics-*"}`),
			Legend:              jsontypes.NewNormalizedValue(string(legendJSON)),
			IgnoreGlobalFilters: types.BoolValue(true),
			Sampling:            types.Float64Value(0.5),
			DonutHole:           types.StringValue(string(kbapi.PieNoESQLDonutHoleSmall)),
			LabelPosition:       types.StringValue(string(kbapi.PieNoESQLLabelPositionInside)),
			Query: &filterSimpleModel{
				Language: types.StringValue("kuery"),
				Query:    types.StringValue("response:200"),
			},
			Metrics: []pieMetricModel{
				{
					Config: customtypes.NewJSONWithDefaultsValue[map[string]any](
						`{"operation":"count","empty_as_null":false,"format":{"type":"number","decimals":2}}`,
						populatePieChartMetricDefaults,
					),
				},
			},
			GroupBy: []pieGroupByModel{
				{
					Config: customtypes.NewJSONWithDefaultsValue[map[string]any](
						`{"operation":"terms","field":"host.name","size":5}`,
						populatePieChartGroupByDefaults,
					),
				},
			},
		}

		panel := panelModel{
			Type:           types.StringValue("lens"),
			PieChartConfig: configModel,
		}

		var apiConfig1 kbapi.DashboardPanelItem_Config
		diags := converter.mapPanelToAPI(panel, &apiConfig1)
		require.False(t, diags.HasError())

		newPanel := panelModel{Type: types.StringValue("lens")}
		diags = converter.populateFromAPIPanel(context.Background(), &newPanel, apiConfig1)
		require.False(t, diags.HasError())
		require.NotNil(t, newPanel.PieChartConfig)
		require.NotNil(t, newPanel.PieChartConfig.Query)
		assert.Equal(t, "Round Trip Pie (NoESQL)", newPanel.PieChartConfig.Title.ValueString())
		assert.Equal(t, "response:200", newPanel.PieChartConfig.Query.Query.ValueString())

		var apiConfig2 kbapi.DashboardPanelItem_Config
		diags = converter.mapPanelToAPI(newPanel, &apiConfig2)
		require.False(t, diags.HasError())

		assert.Equal(t, marshalConfig(t, apiConfig1), marshalConfig(t, apiConfig2))
	})

	t.Run("ESQL", func(t *testing.T) {
		converter := newPieChartPanelConfigConverter()

		legendVisible := kbapi.PieLegendVisibleHide
		legend := kbapi.PieLegend{
			Size:    kbapi.LegendSizeSmall,
			Visible: &legendVisible,
		}
		legendJSON, err := json.Marshal(legend)
		require.NoError(t, err)

		configModel := &pieChartConfigModel{
			Title:               types.StringValue("Round Trip Pie (ESQL)"),
			Description:         types.StringValue("ESQL round-trip test"),
			Dataset:             jsontypes.NewNormalizedValue(`{"type":"esql","query":"FROM metrics-* | KEEP host.name, system.cpu.user.pct | LIMIT 10"}`),
			Legend:              jsontypes.NewNormalizedValue(string(legendJSON)),
			IgnoreGlobalFilters: types.BoolValue(false),
			Sampling:            types.Float64Value(1.0),
			DonutHole:           types.StringValue(string(kbapi.PieESQLDonutHoleLarge)),
			LabelPosition:       types.StringValue(string(kbapi.PieESQLLabelPositionOutside)),
			Query:               nil, // Disambiguates ESQL code path in toAPI()
			Metrics: []pieMetricModel{
				{
					Config: customtypes.NewJSONWithDefaultsValue[map[string]any](
						`{"color":{"type":"static","color":"#FF0000"},"column":"system.cpu.user.pct","format":{"type":"number","decimals":2},"label":"cpu","operation":"value"}`,
						populatePieChartMetricDefaults,
					),
				},
			},
			GroupBy: []pieGroupByModel{
				{
					Config: customtypes.NewJSONWithDefaultsValue[map[string]any](
						`{"collapse_by":"avg","color":{},"column":"host.name","operation":"value"}`,
						populatePieChartGroupByDefaults,
					),
				},
			},
		}

		panel := panelModel{
			Type:           types.StringValue("lens"),
			PieChartConfig: configModel,
		}

		var apiConfig1 kbapi.DashboardPanelItem_Config
		diags := converter.mapPanelToAPI(panel, &apiConfig1)
		require.False(t, diags.HasError())

		newPanel := panelModel{Type: types.StringValue("lens")}
		diags = converter.populateFromAPIPanel(context.Background(), &newPanel, apiConfig1)
		require.False(t, diags.HasError())
		require.NotNil(t, newPanel.PieChartConfig)
		assert.Nil(t, newPanel.PieChartConfig.Query)
		assert.Equal(t, "Round Trip Pie (ESQL)", newPanel.PieChartConfig.Title.ValueString())

		var apiConfig2 kbapi.DashboardPanelItem_Config
		diags = converter.mapPanelToAPI(newPanel, &apiConfig2)
		require.False(t, diags.HasError())

		assert.Equal(t, marshalConfig(t, apiConfig1), marshalConfig(t, apiConfig2))
	})
}
