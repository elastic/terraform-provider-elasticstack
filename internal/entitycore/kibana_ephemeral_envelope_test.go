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

type testKibanaEphemeralModel struct {
	KibanaConnectionField
	Value types.String `tfsdk:"value"`
}

type testKibanaEphemeralCloseState struct {
	Value string
}

func testKibanaEphemeralSchema(_ context.Context) eschema.Schema {
	return eschema.Schema{
		Attributes: map[string]eschema.Attribute{
			"value": eschema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func configureKibanaEphemeral(t *testing.T, r ephemeral.EphemeralResource, factory *clients.ProviderClientFactory) {
	t.Helper()
	var cfgResp ephemeral.ConfigureResponse
	r.(ephemeral.EphemeralResourceWithConfigure).Configure(context.Background(), ephemeral.ConfigureRequest{
		ProviderData: factory,
	}, &cfgResp)
	require.False(t, cfgResp.Diagnostics.HasError(), "Configure must not produce errors: %v", cfgResp.Diagnostics)
}

func testKibanaEphemeralSchemaWithConnection(ctx context.Context) eschema.Schema {
	var resp ephemeral.SchemaResponse
	r := NewKibanaEphemeralResource[testKibanaEphemeralModel, testKibanaEphemeralCloseState]("test_entity", KibanaEphemeralOptions[testKibanaEphemeralModel, testKibanaEphemeralCloseState]{
		Schema: testKibanaEphemeralSchema,
		Open: func(_ context.Context, _ *clients.KibanaScopedClient, req OpenRequest[testKibanaEphemeralModel]) (OpenResult[testKibanaEphemeralModel, testKibanaEphemeralCloseState], diag.Diagnostics) {
			return OpenResult[testKibanaEphemeralModel, testKibanaEphemeralCloseState]{Model: req.Config}, nil
		},
		Close: func(_ context.Context, _ *clients.KibanaScopedClient, _ CloseRequest[testKibanaEphemeralCloseState]) (CloseResponse, diag.Diagnostics) {
			return CloseResponse{}, nil
		},
	})
	r.Schema(ctx, ephemeral.SchemaRequest{}, &resp)
	return resp.Schema
}

func buildKibanaEphemeralOpenConfig(t *testing.T, schema eschema.Schema, values map[string]tftypes.Value) tfsdk.Config {
	t.Helper()
	attrTypes := make(map[string]tftypes.Type, len(schema.Attributes)+1)
	for name := range schema.Attributes {
		attrTypes[name] = tftypes.String
	}
	attrTypes["kibana_connection"] = kibanaConnectionBlockType()

	objValues := make(map[string]tftypes.Value, len(values)+1)
	maps.Copy(objValues, values)
	if _, ok := objValues["kibana_connection"]; !ok {
		objValues["kibana_connection"] = tftypes.NewValue(kibanaConnectionBlockType(), nil)
	}

	return tfsdk.Config{
		Raw: tftypes.NewValue(tftypes.Object{AttributeTypes: attrTypes}, objValues),
		Schema: eschema.Schema{
			Attributes: schema.Attributes,
			Blocks: map[string]eschema.Block{
				"kibana_connection": schema.Blocks["kibana_connection"],
			},
		},
	}
}

func TestNewKibanaEphemeralResource_panicsOnTfsdkCloseState(t *testing.T) {
	t.Parallel()

	type badCloseState struct {
		KeyID types.String
	}

	assertCloseStatePanic(t, func() {
		NewKibanaEphemeralResource[testKibanaEphemeralModel, badCloseState]("test_entity", KibanaEphemeralOptions[testKibanaEphemeralModel, badCloseState]{
			Schema: testKibanaEphemeralSchema,
			Open: func(_ context.Context, _ *clients.KibanaScopedClient, req OpenRequest[testKibanaEphemeralModel]) (OpenResult[testKibanaEphemeralModel, badCloseState], diag.Diagnostics) {
				return OpenResult[testKibanaEphemeralModel, badCloseState]{Model: req.Config}, nil
			},
			Close: func(_ context.Context, _ *clients.KibanaScopedClient, _ CloseRequest[badCloseState]) (CloseResponse, diag.Diagnostics) {
				return CloseResponse{}, nil
			},
		})
	}, "badCloseState", "KeyID", "github.com/hashicorp/terraform-plugin-framework/types")
}

func TestNewKibanaEphemeralResource_typeAssertions(t *testing.T) {
	t.Parallel()

	r := NewKibanaEphemeralResource[testKibanaEphemeralModel, testKibanaEphemeralCloseState]("test_entity", KibanaEphemeralOptions[testKibanaEphemeralModel, testKibanaEphemeralCloseState]{
		Schema: testKibanaEphemeralSchema,
		Open: func(_ context.Context, _ *clients.KibanaScopedClient, req OpenRequest[testKibanaEphemeralModel]) (OpenResult[testKibanaEphemeralModel, testKibanaEphemeralCloseState], diag.Diagnostics) {
			return OpenResult[testKibanaEphemeralModel, testKibanaEphemeralCloseState]{Model: req.Config}, nil
		},
		Close: func(_ context.Context, _ *clients.KibanaScopedClient, _ CloseRequest[testKibanaEphemeralCloseState]) (CloseResponse, diag.Diagnostics) {
			return CloseResponse{}, nil
		},
	})

	require.Implements(t, (*ephemeral.EphemeralResource)(nil), r)
	require.Implements(t, (*ephemeral.EphemeralResourceWithConfigure)(nil), r)
	require.Implements(t, (*ephemeral.EphemeralResourceWithClose)(nil), r)
}

func TestNewKibanaEphemeralResource_doesNotImplementRenew(t *testing.T) {
	t.Parallel()

	r := NewKibanaEphemeralResource[testKibanaEphemeralModel, testKibanaEphemeralCloseState]("test_entity", KibanaEphemeralOptions[testKibanaEphemeralModel, testKibanaEphemeralCloseState]{
		Schema: testKibanaEphemeralSchema,
		Open: func(_ context.Context, _ *clients.KibanaScopedClient, req OpenRequest[testKibanaEphemeralModel]) (OpenResult[testKibanaEphemeralModel, testKibanaEphemeralCloseState], diag.Diagnostics) {
			return OpenResult[testKibanaEphemeralModel, testKibanaEphemeralCloseState]{Model: req.Config}, nil
		},
		Close: func(_ context.Context, _ *clients.KibanaScopedClient, _ CloseRequest[testKibanaEphemeralCloseState]) (CloseResponse, diag.Diagnostics) {
			return CloseResponse{}, nil
		},
	})

	_, ok := any(r).(ephemeral.EphemeralResourceWithRenew)
	require.False(t, ok)
}

func TestNewKibanaEphemeralResource_schemaInjection(t *testing.T) {
	t.Parallel()

	r := NewKibanaEphemeralResource[testKibanaEphemeralModel, testKibanaEphemeralCloseState]("test_entity", KibanaEphemeralOptions[testKibanaEphemeralModel, testKibanaEphemeralCloseState]{
		Schema: testKibanaEphemeralSchema,
		Open: func(_ context.Context, _ *clients.KibanaScopedClient, req OpenRequest[testKibanaEphemeralModel]) (OpenResult[testKibanaEphemeralModel, testKibanaEphemeralCloseState], diag.Diagnostics) {
			return OpenResult[testKibanaEphemeralModel, testKibanaEphemeralCloseState]{Model: req.Config}, nil
		},
		Close: func(_ context.Context, _ *clients.KibanaScopedClient, _ CloseRequest[testKibanaEphemeralCloseState]) (CloseResponse, diag.Diagnostics) {
			return CloseResponse{}, nil
		},
	})

	var resp ephemeral.SchemaResponse
	r.Schema(context.Background(), ephemeral.SchemaRequest{}, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Schema.Blocks, "kibana_connection")
	require.Contains(t, resp.Schema.Attributes, "value")
}

func TestNewKibanaEphemeralResource_Metadata(t *testing.T) {
	t.Parallel()

	r := NewKibanaEphemeralResource[testKibanaEphemeralModel, testKibanaEphemeralCloseState]("service_account_token", KibanaEphemeralOptions[testKibanaEphemeralModel, testKibanaEphemeralCloseState]{
		Schema: testKibanaEphemeralSchema,
		Open: func(_ context.Context, _ *clients.KibanaScopedClient, req OpenRequest[testKibanaEphemeralModel]) (OpenResult[testKibanaEphemeralModel, testKibanaEphemeralCloseState], diag.Diagnostics) {
			return OpenResult[testKibanaEphemeralModel, testKibanaEphemeralCloseState]{Model: req.Config}, nil
		},
		Close: func(_ context.Context, _ *clients.KibanaScopedClient, _ CloseRequest[testKibanaEphemeralCloseState]) (CloseResponse, diag.Diagnostics) {
			return CloseResponse{}, nil
		},
	})

	var resp ephemeral.MetadataResponse
	r.Metadata(context.Background(), ephemeral.MetadataRequest{ProviderTypeName: "elasticstack"}, &resp)
	require.Equal(t, "elasticstack_kibana_service_account_token", resp.TypeName)
}

func TestKibanaEphemeralResource_Open_happyPath(t *testing.T) {
	ctx := context.Background()

	var (
		openCalls int
		gotClient *clients.KibanaScopedClient
		gotConfig testKibanaEphemeralModel
	)

	r := NewKibanaEphemeralResource[testKibanaEphemeralModel, testKibanaEphemeralCloseState]("test_entity", KibanaEphemeralOptions[testKibanaEphemeralModel, testKibanaEphemeralCloseState]{
		Schema: testKibanaEphemeralSchema,
		Open: func(_ context.Context, client *clients.KibanaScopedClient, req OpenRequest[testKibanaEphemeralModel]) (OpenResult[testKibanaEphemeralModel, testKibanaEphemeralCloseState], diag.Diagnostics) {
			openCalls++
			gotClient = client
			gotConfig = req.Config
			model := req.Config
			model.Value = types.StringValue("opened")
			return OpenResult[testKibanaEphemeralModel, testKibanaEphemeralCloseState]{
				Model:      model,
				CloseState: testKibanaEphemeralCloseState{Value: "close-me"},
			}, nil
		},
		Close: func(_ context.Context, _ *clients.KibanaScopedClient, _ CloseRequest[testKibanaEphemeralCloseState]) (CloseResponse, diag.Diagnostics) {
			return CloseResponse{}, nil
		},
	})

	factory := newKibanaFactoryMinimal(t)
	configureKibanaEphemeral(t, r, factory)

	schema := testKibanaEphemeralSchemaWithConnection(ctx)
	cfg := buildKibanaEphemeralOpenConfig(t, schema, map[string]tftypes.Value{
		"value": tftypes.NewValue(tftypes.String, "hello"),
	})

	private := make(ephemeralTestPrivateData)
	var resp ephemeral.OpenResponse
	resp.Result.Schema = schema
	resp.Result.Raw = cfg.Raw
	wrapped := r.(*KibanaEphemeralResource[testKibanaEphemeralModel, testKibanaEphemeralCloseState])
	wrapped.openWithPrivate(ctx, ephemeral.OpenRequest{Config: cfg}, &resp, private)

	require.Equal(t, 1, openCalls)
	require.NotNil(t, gotClient)
	require.Equal(t, "hello", gotConfig.Value.ValueString())
	require.False(t, resp.Diagnostics.HasError())

	var result testKibanaEphemeralModel
	require.False(t, resp.Result.Get(ctx, &result).HasError())
	require.Equal(t, "opened", result.Value.ValueString())

	require.NotEmpty(t, private[ephemeralConnectionKey])
	require.NotEmpty(t, private[ephemeralUserStateKey])

	_, connDiags := decodeKibanaConnection(ctx, private[ephemeralConnectionKey])
	require.False(t, connDiags.HasError())

	decodedState, stateDiags := decodeUserCloseState[testKibanaEphemeralCloseState](private[ephemeralUserStateKey])
	require.False(t, stateDiags.HasError())
	require.Equal(t, "close-me", decodedState.Value)
}

func TestKibanaEphemeralResource_Open_decodeErrorShortCircuits(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	openCalled := false
	r := NewKibanaEphemeralResource[testKibanaEphemeralModel, testKibanaEphemeralCloseState]("test_entity", KibanaEphemeralOptions[testKibanaEphemeralModel, testKibanaEphemeralCloseState]{
		Schema: testKibanaEphemeralSchema,
		Open: func(_ context.Context, _ *clients.KibanaScopedClient, req OpenRequest[testKibanaEphemeralModel]) (OpenResult[testKibanaEphemeralModel, testKibanaEphemeralCloseState], diag.Diagnostics) {
			openCalled = true
			return OpenResult[testKibanaEphemeralModel, testKibanaEphemeralCloseState]{Model: req.Config}, nil
		},
		Close: func(_ context.Context, _ *clients.KibanaScopedClient, _ CloseRequest[testKibanaEphemeralCloseState]) (CloseResponse, diag.Diagnostics) {
			return CloseResponse{}, nil
		},
	})

	factory := newTestConfiguredFactory(ctx, t)
	configureKibanaEphemeral(t, r, factory)

	schema := testKibanaEphemeralSchemaWithConnection(ctx)
	cfg := tfsdk.Config{
		Raw: tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
			"value":             tftypes.Number,
			"kibana_connection": kibanaConnectionBlockType(),
		}}, map[string]tftypes.Value{
			"value":             tftypes.NewValue(tftypes.Number, 42),
			"kibana_connection": tftypes.NewValue(kibanaConnectionBlockType(), nil),
		}),
		Schema: schema,
	}

	var resp ephemeral.OpenResponse
	r.(*KibanaEphemeralResource[testKibanaEphemeralModel, testKibanaEphemeralCloseState]).openWithPrivate(ctx, ephemeral.OpenRequest{Config: cfg}, &resp, make(ephemeralTestPrivateData))

	require.True(t, resp.Diagnostics.HasError())
	require.False(t, openCalled)
}

func TestKibanaEphemeralResource_Open_clientResolutionErrorShortCircuits(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	openCalled := false
	r := NewKibanaEphemeralResource[testKibanaEphemeralModel, testKibanaEphemeralCloseState]("test_entity", KibanaEphemeralOptions[testKibanaEphemeralModel, testKibanaEphemeralCloseState]{
		Schema: testKibanaEphemeralSchema,
		Open: func(_ context.Context, _ *clients.KibanaScopedClient, req OpenRequest[testKibanaEphemeralModel]) (OpenResult[testKibanaEphemeralModel, testKibanaEphemeralCloseState], diag.Diagnostics) {
			openCalled = true
			return OpenResult[testKibanaEphemeralModel, testKibanaEphemeralCloseState]{Model: req.Config}, nil
		},
		Close: func(_ context.Context, _ *clients.KibanaScopedClient, _ CloseRequest[testKibanaEphemeralCloseState]) (CloseResponse, diag.Diagnostics) {
			return CloseResponse{}, nil
		},
	})

	configureKibanaEphemeral(t, r, nonNilTestFactory())

	schema := testKibanaEphemeralSchemaWithConnection(ctx)
	cfg := buildKibanaEphemeralOpenConfig(t, schema, map[string]tftypes.Value{
		"value": tftypes.NewValue(tftypes.String, "hello"),
	})

	var resp ephemeral.OpenResponse
	r.(*KibanaEphemeralResource[testKibanaEphemeralModel, testKibanaEphemeralCloseState]).openWithPrivate(ctx, ephemeral.OpenRequest{Config: cfg}, &resp, make(ephemeralTestPrivateData))

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Provider not configured")
	require.False(t, openCalled)
}

func TestKibanaEphemeralResource_Open_userCallbackDiagnostics(t *testing.T) {
	ctx := context.Background()

	r := NewKibanaEphemeralResource[testKibanaEphemeralModel, testKibanaEphemeralCloseState]("test_entity", KibanaEphemeralOptions[testKibanaEphemeralModel, testKibanaEphemeralCloseState]{
		Schema: testKibanaEphemeralSchema,
		Open: func(_ context.Context, _ *clients.KibanaScopedClient, _ OpenRequest[testKibanaEphemeralModel]) (OpenResult[testKibanaEphemeralModel, testKibanaEphemeralCloseState], diag.Diagnostics) {
			var diags diag.Diagnostics
			diags.AddError("open failed", "injected")
			return OpenResult[testKibanaEphemeralModel, testKibanaEphemeralCloseState]{}, diags
		},
		Close: func(_ context.Context, _ *clients.KibanaScopedClient, _ CloseRequest[testKibanaEphemeralCloseState]) (CloseResponse, diag.Diagnostics) {
			return CloseResponse{}, nil
		},
	})

	factory := newKibanaFactoryMinimal(t)
	configureKibanaEphemeral(t, r, factory)

	schema := testKibanaEphemeralSchemaWithConnection(ctx)
	cfg := buildKibanaEphemeralOpenConfig(t, schema, map[string]tftypes.Value{
		"value": tftypes.NewValue(tftypes.String, "hello"),
	})

	private := make(ephemeralTestPrivateData)
	var resp ephemeral.OpenResponse
	r.(*KibanaEphemeralResource[testKibanaEphemeralModel, testKibanaEphemeralCloseState]).openWithPrivate(ctx, ephemeral.OpenRequest{Config: cfg}, &resp, private)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "open failed")
	require.True(t, resp.Result.Raw.IsNull() || !resp.Result.Raw.IsKnown())
	require.Empty(t, private[ephemeralConnectionKey])
	require.Empty(t, private[ephemeralUserStateKey])
}

func TestKibanaEphemeralResource_Open_versionRequirementErrorShortCircuits(t *testing.T) {
	ctx := context.Background()

	srv := newMockKibanaStatusServer("7.17.0")
	defer srv.Close()

	openCalled := false
	r := NewKibanaEphemeralResource[unsupportedVersionModel, testKibanaEphemeralCloseState]("unsupported_entity", KibanaEphemeralOptions[unsupportedVersionModel, testKibanaEphemeralCloseState]{
		Schema: func(_ context.Context) eschema.Schema {
			return eschema.Schema{
				Attributes: map[string]eschema.Attribute{
					"id": eschema.StringAttribute{Computed: true},
				},
			}
		},
		Open: func(_ context.Context, _ *clients.KibanaScopedClient, req OpenRequest[unsupportedVersionModel]) (OpenResult[unsupportedVersionModel, testKibanaEphemeralCloseState], diag.Diagnostics) {
			openCalled = true
			return OpenResult[unsupportedVersionModel, testKibanaEphemeralCloseState]{Model: req.Config}, nil
		},
		Close: func(_ context.Context, _ *clients.KibanaScopedClient, _ CloseRequest[testKibanaEphemeralCloseState]) (CloseResponse, diag.Diagnostics) {
			return CloseResponse{}, nil
		},
	})

	factory := newKibanaFactoryForURL(t, srv.URL)
	configureKibanaEphemeral(t, r, factory)

	var schemaResp ephemeral.SchemaResponse
	r.Schema(ctx, ephemeral.SchemaRequest{}, &schemaResp)
	cfg := buildKibanaEphemeralOpenConfig(t, schemaResp.Schema, map[string]tftypes.Value{
		"id": tftypes.NewValue(tftypes.String, "test"),
	})

	var resp ephemeral.OpenResponse
	r.(*KibanaEphemeralResource[unsupportedVersionModel, testKibanaEphemeralCloseState]).openWithPrivate(ctx, ephemeral.OpenRequest{Config: cfg}, &resp, make(ephemeralTestPrivateData))

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Unsupported server version")
	require.False(t, openCalled)
}

func TestKibanaEphemeralResource_persistOpenAndCloseRoundTrip(t *testing.T) {
	ctx := context.Background()

	var receivedState testKibanaEphemeralCloseState
	closeCalled := false
	r := NewKibanaEphemeralResource[testKibanaEphemeralModel, testKibanaEphemeralCloseState]("test_entity", KibanaEphemeralOptions[testKibanaEphemeralModel, testKibanaEphemeralCloseState]{
		Schema: testKibanaEphemeralSchema,
		Open: func(_ context.Context, _ *clients.KibanaScopedClient, req OpenRequest[testKibanaEphemeralModel]) (OpenResult[testKibanaEphemeralModel, testKibanaEphemeralCloseState], diag.Diagnostics) {
			model := req.Config
			model.Value = types.StringValue("opened")
			return OpenResult[testKibanaEphemeralModel, testKibanaEphemeralCloseState]{
				Model:      model,
				CloseState: testKibanaEphemeralCloseState{Value: "close-me"},
			}, nil
		},
		Close: func(_ context.Context, _ *clients.KibanaScopedClient, req CloseRequest[testKibanaEphemeralCloseState]) (CloseResponse, diag.Diagnostics) {
			closeCalled = true
			receivedState = req.State
			return CloseResponse{}, nil
		},
	})

	factory := newKibanaFactoryMinimal(t)
	configureKibanaEphemeral(t, r, factory)

	private := make(ephemeralTestPrivateData)
	model := testKibanaEphemeralModel{Value: types.StringValue("opened")}
	openResult := OpenResult[testKibanaEphemeralModel, testKibanaEphemeralCloseState]{
		Model:      model,
		CloseState: testKibanaEphemeralCloseState{Value: "close-me"},
	}
	wrapped := r.(*KibanaEphemeralResource[testKibanaEphemeralModel, testKibanaEphemeralCloseState])
	diags := wrapped.persistOpenPrivateState(ctx, private, openResult)
	require.False(t, diags.HasError())

	var closeResp ephemeral.CloseResponse
	r.(*KibanaEphemeralResource[testKibanaEphemeralModel, testKibanaEphemeralCloseState]).closeFromPrivate(ctx, private, &closeResp)
	require.False(t, closeResp.Diagnostics.HasError())
	require.True(t, closeCalled)
	require.Equal(t, "close-me", receivedState.Value)
}

func TestKibanaEphemeralResource_Close_missingPrivateStateNoOp(t *testing.T) {
	ctx := context.Background()

	closeCalled := false
	r := NewKibanaEphemeralResource[testKibanaEphemeralModel, testKibanaEphemeralCloseState]("test_entity", KibanaEphemeralOptions[testKibanaEphemeralModel, testKibanaEphemeralCloseState]{
		Schema: testKibanaEphemeralSchema,
		Open: func(_ context.Context, _ *clients.KibanaScopedClient, req OpenRequest[testKibanaEphemeralModel]) (OpenResult[testKibanaEphemeralModel, testKibanaEphemeralCloseState], diag.Diagnostics) {
			return OpenResult[testKibanaEphemeralModel, testKibanaEphemeralCloseState]{Model: req.Config}, nil
		},
		Close: func(_ context.Context, _ *clients.KibanaScopedClient, _ CloseRequest[testKibanaEphemeralCloseState]) (CloseResponse, diag.Diagnostics) {
			closeCalled = true
			return CloseResponse{}, nil
		},
	})

	factory := newKibanaFactoryMinimal(t)
	configureKibanaEphemeral(t, r, factory)

	var resp ephemeral.CloseResponse
	r.(ephemeral.EphemeralResourceWithClose).Close(ctx, ephemeral.CloseRequest{}, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.False(t, closeCalled)
}

func TestKibanaEphemeralResource_Close_onlyConnectionSlotNoOp(t *testing.T) {
	ctx := context.Background()

	closeCalled := false
	r := NewKibanaEphemeralResource[testKibanaEphemeralModel, testKibanaEphemeralCloseState]("test_entity", KibanaEphemeralOptions[testKibanaEphemeralModel, testKibanaEphemeralCloseState]{
		Schema: testKibanaEphemeralSchema,
		Open: func(_ context.Context, _ *clients.KibanaScopedClient, req OpenRequest[testKibanaEphemeralModel]) (OpenResult[testKibanaEphemeralModel, testKibanaEphemeralCloseState], diag.Diagnostics) {
			return OpenResult[testKibanaEphemeralModel, testKibanaEphemeralCloseState]{Model: req.Config}, nil
		},
		Close: func(_ context.Context, _ *clients.KibanaScopedClient, _ CloseRequest[testKibanaEphemeralCloseState]) (CloseResponse, diag.Diagnostics) {
			closeCalled = true
			return CloseResponse{}, nil
		},
	})

	factory := newKibanaFactoryMinimal(t)
	configureKibanaEphemeral(t, r, factory)

	private := make(ephemeralTestPrivateData)
	connData, _ := encodeKibanaConnection(ctx, providerschema.KibanaConnectionNullList())
	private[ephemeralConnectionKey] = connData

	var resp ephemeral.CloseResponse
	r.(*KibanaEphemeralResource[testKibanaEphemeralModel, testKibanaEphemeralCloseState]).closeFromPrivate(ctx, private, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.False(t, closeCalled)
}

func TestKibanaEphemeralResource_Close_onlyUserStateSlotNoOp(t *testing.T) {
	ctx := context.Background()

	closeCalled := false
	r := NewKibanaEphemeralResource[testKibanaEphemeralModel, testKibanaEphemeralCloseState]("test_entity", KibanaEphemeralOptions[testKibanaEphemeralModel, testKibanaEphemeralCloseState]{
		Schema: testKibanaEphemeralSchema,
		Open: func(_ context.Context, _ *clients.KibanaScopedClient, req OpenRequest[testKibanaEphemeralModel]) (OpenResult[testKibanaEphemeralModel, testKibanaEphemeralCloseState], diag.Diagnostics) {
			return OpenResult[testKibanaEphemeralModel, testKibanaEphemeralCloseState]{Model: req.Config}, nil
		},
		Close: func(_ context.Context, _ *clients.KibanaScopedClient, _ CloseRequest[testKibanaEphemeralCloseState]) (CloseResponse, diag.Diagnostics) {
			closeCalled = true
			return CloseResponse{}, nil
		},
	})

	factory := newKibanaFactoryMinimal(t)
	configureKibanaEphemeral(t, r, factory)

	private := make(ephemeralTestPrivateData)
	stateData, _ := encodeUserCloseState(testKibanaEphemeralCloseState{Value: "x"})
	private[ephemeralUserStateKey] = stateData

	var resp ephemeral.CloseResponse
	r.(*KibanaEphemeralResource[testKibanaEphemeralModel, testKibanaEphemeralCloseState]).closeFromPrivate(ctx, private, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.False(t, closeCalled)
}

func TestKibanaEphemeralResource_Close_malformedConnectionJSON(t *testing.T) {
	ctx := context.Background()

	closeCalled := false
	r := NewKibanaEphemeralResource[testKibanaEphemeralModel, testKibanaEphemeralCloseState]("test_entity", KibanaEphemeralOptions[testKibanaEphemeralModel, testKibanaEphemeralCloseState]{
		Schema: testKibanaEphemeralSchema,
		Open: func(_ context.Context, _ *clients.KibanaScopedClient, req OpenRequest[testKibanaEphemeralModel]) (OpenResult[testKibanaEphemeralModel, testKibanaEphemeralCloseState], diag.Diagnostics) {
			return OpenResult[testKibanaEphemeralModel, testKibanaEphemeralCloseState]{Model: req.Config}, nil
		},
		Close: func(_ context.Context, _ *clients.KibanaScopedClient, _ CloseRequest[testKibanaEphemeralCloseState]) (CloseResponse, diag.Diagnostics) {
			closeCalled = true
			return CloseResponse{}, nil
		},
	})

	factory := newKibanaFactoryMinimal(t)
	configureKibanaEphemeral(t, r, factory)

	private := make(ephemeralTestPrivateData)
	private[ephemeralConnectionKey] = []byte("not json")
	stateData, _ := encodeUserCloseState(testKibanaEphemeralCloseState{Value: "x"})
	private[ephemeralUserStateKey] = stateData

	var resp ephemeral.CloseResponse
	r.(*KibanaEphemeralResource[testKibanaEphemeralModel, testKibanaEphemeralCloseState]).closeFromPrivate(ctx, private, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.False(t, closeCalled)
}

func TestKibanaEphemeralResource_Configure_invalidProviderDataLeavesPriorClient(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	r := NewKibanaEphemeralResource[testKibanaEphemeralModel, testKibanaEphemeralCloseState]("test_entity", KibanaEphemeralOptions[testKibanaEphemeralModel, testKibanaEphemeralCloseState]{
		Schema: testKibanaEphemeralSchema,
		Open: func(_ context.Context, _ *clients.KibanaScopedClient, req OpenRequest[testKibanaEphemeralModel]) (OpenResult[testKibanaEphemeralModel, testKibanaEphemeralCloseState], diag.Diagnostics) {
			return OpenResult[testKibanaEphemeralModel, testKibanaEphemeralCloseState]{Model: req.Config}, nil
		},
		Close: func(_ context.Context, _ *clients.KibanaScopedClient, _ CloseRequest[testKibanaEphemeralCloseState]) (CloseResponse, diag.Diagnostics) {
			return CloseResponse{}, nil
		},
	})

	f := nonNilTestFactory()
	var okResp ephemeral.ConfigureResponse
	r.(ephemeral.EphemeralResourceWithConfigure).Configure(ctx, ephemeral.ConfigureRequest{ProviderData: f}, &okResp)
	require.False(t, okResp.Diagnostics.HasError())
	require.Same(t, f, r.(*KibanaEphemeralResource[testKibanaEphemeralModel, testKibanaEphemeralCloseState]).Client())

	var badResp ephemeral.ConfigureResponse
	r.(ephemeral.EphemeralResourceWithConfigure).Configure(ctx, ephemeral.ConfigureRequest{ProviderData: "wrong-type"}, &badResp)
	require.True(t, badResp.Diagnostics.HasError())
	require.Same(t, f, r.(*KibanaEphemeralResource[testKibanaEphemeralModel, testKibanaEphemeralCloseState]).Client())
}

func TestKibanaEphemeralResource_Configure_typedNilFactoryLeavesPriorClient(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	r := NewKibanaEphemeralResource[testKibanaEphemeralModel, testKibanaEphemeralCloseState]("test_entity", KibanaEphemeralOptions[testKibanaEphemeralModel, testKibanaEphemeralCloseState]{
		Schema: testKibanaEphemeralSchema,
		Open: func(_ context.Context, _ *clients.KibanaScopedClient, req OpenRequest[testKibanaEphemeralModel]) (OpenResult[testKibanaEphemeralModel, testKibanaEphemeralCloseState], diag.Diagnostics) {
			return OpenResult[testKibanaEphemeralModel, testKibanaEphemeralCloseState]{Model: req.Config}, nil
		},
		Close: func(_ context.Context, _ *clients.KibanaScopedClient, _ CloseRequest[testKibanaEphemeralCloseState]) (CloseResponse, diag.Diagnostics) {
			return CloseResponse{}, nil
		},
	})

	f := nonNilTestFactory()
	var okResp ephemeral.ConfigureResponse
	r.(ephemeral.EphemeralResourceWithConfigure).Configure(ctx, ephemeral.ConfigureRequest{ProviderData: f}, &okResp)
	require.False(t, okResp.Diagnostics.HasError())
	require.Same(t, f, r.(*KibanaEphemeralResource[testKibanaEphemeralModel, testKibanaEphemeralCloseState]).Client())

	var nilFactory *clients.ProviderClientFactory
	var badResp ephemeral.ConfigureResponse
	r.(ephemeral.EphemeralResourceWithConfigure).Configure(ctx, ephemeral.ConfigureRequest{ProviderData: nilFactory}, &badResp)
	require.True(t, badResp.Diagnostics.HasError())
	require.Same(t, f, r.(*KibanaEphemeralResource[testKibanaEphemeralModel, testKibanaEphemeralCloseState]).Client())
}

func TestKibanaConnectionSnapshot_inEnvelopeContext(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	connData, diags := encodeKibanaConnection(ctx, providerschema.KibanaConnectionNullList())
	require.False(t, diags.HasError())

	restored, restoreDiags := decodeKibanaConnection(ctx, connData)
	require.False(t, restoreDiags.HasError())
	require.True(t, restored.IsNull())
}
