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

package lensheatmap

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_alignHeatmapLegendStateFromPlan(t *testing.T) {
	t.Run("clones plan legend when state is nil", func(t *testing.T) {
		t.Parallel()

		plan := &models.HeatmapLegendModel{
			Visibility: types.StringValue("visible"),
			Size:       types.StringValue("m"),
		}
		var state *models.HeatmapLegendModel

		alignHeatmapLegendStateFromPlan(plan, &state)

		require.NotNil(t, state)
		assert.Equal(t, "visible", state.Visibility.ValueString())
		assert.Equal(t, "m", state.Size.ValueString())
	})

	t.Run("preserves known state when plan and state are set", func(t *testing.T) {
		t.Parallel()

		plan := &models.HeatmapLegendModel{
			Visibility: types.StringValue("visible"),
			Size:       types.StringValue("l"),
		}
		state := &models.HeatmapLegendModel{
			Visibility: types.StringValue("hidden"),
			Size:       types.StringValue("m"),
		}

		alignHeatmapLegendStateFromPlan(plan, &state)

		assert.Equal(t, "hidden", state.Visibility.ValueString())
		assert.Equal(t, "m", state.Size.ValueString())
	})

	t.Run("no-op when plan and state are nil", func(t *testing.T) {
		t.Parallel()

		var state *models.HeatmapLegendModel

		alignHeatmapLegendStateFromPlan(nil, &state)

		assert.Nil(t, state)
	})
}
