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

package index

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"
)

func deprecatedSettingsBlockList(ctx context.Context, t *testing.T, entries []settingTfModel) types.List {
	t.Helper()

	settingSet, diags := basetypes.NewSetValueFrom(ctx, settingElementType(), entries)
	require.Empty(t, diags)

	obj, diags := basetypes.NewObjectValue(
		map[string]attr.Type{"setting": basetypes.SetType{ElemType: settingElementType()}},
		map[string]attr.Value{"setting": settingSet},
	)
	require.Empty(t, diags)

	list, diags := basetypes.NewListValue(settingsElementType(), []attr.Value{obj})
	require.Empty(t, diags)

	return list
}

func Test_compareStaticSettings_nilPlan(t *testing.T) {
	t.Parallel()
	_, diags := compareStaticSettings(context.Background(), nil, models.Index{})
	require.True(t, diags.HasError())
}

func Test_compareStaticSettings_noMismatches_allStaticMatch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	sortField, diags := basetypes.NewSetValueFrom(ctx, basetypes.StringType{}, []string{"a", "b"})
	require.Empty(t, diags)
	sortOrder, diags := basetypes.NewListValueFrom(ctx, basetypes.StringType{}, []string{"asc", "desc"})
	require.Empty(t, diags)

	plan := &tfModel{
		Name:                          basetypes.NewStringValue("my-index"),
		NumberOfShards:                basetypes.NewInt64Value(1),
		NumberOfRoutingShards:         basetypes.NewInt64Value(2),
		Codec:                         basetypes.NewStringValue("best_compression"),
		RoutingPartitionSize:          basetypes.NewInt64Value(3),
		LoadFixedBitsetFiltersEagerly: basetypes.NewBoolValue(true),
		ShardCheckOnStartup:           basetypes.NewStringValue("checksum"),
		SortField:                     sortField,
		SortOrder:                     sortOrder,
		MappingCoerce:                 basetypes.NewBoolValue(false),
		Settings:                      basetypes.NewListNull(basetypes.ObjectType{}),
		ElasticsearchConnection:       basetypes.NewListNull(basetypes.ObjectType{}),
	}

	existing := models.Index{
		Settings: map[string]any{
			"index.number_of_shards":                  "1",
			"index.number_of_routing_shards":          "2",
			"index.codec":                             "best_compression",
			"index.routing_partition_size":            "3",
			"index.load_fixed_bitset_filters_eagerly": "true",
			"index.shard.check_on_startup":            "checksum",
			"index.sort.field":                        []any{"b", "a"},
			"index.sort.order":                        []any{"asc", "desc"},
			"index.mapping.coerce":                    "false",
		},
	}

	mismatches, diags := compareStaticSettings(ctx, plan, existing)
	require.Empty(t, diags)
	require.Empty(t, mismatches)
}

func Test_compareStaticSettings_singleMismatch_numberOfShards(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	plan := &tfModel{
		Name:                    basetypes.NewStringValue("i"),
		NumberOfShards:          basetypes.NewInt64Value(1),
		Settings:                basetypes.NewListNull(basetypes.ObjectType{}),
		ElasticsearchConnection: basetypes.NewListNull(basetypes.ObjectType{}),
	}
	existing := models.Index{
		Settings: map[string]any{
			"index.number_of_shards": "2",
		},
	}

	mismatches, diags := compareStaticSettings(ctx, plan, existing)
	require.Empty(t, diags)
	require.Len(t, mismatches, 1)
	require.Equal(t, "number_of_shards", mismatches[0].Attribute)
	require.Equal(t, "1", mismatches[0].Configured)
	require.Equal(t, "2", mismatches[0].Actual)
}

func Test_compareStaticSettings_multipleMismatches(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	plan := &tfModel{
		Name:                    basetypes.NewStringValue("i"),
		NumberOfShards:          basetypes.NewInt64Value(1),
		Codec:                   basetypes.NewStringValue("best_compression"),
		Settings:                basetypes.NewListNull(basetypes.ObjectType{}),
		ElasticsearchConnection: basetypes.NewListNull(basetypes.ObjectType{}),
	}
	existing := models.Index{
		Settings: map[string]any{
			"index.number_of_shards": "3",
			"index.codec":            "default",
		},
	}

	mismatches, diags := compareStaticSettings(ctx, plan, existing)
	require.Empty(t, diags)
	require.Len(t, mismatches, 2)
	attrs := []string{mismatches[0].Attribute, mismatches[1].Attribute}
	require.ElementsMatch(t, []string{"number_of_shards", "codec"}, attrs)
}

func Test_compareStaticSettings_skipsNullAndUnknown(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	plan := &tfModel{
		Name:                    basetypes.NewStringValue("i"),
		NumberOfShards:          basetypes.NewInt64Null(),
		Codec:                   basetypes.NewStringUnknown(),
		Settings:                basetypes.NewListNull(basetypes.ObjectType{}),
		ElasticsearchConnection: basetypes.NewListNull(basetypes.ObjectType{}),
	}
	existing := models.Index{
		Settings: map[string]any{
			"index.number_of_shards": "99",
			"index.codec":            "nope",
		},
	}

	mismatches, diags := compareStaticSettings(ctx, plan, existing)
	require.Empty(t, diags)
	require.Empty(t, mismatches)
}

func Test_compareStaticSettings_normalization_intString(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	plan := &tfModel{
		Name:                    basetypes.NewStringValue("i"),
		RoutingPartitionSize:    basetypes.NewInt64Value(1),
		Settings:                basetypes.NewListNull(basetypes.ObjectType{}),
		ElasticsearchConnection: basetypes.NewListNull(basetypes.ObjectType{}),
	}
	existing := models.Index{
		Settings: map[string]any{
			"index.routing_partition_size": "1",
		},
	}

	mismatches, diags := compareStaticSettings(ctx, plan, existing)
	require.Empty(t, diags)
	require.Empty(t, mismatches)
}

func Test_compareStaticSettings_normalization_intFloat64(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	plan := &tfModel{
		Name:                    basetypes.NewStringValue("i"),
		NumberOfShards:          basetypes.NewInt64Value(2),
		Settings:                basetypes.NewListNull(basetypes.ObjectType{}),
		ElasticsearchConnection: basetypes.NewListNull(basetypes.ObjectType{}),
	}
	existing := models.Index{
		Settings: map[string]any{
			"index.number_of_shards": float64(2),
		},
	}

	mismatches, diags := compareStaticSettings(ctx, plan, existing)
	require.Empty(t, diags)
	require.Empty(t, mismatches)
}

func Test_compareStaticSettings_normalization_boolString(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	plan := &tfModel{
		Name:                    basetypes.NewStringValue("i"),
		MappingCoerce:           basetypes.NewBoolValue(true),
		Settings:                basetypes.NewListNull(basetypes.ObjectType{}),
		ElasticsearchConnection: basetypes.NewListNull(basetypes.ObjectType{}),
	}
	existing := models.Index{
		Settings: map[string]any{
			"index.mapping.coerce": "true",
		},
	}

	mismatches, diags := compareStaticSettings(ctx, plan, existing)
	require.Empty(t, diags)
	require.Empty(t, mismatches)
}

func Test_compareStaticSettings_normalization_string(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	plan := &tfModel{
		Name:                    basetypes.NewStringValue("i"),
		Codec:                   basetypes.NewStringValue("best_compression"),
		Settings:                basetypes.NewListNull(basetypes.ObjectType{}),
		ElasticsearchConnection: basetypes.NewListNull(basetypes.ObjectType{}),
	}
	existing := models.Index{
		Settings: map[string]any{
			"index.codec": "best_compression",
		},
	}

	mismatches, diags := compareStaticSettings(ctx, plan, existing)
	require.Empty(t, diags)
	require.Empty(t, mismatches)
}

func Test_compareStaticSettings_intParseFailureIsMismatch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	plan := &tfModel{
		Name:                    basetypes.NewStringValue("i"),
		NumberOfShards:          basetypes.NewInt64Value(1),
		Settings:                basetypes.NewListNull(basetypes.ObjectType{}),
		ElasticsearchConnection: basetypes.NewListNull(basetypes.ObjectType{}),
	}
	existing := models.Index{
		Settings: map[string]any{
			"index.number_of_shards": "not-a-number",
		},
	}

	mismatches, diags := compareStaticSettings(ctx, plan, existing)
	require.Empty(t, diags)
	require.Len(t, mismatches, 1)
	require.Equal(t, "not-a-number", mismatches[0].Actual)
}

func Test_compareStaticSettings_sortField_setVsArray(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	sortField, diags := basetypes.NewSetValueFrom(ctx, basetypes.StringType{}, []string{"a", "b"})
	require.Empty(t, diags)

	plan := &tfModel{
		Name:                    basetypes.NewStringValue("i"),
		SortField:               sortField,
		Settings:                basetypes.NewListNull(basetypes.ObjectType{}),
		ElasticsearchConnection: basetypes.NewListNull(basetypes.ObjectType{}),
	}
	existing := models.Index{
		Settings: map[string]any{
			"index.sort.field": []any{"a", "b"},
		},
	}

	mismatches, diags := compareStaticSettings(ctx, plan, existing)
	require.Empty(t, diags)
	require.Empty(t, mismatches)
}

func Test_compareStaticSettings_sortField_setVsSingleString(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	sortField, diags := basetypes.NewSetValueFrom(ctx, basetypes.StringType{}, []string{"only"})
	require.Empty(t, diags)

	plan := &tfModel{
		Name:                    basetypes.NewStringValue("i"),
		SortField:               sortField,
		Settings:                basetypes.NewListNull(basetypes.ObjectType{}),
		ElasticsearchConnection: basetypes.NewListNull(basetypes.ObjectType{}),
	}
	existing := models.Index{
		Settings: map[string]any{
			"index.sort.field": "only",
		},
	}

	mismatches, diags := compareStaticSettings(ctx, plan, existing)
	require.Empty(t, diags)
	require.Empty(t, mismatches)
}

func Test_compareStaticSettings_sortOrder_orderedEquality(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	sortOrder, diags := basetypes.NewListValueFrom(ctx, basetypes.StringType{}, []string{"asc", "desc"})
	require.Empty(t, diags)

	plan := &tfModel{
		Name:                    basetypes.NewStringValue("i"),
		SortOrder:               sortOrder,
		Settings:                basetypes.NewListNull(basetypes.ObjectType{}),
		ElasticsearchConnection: basetypes.NewListNull(basetypes.ObjectType{}),
	}
	existing := models.Index{
		Settings: map[string]any{
			"index.sort.order": []any{"asc", "desc"},
		},
	}

	mismatches, diags := compareStaticSettings(ctx, plan, existing)
	require.Empty(t, diags)
	require.Empty(t, mismatches)
}

func Test_compareStaticSettings_sortOrder_singleStringVsList(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	sortOrder, diags := basetypes.NewListValueFrom(ctx, basetypes.StringType{}, []string{"asc"})
	require.Empty(t, diags)

	plan := &tfModel{
		Name:                    basetypes.NewStringValue("i"),
		SortOrder:               sortOrder,
		Settings:                basetypes.NewListNull(basetypes.ObjectType{}),
		ElasticsearchConnection: basetypes.NewListNull(basetypes.ObjectType{}),
	}
	existing := models.Index{
		Settings: map[string]any{
			"index.sort.order": "asc",
		},
	}

	mismatches, diags := compareStaticSettings(ctx, plan, existing)
	require.Empty(t, diags)
	require.Empty(t, mismatches)
}

func Test_compareStaticSettings_bareKeyWithoutIndexPrefix(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	plan := &tfModel{
		Name:                    basetypes.NewStringValue("i"),
		NumberOfShards:          basetypes.NewInt64Value(5),
		Settings:                basetypes.NewListNull(basetypes.ObjectType{}),
		ElasticsearchConnection: basetypes.NewListNull(basetypes.ObjectType{}),
	}
	existing := models.Index{
		Settings: map[string]any{
			"number_of_shards": "5",
		},
	}

	mismatches, diags := compareStaticSettings(ctx, plan, existing)
	require.Empty(t, diags)
	require.Empty(t, mismatches)
}

func Test_compareStaticSettings_prefixedKeyPreferred(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	plan := &tfModel{
		Name:                          basetypes.NewStringValue("i"),
		LoadFixedBitsetFiltersEagerly: basetypes.NewBoolValue(false),
		Settings:                      basetypes.NewListNull(basetypes.ObjectType{}),
		ElasticsearchConnection:       basetypes.NewListNull(basetypes.ObjectType{}),
	}
	existing := models.Index{
		Settings: map[string]any{
			"index.load_fixed_bitset_filters_eagerly": "false",
			"load_fixed_bitset_filters_eagerly":       "true",
		},
	}

	mismatches, diags := compareStaticSettings(ctx, plan, existing)
	require.Empty(t, diags)
	require.Empty(t, mismatches)
}

func Test_compareStaticSettings_missingSettingReportsAbsent(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	plan := &tfModel{
		Name:                    basetypes.NewStringValue("i"),
		NumberOfShards:          basetypes.NewInt64Value(1),
		Settings:                basetypes.NewListNull(basetypes.ObjectType{}),
		ElasticsearchConnection: basetypes.NewListNull(basetypes.ObjectType{}),
	}

	mismatches, diags := compareStaticSettings(ctx, plan, models.Index{Settings: map[string]any{}})
	require.Empty(t, diags)
	require.Len(t, mismatches, 1)
	require.Equal(t, "<absent>", mismatches[0].Actual)
}

func Test_compareStaticSettings_sortOrder_orderMatters(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	sortOrder, diags := basetypes.NewListValueFrom(ctx, basetypes.StringType{}, []string{"asc", "desc"})
	require.Empty(t, diags)

	plan := &tfModel{
		Name:                    basetypes.NewStringValue("i"),
		SortOrder:               sortOrder,
		Settings:                basetypes.NewListNull(basetypes.ObjectType{}),
		ElasticsearchConnection: basetypes.NewListNull(basetypes.ObjectType{}),
	}
	existing := models.Index{
		Settings: map[string]any{
			"index.sort.order": []any{"desc", "asc"},
		},
	}

	mismatches, diags := compareStaticSettings(ctx, plan, existing)
	require.Empty(t, diags)
	require.Len(t, mismatches, 1)
	require.Equal(t, "sort_order", mismatches[0].Attribute)
}

func Test_compareStaticSettings_deprecatedSettingsBlock_staticMismatch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	settings := deprecatedSettingsBlockList(ctx, t, []settingTfModel{
		{Name: basetypes.NewStringValue("number_of_shards"), Value: basetypes.NewStringValue("3")},
	})

	plan := &tfModel{
		Name:                    basetypes.NewStringValue("i"),
		Settings:                settings,
		ElasticsearchConnection: basetypes.NewListNull(basetypes.ObjectType{}),
	}
	existing := models.Index{
		Settings: map[string]any{
			"index.number_of_shards": "2",
		},
	}

	mismatches, diags := compareStaticSettings(ctx, plan, existing)
	require.Empty(t, diags)
	require.Len(t, mismatches, 1)
	require.Equal(t, "number_of_shards", mismatches[0].Attribute)
	require.Equal(t, "3", mismatches[0].Configured)
	require.Equal(t, "2", mismatches[0].Actual)
}

func Test_compareStaticSettings_deprecatedSettingsBlock_staticMatch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	settings := deprecatedSettingsBlockList(ctx, t, []settingTfModel{
		{Name: basetypes.NewStringValue("number_of_shards"), Value: basetypes.NewStringValue("2")},
	})

	plan := &tfModel{
		Name:                    basetypes.NewStringValue("i"),
		Settings:                settings,
		ElasticsearchConnection: basetypes.NewListNull(basetypes.ObjectType{}),
	}
	existing := models.Index{
		Settings: map[string]any{
			"index.number_of_shards": "2",
		},
	}

	mismatches, diags := compareStaticSettings(ctx, plan, existing)
	require.Empty(t, diags)
	require.Empty(t, mismatches)
}

func Test_compareStaticSettings_sortField_mismatch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	sortField, diags := basetypes.NewSetValueFrom(ctx, basetypes.StringType{}, []string{"a", "b"})
	require.Empty(t, diags)

	plan := &tfModel{
		Name:                    basetypes.NewStringValue("i"),
		SortField:               sortField,
		Settings:                basetypes.NewListNull(basetypes.ObjectType{}),
		ElasticsearchConnection: basetypes.NewListNull(basetypes.ObjectType{}),
	}
	existing := models.Index{
		Settings: map[string]any{
			"index.sort.field": []any{"a", "c"},
		},
	}

	mismatches, diags := compareStaticSettings(ctx, plan, existing)
	require.Empty(t, diags)
	require.Len(t, mismatches, 1)
	require.Equal(t, "sort_field", mismatches[0].Attribute)
	require.Equal(t, "a, b", mismatches[0].Configured)
	require.Equal(t, "a, c", mismatches[0].Actual)
}

func Test_formatStaticSettingMismatchesDetail(t *testing.T) {
	t.Parallel()
	detail := formatStaticSettingMismatchesDetail("my-index", []staticSettingMismatch{
		{Attribute: "number_of_shards", Configured: "2", Actual: "1"},
		{Attribute: "codec", Configured: "default", Actual: "best_compression"},
	})
	require.Contains(t, detail, "concrete_name: my-index")
	require.Contains(t, detail, "number_of_shards: configured=2, actual=1")
	require.Contains(t, detail, "codec: configured=default, actual=best_compression")
}

func Test_useExistingDateMathNameMatchesGateRegex(t *testing.T) {
	t.Parallel()
	require.True(t, elasticsearch.DateMathIndexNameRe.MatchString("<logs-{now/d}>"))
	// Same shape as TestAccResourceIndexUseExistingDateMath (random label between angle brackets).
	require.True(t, elasticsearch.DateMathIndexNameRe.MatchString("<useexist-abcdefghij-{now/d}>"))
}

func Test_populateFromAPI_syntheticAdoptPriorState(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	indexName := "synthetic-adopt-unit-index"

	m := &tfModel{
		ElasticsearchConnection: basetypes.NewListNull(basetypes.ObjectType{}),
	}
	apiModel := models.Index{
		Aliases: map[string]models.IndexAlias{
			"alias_a": {IsWriteIndex: true},
		},
		Mappings: map[string]any{
			"properties": map[string]any{
				"foo": map[string]any{"type": "keyword"},
			},
		},
		Settings: map[string]any{
			"index.number_of_shards": "1",
		},
	}

	apiState, convertDiags := modelToIndexState(apiModel)
	require.False(t, convertDiags.HasError())

	diags := m.populateFromAPI(ctx, indexName, apiState)
	require.False(t, diags.HasError())

	require.Equal(t, indexName, m.Name.ValueString())
	require.Equal(t, indexName, m.ConcreteName.ValueString())

	require.False(t, m.Mappings.IsNull())
	require.Contains(t, m.Mappings.ValueString(), "foo")
	require.Contains(t, m.Mappings.ValueString(), "keyword")

	var aliases []aliasTfModel
	diags = m.Alias.ElementsAs(ctx, &aliases, true)
	require.False(t, diags.HasError())
	require.Len(t, aliases, 1)
	require.Equal(t, "alias_a", aliases[0].Name.ValueString())

	require.False(t, m.SettingsRaw.IsNull())
	require.Contains(t, m.SettingsRaw.ValueString(), "number_of_shards")
}
