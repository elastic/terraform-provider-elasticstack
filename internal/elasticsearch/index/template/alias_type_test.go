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

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func strAttr(s string) types.String {
	return types.StringValue(s)
}

func strNull() types.String {
	return types.StringNull()
}

func aliasAttrMap(
	name string,
	indexRouting, routing, searchRouting types.String,
	filter jsontypes.Normalized,
	isHidden, isWriteIndex bool,
) map[string]attr.Value {
	return map[string]attr.Value{
		"name":           types.StringValue(name),
		"index_routing":  indexRouting,
		"routing":        routing,
		"search_routing": searchRouting,
		"filter":         filter,
		"is_hidden":      types.BoolValue(isHidden),
		"is_write_index": types.BoolValue(isWriteIndex),
	}
}

func mustAlias(t *testing.T, attrs map[string]attr.Value) AliasObjectValue {
	t.Helper()
	v, diags := NewAliasObjectValue(attrs)
	require.False(t, diags.HasError(), "%v", diags)
	return v
}

func TestAliasObjectValue_ObjectSemanticEquals_identical(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	attrs := aliasAttrMap("a", strAttr("x"), strAttr("x"), strAttr("x"), jsontypes.NewNormalizedNull(), false, false)
	prior := mustAlias(t, attrs)
	incoming := mustAlias(t, attrs)
	eq, diags := prior.ObjectSemanticEquals(ctx, incoming)
	require.False(t, diags.HasError(), "%v", diags)
	require.True(t, eq)
}

func TestAliasObjectValue_ObjectSemanticEquals_identical_hiddenAndWriteIndexTrue(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	attrs := aliasAttrMap("a", strAttr("x"), strAttr("x"), strAttr("x"), jsontypes.NewNormalizedNull(), true, true)
	prior := mustAlias(t, attrs)
	incoming := mustAlias(t, attrs)
	eq, diags := prior.ObjectSemanticEquals(ctx, incoming)
	require.False(t, diags.HasError(), "%v", diags)
	require.True(t, eq)
}

func TestAliasObjectValue_ObjectSemanticEquals_differingName(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	prior := mustAlias(t, aliasAttrMap("a", strNull(), strAttr("x"), strNull(), jsontypes.NewNormalizedNull(), false, false))
	incoming := mustAlias(t, aliasAttrMap("b", strNull(), strAttr("x"), strNull(), jsontypes.NewNormalizedNull(), false, false))
	eq, diags := prior.ObjectSemanticEquals(ctx, incoming)
	require.False(t, diags.HasError(), "%v", diags)
	require.False(t, eq)
}

func TestAliasObjectValue_ObjectSemanticEquals_differingRouting(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	prior := mustAlias(t, aliasAttrMap("a", strNull(), strAttr("x"), strNull(), jsontypes.NewNormalizedNull(), false, false))
	incoming := mustAlias(t, aliasAttrMap("a", strNull(), strAttr("y"), strNull(), jsontypes.NewNormalizedNull(), false, false))
	eq, diags := prior.ObjectSemanticEquals(ctx, incoming)
	require.False(t, diags.HasError(), "%v", diags)
	require.False(t, eq)
}

func TestAliasObjectValue_ObjectSemanticEquals_differingIsHidden(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	prior := mustAlias(t, aliasAttrMap("a", strNull(), strAttr("x"), strNull(), jsontypes.NewNormalizedNull(), false, false))
	incoming := mustAlias(t, aliasAttrMap("a", strNull(), strAttr("x"), strNull(), jsontypes.NewNormalizedNull(), true, false))
	eq, diags := prior.ObjectSemanticEquals(ctx, incoming)
	require.False(t, diags.HasError(), "%v", diags)
	require.False(t, eq)
}

func TestAliasObjectValue_ObjectSemanticEquals_differingIsWriteIndex(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	prior := mustAlias(t, aliasAttrMap("a", strNull(), strAttr("x"), strNull(), jsontypes.NewNormalizedNull(), false, false))
	incoming := mustAlias(t, aliasAttrMap("a", strNull(), strAttr("x"), strNull(), jsontypes.NewNormalizedNull(), false, true))
	eq, diags := prior.ObjectSemanticEquals(ctx, incoming)
	require.False(t, diags.HasError(), "%v", diags)
	require.False(t, eq)
}

func TestAliasObjectValue_ObjectSemanticEquals_routingOnlyEcho(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	prior := mustAlias(t, aliasAttrMap("a", strNull(), strAttr("x"), strNull(), jsontypes.NewNormalizedNull(), false, false))
	incoming := mustAlias(t, aliasAttrMap("a", strAttr("x"), strAttr("x"), strAttr("x"), jsontypes.NewNormalizedNull(), false, false))
	eq, diags := prior.ObjectSemanticEquals(ctx, incoming)
	require.False(t, diags.HasError(), "%v", diags)
	require.True(t, eq)
}

// Routing-only plan may represent omitted index_routing/search_routing as null; flatten uses "".
func TestAliasObjectValue_ObjectSemanticEquals_routingOnly_planNullRoutingFields_stateEmptyStrings(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	plan := mustAlias(t, aliasAttrMap("routing_only_alias", strNull(), strAttr("shard_1"), strNull(), jsontypes.NewNormalizedNull(), false, false))
	state := mustAlias(t, aliasAttrMap("routing_only_alias", strAttr(""), strAttr("shard_1"), strAttr(""), jsontypes.NewNormalizedNull(), false, false))
	eq, diags := plan.ObjectSemanticEquals(ctx, state)
	require.False(t, diags.HasError(), "%v", diags)
	require.True(t, eq)
}

func TestAliasObjectValue_ObjectSemanticEquals_apiEchoRead_vs_routingOnlyPlan(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	// GET: generic routing omitted, index_routing echoes shard_1, search_routing empty.
	read := mustAlias(t, aliasAttrMap("routing_only_alias", strAttr("shard_1"), strAttr(""), strAttr(""), jsontypes.NewNormalizedNull(), false, false))
	plan := mustAlias(t, aliasAttrMap("routing_only_alias", strNull(), strAttr("shard_1"), strNull(), jsontypes.NewNormalizedNull(), false, false))
	eq, diags := read.ObjectSemanticEquals(ctx, plan)
	require.False(t, diags.HasError(), "%v", diags)
	require.True(t, eq)
}

func TestAliasObjectValue_ObjectSemanticEquals_explicitIndexRouting(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	prior := mustAlias(t, aliasAttrMap("a", strAttr("y"), strAttr("x"), strNull(), jsontypes.NewNormalizedNull(), false, false))
	incoming := mustAlias(t, aliasAttrMap("a", strAttr("y"), strAttr("x"), strAttr("x"), jsontypes.NewNormalizedNull(), false, false))
	eq, diags := prior.ObjectSemanticEquals(ctx, incoming)
	require.False(t, diags.HasError(), "%v", diags)
	require.True(t, eq)
}

func TestAliasObjectValue_ObjectSemanticEquals_esDropsTopLevelRoutingWithSplitRoutings(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	plan := mustAlias(t, aliasAttrMap("detailed_alias",
		strAttr("index_explicit_v1"),
		strAttr("route_common_v1"),
		strAttr("search_explicit_v1"),
		jsontypes.NewNormalizedNull(), true, true))
	api := mustAlias(t, aliasAttrMap("detailed_alias",
		strAttr("index_explicit_v1"),
		strAttr(""),
		strAttr("search_explicit_v1"),
		jsontypes.NewNormalizedNull(), true, true))
	eq, diags := plan.ObjectSemanticEquals(ctx, api)
	require.False(t, diags.HasError(), "%v", diags)
	require.True(t, eq)
}

func TestAliasObjectValue_ObjectSemanticEquals_esEchoesMainRoutingIntoIndexRouting(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	plan := mustAlias(t, aliasAttrMap("detailed_alias",
		strAttr("index_explicit_v1"),
		strAttr("route_common_v1"),
		strAttr("search_explicit_v1"),
		jsontypes.NewNormalizedNull(), true, true))
	api := mustAlias(t, aliasAttrMap("detailed_alias",
		strAttr("route_common_v1"),
		strAttr(""),
		strAttr("search_explicit_v1"),
		jsontypes.NewNormalizedNull(), true, true))
	eq, diags := api.ObjectSemanticEquals(ctx, plan)
	require.False(t, diags.HasError(), "%v", diags)
	require.True(t, eq)
}

func TestAliasObjectValue_ObjectSemanticEquals_frameworkCallOrder_refreshVsPlan(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	plan := mustAlias(t, aliasAttrMap("detailed_alias",
		strAttr("index_explicit_v1"),
		strAttr("route_common_v1"),
		strAttr("search_explicit_v1"),
		jsontypes.NewNormalizedNull(), true, true))
	api := mustAlias(t, aliasAttrMap("detailed_alias",
		strAttr("index_explicit_v1"),
		strAttr(""),
		strAttr("search_explicit_v1"),
		jsontypes.NewNormalizedNull(), true, true))
	eq, diags := api.ObjectSemanticEquals(ctx, plan)
	require.False(t, diags.HasError(), "%v", diags)
	require.True(t, eq)
}

func TestAliasObjectValue_ObjectSemanticEquals_filterWhitespace(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	prior := mustAlias(t, aliasAttrMap("a", strNull(), strAttr("x"), strNull(), jsontypes.NewNormalizedValue(`{"match": {"a": 1}}`), false, false))
	incoming := mustAlias(t, aliasAttrMap("a", strNull(), strAttr("x"), strNull(), jsontypes.NewNormalizedValue(`{"match":{"a":1}}`), false, false))
	eq, diags := prior.ObjectSemanticEquals(ctx, incoming)
	require.False(t, diags.HasError(), "%v", diags)
	require.True(t, eq)
}

func TestAliasObjectValue_ObjectSemanticEquals_priorEmptyStringVsNullDerived(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	priorNull := mustAlias(t, aliasAttrMap("a", strNull(), strAttr("x"), strNull(), jsontypes.NewNormalizedNull(), false, false))
	priorEmpty := mustAlias(t, aliasAttrMap("a", strAttr(""), strAttr("x"), strAttr(""), jsontypes.NewNormalizedNull(), false, false))
	incoming := mustAlias(t, aliasAttrMap("a", strAttr("x"), strAttr("x"), strAttr("x"), jsontypes.NewNormalizedNull(), false, false))

	eq1, d1 := priorNull.ObjectSemanticEquals(ctx, incoming)
	require.False(t, d1.HasError(), "%v", d1)
	require.True(t, eq1)

	eq2, d2 := priorEmpty.ObjectSemanticEquals(ctx, incoming)
	require.False(t, d2.HasError(), "%v", d2)
	require.True(t, eq2)
}

func TestAliasObjectValue_ObjectSemanticEquals_asymmetricExplicitPriorVsDerivedNew(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	// Prior keeps an explicit index_routing; refreshed value echoes only generic routing.
	prior := mustAlias(t, aliasAttrMap("a", strAttr("y"), strAttr("x"), strNull(), jsontypes.NewNormalizedNull(), false, false))
	incoming := mustAlias(t, aliasAttrMap("a", strAttr("x"), strAttr("x"), strAttr("x"), jsontypes.NewNormalizedNull(), false, false))
	eq, diags := prior.ObjectSemanticEquals(ctx, incoming)
	require.False(t, diags.HasError(), "%v", diags)
	require.False(t, eq)
}

func TestAliasObjectValue_ObjectSemanticEquals_asymmetricDerivedPriorVsExplicitNew(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	// Prior had unset/empty index_routing matching routing echo; new side shows an explicit different index_routing.
	prior := mustAlias(t, aliasAttrMap("a", strNull(), strAttr("x"), strNull(), jsontypes.NewNormalizedNull(), false, false))
	incoming := mustAlias(t, aliasAttrMap("a", strAttr("y"), strAttr("x"), strAttr("x"), jsontypes.NewNormalizedNull(), false, false))
	eq, diags := prior.ObjectSemanticEquals(ctx, incoming)
	require.False(t, diags.HasError(), "%v", diags)
	require.False(t, eq)
}

func TestAliasObjectValue_ObjectSemanticEquals_nullAndUnknownObject(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	eq1, d1 := NewAliasObjectNull().ObjectSemanticEquals(ctx, NewAliasObjectNull())
	require.False(t, d1.HasError())
	require.True(t, eq1)

	eq2, d2 := NewAliasObjectUnknown().ObjectSemanticEquals(ctx, NewAliasObjectUnknown())
	require.False(t, d2.HasError())
	require.True(t, eq2)

	eq3, d3 := NewAliasObjectNull().ObjectSemanticEquals(ctx, NewAliasObjectUnknown())
	require.False(t, d3.HasError())
	require.False(t, eq3)
}

func TestAliasObjectValue_ObjectSemanticEquals_unknownNestedAttribute(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	attrs := aliasAttrMap("a", strNull(), strAttr("x"), strNull(), jsontypes.NewNormalizedNull(), false, false)
	attrs["routing"] = types.StringUnknown()
	prior := mustAlias(t, attrs)
	incoming := mustAlias(t, aliasAttrMap("a", strNull(), strAttr("x"), strNull(), jsontypes.NewNormalizedNull(), false, false))
	eq, diags := prior.ObjectSemanticEquals(ctx, incoming)
	require.False(t, diags.HasError(), "%v", diags)
	// Unknown on plan/config is filled from the other operand (refreshed state) before semantic rules.
	require.True(t, eq)
}

// Framework invokes proposedNew.ObjectSemanticEquals(ctx, priorState) (see terraform-plugin-framework fwschemadata).
func TestAliasObjectValue_ObjectSemanticEquals_frameworkOrder_routingOnlyCombinedInPlanSplitInState(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	planNullFields := mustAlias(t, aliasAttrMap("routing_only_alias", strNull(), strAttr("shard_1"), strNull(), jsontypes.NewNormalizedNull(), false, false))
	planEmptyFields := mustAlias(t, aliasAttrMap("routing_only_alias", strAttr(""), strAttr("shard_1"), strAttr(""), jsontypes.NewNormalizedNull(), false, false))
	state := mustAlias(t, aliasAttrMap("routing_only_alias", strAttr("shard_1"), strAttr(""), strAttr("shard_1"), jsontypes.NewNormalizedNull(), false, false))

	eq1, d1 := planNullFields.ObjectSemanticEquals(ctx, state)
	require.False(t, d1.HasError(), "%v", d1)
	require.True(t, eq1)

	eq2, d2 := planEmptyFields.ObjectSemanticEquals(ctx, state)
	require.False(t, d2.HasError(), "%v", d2)
	require.True(t, eq2)
}

func TestAliasObjectValue_ObjectSemanticEquals_explicitIndexSearch_planEmptyRouting_stateEchoesRouting(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	filter := jsontypes.NewNormalizedValue(`{"term":{"status":"active"}}`)
	plan := mustAlias(t, aliasAttrMap("my_alias", strAttr("shard_1"), strAttr(""), strAttr("shard_1"), filter, false, true))
	state := mustAlias(t, aliasAttrMap("my_alias", strAttr("shard_1"), strAttr("shard_1"), strAttr("shard_1"), filter, false, true))
	eq, diags := plan.ObjectSemanticEquals(ctx, state)
	require.False(t, diags.HasError(), "%v", diags)
	require.True(t, eq)
}

// Some Elasticsearch versions echo only index_routing for routing-only template aliases (search_routing empty).
func TestAliasObjectValue_ObjectSemanticEquals_routingOnly_plan_vs_stateIndexOnlyEcho(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	plan := mustAlias(t, aliasAttrMap("routing_only_alias", strNull(), strAttr("shard_1"), strNull(), jsontypes.NewNormalizedNull(), false, false))
	state := mustAlias(t, aliasAttrMap("routing_only_alias", strAttr("shard_1"), strAttr(""), strAttr(""), jsontypes.NewNormalizedNull(), false, false))
	eq, diags := plan.ObjectSemanticEquals(ctx, state)
	require.False(t, diags.HasError(), "%v", diags)
	require.True(t, eq)
}
