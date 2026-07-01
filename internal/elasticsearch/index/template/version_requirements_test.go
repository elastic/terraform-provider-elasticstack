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

	esindex "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/aliasutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/datastreamoptions"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestModel_GetVersionRequirements(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("neither configured", func(t *testing.T) {
		t.Parallel()
		m := Model{
			Template:                        types.ObjectNull(TemplateAttrTypes()),
			IgnoreMissingComponentTemplates: types.ListNull(types.StringType),
		}
		reqs, diags := m.GetVersionRequirements(ctx)
		if diags.HasError() {
			t.Fatal(diags)
		}
		if len(reqs) != 0 {
			t.Fatalf("expected nil or empty, got %v", reqs)
		}
	})

	t.Run("only data_stream_options", func(t *testing.T) {
		t.Parallel()
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
			"alias":               types.SetNull(aliasutil.NewAliasObjectType()),
			"mappings":            esindex.NewMappingsNull(),
			"settings":            customtypes.NewIndexSettingsNull(),
			"lifecycle":           types.ObjectNull(LifecycleAttrTypes()),
			"data_stream_options": dsoObj,
		})
		if diags.HasError() {
			t.Fatal(diags)
		}
		m := Model{Template: tplObj}
		reqs, diags := m.GetVersionRequirements(ctx)
		if diags.HasError() {
			t.Fatal(diags)
		}
		if len(reqs) != 1 {
			t.Fatalf("expected 1 requirement, got %d", len(reqs))
		}
		if !reqs[0].MinVersion.Equal(datastreamoptions.MinSupportedVersion) {
			t.Fatalf("expected min version %s, got %s", datastreamoptions.MinSupportedVersion.String(), reqs[0].MinVersion.String())
		}
	})

	t.Run("only ignore_missing_component_templates", func(t *testing.T) {
		t.Parallel()
		ignoreList, diags := types.ListValueFrom(ctx, types.StringType, []attr.Value{types.StringValue("ct1")})
		if diags.HasError() {
			t.Fatal(diags)
		}
		m := Model{IgnoreMissingComponentTemplates: ignoreList}
		reqs, diags := m.GetVersionRequirements(ctx)
		if diags.HasError() {
			t.Fatal(diags)
		}
		if len(reqs) != 1 {
			t.Fatalf("expected 1 requirement, got %d", len(reqs))
		}
		if !reqs[0].MinVersion.Equal(esindex.MinSupportedIgnoreMissingComponentTemplateVersion) {
			t.Fatalf("expected min version %s, got %s", esindex.MinSupportedIgnoreMissingComponentTemplateVersion.String(), reqs[0].MinVersion.String())
		}
	})

	t.Run("both configured", func(t *testing.T) {
		t.Parallel()
		ignoreList, diags := types.ListValueFrom(ctx, types.StringType, []attr.Value{types.StringValue("ct1")})
		if diags.HasError() {
			t.Fatal(diags)
		}
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
			"alias":               types.SetNull(aliasutil.NewAliasObjectType()),
			"mappings":            esindex.NewMappingsNull(),
			"settings":            customtypes.NewIndexSettingsNull(),
			"lifecycle":           types.ObjectNull(LifecycleAttrTypes()),
			"data_stream_options": dsoObj,
		})
		if diags.HasError() {
			t.Fatal(diags)
		}
		m := Model{
			Template:                        tplObj,
			IgnoreMissingComponentTemplates: ignoreList,
		}
		reqs, diags := m.GetVersionRequirements(ctx)
		if diags.HasError() {
			t.Fatal(diags)
		}
		if len(reqs) != 2 {
			t.Fatalf("expected 2 requirements, got %d", len(reqs))
		}
	})

	t.Run("empty ignore_missing_component_templates", func(t *testing.T) {
		t.Parallel()
		ignoreList, diags := types.ListValueFrom(ctx, types.StringType, []attr.Value{})
		if diags.HasError() {
			t.Fatal(diags)
		}
		m := Model{IgnoreMissingComponentTemplates: ignoreList}
		reqs, diags := m.GetVersionRequirements(ctx)
		if diags.HasError() {
			t.Fatal(diags)
		}
		if len(reqs) != 0 {
			t.Fatalf("expected nil or empty, got %v", reqs)
		}
	})
}
