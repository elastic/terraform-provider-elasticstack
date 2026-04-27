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

func TestEnrichTemplateAliasesRoutingFromReference_splitRouting(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	planAlias, diags := NewAliasObjectValue(aliasAttrMap("detailed_alias",
		strAttr("index_explicit_v1"),
		strAttr("route_common_v1"),
		strAttr("search_explicit_v1"),
		jsontypes.NewNormalizedNull(), true, true))
	require.False(t, diags.HasError(), "%v", diags)

	apiAlias, diags := NewAliasObjectValue(aliasAttrMap("detailed_alias",
		strAttr("index_explicit_v1"),
		strAttr(""),
		strAttr("search_explicit_v1"),
		jsontypes.NewNormalizedNull(), true, true))
	require.False(t, diags.HasError(), "%v", diags)

	planSet, diags := types.SetValueFrom(ctx, NewAliasObjectType(), []attr.Value{planAlias})
	require.False(t, diags.HasError(), "%v", diags)
	apiSet, diags := types.SetValueFrom(ctx, NewAliasObjectType(), []attr.Value{apiAlias})
	require.False(t, diags.HasError(), "%v", diags)

	planTpl, diags := types.ObjectValue(TemplateAttrTypes(), map[string]attr.Value{
		"alias":               planSet,
		"mappings":            jsontypes.NewNormalizedNull(),
		"settings":            customtypes.NewIndexSettingsNull(),
		"lifecycle":           types.ObjectNull(LifecycleAttrTypes()),
		"data_stream_options": types.ObjectNull(DataStreamOptionsAttrTypes()),
	})
	require.False(t, diags.HasError(), "%v", diags)
	apiTpl, diags := types.ObjectValue(TemplateAttrTypes(), map[string]attr.Value{
		"alias":               apiSet,
		"mappings":            jsontypes.NewNormalizedNull(),
		"settings":            customtypes.NewIndexSettingsNull(),
		"lifecycle":           types.ObjectNull(LifecycleAttrTypes()),
		"data_stream_options": types.ObjectNull(DataStreamOptionsAttrTypes()),
	})
	require.False(t, diags.HasError(), "%v", diags)

	read := Model{Template: apiTpl}
	ref := Model{Template: planTpl}

	diags = enrichTemplateAliasesRoutingFromReference(ctx, &read, ref)
	require.False(t, diags.HasError(), "%v", diags)

	var tpl TemplateBlockModel
	diags = read.Template.As(ctx, &tpl, basetypes.ObjectAsOptions{})
	require.False(t, diags.HasError(), "%v", diags)
	elems := tpl.Alias.Elements()
	require.Len(t, elems, 1)
	av, ok, dCoerce := coerceAliasObjectValue(ctx, elems[0])
	require.False(t, dCoerce.HasError(), "%v", dCoerce)
	require.True(t, ok)
	var am AliasElementModel
	diags = av.As(ctx, &am, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})
	require.False(t, diags.HasError(), "%v", diags)
	require.Equal(t, "route_common_v1", am.Routing.ValueString())
}
