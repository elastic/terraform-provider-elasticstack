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

func getTestKibanaResourceSchema() rschema.Schema {
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

func testKibanaCreateFuncFound(_ context.Context, _ *clients.KibanaScopedClient, spaceID string, plan testKibanaResourceModel) (testKibanaResourceModel, diag.Diagnostics) {
	plan.ID = types.StringValue(plan.GetSpaceID().ValueString() + "/" + plan.GetResourceID().ValueString())
	return plan, nil
}

func testKibanaUpdateFuncFound(_ context.Context, _ *clients.KibanaScopedClient, resourceID string, spaceID string, plan testKibanaResourceModel, _ testKibanaResourceModel) (testKibanaResourceModel, diag.Diagnostics) {
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

func testKibanaResourceSchemaWithConnectionBlock() rschema.Schema {
	s := getTestKibanaResourceSchema()
	s.Blocks = map[string]rschema.Block{
		"kibana_connection": providerschema.GetKbFWConnectionBlock(),
	}
	return s
}

func makeTestKibanaResourceCreatePlan(t *testing.T, idValue tftypes.Value, spaceIDValue tftypes.Value) tfsdk.Plan {
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
		Schema: testKibanaResourceSchemaWithConnectionBlock(),
	}
}

func makeTestKibanaResourceState(t *testing.T, id string, spaceID string) tfsdk.State {
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
		"space_id":          tftypes.NewValue(tftypes.String, spaceID),
		"kibana_connection": tftypes.NewValue(connBlockType, nil),
	})

	return tfsdk.State{
		Raw:    objValue,
		Schema: testKibanaResourceSchemaWithConnectionBlock(),
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
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", getTestKibanaResourceSchema, testKibanaReadFuncFound, testKibanaDeleteFunc, testKibanaCreateFuncFound, testKibanaUpdateFuncFound)
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
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", getTestKibanaResourceSchema, testKibanaReadFuncFound, testKibanaDeleteFunc, testKibanaCreateFuncFound, testKibanaUpdateFuncFound)

	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: testProviderTypeName}, &resp)

	require.Equal(t, "elasticstack_kibana_test_entity", resp.TypeName)
}

// =============================================================================
// Subtask 2.4: Schema injection
// =============================================================================

func TestNewKibanaResource_schemaInjection(t *testing.T) {
	t.Parallel()
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", getTestKibanaResourceSchema, testKibanaReadFuncFound, testKibanaDeleteFunc, testKibanaCreateFuncFound, testKibanaUpdateFuncFound)

	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Schema.Blocks, "kibana_connection")
	require.Contains(t, resp.Schema.Attributes, "id")
}

func TestNewKibanaResource_schemaDefensiveClone(t *testing.T) {
	t.Parallel()
	originalSchema := getTestKibanaResourceSchema()
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", func() rschema.Schema {
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
		r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", getTestKibanaResourceSchema, testKibanaReadFuncFound, testKibanaDeleteFunc, testKibanaCreateFuncFound, testKibanaUpdateFuncFound)
		var resp resource.ConfigureResponse
		r.Configure(ctx, resource.ConfigureRequest{ProviderData: nil}, &resp)
		require.False(t, resp.Diagnostics.HasError())
	})

	t.Run("valid_factory", func(t *testing.T) {
		t.Parallel()
		r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", getTestKibanaResourceSchema, testKibanaReadFuncFound, testKibanaDeleteFunc, testKibanaCreateFuncFound, testKibanaUpdateFuncFound)
		f := nonNilTestFactory()
		var resp resource.ConfigureResponse
		r.Configure(ctx, resource.ConfigureRequest{ProviderData: f}, &resp)
		require.False(t, resp.Diagnostics.HasError())
	})

	t.Run("invalid_provider_data", func(t *testing.T) {
		t.Parallel()
		r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", getTestKibanaResourceSchema, testKibanaReadFuncFound, testKibanaDeleteFunc, testKibanaCreateFuncFound, testKibanaUpdateFuncFound)
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

	plan := makeTestKibanaResourceCreatePlan(t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tftypes.NewValue(tftypes.String, "default"))
	objType := testKibanaResourceObjectType()
	respState := tfsdk.State{
		Raw:    tftypes.NewValue(objType, nil),
		Schema: testKibanaResourceSchemaWithConnectionBlock(),
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

	plan := makeTestKibanaResourceCreatePlan(t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tftypes.NewValue(tftypes.String, tftypes.UnknownValue))
	objType := testKibanaResourceObjectType()
	respState := tfsdk.State{
		Raw:    tftypes.NewValue(objType, nil),
		Schema: testKibanaResourceSchemaWithConnectionBlock(),
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

	plan := makeTestKibanaResourceCreatePlan(t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tftypes.NewValue(tftypes.String, nil))
	objType := testKibanaResourceObjectType()
	respState := tfsdk.State{
		Raw:    tftypes.NewValue(objType, nil),
		Schema: testKibanaResourceSchemaWithConnectionBlock(),
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
	plan := tfsdk.Plan{Raw: objValue, Schema: testKibanaResourceSchemaWithConnectionBlock()}
	respState := tfsdk.State{
		Raw:    tftypes.NewValue(objType, nil),
		Schema: testKibanaResourceSchemaWithConnectionBlock(),
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

	plan := makeTestKibanaResourceCreatePlan(t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tftypes.NewValue(tftypes.String, "default"))
	objType := testKibanaResourceObjectType()
	respState := tfsdk.State{
		Raw:    tftypes.NewValue(objType, nil),
		Schema: testKibanaResourceSchemaWithConnectionBlock(),
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

	plan := makeTestKibanaResourceCreatePlan(t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tftypes.NewValue(tftypes.String, "default"))
	objType := testKibanaResourceObjectType()
	respState := tfsdk.State{
		Raw:    tftypes.NewValue(objType, nil),
		Schema: testKibanaResourceSchemaWithConnectionBlock(),
	}
	req := resource.CreateRequest{Plan: plan}
	resp := resource.CreateResponse{State: respState}

	r.Create(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "create error")
	require.True(t, resp.State.Raw.IsNull(), "state should not be mutated when create callback fails")
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

	plan := makeTestKibanaResourceCreatePlan(t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tftypes.NewValue(tftypes.String, "default"))
	objType := testKibanaResourceObjectType()
	respState := tfsdk.State{
		Raw:    tftypes.NewValue(objType, nil),
		Schema: testKibanaResourceSchemaWithConnectionBlock(),
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

	plan := makeTestKibanaResourceCreatePlan(t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tftypes.NewValue(tftypes.String, "default"))
	objType := testKibanaResourceObjectType()
	respState := tfsdk.State{
		Raw:    tftypes.NewValue(objType, nil),
		Schema: testKibanaResourceSchemaWithConnectionBlock(),
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

		plan := makeTestKibanaResourceCreatePlan(t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tftypes.NewValue(tftypes.String, "default"))
		respState := tfsdk.State{
			Raw:    tftypes.NewValue(testKibanaResourceObjectType(), nil),
			Schema: testKibanaResourceSchemaWithConnectionBlock(),
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
		plan := tfsdk.Plan{Raw: objValue, Schema: testKibanaResourceSchemaWithConnectionBlock()}
		respState := tfsdk.State{
			Raw:    tftypes.NewValue(objType, nil),
			Schema: testKibanaResourceSchemaWithConnectionBlock(),
		}
		resp := resource.CreateResponse{State: respState}
		r.Create(ctx, resource.CreateRequest{Plan: plan}, &resp)

		require.True(t, resp.Diagnostics.HasError())
		require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Kibana envelope configuration error")
		require.NotContains(t, resp.Diagnostics.Errors()[0].Summary(), "Invalid space identifier")
	})
}
