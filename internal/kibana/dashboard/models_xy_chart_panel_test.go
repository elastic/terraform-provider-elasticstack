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
	assert.Equal(t, string(kbapi.Xy), converter.visualizationType)
}

func Test_xyAxisModel_fromAPI_toAPI(t *testing.T) {
	tests := []struct {
		name     string
		apiAxis  kbapi.XyAxis
		expected *xyAxisModel
	}{
		{
			name: "all axes populated",
			apiAxis: kbapi.XyAxis{
				X: &xyAxisConfigAPIModel{
					Grid: &struct {
						Visible bool `json:"visible"`
					}{Visible: true},
					Ticks: &struct {
						Visible bool `json:"visible"`
					}{Visible: false},
				},
				Left: &leftYAxisConfigAPIModel{
					Grid: &struct {
						Visible bool `json:"visible"`
					}{Visible: true},
					Ticks: &struct {
						Visible bool `json:"visible"`
					}{Visible: true},
				},
				Right: &rightYAxisConfigAPIModel{
					Grid: &struct {
						Visible bool `json:"visible"`
					}{Visible: false},
					Ticks: &struct {
						Visible bool `json:"visible"`
					}{Visible: true},
				},
			},
			expected: &xyAxisModel{
				X: &xyAxisConfigModel{
					Grid:  types.BoolValue(true),
					Ticks: types.BoolValue(false),
				},
				Left: &yAxisConfigModel{
					Grid:  types.BoolValue(true),
					Ticks: types.BoolValue(true),
				},
				Right: &yAxisConfigModel{
					Grid:  types.BoolValue(false),
					Ticks: types.BoolValue(true),
				},
			},
		},
		{
			name: "nil axes",
			apiAxis: kbapi.XyAxis{
				X:     nil,
				Left:  nil,
				Right: nil,
			},
			expected: &xyAxisModel{
				X:     nil,
				Left:  nil,
				Right: nil,
			},
		},
		{
			name: "only x axis",
			apiAxis: kbapi.XyAxis{
				X: &xyAxisConfigAPIModel{
					Grid: &struct {
						Visible bool `json:"visible"`
					}{Visible: true},
				},
				Left:  nil,
				Right: nil,
			},
			expected: &xyAxisModel{
				X: &xyAxisConfigModel{
					Grid: types.BoolValue(true),
				},
				Left:  nil,
				Right: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test fromAPI
			model := &xyAxisModel{}
			diags := model.fromAPI(tt.apiAxis)
			require.False(t, diags.HasError())

			if tt.expected.X != nil {
				require.NotNil(t, model.X)
				assert.Equal(t, tt.expected.X.Grid, model.X.Grid)
				assert.Equal(t, tt.expected.X.Ticks, model.X.Ticks)
			} else {
				assert.Nil(t, model.X)
			}

			if tt.expected.Left != nil {
				require.NotNil(t, model.Left)
				assert.Equal(t, tt.expected.Left.Grid, model.Left.Grid)
				assert.Equal(t, tt.expected.Left.Ticks, model.Left.Ticks)
			} else {
				assert.Nil(t, model.Left)
			}

			if tt.expected.Right != nil {
				require.NotNil(t, model.Right)
				assert.Equal(t, tt.expected.Right.Grid, model.Right.Grid)
				assert.Equal(t, tt.expected.Right.Ticks, model.Right.Ticks)
			} else {
				assert.Nil(t, model.Right)
			}

			// Test toAPI round-trip
			apiAxis, diags := model.toAPI()
			require.False(t, diags.HasError())

			// Verify round-trip
			model2 := &xyAxisModel{}
			diags = model2.fromAPI(apiAxis)
			require.False(t, diags.HasError())

			if tt.expected.X != nil {
				assert.Equal(t, model.X.Grid, model2.X.Grid)
				assert.Equal(t, model.X.Ticks, model2.X.Ticks)
			}
		})
	}
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

func Test_yAxisConfigModel_fromAPILeft_toAPILeft(t *testing.T) {
	tests := []struct {
		name     string
		apiAxis  *leftYAxisConfigAPIModel
		expected *yAxisConfigModel
	}{
		{
			name: "all fields populated",
			apiAxis: &leftYAxisConfigAPIModel{
				Grid: &struct {
					Visible bool `json:"visible"`
				}{Visible: true},
				Ticks: &struct {
					Visible bool `json:"visible"`
				}{Visible: false},
				Labels: &struct {
					Orientation kbapi.VisApiOrientation `json:"orientation"`
				}{Orientation: kbapi.VisApiOrientation("vertical")},
				Scale: func() *kbapi.XyAxisLeftScale { s := kbapi.XyAxisLeftScale("linear"); return &s }(),
				Title: &struct {
					Text    *string `json:"text,omitempty"`
					Visible *bool   `json:"visible,omitempty"`
				}{
					Text:    new("Y Axis Title"),
					Visible: new(false),
				},
			},
			expected: &yAxisConfigModel{
				Grid:             types.BoolValue(true),
				Ticks:            types.BoolValue(false),
				LabelOrientation: types.StringValue("vertical"),
				Scale:            types.StringValue("linear"),
				Title: &axisTitleModel{
					Value:   types.StringValue("Y Axis Title"),
					Visible: types.BoolValue(false),
				},
			},
		},
		{
			name:     "nil axis",
			apiAxis:  nil,
			expected: nil,
		},
		{
			name: "with scale field",
			apiAxis: &leftYAxisConfigAPIModel{
				Grid: &struct {
					Visible bool `json:"visible"`
				}{Visible: false},
				Scale: func() *kbapi.XyAxisLeftScale { s := kbapi.XyAxisLeftScale("linear"); return &s }(),
			},
			expected: &yAxisConfigModel{
				Grid:  types.BoolValue(false),
				Scale: types.StringValue("linear"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test fromAPILeft
			model := &yAxisConfigModel{}
			diags := model.fromAPILeft(tt.apiAxis)
			require.False(t, diags.HasError())

			if tt.expected == nil {
				return
			}

			assert.Equal(t, tt.expected.Grid, model.Grid)
			assert.Equal(t, tt.expected.Ticks, model.Ticks)
			assert.Equal(t, tt.expected.LabelOrientation, model.LabelOrientation)
			assert.Equal(t, tt.expected.Scale, model.Scale)

			// Test toAPILeft round-trip
			apiAxis, diags := model.toAPILeft()
			require.False(t, diags.HasError())

			if tt.apiAxis != nil {
				assert.NotNil(t, apiAxis)
			}
		})
	}
}

func Test_yAxisConfigModel_fromAPIRight_toAPIRight(t *testing.T) {
	tests := []struct {
		name     string
		apiAxis  *rightYAxisConfigAPIModel
		expected *yAxisConfigModel
	}{
		{
			name: "all fields populated",
			apiAxis: &rightYAxisConfigAPIModel{
				Grid: &struct {
					Visible bool `json:"visible"`
				}{Visible: false},
				Ticks: &struct {
					Visible bool `json:"visible"`
				}{Visible: true},
				Labels: &struct {
					Orientation kbapi.VisApiOrientation `json:"orientation"`
				}{Orientation: kbapi.VisApiOrientation("angled")},
				Scale: func() *kbapi.XyAxisRightScale { s := kbapi.XyAxisRightScale("log"); return &s }(),
				Title: &struct {
					Text    *string `json:"text,omitempty"`
					Visible *bool   `json:"visible,omitempty"`
				}{
					Text:    new("Right Y Axis"),
					Visible: new(true),
				},
			},
			expected: &yAxisConfigModel{
				Grid:             types.BoolValue(false),
				Ticks:            types.BoolValue(true),
				LabelOrientation: types.StringValue("angled"),
				Scale:            types.StringValue("log"),
				Title: &axisTitleModel{
					Value:   types.StringValue("Right Y Axis"),
					Visible: types.BoolValue(true),
				},
			},
		},
		{
			name:     "nil axis",
			apiAxis:  nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test fromAPIRight
			model := &yAxisConfigModel{}
			diags := model.fromAPIRight(tt.apiAxis)
			require.False(t, diags.HasError())

			if tt.expected == nil {
				return
			}

			assert.Equal(t, tt.expected.Grid, model.Grid)
			assert.Equal(t, tt.expected.Ticks, model.Ticks)
			assert.Equal(t, tt.expected.LabelOrientation, model.LabelOrientation)
			assert.Equal(t, tt.expected.Scale, model.Scale)

			// Test toAPIRight round-trip
			apiAxis, diags := model.toAPIRight()
			require.False(t, diags.HasError())

			if tt.apiAxis != nil {
				assert.NotNil(t, apiAxis)
			}
		})
	}
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

func Test_xyDecorationsModel_fromAPI_toAPI(t *testing.T) {
	tests := []struct {
		name            string
		apiDecorations  kbapi.XyDecorations
		expected        *xyDecorationsModel
		expectFillValue float64 // Expected rounded value for fill opacity
	}{
		{
			name: "all fields populated",
			apiDecorations: kbapi.XyDecorations{
				EndZones: &struct {
					Visible bool `json:"visible"`
				}{Visible: true},
				CurrentTimeMarker: &struct {
					Visible bool `json:"visible"`
				}{Visible: false},
				PointVisibility: func() *kbapi.XyDecorationsPointVisibility {
					v := kbapi.XyDecorationsPointVisibilityVisible
					return &v
				}(),
				LineInterpolation: func() *kbapi.XyDecorationsLineInterpolation {
					i := kbapi.XyDecorationsLineInterpolation("linear")
					return &i
				}(),
				MinimumBarHeight: new(float32(5)),
				Values: &struct {
					Visible bool `json:"visible"`
				}{Visible: true},
				FillOpacity: new(float32(0.5)),
			},
			expected: &xyDecorationsModel{
				ShowEndZones:          types.BoolValue(true),
				ShowCurrentTimeMarker: types.BoolValue(false),
				PointVisibility:       types.StringValue("visible"),
				LineInterpolation:     types.StringValue("linear"),
				MinimumBarHeight:      types.Int64Value(5),
				ShowValueLabels:       types.BoolValue(true),
				FillOpacity:           types.Float64Value(0.5),
			},
			expectFillValue: 0.5,
		},
		{
			name:           "nil values",
			apiDecorations: kbapi.XyDecorations{},
			expected: &xyDecorationsModel{
				ShowEndZones:          types.BoolNull(),
				ShowCurrentTimeMarker: types.BoolNull(),
				PointVisibility:       types.StringNull(),
				LineInterpolation:     types.StringNull(),
				MinimumBarHeight:      types.Int64Null(),
				ShowValueLabels:       types.BoolNull(),
				FillOpacity:           types.Float64Null(),
			},
		},
		{
			name: "float precision rounding",
			apiDecorations: kbapi.XyDecorations{
				FillOpacity: new(float32(0.123456)),
			},
			expected: &xyDecorationsModel{
				FillOpacity: types.Float64Value(0.12),
			},
			expectFillValue: 0.12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test fromAPI
			model := &xyDecorationsModel{}
			model.fromAPI(tt.apiDecorations)

			assert.Equal(t, tt.expected.ShowEndZones, model.ShowEndZones)
			assert.Equal(t, tt.expected.ShowCurrentTimeMarker, model.ShowCurrentTimeMarker)
			assert.Equal(t, tt.expected.PointVisibility, model.PointVisibility)
			assert.Equal(t, tt.expected.LineInterpolation, model.LineInterpolation)
			assert.Equal(t, tt.expected.MinimumBarHeight, model.MinimumBarHeight)
			assert.Equal(t, tt.expected.ShowValueLabels, model.ShowValueLabels)

			if !tt.expected.FillOpacity.IsNull() {
				assert.InDelta(t, tt.expectFillValue, model.FillOpacity.ValueFloat64(), 0.001)
			} else {
				assert.True(t, model.FillOpacity.IsNull())
			}

			// Test toAPI
			apiDecorations := model.toAPI()
			assert.NotNil(t, apiDecorations)

			// Verify round-trip preserves known values
			if !model.ShowEndZones.IsNull() && !model.ShowEndZones.IsUnknown() {
				require.NotNil(t, apiDecorations.EndZones)
				assert.Equal(t, model.ShowEndZones.ValueBool(), apiDecorations.EndZones.Visible)
			}
		})
	}
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
				Type:     kbapi.XyFittingType("linear"),
				Dotted:   new(true),
				EndValue: func() *kbapi.XyFittingEndValue { e := kbapi.XyFittingEndValue("zero"); return &e }(),
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
				Type:     kbapi.XyFittingType("none"),
				Dotted:   nil,
				EndValue: nil,
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
							MaxLines *float32 `json:"max_lines,omitempty"`
						} `json:"truncate,omitempty"`
						Type kbapi.XyLegendInsideLayoutType `json:"type"`
					}{
						Truncate: &struct {
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
				visibility := kbapi.XyLegendOutsideVerticalVisibilityHidden
				position := kbapi.XyLegendOutsideVerticalPositionRight
				placement := kbapi.XyLegendOutsideVerticalPlacementOutside
				legend := kbapi.XyLegendOutsideVertical{
					Visibility: &visibility,
					Layout: &struct {
						Truncate *struct {
							MaxLines *float32 `json:"max_lines,omitempty"`
						} `json:"truncate,omitempty"`
						Type kbapi.XyLegendOutsideVerticalLayoutType `json:"type"`
					}{
						Truncate: &struct {
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
		Layers: []xyLayerModel{
			{
				Type: types.StringValue("area"),
				DataLayer: &dataLayerModel{
					DatasetJSON: jsontypes.NewNormalizedValue(`{"type":"dataView","id":"logs-*"}`),
					Y: []yMetricModel{
						{ConfigJSON: jsontypes.NewNormalizedValue(`{"operation":"count","color":"#68BC00","axis":"left"}`)},
					},
				},
			},
		},
		Query: &filterSimpleModel{
			Query:    types.StringValue("*"),
			Language: types.StringValue("kuery"),
		},
	}

	xyChart, diags := model.toAPI()
	require.False(t, diags.HasError())

	var attrs kbapi.LensApiState
	require.NoError(t, attrs.FromXyChart(xyChart))

	converter := newXYChartPanelConfigConverter()
	pm := &panelModel{}
	diags = converter.populateFromAttributes(ctx, pm, attrs)
	require.False(t, diags.HasError())
	require.NotNil(t, pm.XYChartConfig)

	attrs2, diags := converter.buildAttributes(*pm)
	require.False(t, diags.HasError())

	chart2, err := attrs2.AsXyChart()
	require.NoError(t, err)
	assert.Equal(t, kbapi.Xy, chart2.Type)
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
							DatasetJSON: jsontypes.NewNormalizedValue(`{"type":"dataView","id":"logs-*"}`),
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
					Query:    types.StringValue("*"),
					Language: types.StringValue("kuery"),
				},
			},
			expectError: false,
		},
		{
			name: "minimal config",
			model: &xyChartConfigModel{
				Title: types.StringValue("Minimal Chart"),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test toAPI
			apiChart, diags := tt.model.toAPI()
			if tt.expectError {
				assert.True(t, diags.HasError())
				return
			}

			require.False(t, diags.HasError())
			assert.Equal(t, kbapi.Xy, apiChart.Type)

			if tt.model.Title.ValueString() != "" {
				assert.Equal(t, tt.model.Title.ValueString(), *apiChart.Title)
			}

			// Test fromAPI round-trip
			model2 := &xyChartConfigModel{}
			diags = model2.fromAPI(ctx, apiChart)
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

func Test_yAxisConfigModel_toAPILeft_nil(t *testing.T) {
	var model *yAxisConfigModel
	apiAxis, diags := model.toAPILeft()
	assert.False(t, diags.HasError())
	assert.Nil(t, apiAxis)
}

func Test_yAxisConfigModel_toAPIRight_nil(t *testing.T) {
	var model *yAxisConfigModel
	apiAxis, diags := model.toAPIRight()
	assert.False(t, diags.HasError())
	assert.Nil(t, apiAxis)
}

func Test_axisTitleModel_toAPI_nil(t *testing.T) {
	var model *axisTitleModel
	apiTitle := model.toAPI()
	assert.Nil(t, apiTitle)
}

func Test_xyDecorationsModel_toAPI_nil(t *testing.T) {
	var model *xyDecorationsModel
	apiDecorations := model.toAPI()
	assert.NotNil(t, apiDecorations) // Returns empty struct, not nil
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
