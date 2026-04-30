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

package entitycore

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/config"
	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

// testResourceModel satisfies ElasticsearchResourceModel for envelope tests.
type testResourceModel struct {
	ID                      types.String `tfsdk:"id"`
	ElasticsearchConnection types.List   `tfsdk:"elasticsearch_connection"`
}

func (m testResourceModel) GetID() types.String {
	return m.ID
}

func (m testResourceModel) GetElasticsearchConnection() types.List {
	return m.ElasticsearchConnection
}

func getTestResourceSchema() rschema.Schema {
	return rschema.Schema{
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func testReadFuncFound(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, model testResourceModel) (testResourceModel, bool, diag.Diagnostics) {
	return model, true, nil
}

func testDeleteFunc(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, _ testResourceModel) diag.Diagnostics {
	return nil
}

func elasticsearchConnectionBlockType() tftypes.Type {
	nestedObjType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"username":                 tftypes.String,
			"password":                 tftypes.String,
			"api_key":                  tftypes.String,
			"bearer_token":             tftypes.String,
			"es_client_authentication": tftypes.String,
			"endpoints":                tftypes.List{ElementType: tftypes.String},
			"headers":                  tftypes.Map{ElementType: tftypes.String},
			"insecure":                 tftypes.Bool,
			"ca_file":                  tftypes.String,
			"ca_data":                  tftypes.String,
			"cert_file":                tftypes.String,
			"key_file":                 tftypes.String,
			"cert_data":                tftypes.String,
			"key_data":                 tftypes.String,
		},
	}
	return tftypes.List{ElementType: nestedObjType}
}

func newTestConfiguredFactory(t *testing.T) *clients.ProviderClientFactory {
	t.Helper()
	factory, diags := clients.NewProviderClientFactoryFromFramework(context.Background(), config.ProviderConfiguration{}, "test")
	require.False(t, diags.HasError(), "failed to create test factory: %v", diags)
	require.NotNil(t, factory)
	return factory
}

func newResourceEnvelopeWithFactory(t *testing.T, factory *clients.ProviderClientFactory) *ElasticsearchResource[testResourceModel] {
	t.Helper()
	r := NewElasticsearchResource[testResourceModel](
		ComponentElasticsearch,
		"test_entity",
		getTestResourceSchema,
		testReadFuncFound,
		testDeleteFunc,
	)
	r.client = factory
	return r
}

func makeTestResourceState(t *testing.T, id string) tfsdk.State {
	t.Helper()
	connBlockType := elasticsearchConnectionBlockType()
	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":                       tftypes.String,
			"elasticsearch_connection": connBlockType,
		},
	}
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                       tftypes.NewValue(tftypes.String, id),
		"elasticsearch_connection": tftypes.NewValue(connBlockType, nil),
	})

	fullSchema := getTestResourceSchema()
	fullSchema.Blocks = map[string]rschema.Block{
		"elasticsearch_connection": providerschema.GetEsFWConnectionBlock(),
	}

	return tfsdk.State{
		Raw:    objValue,
		Schema: fullSchema,
	}
}

func TestNewElasticsearchResource_typeAssertions(t *testing.T) {
	t.Parallel()
	r := NewElasticsearchResource[testResourceModel](ComponentElasticsearch, "test_entity", getTestResourceSchema, testReadFuncFound, testDeleteFunc)
	require.NotNil(t, r)
	require.Implements(t, (*resource.Resource)(nil), r)
	require.Implements(t, (*resource.ResourceWithConfigure)(nil), r)

}

func TestNewElasticsearchResource_Metadata(t *testing.T) {
	t.Parallel()
	r := NewElasticsearchResource[testResourceModel](ComponentElasticsearch, "test_entity", getTestResourceSchema, testReadFuncFound, testDeleteFunc)

	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: testProviderTypeName}, &resp)

	require.Equal(t, "elasticstack_elasticsearch_test_entity", resp.TypeName)
}

func TestNewElasticsearchResource_schemaInjection(t *testing.T) {
	t.Parallel()
	r := NewElasticsearchResource[testResourceModel](ComponentElasticsearch, "test_entity", getTestResourceSchema, testReadFuncFound, testDeleteFunc)

	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Schema.Blocks, "elasticsearch_connection")
	require.Contains(t, resp.Schema.Attributes, "id")
}

func TestNewElasticsearchResource_schemaDefensiveClone(t *testing.T) {
	t.Parallel()
	originalSchema := getTestResourceSchema()
	r := NewElasticsearchResource[testResourceModel](ComponentElasticsearch, "test_entity", func() rschema.Schema {
		return originalSchema
	}, testReadFuncFound, testDeleteFunc)

	var resp1 resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp1)
	require.False(t, resp1.Diagnostics.HasError())

	var resp2 resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp2)
	require.False(t, resp2.Diagnostics.HasError())

	require.Nil(t, originalSchema.Blocks)
}

func TestNewElasticsearchResource_Configure(t *testing.T) {
	ctx := context.Background()

	t.Run("nil_provider_data", func(t *testing.T) {
		t.Parallel()
		r := NewElasticsearchResource[testResourceModel](ComponentElasticsearch, "test_entity", getTestResourceSchema, testReadFuncFound, testDeleteFunc)
		var resp resource.ConfigureResponse
		r.Configure(ctx, resource.ConfigureRequest{ProviderData: nil}, &resp)
		require.False(t, resp.Diagnostics.HasError())
	})

	t.Run("valid_factory", func(t *testing.T) {
		t.Parallel()
		r := NewElasticsearchResource[testResourceModel](ComponentElasticsearch, "test_entity", getTestResourceSchema, testReadFuncFound, testDeleteFunc)
		f := nonNilTestFactory()
		var resp resource.ConfigureResponse
		r.Configure(ctx, resource.ConfigureRequest{ProviderData: f}, &resp)
		require.False(t, resp.Diagnostics.HasError())
	})

	t.Run("invalid_provider_data", func(t *testing.T) {
		t.Parallel()
		r := NewElasticsearchResource[testResourceModel](ComponentElasticsearch, "test_entity", getTestResourceSchema, testReadFuncFound, testDeleteFunc)
		var resp resource.ConfigureResponse
		r.Configure(ctx, resource.ConfigureRequest{ProviderData: "wrong-type"}, &resp)
		require.True(t, resp.Diagnostics.HasError())
	})
}

func TestNewElasticsearchResource_Read_happyPath(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	factory := newTestConfiguredFactory(t)
	r := newResourceEnvelopeWithFactory(t, factory)

	state := makeTestResourceState(t, "cluster/user1")
	req := resource.ReadRequest{State: state}
	resp := resource.ReadResponse{State: state}

	r.Read(ctx, req, &resp)

	require.False(t, resp.Diagnostics.HasError())

	var result testResourceModel
	diags := resp.State.Get(ctx, &result)
	require.False(t, diags.HasError())
	require.Equal(t, "cluster/user1", result.ID.ValueString())
	require.True(t, result.ElasticsearchConnection.IsNull())
}

func TestNewElasticsearchResource_Read_notFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	factory := newTestConfiguredFactory(t)
	r := NewElasticsearchResource[testResourceModel](
		ComponentElasticsearch,
		"test_entity",
		getTestResourceSchema,
		func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, _ testResourceModel) (testResourceModel, bool, diag.Diagnostics) {
			return testResourceModel{}, false, nil
		},
		testDeleteFunc,
	)
	r.client = factory

	state := makeTestResourceState(t, "cluster/user1")
	req := resource.ReadRequest{State: state}
	resp := resource.ReadResponse{State: state}

	r.Read(ctx, req, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.True(t, resp.State.Raw.IsNull(), "expected state to be removed")
}

func TestNewElasticsearchResource_Read_shortCircuitStateGetError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	factory := newTestConfiguredFactory(t)
	readCalled := false
	r := NewElasticsearchResource[testResourceModel](
		ComponentElasticsearch,
		"test_entity",
		getTestResourceSchema,
		func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, _ testResourceModel) (testResourceModel, bool, diag.Diagnostics) {
			readCalled = true
			return testResourceModel{}, false, nil
		},
		testDeleteFunc,
	)
	r.client = factory

	// Schema missing elasticsearch_connection block while raw value includes it.
	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":                       tftypes.String,
			"elasticsearch_connection": elasticsearchConnectionBlockType(),
		},
	}
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                       tftypes.NewValue(tftypes.String, "cluster/user1"),
		"elasticsearch_connection": tftypes.NewValue(elasticsearchConnectionBlockType(), nil),
	})
	badSchema := getTestResourceSchema() // no elasticsearch_connection block
	state := tfsdk.State{Raw: objValue, Schema: badSchema}

	req := resource.ReadRequest{State: state}
	resp := resource.ReadResponse{State: state}

	r.Read(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.False(t, readCalled, "readFunc should not be called when state.Get fails")
}

func TestNewElasticsearchResource_Read_shortCircuitCompositeIDError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	factory := newTestConfiguredFactory(t)
	readCalled := false
	r := NewElasticsearchResource[testResourceModel](
		ComponentElasticsearch,
		"test_entity",
		getTestResourceSchema,
		func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, _ testResourceModel) (testResourceModel, bool, diag.Diagnostics) {
			readCalled = true
			return testResourceModel{}, false, nil
		},
		testDeleteFunc,
	)
	r.client = factory

	state := makeTestResourceState(t, "invalid-no-slash")
	req := resource.ReadRequest{State: state}
	resp := resource.ReadResponse{State: state}

	r.Read(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.False(t, readCalled, "readFunc should not be called when composite ID parse fails")
}

func TestNewElasticsearchResource_Read_shortCircuitClientError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// zero-value factory has nil defaultClient -> client resolution fails
	factory := nonNilTestFactory()
	readCalled := false
	r := NewElasticsearchResource[testResourceModel](
		ComponentElasticsearch,
		"test_entity",
		getTestResourceSchema,
		func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, _ testResourceModel) (testResourceModel, bool, diag.Diagnostics) {
			readCalled = true
			return testResourceModel{}, false, nil
		},
		testDeleteFunc,
	)
	r.client = factory

	state := makeTestResourceState(t, "cluster/user1")
	req := resource.ReadRequest{State: state}
	resp := resource.ReadResponse{State: state}

	r.Read(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Provider not configured")
	require.False(t, readCalled, "readFunc should not be called when client resolution fails")
}

func TestNewElasticsearchResource_Read_shortCircuitReadFuncError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	factory := newTestConfiguredFactory(t)
	r := NewElasticsearchResource[testResourceModel](
		ComponentElasticsearch,
		"test_entity",
		getTestResourceSchema,
		func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, _ testResourceModel) (testResourceModel, bool, diag.Diagnostics) {
			var diags diag.Diagnostics
			diags.AddError("read error", "something went wrong")
			return testResourceModel{}, false, diags
		},
		testDeleteFunc,
	)
	r.client = factory

	state := makeTestResourceState(t, "cluster/user1")
	req := resource.ReadRequest{State: state}
	resp := resource.ReadResponse{State: state}

	r.Read(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "read error")
	require.False(t, resp.State.Raw.IsNull(), "state should not be removed when readFunc returns an error")
}

func TestNewElasticsearchResource_Delete_happyPath(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	factory := newTestConfiguredFactory(t)
	deleteCalled := false
	r := NewElasticsearchResource[testResourceModel](
		ComponentElasticsearch,
		"test_entity",
		getTestResourceSchema,
		testReadFuncFound,
		func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, _ testResourceModel) diag.Diagnostics {
			deleteCalled = true
			return nil
		},
	)
	r.client = factory

	state := makeTestResourceState(t, "cluster/user1")
	req := resource.DeleteRequest{State: state}
	resp := resource.DeleteResponse{State: state}

	r.Delete(ctx, req, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.True(t, deleteCalled, "deleteFunc should be called")
}

func TestNewElasticsearchResource_Delete_shortCircuitStateGetError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	factory := newTestConfiguredFactory(t)
	deleteCalled := false
	r := NewElasticsearchResource[testResourceModel](
		ComponentElasticsearch,
		"test_entity",
		getTestResourceSchema,
		testReadFuncFound,
		func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, _ testResourceModel) diag.Diagnostics {
			deleteCalled = true
			return nil
		},
	)
	r.client = factory

	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":                       tftypes.String,
			"elasticsearch_connection": elasticsearchConnectionBlockType(),
		},
	}
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                       tftypes.NewValue(tftypes.String, "cluster/user1"),
		"elasticsearch_connection": tftypes.NewValue(elasticsearchConnectionBlockType(), nil),
	})
	badSchema := getTestResourceSchema()
	state := tfsdk.State{Raw: objValue, Schema: badSchema}

	req := resource.DeleteRequest{State: state}
	resp := resource.DeleteResponse{State: state}

	r.Delete(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.False(t, deleteCalled, "deleteFunc should not be called when state.Get fails")
}

func TestNewElasticsearchResource_Delete_shortCircuitCompositeIDError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	factory := newTestConfiguredFactory(t)
	deleteCalled := false
	r := NewElasticsearchResource[testResourceModel](
		ComponentElasticsearch,
		"test_entity",
		getTestResourceSchema,
		testReadFuncFound,
		func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, _ testResourceModel) diag.Diagnostics {
			deleteCalled = true
			return nil
		},
	)
	r.client = factory

	state := makeTestResourceState(t, "invalid-no-slash")
	req := resource.DeleteRequest{State: state}
	resp := resource.DeleteResponse{State: state}

	r.Delete(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.False(t, deleteCalled, "deleteFunc should not be called when composite ID parse fails")
}

func TestNewElasticsearchResource_Delete_shortCircuitClientError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	factory := nonNilTestFactory()
	deleteCalled := false
	r := NewElasticsearchResource[testResourceModel](
		ComponentElasticsearch,
		"test_entity",
		getTestResourceSchema,
		testReadFuncFound,
		func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, _ testResourceModel) diag.Diagnostics {
			deleteCalled = true
			return nil
		},
	)
	r.client = factory

	state := makeTestResourceState(t, "cluster/user1")
	req := resource.DeleteRequest{State: state}
	resp := resource.DeleteResponse{State: state}

	r.Delete(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Provider not configured")
	require.False(t, deleteCalled, "deleteFunc should not be called when client resolution fails")
}

func TestNewElasticsearchResource_Delete_appendsDeleteFuncDiagnostics(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	factory := newTestConfiguredFactory(t)
	r := NewElasticsearchResource[testResourceModel](
		ComponentElasticsearch,
		"test_entity",
		getTestResourceSchema,
		testReadFuncFound,
		func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, _ testResourceModel) diag.Diagnostics {
			var diags diag.Diagnostics
			diags.AddError("delete error", "something went wrong")
			return diags
		},
	)
	r.client = factory

	state := makeTestResourceState(t, "cluster/user1")
	req := resource.DeleteRequest{State: state}
	resp := resource.DeleteResponse{State: state}

	r.Delete(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "delete error")
}

type overridingEnvelopeTestResource struct {
	*ElasticsearchResource[testResourceModel]
	createCalled bool
	updateCalled bool
}

func (r *overridingEnvelopeTestResource) Create(context.Context, resource.CreateRequest, *resource.CreateResponse) {
	r.createCalled = true
}

func (r *overridingEnvelopeTestResource) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {
	r.updateCalled = true
}

func TestNewElasticsearchResource_Create_defaultDiagnostic(t *testing.T) {
	t.Parallel()
	r := NewElasticsearchResource[testResourceModel](ComponentElasticsearch, "test_entity", getTestResourceSchema, testReadFuncFound, testDeleteFunc)

	var resp resource.CreateResponse
	r.Create(context.Background(), resource.CreateRequest{}, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Create not implemented")
}

func TestNewElasticsearchResource_Update_defaultDiagnostic(t *testing.T) {
	t.Parallel()
	r := NewElasticsearchResource[testResourceModel](ComponentElasticsearch, "test_entity", getTestResourceSchema, testReadFuncFound, testDeleteFunc)

	var resp resource.UpdateResponse
	r.Update(context.Background(), resource.UpdateRequest{}, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Update not implemented")
}

func TestNewElasticsearchResource_CreateAndUpdate_concreteOverridesWin(t *testing.T) {
	t.Parallel()
	r := &overridingEnvelopeTestResource{
		ElasticsearchResource: NewElasticsearchResource[testResourceModel](
			ComponentElasticsearch,
			"test_entity",
			getTestResourceSchema,
			testReadFuncFound,
			testDeleteFunc,
		),
	}

	var createResp resource.CreateResponse
	var asResource resource.Resource = r
	asResource.Create(context.Background(), resource.CreateRequest{}, &createResp)
	require.True(t, r.createCalled)
	require.False(t, createResp.Diagnostics.HasError())

	var updateResp resource.UpdateResponse
	asResource.Update(context.Background(), resource.UpdateRequest{}, &updateResp)
	require.True(t, r.updateCalled)
	require.False(t, updateResp.Diagnostics.HasError())
}
