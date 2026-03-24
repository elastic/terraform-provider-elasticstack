//go:build ignore

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

func newSchema() *Schema {
	return &Schema{
		Components: Map{
			"schemas": Map{},
		},
	}
}

func TestCreateRef_InlineSchema(t *testing.T) {
	schema := newSchema()
	root := Map{
		"requestBody": Map{
			"content": Map{
				"application/json": Map{
					"schema": Map{
						"properties": Map{
							"config": Map{
								"type": "object",
								"properties": Map{
									"name": Map{"type": "string"},
								},
							},
						},
					},
				},
			},
		},
	}

	ref := root.CreateRef(schema, "my_config", "requestBody.content.application/json.schema.properties.config")

	if ref["$ref"] != "#/components/schemas/my_config" {
		t.Errorf("expected $ref to my_config, got %v", ref)
	}

	component, ok := schema.Components.Get("schemas.my_config")
	if !ok {
		t.Fatal("expected component to be registered")
	}
	componentMap := component.(Map)
	if componentMap["type"] != "object" {
		t.Errorf("expected extracted component type=object, got %v", componentMap["type"])
	}

	// Verify the original location was replaced with a $ref
	replaced := root.MustGet("requestBody.content.application/json.schema.properties.config")
	replacedMap := replaced.(Map)
	if replacedMap["$ref"] != "#/components/schemas/my_config" {
		t.Errorf("expected inline schema to be replaced with $ref, got %v", replacedMap)
	}
}

func TestCreateRef_AlreadyRef(t *testing.T) {
	schema := newSchema()
	root := Map{
		"responses": Map{
			"200": Map{
				"content": Map{
					"application/json": Map{
						"schema": Map{
							"$ref": "#/components/schemas/existing_component",
						},
					},
				},
			},
		},
	}

	ref := root.CreateRef(schema, "my_component", "responses.200.content.application/json.schema")

	if ref["$ref"] != "#/components/schemas/my_component" {
		t.Errorf("expected $ref return value, got %v", ref)
	}

	// Component should NOT have been registered — target was already a $ref
	if _, ok := schema.Components.Get("schemas.my_component"); ok {
		t.Error("expected component to NOT be registered when target is already a $ref")
	}
}

func TestCreateRef_AlreadyRef_MapStringAny(t *testing.T) {
	schema := newSchema()
	// Use map[string]any instead of Map to test the second type check
	root := Map{
		"responses": Map{
			"200": map[string]any{
				"schema": map[string]any{
					"$ref": "#/components/schemas/existing_component",
				},
			},
		},
	}

	ref := root.CreateRef(schema, "my_component", "responses.200.schema")

	if ref["$ref"] != "#/components/schemas/my_component" {
		t.Errorf("expected $ref return value, got %v", ref)
	}

	if _, ok := schema.Components.Get("schemas.my_component"); ok {
		t.Error("expected component to NOT be registered when target is already a $ref (map[string]any)")
	}
}

func TestCreateRef_PathNotFound(t *testing.T) {
	schema := newSchema()
	root := Map{
		"responses": Map{
			"200": Map{
				"content": Map{},
			},
		},
	}

	// This path doesn't exist — should return gracefully, not panic
	ref := root.CreateRef(schema, "missing_schema", "responses.200.content.application/json.schema")

	if ref["$ref"] != "#/components/schemas/missing_schema" {
		t.Errorf("expected $ref return value, got %v", ref)
	}

	if _, ok := schema.Components.Get("schemas.missing_schema"); ok {
		t.Error("expected component to NOT be registered when path doesn't exist")
	}
}

func TestCreateRef_DuplicateIdenticalComponent(t *testing.T) {
	schema := newSchema()
	inlineSchema := Map{"type": "string", "description": "A name"}

	// First call: register the component
	root1 := Map{"props": Map{"name": deepCopyMap(inlineSchema)}}
	root1.CreateRef(schema, "shared_name", "props.name")

	// Second call with identical schema: should not panic
	root2 := Map{"props": Map{"name": deepCopyMap(inlineSchema)}}
	ref := root2.CreateRef(schema, "shared_name", "props.name")

	if ref["$ref"] != "#/components/schemas/shared_name" {
		t.Errorf("expected $ref return value, got %v", ref)
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
