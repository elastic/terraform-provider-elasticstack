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
	"maps"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
	ResourceTimeoutsField
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

func testKibanaReadFuncDistinguishing(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, model testKibanaResourceModel) (testKibanaResourceModel, bool, diag.Diagnostics) {
	model.Name = types.StringValue(model.Name.ValueString() + "-read")
	return model, true, nil
}

func testKibanaDeleteFunc(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModel) diag.Diagnostics {
	return nil
}

func testKibanaWriteFuncFound(
	_ context.Context,
	_ *clients.KibanaScopedClient,
	req KibanaWriteRequest[testKibanaResourceModel],
) (KibanaWriteResult[testKibanaResourceModel], diag.Diagnostics) {
	model := req.Plan
	model.ID = types.StringValue(req.SpaceID + "/" + req.WriteID)
	return KibanaWriteResult[testKibanaResourceModel]{Model: model}, nil
}

func defaultTestKibanaResourceOptions() KibanaResourceOptions[testKibanaResourceModel] {
	return KibanaResourceOptions[testKibanaResourceModel]{
		Schema: getTestKibanaResourceSchema,
		Read:   testKibanaReadFuncFound,
		Delete: testKibanaDeleteFunc,
		Create: testKibanaWriteFuncFound,
		Update: testKibanaWriteFuncFound,
	}
}

// testKibanaResourceModelUnscoped opts into [KibanaUnscopedSpace] for envelope tests.
type testKibanaResourceModelUnscoped struct {
	testKibanaResourceModel
}

func (testKibanaResourceModelUnscoped) IsUnscopedSpace() bool { return true }

func testKibanaWriteFuncUnscoped(
	_ context.Context,
	_ *clients.KibanaScopedClient,
	req KibanaWriteRequest[testKibanaResourceModelUnscoped],
) (KibanaWriteResult[testKibanaResourceModelUnscoped], diag.Diagnostics) {
	plan := req.Plan
	plan.ID = types.StringValue(plan.GetResourceID().ValueString())
	return KibanaWriteResult[testKibanaResourceModelUnscoped]{Model: plan}, nil
}

func testKibanaReadFuncUnscoped(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, model testKibanaResourceModelUnscoped) (testKibanaResourceModelUnscoped, bool, diag.Diagnostics) {
	return model, true, nil
}

func testKibanaDeleteFuncUnscoped(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModelUnscoped) diag.Diagnostics {
	return nil
}

func testKibanaResourceObjectType() tftypes.Type {
	return tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":                tftypes.String,
			"name":              tftypes.String,
			"space_id":          tftypes.String,
			"kibana_connection": kibanaConnectionBlockType(),
			"timeouts":          resourceTimeoutsObjectType(),
		},
	}
}

func testKibanaResourceSchemaWithConnectionBlock(ctx context.Context) rschema.Schema {
	s := getTestKibanaResourceSchema(ctx)
	s.Blocks = map[string]rschema.Block{
		"kibana_connection": providerschema.GetKbFWConnectionBlock(),
	}
	attrs := make(map[string]rschema.Attribute, len(s.Attributes)+1)
	maps.Copy(attrs, s.Attributes)
	attrs[attrTimeouts] = timeouts.AttributesAll(ctx)
	s.Attributes = attrs
	return s
}

func kibanaTestConfig(plan tfsdk.Plan) tfsdk.Config {
	return tfsdk.Config(plan)
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
		"timeouts":          resourceTimeoutsNullValue(),
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
			"timeouts":          resourceTimeoutsObjectType(),
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
		"timeouts":          resourceTimeoutsNullValue(),
	})

	return tfsdk.State{
		Raw:    objValue,
		Schema: testKibanaResourceSchemaWithConnectionBlock(ctx),
	}
}

func newTestKibanaResourceEnvelopeWithFactory(t *testing.T, factory *clients.ProviderClientFactory) *KibanaResource[testKibanaResourceModel] {
	t.Helper()
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", defaultTestKibanaResourceOptions())
	r.client = factory
	return r
}

// =============================================================================
// Subtask 2.2: Type assertions
// =============================================================================

func TestNewKibanaResource_typeAssertions(t *testing.T) {
	t.Parallel()
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", defaultTestKibanaResourceOptions())
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
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", defaultTestKibanaResourceOptions())

	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: testProviderTypeName}, &resp)

	require.Equal(t, "elasticstack_kibana_test_entity", resp.TypeName)
}

// =============================================================================
// Subtask 2.4: Schema injection
// =============================================================================

func TestNewKibanaResource_schemaInjection(t *testing.T) {
	t.Parallel()
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", defaultTestKibanaResourceOptions())

	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Schema.Blocks, "kibana_connection")
	require.Contains(t, resp.Schema.Attributes, "id")
}

func TestNewKibanaResource_schemaDefensiveClone(t *testing.T) {
	t.Parallel()
	originalSchema := getTestKibanaResourceSchema(context.Background())
	opts := defaultTestKibanaResourceOptions()
	opts.Schema = func(_ context.Context) rschema.Schema {
		return originalSchema
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)

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
		r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", defaultTestKibanaResourceOptions())
		var resp resource.ConfigureResponse
		r.Configure(ctx, resource.ConfigureRequest{ProviderData: nil}, &resp)
		require.False(t, resp.Diagnostics.HasError())
	})

	t.Run("valid_factory", func(t *testing.T) {
		t.Parallel()
		r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", defaultTestKibanaResourceOptions())
		f := nonNilTestFactory()
		var resp resource.ConfigureResponse
		r.Configure(ctx, resource.ConfigureRequest{ProviderData: f}, &resp)
		require.False(t, resp.Diagnostics.HasError())
	})

	t.Run("invalid_provider_data", func(t *testing.T) {
		t.Parallel()
		r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", defaultTestKibanaResourceOptions())
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
	req := resource.CreateRequest{Plan: plan, Config: kibanaTestConfig(plan)}
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
	opts := defaultTestKibanaResourceOptions()
	opts.Create = func(_ context.Context, _ *clients.KibanaScopedClient, _ KibanaWriteRequest[testKibanaResourceModel]) (KibanaWriteResult[testKibanaResourceModel], diag.Diagnostics) {
		createCalled = true
		return KibanaWriteResult[testKibanaResourceModel]{}, nil
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tftypes.NewValue(tftypes.String, tftypes.UnknownValue))
	objType := testKibanaResourceObjectType()
	respState := tfsdk.State{
		Raw:    tftypes.NewValue(objType, nil),
		Schema: testKibanaResourceSchemaWithConnectionBlock(ctx),
	}
	req := resource.CreateRequest{Plan: plan, Config: kibanaTestConfig(plan)}
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
	opts := defaultTestKibanaResourceOptions()
	opts.Create = func(_ context.Context, _ *clients.KibanaScopedClient, _ KibanaWriteRequest[testKibanaResourceModel]) (KibanaWriteResult[testKibanaResourceModel], diag.Diagnostics) {
		createCalled = true
		return KibanaWriteResult[testKibanaResourceModel]{}, nil
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tftypes.NewValue(tftypes.String, nil))
	objType := testKibanaResourceObjectType()
	respState := tfsdk.State{
		Raw:    tftypes.NewValue(objType, nil),
		Schema: testKibanaResourceSchemaWithConnectionBlock(ctx),
	}
	req := resource.CreateRequest{Plan: plan, Config: kibanaTestConfig(plan)}
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
	opts := defaultTestKibanaResourceOptions()
	opts.Create = func(_ context.Context, _ *clients.KibanaScopedClient, _ KibanaWriteRequest[testKibanaResourceModel]) (KibanaWriteResult[testKibanaResourceModel], diag.Diagnostics) {
		createCalled = true
		return KibanaWriteResult[testKibanaResourceModel]{}, nil
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	objType := testKibanaResourceObjectType()
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"name":              tftypes.NewValue(tftypes.String, "my-resource"),
		"space_id":          tftypes.NewValue(tftypes.String, ""),
		"kibana_connection": tftypes.NewValue(kibanaConnectionBlockType(), nil),
		"timeouts":          resourceTimeoutsNullValue(),
	})
	plan := tfsdk.Plan{Raw: objValue, Schema: testKibanaResourceSchemaWithConnectionBlock(ctx)}
	respState := tfsdk.State{
		Raw:    tftypes.NewValue(objType, nil),
		Schema: testKibanaResourceSchemaWithConnectionBlock(ctx),
	}
	req := resource.CreateRequest{Plan: plan, Config: kibanaTestConfig(plan)}
	resp := resource.CreateResponse{State: respState}

	r.Create(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Invalid space identifier")
	require.False(t, createCalled, "create callback should not run when spaceID is empty")
}

func TestNewKibanaResource_Create_allowsEmptySpaceWhenUnscoped(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	createCalled := false
	opts := KibanaResourceOptions[testKibanaResourceModelUnscoped]{
		Schema: getTestKibanaResourceSchema,
		Read:   testKibanaReadFuncUnscoped,
		Delete: testKibanaDeleteFuncUnscoped,
		Create: func(_ context.Context, _ *clients.KibanaScopedClient, req KibanaWriteRequest[testKibanaResourceModelUnscoped]) (KibanaWriteResult[testKibanaResourceModelUnscoped], diag.Diagnostics) {
			createCalled = true
			return testKibanaWriteFuncUnscoped(ctx, nil, req)
		},
		Update: testKibanaWriteFuncUnscoped,
	}
	r := NewKibanaResource[testKibanaResourceModelUnscoped](ComponentKibana, "test_entity", opts)
	r.client = factory

	objType := testKibanaResourceObjectType()
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"name":              tftypes.NewValue(tftypes.String, "my-resource"),
		"space_id":          tftypes.NewValue(tftypes.String, ""),
		"kibana_connection": tftypes.NewValue(kibanaConnectionBlockType(), nil),
		"timeouts":          resourceTimeoutsNullValue(),
	})
	plan := tfsdk.Plan{Raw: objValue, Schema: testKibanaResourceSchemaWithConnectionBlock(ctx)}
	respState := tfsdk.State{
		Raw:    tftypes.NewValue(objType, nil),
		Schema: testKibanaResourceSchemaWithConnectionBlock(ctx),
	}
	req := resource.CreateRequest{Plan: plan, Config: kibanaTestConfig(plan)}
	resp := resource.CreateResponse{State: respState}

	r.Create(ctx, req, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.True(t, createCalled, "create should run when spaceID is empty and model opts into KibanaUnscopedSpace")
	var result testKibanaResourceModelUnscoped
	diags := resp.State.Get(ctx, &result)
	require.False(t, diags.HasError())
	require.Equal(t, "my-resource", result.ID.ValueString())
}

func TestNewKibanaResource_Create_shortCircuitClientError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := nonNilTestFactory()
	createCalled := false
	opts := defaultTestKibanaResourceOptions()
	opts.Create = func(_ context.Context, _ *clients.KibanaScopedClient, _ KibanaWriteRequest[testKibanaResourceModel]) (KibanaWriteResult[testKibanaResourceModel], diag.Diagnostics) {
		createCalled = true
		return KibanaWriteResult[testKibanaResourceModel]{}, nil
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tftypes.NewValue(tftypes.String, "default"))
	objType := testKibanaResourceObjectType()
	respState := tfsdk.State{
		Raw:    tftypes.NewValue(objType, nil),
		Schema: testKibanaResourceSchemaWithConnectionBlock(ctx),
	}
	req := resource.CreateRequest{Plan: plan, Config: kibanaTestConfig(plan)}
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
	opts := defaultTestKibanaResourceOptions()
	opts.Create = func(_ context.Context, _ *clients.KibanaScopedClient, _ KibanaWriteRequest[testKibanaResourceModel]) (KibanaWriteResult[testKibanaResourceModel], diag.Diagnostics) {
		var diags diag.Diagnostics
		diags.AddError("create error", "something went wrong")
		return KibanaWriteResult[testKibanaResourceModel]{}, diags
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tftypes.NewValue(tftypes.String, "default"))
	objType := testKibanaResourceObjectType()
	respState := tfsdk.State{
		Raw:    tftypes.NewValue(objType, nil),
		Schema: testKibanaResourceSchemaWithConnectionBlock(ctx),
	}
	req := resource.CreateRequest{Plan: plan, Config: kibanaTestConfig(plan)}
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
	opts := defaultTestKibanaResourceOptions()
	opts.Create = func(_ context.Context, _ *clients.KibanaScopedClient, _ KibanaWriteRequest[testKibanaResourceModel]) (KibanaWriteResult[testKibanaResourceModel], diag.Diagnostics) {
		createCalled = true
		return KibanaWriteResult[testKibanaResourceModel]{}, nil
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	objType := testKibanaResourceObjectType()
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"name":              tftypes.NewValue(tftypes.String, "my-resource"),
		"space_id":          tftypes.NewValue(tftypes.String, "default"),
		"kibana_connection": tftypes.NewValue(kibanaConnectionBlockType(), nil),
		"timeouts":          resourceTimeoutsNullValue(),
	})
	badSchema := getTestKibanaResourceSchema(context.Background())
	plan := tfsdk.Plan{Raw: objValue, Schema: badSchema}
	respState := tfsdk.State{
		Raw:    tftypes.NewValue(objType, nil),
		Schema: testKibanaResourceSchemaWithConnectionBlock(ctx),
	}
	req := resource.CreateRequest{Plan: plan, Config: kibanaTestConfig(plan)}
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
	var nilCreate KibanaWriteFunc[testKibanaResourceModel]
	opts := defaultTestKibanaResourceOptions()
	opts.Create = nilCreate
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tftypes.NewValue(tftypes.String, "default"))
	objType := testKibanaResourceObjectType()
	respState := tfsdk.State{
		Raw:    tftypes.NewValue(objType, nil),
		Schema: testKibanaResourceSchemaWithConnectionBlock(ctx),
	}
	req := resource.CreateRequest{Plan: plan, Config: kibanaTestConfig(plan)}
	resp := resource.CreateResponse{State: respState}

	r.Create(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Kibana envelope configuration error")
}

func TestNewKibanaResource_Create_placeholderCallbackError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	placeholder := PlaceholderKibanaWriteCallback[testKibanaResourceModel]()
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", KibanaResourceOptions[testKibanaResourceModel]{
		Schema: getTestKibanaResourceSchema,
		Read:   testKibanaReadFuncFound,
		Delete: testKibanaDeleteFunc,
		Create: placeholder,
		Update: placeholder,
	})
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tftypes.NewValue(tftypes.String, "default"))
	objType := testKibanaResourceObjectType()
	respState := tfsdk.State{
		Raw:    tftypes.NewValue(objType, nil),
		Schema: testKibanaResourceSchemaWithConnectionBlock(ctx),
	}
	req := resource.CreateRequest{Plan: plan, Config: kibanaTestConfig(plan)}
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
		var nilCreate KibanaWriteFunc[testKibanaResourceModel]
		opts := defaultTestKibanaResourceOptions()
		opts.Create = nilCreate
		r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
		r.client = nonNilTestFactory()

		plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tftypes.NewValue(tftypes.String, "default"))
		respState := tfsdk.State{
			Raw:    tftypes.NewValue(testKibanaResourceObjectType(), nil),
			Schema: testKibanaResourceSchemaWithConnectionBlock(ctx),
		}
		resp := resource.CreateResponse{State: respState}
		r.Create(ctx, resource.CreateRequest{Plan: plan, Config: kibanaTestConfig(plan)}, &resp)

		require.True(t, resp.Diagnostics.HasError())
		require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Kibana envelope configuration error")
		require.NotContains(t, resp.Diagnostics.Errors()[0].Summary(), "Provider not configured")
	})

	t.Run("Create_precedesInvalidSpaceID", func(t *testing.T) {
		t.Parallel()
		var nilCreate KibanaWriteFunc[testKibanaResourceModel]
		opts := defaultTestKibanaResourceOptions()
		opts.Create = nilCreate
		r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
		r.client = newTestConfiguredFactory(ctx, t)

		objType := testKibanaResourceObjectType()
		objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
			"id":                tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			"name":              tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			"space_id":          tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			"kibana_connection": tftypes.NewValue(kibanaConnectionBlockType(), nil),
			"timeouts":          resourceTimeoutsNullValue(),
		})
		plan := tfsdk.Plan{Raw: objValue, Schema: testKibanaResourceSchemaWithConnectionBlock(ctx)}
		respState := tfsdk.State{
			Raw:    tftypes.NewValue(objType, nil),
			Schema: testKibanaResourceSchemaWithConnectionBlock(ctx),
		}
		resp := resource.CreateResponse{State: respState}
		r.Create(ctx, resource.CreateRequest{Plan: plan, Config: kibanaTestConfig(plan)}, &resp)

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
	opts := defaultTestKibanaResourceOptions()
	opts.Read = func(_ context.Context, _ *clients.KibanaScopedClient, resourceID string, spaceID string, model testKibanaResourceModel) (testKibanaResourceModel, bool, diag.Diagnostics) {
		readCalled = true
		receivedResourceID = resourceID
		receivedSpaceID = spaceID
		return model, true, nil
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
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
	opts := defaultTestKibanaResourceOptions()
	opts.Read = func(_ context.Context, _ *clients.KibanaScopedClient, resourceID string, spaceID string, model testKibanaResourceModel) (testKibanaResourceModel, bool, diag.Diagnostics) {
		readCalled = true
		receivedResourceID = resourceID
		receivedSpaceID = spaceID
		return model, true, nil
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	connBlockType := kibanaConnectionBlockType()
	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":                tftypes.String,
			"name":              tftypes.String,
			"space_id":          tftypes.String,
			"kibana_connection": connBlockType,
			"timeouts":          resourceTimeoutsObjectType(),
		},
	}
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, "not-a-composite-id"),
		"name":              tftypes.NewValue(tftypes.String, "fallback-name"),
		"space_id":          tftypes.NewValue(tftypes.String, "fallback-space"),
		"kibana_connection": tftypes.NewValue(connBlockType, nil),
		"timeouts":          resourceTimeoutsNullValue(),
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
	opts := defaultTestKibanaResourceOptions()
	opts.Read = func(_ context.Context, _ *clients.KibanaScopedClient, resourceID string, spaceID string, model testKibanaResourceModel) (testKibanaResourceModel, bool, diag.Diagnostics) {
		readCalled = true
		receivedResourceID = resourceID
		receivedSpaceID = spaceID
		return model, true, nil
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
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
			"timeouts":          resourceTimeoutsObjectType(),
		},
	}
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, "foo/bar"),
		"name":              tftypes.NewValue(tftypes.String, "different-name"),
		"space_id":          tftypes.NewValue(tftypes.String, "different-space"),
		"kibana_connection": tftypes.NewValue(connBlockType, nil),
		"timeouts":          resourceTimeoutsNullValue(),
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

func TestNewKibanaResource_Update_writeIDFromResolvedIdentity(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	updateCalled := false
	var receivedWriteID, receivedSpaceID string
	opts := defaultTestKibanaResourceOptions()
	opts.Update = func(_ context.Context, _ *clients.KibanaScopedClient, req KibanaWriteRequest[testKibanaResourceModel]) (KibanaWriteResult[testKibanaResourceModel], diag.Diagnostics) {
		updateCalled = true
		receivedWriteID = req.WriteID
		receivedSpaceID = req.SpaceID
		return KibanaWriteResult[testKibanaResourceModel]{Model: req.Plan}, nil
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	connBlockType := kibanaConnectionBlockType()
	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":                tftypes.String,
			"name":              tftypes.String,
			"space_id":          tftypes.String,
			"kibana_connection": connBlockType,
			"timeouts":          resourceTimeoutsObjectType(),
		},
	}
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, "foo/bar"),
		"name":              tftypes.NewValue(tftypes.String, "different-name"),
		"space_id":          tftypes.NewValue(tftypes.String, "different-space"),
		"kibana_connection": tftypes.NewValue(connBlockType, nil),
		"timeouts":          resourceTimeoutsNullValue(),
	})
	plan := tfsdk.Plan{Raw: objValue, Schema: testKibanaResourceSchemaWithConnectionBlock(ctx)}
	prior := makeTestKibanaResourceState(ctx, t, "foo/bar")
	resp := resource.UpdateResponse{State: prior}
	req := resource.UpdateRequest{Plan: plan, State: prior, Config: kibanaTestConfig(plan)}

	r.Update(ctx, req, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.True(t, updateCalled, "update callback should be called")
	require.Equal(t, "bar", receivedWriteID, "update WriteID should come from resolved composite identity")
	require.Equal(t, "foo", receivedSpaceID, "update SpaceID should come from resolved composite identity")
}

// =============================================================================
// Subtask 2.11: Read not-found removes resource from state
// =============================================================================

func TestNewKibanaResource_Read_notFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	opts := defaultTestKibanaResourceOptions()
	opts.Read = func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModel) (testKibanaResourceModel, bool, diag.Diagnostics) {
		return testKibanaResourceModel{}, false, nil
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
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
	opts := defaultTestKibanaResourceOptions()
	opts.Read = func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModel) (testKibanaResourceModel, bool, diag.Diagnostics) {
		readCalled = true
		return testKibanaResourceModel{}, false, nil
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":                tftypes.String,
			"name":              tftypes.String,
			"space_id":          tftypes.String,
			"kibana_connection": kibanaConnectionBlockType(),
			"timeouts":          resourceTimeoutsObjectType(),
		},
	}
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, "default/my-stream"),
		"name":              tftypes.NewValue(tftypes.String, "my-stream"),
		"space_id":          tftypes.NewValue(tftypes.String, "default"),
		"kibana_connection": tftypes.NewValue(kibanaConnectionBlockType(), nil),
		"timeouts":          resourceTimeoutsNullValue(),
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
	opts := defaultTestKibanaResourceOptions()
	opts.Read = func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModel) (testKibanaResourceModel, bool, diag.Diagnostics) {
		readCalled = true
		return testKibanaResourceModel{}, false, nil
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	// Composite parse fails (no slash) and GetResourceID returns empty.
	connBlockType := kibanaConnectionBlockType()
	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":                tftypes.String,
			"name":              tftypes.String,
			"space_id":          tftypes.String,
			"kibana_connection": connBlockType,
			"timeouts":          resourceTimeoutsObjectType(),
		},
	}
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, "plain-id-no-composite"),
		"name":              tftypes.NewValue(tftypes.String, ""),
		"space_id":          tftypes.NewValue(tftypes.String, "default"),
		"kibana_connection": tftypes.NewValue(connBlockType, nil),
		"timeouts":          resourceTimeoutsNullValue(),
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
	opts := defaultTestKibanaResourceOptions()
	opts.Read = func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModel) (testKibanaResourceModel, bool, diag.Diagnostics) {
		readCalled = true
		return testKibanaResourceModel{}, false, nil
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
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
	opts := defaultTestKibanaResourceOptions()
	opts.Read = func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModel) (testKibanaResourceModel, bool, diag.Diagnostics) {
		var diags diag.Diagnostics
		diags.AddError("read error", "something went wrong")
		return testKibanaResourceModel{}, false, diags
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
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
	opts := defaultTestKibanaResourceOptions()
	opts.Read = nilRead
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
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
	req := resource.UpdateRequest{Plan: plan, State: prior, Config: kibanaTestConfig(plan)}

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
	opts := defaultTestKibanaResourceOptions()
	opts.Update = func(_ context.Context, _ *clients.KibanaScopedClient, req KibanaWriteRequest[testKibanaResourceModel]) (KibanaWriteResult[testKibanaResourceModel], diag.Diagnostics) {
		updateCalled = true
		receivedPlan = req.Plan
		if req.Prior != nil {
			receivedPrior = *req.Prior
		}
		return KibanaWriteResult[testKibanaResourceModel]{Model: req.Plan}, nil
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "default/my-resource"), tftypes.NewValue(tftypes.String, "default"))
	prior := makeTestKibanaResourceState(ctx, t, "default/my-resource")
	resp := resource.UpdateResponse{State: prior}
	req := resource.UpdateRequest{Plan: plan, State: prior, Config: kibanaTestConfig(plan)}

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
	opts := defaultTestKibanaResourceOptions()
	opts.Update = func(_ context.Context, _ *clients.KibanaScopedClient, req KibanaWriteRequest[testKibanaResourceModel]) (KibanaWriteResult[testKibanaResourceModel], diag.Diagnostics) {
		updateCalled = true
		receivedResourceID = req.WriteID
		receivedSpaceID = req.SpaceID
		return KibanaWriteResult[testKibanaResourceModel]{Model: req.Plan}, nil
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "abc-uuid"), tftypes.NewValue(tftypes.String, "custom-space"))
	prior := makeTestKibanaResourceState(ctx, t, "abc-uuid")
	resp := resource.UpdateResponse{State: prior}
	req := resource.UpdateRequest{Plan: plan, State: prior, Config: kibanaTestConfig(plan)}

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
	opts := defaultTestKibanaResourceOptions()
	opts.Update = func(_ context.Context, _ *clients.KibanaScopedClient, _ KibanaWriteRequest[testKibanaResourceModel]) (KibanaWriteResult[testKibanaResourceModel], diag.Diagnostics) {
		updateCalled = true
		return KibanaWriteResult[testKibanaResourceModel]{}, nil
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	objType := testKibanaResourceObjectType()
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"name":              tftypes.NewValue(tftypes.String, "my-resource"),
		"space_id":          tftypes.NewValue(tftypes.String, "default"),
		"kibana_connection": tftypes.NewValue(kibanaConnectionBlockType(), nil),
		"timeouts":          resourceTimeoutsNullValue(),
	})
	badSchema := getTestKibanaResourceSchema(context.Background())
	plan := tfsdk.Plan{Raw: objValue, Schema: badSchema}
	prior := makeTestKibanaResourceState(ctx, t, "default/my-resource")
	resp := resource.UpdateResponse{State: prior}
	req := resource.UpdateRequest{Plan: plan, State: prior, Config: kibanaTestConfig(plan)}

	r.Update(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.False(t, updateCalled, "update callback should not run when plan.Get fails")
}

func TestNewKibanaResource_Update_shortCircuitStateGetError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	updateCalled := false
	opts := defaultTestKibanaResourceOptions()
	opts.Update = func(_ context.Context, _ *clients.KibanaScopedClient, _ KibanaWriteRequest[testKibanaResourceModel]) (KibanaWriteResult[testKibanaResourceModel], diag.Diagnostics) {
		updateCalled = true
		return KibanaWriteResult[testKibanaResourceModel]{}, nil
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "default/my-resource"), tftypes.NewValue(tftypes.String, "default"))
	objType := testKibanaResourceObjectType()
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, "default/my-resource"),
		"name":              tftypes.NewValue(tftypes.String, "my-resource"),
		"space_id":          tftypes.NewValue(tftypes.String, "default"),
		"kibana_connection": tftypes.NewValue(kibanaConnectionBlockType(), nil),
		"timeouts":          resourceTimeoutsNullValue(),
	})
	badSchema := getTestKibanaResourceSchema(context.Background())
	prior := tfsdk.State{Raw: objValue, Schema: badSchema}
	resp := resource.UpdateResponse{State: prior}
	req := resource.UpdateRequest{Plan: plan, State: prior, Config: kibanaTestConfig(plan)}

	r.Update(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.False(t, updateCalled, "update callback should not run when state.Get fails")
}

func TestNewKibanaResource_Update_fallbackToPriorStateWhenPlanIdentityEmpty(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	updateCalled := false
	var receivedWriteID, receivedSpaceID string
	opts := defaultTestKibanaResourceOptions()
	opts.Update = func(_ context.Context, _ *clients.KibanaScopedClient, req KibanaWriteRequest[testKibanaResourceModel]) (KibanaWriteResult[testKibanaResourceModel], diag.Diagnostics) {
		updateCalled = true
		receivedWriteID = req.WriteID
		receivedSpaceID = req.SpaceID
		model := req.Plan
		model.ID = types.StringValue(req.SpaceID + "/" + req.WriteID)
		model.Name = types.StringValue(req.WriteID)
		return KibanaWriteResult[testKibanaResourceModel]{Model: model}, nil
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	objType := testKibanaResourceObjectType()
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"name":              tftypes.NewValue(tftypes.String, ""),
		"space_id":          tftypes.NewValue(tftypes.String, "default"),
		"kibana_connection": tftypes.NewValue(kibanaConnectionBlockType(), nil),
		"timeouts":          resourceTimeoutsNullValue(),
	})
	plan := tfsdk.Plan{Raw: objValue, Schema: testKibanaResourceSchemaWithConnectionBlock(ctx)}
	prior := makeTestKibanaResourceState(ctx, t, "default/old-resource")
	resp := resource.UpdateResponse{State: prior}
	req := resource.UpdateRequest{Plan: plan, State: prior, Config: kibanaTestConfig(plan)}

	r.Update(ctx, req, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.True(t, updateCalled, "update callback should be called when prior state has a valid identity")
	require.Equal(t, "old-resource", receivedWriteID, "update WriteID should fall back to prior state identity")
	require.Equal(t, "default", receivedSpaceID, "update SpaceID should fall back to prior state identity")
}

func TestNewKibanaResource_Update_shortCircuitClientError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := nonNilTestFactory()
	updateCalled := false
	opts := defaultTestKibanaResourceOptions()
	opts.Update = func(_ context.Context, _ *clients.KibanaScopedClient, _ KibanaWriteRequest[testKibanaResourceModel]) (KibanaWriteResult[testKibanaResourceModel], diag.Diagnostics) {
		updateCalled = true
		return KibanaWriteResult[testKibanaResourceModel]{}, nil
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "default/my-resource"), tftypes.NewValue(tftypes.String, "default"))
	prior := makeTestKibanaResourceState(ctx, t, "default/my-resource")
	resp := resource.UpdateResponse{State: prior}
	req := resource.UpdateRequest{Plan: plan, State: prior, Config: kibanaTestConfig(plan)}

	r.Update(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Provider not configured")
	require.False(t, updateCalled, "update callback should not run when client resolution fails")
}

func TestNewKibanaResource_Update_shortCircuitCallbackError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	opts := defaultTestKibanaResourceOptions()
	opts.Update = func(_ context.Context, _ *clients.KibanaScopedClient, _ KibanaWriteRequest[testKibanaResourceModel]) (KibanaWriteResult[testKibanaResourceModel], diag.Diagnostics) {
		var diags diag.Diagnostics
		diags.AddError("update error", "something went wrong")
		return KibanaWriteResult[testKibanaResourceModel]{}, diags
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "default/my-resource"), tftypes.NewValue(tftypes.String, "default"))
	prior := makeTestKibanaResourceState(ctx, t, "default/my-resource")
	resp := resource.UpdateResponse{State: prior}
	req := resource.UpdateRequest{Plan: plan, State: prior, Config: kibanaTestConfig(plan)}

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
	var nilUpdate KibanaWriteFunc[testKibanaResourceModel]
	opts := defaultTestKibanaResourceOptions()
	opts.Update = nilUpdate
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "default/my-resource"), tftypes.NewValue(tftypes.String, "default"))
	prior := makeTestKibanaResourceState(ctx, t, "default/my-resource")
	resp := resource.UpdateResponse{State: prior}
	req := resource.UpdateRequest{Plan: plan, State: prior, Config: kibanaTestConfig(plan)}

	r.Update(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Kibana envelope configuration error")
}

func TestNewKibanaResource_Update_placeholderCallbackError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	placeholder := PlaceholderKibanaWriteCallback[testKibanaResourceModel]()
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", KibanaResourceOptions[testKibanaResourceModel]{
		Schema: getTestKibanaResourceSchema,
		Read:   testKibanaReadFuncFound,
		Delete: testKibanaDeleteFunc,
		Create: placeholder,
		Update: placeholder,
	})
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "default/my-resource"), tftypes.NewValue(tftypes.String, "default"))
	prior := makeTestKibanaResourceState(ctx, t, "default/my-resource")
	resp := resource.UpdateResponse{State: prior}
	req := resource.UpdateRequest{Plan: plan, State: prior, Config: kibanaTestConfig(plan)}

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
		var nilUpdate KibanaWriteFunc[testKibanaResourceModel]
		opts := defaultTestKibanaResourceOptions()
		opts.Update = nilUpdate
		r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
		r.client = nonNilTestFactory()

		plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "default/my-resource"), tftypes.NewValue(tftypes.String, "default"))
		prior := makeTestKibanaResourceState(ctx, t, "default/my-resource")
		resp := resource.UpdateResponse{State: prior}
		r.Update(ctx, resource.UpdateRequest{Plan: plan, State: prior, Config: kibanaTestConfig(plan)}, &resp)

		require.True(t, resp.Diagnostics.HasError())
		require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Kibana envelope configuration error")
		require.NotContains(t, resp.Diagnostics.Errors()[0].Summary(), "Provider not configured")
	})

	t.Run("Update_precedesEmptyResourceID", func(t *testing.T) {
		t.Parallel()
		var nilUpdate KibanaWriteFunc[testKibanaResourceModel]
		opts := defaultTestKibanaResourceOptions()
		opts.Update = nilUpdate
		r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
		r.client = newTestConfiguredFactory(ctx, t)

		objType := testKibanaResourceObjectType()
		objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
			"id":                tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			"name":              tftypes.NewValue(tftypes.String, ""),
			"space_id":          tftypes.NewValue(tftypes.String, "default"),
			"kibana_connection": tftypes.NewValue(kibanaConnectionBlockType(), nil),
			"timeouts":          resourceTimeoutsNullValue(),
		})
		plan := tfsdk.Plan{Raw: objValue, Schema: testKibanaResourceSchemaWithConnectionBlock(ctx)}
		prior := makeTestKibanaResourceState(ctx, t, "default/old-resource")
		resp := resource.UpdateResponse{State: prior}
		r.Update(ctx, resource.UpdateRequest{Plan: plan, State: prior, Config: kibanaTestConfig(plan)}, &resp)

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
	opts := defaultTestKibanaResourceOptions()
	opts.Delete = func(_ context.Context, _ *clients.KibanaScopedClient, resourceID string, spaceID string, _ testKibanaResourceModel) diag.Diagnostics {
		deleteCalled = true
		receivedResourceID = resourceID
		receivedSpaceID = spaceID
		return nil
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
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
	opts := defaultTestKibanaResourceOptions()
	opts.Delete = func(_ context.Context, _ *clients.KibanaScopedClient, resourceID string, spaceID string, _ testKibanaResourceModel) diag.Diagnostics {
		deleteCalled = true
		receivedResourceID = resourceID
		receivedSpaceID = spaceID
		return nil
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
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
	opts := defaultTestKibanaResourceOptions()
	opts.Delete = func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModel) diag.Diagnostics {
		deleteCalled = true
		return nil
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":                tftypes.String,
			"name":              tftypes.String,
			"space_id":          tftypes.String,
			"kibana_connection": kibanaConnectionBlockType(),
			"timeouts":          resourceTimeoutsObjectType(),
		},
	}
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, "default/my-stream"),
		"name":              tftypes.NewValue(tftypes.String, "my-stream"),
		"space_id":          tftypes.NewValue(tftypes.String, "default"),
		"kibana_connection": tftypes.NewValue(kibanaConnectionBlockType(), nil),
		"timeouts":          resourceTimeoutsNullValue(),
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
	opts := defaultTestKibanaResourceOptions()
	opts.Delete = func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModel) diag.Diagnostics {
		deleteCalled = true
		return nil
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	connBlockType := kibanaConnectionBlockType()
	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":                tftypes.String,
			"name":              tftypes.String,
			"space_id":          tftypes.String,
			"kibana_connection": connBlockType,
			"timeouts":          resourceTimeoutsObjectType(),
		},
	}
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, "plain-id"),
		"name":              tftypes.NewValue(tftypes.String, ""),
		"space_id":          tftypes.NewValue(tftypes.String, "default"),
		"kibana_connection": tftypes.NewValue(connBlockType, nil),
		"timeouts":          resourceTimeoutsNullValue(),
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
	opts := defaultTestKibanaResourceOptions()
	opts.Delete = func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModel) diag.Diagnostics {
		deleteCalled = true
		return nil
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
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
	opts := defaultTestKibanaResourceOptions()
	opts.Delete = func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModel) diag.Diagnostics {
		var diags diag.Diagnostics
		diags.AddError("delete error", "something went wrong")
		return diags
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
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
	opts := defaultTestKibanaResourceOptions()
	opts.Delete = nilDelete
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	state := makeTestKibanaResourceState(ctx, t, "default/my-stream")
	req := resource.DeleteRequest{State: state}
	resp := resource.DeleteResponse{State: state}

	r.Delete(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Kibana envelope configuration error")
}

// =============================================================================
// Read-after-write and PostRead
// =============================================================================

func TestNewKibanaResource_Create_readAfterWriteUsesWrittenModelIdentity(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	var capturedResourceID, capturedSpaceID string
	createFn := func(
		_ context.Context,
		_ *clients.KibanaScopedClient,
		req KibanaWriteRequest[testKibanaResourceModel],
	) (KibanaWriteResult[testKibanaResourceModel], diag.Diagnostics) {
		model := req.Plan
		model.Name = types.StringValue("server-id")
		model.SpaceID = types.StringValue("server-space")
		return KibanaWriteResult[testKibanaResourceModel]{Model: model}, nil
	}
	readFn := func(
		_ context.Context,
		_ *clients.KibanaScopedClient,
		resourceID string,
		spaceID string,
		model testKibanaResourceModel,
	) (testKibanaResourceModel, bool, diag.Diagnostics) {
		capturedResourceID = resourceID
		capturedSpaceID = spaceID
		model.Name = types.StringValue(model.Name.ValueString() + "-refreshed")
		return model, true, nil
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", KibanaResourceOptions[testKibanaResourceModel]{
		Schema: getTestKibanaResourceSchema,
		Read:   readFn,
		Delete: testKibanaDeleteFunc,
		Create: createFn,
		Update: testKibanaWriteFuncFound,
	})
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tftypes.NewValue(tftypes.String, "plan-space"))
	objType := testKibanaResourceObjectType()
	respState := tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: testKibanaResourceSchemaWithConnectionBlock(ctx)}
	resp := resource.CreateResponse{State: respState}
	r.Create(ctx, resource.CreateRequest{Plan: plan, Config: kibanaTestConfig(plan)}, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.Equal(t, "server-id", capturedResourceID)
	require.Equal(t, "server-space", capturedSpaceID)
	var result testKibanaResourceModel
	require.False(t, resp.State.Get(ctx, &result).HasError())
	require.Equal(t, "server-id-refreshed", result.Name.ValueString())
}

func TestNewKibanaResource_Update_readAfterWriteUsesWrittenModelIdentity(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	var capturedResourceID, capturedSpaceID string
	updateFn := func(
		_ context.Context,
		_ *clients.KibanaScopedClient,
		req KibanaWriteRequest[testKibanaResourceModel],
	) (KibanaWriteResult[testKibanaResourceModel], diag.Diagnostics) {
		model := req.Plan
		model.Name = types.StringValue("server-id")
		model.SpaceID = types.StringValue("server-space")
		return KibanaWriteResult[testKibanaResourceModel]{Model: model}, nil
	}
	readFn := func(
		_ context.Context,
		_ *clients.KibanaScopedClient,
		resourceID string,
		spaceID string,
		model testKibanaResourceModel,
	) (testKibanaResourceModel, bool, diag.Diagnostics) {
		capturedResourceID = resourceID
		capturedSpaceID = spaceID
		model.Name = types.StringValue(model.Name.ValueString() + "-refreshed")
		return model, true, nil
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", KibanaResourceOptions[testKibanaResourceModel]{
		Schema: getTestKibanaResourceSchema,
		Read:   readFn,
		Delete: testKibanaDeleteFunc,
		Create: testKibanaWriteFuncFound,
		Update: updateFn,
	})
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "default/my-resource"), tftypes.NewValue(tftypes.String, "plan-space"))
	prior := makeTestKibanaResourceState(ctx, t, "default/my-resource")
	resp := resource.UpdateResponse{State: prior}
	r.Update(ctx, resource.UpdateRequest{Plan: plan, State: prior, Config: kibanaTestConfig(plan)}, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.Equal(t, "server-id", capturedResourceID)
	require.Equal(t, "server-space", capturedSpaceID)
	var result testKibanaResourceModel
	require.False(t, resp.State.Get(ctx, &result).HasError())
	require.Equal(t, "server-id-refreshed", result.Name.ValueString())
}

func TestNewKibanaResource_Create_readAfterWriteHappyPath(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", KibanaResourceOptions[testKibanaResourceModel]{
		Schema: getTestKibanaResourceSchema,
		Read: func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, model testKibanaResourceModel) (testKibanaResourceModel, bool, diag.Diagnostics) {
			model.Name = types.StringValue("from-readfunc")
			return model, true, nil
		},
		Delete: testKibanaDeleteFunc,
		Create: testKibanaWriteFuncFound,
		Update: testKibanaWriteFuncFound,
	})
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tftypes.NewValue(tftypes.String, "default"))
	objType := testKibanaResourceObjectType()
	respState := tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: testKibanaResourceSchemaWithConnectionBlock(ctx)}
	resp := resource.CreateResponse{State: respState}
	r.Create(ctx, resource.CreateRequest{Plan: plan, Config: kibanaTestConfig(plan)}, &resp)

	require.False(t, resp.Diagnostics.HasError())
	var result testKibanaResourceModel
	diags := resp.State.Get(ctx, &result)
	require.False(t, diags.HasError())
	require.Equal(t, "default/my-resource", result.ID.ValueString())
	require.Equal(t, "from-readfunc", result.Name.ValueString())
}

func TestNewKibanaResource_Update_readAfterWriteHappyPath(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", KibanaResourceOptions[testKibanaResourceModel]{
		Schema: getTestKibanaResourceSchema,
		Read:   testKibanaReadFuncDistinguishing,
		Delete: testKibanaDeleteFunc,
		Create: testKibanaWriteFuncFound,
		Update: testKibanaWriteFuncFound,
	})
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "default/my-resource"), tftypes.NewValue(tftypes.String, "default"))
	prior := makeTestKibanaResourceState(ctx, t, "default/my-resource")
	resp := resource.UpdateResponse{State: prior}
	r.Update(ctx, resource.UpdateRequest{Plan: plan, State: prior, Config: kibanaTestConfig(plan)}, &resp)

	require.False(t, resp.Diagnostics.HasError())
	var result testKibanaResourceModel
	diags := resp.State.Get(ctx, &result)
	require.False(t, diags.HasError())
	require.Equal(t, "my-resource-read", result.Name.ValueString())
}

func TestNewKibanaResource_Create_notFoundAfterWrite(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	createFn := func(
		_ context.Context,
		_ *clients.KibanaScopedClient,
		req KibanaWriteRequest[testKibanaResourceModel],
	) (KibanaWriteResult[testKibanaResourceModel], diag.Diagnostics) {
		model := req.Plan
		model.Name = types.StringValue("server-id")
		model.SpaceID = types.StringValue("server-space")
		return KibanaWriteResult[testKibanaResourceModel]{Model: model}, nil
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", KibanaResourceOptions[testKibanaResourceModel]{
		Schema: getTestKibanaResourceSchema,
		Read: func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModel) (testKibanaResourceModel, bool, diag.Diagnostics) {
			return testKibanaResourceModel{}, false, nil
		},
		Delete: testKibanaDeleteFunc,
		Create: createFn,
		Update: testKibanaWriteFuncFound,
	})
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tftypes.NewValue(tftypes.String, "plan-space"))
	objType := testKibanaResourceObjectType()
	respState := tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: testKibanaResourceSchemaWithConnectionBlock(ctx)}
	resp := resource.CreateResponse{State: respState}
	r.Create(ctx, resource.CreateRequest{Plan: plan, Config: kibanaTestConfig(plan)}, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Resource not found")
	require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), `kibana_test_entity "server-id" in space "server-space" was not found after write`)
	require.NotContains(t, resp.Diagnostics.Errors()[0].Detail(), "my-resource")
	require.True(t, resp.State.Raw.IsNull(), "state should not be mutated when resource not found after create")
}

func TestNewKibanaResource_Update_notFoundAfterWrite(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	updateFn := func(
		_ context.Context,
		_ *clients.KibanaScopedClient,
		req KibanaWriteRequest[testKibanaResourceModel],
	) (KibanaWriteResult[testKibanaResourceModel], diag.Diagnostics) {
		model := req.Plan
		model.Name = types.StringValue("server-id")
		model.SpaceID = types.StringValue("server-space")
		return KibanaWriteResult[testKibanaResourceModel]{Model: model}, nil
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", KibanaResourceOptions[testKibanaResourceModel]{
		Schema: getTestKibanaResourceSchema,
		Read: func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModel) (testKibanaResourceModel, bool, diag.Diagnostics) {
			return testKibanaResourceModel{}, false, nil
		},
		Delete: testKibanaDeleteFunc,
		Create: testKibanaWriteFuncFound,
		Update: updateFn,
	})
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "default/my-resource"), tftypes.NewValue(tftypes.String, "plan-space"))
	prior := makeTestKibanaResourceState(ctx, t, "default/my-resource")
	resp := resource.UpdateResponse{State: prior}
	r.Update(ctx, resource.UpdateRequest{Plan: plan, State: prior, Config: kibanaTestConfig(plan)}, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Resource not found")
	require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), `kibana_test_entity "server-id" in space "server-space" was not found after write`)
	require.False(t, resp.State.Get(ctx, &testKibanaResourceModel{}).HasError())
}

func TestNewKibanaResource_Create_emptyReadSpaceIDAfterWriteScopedResource(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	readCalled := false
	createFn := func(
		_ context.Context,
		_ *clients.KibanaScopedClient,
		req KibanaWriteRequest[testKibanaResourceModel],
	) (KibanaWriteResult[testKibanaResourceModel], diag.Diagnostics) {
		model := req.Plan
		model.Name = types.StringValue("server-id")
		model.SpaceID = types.StringValue("")
		return KibanaWriteResult[testKibanaResourceModel]{Model: model}, nil
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", KibanaResourceOptions[testKibanaResourceModel]{
		Schema: getTestKibanaResourceSchema,
		Read: func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModel) (testKibanaResourceModel, bool, diag.Diagnostics) {
			readCalled = true
			return testKibanaResourceModel{}, false, nil
		},
		Delete: testKibanaDeleteFunc,
		Create: createFn,
		Update: testKibanaWriteFuncFound,
	})
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tftypes.NewValue(tftypes.String, "default"))
	objType := testKibanaResourceObjectType()
	respState := tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: testKibanaResourceSchemaWithConnectionBlock(ctx)}
	resp := resource.CreateResponse{State: respState}
	r.Create(ctx, resource.CreateRequest{Plan: plan, Config: kibanaTestConfig(plan)}, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Invalid space identifier")
	require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "resolved read space is empty after write")
	require.False(t, readCalled, "readFunc should not run when read space is empty on a scoped resource")
	require.True(t, resp.State.Raw.IsNull(), "state should not be mutated")
}

func TestNewKibanaResource_Update_emptyReadSpaceIDAfterWriteScopedResource(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	readCalled := false
	updateFn := func(
		_ context.Context,
		_ *clients.KibanaScopedClient,
		req KibanaWriteRequest[testKibanaResourceModel],
	) (KibanaWriteResult[testKibanaResourceModel], diag.Diagnostics) {
		model := req.Plan
		model.Name = types.StringValue("server-id")
		model.SpaceID = types.StringValue("")
		return KibanaWriteResult[testKibanaResourceModel]{Model: model}, nil
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", KibanaResourceOptions[testKibanaResourceModel]{
		Schema: getTestKibanaResourceSchema,
		Read: func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModel) (testKibanaResourceModel, bool, diag.Diagnostics) {
			readCalled = true
			return testKibanaResourceModel{}, false, nil
		},
		Delete: testKibanaDeleteFunc,
		Create: testKibanaWriteFuncFound,
		Update: updateFn,
	})
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "default/my-resource"), tftypes.NewValue(tftypes.String, "default"))
	prior := makeTestKibanaResourceState(ctx, t, "default/my-resource")
	resp := resource.UpdateResponse{State: prior}
	r.Update(ctx, resource.UpdateRequest{Plan: plan, State: prior, Config: kibanaTestConfig(plan)}, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Invalid space identifier")
	require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "resolved read space is empty after write")
	require.False(t, readCalled, "readFunc should not run when read space is empty on a scoped resource")
	var after testKibanaResourceModel
	require.False(t, resp.State.Get(ctx, &after).HasError())
	require.Equal(t, "default/my-resource", after.ID.ValueString(), "state should not change")
}

func TestNewKibanaResource_Create_emptyReadSpaceIDAfterWriteUnscopedResource(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	readCalled := false
	createFn := func(
		_ context.Context,
		_ *clients.KibanaScopedClient,
		req KibanaWriteRequest[testKibanaResourceModelUnscoped],
	) (KibanaWriteResult[testKibanaResourceModelUnscoped], diag.Diagnostics) {
		plan := req.Plan
		plan.SpaceID = types.StringValue("")
		return KibanaWriteResult[testKibanaResourceModelUnscoped]{Model: plan}, nil
	}
	r := NewKibanaResource[testKibanaResourceModelUnscoped](ComponentKibana, "test_entity", KibanaResourceOptions[testKibanaResourceModelUnscoped]{
		Schema: getTestKibanaResourceSchema,
		Read: func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, model testKibanaResourceModelUnscoped) (testKibanaResourceModelUnscoped, bool, diag.Diagnostics) {
			readCalled = true
			return model, true, nil
		},
		Delete: testKibanaDeleteFuncUnscoped,
		Create: createFn,
		Update: testKibanaWriteFuncUnscoped,
	})
	r.client = factory

	objType := testKibanaResourceObjectType()
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"name":              tftypes.NewValue(tftypes.String, "my-resource"),
		"space_id":          tftypes.NewValue(tftypes.String, ""),
		"kibana_connection": tftypes.NewValue(kibanaConnectionBlockType(), nil),
		"timeouts":          resourceTimeoutsNullValue(),
	})
	plan := tfsdk.Plan{Raw: objValue, Schema: testKibanaResourceSchemaWithConnectionBlock(ctx)}
	respState := tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: testKibanaResourceSchemaWithConnectionBlock(ctx)}
	resp := resource.CreateResponse{State: respState}
	r.Create(ctx, resource.CreateRequest{Plan: plan, Config: kibanaTestConfig(plan)}, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.True(t, readCalled, "readFunc should run when empty read space is allowed for KibanaUnscopedSpace")
}

func TestNewKibanaResource_Create_readFuncErrorAfterWrite(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", KibanaResourceOptions[testKibanaResourceModel]{
		Schema: getTestKibanaResourceSchema,
		Read: func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModel) (testKibanaResourceModel, bool, diag.Diagnostics) {
			var diags diag.Diagnostics
			diags.AddError("read error", "something went wrong after create")
			return testKibanaResourceModel{}, false, diags
		},
		Delete: testKibanaDeleteFunc,
		Create: testKibanaWriteFuncFound,
		Update: testKibanaWriteFuncFound,
	})
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tftypes.NewValue(tftypes.String, "default"))
	objType := testKibanaResourceObjectType()
	respState := tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: testKibanaResourceSchemaWithConnectionBlock(ctx)}
	resp := resource.CreateResponse{State: respState}
	r.Create(ctx, resource.CreateRequest{Plan: plan, Config: kibanaTestConfig(plan)}, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "read error")
	require.True(t, resp.State.Raw.IsNull(), "state should not be mutated when readFunc returns errors after create")
}

func TestNewKibanaResource_Update_readFuncErrorAfterWrite(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", KibanaResourceOptions[testKibanaResourceModel]{
		Schema: getTestKibanaResourceSchema,
		Read: func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModel) (testKibanaResourceModel, bool, diag.Diagnostics) {
			var diags diag.Diagnostics
			diags.AddError("read error", "something went wrong after update")
			return testKibanaResourceModel{}, false, diags
		},
		Delete: testKibanaDeleteFunc,
		Create: testKibanaWriteFuncFound,
		Update: testKibanaWriteFuncFound,
	})
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "default/my-resource"), tftypes.NewValue(tftypes.String, "default"))
	prior := makeTestKibanaResourceState(ctx, t, "default/my-resource")
	resp := resource.UpdateResponse{State: prior}
	r.Update(ctx, resource.UpdateRequest{Plan: plan, State: prior, Config: kibanaTestConfig(plan)}, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "read error")
	var after testKibanaResourceModel
	require.False(t, resp.State.Get(ctx, &after).HasError())
	require.Equal(t, "default/my-resource", after.ID.ValueString())
}

func TestNewKibanaResource_Update_skipReadAfterWriteFromWriteResult(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	readCalled := false
	opts := defaultTestKibanaResourceOptions()
	opts.Read = func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, model testKibanaResourceModel) (testKibanaResourceModel, bool, diag.Diagnostics) {
		readCalled = true
		model.Name = types.StringValue(model.Name.ValueString() + "-read")
		return model, true, nil
	}
	opts.Update = func(_ context.Context, _ *clients.KibanaScopedClient, req KibanaWriteRequest[testKibanaResourceModel]) (KibanaWriteResult[testKibanaResourceModel], diag.Diagnostics) {
		model := req.Plan
		model.Name = types.StringValue("from-write-callback")
		return KibanaWriteResult[testKibanaResourceModel]{
			Model:              model,
			SkipReadAfterWrite: true,
		}, nil
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "default/my-resource"), tftypes.NewValue(tftypes.String, "default"))
	prior := makeTestKibanaResourceState(ctx, t, "default/my-resource")
	resp := resource.UpdateResponse{State: prior}
	r.Update(ctx, resource.UpdateRequest{Plan: plan, State: prior, Config: kibanaTestConfig(plan)}, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.False(t, readCalled, "read callback must not run when write result sets SkipReadAfterWrite")
	var after testKibanaResourceModel
	require.False(t, resp.State.Get(ctx, &after).HasError())
	require.Equal(t, "from-write-callback", after.Name.ValueString())
}

func TestNewKibanaResource_Update_readAfterWriteByDefault(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	readCalled := false
	opts := defaultTestKibanaResourceOptions()
	opts.Read = func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, model testKibanaResourceModel) (testKibanaResourceModel, bool, diag.Diagnostics) {
		readCalled = true
		model.Name = types.StringValue("from-read")
		return model, true, nil
	}
	opts.Update = func(_ context.Context, _ *clients.KibanaScopedClient, req KibanaWriteRequest[testKibanaResourceModel]) (KibanaWriteResult[testKibanaResourceModel], diag.Diagnostics) {
		return KibanaWriteResult[testKibanaResourceModel]{Model: req.Plan}, nil
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "default/my-resource"), tftypes.NewValue(tftypes.String, "default"))
	prior := makeTestKibanaResourceState(ctx, t, "default/my-resource")
	resp := resource.UpdateResponse{State: prior}
	r.Update(ctx, resource.UpdateRequest{Plan: plan, State: prior, Config: kibanaTestConfig(plan)}, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.True(t, readCalled, "read callback must run when SkipReadAfterWrite is false")
	var after testKibanaResourceModel
	require.False(t, resp.State.Get(ctx, &after).HasError())
	require.Equal(t, "from-read", after.Name.ValueString())
}

func TestNewKibanaResource_Update_skipReadAfterWriteSkipsPostRead(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	readCalled := false
	postReadCalled := false
	opts := defaultTestKibanaResourceOptions()
	opts.Read = func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, model testKibanaResourceModel) (testKibanaResourceModel, bool, diag.Diagnostics) {
		readCalled = true
		return model, true, nil
	}
	opts.PostRead = func(_ context.Context, req KibanaPostReadRequest[testKibanaResourceModel]) (testKibanaResourceModel, diag.Diagnostics) {
		postReadCalled = true
		req.State.Name = types.StringValue("post-read-mutated")
		return req.State, nil
	}
	opts.Update = func(_ context.Context, _ *clients.KibanaScopedClient, req KibanaWriteRequest[testKibanaResourceModel]) (KibanaWriteResult[testKibanaResourceModel], diag.Diagnostics) {
		model := req.Plan
		model.Name = types.StringValue("from-write-callback")
		return KibanaWriteResult[testKibanaResourceModel]{
			Model:              model,
			SkipReadAfterWrite: true,
		}, nil
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "default/my-resource"), tftypes.NewValue(tftypes.String, "default"))
	prior := makeTestKibanaResourceState(ctx, t, "default/my-resource")
	resp := resource.UpdateResponse{State: prior}
	r.Update(ctx, resource.UpdateRequest{Plan: plan, State: prior, Config: kibanaTestConfig(plan)}, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.False(t, readCalled)
	require.False(t, postReadCalled, "PostRead must not run when SkipReadAfterWrite is true")
	var after testKibanaResourceModel
	require.False(t, resp.State.Get(ctx, &after).HasError())
	require.Equal(t, "from-write-callback", after.Name.ValueString())
}

func TestNewKibanaResource_Update_skipReadAfterWritePreservesPlanTimeouts(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	opts := defaultTestKibanaResourceOptions()
	opts.Update = func(_ context.Context, _ *clients.KibanaScopedClient, req KibanaWriteRequest[testKibanaResourceModel]) (KibanaWriteResult[testKibanaResourceModel], diag.Diagnostics) {
		model := req.Plan
		model.Name = types.StringValue("written")
		model.ResourceTimeoutsField = ResourceTimeoutsField{}
		return KibanaWriteResult[testKibanaResourceModel]{
			Model:              model,
			SkipReadAfterWrite: true,
		}, nil
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	objType := testKibanaResourceObjectType()
	planValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, "default/my-resource"),
		"name":              tftypes.NewValue(tftypes.String, "my-resource"),
		"space_id":          tftypes.NewValue(tftypes.String, "default"),
		"kibana_connection": tftypes.NewValue(kibanaConnectionBlockType(), nil),
		"timeouts":          resourceTimeoutsWithUpdate("30m"),
	})
	plan := tfsdk.Plan{
		Raw:    planValue,
		Schema: testKibanaResourceSchemaWithConnectionBlock(ctx),
	}
	prior := makeTestKibanaResourceState(ctx, t, "default/my-resource")
	resp := resource.UpdateResponse{State: prior}
	r.Update(ctx, resource.UpdateRequest{Plan: plan, State: prior, Config: kibanaTestConfig(plan)}, &resp)

	require.False(t, resp.Diagnostics.HasError())
	var planModel testKibanaResourceModel
	require.False(t, plan.Get(ctx, &planModel).HasError())
	wantTimeouts := planModel.GetTimeouts()

	var after testKibanaResourceModel
	require.False(t, resp.State.Get(ctx, &after).HasError())
	var gotTimeouts timeouts.Value
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root(attrTimeouts), &gotTimeouts)...)
	require.False(t, resp.Diagnostics.HasError())
	require.Equal(t, wantTimeouts, gotTimeouts, "plan timeouts must be preserved on SkipReadAfterWrite path")
}

func TestNewKibanaResource_Create_callbackReceivesNilPriorAndConfig(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	createFn := func(
		_ context.Context,
		_ *clients.KibanaScopedClient,
		req KibanaWriteRequest[testKibanaResourceModel],
	) (KibanaWriteResult[testKibanaResourceModel], diag.Diagnostics) {
		require.Nil(t, req.Prior)
		require.Equal(t, "my-resource", req.WriteID)
		require.Equal(t, "default", req.SpaceID)
		require.Equal(t, "my-resource", req.Plan.Name.ValueString())
		require.Equal(t, "my-resource", req.Config.Name.ValueString())
		model := req.Plan
		model.ID = types.StringValue(req.SpaceID + "/" + req.WriteID)
		return KibanaWriteResult[testKibanaResourceModel]{Model: model}, nil
	}
	opts := defaultTestKibanaResourceOptions()
	opts.Create = createFn
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tftypes.NewValue(tftypes.String, "default"))
	objType := testKibanaResourceObjectType()
	respState := tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: testKibanaResourceSchemaWithConnectionBlock(ctx)}
	resp := resource.CreateResponse{State: respState}
	r.Create(ctx, resource.CreateRequest{Plan: plan, Config: kibanaTestConfig(plan)}, &resp)

	require.False(t, resp.Diagnostics.HasError())
}

func TestNewKibanaResource_Update_callbackReceivesNonNilPriorAndConfig(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	updateFn := func(
		_ context.Context,
		_ *clients.KibanaScopedClient,
		req KibanaWriteRequest[testKibanaResourceModel],
	) (KibanaWriteResult[testKibanaResourceModel], diag.Diagnostics) {
		require.NotNil(t, req.Prior)
		require.Equal(t, "my-resource", req.WriteID)
		require.Equal(t, "default", req.SpaceID)
		require.Equal(t, "default/my-resource", req.Prior.ID.ValueString())
		require.Equal(t, "my-resource", req.Config.Name.ValueString())
		model := req.Plan
		model.ID = types.StringValue(req.SpaceID + "/" + req.WriteID)
		return KibanaWriteResult[testKibanaResourceModel]{Model: model}, nil
	}
	opts := defaultTestKibanaResourceOptions()
	opts.Update = updateFn
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "default/my-resource"), tftypes.NewValue(tftypes.String, "default"))
	prior := makeTestKibanaResourceState(ctx, t, "default/my-resource")
	resp := resource.UpdateResponse{State: prior}
	r.Update(ctx, resource.UpdateRequest{Plan: plan, State: prior, Config: kibanaTestConfig(plan)}, &resp)

	require.False(t, resp.Diagnostics.HasError())
}

func TestNewKibanaResource_SingleWriteFuncServesCreateAndUpdate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	var sawCreate, sawUpdate bool
	shared := func(
		_ context.Context,
		_ *clients.KibanaScopedClient,
		req KibanaWriteRequest[testKibanaResourceModel],
	) (KibanaWriteResult[testKibanaResourceModel], diag.Diagnostics) {
		if req.Prior == nil {
			sawCreate = true
		} else {
			sawUpdate = true
		}
		model := req.Plan
		model.ID = types.StringValue(req.SpaceID + "/" + req.WriteID)
		return KibanaWriteResult[testKibanaResourceModel]{Model: model}, nil
	}
	opts := defaultTestKibanaResourceOptions()
	opts.Create = shared
	opts.Update = shared
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	createPlan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tftypes.NewValue(tftypes.String, "default"))
	objType := testKibanaResourceObjectType()
	createResp := resource.CreateResponse{State: tfsdk.State{
		Raw:    tftypes.NewValue(objType, nil),
		Schema: testKibanaResourceSchemaWithConnectionBlock(ctx),
	}}
	r.Create(ctx, resource.CreateRequest{Plan: createPlan, Config: kibanaTestConfig(createPlan)}, &createResp)
	require.False(t, createResp.Diagnostics.HasError())

	updatePlan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "default/my-resource"), tftypes.NewValue(tftypes.String, "default"))
	priorState := makeTestKibanaResourceState(ctx, t, "default/my-resource")
	updateResp := resource.UpdateResponse{State: priorState}
	r.Update(ctx, resource.UpdateRequest{Plan: updatePlan, State: priorState, Config: kibanaTestConfig(updatePlan)}, &updateResp)
	require.False(t, updateResp.Diagnostics.HasError())

	require.True(t, sawCreate)
	require.True(t, sawUpdate)
}

func TestNewKibanaResource_Read_invokesPostReadBeforeStateSet(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	postCalled := false
	postRead := func(_ context.Context, req KibanaPostReadRequest[testKibanaResourceModel]) (testKibanaResourceModel, diag.Diagnostics) {
		postCalled = true
		return req.State, nil
	}
	opts := defaultTestKibanaResourceOptions()
	opts.PostRead = postRead
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	state := makeTestKibanaResourceState(ctx, t, "default/my-stream")
	req := resource.ReadRequest{State: state}
	resp := resource.ReadResponse{State: state}
	r.Read(ctx, req, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.True(t, postCalled)
}

func TestNewKibanaResource_Read_skipsPostReadWhenNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	postCalled := false
	postRead := func(_ context.Context, req KibanaPostReadRequest[testKibanaResourceModel]) (testKibanaResourceModel, diag.Diagnostics) {
		postCalled = true
		return req.State, nil
	}
	opts := defaultTestKibanaResourceOptions()
	opts.Read = func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModel) (testKibanaResourceModel, bool, diag.Diagnostics) {
		return testKibanaResourceModel{}, false, nil
	}
	opts.PostRead = postRead
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	state := makeTestKibanaResourceState(ctx, t, "default/my-stream")
	req := resource.ReadRequest{State: state}
	resp := resource.ReadResponse{State: state}
	r.Read(ctx, req, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.False(t, postCalled)
}

func TestNewKibanaResource_Read_skipsPostReadWhenReadFuncError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	postCalled := false
	postRead := func(_ context.Context, req KibanaPostReadRequest[testKibanaResourceModel]) (testKibanaResourceModel, diag.Diagnostics) {
		postCalled = true
		return req.State, nil
	}
	readFn := func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModel) (testKibanaResourceModel, bool, diag.Diagnostics) {
		var d diag.Diagnostics
		d.AddError("read failed", "boom")
		return testKibanaResourceModel{}, false, d
	}
	opts := defaultTestKibanaResourceOptions()
	opts.Read = readFn
	opts.PostRead = postRead
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	state := makeTestKibanaResourceState(ctx, t, "default/my-stream")
	req := resource.ReadRequest{State: state}
	resp := resource.ReadResponse{State: state}
	r.Read(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.False(t, postCalled)
}

func TestNewKibanaResource_Read_postReadRunsBeforeStateSet(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	postCalled := false
	postRead := func(_ context.Context, req KibanaPostReadRequest[testKibanaResourceModel]) (testKibanaResourceModel, diag.Diagnostics) {
		postCalled = true
		return req.State, nil
	}
	opts := defaultTestKibanaResourceOptions()
	opts.PostRead = postRead
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	goodState := makeTestKibanaResourceState(ctx, t, "default/my-stream")
	req := resource.ReadRequest{State: goodState}
	badOut := tfsdk.State{
		Raw:    goodState.Raw,
		Schema: getTestKibanaResourceSchema(context.Background()),
	}
	resp := resource.ReadResponse{State: badOut}
	r.Read(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.True(t, postCalled, "PostRead should run before state set and be called even when state set fails")
}

func TestNewKibanaResource_Read_postReadReceivesFrameworkPrivateHandle(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	var captured any
	postRead := func(_ context.Context, req KibanaPostReadRequest[testKibanaResourceModel]) (testKibanaResourceModel, diag.Diagnostics) {
		captured = req.Private
		return req.State, nil
	}
	opts := defaultTestKibanaResourceOptions()
	opts.PostRead = postRead
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	state := makeTestKibanaResourceState(ctx, t, "default/my-stream")
	req := resource.ReadRequest{State: state}
	resp := resource.ReadResponse{State: state}
	r.Read(ctx, req, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.Equal(t, resp.Private, captured)
}

func TestNewKibanaResource_Create_invokesPostReadAfterReadAfterWrite(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	postCalls := 0
	postRead := func(_ context.Context, req KibanaPostReadRequest[testKibanaResourceModel]) (testKibanaResourceModel, diag.Diagnostics) {
		postCalls++
		return req.State, nil
	}
	opts := defaultTestKibanaResourceOptions()
	opts.Read = testKibanaReadFuncDistinguishing
	opts.PostRead = postRead
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tftypes.NewValue(tftypes.String, "default"))
	objType := testKibanaResourceObjectType()
	respState := tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: testKibanaResourceSchemaWithConnectionBlock(ctx)}
	resp := resource.CreateResponse{State: respState}
	r.Create(ctx, resource.CreateRequest{Plan: plan, Config: kibanaTestConfig(plan)}, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.Equal(t, 1, postCalls)
}

func TestNewKibanaResource_Update_invokesPostReadAfterReadAfterWrite(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	postCalls := 0
	postRead := func(_ context.Context, req KibanaPostReadRequest[testKibanaResourceModel]) (testKibanaResourceModel, diag.Diagnostics) {
		postCalls++
		return req.State, nil
	}
	opts := defaultTestKibanaResourceOptions()
	opts.Read = testKibanaReadFuncDistinguishing
	opts.PostRead = postRead
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "default/my-resource"), tftypes.NewValue(tftypes.String, "default"))
	prior := makeTestKibanaResourceState(ctx, t, "default/my-resource")
	resp := resource.UpdateResponse{State: prior}
	r.Update(ctx, resource.UpdateRequest{Plan: plan, State: prior, Config: kibanaTestConfig(plan)}, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.Equal(t, 1, postCalls)
}

// =============================================================================
// Version requirements
// =============================================================================

// testKibanaResourceModelWithVersionReqs implements WithVersionRequirements
// and always returns error diagnostics from GetVersionRequirements.
type testKibanaResourceModelWithVersionReqs struct {
	ResourceTimeoutsField
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
func (*testKibanaResourceModelWithVersionReqs) GetVersionRequirements(_ context.Context) ([]VersionRequirement, diag.Diagnostics) {
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

func versionReqTestKibanaResourceOptions(createCalled, updateCalled *bool) KibanaResourceOptions[testKibanaResourceModelWithVersionReqs] {
	return KibanaResourceOptions[testKibanaResourceModelWithVersionReqs]{
		Schema: func(ctx context.Context) rschema.Schema {
			s := getTestKibanaResourceSchema(ctx)
			s.Blocks = map[string]rschema.Block{
				"kibana_connection": providerschema.GetKbFWConnectionBlock(),
			}
			return s
		},
		Read: func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModelWithVersionReqs) (testKibanaResourceModelWithVersionReqs, bool, diag.Diagnostics) {
			return testKibanaResourceModelWithVersionReqs{}, false, nil
		},
		Delete: func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModelWithVersionReqs) diag.Diagnostics {
			return nil
		},
		Create: func(
			_ context.Context,
			_ *clients.KibanaScopedClient,
			_ KibanaWriteRequest[testKibanaResourceModelWithVersionReqs],
		) (KibanaWriteResult[testKibanaResourceModelWithVersionReqs], diag.Diagnostics) {
			if createCalled != nil {
				*createCalled = true
			}
			return KibanaWriteResult[testKibanaResourceModelWithVersionReqs]{}, nil
		},
		Update: func(
			_ context.Context,
			_ *clients.KibanaScopedClient,
			_ KibanaWriteRequest[testKibanaResourceModelWithVersionReqs],
		) (KibanaWriteResult[testKibanaResourceModelWithVersionReqs], diag.Diagnostics) {
			if updateCalled != nil {
				*updateCalled = true
			}
			return KibanaWriteResult[testKibanaResourceModelWithVersionReqs]{}, nil
		},
	}
}

func TestKibanaResource_Create_versionReqDiagsStopCreate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	createCalled := false
	r := NewKibanaResource[testKibanaResourceModelWithVersionReqs](ComponentKibana, "test_entity", versionReqTestKibanaResourceOptions(&createCalled, nil))
	r.client = factory

	objType := testKibanaResourceObjectType()
	plan := tfsdk.Plan{
		Raw: tftypes.NewValue(objType, map[string]tftypes.Value{
			"id":                tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			"name":              tftypes.NewValue(tftypes.String, "my-resource"),
			"space_id":          tftypes.NewValue(tftypes.String, "default"),
			"kibana_connection": tftypes.NewValue(kibanaConnectionBlockType(), nil),
			"timeouts":          resourceTimeoutsNullValue(),
		}),
		Schema: testKibanaResourceSchemaWithConnectionBlock(ctx),
	}
	req := resource.CreateRequest{Plan: plan, Config: kibanaTestConfig(plan)}
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
	opts := versionReqTestKibanaResourceOptions(nil, nil)
	opts.Read = func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModelWithVersionReqs) (testKibanaResourceModelWithVersionReqs, bool, diag.Diagnostics) {
		readCalled = true
		return testKibanaResourceModelWithVersionReqs{}, false, nil
	}
	r := NewKibanaResource[testKibanaResourceModelWithVersionReqs](ComponentKibana, "test_entity", opts)
	r.client = factory

	state := tfsdk.State{
		Raw: tftypes.NewValue(testKibanaResourceObjectType(), map[string]tftypes.Value{
			"id":                tftypes.NewValue(tftypes.String, "default/my-resource"),
			"name":              tftypes.NewValue(tftypes.String, "my-resource"),
			"space_id":          tftypes.NewValue(tftypes.String, "default"),
			"kibana_connection": tftypes.NewValue(kibanaConnectionBlockType(), nil),
			"timeouts":          resourceTimeoutsNullValue(),
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

func TestKibanaResource_Delete_versionReqDiagsStopDelete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	deleteCalled := false
	opts := versionReqTestKibanaResourceOptions(nil, nil)
	opts.Delete = func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModelWithVersionReqs) diag.Diagnostics {
		deleteCalled = true
		return nil
	}
	r := NewKibanaResource[testKibanaResourceModelWithVersionReqs](ComponentKibana, "test_entity", opts)
	r.client = factory

	state := tfsdk.State{
		Raw: tftypes.NewValue(testKibanaResourceObjectType(), map[string]tftypes.Value{
			"id":                tftypes.NewValue(tftypes.String, "default/my-resource"),
			"name":              tftypes.NewValue(tftypes.String, "my-resource"),
			"space_id":          tftypes.NewValue(tftypes.String, "default"),
			"kibana_connection": tftypes.NewValue(kibanaConnectionBlockType(), nil),
			"timeouts":          resourceTimeoutsNullValue(),
		}),
		Schema: testKibanaResourceSchemaWithConnectionBlock(ctx),
	}
	req := resource.DeleteRequest{State: state}
	resp := resource.DeleteResponse{State: state}

	r.Delete(ctx, req, &resp)

	require.False(t, deleteCalled, "deleteFunc must NOT be called when GetVersionRequirements returns error diags")
	require.True(t, resp.Diagnostics.HasError(), "Delete must propagate error from GetVersionRequirements")
	summaries := make([]string, 0, len(resp.Diagnostics))
	for _, d := range resp.Diagnostics {
		summaries = append(summaries, d.Summary())
	}
	require.Contains(t, summaries, "version requirements error",
		"diagnostic from GetVersionRequirements must be appended; got: %v", summaries)
}

type testKibanaResourceModelMinVersion955 struct {
	ResourceTimeoutsField
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	SpaceID          types.String `tfsdk:"space_id"`
	KibanaConnection types.List   `tfsdk:"kibana_connection"`
}

func (m testKibanaResourceModelMinVersion955) GetID() types.String         { return m.ID }
func (m testKibanaResourceModelMinVersion955) GetResourceID() types.String { return m.Name }
func (m testKibanaResourceModelMinVersion955) GetSpaceID() types.String    { return m.SpaceID }
func (m testKibanaResourceModelMinVersion955) GetKibanaConnection() types.List {
	return m.KibanaConnection
}

func (*testKibanaResourceModelMinVersion955) GetVersionRequirements(_ context.Context) ([]VersionRequirement, diag.Diagnostics) {
	return []VersionRequirement{
		{
			MinVersion:   *version.Must(version.NewVersion("9.5.0")),
			ErrorMessage: "requires Kibana 9.5.0 or later",
		},
	}, nil
}

func TestKibanaResource_Delete_unsupportedServerStopsBeforeDeleteFunc(t *testing.T) {
	ctx := context.Background()

	srv := newMockKibanaStatusServer("7.17.0")
	defer srv.Close()

	deleteCalled := false
	r := NewKibanaResource[testKibanaResourceModelMinVersion955](ComponentKibana, "test_entity", KibanaResourceOptions[testKibanaResourceModelMinVersion955]{
		Schema: func(ctx context.Context) rschema.Schema {
			s := getTestKibanaResourceSchema(ctx)
			s.Blocks = map[string]rschema.Block{
				"kibana_connection": providerschema.GetKbFWConnectionBlock(),
			}
			return s
		},
		Read: func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModelMinVersion955) (testKibanaResourceModelMinVersion955, bool, diag.Diagnostics) {
			return testKibanaResourceModelMinVersion955{}, false, nil
		},
		Delete: func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModelMinVersion955) diag.Diagnostics {
			deleteCalled = true
			return nil
		},
		Create: PlaceholderKibanaWriteCallback[testKibanaResourceModelMinVersion955](),
		Update: PlaceholderKibanaWriteCallback[testKibanaResourceModelMinVersion955](),
	})
	factory := newKibanaFactoryForURL(t, srv.URL)
	r.client = factory

	state := tfsdk.State{
		Raw: tftypes.NewValue(testKibanaResourceObjectType(), map[string]tftypes.Value{
			"id":                tftypes.NewValue(tftypes.String, "default/my-resource"),
			"name":              tftypes.NewValue(tftypes.String, "my-resource"),
			"space_id":          tftypes.NewValue(tftypes.String, "default"),
			"kibana_connection": tftypes.NewValue(kibanaConnectionBlockType(), nil),
			"timeouts":          resourceTimeoutsNullValue(),
		}),
		Schema: testKibanaResourceSchemaWithConnectionBlock(ctx),
	}
	req := resource.DeleteRequest{State: state}
	resp := resource.DeleteResponse{State: state}

	r.Delete(ctx, req, &resp)

	require.False(t, deleteCalled, "deleteFunc must NOT be called when server is below minimum version")
	require.True(t, resp.Diagnostics.HasError(), "Delete must produce an error diagnostic for unsupported server")

	var foundUnsupported bool
	for _, e := range resp.Diagnostics.Errors() {
		if e.Summary() == "Unsupported server version" {
			foundUnsupported = true
			require.Contains(t, e.Detail(), "requires Kibana 9.5.0 or later")
		}
	}
	require.True(t, foundUnsupported,
		"must have an 'Unsupported server version' diagnostic; got: %v", resp.Diagnostics.Errors())
}

func TestNewKibanaResource_Read_postReadReceivesPriorStateAsPrior(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	var capturedPrior string
	postRead := func(_ context.Context, req KibanaPostReadRequest[testKibanaResourceModel]) (testKibanaResourceModel, diag.Diagnostics) {
		capturedPrior = req.Prior.Name.ValueString()
		return req.State, nil
	}
	opts := defaultTestKibanaResourceOptions()
	opts.PostRead = postRead
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	state := makeTestKibanaResourceState(ctx, t, "default/my-stream")
	req := resource.ReadRequest{State: state}
	resp := resource.ReadResponse{State: state}
	r.Read(ctx, req, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.Equal(t, "my-stream", capturedPrior, "Prior should be the prior state name on Read path")
}

func TestNewKibanaResource_Create_postReadPriorIsPlan(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	var capturedPrior string
	postRead := func(_ context.Context, req KibanaPostReadRequest[testKibanaResourceModel]) (testKibanaResourceModel, diag.Diagnostics) {
		capturedPrior = req.Prior.Name.ValueString()
		return req.State, nil
	}
	opts := defaultTestKibanaResourceOptions()
	opts.Read = testKibanaReadFuncDistinguishing
	opts.PostRead = postRead
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tftypes.NewValue(tftypes.String, "default"))
	objType := testKibanaResourceObjectType()
	respState := tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: testKibanaResourceSchemaWithConnectionBlock(ctx)}
	resp := resource.CreateResponse{State: respState}
	r.Create(ctx, resource.CreateRequest{Plan: plan, Config: kibanaTestConfig(plan)}, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.Equal(t, "my-resource", capturedPrior, "Prior should be the plan name on Create (write) path")
}

func TestNewKibanaResource_Create_postReadStateSetFromReturnedModel(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	postRead := func(_ context.Context, req KibanaPostReadRequest[testKibanaResourceModel]) (testKibanaResourceModel, diag.Diagnostics) {
		returnModel := req.State
		returnModel.Name = types.StringValue("postread-modified")
		return returnModel, nil
	}
	opts := defaultTestKibanaResourceOptions()
	opts.Read = testKibanaReadFuncDistinguishing
	opts.PostRead = postRead
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tftypes.NewValue(tftypes.String, "default"))
	objType := testKibanaResourceObjectType()
	respState := tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: testKibanaResourceSchemaWithConnectionBlock(ctx)}
	resp := resource.CreateResponse{State: respState}
	r.Create(ctx, resource.CreateRequest{Plan: plan, Config: kibanaTestConfig(plan)}, &resp)

	require.False(t, resp.Diagnostics.HasError())
	var result testKibanaResourceModel
	diags := resp.State.Get(ctx, &result)
	require.False(t, diags.HasError())
	require.Equal(t, "postread-modified", result.Name.ValueString(), "State should be set from PostRead's returned model")
}

func TestNewKibanaResource_Create_skipsStateSetWhenPostReadError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	postRead := func(_ context.Context, req KibanaPostReadRequest[testKibanaResourceModel]) (testKibanaResourceModel, diag.Diagnostics) {
		var d diag.Diagnostics
		d.AddError("postread error", "intentional")
		return req.State, d
	}
	opts := defaultTestKibanaResourceOptions()
	opts.Read = testKibanaReadFuncDistinguishing
	opts.PostRead = postRead
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tftypes.NewValue(tftypes.String, "default"))
	objType := testKibanaResourceObjectType()
	respState := tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: testKibanaResourceSchemaWithConnectionBlock(ctx)}
	resp := resource.CreateResponse{State: respState}
	r.Create(ctx, resource.CreateRequest{Plan: plan, Config: kibanaTestConfig(plan)}, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.True(t, resp.State.Raw.IsNull(), "State should remain unset when PostRead returns error")
}

func TestNewKibanaResource_Read_postReadStateSetFromReturnedModel(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	postRead := func(_ context.Context, req KibanaPostReadRequest[testKibanaResourceModel]) (testKibanaResourceModel, diag.Diagnostics) {
		returnModel := req.State
		returnModel.Name = types.StringValue("read-modified")
		return returnModel, nil
	}
	opts := defaultTestKibanaResourceOptions()
	opts.PostRead = postRead
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	state := makeTestKibanaResourceState(ctx, t, "default/my-stream")
	req := resource.ReadRequest{State: state}
	resp := resource.ReadResponse{State: state}
	r.Read(ctx, req, &resp)

	require.False(t, resp.Diagnostics.HasError())
	var result testKibanaResourceModel
	diags := resp.State.Get(ctx, &result)
	require.False(t, diags.HasError())
	require.Equal(t, "read-modified", result.Name.ValueString(), "State should be set from PostRead's returned model")
}

func TestKibanaResource_Update_versionReqDiagsStopUpdate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	updateCalled := false
	r := NewKibanaResource[testKibanaResourceModelWithVersionReqs](ComponentKibana, "test_entity", versionReqTestKibanaResourceOptions(nil, &updateCalled))
	r.client = factory

	objType := testKibanaResourceObjectType()
	plan := tfsdk.Plan{
		Raw: tftypes.NewValue(objType, map[string]tftypes.Value{
			"id":                tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			"name":              tftypes.NewValue(tftypes.String, "my-resource"),
			"space_id":          tftypes.NewValue(tftypes.String, "default"),
			"kibana_connection": tftypes.NewValue(kibanaConnectionBlockType(), nil),
			"timeouts":          resourceTimeoutsNullValue(),
		}),
		Schema: testKibanaResourceSchemaWithConnectionBlock(ctx),
	}
	state := tfsdk.State{
		Raw: tftypes.NewValue(objType, map[string]tftypes.Value{
			"id":                tftypes.NewValue(tftypes.String, "default/my-resource"),
			"name":              tftypes.NewValue(tftypes.String, "my-resource"),
			"space_id":          tftypes.NewValue(tftypes.String, "default"),
			"kibana_connection": tftypes.NewValue(kibanaConnectionBlockType(), nil),
			"timeouts":          resourceTimeoutsNullValue(),
		}),
		Schema: testKibanaResourceSchemaWithConnectionBlock(ctx),
	}
	req := resource.UpdateRequest{Plan: plan, State: state, Config: kibanaTestConfig(plan)}
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
