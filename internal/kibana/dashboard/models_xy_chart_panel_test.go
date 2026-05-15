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
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/lensxy"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_xyAxisModel_fromAPI_toAPI(t *testing.T) {
	raw := []byte(`{
		"x":{"grid":{"visible":true},"ticks":{"visible":false}},
		"y":{"grid":{"visible":true},"ticks":{"visible":true},"domain":{"type":"fit"}},
		"y2":{"grid":{"visible":false},"ticks":{"visible":true},"domain":{"type":"fit"}}
	}`)
	var apiAxis kbapi.VisApiXyAxisConfig
	require.NoError(t, json.Unmarshal(raw, &apiAxis))

	model := &models.XYAxisModel{}
	diags := xyAxisFromAPI(model, apiAxis)
	require.False(t, diags.HasError())

	require.NotNil(t, model.X)
	assert.Equal(t, types.BoolValue(true), model.X.Grid)
	assert.Equal(t, types.BoolValue(false), model.X.Ticks)

	require.NotNil(t, model.Y)
	assert.Equal(t, types.BoolValue(true), model.Y.Grid)
	assert.Equal(t, types.BoolValue(true), model.Y.Ticks)

	require.NotNil(t, model.Y2)
	assert.Equal(t, types.BoolValue(false), model.Y2.Grid)

	out, d := xyAxisToAPI(model)
	require.False(t, d.HasError())
	model2 := &models.XYAxisModel{}
	diags = xyAxisFromAPI(model2, out)
	require.False(t, diags.HasError())
	assert.Equal(t, model.X.Grid, model2.X.Grid)
	assert.Equal(t, model.Y.Grid, model2.Y.Grid)
}

func Test_xyAxisConfigModel_fromAPI_toAPI(t *testing.T) {
	tests := []struct {
		name     string
		apiAxis  *xyAxisConfigAPIModel
		expected *models.XYAxisConfigModel
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
			expected: &models.XYAxisConfigModel{
				Grid:             types.BoolValue(true),
				Ticks:            types.BoolValue(false),
				LabelOrientation: types.StringValue("horizontal"),
				Title: &models.AxisTitleModel{
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
			expected: &models.XYAxisConfigModel{
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
			expected: &models.XYAxisConfigModel{
				Grid:  types.BoolValue(true),
				Ticks: types.BoolValue(true),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test fromAPI
			model := &models.XYAxisConfigModel{}
			diags := xyAxisConfigFromAPI(model, tt.apiAxis)
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
			apiAxis, diags := xyAxisConfigToAPI(model)
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

	model := &models.YAxisConfigModel{}
	diags := YAxisConfigFromAPIY(model, env.Y)
	require.False(t, diags.HasError())

	assert.Equal(t, types.BoolValue(true), model.Grid)
	assert.Equal(t, types.BoolValue(false), model.Ticks)
	assert.Equal(t, types.StringValue("vertical"), model.LabelOrientation)
	assert.Equal(t, types.StringValue("linear"), model.Scale)
	require.NotNil(t, model.Title)
	assert.Equal(t, types.StringValue("Y Axis Title"), model.Title.Value)
	assert.Equal(t, types.BoolValue(false), model.Title.Visible)

	apiY, diags := YAxisConfigToAPIY(model)
	require.False(t, diags.HasError())
	require.NotNil(t, apiY)
}

func Test_yAxisConfigModel_fromAPISecondaryY_toAPISecondaryY(t *testing.T) {
	raw := []byte(`{
		"y2":{
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
	require.NotNil(t, env.Y2)

	model := &models.YAxisConfigModel{}
	diags := YAxisConfigFromAPIY2(model, env.Y2)
	require.False(t, diags.HasError())

	assert.Equal(t, types.BoolValue(false), model.Grid)
	assert.Equal(t, types.BoolValue(true), model.Ticks)
	assert.Equal(t, types.StringValue("angled"), model.LabelOrientation)
	assert.Equal(t, types.StringValue("log"), model.Scale)

	apiY, diags := YAxisConfigToAPIY2(model)
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
		expected *models.AxisTitleModel
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
			expected: &models.AxisTitleModel{
				Value:   types.StringValue("Test Title"),
				Visible: types.BoolValue(true),
			},
		},
		{
			name:     "nil title",
			apiTitle: nil,
			expected: &models.AxisTitleModel{},
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
			expected: &models.AxisTitleModel{
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
			expected: &models.AxisTitleModel{
				Value:   types.StringNull(),
				Visible: types.BoolValue(false),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test fromAPI
			model := &models.AxisTitleModel{}
			lenscommon.AxisTitleFromAPI(model, tt.apiTitle)
			assert.Equal(t, tt.expected.Value, model.Value)
			assert.Equal(t, tt.expected.Visible, model.Visible)

			// Test toAPI
			apiTitle := lenscommon.AxisTitleToAPI(model)
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

	model := &models.XYDecorationsModel{}
	xyDecorationsReadFromStyling(model, styling)

	assert.Equal(t, types.BoolValue(true), model.ShowEndZones)
	assert.Equal(t, types.BoolValue(false), model.ShowCurrentTimeMarker)
	assert.Equal(t, types.StringValue("always"), model.PointVisibility)
	assert.Equal(t, types.StringValue("linear"), model.LineInterpolation)
	assert.Equal(t, types.Int64Value(5), model.MinimumBarHeight)
	assert.Equal(t, types.BoolValue(true), model.ShowValueLabels)
	assert.InDelta(t, 0.5, model.FillOpacity.ValueFloat64(), 0.001)

	var out kbapi.XyStyling
	out.Fitting = kbapi.XyFitting{Type: kbapi.XyFittingTypeNone}
	xyDecorationsWriteToStyling(model, &out)
	model2 := &models.XYDecorationsModel{}
	xyDecorationsReadFromStyling(model2, out)
	assert.Equal(t, model.ShowEndZones, model2.ShowEndZones)
	assert.Equal(t, model.PointVisibility, model2.PointVisibility)
}

func Test_xyFittingModel_fromAPI_toAPI(t *testing.T) {
	tests := []struct {
		name        string
		apiFitting  kbapi.XyFitting
		expected    *models.XYFittingModel
		expectError bool
	}{
		{
			name: "all fields populated",
			apiFitting: kbapi.XyFitting{
				Type:      kbapi.XyFittingType("linear"),
				Emphasize: new(true),
				Extend:    func() *kbapi.XyFittingExtend { e := kbapi.XyFittingExtendZero; return &e }(),
			},
			expected: &models.XYFittingModel{
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
			expected: &models.XYFittingModel{
				Type:     types.StringValue("none"),
				Dotted:   types.BoolNull(),
				EndValue: types.StringNull(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test fromAPI
			model := &models.XYFittingModel{}
			xyFittingFromAPI(model, tt.apiFitting)

			assert.Equal(t, tt.expected.Type, model.Type)
			assert.Equal(t, tt.expected.Dotted, model.Dotted)
			assert.Equal(t, tt.expected.EndValue, model.EndValue)

			// Test toAPI
			apiFitting := xyFittingToAPI(model)
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
		expected  *models.XYLegendModel
	}{
		{
			name: "inside legend with all fields",
			apiLegend: func() kbapi.XyLegend {
				visibility := kbapi.XyLegendInsideVisibilityVisible
				position := kbapi.TopLeft
				legend := kbapi.XyLegendInside{
					Placement:  kbapi.Inside,
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
			expected: &models.XYLegendModel{
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
			model := &models.XYLegendModel{}
			diags := xyLegendFromAPI(ctx, model, tt.apiLegend)
			require.False(t, diags.HasError())

			assert.Equal(t, tt.expected.Inside, model.Inside)
			assert.Equal(t, tt.expected.Visibility, model.Visibility)
			assert.Equal(t, tt.expected.TruncateAfterLines, model.TruncateAfterLines)
			assert.Equal(t, tt.expected.Columns, model.Columns)
			assert.Equal(t, tt.expected.Alignment, model.Alignment)

			// Test toAPI
			apiLegend, diags := xyLegendToAPI(model)
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
		expected  *models.XYLegendModel
	}{
		{
			name: "outside legend with all fields",
			apiLegend: func() kbapi.XyLegend {
				visibility := kbapi.XyLegendOutsideVerticalVisibility("hidden")
				position := kbapi.XyLegendOutsideVerticalPositionRight
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
			expected: &models.XYLegendModel{
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
			model := &models.XYLegendModel{}
			diags := xyLegendFromAPI(ctx, model, tt.apiLegend)
			require.False(t, diags.HasError())

			assert.Equal(t, tt.expected.Inside, model.Inside)
			assert.Equal(t, tt.expected.Visibility, model.Visibility)
			assert.Equal(t, tt.expected.TruncateAfterLines, model.TruncateAfterLines)
			assert.Equal(t, tt.expected.Position, model.Position)
			assert.Equal(t, tt.expected.Size, model.Size)

			// Test toAPI
			apiLegend, diags := xyLegendToAPI(model)
			require.False(t, diags.HasError())
			assert.NotNil(t, apiLegend)
		})
	}
}

func Test_xyChartPanelConfigConverter_populateFromAttributes_buildAttributes_roundTrip(t *testing.T) {
	ctx := context.Background()

	model := &models.XYChartConfigModel{
		Title:       types.StringValue("XY Chart Round-Trip"),
		Description: types.StringValue("Converter test"),
		Axis: &models.XYAxisModel{
			X: &models.XYAxisConfigModel{},
			Y: &models.YAxisConfigModel{},
		},
		Decorations: &models.XYDecorationsModel{},
		Fitting:     &models.XYFittingModel{Type: types.StringValue("none")},
		Layers: []models.XYLayerModel{
			{
				Type: types.StringValue("area"),
				DataLayer: &models.DataLayerModel{
					DataSourceJSON: jsontypes.NewNormalizedValue(`{"type":"dataView","id":"logs-*"}`),
					Y: []models.YMetricModel{
						{ConfigJSON: jsontypes.NewNormalizedValue(`{"operation":"count","color":"#68BC00","axis":"left"}`)},
					},
				},
			},
		},
		Legend: &models.XYLegendModel{
			Inside:     types.BoolValue(false),
			Visibility: types.StringValue("visible"),
		},
		Query: &models.FilterSimpleModel{
			Expression: types.StringValue("*"),
			Language:   types.StringValue("kql"),
		},
	}

	xyChart, diags := xyChartConfigToAPINoESQL(model, nil)
	require.False(t, diags.HasError())

	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	require.NoError(t, attrs.FromXyChartNoESQL(xyChart))

	c := lenscommon.ForType(string(kbapi.XyChartNoESQLTypeXy))
	require.NotNil(t, c)
	visBv := models.VisByValueModel{}
	diags = c.PopulateFromAttributes(ctx, lensChartResolver(nil), &visBv.LensByValueChartBlocks, attrs)
	require.False(t, diags.HasError())
	require.NotNil(t, visBv.XYChartConfig)

	attrs2, diags := c.BuildAttributes(&visBv.LensByValueChartBlocks, lensChartResolver(nil))
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
		model       *models.XYChartConfigModel
		expectError bool
	}{
		{
			name: "complete config",
			model: &models.XYChartConfigModel{
				Title:       types.StringValue("Test XY Chart"),
				Description: types.StringValue("A test chart"),
				Axis: &models.XYAxisModel{
					X: &models.XYAxisConfigModel{
						Grid:  types.BoolValue(true),
						Ticks: types.BoolValue(false),
					},
					Y: &models.YAxisConfigModel{},
				},
				Decorations: &models.XYDecorationsModel{
					ShowEndZones:    types.BoolValue(true),
					PointVisibility: types.StringValue("never"),
				},
				Fitting: &models.XYFittingModel{
					Type:   types.StringValue("linear"),
					Dotted: types.BoolValue(true),
				},
				Layers: []models.XYLayerModel{
					{
						Type: types.StringValue("area"),
						DataLayer: &models.DataLayerModel{
							DataSourceJSON: jsontypes.NewNormalizedValue(`{"type":"dataView","id":"logs-*"}`),
							Y: []models.YMetricModel{
								{
									ConfigJSON: jsontypes.NewNormalizedValue(`{"operation":"count","color":"#68BC00","axis":"left"}`),
								},
							},
						},
					},
				},
				Legend: &models.XYLegendModel{
					Inside:     types.BoolValue(false),
					Visibility: types.StringValue("visible"),
				},
				Query: &models.FilterSimpleModel{
					Expression: types.StringValue("*"),
					Language:   types.StringValue("kql"),
				},
			},
			expectError: false,
		},
		{
			name: "minimal config",
			model: &models.XYChartConfigModel{
				Title: types.StringValue("Minimal Chart"),
				Axis: &models.XYAxisModel{
					X: &models.XYAxisConfigModel{},
					Y: &models.YAxisConfigModel{},
				},
				Decorations: &models.XYDecorationsModel{},
				Fitting:     &models.XYFittingModel{Type: types.StringValue("none")},
				Legend: &models.XYLegendModel{
					Inside:     types.BoolValue(false),
					Visibility: types.StringValue("visible"),
				},
				Query: &models.FilterSimpleModel{
					Expression: types.StringValue("*"),
					Language:   types.StringValue("kql"),
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiChart, diags := xyChartConfigToAPINoESQL(tt.model, nil)
			if tt.expectError {
				assert.True(t, diags.HasError())
				return
			}

			require.False(t, diags.HasError())
			assert.Equal(t, kbapi.XyChartNoESQLTypeXy, apiChart.Type)

			if tt.model.Title.ValueString() != "" {
				assert.Equal(t, tt.model.Title.ValueString(), *apiChart.Title)
			}

			model2 := &models.XYChartConfigModel{}
			diags = xyChartConfigFromAPINoESQL(ctx, model2, nil, nil, apiChart)
			require.False(t, diags.HasError())

			assert.Equal(t, tt.model.Title, model2.Title)
			assert.Equal(t, tt.model.Description, model2.Description)
		})
	}
}

func minimalXYChartConfigForPresentationTests() *models.XYChartConfigModel {
	return &models.XYChartConfigModel{
		Title: types.StringValue("Presentation Wiring"),
		Axis: &models.XYAxisModel{
			X: &models.XYAxisConfigModel{},
			Y: &models.YAxisConfigModel{},
		},
		Decorations: &models.XYDecorationsModel{},
		Fitting:     &models.XYFittingModel{Type: types.StringValue("none")},
		Legend: &models.XYLegendModel{
			Inside:     types.BoolValue(false),
			Visibility: types.StringValue("visible"),
		},
		Query: &models.FilterSimpleModel{
			Expression: types.StringValue("*"),
			Language:   types.StringValue("kql"),
		},
	}
}

func Test_xyChartConfigModel_esqlMode_queryNil(t *testing.T) {
	// ES|QL XY panels should have nil query; query is optional on the schema.
	// Build a model in ES|QL mode (all data layers have esql data_source) and verify
	// that toAPIESQL emits a valid payload without requiring query.
	m := &models.XYChartConfigModel{
		Title: types.StringValue("ESQL XY Chart"),
		Axis: &models.XYAxisModel{
			X: &models.XYAxisConfigModel{},
			Y: &models.YAxisConfigModel{},
		},
		Decorations: &models.XYDecorationsModel{},
		Fitting:     &models.XYFittingModel{Type: types.StringValue("none")},
		Layers: []models.XYLayerModel{
			{
				Type: types.StringValue("area"),
				DataLayer: &models.DataLayerModel{
					DataSourceJSON: jsontypes.NewNormalizedValue(`{"type":"esql","query":"FROM logs-* | LIMIT 10"}`),
					Y: []models.YMetricModel{
						{ConfigJSON: jsontypes.NewNormalizedValue(`{"operation":"count","axis":"left"}`)},
					},
				},
			},
		},
		Legend: &models.XYLegendModel{
			Inside:     types.BoolValue(false),
			Visibility: types.StringValue("visible"),
		},
		Query: nil, // ES|QL mode: no query
	}

	// Verify ES|QL mode detection
	assert.True(t, xyChartConfigXyUsesESQL(m))

	// Build the ES|QL API payload
	esqlChart, diags := xyChartConfigToAPIESQL(m, nil)
	require.False(t, diags.HasError(), "diags: %v", diags)
	assert.Equal(t, kbapi.XyChartESQLTypeXy, esqlChart.Type)
	assert.Len(t, esqlChart.Layers, 1)
}

func Test_xyChartConfigModel_optionalQuery_noESQL_nilQueryNoError(t *testing.T) {
	// Non-ES|QL XY panels with nil query should NOT error; query is optional in the schema.
	// The API payload will simply omit the query field.
	m := &models.XYChartConfigModel{
		Title: types.StringValue("Non-ESQL XY Chart no query"),
		Axis: &models.XYAxisModel{
			X: &models.XYAxisConfigModel{},
			Y: &models.YAxisConfigModel{},
		},
		Decorations: &models.XYDecorationsModel{},
		Fitting:     &models.XYFittingModel{Type: types.StringValue("none")},
		Layers: []models.XYLayerModel{
			{
				Type: types.StringValue("area"),
				DataLayer: &models.DataLayerModel{
					DataSourceJSON: jsontypes.NewNormalizedValue(`{"type":"dataView","id":"logs-*"}`),
					Y: []models.YMetricModel{
						{ConfigJSON: jsontypes.NewNormalizedValue(`{"operation":"count","axis":"left"}`)},
					},
				},
			},
		},
		Legend: &models.XYLegendModel{
			Inside:     types.BoolValue(false),
			Visibility: types.StringValue("visible"),
		},
		Query: nil, // Query is now optional
	}

	assert.False(t, xyChartConfigXyUsesESQL(m))

	_, diags := xyChartConfigToAPINoESQL(m, nil)
	assert.False(t, diags.HasError(), "expected no error when non-ES|QL XY chart has no query; query is optional: %v", diags)
}

func Test_xyChartConfig_lensChartPresentation_timeRange_inheritanceAndMode(t *testing.T) {
	ctx := context.Background()

	dash := &models.DashboardModel{
		TimeRange: &models.TimeRangeModel{
			From: types.StringValue("now-7d"),
			To:   types.StringValue("now"),
		},
	}
	dashTR := timeRangeModelToAPI(dash.TimeRange)

	t.Run("null in plan and API echoes dashboard time range preserves null state", func(t *testing.T) {
		m := minimalXYChartConfigForPresentationTests()
		require.Nil(t, m.TimeRange)

		apiChart, diags := xyChartConfigToAPINoESQL(m, dash)
		require.False(t, diags.HasError())
		assert.True(t, lensTimeRangesAPILiteralEqual(apiChart.TimeRange, dashTR), "write path should inherit dashboard time_range when chart-level is null")

		apiChart.TimeRange = dashTR
		out := &models.XYChartConfigModel{}
		diags = xyChartConfigFromAPINoESQL(ctx, out, dash, m, apiChart)
		require.False(t, diags.HasError())
		assert.Nil(t, out.TimeRange)
	})

	t.Run("prior-null chart time_range and API differs from dashboard populates state", func(t *testing.T) {
		m := minimalXYChartConfigForPresentationTests()
		require.Nil(t, m.TimeRange)

		base := minimalXYChartConfigForPresentationTests()
		apiChart, diags := xyChartConfigToAPINoESQL(base, dash)
		require.False(t, diags.HasError())
		apiChart.TimeRange = kbapi.KbnEsQueryServerTimeRangeSchema{
			From: "now-30d",
			To:   "now-1d",
		}
		require.False(t, lensTimeRangesAPILiteralEqual(apiChart.TimeRange, dashTR), "API chart time_range should differ from dashboard for this scenario")

		out := &models.XYChartConfigModel{}
		diags = xyChartConfigFromAPINoESQL(ctx, out, dash, m, apiChart)
		require.False(t, diags.HasError())
		require.NotNil(t, out.TimeRange)
		assert.Equal(t, "now-30d", out.TimeRange.From.ValueString())
		assert.Equal(t, "now-1d", out.TimeRange.To.ValueString())
	})

	t.Run("explicit chart time_range override round-trips", func(t *testing.T) {
		m := minimalXYChartConfigForPresentationTests()
		m.TimeRange = &models.TimeRangeModel{
			From: types.StringValue("now-30d"),
			To:   types.StringValue("now-1d"),
		}

		apiChart, diags := xyChartConfigToAPINoESQL(m, dash)
		require.False(t, diags.HasError())
		assert.Equal(t, "now-30d", apiChart.TimeRange.From)
		assert.Equal(t, "now-1d", apiChart.TimeRange.To)

		want := *m.TimeRange
		out := &models.XYChartConfigModel{}
		diags = xyChartConfigFromAPINoESQL(ctx, out, dash, m, apiChart)
		require.False(t, diags.HasError())
		require.NotNil(t, out.TimeRange)
		assert.Equal(t, want.From, out.TimeRange.From)
		assert.Equal(t, want.To, out.TimeRange.To)
	})

	t.Run("time_range mode null preserved when API omits mode", func(t *testing.T) {
		prior := minimalXYChartConfigForPresentationTests()
		prior.TimeRange = &models.TimeRangeModel{
			From: types.StringValue("now-7d"),
			To:   types.StringValue("now"),
			Mode: types.StringNull(),
		}

		apiChart := func() kbapi.XyChartNoESQL {
			m := minimalXYChartConfigForPresentationTests()
			api, diags := xyChartConfigToAPINoESQL(m, dash)
			require.False(t, diags.HasError())
			return api
		}()
		apiChart.TimeRange = kbapi.KbnEsQueryServerTimeRangeSchema{
			From: "now-7d",
			To:   "now",
		}

		out := &models.XYChartConfigModel{}
		diags := xyChartConfigFromAPINoESQL(ctx, out, dash, prior, apiChart)
		require.False(t, diags.HasError())
		require.NotNil(t, out.TimeRange)
		assert.True(t, out.TimeRange.Mode.IsNull())
		assert.Equal(t, "now-7d", out.TimeRange.From.ValueString())
		assert.Equal(t, "now", out.TimeRange.To.ValueString())
	})
}

func Test_xyChartConfigModel_lensChartPresentation_boolsReferences_andNullPreservation(t *testing.T) {
	ctx := context.Background()
	dash := &models.DashboardModel{
		TimeRange: &models.TimeRangeModel{
			From: types.StringValue("now-7d"),
			To:   types.StringValue("now"),
		},
	}

	t.Run("hide_title round trip true and false", func(t *testing.T) {
		for _, v := range []bool{true, false} {
			m := minimalXYChartConfigForPresentationTests()
			m.HideTitle = types.BoolValue(v)

			apiChart, diags := xyChartConfigToAPINoESQL(m, dash)
			require.False(t, diags.HasError())
			require.NotNil(t, apiChart.HideTitle)
			assert.Equal(t, v, *apiChart.HideTitle)

			out := &models.XYChartConfigModel{}
			diags = xyChartConfigFromAPINoESQL(ctx, out, dash, m, apiChart)
			require.False(t, diags.HasError())
			assert.Equal(t, types.BoolValue(v), out.HideTitle)
		}
	})

	t.Run("hide_border round trip true and false", func(t *testing.T) {
		for _, v := range []bool{true, false} {
			m := minimalXYChartConfigForPresentationTests()
			m.HideBorder = types.BoolValue(v)

			apiChart, diags := xyChartConfigToAPINoESQL(m, dash)
			require.False(t, diags.HasError())
			require.NotNil(t, apiChart.HideBorder)
			assert.Equal(t, v, *apiChart.HideBorder)

			out := &models.XYChartConfigModel{}
			diags = xyChartConfigFromAPINoESQL(ctx, out, dash, m, apiChart)
			require.False(t, diags.HasError())
			assert.Equal(t, types.BoolValue(v), out.HideBorder)
		}
	})

	t.Run("hide_title null preserved when API omits", func(t *testing.T) {
		prior := minimalXYChartConfigForPresentationTests()
		prior.HideTitle = types.BoolNull()

		m := minimalXYChartConfigForPresentationTests()
		apiChart, diags := xyChartConfigToAPINoESQL(m, dash)
		require.False(t, diags.HasError())
		apiChart.HideTitle = nil

		out := &models.XYChartConfigModel{}
		diags = xyChartConfigFromAPINoESQL(ctx, out, dash, prior, apiChart)
		require.False(t, diags.HasError())
		assert.True(t, out.HideTitle.IsNull())
	})

	t.Run("hide_border null preserved when API omits", func(t *testing.T) {
		prior := minimalXYChartConfigForPresentationTests()
		prior.HideBorder = types.BoolNull()

		m := minimalXYChartConfigForPresentationTests()
		apiChart, diags := xyChartConfigToAPINoESQL(m, dash)
		require.False(t, diags.HasError())
		apiChart.HideBorder = nil

		out := &models.XYChartConfigModel{}
		diags = xyChartConfigFromAPINoESQL(ctx, out, dash, prior, apiChart)
		require.False(t, diags.HasError())
		assert.True(t, out.HideBorder.IsNull())
	})

	t.Run("references_json normalized round trip", func(t *testing.T) {
		raw := `[{"id":"dash1","name":"Target","type":"dashboard"}]`
		m := minimalXYChartConfigForPresentationTests()
		m.ReferencesJSON = jsontypes.NewNormalizedValue(raw)

		apiChart, diags := xyChartConfigToAPINoESQL(m, dash)
		require.False(t, diags.HasError())
		require.NotNil(t, apiChart.References)

		out := &models.XYChartConfigModel{}
		diags = xyChartConfigFromAPINoESQL(ctx, out, dash, m, apiChart)
		require.False(t, diags.HasError())
		require.True(t, typeutils.IsKnown(out.ReferencesJSON))
		assert.JSONEq(t, raw, out.ReferencesJSON.ValueString())
	})

	t.Run("references_json null preserved when API omits references", func(t *testing.T) {
		prior := minimalXYChartConfigForPresentationTests()
		prior.ReferencesJSON = jsontypes.NewNormalizedNull()

		m := minimalXYChartConfigForPresentationTests()
		apiChart, diags := xyChartConfigToAPINoESQL(m, dash)
		require.False(t, diags.HasError())
		apiChart.References = nil

		out := &models.XYChartConfigModel{}
		diags = xyChartConfigFromAPINoESQL(ctx, out, dash, prior, apiChart)
		require.False(t, diags.HasError())
		assert.True(t, out.ReferencesJSON.IsNull())
	})

	t.Run("references_json null preserved when API returns empty slice", func(t *testing.T) {
		prior := minimalXYChartConfigForPresentationTests()
		prior.ReferencesJSON = jsontypes.NewNormalizedNull()

		m := minimalXYChartConfigForPresentationTests()
		apiChart, diags := xyChartConfigToAPINoESQL(m, dash)
		require.False(t, diags.HasError())
		empty := []kbapi.KbnContentManagementUtilsReferenceSchema{}
		apiChart.References = &empty

		out := &models.XYChartConfigModel{}
		diags = xyChartConfigFromAPINoESQL(ctx, out, dash, prior, apiChart)
		require.False(t, diags.HasError())
		assert.True(t, out.ReferencesJSON.IsNull())
	})
}

func Test_xyChartConfigModel_lensChartPresentation_dashboardDrilldown_roundTrip(t *testing.T) {
	ctx := context.Background()
	dash := &models.DashboardModel{
		TimeRange: &models.TimeRangeModel{
			From: types.StringValue("now-7d"),
			To:   types.StringValue("now"),
		},
	}

	m := minimalXYChartConfigForPresentationTests()
	m.Drilldowns = []models.LensDrilldownItemTFModel{
		{
			DashboardDrilldown: &models.LensDashboardDrilldownTFModel{
				DashboardID:  types.StringValue("dash-abc"),
				Label:        types.StringValue("Open related"),
				UseFilters:   types.BoolValue(true),
				UseTimeRange: types.BoolValue(true),
				OpenInNewTab: types.BoolValue(false),
			},
		},
	}

	apiChart, diags := xyChartConfigToAPINoESQL(m, dash)
	require.False(t, diags.HasError())
	require.NotNil(t, apiChart.Drilldowns)
	require.GreaterOrEqual(t, len(*apiChart.Drilldowns), 1)

	raw, err := json.Marshal((*apiChart.Drilldowns)[0])
	require.NoError(t, err)
	var wire map[string]any
	require.NoError(t, json.Unmarshal(raw, &wire))
	assert.Equal(t, "dashboard_drilldown", wire["type"])
	assert.Equal(t, lensDrilldownTriggerOnApplyFilter, wire["trigger"])

	out := &models.XYChartConfigModel{}
	diags = xyChartConfigFromAPINoESQL(ctx, out, dash, m, apiChart)
	require.False(t, diags.HasError())
	require.Len(t, out.Drilldowns, 1)
	require.NotNil(t, out.Drilldowns[0].DashboardDrilldown)
	assert.Equal(t, "dash-abc", out.Drilldowns[0].DashboardDrilldown.DashboardID.ValueString())
	assert.Equal(t, lensDrilldownTriggerOnApplyFilter, out.Drilldowns[0].DashboardDrilldown.Trigger.ValueString())
}

func Test_xyAxisConfigModel_toAPI_nil(t *testing.T) {
	var model *models.XYAxisConfigModel
	apiAxis, diags := xyAxisConfigToAPI(model)
	assert.False(t, diags.HasError())
	assert.Nil(t, apiAxis)
}

func Test_yAxisConfigModel_toAPIY_nil(t *testing.T) {
	var model *models.YAxisConfigModel
	apiAxis, diags := YAxisConfigToAPIY(model)
	assert.False(t, diags.HasError())
	assert.Nil(t, apiAxis)
}

func Test_yAxisConfigModel_toAPIY2_nil(t *testing.T) {
	var model *models.YAxisConfigModel
	apiAxis, diags := YAxisConfigToAPIY2(model)
	assert.False(t, diags.HasError())
	assert.Nil(t, apiAxis)
}

func Test_axisTitleModel_toAPI_nil(t *testing.T) {
	var model *models.AxisTitleModel
	apiTitle := lenscommon.AxisTitleToAPI(model)
	assert.Nil(t, apiTitle)
}

func Test_xyDecorationsModel_writeToStyling_nil(t *testing.T) {
	var model *models.XYDecorationsModel
	var s kbapi.XyStyling
	xyDecorationsWriteToStyling(model, &s)
	assert.Nil(t, s.Overlays.PartialBuckets)
}

func Test_xyFittingModel_toAPI_nil(t *testing.T) {
	var model *models.XYFittingModel
	apiFitting := xyFittingToAPI(model)
	assert.NotNil(t, apiFitting) // Returns empty struct, not nil
}

func Test_xyLegendModel_toAPI_nil(t *testing.T) {
	var model *models.XYLegendModel
	apiLegend, diags := xyLegendToAPI(model)
	assert.False(t, diags.HasError())
	// Check it doesn't panic
	assert.NotNil(t, apiLegend)
}

func Test_alignXYChartStateFromPlanPanels_preservesPractitionerIntent(t *testing.T) {
	planPanels := []models.PanelModel{
		{
			VisConfig: &models.VisConfigModel{
				ByValue: &models.VisByValueModel{
					LensByValueChartBlocks: models.LensByValueChartBlocks{
						XYChartConfig: &models.XYChartConfigModel{
							Title: types.StringValue("Sample XY Chart"),
							Axis: &models.XYAxisModel{
								X: &models.XYAxisConfigModel{
									Title: &models.AxisTitleModel{
										Value:   types.StringValue("Timestamp"),
										Visible: types.BoolValue(true),
									},
									Grid:             types.BoolNull(),
									Ticks:            types.BoolNull(),
									LabelOrientation: types.StringNull(),
									Scale:            types.StringNull(),
									DomainJSON:       jsontypes.NewNormalizedNull(),
								},
								Y: &models.YAxisConfigModel{
									Title: &models.AxisTitleModel{
										Value:   types.StringValue("Count"),
										Visible: types.BoolValue(true),
									},
									Grid:             types.BoolNull(),
									Ticks:            types.BoolNull(),
									LabelOrientation: types.StringNull(),
									Scale:            types.StringValue("linear"),
									DomainJSON:       jsontypes.NewNormalizedValue(`{"type":"fit"}`),
								},
								Y2: &models.YAxisConfigModel{
									Title: &models.AxisTitleModel{
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
							Decorations: &models.XYDecorationsModel{
								ShowEndZones:          types.BoolNull(),
								ShowCurrentTimeMarker: types.BoolNull(),
								PointVisibility:       types.StringNull(),
								LineInterpolation:     types.StringNull(),
								FillOpacity:           types.Float64Value(0.3),
							},
							Legend: &models.XYLegendModel{
								Visibility:         types.StringValue("visible"),
								Inside:             types.BoolValue(false),
								Position:           types.StringValue("right"),
								TruncateAfterLines: types.Int64Null(),
							},
							Layers: []models.XYLayerModel{
								{
									Type: types.StringValue("line"),
									DataLayer: &models.DataLayerModel{
										DataSourceJSON: jsontypes.NewNormalizedValue(`{"type":"data_view_spec","index_pattern":"metrics-*"}`),
										XJSON:          jsontypes.NewNormalizedValue(`{"column":"@timestamp","format":{"type":"number"}}`),
										BreakdownByJSON: jsontypes.NewNormalizedValue(
											`{"column":"host.name","collapse_by":"avg","format":{"type":"number"},` +
												`"color":{"mode":"categorical","palette":"default","mapping":[],` +
												`"unassigned":{"type":"color_code","value":"#D3DAE6"}}}`,
										),
										Y: []models.YMetricModel{
											{ConfigJSON: jsontypes.NewNormalizedValue(`{"operation":"count","empty_as_null":true,"format":{"type":"number"}}`)},
										},
									},
								},
								{
									Type: types.StringValue("reference_lines"),
									ReferenceLineLayer: &models.ReferenceLineLayerModel{
										DataSourceJSON: jsontypes.NewNormalizedValue(`{"type":"data_view_spec","index_pattern":"metrics-*"}`),
										Thresholds: []models.ThresholdModel{
											{
												ValueJSON: jsontypes.NewNormalizedValue(`{"operation":"static_value","value":42,"label":"","format":{"type":"number"}}`),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	statePanels := []models.PanelModel{
		{
			VisConfig: &models.VisConfigModel{
				ByValue: &models.VisByValueModel{
					LensByValueChartBlocks: models.LensByValueChartBlocks{
						XYChartConfig: &models.XYChartConfigModel{
							Title: types.StringValue(""),
							Axis: &models.XYAxisModel{
								X: &models.XYAxisConfigModel{
									Title:            &models.AxisTitleModel{},
									Grid:             types.BoolValue(true),
									Ticks:            types.BoolValue(true),
									LabelOrientation: types.StringValue("horizontal"),
									Scale:            types.StringValue("ordinal"),
									DomainJSON:       jsontypes.NewNormalizedValue(`{"type":"fit","rounding":false}`),
								},
								Y: &models.YAxisConfigModel{
									Title: &models.AxisTitleModel{
										Value:   types.StringValue("Count"),
										Visible: types.BoolValue(true),
									},
									Grid:             types.BoolValue(true),
									Ticks:            types.BoolValue(true),
									LabelOrientation: types.StringValue("horizontal"),
									Scale:            types.StringValue("linear"),
									DomainJSON:       jsontypes.NewNormalizedValue(`{"type":"fit","rounding":true}`),
								},
								Y2: nil,
							},
							Decorations: &models.XYDecorationsModel{
								ShowEndZones:          types.BoolValue(false),
								ShowCurrentTimeMarker: types.BoolValue(false),
								PointVisibility:       types.StringValue("auto"),
								LineInterpolation:     types.StringValue("linear"),
								FillOpacity:           types.Float64Null(),
							},
							Legend: &models.XYLegendModel{
								Visibility:         types.StringValue("visible"),
								Inside:             types.BoolValue(false),
								Position:           types.StringValue("right"),
								TruncateAfterLines: types.Int64Value(1),
							},
							Layers: []models.XYLayerModel{
								{
									Type: types.StringValue("line"),
									DataLayer: &models.DataLayerModel{
										DataSourceJSON: jsontypes.NewNormalizedValue(`{"type":"data_view_spec","index_pattern":"metrics-*","time_field":"@timestamp"}`),
										XJSON:          jsontypes.NewNormalizedValue(`{"column":"@timestamp","format":{"type":"number","decimals":2,"compact":false}}`),
										BreakdownByJSON: jsontypes.NewNormalizedValue(
											`{"column":"host.name","collapse_by":"avg",` +
												`"format":{"type":"number","decimals":2,"compact":false},` +
												`"color":{"mode":"categorical","palette":"default","mapping":[],` +
												`"unassigned":{"type":"color_code","value":"#D3DAE6"}}}`,
										),
										Y: []models.YMetricModel{
											{ConfigJSON: jsontypes.NewNormalizedValue(`{"operation":"count","empty_as_null":true,"format":{"type":"number","decimals":2,"compact":false},"axis_id":"y"}`)},
										},
									},
								},
								{
									Type: types.StringValue("reference_lines"),
									ReferenceLineLayer: &models.ReferenceLineLayerModel{
										DataSourceJSON: jsontypes.NewNormalizedValue(`{"type":"data_view_spec","index_pattern":"metrics-*","time_field":"@timestamp"}`),
										Thresholds: []models.ThresholdModel{
											{
												ValueJSON: jsontypes.NewNormalizedValue(`{"operation":"static_value","value":42,"label":"","format":{"type":"number","decimals":2,"compact":false},"axis_id":"y"}`),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	lensxy.AlignXYChartStateFromPlanPanels(planPanels, statePanels)

	planXY := planPanels[0].VisConfig.ByValue.XYChartConfig
	got := statePanels[0].VisConfig.ByValue.XYChartConfig
	require.NotNil(t, got)
	assert.Equal(t, types.StringValue("Sample XY Chart"), got.Title)
	assert.True(t, got.Axis.X.Scale.IsNull())
	assert.True(t, got.Axis.X.Grid.IsNull())
	assert.True(t, got.Axis.X.Ticks.IsNull())
	assert.True(t, got.Axis.X.LabelOrientation.IsNull())
	assert.True(t, got.Axis.X.DomainJSON.IsNull())
	require.NotNil(t, got.Axis.Y2)
	assert.Equal(t, planXY.Axis.Y2.DomainJSON.ValueString(), got.Axis.Y2.DomainJSON.ValueString())
	assert.Equal(t, planXY.Axis.Y.DomainJSON.ValueString(), got.Axis.Y.DomainJSON.ValueString())
	assert.True(t, got.Decorations.ShowEndZones.IsNull())
	assert.True(t, got.Decorations.ShowCurrentTimeMarker.IsNull())
	assert.True(t, got.Decorations.PointVisibility.IsNull())
	assert.True(t, got.Decorations.LineInterpolation.IsNull())
	assert.Equal(t, types.Float64Value(0.3), got.Decorations.FillOpacity)
	assert.True(t, got.Legend.TruncateAfterLines.IsNull())
	assert.Equal(t, planXY.Layers[0].DataLayer.DataSourceJSON.ValueString(), got.Layers[0].DataLayer.DataSourceJSON.ValueString())
	assert.Equal(t, planXY.Layers[0].DataLayer.XJSON.ValueString(), got.Layers[0].DataLayer.XJSON.ValueString())
	assert.Equal(t, planXY.Layers[0].DataLayer.BreakdownByJSON.ValueString(), got.Layers[0].DataLayer.BreakdownByJSON.ValueString())
	assert.Equal(t, planXY.Layers[0].DataLayer.Y[0].ConfigJSON.ValueString(), got.Layers[0].DataLayer.Y[0].ConfigJSON.ValueString())
	assert.Equal(t, planXY.Layers[1].ReferenceLineLayer.DataSourceJSON.ValueString(), got.Layers[1].ReferenceLineLayer.DataSourceJSON.ValueString())
	assert.Equal(t, planXY.Layers[1].ReferenceLineLayer.Thresholds[0].ValueJSON.ValueString(), got.Layers[1].ReferenceLineLayer.Thresholds[0].ValueJSON.ValueString())
}

func Test_filterSimpleModel_toAPI_nil(t *testing.T) {
	var model *models.FilterSimpleModel
	apiQuery := filterSimpleToAPI(model)
	assert.NotNil(t, apiQuery) // Returns empty struct, not nil
}

func Test_xyAxisModel_toAPI_nil(t *testing.T) {
	var model *models.XYAxisModel
	apiAxis, diags := xyAxisToAPI(model)
	assert.False(t, diags.HasError())
	assert.NotNil(t, apiAxis) // Returns empty struct, not nil
}

func Test_axisTitleIsDefault(t *testing.T) {
	tests := []struct {
		name   string
		title  *models.AxisTitleModel
		expect bool
	}{
		{"nil", nil, true},
		{"value known", &models.AxisTitleModel{Value: types.StringValue("x")}, false},
		{"visible true", &models.AxisTitleModel{Visible: types.BoolValue(true)}, true},
		{"visible false", &models.AxisTitleModel{Visible: types.BoolValue(false)}, false},
		{"empty", &models.AxisTitleModel{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := axisTitleIsDefault(tt.title)
			assert.Equal(t, tt.expect, got)
		})
	}
}
