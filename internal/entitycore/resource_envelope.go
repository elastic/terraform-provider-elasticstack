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
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ElasticsearchResourceModel is the type constraint for models passed to
// [NewElasticsearchResource]. Concrete types must provide value-receiver
// methods GetID and GetElasticsearchConnection.
type ElasticsearchResourceModel interface {
	GetID() types.String
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

// ElasticsearchResource implements [resource.Resource] and related interfaces
// for Elasticsearch-backed resources. It embeds [*ResourceBase] to reuse
// Configure, Metadata, and Client.
//
// The envelope owns Schema (with elasticsearch_connection block injection),
// Read, and Delete. Concrete resources that embed *ElasticsearchResource[T]
// must implement Create and Update, and may choose to implement ImportState.
type ElasticsearchResource[T ElasticsearchResourceModel] struct {
	*ResourceBase
	schemaFactory func() rschema.Schema
	readFunc      elasticsearchReadFunc[T]
	deleteFunc    elasticsearchDeleteFunc[T]
}

// NewElasticsearchResource returns an [*ElasticsearchResource] that owns
// Schema, Read, and Delete. Concrete resources supply a schema factory
// (without elasticsearch_connection block), a read callback, and a delete
// callback.
func NewElasticsearchResource[T ElasticsearchResourceModel](
	component Component,
	name string,
	schemaFactory func() rschema.Schema,
	readFunc elasticsearchReadFunc[T],
	deleteFunc elasticsearchDeleteFunc[T],
) *ElasticsearchResource[T] {
	return &ElasticsearchResource[T]{
		ResourceBase:  NewResourceBase(component, name),
		schemaFactory: schemaFactory,
		readFunc:      readFunc,
		deleteFunc:    deleteFunc,
	}
}

// Schema implements [resource.Resource], injecting the elasticsearch_connection
// block into the schema returned by the concrete schema factory.
func (r *ElasticsearchResource[T]) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	schema := r.schemaFactory()
	blocks := make(map[string]rschema.Block, len(schema.Blocks)+1)
	maps.Copy(blocks, schema.Blocks)
	blocks["elasticsearch_connection"] = providerschema.GetEsFWConnectionBlock()
	schema.Blocks = blocks
	resp.Schema = schema
}

// Create provides a defensive default for the envelope. Concrete resources that
// embed *ElasticsearchResource[T] are expected to override this method with
// their own create logic.
func (r *ElasticsearchResource[T]) Create(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError(
		"Create not implemented",
		"ElasticsearchResource only provides shared Schema, Read, and Delete behavior. Concrete resources embedding ElasticsearchResource must implement Create.",
	)
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

// Update provides a defensive default for the envelope. Concrete resources that
// embed *ElasticsearchResource[T] are expected to override this method with
// their own update logic.
func (r *ElasticsearchResource[T]) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not implemented",
		"ElasticsearchResource only provides shared Schema, Read, and Delete behavior. Concrete resources embedding ElasticsearchResource must implement Update.",
	)
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
