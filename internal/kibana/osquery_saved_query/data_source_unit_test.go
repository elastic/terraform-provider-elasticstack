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
		"id",
		"saved_query_id",
		"space_id",
		"query",
		"description",
		"platform",
		"interval",
		"version",
		"snapshot",
		"removed",
		"ecs_mapping",
		"prebuilt",
	}
	for _, attr := range wantAttrs {
		require.Contains(t, resp.Schema.Attributes, attr, "schema Attributes must contain %q", attr)
	}

	savedQueryIDAttr, ok := resp.Schema.Attributes["saved_query_id"].(dsschema.StringAttribute)
	require.True(t, ok)
	require.True(t, savedQueryIDAttr.IsRequired())
	require.False(t, savedQueryIDAttr.IsComputed())

	idAttr, ok := resp.Schema.Attributes["id"].(dsschema.StringAttribute)
	require.True(t, ok)
	require.True(t, idAttr.IsComputed())
	require.False(t, idAttr.IsRequired())

	prebuiltAttr, ok := resp.Schema.Attributes["prebuilt"].(dsschema.BoolAttribute)
	require.True(t, ok)
	require.True(t, prebuiltAttr.IsComputed())
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

func TestDataSourceModel_populateFromGetAPI_prebuilt(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	prebuilt := true
	query := kbapi.SecurityOsqueryAPIQuery("SELECT 1")
	entity := &kibanaoapi.OsquerySavedQueryGetEntity{
		ID:       "list_all_processes",
		Query:    &query,
		Prebuilt: &prebuilt,
	}

	model := dataSourceModel{
		SavedQueryID: types.StringValue("list_all_processes"),
		SpaceID:      types.StringValue("default"),
	}
	diags := model.populateFromGetAPI(ctx, entity)
	require.False(t, diags.HasError())
	assert.Equal(t, "default/list_all_processes", model.ID.ValueString())
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
