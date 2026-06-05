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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
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

func testKibanaResourceObjectTypeWithTimeouts() tftypes.Type {
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

func TestKibanaResource_Schema_injectsTimeoutsAttribute(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", defaultTestKibanaResourceOptions())
	var resp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &resp)
	require.Contains(t, resp.Schema.Attributes, attrTimeouts)
	nested, ok := resp.Schema.Attributes[attrTimeouts].(rschema.SingleNestedAttribute)
	require.True(t, ok)
	require.Contains(t, nested.Attributes, "create")
	require.Contains(t, nested.Attributes, "read")
	require.Contains(t, nested.Attributes, "update")
	require.Contains(t, nested.Attributes, "delete")
}

func TestKibanaResource_Create_validateSpaceIDBeforeClientUnderTimeout(t *testing.T) {
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

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tftypes.NewValue(tftypes.String, ""))

	var resp resource.CreateResponse
	r.Create(ctx, resource.CreateRequest{Plan: plan, Config: tfsdk.Config(plan)}, &resp)
	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Invalid space identifier")
	require.False(t, createCalled)
}

type kbTimeoutsVersionModel struct {
	ResourceTimeoutsField
	KibanaConnectionField
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	SpaceID types.String `tfsdk:"space_id"`
}

func (m kbTimeoutsVersionModel) GetID() types.String         { return m.ID }
func (m kbTimeoutsVersionModel) GetResourceID() types.String { return m.Name }
func (m kbTimeoutsVersionModel) GetSpaceID() types.String    { return m.SpaceID }

func (kbTimeoutsVersionModel) GetVersionRequirements(_ context.Context) ([]VersionRequirement, diag.Diagnostics) {
	return []VersionRequirement{{
		MinVersion:   *version.Must(version.NewVersion("8.0.0")),
		ErrorMessage: "requires Kibana 8.0.0+",
	}}, nil
}

func TestKibanaResource_Read_versionCheckTimesOutBeforeCallback(t *testing.T) {
	ctx := context.Background()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/status" {
			time.Sleep(3 * time.Second)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"version":{"number":"8.19.0","build_flavor":"default"}}`)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	readCalled := false
	r := NewKibanaResource[kbTimeoutsVersionModel](ComponentKibana, "test_entity", KibanaResourceOptions[kbTimeoutsVersionModel]{
		Schema: func(_ context.Context) rschema.Schema {
			return rschema.Schema{
				Attributes: map[string]rschema.Attribute{
					"id":       rschema.StringAttribute{Computed: true},
					"name":     rschema.StringAttribute{Optional: true},
					"space_id": rschema.StringAttribute{Optional: true},
				},
			}
		},
		Read: func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ kbTimeoutsVersionModel) (kbTimeoutsVersionModel, bool, diag.Diagnostics) {
			readCalled = true
			return kbTimeoutsVersionModel{}, false, nil
		},
		Delete: func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ kbTimeoutsVersionModel) diag.Diagnostics {
			return nil
		},
		Create: func(_ context.Context, _ *clients.KibanaScopedClient, _ KibanaWriteRequest[kbTimeoutsVersionModel]) (KibanaWriteResult[kbTimeoutsVersionModel], diag.Diagnostics) {
			return KibanaWriteResult[kbTimeoutsVersionModel]{}, nil
		},
		Update: func(_ context.Context, _ *clients.KibanaScopedClient, _ KibanaWriteRequest[kbTimeoutsVersionModel]) (KibanaWriteResult[kbTimeoutsVersionModel], diag.Diagnostics) {
			return KibanaWriteResult[kbTimeoutsVersionModel]{}, nil
		},
		Timeouts: ResourceTimeouts{Read: 200 * time.Millisecond},
	})
	factory := newKibanaFactoryForURL(t, srv.URL)
	r.client = factory

	objType := testKibanaResourceObjectTypeWithTimeouts()
	stateValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, "user1"),
		"name":              tftypes.NewValue(tftypes.String, "user1"),
		"space_id":          tftypes.NewValue(tftypes.String, "default"),
		"kibana_connection": tftypes.NewValue(kibanaConnectionBlockType(), nil),
		"timeouts":          resourceTimeoutsNullValue(),
	})
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	state := tfsdk.State{Raw: stateValue, Schema: schemaResp.Schema}
	var resp resource.ReadResponse
	resp.State = state
	r.Read(ctx, resource.ReadRequest{State: state}, &resp)
	require.True(t, resp.Diagnostics.HasError())
	require.False(t, readCalled)
}

func TestKibanaResource_Schema_silentlyOverwritesFactoryTimeouts(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	sentinel := rschema.StringAttribute{Description: "sentinel"}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", KibanaResourceOptions[testKibanaResourceModel]{
		Schema: func(_ context.Context) rschema.Schema {
			return rschema.Schema{
				Attributes: map[string]rschema.Attribute{
					"id":         rschema.StringAttribute{Computed: true},
					"name":       rschema.StringAttribute{Optional: true},
					"space_id":   rschema.StringAttribute{Optional: true},
					attrTimeouts: sentinel,
				},
			}
		},
		Read:   testKibanaReadFuncFound,
		Delete: testKibanaDeleteFunc,
		Create: testKibanaWriteFuncFound,
		Update: testKibanaWriteFuncFound,
	})
	var resp resource.SchemaResponse
	require.NotPanics(t, func() {
		r.Schema(ctx, resource.SchemaRequest{}, &resp)
	})
	require.Equal(t, fmt.Sprintf("%T", timeouts.AttributesAll(ctx)), fmt.Sprintf("%T", resp.Schema.Attributes[attrTimeouts]))
}

func TestKibanaResource_Create_appliesOptionsTimeout(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	var receivedDeadline time.Time
	opts := defaultTestKibanaResourceOptions()
	opts.Create = func(ctx context.Context, _ *clients.KibanaScopedClient, req KibanaWriteRequest[testKibanaResourceModel]) (KibanaWriteResult[testKibanaResourceModel], diag.Diagnostics) {
		deadline, ok := ctx.Deadline()
		require.True(t, ok)
		receivedDeadline = deadline
		m := req.Plan
		m.ID = types.StringValue(req.SpaceID + "/" + req.WriteID)
		return KibanaWriteResult[testKibanaResourceModel]{Model: m}, nil
	}
	opts.Timeouts = ResourceTimeouts{Create: 7 * time.Second}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tftypes.NewValue(tftypes.String, "default"))
	before := time.Now()
	var resp resource.CreateResponse
	resp.State = tfsdk.State{Raw: tftypes.NewValue(testKibanaResourceObjectType(), nil), Schema: plan.Schema}
	r.Create(ctx, resource.CreateRequest{Plan: plan, Config: tfsdk.Config(plan)}, &resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.WithinDuration(t, before.Add(7*time.Second), receivedDeadline, 2*time.Second)
}

func TestKibanaResource_Create_planTimeoutOverridesOptions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	var receivedDeadline time.Time
	opts := defaultTestKibanaResourceOptions()
	opts.Create = func(ctx context.Context, _ *clients.KibanaScopedClient, req KibanaWriteRequest[testKibanaResourceModel]) (KibanaWriteResult[testKibanaResourceModel], diag.Diagnostics) {
		deadline, ok := ctx.Deadline()
		require.True(t, ok)
		receivedDeadline = deadline
		m := req.Plan
		m.ID = types.StringValue(req.SpaceID + "/" + req.WriteID)
		return KibanaWriteResult[testKibanaResourceModel]{Model: m}, nil
	}
	opts.Timeouts = ResourceTimeouts{Create: 7 * time.Second}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	objType := testKibanaResourceObjectTypeWithTimeouts()
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"name":              tftypes.NewValue(tftypes.String, "my-resource"),
		"space_id":          tftypes.NewValue(tftypes.String, "default"),
		"kibana_connection": tftypes.NewValue(kibanaConnectionBlockType(), nil),
		"timeouts":          resourceTimeoutsWithCreate("30m"),
	})
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	plan := tfsdk.Plan{Raw: objValue, Schema: schemaResp.Schema}
	before := time.Now()
	var resp resource.CreateResponse
	resp.State = tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: schemaResp.Schema}
	r.Create(ctx, resource.CreateRequest{Plan: plan, Config: tfsdk.Config(plan)}, &resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.WithinDuration(t, before.Add(30*time.Minute), receivedDeadline, 2*time.Second)
}

func TestKibanaResource_Create_fallsBackToPackageDefaultTimeout(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	var receivedDeadline time.Time
	opts := defaultTestKibanaResourceOptions()
	opts.Create = func(ctx context.Context, _ *clients.KibanaScopedClient, req KibanaWriteRequest[testKibanaResourceModel]) (KibanaWriteResult[testKibanaResourceModel], diag.Diagnostics) {
		deadline, ok := ctx.Deadline()
		require.True(t, ok)
		receivedDeadline = deadline
		m := req.Plan
		m.ID = types.StringValue(req.SpaceID + "/" + req.WriteID)
		return KibanaWriteResult[testKibanaResourceModel]{Model: m}, nil
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory
	plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tftypes.NewValue(tftypes.String, "default"))
	before := time.Now()
	var resp resource.CreateResponse
	resp.State = tfsdk.State{Raw: tftypes.NewValue(testKibanaResourceObjectType(), nil), Schema: plan.Schema}
	r.Create(ctx, resource.CreateRequest{Plan: plan, Config: tfsdk.Config(plan)}, &resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.WithinDuration(t, before.Add(DefaultResourceCreateTimeout), receivedDeadline, time.Second)
}

func TestKibanaResource_Read_nullStoredTimeoutsUsesDefault(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	var receivedDeadline time.Time
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", KibanaResourceOptions[testKibanaResourceModel]{
		Schema: getTestKibanaResourceSchema,
		Read: func(ctx context.Context, _ *clients.KibanaScopedClient, _ string, _ string, model testKibanaResourceModel) (testKibanaResourceModel, bool, diag.Diagnostics) {
			deadline, ok := ctx.Deadline()
			require.True(t, ok)
			receivedDeadline = deadline
			return model, true, nil
		},
		Delete: testKibanaDeleteFunc,
		Create: testKibanaWriteFuncFound,
		Update: testKibanaWriteFuncFound,
	})
	r.client = factory

	objType := testKibanaResourceObjectTypeWithTimeouts()
	stateValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, "default/my-resource"),
		"name":              tftypes.NewValue(tftypes.String, "my-resource"),
		"space_id":          tftypes.NewValue(tftypes.String, "default"),
		"kibana_connection": tftypes.NewValue(kibanaConnectionBlockType(), nil),
		"timeouts":          resourceTimeoutsNullValue(),
	})
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	state := tfsdk.State{Raw: stateValue, Schema: schemaResp.Schema}
	before := time.Now()
	var resp resource.ReadResponse
	resp.State = state
	r.Read(ctx, resource.ReadRequest{State: state}, &resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.WithinDuration(t, before.Add(DefaultResourceReadTimeout), receivedDeadline, time.Second)
}

func TestKibanaResource_operations_setContextDeadline(t *testing.T) {
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)

	type opCase struct {
		name string
		run  func(t *testing.T, r *KibanaResource[testKibanaResourceModel], schema rschema.Schema)
	}
	cases := []opCase{
		{
			name: "read",
			run: func(t *testing.T, r *KibanaResource[testKibanaResourceModel], schema rschema.Schema) {
				state := makeTestKibanaResourceState(ctx, t, "default/my-resource")
				state.Schema = schema
				var resp resource.ReadResponse
				resp.State = state
				r.Read(ctx, resource.ReadRequest{State: state}, &resp)
				require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			},
		},
		{
			name: "delete",
			run: func(t *testing.T, r *KibanaResource[testKibanaResourceModel], schema rschema.Schema) {
				state := makeTestKibanaResourceState(ctx, t, "default/my-resource")
				state.Schema = schema
				var resp resource.DeleteResponse
				resp.State = state
				r.Delete(ctx, resource.DeleteRequest{State: state}, &resp)
				require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			},
		},
		{
			name: "create",
			run: func(t *testing.T, r *KibanaResource[testKibanaResourceModel], schema rschema.Schema) {
				plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tftypes.NewValue(tftypes.String, "default"))
				plan.Schema = schema
				objType := testKibanaResourceObjectType()
				var resp resource.CreateResponse
				resp.State = tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: schema}
				r.Create(ctx, resource.CreateRequest{Plan: plan, Config: tfsdk.Config(plan)}, &resp)
				require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			},
		},
		{
			name: "update",
			run: func(t *testing.T, r *KibanaResource[testKibanaResourceModel], schema rschema.Schema) {
				plan := makeTestKibanaResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "default/my-resource"), tftypes.NewValue(tftypes.String, "default"))
				plan.Schema = schema
				state := makeTestKibanaResourceState(ctx, t, "default/my-resource")
				state.Schema = schema
				var resp resource.UpdateResponse
				resp.State = state
				r.Update(ctx, resource.UpdateRequest{Plan: plan, State: state, Config: tfsdk.Config(plan)}, &resp)
				require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var sawDeadline bool
			opts := defaultTestKibanaResourceOptions()
			opts.Read = func(ctx context.Context, _ *clients.KibanaScopedClient, _ string, _ string, model testKibanaResourceModel) (testKibanaResourceModel, bool, diag.Diagnostics) {
				_, sawDeadline = ctx.Deadline()
				return model, true, nil
			}
			opts.Delete = func(ctx context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ testKibanaResourceModel) diag.Diagnostics {
				_, sawDeadline = ctx.Deadline()
				return nil
			}
			opts.Create = func(ctx context.Context, _ *clients.KibanaScopedClient, req KibanaWriteRequest[testKibanaResourceModel]) (KibanaWriteResult[testKibanaResourceModel], diag.Diagnostics) {
				_, sawDeadline = ctx.Deadline()
				m := req.Plan
				m.ID = types.StringValue(req.SpaceID + "/" + req.WriteID)
				return KibanaWriteResult[testKibanaResourceModel]{Model: m}, nil
			}
			opts.Update = func(ctx context.Context, _ *clients.KibanaScopedClient, req KibanaWriteRequest[testKibanaResourceModel]) (KibanaWriteResult[testKibanaResourceModel], diag.Diagnostics) {
				_, sawDeadline = ctx.Deadline()
				m := req.Plan
				m.ID = types.StringValue(req.SpaceID + "/" + req.WriteID)
				return KibanaWriteResult[testKibanaResourceModel]{Model: m}, nil
			}
			r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
			r.client = factory
			var schemaResp resource.SchemaResponse
			r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
			tc.run(t, r, schemaResp.Schema)
			require.True(t, sawDeadline, "callback must observe context deadline for %s", tc.name)
		})
	}
}

func kibanaReadFuncDropsTimeouts(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, prior testKibanaResourceModel) (testKibanaResourceModel, bool, diag.Diagnostics) {
	return testKibanaResourceModel{
		ID:               prior.ID,
		Name:             prior.Name,
		SpaceID:          prior.SpaceID,
		KibanaConnection: prior.KibanaConnection,
	}, true, nil
}

func TestKibanaResource_Read_preservesTimeoutsWhenCallbackDropsField(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", KibanaResourceOptions[testKibanaResourceModel]{
		Schema: getTestKibanaResourceSchema,
		Read:   kibanaReadFuncDropsTimeouts,
		Delete: testKibanaDeleteFunc,
		Create: testKibanaWriteFuncFound,
		Update: testKibanaWriteFuncFound,
	})
	r.client = factory

	objType := testKibanaResourceObjectTypeWithTimeouts()
	stateValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, "default/my-resource"),
		"name":              tftypes.NewValue(tftypes.String, "my-resource"),
		"space_id":          tftypes.NewValue(tftypes.String, "default"),
		"kibana_connection": tftypes.NewValue(kibanaConnectionBlockType(), nil),
		"timeouts":          resourceTimeoutsWithCreate("10m"),
	})
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	state := tfsdk.State{Raw: stateValue, Schema: schemaResp.Schema}

	var priorModel testKibanaResourceModel
	require.False(t, state.Get(ctx, &priorModel).HasError())
	wantTimeouts := priorModel.GetTimeouts()

	var resp resource.ReadResponse
	resp.State = state
	r.Read(ctx, resource.ReadRequest{State: state}, &resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)

	var gotTimeouts timeouts.Value
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root(attrTimeouts), &gotTimeouts)...)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.Equal(t, wantTimeouts, gotTimeouts)
}

func TestKibanaResource_Create_preservesTimeoutsWhenCallbackDropsField(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	opts := defaultTestKibanaResourceOptions()
	opts.Create = func(_ context.Context, _ *clients.KibanaScopedClient, req KibanaWriteRequest[testKibanaResourceModel]) (KibanaWriteResult[testKibanaResourceModel], diag.Diagnostics) {
		return KibanaWriteResult[testKibanaResourceModel]{Model: testKibanaResourceModel{
			ID:               types.StringValue(req.SpaceID + "/" + req.WriteID),
			Name:             req.Plan.Name,
			SpaceID:          req.Plan.SpaceID,
			KibanaConnection: req.Plan.KibanaConnection,
		}}, nil
	}
	opts.Read = kibanaReadFuncDropsTimeouts
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	objType := testKibanaResourceObjectTypeWithTimeouts()
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"name":              tftypes.NewValue(tftypes.String, "my-resource"),
		"space_id":          tftypes.NewValue(tftypes.String, "default"),
		"kibana_connection": tftypes.NewValue(kibanaConnectionBlockType(), nil),
		"timeouts":          resourceTimeoutsWithCreate("10m"),
	})
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	plan := tfsdk.Plan{Raw: objValue, Schema: schemaResp.Schema}

	var planModel testKibanaResourceModel
	require.False(t, plan.Get(ctx, &planModel).HasError())
	wantTimeouts := planModel.GetTimeouts()

	var resp resource.CreateResponse
	resp.State = tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: schemaResp.Schema}
	r.Create(ctx, resource.CreateRequest{Plan: plan, Config: tfsdk.Config(plan)}, &resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)

	var gotTimeouts timeouts.Value
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root(attrTimeouts), &gotTimeouts)...)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.Equal(t, wantTimeouts, gotTimeouts)
}

func TestKibanaResource_Create_invalidTimeoutPreventsCallback(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	createCalled := false
	opts := defaultTestKibanaResourceOptions()
	opts.Create = func(_ context.Context, _ *clients.KibanaScopedClient, _ KibanaWriteRequest[testKibanaResourceModel]) (KibanaWriteResult[testKibanaResourceModel], diag.Diagnostics) {
		createCalled = true
		t.Fatal("create callback must not run when timeouts.create is invalid")
		return KibanaWriteResult[testKibanaResourceModel]{}, nil
	}
	r := NewKibanaResource[testKibanaResourceModel](ComponentKibana, "test_entity", opts)
	r.client = factory

	objType := testKibanaResourceObjectTypeWithTimeouts()
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"name":              tftypes.NewValue(tftypes.String, "my-resource"),
		"space_id":          tftypes.NewValue(tftypes.String, "default"),
		"kibana_connection": tftypes.NewValue(kibanaConnectionBlockType(), nil),
		"timeouts":          resourceTimeoutsWithCreate("not-a-duration"),
	})
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	plan := tfsdk.Plan{Raw: objValue, Schema: schemaResp.Schema}
	var resp resource.CreateResponse
	r.Create(ctx, resource.CreateRequest{Plan: plan, Config: tfsdk.Config(plan)}, &resp)
	require.True(t, resp.Diagnostics.HasError())
	require.False(t, createCalled)
}
