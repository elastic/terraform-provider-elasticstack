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

package dataview

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFieldAttrsValue_MapSemanticEquals exercises every REQ-015 branch in
// FieldAttrsValue.MapSemanticEquals so the drift-suppression contract is locked in
// independently of the read/update integration paths (TestBuildFieldAttrsMetadataDelta
// only exercises the delta builder, not the semantic-equality logic that decides whether
// the delta is computed at all).
func TestFieldAttrsValue_MapSemanticEquals(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elemType := getFieldAttrElemType()

	tests := []struct {
		name     string
		newVal   FieldAttrsValue
		priorVal FieldAttrsValue
		expected bool
	}{
		{
			name:     "null vs null is equal",
			newVal:   NewFieldAttrsNull(elemType),
			priorVal: NewFieldAttrsNull(elemType),
			expected: true,
		},
		{
			name:     "unknown vs unknown is equal",
			newVal:   NewFieldAttrsUnknown(elemType),
			priorVal: NewFieldAttrsUnknown(elemType),
			expected: true,
		},
		{
			name:     "null vs unknown is not equal (no implicit equivalence)",
			newVal:   NewFieldAttrsNull(elemType),
			priorVal: NewFieldAttrsUnknown(elemType),
			expected: false,
		},
		{
			name:     "unknown vs null is not equal",
			newVal:   NewFieldAttrsUnknown(elemType),
			priorVal: NewFieldAttrsNull(elemType),
			expected: false,
		},
		{
			name:   "null vs prior with only server-injected count is equal (REQ-015 scenario 1)",
			newVal: NewFieldAttrsNull(elemType),
			priorVal: mustNewFieldAttrsValue(ctx, t, elemType, map[string]fieldAttrModel{
				"host.hostname": {CustomLabel: types.StringNull(), Count: types.Int64Value(5)},
			}),
			expected: true,
		},
		{
			name:   "null vs prior with custom_label is a real removal (not equal)",
			newVal: NewFieldAttrsNull(elemType),
			priorVal: mustNewFieldAttrsValue(ctx, t, elemType, map[string]fieldAttrModel{
				"host.hostname": {CustomLabel: types.StringValue("Host"), Count: types.Int64Null()},
			}),
			expected: false,
		},
		{
			name:   "null vs prior mixing server-only and user-managed entry is not equal",
			newVal: NewFieldAttrsNull(elemType),
			priorVal: mustNewFieldAttrsValue(ctx, t, elemType, map[string]fieldAttrModel{
				"host.hostname": {CustomLabel: types.StringNull(), Count: types.Int64Value(5)},
				"keep.me":       {CustomLabel: types.StringValue("Keep"), Count: types.Int64Null()},
			}),
			expected: false,
		},
		{
			name: "same custom_label and null counts is equal",
			newVal: mustNewFieldAttrsValue(ctx, t, elemType, map[string]fieldAttrModel{
				"f1": {CustomLabel: types.StringValue("Same"), Count: types.Int64Null()},
			}),
			priorVal: mustNewFieldAttrsValue(ctx, t, elemType, map[string]fieldAttrModel{
				"f1": {CustomLabel: types.StringValue("Same"), Count: types.Int64Null()},
			}),
			expected: true,
		},
		{
			name: "same custom_label, plan omits count, state has server count is equal (REQ-015)",
			newVal: mustNewFieldAttrsValue(ctx, t, elemType, map[string]fieldAttrModel{
				"f1": {CustomLabel: types.StringValue("Same"), Count: types.Int64Null()},
			}),
			priorVal: mustNewFieldAttrsValue(ctx, t, elemType, map[string]fieldAttrModel{
				"f1": {CustomLabel: types.StringValue("Same"), Count: types.Int64Value(7)},
			}),
			expected: true,
		},
		{
			name: "explicit count change is detected",
			newVal: mustNewFieldAttrsValue(ctx, t, elemType, map[string]fieldAttrModel{
				"f1": {CustomLabel: types.StringValue("L"), Count: types.Int64Value(3)},
			}),
			priorVal: mustNewFieldAttrsValue(ctx, t, elemType, map[string]fieldAttrModel{
				"f1": {CustomLabel: types.StringValue("L"), Count: types.Int64Value(2)},
			}),
			expected: false,
		},
		{
			name: "plan introduces explicit count where state had none is detected",
			newVal: mustNewFieldAttrsValue(ctx, t, elemType, map[string]fieldAttrModel{
				"f1": {CustomLabel: types.StringValue("L"), Count: types.Int64Value(1)},
			}),
			priorVal: mustNewFieldAttrsValue(ctx, t, elemType, map[string]fieldAttrModel{
				"f1": {CustomLabel: types.StringValue("L"), Count: types.Int64Null()},
			}),
			expected: false,
		},
		{
			name: "different custom_label is detected (strict comparison)",
			newVal: mustNewFieldAttrsValue(ctx, t, elemType, map[string]fieldAttrModel{
				"f1": {CustomLabel: types.StringValue("New"), Count: types.Int64Null()},
			}),
			priorVal: mustNewFieldAttrsValue(ctx, t, elemType, map[string]fieldAttrModel{
				"f1": {CustomLabel: types.StringValue("Old"), Count: types.Int64Null()},
			}),
			expected: false,
		},
		{
			name: "plan adds a new field is a real change",
			newVal: mustNewFieldAttrsValue(ctx, t, elemType, map[string]fieldAttrModel{
				"new.field": {CustomLabel: types.StringValue("New"), Count: types.Int64Null()},
			}),
			priorVal: mustNewFieldAttrsValue(ctx, t, elemType, map[string]fieldAttrModel{}),
			expected: false,
		},
		{
			name: "plan adds field where state has only server-only entries is a real change",
			newVal: mustNewFieldAttrsValue(ctx, t, elemType, map[string]fieldAttrModel{
				"new.field": {CustomLabel: types.StringValue("New"), Count: types.Int64Null()},
			}),
			priorVal: mustNewFieldAttrsValue(ctx, t, elemType, map[string]fieldAttrModel{
				"server.only": {CustomLabel: types.StringNull(), Count: types.Int64Value(3)},
			}),
			expected: false,
		},
		{
			name: "plan drops a server-only count entry is equal (suppress drift)",
			newVal: mustNewFieldAttrsValue(ctx, t, elemType, map[string]fieldAttrModel{
				"keep": {CustomLabel: types.StringValue("Keep"), Count: types.Int64Null()},
			}),
			priorVal: mustNewFieldAttrsValue(ctx, t, elemType, map[string]fieldAttrModel{
				"keep":        {CustomLabel: types.StringValue("Keep"), Count: types.Int64Null()},
				"server.only": {CustomLabel: types.StringNull(), Count: types.Int64Value(5)},
			}),
			expected: true,
		},
		{
			name: "plan drops a user-managed entry is a real change",
			newVal: mustNewFieldAttrsValue(ctx, t, elemType, map[string]fieldAttrModel{
				"keep": {CustomLabel: types.StringValue("Keep"), Count: types.Int64Null()},
			}),
			priorVal: mustNewFieldAttrsValue(ctx, t, elemType, map[string]fieldAttrModel{
				"keep":   {CustomLabel: types.StringValue("Keep"), Count: types.Int64Null()},
				"remove": {CustomLabel: types.StringValue("Bye"), Count: types.Int64Null()},
			}),
			expected: false,
		},
		{
			name:     "empty map vs empty map is equal",
			newVal:   mustNewFieldAttrsValue(ctx, t, elemType, map[string]fieldAttrModel{}),
			priorVal: mustNewFieldAttrsValue(ctx, t, elemType, map[string]fieldAttrModel{}),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, diags := tt.newVal.MapSemanticEquals(ctx, tt.priorVal)
			require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
			assert.Equal(t, tt.expected, got)
		})
	}
}

// TestFieldAttrsValue_MapSemanticEquals_TypeMismatch verifies the type-assertion guard surfaces
// a clear diagnostic instead of panicking.
func TestFieldAttrsValue_MapSemanticEquals_TypeMismatch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	elemType := getFieldAttrElemType()
	v := NewFieldAttrsNull(elemType)

	got, diags := v.MapSemanticEquals(ctx, types.MapNull(elemType))
	assert.False(t, got)
	require.True(t, diags.HasError())
	assert.Contains(t, diags[0].Summary(), "Semantic Equality Check Error")
}

func mustNewFieldAttrsValue(ctx context.Context, t *testing.T, elemType attr.Type, entries map[string]fieldAttrModel) FieldAttrsValue {
	t.Helper()
	v, diags := NewFieldAttrsValueFrom(ctx, elemType, entries)
	require.False(t, diags.HasError(), "build FieldAttrsValue: %v", diags)
	return v
}
