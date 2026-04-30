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

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"
)

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
		Name:                               basetypes.NewStringValue("my-index"),
		NumberOfShards:                     basetypes.NewInt64Value(1),
		NumberOfRoutingShards:              basetypes.NewInt64Value(2),
		Codec:                              basetypes.NewStringValue("best_compression"),
		RoutingPartitionSize:               basetypes.NewInt64Value(3),
		LoadFixedBitsetFiltersEagerly:      basetypes.NewBoolValue(true),
		ShardCheckOnStartup:                basetypes.NewStringValue("checksum"),
		SortField:                          sortField,
		SortOrder:                          sortOrder,
		MappingCoerce:                      basetypes.NewBoolValue(false),
		Settings:                           basetypes.NewListNull(basetypes.ObjectType{}),
		ElasticsearchConnection:            basetypes.NewListNull(basetypes.ObjectType{}),
	}

	existing := models.Index{
		Settings: map[string]any{
			"index.number_of_shards":                 "1",
			"index.number_of_routing_shards":         "2",
			"index.codec":                            "best_compression",
			"index.routing_partition_size":           "3",
			"index.load_fixed_bitset_filters_eagerly": "true",
			"index.shard.check_on_startup":           "checksum",
			"index.sort.field":                       []any{"b", "a"},
			"index.sort.order":                       []any{"asc", "desc"},
			"index.mapping.coerce":                   "false",
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
		Name:                    basetypes.NewStringValue("i"),
		LoadFixedBitsetFiltersEagerly: basetypes.NewBoolValue(false),
		Settings:                basetypes.NewListNull(basetypes.ObjectType{}),
		ElasticsearchConnection: basetypes.NewListNull(basetypes.ObjectType{}),
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
