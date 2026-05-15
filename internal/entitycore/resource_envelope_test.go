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
	"strings"
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
	Name                    types.String `tfsdk:"name"`
	ElasticsearchConnection types.List   `tfsdk:"elasticsearch_connection"`
}

func (m testResourceModel) GetID() types.String {
	return m.ID
}

func (m testResourceModel) GetResourceID() types.String {
	return m.Name
}

func (m testResourceModel) GetElasticsearchConnection() types.List {
	return m.ElasticsearchConnection
}

func getTestResourceSchema(_ context.Context) rschema.Schema {
	return rschema.Schema{
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed: true,
			},
			"name": rschema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func testReadFuncFound(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, model testResourceModel) (testResourceModel, bool, diag.Diagnostics) {
	return model, true, nil
}

func testReadFuncDistinguishing(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, model testResourceModel) (testResourceModel, bool, diag.Diagnostics) {
	model.Name = types.StringValue(model.Name.ValueString() + "-read")
	return model, true, nil
}

func testDeleteFunc(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, _ testResourceModel) diag.Diagnostics {
	return nil
}

func defaultTestElasticsearchResourceOptions() ElasticsearchResourceOptions[testResourceModel] {
	return ElasticsearchResourceOptions[testResourceModel]{
		Schema: getTestResourceSchema,
		Read:   testReadFuncFound,
		Delete: testDeleteFunc,
		Create: testWriteFuncFoundCreate,
		Update: testWriteFuncFoundUpdate,
	}
}

func testWriteFuncFoundCreate(
	_ context.Context,
	_ *clients.ElasticsearchScopedClient,
	req WriteRequest[testResourceModel],
) (WriteResult[testResourceModel], diag.Diagnostics) {
	model := req.Plan
	model.ID = types.StringValue("cluster/" + req.WriteID)
	return WriteResult[testResourceModel]{Model: model}, nil
}

func testWriteFuncFoundUpdate(
	_ context.Context,
	_ *clients.ElasticsearchScopedClient,
	req WriteRequest[testResourceModel],
) (WriteResult[testResourceModel], diag.Diagnostics) {
	model := req.Plan
	model.ID = types.StringValue("cluster/" + req.WriteID)
	return WriteResult[testResourceModel]{Model: model}, nil
}

func testResourceObjectType() tftypes.Type {
	return tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":                       tftypes.String,
			"name":                     tftypes.String,
			"elasticsearch_connection": elasticsearchConnectionBlockType(),
		},
	}
}

func testResourceNameFromCompositeID(compositeID string) string {
	_, res, ok := strings.Cut(compositeID, "/")
	if !ok {
		return compositeID
	}
	return res
}

func testResourceSchemaWithConnectionBlock(ctx context.Context) rschema.Schema {
	s := getTestResourceSchema(ctx)
	s.Blocks = map[string]rschema.Block{
		"elasticsearch_connection": providerschema.GetEsFWConnectionBlock(),
	}
	return s
}

func makeTestResourceCreatePlan(ctx context.Context, t *testing.T, idValue tftypes.Value) tfsdk.Plan {
	t.Helper()
	const resourceName = "user1"
	objType := testResourceObjectType()
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                       idValue,
		"name":                     tftypes.NewValue(tftypes.String, resourceName),
		"elasticsearch_connection": tftypes.NewValue(elasticsearchConnectionBlockType(), nil),
	})
	return tfsdk.Plan{
		Raw:    objValue,
		Schema: testResourceSchemaWithConnectionBlock(ctx),
	}
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

func newTestConfiguredFactory(ctx context.Context, t *testing.T) *clients.ProviderClientFactory {
	t.Helper()
	factory, diags := clients.NewProviderClientFactoryFromFramework(ctx, config.ProviderConfiguration{}, "test")
	require.False(t, diags.HasError(), "failed to create test factory: %v", diags)
	require.NotNil(t, factory)
	return factory
}

func newResourceEnvelopeWithFactory(t *testing.T, factory *clients.ProviderClientFactory) *ElasticsearchResource[testResourceModel] {
	t.Helper()
	r := NewElasticsearchResource[testResourceModel]("test_entity", defaultTestElasticsearchResourceOptions())
	r.client = factory
	return r
}

func makeTestResourceState(ctx context.Context, t *testing.T, id string) tfsdk.State {
	t.Helper()
	connBlockType := elasticsearchConnectionBlockType()
	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":                       tftypes.String,
			"name":                     tftypes.String,
			"elasticsearch_connection": connBlockType,
		},
	}
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                       tftypes.NewValue(tftypes.String, id),
		"name":                     tftypes.NewValue(tftypes.String, testResourceNameFromCompositeID(id)),
		"elasticsearch_connection": tftypes.NewValue(connBlockType, nil),
	})

	return tfsdk.State{
		Raw:    objValue,
		Schema: testResourceSchemaWithConnectionBlock(ctx),
	}
}

func TestNewElasticsearchResource_typeAssertions(t *testing.T) {
	t.Parallel()
	r := NewElasticsearchResource[testResourceModel]("test_entity", defaultTestElasticsearchResourceOptions())
	require.NotNil(t, r)
	require.Implements(t, (*resource.Resource)(nil), r)
	require.Implements(t, (*resource.ResourceWithConfigure)(nil), r)
	_, implementsImport := any(r).(resource.ResourceWithImportState)
	require.False(t, implementsImport, "envelope must not implement ImportState; concrete resources add it when needed")
}

func TestNewElasticsearchResource_Metadata(t *testing.T) {
	t.Parallel()
	r := NewElasticsearchResource[testResourceModel]("test_entity", defaultTestElasticsearchResourceOptions())

	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: testProviderTypeName}, &resp)

	require.Equal(t, "elasticstack_elasticsearch_test_entity", resp.TypeName)
}

func TestNewElasticsearchResource_schemaInjection(t *testing.T) {
	t.Parallel()
	r := NewElasticsearchResource[testResourceModel]("test_entity", defaultTestElasticsearchResourceOptions())

	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Schema.Blocks, "elasticsearch_connection")
	require.Contains(t, resp.Schema.Attributes, "id")
}

func TestNewElasticsearchResource_schemaDefensiveClone(t *testing.T) {
	t.Parallel()
	originalSchema := getTestResourceSchema(context.Background())
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema: func(_ context.Context) rschema.Schema {
			return originalSchema
		},
		Read:   testReadFuncFound,
		Delete: testDeleteFunc,
		Create: testWriteFuncFoundCreate,
		Update: testWriteFuncFoundUpdate,
	})

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
		r := NewElasticsearchResource[testResourceModel]("test_entity", defaultTestElasticsearchResourceOptions())
		var resp resource.ConfigureResponse
		r.Configure(ctx, resource.ConfigureRequest{ProviderData: nil}, &resp)
		require.False(t, resp.Diagnostics.HasError())
	})

	t.Run("valid_factory", func(t *testing.T) {
		t.Parallel()
		r := NewElasticsearchResource[testResourceModel]("test_entity", defaultTestElasticsearchResourceOptions())
		f := nonNilTestFactory()
		var resp resource.ConfigureResponse
		r.Configure(ctx, resource.ConfigureRequest{ProviderData: f}, &resp)
		require.False(t, resp.Diagnostics.HasError())
	})

	t.Run("invalid_provider_data", func(t *testing.T) {
		t.Parallel()
		r := NewElasticsearchResource[testResourceModel]("test_entity", defaultTestElasticsearchResourceOptions())
		var resp resource.ConfigureResponse
		r.Configure(ctx, resource.ConfigureRequest{ProviderData: "wrong-type"}, &resp)
		require.True(t, resp.Diagnostics.HasError())
	})
}

func TestNewElasticsearchResource_Read_happyPath(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	factory := newTestConfiguredFactory(ctx, t)
	r := newResourceEnvelopeWithFactory(t, factory)

	state := makeTestResourceState(ctx, t, "cluster/user1")
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

	factory := newTestConfiguredFactory(ctx, t)
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema: getTestResourceSchema,
		Read: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, _ testResourceModel) (testResourceModel, bool, diag.Diagnostics) {
			return testResourceModel{}, false, nil
		},
		Delete: testDeleteFunc,
		Create: testWriteFuncFoundCreate,
		Update: testWriteFuncFoundUpdate,
	})
	r.client = factory

	state := makeTestResourceState(ctx, t, "cluster/user1")
	req := resource.ReadRequest{State: state}
	resp := resource.ReadResponse{State: state}

	r.Read(ctx, req, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.True(t, resp.State.Raw.IsNull(), "expected state to be removed")
}

func TestNewElasticsearchResource_Read_shortCircuitStateGetError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	factory := newTestConfiguredFactory(ctx, t)
	readCalled := false
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema: getTestResourceSchema,
		Read: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, _ testResourceModel) (testResourceModel, bool, diag.Diagnostics) {
			readCalled = true
			return testResourceModel{}, false, nil
		},
		Delete: testDeleteFunc,
		Create: testWriteFuncFoundCreate,
		Update: testWriteFuncFoundUpdate,
	})
	r.client = factory

	// Schema missing elasticsearch_connection block while raw value includes it.
	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":                       tftypes.String,
			"name":                     tftypes.String,
			"elasticsearch_connection": elasticsearchConnectionBlockType(),
		},
	}
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                       tftypes.NewValue(tftypes.String, "cluster/user1"),
		"name":                     tftypes.NewValue(tftypes.String, "user1"),
		"elasticsearch_connection": tftypes.NewValue(elasticsearchConnectionBlockType(), nil),
	})
	badSchema := getTestResourceSchema(context.Background()) // no elasticsearch_connection block
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

	factory := newTestConfiguredFactory(ctx, t)
	readCalled := false
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema: getTestResourceSchema,
		Read: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, _ testResourceModel) (testResourceModel, bool, diag.Diagnostics) {
			readCalled = true
			return testResourceModel{}, false, nil
		},
		Delete: testDeleteFunc,
		Create: testWriteFuncFoundCreate,
		Update: testWriteFuncFoundUpdate,
	})
	r.client = factory

	state := makeTestResourceState(ctx, t, "invalid-no-slash")
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
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema: getTestResourceSchema,
		Read: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, _ testResourceModel) (testResourceModel, bool, diag.Diagnostics) {
			readCalled = true
			return testResourceModel{}, false, nil
		},
		Delete: testDeleteFunc,
		Create: testWriteFuncFoundCreate,
		Update: testWriteFuncFoundUpdate,
	})
	r.client = factory

	state := makeTestResourceState(ctx, t, "cluster/user1")
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

	factory := newTestConfiguredFactory(ctx, t)
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema: getTestResourceSchema,
		Read: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, _ testResourceModel) (testResourceModel, bool, diag.Diagnostics) {
			var diags diag.Diagnostics
			diags.AddError("read error", "something went wrong")
			return testResourceModel{}, false, diags
		},
		Delete: testDeleteFunc,
		Create: testWriteFuncFoundCreate,
		Update: testWriteFuncFoundUpdate,
	})
	r.client = factory

	state := makeTestResourceState(ctx, t, "cluster/user1")
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

	factory := newTestConfiguredFactory(ctx, t)
	deleteCalled := false
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema: getTestResourceSchema,
		Read:   testReadFuncFound,
		Delete: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, _ testResourceModel) diag.Diagnostics {
			deleteCalled = true
			return nil
		},
		Create: testWriteFuncFoundCreate,
		Update: testWriteFuncFoundUpdate,
	})
	r.client = factory

	state := makeTestResourceState(ctx, t, "cluster/user1")
	req := resource.DeleteRequest{State: state}
	resp := resource.DeleteResponse{State: state}

	r.Delete(ctx, req, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.True(t, deleteCalled, "deleteFunc should be called")
}

func TestNewElasticsearchResource_Delete_shortCircuitStateGetError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	factory := newTestConfiguredFactory(ctx, t)
	deleteCalled := false
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema: getTestResourceSchema,
		Read:   testReadFuncFound,
		Delete: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, _ testResourceModel) diag.Diagnostics {
			deleteCalled = true
			return nil
		},
		Create: testWriteFuncFoundCreate,
		Update: testWriteFuncFoundUpdate,
	})
	r.client = factory

	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":                       tftypes.String,
			"name":                     tftypes.String,
			"elasticsearch_connection": elasticsearchConnectionBlockType(),
		},
	}
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                       tftypes.NewValue(tftypes.String, "cluster/user1"),
		"name":                     tftypes.NewValue(tftypes.String, "user1"),
		"elasticsearch_connection": tftypes.NewValue(elasticsearchConnectionBlockType(), nil),
	})
	badSchema := getTestResourceSchema(context.Background())
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

	factory := newTestConfiguredFactory(ctx, t)
	deleteCalled := false
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema: getTestResourceSchema,
		Read:   testReadFuncFound,
		Delete: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, _ testResourceModel) diag.Diagnostics {
			deleteCalled = true
			return nil
		},
		Create: testWriteFuncFoundCreate,
		Update: testWriteFuncFoundUpdate,
	})
	r.client = factory

	state := makeTestResourceState(ctx, t, "invalid-no-slash")
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
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema: getTestResourceSchema,
		Read:   testReadFuncFound,
		Delete: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, _ testResourceModel) diag.Diagnostics {
			deleteCalled = true
			return nil
		},
		Create: testWriteFuncFoundCreate,
		Update: testWriteFuncFoundUpdate,
	})
	r.client = factory

	state := makeTestResourceState(ctx, t, "cluster/user1")
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

	factory := newTestConfiguredFactory(ctx, t)
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema: getTestResourceSchema,
		Read:   testReadFuncFound,
		Delete: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, _ testResourceModel) diag.Diagnostics {
			var diags diag.Diagnostics
			diags.AddError("delete error", "something went wrong")
			return diags
		},
		Create: testWriteFuncFoundCreate,
		Update: testWriteFuncFoundUpdate,
	})
	r.client = factory

	state := makeTestResourceState(ctx, t, "cluster/user1")
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

func TestNewElasticsearchResource_Create_happyPath(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema: getTestResourceSchema,
		Read:   testReadFuncDistinguishing,
		Delete: testDeleteFunc,
		Create: testWriteFuncFoundCreate,
		Update: testWriteFuncFoundUpdate,
	})
	r.client = factory

	plan := makeTestResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue))
	objType := testResourceObjectType()
	respState := tfsdk.State{
		Raw:    tftypes.NewValue(objType, nil),
		Schema: testResourceSchemaWithConnectionBlock(ctx),
	}
	req := resource.CreateRequest{Plan: plan}
	resp := resource.CreateResponse{State: respState}

	r.Create(ctx, req, &resp)

	require.False(t, resp.Diagnostics.HasError())
	var result testResourceModel
	diags := resp.State.Get(ctx, &result)
	require.False(t, diags.HasError())
	require.Equal(t, "cluster/user1", result.ID.ValueString())
	require.Equal(t, "user1-read", result.Name.ValueString())
}

func TestNewElasticsearchResource_Create_nilWriteCallback(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	var nilCreate WriteFunc[testResourceModel]
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema: getTestResourceSchema,
		Read:   testReadFuncFound,
		Delete: testDeleteFunc,
		Create: nilCreate,
		Update: testWriteFuncFoundUpdate,
	})
	r.client = factory

	plan := makeTestResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue))
	objType := testResourceObjectType()
	respState := tfsdk.State{
		Raw:    tftypes.NewValue(objType, nil),
		Schema: testResourceSchemaWithConnectionBlock(ctx),
	}
	req := resource.CreateRequest{Plan: plan}
	resp := resource.CreateResponse{State: respState}

	r.Create(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Elasticsearch envelope configuration error")
}

func TestNewElasticsearchResource_write_nilCallbackPrecedesOtherWritePreludeErrors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("Create_precedesClientError", func(t *testing.T) {
		t.Parallel()
		var nilCreate WriteFunc[testResourceModel]
		r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
			Schema: getTestResourceSchema,
			Read:   testReadFuncFound,
			Delete: testDeleteFunc,
			Create: nilCreate,
			Update: testWriteFuncFoundUpdate,
		})
		r.client = nonNilTestFactory()

		plan := makeTestResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue))
		resp := resource.CreateResponse{State: tfsdk.State{
			Raw:    tftypes.NewValue(testResourceObjectType(), nil),
			Schema: testResourceSchemaWithConnectionBlock(ctx),
		}}
		r.Create(ctx, resource.CreateRequest{Plan: plan}, &resp)

		require.True(t, resp.Diagnostics.HasError())
		require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Elasticsearch envelope configuration error")
		require.NotContains(t, resp.Diagnostics.Errors()[0].Summary(), "Provider not configured")
	})

	t.Run("Create_precedesInvalidWriteID", func(t *testing.T) {
		t.Parallel()
		var nilCreate WriteFunc[testResourceModel]
		r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
			Schema: getTestResourceSchema,
			Read:   testReadFuncFound,
			Delete: testDeleteFunc,
			Create: nilCreate,
			Update: testWriteFuncFoundUpdate,
		})
		r.client = newTestConfiguredFactory(ctx, t)

		objType := testResourceObjectType()
		objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
			"id":                       tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			"name":                     tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			"elasticsearch_connection": tftypes.NewValue(elasticsearchConnectionBlockType(), nil),
		})
		plan := tfsdk.Plan{Raw: objValue, Schema: testResourceSchemaWithConnectionBlock(ctx)}
		resp := resource.CreateResponse{State: tfsdk.State{
			Raw:    tftypes.NewValue(objType, nil),
			Schema: testResourceSchemaWithConnectionBlock(ctx),
		}}
		r.Create(ctx, resource.CreateRequest{Plan: plan}, &resp)

		require.True(t, resp.Diagnostics.HasError())
		require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Elasticsearch envelope configuration error")
		require.NotContains(t, resp.Diagnostics.Errors()[0].Summary(), "Invalid resource identifier")
	})

	t.Run("Update_precedesClientError", func(t *testing.T) {
		t.Parallel()
		var nilUpdate WriteFunc[testResourceModel]
		r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
			Schema: getTestResourceSchema,
			Read:   testReadFuncFound,
			Delete: testDeleteFunc,
			Create: testWriteFuncFoundCreate,
			Update: nilUpdate,
		})
		r.client = nonNilTestFactory()

		plan := makeTestResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "cluster/user1"))
		prior := makeTestResourceState(ctx, t, "cluster/user1")
		resp := resource.UpdateResponse{State: prior}
		r.Update(ctx, resource.UpdateRequest{Plan: plan, State: prior}, &resp)

		require.True(t, resp.Diagnostics.HasError())
		require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Elasticsearch envelope configuration error")
		require.NotContains(t, resp.Diagnostics.Errors()[0].Summary(), "Provider not configured")
	})
}

func TestNewElasticsearchResource_Create_placeholderCallbackError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	placeholder := PlaceholderElasticsearchWriteCallback[testResourceModel]()
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema: getTestResourceSchema,
		Read:   testReadFuncFound,
		Delete: testDeleteFunc,
		Create: placeholder,
		Update: placeholder,
	})
	r.client = factory

	plan := makeTestResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue))
	objType := testResourceObjectType()
	respState := tfsdk.State{
		Raw:    tftypes.NewValue(objType, nil),
		Schema: testResourceSchemaWithConnectionBlock(ctx),
	}
	req := resource.CreateRequest{Plan: plan}
	resp := resource.CreateResponse{State: respState}

	r.Create(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	err0 := resp.Diagnostics.Errors()[0]
	require.Equal(t, placeholderWriteCallbackSummary, err0.Summary())
	require.Equal(t, placeholderWriteCallbackDetail, err0.Detail())
}

func TestNewElasticsearchResource_Create_shortCircuitUnknownWriteID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	createCalled := false
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema: getTestResourceSchema,
		Read:   testReadFuncFound,
		Delete: testDeleteFunc,
		Create: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ WriteRequest[testResourceModel]) (WriteResult[testResourceModel], diag.Diagnostics) {
			createCalled = true
			return WriteResult[testResourceModel]{Model: testResourceModel{}}, nil
		},
		Update: testWriteFuncFoundUpdate,
	})
	r.client = factory

	objType := testResourceObjectType()
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                       tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"name":                     tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"elasticsearch_connection": tftypes.NewValue(elasticsearchConnectionBlockType(), nil),
	})
	plan := tfsdk.Plan{Raw: objValue, Schema: testResourceSchemaWithConnectionBlock(ctx)}
	respState := tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: testResourceSchemaWithConnectionBlock(ctx)}
	req := resource.CreateRequest{Plan: plan}
	resp := resource.CreateResponse{State: respState}

	r.Create(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Invalid resource identifier")
	require.False(t, createCalled, "create callback should not run when write identity is unknown")
}

func TestNewElasticsearchResource_Create_shortCircuitClientError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := nonNilTestFactory()
	createCalled := false
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema: getTestResourceSchema,
		Read:   testReadFuncFound,
		Delete: testDeleteFunc,
		Create: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ WriteRequest[testResourceModel]) (WriteResult[testResourceModel], diag.Diagnostics) {
			createCalled = true
			return WriteResult[testResourceModel]{Model: testResourceModel{}}, nil
		},
		Update: testWriteFuncFoundUpdate,
	})
	r.client = factory

	plan := makeTestResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue))
	objType := testResourceObjectType()
	respState := tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: testResourceSchemaWithConnectionBlock(ctx)}
	req := resource.CreateRequest{Plan: plan}
	resp := resource.CreateResponse{State: respState}

	r.Create(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Provider not configured")
	require.False(t, createCalled, "create callback should not run when client resolution fails")
}

func TestNewElasticsearchResource_Create_shortCircuitCallbackError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema: getTestResourceSchema,
		Read:   testReadFuncFound,
		Delete: testDeleteFunc,
		Create: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ WriteRequest[testResourceModel]) (WriteResult[testResourceModel], diag.Diagnostics) {
			var diags diag.Diagnostics
			diags.AddError("create error", "something went wrong")
			return WriteResult[testResourceModel]{Model: testResourceModel{}}, diags
		},
		Update: testWriteFuncFoundUpdate,
	})
	r.client = factory

	plan := makeTestResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue))
	objType := testResourceObjectType()
	respState := tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: testResourceSchemaWithConnectionBlock(ctx)}
	req := resource.CreateRequest{Plan: plan}
	resp := resource.CreateResponse{State: respState}

	r.Create(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "create error")
	require.True(t, resp.State.Raw.IsNull(), "state should not be mutated when create callback fails")
}

func TestNewElasticsearchResource_Create_readAfterWriteHappyPath(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema: getTestResourceSchema,
		Read: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, model testResourceModel) (testResourceModel, bool, diag.Diagnostics) {
			model.Name = types.StringValue("from-readfunc")
			return model, true, nil
		},
		Delete: testDeleteFunc,
		Create: testWriteFuncFoundCreate,
		Update: testWriteFuncFoundUpdate,
	})
	r.client = factory

	plan := makeTestResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue))
	objType := testResourceObjectType()
	respState := tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: testResourceSchemaWithConnectionBlock(ctx)}
	req := resource.CreateRequest{Plan: plan}
	resp := resource.CreateResponse{State: respState}

	r.Create(ctx, req, &resp)

	require.False(t, resp.Diagnostics.HasError())
	var result testResourceModel
	diags := resp.State.Get(ctx, &result)
	require.False(t, diags.HasError())
	require.Equal(t, "cluster/user1", result.ID.ValueString())
	require.Equal(t, "from-readfunc", result.Name.ValueString())
}

func TestNewElasticsearchResource_Create_notFoundAfterWrite(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema: getTestResourceSchema,
		Read: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, _ testResourceModel) (testResourceModel, bool, diag.Diagnostics) {
			return testResourceModel{}, false, nil
		},
		Delete: testDeleteFunc,
		Create: testWriteFuncFoundCreate,
		Update: testWriteFuncFoundUpdate,
	})
	r.client = factory

	plan := makeTestResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue))
	objType := testResourceObjectType()
	respState := tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: testResourceSchemaWithConnectionBlock(ctx)}
	req := resource.CreateRequest{Plan: plan}
	resp := resource.CreateResponse{State: respState}

	r.Create(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Resource not found")
	require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), `elasticsearch_test_entity "user1" was not found after write`)
	require.True(t, resp.State.Raw.IsNull(), "state should not be mutated when resource not found after create")
}

func TestNewElasticsearchResource_Create_readFuncError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema: getTestResourceSchema,
		Read: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, _ testResourceModel) (testResourceModel, bool, diag.Diagnostics) {
			var diags diag.Diagnostics
			diags.AddError("read error", "something went wrong after create")
			return testResourceModel{}, false, diags
		},
		Delete: testDeleteFunc,
		Create: testWriteFuncFoundCreate,
		Update: testWriteFuncFoundUpdate,
	})
	r.client = factory

	plan := makeTestResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue))
	objType := testResourceObjectType()
	respState := tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: testResourceSchemaWithConnectionBlock(ctx)}
	req := resource.CreateRequest{Plan: plan}
	resp := resource.CreateResponse{State: respState}

	r.Create(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "read error")
	require.True(t, resp.State.Raw.IsNull(), "state should not be mutated when readFunc returns errors after create")
}

func TestNewElasticsearchResource_Update_happyPath(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema: getTestResourceSchema,
		Read:   testReadFuncDistinguishing,
		Delete: testDeleteFunc,
		Create: testWriteFuncFoundCreate,
		Update: testWriteFuncFoundUpdate,
	})
	r.client = factory

	plan := makeTestResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "cluster/user1"))
	prior := makeTestResourceState(ctx, t, "cluster/user1")
	resp := resource.UpdateResponse{State: prior}
	req := resource.UpdateRequest{Plan: plan, State: prior}

	r.Update(ctx, req, &resp)

	require.False(t, resp.Diagnostics.HasError())
	var result testResourceModel
	diags := resp.State.Get(ctx, &result)
	require.False(t, diags.HasError())
	require.Equal(t, "cluster/user1", result.ID.ValueString())
	require.Equal(t, "user1-read", result.Name.ValueString())
}

func TestNewElasticsearchResource_Update_invokesUpdateCallbackNotCreate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	createCalled := false
	updateCalled := false
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema: getTestResourceSchema,
		Read:   testReadFuncFound,
		Delete: testDeleteFunc,
		Create: func(ctx context.Context, client *clients.ElasticsearchScopedClient, req WriteRequest[testResourceModel]) (WriteResult[testResourceModel], diag.Diagnostics) {
			createCalled = true
			return testWriteFuncFoundCreate(ctx, client, req)
		},
		Update: func(ctx context.Context, client *clients.ElasticsearchScopedClient, req WriteRequest[testResourceModel]) (WriteResult[testResourceModel], diag.Diagnostics) {
			updateCalled = true
			return testWriteFuncFoundUpdate(ctx, client, req)
		},
	})
	r.client = factory

	plan := makeTestResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "cluster/user1"))
	prior := makeTestResourceState(ctx, t, "cluster/user1")
	resp := resource.UpdateResponse{State: prior}
	req := resource.UpdateRequest{Plan: plan, State: prior}

	r.Update(ctx, req, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.True(t, updateCalled, "Update should invoke the update callback")
	require.False(t, createCalled, "Update must not invoke the create callback")
}

func TestNewElasticsearchResource_Update_nilWriteCallback(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	var nilUpdate WriteFunc[testResourceModel]
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema: getTestResourceSchema,
		Read:   testReadFuncFound,
		Delete: testDeleteFunc,
		Create: testWriteFuncFoundCreate,
		Update: nilUpdate,
	})
	r.client = factory

	plan := makeTestResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "cluster/user1"))
	prior := makeTestResourceState(ctx, t, "cluster/user1")
	resp := resource.UpdateResponse{State: prior}
	req := resource.UpdateRequest{Plan: plan, State: prior}

	r.Update(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Elasticsearch envelope configuration error")
}

func TestNewElasticsearchResource_Write_shortCircuitEmptyWriteID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	writeCalled := false
	writeFnCreate := func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ WriteRequest[testResourceModel]) (WriteResult[testResourceModel], diag.Diagnostics) {
		writeCalled = true
		return WriteResult[testResourceModel]{Model: testResourceModel{}}, nil
	}
	writeFnUpdate := func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ WriteRequest[testResourceModel]) (WriteResult[testResourceModel], diag.Diagnostics) {
		writeCalled = true
		return WriteResult[testResourceModel]{Model: testResourceModel{}}, nil
	}
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema: getTestResourceSchema,
		Read:   testReadFuncFound,
		Delete: testDeleteFunc,
		Create: writeFnCreate,
		Update: writeFnUpdate,
	})
	r.client = factory

	objType := testResourceObjectType()
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                       tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"name":                     tftypes.NewValue(tftypes.String, ""),
		"elasticsearch_connection": tftypes.NewValue(elasticsearchConnectionBlockType(), nil),
	})
	plan := tfsdk.Plan{Raw: objValue, Schema: testResourceSchemaWithConnectionBlock(ctx)}

	t.Run("Create", func(t *testing.T) {
		respState := tfsdk.State{
			Raw:    tftypes.NewValue(objType, nil),
			Schema: testResourceSchemaWithConnectionBlock(ctx),
		}
		resp := resource.CreateResponse{State: respState}
		r.Create(ctx, resource.CreateRequest{Plan: plan}, &resp)
		require.True(t, resp.Diagnostics.HasError())
		require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Invalid resource identifier")
	})

	t.Run("Update", func(t *testing.T) {
		prior := makeTestResourceState(ctx, t, "cluster/user1")
		resp := resource.UpdateResponse{State: prior}
		r.Update(ctx, resource.UpdateRequest{Plan: plan, State: prior}, &resp)
		require.True(t, resp.Diagnostics.HasError())
		require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Invalid resource identifier")
	})

	require.False(t, writeCalled, "write callbacks should not run when write identity is empty")
}

func TestNewElasticsearchResource_Update_shortCircuitUnknownWriteID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	updateCalled := false
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema: getTestResourceSchema,
		Read:   testReadFuncFound,
		Delete: testDeleteFunc,
		Create: testWriteFuncFoundCreate,
		Update: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ WriteRequest[testResourceModel]) (WriteResult[testResourceModel], diag.Diagnostics) {
			updateCalled = true
			return WriteResult[testResourceModel]{Model: testResourceModel{}}, nil
		},
	})
	r.client = factory

	objType := testResourceObjectType()
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                       tftypes.NewValue(tftypes.String, "cluster/user1"),
		"name":                     tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"elasticsearch_connection": tftypes.NewValue(elasticsearchConnectionBlockType(), nil),
	})
	plan := tfsdk.Plan{Raw: objValue, Schema: testResourceSchemaWithConnectionBlock(ctx)}
	prior := makeTestResourceState(ctx, t, "cluster/user1")
	resp := resource.UpdateResponse{State: prior}
	req := resource.UpdateRequest{Plan: plan, State: prior}

	r.Update(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Invalid resource identifier")
	require.False(t, updateCalled, "update callback should not run when write identity is unknown")
}

func TestNewElasticsearchResource_Update_shortCircuitClientError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := nonNilTestFactory()
	updateCalled := false
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema: getTestResourceSchema,
		Read:   testReadFuncFound,
		Delete: testDeleteFunc,
		Create: testWriteFuncFoundCreate,
		Update: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ WriteRequest[testResourceModel]) (WriteResult[testResourceModel], diag.Diagnostics) {
			updateCalled = true
			return WriteResult[testResourceModel]{Model: testResourceModel{}}, nil
		},
	})
	r.client = factory

	plan := makeTestResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "cluster/user1"))
	prior := makeTestResourceState(ctx, t, "cluster/user1")
	resp := resource.UpdateResponse{State: prior}
	req := resource.UpdateRequest{Plan: plan, State: prior}

	r.Update(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Provider not configured")
	require.False(t, updateCalled, "update callback should not run when client resolution fails")
}

func TestNewElasticsearchResource_Update_shortCircuitCallbackError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema: getTestResourceSchema,
		Read:   testReadFuncFound,
		Delete: testDeleteFunc,
		Create: testWriteFuncFoundCreate,
		Update: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ WriteRequest[testResourceModel]) (WriteResult[testResourceModel], diag.Diagnostics) {
			var diags diag.Diagnostics
			diags.AddError("update error", "something went wrong")
			return WriteResult[testResourceModel]{Model: testResourceModel{}}, diags
		},
	})
	r.client = factory

	plan := makeTestResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "cluster/user1"))
	prior := makeTestResourceState(ctx, t, "cluster/user1")
	resp := resource.UpdateResponse{State: prior}
	req := resource.UpdateRequest{Plan: plan, State: prior}

	r.Update(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "update error")
	var after testResourceModel
	require.False(t, resp.State.Get(ctx, &after).HasError())
	require.Equal(t, "cluster/user1", after.ID.ValueString(), "state should not change when update callback returns errors")
}

func TestNewElasticsearchResource_Update_readAfterWriteHappyPath(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema: getTestResourceSchema,
		Read: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, model testResourceModel) (testResourceModel, bool, diag.Diagnostics) {
			model.Name = types.StringValue("from-readfunc")
			return model, true, nil
		},
		Delete: testDeleteFunc,
		Create: testWriteFuncFoundCreate,
		Update: testWriteFuncFoundUpdate,
	})
	r.client = factory

	plan := makeTestResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "cluster/user1"))
	prior := makeTestResourceState(ctx, t, "cluster/user1")
	resp := resource.UpdateResponse{State: prior}
	req := resource.UpdateRequest{Plan: plan, State: prior}

	r.Update(ctx, req, &resp)

	require.False(t, resp.Diagnostics.HasError())
	var result testResourceModel
	diags := resp.State.Get(ctx, &result)
	require.False(t, diags.HasError())
	require.Equal(t, "cluster/user1", result.ID.ValueString())
	require.Equal(t, "from-readfunc", result.Name.ValueString())
}

func TestNewElasticsearchResource_Update_notFoundAfterWrite(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema: getTestResourceSchema,
		Read: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, _ testResourceModel) (testResourceModel, bool, diag.Diagnostics) {
			return testResourceModel{}, false, nil
		},
		Delete: testDeleteFunc,
		Create: testWriteFuncFoundCreate,
		Update: testWriteFuncFoundUpdate,
	})
	r.client = factory

	plan := makeTestResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "cluster/user1"))
	prior := makeTestResourceState(ctx, t, "cluster/user1")
	resp := resource.UpdateResponse{State: prior}
	req := resource.UpdateRequest{Plan: plan, State: prior}

	r.Update(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Resource not found")
	require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), `elasticsearch_test_entity "user1" was not found after write`)
	var after testResourceModel
	require.False(t, resp.State.Get(ctx, &after).HasError())
	require.Equal(t, "cluster/user1", after.ID.ValueString(), "state should not change when resource not found after update")
}

func TestNewElasticsearchResource_Update_readFuncError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema: getTestResourceSchema,
		Read: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, _ testResourceModel) (testResourceModel, bool, diag.Diagnostics) {
			var diags diag.Diagnostics
			diags.AddError("read error", "something went wrong after update")
			return testResourceModel{}, false, diags
		},
		Delete: testDeleteFunc,
		Create: testWriteFuncFoundCreate,
		Update: testWriteFuncFoundUpdate,
	})
	r.client = factory

	plan := makeTestResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "cluster/user1"))
	prior := makeTestResourceState(ctx, t, "cluster/user1")
	resp := resource.UpdateResponse{State: prior}
	req := resource.UpdateRequest{Plan: plan, State: prior}

	r.Update(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "read error")
	var after testResourceModel
	require.False(t, resp.State.Get(ctx, &after).HasError())
	require.Equal(t, "cluster/user1", after.ID.ValueString(), "state should not change when readFunc returns errors after update")
}

func TestNewElasticsearchResource_Update_placeholderCallbackError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	placeholder := PlaceholderElasticsearchWriteCallback[testResourceModel]()
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema: getTestResourceSchema,
		Read:   testReadFuncFound,
		Delete: testDeleteFunc,
		Create: placeholder,
		Update: placeholder,
	})
	r.client = factory

	plan := makeTestResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "cluster/user1"))
	prior := makeTestResourceState(ctx, t, "cluster/user1")
	resp := resource.UpdateResponse{State: prior}
	req := resource.UpdateRequest{Plan: plan, State: prior}

	r.Update(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	err0 := resp.Diagnostics.Errors()[0]
	require.Equal(t, placeholderWriteCallbackSummary, err0.Summary())
	require.Equal(t, placeholderWriteCallbackDetail, err0.Detail())
}

func TestNewElasticsearchResource_CreateAndUpdate_concreteOverridesWin(t *testing.T) {
	t.Parallel()
	r := &overridingEnvelopeTestResource{
		ElasticsearchResource: NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
			Schema: getTestResourceSchema,
			Read:   testReadFuncFound,
			Delete: testDeleteFunc,
			Create: testWriteFuncFoundCreate,
			Update: testWriteFuncFoundUpdate,
		}),
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

type readIDOverrideModel struct {
	testResourceModel
}

func (readIDOverrideModel) GetReadResourceID() string {
	return "stable-read-id"
}

func testDeleteReadOverride(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, _ readIDOverrideModel) diag.Diagnostics {
	return nil
}

func testWriteReadOverrideCreate(
	_ context.Context,
	_ *clients.ElasticsearchScopedClient,
	req WriteRequest[readIDOverrideModel],
) (WriteResult[readIDOverrideModel], diag.Diagnostics) {
	m := req.Plan
	m.ID = types.StringValue("cluster/" + req.WriteID)
	return WriteResult[readIDOverrideModel]{Model: m}, nil
}

func testWriteReadOverrideUpdate(
	_ context.Context,
	_ *clients.ElasticsearchScopedClient,
	req WriteRequest[readIDOverrideModel],
) (WriteResult[readIDOverrideModel], diag.Diagnostics) {
	m := req.Plan
	m.ID = types.StringValue("cluster/" + req.WriteID)
	return WriteResult[readIDOverrideModel]{Model: m}, nil
}

func TestNewElasticsearchResource_Read_usesReadResourceIDFromModel(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	var captured string
	readFn := func(_ context.Context, _ *clients.ElasticsearchScopedClient, resourceID string, model readIDOverrideModel) (readIDOverrideModel, bool, diag.Diagnostics) {
		captured = resourceID
		return model, true, nil
	}
	r := NewElasticsearchResource[readIDOverrideModel]("test_entity", ElasticsearchResourceOptions[readIDOverrideModel]{
		Schema: getTestResourceSchema,
		Read:   readFn,
		Delete: testDeleteReadOverride,
		Create: testWriteReadOverrideCreate,
		Update: testWriteReadOverrideUpdate,
	})
	r.client = factory

	state := makeTestResourceState(ctx, t, "cluster/user1")
	req := resource.ReadRequest{State: state}
	resp := resource.ReadResponse{State: state}

	r.Read(ctx, req, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.Equal(t, "stable-read-id", captured)
}

func TestNewElasticsearchResource_Create_readAfterWriteUsesReadResourceIDFromWrittenModel(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	var captured string
	readFn := func(_ context.Context, _ *clients.ElasticsearchScopedClient, resourceID string, model readIDOverrideModel) (readIDOverrideModel, bool, diag.Diagnostics) {
		captured = resourceID
		model.Name = types.StringValue(model.Name.ValueString() + "-refreshed")
		return model, true, nil
	}
	r := NewElasticsearchResource[readIDOverrideModel]("test_entity", ElasticsearchResourceOptions[readIDOverrideModel]{
		Schema: getTestResourceSchema,
		Read:   readFn,
		Delete: testDeleteReadOverride,
		Create: testWriteReadOverrideCreate,
		Update: testWriteReadOverrideUpdate,
	})
	r.client = factory

	plan := makeTestResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue))
	objType := testResourceObjectType()
	respState := tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: testResourceSchemaWithConnectionBlock(ctx)}
	resp := resource.CreateResponse{State: respState}

	r.Create(ctx, resource.CreateRequest{Plan: plan}, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.Equal(t, "stable-read-id", captured)
	var result readIDOverrideModel
	require.False(t, resp.State.Get(ctx, &result).HasError())
	require.Equal(t, "user1-refreshed", result.Name.ValueString())
}

type versionReqDiagModel struct {
	testResourceModel
}

func (versionReqDiagModel) GetVersionRequirements() ([]VersionRequirement, diag.Diagnostics) {
	var diags diag.Diagnostics
	diags.AddError("version lookup failed", "boom")
	return nil, diags
}

func TestNewElasticsearchResource_Read_shortCircuitsWhenVersionRequirementsDiag(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	readCalled := false
	readFn := func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, model versionReqDiagModel) (versionReqDiagModel, bool, diag.Diagnostics) {
		readCalled = true
		return model, true, nil
	}
	r := NewElasticsearchResource[versionReqDiagModel]("test_entity", ElasticsearchResourceOptions[versionReqDiagModel]{
		Schema: getTestResourceSchema,
		Read:   readFn,
		Delete: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, _ versionReqDiagModel) diag.Diagnostics {
			return nil
		},
		Create: func(_ context.Context, _ *clients.ElasticsearchScopedClient, req WriteRequest[versionReqDiagModel]) (WriteResult[versionReqDiagModel], diag.Diagnostics) {
			m := req.Plan
			m.ID = types.StringValue("cluster/" + req.WriteID)
			return WriteResult[versionReqDiagModel]{Model: m}, nil
		},
		Update: func(_ context.Context, _ *clients.ElasticsearchScopedClient, req WriteRequest[versionReqDiagModel]) (WriteResult[versionReqDiagModel], diag.Diagnostics) {
			m := req.Plan
			m.ID = types.StringValue("cluster/" + req.WriteID)
			return WriteResult[versionReqDiagModel]{Model: m}, nil
		},
	})
	r.client = factory

	state := makeTestResourceState(ctx, t, "cluster/user1")
	req := resource.ReadRequest{State: state}
	resp := resource.ReadResponse{State: state}

	r.Read(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "version lookup failed")
	require.False(t, readCalled)
}

func TestNewElasticsearchResource_Create_shortCircuitsWhenVersionRequirementsDiag(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	createCalled := false
	r := NewElasticsearchResource[versionReqDiagModel]("test_entity", ElasticsearchResourceOptions[versionReqDiagModel]{
		Schema: getTestResourceSchema,
		Read: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, model versionReqDiagModel) (versionReqDiagModel, bool, diag.Diagnostics) {
			return model, true, nil
		},
		Delete: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, _ versionReqDiagModel) diag.Diagnostics {
			return nil
		},
		Create: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ WriteRequest[versionReqDiagModel]) (WriteResult[versionReqDiagModel], diag.Diagnostics) {
			createCalled = true
			var zero versionReqDiagModel
			return WriteResult[versionReqDiagModel]{Model: zero}, nil
		},
		Update: func(_ context.Context, _ *clients.ElasticsearchScopedClient, req WriteRequest[versionReqDiagModel]) (WriteResult[versionReqDiagModel], diag.Diagnostics) {
			m := req.Plan
			return WriteResult[versionReqDiagModel]{Model: m}, nil
		},
	})
	r.client = factory

	plan := makeTestResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue))
	objType := testResourceObjectType()
	respState := tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: testResourceSchemaWithConnectionBlock(ctx)}
	resp := resource.CreateResponse{State: respState}

	r.Create(ctx, resource.CreateRequest{Plan: plan}, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.False(t, createCalled)
}

func TestNewElasticsearchResource_Read_invokesPostReadAfterSuccessfulStateSet(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	postCalled := false
	postRead := func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ testResourceModel, _ any) diag.Diagnostics {
		postCalled = true
		return nil
	}
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema:   getTestResourceSchema,
		Read:     testReadFuncFound,
		Delete:   testDeleteFunc,
		Create:   testWriteFuncFoundCreate,
		Update:   testWriteFuncFoundUpdate,
		PostRead: postRead,
	})
	r.client = factory

	state := makeTestResourceState(ctx, t, "cluster/user1")
	req := resource.ReadRequest{State: state}
	resp := resource.ReadResponse{State: state}

	r.Read(ctx, req, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.True(t, postCalled)
}

func TestNewElasticsearchResource_Read_skipsPostReadWhenNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	postCalled := false
	postRead := func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ testResourceModel, _ any) diag.Diagnostics {
		postCalled = true
		return nil
	}
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema: getTestResourceSchema,
		Read: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, _ testResourceModel) (testResourceModel, bool, diag.Diagnostics) {
			return testResourceModel{}, false, nil
		},
		Delete:   testDeleteFunc,
		Create:   testWriteFuncFoundCreate,
		Update:   testWriteFuncFoundUpdate,
		PostRead: postRead,
	})
	r.client = factory

	state := makeTestResourceState(ctx, t, "cluster/user1")
	req := resource.ReadRequest{State: state}
	resp := resource.ReadResponse{State: state}

	r.Read(ctx, req, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.False(t, postCalled)
}

func TestNewElasticsearchResource_Update_shortCircuitsWhenVersionRequirementsDiag(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	updateCalled := false
	readFn := func(
		_ context.Context,
		_ *clients.ElasticsearchScopedClient,
		_ string,
		model versionReqDiagModel,
	) (versionReqDiagModel, bool, diag.Diagnostics) {
		return model, true, nil
	}
	createFn := func(
		_ context.Context,
		_ *clients.ElasticsearchScopedClient,
		_ WriteRequest[versionReqDiagModel],
	) (WriteResult[versionReqDiagModel], diag.Diagnostics) {
		var zero versionReqDiagModel
		return WriteResult[versionReqDiagModel]{Model: zero}, nil
	}
	updateFn := func(
		_ context.Context,
		_ *clients.ElasticsearchScopedClient,
		_ WriteRequest[versionReqDiagModel],
	) (WriteResult[versionReqDiagModel], diag.Diagnostics) {
		updateCalled = true
		var zero versionReqDiagModel
		return WriteResult[versionReqDiagModel]{Model: zero}, nil
	}
	r := NewElasticsearchResource[versionReqDiagModel]("test_entity", ElasticsearchResourceOptions[versionReqDiagModel]{
		Schema: getTestResourceSchema,
		Read:   readFn,
		Delete: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, _ versionReqDiagModel) diag.Diagnostics {
			return nil
		},
		Create: createFn,
		Update: updateFn,
	})
	r.client = factory

	plan := makeTestResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "cluster/user1"))
	prior := makeTestResourceState(ctx, t, "cluster/user1")
	cfg := tfsdk.Config(plan)
	resp := resource.UpdateResponse{State: prior}
	req := resource.UpdateRequest{Plan: plan, State: prior, Config: cfg}

	r.Update(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "version lookup failed")
	require.False(t, updateCalled)
}

func TestNewElasticsearchResource_Update_callbackReceivesPlanPriorConfigAndWriteID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)

	plan := makeTestResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "cluster/user1"))
	prior := makeTestResourceState(ctx, t, "cluster/user1")
	cfg := tfsdk.Config(plan)

	updateFn := func(
		_ context.Context,
		_ *clients.ElasticsearchScopedClient,
		req WriteRequest[testResourceModel],
	) (WriteResult[testResourceModel], diag.Diagnostics) {
		require.Equal(t, "user1", req.WriteID)
		require.Equal(t, "user1", req.Plan.Name.ValueString())
		require.NotNil(t, req.Prior, "Update SHALL receive a non-nil Prior pointer")
		require.Equal(t, "cluster/user1", req.Prior.ID.ValueString())
		require.True(t, req.Config.Raw.Equal(plan.Raw))
		require.Equal(t, plan.Schema, req.Config.Schema)

		model := req.Plan
		model.ID = types.StringValue("cluster/" + req.WriteID)
		return WriteResult[testResourceModel]{Model: model}, nil
	}

	opts := defaultTestElasticsearchResourceOptions()
	opts.Update = updateFn
	r := NewElasticsearchResource[testResourceModel]("test_entity", opts)
	r.client = factory

	resp := resource.UpdateResponse{State: prior}
	req := resource.UpdateRequest{Plan: plan, State: prior, Config: cfg}

	r.Update(ctx, req, &resp)

	require.False(t, resp.Diagnostics.HasError())
}

// TestNewElasticsearchResource_Create_callbackReceivesNilPrior verifies that
// Create invokes the WriteFunc with req.Prior == nil so a single shared
// WriteFunc[T] can dispatch on Prior == nil to detect the create branch.
func TestNewElasticsearchResource_Create_callbackReceivesNilPrior(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)

	createFn := func(
		_ context.Context,
		_ *clients.ElasticsearchScopedClient,
		req WriteRequest[testResourceModel],
	) (WriteResult[testResourceModel], diag.Diagnostics) {
		require.Nil(t, req.Prior, "Create SHALL receive a nil Prior pointer")
		require.Equal(t, "user1", req.WriteID)
		require.Equal(t, "user1", req.Plan.Name.ValueString())
		model := req.Plan
		model.ID = types.StringValue("cluster/" + req.WriteID)
		return WriteResult[testResourceModel]{Model: model}, nil
	}

	opts := defaultTestElasticsearchResourceOptions()
	opts.Create = createFn
	r := NewElasticsearchResource[testResourceModel]("test_entity", opts)
	r.client = factory

	plan := makeTestResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue))
	objType := testResourceObjectType()
	respState := tfsdk.State{
		Raw:    tftypes.NewValue(objType, nil),
		Schema: testResourceSchemaWithConnectionBlock(ctx),
	}
	resp := resource.CreateResponse{State: respState}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, &resp)

	require.False(t, resp.Diagnostics.HasError())
}

// TestNewElasticsearchResource_SingleWriteFuncServesCreateAndUpdate verifies a
// concrete resource may wire the same WriteFunc[T] into both Create and Update
// slots, with the function distinguishing the two paths via req.Prior == nil.
func TestNewElasticsearchResource_SingleWriteFuncServesCreateAndUpdate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)

	var sawCreate, sawUpdate bool
	shared := func(
		_ context.Context,
		_ *clients.ElasticsearchScopedClient,
		req WriteRequest[testResourceModel],
	) (WriteResult[testResourceModel], diag.Diagnostics) {
		if req.Prior == nil {
			sawCreate = true
		} else {
			sawUpdate = true
		}
		model := req.Plan
		model.ID = types.StringValue("cluster/" + req.WriteID)
		return WriteResult[testResourceModel]{Model: model}, nil
	}

	opts := defaultTestElasticsearchResourceOptions()
	opts.Create = shared
	opts.Update = shared
	r := NewElasticsearchResource[testResourceModel]("test_entity", opts)
	r.client = factory

	createPlan := makeTestResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue))
	objType := testResourceObjectType()
	createResp := resource.CreateResponse{State: tfsdk.State{
		Raw:    tftypes.NewValue(objType, nil),
		Schema: testResourceSchemaWithConnectionBlock(ctx),
	}}
	r.Create(ctx, resource.CreateRequest{Plan: createPlan}, &createResp)
	require.False(t, createResp.Diagnostics.HasError())

	updatePlan := makeTestResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "cluster/user1"))
	priorState := makeTestResourceState(ctx, t, "cluster/user1")
	updateResp := resource.UpdateResponse{State: priorState}
	r.Update(ctx, resource.UpdateRequest{Plan: updatePlan, State: priorState, Config: tfsdk.Config(updatePlan)}, &updateResp)
	require.False(t, updateResp.Diagnostics.HasError())

	require.True(t, sawCreate, "shared WriteFunc was not invoked from Create")
	require.True(t, sawUpdate, "shared WriteFunc was not invoked from Update")
}

func TestNewElasticsearchResource_Read_skipsPostReadWhenReadFuncError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	postCalled := false
	postRead := func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ testResourceModel, _ any) diag.Diagnostics {
		postCalled = true
		return nil
	}
	readFn := func(
		_ context.Context,
		_ *clients.ElasticsearchScopedClient,
		_ string,
		_ testResourceModel,
	) (testResourceModel, bool, diag.Diagnostics) {
		var d diag.Diagnostics
		d.AddError("read failed", "boom")
		return testResourceModel{}, false, d
	}
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema:   getTestResourceSchema,
		Read:     readFn,
		Delete:   testDeleteFunc,
		Create:   testWriteFuncFoundCreate,
		Update:   testWriteFuncFoundUpdate,
		PostRead: postRead,
	})
	r.client = factory

	state := makeTestResourceState(ctx, t, "cluster/user1")
	req := resource.ReadRequest{State: state}
	resp := resource.ReadResponse{State: state}

	r.Read(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.False(t, postCalled)
}

func TestNewElasticsearchResource_Read_skipsPostReadWhenStateSetFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	postCalled := false
	postRead := func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ testResourceModel, _ any) diag.Diagnostics {
		postCalled = true
		return nil
	}
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema:   getTestResourceSchema,
		Read:     testReadFuncFound,
		Delete:   testDeleteFunc,
		Create:   testWriteFuncFoundCreate,
		Update:   testWriteFuncFoundUpdate,
		PostRead: postRead,
	})
	r.client = factory

	goodState := makeTestResourceState(ctx, t, "cluster/user1")
	req := resource.ReadRequest{State: goodState}
	badOut := tfsdk.State{
		Raw:    goodState.Raw,
		Schema: getTestResourceSchema(context.Background()),
	}
	resp := resource.ReadResponse{State: badOut}

	r.Read(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.False(t, postCalled)
}

func TestNewElasticsearchResource_Read_postReadReceivesFrameworkPrivateHandle(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)

	var captured any
	postRead := func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ testResourceModel, priv any) diag.Diagnostics {
		captured = priv
		return nil
	}
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema:   getTestResourceSchema,
		Read:     testReadFuncFound,
		Delete:   testDeleteFunc,
		Create:   testWriteFuncFoundCreate,
		Update:   testWriteFuncFoundUpdate,
		PostRead: postRead,
	})
	r.client = factory

	state := makeTestResourceState(ctx, t, "cluster/user1")
	req := resource.ReadRequest{State: state}
	resp := resource.ReadResponse{State: state}

	r.Read(ctx, req, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.Equal(t, resp.Private, captured)
}

func TestNewElasticsearchResource_Create_invokesPostReadAfterReadAfterWrite(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	postCalls := 0
	postRead := func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ testResourceModel, _ any) diag.Diagnostics {
		postCalls++
		return nil
	}
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema:   getTestResourceSchema,
		Read:     testReadFuncDistinguishing,
		Delete:   testDeleteFunc,
		Create:   testWriteFuncFoundCreate,
		Update:   testWriteFuncFoundUpdate,
		PostRead: postRead,
	})
	r.client = factory

	plan := makeTestResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue))
	objType := testResourceObjectType()
	respState := tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: testResourceSchemaWithConnectionBlock(ctx)}
	resp := resource.CreateResponse{State: respState}

	r.Create(ctx, resource.CreateRequest{Plan: plan}, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.Equal(t, 1, postCalls)
}

func TestNewElasticsearchResource_Update_invokesPostReadAfterReadAfterWrite(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	postCalls := 0
	postRead := func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ testResourceModel, _ any) diag.Diagnostics {
		postCalls++
		return nil
	}
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema:   getTestResourceSchema,
		Read:     testReadFuncDistinguishing,
		Delete:   testDeleteFunc,
		Create:   testWriteFuncFoundCreate,
		Update:   testWriteFuncFoundUpdate,
		PostRead: postRead,
	})
	r.client = factory

	plan := makeTestResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "cluster/user1"))
	prior := makeTestResourceState(ctx, t, "cluster/user1")
	resp := resource.UpdateResponse{State: prior}
	req := resource.UpdateRequest{Plan: plan, State: prior}

	r.Update(ctx, req, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.Equal(t, 1, postCalls)
}

func TestNewElasticsearchResource_Read_nilReadCallbackConfigurationError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	opts := defaultTestElasticsearchResourceOptions()
	opts.Read = nil
	r := NewElasticsearchResource[testResourceModel]("test_entity", opts)
	r.client = factory

	state := makeTestResourceState(ctx, t, "cluster/user1")
	resp := resource.ReadResponse{State: state}
	r.Read(ctx, resource.ReadRequest{State: state}, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "configuration error")
}

func TestNewElasticsearchResource_Create_nilReadCallbackConfigurationError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	opts := defaultTestElasticsearchResourceOptions()
	opts.Read = nil
	r := NewElasticsearchResource[testResourceModel]("test_entity", opts)
	r.client = factory

	plan := makeTestResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue))
	objType := testResourceObjectType()
	respState := tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: testResourceSchemaWithConnectionBlock(ctx)}
	resp := resource.CreateResponse{State: respState}

	r.Create(ctx, resource.CreateRequest{Plan: plan}, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "configuration error")
}

func TestNewElasticsearchResource_Delete_nilDeleteCallbackConfigurationError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	opts := defaultTestElasticsearchResourceOptions()
	opts.Delete = nil
	r := NewElasticsearchResource[testResourceModel]("test_entity", opts)
	r.client = factory

	state := makeTestResourceState(ctx, t, "cluster/user1")
	resp := resource.DeleteResponse{State: state}
	r.Delete(ctx, resource.DeleteRequest{State: state}, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "configuration error")
}
