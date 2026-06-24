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

package osquerysavedquery

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOsquerySavedQueryModelIdentityMethods(t *testing.T) {
	t.Parallel()

	model := osquerySavedQueryModel{
		ID:           types.StringValue("production/list_processes"),
		SavedQueryID: types.StringValue("list_processes"),
		SpaceID:      types.StringValue("production"),
	}

	assert.Equal(t, "production/list_processes", model.GetID().ValueString())
	assert.Equal(t, "list_processes", model.GetResourceID().ValueString())
	assert.Equal(t, "production", model.GetSpaceID().ValueString())

	reqs, diags := model.GetVersionRequirements(context.Background())
	require.Empty(t, diags)
	require.Len(t, reqs, 1)
	assert.Equal(t, "8.5.0", reqs[0].MinVersion.String())
}

func TestSetCompositeIdentity(t *testing.T) {
	t.Parallel()

	t.Run("uses configured space_id", func(t *testing.T) {
		model := osquerySavedQueryModel{SpaceID: types.StringValue("production")}
		model.setCompositeIdentity("list_processes")

		assert.Equal(t, "production/list_processes", model.ID.ValueString())
		assert.Equal(t, "list_processes", model.SavedQueryID.ValueString())
		assert.Equal(t, "production", model.SpaceID.ValueString())
	})

	t.Run("defaults space_id to default when unset", func(t *testing.T) {
		model := osquerySavedQueryModel{}
		model.setCompositeIdentity("list_processes")

		assert.Equal(t, "default/list_processes", model.ID.ValueString())
		assert.Equal(t, "list_processes", model.SavedQueryID.ValueString())
		assert.Equal(t, "default", model.SpaceID.ValueString())
	})

	t.Run("defaults space_id to default when null", func(t *testing.T) {
		model := osquerySavedQueryModel{SpaceID: types.StringNull()}
		model.setCompositeIdentity("list_processes")

		assert.Equal(t, "default/list_processes", model.ID.ValueString())
		assert.Equal(t, "default", model.SpaceID.ValueString())
	})

	t.Run("preserves unknown space_id while composing default segment", func(t *testing.T) {
		model := osquerySavedQueryModel{SpaceID: types.StringUnknown()}
		model.setCompositeIdentity("list_processes")

		assert.Equal(t, "default/list_processes", model.ID.ValueString())
		assert.True(t, model.SpaceID.IsUnknown())
	})
}

func TestCompositeSpaceID(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "production", compositeSpaceID(types.StringValue("production")))
	assert.Equal(t, "default", compositeSpaceID(types.StringNull()))
	assert.Equal(t, "default", compositeSpaceID(types.StringUnknown()))
	assert.Equal(t, "default", compositeSpaceID(types.StringValue("")))
}

func TestPlatformConversion(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("split and sort from API comma string", func(t *testing.T) {
		platform := kbapi.SecurityOsqueryAPIPlatform("linux,darwin")
		got := platformSetFromAPI(&platform)

		expected := types.SetValueMust(types.StringType, []attr.Value{
			types.StringValue("darwin"),
			types.StringValue("linux"),
		})
		assert.Equal(t, expected, got)
	})

	t.Run("nil platform returns null set", func(t *testing.T) {
		assert.True(t, platformSetFromAPI(nil).IsNull())
	})

	t.Run("empty platform returns null set", func(t *testing.T) {
		platform := kbapi.SecurityOsqueryAPIPlatform("")
		assert.True(t, platformSetFromAPI(&platform).IsNull())
	})

	t.Run("whitespace platform returns null set", func(t *testing.T) {
		platform := kbapi.SecurityOsqueryAPIPlatform("  ,  ")
		assert.True(t, platformSetFromAPI(&platform).IsNull())
	})

	t.Run("join and sort for API write", func(t *testing.T) {
		platform := types.SetValueMust(types.StringType, []attr.Value{
			types.StringValue("linux"),
			types.StringValue("darwin"),
		})

		got, diags := platformToAPI(ctx, platform)
		require.Empty(t, diags)
		require.NotNil(t, got)
		assert.Equal(t, kbapi.SecurityOsqueryAPIPlatform("darwin,linux"), *got)
	})

	t.Run("null platform omits API key", func(t *testing.T) {
		got, diags := platformToAPI(ctx, types.SetNull(types.StringType))
		require.Empty(t, diags)
		assert.Nil(t, got)
	})

	t.Run("unknown platform omits API key", func(t *testing.T) {
		got, diags := platformToAPI(ctx, types.SetUnknown(types.StringType))
		require.Empty(t, diags)
		assert.Nil(t, got)
	})

	t.Run("empty known platform omits API key", func(t *testing.T) {
		got, diags := platformToAPI(ctx, types.SetValueMust(types.StringType, nil))
		require.Empty(t, diags)
		assert.Nil(t, got)
	})
}

func TestEcsMappingConversion(t *testing.T) {
	t.Parallel()

	t.Run("field reference", func(t *testing.T) {
		mapping := ecsMapping{Field: types.StringValue("cmdline")}
		got, diags := mapping.toAPIType()
		require.Empty(t, diags)

		require.NotNil(t, got.Field)
		assert.Equal(t, "cmdline", *got.Field)
		assert.Nil(t, got.Value)
	})

	t.Run("scalar value", func(t *testing.T) {
		mapping := ecsMapping{Value: types.StringValue("process")}
		got, diags := mapping.toAPIType()
		require.Empty(t, diags)

		require.NotNil(t, got.Value)
		scalar, err := got.Value.AsSecurityOsqueryAPIECSMappingItemValue0()
		require.NoError(t, err)
		assert.Equal(t, "process", scalar)
	})

	t.Run("array values", func(t *testing.T) {
		mapping := ecsMapping{
			Values: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("network"),
				types.StringValue("process"),
			}),
		}
		got, diags := mapping.toAPIType()
		require.Empty(t, diags)

		require.NotNil(t, got.Value)
		values, err := got.Value.AsSecurityOsqueryAPIECSMappingItemValue1()
		require.NoError(t, err)
		assert.Equal(t, []string{"network", "process"}, values)
	})

	t.Run("zero arms returns error", func(t *testing.T) {
		_, diags := ecsMapping{}.toAPIType()
		require.True(t, diags.HasError())
	})

	t.Run("multiple arms returns error", func(t *testing.T) {
		mapping := ecsMapping{
			Field: types.StringValue("cmdline"),
			Value: types.StringValue("process"),
		}
		_, diags := mapping.toAPIType()
		require.True(t, diags.HasError())
	})

	t.Run("from API field reference", func(t *testing.T) {
		field := "cmdline"
		got, diags := ecsMappingFromAPIType(kbapi.SecurityOsqueryAPIECSMappingItem{Field: &field})
		require.Empty(t, diags)

		assert.Equal(t, types.StringValue("cmdline"), got.Field)
		assert.True(t, got.Value.IsNull())
		assert.True(t, got.Values.IsNull())
	})

	t.Run("from API scalar value", func(t *testing.T) {
		var value kbapi.SecurityOsqueryAPIECSMappingItem_Value
		require.NoError(t, value.FromSecurityOsqueryAPIECSMappingItemValue0("process"))

		got, diags := ecsMappingFromAPIType(kbapi.SecurityOsqueryAPIECSMappingItem{Value: &value})
		require.Empty(t, diags)
		assert.True(t, got.Field.IsNull())
		assert.Equal(t, types.StringValue("process"), got.Value)
		assert.True(t, got.Values.IsNull())
	})

	t.Run("from API array values", func(t *testing.T) {
		var value kbapi.SecurityOsqueryAPIECSMappingItem_Value
		require.NoError(t, value.FromSecurityOsqueryAPIECSMappingItemValue1([]string{"process", "network"}))

		got, diags := ecsMappingFromAPIType(kbapi.SecurityOsqueryAPIECSMappingItem{Value: &value})
		require.Empty(t, diags)
		assert.True(t, got.Field.IsNull())
		assert.True(t, got.Value.IsNull())
		expected := types.SetValueMust(types.StringType, []attr.Value{
			types.StringValue("network"),
			types.StringValue("process"),
		})
		assert.Equal(t, expected, got.Values)
	})

	t.Run("from API value wins over field", func(t *testing.T) {
		field := "cmdline"
		var value kbapi.SecurityOsqueryAPIECSMappingItem_Value
		require.NoError(t, value.FromSecurityOsqueryAPIECSMappingItemValue0("process"))

		got, diags := ecsMappingFromAPIType(kbapi.SecurityOsqueryAPIECSMappingItem{
			Field: &field,
			Value: &value,
		})
		require.Empty(t, diags)
		assert.True(t, got.Field.IsNull())
		assert.Equal(t, types.StringValue("process"), got.Value)
	})
}

func TestEcsMappingMapFromAPI(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("nil map is null", func(t *testing.T) {
		got, diags := ecsMappingMapFromAPI(ctx, nil)
		require.Empty(t, diags)
		assert.True(t, got.IsNull())
	})

	t.Run("empty map is null", func(t *testing.T) {
		empty := kbapi.SecurityOsqueryAPIECSMapping{}
		got, diags := ecsMappingMapFromAPI(ctx, &empty)
		require.Empty(t, diags)
		assert.True(t, got.IsNull())
	})
}

func TestPopulateFromCreateAPI(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	entity := mustCreateEntity(t, `{
		"data": {
			"id": "list_processes",
			"query": "SELECT * FROM processes",
			"description": "List processes",
			"platform": "linux,darwin",
			"interval": 3600,
			"version": "1.0.0",
			"snapshot": true,
			"removed": false,
			"ecs_mapping": {
				"process.name": { "field": "cmdline" },
				"event.category": { "value": "process" },
				"host.name": { "value": ["web-1", "web-2"] }
			}
		}
	}`)

	model := osquerySavedQueryModel{SpaceID: types.StringValue("default")}
	diags := model.populateFromCreateAPI(ctx, entity)
	require.Empty(t, diags)

	assert.Equal(t, "default/list_processes", model.ID.ValueString())
	assert.Equal(t, "list_processes", model.SavedQueryID.ValueString())
	assert.Equal(t, types.StringValue("SELECT * FROM processes"), model.Query)
	assert.Equal(t, types.StringValue("List processes"), model.Description)
	assert.Equal(t, types.Int64Value(3600), model.Interval)
	assert.Equal(t, types.StringValue("1.0.0"), model.Version)
	assert.Equal(t, types.BoolValue(true), model.Snapshot)
	assert.Equal(t, types.BoolValue(false), model.Removed)

	platform := types.SetValueMust(types.StringType, []attr.Value{
		types.StringValue("darwin"),
		types.StringValue("linux"),
	})
	assert.Equal(t, platform, model.Platform)

	require.False(t, model.EcsMapping.IsNull())
	assertEcsMappingElement(t, model.EcsMapping, "process.name", types.StringValue("cmdline"), types.StringNull(), types.SetNull(types.StringType))
	assertEcsMappingElement(t, model.EcsMapping, "event.category", types.StringNull(), types.StringValue("process"), types.SetNull(types.StringType))
	assertEcsMappingElement(t, model.EcsMapping, "host.name", types.StringNull(), types.StringNull(), types.SetValueMust(types.StringType, []attr.Value{
		types.StringValue("web-1"),
		types.StringValue("web-2"),
	}))
}

func TestPopulateFromGetAPI(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	entity := mustGetEntity(t, `{
		"data": {
			"id": "list_processes",
			"query": "SELECT pid FROM processes",
			"interval": "7200",
			"version": 2
		}
	}`)

	model := osquerySavedQueryModel{SpaceID: types.StringValue("production")}
	diags := model.populateFromGetAPI(ctx, entity)
	require.Empty(t, diags)

	assert.Equal(t, "production/list_processes", model.ID.ValueString())
	assert.Equal(t, types.Int64Value(7200), model.Interval)
	assert.Equal(t, types.StringValue("2"), model.Version)
}

func TestPopulateFromUpdateAPI(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	entity := mustUpdateEntity(t, `{
		"data": {
			"id": "list_processes",
			"query": "SELECT pid FROM processes",
			"interval": 1800,
			"version": "2.1.0",
			"platform": "windows"
		}
	}`)

	model := osquerySavedQueryModel{SavedQueryID: types.StringValue("list_processes")}
	diags := model.populateFromUpdateAPI(ctx, entity)
	require.Empty(t, diags)

	assert.Equal(t, "default/list_processes", model.ID.ValueString())
	assert.Equal(t, types.Int64Value(1800), model.Interval)
	assert.Equal(t, types.StringValue("2.1.0"), model.Version)
	assert.Equal(t, types.SetValueMust(types.StringType, []attr.Value{types.StringValue("windows")}), model.Platform)
}

func TestPopulateFromAPISparseEntity(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	entity := mustGetEntity(t, `{
		"data": {
			"id": "list_processes",
			"query": "SELECT pid FROM processes"
		}
	}`)

	model := osquerySavedQueryModel{SpaceID: types.StringValue("default")}
	diags := model.populateFromGetAPI(ctx, entity)
	require.Empty(t, diags)

	assert.True(t, model.Description.IsNull())
	assert.True(t, model.Platform.IsNull())
	assert.True(t, model.Interval.IsNull())
	assert.True(t, model.Version.IsNull())
	assert.True(t, model.Snapshot.IsNull())
	assert.True(t, model.Removed.IsNull())
	assert.True(t, model.EcsMapping.IsNull())
}

func TestIntervalAndVersionUnionArms(t *testing.T) {
	t.Parallel()

	t.Run("create interval string arm", func(t *testing.T) {
		var interval kbapi.SecurityOsqueryAPICreateSavedQueryResponse_Data_Interval
		require.NoError(t, interval.FromSecurityOsqueryAPICreateSavedQueryResponseDataInterval1("900"))

		got, diags := intervalFromCreateAPI(&interval)
		require.Empty(t, diags)
		assert.Equal(t, types.Int64Value(900), got)
	})

	t.Run("create version int arm", func(t *testing.T) {
		var version kbapi.SecurityOsqueryAPICreateSavedQueryResponse_Data_Version
		require.NoError(t, version.FromSecurityOsqueryAPICreateSavedQueryResponseDataVersion0(4))

		got, diags := versionFromCreateAPI(&version)
		require.Empty(t, diags)
		assert.Equal(t, types.StringValue("4"), got)
	})

	t.Run("get interval int arm", func(t *testing.T) {
		var interval kbapi.SecurityOsqueryAPIFindSavedQueryDetailResponse_Data_Interval
		require.NoError(t, interval.FromSecurityOsqueryAPIFindSavedQueryDetailResponseDataInterval0(7200))

		got, diags := intervalFromGetAPI(&interval)
		require.Empty(t, diags)
		assert.Equal(t, types.Int64Value(7200), got)
	})

	t.Run("get version int arm", func(t *testing.T) {
		var version kbapi.SecurityOsqueryAPIFindSavedQueryDetailResponse_Data_Version
		require.NoError(t, version.FromSecurityOsqueryAPIFindSavedQueryDetailResponseDataVersion0(3))

		got, diags := versionFromGetAPI(&version)
		require.Empty(t, diags)
		assert.Equal(t, types.StringValue("3"), got)
	})

	t.Run("get version string arm", func(t *testing.T) {
		var version kbapi.SecurityOsqueryAPIFindSavedQueryDetailResponse_Data_Version
		require.NoError(t, version.FromSecurityOsqueryAPIFindSavedQueryDetailResponseDataVersion1("3.0.0"))

		got, diags := versionFromGetAPI(&version)
		require.Empty(t, diags)
		assert.Equal(t, types.StringValue("3.0.0"), got)
	})

	t.Run("update interval string arm", func(t *testing.T) {
		var interval kbapi.SecurityOsqueryAPIUpdateSavedQueryResponse_Data_Interval
		require.NoError(t, interval.FromSecurityOsqueryAPIUpdateSavedQueryResponseDataInterval1("1500"))

		got, diags := intervalFromUpdateAPI(&interval)
		require.Empty(t, diags)
		assert.Equal(t, types.Int64Value(1500), got)
	})

	t.Run("update plain string version", func(t *testing.T) {
		version := "2.1.0"
		got := versionFromUpdateAPI(&version)
		assert.Equal(t, types.StringValue("2.1.0"), got)
	})
}

func TestIntervalAndVersionDiagnostics(t *testing.T) {
	t.Parallel()

	t.Run("create interval bad string", func(t *testing.T) {
		var interval kbapi.SecurityOsqueryAPICreateSavedQueryResponse_Data_Interval
		require.NoError(t, interval.FromSecurityOsqueryAPICreateSavedQueryResponseDataInterval1("not-a-number"))

		_, diags := intervalFromCreateAPI(&interval)
		require.True(t, diags.HasError())
	})

	t.Run("get interval invalid union", func(t *testing.T) {
		var interval kbapi.SecurityOsqueryAPIFindSavedQueryDetailResponse_Data_Interval
		require.NoError(t, json.Unmarshal([]byte(`{"unexpected": true}`), &interval))

		_, diags := intervalFromGetAPI(&interval)
		require.True(t, diags.HasError())
	})

	t.Run("create version invalid union", func(t *testing.T) {
		var version kbapi.SecurityOsqueryAPICreateSavedQueryResponse_Data_Version
		require.NoError(t, json.Unmarshal([]byte(`{"unexpected": true}`), &version))

		_, diags := versionFromCreateAPI(&version)
		require.True(t, diags.HasError())
	})
}

func TestPopulateFromAPINilEntity(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	model := osquerySavedQueryModel{}

	require.Empty(t, model.populateFromCreateAPI(ctx, nil))
	require.Empty(t, model.populateFromGetAPI(ctx, nil))
	require.Empty(t, model.populateFromUpdateAPI(ctx, nil))
}

func assertEcsMappingElement(t *testing.T, mapping types.Map, key string, field, value types.String, values types.Set) {
	t.Helper()

	elem, ok := mapping.Elements()[key]
	require.True(t, ok, "missing ecs_mapping key %q", key)

	obj, ok := elem.(types.Object)
	require.True(t, ok)

	attrs := obj.Attributes()
	assert.Equal(t, field, attrs["field"])
	assert.Equal(t, value, attrs["value"])
	assert.Equal(t, values, attrs["values"])
}

func mustCreateEntity(t *testing.T, payload string) *kibanaoapi.OsquerySavedQueryCreateEntity {
	t.Helper()

	var resp kbapi.SecurityOsqueryAPICreateSavedQueryResponse
	require.NoError(t, json.Unmarshal([]byte(payload), &resp))
	require.NotNil(t, resp.Data)

	d := resp.Data
	return &kibanaoapi.OsquerySavedQueryCreateEntity{
		Description: d.Description,
		EcsMapping:  d.EcsMapping,
		ID:          d.Id,
		Interval:    d.Interval,
		Platform:    d.Platform,
		Prebuilt:    d.Prebuilt,
		Query:       d.Query,
		Removed:     d.Removed,
		Snapshot:    d.Snapshot,
		Version:     d.Version,
	}
}

func mustGetEntity(t *testing.T, payload string) *kibanaoapi.OsquerySavedQueryGetEntity {
	t.Helper()

	var resp kbapi.SecurityOsqueryAPIFindSavedQueryDetailResponse
	require.NoError(t, json.Unmarshal([]byte(payload), &resp))
	require.NotNil(t, resp.Data)

	d := resp.Data
	return &kibanaoapi.OsquerySavedQueryGetEntity{
		Description: d.Description,
		EcsMapping:  d.EcsMapping,
		ID:          d.Id,
		Interval:    d.Interval,
		Platform:    d.Platform,
		Prebuilt:    d.Prebuilt,
		Query:       d.Query,
		Removed:     d.Removed,
		Snapshot:    d.Snapshot,
		Version:     d.Version,
	}
}

func mustUpdateEntity(t *testing.T, payload string) *kibanaoapi.OsquerySavedQueryUpdateEntity {
	t.Helper()

	var resp kbapi.SecurityOsqueryAPIUpdateSavedQueryResponse
	require.NoError(t, json.Unmarshal([]byte(payload), &resp))
	require.NotNil(t, resp.Data)

	d := resp.Data
	return &kibanaoapi.OsquerySavedQueryUpdateEntity{
		Description: d.Description,
		EcsMapping:  d.EcsMapping,
		ID:          d.Id,
		Interval:    d.Interval,
		Platform:    d.Platform,
		Prebuilt:    d.Prebuilt,
		Query:       d.Query,
		Removed:     d.Removed,
		Snapshot:    d.Snapshot,
		Version:     d.Version,
	}
}
