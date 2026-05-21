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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/aliasutil"
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
