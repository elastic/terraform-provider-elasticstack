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

package osquerypack

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPolicyIDsToAPI(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("null omits field", func(t *testing.T) {
		result, diags := policyIDsToAPI(ctx, types.SetNull(types.StringType))
		require.False(t, diags.HasError())
		assert.Nil(t, result)
	})

	t.Run("unknown omits field", func(t *testing.T) {
		result, diags := policyIDsToAPI(ctx, types.SetUnknown(types.StringType))
		require.False(t, diags.HasError())
		assert.Nil(t, result)
	})

	t.Run("known empty sends empty slice", func(t *testing.T) {
		set, d := types.SetValueFrom(ctx, types.StringType, []string{})
		require.False(t, d.HasError())

		result, diags := policyIDsToAPI(ctx, set)
		require.False(t, diags.HasError())
		require.NotNil(t, result)
		assert.Empty(t, *result)
	})

	t.Run("known values are sorted before sending", func(t *testing.T) {
		set, d := types.SetValueFrom(ctx, types.StringType, []string{"policy-b", "policy-a"})
		require.False(t, d.HasError())

		result, diags := policyIDsToAPI(ctx, set)
		require.False(t, diags.HasError())
		require.NotNil(t, result)
		assert.Equal(t, kbapi.SecurityOsqueryAPIPolicyIds{"policy-a", "policy-b"}, *result)
	})
}

func TestShardsMapToAPI(t *testing.T) {
	t.Parallel()

	t.Run("null omits field", func(t *testing.T) {
		result := shardsMapToAPI(types.MapNull(types.Float64Type))
		assert.Nil(t, result)
	})

	t.Run("unknown omits field", func(t *testing.T) {
		result := shardsMapToAPI(types.MapUnknown(types.Float64Type))
		assert.Nil(t, result)
	})

	t.Run("known empty sends empty map", func(t *testing.T) {
		empty, d := types.MapValue(types.Float64Type, map[string]attr.Value{})
		require.False(t, d.HasError())

		result := shardsMapToAPI(empty)
		require.NotNil(t, result)
		assert.Empty(t, *result)
	})

	t.Run("converts float64 to float32", func(t *testing.T) {
		shards, d := types.MapValue(types.Float64Type, map[string]attr.Value{
			"policy-a": types.Float64Value(75.5),
		})
		require.False(t, d.HasError())

		result := shardsMapToAPI(shards)
		require.NotNil(t, result)
		assert.InDelta(t, float32(75.5), (*result)["policy-a"], 0.001)
	})
}

func TestToWriteRequestBody_fullV1Fields(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	model := fullWriteModel(ctx, t)

	body, diags := model.toWriteRequestBody(ctx)
	require.False(t, diags.HasError())

	require.NotNil(t, body.Name)
	assert.Equal(t, kbapi.SecurityOsqueryAPIPackName("pack-name"), *body.Name)
	require.NotNil(t, body.Description)
	assert.Equal(t, kbapi.SecurityOsqueryAPIPackDescription("pack description"), *body.Description)
	require.NotNil(t, body.Enabled)
	assert.True(t, *body.Enabled)
	require.NotNil(t, body.PolicyIds)
	assert.Equal(t, kbapi.SecurityOsqueryAPIPolicyIds{"policy-a"}, *body.PolicyIds)
	require.NotNil(t, body.Shards)
	assert.InDelta(t, float32(80), (*body.Shards)["policy-a"], 0.001)

	require.NotNil(t, body.Queries)
	require.Len(t, *body.Queries, 2)

	findProcs := (*body.Queries)["find_procs"]
	require.NotNil(t, findProcs.Query)
	assert.Equal(t, kbapi.SecurityOsqueryAPIQuery("SELECT * FROM processes"), *findProcs.Query)
	require.NotNil(t, findProcs.Platform)
	assert.Equal(t, kbapi.SecurityOsqueryAPIPlatform("linux,windows"), *findProcs.Platform)
	require.NotNil(t, findProcs.EcsMapping)
	require.NotNil(t, (*findProcs.EcsMapping)["process.name"].Field)
	assert.Equal(t, "cmdline", *(*findProcs.EcsMapping)["process.name"].Field)

	findUsers := (*body.Queries)["find_users"]
	require.NotNil(t, findUsers.Query)
	assert.Equal(t, kbapi.SecurityOsqueryAPIQuery("SELECT * FROM users"), *findUsers.Query)
	require.NotNil(t, findUsers.EcsMapping)
	got, err := (*findUsers.EcsMapping)["host.name"].Value.AsSecurityOsqueryAPIECSMappingItemValue0()
	require.NoError(t, err)
	assert.Equal(t, "literal", got)
}

func TestToUpdateRequestBody_fieldParityWithCreate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	model := fullWriteModel(ctx, t)

	createBody, createDiags := model.toCreateRequestBody(ctx)
	require.False(t, createDiags.HasError())

	updateBody, updateDiags := model.toUpdateRequestBody(ctx)
	require.False(t, updateDiags.HasError())

	assert.Equal(t, createBody.Description, updateBody.Description)
	assert.Equal(t, createBody.Enabled, updateBody.Enabled)
	assert.Equal(t, createBody.Name, updateBody.Name)
	assert.Equal(t, createBody.PolicyIds, updateBody.PolicyIds)
	assert.Equal(t, createBody.Queries, updateBody.Queries)
	assert.Equal(t, createBody.Shards, updateBody.Shards)
}

func TestToWriteRequestBody_nullOptionalFieldsOmitted(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	model := osqueryPackModel{
		osqueryPackBaseModel: osqueryPackBaseModel{
			Name:      types.StringValue("pack-name"),
			PolicyIDs: types.SetNull(types.StringType),
			Shards:    types.MapNull(types.Float64Type),
			Queries:   mustQueriesMap(t, singleQueryObject(t, "q1", "SELECT 1")),
		},
	}

	body, diags := model.toWriteRequestBody(ctx)
	require.False(t, diags.HasError())
	assert.Nil(t, body.PolicyIds)
	assert.Nil(t, body.Shards)
}

func TestToWriteRequestBody_knownEmptyOptionalsSent(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	policyIDs, d := types.SetValueFrom(ctx, types.StringType, []string{})
	require.False(t, d.HasError())
	shards, d := types.MapValue(types.Float64Type, map[string]attr.Value{})
	require.False(t, d.HasError())

	model := osqueryPackModel{
		osqueryPackBaseModel: osqueryPackBaseModel{
			Name:      types.StringValue("pack-name"),
			PolicyIDs: policyIDs,
			Shards:    shards,
			Queries:   mustQueriesMap(t, singleQueryObject(t, "q1", "SELECT 1")),
		},
	}

	body, diags := model.toWriteRequestBody(ctx)
	require.False(t, diags.HasError())
	require.NotNil(t, body.PolicyIds)
	assert.Empty(t, *body.PolicyIds)
	require.NotNil(t, body.Shards)
	assert.Empty(t, *body.Shards)
}

func fullWriteModel(ctx context.Context, t *testing.T) osqueryPackModel {
	t.Helper()

	policyIDs, d := types.SetValueFrom(ctx, types.StringType, []string{"policy-a"})
	require.False(t, d.HasError())
	shards, d := types.MapValue(types.Float64Type, map[string]attr.Value{
		"policy-a": types.Float64Value(80),
	})
	require.False(t, d.HasError())

	platform, d := types.SetValueFrom(ctx, types.StringType, []string{"linux", "windows"})
	require.False(t, d.HasError())

	fieldMapping, d := types.ObjectValue(ecsMappingAttrTypes(), map[string]attr.Value{
		"field":  types.StringValue("cmdline"),
		"value":  types.StringNull(),
		"values": types.SetNull(types.StringType),
	})
	require.False(t, d.HasError())
	valueMapping, d := types.ObjectValue(ecsMappingAttrTypes(), map[string]attr.Value{
		"field":  types.StringNull(),
		"value":  types.StringValue("literal"),
		"values": types.SetNull(types.StringType),
	})
	require.False(t, d.HasError())

	findProcs, d := types.ObjectValue(queryAttrTypes(), map[string]attr.Value{
		"query":          types.StringValue("SELECT * FROM processes"),
		"platform":       platform,
		"version":        types.StringNull(),
		"snapshot":       types.BoolNull(),
		"removed":        types.BoolNull(),
		"saved_query_id": types.StringNull(),
		"ecs_mapping": mustMap(t, map[string]attr.Value{
			"process.name": fieldMapping,
		}),
	})
	require.False(t, d.HasError())

	findUsers, d := types.ObjectValue(queryAttrTypes(), map[string]attr.Value{
		"query":          types.StringValue("SELECT * FROM users"),
		"platform":       types.SetNull(types.StringType),
		"version":        types.StringNull(),
		"snapshot":       types.BoolNull(),
		"removed":        types.BoolNull(),
		"saved_query_id": types.StringNull(),
		"ecs_mapping": mustMap(t, map[string]attr.Value{
			"host.name": valueMapping,
		}),
	})
	require.False(t, d.HasError())

	queries, d := types.MapValue(queryMapElemType(), map[string]attr.Value{
		"find_procs": findProcs,
		"find_users": findUsers,
	})
	require.False(t, d.HasError())

	return osqueryPackModel{
		osqueryPackBaseModel: osqueryPackBaseModel{
			Name:        types.StringValue("pack-name"),
			Description: types.StringValue("pack description"),
			Enabled:     types.BoolValue(true),
			PolicyIDs:   policyIDs,
			Shards:      shards,
			Queries:     queries,
		},
	}
}

func singleQueryObject(t *testing.T, name, query string) map[string]attr.Value {
	t.Helper()

	obj, d := types.ObjectValue(queryAttrTypes(), map[string]attr.Value{
		"query":          types.StringValue(query),
		"platform":       types.SetNull(types.StringType),
		"version":        types.StringNull(),
		"snapshot":       types.BoolNull(),
		"removed":        types.BoolNull(),
		"saved_query_id": types.StringNull(),
		"ecs_mapping":    types.MapNull(ecsMappingMapElemType()),
	})
	require.False(t, d.HasError())
	return map[string]attr.Value{name: obj}
}

func mustQueriesMap(t *testing.T, elems map[string]attr.Value) types.Map {
	t.Helper()
	queries, d := types.MapValue(queryMapElemType(), elems)
	require.False(t, d.HasError())
	return queries
}

func mustMap(t *testing.T, elems map[string]attr.Value) types.Map {
	t.Helper()
	m, d := types.MapValue(ecsMappingMapElemType(), elems)
	require.False(t, d.HasError())
	return m
}
