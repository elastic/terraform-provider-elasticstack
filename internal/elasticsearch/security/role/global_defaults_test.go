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

package role

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPopulateGlobalPrivilegesDefaults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input map[string]any
		want  map[string]any
	}{
		{
			name:  "nil is returned as-is",
			input: nil,
			want:  nil,
		},
		{
			name: "empty role object is stripped",
			input: map[string]any{
				"application": map[string]any{},
				"profile":     map[string]any{"write": map[string]any{"applications": []any{"foo"}}},
				"role":        map[string]any{},
			},
			want: map[string]any{
				"application": map[string]any{},
				"profile":     map[string]any{"write": map[string]any{"applications": []any{"foo"}}},
			},
		},
		{
			name: "empty data_source array is stripped",
			input: map[string]any{
				"application": map[string]any{},
				"profile":     map[string]any{"write": map[string]any{"applications": []any{"foo"}}},
				"data_source": []any{},
			},
			want: map[string]any{
				"application": map[string]any{},
				"profile":     map[string]any{"write": map[string]any{"applications": []any{"foo"}}},
			},
		},
		{
			name: "non-empty data_source is preserved",
			input: map[string]any{
				"application": map[string]any{},
				"data_source": []any{"foo"},
			},
			want: map[string]any{
				"application": map[string]any{},
				"data_source": []any{"foo"},
			},
		},
		{
			name: "non-role empty object is preserved",
			input: map[string]any{
				"application": map[string]any{},
				"profile":     map[string]any{"write": map[string]any{"applications": []any{"foo"}}},
			},
			want: map[string]any{
				"application": map[string]any{},
				"profile":     map[string]any{"write": map[string]any{"applications": []any{"foo"}}},
			},
		},
		{
			name: "mixed empty and non-empty defaults",
			input: map[string]any{
				"application": map[string]any{"manage": map[string]any{"applications": []any{"bar"}}},
				"role":        map[string]any{},
				"data_source": []any{},
				"profile":     map[string]any{"write": map[string]any{"applications": []any{"foo"}}},
			},
			want: map[string]any{
				"application": map[string]any{"manage": map[string]any{"applications": []any{"bar"}}},
				"profile":     map[string]any{"write": map[string]any{"applications": []any{"foo"}}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := populateGlobalPrivilegesDefaults(tt.input)
			require.Equal(t, tt.want, got)
		})
	}
}
