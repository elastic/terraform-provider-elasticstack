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
	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	eschema "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ElasticsearchEphemeralModel is the type constraint for models passed to
// [NewElasticsearchEphemeralResource]. Concrete types must provide
// GetElasticsearchConnection, typically by embedding [ElasticsearchConnectionField].
type ElasticsearchEphemeralModel interface {
	GetElasticsearchConnection() types.List
}

// OpenRequest is passed to Elasticsearch ephemeral Open callbacks.
type OpenRequest[T any] struct {
	Config T
}

// OpenResult is returned by Elasticsearch ephemeral Open callbacks.
type OpenResult[T any, S any] struct {
	Model      T
	CloseState S
}

// CloseRequest is passed to Elasticsearch ephemeral Close callbacks.
type CloseRequest[S any] struct {
	State S
}

// CloseResponse is returned by Elasticsearch ephemeral Close callbacks.
type CloseResponse struct{}

type ElasticsearchEphemeralOpenFunc[T ElasticsearchEphemeralModel, S any] func(
	context.Context,
	*clients.ElasticsearchScopedClient,
	OpenRequest[T],
) (OpenResult[T, S], diag.Diagnostics)

type ElasticsearchEphemeralCloseFunc[S any] func(
	context.Context,
	*clients.ElasticsearchScopedClient,
	CloseRequest[S],
) (CloseResponse, diag.Diagnostics)

// ElasticsearchEphemeralOptions configures [NewElasticsearchEphemeralResource].
// Schema, Open, and Close must be non-nil or the constructor panics.
type ElasticsearchEphemeralOptions[T ElasticsearchEphemeralModel, S any] struct {
	Schema func(context.Context) eschema.Schema
	Open   ElasticsearchEphemeralOpenFunc[T, S]
	Close  ElasticsearchEphemeralCloseFunc[S]
}

// ElasticsearchEphemeralResource implements [ephemeral.EphemeralResource] and
// related interfaces for Elasticsearch-backed ephemeral resources.
type ElasticsearchEphemeralResource[T ElasticsearchEphemeralModel, S any] struct {
	*EphemeralBase
	schemaFactory func(context.Context) eschema.Schema
	openFunc      ElasticsearchEphemeralOpenFunc[T, S]
	closeFunc     ElasticsearchEphemeralCloseFunc[S]
}

// NewElasticsearchEphemeralResource returns an [ephemeral.EphemeralResource]
// that owns Metadata, Configure, Schema (with elasticsearch_connection block
// injection), Open, and Close for the Elasticsearch namespace.
func NewElasticsearchEphemeralResource[T ElasticsearchEphemeralModel, S any](
	name string,
	opts ElasticsearchEphemeralOptions[T, S],
) ephemeral.EphemeralResource {
	if opts.Schema == nil {
		panic("entitycore: ElasticsearchEphemeralOptions.Schema must not be nil")
	}
	if opts.Open == nil {
		panic("entitycore: ElasticsearchEphemeralOptions.Open must not be nil")
	}
	if opts.Close == nil {
		panic("entitycore: ElasticsearchEphemeralOptions.Close must not be nil")
	}
	mustBePlainGoCloseState[S]()

	return &ElasticsearchEphemeralResource[T, S]{
		EphemeralBase: NewEphemeralBase(ComponentElasticsearch, name),
		schemaFactory: opts.Schema,
		openFunc:      opts.Open,
		closeFunc:     opts.Close,
	}
}

func (r *ElasticsearchEphemeralResource[T, S]) Metadata(_ context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = r.EphemeralBase.Metadata(req.ProviderTypeName)
}

func (r *ElasticsearchEphemeralResource[T, S]) Configure(_ context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	factory, diags := configureFactoryFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.SetClient(factory)
}

func (r *ElasticsearchEphemeralResource[T, S]) Schema(ctx context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	schema := r.schemaFactory(ctx)
	blocks := make(map[string]eschema.Block, len(schema.Blocks)+1)
	maps.Copy(blocks, schema.Blocks)
	blocks["elasticsearch_connection"] = providerschema.GetEsEphemeralConnectionBlock()
	schema.Blocks = blocks
	resp.Schema = schema
}

func (r *ElasticsearchEphemeralResource[T, S]) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	r.openWithPrivate(ctx, req, resp, resp.Private)
}

func (r *ElasticsearchEphemeralResource[T, S]) openWithPrivate(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse, private ephemeralPrivateState) {
	var model T
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, connDiags := r.Client().GetElasticsearchClient(ctx, model.GetElasticsearchConnection())
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

func (r *ElasticsearchEphemeralResource[T, S]) Close(ctx context.Context, req ephemeral.CloseRequest, resp *ephemeral.CloseResponse) {
	if req.Private == nil {
		return
	}
	r.closeFromPrivate(ctx, req.Private, resp)
}

func (r *ElasticsearchEphemeralResource[T, S]) closeFromPrivate(ctx context.Context, private ephemeralPrivateState, resp *ephemeral.CloseResponse) {
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

	connection, connDiags := decodeElasticsearchConnection(ctx, connectionData)
	resp.Diagnostics.Append(connDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, clientDiags := r.Client().GetElasticsearchClient(ctx, connection)
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

func (r *ElasticsearchEphemeralResource[T, S]) persistOpenPrivateState(ctx context.Context, private ephemeralPrivateState, result OpenResult[T, S]) diag.Diagnostics {
	var diags diag.Diagnostics
	if private == nil {
		diags.AddError(
			"Elasticsearch ephemeral envelope internal error",
			"Open response Private must not be nil",
		)
		return diags
	}

	connectionData, connDiags := encodeElasticsearchConnection(ctx, result.Model.GetElasticsearchConnection())
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

type ephemeralPrivateState interface {
	GetKey(ctx context.Context, key string) ([]byte, diag.Diagnostics)
	SetKey(ctx context.Context, key string, value []byte) diag.Diagnostics
}

var (
	_ ephemeral.EphemeralResource              = (*ElasticsearchEphemeralResource[ElasticsearchConnectionField, struct{}])(nil)
	_ ephemeral.EphemeralResourceWithConfigure = (*ElasticsearchEphemeralResource[ElasticsearchConnectionField, struct{}])(nil)
	_ ephemeral.EphemeralResourceWithClose     = (*ElasticsearchEphemeralResource[ElasticsearchConnectionField, struct{}])(nil)
)
