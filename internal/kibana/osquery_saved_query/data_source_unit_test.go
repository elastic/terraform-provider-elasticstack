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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ entitycore.WithVersionRequirements = (*dataSourceModel)(nil)

func TestNewDataSource_implementsDataSource(t *testing.T) {
	t.Parallel()

	ds := NewDataSource()
	require.NotNil(t, ds)
	require.Implements(t, (*datasource.DataSource)(nil), ds)
	require.Implements(t, (*datasource.DataSourceWithConfigure)(nil), ds)
}

func TestNewDataSource_metadata(t *testing.T) {
	t.Parallel()

	ds := NewDataSource()

	var resp datasource.MetadataResponse
	ds.Metadata(context.Background(), datasource.MetadataRequest{
		ProviderTypeName: "elasticstack",
	}, &resp)

	require.Equal(t, "elasticstack_kibana_osquery_saved_query", resp.TypeName)
}

func TestNewDataSource_schemaAttributes(t *testing.T) {
	t.Parallel()

	ds := NewDataSource()

	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	require.False(t, resp.Diagnostics.HasError(), "Schema must not produce diagnostics: %v", resp.Diagnostics)

	require.Contains(t, resp.Schema.Blocks, "kibana_connection")

	wantAttrs := []string{
		attrID,
		attrSavedObjectID,
		attrSavedQueryID,
		attrSpaceID,
		attrQuery,
		attrDescription,
		attrPlatform,
		attrInterval,
		attrVersion,
		attrSnapshot,
		attrRemoved,
		attrEcsMapping,
		attrPrebuilt,
	}
	for _, attr := range wantAttrs {
		require.Contains(t, resp.Schema.Attributes, attr, "schema Attributes must contain %q", attr)
	}
}

func TestDataSourceSchema_attributeMetadata(t *testing.T) {
	t.Parallel()

	s := getDataSourceSchema(context.Background())

	savedQueryIDAttr, ok := s.Attributes[attrSavedQueryID].(dsschema.StringAttribute)
	require.True(t, ok)
	assert.True(t, savedQueryIDAttr.IsRequired())
	assert.False(t, savedQueryIDAttr.IsComputed())
	assert.False(t, savedQueryIDAttr.IsOptional())

	spaceIDAttr, ok := s.Attributes[attrSpaceID].(dsschema.StringAttribute)
	require.True(t, ok)
	assert.True(t, spaceIDAttr.IsOptional())
	assert.True(t, spaceIDAttr.IsComputed())
	assert.False(t, spaceIDAttr.IsRequired())

	computedOnlyStringAttrs := []string{attrID, attrSavedObjectID, attrQuery, attrDescription, attrVersion}
	for _, name := range computedOnlyStringAttrs {
		attr, ok := s.Attributes[name].(dsschema.StringAttribute)
		require.True(t, ok, "expected %q to be StringAttribute", name)
		assert.True(t, attr.IsComputed(), "%q must be computed", name)
		assert.False(t, attr.IsRequired(), "%q must not be required", name)
		assert.False(t, attr.IsOptional(), "%q must not be optional", name)
	}

	platformAttr, ok := s.Attributes[attrPlatform].(dsschema.SetAttribute)
	require.True(t, ok)
	assert.True(t, platformAttr.IsComputed())
	assert.False(t, platformAttr.IsRequired())
	assert.False(t, platformAttr.IsOptional())

	intervalAttr, ok := s.Attributes[attrInterval].(dsschema.Int64Attribute)
	require.True(t, ok)
	assert.True(t, intervalAttr.IsComputed())
	assert.False(t, intervalAttr.IsRequired())
	assert.False(t, intervalAttr.IsOptional())

	computedOnlyBoolAttrs := []string{attrSnapshot, attrRemoved, attrPrebuilt}
	for _, name := range computedOnlyBoolAttrs {
		attr, ok := s.Attributes[name].(dsschema.BoolAttribute)
		require.True(t, ok, "expected %q to be BoolAttribute", name)
		assert.True(t, attr.IsComputed(), "%q must be computed", name)
		assert.False(t, attr.IsRequired(), "%q must not be required", name)
		assert.False(t, attr.IsOptional(), "%q must not be optional", name)
	}

	ecsMappingAttr, ok := s.Attributes[attrEcsMapping].(dsschema.MapNestedAttribute)
	require.True(t, ok)
	assert.True(t, ecsMappingAttr.IsComputed())
	assert.False(t, ecsMappingAttr.IsRequired())
	assert.False(t, ecsMappingAttr.IsOptional())
}

func TestEcsMappingDataSourceElemType_matchesSchema(t *testing.T) {
	t.Parallel()

	ecsMappingAttr, ok := getDataSourceSchema(context.Background()).Attributes[attrEcsMapping].(dsschema.MapNestedAttribute)
	require.True(t, ok, "expected ecs_mapping to be MapNestedAttribute")

	schemaElem := dataSourceSchemaNestedObjectElemType(ecsMappingAttr.NestedObject)
	require.Equal(t, schemaElem, getEcsMappingElemType(),
		"getEcsMappingElemType() drifted from data source ecs_mapping nested object; update both together")
}

func TestResolveDataSourceSpaceID(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "production", resolveDataSourceSpaceID(types.StringValue("production")))
	assert.Equal(t, clients.DefaultSpaceID, resolveDataSourceSpaceID(types.StringNull()))
	assert.Equal(t, clients.DefaultSpaceID, resolveDataSourceSpaceID(types.StringValue("")))
}

func TestDataSourceModel_GetVersionRequirements(t *testing.T) {
	t.Parallel()

	var m dataSourceModel
	reqs, diags := m.GetVersionRequirements(context.Background())
	require.False(t, diags.HasError())
	require.Len(t, reqs, 1)
	require.Equal(t, "8.5.0", reqs[0].MinVersion.String())
	require.NotEmpty(t, reqs[0].ErrorMessage)
}

func TestFinishOsquerySavedQueryDataSourceRead_notFound(t *testing.T) {
	t.Parallel()

	config := dataSourceModel{SavedQueryID: types.StringValue("missing-query")}
	_, diags := finishOsquerySavedQueryDataSourceRead(context.Background(), config, nil, "default")
	require.True(t, diags.HasError())
	assert.Equal(t, "Osquery saved query not found", diags.Errors()[0].Summary())
	assert.Contains(t, diags.Errors()[0].Detail(), "missing-query")
	assert.Contains(t, diags.Errors()[0].Detail(), "default")
}

func TestFinishOsquerySavedQueryDataSourceRead_successWithDefaultSpace(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	prebuilt := false
	query := kbapi.SecurityOsqueryAPIQuery("SELECT pid FROM processes")
	description := kbapi.SecurityOsqueryAPISavedQueryDescription("List processes")
	platform := kbapi.SecurityOsqueryAPIPlatform("linux,darwin")
	snapshot := true
	removed := false
	entity := &kibanaoapi.OsquerySavedQueryGetEntity{
		ID:            "list_processes",
		SavedObjectID: "saved-object-123",
		Query:         &query,
		Description:   &description,
		Platform:      &platform,
		Prebuilt:      &prebuilt,
		Snapshot:      &snapshot,
		Removed:       &removed,
	}

	config := dataSourceModel{
		SavedQueryID: types.StringValue("list_processes"),
		SpaceID:      types.StringNull(),
	}

	result, diags := finishOsquerySavedQueryDataSourceRead(ctx, config, entity, clients.DefaultSpaceID)
	require.False(t, diags.HasError())
	assert.Equal(t, clients.DefaultSpaceID, result.SpaceID.ValueString())
	assert.Equal(t, "default/list_processes", result.ID.ValueString())
	assert.Equal(t, "saved-object-123", result.SavedObjectID.ValueString())
	assert.Equal(t, "list_processes", result.SavedQueryID.ValueString())
	assert.Equal(t, "SELECT pid FROM processes", result.Query.ValueString())
	assert.Equal(t, "List processes", result.Description.ValueString())
	assert.False(t, result.Prebuilt.ValueBool())
	assert.True(t, result.Snapshot.ValueBool())
	assert.False(t, result.Removed.ValueBool())
	require.False(t, result.Platform.IsNull())
}

func TestDataSourceModel_populateFromGetAPI_prebuilt(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	prebuilt := true
	query := kbapi.SecurityOsqueryAPIQuery("SELECT 1")
	entity := &kibanaoapi.OsquerySavedQueryGetEntity{
		ID:            "list_all_processes",
		SavedObjectID: "saved-object-prebuilt",
		Query:         &query,
		Prebuilt:      &prebuilt,
	}

	model := dataSourceModel{
		SavedQueryID: types.StringValue("list_all_processes"),
		SpaceID:      types.StringValue("default"),
	}
	diags := model.populateFromGetAPI(ctx, entity)
	require.False(t, diags.HasError())
	assert.Equal(t, "default/list_all_processes", model.ID.ValueString())
	assert.Equal(t, "saved-object-prebuilt", model.SavedObjectID.ValueString())
	assert.Equal(t, "SELECT 1", model.Query.ValueString())
	assert.True(t, model.Prebuilt.ValueBool())
}

func TestDataSourceModel_populateFromGetAPI_userManaged(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	prebuilt := false
	query := kbapi.SecurityOsqueryAPIQuery("SELECT pid FROM processes")
	entity := &kibanaoapi.OsquerySavedQueryGetEntity{
		ID:       "list_processes",
		Query:    &query,
		Prebuilt: &prebuilt,
	}

	model := dataSourceModel{
		SavedQueryID: types.StringValue("list_processes"),
		SpaceID:      types.StringValue("production"),
	}
	diags := model.populateFromGetAPI(ctx, entity)
	require.False(t, diags.HasError())
	assert.Equal(t, "production/list_processes", model.ID.ValueString())
	assert.False(t, model.Prebuilt.ValueBool())
}

func TestDataSourceModel_populateFromGetAPI_prebuiltOmitted(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	query := kbapi.SecurityOsqueryAPIQuery("SELECT 1")
	entity := &kibanaoapi.OsquerySavedQueryGetEntity{
		ID:    "list_processes",
		Query: &query,
	}

	model := dataSourceModel{
		SavedQueryID: types.StringValue("list_processes"),
		SpaceID:      types.StringValue("default"),
	}
	diags := model.populateFromGetAPI(ctx, entity)
	require.False(t, diags.HasError())
	assert.False(t, model.Prebuilt.ValueBool())
	assert.False(t, model.Prebuilt.IsNull())
}

func TestPrebuiltFromAPI(t *testing.T) {
	t.Parallel()

	t.Run("nil defaults to false", func(t *testing.T) {
		result := prebuiltFromAPI(nil)
		assert.False(t, result.ValueBool())
		assert.False(t, result.IsNull())
	})

	t.Run("false stays false", func(t *testing.T) {
		prebuilt := false
		result := prebuiltFromAPI(&prebuilt)
		assert.False(t, result.ValueBool())
	})

	t.Run("true stays true", func(t *testing.T) {
		prebuilt := true
		result := prebuiltFromAPI(&prebuilt)
		assert.True(t, result.ValueBool())
	})
}

func dataSourceSchemaNestedObjectElemType(no dsschema.NestedAttributeObject) attr.Type {
	attrTypes := make(map[string]attr.Type, len(no.Attributes))
	for name, a := range no.Attributes {
		attrTypes[name] = a.GetType()
	}
	return types.ObjectType{AttrTypes: attrTypes}
}
