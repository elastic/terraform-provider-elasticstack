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
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	eschema "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

func TestKibanaConnectionSnapshot_inEnvelopeContext(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	connData, diags := encodeKibanaConnection(ctx, providerschema.KibanaConnectionNullList())
	require.False(t, diags.HasError())

	restored, restoreDiags := decodeKibanaConnection(ctx, connData)
	require.False(t, restoreDiags.HasError())
	require.True(t, restored.IsNull())
}
