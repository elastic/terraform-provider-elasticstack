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

package lensxy

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAlignXYLegendStateFromPlan(t *testing.T) {
	t.Run("clones plan legend when state is nil", func(t *testing.T) {
		plan := &models.XYChartConfigModel{
			Legend: &models.XYLegendModel{
				Visibility: types.StringValue("visible"),
				Position:   types.StringValue("right"),
			},
		}
		state := &models.XYChartConfigModel{Legend: nil}

		alignXYChartStateFromPlan(plan, state)

		require.NotNil(t, state.Legend)
		assert.Equal(t, "visible", state.Legend.Visibility.ValueString())
		assert.Equal(t, "right", state.Legend.Position.ValueString())
	})

	t.Run("Visibility copies plan when state field is null", func(t *testing.T) {
		plan := &models.XYLegendModel{Visibility: types.StringValue("visible")}
		state := &models.XYLegendModel{Visibility: types.StringNull()}

		alignXYLegendStateFromPlan(plan, state)

		assert.Equal(t, "visible", state.Visibility.ValueString())
	})

	t.Run("clones plan legend when state legend is nil after Kibana omits block", func(t *testing.T) {
		plan := &models.XYChartConfigModel{
			Legend: &models.XYLegendModel{
				Visibility: types.StringValue("visible"),
				Position:   types.StringValue("right"),
				Size:       types.StringValue("m"),
				Inside:     types.BoolValue(false),
			},
		}
		state := &models.XYChartConfigModel{Legend: nil}

		alignXYChartStateFromPlan(plan, state)

		require.NotNil(t, state.Legend)
		assert.Equal(t, "m", state.Legend.Size.ValueString())
	})
}

func TestAlignXYAxisStateFromPlan(t *testing.T) {
	t.Run("Y2 axis is aligned with the same rules as the primary Y axis", func(t *testing.T) {
		plan := &models.XYAxisModel{
			Y2: &models.YAxisConfigModel{
				LabelOrientation: types.StringValue("vertical"),
				Ticks:            types.BoolValue(false),
			},
		}
		state := &models.XYAxisModel{
			Y2: &models.YAxisConfigModel{
				LabelOrientation: types.StringNull(),
				Ticks:            types.BoolNull(),
			},
		}

		alignXYAxisStateFromPlan(plan, state)

		require.NotNil(t, state.Y2)
		assert.Equal(t, "vertical", state.Y2.LabelOrientation.ValueString())
		assert.False(t, state.Y2.Ticks.ValueBool())
	})

	t.Run("Y2 clones plan when state omits the block", func(t *testing.T) {
		plan := &models.XYAxisModel{
			Y2: &models.YAxisConfigModel{LabelOrientation: types.StringValue("vertical")},
		}
		state := &models.XYAxisModel{Y2: nil}

		alignXYAxisStateFromPlan(plan, state)

		require.NotNil(t, state.Y2)
		assert.Equal(t, "vertical", state.Y2.LabelOrientation.ValueString())
	})
}

func TestAlignXYFittingStateFromPlan(t *testing.T) {
	t.Run("Type copies plan when state is null", func(t *testing.T) {
		plan := &models.XYFittingModel{Type: types.StringValue("none")}
		state := &models.XYFittingModel{Type: types.StringNull()}

		alignXYFittingStateFromPlan(plan, state)

		assert.Equal(t, "none", state.Type.ValueString())
	})

	t.Run("Type leaves state unchanged when state is already known", func(t *testing.T) {
		plan := &models.XYFittingModel{Type: types.StringValue("none")}
		state := &models.XYFittingModel{Type: types.StringValue("linear")}

		alignXYFittingStateFromPlan(plan, state)

		assert.Equal(t, "linear", state.Type.ValueString())
	})

	t.Run("both nil pointers do not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			alignXYFittingStateFromPlan(nil, nil)
		})
	})

	t.Run("plan nil leaves state unchanged", func(t *testing.T) {
		state := &models.XYFittingModel{Type: types.StringValue("linear")}

		alignXYFittingStateFromPlan(nil, state)

		assert.Equal(t, "linear", state.Type.ValueString())
	})

	t.Run("state nil does not panic", func(t *testing.T) {
		plan := &models.XYFittingModel{Type: types.StringValue("none")}

		assert.NotPanics(t, func() {
			alignXYFittingStateFromPlan(plan, nil)
		})
	})

	t.Run("Dotted copies plan when state is null", func(t *testing.T) {
		plan := &models.XYFittingModel{Dotted: types.BoolValue(true)}
		state := &models.XYFittingModel{Dotted: types.BoolNull()}

		alignXYFittingStateFromPlan(plan, state)

		assert.True(t, state.Dotted.ValueBool())
	})

	t.Run("EndValue copies plan when state is null", func(t *testing.T) {
		plan := &models.XYFittingModel{EndValue: types.StringValue("nearest")}
		state := &models.XYFittingModel{EndValue: types.StringNull()}

		alignXYFittingStateFromPlan(plan, state)

		assert.Equal(t, "nearest", state.EndValue.ValueString())
	})
}

func TestAlignXYLayerStateFromPlan_omittedAxisPreservesPlanConfigJSON(t *testing.T) {
	t.Parallel()

	planJSON := `{"empty_as_null":true,"operation":"count"}`
	stateJSON := `{"empty_as_null":true,"operation":"count","axis":"y","color":{"type":"auto"}}`

	plan := &models.XYChartConfigModel{
		Layers: []models.XYLayerModel{{
			DataLayer: &models.DataLayerModel{
				Y: []models.YMetricModel{{
					ConfigJSON: jsontypes.NewNormalizedValue(planJSON),
				}},
			},
		}},
	}
	state := &models.XYChartConfigModel{
		Layers: []models.XYLayerModel{{
			DataLayer: &models.DataLayerModel{
				Y: []models.YMetricModel{{
					ConfigJSON: jsontypes.NewNormalizedValue(stateJSON),
				}},
			},
		}},
	}

	alignXYChartStateFromPlan(plan, state)

	assert.JSONEq(t, planJSON, state.Layers[0].DataLayer.Y[0].ConfigJSON.ValueString())
}

func TestAlignXYLayerStateFromPlan_explicitY2AxisIsNotOverridden(t *testing.T) {
	t.Parallel()

	planJSON := `{"axis":"y2","empty_as_null":true,"operation":"count"}`
	stateJSON := `{"empty_as_null":true,"operation":"count","axis":"y2","color":{"type":"auto"}}`

	plan := &models.XYChartConfigModel{
		Layers: []models.XYLayerModel{{
			DataLayer: &models.DataLayerModel{
				Y: []models.YMetricModel{{
					ConfigJSON: jsontypes.NewNormalizedValue(planJSON),
				}},
			},
		}},
	}
	state := &models.XYChartConfigModel{
		Layers: []models.XYLayerModel{{
			DataLayer: &models.DataLayerModel{
				Y: []models.YMetricModel{{
					ConfigJSON: jsontypes.NewNormalizedValue(stateJSON),
				}},
			},
		}},
	}

	alignXYChartStateFromPlan(plan, state)

	assert.JSONEq(t, planJSON, state.Layers[0].DataLayer.Y[0].ConfigJSON.ValueString())
}

func TestAlignXYLayerStateFromPlan_countEmptyAsNullGatingComposesWithAxis(t *testing.T) {
	t.Parallel()

	planJSON := `{"operation":"count"}`
	stateJSON := `{"operation":"count","empty_as_null":false,"axis":"y","color":{"type":"auto"}}`

	plan := &models.XYChartConfigModel{
		Layers: []models.XYLayerModel{{
			DataLayer: &models.DataLayerModel{
				Y: []models.YMetricModel{{
					ConfigJSON: jsontypes.NewNormalizedValue(planJSON),
				}},
			},
		}},
	}
	state := &models.XYChartConfigModel{
		Layers: []models.XYLayerModel{{
			DataLayer: &models.DataLayerModel{
				Y: []models.YMetricModel{{
					ConfigJSON: jsontypes.NewNormalizedValue(stateJSON),
				}},
			},
		}},
	}

	alignXYChartStateFromPlan(plan, state)

	assert.JSONEq(t, planJSON, state.Layers[0].DataLayer.Y[0].ConfigJSON.ValueString())
}

func TestAlignXYLayerStateFromPlan_staticYColorIsNotNormalizedToAuto(t *testing.T) {
	t.Parallel()

	planJSON := `{"column":"system.cpu.user.pct","color":{"type":"static","color":"#54B399"},"format":{"type":"number"}}`
	stateJSON := `{"column":"system.cpu.user.pct","axis":"y","color":{"type":"auto"},"format":{"compact":false,"decimals":2,"type":"number"}}`

	planLayers := []models.XYLayerModel{{
		DataLayer: &models.DataLayerModel{
			Y: []models.YMetricModel{{
				ConfigJSON: jsontypes.NewNormalizedValue(planJSON),
			}},
		},
	}}
	stateLayers := []models.XYLayerModel{{
		DataLayer: &models.DataLayerModel{
			Y: []models.YMetricModel{{
				ConfigJSON: jsontypes.NewNormalizedValue(stateJSON),
			}},
		},
	}}

	alignXYLayerStateFromPlan(planLayers, stateLayers)

	assert.JSONEq(t, stateJSON, stateLayers[0].DataLayer.Y[0].ConfigJSON.ValueString())
}

func TestAlignXYLayerStateFromPlan_omittedReferenceLineThresholdDefaults(t *testing.T) {
	t.Parallel()

	planValue := `{"format":{"compact":false,"decimals":2,"type":"number"},"label":"","operation":"static_value","value":42}`
	plan := &models.XYChartConfigModel{
		Layers: []models.XYLayerModel{{
			ReferenceLineLayer: &models.ReferenceLineLayerModel{
				Thresholds: []models.ThresholdModel{{
					Axis:      types.StringNull(),
					Operation: types.StringNull(),
					ColorJSON: jsontypes.NewNormalizedNull(),
					ValueJSON: jsontypes.NewNormalizedValue(planValue),
				}},
			},
		}},
	}
	state := &models.XYChartConfigModel{
		Layers: []models.XYLayerModel{{
			ReferenceLineLayer: &models.ReferenceLineLayerModel{
				Thresholds: []models.ThresholdModel{{
					Axis:      types.StringValue("y"),
					Operation: types.StringValue("static_value"),
					ColorJSON: jsontypes.NewNormalizedValue(`{"type":"auto"}`),
					ValueJSON: jsontypes.NewNormalizedValue(`42`),
				}},
			},
		}},
	}

	alignXYChartStateFromPlan(plan, state)

	got := state.Layers[0].ReferenceLineLayer.Thresholds[0]
	assert.True(t, got.Axis.IsNull())
	assert.True(t, got.Operation.IsNull())
	assert.True(t, got.ColorJSON.IsNull())
	assert.JSONEq(t, planValue, got.ValueJSON.ValueString())
}

func TestAlignXYLayerStateFromPlan_siblingStaticOperationRestoresCollapsedValueJSON(t *testing.T) {
	t.Parallel()

	planValue := `{"value":42}`
	plan := &models.XYChartConfigModel{
		Layers: []models.XYLayerModel{{
			ReferenceLineLayer: &models.ReferenceLineLayerModel{
				Thresholds: []models.ThresholdModel{{
					Operation: types.StringValue("static_value"),
					ValueJSON: jsontypes.NewNormalizedValue(planValue),
				}},
			},
		}},
	}
	state := &models.XYChartConfigModel{
		Layers: []models.XYLayerModel{{
			ReferenceLineLayer: &models.ReferenceLineLayerModel{
				Thresholds: []models.ThresholdModel{{
					Operation: types.StringValue("static_value"),
					ValueJSON: jsontypes.NewNormalizedValue(`42`),
				}},
			},
		}},
	}

	alignXYChartStateFromPlan(plan, state)

	assert.JSONEq(t, planValue, state.Layers[0].ReferenceLineLayer.Thresholds[0].ValueJSON.ValueString())
}

func TestAlignXYLayerStateFromPlan_explicitReferenceLineThresholdAxisIsNotOverridden(t *testing.T) {
	t.Parallel()

	plan := &models.XYChartConfigModel{
		Layers: []models.XYLayerModel{{
			ReferenceLineLayer: &models.ReferenceLineLayerModel{
				Thresholds: []models.ThresholdModel{{
					Axis: types.StringValue("y2"),
				}},
			},
		}},
	}
	state := &models.XYChartConfigModel{
		Layers: []models.XYLayerModel{{
			ReferenceLineLayer: &models.ReferenceLineLayerModel{
				Thresholds: []models.ThresholdModel{{
					Axis: types.StringValue("y2"),
				}},
			},
		}},
	}

	alignXYChartStateFromPlan(plan, state)

	assert.Equal(t, "y2", state.Layers[0].ReferenceLineLayer.Thresholds[0].Axis.ValueString())
}

func TestAlignXYLayerStateFromPlan_explicitLeftThresholdAxisIsNotTreatedAsOmitDefaultY(t *testing.T) {
	t.Parallel()

	plan := &models.XYChartConfigModel{
		Layers: []models.XYLayerModel{{
			ReferenceLineLayer: &models.ReferenceLineLayerModel{
				Thresholds: []models.ThresholdModel{{
					Axis: types.StringValue("left"),
				}},
			},
		}},
	}
	state := &models.XYChartConfigModel{
		Layers: []models.XYLayerModel{{
			ReferenceLineLayer: &models.ReferenceLineLayerModel{
				Thresholds: []models.ThresholdModel{{
					Axis: types.StringValue("y"),
				}},
			},
		}},
	}

	alignXYChartStateFromPlan(plan, state)

	assert.Equal(t, "y", state.Layers[0].ReferenceLineLayer.Thresholds[0].Axis.ValueString())
	assert.False(t, state.Layers[0].ReferenceLineLayer.Thresholds[0].Axis.IsNull())
}

func TestPreserveThresholdValueJSONIfStateIsPlanValue_planWithoutValueDoesNotRestoreOnNullState(t *testing.T) {
	t.Parallel()

	planJSON := `{"operation":"static_value","label":""}`
	plan := jsontypes.NewNormalizedValue(planJSON)
	state := jsontypes.NewNormalizedValue(`null`)

	preserveThresholdValueJSONIfStateIsPlanValue(plan, &state, types.StringNull())

	assert.JSONEq(t, `null`, state.ValueString())
}

func TestPreserveThresholdValueJSONIfStateIsPlanValue_nonStaticValueWithValueKeyDoesNotRestore(t *testing.T) {
	t.Parallel()

	planJSON := `{"operation":"formula","value":42,"label":""}`
	plan := jsontypes.NewNormalizedValue(planJSON)
	state := jsontypes.NewNormalizedValue(`42`)

	preserveThresholdValueJSONIfStateIsPlanValue(plan, &state, types.StringNull())

	assert.JSONEq(t, `42`, state.ValueString())
}
