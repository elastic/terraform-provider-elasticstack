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
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"
)

// ES omits top-level routing in GET when index_routing and search_routing differ (see manual curl against 8.x).
func TestEnrichTemplateAliasesRoutingFromReference_strictEqualToPlanAfterEnrich(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	planAlias, diags := NewAliasObjectValue(aliasAttrMap("detailed_alias",
		strAttr("index_explicit_v1"),
		strAttr("route_common_v1"),
		strAttr("search_explicit_v1"),
		jsontypes.NewNormalizedNull(), true, true))
	require.False(t, diags.HasError(), "%v", diags)

	ia := models.IndexAlias{
		IndexRouting:  "index_explicit_v1",
		SearchRouting: "search_explicit_v1",
		Routing:       "",
		IsHidden:      true,
		IsWriteIndex:  true,
	}
	readElem, d := flattenAliasElement("detailed_alias", ia)
	require.False(t, d.HasError(), "%v", d)

	planSet, d := types.SetValueFrom(ctx, NewAliasObjectType(), []attr.Value{planAlias})
	require.False(t, d.HasError(), "%v", d)
	readSet, d := types.SetValueFrom(ctx, NewAliasObjectType(), []attr.Value{readElem})
	require.False(t, d.HasError(), "%v", d)

	planTpl, d := types.ObjectValue(TemplateAttrTypes(), map[string]attr.Value{
		"alias":               planSet,
		"mappings":            jsontypes.NewNormalizedNull(),
		"settings":            customtypes.NewIndexSettingsNull(),
		"lifecycle":           types.ObjectNull(LifecycleAttrTypes()),
		"data_stream_options": types.ObjectNull(DataStreamOptionsAttrTypes()),
	})
	require.False(t, d.HasError(), "%v", d)
	readTpl, d := types.ObjectValue(TemplateAttrTypes(), map[string]attr.Value{
		"alias":               readSet,
		"mappings":            jsontypes.NewNormalizedNull(),
		"settings":            customtypes.NewIndexSettingsNull(),
		"lifecycle":           types.ObjectNull(LifecycleAttrTypes()),
		"data_stream_options": types.ObjectNull(DataStreamOptionsAttrTypes()),
	})
	require.False(t, d.HasError(), "%v", d)

	read := Model{Template: readTpl}
	ref := Model{Template: planTpl}
	d = enrichTemplateAliasesRoutingFromReference(ctx, &read, ref)
	require.False(t, d.HasError(), "%v", d)

	var tpl TemplateBlockModel
	d = read.Template.As(ctx, &tpl, basetypes.ObjectAsOptions{})
	require.False(t, d.HasError(), "%v", d)
	elems := tpl.Alias.Elements()
	require.Len(t, elems, 1)
	enriched, ok := elems[0].(AliasObjectValue)
	require.True(t, ok)
	require.True(t, enriched.Equal(planAlias), "enriched alias must strictly equal plan for post-apply consistency")
}
