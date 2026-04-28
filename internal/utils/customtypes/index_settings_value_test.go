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

package customtypes

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"
)

func TestCanonicalIndexSettingsJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "top-level keys get index wrapper",
			in:   `{"number_of_shards":"3"}`,
			want: `{"index":{"number_of_shards":"3"}}`,
		},
		{
			name: "already nested unchanged shape",
			in:   `{"index":{"number_of_shards":"3"}}`,
			want: `{"index":{"number_of_shards":"3"}}`,
		},
		{
			name: "empty object",
			in:   `{}`,
			want: `{}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := CanonicalIndexSettingsJSON(tc.in)
			require.NoError(t, err)
			require.JSONEq(t, tc.want, got)
		})
	}
}

func TestCanonicalIndexSettingsJSON_rejectsNull(t *testing.T) {
	t.Parallel()
	_, err := CanonicalIndexSettingsJSON("null")
	require.Error(t, err)
}

func TestCanonicalIndexSettingsJSON_nestedFormWinsOverFlatSibling(t *testing.T) {
	t.Parallel()
	got, err := CanonicalIndexSettingsJSON(`{"number_of_shards":"3","index":{"number_of_shards":"5"}}`)
	require.NoError(t, err)
	require.JSONEq(t, `{"index":{"number_of_shards":"5"}}`, got)
}

func TestCanonicalIndexSettingsJSON_byteIdenticalAcrossCalls(t *testing.T) {
	t.Parallel()
	in := `{"index":{"zebra":"1","alpha":"2"},"other":"3"}`
	first, err := CanonicalIndexSettingsJSON(in)
	require.NoError(t, err)
	for range 30 {
		got, err := CanonicalIndexSettingsJSON(in)
		require.NoError(t, err)
		require.Equal(t, first, got)
	}
}

func TestIndexSettingsValue_StringSemanticEquals_nestedWinsConflict(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	a := NewIndexSettingsValue(`{"number_of_shards":"3","index":{"number_of_shards":"5"}}`)
	b := NewIndexSettingsValue(`{"index":{"number_of_shards":"5"},"number_of_shards":"3"}`)
	eq, diags := a.StringSemanticEquals(ctx, b)
	require.False(t, diags.HasError(), "%v", diags)
	require.True(t, eq)
}

func TestIndexSettingsValue_Type(t *testing.T) {
	require.Equal(t, IndexSettingsType{}, IndexSettingsValue{}.Type(context.Background()))
}

func TestIndexSettingsValue_StringSemanticEquals(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name   string
		old    IndexSettingsValue
		newVal IndexSettingsValue
		equal  bool
	}{
		{
			name:   "dotted vs nested keys equivalent",
			old:    NewIndexSettingsValue(`{"key1.key2": 2, "index.key2.key1": "3"}`),
			newVal: NewIndexSettingsValue(`{"index": {"key1.key2": "2", "key2.key1": "3"}}`),
			equal:  true,
		},
		{
			name:   "nested index object vs top-level keys",
			old:    NewIndexSettingsValue(`{"key1": "2", "key2": "3"}`),
			newVal: NewIndexSettingsValue(`{"index": {"key1": "2", "key2": "3"}}`),
			equal:  true,
		},
		{
			name:   "nested index object same shape",
			old:    NewIndexSettingsValue(`{"index":{"key1": "2", "key2": "3"}}`),
			newVal: NewIndexSettingsValue(`{"index": {"key1": "2", "key2": "3"}}`),
			equal:  true,
		},
		{
			name:   "index prefix normalization on keys",
			old:    NewIndexSettingsValue(`{"key1": "2", "key2": "3"}`),
			newVal: NewIndexSettingsValue(`{"index.key1": "2", "index.key2": "3"}`),
			equal:  true,
		},
		{
			name:   "numeric vs string value mismatch after normalization",
			old:    NewIndexSettingsValue(`{"key1": 1, "key2": 2}`),
			newVal: NewIndexSettingsValue(`{"key1": "2", "index.key2": "3"}`),
			equal:  false,
		},
		{
			name:   "number_of_shards nested vs dotted string",
			old:    NewIndexSettingsValue(`{"index": {"number_of_shards": 1}}`),
			newVal: NewIndexSettingsValue(`{"index.number_of_shards": "1"}`),
			equal:  true,
		},
		{
			name:   "refresh_interval without index prefix vs with prefix",
			old:    NewIndexSettingsValue(`{"refresh_interval": "1s"}`),
			newVal: NewIndexSettingsValue(`{"index.refresh_interval": "1s"}`),
			equal:  true,
		},
		{
			name:   "both null",
			old:    NewIndexSettingsNull(),
			newVal: NewIndexSettingsNull(),
			equal:  true,
		},
		{
			name:   "both unknown",
			old:    NewIndexSettingsUnknown(),
			newVal: NewIndexSettingsUnknown(),
			equal:  true,
		},
		{
			name:   "null vs known",
			old:    NewIndexSettingsNull(),
			newVal: NewIndexSettingsValue(`{}`),
			equal:  false,
		},
		{
			name:   "known vs unknown",
			old:    NewIndexSettingsValue(`{}`),
			newVal: NewIndexSettingsUnknown(),
			equal:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			eq, diags := tc.old.StringSemanticEquals(ctx, tc.newVal)
			require.False(t, diags.HasError(), "%v", diags)
			require.Equal(t, tc.equal, eq)
		})
	}
}

func TestIndexSettingsValue_StringSemanticEquals_wrongValuableType(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	v := NewIndexSettingsValue(`{}`)
	eq, diags := v.StringSemanticEquals(ctx, basetypes.NewStringValue(`{}`))
	require.True(t, diags.HasError())
	require.False(t, eq)
}

func TestIndexSettingsValue_ValidateAttribute(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name       string
		val        IndexSettingsValue
		wantErrors int
	}{
		{name: "null skips validation", val: NewIndexSettingsNull()},
		{name: "unknown skips validation", val: NewIndexSettingsUnknown()},
		{name: "empty object", val: NewIndexSettingsValue(`{}`)},
		{
			name:       "json null literal rejected",
			val:        NewIndexSettingsValue(`null`),
			wantErrors: 1,
		},
		{
			name:       "array rejected",
			val:        NewIndexSettingsValue(`[]`),
			wantErrors: 1,
		},
		{
			name:       "invalid json",
			val:        NewIndexSettingsValue(`not-json`),
			wantErrors: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := xattr.ValidateAttributeRequest{Path: path.Root("settings")}
			resp := &xattr.ValidateAttributeResponse{}
			tc.val.ValidateAttribute(ctx, req, resp)
			if tc.wantErrors > 0 {
				require.True(t, resp.Diagnostics.HasError(), "expected errors")
				require.GreaterOrEqual(t, resp.Diagnostics.ErrorsCount(), tc.wantErrors)
			} else {
				require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			}
		})
	}
}
