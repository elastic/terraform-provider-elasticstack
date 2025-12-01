package alias

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAliasIndexConfig_Equals(t *testing.T) {
	tests := []struct {
		name     string
		a        AliasIndexConfig
		b        AliasIndexConfig
		expected bool
	}{
		{
			name: "identical configs",
			a: AliasIndexConfig{
				Name:          "test-index",
				IsWriteIndex:  true,
				Filter:        map[string]interface{}{"user": "admin", "status": "active"},
				IndexRouting:  "1",
				IsHidden:      false,
				Routing:       "2",
				SearchRouting: "3",
			},
			b: AliasIndexConfig{
				Name:          "test-index",
				IsWriteIndex:  true,
				Filter:        map[string]interface{}{"user": "admin", "status": "active"},
				IndexRouting:  "1",
				IsHidden:      false,
				Routing:       "2",
				SearchRouting: "3",
			},
			expected: true,
		},
		{
			name: "different name",
			a: AliasIndexConfig{
				Name:         "test-index-1",
				IsWriteIndex: true,
			},
			b: AliasIndexConfig{
				Name:         "test-index-2",
				IsWriteIndex: true,
			},
			expected: false,
		},
		{
			name: "different IsWriteIndex",
			a: AliasIndexConfig{
				Name:         "test-index",
				IsWriteIndex: true,
			},
			b: AliasIndexConfig{
				Name:         "test-index",
				IsWriteIndex: false,
			},
			expected: false,
		},
		{
			name: "different IndexRouting",
			a: AliasIndexConfig{
				Name:         "test-index",
				IndexRouting: "1",
			},
			b: AliasIndexConfig{
				Name:         "test-index",
				IndexRouting: "2",
			},
			expected: false,
		},
		{
			name: "different IsHidden",
			a: AliasIndexConfig{
				Name:     "test-index",
				IsHidden: true,
			},
			b: AliasIndexConfig{
				Name:     "test-index",
				IsHidden: false,
			},
			expected: false,
		},
		{
			name: "different Routing",
			a: AliasIndexConfig{
				Name:    "test-index",
				Routing: "route-1",
			},
			b: AliasIndexConfig{
				Name:    "test-index",
				Routing: "route-2",
			},
			expected: false,
		},
		{
			name: "different SearchRouting",
			a: AliasIndexConfig{
				Name:          "test-index",
				SearchRouting: "search-1",
			},
			b: AliasIndexConfig{
				Name:          "test-index",
				SearchRouting: "search-2",
			},
			expected: false,
		},
		{
			name: "different Filter - different values",
			a: AliasIndexConfig{
				Name:   "test-index",
				Filter: map[string]interface{}{"user": "admin"},
			},
			b: AliasIndexConfig{
				Name:   "test-index",
				Filter: map[string]interface{}{"user": "guest"},
			},
			expected: false,
		},
		{
			name: "different Filter - different keys",
			a: AliasIndexConfig{
				Name:   "test-index",
				Filter: map[string]interface{}{"user": "admin"},
			},
			b: AliasIndexConfig{
				Name:   "test-index",
				Filter: map[string]interface{}{"role": "admin"},
			},
			expected: false,
		},
		{
			name: "one nil Filter, one non-nil",
			a: AliasIndexConfig{
				Name:   "test-index",
				Filter: map[string]interface{}{"term": "value"},
			},
			b: AliasIndexConfig{
				Name:   "test-index",
				Filter: nil,
			},
			expected: false,
		},
		{
			name: "both nil Filters",
			a: AliasIndexConfig{
				Name:   "test-index",
				Filter: nil,
			},
			b: AliasIndexConfig{
				Name:   "test-index",
				Filter: nil,
			},
			expected: true,
		},
		{
			name: "both empty Filters",
			a: AliasIndexConfig{
				Name:   "test-index",
				Filter: map[string]interface{}{},
			},
			b: AliasIndexConfig{
				Name:   "test-index",
				Filter: map[string]interface{}{},
			},
			expected: true,
		},
		{
			name: "complex nested Filter match",
			a: AliasIndexConfig{
				Name:   "test-index",
				Filter: map[string]interface{}{"environment": "prod", "tier": "premium"},
			},
			b: AliasIndexConfig{
				Name:   "test-index",
				Filter: map[string]interface{}{"environment": "prod", "tier": "premium"},
			},
			expected: true,
		},
		{
			name: "all empty string fields",
			a: AliasIndexConfig{
				Name:          "test-index",
				IndexRouting:  "",
				Routing:       "",
				SearchRouting: "",
			},
			b: AliasIndexConfig{
				Name:          "test-index",
				IndexRouting:  "",
				Routing:       "",
				SearchRouting: "",
			},
			expected: true,
		},
		{
			name: "empty string vs populated string",
			a: AliasIndexConfig{
				Name:    "test-index",
				Routing: "",
			},
			b: AliasIndexConfig{
				Name:    "test-index",
				Routing: "route",
			},
			expected: false,
		},
		{
			name: "multiple fields different",
			a: AliasIndexConfig{
				Name:          "test-index-1",
				IsWriteIndex:  true,
				IndexRouting:  "1",
				SearchRouting: "search-1",
			},
			b: AliasIndexConfig{
				Name:          "test-index-2",
				IsWriteIndex:  false,
				IndexRouting:  "2",
				SearchRouting: "search-2",
			},
			expected: false,
		},
		{
			name: "fully populated identical configs",
			a: AliasIndexConfig{
				Name:          "production-index",
				IsWriteIndex:  true,
				Filter:        map[string]interface{}{"environment": "prod"},
				IndexRouting:  "prod-route",
				IsHidden:      true,
				Routing:       "main-route",
				SearchRouting: "search-route",
			},
			b: AliasIndexConfig{
				Name:          "production-index",
				IsWriteIndex:  true,
				Filter:        map[string]interface{}{"environment": "prod"},
				IndexRouting:  "prod-route",
				IsHidden:      true,
				Routing:       "main-route",
				SearchRouting: "search-route",
			},
			expected: true,
		},
		{
			name: "Filter with nested maps",
			a: AliasIndexConfig{
				Name: "test-index",
				Filter: map[string]interface{}{
					"term": map[string]interface{}{"user": "admin"},
				},
			},
			b: AliasIndexConfig{
				Name: "test-index",
				Filter: map[string]interface{}{
					"term": map[string]interface{}{"user": "admin"},
				},
			},
			expected: true,
		},
		{
			name: "Filter with slices",
			a: AliasIndexConfig{
				Name: "test-index",
				Filter: map[string]interface{}{
					"bool": map[string]interface{}{
						"must": []interface{}{
							map[string]interface{}{"term": map[string]interface{}{"status": "active"}},
						},
					},
				},
			},
			b: AliasIndexConfig{
				Name: "test-index",
				Filter: map[string]interface{}{
					"bool": map[string]interface{}{
						"must": []interface{}{
							map[string]interface{}{"term": map[string]interface{}{"status": "active"}},
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
