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

package typeutils_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeAnyToAttrValue_Scalars(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name    string
		t       attr.Type
		raw     any
		want    attr.Value
		wantErr bool
	}{
		{name: "string passthrough", t: types.StringType, raw: "hello", want: types.StringValue("hello")},
		{name: "string rejects number", t: types.StringType, raw: float64(1), wantErr: true},
		{name: "nil decodes to null", t: types.StringType, raw: nil, want: types.StringNull()},

		{name: "bool native", t: types.BoolType, raw: true, want: types.BoolValue(true)},
		{name: "bool string synonym", t: types.BoolType, raw: "true", want: types.BoolValue(true)},
		{name: "bool invalid string", t: types.BoolType, raw: "nope", wantErr: true},
		{name: "bool rejects number", t: types.BoolType, raw: float64(1), wantErr: true},

		{name: "int64 from float64", t: types.Int64Type, raw: float64(42), want: types.Int64Value(42)},
		{name: "int64 from int", t: types.Int64Type, raw: 42, want: types.Int64Value(42)},
		{name: "int64 from json.Number", t: types.Int64Type, raw: json.Number("42"), want: types.Int64Value(42)},
		{name: "int64 from numeric string", t: types.Int64Type, raw: "42", want: types.Int64Value(42)},
		{name: "int64 invalid string", t: types.Int64Type, raw: "not-a-number", wantErr: true},

		{name: "normalized json string", t: jsontypes.NormalizedType{}, raw: `{"a":1}`, want: jsontypes.NewNormalizedValue(`{"a":1}`)},
		{name: "normalized rejects non-string", t: jsontypes.NormalizedType{}, raw: 1, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, diags := typeutils.DecodeAnyToAttrValue(ctx, tt.t, tt.raw)
			if tt.wantErr {
				require.True(t, diags.HasError())
				return
			}
			require.False(t, diags.HasError(), "%v", diags)
			assert.True(t, tt.want.Equal(got), "got %v, want %v", got, tt.want)
		})
	}
}

func TestDecodeAnyToAttrValue_List(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("list of strings", func(t *testing.T) {
		t.Parallel()
		got, diags := typeutils.DecodeAnyToAttrValue(ctx, types.ListType{ElemType: types.StringType}, []any{"a", "b"})
		require.False(t, diags.HasError(), "%v", diags)
		want := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("a"), types.StringValue("b")})
		assert.True(t, want.Equal(got))
	})

	t.Run("non-array raw is an error for scalar element lists", func(t *testing.T) {
		t.Parallel()
		_, diags := typeutils.DecodeAnyToAttrValue(ctx, types.ListType{ElemType: types.StringType}, "not-a-list")
		require.True(t, diags.HasError())
	})

	objType := types.ObjectType{AttrTypes: map[string]attr.Type{"enabled": types.BoolType}}

	t.Run("list of objects takes first element", func(t *testing.T) {
		t.Parallel()
		got, diags := typeutils.DecodeAnyToAttrValue(ctx, types.ListType{ElemType: objType}, []any{
			map[string]any{"enabled": true},
		})
		require.False(t, diags.HasError(), "%v", diags)
		list, ok := got.(types.List)
		require.True(t, ok)
		require.Len(t, list.Elements(), 1)
	})

	t.Run("non-array raw is null for object element lists", func(t *testing.T) {
		t.Parallel()
		got, diags := typeutils.DecodeAnyToAttrValue(ctx, types.ListType{ElemType: objType}, "not-a-list")
		require.False(t, diags.HasError(), "%v", diags)
		assert.True(t, got.(types.List).IsNull())
	})
}

func TestDecodeAnyToAttrValue_Set(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	got, diags := typeutils.DecodeAnyToAttrValue(ctx, types.SetType{ElemType: types.StringType}, []any{"a", "b"})
	require.False(t, diags.HasError(), "%v", diags)
	want := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("a"), types.StringValue("b")})
	assert.True(t, want.Equal(got))
}

func TestDecodeAnyToAttrValue_Object(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ot := types.ObjectType{AttrTypes: map[string]attr.Type{
		"name":  types.StringType,
		"count": types.Int64Type,
	}}

	t.Run("decodes present keys and nulls missing ones", func(t *testing.T) {
		t.Parallel()
		got, diags := typeutils.DecodeAnyToAttrValue(ctx, ot, map[string]any{"name": "foo"})
		require.False(t, diags.HasError(), "%v", diags)
		obj, ok := got.(types.Object)
		require.True(t, ok)
		assert.Equal(t, "foo", obj.Attributes()["name"].(types.String).ValueString())
		assert.True(t, obj.Attributes()["count"].(types.Int64).IsNull())
	})

	t.Run("accepts single-element slice wrapping a map", func(t *testing.T) {
		t.Parallel()
		got, diags := typeutils.DecodeAnyToAttrValue(ctx, ot, []any{map[string]any{"name": "foo"}})
		require.False(t, diags.HasError(), "%v", diags)
		obj, ok := got.(types.Object)
		require.True(t, ok)
		assert.Equal(t, "foo", obj.Attributes()["name"].(types.String).ValueString())
	})

	t.Run("rejects non-map raw", func(t *testing.T) {
		t.Parallel()
		_, diags := typeutils.DecodeAnyToAttrValue(ctx, ot, "not-a-map")
		require.True(t, diags.HasError())
	})
}

func TestNullValueForType(t *testing.T) {
	t.Parallel()

	assert.True(t, typeutils.NullValueForType(types.StringType).(types.String).IsNull())
	assert.True(t, typeutils.NullValueForType(types.Int64Type).(types.Int64).IsNull())
	assert.True(t, typeutils.NullValueForType(types.BoolType).(types.Bool).IsNull())
	assert.True(t, typeutils.NullValueForType(types.ListType{ElemType: types.StringType}).(types.List).IsNull())
	assert.True(t, typeutils.NullValueForType(types.SetType{ElemType: types.StringType}).(types.Set).IsNull())
	assert.True(t, typeutils.NullValueForType(types.ObjectType{AttrTypes: map[string]attr.Type{}}).(types.Object).IsNull())
	assert.True(t, typeutils.NullValueForType(jsontypes.NormalizedType{}).(jsontypes.Normalized).IsNull())
}
