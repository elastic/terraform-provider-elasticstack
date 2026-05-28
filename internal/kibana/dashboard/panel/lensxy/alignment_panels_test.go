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
