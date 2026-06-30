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
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
		require.NotNil(t, tpl.DataStreamOptions.FailureStore.Enabled)
		assert.True(t, *tpl.DataStreamOptions.FailureStore.Enabled)
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

func TestExpandMetadataJSON(t *testing.T) {
	t.Run("null value returns nil", func(t *testing.T) {
		var diags diag.Diagnostics
		result := ExpandMetadataJSON(jsontypes.NewNormalizedNull(), &diags)
		require.False(t, diags.HasError())
		assert.Nil(t, result)
	})

	t.Run("unknown value returns nil", func(t *testing.T) {
		var diags diag.Diagnostics
		result := ExpandMetadataJSON(jsontypes.NewNormalizedUnknown(), &diags)
		require.False(t, diags.HasError())
		assert.Nil(t, result)
	})

	t.Run("empty string returns nil", func(t *testing.T) {
		var diags diag.Diagnostics
		result := ExpandMetadataJSON(jsontypes.NewNormalizedValue(""), &diags)
		require.False(t, diags.HasError())
		assert.Nil(t, result)
	})

	t.Run("whitespace-only string returns nil", func(t *testing.T) {
		var diags diag.Diagnostics
		result := ExpandMetadataJSON(jsontypes.NewNormalizedValue("   "), &diags)
		require.False(t, diags.HasError())
		assert.Nil(t, result)
	})

	t.Run("valid JSON returns parsed map", func(t *testing.T) {
		var diags diag.Diagnostics
		result := ExpandMetadataJSON(jsontypes.NewNormalizedValue(`{"key":"value","num":42}`), &diags)
		require.False(t, diags.HasError())
		require.NotNil(t, result)
		assert.Equal(t, "value", result["key"])
		assert.InDelta(t, float64(42), result["num"], 0)
	})

	t.Run("whitespace-padded valid JSON is accepted", func(t *testing.T) {
		var diags diag.Diagnostics
		result := ExpandMetadataJSON(jsontypes.NewNormalizedValue(`  {"key":"value"}  `), &diags)
		require.False(t, diags.HasError())
		require.NotNil(t, result)
		assert.Equal(t, "value", result["key"])
	})

	t.Run("invalid JSON adds error diagnostic", func(t *testing.T) {
		var diags diag.Diagnostics
		result := ExpandMetadataJSON(jsontypes.NewNormalizedValue(`{invalid`), &diags)
		require.True(t, diags.HasError())
		assert.Nil(t, result)
		assert.Equal(t, "Normalized JSON Unmarshal Error", diags.Errors()[0].Summary())
	})
}
