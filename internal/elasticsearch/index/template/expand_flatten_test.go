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

package template

import (
	"context"
	"encoding/json"
	"testing"

	esindex "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/aliasutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/datastreamoptions"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestExpandTemplate_minimal(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	patterns, diags := types.SetValueFrom(ctx, types.StringType, []attr.Value{types.StringValue("logs-*")})
	if diags.HasError() {
		t.Fatal(diags)
	}
	m := Model{
		Name:                            types.StringValue("mytpl"),
		ComposedOf:                      types.ListNull(types.StringType),
		IgnoreMissingComponentTemplates: types.ListNull(types.StringType),
		IndexPatterns:                   patterns,
		Metadata:                        jsontypes.NewNormalizedNull(),
		Priority:                        types.Int64Null(),
		Version:                         types.Int64Null(),
		DataStream:                      types.ObjectNull(DataStreamAttrTypes()),
		Template:                        types.ObjectNull(TemplateAttrTypes()),
	}
	out, diags := m.toAPIModel(ctx)
	if diags.HasError() {
		t.Fatal(diags)
	}
	if out.Name != "mytpl" {
		t.Fatalf("name: got %q", out.Name)
	}
	if len(out.IndexPatterns) != 1 || out.IndexPatterns[0] != "logs-*" {
		t.Fatalf("index_patterns: %#v", out.IndexPatterns)
	}
	if out.Template != nil || out.DataStream != nil {
		t.Fatalf("expected nil optional blocks, template=%v data_stream=%v", out.Template, out.DataStream)
	}
}

func TestFlattenIndexTemplate_minimalRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pr64 := int64(42)
	ver64 := int64(7)
	tpl := &models.IndexTemplate{
		ComposedOf:                      []string{"a"},
		IgnoreMissingComponentTemplates: []string{"missing"},
		IndexPatterns:                   []string{"ix-*"},
		Meta:                            map[string]any{"k": "v"},
		Priority:                        &pr64,
		Version:                         &ver64,
	}
	var m Model
	diags := m.fromAPIModel(ctx, "tname", tpl)
	if diags.HasError() {
		t.Fatal(diags)
	}
	if m.Name.ValueString() != "tname" {
		t.Fatalf("name %q", m.Name.ValueString())
	}
	api, diags := m.toAPIModel(ctx)
	if diags.HasError() {
		t.Fatal(diags)
	}
	if api.Name != "tname" {
		t.Fatalf("expand name %q", api.Name)
	}
	if len(api.ComposedOf) != 1 || api.ComposedOf[0] != "a" {
		t.Fatalf("composed_of %#v", api.ComposedOf)
	}
	if len(api.IgnoreMissingComponentTemplates) != 1 || api.IgnoreMissingComponentTemplates[0] != "missing" {
		t.Fatalf("ignore_missing %#v", api.IgnoreMissingComponentTemplates)
	}
	if len(api.IndexPatterns) != 1 || api.IndexPatterns[0] != "ix-*" {
		t.Fatalf("index_patterns %#v", api.IndexPatterns)
	}
	if api.Meta == nil || api.Meta["k"] != "v" {
		t.Fatalf("meta %#v", api.Meta)
	}
	if api.Priority == nil || *api.Priority != 42 {
		t.Fatalf("priority %#v", api.Priority)
	}
	if api.Version == nil || *api.Version != 7 {
		t.Fatalf("version %#v", api.Version)
	}
}

// TestFlattenIndexTemplate_preservesUnmodeledSettings is a regression test for
// https://github.com/elastic/terraform-provider-elasticstack/issues/3124. The
// raw decoder must preserve every settings sub-key the API returns, even when
// the typed go-elasticsearch SlowlogSettings struct lacks the corresponding
// field (e.g. include.user).
func TestFlattenIndexTemplate_preservesUnmodeledSettings(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tpl := &models.IndexTemplate{
		IndexPatterns: []string{"ix-*"},
		Template: &models.Template{
			Settings: map[string]any{
				"index": map[string]any{
					"number_of_shards":   "1",
					"number_of_replicas": "0",
					"search": map[string]any{
						"slowlog": map[string]any{
							"include": map[string]any{
								"user": "true",
							},
						},
					},
					"lifecycle": map[string]any{
						"parse_origination_date": "true",
					},
				},
			},
		},
	}
	var m Model
	diags := m.fromAPIModel(ctx, "tname", tpl)
	if diags.HasError() {
		t.Fatal(diags)
	}
	if m.Template.IsNull() || m.Template.IsUnknown() {
		t.Fatalf("expected template object, got %#v", m.Template)
	}
	settingsAttr := m.Template.Attributes()["settings"]
	settings, ok := settingsAttr.(customtypes.IndexSettingsValue)
	if !ok {
		t.Fatalf("expected IndexSettingsValue, got %T", settingsAttr)
	}
	got := settings.ValueString()

	// Round-trip the JSON to a generic map so key order does not matter.
	var gotMap map[string]any
	if err := json.Unmarshal([]byte(got), &gotMap); err != nil {
		t.Fatalf("settings is not valid JSON: %v (raw=%q)", err, got)
	}
	idx, _ := gotMap["index"].(map[string]any)
	search, _ := idx["search"].(map[string]any)
	slowlog, _ := search["slowlog"].(map[string]any)
	include, ok := slowlog["include"].(map[string]any)
	if !ok {
		t.Fatalf("expected index.search.slowlog.include to survive flatten, got settings=%q", got)
	}
	if include["user"] != "true" {
		t.Fatalf("expected include.user to survive flatten as string \"true\", got %#v in settings=%q", include["user"], got)
	}
	lc, _ := idx["lifecycle"].(map[string]any)
	if lc == nil || lc["parse_origination_date"] != "true" {
		t.Fatalf("expected lifecycle.parse_origination_date to survive flatten, got %#v in settings=%q", lc, got)
	}
}

func TestExpandTemplate_dataStreamAllowCustomRoutingOnlyWhenTrue(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	patterns, diags := types.SetValueFrom(ctx, types.StringType, []attr.Value{types.StringValue("p")})
	if diags.HasError() {
		t.Fatal(diags)
	}
	hidden := true
	acrFalse := false
	dsObj, diags := types.ObjectValue(DataStreamAttrTypes(), map[string]attr.Value{
		"hidden":               types.BoolValue(hidden),
		"allow_custom_routing": types.BoolValue(acrFalse),
	})
	if diags.HasError() {
		t.Fatal(diags)
	}
	m := Model{
		Name:                            types.StringValue("ds"),
		ComposedOf:                      types.ListNull(types.StringType),
		IgnoreMissingComponentTemplates: types.ListNull(types.StringType),
		IndexPatterns:                   patterns,
		Metadata:                        jsontypes.NewNormalizedNull(),
		Priority:                        types.Int64Null(),
		Version:                         types.Int64Null(),
		DataStream:                      dsObj,
		Template:                        types.ObjectNull(TemplateAttrTypes()),
	}
	out, diags := m.toAPIModel(ctx)
	if diags.HasError() {
		t.Fatal(diags)
	}
	if out.DataStream == nil || out.DataStream.Hidden == nil || !*out.DataStream.Hidden {
		t.Fatalf("hidden not set: %#v", out.DataStream)
	}
	if out.DataStream.AllowCustomRouting != nil {
		t.Fatalf("allow_custom_routing should be omitted when false, got %#v", out.DataStream.AllowCustomRouting)
	}
}

func TestModel_GetVersionRequirements_ignoreMissing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ignoreList, diags := types.ListValueFrom(ctx, types.StringType, []attr.Value{types.StringValue("ct1")})
	if diags.HasError() {
		t.Fatal(diags)
	}
	plan := Model{IgnoreMissingComponentTemplates: ignoreList}
	reqs, diags := plan.GetVersionRequirements(ctx)
	if diags.HasError() {
		t.Fatal(diags)
	}
	if len(reqs) != 1 {
		t.Fatalf("expected 1 requirement, got %d", len(reqs))
	}
	emptyPlan := Model{IgnoreMissingComponentTemplates: types.ListNull(types.StringType)}
	reqs, diags = emptyPlan.GetVersionRequirements(ctx)
	if diags.HasError() {
		t.Fatal(diags)
	}
	if len(reqs) != 0 {
		t.Fatalf("expected 0 requirements, got %d", len(reqs))
	}
}

func TestModel_GetVersionRequirements_dataStreamOptions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fsObj, diags := types.ObjectValue(datastreamoptions.FailureStoreAttrTypes(), map[string]attr.Value{
		"enabled":   types.BoolValue(true),
		"lifecycle": types.ObjectNull(datastreamoptions.FailureStoreLifecycleAttrTypes()),
	})
	if diags.HasError() {
		t.Fatal(diags)
	}
	dsoObj, diags := types.ObjectValue(datastreamoptions.AttrTypes(), map[string]attr.Value{
		"failure_store": fsObj,
	})
	if diags.HasError() {
		t.Fatal(diags)
	}
	tplObj, diags := types.ObjectValue(TemplateAttrTypes(), map[string]attr.Value{
		"alias":               types.SetNull(NewAliasObjectType()),
		"mappings":            esindex.NewMappingsNull(),
		"settings":            customtypes.NewIndexSettingsNull(),
		"lifecycle":           types.ObjectNull(LifecycleAttrTypes()),
		"data_stream_options": dsoObj,
	})
	if diags.HasError() {
		t.Fatal(diags)
	}
	model := Model{Template: tplObj}
	reqs, diags := model.GetVersionRequirements(ctx)
	if diags.HasError() {
		t.Fatal(diags)
	}
	if len(reqs) != 1 {
		t.Fatalf("expected 1 requirement, got %d", len(reqs))
	}
	noDsoTpl, diags := types.ObjectValue(TemplateAttrTypes(), map[string]attr.Value{
		"alias":               types.SetNull(NewAliasObjectType()),
		"mappings":            esindex.NewMappingsNull(),
		"settings":            customtypes.NewIndexSettingsNull(),
		"lifecycle":           types.ObjectNull(LifecycleAttrTypes()),
		"data_stream_options": types.ObjectNull(datastreamoptions.AttrTypes()),
	})
	if diags.HasError() {
		t.Fatal(diags)
	}
	modelNoDso := Model{Template: noDsoTpl}
	reqs, diags = modelNoDso.GetVersionRequirements(ctx)
	if diags.HasError() {
		t.Fatal(diags)
	}
	if len(reqs) != 0 {
		t.Fatalf("expected 0 requirements, got %d", len(reqs))
	}
}

func TestFlattenAliasElement_emptyFilterMapIsNull(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	av, diags := flattenAliasElement("a", models.IndexAlias{
		Filter:        map[string]any{},
		IndexRouting:  "ir",
		Routing:       "r",
		SearchRouting: "sr",
	})
	if diags.HasError() {
		t.Fatal(diags)
	}
	alias, ok := av.(AliasObjectValue)
	if !ok {
		t.Fatalf("got %T", av)
	}
	var am aliasutil.AliasModel
	diags = alias.As(ctx, &am, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		t.Fatal(diags)
	}
	if !am.Filter.IsNull() {
		t.Fatalf("expected null filter for empty API map, got %#v", am.Filter)
	}
}
