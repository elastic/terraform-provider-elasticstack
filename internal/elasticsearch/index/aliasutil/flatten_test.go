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

package aliasutil_test

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/aliasutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNormalizeAliasFilterMap_nil(t *testing.T) {
	t.Parallel()
	result, diags := aliasutil.NormalizeAliasFilterMap(nil)
	if diags.HasError() {
		t.Fatal(diags)
	}
	if !result.IsNull() {
		t.Fatalf("expected null for nil map, got %q", result.ValueString())
	}
}

func TestNormalizeAliasFilterMap_empty(t *testing.T) {
	t.Parallel()
	result, diags := aliasutil.NormalizeAliasFilterMap(map[string]any{})
	if diags.HasError() {
		t.Fatal(diags)
	}
	if !result.IsNull() {
		t.Fatalf("expected null for empty map, got %q", result.ValueString())
	}
}

func TestNormalizeAliasFilterMap_simple(t *testing.T) {
	t.Parallel()
	filterMap := map[string]any{
		"term": map[string]any{"status": "active"},
	}
	result, diags := aliasutil.NormalizeAliasFilterMap(filterMap)
	if diags.HasError() {
		t.Fatal(diags)
	}
	if result.IsNull() {
		t.Fatal("expected non-null result")
	}
	got := result.ValueString()
	if got == "" {
		t.Fatal("expected non-empty JSON string")
	}
}

func TestNormalizeAliasFilterMap_normalizesExpandedForm(t *testing.T) {
	t.Parallel()
	// The typed client may produce {"term":{"field":{"value":"x"}}}
	// NormalizeQueryFilter compacts this to {"term":{"field":"x"}}
	filterMap := map[string]any{
		"term": map[string]any{
			"field": map[string]any{"value": "x"},
		},
	}
	result, diags := aliasutil.NormalizeAliasFilterMap(filterMap)
	if diags.HasError() {
		t.Fatal(diags)
	}
	if result.IsNull() {
		t.Fatal("expected non-null result")
	}
	got := result.ValueString()
	if got != `{"term":{"field":"x"}}` {
		t.Fatalf("expected compact form, got %q", got)
	}
}

func TestNormalizeAliasFilterAnyToMap_nil(t *testing.T) {
	t.Parallel()
	result, diags := aliasutil.NormalizeAliasFilterAnyToMap(nil)
	if diags.HasError() {
		t.Fatal(diags)
	}
	if result != nil {
		t.Fatalf("expected nil for nil input, got %#v", result)
	}
}

func TestNormalizeAliasFilterAnyToMap_map(t *testing.T) {
	t.Parallel()
	input := map[string]any{
		"term": map[string]any{"status": "active"},
	}
	result, diags := aliasutil.NormalizeAliasFilterAnyToMap(input)
	if diags.HasError() {
		t.Fatal(diags)
	}
	if result == nil {
		t.Fatal("expected non-nil map")
	}
	if _, ok := result["term"]; !ok {
		t.Fatalf("expected 'term' key in result, got %#v", result)
	}
}

func TestNormalizeAliasFilterFromAny_nil(t *testing.T) {
	t.Parallel()
	result, diags := aliasutil.NormalizeAliasFilterFromAny(nil)
	if diags.HasError() {
		t.Fatal(diags)
	}
	if !result.IsNull() {
		t.Fatalf("expected null for nil input, got %q", result.ValueString())
	}
}

func TestNormalizeAliasFilterFromAny_map(t *testing.T) {
	t.Parallel()
	input := map[string]any{
		"term": map[string]any{"status": "active"},
	}
	result, diags := aliasutil.NormalizeAliasFilterFromAny(input)
	if diags.HasError() {
		t.Fatal(diags)
	}
	if result.IsNull() {
		t.Fatal("expected non-null result")
	}
	got := result.ValueString()
	if got == "" {
		t.Fatal("expected non-empty JSON string")
	}
}

func TestNormalizeAliasFilterFromAny_expandedForm(t *testing.T) {
	t.Parallel()
	// Simulate what the typed ES client produces: expanded single-value form
	input := map[string]any{
		"term": map[string]any{
			"field": map[string]any{"value": "x"},
		},
	}
	result, diags := aliasutil.NormalizeAliasFilterFromAny(input)
	if diags.HasError() {
		t.Fatal(diags)
	}
	if result.IsNull() {
		t.Fatal("expected non-null result")
	}
	got := result.ValueString()
	if got != `{"term":{"field":"x"}}` {
		t.Fatalf("expected compact form, got %q", got)
	}
}

func TestAliasAttrsFromModelWithRouting_noPreservation(t *testing.T) {
	t.Parallel()
	a := models.IndexAlias{Routing: "r1"}
	attrs, diags := aliasutil.AliasAttrsFromModelWithRouting("myalias", a, map[string]string{"myalias": "preserved"})
	if diags.HasError() {
		t.Fatal(diags)
	}
	// API returned a non-empty routing value: preserved routing must be ignored.
	if v, ok := attrs["routing"]; !ok || v.String() != `"r1"` {
		t.Fatalf("expected routing=r1, got %v", attrs["routing"])
	}
}

func TestAliasAttrsFromModelWithRouting_restoresWhenAPIOmits(t *testing.T) {
	t.Parallel()
	a := models.IndexAlias{Routing: ""}
	attrs, diags := aliasutil.AliasAttrsFromModelWithRouting("myalias", a, map[string]string{"myalias": "preserved"})
	if diags.HasError() {
		t.Fatal(diags)
	}
	// API omitted routing (empty string): preserved value must be restored.
	if v, ok := attrs["routing"]; !ok || v.String() != `"preserved"` {
		t.Fatalf("expected routing=preserved, got %v", attrs["routing"])
	}
}

func TestAliasAttrsFromModelWithRouting_noPreservationEntryForAlias(t *testing.T) {
	t.Parallel()
	a := models.IndexAlias{Routing: ""}
	attrs, diags := aliasutil.AliasAttrsFromModelWithRouting("myalias", a, map[string]string{"otheralias": "preserved"})
	if diags.HasError() {
		t.Fatal(diags)
	}
	// No preserved routing for this alias name: routing stays empty string.
	if v, ok := attrs["routing"]; !ok || v.String() != `""` {
		t.Fatalf("expected routing empty string, got %v", attrs["routing"])
	}
}

func TestAliasAttrsFromModelWithRouting_nilPreservedRouting(t *testing.T) {
	t.Parallel()
	a := models.IndexAlias{Routing: ""}
	attrs, diags := aliasutil.AliasAttrsFromModelWithRouting("myalias", a, nil)
	if diags.HasError() {
		t.Fatal(diags)
	}
	// Nil preservedRouting map: routing stays empty string.
	if v, ok := attrs["routing"]; !ok || v.String() != `""` {
		t.Fatalf("expected routing empty string, got %v", attrs["routing"])
	}
}

func testAliasAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":           types.StringType,
		"filter":         jsontypes.NormalizedType{},
		"index_routing":  types.StringType,
		"routing":        types.StringType,
		"search_routing": types.StringType,
		"is_hidden":      types.BoolType,
		"is_write_index": types.BoolType,
	}
}

func TestFlattenAliasElement_basicFields(t *testing.T) {
	t.Parallel()
	a := models.IndexAlias{
		Routing:       "r1",
		IndexRouting:  "ir1",
		SearchRouting: "sr1",
		IsHidden:      true,
		IsWriteIndex:  false,
	}
	val, diags := aliasutil.FlattenAliasElement("myalias", a, nil, testAliasAttrTypes())
	if diags.HasError() {
		t.Fatal(diags)
	}
	obj, ok := val.(types.Object)
	if !ok {
		t.Fatalf("expected types.Object, got %T", val)
	}
	attrs := obj.Attributes()
	if v := attrs["name"].(types.String).ValueString(); v != "myalias" {
		t.Errorf("name: got %q, want %q", v, "myalias")
	}
	if v := attrs["routing"].(types.String).ValueString(); v != "r1" {
		t.Errorf("routing: got %q, want %q", v, "r1")
	}
	if v := attrs["is_hidden"].(types.Bool).ValueBool(); !v {
		t.Errorf("is_hidden: got false, want true")
	}
}

func TestFlattenAliasElement_preservedRoutingRestored(t *testing.T) {
	t.Parallel()
	a := models.IndexAlias{Routing: ""}
	val, diags := aliasutil.FlattenAliasElement("myalias", a, map[string]string{"myalias": "preserved"}, testAliasAttrTypes())
	if diags.HasError() {
		t.Fatal(diags)
	}
	obj := val.(types.Object)
	attrs := obj.Attributes()
	if v := attrs["routing"].(types.String).ValueString(); v != "preserved" {
		t.Errorf("routing: got %q, want %q", v, "preserved")
	}
}

func TestFlattenAliasElement_preservedRoutingIgnoredWhenAPIReturnsValue(t *testing.T) {
	t.Parallel()
	a := models.IndexAlias{Routing: "api_routing"}
	val, diags := aliasutil.FlattenAliasElement("myalias", a, map[string]string{"myalias": "preserved"}, testAliasAttrTypes())
	if diags.HasError() {
		t.Fatal(diags)
	}
	obj := val.(types.Object)
	attrs := obj.Attributes()
	if v := attrs["routing"].(types.String).ValueString(); v != "api_routing" {
		t.Errorf("routing: got %q, want %q", v, "api_routing")
	}
}

func TestFlattenAliasSet_nilAliasesReturnsNull(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	elemType := types.ObjectType{AttrTypes: testAliasAttrTypes()}
	sv, diags := aliasutil.FlattenAliasSet(ctx, nil, nil, elemType, testAliasAttrTypes())
	if diags.HasError() {
		t.Fatal(diags)
	}
	if !sv.IsNull() {
		t.Fatalf("expected null set for nil aliases, got %#v", sv)
	}
}

func TestFlattenAliasSet_emptyAliasesReturnsNull(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	elemType := types.ObjectType{AttrTypes: testAliasAttrTypes()}
	sv, diags := aliasutil.FlattenAliasSet(ctx, map[string]models.IndexAlias{}, nil, elemType, testAliasAttrTypes())
	if diags.HasError() {
		t.Fatal(diags)
	}
	if !sv.IsNull() {
		t.Fatalf("expected null set for empty aliases, got %#v", sv)
	}
}

func TestFlattenAliasSet_sortedOrder(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	attrTypes := testAliasAttrTypes()
	elemType := types.ObjectType{AttrTypes: attrTypes}
	aliases := map[string]models.IndexAlias{
		"z-alias": {Routing: "r-z"},
		"a-alias": {Routing: "r-a"},
		"m-alias": {Routing: "r-m"},
	}
	sv, diags := aliasutil.FlattenAliasSet(ctx, aliases, nil, elemType, attrTypes)
	if diags.HasError() {
		t.Fatal(diags)
	}
	if sv.IsNull() {
		t.Fatal("expected non-null set")
	}
	elems := sv.Elements()
	if len(elems) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(elems))
	}
	// Elements in a Set are not ordered the same way as insertion, but we
	// verify all names are present and routing is correctly assigned.
	names := make(map[string]string)
	for _, el := range elems {
		obj := el.(types.Object)
		attrs := obj.Attributes()
		name := attrs["name"].(types.String).ValueString()
		routing := attrs["routing"].(types.String).ValueString()
		names[name] = routing
	}
	if names["a-alias"] != "r-a" {
		t.Errorf("a-alias routing: got %q, want %q", names["a-alias"], "r-a")
	}
	if names["m-alias"] != "r-m" {
		t.Errorf("m-alias routing: got %q, want %q", names["m-alias"], "r-m")
	}
	if names["z-alias"] != "r-z" {
		t.Errorf("z-alias routing: got %q, want %q", names["z-alias"], "r-z")
	}
}

func TestFlattenAliasSet_routingPreservation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	attrTypes := testAliasAttrTypes()
	elemType := types.ObjectType{AttrTypes: attrTypes}
	aliases := map[string]models.IndexAlias{
		"a-alias": {Routing: ""},
		"b-alias": {Routing: "api-routing"},
	}
	preservedRouting := map[string]string{"a-alias": "preserved-a"}
	sv, diags := aliasutil.FlattenAliasSet(ctx, aliases, preservedRouting, elemType, attrTypes)
	if diags.HasError() {
		t.Fatal(diags)
	}
	elems := sv.Elements()
	routingMap := make(map[string]string)
	for _, el := range elems {
		obj := el.(types.Object)
		attrs := obj.Attributes()
		name := attrs["name"].(types.String).ValueString()
		routing := attrs["routing"].(types.String).ValueString()
		routingMap[name] = routing
	}
	if routingMap["a-alias"] != "preserved-a" {
		t.Errorf("a-alias routing: got %q, want preserved-a", routingMap["a-alias"])
	}
	if routingMap["b-alias"] != "api-routing" {
		t.Errorf("b-alias routing: got %q, want api-routing", routingMap["b-alias"])
	}
}
