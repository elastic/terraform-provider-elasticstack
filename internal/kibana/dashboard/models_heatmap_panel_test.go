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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_newHeatmapPanelConfigConverter(t *testing.T) {
	converter := newHeatmapPanelConfigConverter()
	assert.NotNil(t, converter)
	assert.Equal(t, "heatmap", converter.visualizationType)
}

func Test_heatmapConfigModel_fromAPI_toAPI_noESQL(t *testing.T) {
	heatmap := kbapi.HeatmapNoESQL{
		Type:                kbapi.HeatmapNoESQLTypeHeatmap,
		Title:               new("Test Heatmap"),
		Description:         new("Heatmap description"),
		IgnoreGlobalFilters: new(true),
		Sampling:            new(float32(0.5)),
		Query: kbapi.FilterSimple{
			Expression: "status:200",
			Language: func() *kbapi.FilterSimpleLanguage {
				lang := kbapi.FilterSimpleLanguage("kuery")
				return &lang
			}(),
		},
		Axes: kbapi.HeatmapAxes{
			X: kbapi.HeatmapXAxis{
				Labels: &struct {
					Orientation kbapi.VisApiOrientation `json:"orientation"`
					Visible     *bool                   `json:"visible,omitempty"`
				}{
					Orientation: kbapi.VisApiOrientation("horizontal"),
					Visible:     new(true),
				},
				Title: &struct {
					Text    *string `json:"text,omitempty"`
					Visible *bool   `json:"visible,omitempty"`
				}{
					Text:    new("X Axis"),
					Visible: new(true),
				},
			},
			Y: kbapi.HeatmapYAxis{
				Labels: &struct {
					Visible *bool `json:"visible,omitempty"`
				}{
					Visible: new(false),
				},
				Title: &struct {
					Text    *string `json:"text,omitempty"`
					Visible *bool   `json:"visible,omitempty"`
				}{
					Text:    new("Y Axis"),
					Visible: new(true),
				},
			},
		},
		Cells: kbapi.HeatmapCells{
			Labels: &struct {
				Visible *bool `json:"visible,omitempty"`
			}{
				Visible: new(true),
			},
		},
		Legend: kbapi.HeatmapLegend{
			Size: kbapi.LegendSizeM,
			Visibility: func() *kbapi.HeatmapLegendVisibility {
				visibility := kbapi.HeatmapLegendVisibilityVisible
				return &visibility
			}(),
			TruncateAfterLines: new(float32(4)),
		},
	}

	require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"metrics-*"}`), &heatmap.Dataset))
	require.NoError(t, json.Unmarshal([]byte(`{"operation":"count"}`), &heatmap.Metric))
	require.NoError(t, json.Unmarshal([]byte(`{"operation":"filters","filters":[{"label":"All","filter":{"query":"*","language":"kuery"}}]}`), &heatmap.X))
	var yAxis kbapi.HeatmapNoESQL_Y
	require.NoError(t, json.Unmarshal([]byte(`{"operation":"filters","filters":[{"label":"All","filter":{"query":"*","language":"kuery"}}]}`), &yAxis))
	heatmap.Y = &yAxis

	var fItem kbapi.LensPanelFilters_Item
	require.NoError(t, json.Unmarshal([]byte(`{"type":"condition","condition":{"field":"status","operator":"is","value":"200"}}`), &fItem))
	filters := []kbapi.LensPanelFilters_Item{fItem}
	heatmap.Filters = filters

	model := &heatmapConfigModel{}
	diags := model.fromAPINoESQL(context.Background(), heatmap)
	require.False(t, diags.HasError())

	assert.Equal(t, types.StringValue("Test Heatmap"), model.Title)
	assert.Equal(t, types.StringValue("Heatmap description"), model.Description)
	assert.Equal(t, types.BoolValue(true), model.IgnoreGlobalFilters)
	assert.Equal(t, types.Float64Value(0.5), model.Sampling)
	require.NotNil(t, model.Query)
	assert.Equal(t, types.StringValue("status:200"), model.Query.Query)
	assert.Equal(t, types.StringValue("kuery"), model.Query.Language)
	assert.False(t, model.DatasetJSON.IsNull())
	assert.False(t, model.MetricJSON.IsNull())
	assert.False(t, model.XAxisJSON.IsNull())
	assert.False(t, model.YAxisJSON.IsNull())
	require.NotNil(t, model.Axes)
	require.NotNil(t, model.Cells)
	require.NotNil(t, model.Legend)

	chart, diags := model.toAPI()
	require.False(t, diags.HasError())

	heatmapRoundTrip, err := chart.AsHeatmapNoESQL()
	require.NoError(t, err)
	assert.Equal(t, kbapi.HeatmapNoESQLTypeHeatmap, heatmapRoundTrip.Type)
	require.NotNil(t, heatmapRoundTrip.Title)
	assert.Equal(t, "Test Heatmap", *heatmapRoundTrip.Title)
	assert.Equal(t, kbapi.LegendSizeM, heatmapRoundTrip.Legend.Size)
	assert.Equal(t, "status:200", heatmapRoundTrip.Query.Expression)
}

func Test_heatmapConfigModel_fromAPI_toAPI_esql(t *testing.T) {
	const esqlHeatmapJSON = `{
		"type": "heatmap",
		"title": "ESQL Heatmap",
		"description": "ESQL heatmap description",
		"ignore_global_filters": false,
		"sampling": 1,
		"axis": {
			"x": { "labels": { "orientation": "angled", "visible": true } },
			"y": { "labels": { "visible": true } }
		},
		"cells": { "labels": { "visible": false } },
		"legend": { "size": "small", "visible": false },
		"dataset": {"type":"esql","query":"FROM logs-* | LIMIT 10"},
		"metric": {
			"color": {"type":"dynamic","range":"absolute","steps":[{"type":"from","from":0,"color":"#000000"}]},
			"column": "bytes",
			"format": {"type":"number"},
			"operation": "value"
		},
		"x": {"column":"host","format":{"type":"number"},"operation":"value"},
		"y": {"column":"service","format":{"type":"number"},"operation":"value"}
	}`
	var heatmap kbapi.HeatmapESQL
	require.NoError(t, json.Unmarshal([]byte(esqlHeatmapJSON), &heatmap))

	model := &heatmapConfigModel{}
	diags := model.fromAPIESQL(context.Background(), heatmap)
	require.False(t, diags.HasError())
	assert.Nil(t, model.Query)
	assert.Equal(t, types.StringValue("ESQL Heatmap"), model.Title)
	assert.Equal(t, types.StringValue("ESQL heatmap description"), model.Description)

	chart, diags := model.toAPI()
	require.False(t, diags.HasError())

	heatmapRoundTrip, err := chart.AsHeatmapESQL()
	require.NoError(t, err)
	assert.Equal(t, kbapi.HeatmapESQLTypeHeatmap, heatmapRoundTrip.Type)
	assert.Equal(t, "bytes", heatmapRoundTrip.Metric.Column)
}

func Test_heatmapPanelConfigConverter_populateFromAttributes_buildAttributes_roundTrip_NoESQL(t *testing.T) {
	ctx := context.Background()

	heatmap := kbapi.HeatmapNoESQL{
		Type:                kbapi.HeatmapNoESQLTypeHeatmap,
		Title:               new("Heatmap NoESQL Round-Trip"),
		Description:         new("Converter test"),
		IgnoreGlobalFilters: new(true),
		Sampling:            new(float32(0.5)),
		Query: kbapi.FilterSimple{
			Expression: "status:200",
			Language:   new(kbapi.FilterSimpleLanguage("kuery")),
		},
	}
	require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"metrics-*"}`), &heatmap.Dataset))
	require.NoError(t, json.Unmarshal([]byte(`{"operation":"count"}`), &heatmap.Metric))
	require.NoError(t, json.Unmarshal([]byte(`{"operation":"filters","filters":[{"label":"All","filter":{"query":"*","language":"kuery"}}]}`), &heatmap.X))

	var heatmapChart kbapi.HeatmapChart
	require.NoError(t, heatmapChart.FromHeatmapNoESQL(heatmap))

	var attrs kbapi.LensApiState
	require.NoError(t, attrs.FromHeatmapChart(heatmapChart))

	converter := newHeatmapPanelConfigConverter()
	pm := &panelModel{}
	diags := converter.populateFromAttributes(ctx, pm, attrs)
	require.False(t, diags.HasError())
	require.NotNil(t, pm.HeatmapConfig)

	attrs2, diags := converter.buildAttributes(*pm)
	require.False(t, diags.HasError())

	chart2, err := attrs2.AsHeatmapChart()
	require.NoError(t, err)
	noESQL2, err := chart2.AsHeatmapNoESQL()
	require.NoError(t, err)
	assert.Equal(t, "Heatmap NoESQL Round-Trip", *noESQL2.Title)
	assert.Equal(t, kbapi.HeatmapNoESQLTypeHeatmap, noESQL2.Type)
}

func Test_heatmapPanelConfigConverter_populateFromAttributes_buildAttributes_roundTrip_ESQL(t *testing.T) {
	ctx := context.Background()

	const esqlRoundTripJSON = `{
		"type": "heatmap",
		"title": "Heatmap ESQL Round-Trip",
		"description": "Converter test",
		"ignore_global_filters": false,
		"sampling": 1,
		"axis": { "x": {}, "y": {} },
		"cells": {},
		"legend": { "size": "medium" },
		"dataset": {"type":"esql","query":"FROM logs-* | LIMIT 10"},
		"metric": {
			"color": {"type":"dynamic","range":"absolute","steps":[{"type":"from","from":0,"color":"#000000"}]},
			"column": "bytes",
			"format": {"type":"number"},
			"operation": "value"
		},
		"x": {"column":"host","format":{"type":"number"},"operation":"value"},
		"y": {"column":"service","format":{"type":"number"},"operation":"value"}
	}`
	var heatmap kbapi.HeatmapESQL
	require.NoError(t, json.Unmarshal([]byte(esqlRoundTripJSON), &heatmap))

	var heatmapChart kbapi.HeatmapChart
	require.NoError(t, heatmapChart.FromHeatmapESQL(heatmap))

	var attrs kbapi.LensApiState
	require.NoError(t, attrs.FromHeatmapChart(heatmapChart))

	converter := newHeatmapPanelConfigConverter()
	pm := &panelModel{}
	diags := converter.populateFromAttributes(ctx, pm, attrs)
	require.False(t, diags.HasError())
	require.NotNil(t, pm.HeatmapConfig)

	attrs2, diags := converter.buildAttributes(*pm)
	require.False(t, diags.HasError())

	chart2, err := attrs2.AsHeatmapChart()
	require.NoError(t, err)
	esql2, err := chart2.AsHeatmapESQL()
	require.NoError(t, err)
	assert.Equal(t, "Heatmap ESQL Round-Trip", *esql2.Title)
	assert.Equal(t, kbapi.HeatmapESQLTypeHeatmap, esql2.Type)
	assert.Equal(t, "host", esql2.X.Column)
}
