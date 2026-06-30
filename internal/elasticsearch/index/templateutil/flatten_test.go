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

	esindex "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsKnownSemanticallyEmpty(t *testing.T) {
	t.Run("mappings null returns false", func(t *testing.T) {
		assert.False(t, IsKnownSemanticallyEmpty(esindex.NewMappingsNull()))
	})

	t.Run("mappings unknown returns false", func(t *testing.T) {
		assert.False(t, IsKnownSemanticallyEmpty(esindex.NewMappingsUnknown()))
	})

	t.Run("mappings empty JSON object returns true", func(t *testing.T) {
		assert.True(t, IsKnownSemanticallyEmpty(esindex.NewMappingsValue("{}")))
	})

	t.Run("mappings whitespace-padded empty JSON object returns true", func(t *testing.T) {
		assert.True(t, IsKnownSemanticallyEmpty(esindex.NewMappingsValue("  {}  ")))
	})

	t.Run("mappings non-empty JSON object returns false", func(t *testing.T) {
		assert.False(t, IsKnownSemanticallyEmpty(esindex.NewMappingsValue(`{"properties":{}}`)))
	})

	t.Run("settings null returns false", func(t *testing.T) {
		assert.False(t, IsKnownSemanticallyEmpty(customtypes.NewIndexSettingsNull()))
	})

	t.Run("settings unknown returns false", func(t *testing.T) {
		assert.False(t, IsKnownSemanticallyEmpty(customtypes.NewIndexSettingsUnknown()))
	})

	t.Run("settings empty JSON object returns true", func(t *testing.T) {
		assert.True(t, IsKnownSemanticallyEmpty(customtypes.NewIndexSettingsValue("{}")))
	})

	t.Run("settings whitespace-padded empty JSON object returns true", func(t *testing.T) {
		assert.True(t, IsKnownSemanticallyEmpty(customtypes.NewIndexSettingsValue("  {}  ")))
	})

	t.Run("settings non-empty JSON object returns false", func(t *testing.T) {
		assert.False(t, IsKnownSemanticallyEmpty(customtypes.NewIndexSettingsValue(`{"number_of_shards":1}`)))
	})
}

// aliasAttrTypesForTest returns the shared alias attribute types used in tests.
func aliasAttrTypesForTest() map[string]attr.Type {
	return map[string]attr.Type{
		"name":           types.StringType,
		"filter":         jsontypes.NormalizedType{},
		"index_routing":  types.StringType,
		"routing":        types.StringType,
		"search_routing": types.StringType,
		"is_hidden":      types.BoolType,
		"is_write_index": types.BoolType,
	}
}

func TestFlattenTemplateCore_nilTemplate(t *testing.T) {
	ctx := context.Background()
	attrTypes := aliasAttrTypesForTest()
	elemType := types.ObjectType{AttrTypes: attrTypes}

	// FlattenTemplateCore should handle a template with all nil fields gracefully.
	tpl := &models.Template{}
	result, diags := FlattenTemplateCore(
		ctx,
		tpl,
		esindex.NewMappingsNull(),
		customtypes.NewIndexSettingsNull(),
		nil,
		elemType,
		attrTypes,
	)
	require.False(t, diags.HasError(), "unexpected diags: %v", diags)
	assert.True(t, result.AliasSet.IsNull(), "alias set should be null when template has no aliases")
	assert.True(t, result.Mappings.IsNull(), "mappings should be null when API omits them")
	assert.True(t, result.Settings.IsNull(), "settings should be null when API omits them")
	assert.True(t, result.DsoObj.IsNull(), "data_stream_options should be null when API omits them")
}

func TestFlattenTemplateCore_mappingsAndSettingsPreservation(t *testing.T) {
	ctx := context.Background()
	attrTypes := aliasAttrTypesForTest()
	elemType := types.ObjectType{AttrTypes: attrTypes}

	cases := []struct {
		name             string
		apiMappings      map[string]any
		apiSettings      map[string]any
		priorMappings    esindex.MappingsValue
		priorSettings    customtypes.IndexSettingsValue
		wantMappingsNull bool
		wantSettingsNull bool
	}{
		{
			name:             "nil API fields and null priors produce null values",
			apiMappings:      nil,
			apiSettings:      nil,
			priorMappings:    esindex.NewMappingsNull(),
			priorSettings:    customtypes.NewIndexSettingsNull(),
			wantMappingsNull: true,
			wantSettingsNull: true,
		},
		{
			name:             "empty API fields and empty-object priors are preserved",
			apiMappings:      map[string]any{},
			apiSettings:      map[string]any{},
			priorMappings:    esindex.NewMappingsValue(`{}`),
			priorSettings:    customtypes.NewIndexSettingsValue(`{}`),
			wantMappingsNull: false,
			wantSettingsNull: false,
		},
		{
			name:             "non-empty API fields override priors",
			apiMappings:      map[string]any{"properties": map[string]any{"f": map[string]any{"type": "keyword"}}},
			apiSettings:      map[string]any{"number_of_shards": "1"},
			priorMappings:    esindex.NewMappingsValue(`{}`),
			priorSettings:    customtypes.NewIndexSettingsValue(`{}`),
			wantMappingsNull: false,
			wantSettingsNull: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tpl := &models.Template{
				Mappings: tc.apiMappings,
				Settings: tc.apiSettings,
			}
			result, diags := FlattenTemplateCore(ctx, tpl, tc.priorMappings, tc.priorSettings, nil, elemType, attrTypes)
			require.False(t, diags.HasError(), "unexpected diags: %v", diags)
			assert.Equal(t, tc.wantMappingsNull, result.Mappings.IsNull(), "mappings null mismatch")
			assert.Equal(t, tc.wantSettingsNull, result.Settings.IsNull(), "settings null mismatch")
		})
	}
}

func TestFlattenTemplateCore_aliasesSortedAndRoutingPreserved(t *testing.T) {
	ctx := context.Background()
	attrTypes := aliasAttrTypesForTest()
	elemType := types.ObjectType{AttrTypes: attrTypes}

	aliases := map[string]models.IndexAlias{
		"z-alias": {Routing: ""},
		"a-alias": {Routing: "api-routing"},
		"m-alias": {Routing: ""},
	}
	preservedRouting := map[string]string{
		"z-alias": "preserved-z",
		"m-alias": "preserved-m",
	}

	tpl := &models.Template{Aliases: aliases}
	result, diags := FlattenTemplateCore(ctx, tpl, esindex.NewMappingsNull(), customtypes.NewIndexSettingsNull(), preservedRouting, elemType, attrTypes)
	require.False(t, diags.HasError(), "unexpected diags: %v", diags)
	require.False(t, result.AliasSet.IsNull(), "alias set should not be null")

	// Verify elements are present and routing is correctly preserved/not-preserved.
	elems := result.AliasSet.Elements()
	require.Len(t, elems, 3)

	nameToRouting := make(map[string]string)
	for _, el := range elems {
		obj := el.(types.Object)
		attrs := obj.Attributes()
		name := attrs["name"].(types.String).ValueString()
		routing := attrs["routing"].(types.String).ValueString()
		nameToRouting[name] = routing
	}

	assert.Equal(t, "api-routing", nameToRouting["a-alias"], "a-alias routing should come from API")
	assert.Equal(t, "preserved-z", nameToRouting["z-alias"], "z-alias routing should be preserved")
	assert.Equal(t, "preserved-m", nameToRouting["m-alias"], "m-alias routing should be preserved")
}
