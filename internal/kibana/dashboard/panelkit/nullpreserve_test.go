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

package panelkit

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

// NullPreserveFromPrior (generic core)

func TestNullPreserveFromPrior_knownPrior_leavesExistingUnchanged(t *testing.T) {
	t.Parallel()
	existing := types.StringValue("original")
	NullPreserveFromPrior(types.StringValue("prior"), &existing)
	assert.Equal(t, "original", existing.ValueString())
}

func TestNullPreserveFromPrior_nullPrior_setsExistingToNull(t *testing.T) {
	t.Parallel()
	existing := types.Int64Value(42)
	NullPreserveFromPrior(types.Int64Null(), &existing)
	assert.True(t, existing.IsNull(), "null prior should set existing to null")
	assert.False(t, existing.IsUnknown())
}

func TestNullPreserveFromPrior_unknownPrior_setsExistingToUnknown(t *testing.T) {
	t.Parallel()
	existing := types.BoolValue(true)
	NullPreserveFromPrior(types.BoolUnknown(), &existing)
	assert.True(t, existing.IsUnknown(), "unknown prior should set existing to unknown, not null")
}

func TestNullPreserveFromPrior_nilExisting_noopNoPanic(t *testing.T) {
	t.Parallel()
	assert.NotPanics(t, func() {
		NullPreserveFromPrior[types.String](types.StringNull(), nil)
	})
}

// NullPreserveSetFromPrior

func TestNullPreserveSetFromPrior_knownPrior_leavesExistingUnchanged(t *testing.T) {
	t.Parallel()
	existing := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("original")})
	prior := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("prior")})
	NullPreserveSetFromPrior(prior, &existing)
	assert.Equal(t, []attr.Value{types.StringValue("original")}, existing.Elements())
}

func TestNullPreserveSetFromPrior_nullPrior_setsExistingToNull(t *testing.T) {
	t.Parallel()
	existing := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("original")})
	NullPreserveSetFromPrior(types.SetNull(types.StringType), &existing)
	assert.True(t, existing.IsNull())
}

func TestNullPreserveSetFromPrior_nilExisting_noopNoPanic(t *testing.T) {
	t.Parallel()
	assert.NotPanics(t, func() {
		NullPreserveSetFromPrior(types.SetNull(types.StringType), nil)
	})
}

// NullPreserveListFromPrior

func TestNullPreserveListFromPrior_knownPrior_leavesExistingUnchanged(t *testing.T) {
	t.Parallel()
	existing := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("original")})
	prior := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("prior")})
	NullPreserveListFromPrior(prior, &existing)
	assert.Equal(t, []attr.Value{types.StringValue("original")}, existing.Elements())
}

func TestNullPreserveListFromPrior_unknownPrior_setsExistingToUnknown(t *testing.T) {
	t.Parallel()
	existing := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("original")})
	NullPreserveListFromPrior(types.ListUnknown(types.StringType), &existing)
	assert.True(t, existing.IsUnknown())
}

func TestNullPreserveListFromPrior_nilExisting_noopNoPanic(t *testing.T) {
	t.Parallel()
	assert.NotPanics(t, func() {
		NullPreserveListFromPrior(types.ListNull(types.StringType), nil)
	})
}

// NullPreserveMapFromPrior

func TestNullPreserveMapFromPrior_knownPrior_leavesExistingUnchanged(t *testing.T) {
	t.Parallel()
	existing := types.MapValueMust(types.StringType, map[string]attr.Value{"k": types.StringValue("original")})
	prior := types.MapValueMust(types.StringType, map[string]attr.Value{"k": types.StringValue("prior")})
	NullPreserveMapFromPrior(prior, &existing)
	assert.Equal(t, map[string]attr.Value{"k": types.StringValue("original")}, existing.Elements())
}

func TestNullPreserveMapFromPrior_nullPrior_setsExistingToNull(t *testing.T) {
	t.Parallel()
	existing := types.MapValueMust(types.StringType, map[string]attr.Value{"k": types.StringValue("original")})
	NullPreserveMapFromPrior(types.MapNull(types.StringType), &existing)
	assert.True(t, existing.IsNull())
}

func TestNullPreserveMapFromPrior_nilExisting_noopNoPanic(t *testing.T) {
	t.Parallel()
	assert.NotPanics(t, func() {
		NullPreserveMapFromPrior(types.MapNull(types.StringType), nil)
	})
}

// NullPreserveStringFromPrior

func TestNullPreserveStringFromPrior_knownPrior_leavesExistingUnchanged(t *testing.T) {
	t.Parallel()
	existing := types.StringValue("original")
	NullPreserveStringFromPrior(types.StringValue("prior"), &existing)
	assert.Equal(t, "original", existing.ValueString())
}

func TestNullPreserveStringFromPrior_nullPrior_setsExistingToNull(t *testing.T) {
	t.Parallel()
	existing := types.StringValue("original")
	NullPreserveStringFromPrior(types.StringNull(), &existing)
	assert.True(t, existing.IsNull(), "null prior should set existing to null")
	assert.False(t, existing.IsUnknown())
}

func TestNullPreserveStringFromPrior_unknownPrior_setsExistingToUnknown(t *testing.T) {
	t.Parallel()
	existing := types.StringValue("original")
	NullPreserveStringFromPrior(types.StringUnknown(), &existing)
	assert.True(t, existing.IsUnknown(), "unknown prior should set existing to unknown, not null")
}

func TestNullPreserveStringFromPrior_nilExisting_noopNoPanic(t *testing.T) {
	t.Parallel()
	assert.NotPanics(t, func() {
		NullPreserveStringFromPrior(types.StringNull(), nil)
	})
}

// NullPreserveBoolFromPrior

func TestNullPreserveBoolFromPrior_knownPrior_leavesExistingUnchanged(t *testing.T) {
	t.Parallel()
	existing := types.BoolValue(true)
	NullPreserveBoolFromPrior(types.BoolValue(false), &existing)
	assert.True(t, existing.ValueBool())
}

func TestNullPreserveBoolFromPrior_nullPrior_setsExistingToNull(t *testing.T) {
	t.Parallel()
	existing := types.BoolValue(true)
	NullPreserveBoolFromPrior(types.BoolNull(), &existing)
	assert.True(t, existing.IsNull(), "null prior should set existing to null")
	assert.False(t, existing.IsUnknown())
}

func TestNullPreserveBoolFromPrior_unknownPrior_setsExistingToUnknown(t *testing.T) {
	t.Parallel()
	existing := types.BoolValue(true)
	NullPreserveBoolFromPrior(types.BoolUnknown(), &existing)
	assert.True(t, existing.IsUnknown(), "unknown prior should set existing to unknown, not null")
}

func TestNullPreserveBoolFromPrior_nilExisting_noopNoPanic(t *testing.T) {
	t.Parallel()
	assert.NotPanics(t, func() {
		NullPreserveBoolFromPrior(types.BoolNull(), nil)
	})
}

// NullPreserveFloat32FromPrior

func TestNullPreserveFloat32FromPrior_knownPrior_leavesExistingUnchanged(t *testing.T) {
	t.Parallel()
	existing := types.Float32Value(1.5)
	NullPreserveFloat32FromPrior(types.Float32Value(2.5), &existing)
	assert.InDelta(t, 1.5, existing.ValueFloat32(), 1e-6)
}

func TestNullPreserveFloat32FromPrior_nullPrior_setsExistingToNull(t *testing.T) {
	t.Parallel()
	existing := types.Float32Value(1.5)
	NullPreserveFloat32FromPrior(types.Float32Null(), &existing)
	assert.True(t, existing.IsNull(), "null prior should set existing to null")
	assert.False(t, existing.IsUnknown())
}

func TestNullPreserveFloat32FromPrior_unknownPrior_setsExistingToUnknown(t *testing.T) {
	t.Parallel()
	existing := types.Float32Value(1.5)
	NullPreserveFloat32FromPrior(types.Float32Unknown(), &existing)
	assert.True(t, existing.IsUnknown(), "unknown prior should set existing to unknown, not null")
}

func TestNullPreserveFloat32FromPrior_nilExisting_noopNoPanic(t *testing.T) {
	t.Parallel()
	assert.NotPanics(t, func() {
		NullPreserveFloat32FromPrior(types.Float32Null(), nil)
	})
}

// NullPreserveInt64FromPrior

func TestNullPreserveInt64FromPrior_knownPrior_leavesExistingUnchanged(t *testing.T) {
	t.Parallel()
	existing := types.Int64Value(42)
	NullPreserveInt64FromPrior(types.Int64Value(99), &existing)
	assert.Equal(t, int64(42), existing.ValueInt64())
}

func TestNullPreserveInt64FromPrior_nullPrior_setsExistingToNull(t *testing.T) {
	t.Parallel()
	existing := types.Int64Value(42)
	NullPreserveInt64FromPrior(types.Int64Null(), &existing)
	assert.True(t, existing.IsNull(), "null prior should set existing to null")
	assert.False(t, existing.IsUnknown())
}

func TestNullPreserveInt64FromPrior_unknownPrior_setsExistingToUnknown(t *testing.T) {
	t.Parallel()
	existing := types.Int64Value(42)
	NullPreserveInt64FromPrior(types.Int64Unknown(), &existing)
	assert.True(t, existing.IsUnknown(), "unknown prior should set existing to unknown, not null")
}

func TestNullPreserveInt64FromPrior_nilExisting_noopNoPanic(t *testing.T) {
	t.Parallel()
	assert.NotPanics(t, func() {
		NullPreserveInt64FromPrior(types.Int64Null(), nil)
	})
}

// PreserveKnownString

func TestPreserveKnownString_knownExistingAndAPIPresent_updatesFromAPI(t *testing.T) {
	t.Parallel()
	api := "new"
	got := PreserveKnownString(types.StringValue("old"), &api)
	assert.Equal(t, types.StringValue("new"), got)
}

func TestPreserveKnownString_knownExistingAndAPINil_leavesExistingUnchanged(t *testing.T) {
	t.Parallel()
	got := PreserveKnownString(types.StringValue("old"), nil)
	assert.Equal(t, types.StringValue("old"), got, "a known existing value must not be nulled out just because the API omitted the field")
}

func TestPreserveKnownString_nullExisting_leavesNullRegardlessOfAPI(t *testing.T) {
	t.Parallel()
	api := "new"
	got := PreserveKnownString(types.StringNull(), &api)
	assert.True(t, got.IsNull())
}

func TestPreserveKnownString_unknownExisting_preservesUnknown(t *testing.T) {
	t.Parallel()
	api := "new"
	got := PreserveKnownString(types.StringUnknown(), &api)
	assert.True(t, got.IsUnknown())
}

// PreserveKnownBool

func TestPreserveKnownBool_knownExistingAndAPIPresent_updatesFromAPI(t *testing.T) {
	t.Parallel()
	api := true
	got := PreserveKnownBool(types.BoolValue(false), &api)
	assert.Equal(t, types.BoolValue(true), got)
}

func TestPreserveKnownBool_knownExistingAndAPINil_leavesExistingUnchanged(t *testing.T) {
	t.Parallel()
	got := PreserveKnownBool(types.BoolValue(false), nil)
	assert.Equal(t, types.BoolValue(false), got, "a known existing value must not be nulled out just because the API omitted the field")
}

func TestPreserveKnownBool_nullExisting_leavesNullRegardlessOfAPI(t *testing.T) {
	t.Parallel()
	api := true
	got := PreserveKnownBool(types.BoolNull(), &api)
	assert.True(t, got.IsNull())
}

// PreserveKnownFloat32

func TestPreserveKnownFloat32_knownExistingAndAPIPresent_updatesFromAPI(t *testing.T) {
	t.Parallel()
	api := float32(2.5)
	got := PreserveKnownFloat32(types.Float32Value(1.5), &api)
	assert.InDelta(t, 2.5, got.ValueFloat32(), 1e-6)
}

func TestPreserveKnownFloat32_knownExistingAndAPINil_leavesExistingUnchanged(t *testing.T) {
	t.Parallel()
	got := PreserveKnownFloat32(types.Float32Value(1.5), nil)
	assert.InDelta(t, 1.5, got.ValueFloat32(), 1e-6, "a known existing value must not be nulled out just because the API omitted the field")
}

func TestPreserveKnownFloat32_nullExisting_leavesNullRegardlessOfAPI(t *testing.T) {
	t.Parallel()
	api := float32(2.5)
	got := PreserveKnownFloat32(types.Float32Null(), &api)
	assert.True(t, got.IsNull())
}

// PreserveKnownList

func TestPreserveKnownList_knownExistingAndPresent_updatesFromNext(t *testing.T) {
	t.Parallel()
	existing := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("old")})
	next := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("new")})
	got := PreserveKnownList(existing, next, true)
	assert.Equal(t, next, got)
}

func TestPreserveKnownList_knownExistingAndNotPresent_leavesExistingUnchanged(t *testing.T) {
	t.Parallel()
	existing := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("old")})
	next := types.ListNull(types.StringType)
	got := PreserveKnownList(existing, next, false)
	assert.Equal(t, existing, got, "a known existing value must not be nulled out just because the API omitted the field")
}

func TestPreserveKnownList_nullExisting_leavesNullRegardlessOfPresent(t *testing.T) {
	t.Parallel()
	next := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("new")})
	got := PreserveKnownList(types.ListNull(types.StringType), next, true)
	assert.True(t, got.IsNull())
}
