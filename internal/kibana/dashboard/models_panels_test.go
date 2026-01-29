package dashboard

import (
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
					"type": "DASHBOARD_MARKDOWN",
					"config": {
						"title": "My Panel",
						"content": "some content",
                        "hidePanelTitles": true
					}
				}
			]`,
			expectedPanels: []panelModel{
				{
					Type: types.StringValue("DASHBOARD_MARKDOWN"),
					Grid: panelGridModel{
						X: types.Int64Value(0),
						Y: types.Int64Value(1),
						W: types.Int64Value(2),
						H: types.Int64Value(3),
					},
					ID: types.StringValue("1"),
					MarkdownConfig: &markdownConfigModel{
						Title:           types.StringValue("My Panel"),
						Content:         types.StringValue("some content"),
						HidePanelTitles: types.BoolValue(true),
						Description:     types.StringNull(),
					},
					ConfigJSON: jsontypes.NewNormalizedValue(`{
						"title": "My Panel",
						"content": "some content",
                        "hidePanelTitles": true
					}`),
				},
			},
		},
		{
			name: "panel with unstructured config (JSON)",
			apiPanelsJSON: `[
				{
					"grid": {
						"x": 10,
						"y": 20
					},
					"type": "search",
					"config": {"unknownField": "something"}
				}
			]`,
			expectedPanels: []panelModel{
				{
					Type: types.StringValue("search"),
					Grid: panelGridModel{
						X: types.Int64Value(10),
						Y: types.Int64Value(20),
						W: types.Int64Null(),
						H: types.Int64Null(),
					},
					ID:             types.StringNull(),
					MarkdownConfig: nil,
					ConfigJSON:     jsontypes.NewNormalizedValue(`{"unknownField": "something"}`),
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
							"type": "visualization",
							"grid": { "x": 0, "y": 0, "w": 4, "h": 4 },
							"config": { "title": "Inner Panel" }
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
							Type: types.StringValue("visualization"),
							Grid: panelGridModel{
								X: types.Int64Value(0),
								Y: types.Int64Value(0),
								W: types.Int64Value(4),
								H: types.Int64Value(4),
							},
							ID:             types.StringNull(),
							MarkdownConfig: nil,
							ConfigJSON:     jsontypes.NewNormalizedValue(`{ "title": "Inner Panel" }`),
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
					"type": "visualization",
					"uid": "panel1",
					"config": { "title": "Panel 1" }
				},
				{
					"title": "Section 1",
					"grid": { "y": 100 },
					"uid": "section1",
					"panels": [
						{
							"type": "visualization",
							"grid": { "x": 0, "y": 0, "w": 6, "h": 6 },
							"config": { "title": "Inner Panel" }
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
					Type: types.StringValue("visualization"),
					Grid: panelGridModel{
						X: types.Int64Value(0),
						Y: types.Int64Value(0),
						W: types.Int64Value(6),
						H: types.Int64Value(6),
					},
					ID:             types.StringValue("panel1"),
					MarkdownConfig: nil,
					ConfigJSON:     jsontypes.NewNormalizedValue(`{ "title": "Panel 1" }`),
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
					ConfigJSON:     jsontypes.NewNormalizedValue(`{ "title": "Panel 2" }`),
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
							Type: types.StringValue("visualization"),
							Grid: panelGridModel{
								X: types.Int64Value(0),
								Y: types.Int64Value(0),
								W: types.Int64Value(6),
								H: types.Int64Value(6),
							},
							ID:             types.StringNull(),
							MarkdownConfig: nil,
							ConfigJSON:     jsontypes.NewNormalizedValue(`{ "title": "Inner Panel" }`),
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

			assert.Equal(t, tt.expectedPanels, panels)
			assert.Equal(t, tt.expectedSections, sections)
		})
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
						Type: types.StringValue("visualization"),
						Grid: panelGridModel{
							X: types.Int64Value(0),
							Y: types.Int64Value(1),
							W: types.Int64Value(2),
							H: types.Int64Value(3),
						},
						ID: types.StringValue("1"),
						MarkdownConfig: &markdownConfigModel{
							Title:           types.StringValue("My Panel"),
							Content:         types.StringValue("some content"),
							HidePanelTitles: types.BoolValue(true),
						},
						ConfigJSON: jsontypes.NewNormalizedNull(),
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
					"type": "visualization",
					"config": {
						"content": "some content",
                        "hidePanelTitles": true,
						"title": "My Panel"
					}
				}
			]`,
		},
		{
			name: "panel with unstructured config (JSON)",
			model: dashboardModel{
				Panels: []panelModel{
					{
						Type: types.StringValue("search"),
						Grid: panelGridModel{
							X: types.Int64Value(10),
							Y: types.Int64Value(20),
						},
						ID:             types.StringNull(),
						MarkdownConfig: nil,
						ConfigJSON:     jsontypes.NewNormalizedValue(`{"unknownField":"something"}`),
					},
				},
			},
			expected: `[
				{
					"grid": {
						"x": 10,
						"y": 20
					},
					"type": "search",
					"config": {
						"unknownField": "something"
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
								Type: types.StringValue("text"),
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
						{"grid":{"h":5,"w":5,"x":0,"y":0},"type":"text","config":{"content":"","title":"Inner Text"}}
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

			var expectedJSON interface{}
			var actualJSON interface{}

			require.NoError(t, json.Unmarshal([]byte(tt.expected), &expectedJSON))
			require.NoError(t, json.Unmarshal(jsonBytes, &actualJSON))

			assert.Equal(t, expectedJSON, actualJSON)
		})
	}
}
