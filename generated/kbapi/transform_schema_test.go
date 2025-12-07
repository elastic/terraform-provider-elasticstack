//go:build ignore
// +build ignore

package main

import (
	"reflect"
	"testing"
)

func TestRemoveDuplicateOneOfRefsFromNode(t *testing.T) {
	tests := []struct {
		name     string
		input    Map
		expected Map
	}{
		{
			name: "no oneOf field",
			input: Map{
				"type": "object",
				"properties": Map{
					"field1": Map{"type": "string"},
				},
			},
			expected: Map{
				"type": "object",
				"properties": Map{
					"field1": Map{"type": "string"},
				},
			},
		},
		{
			name: "oneOf with no duplicates",
			input: Map{
				"oneOf": Slice{
					Map{"$ref": "#/components/schemas/Schema1"},
					Map{"$ref": "#/components/schemas/Schema2"},
					Map{"$ref": "#/components/schemas/Schema3"},
				},
			},
			expected: Map{
				"oneOf": Slice{
					Map{"$ref": "#/components/schemas/Schema1"},
					Map{"$ref": "#/components/schemas/Schema2"},
					Map{"$ref": "#/components/schemas/Schema3"},
				},
			},
		},
		{
			name: "oneOf with duplicate refs",
			input: Map{
				"oneOf": Slice{
					Map{"$ref": "#/components/schemas/Schema1"},
					Map{"$ref": "#/components/schemas/Schema2"},
					Map{"$ref": "#/components/schemas/Schema1"},
					Map{"$ref": "#/components/schemas/Schema3"},
					Map{"$ref": "#/components/schemas/Schema2"},
				},
			},
			expected: Map{
				"oneOf": Slice{
					Map{"$ref": "#/components/schemas/Schema1"},
					Map{"$ref": "#/components/schemas/Schema2"},
					Map{"$ref": "#/components/schemas/Schema3"},
				},
			},
		},
		{
			name: "oneOf with all duplicates",
			input: Map{
				"oneOf": Slice{
					Map{"$ref": "#/components/schemas/Schema1"},
					Map{"$ref": "#/components/schemas/Schema1"},
					Map{"$ref": "#/components/schemas/Schema1"},
				},
			},
			expected: Map{
				"oneOf": Slice{
					Map{"$ref": "#/components/schemas/Schema1"},
				},
			},
		},
		{
			name: "oneOf with non-ref items",
			input: Map{
				"oneOf": Slice{
					Map{"type": "string"},
					Map{"$ref": "#/components/schemas/Schema1"},
					Map{"type": "number"},
					Map{"$ref": "#/components/schemas/Schema1"},
				},
			},
			expected: Map{
				"oneOf": Slice{
					Map{"type": "string"},
					Map{"$ref": "#/components/schemas/Schema1"},
					Map{"type": "number"},
				},
			},
		},
		{
			name: "oneOf with mixed items including duplicates",
			input: Map{
				"oneOf": Slice{
					Map{"$ref": "#/components/schemas/Schema1"},
					Map{"type": "string"},
					Map{"$ref": "#/components/schemas/Schema2"},
					Map{"type": "number"},
					Map{"$ref": "#/components/schemas/Schema1"},
					Map{"$ref": "#/components/schemas/Schema3"},
				},
			},
			expected: Map{
				"oneOf": Slice{
					Map{"$ref": "#/components/schemas/Schema1"},
					Map{"type": "string"},
					Map{"$ref": "#/components/schemas/Schema2"},
					Map{"type": "number"},
					Map{"$ref": "#/components/schemas/Schema3"},
				},
			},
		},
		{
			name: "oneOf with non-string ref value",
			input: Map{
				"oneOf": Slice{
					Map{"$ref": 123},
					Map{"$ref": "#/components/schemas/Schema1"},
				},
			},
			expected: Map{
				"oneOf": Slice{
					Map{"$ref": 123},
					Map{"$ref": "#/components/schemas/Schema1"},
				},
			},
		},
		{
			name: "oneOf with non-map items",
			input: Map{
				"oneOf": Slice{
					"string-item",
					Map{"$ref": "#/components/schemas/Schema1"},
					42,
				},
			},
			expected: Map{
				"oneOf": Slice{
					"string-item",
					Map{"$ref": "#/components/schemas/Schema1"},
					42,
				},
			},
		},
		{
			name: "nested properties with oneOf and duplicates",
			input: Map{
				"type": "object",
				"properties": Map{
					"field1": Map{
						"oneOf": Slice{
							Map{"$ref": "#/components/schemas/Schema1"},
							Map{"$ref": "#/components/schemas/Schema1"},
						},
					},
					"field2": Map{
						"type": "string",
					},
				},
			},
			expected: Map{
				"type": "object",
				"properties": Map{
					"field1": Map{
						"oneOf": Slice{
							Map{"$ref": "#/components/schemas/Schema1"},
						},
					},
					"field2": Map{
						"type": "string",
					},
				},
			},
		},
		{
			name: "deeply nested properties with duplicates",
			input: Map{
				"type": "object",
				"properties": Map{
					"level1": Map{
						"type": "object",
						"properties": Map{
							"level2": Map{
								"oneOf": Slice{
									Map{"$ref": "#/components/schemas/Schema1"},
									Map{"$ref": "#/components/schemas/Schema2"},
									Map{"$ref": "#/components/schemas/Schema1"},
								},
							},
						},
					},
				},
			},
			expected: Map{
				"type": "object",
				"properties": Map{
					"level1": Map{
						"type": "object",
						"properties": Map{
							"level2": Map{
								"oneOf": Slice{
									Map{"$ref": "#/components/schemas/Schema1"},
									Map{"$ref": "#/components/schemas/Schema2"},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "empty oneOf slice",
			input: Map{
				"oneOf": Slice{},
			},
			expected: Map{
				"oneOf": Slice{},
			},
		},
		{
			name: "properties without oneOf",
			input: Map{
				"type": "object",
				"properties": Map{
					"field1": Map{"type": "string"},
					"field2": Map{"type": "number"},
				},
			},
			expected: Map{
				"type": "object",
				"properties": Map{
					"field1": Map{"type": "string"},
					"field2": Map{"type": "number"},
				},
			},
		},
		{
			name: "multiple levels with multiple oneOf duplicates",
			input: Map{
				"oneOf": Slice{
					Map{"$ref": "#/components/schemas/Schema1"},
					Map{"$ref": "#/components/schemas/Schema1"},
				},
				"properties": Map{
					"prop1": Map{
						"oneOf": Slice{
							Map{"$ref": "#/components/schemas/Schema2"},
							Map{"$ref": "#/components/schemas/Schema3"},
							Map{"$ref": "#/components/schemas/Schema2"},
						},
					},
					"prop2": Map{
						"properties": Map{
							"nested": Map{
								"oneOf": Slice{
									Map{"$ref": "#/components/schemas/Schema4"},
									Map{"$ref": "#/components/schemas/Schema4"},
									Map{"$ref": "#/components/schemas/Schema4"},
								},
							},
						},
					},
				},
			},
			expected: Map{
				"oneOf": Slice{
					Map{"$ref": "#/components/schemas/Schema1"},
				},
				"properties": Map{
					"prop1": Map{
						"oneOf": Slice{
							Map{"$ref": "#/components/schemas/Schema2"},
							Map{"$ref": "#/components/schemas/Schema3"},
						},
					},
					"prop2": Map{
						"properties": Map{
							"nested": Map{
								"oneOf": Slice{
									Map{"$ref": "#/components/schemas/Schema4"},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a deep copy of input to ensure function modifies in place
			inputCopy := deepCopyMap(tt.input)

			removeDuplicateOneOfRefsFromNode("", inputCopy)

			if !reflect.DeepEqual(inputCopy, tt.expected) {
				t.Errorf("removeDuplicateOneOfRefsFromNode() =\n%+v\nwant:\n%+v", inputCopy, tt.expected)
			}
		})
	}
}

// deepCopyMap creates a deep copy of a Map for testing purposes
func deepCopyMap(m Map) Map {
	result := make(Map)
	for k, v := range m {
		result[k] = deepCopyValue(v)
	}
	return result
}

func deepCopyValue(v any) any {
	switch val := v.(type) {
	case Map:
		return deepCopyMap(val)
	case map[string]any:
		return deepCopyMap(Map(val))
	case Slice:
		result := make(Slice, len(val))
		for i, item := range val {
			result[i] = deepCopyValue(item)
		}
		return result
	case []any:
		result := make(Slice, len(val))
		for i, item := range val {
			result[i] = deepCopyValue(item)
		}
		return result
	default:
		// For primitive types, return as-is
		return v
	}
}
