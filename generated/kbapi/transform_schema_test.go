//go:build ignore

package main

import (
	"io"
	"log"
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

func TestTransformRemoveAnyOfWhenOneOfPresent(t *testing.T) {
	tests := []struct {
		name       string
		components Map
		expected   Map
	}{
		{
			name: "removes anyOf when oneOf is present",
			components: Map{
				"schemas": Map{
					"top_level": Map{
						"anyOf": Slice{
							Map{"type": "string"},
						},
						"oneOf": Slice{
							Map{"$ref": "#/components/schemas/Schema1"},
						},
					},
				},
			},
			expected: Map{
				"schemas": Map{
					"top_level": Map{
						"oneOf": Slice{
							Map{"$ref": "#/components/schemas/Schema1"},
						},
					},
				},
			},
		},
		{
			name: "keeps anyOf when oneOf is absent",
			components: Map{
				"schemas": Map{
					"any_of_only": Map{
						"anyOf": Slice{
							Map{"type": "string"},
						},
					},
					"one_of_only": Map{
						"oneOf": Slice{
							Map{"type": "number"},
						},
					},
				},
			},
			expected: Map{
				"schemas": Map{
					"any_of_only": Map{
						"anyOf": Slice{
							Map{"type": "string"},
						},
					},
					"one_of_only": Map{
						"oneOf": Slice{
							Map{"type": "number"},
						},
					},
				},
			},
		},
		{
			name: "removes nested anyOf when nested oneOf is present",
			components: Map{
				"schemas": Map{
					"nested": Map{
						"type": "object",
						"properties": Map{
							"child": Map{
								"anyOf": Slice{
									Map{"type": "integer"},
								},
								"oneOf": Slice{
									Map{"type": "number"},
								},
							},
							"unchanged": Map{
								"anyOf": Slice{
									Map{"type": "boolean"},
								},
							},
						},
					},
				},
			},
			expected: Map{
				"schemas": Map{
					"nested": Map{
						"type": "object",
						"properties": Map{
							"child": Map{
								"oneOf": Slice{
									Map{"type": "number"},
								},
							},
							"unchanged": Map{
								"anyOf": Slice{
									Map{"type": "boolean"},
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
			schema := &Schema{
				Components: deepCopyMap(tt.components),
			}

			transformRemoveAnyOfWhenOneOfPresent(schema)

			if !reflect.DeepEqual(schema.Components, tt.expected) {
				t.Errorf("transformRemoveAnyOfWhenOneOfPresent() =\n%+v\nwant:\n%+v", schema.Components, tt.expected)
			}
		})
	}
}

func TestCreateRefCreatesComponentFromMapField(t *testing.T) {
	schema := &Schema{
		Components: Map{
			"schemas": Map{},
		},
	}
	input := Map{
		"properties": Map{
			"child": Map{
				"type": "string",
			},
		},
	}

	ref := input.CreateRef(schema, "Child", "properties.child")

	expectedRef := Map{"$ref": "#/components/schemas/Child"}
	expectedInput := Map{
		"properties": Map{
			"child": expectedRef,
		},
	}
	expectedComponents := Map{
		"schemas": Map{
			"Child": Map{
				"type": "string",
			},
		},
	}

	if !reflect.DeepEqual(ref, expectedRef) {
		t.Fatalf("CreateRef() returned %+v, want %+v", ref, expectedRef)
	}
	if !reflect.DeepEqual(input, expectedInput) {
		t.Fatalf("CreateRef() mutated input to %+v, want %+v", input, expectedInput)
	}
	if !reflect.DeepEqual(schema.Components, expectedComponents) {
		t.Fatalf("CreateRef() wrote components %+v, want %+v", schema.Components, expectedComponents)
	}
}

func TestCreateRefCreatesComponentFromSliceElement(t *testing.T) {
	schema := &Schema{
		Components: Map{
			"schemas": Map{},
		},
	}
	input := Map{
		"oneOf": Slice{
			Map{"type": "string"},
			Map{"type": "number"},
		},
	}

	ref := input.CreateRef(schema, "NumberVariant", "oneOf.1")

	expectedRef := Map{"$ref": "#/components/schemas/NumberVariant"}
	expectedInput := Map{
		"oneOf": Slice{
			Map{"type": "string"},
			expectedRef,
		},
	}
	expectedComponents := Map{
		"schemas": Map{
			"NumberVariant": Map{"type": "number"},
		},
	}

	if !reflect.DeepEqual(ref, expectedRef) {
		t.Fatalf("CreateRef() returned %+v, want %+v", ref, expectedRef)
	}
	if !reflect.DeepEqual(input, expectedInput) {
		t.Fatalf("CreateRef() mutated input to %+v, want %+v", input, expectedInput)
	}
	if !reflect.DeepEqual(schema.Components, expectedComponents) {
		t.Fatalf("CreateRef() wrote components %+v, want %+v", schema.Components, expectedComponents)
	}
}

func TestCreateRefReusesExistingEquivalentSliceComponent(t *testing.T) {
	existingChoice := []any{
		map[string]any{"type": "string"},
		map[string]any{"type": "number"},
	}
	schema := &Schema{
		Components: Map{
			"schemas": map[string]any{
				"Choice": existingChoice,
			},
		},
	}
	input := Map{
		"oneOf": Slice{
			Map{"type": "string"},
			Map{"type": "number"},
		},
	}

	ref := input.CreateRef(schema, "Choice", "oneOf")

	expectedRef := Map{"$ref": "#/components/schemas/Choice"}
	expectedInput := Map{
		"oneOf": expectedRef,
	}

	if !reflect.DeepEqual(ref, expectedRef) {
		t.Fatalf("CreateRef() returned %+v, want %+v", ref, expectedRef)
	}
	if !reflect.DeepEqual(input, expectedInput) {
		t.Fatalf("CreateRef() mutated input to %+v, want %+v", input, expectedInput)
	}
	gotChoice := schema.Components.MustGet("schemas.Choice")
	if !reflect.DeepEqual(gotChoice, existingChoice) {
		t.Fatalf("CreateRef() rewrote equivalent component to %+v, want %+v", gotChoice, existingChoice)
	}
	if _, ok := gotChoice.([]any); !ok {
		t.Fatalf("CreateRef() rewrote equivalent component type to %T, want []any", gotChoice)
	}
}

func TestCreateRefPanicsWhenExistingComponentDiffers(t *testing.T) {
	schema := &Schema{
		Components: Map{
			"schemas": Map{
				"Choice": Map{"type": "number"},
			},
		},
	}
	input := Map{
		"oneOf": Slice{
			Map{"type": "string"},
		},
	}
	expectedInput := deepCopyMap(input)
	expectedComponents := deepCopyMap(schema.Components)
	originalLogWriter := log.Writer()
	log.SetOutput(io.Discard)
	defer log.SetOutput(originalLogWriter)

	defer func() {
		if recover() == nil {
			t.Fatal("CreateRef() did not panic for conflicting component schema")
		}
		if !reflect.DeepEqual(input, expectedInput) {
			t.Fatalf("CreateRef() mutated input to %+v before panic, want %+v", input, expectedInput)
		}
		if !reflect.DeepEqual(schema.Components, expectedComponents) {
			t.Fatalf("CreateRef() mutated components to %+v before panic, want %+v", schema.Components, expectedComponents)
		}
	}()

	input.CreateRef(schema, "Choice", "oneOf")
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
