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

package lenscommon

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func alignPartitionValueDisplayForTest(t *testing.T, plan, state *models.PartitionValueDisplay) *models.PartitionValueDisplay {
	t.Helper()

	title := types.StringNull()
	desc := types.StringNull()
	ds := jsontypes.NewNormalizedNull()
	ignore := types.BoolNull()
	sampling := types.Float64Null()
	var groupBy customtypes.JSONWithDefaultsValue[[]map[string]any]
	var metrics customtypes.JSONWithDefaultsValue[[]map[string]any]
	stateVD := state

	AlignStandardPartitionChartStateFromPlan(
		t.Context(),
		types.StringNull(), types.StringNull(),
		jsontypes.NewNormalizedNull(),
		types.BoolNull(), types.Float64Null(),
		groupBy, metrics,
		nil, plan,
		&title, &desc, &ds, &ignore, &sampling,
		&groupBy, &metrics,
		nil, &stateVD,
	)
	return stateVD
}

func TestPartitionValueDisplayMatchesKibanaDefault_percentageModeWithPercentDecimalsTwo(t *testing.T) {
	t.Parallel()

	state := &models.PartitionValueDisplay{
		Mode:            types.StringValue("percentage"),
		PercentDecimals: types.Float64Value(2),
	}

	assert.True(t, PartitionValueDisplayMatchesKibanaDefault(state))
}

func TestPartitionValueDisplayMatchesKibanaDefault_percentageModeWithNullPercentDecimals(t *testing.T) {
	t.Parallel()

	state := &models.PartitionValueDisplay{
		Mode:            types.StringValue("percentage"),
		PercentDecimals: types.Float64Null(),
	}

	assert.True(t, PartitionValueDisplayMatchesKibanaDefault(state))
}

func TestAlignStandardPartitionChartStateFromPlan_preservesNullPercentDecimalsWhenStateIsTwo(t *testing.T) {
	t.Parallel()

	plan := &models.PartitionValueDisplay{
		Mode:            types.StringValue("absolute"),
		PercentDecimals: types.Float64Null(),
	}
	state := &models.PartitionValueDisplay{
		Mode:            types.StringValue("absolute"),
		PercentDecimals: types.Float64Value(2),
	}

	got := alignPartitionValueDisplayForTest(t, plan, state)

	require.NotNil(t, got)
	assert.Equal(t, "absolute", got.Mode.ValueString())
	assert.True(t, got.PercentDecimals.IsNull())
}

func TestPartitionValueDisplayMatchesKibanaDefault_nonDefaultPercentDecimalsIsNotDefault(t *testing.T) {
	t.Parallel()

	state := &models.PartitionValueDisplay{
		Mode:            types.StringValue("percentage"),
		PercentDecimals: types.Float64Value(5),
	}

	assert.False(t, PartitionValueDisplayMatchesKibanaDefault(state))
}

func TestAlignStandardPartitionChartStateFromPlan_dropsInjectedDefaultValueDisplayBlock(t *testing.T) {
	t.Parallel()

	state := &models.PartitionValueDisplay{
		Mode:            types.StringValue("percentage"),
		PercentDecimals: types.Float64Value(2),
	}

	got := alignPartitionValueDisplayForTest(t, nil, state)

	assert.Nil(t, got)
}

func TestAlignStandardPartitionChartStateFromPlan_keepsExplicitPercentDecimalsTwo(t *testing.T) {
	t.Parallel()

	plan := &models.PartitionValueDisplay{
		Mode:            types.StringValue("percentage"),
		PercentDecimals: types.Float64Value(2),
	}
	state := &models.PartitionValueDisplay{
		Mode:            types.StringValue("percentage"),
		PercentDecimals: types.Float64Value(2),
	}

	got := alignPartitionValueDisplayForTest(t, plan, state)

	require.NotNil(t, got)
	assert.InDelta(t, float64(2), got.PercentDecimals.ValueFloat64(), 1e-9)
}

func TestAlignStandardPartitionChartStateFromPlan_doesNotDropNonDefaultPercentDecimals(t *testing.T) {
	t.Parallel()

	state := &models.PartitionValueDisplay{
		Mode:            types.StringValue("percentage"),
		PercentDecimals: types.Float64Value(5),
	}

	got := alignPartitionValueDisplayForTest(t, nil, state)

	require.NotNil(t, got)
	assert.InDelta(t, float64(5), got.PercentDecimals.ValueFloat64(), 1e-9)
}

func TestAlignStandardPartitionChartStateFromPlan_dropsInjectedNullPercentDecimalsDefaultBlock(t *testing.T) {
	t.Parallel()

	state := &models.PartitionValueDisplay{
		Mode:            types.StringValue("percentage"),
		PercentDecimals: types.Float64Null(),
	}

	got := alignPartitionValueDisplayForTest(t, nil, state)

	assert.Nil(t, got)
}

func TestAlignPartitionLegendStateFromPlan_preservesNullTruncateAndNestedWhenKibanaInjectsDefaults(t *testing.T) {
	t.Parallel()

	plan := &models.PartitionLegendModel{
		Visible:           types.StringNull(),
		TruncateAfterLine: types.Int64Null(),
		Nested:            types.BoolNull(),
	}
	state := &models.PartitionLegendModel{
		Visible:           types.StringValue("auto"),
		TruncateAfterLine: types.Int64Value(1),
		Nested:            types.BoolValue(false),
	}

	AlignPartitionLegendStateFromPlan(plan, state)

	assert.True(t, state.Visible.IsNull())
	assert.True(t, state.TruncateAfterLine.IsNull())
	assert.True(t, state.Nested.IsNull())
}

func TestAlignPartitionLegendStateFromPlan_doesNotOverwriteExplicitTruncateAndNested(t *testing.T) {
	t.Parallel()

	plan := &models.PartitionLegendModel{
		Visible:           types.StringValue("visible"),
		TruncateAfterLine: types.Int64Value(10),
		Nested:            types.BoolValue(true),
	}
	state := &models.PartitionLegendModel{
		Visible:           types.StringValue("visible"),
		TruncateAfterLine: types.Int64Value(10),
		Nested:            types.BoolValue(true),
	}

	AlignPartitionLegendStateFromPlan(plan, state)

	assert.Equal(t, "visible", state.Visible.ValueString())
	assert.Equal(t, int64(10), state.TruncateAfterLine.ValueInt64())
	assert.True(t, state.Nested.ValueBool())
}
