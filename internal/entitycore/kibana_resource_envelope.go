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
// flow uses Model when resolving refresh identity and calling read.
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
type KibanaResourceOptions[T KibanaResourceModel] struct {
	Schema   func(context.Context) rschema.Schema
	Read     kibanaReadFunc[T]
	Delete   kibanaDeleteFunc[T]
	Create   KibanaWriteFunc[T]
	Update   KibanaWriteFunc[T]
	PostRead KibanaPostReadFunc[T]
}

// KibanaResource implements [resource.Resource] and related interfaces
// for Kibana-backed resources. It embeds [baseResourceEnvelope] to reuse
// Configure, Metadata, Client, Schema, Read, and Delete.
//
// The envelope owns Schema (with kibana_connection block injection),
// Create, Read, Update, and Delete. Concrete resources may override Create or
// Update when their lifecycle does not fit the callback contract, and may
// choose to implement ImportState.
type KibanaResource[T KibanaResourceModel] struct {
	baseResourceEnvelope[T, *clients.KibanaScopedClient]
	createFunc KibanaWriteFunc[T]
	updateFunc KibanaWriteFunc[T]
	readFunc   kibanaReadFunc[T]
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
	r := &KibanaResource[T]{}
	rb := NewResourceBase(component, name)
	r.baseResourceEnvelope = baseResourceEnvelope[T, *clients.KibanaScopedClient]{
		ResourceBase:    rb,
		schemaFactory:   opts.Schema,
		connectionKey:   blockKibanaConnection,
		connectionBlock: providerschema.GetKbFWConnectionBlock(),
		resolveID: func(m T) (string, diag.Diagnostics) {
			resourceID, _ := resolveKibanaResourceIdentity(m)
			return resourceID, nil
		},
		getClient: func(ctx context.Context, m T) (*clients.KibanaScopedClient, diag.Diagnostics) {
			return rb.Client().GetKibanaClient(ctx, m.GetKibanaConnection())
		},
		postRead: func(ctx context.Context, client *clients.KibanaScopedClient, prior, state T, private PrivateStateStorage) (T, diag.Diagnostics) {
			if opts.PostRead == nil {
				return state, nil
			}
			return opts.PostRead(ctx, KibanaPostReadRequest[T]{
				Client:  client,
				Prior:   prior,
				State:   state,
				Private: private,
			})
		},
	}
	r.readFunc = opts.Read
	if opts.Read != nil {
		r.read = func(ctx context.Context, client *clients.KibanaScopedClient, id string, m T) (T, bool, diag.Diagnostics) {
			_, spaceID := resolveKibanaResourceIdentity(m)
			return opts.Read(ctx, client, id, spaceID, m)
		}
	}
	if opts.Delete != nil {
		r.delete = func(ctx context.Context, client *clients.KibanaScopedClient, id string, m T) diag.Diagnostics {
			_, spaceID := resolveKibanaResourceIdentity(m)
			return opts.Delete(ctx, client, id, spaceID, m)
		}
	}
	r.createFunc = opts.Create
	r.updateFunc = opts.Update
	return r
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

func (r *KibanaResource[T]) runKibanaWrite(ctx context.Context, inv resourceWriteInvocation) diag.Diagnostics {
	// resolvedSpaceID and readSpaceID are captured across adapter callbacks via
	// closure: validateIdentity sets resolvedSpaceID; resolveReadIdentity sets
	// readSpaceID; invokeWrite, doRead, and notFoundDetail consume them.
	var resolvedSpaceID string
	var readSpaceID string

	adapter := writeCommonAdapter[T, *clients.KibanaScopedClient]{
		validateIdentity: func(planModel T, priorPtr *T, isUpdate bool) (string, diag.Diagnostics) {
			if diags := r.validateSpaceID(planModel); diags.HasError() {
				return "", diags
			}

			writeID := planModel.GetResourceID().ValueString()
			spaceID := planModel.GetSpaceID().ValueString()

			if isUpdate {
				writeID, spaceID = resolveKibanaResourceIdentity(planModel)
				// When the plan identity is empty (for example because computed
				// fields were marked unknown in ModifyPlan), fall back to the
				// prior state's identity. This handles resources whose computed
				// identifiers change during Update.
				if writeID == "" && priorPtr != nil {
					writeID, spaceID = resolveKibanaResourceIdentity(*priorPtr)
				}
				if writeID == "" {
					var diags diag.Diagnostics
					diags.AddError(
						"Invalid resource identifier",
						"The resource identifier is empty; cannot update.",
					)
					return "", diags
				}
			}
			resolvedSpaceID = spaceID
			return writeID, nil
		},
		getClient: func(ctx context.Context, planModel T) (*clients.KibanaScopedClient, diag.Diagnostics) {
			return r.Client().GetKibanaClient(ctx, planModel.GetKibanaConnection())
		},
		checkReadFunc: func() diag.Diagnostics {
			if r.readFunc == nil {
				return requireReadFuncDiag(r.component)
			}
			return nil
		},
		invokeWrite: func(ctx context.Context, client *clients.KibanaScopedClient, planModel T, priorPtr *T, configModel T, writeID string, isUpdate bool, _ PrivateStateStorage) (T, diag.Diagnostics) {
			writeFn := r.createFunc
			if isUpdate {
				writeFn = r.updateFunc
			}
			written, d := writeFn(ctx, client, KibanaWriteRequest[T]{
				Plan:    planModel,
				Prior:   priorPtr,
				Config:  configModel,
				WriteID: writeID,
				SpaceID: resolvedSpaceID,
			})
			return written.Model, d
		},
		resolveReadIdentity: func(writtenModel T, _ string) (string, diag.Diagnostics) {
			var diags diag.Diagnostics
			readID := writtenModel.GetResourceID().ValueString()
			readSpaceID = writtenModel.GetSpaceID().ValueString()
			if readID == "" {
				diags.AddError(
					"Invalid resource identifier",
					"The resolved read identity is empty after write; cannot refresh.",
				)
				return "", diags
			}
			if !isKibanaUnscoped(writtenModel) && readSpaceID == "" {
				diags.AddError(
					"Invalid space identifier",
					"The resolved read space is empty after write; cannot refresh.",
				)
				return "", diags
			}
			return readID, diags
		},
		doRead: func(ctx context.Context, client *clients.KibanaScopedClient, readID string, writtenModel T) (T, bool, diag.Diagnostics) {
			return r.readFunc(ctx, client, readID, readSpaceID, writtenModel)
		},
		notFoundDetail: func(_, readID string) string {
			if readSpaceID != "" {
				return fmt.Sprintf("%s_%s %q in space %q was not found after write", r.component, r.resourceName, readID, readSpaceID)
			}
			return fmt.Sprintf("%s_%s %q was not found after write", r.component, r.resourceName, readID)
		},
		doPostRead: func(ctx context.Context, client *clients.KibanaScopedClient, priorModel, stateModel T, private PrivateStateStorage) (T, diag.Diagnostics) {
			if r.postRead == nil {
				return stateModel, nil
			}
			return r.postRead(ctx, client, priorModel, stateModel, private)
		},
	}
	return runWriteCommon(ctx, inv, r.component, r.createFunc == nil, r.updateFunc == nil, adapter)
}

var (
	_ resource.Resource              = (*KibanaResource[KibanaResourceModel])(nil)
	_ resource.ResourceWithConfigure = (*KibanaResource[KibanaResourceModel])(nil)
)
