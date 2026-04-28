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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/go-version"
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
	pr := 42
	ver := 7
	tpl := &models.IndexTemplate{
		ComposedOf:                      []string{"a"},
		IgnoreMissingComponentTemplates: []string{"missing"},
		IndexPatterns:                   []string{"ix-*"},
		Meta:                            map[string]any{"k": "v"},
		Priority:                        &pr,
		Version:                         &ver,
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

func TestValidateIgnoreMissingComponentTemplatesVersion(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	old := version.Must(version.NewVersion("8.0.0"))
	okVer := version.Must(version.NewVersion("8.8.0"))

	ignoreList, diags := types.ListValueFrom(ctx, types.StringType, []attr.Value{types.StringValue("ct1")})
	if diags.HasError() {
		t.Fatal(diags)
	}
	plan := Model{IgnoreMissingComponentTemplates: ignoreList}
	if diags := validateIgnoreMissingComponentTemplatesVersion(plan, old); !diags.HasError() {
		t.Fatal("expected error on old cluster")
	}
	if diags := validateIgnoreMissingComponentTemplatesVersion(plan, okVer); diags.HasError() {
		t.Fatal(diags)
	}
	emptyPlan := Model{IgnoreMissingComponentTemplates: types.ListNull(types.StringType)}
	if diags := validateIgnoreMissingComponentTemplatesVersion(emptyPlan, old); diags.HasError() {
		t.Fatal(diags)
	}
}

func TestValidateDataStreamOptionsVersion(t *testing.T) {
	t.Parallel()
	old := version.Must(version.NewVersion("8.17.0"))
	okVer := version.Must(version.NewVersion("9.2.0"))

	fsObj, diags := types.ObjectValue(FailureStoreAttrTypes(), map[string]attr.Value{
		"enabled":   types.BoolValue(true),
		"lifecycle": types.ObjectNull(FailureStoreLifecycleAttrTypes()),
	})
	if diags.HasError() {
		t.Fatal(diags)
	}
	dsoObj, diags := types.ObjectValue(DataStreamOptionsAttrTypes(), map[string]attr.Value{
		"failure_store": fsObj,
	})
	if diags.HasError() {
		t.Fatal(diags)
	}
	tplObj, diags := types.ObjectValue(TemplateAttrTypes(), map[string]attr.Value{
		"alias":               types.SetNull(NewAliasObjectType()),
		"mappings":            jsontypes.NewNormalizedNull(),
		"settings":            customtypes.NewIndexSettingsNull(),
		"lifecycle":           types.ObjectNull(LifecycleAttrTypes()),
		"data_stream_options": dsoObj,
	})
	if diags.HasError() {
		t.Fatal(diags)
	}
	plan := Model{Template: tplObj}
	if diags := validateDataStreamOptionsVersion(plan, old); !diags.HasError() {
		t.Fatal("expected error on old cluster")
	}
	if diags := validateDataStreamOptionsVersion(plan, okVer); diags.HasError() {
		t.Fatal(diags)
	}
	noDsoTpl, diags := types.ObjectValue(TemplateAttrTypes(), map[string]attr.Value{
		"alias":               types.SetNull(NewAliasObjectType()),
		"mappings":            jsontypes.NewNormalizedNull(),
		"settings":            customtypes.NewIndexSettingsNull(),
		"lifecycle":           types.ObjectNull(LifecycleAttrTypes()),
		"data_stream_options": types.ObjectNull(DataStreamOptionsAttrTypes()),
	})
	if diags.HasError() {
		t.Fatal(diags)
	}
	planNo := Model{Template: noDsoTpl}
	if diags := validateDataStreamOptionsVersion(planNo, old); diags.HasError() {
		t.Fatal(diags)
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
	var am AliasElementModel
	diags = alias.As(ctx, &am, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		t.Fatal(diags)
	}
	if !am.Filter.IsNull() {
		t.Fatalf("expected null filter for empty API map, got %#v", am.Filter)
	}
}
