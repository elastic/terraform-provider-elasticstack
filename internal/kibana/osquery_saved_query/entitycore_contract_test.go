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
	"reflect"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/providerfwtest"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOsquerySavedQueryModel_satisfiesKibanaResourceModel(t *testing.T) {
	t.Parallel()
	var _ entitycore.KibanaResourceModel = osquerySavedQueryModel{}
}

func TestOsquerySavedQueryModel_satisfiesWithVersionRequirements(t *testing.T) {
	t.Parallel()
	var _ entitycore.WithVersionRequirements = osquerySavedQueryModel{}
}

func TestNewResource_satisfiesFrameworkInterfaces(t *testing.T) {
	t.Parallel()
	var _ resource.Resource = newResource()
	var _ resource.ResourceWithConfigure = newResource()
	var _ resource.ResourceWithImportState = newResource()
}

func TestResource_embedsEntityCoreKibanaResource(t *testing.T) {
	t.Parallel()

	rt := reflect.TypeFor[Resource]()
	field, ok := rt.FieldByName("KibanaResource")
	require.True(t, ok)
	require.True(t, field.Anonymous)
	require.Equal(t, reflect.TypeFor[*entitycore.KibanaResource[osquerySavedQueryModel]](), field.Type)
}

func TestResource_importState_seedsCompositeIdentity(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	r, ok := any(newResource()).(resource.ResourceWithImportState)
	require.True(t, ok)

	st := providerfwtest.EmptyImportState(t, r)
	resp := &resource.ImportStateResponse{State: st}

	const importID = "production/list_all_processes"
	r.ImportState(ctx, resource.ImportStateRequest{ID: importID}, resp)
	require.False(t, resp.Diagnostics.HasError())

	var id, savedQueryID, spaceID types.String
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("id"), &id)...)
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("saved_query_id"), &savedQueryID)...)
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("space_id"), &spaceID)...)
	require.False(t, resp.Diagnostics.HasError())

	assert.Equal(t, importID, id.ValueString())
	assert.Equal(t, "list_all_processes", savedQueryID.ValueString())
	assert.Equal(t, "production", spaceID.ValueString())
}

func TestResource_importState_rejectsEmptySpaceSegment(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	r, ok := any(newResource()).(resource.ResourceWithImportState)
	require.True(t, ok)

	st := providerfwtest.EmptyImportState(t, r)
	resp := &resource.ImportStateResponse{State: st}

	r.ImportState(ctx, resource.ImportStateRequest{ID: "/list_all_processes"}, resp)
	require.True(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "<space_id>/<saved_query_id>")
}

func TestPrebuiltGuardDiagnostic(t *testing.T) {
	t.Parallel()

	t.Run("nil prebuilt is allowed", func(t *testing.T) {
		diags := prebuiltGuardDiagnostic(nil)
		assert.Empty(t, diags)
	})

	t.Run("false prebuilt is allowed", func(t *testing.T) {
		prebuilt := false
		diags := prebuiltGuardDiagnostic(&prebuilt)
		assert.Empty(t, diags)
	})

	t.Run("true prebuilt returns error", func(t *testing.T) {
		prebuilt := true
		diags := prebuiltGuardDiagnostic(&prebuilt)
		require.True(t, diags.HasError())
		assert.Equal(t, prebuiltSavedQueryDiagnosticSummary, diags.Errors()[0].Summary())
		assert.Equal(t, prebuiltSavedQueryDiagnosticDetail, diags.Errors()[0].Detail())
	})
}

func TestToAPICreateRequest_minimalOmitsOptionalFields(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	model := osquerySavedQueryModel{
		SavedQueryID: types.StringValue("list_processes"),
		Query:        types.StringValue("SELECT 1"),
	}

	body, diags := model.toAPICreateRequest(ctx)
	require.Empty(t, diags)

	require.NotNil(t, body.Id)
	assert.Equal(t, "list_processes", *body.Id)
	require.NotNil(t, body.Query)
	assert.Equal(t, "SELECT 1", *body.Query)
	assert.Nil(t, body.Description)
	assert.Nil(t, body.Platform)
	assert.Nil(t, body.Interval)
	assert.Nil(t, body.Version)
	assert.Nil(t, body.Snapshot)
	assert.Nil(t, body.Removed)
	assert.Nil(t, body.EcsMapping)
}

func TestToAPICreateRequest_omitsNullOptionalFields(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	model := osquerySavedQueryModel{
		SavedQueryID: types.StringValue("list_processes"),
		Query:        types.StringValue("SELECT 1"),
		Description:  types.StringNull(),
		Platform:     types.SetNull(types.StringType),
		Interval:     types.Int64Null(),
		Version:      types.StringNull(),
		Snapshot:     types.BoolNull(),
		Removed:      types.BoolNull(),
		EcsMapping:   types.MapNull(getEcsMappingElemType()),
	}

	body, diags := model.toAPICreateRequest(ctx)
	require.Empty(t, diags)

	require.NotNil(t, body.Id)
	require.NotNil(t, body.Query)
	assert.Nil(t, body.Description)
	assert.Nil(t, body.Platform)
	assert.Nil(t, body.Interval)
	assert.Nil(t, body.Version)
	assert.Nil(t, body.Snapshot)
	assert.Nil(t, body.Removed)
	assert.Nil(t, body.EcsMapping)
}

func TestToAPICreateRequest(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ecsObj, objDiags := types.ObjectValue(ecsMappingAttrTypes, map[string]attr.Value{
		attrEcsMappingField:  types.StringValue("cmdline"),
		attrEcsMappingValue:  types.StringNull(),
		attrEcsMappingValues: types.SetNull(types.StringType),
	})
	require.Empty(t, objDiags)

	ecsMap, mapDiags := types.MapValue(getEcsMappingElemType(), map[string]attr.Value{
		"process.name": ecsObj,
	})
	require.Empty(t, mapDiags)

	model := osquerySavedQueryModel{
		SavedQueryID: types.StringValue("list_processes"),
		Query:        types.StringValue("SELECT * FROM processes"),
		Description:  types.StringValue("List processes"),
		Platform:     stringSetValue([]string{"linux", "darwin"}),
		Interval:     types.Int64Value(3600),
		Version:      types.StringValue("5.0.0"),
		Snapshot:     types.BoolValue(true),
		Removed:      types.BoolValue(false),
		EcsMapping:   ecsMap,
	}

	body, diags := model.toAPICreateRequest(ctx)
	require.Empty(t, diags)

	require.NotNil(t, body.Id)
	assert.Equal(t, "list_processes", *body.Id)
	require.NotNil(t, body.Query)
	assert.Equal(t, "SELECT * FROM processes", *body.Query)
	require.NotNil(t, body.Description)
	assert.Equal(t, "List processes", *body.Description)
	require.NotNil(t, body.Platform)
	assert.Equal(t, "darwin,linux", *body.Platform)
	require.NotNil(t, body.Interval)
	assert.Equal(t, "3600", *body.Interval)
	require.NotNil(t, body.Version)
	assert.Equal(t, "5.0.0", *body.Version)
	require.NotNil(t, body.Snapshot)
	assert.True(t, *body.Snapshot)
	require.NotNil(t, body.Removed)
	assert.False(t, *body.Removed)
	require.NotNil(t, body.EcsMapping)
	require.Contains(t, *body.EcsMapping, "process.name")
}

func TestToAPIUpdateRequest_omitsUnsetOptionalFields(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	model := osquerySavedQueryModel{
		Query: types.StringValue("SELECT 1"),
	}

	body, diags := model.toAPIUpdateRequest(ctx)
	require.Empty(t, diags)

	require.NotNil(t, body.Query)
	assert.Equal(t, "SELECT 1", *body.Query)
	assert.Nil(t, body.Description)
	assert.Nil(t, body.Platform)
	assert.Nil(t, body.Interval)
	assert.Nil(t, body.Version)
	assert.Nil(t, body.Snapshot)
	assert.Nil(t, body.Removed)
	assert.Nil(t, body.EcsMapping)
	assert.Nil(t, body.Id)
}
