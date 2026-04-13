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
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_newXYChartPanelConfigConverter(t *testing.T) {
	converter := newXYChartPanelConfigConverter()
	assert.NotNil(t, converter)
	assert.Equal(t, string(kbapi.XyChartNoESQLTypeXy), converter.visualizationType)
}

func Test_xyAxisModel_fromAPI_toAPI(t *testing.T) {
	raw := []byte(`{
		"x":{"grid":{"visible":true},"ticks":{"visible":false}},
		"y":{"grid":{"visible":true},"ticks":{"visible":true},"domain":{"type":"fit"}},
		"secondary_y":{"grid":{"visible":false},"ticks":{"visible":true},"domain":{"type":"fit"}}
	}`)
	var apiAxis kbapi.VisApiXyAxisConfig
	require.NoError(t, json.Unmarshal(raw, &apiAxis))

	model := &xyAxisModel{}
	diags := model.fromAPI(apiAxis)
	require.False(t, diags.HasError())

	require.NotNil(t, model.X)
	assert.Equal(t, types.BoolValue(true), model.X.Grid)
	assert.Equal(t, types.BoolValue(false), model.X.Ticks)

	require.NotNil(t, model.Y)
	assert.Equal(t, types.BoolValue(true), model.Y.Grid)
	assert.Equal(t, types.BoolValue(true), model.Y.Ticks)

	require.NotNil(t, model.SecondaryY)
	assert.Equal(t, types.BoolValue(false), model.SecondaryY.Grid)

	out, d := model.toAPI()
	require.False(t, d.HasError())
	model2 := &xyAxisModel{}
	diags = model2.fromAPI(out)
	require.False(t, diags.HasError())
	assert.Equal(t, model.X.Grid, model2.X.Grid)
	assert.Equal(t, model.Y.Grid, model2.Y.Grid)
}

func Test_xyAxisConfigModel_fromAPI_toAPI(t *testing.T) {
	tests := []struct {
		name     string
		apiAxis  *xyAxisConfigAPIModel
		expected *xyAxisConfigModel
	}{
		{
			name: "all fields populated",
			apiAxis: &xyAxisConfigAPIModel{
				Grid: &struct {
					Visible bool `json:"visible"`
				}{Visible: true},
				Ticks: &struct {
					Visible bool `json:"visible"`
				}{Visible: false},
				Labels: &struct {
					Orientation kbapi.VisApiOrientation `json:"orientation"`
				}{Orientation: kbapi.VisApiOrientation("horizontal")},
				Title: &struct {
					Text    *string `json:"text,omitempty"`
					Visible *bool   `json:"visible,omitempty"`
				}{
					Text:    new("X Axis Title"),
					Visible: new(true),
				},
			},
			expected: &xyAxisConfigModel{
				Grid:             types.BoolValue(true),
				Ticks:            types.BoolValue(false),
				LabelOrientation: types.StringValue("horizontal"),
				Title: &axisTitleModel{
					Value:   types.StringValue("X Axis Title"),
					Visible: types.BoolValue(true),
				},
			},
		},
		{
			name:     "nil axis",
			apiAxis:  nil,
			expected: nil,
		},
		{
			name: "only required fields",
			apiAxis: &xyAxisConfigAPIModel{
				Grid: &struct {
					Visible bool `json:"visible"`
				}{Visible: false},
				Ticks:  nil,
				Labels: nil,
				Title:  nil,
			},
			expected: &xyAxisConfigModel{
				Grid:             types.BoolValue(false),
				Ticks:            types.BoolNull(),
				LabelOrientation: types.StringNull(),
				Title:            nil,
			},
		},
		{
			name: "with all boolean fields",
			apiAxis: &xyAxisConfigAPIModel{
				Grid: &struct {
					Visible bool `json:"visible"`
				}{Visible: true},
				Ticks: &struct {
					Visible bool `json:"visible"`
				}{Visible: true},
			},
			expected: &xyAxisConfigModel{
				Grid:  types.BoolValue(true),
				Ticks: types.BoolValue(true),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test fromAPI
			model := &xyAxisConfigModel{}
			diags := model.fromAPI(tt.apiAxis)
			require.False(t, diags.HasError())

			if tt.expected == nil {
				return
			}

			assert.Equal(t, tt.expected.Grid, model.Grid)
			assert.Equal(t, tt.expected.Ticks, model.Ticks)
			assert.Equal(t, tt.expected.LabelOrientation, model.LabelOrientation)

			if tt.expected.Title != nil {
				require.NotNil(t, model.Title)
				assert.Equal(t, tt.expected.Title.Value, model.Title.Value)
				assert.Equal(t, tt.expected.Title.Visible, model.Title.Visible)
			}

			// Test toAPI round-trip
			apiAxis, diags := model.toAPI()
			require.False(t, diags.HasError())

			if tt.apiAxis != nil {
				assert.NotNil(t, apiAxis)
			}
		})
	}
}

func Test_yAxisConfigModel_fromAPIY_toAPIY(t *testing.T) {
	raw := []byte(`{
		"y":{
			"grid":{"visible":true},
			"ticks":{"visible":false},
			"labels":{"orientation":"vertical"},
			"scale":"linear",
			"title":{"text":"Y Axis Title","visible":false},
			"domain":{"type":"fit"}
		}
	}`)
	var env kbapi.VisApiXyAxisConfig
	require.NoError(t, json.Unmarshal(raw, &env))
	require.NotNil(t, env.Y)

	model := &yAxisConfigModel{}
	diags := model.fromAPIY(env.Y)
	require.False(t, diags.HasError())

	assert.Equal(t, types.BoolValue(true), model.Grid)
	assert.Equal(t, types.BoolValue(false), model.Ticks)
	assert.Equal(t, types.StringValue("vertical"), model.LabelOrientation)
	assert.Equal(t, types.StringValue("linear"), model.Scale)
	require.NotNil(t, model.Title)
	assert.Equal(t, types.StringValue("Y Axis Title"), model.Title.Value)
	assert.Equal(t, types.BoolValue(false), model.Title.Visible)

	apiY, diags := model.toAPIY()
	require.False(t, diags.HasError())
	require.NotNil(t, apiY)
}

func Test_yAxisConfigModel_fromAPISecondaryY_toAPISecondaryY(t *testing.T) {
	raw := []byte(`{
		"secondary_y":{
			"grid":{"visible":false},
			"ticks":{"visible":true},
			"labels":{"orientation":"angled"},
			"scale":"log",
			"title":{"text":"Right Y Axis","visible":true},
			"domain":{"type":"fit"}
		}
	}`)
	var env kbapi.VisApiXyAxisConfig
	require.NoError(t, json.Unmarshal(raw, &env))
	require.NotNil(t, env.SecondaryY)

	model := &yAxisConfigModel{}
	diags := model.fromAPISecondaryY(env.SecondaryY)
	require.False(t, diags.HasError())

	assert.Equal(t, types.BoolValue(false), model.Grid)
	assert.Equal(t, types.BoolValue(true), model.Ticks)
	assert.Equal(t, types.StringValue("angled"), model.LabelOrientation)
	assert.Equal(t, types.StringValue("log"), model.Scale)

	apiY, diags := model.toAPISecondaryY()
	require.False(t, diags.HasError())
	require.NotNil(t, apiY)
}

func Test_axisTitleModel_fromAPI_toAPI(t *testing.T) {
	tests := []struct {
		name     string
		apiTitle *struct {
			Text    *string `json:"text,omitempty"`
			Visible *bool   `json:"visible,omitempty"`
		}
		expected *axisTitleModel
	}{
		{
			name: "all fields populated",
			apiTitle: &struct {
				Text    *string `json:"text,omitempty"`
				Visible *bool   `json:"visible,omitempty"`
			}{
				Text:    new("Test Title"),
				Visible: new(true),
			},
			expected: &axisTitleModel{
				Value:   types.StringValue("Test Title"),
				Visible: types.BoolValue(true),
			},
		},
		{
			name:     "nil title",
			apiTitle: nil,
			expected: &axisTitleModel{},
		},
		{
			name: "only value",
			apiTitle: &struct {
				Text    *string `json:"text,omitempty"`
				Visible *bool   `json:"visible,omitempty"`
			}{
				Text:    new("Only Value"),
				Visible: nil,
			},
			expected: &axisTitleModel{
				Value:   types.StringValue("Only Value"),
				Visible: types.BoolNull(),
			},
		},
		{
			name: "only visible",
			apiTitle: &struct {
				Text    *string `json:"text,omitempty"`
				Visible *bool   `json:"visible,omitempty"`
			}{
				Text:    nil,
				Visible: new(false),
			},
			expected: &axisTitleModel{
				Value:   types.StringNull(),
				Visible: types.BoolValue(false),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test fromAPI
			model := &axisTitleModel{}
			model.fromAPI(tt.apiTitle)
			assert.Equal(t, tt.expected.Value, model.Value)
			assert.Equal(t, tt.expected.Visible, model.Visible)

			// Test toAPI
			apiTitle := model.toAPI()
			assert.NotNil(t, apiTitle)
		})
	}
}

func Test_xyDecorationsModel_readFromStyling_writeToStyling(t *testing.T) {
	interp := kbapi.Linear
	pts := kbapi.XyStylingPointsVisibilityVisible
	styling := kbapi.XyStyling{
		Areas: kbapi.XyStylingAreas{FillOpacity: new(float32(0.5))},
		Bars: kbapi.XyStylingBars{
			MinimumHeight: new(float32(5)),
			DataLabels: func() *struct {
				Visible *bool `json:"visible,omitempty"`
			} {
				v := true
				return &struct {
					Visible *bool `json:"visible,omitempty"`
				}{Visible: &v}
			}(),
		},
		Fitting:       kbapi.XyFitting{Type: kbapi.XyFittingTypeNone},
		Interpolation: &interp,
		Overlays: kbapi.XyStylingOverlays{
			PartialBuckets: func() *struct {
				Visible *bool `json:"visible,omitempty"`
			} {
				v := true
				return &struct {
					Visible *bool `json:"visible,omitempty"`
				}{Visible: &v}
			}(),
			CurrentTimeMarker: func() *struct {
				Visible *bool `json:"visible,omitempty"`
			} {
				v := false
				return &struct {
					Visible *bool `json:"visible,omitempty"`
				}{Visible: &v}
			}(),
		},
		Points: kbapi.XyStylingPoints{Visibility: &pts},
	}

	model := &xyDecorationsModel{}
	model.readFromStyling(styling)

	assert.Equal(t, types.BoolValue(true), model.ShowEndZones)
	assert.Equal(t, types.BoolValue(false), model.ShowCurrentTimeMarker)
	assert.Equal(t, types.StringValue("always"), model.PointVisibility)
	assert.Equal(t, types.StringValue("linear"), model.LineInterpolation)
	assert.Equal(t, types.Int64Value(5), model.MinimumBarHeight)
	assert.Equal(t, types.BoolValue(true), model.ShowValueLabels)
	assert.InDelta(t, 0.5, model.FillOpacity.ValueFloat64(), 0.001)

	var out kbapi.XyStyling
	out.Fitting = kbapi.XyFitting{Type: kbapi.XyFittingTypeNone}
	model.writeToStyling(&out)
	model2 := &xyDecorationsModel{}
	model2.readFromStyling(out)
	assert.Equal(t, model.ShowEndZones, model2.ShowEndZones)
	assert.Equal(t, model.PointVisibility, model2.PointVisibility)
}

func Test_xyFittingModel_fromAPI_toAPI(t *testing.T) {
	tests := []struct {
		name        string
		apiFitting  kbapi.XyFitting
		expected    *xyFittingModel
		expectError bool
	}{
		{
			name: "all fields populated",
			apiFitting: kbapi.XyFitting{
				Type:      kbapi.XyFittingType("linear"),
				Emphasize: new(true),
				Extend:    func() *kbapi.XyFittingExtend { e := kbapi.XyFittingExtendZero; return &e }(),
			},
			expected: &xyFittingModel{
				Type:     types.StringValue("linear"),
				Dotted:   types.BoolValue(true),
				EndValue: types.StringValue("zero"),
			},
		},
		{
			name: "only required field",
			apiFitting: kbapi.XyFitting{
				Type:      kbapi.XyFittingType("none"),
				Emphasize: nil,
				Extend:    nil,
			},
			expected: &xyFittingModel{
				Type:     types.StringValue("none"),
				Dotted:   types.BoolNull(),
				EndValue: types.StringNull(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test fromAPI
			model := &xyFittingModel{}
			model.fromAPI(tt.apiFitting)

			assert.Equal(t, tt.expected.Type, model.Type)
			assert.Equal(t, tt.expected.Dotted, model.Dotted)
			assert.Equal(t, tt.expected.EndValue, model.EndValue)

			// Test toAPI
			apiFitting := model.toAPI()
			assert.NotNil(t, apiFitting)

			// Verify type is preserved
			assert.Equal(t, string(tt.apiFitting.Type), string(apiFitting.Type))
		})
	}
}

func Test_xyLegendModel_fromAPI_toAPI_Inside(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		apiLegend kbapi.XyLegend
		expected  *xyLegendModel
	}{
		{
			name: "inside legend with all fields",
			apiLegend: func() kbapi.XyLegend {
				visibility := kbapi.XyLegendInsideVisibilityVisible
				position := kbapi.TopLeft
				legend := kbapi.XyLegendInside{
					Placement:  kbapi.XyLegendInsidePlacementInside,
					Visibility: &visibility,
					Layout: &struct {
						Truncate *struct {
							Enabled  *bool    `json:"enabled,omitempty"`
							MaxLines *float32 `json:"max_lines,omitempty"`
						} `json:"truncate,omitempty"`
						Type kbapi.XyLegendInsideLayoutType `json:"type"`
					}{
						Truncate: &struct {
							Enabled  *bool    `json:"enabled,omitempty"`
							MaxLines *float32 `json:"max_lines,omitempty"`
						}{
							MaxLines: new(float32(3)),
						},
						Type: kbapi.XyLegendInsideLayoutTypeGrid,
					},
					Columns:  new(float32(2)),
					Position: &position,
					Statistics: &[]kbapi.XyLegendInsideStatistics{
						kbapi.XyLegendInsideStatistics("mean"),
						kbapi.XyLegendInsideStatistics("max"),
					},
				}
				var result kbapi.XyLegend
				_ = result.FromXyLegendInside(legend)
				return result
			}(),
			expected: &xyLegendModel{
				Inside:             types.BoolValue(true),
				Visibility:         types.StringValue("visible"),
				TruncateAfterLines: types.Int64Value(3),
				Columns:            types.Int64Value(2),
				Alignment:          types.StringValue("top_left"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test fromAPI
			model := &xyLegendModel{}
			diags := model.fromAPI(ctx, tt.apiLegend)
			require.False(t, diags.HasError())

			assert.Equal(t, tt.expected.Inside, model.Inside)
			assert.Equal(t, tt.expected.Visibility, model.Visibility)
			assert.Equal(t, tt.expected.TruncateAfterLines, model.TruncateAfterLines)
			assert.Equal(t, tt.expected.Columns, model.Columns)
			assert.Equal(t, tt.expected.Alignment, model.Alignment)

			// Test toAPI
			apiLegend, diags := model.toAPI()
			require.False(t, diags.HasError())
			assert.NotNil(t, apiLegend)
		})
	}
}

func Test_xyLegendModel_fromAPI_toAPI_Outside(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		apiLegend kbapi.XyLegend
		expected  *xyLegendModel
	}{
		{
			name: "outside legend with all fields",
			apiLegend: func() kbapi.XyLegend {
				visibility := kbapi.XyLegendOutsideVerticalVisibility("hidden")
				position := kbapi.Right
				placement := kbapi.XyLegendOutsideVerticalPlacementOutside
				legend := kbapi.XyLegendOutsideVertical{
					Visibility: &visibility,
					Layout: &struct {
						Truncate *struct {
							Enabled  *bool    `json:"enabled,omitempty"`
							MaxLines *float32 `json:"max_lines,omitempty"`
						} `json:"truncate,omitempty"`
						Type kbapi.XyLegendOutsideVerticalLayoutType `json:"type"`
					}{
						Truncate: &struct {
							Enabled  *bool    `json:"enabled,omitempty"`
							MaxLines *float32 `json:"max_lines,omitempty"`
						}{
							MaxLines: new(float32(5)),
						},
						Type: kbapi.Grid,
					},
					Placement: &placement,
					Position:  &position,
					Size:      kbapi.LegendSizeM,
					Statistics: &[]kbapi.XyLegendOutsideVerticalStatistics{
						kbapi.XyLegendOutsideVerticalStatistics("min"),
					},
				}
				var result kbapi.XyLegend
				_ = result.FromXyLegendOutsideVertical(legend)
				return result
			}(),
			expected: &xyLegendModel{
				Inside:             types.BoolValue(false),
				Visibility:         types.StringValue("hidden"),
				TruncateAfterLines: types.Int64Value(5),
				Position:           types.StringValue("right"),
				Size:               types.StringValue("m"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test fromAPI
			model := &xyLegendModel{}
			diags := model.fromAPI(ctx, tt.apiLegend)
			require.False(t, diags.HasError())

			assert.Equal(t, tt.expected.Inside, model.Inside)
			assert.Equal(t, tt.expected.Visibility, model.Visibility)
			assert.Equal(t, tt.expected.TruncateAfterLines, model.TruncateAfterLines)
			assert.Equal(t, tt.expected.Position, model.Position)
			assert.Equal(t, tt.expected.Size, model.Size)

			// Test toAPI
			apiLegend, diags := model.toAPI()
			require.False(t, diags.HasError())
			assert.NotNil(t, apiLegend)
		})
	}
}

func Test_xyChartPanelConfigConverter_populateFromAttributes_buildAttributes_roundTrip(t *testing.T) {
	ctx := context.Background()

	model := &xyChartConfigModel{
		Title:       types.StringValue("XY Chart Round-Trip"),
		Description: types.StringValue("Converter test"),
		Axis: &xyAxisModel{
			X: &xyAxisConfigModel{},
			Y: &yAxisConfigModel{},
		},
		Decorations: &xyDecorationsModel{},
		Fitting:     &xyFittingModel{Type: types.StringValue("none")},
		Layers: []xyLayerModel{
			{
				Type: types.StringValue("area"),
				DataLayer: &dataLayerModel{
					DataSourceJSON: jsontypes.NewNormalizedValue(`{"type":"dataView","id":"logs-*"}`),
					Y: []yMetricModel{
						{ConfigJSON: jsontypes.NewNormalizedValue(`{"operation":"count","color":"#68BC00","axis":"left"}`)},
					},
				},
			},
		},
		Legend: &xyLegendModel{
			Inside:     types.BoolValue(false),
			Visibility: types.StringValue("visible"),
		},
		Query: &filterSimpleModel{
			Expression: types.StringValue("*"),
			Language:   types.StringValue("kql"),
		},
	}

	xyChart, diags := model.toAPINoESQL()
	require.False(t, diags.HasError())

	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	require.NoError(t, attrs.FromXyChartNoESQL(xyChart))

	converter := newXYChartPanelConfigConverter()
	pm := &panelModel{}
	diags = converter.populateFromAttributes(ctx, pm, attrs)
	require.False(t, diags.HasError())
	require.NotNil(t, pm.XYChartConfig)

	attrs2, diags := converter.buildAttributes(*pm)
	require.False(t, diags.HasError())

	chart2, err := attrs2.AsXyChartNoESQL()
	require.NoError(t, err)
	assert.Equal(t, kbapi.XyChartNoESQLTypeXy, chart2.Type)
	assert.Equal(t, "XY Chart Round-Trip", *chart2.Title)
	assert.Len(t, chart2.Layers, 1)
}

func Test_xyChartConfigModel_toAPI_fromAPI(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		model       *xyChartConfigModel
		expectError bool
	}{
		{
			name: "complete config",
			model: &xyChartConfigModel{
				Title:       types.StringValue("Test XY Chart"),
				Description: types.StringValue("A test chart"),
				Axis: &xyAxisModel{
					X: &xyAxisConfigModel{
						Grid:  types.BoolValue(true),
						Ticks: types.BoolValue(false),
					},
					Y: &yAxisConfigModel{},
				},
				Decorations: &xyDecorationsModel{
					ShowEndZones:    types.BoolValue(true),
					PointVisibility: types.StringValue("never"),
				},
				Fitting: &xyFittingModel{
					Type:   types.StringValue("linear"),
					Dotted: types.BoolValue(true),
				},
				Layers: []xyLayerModel{
					{
						Type: types.StringValue("area"),
						DataLayer: &dataLayerModel{
							DataSourceJSON: jsontypes.NewNormalizedValue(`{"type":"dataView","id":"logs-*"}`),
							Y: []yMetricModel{
								{
									ConfigJSON: jsontypes.NewNormalizedValue(`{"operation":"count","color":"#68BC00","axis":"left"}`),
								},
							},
						},
					},
				},
				Legend: &xyLegendModel{
					Inside:     types.BoolValue(false),
					Visibility: types.StringValue("visible"),
				},
				Query: &filterSimpleModel{
					Expression: types.StringValue("*"),
					Language:   types.StringValue("kql"),
				},
			},
			expectError: false,
		},
		{
			name: "minimal config",
			model: &xyChartConfigModel{
				Title: types.StringValue("Minimal Chart"),
				Axis: &xyAxisModel{
					X: &xyAxisConfigModel{},
					Y: &yAxisConfigModel{},
				},
				Decorations: &xyDecorationsModel{},
				Fitting:     &xyFittingModel{Type: types.StringValue("none")},
				Legend: &xyLegendModel{
					Inside:     types.BoolValue(false),
					Visibility: types.StringValue("visible"),
				},
				Query: &filterSimpleModel{
					Expression: types.StringValue("*"),
					Language:   types.StringValue("kql"),
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiChart, diags := tt.model.toAPINoESQL()
			if tt.expectError {
				assert.True(t, diags.HasError())
				return
			}

			require.False(t, diags.HasError())
			assert.Equal(t, kbapi.XyChartNoESQLTypeXy, apiChart.Type)

			if tt.model.Title.ValueString() != "" {
				assert.Equal(t, tt.model.Title.ValueString(), *apiChart.Title)
			}

			model2 := &xyChartConfigModel{}
			diags = model2.fromAPINoESQL(ctx, apiChart)
			require.False(t, diags.HasError())

			assert.Equal(t, tt.model.Title, model2.Title)
			assert.Equal(t, tt.model.Description, model2.Description)
		})
	}
}

func Test_xyAxisConfigModel_toAPI_nil(t *testing.T) {
	var model *xyAxisConfigModel
	apiAxis, diags := model.toAPI()
	assert.False(t, diags.HasError())
	assert.Nil(t, apiAxis)
}

func Test_yAxisConfigModel_toAPIY_nil(t *testing.T) {
	var model *yAxisConfigModel
	apiAxis, diags := model.toAPIY()
	assert.False(t, diags.HasError())
	assert.Nil(t, apiAxis)
}

func Test_yAxisConfigModel_toAPISecondaryY_nil(t *testing.T) {
	var model *yAxisConfigModel
	apiAxis, diags := model.toAPISecondaryY()
	assert.False(t, diags.HasError())
	assert.Nil(t, apiAxis)
}

func Test_axisTitleModel_toAPI_nil(t *testing.T) {
	var model *axisTitleModel
	apiTitle := model.toAPI()
	assert.Nil(t, apiTitle)
}

func Test_xyDecorationsModel_writeToStyling_nil(t *testing.T) {
	var model *xyDecorationsModel
	var s kbapi.XyStyling
	model.writeToStyling(&s)
	assert.Nil(t, s.Overlays.PartialBuckets)
}

func Test_xyFittingModel_toAPI_nil(t *testing.T) {
	var model *xyFittingModel
	apiFitting := model.toAPI()
	assert.NotNil(t, apiFitting) // Returns empty struct, not nil
}

func Test_xyLegendModel_toAPI_nil(t *testing.T) {
	var model *xyLegendModel
	apiLegend, diags := model.toAPI()
	assert.False(t, diags.HasError())
	// Check it doesn't panic
	assert.NotNil(t, apiLegend)
}

func Test_alignXYChartStateFromPlanPanels_preservesPractitionerIntent(t *testing.T) {
	planPanels := []panelModel{
		{
			XYChartConfig: &xyChartConfigModel{
				Title: types.StringValue("Sample XY Chart"),
				Axis: &xyAxisModel{
					X: &xyAxisConfigModel{
						Title: &axisTitleModel{
							Value:   types.StringValue("Timestamp"),
							Visible: types.BoolValue(true),
						},
						Grid:             types.BoolNull(),
						Ticks:            types.BoolNull(),
						LabelOrientation: types.StringNull(),
						Scale:            types.StringNull(),
						DomainJSON:       jsontypes.NewNormalizedNull(),
					},
					Y: &yAxisConfigModel{
						Title: &axisTitleModel{
							Value:   types.StringValue("Count"),
							Visible: types.BoolValue(true),
						},
						Grid:             types.BoolNull(),
						Ticks:            types.BoolNull(),
						LabelOrientation: types.StringNull(),
						Scale:            types.StringValue("linear"),
						DomainJSON:       jsontypes.NewNormalizedValue(`{"type":"fit"}`),
					},
					SecondaryY: &yAxisConfigModel{
						Title: &axisTitleModel{
							Value:   types.StringValue("Rate"),
							Visible: types.BoolValue(true),
						},
						Grid:             types.BoolValue(false),
						Ticks:            types.BoolValue(false),
						LabelOrientation: types.StringValue("vertical"),
						Scale:            types.StringValue("sqrt"),
						DomainJSON:       jsontypes.NewNormalizedValue(`{"type":"fit"}`),
					},
				},
				Decorations: &xyDecorationsModel{
					ShowEndZones:          types.BoolNull(),
					ShowCurrentTimeMarker: types.BoolNull(),
					PointVisibility:       types.StringNull(),
					LineInterpolation:     types.StringNull(),
					FillOpacity:           types.Float64Value(0.3),
				},
				Legend: &xyLegendModel{
					Visibility:         types.StringValue("visible"),
					Inside:             types.BoolValue(false),
					Position:           types.StringValue("right"),
					TruncateAfterLines: types.Int64Null(),
				},
				Layers: []xyLayerModel{
					{
						Type: types.StringValue("line"),
						DataLayer: &dataLayerModel{
							DataSourceJSON: jsontypes.NewNormalizedValue(`{"type":"data_view_spec","index_pattern":"metrics-*"}`),
							XJSON:          jsontypes.NewNormalizedValue(`{"column":"@timestamp","format":{"type":"number"}}`),
							BreakdownByJSON: jsontypes.NewNormalizedValue(
								`{"column":"host.name","collapse_by":"avg","format":{"type":"number"},` +
									`"color":{"mode":"categorical","palette":"default","mapping":[],` +
									`"unassigned":{"type":"color_code","value":"#D3DAE6"}}}`,
							),
							Y: []yMetricModel{
								{ConfigJSON: jsontypes.NewNormalizedValue(`{"operation":"count","empty_as_null":true,"format":{"type":"number"}}`)},
							},
						},
					},
					{
						Type: types.StringValue("reference_lines"),
						ReferenceLineLayer: &referenceLineLayerModel{
							DataSourceJSON: jsontypes.NewNormalizedValue(`{"type":"data_view_spec","index_pattern":"metrics-*"}`),
							Thresholds: []thresholdModel{
								{
									ValueJSON: jsontypes.NewNormalizedValue(`{"operation":"static_value","value":42,"label":"","format":{"type":"number"}}`),
								},
							},
						},
					},
				},
			},
		},
	}

	statePanels := []panelModel{
		{
			XYChartConfig: &xyChartConfigModel{
				Title: types.StringValue(""),
				Axis: &xyAxisModel{
					X: &xyAxisConfigModel{
						Title:            &axisTitleModel{},
						Grid:             types.BoolValue(true),
						Ticks:            types.BoolValue(true),
						LabelOrientation: types.StringValue("horizontal"),
						Scale:            types.StringValue("ordinal"),
						DomainJSON:       jsontypes.NewNormalizedValue(`{"type":"fit","rounding":false}`),
					},
					Y: &yAxisConfigModel{
						Title: &axisTitleModel{
							Value:   types.StringValue("Count"),
							Visible: types.BoolValue(true),
						},
						Grid:             types.BoolValue(true),
						Ticks:            types.BoolValue(true),
						LabelOrientation: types.StringValue("horizontal"),
						Scale:            types.StringValue("linear"),
						DomainJSON:       jsontypes.NewNormalizedValue(`{"type":"fit","rounding":true}`),
					},
					SecondaryY: nil,
				},
				Decorations: &xyDecorationsModel{
					ShowEndZones:          types.BoolValue(false),
					ShowCurrentTimeMarker: types.BoolValue(false),
					PointVisibility:       types.StringValue("auto"),
					LineInterpolation:     types.StringValue("linear"),
					FillOpacity:           types.Float64Null(),
				},
				Legend: &xyLegendModel{
					Visibility:         types.StringValue("visible"),
					Inside:             types.BoolValue(false),
					Position:           types.StringValue("right"),
					TruncateAfterLines: types.Int64Value(1),
				},
				Layers: []xyLayerModel{
					{
						Type: types.StringValue("line"),
						DataLayer: &dataLayerModel{
							DataSourceJSON: jsontypes.NewNormalizedValue(`{"type":"data_view_spec","index_pattern":"metrics-*","time_field":"@timestamp"}`),
							XJSON:          jsontypes.NewNormalizedValue(`{"column":"@timestamp","format":{"type":"number","decimals":2,"compact":false}}`),
							BreakdownByJSON: jsontypes.NewNormalizedValue(
								`{"column":"host.name","collapse_by":"avg",` +
									`"format":{"type":"number","decimals":2,"compact":false},` +
									`"color":{"mode":"categorical","palette":"default","mapping":[],` +
									`"unassigned":{"type":"color_code","value":"#D3DAE6"}}}`,
							),
							Y: []yMetricModel{
								{ConfigJSON: jsontypes.NewNormalizedValue(`{"operation":"count","empty_as_null":true,"format":{"type":"number","decimals":2,"compact":false},"axis_id":"y"}`)},
							},
						},
					},
					{
						Type: types.StringValue("reference_lines"),
						ReferenceLineLayer: &referenceLineLayerModel{
							DataSourceJSON: jsontypes.NewNormalizedValue(`{"type":"data_view_spec","index_pattern":"metrics-*","time_field":"@timestamp"}`),
							Thresholds: []thresholdModel{
								{
									ValueJSON: jsontypes.NewNormalizedValue(`{"operation":"static_value","value":42,"label":"","format":{"type":"number","decimals":2,"compact":false},"axis_id":"y"}`),
								},
							},
						},
					},
				},
			},
		},
	}

	alignXYChartStateFromPlanPanels(planPanels, statePanels)

	got := statePanels[0].XYChartConfig
	require.NotNil(t, got)
	assert.Equal(t, types.StringValue("Sample XY Chart"), got.Title)
	assert.True(t, got.Axis.X.Scale.IsNull())
	assert.True(t, got.Axis.X.Grid.IsNull())
	assert.True(t, got.Axis.X.Ticks.IsNull())
	assert.True(t, got.Axis.X.LabelOrientation.IsNull())
	assert.True(t, got.Axis.X.DomainJSON.IsNull())
	require.NotNil(t, got.Axis.SecondaryY)
	assert.Equal(t, planPanels[0].XYChartConfig.Axis.SecondaryY.DomainJSON.ValueString(), got.Axis.SecondaryY.DomainJSON.ValueString())
	assert.Equal(t, planPanels[0].XYChartConfig.Axis.Y.DomainJSON.ValueString(), got.Axis.Y.DomainJSON.ValueString())
	assert.True(t, got.Decorations.ShowEndZones.IsNull())
	assert.True(t, got.Decorations.ShowCurrentTimeMarker.IsNull())
	assert.True(t, got.Decorations.PointVisibility.IsNull())
	assert.True(t, got.Decorations.LineInterpolation.IsNull())
	assert.Equal(t, types.Float64Value(0.3), got.Decorations.FillOpacity)
	assert.True(t, got.Legend.TruncateAfterLines.IsNull())
	assert.Equal(t, planPanels[0].XYChartConfig.Layers[0].DataLayer.DataSourceJSON.ValueString(), got.Layers[0].DataLayer.DataSourceJSON.ValueString())
	assert.Equal(t, planPanels[0].XYChartConfig.Layers[0].DataLayer.XJSON.ValueString(), got.Layers[0].DataLayer.XJSON.ValueString())
	assert.Equal(t, planPanels[0].XYChartConfig.Layers[0].DataLayer.BreakdownByJSON.ValueString(), got.Layers[0].DataLayer.BreakdownByJSON.ValueString())
	assert.Equal(t, planPanels[0].XYChartConfig.Layers[0].DataLayer.Y[0].ConfigJSON.ValueString(), got.Layers[0].DataLayer.Y[0].ConfigJSON.ValueString())
	assert.Equal(t, planPanels[0].XYChartConfig.Layers[1].ReferenceLineLayer.DataSourceJSON.ValueString(), got.Layers[1].ReferenceLineLayer.DataSourceJSON.ValueString())
	assert.Equal(t, planPanels[0].XYChartConfig.Layers[1].ReferenceLineLayer.Thresholds[0].ValueJSON.ValueString(), got.Layers[1].ReferenceLineLayer.Thresholds[0].ValueJSON.ValueString())
}

func Test_filterSimpleModel_toAPI_nil(t *testing.T) {
	var model *filterSimpleModel
	apiQuery := model.toAPI()
	assert.NotNil(t, apiQuery) // Returns empty struct, not nil
}

func Test_xyAxisModel_toAPI_nil(t *testing.T) {
	var model *xyAxisModel
	apiAxis, diags := model.toAPI()
	assert.False(t, diags.HasError())
	assert.NotNil(t, apiAxis) // Returns empty struct, not nil
}

func Test_axisTitleIsDefault(t *testing.T) {
	tests := []struct {
		name   string
		title  *axisTitleModel
		expect bool
	}{
		{"nil", nil, true},
		{"value known", &axisTitleModel{Value: types.StringValue("x")}, false},
		{"visible true", &axisTitleModel{Visible: types.BoolValue(true)}, true},
		{"visible false", &axisTitleModel{Visible: types.BoolValue(false)}, false},
		{"empty", &axisTitleModel{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := axisTitleIsDefault(tt.title)
			assert.Equal(t, tt.expect, got)
		})
	}
}
