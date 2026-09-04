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
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type jsonDefaultsSliceTestItem struct {
	Field string `json:"field"`
}

type jsonDefaultsSliceTestModel struct {
	Config customtypes.JSONWithDefaultsValue[map[string]any]
}

func testModelConfigOf(m *jsonDefaultsSliceTestModel) *customtypes.JSONWithDefaultsValue[map[string]any] {
	return &m.Config
}

func identityDefaults(m map[string]any) map[string]any {
	return m
}

func TestPopulateJSONWithDefaultsSlice_NoPrior(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	items := []jsonDefaultsSliceTestItem{{Field: "a"}, {Field: "b"}}
	var diags diag.Diagnostics

	got := PopulateJSONWithDefaultsSlice[jsonDefaultsSliceTestItem, jsonDefaultsSliceTestModel, map[string]any](
		ctx, items, nil, testModelConfigOf, identityDefaults, "item", &diags,
	)

	require.False(t, diags.HasError())
	require.Len(t, got, 2)
	assert.JSONEq(t, `{"field":"a"}`, got[0].Config.ValueString())
	assert.JSONEq(t, `{"field":"b"}`, got[1].Config.ValueString())
}

func TestPopulateJSONWithDefaultsSlice_PreservesSemanticallyEquivalentPrior(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	items := []jsonDefaultsSliceTestItem{{Field: "a"}}

	// Prior has extra whitespace/formatting but is semantically identical once defaults are
	// applied. A freshly-marshaled current value would be compact (no spaces), so if the
	// returned config keeps prior's exact spacing, that proves prior was preserved rather than
	// the newly computed value being used.
	priorJSON := `{"field": "a"}`
	prior := []jsonDefaultsSliceTestModel{
		{Config: customtypes.NewJSONWithDefaultsValue(priorJSON, identityDefaults)},
	}
	var diags diag.Diagnostics

	got := PopulateJSONWithDefaultsSlice[jsonDefaultsSliceTestItem, jsonDefaultsSliceTestModel, map[string]any](
		ctx, items, prior, testModelConfigOf, identityDefaults, "item", &diags,
	)

	require.False(t, diags.HasError())
	require.Len(t, got, 1)
	// Exact-string check that prior was preserved verbatim, not recomputed; JSONEq would ignore
	// the whitespace difference this test relies on.
	assert.Equal(t, priorJSON, got[0].Config.ValueString()) //nolint:testifylint
}

func TestPopulateJSONWithDefaultsSlice_DifferentPriorIsReplaced(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	items := []jsonDefaultsSliceTestItem{{Field: "a"}}
	prior := []jsonDefaultsSliceTestModel{
		{Config: customtypes.NewJSONWithDefaultsValue(`{"field":"different"}`, identityDefaults)},
	}
	var diags diag.Diagnostics

	got := PopulateJSONWithDefaultsSlice[jsonDefaultsSliceTestItem, jsonDefaultsSliceTestModel, map[string]any](
		ctx, items, prior, testModelConfigOf, identityDefaults, "item", &diags,
	)

	require.False(t, diags.HasError())
	require.Len(t, got, 1)
	assert.JSONEq(t, `{"field":"a"}`, got[0].Config.ValueString())
}

func TestPopulateJSONWithDefaultsSlice_ShorterPriorLeavesExtraItemsUnpreserved(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	items := []jsonDefaultsSliceTestItem{{Field: "a"}, {Field: "b"}}
	prior := []jsonDefaultsSliceTestModel{
		{Config: customtypes.NewJSONWithDefaultsValue(`{"field":"a"}`, identityDefaults)},
	}
	var diags diag.Diagnostics

	got := PopulateJSONWithDefaultsSlice[jsonDefaultsSliceTestItem, jsonDefaultsSliceTestModel, map[string]any](
		ctx, items, prior, testModelConfigOf, identityDefaults, "item", &diags,
	)

	require.False(t, diags.HasError())
	require.Len(t, got, 2)
	assert.JSONEq(t, `{"field":"a"}`, got[0].Config.ValueString())
	assert.JSONEq(t, `{"field":"b"}`, got[1].Config.ValueString())
}

func testModelNormalizedConfigOf(m *jsonDefaultsSliceTestModel) *jsontypes.Normalized {
	return &m.Config.Normalized
}

func TestPopulateNormalizedJSONSlice_Success(t *testing.T) {
	t.Parallel()
	items := []jsonDefaultsSliceTestItem{{Field: "a"}, {Field: "b"}}
	var diags diag.Diagnostics

	got, ok := PopulateNormalizedJSONSlice[jsonDefaultsSliceTestItem, jsonDefaultsSliceTestModel](
		items, testModelNormalizedConfigOf, "item", &diags,
	)

	require.True(t, ok)
	require.False(t, diags.HasError())
	require.Len(t, got, 2)
	assert.JSONEq(t, `{"field":"a"}`, got[0].Config.ValueString())
	assert.JSONEq(t, `{"field":"b"}`, got[1].Config.ValueString())
}

func TestPopulateNormalizedJSONSlice_MarshalErrorReturnsFalse(t *testing.T) {
	t.Parallel()
	items := []float64{1, 2, 3}
	var diags diag.Diagnostics

	// math.NaN is not marshalable and should abort immediately, matching the early-return
	// behavior of the panel converters this helper replaces.
	items[1] = notANumber()

	got, ok := PopulateNormalizedJSONSlice[float64, jsonDefaultsSliceTestModel](
		items, testModelNormalizedConfigOf, "item", &diags,
	)

	assert.False(t, ok)
	assert.Nil(t, got)
	assert.True(t, diags.HasError())
}

func notANumber() float64 {
	var zero float64
	return zero / zero
}
