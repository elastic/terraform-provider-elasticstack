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

func TestTreeVisitor(t *testing.T) {
	t.Parallel()

	t.Run("zero value returns structural copy", func(t *testing.T) {
		t.Parallel()
		input := map[string]any{
			"a": []any{"x", "y"},
			"b": map[string]any{"c": float64(1)},
		}
		var vis typeutils.TreeVisitor
		require.Equal(t, input, vis.Walk(input))
	})

	t.Run("slice collapse short-circuits elements", func(t *testing.T) {
		t.Parallel()
		vis := typeutils.TreeVisitor{
			Slice: func(s []any) (any, bool) {
				if len(s) == 1 {
					if str, ok := s[0].(string); ok {
						return str, true
					}
				}
				return nil, false
			},
		}
		require.Equal(t, "solo", vis.Walk([]any{"solo"}))
		require.Equal(t, []any{"a", "b"}, vis.Walk([]any{"a", "b"}))
	})

	t.Run("map short-circuits children", func(t *testing.T) {
		t.Parallel()
		vis := typeutils.TreeVisitor{
			Map: func(m map[string]any) (any, bool) {
				if len(m) == 1 {
					if inner, ok := m["value"]; ok {
						return inner, true
					}
				}
				return nil, false
			},
		}
		require.Equal(t, "x", vis.Walk(map[string]any{"value": "x"}))
		require.Equal(t,
			map[string]any{"term": "x"},
			vis.Walk(map[string]any{"term": map[string]any{"value": "x"}}),
		)
	})

	t.Run("map mutates before continuing recursion", func(t *testing.T) {
		t.Parallel()
		vis := typeutils.TreeVisitor{
			Map: func(m map[string]any) (any, bool) {
				if field, ok := m["field"].(map[string]any); ok {
					for k, v := range field {
						if arr, ok := v.([]any); ok && len(arr) == 1 {
							field[k] = arr[0]
						}
					}
				}
				return nil, false
			},
		}
		got := vis.Walk(map[string]any{
			"field": map[string]any{"key": []any{"username"}},
		})
		require.Equal(t, map[string]any{
			"field": map[string]any{"key": "username"},
		}, got)
	})

	t.Run("map child and post map", func(t *testing.T) {
		t.Parallel()
		vis := typeutils.TreeVisitor{
			MapChild: func(key string, walked any) any {
				if key == "match" {
					if arr, ok := walked.([]any); ok && len(arr) == 1 {
						if s, ok := arr[0].(string); ok {
							return s
						}
					}
				}
				return walked
			},
			PostMap: func(m map[string]any) any {
				if m["type"] == "object" {
					if _, ok := m["properties"]; ok {
						delete(m, "type")
					}
				}
				return m
			},
		}
		got := vis.Walk(map[string]any{
			"type":       "object",
			"properties": map[string]any{},
			"match":      []any{"name_*"},
		})
		require.Equal(t, map[string]any{
			"properties": map[string]any{},
			"match":      "name_*",
		}, got)
	})

	t.Run("leaf transforms scalars only", func(t *testing.T) {
		t.Parallel()
		vis := typeutils.TreeVisitor{
			Leaf: func(v any) any {
				if s, ok := v.(string); ok && s == "true" {
					return true
				}
				return v
			},
		}
		got := vis.Walk(map[string]any{"dynamic": "true", "other": "x"})
		require.Equal(t, map[string]any{"dynamic": true, "other": "x"}, got)
	})

	t.Run("nested structures", func(t *testing.T) {
		t.Parallel()
		vis := typeutils.TreeVisitor{
			Slice: func(s []any) (any, bool) {
				if len(s) == 1 {
					switch s[0].(type) {
					case string, float64, bool:
						return s[0], true
					}
				}
				return nil, false
			},
		}
		input := map[string]any{
			"remove": map[string]any{
				"field": []any{"my_field"},
				"on_failure": []any{
					map[string]any{
						"set": map[string]any{
							"field": "error.message",
							"value": []any{"failed"},
						},
					},
				},
			},
		}
		want := map[string]any{
			"remove": map[string]any{
				"field": "my_field",
				"on_failure": []any{
					map[string]any{
						"set": map[string]any{
							"field": "error.message",
							"value": "failed",
						},
					},
				},
			},
		}
		require.Equal(t, want, vis.Walk(input))
	})
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
