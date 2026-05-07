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

// ElasticsearchCreateFunc performs the create after the envelope decodes the
// plan, checks the write identity, resolves the scoped Elasticsearch client,
// and passes the planned model. The callback should call the remote create
// API, set the composite ID on the returned model when readFunc expects to
// carry it through (e.g. via client.ID()), include any create-only field
// values, and return the model. The envelope invokes readFunc after a
// successful callback and sets state from the read result; the callback must
// not call readFunc.
type ElasticsearchCreateFunc[T ElasticsearchResourceModel] func(
	context.Context,
	*clients.ElasticsearchScopedClient,
	string,
	T,
) (T, diag.Diagnostics)

// ElasticsearchUpdateFunc performs the update with the same prelude as
// [ElasticsearchCreateFunc]. The callback should call the remote update API,
// set the composite ID on the returned model when readFunc expects to carry
// it through, and return it. The envelope invokes readFunc after a successful
// callback and sets state from the read result; the callback must not call
// readFunc.
type ElasticsearchUpdateFunc[T ElasticsearchResourceModel] func(
	context.Context,
	*clients.ElasticsearchScopedClient,
	string,
	T,
) (T, diag.Diagnostics)

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
	createFunc    ElasticsearchCreateFunc[T]
	updateFunc    ElasticsearchUpdateFunc[T]
}

// PlaceholderElasticsearchWriteCallbacks returns create and update callbacks
// that fail if invoked. Use when a concrete resource type still defines its own
// Create and Update methods that override the envelope so Terraform never calls
// these placeholders.
const (
	placeholderWriteCallbackSummary = "Elasticsearch envelope"
	placeholderWriteCallbackDetail  = "Internal error: write callback placeholder was invoked; " +
		"the concrete resource should override Create and Update or pass real callbacks to NewElasticsearchResource."
)

func PlaceholderElasticsearchWriteCallbacks[T ElasticsearchResourceModel]() (ElasticsearchCreateFunc[T], ElasticsearchUpdateFunc[T]) {
	fn := func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, _ T) (T, diag.Diagnostics) {
		var diags diag.Diagnostics
		diags.AddError(
			placeholderWriteCallbackSummary,
			placeholderWriteCallbackDetail,
		)
		var zero T
		return zero, diags
	}
	return fn, fn
}

// NewElasticsearchResource returns an [*ElasticsearchResource] that owns
// Schema, Create, Read, Update, and Delete. Concrete resources supply a schema
// factory (without elasticsearch_connection block), read, delete, create, and
// update callbacks. All callbacks must be non-nil; otherwise Create or Update
// surface a configuration error diagnostic instead of invoking the callback.
func NewElasticsearchResource[T ElasticsearchResourceModel](
	component Component,
	name string,
	schemaFactory func(context.Context) rschema.Schema,
	readFunc elasticsearchReadFunc[T],
	deleteFunc elasticsearchDeleteFunc[T],
	createFunc ElasticsearchCreateFunc[T],
	updateFunc ElasticsearchUpdateFunc[T],
) *ElasticsearchResource[T] {
	return &ElasticsearchResource[T]{
		ResourceBase:  NewResourceBase(component, name),
		schemaFactory: schemaFactory,
		readFunc:      readFunc,
		deleteFunc:    deleteFunc,
		createFunc:    createFunc,
		updateFunc:    updateFunc,
	}
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
// the create callback, then persist the returned model.
func (r *ElasticsearchResource[T]) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(r.writeFromPlan(ctx, req.Plan, &resp.State, r.createFunc)...)
}

// Read implements [resource.Resource] with the standard prelude: decode state,
// parse composite ID, resolve scoped Elasticsearch client, then delegate to the
// concrete readFunc. When readFunc reports found==true, the returned model is
// persisted via resp.State.Set; when found==false, the resource is removed
// from state.
func (r *ElasticsearchResource[T]) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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

	resultModel, found, diags := r.readFunc(ctx, client, compID.ResourceID, model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if found {
		resp.Diagnostics.Append(resp.State.Set(ctx, &resultModel)...)
	} else {
		resp.State.RemoveResource(ctx)
	}
}

// Update implements [resource.Resource] with the same prelude as Create.
func (r *ElasticsearchResource[T]) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(r.writeFromPlan(ctx, req.Plan, &resp.State, r.updateFunc)...)
}

func (r *ElasticsearchResource[T]) writeFromPlan(
	ctx context.Context,
	plan tfsdk.Plan,
	state *tfsdk.State,
	op func(context.Context, *clients.ElasticsearchScopedClient, string, T) (T, diag.Diagnostics),
) diag.Diagnostics {
	var model T
	var diags diag.Diagnostics
	if op == nil {
		diags.AddError(
			"Elasticsearch envelope configuration error",
			"The create or update callback passed to NewElasticsearchResource must not be nil.",
		)
		return diags
	}

	diags.Append(plan.Get(ctx, &model)...)
	if diags.HasError() {
		return diags
	}

	writeID := model.GetResourceID()
	if !typeutils.IsKnown(writeID) || writeID.ValueString() == "" {
		diags.AddError(
			"Invalid resource identifier",
			"The resource write identity from configuration is unknown or empty; cannot create or update.",
		)
		return diags
	}

	client, connDiags := r.Client().GetElasticsearchClient(ctx, model.GetElasticsearchConnection())
	diags.Append(connDiags...)
	if diags.HasError() {
		return diags
	}

	if r.readFunc == nil {
		diags.AddError(
			"Elasticsearch envelope configuration error",
			"The read callback passed to NewElasticsearchResource must not be nil.",
		)
		return diags
	}

	writtenModel, callDiags := op(ctx, client, writeID.ValueString(), model)
	diags.Append(callDiags...)
	if diags.HasError() {
		return diags
	}

	stateModel, found, readDiags := r.readFunc(ctx, client, writeID.ValueString(), writtenModel)
	diags.Append(readDiags...)
	if diags.HasError() {
		return diags
	}

	if !found {
		diags.AddError(
			"Resource not found",
			fmt.Sprintf("%s_%s %q was not found after write", r.component, r.resourceName, writeID.ValueString()),
		)
		return diags
	}

	diags.Append(state.Set(ctx, &stateModel)...)
	return diags
}

// Delete implements [resource.Resource] with the standard prelude, then
// delegates to the concrete deleteFunc.
func (r *ElasticsearchResource[T]) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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
