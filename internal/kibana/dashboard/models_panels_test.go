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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildLensMosaicPanelForTest creates a panel model with MosaicConfig for panelsToAPI tests.
func buildLensMosaicPanelForTest(t *testing.T) panelModel {
	t.Helper()
	groupBy := `[{"operation":"terms","collapse_by":"avg","fields":["host.name"],` +
		`"color":{"mode":"categorical","palette":"default","mapping":[],"unassignedColor":{"type":"color_code","value":"#D3DAE6"}}}]`
	groupBreakdownBy := `[{"operation":"terms","collapse_by":"avg","fields":["service.name"],` +
		`"color":{"mode":"categorical","palette":"default","mapping":[],"unassignedColor":{"type":"color_code","value":"#D3DAE6"}}}]`
	apiJSON := `{
		"type": "mosaic",
		"title": "Lens Mosaic",
		"dataset": {"type":"dataView","id":"metrics-*"},
		"query": {"language":"kuery","query":""},
		"legend": {"size":"small"},
		"metric": {"operation":"count"},
		"group_by": ` + groupBy + `,
		"group_breakdown_by": ` + groupBreakdownBy + `
	}`
	var api kbapi.MosaicNoESQL
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))

	var chart kbapi.MosaicChart
	require.NoError(t, chart.FromMosaicNoESQL(api))

	var attrs kbapi.KbnDashboardPanelLens_Config_0_Attributes
	require.NoError(t, attrs.FromMosaicChart(chart))

	converter := newMosaicPanelConfigConverter()
	pm := &panelModel{}
	diags := converter.populateFromAttributes(context.Background(), pm, attrs)
	require.False(t, diags.HasError())

	return panelModel{
		Type:         types.StringValue("lens"),
		ID:           types.StringValue("mosaic-1"),
		Grid:         panelGridModel{X: types.Int64Value(0), Y: types.Int64Value(0), W: types.Int64Value(6), H: types.Int64Value(6)},
		MosaicConfig: pm.MosaicConfig,
	}
}

// buildLensTreemapPanelForTest creates a panel model with TreemapConfig for panelsToAPI tests.
func buildLensTreemapPanelForTest(t *testing.T) panelModel {
	t.Helper()
	apiJSON := `{
		"type": "treemap",
		"title": "Lens Treemap",
		"dataset": {"type":"dataView","id":"metrics-*"},
		"query": {"language":"kuery","query":""},
		"legend": {"size":"small"},
		"metrics": [{"operation":"count"}],
		"group_by": [{"operation":"terms","field":"host.name","collapse_by":"avg"}]
	}`
	var api kbapi.TreemapNoESQL
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))

	var chart kbapi.TreemapChart
	require.NoError(t, chart.FromTreemapNoESQL(api))

	var attrs kbapi.KbnDashboardPanelLens_Config_0_Attributes
	require.NoError(t, attrs.FromTreemapChart(chart))

	converter := newTreemapPanelConfigConverter()
	pm := &panelModel{}
	diags := converter.populateFromAttributes(context.Background(), pm, attrs)
	require.False(t, diags.HasError())

	return panelModel{
		Type:          types.StringValue("lens"),
		ID:            types.StringValue("treemap-1"),
		Grid:          panelGridModel{X: types.Int64Value(0), Y: types.Int64Value(0), W: types.Int64Value(6), H: types.Int64Value(6)},
		TreemapConfig: pm.TreemapConfig,
	}
}

// buildLensWafflePanelForTest creates a panel model with WaffleConfig for panelsToAPI tests.
func buildLensWafflePanelForTest(t *testing.T) panelModel {
	t.Helper()
	apiJSON := `{
		"type": "waffle",
		"title": "Lens Waffle",
		"dataset": {"type":"dataView","id":"metrics-*"},
		"query": {"language":"kuery","query":""},
		"legend": {"size":"small"},
		"metrics": [{"operation":"count"}]
	}`
	var api kbapi.WaffleNoESQL
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))

	var chart kbapi.WaffleChart
	require.NoError(t, chart.FromWaffleNoESQL(api))

	var attrs kbapi.KbnDashboardPanelLens_Config_0_Attributes
	require.NoError(t, attrs.FromWaffleChart(chart))

	converter := newWafflePanelConfigConverter()
	pm := &panelModel{}
	diags := converter.populateFromAttributes(context.Background(), pm, attrs)
	require.False(t, diags.HasError())

	return panelModel{
		Type:         types.StringValue("lens"),
		ID:           types.StringValue("waffle-1"),
		Grid:         panelGridModel{X: types.Int64Value(0), Y: types.Int64Value(0), W: types.Int64Value(8), H: types.Int64Value(10)},
		WaffleConfig: pm.WaffleConfig,
	}
}

func Test_lensPanelTimeRange(t *testing.T) {
	tr := lensPanelTimeRange()
	assert.Equal(t, "now-15m", tr.From)
	assert.Equal(t, "now", tr.To)
}

func Test_mapPanelsFromAPI(t *testing.T) {
	tests := []struct {
		name             string
		apiPanelsJSON    string
		expectedPanels   []panelModel
		expectedSections []sectionModel
	}{
		{
			name:             "empty panels",
			apiPanelsJSON:    "[]",
			expectedPanels:   nil,
			expectedSections: nil,
		},
		{
			name: "basic panel with structured config",
			apiPanelsJSON: `[
				{
					"grid": {
						"x": 0,
						"y": 1,
						"w": 2,
						"h": 3
					},
					"uid": "1",
					"type": "markdown",
					"config": {
						"title": "My Panel",
						"content": "some content",
                        "hide_title": true
					}
				}
			]`,
			expectedPanels: []panelModel{
				{
					Type: types.StringValue("markdown"),
					Grid: panelGridModel{
						X: types.Int64Value(0),
						Y: types.Int64Value(1),
						W: types.Int64Value(2),
						H: types.Int64Value(3),
					},
					ID: types.StringValue("1"),
					MarkdownConfig: &markdownConfigModel{
						Title:       types.StringValue("My Panel"),
						Content:     types.StringValue("some content"),
						HideTitle:   types.BoolValue(true),
						Description: types.StringNull(),
					},
					ConfigJSON: customtypes.NewJSONWithDefaultsValue(`{
						"title": "My Panel",
						"content": "some content",
                        "hide_title": true
					}`, populatePanelConfigJSONDefaults),
				},
			},
		},
		{
			name: "panel with markdown config_json fallback",
			apiPanelsJSON: `[
				{
					"grid": {
						"x": 10,
						"y": 20
					},
					"type": "markdown",
					"config": {"unknownField": "something"}
				}
			]`,
			expectedPanels: []panelModel{
				{
					Type: types.StringValue("markdown"),
					Grid: panelGridModel{
						X: types.Int64Value(10),
						Y: types.Int64Value(20),
						W: types.Int64Null(),
						H: types.Int64Null(),
					},
					ID: types.StringNull(),
					MarkdownConfig: &markdownConfigModel{
						Title:       types.StringNull(),
						Content:     types.StringNull(),
						HideTitle:   types.BoolNull(),
						Description: types.StringNull(),
					},
					ConfigJSON: customtypes.NewJSONWithDefaultsValue(`{"unknownField": "something"}`, populatePanelConfigJSONDefaults),
				},
			},
		},
		{
			name: "section with panels",
			apiPanelsJSON: `[
				{
					"title": "My Section",
					"grid": { "y": 100 },
					"collapsed": true,
					"uid": "section1",
					"panels": [
						{
							"type": "markdown",
							"grid": { "x": 0, "y": 0, "w": 4, "h": 4 },
							"config": { "title": "Inner Panel", "content": "Inner content" }
						}
					]
				}
			]`,
			expectedSections: []sectionModel{
				{
					Title:     types.StringValue("My Section"),
					ID:        types.StringValue("section1"),
					Collapsed: types.BoolValue(true),
					Grid: sectionGridModel{
						Y: types.Int64Value(100),
					},
					Panels: []panelModel{
						{
							Type: types.StringValue("markdown"),
							Grid: panelGridModel{
								X: types.Int64Value(0),
								Y: types.Int64Value(0),
								W: types.Int64Value(4),
								H: types.Int64Value(4),
							},
							ID: types.StringNull(),
							MarkdownConfig: &markdownConfigModel{
								Title:       types.StringValue("Inner Panel"),
								Content:     types.StringValue("Inner content"),
								HideTitle:   types.BoolNull(),
								Description: types.StringNull(),
							},
							ConfigJSON: customtypes.NewJSONWithDefaultsValue(`{ "title": "Inner Panel", "content": "Inner content" }`, populatePanelConfigJSONDefaults),
						},
					},
				},
			},
		},
		{
			name: "mix of panels and sections",
			apiPanelsJSON: `[
				{
					"grid": { "x": 0, "y": 0, "w": 6, "h": 6 },
					"type": "markdown",
					"uid": "panel1",
					"config": { "title": "Panel 1", "content": "Panel 1 body" }
				},
				{
					"title": "Section 1",
					"grid": { "y": 100 },
					"uid": "section1",
					"panels": [
						{
							"type": "markdown",
							"grid": { "x": 0, "y": 0, "w": 6, "h": 6 },
							"config": { "title": "Inner Panel", "content": "Inner panel body" }
						}
					]
				},
				{
					"grid": { "x": 6, "y": 0, "w": 6, "h": 6 },
					"type": "lens",
					"uid": "panel2",
					"config": { "title": "Panel 2" }
				}
			]`,
			expectedPanels: []panelModel{
				{
					Type: types.StringValue("markdown"),
					Grid: panelGridModel{
						X: types.Int64Value(0),
						Y: types.Int64Value(0),
						W: types.Int64Value(6),
						H: types.Int64Value(6),
					},
					ID: types.StringValue("panel1"),
					MarkdownConfig: &markdownConfigModel{
						Title:       types.StringValue("Panel 1"),
						Content:     types.StringValue("Panel 1 body"),
						HideTitle:   types.BoolNull(),
						Description: types.StringNull(),
					},
					ConfigJSON: customtypes.NewJSONWithDefaultsValue(`{ "title": "Panel 1", "content": "Panel 1 body" }`, populatePanelConfigJSONDefaults),
				},
				{
					Type: types.StringValue("lens"),
					Grid: panelGridModel{
						X: types.Int64Value(6),
						Y: types.Int64Value(0),
						W: types.Int64Value(6),
						H: types.Int64Value(6),
					},
					ID:             types.StringValue("panel2"),
					MarkdownConfig: nil,
					ConfigJSON:     customtypes.NewJSONWithDefaultsValue(`{ "title": "Panel 2" }`, populatePanelConfigJSONDefaults),
				},
			},
			expectedSections: []sectionModel{
				{
					Title:     types.StringValue("Section 1"),
					ID:        types.StringValue("section1"),
					Collapsed: types.BoolNull(),
					Grid: sectionGridModel{
						Y: types.Int64Value(100),
					},
					Panels: []panelModel{
						{
							Type: types.StringValue("markdown"),
							Grid: panelGridModel{
								X: types.Int64Value(0),
								Y: types.Int64Value(0),
								W: types.Int64Value(6),
								H: types.Int64Value(6),
							},
							ID: types.StringNull(),
							MarkdownConfig: &markdownConfigModel{
								Title:       types.StringValue("Inner Panel"),
								Content:     types.StringValue("Inner panel body"),
								HideTitle:   types.BoolNull(),
								Description: types.StringNull(),
							},
							ConfigJSON: customtypes.NewJSONWithDefaultsValue(`{ "title": "Inner Panel", "content": "Inner panel body" }`, populatePanelConfigJSONDefaults),
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var apiPanels kbapi.DashboardPanels
			err := json.Unmarshal([]byte(tt.apiPanelsJSON), &apiPanels)
			require.NoError(t, err)

			model := &dashboardModel{}
			panels, sections, diags := model.mapPanelsFromAPI(t.Context(), &apiPanels)
			require.False(t, diags.HasError())

			assert.Len(t, panels, len(tt.expectedPanels))
			for i := range panels {
				assertPanelsEqual(t, tt.expectedPanels[i], panels[i])
			}
			assert.Len(t, sections, len(tt.expectedSections))
			for i := range sections {
				assertSectionsEqual(t, tt.expectedSections[i], sections[i])
			}
		})
	}
}

// assertPanelsEqual compares two panelModels, using semantic equality for ConfigJSON
// since the API may return different JSON formatting (whitespace, key order).
func assertPanelsEqual(t *testing.T, expected, actual panelModel) {
	t.Helper()
	assert.Equal(t, expected.Type, actual.Type)
	assert.Equal(t, expected.Grid, actual.Grid)
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.MarkdownConfig, actual.MarkdownConfig)
	assert.Equal(t, expected.XYChartConfig, actual.XYChartConfig)
	assert.Equal(t, expected.TreemapConfig, actual.TreemapConfig)
	assert.Equal(t, expected.DatatableConfig, actual.DatatableConfig)
	assert.Equal(t, expected.TagcloudConfig, actual.TagcloudConfig)
	assert.Equal(t, expected.MetricChartConfig, actual.MetricChartConfig)
	assert.Equal(t, expected.PieChartConfig, actual.PieChartConfig)
	assert.Equal(t, expected.GaugeConfig, actual.GaugeConfig)
	assert.Equal(t, expected.LegacyMetricConfig, actual.LegacyMetricConfig)
	assert.Equal(t, expected.RegionMapConfig, actual.RegionMapConfig)
	assert.Equal(t, expected.HeatmapConfig, actual.HeatmapConfig)
	// ConfigJSON: use semantic equality (handles formatting differences)
	ctx := context.Background()
	eq, diags := expected.ConfigJSON.StringSemanticEquals(ctx, actual.ConfigJSON)
	require.False(t, diags.HasError())
	assert.True(t, eq, "ConfigJSON should be semantically equal: expected %q, got %q",
		expected.ConfigJSON.ValueString(), actual.ConfigJSON.ValueString())
}

func assertSectionsEqual(t *testing.T, expected, actual sectionModel) {
	t.Helper()
	assert.Equal(t, expected.Title, actual.Title)
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.Collapsed, actual.Collapsed)
	assert.Equal(t, expected.Grid, actual.Grid)
	assert.Len(t, actual.Panels, len(expected.Panels))
	for i := range actual.Panels {
		assertPanelsEqual(t, expected.Panels[i], actual.Panels[i])
	}
}

func Test_panelsToAPI(t *testing.T) {
	tests := []struct {
		name     string
		model    dashboardModel
		expected string // JSON representation of API panels for easy comparison
	}{
		{
			name: "basic panel with structured config",
			model: dashboardModel{
				Panels: []panelModel{
					{
						Type: types.StringValue("markdown"),
						Grid: panelGridModel{
							X: types.Int64Value(0),
							Y: types.Int64Value(1),
							W: types.Int64Value(2),
							H: types.Int64Value(3),
						},
						ID: types.StringValue("1"),
						MarkdownConfig: &markdownConfigModel{
							Title:     types.StringValue("My Panel"),
							Content:   types.StringValue("some content"),
							HideTitle: types.BoolValue(true),
						},
						ConfigJSON: customtypes.NewJSONWithDefaultsNull(populatePanelConfigJSONDefaults),
					},
				},
			},
			expected: `[
				{
					"grid": {
						"h": 3,
						"w": 2,
						"x": 0,
						"y": 1
					},
					"uid": "1",
					"type": "markdown",
					"config": {
						"content": "some content",
                        "hide_title": true,
						"title": "My Panel"
					}
				}
			]`,
		},
		{
			name: "panel with markdown config_json",
			model: dashboardModel{
				Panels: []panelModel{
					{
						Type: types.StringValue("markdown"),
						Grid: panelGridModel{
							X: types.Int64Value(10),
							Y: types.Int64Value(20),
						},
						ID:             types.StringNull(),
						MarkdownConfig: nil,
						ConfigJSON:     customtypes.NewJSONWithDefaultsValue(`{"content":"from json"}`, populatePanelConfigJSONDefaults),
					},
				},
			},
			expected: `[
				{
					"grid": {
						"x": 10,
						"y": 20
					},
					"type": "markdown",
					"config": {
						"content": "from json"
					}
				}
			]`,
		},
		{
			name: "lens panel with treemap config",
			model: dashboardModel{
				Panels: []panelModel{
					buildLensTreemapPanelForTest(t),
				},
			},
			expected: `[
				{
					"grid": {"h": 6, "w": 6, "x": 0, "y": 0},
					"uid": "treemap-1",
					"type": "lens",
					"config": {
						"attributes": {
							"type": "treemap",
							"title": "Lens Treemap",
							"dataset": {"type":"dataView","id":"metrics-*"},
							"query": {"language":"kuery","query":""},
							"legend": {"size":"small"},
							"metrics": [{"operation":"count"}],
							"group_by": [{"operation":"terms","field":"host.name","collapse_by":"avg"}],
							"value_display": {"mode": ""}
						},
						"time_range": {"from": "now-15m", "to": "now"}
					}
				}
			]`,
		},
		{
			name: "lens panel with mosaic config",
			model: dashboardModel{
				Panels: []panelModel{
					buildLensMosaicPanelForTest(t),
				},
			},
			expected: `[
				{
					"grid": {"h": 6, "w": 6, "x": 0, "y": 0},
					"uid": "mosaic-1",
					"type": "lens",
					"config": {
						"attributes": {
							"type": "mosaic",
							"title": "Lens Mosaic",
							"dataset": {"type":"dataView","id":"metrics-*"},
							"query": {"language":"kuery","query":""},
							"legend": {"size":"small"},
							"metric": {"operation":"count"},
							"group_by": [{"operation":"terms","collapse_by":"avg","fields":["host.name"],
								"color":{"mode":"categorical","palette":"default","mapping":[],
								"unassignedColor":{"type":"color_code","value":"#D3DAE6"}}}],
							"group_breakdown_by": [{"operation":"terms","collapse_by":"avg","fields":["service.name"],
								"color":{"mode":"categorical","palette":"default","mapping":[],
								"unassignedColor":{"type":"color_code","value":"#D3DAE6"}}}],
							"value_display": {"mode": ""}
						},
						"time_range": {"from": "now-15m", "to": "now"}
					}
				}
			]`,
		},
		{
			name: "lens panel with waffle config",
			model: dashboardModel{
				Panels: []panelModel{
					buildLensWafflePanelForTest(t),
				},
			},
			expected: `[
				{
					"grid": {"h": 10, "w": 8, "x": 0, "y": 0},
					"uid": "waffle-1",
					"type": "lens",
					"config": {
						"attributes": {
							"type": "waffle",
							"title": "Lens Waffle",
							"dataset": {"type":"dataView","id":"metrics-*"},
							"query": {"language":"kuery","query":""},
							"legend": {"size":"small"},
							"metrics": [{"operation":"count"}],
							"value_display": {"mode": "percentage"}
						},
						"time_range": {"from": "now-15m", "to": "now"}
					}
				}
			]`,
		},
		{
			name: "section with panel",
			model: dashboardModel{
				Sections: []sectionModel{
					{
						Title:     types.StringValue("Test Section"),
						ID:        types.StringValue("sec-1"),
						Collapsed: types.BoolValue(true),
						Grid:      sectionGridModel{Y: types.Int64Value(50)},
						Panels: []panelModel{
							{
								Type: types.StringValue("markdown"),
								Grid: panelGridModel{X: types.Int64Value(0), Y: types.Int64Value(0), W: types.Int64Value(5), H: types.Int64Value(5)},
								MarkdownConfig: &markdownConfigModel{
									Title: types.StringValue("Inner Text"),
								},
							},
						},
					},
				},
			},
			expected: `[
				{
					"title": "Test Section",
					"uid": "sec-1",
					"collapsed": true,
					"grid": {"y": 50},
					"panels": [
						{"grid":{"h":5,"w":5,"x":0,"y":0},"type":"markdown","config":{"title":"Inner Text"}}
					]
				}
			]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, diags := tt.model.panelsToAPI()
			require.False(t, diags.HasError())

			jsonBytes, err := json.Marshal(result)
			require.NoError(t, err)

			var expectedJSON any
			var actualJSON any

			require.NoError(t, json.Unmarshal([]byte(tt.expected), &expectedJSON))
			require.NoError(t, json.Unmarshal(jsonBytes, &actualJSON))

			assert.Equal(t, expectedJSON, actualJSON)
		})
	}
}
