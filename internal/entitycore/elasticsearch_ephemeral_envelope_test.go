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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	eschema "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

type testEphemeralModel struct {
	ElasticsearchConnectionField
	Value types.String `tfsdk:"value"`
}

type testEphemeralCloseState struct {
	Value string
}

type ephemeralTestPrivateData map[string][]byte

func (p ephemeralTestPrivateData) GetKey(_ context.Context, key string) ([]byte, diag.Diagnostics) {
	if val, ok := p[key]; ok {
		return val, nil
	}
	return nil, nil
}

func (p ephemeralTestPrivateData) SetKey(_ context.Context, key string, value []byte) diag.Diagnostics {
	p[key] = value
	return nil
}

func testEphemeralSchema(_ context.Context) eschema.Schema {
	return eschema.Schema{
		Attributes: map[string]eschema.Attribute{
			"value": eschema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func testEphemeralSchemaWithConnection(ctx context.Context) eschema.Schema {
	var resp ephemeral.SchemaResponse
	r := NewElasticsearchEphemeralResource[testEphemeralModel, testEphemeralCloseState]("test_entity", ElasticsearchEphemeralOptions[testEphemeralModel, testEphemeralCloseState]{
		Schema: testEphemeralSchema,
		Open: func(_ context.Context, _ *clients.ElasticsearchScopedClient, req OpenRequest[testEphemeralModel]) (OpenResult[testEphemeralModel, testEphemeralCloseState], diag.Diagnostics) {
			return OpenResult[testEphemeralModel, testEphemeralCloseState]{
				Model:      req.Config,
				CloseState: testEphemeralCloseState{Value: req.Config.Value.ValueString()},
			}, nil
		},
		Close: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ CloseRequest[testEphemeralCloseState]) (CloseResponse, diag.Diagnostics) {
			return CloseResponse{}, nil
		},
	})
	r.Schema(ctx, ephemeral.SchemaRequest{}, &resp)
	return resp.Schema
}

func configureElasticsearchEphemeral(t *testing.T, r ephemeral.EphemeralResource, factory *clients.ProviderClientFactory) {
	t.Helper()
	var cfgResp ephemeral.ConfigureResponse
	r.(ephemeral.EphemeralResourceWithConfigure).Configure(context.Background(), ephemeral.ConfigureRequest{
		ProviderData: factory,
	}, &cfgResp)
	require.False(t, cfgResp.Diagnostics.HasError(), "Configure must not produce errors: %v", cfgResp.Diagnostics)
}

func buildEphemeralOpenConfig(t *testing.T, schema eschema.Schema, values map[string]tftypes.Value) tfsdk.Config {
	t.Helper()
	attrTypes := make(map[string]tftypes.Type, len(schema.Attributes)+1)
	for name := range schema.Attributes {
		attrTypes[name] = tftypes.String
	}
	attrTypes["elasticsearch_connection"] = elasticsearchConnectionBlockType()

	objValues := make(map[string]tftypes.Value, len(values)+1)
	maps.Copy(objValues, values)
	if _, ok := objValues["elasticsearch_connection"]; !ok {
		objValues["elasticsearch_connection"] = tftypes.NewValue(elasticsearchConnectionBlockType(), nil)
	}

	return tfsdk.Config{
		Raw: tftypes.NewValue(tftypes.Object{AttributeTypes: attrTypes}, objValues),
		Schema: eschema.Schema{
			Attributes: schema.Attributes,
			Blocks: map[string]eschema.Block{
				"elasticsearch_connection": schema.Blocks["elasticsearch_connection"],
			},
		},
	}
}

func TestNewElasticsearchEphemeralResource_panicsOnTfsdkCloseState(t *testing.T) {
	t.Parallel()

	type badCloseState struct {
		KeyID types.String
	}

	assertCloseStatePanic(t, func() {
		NewElasticsearchEphemeralResource[testEphemeralModel, badCloseState]("test_entity", ElasticsearchEphemeralOptions[testEphemeralModel, badCloseState]{
			Schema: testEphemeralSchema,
			Open: func(_ context.Context, _ *clients.ElasticsearchScopedClient, req OpenRequest[testEphemeralModel]) (OpenResult[testEphemeralModel, badCloseState], diag.Diagnostics) {
				return OpenResult[testEphemeralModel, badCloseState]{Model: req.Config}, nil
			},
			Close: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ CloseRequest[badCloseState]) (CloseResponse, diag.Diagnostics) {
				return CloseResponse{}, nil
			},
		})
	}, "badCloseState", "KeyID", "github.com/hashicorp/terraform-plugin-framework/types")
}

func TestNewElasticsearchEphemeralResource_panicsOnNilCallbacks(t *testing.T) {
	t.Parallel()

	require.Panics(t, func() {
		NewElasticsearchEphemeralResource[testEphemeralModel, testEphemeralCloseState]("test_entity", ElasticsearchEphemeralOptions[testEphemeralModel, testEphemeralCloseState]{
			Schema: nil,
			Open:   nil,
			Close:  nil,
		})
	})
}

func TestNewElasticsearchEphemeralResource_typeAssertions(t *testing.T) {
	t.Parallel()

	r := NewElasticsearchEphemeralResource[testEphemeralModel, testEphemeralCloseState]("test_entity", ElasticsearchEphemeralOptions[testEphemeralModel, testEphemeralCloseState]{
		Schema: testEphemeralSchema,
		Open: func(_ context.Context, _ *clients.ElasticsearchScopedClient, req OpenRequest[testEphemeralModel]) (OpenResult[testEphemeralModel, testEphemeralCloseState], diag.Diagnostics) {
			return OpenResult[testEphemeralModel, testEphemeralCloseState]{Model: req.Config}, nil
		},
		Close: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ CloseRequest[testEphemeralCloseState]) (CloseResponse, diag.Diagnostics) {
			return CloseResponse{}, nil
		},
	})

	require.Implements(t, (*ephemeral.EphemeralResource)(nil), r)
	require.Implements(t, (*ephemeral.EphemeralResourceWithConfigure)(nil), r)
	require.Implements(t, (*ephemeral.EphemeralResourceWithClose)(nil), r)
}

func TestNewElasticsearchEphemeralResource_doesNotImplementRenew(t *testing.T) {
	t.Parallel()

	r := NewElasticsearchEphemeralResource[testEphemeralModel, testEphemeralCloseState]("test_entity", ElasticsearchEphemeralOptions[testEphemeralModel, testEphemeralCloseState]{
		Schema: testEphemeralSchema,
		Open: func(_ context.Context, _ *clients.ElasticsearchScopedClient, req OpenRequest[testEphemeralModel]) (OpenResult[testEphemeralModel, testEphemeralCloseState], diag.Diagnostics) {
			return OpenResult[testEphemeralModel, testEphemeralCloseState]{Model: req.Config}, nil
		},
		Close: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ CloseRequest[testEphemeralCloseState]) (CloseResponse, diag.Diagnostics) {
			return CloseResponse{}, nil
		},
	})

	_, ok := any(r).(ephemeral.EphemeralResourceWithRenew)
	require.False(t, ok)
}

func TestNewElasticsearchEphemeralResource_schemaInjection(t *testing.T) {
	t.Parallel()

	r := NewElasticsearchEphemeralResource[testEphemeralModel, testEphemeralCloseState]("test_entity", ElasticsearchEphemeralOptions[testEphemeralModel, testEphemeralCloseState]{
		Schema: testEphemeralSchema,
		Open: func(_ context.Context, _ *clients.ElasticsearchScopedClient, req OpenRequest[testEphemeralModel]) (OpenResult[testEphemeralModel, testEphemeralCloseState], diag.Diagnostics) {
			return OpenResult[testEphemeralModel, testEphemeralCloseState]{Model: req.Config}, nil
		},
		Close: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ CloseRequest[testEphemeralCloseState]) (CloseResponse, diag.Diagnostics) {
			return CloseResponse{}, nil
		},
	})

	var resp ephemeral.SchemaResponse
	r.Schema(context.Background(), ephemeral.SchemaRequest{}, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Schema.Blocks, "elasticsearch_connection")
	require.Contains(t, resp.Schema.Attributes, "value")
}

func TestNewElasticsearchEphemeralResource_Metadata(t *testing.T) {
	t.Parallel()

	r := NewElasticsearchEphemeralResource[testEphemeralModel, testEphemeralCloseState]("security_api_key", ElasticsearchEphemeralOptions[testEphemeralModel, testEphemeralCloseState]{
		Schema: testEphemeralSchema,
		Open: func(_ context.Context, _ *clients.ElasticsearchScopedClient, req OpenRequest[testEphemeralModel]) (OpenResult[testEphemeralModel, testEphemeralCloseState], diag.Diagnostics) {
			return OpenResult[testEphemeralModel, testEphemeralCloseState]{Model: req.Config}, nil
		},
		Close: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ CloseRequest[testEphemeralCloseState]) (CloseResponse, diag.Diagnostics) {
			return CloseResponse{}, nil
		},
	})

	var resp ephemeral.MetadataResponse
	r.Metadata(context.Background(), ephemeral.MetadataRequest{ProviderTypeName: "elasticstack"}, &resp)
	require.Equal(t, "elasticstack_elasticsearch_security_api_key", resp.TypeName)
}

func TestElasticsearchEphemeralResource_Open_happyPath(t *testing.T) {
	ctx := context.Background()

	var (
		openCalls int
		gotClient *clients.ElasticsearchScopedClient
		gotConfig testEphemeralModel
	)

	r := NewElasticsearchEphemeralResource[testEphemeralModel, testEphemeralCloseState]("test_entity", ElasticsearchEphemeralOptions[testEphemeralModel, testEphemeralCloseState]{
		Schema: testEphemeralSchema,
		Open: func(_ context.Context, client *clients.ElasticsearchScopedClient, req OpenRequest[testEphemeralModel]) (OpenResult[testEphemeralModel, testEphemeralCloseState], diag.Diagnostics) {
			openCalls++
			gotClient = client
			gotConfig = req.Config
			model := req.Config
			model.Value = types.StringValue("opened")
			return OpenResult[testEphemeralModel, testEphemeralCloseState]{
				Model:      model,
				CloseState: testEphemeralCloseState{Value: "close-me"},
			}, nil
		},
		Close: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ CloseRequest[testEphemeralCloseState]) (CloseResponse, diag.Diagnostics) {
			return CloseResponse{}, nil
		},
	})

	factory := newElasticsearchFactoryMinimal(t)
	configureElasticsearchEphemeral(t, r, factory)

	schema := testEphemeralSchemaWithConnection(ctx)
	cfg := buildEphemeralOpenConfig(t, schema, map[string]tftypes.Value{
		"value": tftypes.NewValue(tftypes.String, "hello"),
	})

	private := make(ephemeralTestPrivateData)
	var resp ephemeral.OpenResponse
	resp.Result.Schema = schema
	resp.Result.Raw = cfg.Raw
	wrapped := r.(*ElasticsearchEphemeralResource[testEphemeralModel, testEphemeralCloseState])
	wrapped.openWithPrivate(ctx, ephemeral.OpenRequest{Config: cfg}, &resp, private)

	require.Equal(t, 1, openCalls)
	require.NotNil(t, gotClient)
	require.Equal(t, "hello", gotConfig.Value.ValueString())
	require.False(t, resp.Diagnostics.HasError())

	var result testEphemeralModel
	require.False(t, resp.Result.Get(ctx, &result).HasError())
	require.Equal(t, "opened", result.Value.ValueString())

	require.NotEmpty(t, private[ephemeralConnectionKey])
	require.NotEmpty(t, private[ephemeralUserStateKey])

	_, connDiags := decodeElasticsearchConnection(ctx, private[ephemeralConnectionKey])
	require.False(t, connDiags.HasError())

	decodedState, stateDiags := decodeUserCloseState[testEphemeralCloseState](private[ephemeralUserStateKey])
	require.False(t, stateDiags.HasError())
	require.Equal(t, "close-me", decodedState.Value)
}

func TestElasticsearchEphemeralResource_Open_decodeErrorShortCircuits(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	openCalled := false
	r := NewElasticsearchEphemeralResource[testEphemeralModel, testEphemeralCloseState]("test_entity", ElasticsearchEphemeralOptions[testEphemeralModel, testEphemeralCloseState]{
		Schema: testEphemeralSchema,
		Open: func(_ context.Context, _ *clients.ElasticsearchScopedClient, req OpenRequest[testEphemeralModel]) (OpenResult[testEphemeralModel, testEphemeralCloseState], diag.Diagnostics) {
			openCalled = true
			return OpenResult[testEphemeralModel, testEphemeralCloseState]{Model: req.Config}, nil
		},
		Close: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ CloseRequest[testEphemeralCloseState]) (CloseResponse, diag.Diagnostics) {
			return CloseResponse{}, nil
		},
	})

	factory := newTestConfiguredFactory(ctx, t)
	configureElasticsearchEphemeral(t, r, factory)

	schema := testEphemeralSchemaWithConnection(ctx)
	cfg := tfsdk.Config{
		Raw: tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
			"value":                    tftypes.Number,
			"elasticsearch_connection": elasticsearchConnectionBlockType(),
		}}, map[string]tftypes.Value{
			"value":                    tftypes.NewValue(tftypes.Number, 42),
			"elasticsearch_connection": tftypes.NewValue(elasticsearchConnectionBlockType(), nil),
		}),
		Schema: schema,
	}

	var resp ephemeral.OpenResponse
	r.(ephemeral.EphemeralResourceWithClose).Open(ctx, ephemeral.OpenRequest{Config: cfg}, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.False(t, openCalled)
}

func TestElasticsearchEphemeralResource_Open_clientResolutionErrorShortCircuits(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	openCalled := false
	r := NewElasticsearchEphemeralResource[testEphemeralModel, testEphemeralCloseState]("test_entity", ElasticsearchEphemeralOptions[testEphemeralModel, testEphemeralCloseState]{
		Schema: testEphemeralSchema,
		Open: func(_ context.Context, _ *clients.ElasticsearchScopedClient, req OpenRequest[testEphemeralModel]) (OpenResult[testEphemeralModel, testEphemeralCloseState], diag.Diagnostics) {
			openCalled = true
			return OpenResult[testEphemeralModel, testEphemeralCloseState]{Model: req.Config}, nil
		},
		Close: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ CloseRequest[testEphemeralCloseState]) (CloseResponse, diag.Diagnostics) {
			return CloseResponse{}, nil
		},
	})

	configureElasticsearchEphemeral(t, r, nonNilTestFactory())

	schema := testEphemeralSchemaWithConnection(ctx)
	cfg := buildEphemeralOpenConfig(t, schema, map[string]tftypes.Value{
		"value": tftypes.NewValue(tftypes.String, "hello"),
	})

	var resp ephemeral.OpenResponse
	r.(ephemeral.EphemeralResourceWithClose).Open(ctx, ephemeral.OpenRequest{Config: cfg}, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Provider not configured")
	require.False(t, openCalled)
}

func TestElasticsearchEphemeralResource_Open_userCallbackDiagnostics(t *testing.T) {
	ctx := context.Background()

	r := NewElasticsearchEphemeralResource[testEphemeralModel, testEphemeralCloseState]("test_entity", ElasticsearchEphemeralOptions[testEphemeralModel, testEphemeralCloseState]{
		Schema: testEphemeralSchema,
		Open: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ OpenRequest[testEphemeralModel]) (OpenResult[testEphemeralModel, testEphemeralCloseState], diag.Diagnostics) {
			var diags diag.Diagnostics
			diags.AddError("open failed", "injected")
			return OpenResult[testEphemeralModel, testEphemeralCloseState]{}, diags
		},
		Close: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ CloseRequest[testEphemeralCloseState]) (CloseResponse, diag.Diagnostics) {
			return CloseResponse{}, nil
		},
	})

	factory := newElasticsearchFactoryMinimal(t)
	configureElasticsearchEphemeral(t, r, factory)

	schema := testEphemeralSchemaWithConnection(ctx)
	cfg := buildEphemeralOpenConfig(t, schema, map[string]tftypes.Value{
		"value": tftypes.NewValue(tftypes.String, "hello"),
	})

	var resp ephemeral.OpenResponse
	private := make(ephemeralTestPrivateData)
	wrapped := r.(*ElasticsearchEphemeralResource[testEphemeralModel, testEphemeralCloseState])
	wrapped.openWithPrivate(ctx, ephemeral.OpenRequest{Config: cfg}, &resp, private)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "open failed")
	require.True(t, resp.Result.Raw.IsNull() || !resp.Result.Raw.IsKnown())
	require.Empty(t, private[ephemeralConnectionKey])
	require.Empty(t, private[ephemeralUserStateKey])
}

func TestElasticsearchEphemeralResource_persistOpenPrivateState(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	r := NewElasticsearchEphemeralResource[testEphemeralModel, testEphemeralCloseState]("test_entity", ElasticsearchEphemeralOptions[testEphemeralModel, testEphemeralCloseState]{
		Schema: testEphemeralSchema,
		Open: func(_ context.Context, _ *clients.ElasticsearchScopedClient, req OpenRequest[testEphemeralModel]) (OpenResult[testEphemeralModel, testEphemeralCloseState], diag.Diagnostics) {
			return OpenResult[testEphemeralModel, testEphemeralCloseState]{Model: req.Config}, nil
		},
		Close: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ CloseRequest[testEphemeralCloseState]) (CloseResponse, diag.Diagnostics) {
			return CloseResponse{}, nil
		},
	})

	private := make(ephemeralTestPrivateData)
	model := testEphemeralModel{Value: types.StringValue("opened")}
	diags := r.(*ElasticsearchEphemeralResource[testEphemeralModel, testEphemeralCloseState]).persistOpenPrivateState(ctx, private, OpenResult[testEphemeralModel, testEphemeralCloseState]{
		Model:      model,
		CloseState: testEphemeralCloseState{Value: "close-me"},
	})
	require.False(t, diags.HasError())
	require.NotEmpty(t, private[ephemeralConnectionKey])
	require.NotEmpty(t, private[ephemeralUserStateKey])
}

func TestElasticsearchEphemeralResource_Close_missingPrivateStateNoOp(t *testing.T) {
	ctx := context.Background()

	closeCalled := false
	r := NewElasticsearchEphemeralResource[testEphemeralModel, testEphemeralCloseState]("test_entity", ElasticsearchEphemeralOptions[testEphemeralModel, testEphemeralCloseState]{
		Schema: testEphemeralSchema,
		Open: func(_ context.Context, _ *clients.ElasticsearchScopedClient, req OpenRequest[testEphemeralModel]) (OpenResult[testEphemeralModel, testEphemeralCloseState], diag.Diagnostics) {
			return OpenResult[testEphemeralModel, testEphemeralCloseState]{Model: req.Config}, nil
		},
		Close: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ CloseRequest[testEphemeralCloseState]) (CloseResponse, diag.Diagnostics) {
			closeCalled = true
			return CloseResponse{}, nil
		},
	})

	factory := newElasticsearchFactoryMinimal(t)
	configureElasticsearchEphemeral(t, r, factory)

	var resp ephemeral.CloseResponse
	r.(ephemeral.EphemeralResourceWithClose).Close(ctx, ephemeral.CloseRequest{}, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.False(t, closeCalled)
}

func TestElasticsearchEphemeralResource_Close_onlyConnectionSlotNoOp(t *testing.T) {
	ctx := context.Background()

	closeCalled := false
	r := NewElasticsearchEphemeralResource[testEphemeralModel, testEphemeralCloseState]("test_entity", ElasticsearchEphemeralOptions[testEphemeralModel, testEphemeralCloseState]{
		Schema: testEphemeralSchema,
		Open: func(_ context.Context, _ *clients.ElasticsearchScopedClient, req OpenRequest[testEphemeralModel]) (OpenResult[testEphemeralModel, testEphemeralCloseState], diag.Diagnostics) {
			return OpenResult[testEphemeralModel, testEphemeralCloseState]{Model: req.Config}, nil
		},
		Close: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ CloseRequest[testEphemeralCloseState]) (CloseResponse, diag.Diagnostics) {
			closeCalled = true
			return CloseResponse{}, nil
		},
	})

	factory := newElasticsearchFactoryMinimal(t)
	configureElasticsearchEphemeral(t, r, factory)

	private := make(ephemeralTestPrivateData)
	connData, _ := encodeElasticsearchConnection(ctx, providerschema.ElasticsearchConnectionNullList())
	private[ephemeralConnectionKey] = connData

	var resp ephemeral.CloseResponse
	r.(*ElasticsearchEphemeralResource[testEphemeralModel, testEphemeralCloseState]).closeFromPrivate(ctx, private, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.False(t, closeCalled)
}

func TestElasticsearchEphemeralResource_Close_onlyUserStateSlotNoOp(t *testing.T) {
	ctx := context.Background()

	closeCalled := false
	r := NewElasticsearchEphemeralResource[testEphemeralModel, testEphemeralCloseState]("test_entity", ElasticsearchEphemeralOptions[testEphemeralModel, testEphemeralCloseState]{
		Schema: testEphemeralSchema,
		Open: func(_ context.Context, _ *clients.ElasticsearchScopedClient, req OpenRequest[testEphemeralModel]) (OpenResult[testEphemeralModel, testEphemeralCloseState], diag.Diagnostics) {
			return OpenResult[testEphemeralModel, testEphemeralCloseState]{Model: req.Config}, nil
		},
		Close: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ CloseRequest[testEphemeralCloseState]) (CloseResponse, diag.Diagnostics) {
			closeCalled = true
			return CloseResponse{}, nil
		},
	})

	factory := newElasticsearchFactoryMinimal(t)
	configureElasticsearchEphemeral(t, r, factory)

	private := make(ephemeralTestPrivateData)
	stateData, _ := encodeUserCloseState(testEphemeralCloseState{Value: "x"})
	private[ephemeralUserStateKey] = stateData

	var resp ephemeral.CloseResponse
	r.(*ElasticsearchEphemeralResource[testEphemeralModel, testEphemeralCloseState]).closeFromPrivate(ctx, private, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.False(t, closeCalled)
}

func TestElasticsearchEphemeralResource_Close_malformedConnectionJSON(t *testing.T) {
	ctx := context.Background()

	closeCalled := false
	r := NewElasticsearchEphemeralResource[testEphemeralModel, testEphemeralCloseState]("test_entity", ElasticsearchEphemeralOptions[testEphemeralModel, testEphemeralCloseState]{
		Schema: testEphemeralSchema,
		Open: func(_ context.Context, _ *clients.ElasticsearchScopedClient, req OpenRequest[testEphemeralModel]) (OpenResult[testEphemeralModel, testEphemeralCloseState], diag.Diagnostics) {
			return OpenResult[testEphemeralModel, testEphemeralCloseState]{Model: req.Config}, nil
		},
		Close: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ CloseRequest[testEphemeralCloseState]) (CloseResponse, diag.Diagnostics) {
			closeCalled = true
			return CloseResponse{}, nil
		},
	})

	factory := newElasticsearchFactoryMinimal(t)
	configureElasticsearchEphemeral(t, r, factory)

	private := make(ephemeralTestPrivateData)
	private[ephemeralConnectionKey] = []byte("not json")
	stateData, _ := encodeUserCloseState(testEphemeralCloseState{Value: "x"})
	private[ephemeralUserStateKey] = stateData

	var resp ephemeral.CloseResponse
	r.(*ElasticsearchEphemeralResource[testEphemeralModel, testEphemeralCloseState]).closeFromPrivate(ctx, private, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.False(t, closeCalled)
}

func TestElasticsearchEphemeralResource_Close_restoresConnectionAndInvokesCallback(t *testing.T) {
	ctx := context.Background()

	var receivedState testEphemeralCloseState
	closeCalled := false
	r := NewElasticsearchEphemeralResource[testEphemeralModel, testEphemeralCloseState]("test_entity", ElasticsearchEphemeralOptions[testEphemeralModel, testEphemeralCloseState]{
		Schema: testEphemeralSchema,
		Open: func(_ context.Context, _ *clients.ElasticsearchScopedClient, req OpenRequest[testEphemeralModel]) (OpenResult[testEphemeralModel, testEphemeralCloseState], diag.Diagnostics) {
			return OpenResult[testEphemeralModel, testEphemeralCloseState]{Model: req.Config}, nil
		},
		Close: func(_ context.Context, client *clients.ElasticsearchScopedClient, req CloseRequest[testEphemeralCloseState]) (CloseResponse, diag.Diagnostics) {
			closeCalled = true
			receivedState = req.State
			require.NotNil(t, client)
			return CloseResponse{}, nil
		},
	})

	factory := newElasticsearchFactoryMinimal(t)
	configureElasticsearchEphemeral(t, r, factory)

	private := make(ephemeralTestPrivateData)
	connData, connDiags := encodeElasticsearchConnection(ctx, providerschema.ElasticsearchConnectionNullList())
	require.False(t, connDiags.HasError())
	stateData, stateDiags := encodeUserCloseState(testEphemeralCloseState{Value: "persisted"})
	require.False(t, stateDiags.HasError())
	private[ephemeralConnectionKey] = connData
	private[ephemeralUserStateKey] = stateData

	var resp ephemeral.CloseResponse
	r.(*ElasticsearchEphemeralResource[testEphemeralModel, testEphemeralCloseState]).closeFromPrivate(ctx, private, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.True(t, closeCalled)
	require.Equal(t, "persisted", receivedState.Value)
}

func TestElasticsearchEphemeralResource_Close_userCallbackDiagnostics(t *testing.T) {
	ctx := context.Background()

	r := NewElasticsearchEphemeralResource[testEphemeralModel, testEphemeralCloseState]("test_entity", ElasticsearchEphemeralOptions[testEphemeralModel, testEphemeralCloseState]{
		Schema: testEphemeralSchema,
		Open: func(_ context.Context, _ *clients.ElasticsearchScopedClient, req OpenRequest[testEphemeralModel]) (OpenResult[testEphemeralModel, testEphemeralCloseState], diag.Diagnostics) {
			return OpenResult[testEphemeralModel, testEphemeralCloseState]{Model: req.Config}, nil
		},
		Close: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ CloseRequest[testEphemeralCloseState]) (CloseResponse, diag.Diagnostics) {
			var diags diag.Diagnostics
			diags.AddError("close failed", "injected")
			return CloseResponse{}, diags
		},
	})

	factory := newElasticsearchFactoryMinimal(t)
	configureElasticsearchEphemeral(t, r, factory)

	private := make(ephemeralTestPrivateData)
	connData, _ := encodeElasticsearchConnection(ctx, providerschema.ElasticsearchConnectionNullList())
	stateData, _ := encodeUserCloseState(testEphemeralCloseState{Value: "x"})
	private[ephemeralConnectionKey] = connData
	private[ephemeralUserStateKey] = stateData

	var resp ephemeral.CloseResponse
	r.(*ElasticsearchEphemeralResource[testEphemeralModel, testEphemeralCloseState]).closeFromPrivate(ctx, private, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "close failed")
}

func TestElasticsearchEphemeralResource_Open_versionRequirementErrorShortCircuits(t *testing.T) {
	ctx := context.Background()

	srv := newMockElasticsearchStatusServer("7.17.0")
	defer srv.Close()

	openCalled := false
	r := NewElasticsearchEphemeralResource[esUnsupportedVersionModel, testEphemeralCloseState]("unsupported_entity", ElasticsearchEphemeralOptions[esUnsupportedVersionModel, testEphemeralCloseState]{
		Schema: func(_ context.Context) eschema.Schema {
			return eschema.Schema{
				Attributes: map[string]eschema.Attribute{
					"id": eschema.StringAttribute{Computed: true},
				},
			}
		},
		Open: func(_ context.Context, _ *clients.ElasticsearchScopedClient, req OpenRequest[esUnsupportedVersionModel]) (OpenResult[esUnsupportedVersionModel, testEphemeralCloseState], diag.Diagnostics) {
			openCalled = true
			return OpenResult[esUnsupportedVersionModel, testEphemeralCloseState]{Model: req.Config}, nil
		},
		Close: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ CloseRequest[testEphemeralCloseState]) (CloseResponse, diag.Diagnostics) {
			return CloseResponse{}, nil
		},
	})

	factory := newElasticsearchFactoryForURL(t, srv.URL)
	configureElasticsearchEphemeral(t, r, factory)

	var schemaResp ephemeral.SchemaResponse
	r.Schema(ctx, ephemeral.SchemaRequest{}, &schemaResp)
	cfg := buildEphemeralOpenConfig(t, schemaResp.Schema, map[string]tftypes.Value{
		"id": tftypes.NewValue(tftypes.String, "test"),
	})

	var resp ephemeral.OpenResponse
	r.(ephemeral.EphemeralResourceWithClose).Open(ctx, ephemeral.OpenRequest{Config: cfg}, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Unsupported server version")
	require.False(t, openCalled)
}

func TestElasticsearchEphemeralResource_Configure_invalidProviderDataLeavesPriorClient(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	r := NewElasticsearchEphemeralResource[testEphemeralModel, testEphemeralCloseState]("test_entity", ElasticsearchEphemeralOptions[testEphemeralModel, testEphemeralCloseState]{
		Schema: testEphemeralSchema,
		Open: func(_ context.Context, _ *clients.ElasticsearchScopedClient, req OpenRequest[testEphemeralModel]) (OpenResult[testEphemeralModel, testEphemeralCloseState], diag.Diagnostics) {
			return OpenResult[testEphemeralModel, testEphemeralCloseState]{Model: req.Config}, nil
		},
		Close: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ CloseRequest[testEphemeralCloseState]) (CloseResponse, diag.Diagnostics) {
			return CloseResponse{}, nil
		},
	})

	f := nonNilTestFactory()
	var okResp ephemeral.ConfigureResponse
	r.(ephemeral.EphemeralResourceWithConfigure).Configure(ctx, ephemeral.ConfigureRequest{ProviderData: f}, &okResp)
	require.False(t, okResp.Diagnostics.HasError())
	require.Same(t, f, r.(*ElasticsearchEphemeralResource[testEphemeralModel, testEphemeralCloseState]).Client())

	var badResp ephemeral.ConfigureResponse
	r.(ephemeral.EphemeralResourceWithConfigure).Configure(ctx, ephemeral.ConfigureRequest{ProviderData: "wrong-type"}, &badResp)
	require.True(t, badResp.Diagnostics.HasError())
	require.Same(t, f, r.(*ElasticsearchEphemeralResource[testEphemeralModel, testEphemeralCloseState]).Client())
}

func TestElasticsearchEphemeralResource_Configure_typedNilFactoryLeavesPriorClient(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	r := NewElasticsearchEphemeralResource[testEphemeralModel, testEphemeralCloseState]("test_entity", ElasticsearchEphemeralOptions[testEphemeralModel, testEphemeralCloseState]{
		Schema: testEphemeralSchema,
		Open: func(_ context.Context, _ *clients.ElasticsearchScopedClient, req OpenRequest[testEphemeralModel]) (OpenResult[testEphemeralModel, testEphemeralCloseState], diag.Diagnostics) {
			return OpenResult[testEphemeralModel, testEphemeralCloseState]{Model: req.Config}, nil
		},
		Close: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ CloseRequest[testEphemeralCloseState]) (CloseResponse, diag.Diagnostics) {
			return CloseResponse{}, nil
		},
	})

	f := nonNilTestFactory()
	var okResp ephemeral.ConfigureResponse
	r.(ephemeral.EphemeralResourceWithConfigure).Configure(ctx, ephemeral.ConfigureRequest{ProviderData: f}, &okResp)
	require.False(t, okResp.Diagnostics.HasError())
	require.Same(t, f, r.(*ElasticsearchEphemeralResource[testEphemeralModel, testEphemeralCloseState]).Client())

	var nilFactory *clients.ProviderClientFactory
	var badResp ephemeral.ConfigureResponse
	r.(ephemeral.EphemeralResourceWithConfigure).Configure(ctx, ephemeral.ConfigureRequest{ProviderData: nilFactory}, &badResp)
	require.True(t, badResp.Diagnostics.HasError())
	require.Same(t, f, r.(*ElasticsearchEphemeralResource[testEphemeralModel, testEphemeralCloseState]).Client())
}
