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
	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

// testKibanaResourceModel satisfies KibanaResourceModel for envelope tests.
// It is a user-ID variant: GetResourceID returns the user-specified name.
type testKibanaResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	SpaceID          types.String `tfsdk:"space_id"`
	KibanaConnection types.List   `tfsdk:"kibana_connection"`
}

func (m testKibanaResourceModel) GetID() types.String {
	return m.ID
}

func (m testKibanaResourceModel) GetResourceID() types.String {
	return m.Name
}

func (m testKibanaResourceModel) GetSpaceID() types.String {
	return m.SpaceID
}

func (m testKibanaResourceModel) GetKibanaConnection() types.List {
	return m.KibanaConnection
}

func getTestKibanaResourceSchema(_ context.Context) rschema.Schema {
	return rschema.Schema{
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed: true,
			},
			"name": rschema.StringAttribute{
				Optional: true,
			},
			"space_id": rschema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func testKibanaReadFuncFound(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, model testKibanaResourceModel) (testKibanaResourceModel, bool, diag.Diagnostics) {
	return model, true, nil
}

func testKibanaDeleteFunc(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModel) diag.Diagnostics {
	return nil
}

func testKibanaCreateFuncFound(_ context.Context, _ *clients.KibanaScopedClient, _ string, plan testKibanaResourceModel) (testKibanaResourceModel, diag.Diagnostics) {
	plan.ID = types.StringValue(plan.GetSpaceID().ValueString() + "/" + plan.GetResourceID().ValueString())
	return plan, nil
}

func testKibanaUpdateFuncFound(
	_ context.Context,
	_ *clients.KibanaScopedClient,
	resourceID string,
	spaceID string,
	plan testKibanaResourceModel,
	_ testKibanaResourceModel,
) (testKibanaResourceModel, diag.Diagnostics) {
	plan.ID = types.StringValue(spaceID + "/" + resourceID)
	return plan, nil
}

func testKibanaResourceObjectType() tftypes.Type {
	return tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":                tftypes.String,
			"name":              tftypes.String,
			"space_id":          tftypes.String,
			"kibana_connection": kibanaConnectionBlockType(),
		},
	}
}

func testKibanaResourceSchemaWithConnectionBlock(ctx context.Context) rschema.Schema {
	s := getTestKibanaResourceSchema(ctx)
	s.Blocks = map[string]rschema.Block{
		"kibana_connection": providerschema.GetKbFWConnectionBlock(),
	}
	return s
}

func makeTestKibanaResourceCreatePlan(ctx context.Context, t *testing.T, idValue tftypes.Value, spaceIDValue tftypes.Value) tfsdk.Plan {
	t.Helper()
	const resourceName = "my-resource"
	objType := testKibanaResourceObjectType()
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                idValue,
		"name":              tftypes.NewValue(tftypes.String, resourceName),
		"space_id":          spaceIDValue,
		"kibana_connection": tftypes.NewValue(kibanaConnectionBlockType(), nil),
	})
	return tfsdk.Plan{
		Raw:    objValue,
		Schema: testKibanaResourceSchemaWithConnectionBlock(ctx),
	}
}

func makeTestKibanaResourceState(ctx context.Context, t *testing.T, id string) tfsdk.State {
	t.Helper()
	connBlockType := kibanaConnectionBlockType()
	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":                tftypes.String,
			"name":              tftypes.String,
			"space_id":          tftypes.String,
			"kibana_connection": connBlockType,
		},
	}
	name := id
	// If id is a composite ID, extract the resourceID for the name field.
	if compID, _ := clients.CompositeIDFromStr(id); compID != nil {
		name = compID.ResourceID
	}
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, id),
		"name":              tftypes.NewValue(tftypes.String, name),
		"space_id":          tftypes.NewValue(tftypes.String, "default"),
		"kibana_connection": tftypes.NewValue(connBlockType, nil),
	})

	return tfsdk.State{
		Raw:    objValue,
		Schema: testKibanaResourceSchemaWithConnectionBlock(ctx),
	}
}

func newTestKibanaResourceEnvelopeWithFactory(t *testing.T, factory *clients.ProviderClientFactory) *KibanaResource[testKibanaResourceModel] {
	t.Helper()
	r := NewKibanaResource[testKibanaResourceModel](
		ComponentKibana,
		"test_entity",
		getTestKibanaResourceSchema,
		testKibanaReadFuncFound,
		testKibanaDeleteFunc,
		testKibanaCreateFuncFound,
		testKibanaUpdateFuncFound,
	)
	r.client = factory
	return r
}

// =============================================================================
// Subtask 2.2: Type assertions
// =============================================================================

func TestNewKibanaResource_typeAssertions(t *testing.T) {
	t.Parallel()
	r := NewKibanaResource[testKibanaResourceModel](
		ComponentKibana,
		"test_entity",
		getTestKibanaResourceSchema,
		testKibanaReadFuncFound,
		testKibanaDeleteFunc,
		testKibanaCreateFuncFound,
		testKibanaUpdateFuncFound,
	)
	require.NotNil(t, r)
	require.Implements(t, (*resource.Resource)(nil), r)
	require.Implements(t, (*resource.ResourceWithConfigure)(nil), r)
	_, implementsImport := any(r).(resource.ResourceWithImportState)
	require.False(t, implementsImport, "envelope must not implement ImportState; concrete resources add it when needed")
}

// =============================================================================
// Subtask 2.3: Metadata
// =============================================================================

func TestNewKibanaResource_Metadata(t *testing.T) {
	t.Parallel()
	r := NewKibanaResource[testKibanaResourceModel](
		ComponentKibana,
		"test_entity",
		getTestKibanaResourceSchema,
		testKibanaReadFuncFound,
		testKibanaDeleteFunc,
		testKibanaCreateFuncFound,
		testKibanaUpdateFuncFound,
	)

	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: testProviderTypeName}, &resp)

	require.Equal(t, "elasticstack_kibana_test_entity", resp.TypeName)
}

// =============================================================================
// Subtask 2.4: Schema injection
// =============================================================================

func TestNewKibanaResource_schemaInjection(t *testing.T) {
	t.Parallel()
	r := NewKibanaResource[testKibanaResourceModel](
		ComponentKibana,
		"test_entity",
		getTestKibanaResourceSchema,
		testKibanaReadFuncFound,
		testKibanaDeleteFunc,
		testKibanaCreateFuncFound,
		testKibanaUpdateFuncFound,
	)

	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Schema.Blocks, "kibana_connection")
	require.Contains(t, resp.Schema.Attributes, "id")
}

func TestNewKibanaResource_schemaDefensiveClone(t *testing.T) {
	t.Parallel()
	originalSchema := getTestKibanaResourceSchema(context.Background())
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", func(_ context.Context) rschema.Schema {
		return originalSchema
	}, testKibanaReadFuncFound, testKibanaDeleteFunc, testKibanaCreateFuncFound, testKibanaUpdateFuncFound)

	var resp1 resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp1)
	require.False(t, resp1.Diagnostics.HasError())

	var resp2 resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp2)
	require.False(t, resp2.Diagnostics.HasError())

	require.Nil(t, originalSchema.Blocks)
}

// =============================================================================
// Subtask 2.5: Configure
// =============================================================================

func TestNewKibanaResource_Configure(t *testing.T) {
	ctx := context.Background()

	t.Run("nil_provider_data", func(t *testing.T) {
		t.Parallel()
		r := NewKibanaResource[testKibanaResourceModel](
			ComponentKibana,
			"test_entity",
			getTestKibanaResourceSchema,
			testKibanaReadFuncFound,
			testKibanaDeleteFunc,
			testKibanaCreateFuncFound,
			testKibanaUpdateFuncFound,
		)
		var resp resource.ConfigureResponse
		r.Configure(ctx, resource.ConfigureRequest{ProviderData: nil}, &resp)
		require.False(t, resp.Diagnostics.HasError())
	})

	t.Run("valid_factory", func(t *testing.T) {
		t.Parallel()
		r := NewKibanaResource[testKibanaResourceModel](
			ComponentKibana,
			"test_entity",
			getTestKibanaResourceSchema,
			testKibanaReadFuncFound,
			testKibanaDeleteFunc,
			testKibanaCreateFuncFound,
			testKibanaUpdateFuncFound,
		)
		f := nonNilTestFactory()
		var resp resource.ConfigureResponse
		r.Configure(ctx, resource.ConfigureRequest{ProviderData: f}, &resp)
		require.False(t, resp.Diagnostics.HasError())
	})

	t.Run("invalid_provider_data", func(t *testing.T) {
		t.Parallel()
		r := NewKibanaResource[testKibanaResourceModel](
			ComponentKibana,
			"test_entity",
			getTestKibanaResourceSchema,
			testKibanaReadFuncFound,
			testKibanaDeleteFunc,
			testKibanaCreateFuncFound,
			testKibanaUpdateFuncFound,
		)
		var resp resource.ConfigureResponse
		r.Configure(ctx, resource.ConfigureRequest{ProviderData: "wrong-type"}, &resp)
		require.True(t, resp.Diagnostics.HasError())
	})
}

// =============================================================================
// Subtask 2.6: Create happy path
// =============================================================================

func TestNewKibanaResource_Create_happyPath(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	r := newTestKibanaResourceEnvelopeWithFactory(t, factory)

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tftypes.NewValue(tftypes.String, "default"))
	objType := testKibanaResourceObjectType()
	respState := tfsdk.State{
		Raw:    tftypes.NewValue(objType, nil),
		Schema: testKibanaResourceSchemaWithConnectionBlock(ctx),
	}
	req := resource.CreateRequest{Plan: plan}
	resp := resource.CreateResponse{State: respState}

	r.Create(ctx, req, &resp)

	require.False(t, resp.Diagnostics.HasError())
	var result testKibanaResourceModel
	diags := resp.State.Get(ctx, &result)
	require.False(t, diags.HasError())
	require.Equal(t, "default/my-resource", result.ID.ValueString())
	require.Equal(t, "my-resource", result.Name.ValueString())
	require.Equal(t, "default", result.SpaceID.ValueString())
}

// =============================================================================
// Subtask 2.7: Create short-circuits
// =============================================================================

func TestNewKibanaResource_Create_shortCircuitSpaceIDUnknown(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	createCalled := false
	r := NewKibanaResource[testKibanaResourceModel](
		ComponentKibana,
		"test_entity",
		getTestKibanaResourceSchema,
		testKibanaReadFuncFound,
		testKibanaDeleteFunc,
		func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ testKibanaResourceModel) (testKibanaResourceModel, diag.Diagnostics) {
			createCalled = true
			return testKibanaResourceModel{}, nil
		},
		testKibanaUpdateFuncFound,
	)
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tftypes.NewValue(tftypes.String, tftypes.UnknownValue))
	objType := testKibanaResourceObjectType()
	respState := tfsdk.State{
		Raw:    tftypes.NewValue(objType, nil),
		Schema: testKibanaResourceSchemaWithConnectionBlock(ctx),
	}
	req := resource.CreateRequest{Plan: plan}
	resp := resource.CreateResponse{State: respState}

	r.Create(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Invalid space identifier")
	require.False(t, createCalled, "create callback should not run when spaceID is unknown")
}

func TestNewKibanaResource_Create_shortCircuitSpaceIDNull(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	createCalled := false
	r := NewKibanaResource[testKibanaResourceModel](
		ComponentKibana,
		"test_entity",
		getTestKibanaResourceSchema,
		testKibanaReadFuncFound,
		testKibanaDeleteFunc,
		func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ testKibanaResourceModel) (testKibanaResourceModel, diag.Diagnostics) {
			createCalled = true
			return testKibanaResourceModel{}, nil
		},
		testKibanaUpdateFuncFound,
	)
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tftypes.NewValue(tftypes.String, nil))
	objType := testKibanaResourceObjectType()
	respState := tfsdk.State{
		Raw:    tftypes.NewValue(objType, nil),
		Schema: testKibanaResourceSchemaWithConnectionBlock(ctx),
	}
	req := resource.CreateRequest{Plan: plan}
	resp := resource.CreateResponse{State: respState}

	r.Create(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Invalid space identifier")
	require.False(t, createCalled, "create callback should not run when spaceID is null")
}

func TestNewKibanaResource_Create_shortCircuitSpaceIDEmpty(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	createCalled := false
	r := NewKibanaResource[testKibanaResourceModel](
		ComponentKibana,
		"test_entity",
		getTestKibanaResourceSchema,
		testKibanaReadFuncFound,
		testKibanaDeleteFunc,
		func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ testKibanaResourceModel) (testKibanaResourceModel, diag.Diagnostics) {
			createCalled = true
			return testKibanaResourceModel{}, nil
		},
		testKibanaUpdateFuncFound,
	)
	r.client = factory

	objType := testKibanaResourceObjectType()
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"name":              tftypes.NewValue(tftypes.String, "my-resource"),
		"space_id":          tftypes.NewValue(tftypes.String, ""),
		"kibana_connection": tftypes.NewValue(kibanaConnectionBlockType(), nil),
	})
	plan := tfsdk.Plan{Raw: objValue, Schema: testKibanaResourceSchemaWithConnectionBlock(ctx)}
	respState := tfsdk.State{
		Raw:    tftypes.NewValue(objType, nil),
		Schema: testKibanaResourceSchemaWithConnectionBlock(ctx),
	}
	req := resource.CreateRequest{Plan: plan}
	resp := resource.CreateResponse{State: respState}

	r.Create(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Invalid space identifier")
	require.False(t, createCalled, "create callback should not run when spaceID is empty")
}

func TestNewKibanaResource_Create_shortCircuitClientError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := nonNilTestFactory()
	createCalled := false
	r := NewKibanaResource[testKibanaResourceModel](
		ComponentKibana,
		"test_entity",
		getTestKibanaResourceSchema,
		testKibanaReadFuncFound,
		testKibanaDeleteFunc,
		func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ testKibanaResourceModel) (testKibanaResourceModel, diag.Diagnostics) {
			createCalled = true
			return testKibanaResourceModel{}, nil
		},
		testKibanaUpdateFuncFound,
	)
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tftypes.NewValue(tftypes.String, "default"))
	objType := testKibanaResourceObjectType()
	respState := tfsdk.State{
		Raw:    tftypes.NewValue(objType, nil),
		Schema: testKibanaResourceSchemaWithConnectionBlock(ctx),
	}
	req := resource.CreateRequest{Plan: plan}
	resp := resource.CreateResponse{State: respState}

	r.Create(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Provider not configured")
	require.False(t, createCalled, "create callback should not run when client resolution fails")
}

func TestNewKibanaResource_Create_shortCircuitCallbackError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	r := NewKibanaResource[testKibanaResourceModel](
		ComponentKibana,
		"test_entity",
		getTestKibanaResourceSchema,
		testKibanaReadFuncFound,
		testKibanaDeleteFunc,
		func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ testKibanaResourceModel) (testKibanaResourceModel, diag.Diagnostics) {
			var diags diag.Diagnostics
			diags.AddError("create error", "something went wrong")
			return testKibanaResourceModel{}, diags
		},
		testKibanaUpdateFuncFound,
	)
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tftypes.NewValue(tftypes.String, "default"))
	objType := testKibanaResourceObjectType()
	respState := tfsdk.State{
		Raw:    tftypes.NewValue(objType, nil),
		Schema: testKibanaResourceSchemaWithConnectionBlock(ctx),
	}
	req := resource.CreateRequest{Plan: plan}
	resp := resource.CreateResponse{State: respState}

	r.Create(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "create error")
	require.True(t, resp.State.Raw.IsNull(), "state should not be mutated when create callback fails")
}

func TestNewKibanaResource_Create_shortCircuitPlanGetError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	createCalled := false
	r := NewKibanaResource[testKibanaResourceModel](
		ComponentKibana,
		"test_entity",
		getTestKibanaResourceSchema,
		testKibanaReadFuncFound,
		testKibanaDeleteFunc,
		func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ testKibanaResourceModel) (testKibanaResourceModel, diag.Diagnostics) {
			createCalled = true
			return testKibanaResourceModel{}, nil
		},
		testKibanaUpdateFuncFound,
	)
	r.client = factory

	objType := testKibanaResourceObjectType()
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"name":              tftypes.NewValue(tftypes.String, "my-resource"),
		"space_id":          tftypes.NewValue(tftypes.String, "default"),
		"kibana_connection": tftypes.NewValue(kibanaConnectionBlockType(), nil),
	})
	badSchema := getTestKibanaResourceSchema(context.Background())
	plan := tfsdk.Plan{Raw: objValue, Schema: badSchema}
	respState := tfsdk.State{
		Raw:    tftypes.NewValue(objType, nil),
		Schema: testKibanaResourceSchemaWithConnectionBlock(ctx),
	}
	req := resource.CreateRequest{Plan: plan}
	resp := resource.CreateResponse{State: respState}

	r.Create(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.False(t, createCalled, "create callback should not run when plan.Get fails")
}

// =============================================================================
// Subtask 2.8: Create with nil and placeholder write callbacks
// =============================================================================

func TestNewKibanaResource_Create_nilWriteCallback(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	var nilCreate KibanaCreateFunc[testKibanaResourceModel]
	r := NewKibanaResource[testKibanaResourceModel](
		ComponentKibana,
		"test_entity",
		getTestKibanaResourceSchema,
		testKibanaReadFuncFound,
		testKibanaDeleteFunc,
		nilCreate,
		testKibanaUpdateFuncFound,
	)
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tftypes.NewValue(tftypes.String, "default"))
	objType := testKibanaResourceObjectType()
	respState := tfsdk.State{
		Raw:    tftypes.NewValue(objType, nil),
		Schema: testKibanaResourceSchemaWithConnectionBlock(ctx),
	}
	req := resource.CreateRequest{Plan: plan}
	resp := resource.CreateResponse{State: respState}

	r.Create(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Kibana envelope configuration error")
}

func TestNewKibanaResource_Create_placeholderCallbackError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	createFn, updateFn := PlaceholderKibanaWriteCallbacks[testKibanaResourceModel]()
	r := NewKibanaResource[testKibanaResourceModel](
		ComponentKibana,
		"test_entity",
		getTestKibanaResourceSchema,
		testKibanaReadFuncFound,
		testKibanaDeleteFunc,
		createFn,
		updateFn,
	)
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tftypes.NewValue(tftypes.String, "default"))
	objType := testKibanaResourceObjectType()
	respState := tfsdk.State{
		Raw:    tftypes.NewValue(objType, nil),
		Schema: testKibanaResourceSchemaWithConnectionBlock(ctx),
	}
	req := resource.CreateRequest{Plan: plan}
	resp := resource.CreateResponse{State: respState}

	r.Create(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	err0 := resp.Diagnostics.Errors()[0]
	require.Equal(t, placeholderKibanaWriteCallbackSummary, err0.Summary())
	require.Equal(t, placeholderKibanaWriteCallbackDetail, err0.Detail())
}

func TestNewKibanaResource_Create_nilCallbackPrecedesOtherPreludeErrors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("Create_precedesClientError", func(t *testing.T) {
		t.Parallel()
		var nilCreate KibanaCreateFunc[testKibanaResourceModel]
		r := NewKibanaResource[testKibanaResourceModel](
			ComponentKibana,
			"test_entity",
			getTestKibanaResourceSchema,
			testKibanaReadFuncFound,
			testKibanaDeleteFunc,
			nilCreate,
			testKibanaUpdateFuncFound,
		)
		r.client = nonNilTestFactory()

		plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tftypes.NewValue(tftypes.String, "default"))
		respState := tfsdk.State{
			Raw:    tftypes.NewValue(testKibanaResourceObjectType(), nil),
			Schema: testKibanaResourceSchemaWithConnectionBlock(ctx),
		}
		resp := resource.CreateResponse{State: respState}
		r.Create(ctx, resource.CreateRequest{Plan: plan}, &resp)

		require.True(t, resp.Diagnostics.HasError())
		require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Kibana envelope configuration error")
		require.NotContains(t, resp.Diagnostics.Errors()[0].Summary(), "Provider not configured")
	})

	t.Run("Create_precedesInvalidSpaceID", func(t *testing.T) {
		t.Parallel()
		var nilCreate KibanaCreateFunc[testKibanaResourceModel]
		r := NewKibanaResource[testKibanaResourceModel](
			ComponentKibana,
			"test_entity",
			getTestKibanaResourceSchema,
			testKibanaReadFuncFound,
			testKibanaDeleteFunc,
			nilCreate,
			testKibanaUpdateFuncFound,
		)
		r.client = newTestConfiguredFactory(ctx, t)

		objType := testKibanaResourceObjectType()
		objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
			"id":                tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			"name":              tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			"space_id":          tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			"kibana_connection": tftypes.NewValue(kibanaConnectionBlockType(), nil),
		})
		plan := tfsdk.Plan{Raw: objValue, Schema: testKibanaResourceSchemaWithConnectionBlock(ctx)}
		respState := tfsdk.State{
			Raw:    tftypes.NewValue(objType, nil),
			Schema: testKibanaResourceSchemaWithConnectionBlock(ctx),
		}
		resp := resource.CreateResponse{State: respState}
		r.Create(ctx, resource.CreateRequest{Plan: plan}, &resp)

		require.True(t, resp.Diagnostics.HasError())
		require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Kibana envelope configuration error")
		require.NotContains(t, resp.Diagnostics.Errors()[0].Summary(), "Invalid space identifier")
	})
}

// =============================================================================
// Subtask 2.9: Read happy path (found) — composite ID parse path
// =============================================================================

func TestNewKibanaResource_Read_happyPath_compositeID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	r := newTestKibanaResourceEnvelopeWithFactory(t, factory)

	state := makeTestKibanaResourceState(ctx, t, "default/my-stream")
	req := resource.ReadRequest{State: state}
	resp := resource.ReadResponse{State: state}

	r.Read(ctx, req, &resp)

	require.False(t, resp.Diagnostics.HasError())
	var result testKibanaResourceModel
	diags := resp.State.Get(ctx, &result)
	require.False(t, diags.HasError())
	require.Equal(t, "default/my-stream", result.ID.ValueString())
	require.Equal(t, "my-stream", result.Name.ValueString())
	require.Equal(t, "default", result.SpaceID.ValueString())
}

// =============================================================================
// Subtask 2.10: Read happy path (found) — fallback path (plain-UUID resource)
// =============================================================================

func TestNewKibanaResource_Read_happyPath_fallback(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	readCalled := false
	var receivedResourceID, receivedSpaceID string
	r := NewKibanaResource[testKibanaResourceModel](
		ComponentKibana,
		"test_entity",
		getTestKibanaResourceSchema,
		func(_ context.Context, _ *clients.KibanaScopedClient, resourceID string, spaceID string, model testKibanaResourceModel) (testKibanaResourceModel, bool, diag.Diagnostics) {
			readCalled = true
			receivedResourceID = resourceID
			receivedSpaceID = spaceID
			return model, true, nil
		},
		testKibanaDeleteFunc,
		testKibanaCreateFuncFound,
		testKibanaUpdateFuncFound,
	)
	r.client = factory

	state := makeTestKibanaResourceState(ctx, t, "abc-uuid")
	req := resource.ReadRequest{State: state}
	resp := resource.ReadResponse{State: state}

	r.Read(ctx, req, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.True(t, readCalled, "readFunc should be called")
	require.Equal(t, "abc-uuid", receivedResourceID, "fallback should use GetResourceID")
	require.Equal(t, "default", receivedSpaceID, "fallback should use GetSpaceID")
}

// =============================================================================
// Additional: multi-slash ID falls back through composite parse
// =============================================================================

func TestNewKibanaResource_Read_multiSlashIDFallback(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	readCalled := false
	var receivedResourceID, receivedSpaceID string
	r := NewKibanaResource[testKibanaResourceModel](
		ComponentKibana,
		"test_entity",
		getTestKibanaResourceSchema,
		func(_ context.Context, _ *clients.KibanaScopedClient, resourceID string, spaceID string, model testKibanaResourceModel) (testKibanaResourceModel, bool, diag.Diagnostics) {
			readCalled = true
			receivedResourceID = resourceID
			receivedSpaceID = spaceID
			return model, true, nil
		},
		testKibanaDeleteFunc,
		testKibanaCreateFuncFound,
		testKibanaUpdateFuncFound,
	)
	r.client = factory

	connBlockType := kibanaConnectionBlockType()
	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":                tftypes.String,
			"name":              tftypes.String,
			"space_id":          tftypes.String,
			"kibana_connection": connBlockType,
		},
	}
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, "a/b/c"),
		"name":              tftypes.NewValue(tftypes.String, "fallback-name"),
		"space_id":          tftypes.NewValue(tftypes.String, "fallback-space"),
		"kibana_connection": tftypes.NewValue(connBlockType, nil),
	})
	state := tfsdk.State{Raw: objValue, Schema: testKibanaResourceSchemaWithConnectionBlock(ctx)}

	req := resource.ReadRequest{State: state}
	resp := resource.ReadResponse{State: state}

	r.Read(ctx, req, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.True(t, readCalled, "readFunc should be called")
	require.Equal(t, "fallback-name", receivedResourceID, "fallback should use GetResourceID")
	require.Equal(t, "fallback-space", receivedSpaceID, "fallback should use GetSpaceID")
}

// =============================================================================
// Additional C: resolveResourceIdentity with "/"-containing ID
// =============================================================================

func TestNewKibanaResource_Read_compositeIDWinsOverDifferentResourceID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	readCalled := false
	var receivedResourceID, receivedSpaceID string
	r := NewKibanaResource[testKibanaResourceModel](
		ComponentKibana,
		"test_entity",
		getTestKibanaResourceSchema,
		func(_ context.Context, _ *clients.KibanaScopedClient, resourceID string, spaceID string, model testKibanaResourceModel) (testKibanaResourceModel, bool, diag.Diagnostics) {
			readCalled = true
			receivedResourceID = resourceID
			receivedSpaceID = spaceID
			return model, true, nil
		},
		testKibanaDeleteFunc,
		testKibanaCreateFuncFound,
		testKibanaUpdateFuncFound,
	)
	r.client = factory

	// ID looks like a composite ID, but Name (GetResourceID) is different.
	// The composite path should win.
	connBlockType := kibanaConnectionBlockType()
	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":                tftypes.String,
			"name":              tftypes.String,
			"space_id":          tftypes.String,
			"kibana_connection": connBlockType,
		},
	}
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, "foo/bar"),
		"name":              tftypes.NewValue(tftypes.String, "different-name"),
		"space_id":          tftypes.NewValue(tftypes.String, "different-space"),
		"kibana_connection": tftypes.NewValue(connBlockType, nil),
	})
	state := tfsdk.State{Raw: objValue, Schema: testKibanaResourceSchemaWithConnectionBlock(ctx)}

	req := resource.ReadRequest{State: state}
	resp := resource.ReadResponse{State: state}

	r.Read(ctx, req, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.True(t, readCalled, "readFunc should be called")
	require.Equal(t, "bar", receivedResourceID, "composite ID resourceID should win")
	require.Equal(t, "foo", receivedSpaceID, "composite ID spaceID should win")
}

// =============================================================================
// Subtask 2.11: Read not-found removes resource from state
// =============================================================================

func TestNewKibanaResource_Read_notFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	r := NewKibanaResource[testKibanaResourceModel](
		ComponentKibana,
		"test_entity",
		getTestKibanaResourceSchema,
		func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModel) (testKibanaResourceModel, bool, diag.Diagnostics) {
			return testKibanaResourceModel{}, false, nil
		},
		testKibanaDeleteFunc,
		testKibanaCreateFuncFound,
		testKibanaUpdateFuncFound,
	)
	r.client = factory

	state := makeTestKibanaResourceState(ctx, t, "default/my-stream")
	req := resource.ReadRequest{State: state}
	resp := resource.ReadResponse{State: state}

	r.Read(ctx, req, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.True(t, resp.State.Raw.IsNull(), "expected state to be removed")
}

// =============================================================================
// Subtask 2.12: Read short-circuits
// =============================================================================

func TestNewKibanaResource_Read_shortCircuitStateGetError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	readCalled := false
	r := NewKibanaResource[testKibanaResourceModel](
		ComponentKibana,
		"test_entity",
		getTestKibanaResourceSchema,
		func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModel) (testKibanaResourceModel, bool, diag.Diagnostics) {
			readCalled = true
			return testKibanaResourceModel{}, false, nil
		},
		testKibanaDeleteFunc,
		testKibanaCreateFuncFound,
		testKibanaUpdateFuncFound,
	)
	r.client = factory

	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":                tftypes.String,
			"name":              tftypes.String,
			"space_id":          tftypes.String,
			"kibana_connection": kibanaConnectionBlockType(),
		},
	}
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, "default/my-stream"),
		"name":              tftypes.NewValue(tftypes.String, "my-stream"),
		"space_id":          tftypes.NewValue(tftypes.String, "default"),
		"kibana_connection": tftypes.NewValue(kibanaConnectionBlockType(), nil),
	})
	badSchema := getTestKibanaResourceSchema(context.Background())
	state := tfsdk.State{Raw: objValue, Schema: badSchema}

	req := resource.ReadRequest{State: state}
	resp := resource.ReadResponse{State: state}

	r.Read(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.False(t, readCalled, "readFunc should not be called when state.Get fails")
}

func TestNewKibanaResource_Read_shortCircuitEmptyResourceID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	readCalled := false
	r := NewKibanaResource[testKibanaResourceModel](
		ComponentKibana,
		"test_entity",
		getTestKibanaResourceSchema,
		func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModel) (testKibanaResourceModel, bool, diag.Diagnostics) {
			readCalled = true
			return testKibanaResourceModel{}, false, nil
		},
		testKibanaDeleteFunc,
		testKibanaCreateFuncFound,
		testKibanaUpdateFuncFound,
	)
	r.client = factory

	// Composite parse fails (no slash) and GetResourceID returns empty.
	connBlockType := kibanaConnectionBlockType()
	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":                tftypes.String,
			"name":              tftypes.String,
			"space_id":          tftypes.String,
			"kibana_connection": connBlockType,
		},
	}
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, "plain-id-no-composite"),
		"name":              tftypes.NewValue(tftypes.String, ""),
		"space_id":          tftypes.NewValue(tftypes.String, "default"),
		"kibana_connection": tftypes.NewValue(connBlockType, nil),
	})
	state := tfsdk.State{Raw: objValue, Schema: testKibanaResourceSchemaWithConnectionBlock(ctx)}

	req := resource.ReadRequest{State: state}
	resp := resource.ReadResponse{State: state}

	r.Read(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Invalid resource identifier")
	require.False(t, readCalled, "readFunc should not be called when resourceID is empty")
}

func TestNewKibanaResource_Read_shortCircuitClientError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := nonNilTestFactory()
	readCalled := false
	r := NewKibanaResource[testKibanaResourceModel](
		ComponentKibana,
		"test_entity",
		getTestKibanaResourceSchema,
		func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModel) (testKibanaResourceModel, bool, diag.Diagnostics) {
			readCalled = true
			return testKibanaResourceModel{}, false, nil
		},
		testKibanaDeleteFunc,
		testKibanaCreateFuncFound,
		testKibanaUpdateFuncFound,
	)
	r.client = factory

	state := makeTestKibanaResourceState(ctx, t, "default/my-stream")
	req := resource.ReadRequest{State: state}
	resp := resource.ReadResponse{State: state}

	r.Read(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Provider not configured")
	require.False(t, readCalled, "readFunc should not be called when client resolution fails")
}

func TestNewKibanaResource_Read_shortCircuitReadFuncError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	r := NewKibanaResource[testKibanaResourceModel](
		ComponentKibana,
		"test_entity",
		getTestKibanaResourceSchema,
		func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModel) (testKibanaResourceModel, bool, diag.Diagnostics) {
			var diags diag.Diagnostics
			diags.AddError("read error", "something went wrong")
			return testKibanaResourceModel{}, false, diags
		},
		testKibanaDeleteFunc,
		testKibanaCreateFuncFound,
		testKibanaUpdateFuncFound,
	)
	r.client = factory

	state := makeTestKibanaResourceState(ctx, t, "default/my-stream")
	req := resource.ReadRequest{State: state}
	resp := resource.ReadResponse{State: state}

	r.Read(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "read error")
	require.False(t, resp.State.Raw.IsNull(), "state should not be removed when readFunc returns an error")
}

func TestNewKibanaResource_Read_nilReadCallback(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	var nilRead kibanaReadFunc[testKibanaResourceModel]
	r := NewKibanaResource[testKibanaResourceModel](
		ComponentKibana,
		"test_entity",
		getTestKibanaResourceSchema,
		nilRead,
		testKibanaDeleteFunc,
		testKibanaCreateFuncFound,
		testKibanaUpdateFuncFound,
	)
	r.client = factory

	state := makeTestKibanaResourceState(ctx, t, "default/my-stream")
	req := resource.ReadRequest{State: state}
	resp := resource.ReadResponse{State: state}

	r.Read(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Kibana envelope configuration error")
}

// =============================================================================
// Subtask 2.13: Update happy path
// =============================================================================

func TestNewKibanaResource_Update_happyPath(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	r := newTestKibanaResourceEnvelopeWithFactory(t, factory)

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "default/my-resource"), tftypes.NewValue(tftypes.String, "default"))
	prior := makeTestKibanaResourceState(ctx, t, "default/my-resource")
	resp := resource.UpdateResponse{State: prior}
	req := resource.UpdateRequest{Plan: plan, State: prior}

	r.Update(ctx, req, &resp)

	require.False(t, resp.Diagnostics.HasError())
	var result testKibanaResourceModel
	diags := resp.State.Get(ctx, &result)
	require.False(t, diags.HasError())
	require.Equal(t, "default/my-resource", result.ID.ValueString())
	require.Equal(t, "my-resource", result.Name.ValueString())
}

func TestNewKibanaResource_Update_callbackReceivesPlanAndPrior(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	updateCalled := false
	var receivedPlan, receivedPrior testKibanaResourceModel
	r := NewKibanaResource[testKibanaResourceModel](
		ComponentKibana,
		"test_entity",
		getTestKibanaResourceSchema,
		testKibanaReadFuncFound,
		testKibanaDeleteFunc,
		testKibanaCreateFuncFound,
		func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, plan testKibanaResourceModel, prior testKibanaResourceModel) (testKibanaResourceModel, diag.Diagnostics) {
			updateCalled = true
			receivedPlan = plan
			receivedPrior = prior
			return plan, nil
		},
	)
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "default/my-resource"), tftypes.NewValue(tftypes.String, "default"))
	prior := makeTestKibanaResourceState(ctx, t, "default/my-resource")
	resp := resource.UpdateResponse{State: prior}
	req := resource.UpdateRequest{Plan: plan, State: prior}

	r.Update(ctx, req, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.True(t, updateCalled, "update callback should be called")
	require.Equal(t, "my-resource", receivedPlan.Name.ValueString())
	require.Equal(t, "default/my-resource", receivedPrior.ID.ValueString())
}

func TestNewKibanaResource_Update_happyPath_fallback(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	updateCalled := false
	var receivedResourceID, receivedSpaceID string
	r := NewKibanaResource[testKibanaResourceModel](
		ComponentKibana,
		"test_entity",
		getTestKibanaResourceSchema,
		testKibanaReadFuncFound,
		testKibanaDeleteFunc,
		testKibanaCreateFuncFound,
		func(_ context.Context, _ *clients.KibanaScopedClient, resourceID string, spaceID string, plan testKibanaResourceModel, _ testKibanaResourceModel) (testKibanaResourceModel, diag.Diagnostics) {
			updateCalled = true
			receivedResourceID = resourceID
			receivedSpaceID = spaceID
			return plan, nil
		},
	)
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "abc-uuid"), tftypes.NewValue(tftypes.String, "custom-space"))
	prior := makeTestKibanaResourceState(ctx, t, "abc-uuid")
	resp := resource.UpdateResponse{State: prior}
	req := resource.UpdateRequest{Plan: plan, State: prior}

	r.Update(ctx, req, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.True(t, updateCalled, "update callback should be called")
	require.Equal(t, "my-resource", receivedResourceID, "fallback should use GetResourceID")
	require.Equal(t, "custom-space", receivedSpaceID, "fallback should use GetSpaceID")
}

// =============================================================================
// Subtask 2.14: Update short-circuits
// =============================================================================

func TestNewKibanaResource_Update_shortCircuitPlanGetError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	updateCalled := false
	r := NewKibanaResource[testKibanaResourceModel](
		ComponentKibana,
		"test_entity",
		getTestKibanaResourceSchema,
		testKibanaReadFuncFound,
		testKibanaDeleteFunc,
		testKibanaCreateFuncFound,
		func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModel, _ testKibanaResourceModel) (testKibanaResourceModel, diag.Diagnostics) {
			updateCalled = true
			return testKibanaResourceModel{}, nil
		},
	)
	r.client = factory

	objType := testKibanaResourceObjectType()
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"name":              tftypes.NewValue(tftypes.String, "my-resource"),
		"space_id":          tftypes.NewValue(tftypes.String, "default"),
		"kibana_connection": tftypes.NewValue(kibanaConnectionBlockType(), nil),
	})
	badSchema := getTestKibanaResourceSchema(context.Background())
	plan := tfsdk.Plan{Raw: objValue, Schema: badSchema}
	prior := makeTestKibanaResourceState(ctx, t, "default/my-resource")
	resp := resource.UpdateResponse{State: prior}
	req := resource.UpdateRequest{Plan: plan, State: prior}

	r.Update(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.False(t, updateCalled, "update callback should not run when plan.Get fails")
}

func TestNewKibanaResource_Update_shortCircuitStateGetError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	updateCalled := false
	r := NewKibanaResource[testKibanaResourceModel](
		ComponentKibana,
		"test_entity",
		getTestKibanaResourceSchema,
		testKibanaReadFuncFound,
		testKibanaDeleteFunc,
		testKibanaCreateFuncFound,
		func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModel, _ testKibanaResourceModel) (testKibanaResourceModel, diag.Diagnostics) {
			updateCalled = true
			return testKibanaResourceModel{}, nil
		},
	)
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "default/my-resource"), tftypes.NewValue(tftypes.String, "default"))
	objType := testKibanaResourceObjectType()
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, "default/my-resource"),
		"name":              tftypes.NewValue(tftypes.String, "my-resource"),
		"space_id":          tftypes.NewValue(tftypes.String, "default"),
		"kibana_connection": tftypes.NewValue(kibanaConnectionBlockType(), nil),
	})
	badSchema := getTestKibanaResourceSchema(context.Background())
	prior := tfsdk.State{Raw: objValue, Schema: badSchema}
	resp := resource.UpdateResponse{State: prior}
	req := resource.UpdateRequest{Plan: plan, State: prior}

	r.Update(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.False(t, updateCalled, "update callback should not run when state.Get fails")
}

func TestNewKibanaResource_Update_shortCircuitEmptyResourceID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	updateCalled := false
	r := NewKibanaResource[testKibanaResourceModel](
		ComponentKibana,
		"test_entity",
		getTestKibanaResourceSchema,
		testKibanaReadFuncFound,
		testKibanaDeleteFunc,
		testKibanaCreateFuncFound,
		func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModel, _ testKibanaResourceModel) (testKibanaResourceModel, diag.Diagnostics) {
			updateCalled = true
			return testKibanaResourceModel{}, nil
		},
	)
	r.client = factory

	objType := testKibanaResourceObjectType()
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"name":              tftypes.NewValue(tftypes.String, ""),
		"space_id":          tftypes.NewValue(tftypes.String, "default"),
		"kibana_connection": tftypes.NewValue(kibanaConnectionBlockType(), nil),
	})
	plan := tfsdk.Plan{Raw: objValue, Schema: testKibanaResourceSchemaWithConnectionBlock(ctx)}
	prior := makeTestKibanaResourceState(ctx, t, "default/old-resource")
	resp := resource.UpdateResponse{State: prior}
	req := resource.UpdateRequest{Plan: plan, State: prior}

	r.Update(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Invalid resource identifier")
	require.False(t, updateCalled, "update callback should not run when resourceID is empty")
}

func TestNewKibanaResource_Update_shortCircuitClientError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := nonNilTestFactory()
	updateCalled := false
	r := NewKibanaResource[testKibanaResourceModel](
		ComponentKibana,
		"test_entity",
		getTestKibanaResourceSchema,
		testKibanaReadFuncFound,
		testKibanaDeleteFunc,
		testKibanaCreateFuncFound,
		func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModel, _ testKibanaResourceModel) (testKibanaResourceModel, diag.Diagnostics) {
			updateCalled = true
			return testKibanaResourceModel{}, nil
		},
	)
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "default/my-resource"), tftypes.NewValue(tftypes.String, "default"))
	prior := makeTestKibanaResourceState(ctx, t, "default/my-resource")
	resp := resource.UpdateResponse{State: prior}
	req := resource.UpdateRequest{Plan: plan, State: prior}

	r.Update(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Provider not configured")
	require.False(t, updateCalled, "update callback should not run when client resolution fails")
}

func TestNewKibanaResource_Update_shortCircuitCallbackError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	r := NewKibanaResource[testKibanaResourceModel](
		ComponentKibana,
		"test_entity",
		getTestKibanaResourceSchema,
		testKibanaReadFuncFound,
		testKibanaDeleteFunc,
		testKibanaCreateFuncFound,
		func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModel, _ testKibanaResourceModel) (testKibanaResourceModel, diag.Diagnostics) {
			var diags diag.Diagnostics
			diags.AddError("update error", "something went wrong")
			return testKibanaResourceModel{}, diags
		},
	)
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "default/my-resource"), tftypes.NewValue(tftypes.String, "default"))
	prior := makeTestKibanaResourceState(ctx, t, "default/my-resource")
	resp := resource.UpdateResponse{State: prior}
	req := resource.UpdateRequest{Plan: plan, State: prior}

	r.Update(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "update error")
	var after testKibanaResourceModel
	require.False(t, resp.State.Get(ctx, &after).HasError())
	require.Equal(t, "default/my-resource", after.ID.ValueString(), "state should not change when update callback returns errors")
}

// =============================================================================
// Subtask 2.15: Update with nil and placeholder write callbacks
// =============================================================================

func TestNewKibanaResource_Update_nilWriteCallback(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	var nilUpdate KibanaUpdateFunc[testKibanaResourceModel]
	r := NewKibanaResource[testKibanaResourceModel](
		ComponentKibana,
		"test_entity",
		getTestKibanaResourceSchema,
		testKibanaReadFuncFound,
		testKibanaDeleteFunc,
		testKibanaCreateFuncFound,
		nilUpdate,
	)
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "default/my-resource"), tftypes.NewValue(tftypes.String, "default"))
	prior := makeTestKibanaResourceState(ctx, t, "default/my-resource")
	resp := resource.UpdateResponse{State: prior}
	req := resource.UpdateRequest{Plan: plan, State: prior}

	r.Update(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Kibana envelope configuration error")
}

func TestNewKibanaResource_Update_placeholderCallbackError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	createFn, updateFn := PlaceholderKibanaWriteCallbacks[testKibanaResourceModel]()
	r := NewKibanaResource[testKibanaResourceModel](
		ComponentKibana,
		"test_entity",
		getTestKibanaResourceSchema,
		testKibanaReadFuncFound,
		testKibanaDeleteFunc,
		createFn,
		updateFn,
	)
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "default/my-resource"), tftypes.NewValue(tftypes.String, "default"))
	prior := makeTestKibanaResourceState(ctx, t, "default/my-resource")
	resp := resource.UpdateResponse{State: prior}
	req := resource.UpdateRequest{Plan: plan, State: prior}

	r.Update(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	err0 := resp.Diagnostics.Errors()[0]
	require.Equal(t, placeholderKibanaWriteCallbackSummary, err0.Summary())
	require.Equal(t, placeholderKibanaWriteCallbackDetail, err0.Detail())
}

func TestNewKibanaResource_Update_nilCallbackPrecedesOtherPreludeErrors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("Update_precedesClientError", func(t *testing.T) {
		t.Parallel()
		var nilUpdate KibanaUpdateFunc[testKibanaResourceModel]
		r := NewKibanaResource[testKibanaResourceModel](
			ComponentKibana,
			"test_entity",
			getTestKibanaResourceSchema,
			testKibanaReadFuncFound,
			testKibanaDeleteFunc,
			testKibanaCreateFuncFound,
			nilUpdate,
		)
		r.client = nonNilTestFactory()

		plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "default/my-resource"), tftypes.NewValue(tftypes.String, "default"))
		prior := makeTestKibanaResourceState(ctx, t, "default/my-resource")
		resp := resource.UpdateResponse{State: prior}
		r.Update(ctx, resource.UpdateRequest{Plan: plan, State: prior}, &resp)

		require.True(t, resp.Diagnostics.HasError())
		require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Kibana envelope configuration error")
		require.NotContains(t, resp.Diagnostics.Errors()[0].Summary(), "Provider not configured")
	})

	t.Run("Update_precedesEmptyResourceID", func(t *testing.T) {
		t.Parallel()
		var nilUpdate KibanaUpdateFunc[testKibanaResourceModel]
		r := NewKibanaResource[testKibanaResourceModel](
			ComponentKibana,
			"test_entity",
			getTestKibanaResourceSchema,
			testKibanaReadFuncFound,
			testKibanaDeleteFunc,
			testKibanaCreateFuncFound,
			nilUpdate,
		)
		r.client = newTestConfiguredFactory(ctx, t)

		objType := testKibanaResourceObjectType()
		objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
			"id":                tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			"name":              tftypes.NewValue(tftypes.String, ""),
			"space_id":          tftypes.NewValue(tftypes.String, "default"),
			"kibana_connection": tftypes.NewValue(kibanaConnectionBlockType(), nil),
		})
		plan := tfsdk.Plan{Raw: objValue, Schema: testKibanaResourceSchemaWithConnectionBlock(ctx)}
		prior := makeTestKibanaResourceState(ctx, t, "default/old-resource")
		resp := resource.UpdateResponse{State: prior}
		r.Update(ctx, resource.UpdateRequest{Plan: plan, State: prior}, &resp)

		require.True(t, resp.Diagnostics.HasError())
		require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Kibana envelope configuration error")
		require.NotContains(t, resp.Diagnostics.Errors()[0].Summary(), "Invalid resource identifier")
	})
}

// =============================================================================
// Subtask 2.16: Delete happy path
// =============================================================================

func TestNewKibanaResource_Delete_happyPath(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	deleteCalled := false
	var receivedResourceID, receivedSpaceID string
	r := NewKibanaResource[testKibanaResourceModel](
		ComponentKibana,
		"test_entity",
		getTestKibanaResourceSchema,
		testKibanaReadFuncFound,
		func(_ context.Context, _ *clients.KibanaScopedClient, resourceID string, spaceID string, _ testKibanaResourceModel) diag.Diagnostics {
			deleteCalled = true
			receivedResourceID = resourceID
			receivedSpaceID = spaceID
			return nil
		},
		testKibanaCreateFuncFound,
		testKibanaUpdateFuncFound,
	)
	r.client = factory

	state := makeTestKibanaResourceState(ctx, t, "default/my-stream")
	req := resource.DeleteRequest{State: state}
	resp := resource.DeleteResponse{State: state}

	r.Delete(ctx, req, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.True(t, deleteCalled, "deleteFunc should be called")
	require.Equal(t, "my-stream", receivedResourceID)
	require.Equal(t, "default", receivedSpaceID)
}

func TestNewKibanaResource_Delete_happyPath_fallback(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	deleteCalled := false
	var receivedResourceID, receivedSpaceID string
	r := NewKibanaResource[testKibanaResourceModel](
		ComponentKibana,
		"test_entity",
		getTestKibanaResourceSchema,
		testKibanaReadFuncFound,
		func(_ context.Context, _ *clients.KibanaScopedClient, resourceID string, spaceID string, _ testKibanaResourceModel) diag.Diagnostics {
			deleteCalled = true
			receivedResourceID = resourceID
			receivedSpaceID = spaceID
			return nil
		},
		testKibanaCreateFuncFound,
		testKibanaUpdateFuncFound,
	)
	r.client = factory

	state := makeTestKibanaResourceState(ctx, t, "abc-uuid")
	req := resource.DeleteRequest{State: state}
	resp := resource.DeleteResponse{State: state}

	r.Delete(ctx, req, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.True(t, deleteCalled, "deleteFunc should be called")
	require.Equal(t, "abc-uuid", receivedResourceID, "fallback should use GetResourceID")
	require.Equal(t, "default", receivedSpaceID, "fallback should use GetSpaceID")
}

// =============================================================================
// Subtask 2.17: Delete short-circuits
// =============================================================================

func TestNewKibanaResource_Delete_shortCircuitStateGetError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	deleteCalled := false
	r := NewKibanaResource[testKibanaResourceModel](
		ComponentKibana,
		"test_entity",
		getTestKibanaResourceSchema,
		testKibanaReadFuncFound,
		func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModel) diag.Diagnostics {
			deleteCalled = true
			return nil
		},
		testKibanaCreateFuncFound,
		testKibanaUpdateFuncFound,
	)
	r.client = factory

	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":                tftypes.String,
			"name":              tftypes.String,
			"space_id":          tftypes.String,
			"kibana_connection": kibanaConnectionBlockType(),
		},
	}
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, "default/my-stream"),
		"name":              tftypes.NewValue(tftypes.String, "my-stream"),
		"space_id":          tftypes.NewValue(tftypes.String, "default"),
		"kibana_connection": tftypes.NewValue(kibanaConnectionBlockType(), nil),
	})
	badSchema := getTestKibanaResourceSchema(context.Background())
	state := tfsdk.State{Raw: objValue, Schema: badSchema}

	req := resource.DeleteRequest{State: state}
	resp := resource.DeleteResponse{State: state}

	r.Delete(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.False(t, deleteCalled, "deleteFunc should not be called when state.Get fails")
}

func TestNewKibanaResource_Delete_shortCircuitEmptyResourceID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	deleteCalled := false
	r := NewKibanaResource[testKibanaResourceModel](
		ComponentKibana,
		"test_entity",
		getTestKibanaResourceSchema,
		testKibanaReadFuncFound,
		func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModel) diag.Diagnostics {
			deleteCalled = true
			return nil
		},
		testKibanaCreateFuncFound,
		testKibanaUpdateFuncFound,
	)
	r.client = factory

	connBlockType := kibanaConnectionBlockType()
	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":                tftypes.String,
			"name":              tftypes.String,
			"space_id":          tftypes.String,
			"kibana_connection": connBlockType,
		},
	}
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, "plain-id"),
		"name":              tftypes.NewValue(tftypes.String, ""),
		"space_id":          tftypes.NewValue(tftypes.String, "default"),
		"kibana_connection": tftypes.NewValue(connBlockType, nil),
	})
	state := tfsdk.State{Raw: objValue, Schema: testKibanaResourceSchemaWithConnectionBlock(ctx)}

	req := resource.DeleteRequest{State: state}
	resp := resource.DeleteResponse{State: state}

	r.Delete(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Invalid resource identifier")
	require.False(t, deleteCalled, "deleteFunc should not be called when resourceID is empty")
}

func TestNewKibanaResource_Delete_shortCircuitClientError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := nonNilTestFactory()
	deleteCalled := false
	r := NewKibanaResource[testKibanaResourceModel](
		ComponentKibana,
		"test_entity",
		getTestKibanaResourceSchema,
		testKibanaReadFuncFound,
		func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModel) diag.Diagnostics {
			deleteCalled = true
			return nil
		},
		testKibanaCreateFuncFound,
		testKibanaUpdateFuncFound,
	)
	r.client = factory

	state := makeTestKibanaResourceState(ctx, t, "default/my-stream")
	req := resource.DeleteRequest{State: state}
	resp := resource.DeleteResponse{State: state}

	r.Delete(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Provider not configured")
	require.False(t, deleteCalled, "deleteFunc should not be called when client resolution fails")
}

func TestNewKibanaResource_Delete_appendsDeleteFuncDiagnostics(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	r := NewKibanaResource[testKibanaResourceModel](
		ComponentKibana,
		"test_entity",
		getTestKibanaResourceSchema,
		testKibanaReadFuncFound,
		func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModel) diag.Diagnostics {
			var diags diag.Diagnostics
			diags.AddError("delete error", "something went wrong")
			return diags
		},
		testKibanaCreateFuncFound,
		testKibanaUpdateFuncFound,
	)
	r.client = factory

	state := makeTestKibanaResourceState(ctx, t, "default/my-stream")
	req := resource.DeleteRequest{State: state}
	resp := resource.DeleteResponse{State: state}

	r.Delete(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "delete error")
}

func TestNewKibanaResource_Delete_nilDeleteCallback(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	var nilDelete kibanaDeleteFunc[testKibanaResourceModel]
	r := NewKibanaResource[testKibanaResourceModel](
		ComponentKibana,
		"test_entity",
		getTestKibanaResourceSchema,
		testKibanaReadFuncFound,
		nilDelete,
		testKibanaCreateFuncFound,
		testKibanaUpdateFuncFound,
	)
	r.client = factory

	state := makeTestKibanaResourceState(ctx, t, "default/my-stream")
	req := resource.DeleteRequest{State: state}
	resp := resource.DeleteResponse{State: state}

	r.Delete(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Kibana envelope configuration error")
}

// =============================================================================
// Version requirements
// =============================================================================

// testKibanaResourceModelWithVersionReqs implements WithVersionRequirements
// and always returns error diagnostics from GetVersionRequirements.
type testKibanaResourceModelWithVersionReqs struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	SpaceID          types.String `tfsdk:"space_id"`
	KibanaConnection types.List   `tfsdk:"kibana_connection"`
}

func (m testKibanaResourceModelWithVersionReqs) GetID() types.String         { return m.ID }
func (m testKibanaResourceModelWithVersionReqs) GetResourceID() types.String { return m.Name }
func (m testKibanaResourceModelWithVersionReqs) GetSpaceID() types.String    { return m.SpaceID }
func (m testKibanaResourceModelWithVersionReqs) GetKibanaConnection() types.List {
	return m.KibanaConnection
}
func (*testKibanaResourceModelWithVersionReqs) GetVersionRequirements() ([]VersionRequirement, diag.Diagnostics) {
	return nil, diag.Diagnostics{
		diag.NewErrorDiagnostic("version requirements error", "injected GetVersionRequirements failure"),
	}
}

func TestWithVersionRequirements_resourcePointerAssertionTrue(t *testing.T) {
	t.Parallel()
	var m testKibanaResourceModelWithVersionReqs
	_, ok := any(m).(WithVersionRequirements)
	require.False(t, ok, "value testKibanaResourceModelWithVersionReqs must not satisfy interface")
	_, ok = any(&m).(WithVersionRequirements)
	require.True(t, ok, "*testKibanaResourceModelWithVersionReqs must satisfy WithVersionRequirements")
}

func TestKibanaResource_Create_versionReqDiagsStopCreate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	createCalled := false
	r := NewKibanaResource[testKibanaResourceModelWithVersionReqs](
		ComponentKibana,
		"test_entity",
		func(ctx context.Context) rschema.Schema {
			s := getTestKibanaResourceSchema(ctx)
			s.Blocks = map[string]rschema.Block{
				"kibana_connection": providerschema.GetKbFWConnectionBlock(),
			}
			return s
		},
		func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModelWithVersionReqs) (testKibanaResourceModelWithVersionReqs, bool, diag.Diagnostics) {
			return testKibanaResourceModelWithVersionReqs{}, false, nil
		},
		func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModelWithVersionReqs) diag.Diagnostics {
			return nil
		},
		func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ testKibanaResourceModelWithVersionReqs) (testKibanaResourceModelWithVersionReqs, diag.Diagnostics) {
			createCalled = true
			return testKibanaResourceModelWithVersionReqs{}, nil
		},
		func(
			_ context.Context,
			_ *clients.KibanaScopedClient,
			_ string,
			_ string,
			_ testKibanaResourceModelWithVersionReqs,
			_ testKibanaResourceModelWithVersionReqs,
		) (testKibanaResourceModelWithVersionReqs, diag.Diagnostics) {
			return testKibanaResourceModelWithVersionReqs{}, nil
		},
	)
	r.client = factory

	objType := testKibanaResourceObjectType()
	plan := tfsdk.Plan{
		Raw: tftypes.NewValue(objType, map[string]tftypes.Value{
			"id":                tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			"name":              tftypes.NewValue(tftypes.String, "my-resource"),
			"space_id":          tftypes.NewValue(tftypes.String, "default"),
			"kibana_connection": tftypes.NewValue(kibanaConnectionBlockType(), nil),
		}),
		Schema: testKibanaResourceSchemaWithConnectionBlock(ctx),
	}
	req := resource.CreateRequest{Plan: plan}
	respState := tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: testKibanaResourceSchemaWithConnectionBlock(ctx)}
	resp := resource.CreateResponse{State: respState}

	r.Create(ctx, req, &resp)

	require.False(t, createCalled, "createFunc must NOT be called when GetVersionRequirements returns error diags")
	require.True(t, resp.Diagnostics.HasError(), "Create must propagate error from GetVersionRequirements")
	summaries := make([]string, 0, len(resp.Diagnostics))
	for _, d := range resp.Diagnostics {
		summaries = append(summaries, d.Summary())
	}
	require.Contains(t, summaries, "version requirements error",
		"diagnostic from GetVersionRequirements must be appended; got: %v", summaries)
}

func TestKibanaResource_Read_versionReqDiagsStopRead(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	readCalled := false
	r := NewKibanaResource[testKibanaResourceModelWithVersionReqs](
		ComponentKibana,
		"test_entity",
		func(ctx context.Context) rschema.Schema {
			s := getTestKibanaResourceSchema(ctx)
			s.Blocks = map[string]rschema.Block{
				"kibana_connection": providerschema.GetKbFWConnectionBlock(),
			}
			return s
		},
		func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModelWithVersionReqs) (testKibanaResourceModelWithVersionReqs, bool, diag.Diagnostics) {
			readCalled = true
			return testKibanaResourceModelWithVersionReqs{}, false, nil
		},
		func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModelWithVersionReqs) diag.Diagnostics {
			return nil
		},
		func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ testKibanaResourceModelWithVersionReqs) (testKibanaResourceModelWithVersionReqs, diag.Diagnostics) {
			return testKibanaResourceModelWithVersionReqs{}, nil
		},
		func(
			_ context.Context,
			_ *clients.KibanaScopedClient,
			_ string,
			_ string,
			_ testKibanaResourceModelWithVersionReqs,
			_ testKibanaResourceModelWithVersionReqs,
		) (testKibanaResourceModelWithVersionReqs, diag.Diagnostics) {
			return testKibanaResourceModelWithVersionReqs{}, nil
		},
	)
	r.client = factory

	state := tfsdk.State{
		Raw: tftypes.NewValue(testKibanaResourceObjectType(), map[string]tftypes.Value{
			"id":                tftypes.NewValue(tftypes.String, "default/my-resource"),
			"name":              tftypes.NewValue(tftypes.String, "my-resource"),
			"space_id":          tftypes.NewValue(tftypes.String, "default"),
			"kibana_connection": tftypes.NewValue(kibanaConnectionBlockType(), nil),
		}),
		Schema: testKibanaResourceSchemaWithConnectionBlock(ctx),
	}
	req := resource.ReadRequest{State: state}
	resp := resource.ReadResponse{State: state}

	r.Read(ctx, req, &resp)

	require.False(t, readCalled, "readFunc must NOT be called when GetVersionRequirements returns error diags")
	require.True(t, resp.Diagnostics.HasError(), "Read must propagate error from GetVersionRequirements")
	summaries := make([]string, 0, len(resp.Diagnostics))
	for _, d := range resp.Diagnostics {
		summaries = append(summaries, d.Summary())
	}
	require.Contains(t, summaries, "version requirements error",
		"diagnostic from GetVersionRequirements must be appended; got: %v", summaries)
}

func TestKibanaResource_Update_versionReqDiagsStopUpdate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	updateCalled := false
	r := NewKibanaResource[testKibanaResourceModelWithVersionReqs](
		ComponentKibana,
		"test_entity",
		func(ctx context.Context) rschema.Schema {
			s := getTestKibanaResourceSchema(ctx)
			s.Blocks = map[string]rschema.Block{
				"kibana_connection": providerschema.GetKbFWConnectionBlock(),
			}
			return s
		},
		func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModelWithVersionReqs) (testKibanaResourceModelWithVersionReqs, bool, diag.Diagnostics) {
			return testKibanaResourceModelWithVersionReqs{}, false, nil
		},
		func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModelWithVersionReqs) diag.Diagnostics {
			return nil
		},
		func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ testKibanaResourceModelWithVersionReqs) (testKibanaResourceModelWithVersionReqs, diag.Diagnostics) {
			return testKibanaResourceModelWithVersionReqs{}, nil
		},
		func(
			_ context.Context,
			_ *clients.KibanaScopedClient,
			_ string,
			_ string,
			_ testKibanaResourceModelWithVersionReqs,
			_ testKibanaResourceModelWithVersionReqs,
		) (testKibanaResourceModelWithVersionReqs, diag.Diagnostics) {
			updateCalled = true
			return testKibanaResourceModelWithVersionReqs{}, nil
		},
	)
	r.client = factory

	objType := testKibanaResourceObjectType()
	plan := tfsdk.Plan{
		Raw: tftypes.NewValue(objType, map[string]tftypes.Value{
			"id":                tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			"name":              tftypes.NewValue(tftypes.String, "my-resource"),
			"space_id":          tftypes.NewValue(tftypes.String, "default"),
			"kibana_connection": tftypes.NewValue(kibanaConnectionBlockType(), nil),
		}),
		Schema: testKibanaResourceSchemaWithConnectionBlock(ctx),
	}
	state := tfsdk.State{
		Raw: tftypes.NewValue(objType, map[string]tftypes.Value{
			"id":                tftypes.NewValue(tftypes.String, "default/my-resource"),
			"name":              tftypes.NewValue(tftypes.String, "my-resource"),
			"space_id":          tftypes.NewValue(tftypes.String, "default"),
			"kibana_connection": tftypes.NewValue(kibanaConnectionBlockType(), nil),
		}),
		Schema: testKibanaResourceSchemaWithConnectionBlock(ctx),
	}
	req := resource.UpdateRequest{Plan: plan, State: state}
	resp := resource.UpdateResponse{State: state}

	r.Update(ctx, req, &resp)

	require.False(t, updateCalled, "updateFunc must NOT be called when GetVersionRequirements returns error diags")
	require.True(t, resp.Diagnostics.HasError(), "Update must propagate error from GetVersionRequirements")
	summaries := make([]string, 0, len(resp.Diagnostics))
	for _, d := range resp.Diagnostics {
		summaries = append(summaries, d.Summary())
	}
	require.Contains(t, summaries, "version requirements error",
		"diagnostic from GetVersionRequirements must be appended; got: %v", summaries)
}
