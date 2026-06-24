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
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOsqueryPackModel_GetVersionRequirements(t *testing.T) {
	t.Parallel()

	reqs, diags := osqueryPackModel{}.GetVersionRequirements(context.Background())
	require.False(t, diags.HasError())
	require.Len(t, reqs, 1)
	require.Equal(t, *osqueryPackMinVersion, reqs[0].MinVersion)
	require.Contains(t, reqs[0].ErrorMessage, "8.5.0")
}

func TestOsqueryPackModel_ImplementsEntityCoreContracts(t *testing.T) {
	t.Parallel()

	var _ entitycore.KibanaResourceModel = osqueryPackModel{}
	var _ entitycore.WithVersionRequirements = osqueryPackModel{}
}

func TestShardsMapFromAPI(t *testing.T) {
	t.Parallel()

	t.Run("returns null for empty shards", func(t *testing.T) {
		result := shardsMapFromAPI(nil)
		assert.True(t, result.IsNull())
	})

	t.Run("normalizes float64 shard map", func(t *testing.T) {
		result := shardsMapFromAPI(kibanaoapi.OsqueryPackShards{
			"policy-a": 50,
			"policy-b": 100,
		})

		require.False(t, result.IsNull())
		assert.Equal(t, types.Float64Value(50), result.Elements()["policy-a"])
		assert.Equal(t, types.Float64Value(100), result.Elements()["policy-b"])
	})
}

func TestPlatformCommaStringSetRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("comma string to sorted set", func(t *testing.T) {
		platform := kbapi.SecurityOsqueryAPIPlatform("darwin,linux")
		set := platformSetFromAPI(&platform)

		require.False(t, set.IsNull())
		var platforms []string
		diags := set.ElementsAs(ctx, &platforms, false)
		require.False(t, diags.HasError())
		assert.Equal(t, []string{"darwin", "linux"}, platforms)
	})

	t.Run("set to comma string", func(t *testing.T) {
		set, d := types.SetValueFrom(ctx, types.StringType, []string{"windows", "linux"})
		require.False(t, d.HasError())

		platform, diags := platformCommaStringFromSet(ctx, set)
		require.False(t, diags.HasError())
		require.NotNil(t, platform)
		assert.Equal(t, kbapi.SecurityOsqueryAPIPlatform("linux,windows"), *platform)
	})

	t.Run("round trip", func(t *testing.T) {
		initial := kbapi.SecurityOsqueryAPIPlatform("linux,darwin")
		set := platformSetFromAPI(&initial)

		platform, diags := platformCommaStringFromSet(ctx, set)
		require.False(t, diags.HasError())
		assert.Equal(t, kbapi.SecurityOsqueryAPIPlatform("darwin,linux"), *platform)
	})

	t.Run("null platform stays null", func(t *testing.T) {
		assert.True(t, platformSetFromAPI(nil).IsNull())

		platform, diags := platformCommaStringFromSet(ctx, types.SetNull(types.StringType))
		require.False(t, diags.HasError())
		assert.Nil(t, platform)
	})
}

func TestEcsMappingThreeShapes(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("field reference", func(t *testing.T) {
		api := kbapi.SecurityOsqueryAPIECSMapping{
			"process.name": {Field: strPtr("cmdline")},
		}

		mapping, diags := ecsMappingMapFromAPI(ctx, &api)
		require.False(t, diags.HasError())

		var m ecsMappingModel
		asDiags := mapping.Elements()["process.name"].(basetypes.ObjectValue).As(ctx, &m, basetypes.ObjectAsOptions{})
		require.False(t, asDiags.HasError())
		assert.Equal(t, "cmdline", m.Field.ValueString())
		assert.True(t, m.Value.IsNull())
		assert.True(t, m.Values.IsNull())

		roundTrip, diags := ecsMappingMapToAPI(ctx, mapping)
		require.False(t, diags.HasError())
		require.NotNil(t, roundTrip)
		assert.Equal(t, strPtr("cmdline"), (*roundTrip)["process.name"].Field)
	})

	t.Run("scalar value", func(t *testing.T) {
		var val kbapi.SecurityOsqueryAPIECSMappingItem_Value
		require.NoError(t, val.FromSecurityOsqueryAPIECSMappingItemValue0("literal"))

		api := kbapi.SecurityOsqueryAPIECSMapping{
			"host.name": {Value: &val},
		}

		mapping, diags := ecsMappingMapFromAPI(ctx, &api)
		require.False(t, diags.HasError())

		var m ecsMappingModel
		asDiags := mapping.Elements()["host.name"].(basetypes.ObjectValue).As(ctx, &m, basetypes.ObjectAsOptions{})
		require.False(t, asDiags.HasError())
		assert.True(t, m.Field.IsNull())
		assert.Equal(t, "literal", m.Value.ValueString())
		assert.True(t, m.Values.IsNull())

		roundTrip, diags := ecsMappingMapToAPI(ctx, mapping)
		require.False(t, diags.HasError())
		require.NotNil(t, roundTrip)

		got, err := (*roundTrip)["host.name"].Value.AsSecurityOsqueryAPIECSMappingItemValue0()
		require.NoError(t, err)
		assert.Equal(t, "literal", got)
	})

	t.Run("array values", func(t *testing.T) {
		var val kbapi.SecurityOsqueryAPIECSMappingItem_Value
		require.NoError(t, val.FromSecurityOsqueryAPIECSMappingItemValue1([]string{"a", "b"}))

		api := kbapi.SecurityOsqueryAPIECSMapping{
			"tags": {Value: &val},
		}

		mapping, diags := ecsMappingMapFromAPI(ctx, &api)
		require.False(t, diags.HasError())

		var m ecsMappingModel
		asDiags := mapping.Elements()["tags"].(basetypes.ObjectValue).As(ctx, &m, basetypes.ObjectAsOptions{})
		require.False(t, asDiags.HasError())
		assert.True(t, m.Field.IsNull())
		assert.True(t, m.Value.IsNull())

		var values []string
		valueDiags := m.Values.ElementsAs(ctx, &values, false)
		require.False(t, valueDiags.HasError())
		assert.Equal(t, []string{"a", "b"}, values)

		roundTrip, diags := ecsMappingMapToAPI(ctx, mapping)
		require.False(t, diags.HasError())
		require.NotNil(t, roundTrip)

		got, err := (*roundTrip)["tags"].Value.AsSecurityOsqueryAPIECSMappingItemValue1()
		require.NoError(t, err)
		assert.Equal(t, []string{"a", "b"}, got)
	})
}

func TestQueryModelVersionRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	version := kbapi.SecurityOsqueryAPIVersion("5.0.0")
	apiItem := kbapi.SecurityOsqueryAPIObjectQueriesItem{
		Query:   queryPtr("SELECT 1"),
		Version: &version,
	}

	var q queryModel
	require.False(t, q.fromAPIType(ctx, apiItem).HasError())
	assert.Equal(t, "5.0.0", q.Version.ValueString())

	roundTrip, diags := q.toAPIType(ctx)
	require.False(t, diags.HasError())
	require.NotNil(t, roundTrip.Version)
	assert.Equal(t, kbapi.SecurityOsqueryAPIVersion("5.0.0"), *roundTrip.Version)
}

func TestPopulateFromAPI(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	desc := kbapi.SecurityOsqueryAPIPackDescription("pack description")
	enabled := kbapi.SecurityOsqueryAPIEnabled(true)
	policyIDs := kbapi.SecurityOsqueryAPIPolicyIds{"policy-1", "policy-2"}
	platform := kbapi.SecurityOsqueryAPIPlatform("linux,darwin")
	version := kbapi.SecurityOsqueryAPIVersion("5.10.2")
	snapshot := kbapi.SecurityOsqueryAPISnapshot(true)
	removed := kbapi.SecurityOsqueryAPIRemoved(false)
	savedQueryID := kbapi.SecurityOsqueryAPISavedQueryId("saved-query-id")

	var fieldVal kbapi.SecurityOsqueryAPIECSMappingItem_Value
	require.NoError(t, fieldVal.FromSecurityOsqueryAPIECSMappingItemValue0("static"))

	queries := kbapi.SecurityOsqueryAPIObjectQueries{
		"find_procs": {
			Query:        queryPtr("SELECT * FROM processes"),
			Platform:     &platform,
			Version:      &version,
			Snapshot:     &snapshot,
			Removed:      &removed,
			SavedQueryId: &savedQueryID,
			EcsMapping: &kbapi.SecurityOsqueryAPIECSMapping{
				"process.name": {Field: strPtr("cmdline")},
				"host.name":    {Value: &fieldVal},
			},
		},
	}

	detail := &kibanaoapi.OsqueryPackDetail{
		Name:          "test-pack",
		Description:   &desc,
		Enabled:       &enabled,
		PolicyIds:     &policyIDs,
		SavedObjectId: "pack-uuid-123",
		Shards: kibanaoapi.OsqueryPackShards{
			"policy-1": 75,
		},
		Queries: &queries,
	}

	var model osqueryPackModel
	diags := model.populateFromAPI(ctx, "production", detail)
	require.False(t, diags.HasError())

	assert.Equal(t, "production/pack-uuid-123", model.ID.ValueString())
	assert.Equal(t, "pack-uuid-123", model.PackID.ValueString())
	assert.Equal(t, "production", model.SpaceID.ValueString())
	assert.Equal(t, "test-pack", model.Name.ValueString())
	assert.Equal(t, "pack description", model.Description.ValueString())
	assert.True(t, model.Enabled.ValueBool())

	var policies []string
	policyDiags := model.PolicyIDs.ElementsAs(ctx, &policies, false)
	require.False(t, policyDiags.HasError())
	assert.Equal(t, []string{"policy-1", "policy-2"}, policies)

	assert.Equal(t, types.Float64Value(75), model.Shards.Elements()["policy-1"])

	queryObj, ok := model.Queries.Elements()["find_procs"].(basetypes.ObjectValue)
	require.True(t, ok)

	var q queryModel
	queryDiags := queryObj.As(ctx, &q, basetypes.ObjectAsOptions{})
	require.False(t, queryDiags.HasError())
	assert.Equal(t, "SELECT * FROM processes", q.Query.ValueString())
	assert.Equal(t, "5.10.2", q.Version.ValueString())
	assert.True(t, q.Snapshot.ValueBool())
	assert.False(t, q.Removed.ValueBool())
	assert.Equal(t, "saved-query-id", q.SavedQueryID.ValueString())

	var platforms []string
	platformDiags := q.Platform.ElementsAs(ctx, &platforms, false)
	require.False(t, platformDiags.HasError())
	assert.Equal(t, []string{"darwin", "linux"}, platforms)

	require.Len(t, q.EcsMapping.Elements(), 2)

	var fieldMapping ecsMappingModel
	fieldDiags := q.EcsMapping.Elements()["process.name"].(basetypes.ObjectValue).As(ctx, &fieldMapping, basetypes.ObjectAsOptions{})
	require.False(t, fieldDiags.HasError())
	assert.Equal(t, "cmdline", fieldMapping.Field.ValueString())

	var valueMapping ecsMappingModel
	valueDiags := q.EcsMapping.Elements()["host.name"].(basetypes.ObjectValue).As(ctx, &valueMapping, basetypes.ObjectAsOptions{})
	require.False(t, valueDiags.HasError())
	assert.Equal(t, "static", valueMapping.Value.ValueString())
}

func TestPopulateFromAPI_DefaultSpaceAndEmptyOptionals(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	detail := &kibanaoapi.OsqueryPackDetail{
		Name:          "minimal-pack",
		SavedObjectId: "minimal-id",
		Queries: &kbapi.SecurityOsqueryAPIObjectQueries{
			"q1": {Query: queryPtr("SELECT 1")},
		},
	}

	var model osqueryPackModel
	diags := model.populateFromAPI(ctx, "", detail)
	require.False(t, diags.HasError())

	assert.Equal(t, "default/minimal-id", model.ID.ValueString())
	assert.Equal(t, "default", model.SpaceID.ValueString())
	assert.True(t, model.Description.IsNull())
	assert.True(t, model.Enabled.IsNull())
	assert.True(t, model.PolicyIDs.IsNull())
	assert.True(t, model.Shards.IsNull())
}

func strPtr(s string) *string { return &s }

func queryPtr(q string) *kbapi.SecurityOsqueryAPIQuery {
	v := kbapi.SecurityOsqueryAPIQuery(q)
	return &v
}
