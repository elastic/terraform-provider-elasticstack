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
	"strings"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ElasticsearchResourceModel is the type constraint for models passed to
// [NewElasticsearchResource]. Concrete types must provide value-receiver
// methods GetID, GetResourceID, GetElasticsearchConnection, and
// [WithResourceTimeouts] (typically by embedding [ResourceTimeoutsField]).
type ElasticsearchResourceModel interface {
	GetID() types.String
	// GetResourceID returns the plan-safe write identity (for example name or
	// username). Create and Update use this instead of GetID because computed
	// id values may be unknown in create plans.
	GetResourceID() types.String
	GetElasticsearchConnection() types.List
	WithResourceTimeouts
}

// WithReadResourceID is an optional interface for models that need a stable read
// identity distinct from the composite state ID segment used as the default.
type WithReadResourceID interface {
	GetReadResourceID() string
}

// WithOptionalWriteIdentity marks models whose write identity (GetResourceID) may
// be empty on Create when the API auto-generates an identifier (for example
// POST /_connector without a connector_id).
type WithOptionalWriteIdentity interface {
	AllowsEmptyWriteIdentityOnCreate() bool
}

// PrivateStateStorage is the subset of Terraform resource private state used by
// envelope write callbacks.
type PrivateStateStorage interface {
	GetKey(ctx context.Context, key string) ([]byte, diag.Diagnostics)
	SetKey(ctx context.Context, key string, value []byte) diag.Diagnostics
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

// WriteRequest is passed to [WriteFunc]. Config is the Terraform configuration
// decoded into T by the envelope before the callback is invoked. Prior is non-nil
// only for Update; Create receives Prior == nil. The same WriteRequest type is
// shared by Create and Update so a single function can serve both when the logic
// does not differ.
type WriteRequest[T ElasticsearchResourceModel] struct {
	Plan    T
	Prior   *T
	Config  T
	WriteID string
	// Private is the framework response Private field (typically
	// *internal/privatestate.Data). Nil when the callback does not need it.
	Private PrivateStateStorage
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

// ElasticsearchPostReadRequest is passed to [PostReadFunc]. Prior is the model
// before the read (plan on write path, prior state on plain Read path). State is
// the freshly-read model returned by the read callback.
type ElasticsearchPostReadRequest[T ElasticsearchResourceModel] struct {
	Client  *clients.ElasticsearchScopedClient
	Prior   T
	State   T
	Private PrivateStateStorage
}

// PostReadFunc runs after a successful read and before state is persisted,
// including read-after-write refresh. It is optional.
type PostReadFunc[T ElasticsearchResourceModel] func(
	ctx context.Context,
	req ElasticsearchPostReadRequest[T],
) (T, diag.Diagnostics)

// ElasticsearchResourceOptions configures [NewElasticsearchResource]. PostRead
// is optional; Schema, Read, Delete, Create, and Update must be non-nil or the
// envelope surfaces configuration diagnostics instead of invoking nil callbacks.
// Create and Update share the [WriteFunc] type so callers may pass the same
// function for both when the logic is identical.
//
// Timeouts supplies per-operation default durations when configuration omits
// `timeouts.<op>`; zero fields fall back to [DefaultResourceCreateTimeout],
// [DefaultResourceReadTimeout], [DefaultResourceUpdateTimeout], and
// [DefaultResourceDeleteTimeout]. Concrete schema factories MUST NOT include
// a `timeouts` attribute; the envelope injects it and silently overwrites any
// factory-supplied attribute with the same key.
type ElasticsearchResourceOptions[T ElasticsearchResourceModel] struct {
	Schema   func(context.Context) rschema.Schema
	Read     elasticsearchReadFunc[T]
	Delete   elasticsearchDeleteFunc[T]
	Create   WriteFunc[T]
	Update   WriteFunc[T]
	PostRead PostReadFunc[T]
	Timeouts ResourceTimeouts
	// SkipReadAfterWrite, when true, persists the write callback's WriteResult.Model
	// to state directly instead of re-reading via the read callback after Create/Update.
	// Use for resources whose write already returns the authoritative post-write state
	// and where a generic re-read would lose information (e.g. a transient state that the
	// write path detected but a subsequent read cannot reconstruct). The read callback is
	// still used for Read/refresh.
	SkipReadAfterWrite bool
}

// ElasticsearchResource implements [resource.Resource] and related interfaces
// for Elasticsearch-backed resources. It embeds [baseResourceEnvelope] to reuse
// Configure, Metadata, Client, Schema, Read, and Delete.
//
// The envelope owns Schema (with elasticsearch_connection block and timeouts
// attribute injection), Create, Read, Update, and Delete. Concrete resources may
// override Create or Update when their lifecycle does not fit the callback
// contract, and may choose to implement ImportState.
type ElasticsearchResource[T ElasticsearchResourceModel] struct {
	baseResourceEnvelope[T, *clients.ElasticsearchScopedClient]
	createFunc WriteFunc[T]
	updateFunc WriteFunc[T]
	// skipReadAfterWrite, when true, persists the write callback's WriteResult.Model
	// to state directly instead of re-reading via the read callback after Create/Update.
	skipReadAfterWrite bool
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

// UpdateNotSupportedWriteCallback returns a write callback that always returns
// an error diagnostic. Use for resources where all mutable attributes carry
// RequiresReplace, so Terraform never reaches an in-place update.
func UpdateNotSupportedWriteCallback[T ElasticsearchResourceModel]() WriteFunc[T] {
	return func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ WriteRequest[T]) (WriteResult[T], diag.Diagnostics) {
		var diags diag.Diagnostics
		diags.AddError("Update not supported", "Update not supported")
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
	r := &ElasticsearchResource[T]{}
	rb := NewResourceBase(ComponentElasticsearch, name)
	r.baseResourceEnvelope = baseResourceEnvelope[T, *clients.ElasticsearchScopedClient]{
		ResourceBase:    rb,
		schemaFactory:   opts.Schema,
		connectionKey:   blockElasticsearchConnection,
		connectionBlock: providerschema.GetEsFWConnectionBlock(),
		timeouts:        opts.Timeouts,
		resolveID: func(m T) (string, diag.Diagnostics) {
			return resolveElasticsearchReadResourceID(m, "")
		},
		getClient: func(ctx context.Context, m T) (*clients.ElasticsearchScopedClient, diag.Diagnostics) {
			return rb.Client().GetElasticsearchClient(ctx, m.GetElasticsearchConnection())
		},
		postRead: func(ctx context.Context, client *clients.ElasticsearchScopedClient, prior, state T, private PrivateStateStorage) (T, diag.Diagnostics) {
			if opts.PostRead == nil {
				return state, nil
			}
			return opts.PostRead(ctx, ElasticsearchPostReadRequest[T]{
				Client:  client,
				Prior:   prior,
				State:   state,
				Private: private,
			})
		},
	}
	r.skipReadAfterWrite = opts.SkipReadAfterWrite
	if opts.Read != nil {
		r.read = func(ctx context.Context, client *clients.ElasticsearchScopedClient, id string, m T) (T, bool, diag.Diagnostics) {
			return opts.Read(ctx, client, id, m)
		}
	}
	if opts.Delete != nil {
		r.delete = func(ctx context.Context, client *clients.ElasticsearchScopedClient, id string, m T) diag.Diagnostics {
			return opts.Delete(ctx, client, id, m)
		}
	}
	r.createFunc = opts.Create
	r.updateFunc = opts.Update
	return r
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
	compID, compDiags := clients.CompositeIDFromStr(model.GetID().ValueString())
	if compDiags.HasError() {
		// Fall back to GetResourceID when the state ID is not a composite.
		// This supports resources that were created by older provider versions
		// or imported with a plain resource identifier.
		if id := strings.TrimSpace(model.GetResourceID().ValueString()); id != "" {
			return id, diags
		}
		// Third fallback: use the raw ID string itself, since some older
		// provider versions stored the plain resource identifier as the state id.
		if id := strings.TrimSpace(model.GetID().ValueString()); id != "" {
			return id, diags
		}
		// All fallbacks exhausted. Return empty without the parse diagnostic
		// so callers can report their own "Invalid resource identifier" error.
		return "", diags
	}
	return compID.ResourceID, diags
}

// Create implements [resource.Resource]: decode plan, resolve client, invoke
// the create callback, read-after-write, then persist state from readFunc.
func (r *ElasticsearchResource[T]) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(r.runWrite(ctx, resourceWriteInvocation{
		plan:         req.Plan,
		config:       req.Config,
		outState:     &resp.State,
		privateState: resp.Private,
		isUpdate:     false,
	})...)
}

// Update implements [resource.Resource] with the same prelude as Create,
// additionally decoding prior state for the update callback.
func (r *ElasticsearchResource[T]) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(r.runWrite(ctx, resourceWriteInvocation{
		plan:         req.Plan,
		priorState:   &req.State,
		config:       req.Config,
		outState:     &resp.State,
		privateState: resp.Private,
		isUpdate:     true,
	})...)
}

func (r *ElasticsearchResource[T]) runWrite(ctx context.Context, inv resourceWriteInvocation) diag.Diagnostics {
	var diags diag.Diagnostics
	if (inv.isUpdate && r.updateFunc == nil) || (!inv.isUpdate && r.createFunc == nil) {
		op := envelopeWriteOpCreate
		if inv.isUpdate {
			op = envelopeWriteOpUpdate
		}
		return requireCallbackDiag(r.component, op)
	}

	var planModel T
	diags.Append(inv.plan.Get(ctx, &planModel)...)
	if diags.HasError() {
		return diags
	}

	var opTimeout time.Duration
	var timeoutDiags diag.Diagnostics
	if inv.isUpdate {
		opTimeout, timeoutDiags = planModel.GetTimeouts().Update(ctx, r.timeouts.UpdateOrDefault())
	} else {
		opTimeout, timeoutDiags = planModel.GetTimeouts().Create(ctx, r.timeouts.CreateOrDefault())
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

	writeID := planModel.GetResourceID()
	if !typeutils.IsKnown(writeID) {
		diags.AddError(
			"Invalid resource identifier",
			"The resource write identity from configuration is unknown; cannot create or update.",
		)
		return diags
	}
	allowEmptyCreate := false
	if opt, ok := any(planModel).(WithOptionalWriteIdentity); ok {
		allowEmptyCreate = opt.AllowsEmptyWriteIdentityOnCreate()
	}
	writeKey := writeID.ValueString()
	if writeKey == "" && (inv.isUpdate || !allowEmptyCreate) {
		diags.AddError(
			"Invalid resource identifier",
			"The resource write identity from configuration is empty; cannot create or update.",
		)
		return diags
	}

	client, connDiags := r.Client().GetElasticsearchClient(ctx, planModel.GetElasticsearchConnection())
	diags.Append(connDiags...)
	if diags.HasError() {
		return diags
	}

	if vDiags := EnforceVersionRequirements(ctx, client, &planModel); vDiags.HasError() {
		diags.Append(vDiags...)
		return diags
	}

	if r.read == nil {
		return requireReadFuncDiag(r.component)
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
	written, callDiags := writeFn(ctx, client, WriteRequest[T]{
		Plan:    planModel,
		Prior:   priorPtr,
		Config:  configModel,
		WriteID: writeKey,
		Private: inv.privateState,
	})
	diags.Append(callDiags...)
	if diags.HasError() {
		return diags
	}

	var stateModel T
	if r.skipReadAfterWrite {
		stateModel = written.Model
	} else {
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

		var found bool
		var readDiags diag.Diagnostics
		stateModel, found, readDiags = r.read(ctx, client, readResourceID, written.Model)
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
	}

	priorModel := planModel

	if r.postRead != nil {
		var prDiags diag.Diagnostics
		stateModel, prDiags = r.postRead(ctx, client, priorModel, stateModel, inv.privateState)
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

var (
	_ resource.Resource              = (*ElasticsearchResource[ElasticsearchResourceModel])(nil)
	_ resource.ResourceWithConfigure = (*ElasticsearchResource[ElasticsearchResourceModel])(nil)
)
