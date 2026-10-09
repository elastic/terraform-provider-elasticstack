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

package securityexceptionitem

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

// marshalEntry marshals a generated "entry" union type to a map for assertions.
func marshalEntry(t *testing.T, v interface{ MarshalJSON() ([]byte, error) }) map[string]any {
	t.Helper()
	b, err := v.MarshalJSON()
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(b, &m))
	return m
}

func TestConvertSingleValueEntryFromAPI(t *testing.T) {
	t.Parallel()

	t.Run("flat EntryModel: sets value and resets Values", func(t *testing.T) {
		t.Parallel()
		values, d := types.ListValueFrom(context.Background(), types.StringType, []string{"stale"})
		require.False(t, d.HasError())
		entry := &EntryModel{Values: values}

		convertSingleValueEntryFromAPI(map[string]any{"value": "foo"}, entry)
		require.Equal(t, types.StringValue("foo"), entry.Value)
		require.True(t, entry.Values.IsNull())
	})

	t.Run("flat EntryModel: missing value nulls Value", func(t *testing.T) {
		t.Parallel()
		entry := &EntryModel{}
		convertSingleValueEntryFromAPI(map[string]any{}, entry)
		require.True(t, entry.Value.IsNull())
	})

	t.Run("nested NestedEntryModel: sets value and resets Values", func(t *testing.T) {
		t.Parallel()
		entry := &NestedEntryModel{}
		convertSingleValueEntryFromAPI(map[string]any{"value": "bar"}, entry)
		require.Equal(t, types.StringValue("bar"), entry.Value)
		require.True(t, entry.Values.IsNull())
	})

	t.Run("nested NestedEntryModel: missing value nulls Value", func(t *testing.T) {
		t.Parallel()
		entry := &NestedEntryModel{}
		convertSingleValueEntryFromAPI(map[string]any{}, entry)
		require.True(t, entry.Value.IsNull())
	})
}

func TestConvertMultiValueEntryFromAPI(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("flat EntryModel: sets values and resets Value", func(t *testing.T) {
		t.Parallel()
		entry := &EntryModel{Value: types.StringValue("stale")}
		diags := convertMultiValueEntryFromAPI(ctx, map[string]any{"value": []any{"a", "b"}}, entry)
		require.False(t, diags.HasError())
		require.True(t, entry.Value.IsNull())

		var values []string
		require.False(t, entry.Values.ElementsAs(ctx, &values, false).HasError())
		require.Equal(t, []string{"a", "b"}, values)
	})

	t.Run("flat EntryModel: non-string elements are dropped", func(t *testing.T) {
		t.Parallel()
		entry := &EntryModel{}
		diags := convertMultiValueEntryFromAPI(ctx, map[string]any{"value": []any{"a", 1, "c"}}, entry)
		require.False(t, diags.HasError())

		var values []string
		require.False(t, entry.Values.ElementsAs(ctx, &values, false).HasError())
		require.Equal(t, []string{"a", "c"}, values)
	})

	t.Run("flat EntryModel: missing value nulls Values", func(t *testing.T) {
		t.Parallel()
		entry := &EntryModel{}
		diags := convertMultiValueEntryFromAPI(ctx, map[string]any{}, entry)
		require.False(t, diags.HasError())
		require.True(t, entry.Values.IsNull())
	})

	t.Run("nested NestedEntryModel: sets values and resets Value", func(t *testing.T) {
		t.Parallel()
		entry := &NestedEntryModel{Value: types.StringValue("stale")}
		diags := convertMultiValueEntryFromAPI(ctx, map[string]any{"value": []any{"x"}}, entry)
		require.False(t, diags.HasError())
		require.True(t, entry.Value.IsNull())

		var values []string
		require.False(t, entry.Values.ElementsAs(ctx, &values, false).HasError())
		require.Equal(t, []string{"x"}, values)
	})
}

func TestConvertMatchEntryToAPI(t *testing.T) {
	t.Parallel()

	t.Run("flat: valid value produces a match entry", func(t *testing.T) {
		t.Parallel()
		result, diags := convertMatchEntryToAPI(EntryModel{Value: types.StringValue("foo")}, "field.name", "included")
		require.False(t, diags.HasError())
		m := marshalEntry(t, &result)
		require.Equal(t, "match", m["type"])
		require.Equal(t, "field.name", m["field"])
		require.Equal(t, "included", m["operator"])
		require.Equal(t, "foo", m["value"])
	})

	t.Run("flat: empty value is rejected", func(t *testing.T) {
		t.Parallel()
		_, diags := convertMatchEntryToAPI(EntryModel{Value: types.StringValue("")}, "field.name", "included")
		require.True(t, diags.HasError())
		require.Contains(t, diags.Errors()[0].Detail(), "Attribute 'value' is required when type is 'match'")
	})

	t.Run("nested: valid value produces a nested match entry", func(t *testing.T) {
		t.Parallel()
		result, diags := convertNestedMatchEntryToAPI(NestedEntryModel{Value: types.StringValue("bar")}, "field.name", "excluded")
		require.False(t, diags.HasError())
		m := marshalEntry(t, &result)
		require.Equal(t, "match", m["type"])
		require.Equal(t, "bar", m["value"])
	})

	t.Run("nested: unknown value is rejected with nested-specific message", func(t *testing.T) {
		t.Parallel()
		_, diags := convertNestedMatchEntryToAPI(NestedEntryModel{Value: types.StringNull()}, "field.name", "excluded")
		require.True(t, diags.HasError())
		require.Contains(t, diags.Errors()[0].Detail(), "Attribute 'value' is required for nested entry when type is 'match'")
	})
}

func TestConvertMatchAnyEntryToAPI(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("flat: valid values produce a match_any entry", func(t *testing.T) {
		t.Parallel()
		values, d := types.ListValueFrom(ctx, types.StringType, []string{"a", "b"})
		require.False(t, d.HasError())

		result, diags := convertMatchAnyEntryToAPI(ctx, EntryModel{Values: values}, "field.name", "included")
		require.False(t, diags.HasError())
		m := marshalEntry(t, &result)
		require.Equal(t, "match_any", m["type"])
		require.Equal(t, []any{"a", "b"}, m["value"])
	})

	t.Run("flat: empty values list is rejected", func(t *testing.T) {
		t.Parallel()
		values, d := types.ListValueFrom(ctx, types.StringType, []string{})
		require.False(t, d.HasError())

		_, diags := convertMatchAnyEntryToAPI(ctx, EntryModel{Values: values}, "field.name", "included")
		require.True(t, diags.HasError())
		require.Contains(t, diags.Errors()[0].Detail(), "must contain at least one value when type is 'match_any'")
	})

	t.Run("nested: unknown values is rejected with nested-specific message", func(t *testing.T) {
		t.Parallel()
		_, diags := convertNestedMatchAnyEntryToAPI(ctx, NestedEntryModel{Values: types.ListNull(types.StringType)}, "field.name", "included")
		require.True(t, diags.HasError())
		require.Contains(t, diags.Errors()[0].Detail(), "Attribute 'values' is required for nested entry when type is 'match_any'")
	})
}

func TestConvertExistsEntryToAPI(t *testing.T) {
	t.Parallel()

	t.Run("flat: produces an exists entry", func(t *testing.T) {
		t.Parallel()
		result, diags := convertExistsEntryToAPI("field.name", "included")
		require.False(t, diags.HasError())
		m := marshalEntry(t, &result)
		require.Equal(t, "exists", m["type"])
		require.Equal(t, "field.name", m["field"])
	})

	t.Run("nested: produces a nested exists entry", func(t *testing.T) {
		t.Parallel()
		result, diags := convertNestedExistsEntryToAPI("field.name", "excluded")
		require.False(t, diags.HasError())
		m := marshalEntry(t, &result)
		require.Equal(t, "exists", m["type"])
	})
}
