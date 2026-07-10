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

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

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
