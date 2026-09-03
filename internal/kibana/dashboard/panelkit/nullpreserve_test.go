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

func TestNullPreserveFromPrior(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		prior    attr.Value
		existing attr.Value
		want     attr.Value
		nilDest  bool
	}{
		{
			name:     "known prior leaves existing unchanged",
			prior:    types.StringValue("prior"),
			existing: types.StringValue("original"),
			want:     types.StringValue("original"),
		},
		{
			name:     "null prior sets existing to null",
			prior:    types.Int64Null(),
			existing: types.Int64Value(42),
			want:     types.Int64Null(),
		},
		{
			name:     "unknown prior sets existing to unknown",
			prior:    types.BoolUnknown(),
			existing: types.BoolValue(true),
			want:     types.BoolUnknown(),
		},
		{
			name:    "nil existing is a no-op",
			prior:   types.StringNull(),
			nilDest: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.nilDest {
				assert.NotPanics(t, func() {
					NullPreserveFromPrior(tt.prior, (*attr.Value)(nil))
				})
				return
			}
			existing := tt.existing
			NullPreserveFromPrior(tt.prior, &existing)
			assert.Equal(t, tt.want, existing)
		})
	}
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
