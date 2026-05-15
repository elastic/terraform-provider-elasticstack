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
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubResolver struct{}

func (stubResolver) ResolveChartTimeRange(chartLevel *models.TimeRangeModel) kbapi.KbnEsQueryServerTimeRangeSchema {
	_ = chartLevel
	return kbapi.KbnEsQueryServerTimeRangeSchema{}
}

func (stubResolver) DashboardLensComparableTimeRange() (kbapi.KbnEsQueryServerTimeRangeSchema, bool) {
	return kbapi.KbnEsQueryServerTimeRangeSchema{}, false
}

func TestConverter_VizType(t *testing.T) {
	var c converter
	require.Equal(t, string(kbapi.HeatmapNoESQLTypeHeatmap), c.VizType())
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
	resolver := stubResolver{}

	heatmap := kbapi.HeatmapNoESQL{
		Type:                kbapi.HeatmapNoESQLTypeHeatmap,
		Title:               new("Test Heatmap"),
		Description:         new("Heatmap description"),
		IgnoreGlobalFilters: new(true),
		Sampling:            newFloat32(0.5),
		Query: kbapi.FilterSimple{
			Expression: "status:200",
			Language: func() *kbapi.FilterSimpleLanguage {
				lang := kbapi.FilterSimpleLanguage("kql")
				return &lang
			}(),
		},
		Axis: kbapi.HeatmapAxes{
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
		Styling: kbapi.HeatmapStyling{
			Cells: kbapi.HeatmapCells{
				Labels: &struct {
					Visible *bool `json:"visible,omitempty"`
				}{
					Visible: new(true),
				},
			},
		},
		Legend: kbapi.HeatmapLegend{
			Size: kbapi.LegendSizeM,
			Visibility: func() *kbapi.HeatmapLegendVisibility {
				visibility := kbapi.HeatmapLegendVisibilityVisible
				return &visibility
			}(),
			TruncateAfterLines: newFloat32(4),
		},
	}

	require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"metrics-*"}`), &heatmap.DataSource))
	require.NoError(t, json.Unmarshal([]byte(`{"operation":"count"}`), &heatmap.Metric))
	require.NoError(t, json.Unmarshal([]byte(`{"operation":"filters","filters":[{"label":"All","filter":{"query":"*","language":"kql"}}]}`), &heatmap.X))
	var yAxis kbapi.HeatmapNoESQL_Y
	require.NoError(t, json.Unmarshal([]byte(`{"operation":"filters","filters":[{"label":"All","filter":{"query":"*","language":"kql"}}]}`), &yAxis))
	heatmap.Y = &yAxis

	var fItem kbapi.LensPanelFilters_Item
	require.NoError(t, json.Unmarshal([]byte(`{"type":"condition","condition":{"field":"status","operator":"is","value":"200"}}`), &fItem))
	heatmap.Filters = []kbapi.LensPanelFilters_Item{fItem}

	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	require.NoError(t, attrs.FromHeatmapNoESQL(heatmap))

	blocks := &models.LensByValueChartBlocks{}
	diags := c.PopulateFromAttributes(ctx, resolver, blocks, attrs)
	require.False(t, diags.HasError(), "%v", diags)

	attrs2, diags := c.BuildAttributes(blocks, resolver)
	require.False(t, diags.HasError(), "%v", diags)

	heatmapRoundTrip, err := attrs2.AsHeatmapNoESQL()
	require.NoError(t, err)
	assert.Equal(t, kbapi.HeatmapNoESQLTypeHeatmap, heatmapRoundTrip.Type)
	require.NotNil(t, heatmapRoundTrip.Title)
	assert.Equal(t, "Test Heatmap", *heatmapRoundTrip.Title)
	assert.Equal(t, "status:200", heatmapRoundTrip.Query.Expression)
}

func TestConverter_roundTrip_ESQL_heatmap(t *testing.T) {
	ctx := context.Background()
	var c converter
	resolver := stubResolver{}

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
	var heatmap kbapi.HeatmapESQL
	require.NoError(t, json.Unmarshal([]byte(esqlJSON), &heatmap))

	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	require.NoError(t, attrs.FromHeatmapESQL(heatmap))

	blocks := &models.LensByValueChartBlocks{}
	diags := c.PopulateFromAttributes(ctx, resolver, blocks, attrs)
	require.False(t, diags.HasError(), "%v", diags)
	require.NotNil(t, blocks.HeatmapConfig)
	require.Nil(t, blocks.HeatmapConfig.Query)
	assert.Contains(t, blocks.HeatmapConfig.DataSourceJSON.ValueString(), "FROM logs-*")

	attrs2, diags := c.BuildAttributes(blocks, resolver)
	require.False(t, diags.HasError(), "%v", diags)

	out, err := attrs2.AsHeatmapESQL()
	require.NoError(t, err)
	assert.Equal(t, kbapi.HeatmapESQLTypeHeatmap, out.Type)
	require.NotNil(t, out.Title)
	assert.Equal(t, "Heatmap ESQL RT", *out.Title)
	assert.Equal(t, "host", out.X.Column)
	dsBytes, err := json.Marshal(out.DataSource)
	require.NoError(t, err)
	assert.Contains(t, string(dsBytes), "FROM logs-*")
}

//go:fix inline
func newFloat32(f float32) *float32 { return new(f) }
