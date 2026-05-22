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

// KibanaEphemeralModel is the type constraint for models passed to
// [NewKibanaEphemeralResource]. Concrete types must provide GetKibanaConnection,
// typically by embedding [KibanaConnectionField].
type KibanaEphemeralModel interface {
	GetKibanaConnection() types.List
}

type KibanaEphemeralOpenFunc[T KibanaEphemeralModel, S any] func(
	context.Context,
	*clients.KibanaScopedClient,
	OpenRequest[T],
) (OpenResult[T, S], diag.Diagnostics)

type KibanaEphemeralCloseFunc[S any] func(
	context.Context,
	*clients.KibanaScopedClient,
	CloseRequest[S],
) (CloseResponse, diag.Diagnostics)

// KibanaEphemeralOptions configures [NewKibanaEphemeralResource].
// Schema, Open, and Close must be non-nil or the constructor panics.
type KibanaEphemeralOptions[T KibanaEphemeralModel, S any] struct {
	Schema func(context.Context) eschema.Schema
	Open   KibanaEphemeralOpenFunc[T, S]
	Close  KibanaEphemeralCloseFunc[S]
}

// KibanaEphemeralResource implements [ephemeral.EphemeralResource] and related
// interfaces for Kibana-backed ephemeral resources.
type KibanaEphemeralResource[T KibanaEphemeralModel, S any] struct {
	*EphemeralBase
	schemaFactory func(context.Context) eschema.Schema
	openFunc      KibanaEphemeralOpenFunc[T, S]
	closeFunc     KibanaEphemeralCloseFunc[S]
}

// NewKibanaEphemeralResource returns an [ephemeral.EphemeralResource] that
// owns Metadata, Configure, Schema (with kibana_connection block injection),
// Open, and Close for the Kibana namespace.
func NewKibanaEphemeralResource[T KibanaEphemeralModel, S any](
	name string,
	opts KibanaEphemeralOptions[T, S],
) ephemeral.EphemeralResource {
	if opts.Schema == nil {
		panic("entitycore: KibanaEphemeralOptions.Schema must not be nil")
	}
	if opts.Open == nil {
		panic("entitycore: KibanaEphemeralOptions.Open must not be nil")
	}
	if opts.Close == nil {
		panic("entitycore: KibanaEphemeralOptions.Close must not be nil")
	}
	mustBePlainGoCloseState[S]()

	return &KibanaEphemeralResource[T, S]{
		EphemeralBase: NewEphemeralBase(ComponentKibana, name),
		schemaFactory: opts.Schema,
		openFunc:      opts.Open,
		closeFunc:     opts.Close,
	}
}

func (r *KibanaEphemeralResource[T, S]) Metadata(_ context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = r.EphemeralBase.Metadata(req.ProviderTypeName)
}

func (r *KibanaEphemeralResource[T, S]) Configure(_ context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	factory, diags := configureFactoryFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.SetClient(factory)
}

func (r *KibanaEphemeralResource[T, S]) Schema(ctx context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	schema := r.schemaFactory(ctx)
	blocks := make(map[string]eschema.Block, len(schema.Blocks)+1)
	maps.Copy(blocks, schema.Blocks)
	blocks["kibana_connection"] = providerschema.GetKbEphemeralConnectionBlock()
	schema.Blocks = blocks
	resp.Schema = schema
}

func (r *KibanaEphemeralResource[T, S]) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	r.openWithPrivate(ctx, req, resp, resp.Private)
}

func (r *KibanaEphemeralResource[T, S]) openWithPrivate(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse, private ephemeralPrivateState) {
	var model T
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, connDiags := r.Client().GetKibanaClient(ctx, model.GetKibanaConnection())
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

func (r *KibanaEphemeralResource[T, S]) Close(ctx context.Context, req ephemeral.CloseRequest, resp *ephemeral.CloseResponse) {
	if req.Private == nil {
		return
	}
	r.closeFromPrivate(ctx, req.Private, resp)
}

func (r *KibanaEphemeralResource[T, S]) closeFromPrivate(ctx context.Context, private ephemeralPrivateState, resp *ephemeral.CloseResponse) {
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

	connection, connDiags := decodeKibanaConnection(ctx, connectionData)
	resp.Diagnostics.Append(connDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, clientDiags := r.Client().GetKibanaClient(ctx, connection)
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

func (r *KibanaEphemeralResource[T, S]) persistOpenPrivateState(ctx context.Context, private ephemeralPrivateState, result OpenResult[T, S]) diag.Diagnostics {
	var diags diag.Diagnostics
	if private == nil {
		diags.AddError(
			"Kibana ephemeral envelope internal error",
			"Open response Private must not be nil",
		)
		return diags
	}

	connectionData, connDiags := encodeKibanaConnection(ctx, result.Model.GetKibanaConnection())
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

var (
	_ ephemeral.EphemeralResource              = (*KibanaEphemeralResource[KibanaConnectionField, struct{}])(nil)
	_ ephemeral.EphemeralResourceWithConfigure = (*KibanaEphemeralResource[KibanaConnectionField, struct{}])(nil)
	_ ephemeral.EphemeralResourceWithClose     = (*KibanaEphemeralResource[KibanaConnectionField, struct{}])(nil)
)
