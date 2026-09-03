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
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/templateutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"
)

func TestReconcilePlanWithPriorStateForSemanticDrift_settingsNestedPlanDottedState(t *testing.T) {
	ctx := context.Background()
	planSettings := customtypes.NewIndexSettingsValue(`{"index":{"number_of_shards":1}}`)
	stateSettings := customtypes.NewIndexSettingsValue(`{"index.number_of_shards":"1"}`)

	emptyAlias, diags := types.SetValueFrom(ctx, aliasutil.NewAliasObjectType(), []attr.Value{})
	require.False(t, diags.HasError(), "%v", diags)

	planTpl, diags := types.ObjectValue(TemplateAttrTypes(), map[string]attr.Value{
		"alias":               emptyAlias,
		"mappings":            esindex.NewMappingsNull(),
		"settings":            planSettings,
		"lifecycle":           types.ObjectNull(LifecycleAttrTypes()),
		"data_stream_options": types.ObjectNull(datastreamoptions.AttrTypes()),
	})
	require.False(t, diags.HasError(), "%v", diags)
	stateTpl, diags := types.ObjectValue(TemplateAttrTypes(), map[string]attr.Value{
		"alias":               emptyAlias,
		"mappings":            esindex.NewMappingsNull(),
		"settings":            stateSettings,
		"lifecycle":           types.ObjectNull(LifecycleAttrTypes()),
		"data_stream_options": types.ObjectNull(datastreamoptions.AttrTypes()),
	})
	require.False(t, diags.HasError(), "%v", diags)

	var plan, state, config Model
	plan.Template = planTpl
	state.Template = stateTpl
	config.Template = planTpl
	merged, diags := templateutil.ReconcilePlanModelForSemanticDrift(ctx, plan, state, config, TemplateAttrTypes)
	require.False(t, diags.HasError(), "%v", diags)
	require.NotNil(t, merged)
	var mt TemplateBlockModel
	require.False(t, merged.Template.As(ctx, &mt, basetypes.ObjectAsOptions{}).HasError())
	require.True(t, mt.Settings.Equal(stateSettings), "plan should adopt state settings encoding when semantically equal")
}

func TestApplyTemplateAliasReconciliationFromReference_routingOnlyVsSplitEcho(t *testing.T) {
	ctx := context.Background()
	filter := jsontypes.NewNormalizedNull()
	refAlias, diags := aliasutil.NewAliasObjectValue(testAliasAttrMap("routing_only_alias", testStrNull(), testStrAttr("shard_1"), testStrNull(), filter, false, false))
	require.False(t, diags.HasError(), "%v", diags)
	apiAlias, diags := aliasutil.NewAliasObjectValue(testAliasAttrMap("routing_only_alias", testStrAttr("shard_1"), testStrAttr(""), testStrAttr("shard_1"), filter, false, false))
	require.False(t, diags.HasError(), "%v", diags)

	refSet, diags := types.SetValue(aliasutil.NewAliasObjectType(), []attr.Value{refAlias})
	require.False(t, diags.HasError(), "%v", diags)
	apiSet, diags := types.SetValue(aliasutil.NewAliasObjectType(), []attr.Value{apiAlias})
	require.False(t, diags.HasError(), "%v", diags)

	refTpl, diags := types.ObjectValue(TemplateAttrTypes(), map[string]attr.Value{
		"alias":               refSet,
		"mappings":            esindex.NewMappingsNull(),
		"settings":            customtypes.NewIndexSettingsValue(`{}`),
		"lifecycle":           types.ObjectNull(LifecycleAttrTypes()),
		"data_stream_options": types.ObjectNull(datastreamoptions.AttrTypes()),
	})
	require.False(t, diags.HasError(), "%v", diags)
	apiTpl, diags := types.ObjectValue(TemplateAttrTypes(), map[string]attr.Value{
		"alias":               apiSet,
		"mappings":            esindex.NewMappingsNull(),
		"settings":            customtypes.NewIndexSettingsValue(`{}`),
		"lifecycle":           types.ObjectNull(LifecycleAttrTypes()),
		"data_stream_options": types.ObjectNull(datastreamoptions.AttrTypes()),
	})
	require.False(t, diags.HasError(), "%v", diags)

	var out, ref Model
	out.Template = apiTpl
	ref.Template = refTpl
	diags = aliasutil.ApplyTemplateAliasReconciliationFromReference(ctx, &out.Template, ref.Template, TemplateAttrTypes())
	require.False(t, diags.HasError(), "%v", diags)
	var mt TemplateBlockModel
	require.False(t, out.Template.As(ctx, &mt, basetypes.ObjectAsOptions{}).HasError())
	require.True(t, mt.Alias.Equal(refSet), "API echo should adopt reference (config) alias encoding when semantically equal")
}

func testStrAttr(s string) types.String {
	return types.StringValue(s)
}

func testStrNull() types.String {
	return types.StringNull()
}

func testAliasAttrMap(
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
