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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	eschema "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// OpenRequest is passed to ephemeral Open callbacks.
type OpenRequest[T any] struct {
	Config T
}

// OpenResult is returned by ephemeral Open callbacks.
type OpenResult[T any, S any] struct {
	Model      T
	CloseState S
}

// CloseRequest is passed to ephemeral Close callbacks.
type CloseRequest[S any] struct {
	State S
}

// CloseResponse is returned by ephemeral Close callbacks.
type CloseResponse struct{}

// ephemeralPrivateState is the interface for ephemeral resource private state storage.
type ephemeralPrivateState interface {
	GetKey(ctx context.Context, key string) ([]byte, diag.Diagnostics)
	SetKey(ctx context.Context, key string, value []byte) diag.Diagnostics
}

// ephemeralAdapter holds connection-type-specific operations for [genericEphemeralResource].
// It captures the differences between Elasticsearch and Kibana connection handling.
type ephemeralAdapter[T any, Client MinVersionClient] struct {
	getConnection      func(model T) types.List
	getClient          func(ctx context.Context, factory *clients.ProviderClientFactory, connection types.List) (Client, diag.Diagnostics)
	encodeConn         func(ctx context.Context, connection types.List) ([]byte, diag.Diagnostics)
	decodeConn         func(ctx context.Context, data []byte) (types.List, diag.Diagnostics)
	schemaBlockKey     string
	schemaBlockFactory func() eschema.Block
	errorSummary       string
}

// genericEphemeralResource implements [ephemeral.EphemeralResource] and related
// interfaces for any connection-backed ephemeral resource. All lifecycle
// boilerplate lives here; connection-type-specific operations are delegated to
// the adapter. [ElasticsearchEphemeralResource] and [KibanaEphemeralResource]
// are type aliases over this struct.
type genericEphemeralResource[T any, S any, Client MinVersionClient] struct {
	*EphemeralBase
	schemaFactory func(context.Context) eschema.Schema
	openFunc      func(context.Context, Client, OpenRequest[T]) (OpenResult[T, S], diag.Diagnostics)
	closeFunc     func(context.Context, Client, CloseRequest[S]) (CloseResponse, diag.Diagnostics)
	adapter       ephemeralAdapter[T, Client]
}

func (r *genericEphemeralResource[T, S, Client]) Metadata(_ context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = r.EphemeralBase.Metadata(req.ProviderTypeName)
}

func (r *genericEphemeralResource[T, S, Client]) Configure(_ context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	factory, diags := configureFactoryFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.SetClient(factory)
}

func (r *genericEphemeralResource[T, S, Client]) Schema(ctx context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	schema := r.schemaFactory(ctx)
	blocks := make(map[string]eschema.Block, len(schema.Blocks)+1)
	maps.Copy(blocks, schema.Blocks)
	blocks[r.adapter.schemaBlockKey] = r.adapter.schemaBlockFactory()
	schema.Blocks = blocks
	resp.Schema = schema
}

func (r *genericEphemeralResource[T, S, Client]) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	r.openWithPrivate(ctx, req, resp, resp.Private)
}

func (r *genericEphemeralResource[T, S, Client]) openWithPrivate(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse, private ephemeralPrivateState) {
	var model T
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, connDiags := r.adapter.getClient(ctx, r.Client(), r.adapter.getConnection(model))
	resp.Diagnostics.Append(connDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if vDiags := EnforceVersionRequirements(ctx, client, &model); vDiags.HasError() {
		resp.Diagnostics.Append(vDiags...)
		return
	}

	result, callDiags := r.openFunc(ctx, client, OpenRequest[T]{Config: model})
	resp.Diagnostics.Append(callDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.persistOpenPrivateState(ctx, private, result)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Result.Set(ctx, &result.Model)...)
}

func (r *genericEphemeralResource[T, S, Client]) Close(ctx context.Context, req ephemeral.CloseRequest, resp *ephemeral.CloseResponse) {
	if req.Private == nil {
		return
	}
	r.closeFromPrivate(ctx, req.Private, resp)
}

func (r *genericEphemeralResource[T, S, Client]) closeFromPrivate(ctx context.Context, private ephemeralPrivateState, resp *ephemeral.CloseResponse) {
	if private == nil {
		return
	}

	connectionData, connKeyDiags := private.GetKey(ctx, ephemeralConnectionKey)
	resp.Diagnostics.Append(connKeyDiags...)
	if resp.Diagnostics.HasError() || len(connectionData) == 0 {
		return
	}

	userStateData, userKeyDiags := private.GetKey(ctx, ephemeralUserStateKey)
	resp.Diagnostics.Append(userKeyDiags...)
	if resp.Diagnostics.HasError() || len(userStateData) == 0 {
		return
	}

	connection, connDiags := r.adapter.decodeConn(ctx, connectionData)
	resp.Diagnostics.Append(connDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, clientDiags := r.adapter.getClient(ctx, r.Client(), connection)
	resp.Diagnostics.Append(clientDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state, stateDiags := decodeUserCloseState[S](userStateData)
	resp.Diagnostics.Append(stateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	closeResult, closeDiags := r.closeFunc(ctx, client, CloseRequest[S]{State: state})
	resp.Diagnostics.Append(closeDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	_ = closeResult
}

func (r *genericEphemeralResource[T, S, Client]) persistOpenPrivateState(ctx context.Context, private ephemeralPrivateState, result OpenResult[T, S]) diag.Diagnostics {
	var diags diag.Diagnostics
	if private == nil {
		diags.AddError(r.adapter.errorSummary, "Open response Private must not be nil")
		return diags
	}

	connectionData, connDiags := r.adapter.encodeConn(ctx, r.adapter.getConnection(result.Model))
	diags.Append(connDiags...)
	if diags.HasError() {
		return diags
	}
	diags.Append(private.SetKey(ctx, ephemeralConnectionKey, connectionData)...)

	closeStateData, closeDiags := encodeUserCloseState(result.CloseState)
	diags.Append(closeDiags...)
	if diags.HasError() {
		return diags
	}
	diags.Append(private.SetKey(ctx, ephemeralUserStateKey, closeStateData)...)

	return diags
}
