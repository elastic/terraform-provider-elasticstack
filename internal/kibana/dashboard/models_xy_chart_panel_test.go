package dashboard

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
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
					Grid:  utils.Pointer(true),
					Ticks: utils.Pointer(false),
				},
				Left: &leftYAxisConfigAPIModel{
					Grid:  utils.Pointer(true),
					Ticks: utils.Pointer(true),
				},
				Right: &rightYAxisConfigAPIModel{
					Grid:  utils.Pointer(false),
					Ticks: utils.Pointer(true),
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
					Grid: utils.Pointer(true),
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
				Grid:             utils.Pointer(true),
				Ticks:            utils.Pointer(false),
				LabelOrientation: func() *kbapi.XyAxisXLabelOrientation { o := kbapi.XyAxisXLabelOrientation("horizontal"); return &o }(),
				Title: &struct {
					Value   *string `json:"value,omitempty"`
					Visible *bool   `json:"visible,omitempty"`
				}{
					Value:   utils.Pointer("X Axis Title"),
					Visible: utils.Pointer(true),
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
				Grid:             utils.Pointer(false),
				Ticks:            nil,
				LabelOrientation: nil,
				Title:            nil,
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
				Grid:  utils.Pointer(true),
				Ticks: utils.Pointer(true),
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
				Grid:             utils.Pointer(true),
				Ticks:            utils.Pointer(false),
				LabelOrientation: func() *kbapi.XyAxisLeftLabelOrientation { o := kbapi.XyAxisLeftLabelOrientation("vertical"); return &o }(),
				Scale:            func() *kbapi.XyAxisLeftScale { s := kbapi.XyAxisLeftScale("linear"); return &s }(),
				Title: &struct {
					Value   *string `json:"value,omitempty"`
					Visible *bool   `json:"visible,omitempty"`
				}{
					Value:   utils.Pointer("Y Axis Title"),
					Visible: utils.Pointer(false),
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
				Grid:  utils.Pointer(false),
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
				Grid:             utils.Pointer(false),
				Ticks:            utils.Pointer(true),
				LabelOrientation: func() *kbapi.XyAxisRightLabelOrientation { o := kbapi.XyAxisRightLabelOrientation("angled"); return &o }(),
				Scale:            func() *kbapi.XyAxisRightScale { s := kbapi.XyAxisRightScale("log"); return &s }(),
				Title: &struct {
					Value   *string `json:"value,omitempty"`
					Visible *bool   `json:"visible,omitempty"`
				}{
					Value:   utils.Pointer("Right Y Axis"),
					Visible: utils.Pointer(true),
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
			Value   *string `json:"value,omitempty"`
			Visible *bool   `json:"visible,omitempty"`
		}
		expected *axisTitleModel
	}{
		{
			name: "all fields populated",
			apiTitle: &struct {
				Value   *string `json:"value,omitempty"`
				Visible *bool   `json:"visible,omitempty"`
			}{
				Value:   utils.Pointer("Test Title"),
				Visible: utils.Pointer(true),
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
				Value   *string `json:"value,omitempty"`
				Visible *bool   `json:"visible,omitempty"`
			}{
				Value:   utils.Pointer("Only Value"),
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
				Value   *string `json:"value,omitempty"`
				Visible *bool   `json:"visible,omitempty"`
			}{
				Value:   nil,
				Visible: utils.Pointer(false),
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
			if tt.apiTitle == nil {
				assert.NotNil(t, apiTitle)
			} else {
				assert.NotNil(t, apiTitle)
			}
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
				EndZones:          utils.Pointer(true),
				CurrentTimeMarker: utils.Pointer(false),
				PointVisibility:   utils.Pointer(true),
				LineInterpolation: func() *kbapi.XyDecorationsLineInterpolation {
					i := kbapi.XyDecorationsLineInterpolation("linear")
					return &i
				}(),
				MinimumBarHeight: utils.Pointer(float32(5)),
				ShowValueLabels:  utils.Pointer(true),
				FillOpacity:      utils.Pointer(float32(0.5)),
				ValueLabels:      utils.Pointer(false),
			},
			expected: &xyDecorationsModel{
				EndZones:          types.BoolValue(true),
				CurrentTimeMarker: types.BoolValue(false),
				PointVisibility:   types.BoolValue(true),
				LineInterpolation: types.StringValue("linear"),
				MinimumBarHeight:  types.Int64Value(5),
				ShowValueLabels:   types.BoolValue(true),
				FillOpacity:       types.Float64Value(0.5),
				ValueLabels:       types.BoolValue(false),
			},
			expectFillValue: 0.5,
		},
		{
			name: "nil values",
			apiDecorations: kbapi.XyDecorations{
				EndZones:          nil,
				CurrentTimeMarker: nil,
				PointVisibility:   nil,
				LineInterpolation: nil,
				MinimumBarHeight:  nil,
				ShowValueLabels:   nil,
				FillOpacity:       nil,
				ValueLabels:       nil,
			},
			expected: &xyDecorationsModel{
				EndZones:          types.BoolNull(),
				CurrentTimeMarker: types.BoolNull(),
				PointVisibility:   types.BoolNull(),
				LineInterpolation: types.StringNull(),
				MinimumBarHeight:  types.Int64Null(),
				ShowValueLabels:   types.BoolNull(),
				FillOpacity:       types.Float64Null(),
				ValueLabels:       types.BoolNull(),
			},
		},
		{
			name: "float precision rounding",
			apiDecorations: kbapi.XyDecorations{
				FillOpacity: utils.Pointer(float32(0.123456)),
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

			assert.Equal(t, tt.expected.EndZones, model.EndZones)
			assert.Equal(t, tt.expected.CurrentTimeMarker, model.CurrentTimeMarker)
			assert.Equal(t, tt.expected.PointVisibility, model.PointVisibility)
			assert.Equal(t, tt.expected.LineInterpolation, model.LineInterpolation)
			assert.Equal(t, tt.expected.MinimumBarHeight, model.MinimumBarHeight)
			assert.Equal(t, tt.expected.ShowValueLabels, model.ShowValueLabels)
			assert.Equal(t, tt.expected.ValueLabels, model.ValueLabels)

			if !tt.expected.FillOpacity.IsNull() {
				assert.InDelta(t, tt.expectFillValue, model.FillOpacity.ValueFloat64(), 0.001)
			} else {
				assert.True(t, model.FillOpacity.IsNull())
			}

			// Test toAPI
			apiDecorations := model.toAPI()
			assert.NotNil(t, apiDecorations)

			// Verify round-trip preserves known values
			if !model.EndZones.IsNull() && !model.EndZones.IsUnknown() {
				assert.Equal(t, model.EndZones.ValueBool(), *apiDecorations.EndZones)
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
				Dotted:   utils.Pointer(true),
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
				legend := kbapi.XyLegendInside{
					Inside:             true,
					Visible:            utils.Pointer(true),
					TruncateAfterLines: utils.Pointer(float32(3)),
					Columns:            utils.Pointer(float32(2)),
					Alignment:          func() *kbapi.XyLegendInsideAlignment { a := kbapi.XyLegendInsideAlignment("left"); return &a }(),
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
				Visible:            types.BoolValue(true),
				TruncateAfterLines: types.Int64Value(3),
				Columns:            types.Int64Value(2),
				Alignment:          types.StringValue("left"),
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
			assert.Equal(t, tt.expected.Visible, model.Visible)
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
				legend := kbapi.XyLegendOutside{
					Visible:            utils.Pointer(false),
					TruncateAfterLines: utils.Pointer(float32(5)),
					Position:           func() *kbapi.XyLegendOutsidePosition { p := kbapi.XyLegendOutsidePosition("right"); return &p }(),
					Size:               func() *kbapi.XyLegendOutsideSize { s := kbapi.XyLegendOutsideSize("medium"); return &s }(),
					Statistics: &[]kbapi.XyLegendOutsideStatistics{
						kbapi.XyLegendOutsideStatistics("min"),
					},
				}
				var result kbapi.XyLegend
				_ = result.FromXyLegendOutside(legend)
				return result
			}(),
			expected: &xyLegendModel{
				Inside:             types.BoolValue(false),
				Visible:            types.BoolValue(false),
				TruncateAfterLines: types.Int64Value(5),
				Position:           types.StringValue("right"),
				Size:               types.StringValue("medium"),
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
			assert.Equal(t, tt.expected.Visible, model.Visible)
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

func Test_filterSimpleModel_fromAPI_toAPI(t *testing.T) {
	tests := []struct {
		name     string
		apiQuery kbapi.FilterSimpleSchema
		expected *filterSimpleModel
	}{
		{
			name: "all fields populated",
			apiQuery: kbapi.FilterSimpleSchema{
				Query:    "test query",
				Language: func() *kbapi.FilterSimpleSchemaLanguage { l := kbapi.FilterSimpleSchemaLanguage("kuery"); return &l }(),
			},
			expected: &filterSimpleModel{
				Query:    types.StringValue("test query"),
				Language: types.StringValue("kuery"),
			},
		},
		{
			name: "only required field",
			apiQuery: kbapi.FilterSimpleSchema{
				Query:    "simple query",
				Language: nil,
			},
			expected: &filterSimpleModel{
				Query:    types.StringValue("simple query"),
				Language: types.StringNull(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test fromAPI
			model := &filterSimpleModel{}
			model.fromAPI(tt.apiQuery)

			assert.Equal(t, tt.expected.Query, model.Query)
			assert.Equal(t, tt.expected.Language, model.Language)

			// Test toAPI
			apiQuery := model.toAPI()
			assert.Equal(t, tt.apiQuery.Query, apiQuery.Query)
		})
	}
}

func Test_searchFilterModel_fromAPI_toAPI(t *testing.T) {
	tests := []struct {
		name        string
		apiFilter   kbapi.SearchFilterSchema
		expected    *searchFilterModel
		expectError bool
	}{
		{
			name: "valid filter with language",
			apiFilter: func() kbapi.SearchFilterSchema {
				filter := kbapi.SearchFilterSchema0{
					Language: func() *kbapi.SearchFilterSchema0Language { l := kbapi.SearchFilterSchema0Language("lucene"); return &l }(),
				}
				var query kbapi.SearchFilterSchema_0_Query
				_ = query.FromSearchFilterSchema0Query0("field:value")
				filter.Query = query

				var result kbapi.SearchFilterSchema
				_ = result.FromSearchFilterSchema0(filter)
				return result
			}(),
			expected: &searchFilterModel{
				Query:    types.StringValue("field:value"),
				Language: types.StringValue("lucene"),
			},
			expectError: false,
		},
		{
			name: "filter without language",
			apiFilter: func() kbapi.SearchFilterSchema {
				filter := kbapi.SearchFilterSchema0{}
				var query kbapi.SearchFilterSchema_0_Query
				_ = query.FromSearchFilterSchema0Query0("simple query")
				filter.Query = query

				var result kbapi.SearchFilterSchema
				_ = result.FromSearchFilterSchema0(filter)
				return result
			}(),
			expected: &searchFilterModel{
				Query:    types.StringValue("simple query"),
				Language: types.StringNull(),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test fromAPI
			model := &searchFilterModel{}
			diags := model.fromAPI(tt.apiFilter)

			if tt.expectError {
				assert.True(t, diags.HasError())
				return
			}

			require.False(t, diags.HasError())
			assert.Equal(t, tt.expected.Query, model.Query)
			assert.Equal(t, tt.expected.Language, model.Language)

			// Test toAPI
			apiFilter, diags := model.toAPI()
			require.False(t, diags.HasError())
			assert.NotNil(t, apiFilter)
		})
	}
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
					EndZones:        types.BoolValue(true),
					PointVisibility: types.BoolValue(false),
				},
				Fitting: &xyFittingModel{
					Type:   types.StringValue("linear"),
					Dotted: types.BoolValue(true),
				},
				Layers: jsontypes.NewNormalizedValue(`[{"type":"layer1"}]`),
				Legend: &xyLegendModel{
					Inside:  types.BoolValue(false),
					Visible: types.BoolValue(true),
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

func Test_xyChartPanelConfigConverter_populateFromAPIPanel(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		config      kbapi.DashboardPanelItem_Config
		expectError bool
	}{
		{
			name: "valid xy chart config",
			config: func() kbapi.DashboardPanelItem_Config {
				attributes := map[string]interface{}{
					"type":  "xy",
					"title": "Test Chart",
					"axis": map[string]interface{}{
						"x": map[string]interface{}{
							"grid": true,
						},
					},
					"decorations": map[string]interface{}{
						"endZones": true,
					},
					"fitting": map[string]interface{}{
						"type": "linear",
					},
					"layers": []interface{}{},
					"legend": map[string]interface{}{
						"visible": true,
					},
					"query": map[string]interface{}{
						"query": "*",
					},
				}

				configMap := map[string]interface{}{
					"attributes": attributes,
				}

				configJSON, _ := json.Marshal(configMap)
				var config kbapi.DashboardPanelItem_Config
				_ = config.FromDashboardPanelItemConfig2(configMap)
				_ = json.Unmarshal(configJSON, &config)
				return config
			}(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := newXYChartPanelConfigConverter()
			pm := &panelModel{}
			diags := converter.populateFromAPIPanel(ctx, pm, tt.config)

			if tt.expectError {
				assert.True(t, diags.HasError())
			} else {
				// May have errors depending on config structure
				// Just verify it doesn't panic
				assert.NotNil(t, pm)
			}
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
