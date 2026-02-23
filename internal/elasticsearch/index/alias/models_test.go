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

package alias

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndexConfig_Equals(t *testing.T) {
	tests := []struct {
		name     string
		a        IndexConfig
		b        IndexConfig
		expected bool
	}{
		{
			name: "identical configs",
			a: IndexConfig{
				Name:          "test-index",
				IsWriteIndex:  true,
				Filter:        map[string]any{"user": "admin", "status": "active"},
				IndexRouting:  "1",
				IsHidden:      false,
				Routing:       "2",
				SearchRouting: "3",
			},
			b: IndexConfig{
				Name:          "test-index",
				IsWriteIndex:  true,
				Filter:        map[string]any{"user": "admin", "status": "active"},
				IndexRouting:  "1",
				IsHidden:      false,
				Routing:       "2",
				SearchRouting: "3",
			},
			expected: true,
		},
		{
			name: "different name",
			a: IndexConfig{
				Name:         "test-index-1",
				IsWriteIndex: true,
			},
			b: IndexConfig{
				Name:         "test-index-2",
				IsWriteIndex: true,
			},
			expected: false,
		},
		{
			name: "different IsWriteIndex",
			a: IndexConfig{
				Name:         "test-index",
				IsWriteIndex: true,
			},
			b: IndexConfig{
				Name:         "test-index",
				IsWriteIndex: false,
			},
			expected: false,
		},
		{
			name: "different IndexRouting",
			a: IndexConfig{
				Name:         "test-index",
				IndexRouting: "1",
			},
			b: IndexConfig{
				Name:         "test-index",
				IndexRouting: "2",
			},
			expected: false,
		},
		{
			name: "different IsHidden",
			a: IndexConfig{
				Name:     "test-index",
				IsHidden: true,
			},
			b: IndexConfig{
				Name:     "test-index",
				IsHidden: false,
			},
			expected: false,
		},
		{
			name: "different Routing",
			a: IndexConfig{
				Name:    "test-index",
				Routing: "route-1",
			},
			b: IndexConfig{
				Name:    "test-index",
				Routing: "route-2",
			},
			expected: false,
		},
		{
			name: "different SearchRouting",
			a: IndexConfig{
				Name:          "test-index",
				SearchRouting: "search-1",
			},
			b: IndexConfig{
				Name:          "test-index",
				SearchRouting: "search-2",
			},
			expected: false,
		},
		{
			name: "different Filter - different values",
			a: IndexConfig{
				Name:   "test-index",
				Filter: map[string]any{"user": "admin"},
			},
			b: IndexConfig{
				Name:   "test-index",
				Filter: map[string]any{"user": "guest"},
			},
			expected: false,
		},
		{
			name: "different Filter - different keys",
			a: IndexConfig{
				Name:   "test-index",
				Filter: map[string]any{"user": "admin"},
			},
			b: IndexConfig{
				Name:   "test-index",
				Filter: map[string]any{"role": "admin"},
			},
			expected: false,
		},
		{
			name: "one nil Filter, one non-nil",
			a: IndexConfig{
				Name:   "test-index",
				Filter: map[string]any{"term": "value"},
			},
			b: IndexConfig{
				Name:   "test-index",
				Filter: nil,
			},
			expected: false,
		},
		{
			name: "both nil Filters",
			a: IndexConfig{
				Name:   "test-index",
				Filter: nil,
			},
			b: IndexConfig{
				Name:   "test-index",
				Filter: nil,
			},
			expected: true,
		},
		{
			name: "both empty Filters",
			a: IndexConfig{
				Name:   "test-index",
				Filter: map[string]any{},
			},
			b: IndexConfig{
				Name:   "test-index",
				Filter: map[string]any{},
			},
			expected: true,
		},
		{
			name: "complex nested Filter match",
			a: IndexConfig{
				Name:   "test-index",
				Filter: map[string]any{"environment": "prod", "tier": "premium"},
			},
			b: IndexConfig{
				Name:   "test-index",
				Filter: map[string]any{"environment": "prod", "tier": "premium"},
			},
			expected: true,
		},
		{
			name: "all empty string fields",
			a: IndexConfig{
				Name:          "test-index",
				IndexRouting:  "",
				Routing:       "",
				SearchRouting: "",
			},
			b: IndexConfig{
				Name:          "test-index",
				IndexRouting:  "",
				Routing:       "",
				SearchRouting: "",
			},
			expected: true,
		},
		{
			name: "empty string vs populated string",
			a: IndexConfig{
				Name:    "test-index",
				Routing: "",
			},
			b: IndexConfig{
				Name:    "test-index",
				Routing: "route",
			},
			expected: false,
		},
		{
			name: "multiple fields different",
			a: IndexConfig{
				Name:          "test-index-1",
				IsWriteIndex:  true,
				IndexRouting:  "1",
				SearchRouting: "search-1",
			},
			b: IndexConfig{
				Name:          "test-index-2",
				IsWriteIndex:  false,
				IndexRouting:  "2",
				SearchRouting: "search-2",
			},
			expected: false,
		},
		{
			name: "fully populated identical configs",
			a: IndexConfig{
				Name:          "production-index",
				IsWriteIndex:  true,
				Filter:        map[string]any{"environment": "prod"},
				IndexRouting:  "prod-route",
				IsHidden:      true,
				Routing:       "main-route",
				SearchRouting: "search-route",
			},
			b: IndexConfig{
				Name:          "production-index",
				IsWriteIndex:  true,
				Filter:        map[string]any{"environment": "prod"},
				IndexRouting:  "prod-route",
				IsHidden:      true,
				Routing:       "main-route",
				SearchRouting: "search-route",
			},
			expected: true,
		},
		{
			name: "Filter with nested maps",
			a: IndexConfig{
				Name: "test-index",
				Filter: map[string]any{
					"term": map[string]any{"user": "admin"},
				},
			},
			b: IndexConfig{
				Name: "test-index",
				Filter: map[string]any{
					"term": map[string]any{"user": "admin"},
				},
			},
			expected: true,
		},
		{
			name: "Filter with slices",
			a: IndexConfig{
				Name: "test-index",
				Filter: map[string]any{
					"bool": map[string]any{
						"must": []any{
							map[string]any{"term": map[string]any{"status": "active"}},
						},
					},
				},
			},
			b: IndexConfig{
				Name: "test-index",
				Filter: map[string]any{
					"bool": map[string]any{
						"must": []any{
							map[string]any{"term": map[string]any{"status": "active"}},
						},
					},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip tests that would panic due to maps.Equal limitations
			// if tt.name == "Filter with nested maps - demonstrates maps.Equal limitation" ||
			// 	tt.name == "Filter with slices - demonstrates maps.Equal panic" {
			// 	t.Skip("This test demonstrates the limitation of maps.Equal with uncomparable types - it would panic")
			// }
			result := tt.a.Equals(tt.b)
			assert.Equal(t, tt.expected, result, "Equals() returned unexpected result")
		})
	}
}
