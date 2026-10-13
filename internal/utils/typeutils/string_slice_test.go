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

package typeutils

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_StringSliceFromAny(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		v      any
		want   []string
		wantOk bool
	}{
		{
			name:   "nil is unrecognised",
			v:      nil,
			want:   nil,
			wantOk: false,
		},
		{
			name:   "string slice passthrough",
			v:      []string{"a", "b"},
			want:   []string{"a", "b"},
			wantOk: true,
		},
		{
			name:   "any slice of strings",
			v:      []any{"x", "y"},
			want:   []string{"x", "y"},
			wantOk: true,
		},
		{
			name:   "empty any slice is a legal empty set",
			v:      []any{},
			want:   []string{},
			wantOk: true,
		},
		{
			name:   "any slice with non-string elements fails closed",
			v:      []any{1.0, true},
			want:   nil,
			wantOk: false,
		},
		{
			name:   "scalar string",
			v:      "hello",
			want:   []string{"hello"},
			wantOk: true,
		},
		{
			name:   "scalar string with leading/trailing spaces is trimmed",
			v:      "  hello  ",
			want:   []string{"hello"},
			wantOk: true,
		},
		{
			name:   "empty string is unrecognised",
			v:      "",
			want:   nil,
			wantOk: false,
		},
		{
			name:   "whitespace-only string is unrecognised",
			v:      "   ",
			want:   nil,
			wantOk: false,
		},
		{
			name:   "JSON array string is parsed",
			v:      `["a","b"]`,
			want:   []string{"a", "b"},
			wantOk: true,
		},
		{
			name:   "JSON array string with spaces",
			v:      `  ["a","b"]  `,
			want:   []string{"a", "b"},
			wantOk: true,
		},
		{
			name:   "empty JSON array string is a legal empty set",
			v:      `[]`,
			want:   []string{},
			wantOk: true,
		},
		{
			name:   "malformed JSON array string falls back to scalar",
			v:      "[not json",
			want:   []string{"[not json"},
			wantOk: true,
		},
		{
			name:   "unrecognised type",
			v:      map[string]any{"key": "val"},
			want:   nil,
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := StringSliceFromAny(tt.v)
			require.Equal(t, tt.wantOk, ok)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_StringSliceFromRawJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		raw    json.RawMessage
		want   []string
		wantOk bool
	}{
		{
			name:   "real JSON array",
			raw:    json.RawMessage(`["title","body"]`),
			want:   []string{"title", "body"},
			wantOk: true,
		},
		{
			name:   "JSON-encoded array embedded in a string",
			raw:    json.RawMessage(`"[\"title\",\"body\"]"`),
			want:   []string{"title", "body"},
			wantOk: true,
		},
		{
			name:   "scalar string",
			raw:    json.RawMessage(`"title"`),
			want:   []string{"title"},
			wantOk: true,
		},
		{
			name:   "empty array",
			raw:    json.RawMessage(`[]`),
			want:   []string{},
			wantOk: true,
		},
		{
			name:   "array with non-string element fails closed",
			raw:    json.RawMessage(`[1,2]`),
			want:   nil,
			wantOk: false,
		},
		{
			name:   "invalid JSON",
			raw:    json.RawMessage(`not json`),
			want:   nil,
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := StringSliceFromRawJSON(tt.raw)
			require.Equal(t, tt.wantOk, ok)
			require.Equal(t, tt.want, got)
		})
	}
}
