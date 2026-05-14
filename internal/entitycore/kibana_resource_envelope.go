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
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// KibanaResourceModel is the type constraint for models passed to
// [NewKibanaResource]. Concrete types must provide value-receiver methods
// GetID, GetResourceID, GetSpaceID, and GetKibanaConnection.
type KibanaResourceModel interface {
	GetID() types.String
	// GetResourceID returns the plan-safe write identity (for example name or
	// API-assigned UUID). Read, Update, and Delete use this when the state ID
	// is not a composite.
	GetResourceID() types.String
	GetSpaceID() types.String
	GetKibanaConnection() types.List
}

type kibanaReadFunc[T KibanaResourceModel] func(
	context.Context,
	*clients.KibanaScopedClient,
	string,
	string,
	T,
) (T, bool, diag.Diagnostics)

type kibanaDeleteFunc[T KibanaResourceModel] func(
	context.Context,
	*clients.KibanaScopedClient,
	string,
	string,
	T,
) diag.Diagnostics

// KibanaCreateFunc performs the create after the envelope decodes the plan,
// validates the space ID, resolves the scoped Kibana client, and passes the
// planned model. It returns the model to persist in state.
type KibanaCreateFunc[T KibanaResourceModel] func(
	context.Context,
	*clients.KibanaScopedClient,
	string,
	T,
) (T, diag.Diagnostics)

// KibanaUpdateFunc performs the update after the envelope decodes the plan
// and prior state, resolves the resource identity, resolves the scoped Kibana
// client, and passes both the plan and prior models. It returns the model to
// persist in state.
type KibanaUpdateFunc[T KibanaResourceModel] func(
	context.Context,
	*clients.KibanaScopedClient,
	string,
	string,
	T,
	T,
) (T, diag.Diagnostics)

// KibanaResource implements [resource.Resource] and related interfaces
// for Kibana-backed resources. It embeds [*ResourceBase] to reuse
// Configure, Metadata, and Client.
//
// The envelope owns Schema (with kibana_connection block injection),
// Create, Read, Update, and Delete. Concrete resources may override Create or
// Update when their lifecycle does not fit the callback contract, and may
// choose to implement ImportState.
type KibanaResource[T KibanaResourceModel] struct {
	*ResourceBase
	schemaFactory func(context.Context) rschema.Schema
	readFunc      kibanaReadFunc[T]
	deleteFunc    kibanaDeleteFunc[T]
	createFunc    KibanaCreateFunc[T]
	updateFunc    KibanaUpdateFunc[T]
}

const (
	placeholderKibanaWriteCallbackSummary = "Kibana envelope"
	placeholderKibanaWriteCallbackDetail  = "Internal error: write callback placeholder was invoked; " +
		"the concrete resource should override Create and Update or pass real callbacks to NewKibanaResource."
)

// PlaceholderKibanaWriteCallbacks returns create and update callbacks
// that fail if invoked. Use when a concrete resource type still defines its own
// Create and Update methods that override the envelope so Terraform never calls
// these placeholders.
func PlaceholderKibanaWriteCallbacks[T KibanaResourceModel]() (KibanaCreateFunc[T], KibanaUpdateFunc[T]) {
	createFn := func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ T) (T, diag.Diagnostics) {
		var diags diag.Diagnostics
		diags.AddError(
			placeholderKibanaWriteCallbackSummary,
			placeholderKibanaWriteCallbackDetail,
		)
		var zero T
		return zero, diags
	}
	updateFn := func(_ context.Context, _ *clients.KibanaScopedClient, _ string, _ string, _ T, _ T) (T, diag.Diagnostics) {
		var diags diag.Diagnostics
		diags.AddError(
			placeholderKibanaWriteCallbackSummary,
			placeholderKibanaWriteCallbackDetail,
		)
		var zero T
		return zero, diags
	}
	return createFn, updateFn
}

// NewKibanaResource returns an [*KibanaResource] that owns
// Schema, Create, Read, Update, and Delete. Concrete resources supply a schema
// factory (without kibana_connection block), read, delete, create, and
// update callbacks. All callbacks must be non-nil; otherwise Create or Update
// surface a configuration error diagnostic instead of invoking the callback.
func NewKibanaResource[T KibanaResourceModel](
	component Component,
	name string,
	schemaFactory func(context.Context) rschema.Schema,
	readFunc kibanaReadFunc[T],
	deleteFunc kibanaDeleteFunc[T],
	createFunc KibanaCreateFunc[T],
	updateFunc KibanaUpdateFunc[T],
) *KibanaResource[T] {
	return &KibanaResource[T]{
		ResourceBase:  NewResourceBase(component, name),
		schemaFactory: schemaFactory,
		readFunc:      readFunc,
		deleteFunc:    deleteFunc,
		createFunc:    createFunc,
		updateFunc:    updateFunc,
	}
}

// Schema implements [resource.Resource], injecting the kibana_connection
// block into the schema returned by the concrete schema factory.
func (r *KibanaResource[T]) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	schema := r.schemaFactory(ctx)
	blocks := make(map[string]rschema.Block, len(schema.Blocks)+1)
	maps.Copy(blocks, schema.Blocks)
	blocks["kibana_connection"] = providerschema.GetKbFWConnectionBlock()
	schema.Blocks = blocks
	resp.Schema = schema
}

// resolveResourceIdentity uses the composite-ID-or-fallback rule to determine
// the resourceID and spaceID for a model. It attempts to parse GetID() as a
// composite ID; on failure (nil result) it falls back to GetResourceID() and
// GetSpaceID(). Any diagnostics from the composite parse are discarded.
func (r *KibanaResource[T]) resolveResourceIdentity(model T) (resourceID string, spaceID string) {
	compID, _ := clients.CompositeIDFromStrFw(model.GetID().ValueString())
	if compID != nil {
		return compID.ResourceID, compID.ClusterID
	}
	return model.GetResourceID().ValueString(), model.GetSpaceID().ValueString()
}

// Create implements [resource.Resource]: decode plan, validate spaceID,
// resolve client, invoke the create callback, then persist the returned model.
func (r *KibanaResource[T]) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.createFunc == nil {
		resp.Diagnostics.AddError(
			"Kibana envelope configuration error",
			"The create callback passed to NewKibanaResource must not be nil.",
		)
		return
	}

	var plan T
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	spaceID := plan.GetSpaceID()
	if !typeutils.IsKnown(spaceID) || spaceID.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Invalid space identifier",
			"The space identifier from configuration is unknown or empty; cannot create.",
		)
		return
	}

	client, diags := r.Client().GetKibanaClient(ctx, plan.GetKibanaConnection())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if vDiags := enforceVersionRequirements(ctx, client, &plan); vDiags.HasError() {
		resp.Diagnostics.Append(vDiags...)
		return
	}

	resultModel, callDiags := r.createFunc(ctx, client, spaceID.ValueString(), plan)
	resp.Diagnostics.Append(callDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &resultModel)...)
}

// Read implements [resource.Resource] with the standard prelude: decode state,
// resolve identity via composite-ID-or-fallback, validate resourceID, resolve
// scoped Kibana client, then delegate to the concrete readFunc. When readFunc
// reports found==true, the returned model is persisted via resp.State.Set;
// when found==false, the resource is removed from state.
func (r *KibanaResource[T]) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.readFunc == nil {
		resp.Diagnostics.AddError(
			"Kibana envelope configuration error",
			"The read callback passed to NewKibanaResource must not be nil.",
		)
		return
	}

	var model T
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceID, spaceID := r.resolveResourceIdentity(model)
	if resourceID == "" {
		resp.Diagnostics.AddError(
			"Invalid resource identifier",
			"The resource identifier is empty; cannot read.",
		)
		return
	}

	client, diags := r.Client().GetKibanaClient(ctx, model.GetKibanaConnection())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if vDiags := enforceVersionRequirements(ctx, client, &model); vDiags.HasError() {
		resp.Diagnostics.Append(vDiags...)
		return
	}

	resultModel, found, callDiags := r.readFunc(ctx, client, resourceID, spaceID, model)
	resp.Diagnostics.Append(callDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if found {
		resp.Diagnostics.Append(resp.State.Set(ctx, &resultModel)...)
	} else {
		resp.State.RemoveResource(ctx)
	}
}

// Update implements [resource.Resource]: decode plan and prior state, resolve
// identity on the plan model, validate resourceID, resolve client, invoke the
// update callback, then persist the returned model.
func (r *KibanaResource[T]) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.updateFunc == nil {
		resp.Diagnostics.AddError(
			"Kibana envelope configuration error",
			"The update callback passed to NewKibanaResource must not be nil.",
		)
		return
	}

	var plan T
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var prior T
	resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceID, spaceID := r.resolveResourceIdentity(plan)
	if resourceID == "" {
		resp.Diagnostics.AddError(
			"Invalid resource identifier",
			"The resource identifier is empty; cannot update.",
		)
		return
	}

	client, diags := r.Client().GetKibanaClient(ctx, plan.GetKibanaConnection())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if vDiags := enforceVersionRequirements(ctx, client, &plan); vDiags.HasError() {
		resp.Diagnostics.Append(vDiags...)
		return
	}

	resultModel, callDiags := r.updateFunc(ctx, client, resourceID, spaceID, plan, prior)
	resp.Diagnostics.Append(callDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &resultModel)...)
}

// Delete implements [resource.Resource] with the standard prelude, then
// delegates to the concrete deleteFunc.
func (r *KibanaResource[T]) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.deleteFunc == nil {
		resp.Diagnostics.AddError(
			"Kibana envelope configuration error",
			"The delete callback passed to NewKibanaResource must not be nil.",
		)
		return
	}

	var model T
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceID, spaceID := r.resolveResourceIdentity(model)
	if resourceID == "" {
		resp.Diagnostics.AddError(
			"Invalid resource identifier",
			"The resource identifier is empty; cannot delete.",
		)
		return
	}

	client, diags := r.Client().GetKibanaClient(ctx, model.GetKibanaConnection())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.deleteFunc(ctx, client, resourceID, spaceID, model)...)
}

var (
	_ resource.Resource              = (*KibanaResource[KibanaResourceModel])(nil)
	_ resource.ResourceWithConfigure = (*KibanaResource[KibanaResourceModel])(nil)
)
