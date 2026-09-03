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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalJSONDiag(t *testing.T) {
	t.Parallel()

	t.Run("valid JSON into map succeeds", func(t *testing.T) {
		t.Parallel()
		got, diags := typeutils.UnmarshalJSONDiag[map[string]any](`{"key":"value"}`, "parse error")
		require.False(t, diags.HasError())
		require.Equal(t, map[string]any{"key": "value"}, got)
	})

	t.Run("invalid JSON returns error diag", func(t *testing.T) {
		t.Parallel()
		_, diags := typeutils.UnmarshalJSONDiag[map[string]any]("not-json", "parse error")
		require.True(t, diags.HasError())
		require.Equal(t, "parse error", diags[0].Summary())
	})

	t.Run("valid JSON into slice succeeds", func(t *testing.T) {
		t.Parallel()
		got, diags := typeutils.UnmarshalJSONDiag[[]string](`["a","b"]`, "parse error")
		require.False(t, diags.HasError())
		require.Equal(t, []string{"a", "b"}, got)
	})

	t.Run("error summary is preserved", func(t *testing.T) {
		t.Parallel()
		_, diags := typeutils.UnmarshalJSONDiag[map[string]any]("{bad", "custom summary")
		require.True(t, diags.HasError())
		require.Equal(t, "custom summary", diags[0].Summary())
	})
}

func TestNormalizeJSONScalar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   any
		want any
	}{
		{name: "true string", in: "true", want: true},
		{name: "false string", in: "false", want: false},
		{name: "null string", in: "null", want: nil},
		{name: "other string unchanged", in: "hello", want: "hello"},
		{name: "bool passthrough", in: true, want: true},
		{name: "float passthrough", in: float64(42), want: float64(42)},
		{name: "nil passthrough", in: nil, want: nil},
		{
			name: "map with string-encoded scalars",
			in:   map[string]any{"enabled": "true", "dynamic": "false", "meta": "null", "name": "foo"},
			want: map[string]any{"enabled": true, "dynamic": false, "meta": nil, "name": "foo"},
		},
		{
			name: "slice with mixed values",
			in:   []any{"true", "false", "null", "bar", float64(1)},
			want: []any{true, false, nil, "bar", float64(1)},
		},
		{
			name: "nested map",
			in:   map[string]any{"outer": map[string]any{"flag": "true"}},
			want: map[string]any{"outer": map[string]any{"flag": true}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := typeutils.NormalizeJSONScalar(tt.in)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestIsEmptyJSONObject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want bool
	}{
		{name: "empty string", in: "", want: true},
		{name: "whitespace only", in: "  ", want: true},
		{name: "empty JSON object", in: "{}", want: true},
		{name: "whitespace-padded empty object", in: "  {}  ", want: true},
		{name: "non-empty JSON object", in: `{"k":"v"}`, want: false},
		{name: "JSON array", in: "[]", want: false},
		{name: "JSON null literal", in: "null", want: false},
		{name: "JSON string", in: `"string"`, want: false},
		{name: "invalid JSON", in: "not-json", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, typeutils.IsEmptyJSONObject(tt.in))
		})
	}
}

func TestJSONBytesEqual(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		a, b    []byte
		want    bool
		wantErr bool
	}{
		{
			name: "identical JSON",
			a:    []byte(`{"a":1,"b":2}`),
			b:    []byte(`{"a":1,"b":2}`),
			want: true,
		},
		{
			name: "semantically equivalent with different key order",
			a:    []byte(`{"a":1,"b":2}`),
			b:    []byte(`{"b":2,"a":1}`),
			want: true,
		},
		{
			name: "different values",
			a:    []byte(`{"a":1}`),
			b:    []byte(`{"a":2}`),
			want: false,
		},
		{
			name:    "invalid JSON in a",
			a:       []byte(`not json`),
			b:       []byte(`{}`),
			wantErr: true,
		},
		{
			name:    "invalid JSON in b",
			a:       []byte(`{}`),
			b:       []byte(`not json`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := typeutils.JSONBytesEqual(tt.a, tt.b)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestMarshalToNormalized(t *testing.T) {
	t.Parallel()

	t.Run("nil returns null", func(t *testing.T) {
		t.Parallel()
		var d diag.Diagnostics
		result := typeutils.MarshalToNormalized(nil, path.Root("field"), &d)
		require.False(t, d.HasError())
		require.True(t, result.IsNull())
	})

	t.Run("typed nil map returns null", func(t *testing.T) {
		t.Parallel()
		var value map[string]any
		var d diag.Diagnostics
		result := typeutils.MarshalToNormalized(value, path.Root("field"), &d)
		require.False(t, d.HasError())
		require.True(t, result.IsNull())
	})

	t.Run("pointer to nil map returns null", func(t *testing.T) {
		t.Parallel()
		var value map[string]any
		var d diag.Diagnostics
		result := typeutils.MarshalToNormalized(&value, path.Root("field"), &d)
		require.False(t, d.HasError())
		require.True(t, result.IsNull())
	})

	t.Run("map marshals correctly", func(t *testing.T) {
		t.Parallel()
		var d diag.Diagnostics
		result := typeutils.MarshalToNormalized(map[string]any{"key": "val"}, path.Root("field"), &d)
		require.False(t, d.HasError())
		require.False(t, result.IsNull())
		require.JSONEq(t, `{"key":"val"}`, result.ValueString())
	})

	t.Run("string marshals to quoted JSON", func(t *testing.T) {
		t.Parallel()
		var d diag.Diagnostics
		result := typeutils.MarshalToNormalized("hello", path.Root("field"), &d)
		require.False(t, d.HasError())
		require.Equal(t, `"hello"`, result.ValueString())
	})

	t.Run("struct marshals correctly", func(t *testing.T) {
		t.Parallel()
		type inner struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}
		var d diag.Diagnostics
		result := typeutils.MarshalToNormalized(inner{Name: "alice", Age: 30}, path.Root("field"), &d)
		require.False(t, d.HasError())
		require.JSONEq(t, `{"name":"alice","age":30}`, result.ValueString())
	})

	t.Run("unmarshalable value adds error and returns null", func(t *testing.T) {
		t.Parallel()
		var d diag.Diagnostics
		result := typeutils.MarshalToNormalized(make(chan int), path.Root("field"), &d)
		require.True(t, d.HasError())
		require.True(t, result.IsNull())
		require.Contains(t, d[0].Summary(), "marshal failure")
	})
}
