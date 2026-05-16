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
	"maps"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ElasticsearchResourceModel is the type constraint for models passed to
// [NewElasticsearchResource]. Concrete types must provide value-receiver
// methods GetID, GetResourceID, and GetElasticsearchConnection.
type ElasticsearchResourceModel interface {
	GetID() types.String
	// GetResourceID returns the plan-safe write identity (for example name or
	// username). Create and Update use this instead of GetID because computed
	// id values may be unknown in create plans.
	GetResourceID() types.String
	GetElasticsearchConnection() types.List
}

// WithReadResourceID is an optional interface for models that need a stable read
// identity distinct from the composite state ID segment used as the default.
type WithReadResourceID interface {
	GetReadResourceID() string
}

type elasticsearchReadFunc[T ElasticsearchResourceModel] func(
	context.Context,
	*clients.ElasticsearchScopedClient,
	string,
	T,
) (T, bool, diag.Diagnostics)

type elasticsearchDeleteFunc[T ElasticsearchResourceModel] func(
	context.Context,
	*clients.ElasticsearchScopedClient,
	string,
	T,
) diag.Diagnostics

// WriteRequest is passed to [WriteFunc] after plan decoding, prior-state
// decoding (Update only), write identity validation, client resolution, and
// optional version checks. Prior is non-nil only for Update; Create receives
// Prior == nil. The same WriteRequest type is shared by Create and Update so a
// single function can serve both when the logic does not differ.
type WriteRequest[T ElasticsearchResourceModel] struct {
	Plan    T
	Prior   *T
	Config  tfsdk.Config
	WriteID string
}

// WriteResult is returned by write callbacks; the envelope read-after-write
// flow uses Model when resolving refresh identity and calling readFunc.
type WriteResult[T ElasticsearchResourceModel] struct {
	Model T
}

// WriteFunc performs Create or Update after the envelope decodes the plan
// (and prior state for Update), validates the write identity, resolves the
// scoped Elasticsearch client, and evaluates optional version requirements.
// Inspect req.Prior == nil to detect Create when sharing a single function for
// both Create and Update.
type WriteFunc[T ElasticsearchResourceModel] func(
	context.Context,
	*clients.ElasticsearchScopedClient,
	WriteRequest[T],
) (WriteResult[T], diag.Diagnostics)

// PostReadFunc runs after a successful read that persisted state, including
// read-after-write refresh. It is optional. The privateState argument is the
// framework response Private field (typically *internal/privatestate.ProviderData).
type PostReadFunc[T ElasticsearchResourceModel] func(
	ctx context.Context,
	client *clients.ElasticsearchScopedClient,
	model T,
	privateState any,
) diag.Diagnostics

// ElasticsearchResourceOptions configures [NewElasticsearchResource]. PostRead
// is optional; Schema, Read, Delete, Create, and Update must be non-nil or the
// envelope surfaces configuration diagnostics instead of invoking nil callbacks.
// Create and Update share the [WriteFunc] type so callers may pass the same
// function for both when the logic is identical.
type ElasticsearchResourceOptions[T ElasticsearchResourceModel] struct {
	Schema   func(context.Context) rschema.Schema
	Read     elasticsearchReadFunc[T]
	Delete   elasticsearchDeleteFunc[T]
	Create   WriteFunc[T]
	Update   WriteFunc[T]
	PostRead PostReadFunc[T]
}

// ElasticsearchResource implements [resource.Resource] and related interfaces
// for Elasticsearch-backed resources. It embeds [*ResourceBase] to reuse
// Configure, Metadata, and Client.
//
// The envelope owns Schema (with elasticsearch_connection block injection),
// Create, Read, Update, and Delete. Concrete resources may override Create or
// Update when their lifecycle does not fit the callback contract, and may
// choose to implement ImportState.
type ElasticsearchResource[T ElasticsearchResourceModel] struct {
	*ResourceBase
	schemaFactory func(context.Context) rschema.Schema
	readFunc      elasticsearchReadFunc[T]
	deleteFunc    elasticsearchDeleteFunc[T]
	createFunc    WriteFunc[T]
	updateFunc    WriteFunc[T]
	postReadFunc  PostReadFunc[T]
}

// PlaceholderElasticsearchWriteCallback returns a write callback that fails if
// invoked. Use it for both Create and Update when a concrete resource type
// still defines its own Create and Update methods that shadow the envelope so
// Terraform never calls the placeholder.
const (
	placeholderWriteCallbackSummary = "Elasticsearch envelope"
	placeholderWriteCallbackDetail  = "Internal error: write callback placeholder was invoked; " +
		"the concrete resource should override Create and Update or pass real callbacks via ElasticsearchResourceOptions."
)

func PlaceholderElasticsearchWriteCallback[T ElasticsearchResourceModel]() WriteFunc[T] {
	return func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ WriteRequest[T]) (WriteResult[T], diag.Diagnostics) {
		var diags diag.Diagnostics
		diags.AddError(
			placeholderWriteCallbackSummary,
			placeholderWriteCallbackDetail,
		)
		var zero T
		return WriteResult[T]{Model: zero}, diags
	}
}

// NewElasticsearchResource returns an [*ElasticsearchResource] that owns
// Schema, Create, Read, Update, and Delete for the Elasticsearch namespace.
// Concrete resources supply callbacks in opts; Schema, Read, Delete, Create,
// and Update must be non-nil or the envelope surfaces configuration error
// diagnostics instead of invoking nil callbacks.
func NewElasticsearchResource[T ElasticsearchResourceModel](name string, opts ElasticsearchResourceOptions[T]) *ElasticsearchResource[T] {
	return &ElasticsearchResource[T]{
		ResourceBase:  NewResourceBase(ComponentElasticsearch, name),
		schemaFactory: opts.Schema,
		readFunc:      opts.Read,
		deleteFunc:    opts.Delete,
		createFunc:    opts.Create,
		updateFunc:    opts.Update,
		postReadFunc:  opts.PostRead,
	}
}

func resolveElasticsearchReadResourceID(model ElasticsearchResourceModel, writeFallback string) (string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if m, ok := any(model).(WithReadResourceID); ok {
		if id := strings.TrimSpace(m.GetReadResourceID()); id != "" {
			return id, diags
		}
	}
	if writeFallback != "" {
		return writeFallback, diags
	}
	compID, compDiags := clients.CompositeIDFromStrFw(model.GetID().ValueString())
	diags.Append(compDiags...)
	if diags.HasError() {
		return "", diags
	}
	return compID.ResourceID, diags
}

// Schema implements [resource.Resource], injecting the elasticsearch_connection
// block into the schema returned by the concrete schema factory.
func (r *ElasticsearchResource[T]) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	schema := r.schemaFactory(ctx)
	blocks := make(map[string]rschema.Block, len(schema.Blocks)+1)
	maps.Copy(blocks, schema.Blocks)
	blocks["elasticsearch_connection"] = providerschema.GetEsFWConnectionBlock()
	schema.Blocks = blocks
	resp.Schema = schema
}

// Create implements [resource.Resource]: decode plan, resolve client, invoke
// the create callback, read-after-write, then persist state from readFunc.
func (r *ElasticsearchResource[T]) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(r.runWrite(ctx, writeInvocation[T]{
		plan:         req.Plan,
		config:       req.Config,
		outState:     &resp.State,
		privateState: resp.Private,
		isUpdate:     false,
	})...)
}

// Read implements [resource.Resource] with the standard prelude: deserialize
// prior state into the generic model T, resolve read identity from the model
// and/or composite ID, resolve the scoped Elasticsearch client, enforce optional
// version requirements when the model reports requirement diagnostics with an
// error severity (matching Kibana envelope semantics), then delegate to the
// concrete readFunc.
func (r *ElasticsearchResource[T]) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model T
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceID, idDiags := resolveElasticsearchReadResourceID(model, "")
	resp.Diagnostics.Append(idDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if resourceID == "" {
		resp.Diagnostics.AddError(
			"Invalid resource identifier",
			"The resolved read identity is empty; cannot read.",
		)
		return
	}

	client, diags := r.Client().GetElasticsearchClient(ctx, model.GetElasticsearchConnection())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if vDiags := enforceVersionRequirements(ctx, client, &model); vDiags.HasError() {
		resp.Diagnostics.Append(vDiags...)
		return
	}

	if d := r.requireReadFunc(); d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}

	resultModel, found, callDiags := r.readFunc(ctx, client, resourceID, model)
	resp.Diagnostics.Append(callDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if found {
		resp.Diagnostics.Append(resp.State.Set(ctx, &resultModel)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if r.postReadFunc != nil {
			resp.Diagnostics.Append(r.postReadFunc(ctx, client, resultModel, resp.Private)...)
		}
	} else {
		resp.State.RemoveResource(ctx)
	}
}

// Update implements [resource.Resource] with the same prelude as Create,
// additionally decoding prior state for the update callback.
func (r *ElasticsearchResource[T]) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(r.runWrite(ctx, writeInvocation[T]{
		plan:         req.Plan,
		priorState:   &req.State,
		config:       req.Config,
		outState:     &resp.State,
		privateState: resp.Private,
		isUpdate:     true,
	})...)
}

type writeInvocation[T ElasticsearchResourceModel] struct {
	plan         tfsdk.Plan
	priorState   *tfsdk.State
	config       tfsdk.Config
	outState     *tfsdk.State
	privateState any
	isUpdate     bool
}

func (r *ElasticsearchResource[T]) requireReadFunc() diag.Diagnostics {
	var diags diag.Diagnostics
	if r.readFunc == nil {
		diags.AddError(
			"Elasticsearch envelope configuration error",
			"The read callback passed via ElasticsearchResourceOptions must not be nil.",
		)
	}
	return diags
}

func (r *ElasticsearchResource[T]) runWrite(ctx context.Context, inv writeInvocation[T]) diag.Diagnostics {
	var diags diag.Diagnostics
	if (inv.isUpdate && r.updateFunc == nil) || (!inv.isUpdate && r.createFunc == nil) {
		op := "create"
		if inv.isUpdate {
			op = "update"
		}
		diags.AddError(
			"Elasticsearch envelope configuration error",
			fmt.Sprintf("The %s callback passed via ElasticsearchResourceOptions must not be nil.", op),
		)
		return diags
	}

	var planModel T
	diags.Append(inv.plan.Get(ctx, &planModel)...)
	if diags.HasError() {
		return diags
	}

	var priorPtr *T
	if inv.isUpdate && inv.priorState != nil {
		var priorModel T
		diags.Append(inv.priorState.Get(ctx, &priorModel)...)
		if diags.HasError() {
			return diags
		}
		priorPtr = &priorModel
	}

	writeID := planModel.GetResourceID()
	if !typeutils.IsKnown(writeID) || writeID.ValueString() == "" {
		diags.AddError(
			"Invalid resource identifier",
			"The resource write identity from configuration is unknown or empty; cannot create or update.",
		)
		return diags
	}

	client, connDiags := r.Client().GetElasticsearchClient(ctx, planModel.GetElasticsearchConnection())
	diags.Append(connDiags...)
	if diags.HasError() {
		return diags
	}

	if vDiags := enforceVersionRequirements(ctx, client, &planModel); vDiags.HasError() {
		diags.Append(vDiags...)
		return diags
	}

	if d := r.requireReadFunc(); d.HasError() {
		return d
	}

	writeFn := r.createFunc
	if inv.isUpdate {
		writeFn = r.updateFunc
	}
	writeKey := writeID.ValueString()
	written, callDiags := writeFn(ctx, client, WriteRequest[T]{
		Plan:    planModel,
		Prior:   priorPtr,
		Config:  inv.config,
		WriteID: writeKey,
	})
	diags.Append(callDiags...)
	if diags.HasError() {
		return diags
	}

	readResourceID, idDiags := resolveElasticsearchReadResourceID(written.Model, writeKey)
	diags.Append(idDiags...)
	if diags.HasError() {
		return diags
	}
	if readResourceID == "" {
		diags.AddError(
			"Invalid resource identifier",
			"The resolved read identity is empty after write; cannot refresh.",
		)
		return diags
	}

	stateModel, found, readDiags := r.readFunc(ctx, client, readResourceID, written.Model)
	diags.Append(readDiags...)
	if diags.HasError() {
		return diags
	}

	if !found {
		diags.AddError(
			"Resource not found",
			fmt.Sprintf("%s_%s %q was not found after write", r.component, r.resourceName, writeKey),
		)
		return diags
	}

	diags.Append(inv.outState.Set(ctx, &stateModel)...)
	if diags.HasError() {
		return diags
	}

	if r.postReadFunc != nil {
		diags.Append(r.postReadFunc(ctx, client, stateModel, inv.privateState)...)
	}

	return diags
}

// Delete implements [resource.Resource] with the standard prelude, then
// delegates to the concrete deleteFunc.
func (r *ElasticsearchResource[T]) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.deleteFunc == nil {
		resp.Diagnostics.AddError(
			"Elasticsearch envelope configuration error",
			"The delete callback passed via ElasticsearchResourceOptions must not be nil.",
		)
		return
	}

	var model T
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	compID, diags := clients.CompositeIDFromStrFw(model.GetID().ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, diags := r.Client().GetElasticsearchClient(ctx, model.GetElasticsearchConnection())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.deleteFunc(ctx, client, compID.ResourceID, model)...)
}

var (
	_ resource.Resource              = (*ElasticsearchResource[ElasticsearchResourceModel])(nil)
	_ resource.ResourceWithConfigure = (*ElasticsearchResource[ElasticsearchResourceModel])(nil)
)
