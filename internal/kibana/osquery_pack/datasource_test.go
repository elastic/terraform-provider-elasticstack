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
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	require.Equal(t, "elasticstack_kibana_osquery_pack", resp.TypeName)
}

func TestNewDataSource_schemaAttributes(t *testing.T) {
	t.Parallel()
	ds := NewDataSource()

	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)

	require.Contains(t, resp.Schema.Blocks, "kibana_connection")

	wantAttrs := []string{
		"id",
		"pack_id",
		"space_id",
		"name",
		"description",
		"enabled",
		"policy_ids",
		"shards",
		"queries",
		"read_only",
	}
	for _, attr := range wantAttrs {
		require.Contains(t, resp.Schema.Attributes, attr, "schema Attributes must contain %q", attr)
	}

	packIDAttr, ok := resp.Schema.Attributes["pack_id"].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, packIDAttr.IsRequired())
	assert.False(t, packIDAttr.IsComputed())

	readOnlyAttr, ok := resp.Schema.Attributes["read_only"].(schema.BoolAttribute)
	require.True(t, ok)
	assert.True(t, readOnlyAttr.IsComputed())

	queriesAttr, ok := resp.Schema.Attributes["queries"].(schema.MapNestedAttribute)
	require.True(t, ok)
	assert.True(t, queriesAttr.IsComputed())
	assert.False(t, queriesAttr.IsRequired())

	for _, attr := range []string{"id", "name", "description", "enabled", "policy_ids", "shards"} {
		switch attr {
		case "policy_ids":
			listAttr, ok := resp.Schema.Attributes[attr].(schema.ListAttribute)
			require.True(t, ok, "%q should be ListAttribute", attr)
			assert.True(t, listAttr.IsComputed(), "%q should be computed-only", attr)
			assert.False(t, listAttr.IsRequired(), "%q should not be required", attr)
			assert.False(t, listAttr.IsOptional(), "%q should not be optional", attr)
		case "shards":
			mapAttr, ok := resp.Schema.Attributes[attr].(schema.MapAttribute)
			require.True(t, ok, "%q should be MapAttribute", attr)
			assert.True(t, mapAttr.IsComputed(), "%q should be computed-only", attr)
			assert.False(t, mapAttr.IsRequired(), "%q should not be required", attr)
			assert.False(t, mapAttr.IsOptional(), "%q should not be optional", attr)
		case "enabled":
			boolAttr, ok := resp.Schema.Attributes[attr].(schema.BoolAttribute)
			require.True(t, ok, "%q should be BoolAttribute", attr)
			assert.True(t, boolAttr.IsComputed(), "%q should be computed-only", attr)
			assert.False(t, boolAttr.IsRequired(), "%q should not be required", attr)
			assert.False(t, boolAttr.IsOptional(), "%q should not be optional", attr)
		default:
			strAttr, ok := resp.Schema.Attributes[attr].(schema.StringAttribute)
			require.True(t, ok, "%q should be StringAttribute", attr)
			assert.True(t, strAttr.IsComputed(), "%q should be computed-only", attr)
			assert.False(t, strAttr.IsRequired(), "%q should not be required", attr)
			assert.False(t, strAttr.IsOptional(), "%q should not be optional", attr)
		}
	}

	spaceIDAttr, ok := resp.Schema.Attributes["space_id"].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, spaceIDAttr.IsOptional())
	assert.False(t, spaceIDAttr.IsRequired())
	assert.False(t, spaceIDAttr.IsComputed())
}

func TestDataSourceSchema_spaceIDOptionalWithoutSchemaDefault(t *testing.T) {
	t.Parallel()

	spaceIDAttr, ok := getDataSourceSchema(context.Background()).Attributes["space_id"].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, spaceIDAttr.IsOptional())
	assert.False(t, spaceIDAttr.IsComputed())

	// Datasource schema lacks resource-style Default; read resolves omitted space_id to default.
	spaceID, _ := clients.ResolveCompositeSpaceAndID(types.StringNull(), "pack-id")
	assert.Equal(t, clients.DefaultSpaceID, spaceID)
}

func TestDataSourceModel_GetVersionRequirements(t *testing.T) {
	t.Parallel()
	var m dataSourceModel
	reqs, diags := m.GetVersionRequirements(context.Background())
	require.False(t, diags.HasError())
	require.Len(t, reqs, 1)
	require.Equal(t, osqueryPackMinVersion, &reqs[0].MinVersion)
	require.NotEmpty(t, reqs[0].ErrorMessage)
}

func TestDataSourceModel_satisfiesWithVersionRequirements(t *testing.T) {
	t.Parallel()
	var _ entitycore.WithVersionRequirements = (*dataSourceModel)(nil)
}

func TestDataSourceModel_populateFromAPI_readOnlyFalse(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	readOnly := false
	detail := &kibanaoapi.OsqueryPackDetail{
		SavedObjectId: "pack-id",
		Name:          kbapi.SecurityOsqueryAPIPackName("Managed pack"),
		ReadOnly:      &readOnly,
	}

	var model dataSourceModel
	diags := model.populateFromAPI(ctx, "default", detail)
	require.False(t, diags.HasError(), "%v", diags)
	require.False(t, model.ReadOnly.ValueBool())
}

func TestDataSourceModel_populateFromAPI_readOnlyNil(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	detail := &kibanaoapi.OsqueryPackDetail{
		SavedObjectId: "pack-id",
		Name:          kbapi.SecurityOsqueryAPIPackName("Managed pack"),
		ReadOnly:      nil,
	}

	var model dataSourceModel
	diags := model.populateFromAPI(ctx, "default", detail)
	require.False(t, diags.HasError(), "%v", diags)
	require.True(t, model.ReadOnly.IsNull())
}

func TestDataSourceModel_populateFromAPI_prebuiltPack(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	readOnly := true
	name := kbapi.SecurityOsqueryAPIPackName("Prebuilt pack")
	query := kbapi.SecurityOsqueryAPIQuery("SELECT 1;")
	detail := &kibanaoapi.OsqueryPackDetail{
		SavedObjectId: "3c42c847-eb30-4452-80e0-728584042334",
		Name:          name,
		ReadOnly:      &readOnly,
		Queries: &kbapi.SecurityOsqueryAPIObjectQueries{
			"find_procs": {
				Query: &query,
			},
		},
	}

	var model dataSourceModel
	diags := model.populateFromAPI(ctx, "default", detail)
	require.False(t, diags.HasError(), "%v", diags)
	require.Equal(t, "default/3c42c847-eb30-4452-80e0-728584042334", model.ID.ValueString())
	require.Equal(t, "3c42c847-eb30-4452-80e0-728584042334", model.PackID.ValueString())
	require.True(t, model.ReadOnly.ValueBool())
	require.Equal(t, "Prebuilt pack", model.Name.ValueString())
}

func TestQueryDataSourceElemType_matchesSchema(t *testing.T) {
	t.Parallel()

	queriesAttr, ok := getDataSourceSchema(context.Background()).Attributes["queries"].(schema.MapNestedAttribute)
	require.True(t, ok)

	schemaElem := dataSourceNestedObjectElemType(queriesAttr.NestedObject)
	require.Equal(t, schemaElem, queryMapElemType(),
		"queryMapElemType() drifted from the data source queries nested object; update both together")
}

func TestEcsMappingDataSourceElemType_matchesSchema(t *testing.T) {
	t.Parallel()

	queriesAttr, ok := getDataSourceSchema(context.Background()).Attributes["queries"].(schema.MapNestedAttribute)
	require.True(t, ok)

	ecsMappingAttr, ok := queriesAttr.NestedObject.Attributes["ecs_mapping"].(schema.MapNestedAttribute)
	require.True(t, ok)

	schemaElem := dataSourceNestedObjectElemType(ecsMappingAttr.NestedObject)
	require.Equal(t, schemaElem, ecsMappingMapElemType(),
		"ecsMappingMapElemType() drifted from the data source ecs_mapping nested object; update both together")
}

func dataSourceNestedObjectElemType(no schema.NestedAttributeObject) attr.Type {
	attrTypes := make(map[string]attr.Type, len(no.Attributes))
	for name, a := range no.Attributes {
		attrTypes[name] = a.GetType()
	}
	return types.ObjectType{AttrTypes: attrTypes}
}
