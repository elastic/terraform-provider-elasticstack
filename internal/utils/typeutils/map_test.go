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
	"github.com/stretchr/testify/require"
)

func TestPointerInterfaceMapFromAnyMap(t *testing.T) {
	t.Parallel()

	t.Run("nil input returns empty map", func(t *testing.T) {
		t.Parallel()
		result := typeutils.PointerInterfaceMapFromAnyMap(nil)
		if len(result) != 0 {
			t.Fatalf("expected empty map, got %v", result)
		}
	})

	t.Run("empty input returns empty map", func(t *testing.T) {
		t.Parallel()
		result := typeutils.PointerInterfaceMapFromAnyMap(map[string]any{})
		if len(result) != 0 {
			t.Fatalf("expected empty map, got %v", result)
		}
	})

	t.Run("each key gets a distinct pointer with correct value", func(t *testing.T) {
		t.Parallel()
		input := map[string]any{
			"str":  "hello",
			"num":  42,
			"bool": true,
		}
		result := typeutils.PointerInterfaceMapFromAnyMap(input)

		if len(result) != len(input) {
			t.Fatalf("expected %d entries, got %d", len(input), len(result))
		}

		for k, v := range input {
			ptr, ok := result[k]
			if !ok {
				t.Errorf("key %q missing from result", k)
				continue
			}
			if ptr == nil {
				t.Errorf("key %q has nil pointer", k)
				continue
			}
			if *ptr != v {
				t.Errorf("key %q: expected %v, got %v", k, v, *ptr)
			}
		}
	})

	t.Run("pointers are distinct (not aliased)", func(t *testing.T) {
		t.Parallel()
		input := map[string]any{"a": 1, "b": 2}
		result := typeutils.PointerInterfaceMapFromAnyMap(input)

		ptrA := result["a"]
		ptrB := result["b"]
		if ptrA == ptrB {
			t.Error("expected distinct pointers for different keys")
		}
	})

	t.Run("nil value produces non-nil pointer to nil", func(t *testing.T) {
		t.Parallel()
		input := map[string]any{"key": nil}
		result := typeutils.PointerInterfaceMapFromAnyMap(input)

		ptr, ok := result["key"]
		if !ok {
			t.Fatal("key missing from result")
		}
		if ptr == nil {
			t.Fatal("expected non-nil pointer, got nil")
		}
		if *ptr != nil {
			t.Fatalf("expected dereferenced value to be nil, got %v", *ptr)
		}
	})
}

func TestFlipMap(t *testing.T) {
	t.Parallel()

	t.Run("flips keys and values", func(t *testing.T) {
		t.Parallel()
		m := map[string]string{"a": "x", "b": "y"}
		got := typeutils.FlipMap(m)
		require.Equal(t, map[string]string{"x": "a", "y": "b"}, got)
	})

	t.Run("empty map returns empty map", func(t *testing.T) {
		t.Parallel()
		got := typeutils.FlipMap(map[string]string{})
		require.Empty(t, got)
	})
}

func TestFlattenMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input map[string]any
		want  map[string]any
	}{
		{
			name:  "already flat",
			input: map[string]any{"a": 1, "b": 2},
			want:  map[string]any{"a": 1, "b": 2},
		},
		{
			name:  "one level nesting",
			input: map[string]any{"index": map[string]any{"key": 1}},
			want:  map[string]any{"index.key": 1},
		},
		{
			name:  "deep nesting",
			input: map[string]any{"a": map[string]any{"b": map[string]any{"c": "v"}}},
			want:  map[string]any{"a.b.c": "v"},
		},
		{
			name:  "mixed flat and nested",
			input: map[string]any{"x": 1, "y": map[string]any{"z": 2}},
			want:  map[string]any{"x": 1, "y.z": 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, typeutils.FlattenMap(tt.input))
		})
	}
}
