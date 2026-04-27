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

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"
)

func TestAliasSetsSemanticallyEqual_routingOnlyPlanVsIndexEchoState(t *testing.T) {
	ctx := context.Background()
	planEl, diags := NewAliasObjectValue(map[string]attr.Value{
		"name":           types.StringValue("routing_only_alias"),
		"routing":        types.StringValue("shard_1"),
		"index_routing":  types.StringValue(""),
		"search_routing": types.StringValue(""),
		"filter":         jsontypes.NewNormalizedNull(),
		"is_hidden":      types.BoolValue(false),
		"is_write_index": types.BoolValue(false),
	})
	if diags.HasError() {
		t.Fatal(diags)
	}
	// Flatten uses StringValue("") for empty API strings, routing omitted from GET.
	stateEl, diags := NewAliasObjectValue(map[string]attr.Value{
		"name":           types.StringValue("routing_only_alias"),
		"routing":        types.StringValue(""),
		"index_routing":  types.StringValue("shard_1"),
		"search_routing": types.StringValue(""),
		"filter":         jsontypes.NewNormalizedNull(),
		"is_hidden":      types.BoolValue(false),
		"is_write_index": types.BoolValue(false),
	})
	if diags.HasError() {
		t.Fatal(diags)
	}
	planSet, diags := types.SetValueFrom(ctx, NewAliasObjectType(), []attr.Value{planEl})
	if diags.HasError() {
		t.Fatal(diags)
	}
	stateSet, diags := types.SetValueFrom(ctx, NewAliasObjectType(), []attr.Value{stateEl})
	if diags.HasError() {
		t.Fatal(diags)
	}
	ok, diags := aliasPlanAndStateSetsSemanticallyEqual(ctx, planSet, stateSet)
	if diags.HasError() {
		t.Fatal(diags)
	}
	if !ok {
		t.Fatal("expected semantic equality for routing-only plan vs index_routing echo state")
	}
}

func TestAliasSetsSemanticallyEqual_routingOnlyPlanNullRoutingStrings_stateEmptyStrings(t *testing.T) {
	ctx := context.Background()
	planEl, diags := NewAliasObjectValue(map[string]attr.Value{
		"name":           types.StringValue("routing_only_alias"),
		"routing":        types.StringValue("shard_1"),
		"index_routing":  types.StringNull(),
		"search_routing": types.StringNull(),
		"filter":         jsontypes.NewNormalizedNull(),
		"is_hidden":      types.BoolValue(false),
		"is_write_index": types.BoolValue(false),
	})
	if diags.HasError() {
		t.Fatal(diags)
	}
	stateEl, diags := NewAliasObjectValue(map[string]attr.Value{
		"name":           types.StringValue("routing_only_alias"),
		"routing":        types.StringValue("shard_1"),
		"index_routing":  types.StringValue(""),
		"search_routing": types.StringValue(""),
		"filter":         jsontypes.NewNormalizedNull(),
		"is_hidden":      types.BoolValue(false),
		"is_write_index": types.BoolValue(false),
	})
	if diags.HasError() {
		t.Fatal(diags)
	}
	planSet, diags := types.SetValueFrom(ctx, NewAliasObjectType(), []attr.Value{planEl})
	if diags.HasError() {
		t.Fatal(diags)
	}
	stateSet, diags := types.SetValueFrom(ctx, NewAliasObjectType(), []attr.Value{stateEl})
	if diags.HasError() {
		t.Fatal(diags)
	}
	ok, diags := aliasPlanAndStateSetsSemanticallyEqual(ctx, planSet, stateSet)
	if diags.HasError() {
		t.Fatal(diags)
	}
	if !ok {
		t.Fatal("expected semantic equality when plan uses null for omitted routing fields and state uses empty strings")
	}
}

func TestAliasSetsSemanticallyEqual_planNullBools_stateFalseBools(t *testing.T) {
	ctx := context.Background()
	planEl, diags := NewAliasObjectValue(map[string]attr.Value{
		"name":           types.StringValue("routing_only_alias"),
		"routing":        types.StringValue("shard_1"),
		"index_routing":  types.StringNull(),
		"search_routing": types.StringNull(),
		"filter":         jsontypes.NewNormalizedNull(),
		"is_hidden":      types.BoolNull(),
		"is_write_index": types.BoolNull(),
	})
	if diags.HasError() {
		t.Fatal(diags)
	}
	stateEl, diags := NewAliasObjectValue(map[string]attr.Value{
		"name":           types.StringValue("routing_only_alias"),
		"routing":        types.StringValue("shard_1"),
		"index_routing":  types.StringValue(""),
		"search_routing": types.StringValue(""),
		"filter":         jsontypes.NewNormalizedNull(),
		"is_hidden":      types.BoolValue(false),
		"is_write_index": types.BoolValue(false),
	})
	if diags.HasError() {
		t.Fatal(diags)
	}
	planSet, diags := types.SetValueFrom(ctx, NewAliasObjectType(), []attr.Value{planEl})
	if diags.HasError() {
		t.Fatal(diags)
	}
	stateSet, diags := types.SetValueFrom(ctx, NewAliasObjectType(), []attr.Value{stateEl})
	if diags.HasError() {
		t.Fatal(diags)
	}
	ok, diags := aliasPlanAndStateSetsSemanticallyEqual(ctx, planSet, stateSet)
	if diags.HasError() {
		t.Fatal(diags)
	}
	if !ok {
		t.Fatal("expected semantic equality when plan omits bools (null) and state has explicit false")
	}
}

func TestReconcilePlanWithPriorStateForSemanticDrift_routingOnlyNullBoolPlan(t *testing.T) {
	ctx := context.Background()
	planEl, diags := NewAliasObjectValue(map[string]attr.Value{
		"name":           types.StringValue("routing_only_alias"),
		"routing":        types.StringValue("shard_1"),
		"index_routing":  types.StringNull(),
		"search_routing": types.StringNull(),
		"filter":         jsontypes.NewNormalizedNull(),
		"is_hidden":      types.BoolNull(),
		"is_write_index": types.BoolNull(),
	})
	require.False(t, diags.HasError(), "%v", diags)
	stateEl, diags := NewAliasObjectValue(map[string]attr.Value{
		"name":           types.StringValue("routing_only_alias"),
		"routing":        types.StringValue("shard_1"),
		"index_routing":  types.StringValue(""),
		"search_routing": types.StringValue(""),
		"filter":         jsontypes.NewNormalizedNull(),
		"is_hidden":      types.BoolValue(false),
		"is_write_index": types.BoolValue(false),
	})
	require.False(t, diags.HasError(), "%v", diags)
	planSet, diags := types.SetValueFrom(ctx, NewAliasObjectType(), []attr.Value{planEl})
	require.False(t, diags.HasError(), "%v", diags)
	stateSet, diags := types.SetValueFrom(ctx, NewAliasObjectType(), []attr.Value{stateEl})
	require.False(t, diags.HasError(), "%v", diags)
	emptyTpl := map[string]attr.Value{
		"alias":               planSet,
		"mappings":            jsontypes.NewNormalizedNull(),
		"settings":            customtypes.NewIndexSettingsNull(),
		"lifecycle":           types.ObjectNull(LifecycleAttrTypes()),
		"data_stream_options": types.ObjectNull(DataStreamOptionsAttrTypes()),
	}
	planTpl, diags := types.ObjectValue(TemplateAttrTypes(), emptyTpl)
	require.False(t, diags.HasError(), "%v", diags)
	stateTpl, diags := types.ObjectValue(TemplateAttrTypes(), map[string]attr.Value{
		"alias":               stateSet,
		"mappings":            jsontypes.NewNormalizedNull(),
		"settings":            customtypes.NewIndexSettingsNull(),
		"lifecycle":           types.ObjectNull(LifecycleAttrTypes()),
		"data_stream_options": types.ObjectNull(DataStreamOptionsAttrTypes()),
	})
	require.False(t, diags.HasError(), "%v", diags)
	var plan, state Model
	plan.Template = planTpl
	state.Template = stateTpl
	merged, diags := reconcilePlanWithPriorStateForSemanticDrift(ctx, plan, state)
	require.False(t, diags.HasError(), "%v", diags)
	require.NotNil(t, merged, "expected alias merge when plan null-bools match state false semantically")
	var mt TemplateBlockModel
	require.False(t, merged.Template.As(ctx, &mt, basetypes.ObjectAsOptions{}).HasError())
	require.True(t, mt.Alias.Equal(stateSet), "merged plan should adopt state alias encoding")
}
