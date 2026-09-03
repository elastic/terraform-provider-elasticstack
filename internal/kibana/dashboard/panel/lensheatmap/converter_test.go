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

package lensheatmap

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConverter_VizType(t *testing.T) {
	var c converter
	require.Equal(t, string(kbapi.KibanaHTTPAPIsHeatmapNoESQLByValuePanelTypeHeatmap), c.VizType())
}

func TestConverter_HandlesBlocks(t *testing.T) {
	var c converter
	require.False(t, c.HandlesBlocks(nil))
	require.False(t, c.HandlesBlocks(&models.LensByValueChartBlocks{}))
	require.True(t, c.HandlesBlocks(&models.LensByValueChartBlocks{
		HeatmapConfig: &models.HeatmapConfigModel{},
	}))
}

func TestConverter_roundTrip_NoESQL(t *testing.T) {
	ctx := context.Background()
	var c converter
	query := kbapi.KibanaHTTPAPIsFilterSimple{
		Expression: "status:200",
		Language: func() *kbapi.KibanaHTTPAPIsFilterSimpleLanguage {
			lang := kbapi.KibanaHTTPAPIsFilterSimpleLanguage("kql")
			return &lang
		}(),
	}
	xOrientation := kbapi.KibanaHTTPAPIsVisApiOrientation("horizontal")
	xAxis := kbapi.KibanaHTTPAPIsHeatmapXAxis{
		Labels: &struct {
			Orientation *kbapi.KibanaHTTPAPIsVisApiOrientation `json:"orientation,omitempty"`
			Visible     *bool                                  `json:"visible,omitempty"`
		}{
			Orientation: &xOrientation,
			Visible:     new(true),
		},
		Title: &struct {
			Text    *string `json:"text,omitempty"`
			Visible *bool   `json:"visible,omitempty"`
		}{
			Text:    new("X Axis"),
			Visible: new(true),
		},
	}
	yAxisConfig := kbapi.KibanaHTTPAPIsHeatmapYAxis{
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
	}
	cells := kbapi.KibanaHTTPAPIsHeatmapCells{
		Labels: &struct {
			Visible *bool `json:"visible,omitempty"`
		}{
			Visible: new(true),
		},
	}
	legendSize := kbapi.KibanaHTTPAPIsLegendSizeM
	heatmap := kbapi.KibanaHTTPAPIsHeatmapNoESQLByValuePanel{
		Type:                kbapi.KibanaHTTPAPIsHeatmapNoESQLByValuePanelTypeHeatmap,
		Title:               new("Test Heatmap"),
		Description:         new("Heatmap description"),
		IgnoreGlobalFilters: new(true),
		Sampling:            new(float32(0.5)),
		Query:               &query,
		Axis: &kbapi.KibanaHTTPAPIsHeatmapAxes{
			X: &xAxis,
			Y: &yAxisConfig,
		},
		Styling: &kbapi.KibanaHTTPAPIsHeatmapStyling{
			Cells: &cells,
		},
		Legend: &kbapi.KibanaHTTPAPIsHeatmapLegend{
			Size: &legendSize,
			Visibility: func() *kbapi.KibanaHTTPAPIsHeatmapLegendVisibility {
				visibility := kbapi.KibanaHTTPAPIsHeatmapLegendVisibilityVisible
				return &visibility
			}(),
			TruncateAfterLines: new(float32(4)),
		},
	}

	require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"metrics-*"}`), &heatmap.DataSource))
	require.NoError(t, json.Unmarshal([]byte(`{"operation":"count"}`), &heatmap.Metric))
	require.NoError(t, json.Unmarshal([]byte(`{"operation":"filters","filters":[{"label":"All","filter":{"query":"*","language":"kql"}}]}`), &heatmap.X))
	var yAxis kbapi.KibanaHTTPAPIsHeatmapNoESQLByValuePanel_Y
	require.NoError(t, json.Unmarshal([]byte(`{"operation":"filters","filters":[{"label":"All","filter":{"query":"*","language":"kql"}}]}`), &yAxis))
	heatmap.Y = &yAxis

	var fItem kbapi.KibanaHTTPAPIsLensPanelFilters_Item
	require.NoError(t, json.Unmarshal([]byte(`{"type":"condition","condition":{"field":"status","operator":"is","value":"200"}}`), &fItem))
	filters := kbapi.KibanaHTTPAPIsLensPanelFilters{fItem}
	heatmap.Filters = &filters

	var attrs lenscommon.VisByValueConfig0
	require.NoError(t, attrs.FromKibanaHTTPAPIsHeatmapNoESQLByValuePanel(heatmap))

	blocks := &models.LensByValueChartBlocks{}
	diags := c.PopulateFromAttributes(ctx, blocks, attrs)
	require.False(t, diags.HasError(), "%v", diags)

	attrs2, diags := c.BuildAttributes(blocks)
	require.False(t, diags.HasError(), "%v", diags)

	heatmapRoundTrip, err := attrs2.AsKibanaHTTPAPIsHeatmapNoESQLByValuePanel()
	require.NoError(t, err)
	assert.Equal(t, kbapi.KibanaHTTPAPIsHeatmapNoESQLByValuePanelTypeHeatmap, heatmapRoundTrip.Type)
	require.NotNil(t, heatmapRoundTrip.Title)
	assert.Equal(t, "Test Heatmap", *heatmapRoundTrip.Title)
	require.NotNil(t, heatmapRoundTrip.Query)
	assert.Equal(t, "status:200", heatmapRoundTrip.Query.Expression)
}

func TestConverter_roundTrip_ESQL_heatmap(t *testing.T) {
	ctx := context.Background()
	var c converter
	const esqlJSON = `{
		"type": "heatmap",
		"title": "Heatmap ESQL RT",
		"description": "Converter test",
		"ignore_global_filters": false,
		"sampling": 1,
		"axis": { "x": {}, "y": {} },
		"styling": { "cells": {} },
		"legend": { "size": "m" },
		"data_source": {"type":"esql","query":"FROM logs-* | LIMIT 10"},
		"metric": {
			"color": {"type":"dynamic","range":"absolute","steps":[{"type":"from","from":0,"color":"#000000"}]},
			"column": "bytes",
			"format": {"type":"number"},
			"operation": "value"
		},
		"x": {"column":"host","format":{"type":"number"},"operation":"value"},
		"y": {"column":"service","format":{"type":"number"},"operation":"value"}
	}`
	var heatmap kbapi.KibanaHTTPAPIsHeatmapESQLByValuePanel
	require.NoError(t, json.Unmarshal([]byte(esqlJSON), &heatmap))

	var attrs lenscommon.VisByValueConfig0
	require.NoError(t, attrs.FromKibanaHTTPAPIsHeatmapESQLByValuePanel(heatmap))

	blocks := &models.LensByValueChartBlocks{}
	diags := c.PopulateFromAttributes(ctx, blocks, attrs)
	require.False(t, diags.HasError(), "%v", diags)
	require.NotNil(t, blocks.HeatmapConfig)
	require.Nil(t, blocks.HeatmapConfig.Query)
	assert.Contains(t, blocks.HeatmapConfig.DataSourceJSON.ValueString(), "FROM logs-*")

	attrs2, diags := c.BuildAttributes(blocks)
	require.False(t, diags.HasError(), "%v", diags)

	out, err := attrs2.AsKibanaHTTPAPIsHeatmapESQLByValuePanel()
	require.NoError(t, err)
	assert.Equal(t, kbapi.KibanaHTTPAPIsHeatmapESQLByValuePanelTypeHeatmap, out.Type)
	require.NotNil(t, out.Title)
	assert.Equal(t, "Heatmap ESQL RT", *out.Title)
	assert.Equal(t, "host", out.X.Column)
	dsBytes, err := json.Marshal(out.DataSource)
	require.NoError(t, err)
	assert.Contains(t, string(dsBytes), "FROM logs-*")
}

// Test_heatmapAxesFromAPI_preservesPriorWhenAPIDropsAxis covers a regression where
// Kibana may omit axis.x or axis.y entirely from a GET response (the kbapi spec
// types both as pointers with omitempty). Previously these were value-typed and
// always present, so the prior-preserve logic inside the per-axis converters was
// sufficient. Now that they are pointers, axis-level preservation must happen
// in heatmapAxesFromAPI itself, otherwise round-tripping a config that set
// axis.y produces an "inconsistent result after apply" because the read-back
// state drops axis.y to null.
func Test_heatmapAxesFromAPI_preservesPriorWhenAPIDropsAxis(t *testing.T) {
	prior := &models.HeatmapAxesModel{
		Y: &models.HeatmapYAxisModel{
			Labels: &models.HeatmapYAxisLabelsModel{Visible: types.BoolValue(true)},
			Title: &models.AxisTitleModel{
				Value:   types.StringValue("Y Axis"),
				Visible: types.BoolValue(true),
			},
		},
		X: &models.HeatmapXAxisModel{
			Labels: &models.HeatmapXAxisLabelsModel{
				Orientation: types.StringValue("horizontal"),
				Visible:     types.BoolValue(true),
			},
			Title: &models.AxisTitleModel{
				Value:   types.StringValue("X Axis"),
				Visible: types.BoolValue(true),
			},
		},
	}

	t.Run("api.Y nil preserves prior.Y", func(t *testing.T) {
		api := &kbapi.KibanaHTTPAPIsHeatmapAxes{
			X: &kbapi.KibanaHTTPAPIsHeatmapXAxis{},
		}
		got := &models.HeatmapAxesModel{}
		diags := heatmapAxesFromAPI(got, api, prior)
		require.False(t, diags.HasError(), "%v", diags)
		require.NotNil(t, got.Y, "axis.y was lost when API omitted it")
		assert.Equal(t, prior.Y, got.Y)
	})

	t.Run("api.X nil preserves prior.X", func(t *testing.T) {
		api := &kbapi.KibanaHTTPAPIsHeatmapAxes{
			Y: &kbapi.KibanaHTTPAPIsHeatmapYAxis{},
		}
		got := &models.HeatmapAxesModel{}
		diags := heatmapAxesFromAPI(got, api, prior)
		require.False(t, diags.HasError(), "%v", diags)
		require.NotNil(t, got.X, "axis.x was lost when API omitted it")
		assert.Equal(t, prior.X, got.X)
	})
}
