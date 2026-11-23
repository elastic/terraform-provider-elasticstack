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

package lenscommon

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricFormatUsesUnsupportedDurationParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		model map[string]any
		want  bool
	}{
		{
			name: "type decimals only",
			model: map[string]any{"format": map[string]any{
				"type": "duration", "from": "s", "to": "ms", "decimals": float64(0),
			}},
			want: true,
		},
		{
			name: "type compact only",
			model: map[string]any{"format": map[string]any{
				"type": "duration", "from": "s", "to": "ms", "compact": true,
			}},
			want: true,
		},
		{
			name: "type decimals and compact",
			model: map[string]any{"format": map[string]any{
				"type": "duration", "from": "s", "to": "ms", "decimals": float64(0), "compact": true,
			}},
			want: true,
		},
		{
			name: "type decimals under params",
			model: map[string]any{"format": map[string]any{
				"type": "duration", "params": map[string]any{"decimals": float64(0)},
			}},
			want: true,
		},
		{
			name: "type compact under params",
			model: map[string]any{"format": map[string]any{
				"type": "duration", "params": map[string]any{"compact": true},
			}},
			want: true,
		},
		{
			name: "legacy id decimals",
			model: map[string]any{"format": map[string]any{
				"id": "duration", "decimals": float64(1),
			}},
			want: true,
		},
		{
			name: "legacy id compact",
			model: map[string]any{"format": map[string]any{
				"id": "duration", "compact": true,
			}},
			want: true,
		},
		{
			name: "legacy id decimals and compact",
			model: map[string]any{"format": map[string]any{
				"id": "duration", "decimals": float64(0), "compact": true,
			}},
			want: true,
		},
		{
			name: "legacy id decimals under params",
			model: map[string]any{"format": map[string]any{
				"id": "duration", "params": map[string]any{"decimals": float64(1)},
			}},
			want: true,
		},
		{
			name: "legacy id compact under params",
			model: map[string]any{"format": map[string]any{
				"id": "duration", "params": map[string]any{"compact": true},
			}},
			want: true,
		},
		{
			name: "type duration without params",
			model: map[string]any{"format": map[string]any{
				"type": "duration", "from": "s", "to": "ms",
			}},
		},
		{
			name: "legacy id without params",
			model: map[string]any{"format": map[string]any{
				"id": "duration", "from": "s", "to": "ms",
			}},
		},
		{
			name: "number format",
			model: map[string]any{"format": map[string]any{
				"type": "number", "decimals": float64(0), "compact": true,
			}},
		},
		{
			name: "percent params",
			model: map[string]any{"format": map[string]any{
				"type": "percent", "params": map[string]any{"compact": true, "decimals": float64(2)},
			}},
		},
		{name: "null format", model: map[string]any{"format": nil}},
		{name: "nil model"},
		{name: "missing format", model: map[string]any{}},
		{name: "format is not a map", model: map[string]any{"format": "duration"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, MetricFormatUsesUnsupportedDurationParams(tc.model))
		})
	}
}
