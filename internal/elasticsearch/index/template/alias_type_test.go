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

func TestAliasObjectValue_ObjectSemanticEquals_routingOnlyEcho(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	prior := mustAlias(t, aliasAttrMap("a", strNull(), strAttr("x"), strNull(), jsontypes.NewNormalizedNull(), false, false))
	incoming := mustAlias(t, aliasAttrMap("a", strAttr("x"), strAttr("x"), strAttr("x"), jsontypes.NewNormalizedNull(), false, false))
	eq, diags := prior.ObjectSemanticEquals(ctx, incoming)
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

func TestAliasObjectValue_ObjectSemanticEquals_wrongType(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	prior := mustAlias(t, aliasAttrMap("a", strNull(), strAttr("x"), strNull(), jsontypes.NewNormalizedNull(), false, false))
	_, diags := prior.ObjectSemanticEquals(ctx, types.ObjectNull(AliasAttributeTypes()))
	require.True(t, diags.HasError())
}
