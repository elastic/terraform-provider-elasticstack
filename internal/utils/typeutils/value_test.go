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

func TestIsEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input any
		want  bool
	}{
		{name: "nil", input: nil, want: true},
		{name: "zero int", input: 0, want: true},
		{name: "non-zero int", input: 1, want: false},
		{name: "zero float64 (not empty due to interface comparison semantics)", input: float64(0), want: false},
		{name: "non-zero float64", input: float64(1.5), want: false},
		{name: "blank string", input: "   ", want: true},
		{name: "empty string", input: "", want: true},
		{name: "non-empty string", input: "hello", want: false},
		{name: "empty slice", input: []any{}, want: true},
		{name: "non-empty slice", input: []any{1}, want: false},
		{name: "empty map", input: map[any]any{}, want: true},
		{name: "non-empty map", input: map[any]any{"a": 1}, want: false},
		{name: "struct value is not empty", input: struct{}{}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, typeutils.IsEmpty(tt.input))
		})
	}
}

func TestNonZero(t *testing.T) {
	t.Parallel()

	t.Run("string", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, "fallback", typeutils.NonZero("", "fallback"))
		require.Equal(t, "value", typeutils.NonZero("value", "fallback"))
	})

	t.Run("int", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, 42, typeutils.NonZero(0, 42))
		require.Equal(t, 7, typeutils.NonZero(7, 42))
	})

	t.Run("bool", func(t *testing.T) {
		t.Parallel()
		require.True(t, typeutils.NonZero(false, true))
		require.True(t, typeutils.NonZero(true, true))
	})

	t.Run("struct", func(t *testing.T) {
		t.Parallel()
		type foo struct{ x int }
		fallback := foo{x: 42}
		require.Equal(t, fallback, typeutils.NonZero(foo{}, fallback))
		require.Equal(t, foo{x: 7}, typeutils.NonZero(foo{x: 7}, fallback))
	})
}
