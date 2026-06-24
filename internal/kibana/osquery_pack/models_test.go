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
	"encoding/json"
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
		set := platformSetFromAPI(ctx, &platform)

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
		set := platformSetFromAPI(ctx, &initial)

		platform, diags := platformCommaStringFromSet(ctx, set)
		require.False(t, diags.HasError())
		assert.Equal(t, kbapi.SecurityOsqueryAPIPlatform("darwin,linux"), *platform)
	})

	t.Run("null platform stays null", func(t *testing.T) {
		assert.True(t, platformSetFromAPI(ctx, nil).IsNull())

		empty := kbapi.SecurityOsqueryAPIPlatform("")
		assert.True(t, platformSetFromAPI(ctx, &empty).IsNull())

		platform, diags := platformCommaStringFromSet(ctx, types.SetNull(types.StringType))
		require.False(t, diags.HasError())
		assert.Nil(t, platform)
	})

	t.Run("trims whitespace after comma", func(t *testing.T) {
		platform := kbapi.SecurityOsqueryAPIPlatform("linux, darwin ,windows")
		set := platformSetFromAPI(ctx, &platform)

		var platforms []string
		diags := set.ElementsAs(ctx, &platforms, false)
		require.False(t, diags.HasError())
		assert.Equal(t, []string{"darwin", "linux", "windows"}, platforms)
	})
}

func TestEcsMappingThreeShapes(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("field reference", func(t *testing.T) {
		api := kbapi.SecurityOsqueryAPIECSMapping{
			"process.name": {Field: new("cmdline")},
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
		assert.Equal(t, new("cmdline"), (*roundTrip)["process.name"].Field)
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

	t.Run("invalid value type returns error", func(t *testing.T) {
		var item kbapi.SecurityOsqueryAPIECSMappingItem
		require.NoError(t, json.Unmarshal([]byte(`{"value": 123}`), &item))

		api := kbapi.SecurityOsqueryAPIECSMapping{"bad.key": item}

		_, diags := ecsMappingMapFromAPI(ctx, &api)
		require.True(t, diags.HasError())
		assert.Contains(t, diags[0].Detail(), `ecs_mapping["bad.key"]`)
	})

	t.Run("field and value both set returns error", func(t *testing.T) {
		var val kbapi.SecurityOsqueryAPIECSMappingItem_Value
		require.NoError(t, val.FromSecurityOsqueryAPIECSMappingItemValue0("literal"))

		api := kbapi.SecurityOsqueryAPIECSMapping{
			"conflict": {Field: new("col"), Value: &val},
		}

		_, diags := ecsMappingMapFromAPI(ctx, &api)
		require.True(t, diags.HasError())
		assert.Contains(t, diags[0].Detail(), `ecs_mapping["conflict"]`)
		assert.Contains(t, diags[0].Detail(), "both field and value")
	})
}

func TestQueryModelFullRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	platform := kbapi.SecurityOsqueryAPIPlatform("windows,linux")
	version := kbapi.SecurityOsqueryAPIVersion("5.0.0")
	snapshot := kbapi.SecurityOsqueryAPISnapshot(true)
	removed := kbapi.SecurityOsqueryAPIRemoved(false)
	savedQueryID := kbapi.SecurityOsqueryAPISavedQueryId("my-saved-query")

	var scalarVal kbapi.SecurityOsqueryAPIECSMappingItem_Value
	require.NoError(t, scalarVal.FromSecurityOsqueryAPIECSMappingItemValue0("literal"))

	apiItem := kbapi.SecurityOsqueryAPIObjectQueriesItem{
		Query:        new(kbapi.SecurityOsqueryAPIQuery("SELECT * FROM users")),
		Platform:     &platform,
		Version:      &version,
		Snapshot:     &snapshot,
		Removed:      &removed,
		SavedQueryId: &savedQueryID,
		EcsMapping: &kbapi.SecurityOsqueryAPIECSMapping{
			"user.name": {Field: new("username")},
			"host.name": {Value: &scalarVal},
		},
	}

	var q queryModel
	require.False(t, q.fromAPIType(ctx, apiItem).HasError())

	roundTrip, diags := q.toAPIType(ctx)
	require.False(t, diags.HasError())

	require.NotNil(t, roundTrip.Query)
	assert.Equal(t, kbapi.SecurityOsqueryAPIQuery("SELECT * FROM users"), *roundTrip.Query)
	require.NotNil(t, roundTrip.Platform)
	assert.Equal(t, kbapi.SecurityOsqueryAPIPlatform("linux,windows"), *roundTrip.Platform)
	require.NotNil(t, roundTrip.Version)
	assert.Equal(t, kbapi.SecurityOsqueryAPIVersion("5.0.0"), *roundTrip.Version)
	require.NotNil(t, roundTrip.Snapshot)
	assert.True(t, *roundTrip.Snapshot)
	require.NotNil(t, roundTrip.Removed)
	assert.False(t, *roundTrip.Removed)
	require.NotNil(t, roundTrip.SavedQueryId)
	assert.Equal(t, kbapi.SecurityOsqueryAPISavedQueryId("my-saved-query"), *roundTrip.SavedQueryId)

	require.NotNil(t, roundTrip.EcsMapping)
	assert.Equal(t, new("username"), (*roundTrip.EcsMapping)["user.name"].Field)
	got, err := (*roundTrip.EcsMapping)["host.name"].Value.AsSecurityOsqueryAPIECSMappingItemValue0()
	require.NoError(t, err)
	assert.Equal(t, "literal", got)
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
			Query:        new(kbapi.SecurityOsqueryAPIQuery("SELECT * FROM processes")),
			Platform:     &platform,
			Version:      &version,
			Snapshot:     &snapshot,
			Removed:      &removed,
			SavedQueryId: &savedQueryID,
			EcsMapping: &kbapi.SecurityOsqueryAPIECSMapping{
				"process.name": {Field: new("cmdline")},
				"host.name":    {Value: &fieldVal},
			},
		},
	}

	detail := &kibanaoapi.OsqueryPackDetail{
		Name:          "test-pack",
		Description:   &desc,
		Enabled:       &enabled,
		PolicyIDs:     &policyIDs,
		SavedObjectID: "pack-uuid-123",
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
		SavedObjectID: "minimal-id",
		Queries: &kbapi.SecurityOsqueryAPIObjectQueries{
			"q1": {Query: new(kbapi.SecurityOsqueryAPIQuery("SELECT 1"))},
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

func TestPopulateFromAPI_NilDetail(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	var model osqueryPackModel
	model.ID = types.StringValue("prior/id")
	diags := model.populateFromAPI(ctx, "default", nil)
	require.False(t, diags.HasError())
	assert.Equal(t, "prior/id", model.ID.ValueString())
}

func TestPopulateFromAPI_NilAndEmptyQueries(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("nil queries", func(t *testing.T) {
		detail := &kibanaoapi.OsqueryPackDetail{
			Name:          "no-queries",
			SavedObjectID: "no-queries-id",
		}

		var model osqueryPackModel
		diags := model.populateFromAPI(ctx, "default", detail)
		require.False(t, diags.HasError())
		assert.True(t, model.Queries.IsNull())
	})

	t.Run("empty queries map", func(t *testing.T) {
		empty := kbapi.SecurityOsqueryAPIObjectQueries{}
		detail := &kibanaoapi.OsqueryPackDetail{
			Name:          "empty-queries",
			SavedObjectID: "empty-queries-id",
			Queries:       &empty,
		}

		var model osqueryPackModel
		diags := model.populateFromAPI(ctx, "default", detail)
		require.False(t, diags.HasError())
		assert.True(t, model.Queries.IsNull())
	})
}

func TestPopulateFromAPI_EmptyPolicyIDs(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	emptyPolicies := kbapi.SecurityOsqueryAPIPolicyIds{}
	detail := &kibanaoapi.OsqueryPackDetail{
		Name:          "empty-policies",
		SavedObjectID: "empty-policies-id",
		PolicyIDs:     &emptyPolicies,
		Queries: &kbapi.SecurityOsqueryAPIObjectQueries{
			"q1": {Query: new(kbapi.SecurityOsqueryAPIQuery("SELECT 1"))},
		},
	}

	var model osqueryPackModel
	diags := model.populateFromAPI(ctx, "default", detail)
	require.False(t, diags.HasError())
	assert.True(t, model.PolicyIDs.IsNull())
}

func TestPopulateFromAPI_QueryRequiredOnlyNullOptionals(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	detail := &kibanaoapi.OsqueryPackDetail{
		Name:          "query-only",
		SavedObjectID: "query-only-id",
		Queries: &kbapi.SecurityOsqueryAPIObjectQueries{
			"q1": {Query: new(kbapi.SecurityOsqueryAPIQuery("SELECT 1"))},
		},
	}

	var model osqueryPackModel
	diags := model.populateFromAPI(ctx, "default", detail)
	require.False(t, diags.HasError())

	queryObj := model.Queries.Elements()["q1"].(basetypes.ObjectValue)
	var q queryModel
	require.False(t, queryObj.As(ctx, &q, basetypes.ObjectAsOptions{}).HasError())
	assert.Equal(t, "SELECT 1", q.Query.ValueString())
	assert.True(t, q.Platform.IsNull())
	assert.True(t, q.Version.IsNull())
	assert.True(t, q.Snapshot.IsNull())
	assert.True(t, q.Removed.IsNull())
	assert.True(t, q.SavedQueryID.IsNull())
	assert.True(t, q.EcsMapping.IsNull())
}
