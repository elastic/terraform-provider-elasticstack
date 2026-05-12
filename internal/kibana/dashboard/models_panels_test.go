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
		`"color":{"mode":"categorical","palette":"default","mapping":[],"unassigned":{"type":"color_code","value":"#D3DAE6"}}}]`
	groupBreakdownBy := `[{"operation":"terms","collapse_by":"avg","fields":["service.name"],` +
		`"color":{"mode":"categorical","palette":"default","mapping":[],"unassigned":{"type":"color_code","value":"#D3DAE6"}}}]`
	apiJSON := `{
		"type": "mosaic",
		"title": "Lens Mosaic",
		"data_source": {"type":"dataView","id":"metrics-*"},
		"query": {"language":"kql","expression":""},
		"legend": {"size":"small"},
		"metric": {"operation":"count"},
		"group_by": ` + groupBy + `,
		"group_breakdown_by": ` + groupBreakdownBy + `
	}`
	var api kbapi.MosaicNoESQL
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))

	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	require.NoError(t, attrs.FromMosaicNoESQL(api))

	converter := newMosaicPanelConfigConverter()
	visBv := visByValueModel{}
	diags := converter.populateFromAttributes(context.Background(), nil, nil, &visBv.lensByValueChartBlocks, attrs)
	require.False(t, diags.HasError())

	return panelModel{
		Type: types.StringValue("vis"),
		ID:   types.StringValue("mosaic-1"),
		Grid: panelGridModel{X: types.Int64Value(0), Y: types.Int64Value(0), W: types.Int64Value(6), H: types.Int64Value(6)},
		VisConfig: &visConfigModel{
			ByValue: &visBv,
		},
		ConfigJSON: customtypes.NewJSONWithDefaultsNull(populatePanelConfigJSONDefaults),
	}
}

// buildLensTreemapPanelForTest creates a panel model with TreemapConfig for panelsToAPI tests.
func buildLensTreemapPanelForTest(t *testing.T) panelModel {
	t.Helper()
	apiJSON := `{
		"type": "treemap",
		"title": "Lens Treemap",
		"data_source": {"type":"dataView","id":"metrics-*"},
		"query": {"language":"kql","expression":""},
		"legend": {"size":"small"},
		"metrics": [{"operation":"count"}],
		"group_by": [{"operation":"terms","field":"host.name","collapse_by":"avg"}]
	}`
	var api kbapi.TreemapNoESQL
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))

	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	require.NoError(t, attrs.FromTreemapNoESQL(api))

	converter := newTreemapPanelConfigConverter()
	visBv := visByValueModel{}
	diags := converter.populateFromAttributes(context.Background(), nil, nil, &visBv.lensByValueChartBlocks, attrs)
	require.False(t, diags.HasError())

	return panelModel{
		Type: types.StringValue("vis"),
		ID:   types.StringValue("treemap-1"),
		Grid: panelGridModel{X: types.Int64Value(0), Y: types.Int64Value(0), W: types.Int64Value(6), H: types.Int64Value(6)},
		VisConfig: &visConfigModel{
			ByValue: &visBv,
		},
		ConfigJSON: customtypes.NewJSONWithDefaultsNull(populatePanelConfigJSONDefaults),
	}
}

// buildLensWafflePanelForTest creates a panel model with WaffleConfig for panelsToAPI tests.
func buildLensWafflePanelForTest(t *testing.T) panelModel {
	t.Helper()
	apiJSON := `{
		"type": "waffle",
		"title": "Lens Waffle",
		"data_source": {"type":"dataView","id":"metrics-*"},
		"query": {"language":"kql","expression":""},
		"legend": {"size":"small"},
		"metrics": [{"operation":"count"}]
	}`
	var api kbapi.WaffleNoESQL
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))

	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	require.NoError(t, attrs.FromWaffleNoESQL(api))

	converter := newWafflePanelConfigConverter()
	visBv := visByValueModel{}
	diags := converter.populateFromAttributes(context.Background(), nil, nil, &visBv.lensByValueChartBlocks, attrs)
	require.False(t, diags.HasError())

	return panelModel{
		Type: types.StringValue("vis"),
		ID:   types.StringValue("waffle-1"),
		Grid: panelGridModel{X: types.Int64Value(0), Y: types.Int64Value(0), W: types.Int64Value(8), H: types.Int64Value(10)},
		VisConfig: &visConfigModel{
			ByValue: &visBv,
		},
		ConfigJSON: customtypes.NewJSONWithDefaultsNull(populatePanelConfigJSONDefaults),
	}
}

func Test_resolveChartTimeRange_defaultWhenNoDashboard(t *testing.T) {
	dash := &dashboardModel{
		TimeRange: &timeRangeModel{
			From: types.StringValue("now-7d"),
			To:   types.StringValue("now"),
		},
	}

	chartTR := &timeRangeModel{
		From: types.StringValue("now-30d"),
		To:   types.StringValue("now-1d"),
	}

	got := resolveChartTimeRange(dash, chartTR)
	assert.Equal(t, "now-30d", got.From)
	assert.Equal(t, "now-1d", got.To)

	gotInherit := resolveChartTimeRange(dash, nil)
	assert.Equal(t, "now-7d", gotInherit.From)
	assert.Equal(t, "now", gotInherit.To)

	// Scratch paths (no dashboard model) fall back to the legacy window when chart time is unset.
	gotScratch := resolveChartTimeRange(nil, nil)
	assert.Equal(t, "now-15m", gotScratch.From)
	assert.Equal(t, "now", gotScratch.To)
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
					"id": "1",
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
						ByValue: &markdownConfigByValueModel{
							Content:     types.StringValue("some content"),
							Title:       types.StringValue("My Panel"),
							HideTitle:   types.BoolValue(true),
							Description: types.StringNull(),
							HideBorder:  types.BoolNull(),
							Settings: &markdownConfigSettingsModel{
								OpenLinksInNewTab: types.BoolNull(),
							},
						},
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
						ByValue: &markdownConfigByValueModel{
							Content:     types.StringValue(""),
							Title:       types.StringNull(),
							HideTitle:   types.BoolNull(),
							Description: types.StringNull(),
							HideBorder:  types.BoolNull(),
							Settings: &markdownConfigSettingsModel{
								OpenLinksInNewTab: types.BoolNull(),
							},
						},
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
					"id": "section1",
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
								ByValue: &markdownConfigByValueModel{
									Content:     types.StringValue("Inner content"),
									Title:       types.StringValue("Inner Panel"),
									HideTitle:   types.BoolNull(),
									Description: types.StringNull(),
									HideBorder:  types.BoolNull(),
									Settings: &markdownConfigSettingsModel{
										OpenLinksInNewTab: types.BoolNull(),
									},
								},
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
					"id": "panel1",
					"config": { "title": "Panel 1", "content": "Panel 1 body" }
				},
				{
					"title": "Section 1",
					"grid": { "y": 100 },
					"id": "section1",
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
					"type": "vis",
					"id": "panel2",
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
						ByValue: &markdownConfigByValueModel{
							Content:     types.StringValue("Panel 1 body"),
							Title:       types.StringValue("Panel 1"),
							HideTitle:   types.BoolNull(),
							Description: types.StringNull(),
							HideBorder:  types.BoolNull(),
							Settings: &markdownConfigSettingsModel{
								OpenLinksInNewTab: types.BoolNull(),
							},
						},
					},
					ConfigJSON: customtypes.NewJSONWithDefaultsValue(`{ "title": "Panel 1", "content": "Panel 1 body" }`, populatePanelConfigJSONDefaults),
				},
				{
					Type: types.StringValue("vis"),
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
								ByValue: &markdownConfigByValueModel{
									Content:     types.StringValue("Inner panel body"),
									Title:       types.StringValue("Inner Panel"),
									HideTitle:   types.BoolNull(),
									Description: types.StringNull(),
									HideBorder:  types.BoolNull(),
									Settings: &markdownConfigSettingsModel{
										OpenLinksInNewTab: types.BoolNull(),
									},
								},
							},
							ConfigJSON: customtypes.NewJSONWithDefaultsValue(`{ "title": "Inner Panel", "content": "Inner panel body" }`, populatePanelConfigJSONDefaults),
						},
					},
				},
			},
		},
		{
			name: "unknown panel type preserves id, grid, and config",
			apiPanelsJSON: `[
				{
					"grid": {"x": 0, "y": 1, "w": 48, "h": 15},
					"id": "unknown-1",
					"type": "custom_unknown_panel",
					"config": {
						"timeRange": {"from": "now-30d", "to": "now"},
						"columns": ["_source"],
						"sort": [{"@timestamp": "desc"}]
					}
				}
			]`,
			expectedPanels: []panelModel{
				{
					Type: types.StringValue("custom_unknown_panel"),
					Grid: panelGridModel{
						X: types.Int64Value(0),
						Y: types.Int64Value(1),
						W: types.Int64Value(48),
						H: types.Int64Value(15),
					},
					ID:         types.StringValue("unknown-1"),
					ConfigJSON: customtypes.NewJSONWithDefaultsValue(`{"timeRange":{"from":"now-30d","to":"now"},"columns":["_source"],"sort":[{"@timestamp":"desc"}]}`, populatePanelConfigJSONDefaults),
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
	assert.Equal(t, expected.VisConfig, actual.VisConfig)
	assert.Equal(t, expected.LensDashboardAppConfig, actual.LensDashboardAppConfig)
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
							ByValue: &markdownConfigByValueModel{
								Content:   types.StringValue("some content"),
								Title:     types.StringValue("My Panel"),
								HideTitle: types.BoolValue(true),
								Settings: &markdownConfigSettingsModel{
									OpenLinksInNewTab: types.BoolNull(),
								},
							},
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
					"id": "1",
					"type": "markdown",
					"config": {
						"content": "some content",
                        "hide_title": true,
						"settings": {},
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
				TimeRange: &timeRangeModel{
					From: types.StringValue("now-15m"),
					To:   types.StringValue("now"),
				},
				Panels: []panelModel{
					buildLensTreemapPanelForTest(t),
				},
			},
			expected: `[
				{
					"grid": {"h": 6, "w": 6, "x": 0, "y": 0},
					"id": "treemap-1",
					"type": "vis",
					"config": {
						"type": "treemap",
						"title": "Lens Treemap",
						"data_source": {"type":"dataView","id":"metrics-*"},
						"filters": [],
						"query": {"language":"kql","expression":""},
						"legend": {"size":"small"},
						"metrics": [{"operation":"count"}],
						"group_by": [{"operation":"terms","field":"host.name","collapse_by":"avg"}],
						"styling": {"values": {}},
						"time_range": {"from": "now-15m", "to": "now"}
					}
				}
			]`,
		},
		{
			name: "lens panel with mosaic config",
			model: dashboardModel{
				TimeRange: &timeRangeModel{
					From: types.StringValue("now-15m"),
					To:   types.StringValue("now"),
				},
				Panels: []panelModel{
					buildLensMosaicPanelForTest(t),
				},
			},
			expected: `[
				{
					"grid": {"h": 6, "w": 6, "x": 0, "y": 0},
					"id": "mosaic-1",
					"type": "vis",
					"config": {
						"type": "mosaic",
						"title": "Lens Mosaic",
						"data_source": {"type":"dataView","id":"metrics-*"},
						"filters": [],
						"query": {"language":"kql","expression":""},
						"legend": {"size":"small"},
						"metric": {"operation":"count"},
						"group_by": [{"operation":"terms","collapse_by":"avg","fields":["host.name"],
							"color":{"mode":"categorical","palette":"default","mapping":[],
							"unassigned":{"type":"color_code","value":"#D3DAE6"}}}],
						"group_breakdown_by": [{"operation":"terms","collapse_by":"avg","fields":["service.name"],
							"color":{"mode":"categorical","palette":"default","mapping":[],
							"unassigned":{"type":"color_code","value":"#D3DAE6"}}}],
						"styling": {"values": {}},
						"time_range": {"from": "now-15m", "to": "now"}
					}
				}
			]`,
		},
		{
			name: "lens panel with waffle config",
			model: dashboardModel{
				TimeRange: &timeRangeModel{
					From: types.StringValue("now-15m"),
					To:   types.StringValue("now"),
				},
				Panels: []panelModel{
					buildLensWafflePanelForTest(t),
				},
			},
			expected: `[
				{
					"grid": {"h": 10, "w": 8, "x": 0, "y": 0},
					"id": "waffle-1",
					"type": "vis",
					"config": {
						"type": "waffle",
						"title": "Lens Waffle",
						"data_source": {"type":"dataView","id":"metrics-*"},
						"filters": [],
						"query": {"language":"kql","expression":""},
						"legend": {"size":"small"},
						"metrics": [{"operation":"count"}],
						"styling": {"values": {"mode": "percentage"}},
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
									ByValue: &markdownConfigByValueModel{
										Content: types.StringValue(""),
										Title:   types.StringValue("Inner Text"),
										Settings: &markdownConfigSettingsModel{
											OpenLinksInNewTab: types.BoolNull(),
										},
									},
								},
							},
						},
					},
				},
			},
			expected: `[
				{
					"title": "Test Section",
					"id": "sec-1",
					"collapsed": true,
					"grid": {"y": 50},
					"panels": [
						{"grid":{"h":5,"w":5,"x":0,"y":0},"type":"markdown","config":{"content":"","settings":{},"title":"Inner Text"}}
					]
				}
			]`,
		},
		{
			name: "unknown panel type replays config_json",
			model: dashboardModel{
				Panels: []panelModel{
					{
						Type: types.StringValue("custom_unknown_panel"),
						Grid: panelGridModel{
							X: types.Int64Value(0),
							Y: types.Int64Value(1),
							W: types.Int64Value(48),
							H: types.Int64Value(15),
						},
						ID:         types.StringValue("unknown-1"),
						ConfigJSON: customtypes.NewJSONWithDefaultsValue(`{"timeRange":{"from":"now-30d","to":"now"},"columns":["_source"],"sort":[{"@timestamp":"desc"}]}`, populatePanelConfigJSONDefaults),
					},
				},
			},
			expected: `[
				{
					"grid": {"x": 0, "y": 1, "w": 48, "h": 15},
					"id": "unknown-1",
					"type": "custom_unknown_panel",
					"config": {
						"timeRange": {"from": "now-30d", "to": "now"},
						"columns": ["_source"],
						"sort": [{"@timestamp": "desc"}]
					}
				}
			]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, diags := tt.model.panelsToAPI(context.Background())
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

func Test_panelModel_toAPI_configJSONErrors(t *testing.T) {
	tests := []struct {
		name          string
		panel         panelModel
		errorSummary  string
		errorContains string
	}{
		{
			name: "rejects panel-level config_json for lens-dashboard-app (REQ-025 write path)",
			panel: panelModel{
				Type:       types.StringValue("lens-dashboard-app"),
				Grid:       panelGridModel{X: types.Int64Value(0), Y: types.Int64Value(0)},
				ConfigJSON: customtypes.NewJSONWithDefaultsValue(`{"x":1}`, populatePanelConfigJSONDefaults),
			},
			errorSummary:  "Unsupported panel type for config_json",
			errorContains: "Panel-level `config_json` is not supported for `lens-dashboard-app`",
		},
		{
			name: "rejects panel-level config_json for image",
			panel: panelModel{
				Type:       types.StringValue("image"),
				Grid:       panelGridModel{X: types.Int64Value(0), Y: types.Int64Value(0)},
				ConfigJSON: customtypes.NewJSONWithDefaultsValue(`{}`, populatePanelConfigJSONDefaults),
			},
			errorSummary:  "Unsupported panel type for config_json",
			errorContains: "not supported for `image`",
		},
		{
			name: "rejects panel-level config_json for slo_alerts",
			panel: panelModel{
				Type:       types.StringValue(panelTypeSloAlerts),
				Grid:       panelGridModel{X: types.Int64Value(0), Y: types.Int64Value(0)},
				ConfigJSON: customtypes.NewJSONWithDefaultsValue(`{}`, populatePanelConfigJSONDefaults),
			},
			errorSummary:  "Unsupported panel type for config_json",
			errorContains: "not supported for `slo_alerts`",
		},
		{
			name: "rejects missing slo_alerts_config",
			panel: panelModel{
				Type: types.StringValue(panelTypeSloAlerts),
				Grid: panelGridModel{X: types.Int64Value(0), Y: types.Int64Value(0)},
			},
			errorSummary:  "Missing SLO alerts panel configuration",
			errorContains: "require `slo_alerts_config`",
		},
		{
			name: "rejects missing panel configuration",
			panel: panelModel{
				Type:       types.StringValue("markdown"),
				Grid:       panelGridModel{X: types.Int64Value(0), Y: types.Int64Value(0)},
				ConfigJSON: customtypes.NewJSONWithDefaultsNull(populatePanelConfigJSONDefaults),
			},
			errorSummary:  "Unsupported panel configuration",
			errorContains: "No panel configuration block was provided",
		},
		{
			name: "rejects invalid markdown config_json",
			panel: panelModel{
				Type:       types.StringValue("markdown"),
				Grid:       panelGridModel{X: types.Int64Value(0), Y: types.Int64Value(0)},
				ConfigJSON: customtypes.NewJSONWithDefaultsValue(`{"content":`, populatePanelConfigJSONDefaults),
			},
			errorSummary:  "Failed to create markdown panel",
			errorContains: "unexpected end of JSON input",
		},
		{
			name: "rejects invalid vis config_json",
			panel: panelModel{
				Type:       types.StringValue("vis"),
				Grid:       panelGridModel{X: types.Int64Value(0), Y: types.Int64Value(0)},
				ConfigJSON: customtypes.NewJSONWithDefaultsValue(`{"attributes":`, populatePanelConfigJSONDefaults),
			},
			errorSummary:  "Failed to create visualization panel",
			errorContains: "unexpected end of JSON input",
		},
		{
			name: "rejects panel-level config_json for discover_session",
			panel: panelModel{
				Type:       types.StringValue(panelTypeDiscoverSession),
				Grid:       panelGridModel{X: types.Int64Value(0), Y: types.Int64Value(0)},
				ConfigJSON: customtypes.NewJSONWithDefaultsValue(`{}`, populatePanelConfigJSONDefaults),
			},
			errorSummary:  "Unsupported panel type for config_json",
			errorContains: "not supported for `discover_session`",
		},
		{
			name: "rejects missing discover_session_config",
			panel: panelModel{
				Type: types.StringValue(panelTypeDiscoverSession),
				Grid: panelGridModel{X: types.Int64Value(0), Y: types.Int64Value(0)},
			},
			errorSummary:  "Missing discover_session panel configuration",
			errorContains: "require `discover_session_config`",
		},
		{
			name: "rejects missing slo burn rate config",
			panel: panelModel{
				Type: types.StringValue("slo_burn_rate"),
				Grid: panelGridModel{X: types.Int64Value(0), Y: types.Int64Value(0)},
			},
			errorSummary:  "Missing SLO burn rate panel configuration",
			errorContains: "SLO burn rate panels require `slo_burn_rate_config`.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, diags := tt.panel.toAPI(context.Background(), nil)
			require.True(t, diags.HasError())
			require.Equal(t, tt.errorSummary, diags[0].Summary())
			require.Contains(t, diags[0].Detail(), tt.errorContains)
		})
	}
}

func Test_unknownPanelRoundTrip(t *testing.T) {
	apiJSON := `[
		{
			"grid": {"x": 0, "y": 1, "w": 48, "h": 15},
			"id": "unknown-1",
			"type": "custom_unknown_panel",
			"config": {
				"timeRange": {"from": "now-30d", "to": "now"},
				"columns": ["_source"],
				"sort": [{"@timestamp": "desc"}]
			}
		}
	]`

	// Step 1: unmarshal API JSON
	var apiPanels kbapi.DashboardPanels
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &apiPanels))

	// Step 2: read into TF model
	model := &dashboardModel{}
	panels, sections, diags := model.mapPanelsFromAPI(t.Context(), &apiPanels)
	require.False(t, diags.HasError())
	require.Empty(t, sections)
	require.Len(t, panels, 1)

	// Step 3: reconstruct model with just these panels and write back to API
	model.Panels = panels
	result, diags := model.panelsToAPI(context.Background())
	require.False(t, diags.HasError())

	// Step 4: verify round-trip equality (semantic JSON comparison)
	actualBytes, err := json.Marshal(result)
	require.NoError(t, err)

	var expectedAny, actualAny any
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &expectedAny))
	require.NoError(t, json.Unmarshal(actualBytes, &actualAny))

	assert.Equal(t, expectedAny, actualAny, "Round-trip should produce identical API JSON for unknown panel types")
}

func Test_unknownPanelToAPIErrorWithoutConfigJSON(t *testing.T) {
	// When an unknown panel type is authored in config (not read from the API)
	// and has no config_json, toAPI should produce an "unsupported panel type" error.
	panel := panelModel{
		Type:       types.StringValue("custom_unknown_panel"),
		Grid:       panelGridModel{X: types.Int64Value(0), Y: types.Int64Value(0), W: types.Int64Value(48), H: types.Int64Value(15)},
		ID:         types.StringNull(),
		ConfigJSON: customtypes.NewJSONWithDefaultsNull(populatePanelConfigJSONDefaults),
	}
	_, diags := panel.toAPI(context.Background(), nil)
	require.True(t, diags.HasError())
	require.Equal(t, "Unsupported panel type", diags[0].Summary())
	require.Contains(t, diags[0].Detail(), "not yet supported")
}
