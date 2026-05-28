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

package index

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_stringSliceFromAny(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		v    any
		want []string
	}{
		{
			name: "nil returns nil",
			v:    nil,
			want: nil,
		},
		{
			name: "string slice passthrough",
			v:    []string{"a", "b"},
			want: []string{"a", "b"},
		},
		{
			name: "any slice of strings",
			v:    []any{"x", "y"},
			want: []string{"x", "y"},
		},
		{
			name: "any slice with non-string elements",
			v:    []any{1.0, true},
			want: []string{"1", "true"},
		},
		{
			name: "scalar string",
			v:    "hello",
			want: []string{"hello"},
		},
		{
			name: "scalar string with leading/trailing spaces",
			v:    "  hello  ",
			want: []string{"hello"},
		},
		{
			name: "empty string returns nil",
			v:    "",
			want: nil,
		},
		{
			name: "whitespace-only string returns nil",
			v:    "   ",
			want: nil,
		},
		{
			name: "JSON array string is parsed",
			v:    `["a","b"]`,
			want: []string{"a", "b"},
		},
		{
			name: "JSON array string with spaces",
			v:    `  ["a","b"]  `,
			want: []string{"a", "b"},
		},
		{
			name: "malformed JSON array string falls back to scalar",
			v:    "[not json",
			want: []string{"[not json"},
		},
		{
			name: "unrecognised type returns nil",
			v:    map[string]any{"key": "val"},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := stringSliceFromAny(tt.v)
			require.Equal(t, tt.want, got)
		})
	}
}
