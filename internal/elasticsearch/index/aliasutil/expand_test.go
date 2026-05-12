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
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestExpandAliasFields_minimal(t *testing.T) {
	t.Parallel()
	f := aliasutil.AliasFields{
		Name:          types.StringValue("my-alias"),
		Filter:        jsontypes.NewNormalizedNull(),
		IndexRouting:  types.StringNull(),
		SearchRouting: types.StringNull(),
		Routing:       types.StringNull(),
		IsHidden:      types.BoolNull(),
		IsWriteIndex:  types.BoolNull(),
	}
	ia, diags := aliasutil.ExpandAliasFields(f)
	if diags.HasError() {
		t.Fatal(diags)
	}
	if ia.Name != "my-alias" {
		t.Fatalf("name: got %q", ia.Name)
	}
	if ia.Filter != nil {
		t.Fatalf("expected nil filter, got %#v", ia.Filter)
	}
	if ia.IndexRouting != "" {
		t.Fatalf("expected empty index_routing, got %q", ia.IndexRouting)
	}
	if ia.SearchRouting != "" {
		t.Fatalf("expected empty search_routing, got %q", ia.SearchRouting)
	}
	if ia.Routing != "" {
		t.Fatalf("expected empty routing, got %q", ia.Routing)
	}
	if ia.IsHidden {
		t.Fatalf("expected is_hidden false, got true")
	}
	if ia.IsWriteIndex {
		t.Fatalf("expected is_write_index false, got true")
	}
}

func TestExpandAliasFields_allSet(t *testing.T) {
	t.Parallel()
	f := aliasutil.AliasFields{
		Name:          types.StringValue("full-alias"),
		Filter:        jsontypes.NewNormalizedValue(`{"term":{"status":"active"}}`),
		IndexRouting:  types.StringValue("shard-a"),
		SearchRouting: types.StringValue("shard-b"),
		Routing:       types.StringValue("shard-c"),
		IsHidden:      types.BoolValue(true),
		IsWriteIndex:  types.BoolValue(true),
	}
	ia, diags := aliasutil.ExpandAliasFields(f)
	if diags.HasError() {
		t.Fatal(diags)
	}
	if ia.Name != "full-alias" {
		t.Fatalf("name: got %q", ia.Name)
	}
	if ia.Filter == nil {
		t.Fatal("expected non-nil filter")
	}
	if ia.IndexRouting != "shard-a" {
		t.Fatalf("index_routing: got %q", ia.IndexRouting)
	}
	if ia.SearchRouting != "shard-b" {
		t.Fatalf("search_routing: got %q", ia.SearchRouting)
	}
	if ia.Routing != "shard-c" {
		t.Fatalf("routing: got %q", ia.Routing)
	}
	if !ia.IsHidden {
		t.Fatal("expected is_hidden true")
	}
	if !ia.IsWriteIndex {
		t.Fatal("expected is_write_index true")
	}
}

func TestExpandAliasFields_unknownRoutingNotSet(t *testing.T) {
	t.Parallel()
	f := aliasutil.AliasFields{
		Name:          types.StringValue("alias"),
		Filter:        jsontypes.NewNormalizedNull(),
		IndexRouting:  types.StringUnknown(),
		SearchRouting: types.StringUnknown(),
		Routing:       types.StringUnknown(),
		IsHidden:      types.BoolUnknown(),
		IsWriteIndex:  types.BoolUnknown(),
	}
	ia, diags := aliasutil.ExpandAliasFields(f)
	if diags.HasError() {
		t.Fatal(diags)
	}
	if ia.IndexRouting != "" {
		t.Fatalf("unknown index_routing should not set field, got %q", ia.IndexRouting)
	}
	if ia.SearchRouting != "" {
		t.Fatalf("unknown search_routing should not set field, got %q", ia.SearchRouting)
	}
	if ia.Routing != "" {
		t.Fatalf("unknown routing should not set field, got %q", ia.Routing)
	}
	if ia.IsHidden {
		t.Fatal("unknown is_hidden should not set field")
	}
	if ia.IsWriteIndex {
		t.Fatal("unknown is_write_index should not set field")
	}
}

func TestExpandAliasFields_invalidFilterJSON(t *testing.T) {
	t.Parallel()
	f := aliasutil.AliasFields{
		Name:   types.StringValue("bad"),
		Filter: jsontypes.NewNormalizedValue(`{not valid json`),
	}
	_, diags := aliasutil.ExpandAliasFields(f)
	if !diags.HasError() {
		t.Fatal("expected error for invalid filter JSON")
	}
}

func TestExpandAliasFields_emptyFilterString(t *testing.T) {
	t.Parallel()
	f := aliasutil.AliasFields{
		Name:   types.StringValue("alias"),
		Filter: jsontypes.NewNormalizedValue("   "),
	}
	ia, diags := aliasutil.ExpandAliasFields(f)
	if diags.HasError() {
		t.Fatal(diags)
	}
	if ia.Filter != nil {
		t.Fatalf("expected nil filter for blank filter string, got %#v", ia.Filter)
	}
}

func TestExpandAliasFields_nullRoutingNotSet(t *testing.T) {
	t.Parallel()
	// Null routing fields must not set any value — null is equivalent to "not configured".
	f := aliasutil.AliasFields{
		Name:          types.StringValue("alias"),
		Filter:        jsontypes.NewNormalizedNull(),
		IndexRouting:  types.StringNull(),
		SearchRouting: types.StringNull(),
		Routing:       types.StringNull(),
		IsHidden:      types.BoolNull(),
		IsWriteIndex:  types.BoolNull(),
	}
	ia, diags := aliasutil.ExpandAliasFields(f)
	if diags.HasError() {
		t.Fatal(diags)
	}
	// Go zero value for string is ""; the fields must remain at zero, not be overwritten.
	if ia.IndexRouting != "" || ia.SearchRouting != "" || ia.Routing != "" {
		t.Fatalf("null routing fields must not be set: ir=%q sr=%q r=%q",
			ia.IndexRouting, ia.SearchRouting, ia.Routing)
	}
}
