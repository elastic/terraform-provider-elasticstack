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

func TestFixDashboardPanelItemRefsExtractsCurrentSpecShape(t *testing.T) {
	visRef := "#/components/schemas/Kibana_HTTP_APIs_kbn-dashboard-panel-type-vis"
	sectionRef := "#/components/schemas/Kibana_HTTP_APIs_kbn-dashboard-section"
	filterRef := "#/components/schemas/Kibana_HTTP_APIs_kbn-as-code-filters-schema_asCodeFilterSchema"
	panelUnion := Map{
		"discriminator": Map{
			"propertyName": "type",
			"mapping": Map{
				"vis": visRef,
			},
		},
		"oneOf": Slice{
			Map{"$ref": visRef},
		},
	}
	panels := Map{
		"description": "Panels and sections.",
		"items": Map{
			"anyOf": Slice{
				panelUnion,
				Map{"$ref": sectionRef},
			},
		},
		"type": "array",
	}
	filters := Map{
		"description": "Filters.",
		"items":       Map{"$ref": filterRef},
		"type":        "array",
	}
	pinnedPanels := Map{
		"description": "Pinned panels.",
		"items":       Map{"type": "object"},
		"type":        "array",
	}

	schema := &Schema{
		Components: Map{
			"schemas": Map{
				"Kibana_HTTP_APIs_kbn-dashboard-data": Map{
					"properties": Map{
						"filters":       filters,
						"panels":        panels,
						"pinned_panels": pinnedPanels,
					},
				},
				"Kibana_HTTP_APIs_kbn-dashboard-section": Map{
					"properties": Map{
						"panels": Map{
							"items": Map{
								"oneOf": Slice{Map{"$ref": visRef}},
							},
						},
					},
				},
			},
		},
		Paths: map[string]*Path{
			"/api/dashboards/{id}": {
				Put: Map{
					"requestBody": Map{
						"content": Map{
							"application/json": Map{
								"schema": Map{
									"properties": Map{
										"filters":       Map{"type": "array"},
										"panels":        Map{"type": "array"},
										"pinned_panels": Map{"type": "array"},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	fixDashboardPanelItemRefs(schema)

	gotPanelItem := schema.Components.MustGetMap("schemas.dashboard_panel_item")
	if !reflect.DeepEqual(gotPanelItem, panelUnion) {
		t.Fatalf("dashboard_panel_item = %#v, want %#v", gotPanelItem, panelUnion)
	}

	wantDataPanels := Map{
		"description": "Panels and sections.",
		"items": Map{
			"anyOf": Slice{
				Map{"$ref": "#/components/schemas/dashboard_panel_item"},
				Map{"$ref": sectionRef},
			},
		},
		"type": "array",
	}
	gotDataPanelsRef := schema.Components.MustGetMap("schemas.Kibana_HTTP_APIs_kbn-dashboard-data.properties.panels")
	if !reflect.DeepEqual(gotDataPanelsRef, Map{"$ref": "#/components/schemas/dashboard_panels"}) {
		t.Fatalf("data.panels = %#v, want $ref dashboard_panels", gotDataPanelsRef)
	}

	gotSectionItems := schema.Components.MustGetMap("schemas.Kibana_HTTP_APIs_kbn-dashboard-section.properties.panels.items")
	wantSectionItems := Map{"$ref": "#/components/schemas/dashboard_panel_item"}
	if !reflect.DeepEqual(gotSectionItems, wantSectionItems) {
		t.Fatalf("section.panels.items = %#v, want %#v", gotSectionItems, wantSectionItems)
	}

	if !reflect.DeepEqual(schema.Components.MustGetMap("schemas.dashboard_panels"), wantDataPanels) {
		t.Fatalf("dashboard_panels = %#v, want %#v", schema.Components.MustGet("schemas.dashboard_panels"), wantDataPanels)
	}
	if !reflect.DeepEqual(schema.Components.MustGetMap("schemas.dashboard_filters"), filters) {
		t.Fatalf("dashboard_filters = %#v, want %#v", schema.Components.MustGet("schemas.dashboard_filters"), filters)
	}
	if !reflect.DeepEqual(schema.Components.MustGetMap("schemas.dashboard_pinned_panels"), pinnedPanels) {
		t.Fatalf("dashboard_pinned_panels = %#v, want %#v", schema.Components.MustGet("schemas.dashboard_pinned_panels"), pinnedPanels)
	}

	put := schema.MustGetPath("/api/dashboards/{id}").Put
	for _, name := range []string{"panels", "filters", "pinned_panels"} {
		got := put.MustGetMap("requestBody.content.application/json.schema.properties." + name)
		want := Map{"$ref": "#/components/schemas/dashboard_" + name}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("PUT %s = %#v, want %#v", name, got, want)
		}
	}
}

func TestFixVisByValueConfigFlattensAndRenamesLeaves(t *testing.T) {
	xyNoESQLRef := "#/components/schemas/Kibana_HTTP_APIs_xyChartNoESQL"
	xyChartRef := "#/components/schemas/Kibana_HTTP_APIs_xyChart"
	lensRef := "#/components/schemas/Kibana_HTTP_APIs_lensApiConfig"
	schema := &Schema{
		Components: Map{
			"schemas": Map{
				"Kibana_HTTP_APIs_xyChartNoESQL": Map{
					"type": "object",
					"properties": Map{
						"title": Map{"type": "string"},
						"type":  Map{"enum": Slice{"xy"}, "type": "string"},
					},
				},
				"Kibana_HTTP_APIs_xyChart": Map{
					"anyOf": Slice{Map{"$ref": xyNoESQLRef}},
				},
				"Kibana_HTTP_APIs_lensApiConfig": Map{
					"anyOf": Slice{Map{"$ref": xyChartRef}},
				},
				"Kibana_HTTP_APIs_kbn-dashboard-panel-type-vis": Map{
					"properties": Map{
						"config": Map{
							"anyOf": Slice{
								Map{
									"allOf": Slice{
										Map{"$ref": lensRef},
										Map{
											"type": "object",
											"properties": Map{
												"drilldowns": Map{"type": "array"},
												"hide_title": Map{"type": "boolean"},
												"title":      Map{"type": "string", "description": "panel title"},
											},
										},
									},
								},
								Map{"type": "object"},
							},
						},
					},
				},
			},
		},
	}

	fixVisByValueConfig(schema)

	if _, ok := schema.Components.Get("schemas.Kibana_HTTP_APIs_xyChartNoESQL"); ok {
		t.Fatal("expected xyChartNoESQL to be renamed")
	}
	leaf := schema.Components.MustGetMap("schemas.Kibana_HTTP_APIs_xyChartNoESQLByValuePanel")
	props := leaf.MustGetMap("properties")
	if _, ok := props["drilldowns"]; !ok {
		t.Fatalf("leaf missing chrome drilldowns: %#v", props)
	}
	if _, ok := props["hide_title"]; !ok {
		t.Fatalf("leaf missing chrome hide_title: %#v", props)
	}
	if desc, _ := props.MustGetMap("title")["description"]; desc == "panel title" {
		t.Fatal("chrome title should not overwrite the chart title")
	}

	gotLens := schema.Components.MustGetSlice("schemas.Kibana_HTTP_APIs_lensApiConfig.anyOf")
	wantLens := Slice{Map{"$ref": "#/components/schemas/Kibana_HTTP_APIs_xyChartNoESQLByValuePanel"}}
	if !reflect.DeepEqual(gotLens, wantLens) {
		t.Fatalf("lensApiConfig.anyOf = %#v, want %#v", gotLens, wantLens)
	}

	gotByValue := schema.Components.MustGetMap("schemas.Kibana_HTTP_APIs_kbn-dashboard-panel-type-vis.properties.config.anyOf.0")
	wantByValue := Map{"anyOf": wantLens}
	if !reflect.DeepEqual(gotByValue, wantByValue) {
		t.Fatalf("vis config.anyOf.0 = %#v, want %#v", gotByValue, wantByValue)
	}

	fixVisByValueConfig(schema)
	if !reflect.DeepEqual(schema.Components.MustGetMap("schemas.Kibana_HTTP_APIs_kbn-dashboard-panel-type-vis.properties.config.anyOf.0"), wantByValue) {
		t.Fatal("expected fixVisByValueConfig to be idempotent")
	}
}

func TestTransformUnwrapSingleBranchAnyOf(t *testing.T) {
	schema := &Schema{
		Components: Map{
			"schemas": Map{
				"Embeddable": Map{
					"properties": Map{
						"drilldowns": Map{
							"type": "array",
							"items": Map{
								"anyOf": Slice{
									Map{
										"title": "url_drilldown",
										"allOf": Slice{
											Map{"type": "object", "properties": Map{"url": Map{"type": "string"}}},
											Map{"type": "object", "properties": Map{"label": Map{"type": "string"}}},
										},
									},
								},
							},
						},
					},
				},
				"KeepUnion": Map{
					"properties": Map{
						"drilldowns": Map{
							"items": Map{
								"anyOf": Slice{
									Map{"type": "object"},
									Map{"type": "string"},
								},
							},
						},
					},
				},
			},
		},
	}

	transformUnwrapSingleBranchAnyOf(schema)

	got := schema.Components.MustGetMap("schemas.Embeddable.properties.drilldowns.items")
	if _, ok := got["anyOf"]; ok {
		t.Fatalf("single-branch drilldowns.items still has anyOf: %#v", got)
	}
	if got["title"] != "url_drilldown" {
		t.Fatalf("items.title = %#v, want url_drilldown", got["title"])
	}
	if _, ok := got.GetSlice("allOf"); !ok {
		t.Fatalf("items missing allOf after unwrap: %#v", got)
	}

	keep := schema.Components.MustGetMap("schemas.KeepUnion.properties.drilldowns.items")
	if _, ok := keep["anyOf"]; !ok {
		t.Fatal("multi-branch drilldowns.items should stay anyOf")
	}
}

func TestTransformRemoveAllOfObjectDefaults(t *testing.T) {
	schema := &Schema{
		Components: Map{
			"schemas": Map{
				"Drilldown": Map{
					"allOf": Slice{
						Map{
							"type": "object",
							"default": Map{
								"open_in_new_tab": false,
							},
							"properties": Map{
								"open_in_new_tab": Map{"type": "boolean", "default": false},
							},
						},
						Map{
							"type": "object",
							"properties": Map{
								"dashboard_id": Map{"type": "string"},
							},
						},
					},
				},
			},
		},
	}

	transformRemoveAllOfObjectDefaults(schema)

	got := schema.Components.MustGetMap("schemas.Drilldown.allOf.0")
	if _, ok := got["default"]; ok {
		t.Fatalf("expected object default to be removed, got %#v", got)
	}
	propDefault := got.MustGet("properties.open_in_new_tab.default")
	if propDefault != false {
		t.Fatalf("property default = %#v, want false", propDefault)
	}
}

func TestTransformCollapseStringEnumAnyOf(t *testing.T) {
	schema := &Schema{
		Components: Map{
			"schemas": Map{
				"TimeRange": Map{
					"properties": Map{
						"mode": Map{
							"anyOf": Slice{
								Map{"enum": Slice{"absolute"}, "type": "string"},
								Map{"enum": Slice{"relative"}, "type": "string"},
							},
						},
					},
				},
			},
		},
	}

	transformCollapseStringEnumAnyOf(schema)

	got := schema.Components.MustGetMap("schemas.TimeRange.properties.mode")
	if _, ok := got["anyOf"]; ok {
		t.Fatalf("mode still has anyOf: %#v", got)
	}
	if got.MustGet("type") != "string" {
		t.Fatalf("mode.type = %#v, want string", got.MustGet("type"))
	}
	if !reflect.DeepEqual(got.MustGetSlice("enum"), Slice{"absolute", "relative"}) {
		t.Fatalf("mode.enum = %#v, want [absolute relative]", got.MustGet("enum"))
	}
}

func TestTransformUnwrapAllOfContainingUnion(t *testing.T) {
	countRef := "#/components/schemas/countMetric"
	schema := &Schema{
		Components: Map{
			"schemas": Map{
				"MetricItem": Map{
					"allOf": Slice{
						Map{
							"anyOf": Slice{
								Map{"$ref": countRef},
								Map{"type": "object"},
							},
						},
						Map{
							"type": "object",
							"properties": Map{
								"color": Map{"type": "string"},
							},
						},
					},
				},
			},
		},
	}

	transformUnwrapAllOfContainingUnion(schema)

	got := schema.Components.MustGetMap("schemas.MetricItem")
	if _, ok := got["allOf"]; ok {
		t.Fatalf("allOf still present: %#v", got)
	}
	anyOf, ok := got.GetSlice("anyOf")
	if !ok || len(anyOf) != 2 {
		t.Fatalf("anyOf = %#v", got)
	}
}

func TestFixControlValuesSourceSchemas(t *testing.T) {
	fieldRef := "#/components/schemas/Kibana_HTTP_APIs_kbn-controls-schemas-options-list-dsl-control-schema-field"
	esqlRef := "#/components/schemas/Kibana_HTTP_APIs_kbn-controls-schemas-options-list-dsl-control-schema-esql"
	schema := &Schema{
		Components: Map{
			"schemas": Map{
				"Kibana_HTTP_APIs_kbn-controls-schemas-options-list-dsl-control-schema-field": Map{
					"properties": Map{
						"values_source": Map{
							"enum":    Slice{"field"},
							"type":    "string",
							"default": "field",
						},
					},
				},
				"Kibana_HTTP_APIs_kbn-controls-schemas-options-list-dsl-control-schema-esql": Map{
					"properties": Map{
						"values_source": Map{
							"enum": Slice{"esql"},
							"type": "string",
						},
					},
				},
				"ControlConfig": Map{
					"discriminator": Map{
						"propertyName": "values_source",
						"mapping": Map{
							"esql": esqlRef,
						},
					},
					"oneOf": Slice{
						Map{"$ref": esqlRef},
						Map{"$ref": fieldRef},
					},
				},
			},
		},
	}

	fixControlValuesSourceSchemas(schema)

	if schema.Components.MustGetMap("schemas.ControlConfig").Has("discriminator") {
		t.Fatal("expected values_source discriminator to be removed")
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
