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

package templateutil

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/aliasutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/datastreamoptions"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var ctx = context.Background()

func aliasAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":           types.StringType,
		"filter":         jsontypes.NormalizedType{},
		"index_routing":  types.StringType,
		"is_hidden":      types.BoolType,
		"is_write_index": types.BoolType,
		"routing":        types.StringType,
		"search_routing": types.StringType,
	}
}

func mustAliasSet(t *testing.T, elems ...aliasutil.AliasModel) types.Set {
	attrValues := make([]attr.Value, len(elems))
	for i, e := range elems {
		obj, diags := types.ObjectValueFrom(ctx, aliasAttrTypes(), e)
		require.False(t, diags.HasError(), "object value error: %v", diags.Errors())
		attrValues[i] = obj
	}
	set, diags := types.SetValue(types.ObjectType{AttrTypes: aliasAttrTypes()}, attrValues)
	require.False(t, diags.HasError(), "set value error: %v", diags.Errors())
	return set
}

func TestExpandTemplateCore(t *testing.T) {
	t.Run("all null", func(t *testing.T) {
		tpl, diags := ExpandTemplateCore(
			ctx,
			types.SetNull(types.ObjectType{AttrTypes: aliasAttrTypes()}),
			index.NewMappingsNull(),
			customtypes.NewIndexSettingsNull(),
			types.ObjectNull(datastreamoptions.AttrTypes()),
		)
		require.False(t, diags.HasError())
		assert.NotNil(t, tpl)
		assert.Empty(t, tpl.Aliases)
		assert.Empty(t, tpl.Mappings)
		assert.Empty(t, tpl.Settings)
		assert.Nil(t, tpl.DataStreamOptions)
	})

	t.Run("all unknown", func(t *testing.T) {
		tpl, diags := ExpandTemplateCore(
			ctx,
			types.SetUnknown(types.ObjectType{AttrTypes: aliasAttrTypes()}),
			index.NewMappingsUnknown(),
			customtypes.NewIndexSettingsUnknown(),
			types.ObjectUnknown(datastreamoptions.AttrTypes()),
		)
		require.False(t, diags.HasError())
		assert.NotNil(t, tpl)
		assert.Empty(t, tpl.Aliases)
		assert.Empty(t, tpl.Mappings)
		assert.Empty(t, tpl.Settings)
		assert.Nil(t, tpl.DataStreamOptions)
	})

	t.Run("aliases populated", func(t *testing.T) {
		aliasSet := mustAliasSet(t, aliasutil.AliasModel{
			Name:   types.StringValue("my-alias"),
			Filter: jsontypes.NewNormalizedNull(),
		})

		tpl, diags := ExpandTemplateCore(
			ctx, aliasSet,
			index.NewMappingsNull(),
			customtypes.NewIndexSettingsNull(),
			types.ObjectNull(datastreamoptions.AttrTypes()),
		)
		require.False(t, diags.HasError())
		assert.Len(t, tpl.Aliases, 1)
		assert.Equal(t, "my-alias", tpl.Aliases["my-alias"].Name)
	})

	t.Run("mappings populated", func(t *testing.T) {
		tpl, diags := ExpandTemplateCore(
			ctx,
			types.SetNull(types.ObjectType{AttrTypes: aliasAttrTypes()}),
			index.NewMappingsValue(`{"properties":{"title":{"type":"text"}}}`),
			customtypes.NewIndexSettingsNull(),
			types.ObjectNull(datastreamoptions.AttrTypes()),
		)
		require.False(t, diags.HasError())
		assert.Equal(t, map[string]any{"properties": map[string]any{"title": map[string]any{"type": "text"}}}, tpl.Mappings)
	})

	t.Run("mappings invalid json", func(t *testing.T) {
		_, diags := ExpandTemplateCore(
			ctx,
			types.SetNull(types.ObjectType{AttrTypes: aliasAttrTypes()}),
			index.NewMappingsValue(`{invalid`),
			customtypes.NewIndexSettingsNull(),
			types.ObjectNull(datastreamoptions.AttrTypes()),
		)
		require.True(t, diags.HasError())
		assert.Contains(t, diags.Errors()[0].Summary(), "Invalid template.mappings JSON")
	})

	t.Run("settings populated", func(t *testing.T) {
		tpl, diags := ExpandTemplateCore(
			ctx,
			types.SetNull(types.ObjectType{AttrTypes: aliasAttrTypes()}),
			index.NewMappingsNull(),
			customtypes.NewIndexSettingsValue(`{"number_of_shards":1}`),
			types.ObjectNull(datastreamoptions.AttrTypes()),
		)
		require.False(t, diags.HasError())
		assert.Equal(t, map[string]any{"number_of_shards": float64(1)}, tpl.Settings)
	})

	t.Run("settings invalid json", func(t *testing.T) {
		_, diags := ExpandTemplateCore(
			ctx,
			types.SetNull(types.ObjectType{AttrTypes: aliasAttrTypes()}),
			index.NewMappingsNull(),
			customtypes.NewIndexSettingsValue(`{invalid`),
			types.ObjectNull(datastreamoptions.AttrTypes()),
		)
		require.True(t, diags.HasError())
		assert.Contains(t, diags.Errors()[0].Summary(), "Invalid template.settings JSON")
	})

	t.Run("data stream options populated", func(t *testing.T) {
		fsAttrs := map[string]attr.Value{
			"enabled":   types.BoolValue(true),
			"lifecycle": types.ObjectNull(datastreamoptions.FailureStoreLifecycleAttrTypes()),
		}
		fsObj, diags := types.ObjectValue(datastreamoptions.FailureStoreAttrTypes(), fsAttrs)
		require.False(t, diags.HasError())
		dsoAttrs := map[string]attr.Value{
			"failure_store": fsObj,
		}
		dsoObj, diags := types.ObjectValue(datastreamoptions.AttrTypes(), dsoAttrs)
		require.False(t, diags.HasError())

		tpl, diags := ExpandTemplateCore(ctx, types.SetNull(types.ObjectType{AttrTypes: aliasAttrTypes()}), index.NewMappingsNull(), customtypes.NewIndexSettingsNull(), dsoObj)
		require.False(t, diags.HasError())
		require.NotNil(t, tpl.DataStreamOptions)
		require.NotNil(t, tpl.DataStreamOptions.FailureStore)
		assert.True(t, tpl.DataStreamOptions.FailureStore.Enabled)
	})

	t.Run("data stream options null", func(t *testing.T) {
		tpl, diags := ExpandTemplateCore(
			ctx,
			types.SetNull(types.ObjectType{AttrTypes: aliasAttrTypes()}),
			index.NewMappingsNull(),
			customtypes.NewIndexSettingsNull(),
			types.ObjectNull(datastreamoptions.AttrTypes()),
		)
		require.False(t, diags.HasError())
		assert.Nil(t, tpl.DataStreamOptions)
	})

	t.Run("all fields populated", func(t *testing.T) {
		aliasSet := mustAliasSet(t, aliasutil.AliasModel{
			Name:   types.StringValue("my-alias"),
			Filter: jsontypes.NewNormalizedNull(),
		})
		mappings := index.NewMappingsValue(`{"properties":{"title":{"type":"text"}}}`)
		settings := customtypes.NewIndexSettingsValue(`{"number_of_shards":1}`)

		fsAttrs := map[string]attr.Value{
			"enabled":   types.BoolValue(true),
			"lifecycle": types.ObjectNull(datastreamoptions.FailureStoreLifecycleAttrTypes()),
		}
		fsObj, d := types.ObjectValue(datastreamoptions.FailureStoreAttrTypes(), fsAttrs)
		require.False(t, d.HasError())
		dsoAttrs := map[string]attr.Value{
			"failure_store": fsObj,
		}
		dsoObj, d := types.ObjectValue(datastreamoptions.AttrTypes(), dsoAttrs)
		require.False(t, d.HasError())

		tpl, diags := ExpandTemplateCore(ctx, aliasSet, mappings, settings, dsoObj)
		require.False(t, diags.HasError())
		assert.Len(t, tpl.Aliases, 1)
		assert.Equal(t, "my-alias", tpl.Aliases["my-alias"].Name)
		assert.Equal(t, map[string]any{"properties": map[string]any{"title": map[string]any{"type": "text"}}}, tpl.Mappings)
		assert.Equal(t, map[string]any{"number_of_shards": float64(1)}, tpl.Settings)
		assert.NotNil(t, tpl.DataStreamOptions)
	})
}

func TestDecodeTemplateObject(t *testing.T) {
	t.Run("null object returns empty diags", func(t *testing.T) {
		var m testModel
		diags := DecodeTemplateObject(ctx, types.ObjectNull(map[string]attr.Type{}), &m)
		require.False(t, diags.HasError())
		assert.True(t, m.Field.IsNull())
	})

	t.Run("unknown object returns empty diags", func(t *testing.T) {
		var m testModel
		diags := DecodeTemplateObject(ctx, types.ObjectUnknown(map[string]attr.Type{}), &m)
		require.False(t, diags.HasError())
		assert.True(t, m.Field.IsNull())
	})

	t.Run("valid object decodes into model", func(t *testing.T) {
		attrTypes := map[string]attr.Type{
			"field": types.StringType,
		}
		obj, d := types.ObjectValue(attrTypes, map[string]attr.Value{"field": types.StringValue("hello")})
		require.False(t, d.HasError())

		var m testModel
		diags := DecodeTemplateObject(ctx, obj, &m)
		require.False(t, diags.HasError())
		assert.Equal(t, "hello", m.Field.ValueString())
	})
}

type testModel struct {
	Field types.String `tfsdk:"field"`
}
