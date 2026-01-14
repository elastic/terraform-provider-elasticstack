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
		name          string
		apiPanelsJSON string
		expected      []panelModel
	}{
		{
			name:          "empty panels",
			apiPanelsJSON: "[]",
			expected:      nil,
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
					"type": "visualization",
					"config": {
						"title": "My Panel",
						"content": "some content",
                        "hidePanelTitles": true
					}
				}
			]`,
			expected: []panelModel{
				{
					Type: types.StringValue("visualization"),
					Grid: panelGridModel{
						X: types.Int64Value(0),
						Y: types.Int64Value(1),
						W: types.Int64Value(2),
						H: types.Int64Value(3),
					},
					PanelID: types.StringValue("1"),
					EmbeddableConfig: &embeddableConfigModel{
						Title:           types.StringValue("My Panel"),
						Content:         types.StringValue("some content"),
						HidePanelTitles: types.BoolValue(true),
						Description:     types.StringNull(),
					},
					EmbeddableConfigJSON: jsontypes.NewNormalizedNull(),
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
			expected: []panelModel{
				{
					Type: types.StringValue("search"),
					Grid: panelGridModel{
						X: types.Int64Value(10),
						Y: types.Int64Value(20),
						W: types.Int64Null(),
						H: types.Int64Null(),
					},
					PanelID:              types.StringNull(),
					EmbeddableConfig:     nil,
					EmbeddableConfigJSON: jsontypes.NewNormalizedValue(`{"unknownField": "something"}`),
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
			result, diags := model.mapPanelsFromAPI(&apiPanels)
			require.False(t, diags.HasError())

			// Normalize JSON strings for comparison if needed, or rely on assert.Equal handling
			// Since we use jsontypes.Normalized which stores string, we might need to be careful with JSON formatting in test expectation.
			// In the 'unstructured' case, I used a compact JSON string.
			assert.Equal(t, tt.expected, result)
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
						PanelID: types.StringValue("1"),
						EmbeddableConfig: &embeddableConfigModel{
							Title:           types.StringValue("My Panel"),
							Content:         types.StringValue("some content"),
							HidePanelTitles: types.BoolValue(true),
						},
						EmbeddableConfigJSON: jsontypes.NewNormalizedNull(),
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
						PanelID:              types.StringNull(),
						EmbeddableConfig:     nil,
						EmbeddableConfigJSON: jsontypes.NewNormalizedValue(`{"unknownField":"something"}`),
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
