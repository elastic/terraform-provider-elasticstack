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
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// KibanaResourceModel is the type constraint for models passed to
// [NewKibanaResource]. Concrete types must provide value-receiver methods
// GetID, GetResourceID, GetSpaceID, GetKibanaConnection, and
// [WithResourceTimeouts] (typically by embedding [ResourceTimeoutsField]).
type KibanaResourceModel interface {
	GetID() types.String
	// GetResourceID returns the plan-safe write identity (for example name or
	// API-assigned UUID). Read, Update, and Delete use this when the state ID
	// is not a composite.
	GetResourceID() types.String
	GetSpaceID() types.String
	GetKibanaConnection() types.List
	WithResourceTimeouts
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

// KibanaWriteRequest is passed to [KibanaWriteFunc]. Config is the Terraform
// configuration decoded into T by the envelope before the callback is invoked.
// Prior is non-nil only for Update; Create receives Prior == nil.
type KibanaWriteRequest[T KibanaResourceModel] struct {
	Plan    T
	Prior   *T
	Config  T
	WriteID string
	SpaceID string
}

// KibanaWriteResult is returned by write callbacks; the envelope read-after-write
// flow uses Model when resolving refresh identity and calling readFunc.
type KibanaWriteResult[T KibanaResourceModel] struct {
	Model T
}

// KibanaWriteFunc performs Create or Update after the envelope decodes the plan
// (and prior state for Update), validates spaceID, resolves the scoped Kibana
// client, and evaluates optional version requirements. Inspect req.Prior == nil
// to detect Create when sharing a single function for both Create and Update.
type KibanaWriteFunc[T KibanaResourceModel] func(
	context.Context,
	*clients.KibanaScopedClient,
	KibanaWriteRequest[T],
) (KibanaWriteResult[T], diag.Diagnostics)

// KibanaPostReadRequest is passed to [KibanaPostReadFunc]. Prior is the model
// before the read (plan on write path, prior state on plain Read path). State is
// the freshly-read model returned by the read callback.
type KibanaPostReadRequest[T KibanaResourceModel] struct {
	Client  *clients.KibanaScopedClient
	Prior   T
	State   T
	Private PrivateStateStorage
}

// KibanaPostReadFunc runs after a successful read and before state is persisted,
// including read-after-write refresh. It is optional.
type KibanaPostReadFunc[T KibanaResourceModel] func(
	ctx context.Context,
	req KibanaPostReadRequest[T],
) (T, diag.Diagnostics)

// KibanaResourceOptions configures [NewKibanaResource]. PostRead is optional;
// Schema, Read, Delete, Create, and Update must be non-nil or the envelope
// surfaces configuration diagnostics instead of invoking nil callbacks.
//
// Timeouts supplies per-operation default durations when configuration omits
// `timeouts.<op>`; zero fields fall back to [DefaultResourceCreateTimeout],
// [DefaultResourceReadTimeout], [DefaultResourceUpdateTimeout], and
// [DefaultResourceDeleteTimeout]. Concrete schema factories MUST NOT include
// a `timeouts` attribute; the envelope injects it and silently overwrites any
// factory-supplied attribute with the same key.
type KibanaResourceOptions[T KibanaResourceModel] struct {
	Schema   func(context.Context) rschema.Schema
	Read     kibanaReadFunc[T]
	Delete   kibanaDeleteFunc[T]
	Create   KibanaWriteFunc[T]
	Update   KibanaWriteFunc[T]
	PostRead KibanaPostReadFunc[T]
	Timeouts ResourceTimeouts
}

// KibanaResource implements [resource.Resource] and related interfaces
// for Kibana-backed resources. It embeds [*ResourceBase] to reuse
// Configure, Metadata, and Client.
//
// The envelope owns Schema (with kibana_connection block and timeouts attribute
// injection), Create, Read, Update, and Delete. Concrete resources may override
// Create or Update when their lifecycle does not fit the callback contract, and
// may choose to implement ImportState.
type KibanaResource[T KibanaResourceModel] struct {
	*ResourceBase
	schemaFactory func(context.Context) rschema.Schema
	readFunc      kibanaReadFunc[T]
	deleteFunc    kibanaDeleteFunc[T]
	createFunc    KibanaWriteFunc[T]
	updateFunc    KibanaWriteFunc[T]
	postReadFunc  KibanaPostReadFunc[T]
	timeouts      ResourceTimeouts
}

const (
	placeholderKibanaWriteCallbackSummary = "Kibana envelope"
	placeholderKibanaWriteCallbackDetail  = "Internal error: write callback placeholder was invoked; " +
		"the concrete resource should override Create and Update or pass real callbacks via KibanaResourceOptions."
)

// PlaceholderKibanaWriteCallback returns a write callback that fails if invoked.
// Use for both Create and Update when a concrete resource type still defines its own
// Create and Update methods that override the envelope so Terraform never calls
// the placeholder.
func PlaceholderKibanaWriteCallback[T KibanaResourceModel]() KibanaWriteFunc[T] {
	return func(_ context.Context, _ *clients.KibanaScopedClient, _ KibanaWriteRequest[T]) (KibanaWriteResult[T], diag.Diagnostics) {
		var diags diag.Diagnostics
		diags.AddError(
			placeholderKibanaWriteCallbackSummary,
			placeholderKibanaWriteCallbackDetail,
		)
		var zero T
		return KibanaWriteResult[T]{Model: zero}, diags
	}
}

// NewKibanaResource returns an [*KibanaResource] that owns
// Schema, Create, Read, Update, and Delete. Concrete resources supply callbacks
// in opts; Schema, Read, Delete, Create, and Update must be non-nil or the
// envelope surfaces configuration error diagnostics instead of invoking nil callbacks.
func NewKibanaResource[T KibanaResourceModel](
	component Component,
	name string,
	opts KibanaResourceOptions[T],
) *KibanaResource[T] {
	return &KibanaResource[T]{
		ResourceBase:  NewResourceBase(component, name),
		schemaFactory: opts.Schema,
		readFunc:      opts.Read,
		deleteFunc:    opts.Delete,
		createFunc:    opts.Create,
		updateFunc:    opts.Update,
		postReadFunc:  opts.PostRead,
		timeouts:      opts.Timeouts,
	}
}

// Schema implements [resource.Resource], injecting the kibana_connection
// block and the `timeouts` attribute into the schema returned by the concrete
// schema factory. A pre-existing `timeouts` attribute in the factory output is
// silently replaced.
func (r *KibanaResource[T]) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	schema := r.schemaFactory(ctx)
	blocks := make(map[string]rschema.Block, len(schema.Blocks)+1)
	maps.Copy(blocks, schema.Blocks)
	blocks[blockKibanaConnection] = providerschema.GetKbFWConnectionBlock()
	schema.Blocks = blocks

	attrs := make(map[string]rschema.Attribute, len(schema.Attributes)+1)
	maps.Copy(attrs, schema.Attributes)
	attrs[attrTimeouts] = timeouts.AttributesAll(ctx)
	schema.Attributes = attrs

	resp.Schema = schema
}

// resolveKibanaResourceIdentity uses the composite-ID-or-fallback rule to
// determine the resourceID and spaceID for a model. It attempts to parse
// GetID() as a composite ID; on failure (nil result) it falls back to
// GetResourceID() and GetSpaceID(). Composite-parse diagnostics are discarded.
func resolveKibanaResourceIdentity[T KibanaResourceModel](model T) (resourceID string, spaceID string) {
	compID, _ := clients.CompositeIDFromStr(model.GetID().ValueString())
	if compID != nil {
		return compID.ResourceID, compID.ClusterID
	}
	return model.GetResourceID().ValueString(), model.GetSpaceID().ValueString()
}

// isKibanaUnscoped reports whether model opts out of space-identifier
// validation via the [KibanaUnscopedSpace] interface.
func isKibanaUnscoped[T KibanaResourceModel](model T) bool {
	u, ok := any(model).(KibanaUnscopedSpace)
	return ok && u.IsUnscopedSpace()
}

func (r *KibanaResource[T]) validateSpaceID(plan T) diag.Diagnostics {
	var diags diag.Diagnostics
	spaceID := plan.GetSpaceID()
	if !typeutils.IsKnown(spaceID) {
		diags.AddError(
			"Invalid space identifier",
			"The space identifier from configuration is unknown; cannot create or update.",
		)
		return diags
	}
	if !isKibanaUnscoped(plan) && spaceID.ValueString() == "" {
		diags.AddError(
			"Invalid space identifier",
			"The space identifier from configuration is unknown or empty; cannot create or update.",
		)
	}
	return diags
}

// Create implements [resource.Resource]: decode plan and config, validate spaceID,
// resolve client, invoke the create callback, read-after-write, then persist state.
func (r *KibanaResource[T]) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(r.runKibanaWrite(ctx, resourceWriteInvocation{
		plan:         req.Plan,
		config:       req.Config,
		outState:     &resp.State,
		privateState: resp.Private,
		isUpdate:     false,
	})...)
}

// Read implements [resource.Resource] with the standard prelude: decode state,
// resolve identity via composite-ID-or-fallback, validate resourceID, resolve
// scoped Kibana client, then delegate to the concrete readFunc. When readFunc
// reports found==true, the returned model is persisted via resp.State.Set;
// when found==false, the resource is removed from state.
func (r *KibanaResource[T]) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if d := r.requireReadFunc(); d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}

	var model T
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	defaultTimeout := r.timeouts.Read
	if defaultTimeout <= 0 {
		defaultTimeout = DefaultResourceReadTimeout
	}
	readTimeout, timeoutDiags := model.GetTimeouts().Read(ctx, defaultTimeout)
	resp.Diagnostics.Append(timeoutDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	resourceID, spaceID := resolveKibanaResourceIdentity(model)
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

	if vDiags := EnforceVersionRequirements(ctx, client, &model); vDiags.HasError() {
		resp.Diagnostics.Append(vDiags...)
		return
	}

	resultModel, found, callDiags := r.readFunc(ctx, client, resourceID, spaceID, model)
	resp.Diagnostics.Append(callDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if found {
		if r.postReadFunc != nil {
			var prDiags diag.Diagnostics
			resultModel, prDiags = r.postReadFunc(ctx, KibanaPostReadRequest[T]{
				Client:  client,
				Prior:   model,
				State:   resultModel,
				Private: resp.Private,
			})
			resp.Diagnostics.Append(prDiags...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
		preserveModelTimeouts(&resultModel, model.GetTimeouts())
		resp.Diagnostics.Append(resp.State.Set(ctx, &resultModel)...)
		if resp.Diagnostics.HasError() {
			return
		}
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(attrTimeouts), model.GetTimeouts())...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		resp.State.RemoveResource(ctx)
	}
}

// Update implements [resource.Resource]: decode plan, prior state, and config,
// validate identity and spaceID, resolve client, invoke the update callback,
// read-after-write, then persist state.
func (r *KibanaResource[T]) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(r.runKibanaWrite(ctx, resourceWriteInvocation{
		plan:         req.Plan,
		priorState:   &req.State,
		config:       req.Config,
		outState:     &resp.State,
		privateState: resp.Private,
		isUpdate:     true,
	})...)
}

func (r *KibanaResource[T]) requireReadFunc() diag.Diagnostics {
	if r.readFunc == nil {
		return requireReadFuncDiag(r.component)
	}
	return nil
}

func (r *KibanaResource[T]) runKibanaWrite(ctx context.Context, inv resourceWriteInvocation) diag.Diagnostics {
	var diags diag.Diagnostics
	if (inv.isUpdate && r.updateFunc == nil) || (!inv.isUpdate && r.createFunc == nil) {
		op := envelopeWriteOpCreate
		if inv.isUpdate {
			op = envelopeWriteOpUpdate
		}
		diags.AddError(
			"Kibana envelope configuration error",
			fmt.Sprintf("The %s callback passed via KibanaResourceOptions must not be nil.", op),
		)
		return diags
	}

	var planModel T
	diags.Append(inv.plan.Get(ctx, &planModel)...)
	if diags.HasError() {
		return diags
	}

	var defaultTimeout time.Duration
	if inv.isUpdate {
		defaultTimeout = r.timeouts.Update
		if defaultTimeout <= 0 {
			defaultTimeout = DefaultResourceUpdateTimeout
		}
	} else {
		defaultTimeout = r.timeouts.Create
		if defaultTimeout <= 0 {
			defaultTimeout = DefaultResourceCreateTimeout
		}
	}
	var opTimeout time.Duration
	var timeoutDiags diag.Diagnostics
	if inv.isUpdate {
		opTimeout, timeoutDiags = planModel.GetTimeouts().Update(ctx, defaultTimeout)
	} else {
		opTimeout, timeoutDiags = planModel.GetTimeouts().Create(ctx, defaultTimeout)
	}
	diags.Append(timeoutDiags...)
	if diags.HasError() {
		return diags
	}
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	var priorPtr *T
	if inv.isUpdate && inv.priorState != nil {
		var priorModel T
		diags.Append(inv.priorState.Get(ctx, &priorModel)...)
		if diags.HasError() {
			return diags
		}
		priorPtr = &priorModel
	}

	diags.Append(r.validateSpaceID(planModel)...)
	if diags.HasError() {
		return diags
	}

	writeID := planModel.GetResourceID().ValueString()
	spaceID := planModel.GetSpaceID().ValueString()

	if inv.isUpdate {
		writeID, spaceID = resolveKibanaResourceIdentity(planModel)
		// When the plan identity is empty (for example because computed fields
		// were marked unknown in ModifyPlan), fall back to the prior state's
		// identity. This handles resources whose computed identifiers change
		// during Update.
		if writeID == "" && priorPtr != nil {
			writeID, spaceID = resolveKibanaResourceIdentity(*priorPtr)
		}
		if writeID == "" {
			diags.AddError(
				"Invalid resource identifier",
				"The resource identifier is empty; cannot update.",
			)
			return diags
		}
	}

	client, connDiags := r.Client().GetKibanaClient(ctx, planModel.GetKibanaConnection())
	diags.Append(connDiags...)
	if diags.HasError() {
		return diags
	}

	if vDiags := EnforceVersionRequirements(ctx, client, &planModel); vDiags.HasError() {
		diags.Append(vDiags...)
		return diags
	}

	if d := r.requireReadFunc(); d.HasError() {
		return d
	}

	var configModel T
	diags.Append(inv.config.Get(ctx, &configModel)...)
	if diags.HasError() {
		return diags
	}

	writeFn := r.createFunc
	if inv.isUpdate {
		writeFn = r.updateFunc
	}
	written, callDiags := writeFn(ctx, client, KibanaWriteRequest[T]{
		Plan:    planModel,
		Prior:   priorPtr,
		Config:  configModel,
		WriteID: writeID,
		SpaceID: spaceID,
	})
	diags.Append(callDiags...)
	if diags.HasError() {
		return diags
	}

	readResourceID := written.Model.GetResourceID().ValueString()
	readSpaceID := written.Model.GetSpaceID().ValueString()
	if readResourceID == "" {
		diags.AddError(
			"Invalid resource identifier",
			"The resolved read identity is empty after write; cannot refresh.",
		)
		return diags
	}

	if !isKibanaUnscoped(written.Model) && readSpaceID == "" {
		diags.AddError(
			"Invalid space identifier",
			"The resolved read space is empty after write; cannot refresh.",
		)
		return diags
	}

	stateModel, found, readDiags := r.readFunc(ctx, client, readResourceID, readSpaceID, written.Model)
	diags.Append(readDiags...)
	if diags.HasError() {
		return diags
	}

	if !found {
		notFoundDetail := fmt.Sprintf("%s_%s %q was not found after write", r.component, r.resourceName, readResourceID)
		if readSpaceID != "" {
			notFoundDetail = fmt.Sprintf("%s_%s %q in space %q was not found after write", r.component, r.resourceName, readResourceID, readSpaceID)
		}
		diags.AddError(
			"Resource not found",
			notFoundDetail,
		)
		return diags
	}

	priorModel := planModel

	if r.postReadFunc != nil {
		var prDiags diag.Diagnostics
		stateModel, prDiags = r.postReadFunc(ctx, KibanaPostReadRequest[T]{
			Client:  client,
			Prior:   priorModel,
			State:   stateModel,
			Private: inv.privateState,
		})
		diags.Append(prDiags...)
		if diags.HasError() {
			return diags
		}
	}

	preserveModelTimeouts(&stateModel, planModel.GetTimeouts())
	diags.Append(inv.outState.Set(ctx, &stateModel)...)
	if diags.HasError() {
		return diags
	}

	diags.Append(inv.outState.SetAttribute(ctx, path.Root(attrTimeouts), planModel.GetTimeouts())...)
	if diags.HasError() {
		return diags
	}

	return diags
}

// Delete implements [resource.Resource] with the standard prelude, then
// delegates to the concrete deleteFunc.
func (r *KibanaResource[T]) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.deleteFunc == nil {
		resp.Diagnostics.AddError(
			"Kibana envelope configuration error",
			"The delete callback passed via KibanaResourceOptions must not be nil.",
		)
		return
	}

	var model T
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	defaultTimeout := r.timeouts.Delete
	if defaultTimeout <= 0 {
		defaultTimeout = DefaultResourceDeleteTimeout
	}
	deleteTimeout, timeoutDiags := model.GetTimeouts().Delete(ctx, defaultTimeout)
	resp.Diagnostics.Append(timeoutDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	resourceID, spaceID := resolveKibanaResourceIdentity(model)
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
