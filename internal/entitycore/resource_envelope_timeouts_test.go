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

func resourceTimeoutsObjectType() tftypes.Type {
	return tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"create": tftypes.String,
			"read":   tftypes.String,
			"update": tftypes.String,
			"delete": tftypes.String,
		},
	}
}

func resourceTimeoutsNullValue() tftypes.Value {
	return tftypes.NewValue(resourceTimeoutsObjectType(), nil)
}

func resourceTimeoutsWithCreate(create string) tftypes.Value {
	return tftypes.NewValue(resourceTimeoutsObjectType(), map[string]tftypes.Value{
		"create": tftypes.NewValue(tftypes.String, create),
		"read":   tftypes.NewValue(tftypes.String, nil),
		"update": tftypes.NewValue(tftypes.String, nil),
		"delete": tftypes.NewValue(tftypes.String, nil),
	})
}

func resourceTimeoutsWithUpdate(update string) tftypes.Value {
	return tftypes.NewValue(resourceTimeoutsObjectType(), map[string]tftypes.Value{
		"create": tftypes.NewValue(tftypes.String, nil),
		"read":   tftypes.NewValue(tftypes.String, nil),
		"update": tftypes.NewValue(tftypes.String, update),
		"delete": tftypes.NewValue(tftypes.String, nil),
	})
}

func testResourceObjectTypeWithTimeouts() tftypes.Type {
	return tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":                       tftypes.String,
			"name":                     tftypes.String,
			"elasticsearch_connection": elasticsearchConnectionBlockType(),
			"timeouts":                 resourceTimeoutsObjectType(),
		},
	}
}

func TestElasticsearchResource_Schema_injectsTimeoutsAttribute(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := NewElasticsearchResource[testResourceModel]("test_entity", defaultTestElasticsearchResourceOptions())
	var resp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &resp)
	require.Contains(t, resp.Schema.Attributes, attrTimeouts)
	nested, ok := resp.Schema.Attributes[attrTimeouts].(rschema.SingleNestedAttribute)
	require.True(t, ok, "timeouts must be a single nested attribute")
	require.Contains(t, nested.Attributes, "create")
	require.Contains(t, nested.Attributes, "read")
	require.Contains(t, nested.Attributes, "update")
	require.Contains(t, nested.Attributes, "delete")
}

func TestElasticsearchResource_Schema_silentlyOverwritesFactoryTimeouts(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	sentinel := rschema.StringAttribute{Description: "sentinel"}
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema: func(_ context.Context) rschema.Schema {
			return rschema.Schema{
				Attributes: map[string]rschema.Attribute{
					"id":         rschema.StringAttribute{Computed: true},
					"name":       rschema.StringAttribute{Optional: true},
					attrTimeouts: sentinel,
				},
			}
		},
		Read:   testReadFuncFound,
		Delete: testDeleteFunc,
		Create: testWriteFuncFoundCreate,
		Update: testWriteFuncFoundUpdate,
	})
	var resp resource.SchemaResponse
	require.NotPanics(t, func() {
		r.Schema(ctx, resource.SchemaRequest{}, &resp)
	})
	require.False(t, resp.Diagnostics.HasError())
	require.Equal(t, fmt.Sprintf("%T", timeouts.AttributesAll(ctx)), fmt.Sprintf("%T", resp.Schema.Attributes[attrTimeouts]))
}

func TestElasticsearchResource_Create_appliesOptionsTimeout(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	var receivedDeadline time.Time
	createFn := func(ctx context.Context, _ *clients.ElasticsearchScopedClient, req WriteRequest[testResourceModel]) (WriteResult[testResourceModel], diag.Diagnostics) {
		deadline, ok := ctx.Deadline()
		require.True(t, ok)
		receivedDeadline = deadline
		m := req.Plan
		m.ID = types.StringValue("cluster/" + req.WriteID)
		return WriteResult[testResourceModel]{Model: m}, nil
	}
	opts := defaultTestElasticsearchResourceOptions()
	opts.Create = createFn
	opts.Read = testReadFuncFound
	opts.Timeouts = ResourceTimeouts{Create: 7 * time.Second}
	r := NewElasticsearchResource[testResourceModel]("test_entity", opts)
	r.client = factory

	plan := makeTestResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue))
	resp := resource.CreateResponse{State: tfsdk.State{Raw: tftypes.NewValue(testResourceObjectType(), nil), Schema: testResourceSchemaWithConnectionBlock(ctx)}}
	before := time.Now()
	r.Create(ctx, resource.CreateRequest{Plan: plan, Config: tfsdk.Config(plan)}, &resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.WithinDuration(t, before.Add(7*time.Second), receivedDeadline, 2*time.Second)
}

func TestElasticsearchResource_Create_planTimeoutOverridesOptions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	var receivedDeadline time.Time
	opts := defaultTestElasticsearchResourceOptions()
	opts.Create = func(ctx context.Context, _ *clients.ElasticsearchScopedClient, req WriteRequest[testResourceModel]) (WriteResult[testResourceModel], diag.Diagnostics) {
		deadline, ok := ctx.Deadline()
		require.True(t, ok)
		receivedDeadline = deadline
		m := req.Plan
		m.ID = types.StringValue("cluster/" + req.WriteID)
		return WriteResult[testResourceModel]{Model: m}, nil
	}
	opts.Timeouts = ResourceTimeouts{Create: 7 * time.Second}
	r := NewElasticsearchResource[testResourceModel]("test_entity", opts)
	r.client = factory

	objType := testResourceObjectTypeWithTimeouts()
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                       tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"name":                     tftypes.NewValue(tftypes.String, "user1"),
		"elasticsearch_connection": tftypes.NewValue(elasticsearchConnectionBlockType(), nil),
		"timeouts":                 resourceTimeoutsWithCreate("30m"),
	})
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	plan := tfsdk.Plan{Raw: objValue, Schema: schemaResp.Schema}
	before := time.Now()
	resp := resource.CreateResponse{State: tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: schemaResp.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: plan, Config: tfsdk.Config(plan)}, &resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.WithinDuration(t, before.Add(30*time.Minute), receivedDeadline, 2*time.Second)
}

func TestElasticsearchResource_Create_fallsBackToPackageDefaultTimeout(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	var receivedDeadline time.Time
	opts := defaultTestElasticsearchResourceOptions()
	opts.Create = func(ctx context.Context, _ *clients.ElasticsearchScopedClient, req WriteRequest[testResourceModel]) (WriteResult[testResourceModel], diag.Diagnostics) {
		deadline, ok := ctx.Deadline()
		require.True(t, ok)
		receivedDeadline = deadline
		m := req.Plan
		m.ID = types.StringValue("cluster/" + req.WriteID)
		return WriteResult[testResourceModel]{Model: m}, nil
	}
	r := NewElasticsearchResource[testResourceModel]("test_entity", opts)
	r.client = factory
	plan := makeTestResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue))
	before := time.Now()
	resp := resource.CreateResponse{State: tfsdk.State{Raw: tftypes.NewValue(testResourceObjectType(), nil), Schema: testResourceSchemaWithConnectionBlock(ctx)}}
	r.Create(ctx, resource.CreateRequest{Plan: plan, Config: tfsdk.Config(plan)}, &resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.WithinDuration(t, before.Add(DefaultResourceCreateTimeout), receivedDeadline, time.Second)
}

func TestElasticsearchResource_Read_nullStoredTimeoutsUsesDefault(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	var receivedDeadline time.Time
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema: getTestResourceSchema,
		Read: func(ctx context.Context, _ *clients.ElasticsearchScopedClient, _ string, model testResourceModel) (testResourceModel, bool, diag.Diagnostics) {
			deadline, ok := ctx.Deadline()
			require.True(t, ok)
			receivedDeadline = deadline
			return model, true, nil
		},
		Delete: testDeleteFunc,
		Create: testWriteFuncFoundCreate,
		Update: testWriteFuncFoundUpdate,
	})
	r.client = factory

	objType := testResourceObjectTypeWithTimeouts()
	stateValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                       tftypes.NewValue(tftypes.String, "cluster/user1"),
		"name":                     tftypes.NewValue(tftypes.String, "user1"),
		"elasticsearch_connection": tftypes.NewValue(elasticsearchConnectionBlockType(), nil),
		"timeouts":                 resourceTimeoutsNullValue(),
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

type esTimeoutsVersionModel struct {
	ResourceTimeoutsField
	ElasticsearchConnectionField
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (m esTimeoutsVersionModel) GetID() types.String         { return m.ID }
func (m esTimeoutsVersionModel) GetResourceID() types.String { return m.Name }

func (esTimeoutsVersionModel) GetVersionRequirements(_ context.Context) ([]VersionRequirement, diag.Diagnostics) {
	return []VersionRequirement{{
		MinVersion:   *version.Must(version.NewVersion("8.0.0")),
		ErrorMessage: "requires Elasticsearch 8.0.0+",
	}}, nil
}

func newSlowElasticsearchStatusServer(block time.Duration) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/" {
			time.Sleep(block)
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Elastic-Product", "Elasticsearch")
			fmt.Fprint(w, `{"cluster_uuid":"test-cluster","version":{"number":"8.19.0","build_flavor":"default"}}`)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func TestElasticsearchResource_Read_versionCheckTimesOutBeforeCallback(t *testing.T) {
	ctx := context.Background()
	srv := newSlowElasticsearchStatusServer(3 * time.Second)
	defer srv.Close()

	readCalled := false
	r := NewElasticsearchResource[esTimeoutsVersionModel]("test_entity", ElasticsearchResourceOptions[esTimeoutsVersionModel]{
		Schema: func(_ context.Context) rschema.Schema {
			return rschema.Schema{
				Attributes: map[string]rschema.Attribute{
					"id":   rschema.StringAttribute{Computed: true},
					"name": rschema.StringAttribute{Optional: true},
				},
			}
		},
		Read: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, _ esTimeoutsVersionModel) (esTimeoutsVersionModel, bool, diag.Diagnostics) {
			readCalled = true
			return esTimeoutsVersionModel{}, false, nil
		},
		Delete: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, _ esTimeoutsVersionModel) diag.Diagnostics {
			return nil
		},
		Create: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ WriteRequest[esTimeoutsVersionModel]) (WriteResult[esTimeoutsVersionModel], diag.Diagnostics) {
			return WriteResult[esTimeoutsVersionModel]{}, nil
		},
		Update: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ WriteRequest[esTimeoutsVersionModel]) (WriteResult[esTimeoutsVersionModel], diag.Diagnostics) {
			return WriteResult[esTimeoutsVersionModel]{}, nil
		},
		Timeouts: ResourceTimeouts{Read: 200 * time.Millisecond},
	})
	factory := newElasticsearchFactoryForURL(t, srv.URL)
	r.client = factory

	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":                       tftypes.String,
			"name":                     tftypes.String,
			"elasticsearch_connection": elasticsearchConnectionBlockType(),
			"timeouts":                 resourceTimeoutsObjectType(),
		},
	}
	stateValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                       tftypes.NewValue(tftypes.String, "cluster/user1"),
		"name":                     tftypes.NewValue(tftypes.String, "user1"),
		"elasticsearch_connection": tftypes.NewValue(elasticsearchConnectionBlockType(), nil),
		"timeouts":                 resourceTimeoutsNullValue(),
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

func TestElasticsearchResource_operations_setContextDeadline(t *testing.T) {
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)

	type opCase struct {
		name string
		run  func(t *testing.T, r *ElasticsearchResource[testResourceModel], schema rschema.Schema)
	}
	cases := []opCase{
		{
			name: "read",
			run: func(t *testing.T, r *ElasticsearchResource[testResourceModel], schema rschema.Schema) {
				state := makeTestResourceState(ctx, t, "cluster/user1")
				state.Schema = schema
				var resp resource.ReadResponse
				resp.State = state
				r.Read(ctx, resource.ReadRequest{State: state}, &resp)
				require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			},
		},
		{
			name: "delete",
			run: func(t *testing.T, r *ElasticsearchResource[testResourceModel], schema rschema.Schema) {
				state := makeTestResourceState(ctx, t, "cluster/user1")
				state.Schema = schema
				var resp resource.DeleteResponse
				resp.State = state
				r.Delete(ctx, resource.DeleteRequest{State: state}, &resp)
				require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			},
		},
		{
			name: "create",
			run: func(t *testing.T, r *ElasticsearchResource[testResourceModel], schema rschema.Schema) {
				plan := makeTestResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue))
				plan.Schema = schema
				objType := testResourceObjectType()
				var resp resource.CreateResponse
				resp.State = tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: schema}
				r.Create(ctx, resource.CreateRequest{Plan: plan, Config: tfsdk.Config(plan)}, &resp)
				require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			},
		},
		{
			name: "update",
			run: func(t *testing.T, r *ElasticsearchResource[testResourceModel], schema rschema.Schema) {
				plan := makeTestResourceCreatePlan(ctx, t, tftypes.NewValue(tftypes.String, "cluster/user1"))
				plan.Schema = schema
				state := makeTestResourceState(ctx, t, "cluster/user1")
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
			opts := defaultTestElasticsearchResourceOptions()
			opts.Read = func(ctx context.Context, _ *clients.ElasticsearchScopedClient, _ string, model testResourceModel) (testResourceModel, bool, diag.Diagnostics) {
				_, sawDeadline = ctx.Deadline()
				return model, true, nil
			}
			opts.Delete = func(ctx context.Context, _ *clients.ElasticsearchScopedClient, _ string, _ testResourceModel) diag.Diagnostics {
				_, sawDeadline = ctx.Deadline()
				return nil
			}
			opts.Create = func(ctx context.Context, _ *clients.ElasticsearchScopedClient, req WriteRequest[testResourceModel]) (WriteResult[testResourceModel], diag.Diagnostics) {
				_, sawDeadline = ctx.Deadline()
				m := req.Plan
				m.ID = types.StringValue("cluster/" + req.WriteID)
				return WriteResult[testResourceModel]{Model: m}, nil
			}
			opts.Update = func(ctx context.Context, _ *clients.ElasticsearchScopedClient, req WriteRequest[testResourceModel]) (WriteResult[testResourceModel], diag.Diagnostics) {
				_, sawDeadline = ctx.Deadline()
				m := req.Plan
				m.ID = types.StringValue("cluster/" + req.WriteID)
				return WriteResult[testResourceModel]{Model: m}, nil
			}
			r := NewElasticsearchResource[testResourceModel]("test_entity", opts)
			r.client = factory
			var schemaResp resource.SchemaResponse
			r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
			tc.run(t, r, schemaResp.Schema)
			require.True(t, sawDeadline, "callback must observe context deadline for %s", tc.name)
		})
	}
}

// readFuncDropsTimeouts reconstructs the model like mapSpaceResponseToResourceModel-style
// callbacks that omit ResourceTimeoutsField, yielding a zero timeouts.Value{}.
func readFuncDropsTimeouts(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, prior testResourceModel) (testResourceModel, bool, diag.Diagnostics) {
	return testResourceModel{
		ID:                      prior.ID,
		Name:                    prior.Name,
		ElasticsearchConnection: prior.ElasticsearchConnection,
	}, true, nil
}

func TestElasticsearchResource_Read_preservesTimeoutsWhenCallbackDropsField(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	r := NewElasticsearchResource[testResourceModel]("test_entity", ElasticsearchResourceOptions[testResourceModel]{
		Schema: getTestResourceSchema,
		Read:   readFuncDropsTimeouts,
		Delete: testDeleteFunc,
		Create: testWriteFuncFoundCreate,
		Update: testWriteFuncFoundUpdate,
	})
	r.client = factory

	objType := testResourceObjectTypeWithTimeouts()
	stateValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                       tftypes.NewValue(tftypes.String, "cluster/user1"),
		"name":                     tftypes.NewValue(tftypes.String, "user1"),
		"elasticsearch_connection": tftypes.NewValue(elasticsearchConnectionBlockType(), nil),
		"timeouts":                 resourceTimeoutsWithCreate("10m"),
	})
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	state := tfsdk.State{Raw: stateValue, Schema: schemaResp.Schema}

	var priorModel testResourceModel
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

func TestElasticsearchResource_Create_preservesTimeoutsWhenCallbackDropsField(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	opts := defaultTestElasticsearchResourceOptions()
	opts.Create = func(_ context.Context, _ *clients.ElasticsearchScopedClient, req WriteRequest[testResourceModel]) (WriteResult[testResourceModel], diag.Diagnostics) {
		return WriteResult[testResourceModel]{Model: testResourceModel{
			ID:                      types.StringValue("cluster/" + req.WriteID),
			Name:                    req.Plan.Name,
			ElasticsearchConnection: req.Plan.ElasticsearchConnection,
		}}, nil
	}
	opts.Read = readFuncDropsTimeouts
	r := NewElasticsearchResource[testResourceModel]("test_entity", opts)
	r.client = factory

	objType := testResourceObjectTypeWithTimeouts()
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                       tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"name":                     tftypes.NewValue(tftypes.String, "user1"),
		"elasticsearch_connection": tftypes.NewValue(elasticsearchConnectionBlockType(), nil),
		"timeouts":                 resourceTimeoutsWithCreate("10m"),
	})
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	plan := tfsdk.Plan{Raw: objValue, Schema: schemaResp.Schema}

	var planModel testResourceModel
	require.False(t, plan.Get(ctx, &planModel).HasError())
	wantTimeouts := planModel.GetTimeouts()

	resp := resource.CreateResponse{State: tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: schemaResp.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: plan, Config: tfsdk.Config(plan)}, &resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)

	var gotTimeouts timeouts.Value
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root(attrTimeouts), &gotTimeouts)...)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.Equal(t, wantTimeouts, gotTimeouts)
}

func TestElasticsearchResource_Create_invalidTimeoutPreventsCallback(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)
	createCalled := false
	opts := defaultTestElasticsearchResourceOptions()
	opts.Create = func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ WriteRequest[testResourceModel]) (WriteResult[testResourceModel], diag.Diagnostics) {
		createCalled = true
		t.Fatal("create callback must not run when timeouts.create is invalid")
		return WriteResult[testResourceModel]{}, nil
	}
	r := NewElasticsearchResource[testResourceModel]("test_entity", opts)
	r.client = factory

	objType := testResourceObjectTypeWithTimeouts()
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                       tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"name":                     tftypes.NewValue(tftypes.String, "user1"),
		"elasticsearch_connection": tftypes.NewValue(elasticsearchConnectionBlockType(), nil),
		"timeouts":                 resourceTimeoutsWithCreate("not-a-duration"),
	})
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	plan := tfsdk.Plan{Raw: objValue, Schema: schemaResp.Schema}
	resp := resource.CreateResponse{State: tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: schemaResp.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: plan, Config: tfsdk.Config(plan)}, &resp)
	require.True(t, resp.Diagnostics.HasError())
	require.False(t, createCalled)
}
