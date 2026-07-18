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

package componenttemplate

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

	planTpl, diags := types.ObjectValue(templateAttrTypes(), map[string]attr.Value{
		attrAlias:             emptyAlias,
		attrMappings:          esindex.NewMappingsNull(),
		attrSettings:          planSettings,
		attrDataStreamOptions: types.ObjectNull(datastreamoptions.AttrTypes()),
	})
	require.False(t, diags.HasError(), "%v", diags)
	stateTpl, diags := types.ObjectValue(templateAttrTypes(), map[string]attr.Value{
		attrAlias:             emptyAlias,
		attrMappings:          esindex.NewMappingsNull(),
		attrSettings:          stateSettings,
		attrDataStreamOptions: types.ObjectNull(datastreamoptions.AttrTypes()),
	})
	require.False(t, diags.HasError(), "%v", diags)

	var plan, state, config Data
	plan.Template = planTpl
	state.Template = stateTpl
	config.Template = planTpl
	merged, diags := templateutil.ReconcilePlanModelForSemanticDrift(ctx, plan, state, config, templateAttrTypes)
	require.False(t, diags.HasError(), "%v", diags)
	require.NotNil(t, merged)
	var mt TemplateModel
	require.False(t, merged.Template.As(ctx, &mt, basetypes.ObjectAsOptions{}).HasError())
	require.True(t, mt.Settings.Equal(stateSettings), "plan should adopt state settings encoding when semantically equal")
}

func TestReconcilePlanWithPriorStateForSemanticDrift_aliasRoutingOnlyPlanVsSplitState(t *testing.T) {
	ctx := context.Background()
	filter := jsontypes.NewNormalizedNull()

	planAlias, diags := aliasutil.NewAliasObjectValue(testAliasAttrMap(
		"routing_only_alias", testStrNull(), testStrAttr("shard_1"), testStrNull(), filter, false, false,
	))
	require.False(t, diags.HasError(), "%v", diags)
	stateAlias, diags := aliasutil.NewAliasObjectValue(testAliasAttrMap(
		"routing_only_alias", testStrAttr("shard_1"), testStrAttr(""), testStrAttr(""), filter, false, false,
	))
	require.False(t, diags.HasError(), "%v", diags)

	planSet, diags := types.SetValue(aliasutil.NewAliasObjectType(), []attr.Value{planAlias})
	require.False(t, diags.HasError(), "%v", diags)
	stateSet, diags := types.SetValue(aliasutil.NewAliasObjectType(), []attr.Value{stateAlias})
	require.False(t, diags.HasError(), "%v", diags)

	settings := customtypes.NewIndexSettingsValue(`{}`)
	planTpl, diags := types.ObjectValue(templateAttrTypes(), map[string]attr.Value{
		attrAlias:             planSet,
		attrMappings:          esindex.NewMappingsNull(),
		attrSettings:          settings,
		attrDataStreamOptions: types.ObjectNull(datastreamoptions.AttrTypes()),
	})
	require.False(t, diags.HasError(), "%v", diags)
	stateTpl, diags := types.ObjectValue(templateAttrTypes(), map[string]attr.Value{
		attrAlias:             stateSet,
		attrMappings:          esindex.NewMappingsNull(),
		attrSettings:          settings,
		attrDataStreamOptions: types.ObjectNull(datastreamoptions.AttrTypes()),
	})
	require.False(t, diags.HasError(), "%v", diags)

	var plan, state, config Data
	plan.Template = planTpl
	state.Template = stateTpl
	config.Template = planTpl
	merged, diags := templateutil.ReconcilePlanModelForSemanticDrift(ctx, plan, state, config, templateAttrTypes)
	require.False(t, diags.HasError(), "%v", diags)
	require.NotNil(t, merged)
	var mt TemplateModel
	require.False(t, merged.Template.As(ctx, &mt, basetypes.ObjectAsOptions{}).HasError())
	require.True(t, mt.Alias.Equal(stateSet), "plan should adopt state alias encoding when semantically equal")
}

func TestReconcilePlanWithPriorStateForSemanticDrift_nullOrUnknownTemplateNoOp(t *testing.T) {
	ctx := context.Background()

	t.Run("null template", func(t *testing.T) {
		var plan, state, config Data
		plan.Template = types.ObjectNull(templateAttrTypes())
		state.Template = types.ObjectNull(templateAttrTypes())
		merged, diags := templateutil.ReconcilePlanModelForSemanticDrift(ctx, plan, state, config, templateAttrTypes)
		require.False(t, diags.HasError(), "%v", diags)
		require.Nil(t, merged)
	})

	t.Run("unknown plan template", func(t *testing.T) {
		settings := customtypes.NewIndexSettingsValue(`{}`)
		stateTpl, diags := types.ObjectValue(templateAttrTypes(), map[string]attr.Value{
			attrAlias:             types.SetNull(aliasutil.NewAliasObjectType()),
			attrMappings:          esindex.NewMappingsNull(),
			attrSettings:          settings,
			attrDataStreamOptions: types.ObjectNull(datastreamoptions.AttrTypes()),
		})
		require.False(t, diags.HasError(), "%v", diags)

		var plan, state, config Data
		plan.Template = types.ObjectUnknown(templateAttrTypes())
		state.Template = stateTpl
		merged, diags := templateutil.ReconcilePlanModelForSemanticDrift(ctx, plan, state, config, templateAttrTypes)
		require.False(t, diags.HasError(), "%v", diags)
		require.Nil(t, merged)
	})
}

func TestReconcilePlanWithPriorStateForSemanticDrift_noSemanticDriftReturnsNil(t *testing.T) {
	ctx := context.Background()
	settings := customtypes.NewIndexSettingsValue(`{"index":{"number_of_shards":1}}`)
	emptyAlias, diags := types.SetValueFrom(ctx, aliasutil.NewAliasObjectType(), []attr.Value{})
	require.False(t, diags.HasError(), "%v", diags)

	tpl, diags := types.ObjectValue(templateAttrTypes(), map[string]attr.Value{
		attrAlias:             emptyAlias,
		attrMappings:          esindex.NewMappingsNull(),
		attrSettings:          settings,
		attrDataStreamOptions: types.ObjectNull(datastreamoptions.AttrTypes()),
	})
	require.False(t, diags.HasError(), "%v", diags)

	var plan, state, config Data
	plan.Template = tpl
	state.Template = tpl
	config.Template = tpl
	merged, diags := templateutil.ReconcilePlanModelForSemanticDrift(ctx, plan, state, config, templateAttrTypes)
	require.False(t, diags.HasError(), "%v", diags)
	require.Nil(t, merged)
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
