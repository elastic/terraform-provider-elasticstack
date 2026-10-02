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
	actiontimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/action/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/action"
	actionschema "github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DefaultActionInvokeTimeout is used when an [ActionOptions] entry leaves
// DefaultInvokeTimeout zero. It is generous because actions typically wrap
// long-running imperative operations (snapshot, restore, reindex).
const DefaultActionInvokeTimeout = 20 * time.Minute

// ActionTimeoutsField is an embeddable struct that provides the action
// `timeouts` block field for action models used with [NewElasticsearchAction]
// or [NewKibanaAction]. Embedding it satisfies [WithActionTimeouts] without
// requiring the concrete model to redeclare the framework type.
type ActionTimeoutsField struct {
	Timeouts actiontimeouts.Value `tfsdk:"timeouts"`
}

// GetTimeouts returns the timeouts block value.
func (f ActionTimeoutsField) GetTimeouts() actiontimeouts.Value {
	return f.Timeouts
}

// WithActionTimeouts is the timeouts portion of the action model contract.
// Concrete action models satisfy it by embedding [ActionTimeoutsField] (or by
// declaring an equivalent field plus method).
type WithActionTimeouts interface {
	GetTimeouts() actiontimeouts.Value
}

// ElasticsearchActionModel is the type constraint for models passed to
// [NewElasticsearchAction]. Concrete types satisfy it by embedding both
// [ElasticsearchConnectionField] and [ActionTimeoutsField].
type ElasticsearchActionModel interface {
	GetElasticsearchConnection() types.List
	WithActionTimeouts
}

// KibanaActionModel is the type constraint for models passed to
// [NewKibanaAction]. Concrete types satisfy it by embedding both
// [KibanaConnectionField] and [ActionTimeoutsField].
type KibanaActionModel interface {
	GetKibanaConnection() types.List
	WithActionTimeouts
}

// ActionRequest is passed to action Invoke callbacks. Config is the decoded
// model from the Terraform configuration. SendProgress mirrors
// [action.InvokeResponse.SendProgress] so callbacks can stream progress
// events to Terraform without holding a reference to the framework response
// struct.
type ActionRequest[T any] struct {
	Config       T
	SendProgress func(action.InvokeProgressEvent)
}

// ActionInvokeFunc performs the action's work after the envelope has decoded
// the configuration, resolved the scoped client, evaluated optional version
// requirements, and applied the invoke timeout to ctx via
// [context.WithTimeout]. The callback returns diagnostics; the envelope
// appends them to the framework response.
type ActionInvokeFunc[T any, Client MinVersionClient] func(
	ctx context.Context,
	client Client,
	req ActionRequest[T],
) diag.Diagnostics

// ElasticsearchActionOptions configures [NewElasticsearchAction].
// Schema and Invoke must be non-nil or the constructor panics.
// DefaultInvokeTimeout is used when the configuration omits `timeouts.invoke`;
// zero falls back to [DefaultActionInvokeTimeout].
type ElasticsearchActionOptions[T ElasticsearchActionModel] struct {
	Schema               func(context.Context) actionschema.Schema
	Invoke               ActionInvokeFunc[T, *clients.ElasticsearchScopedClient]
	DefaultInvokeTimeout time.Duration
}

// KibanaActionOptions configures [NewKibanaAction].
// Schema and Invoke must be non-nil or the constructor panics.
// DefaultInvokeTimeout is used when the configuration omits `timeouts.invoke`;
// zero falls back to [DefaultActionInvokeTimeout].
type KibanaActionOptions[T KibanaActionModel] struct {
	Schema               func(context.Context) actionschema.Schema
	Invoke               ActionInvokeFunc[T, *clients.KibanaScopedClient]
	DefaultInvokeTimeout time.Duration
}

// ActionBase holds shared Plugin Framework action wiring: typed naming parts
// and the provider client factory from Configure. It is the action analogue
// of [ResourceBase] / [DataSourceBase] / [EphemeralBase].
type ActionBase struct {
	component  Component
	actionName string
	client     *clients.ProviderClientFactory
}

// NewActionBase returns an [ActionBase] for the given namespace segment and
// literal action name suffix. actionName is not normalized; see package
// documentation.
func NewActionBase(component Component, actionName string) *ActionBase {
	return &ActionBase{component: component, actionName: actionName}
}

// Metadata implements the Metadata method of [action.Action], setting the
// Terraform type name to "<providerTypeName>_<component>_<actionName>".
func (a *ActionBase) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_%s_%s", req.ProviderTypeName, a.component, a.actionName)
}

// Configure implements [action.ActionWithConfigure], converting provider data
// with [clients.ConvertProviderDataToFactory] and appending diagnostics. If
// the response has error diagnostics it returns without assigning a new
// factory, leaving any prior successful client unchanged. ProviderData == nil
// is permitted because the framework calls Configure twice (once before
// provider config and once after) and we must not surface a spurious error
// during the early call.
func (a *ActionBase) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	factory, diags := clients.ConvertProviderDataToFactory(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	a.client = factory
}

// Client returns the client factory from the last successful
// [ActionBase.Configure] assignment, or nil if none has been stored yet.
func (a *ActionBase) Client() *clients.ProviderClientFactory {
	if a == nil {
		return nil
	}
	return a.client
}

// actionAdapter holds connection-type-specific operations for
// [genericAction]. It captures the only meaningful differences between
// Elasticsearch and Kibana actions: which connection list to read from the
// model, which scoped client to resolve, and which connection block to
// inject into the schema.
type actionAdapter[T any, Client MinVersionClient] struct {
	getConnection      func(model T) types.List
	getClient          func(ctx context.Context, factory *clients.ProviderClientFactory, connection types.List) (Client, diag.Diagnostics)
	schemaBlockKey     string
	schemaBlockFactory func() actionschema.Block
	errorSummary       string
}

// genericAction implements [action.Action] and [action.ActionWithConfigure]
// for any connection-backed action. All lifecycle boilerplate lives here;
// connection-type-specific operations are delegated to the adapter. The
// envelope owns Metadata, Configure, Schema (with the connection and
// timeouts blocks injected), and Invoke prelude (decode, factory check,
// scoped client resolution, version-requirement enforcement, timeout
// application).
type genericAction[T WithActionTimeouts, Client MinVersionClient] struct {
	*ActionBase
	schemaFactory  func(context.Context) actionschema.Schema
	invokeFunc     ActionInvokeFunc[T, Client]
	defaultTimeout time.Duration
	adapter        actionAdapter[T, Client]
}

// Schema implements [action.Action], injecting the `timeouts` block (always)
// and the connection block (`elasticsearch_connection` or `kibana_connection`)
// into the schema returned by the concrete schema factory. Concrete actions
// SHOULD NOT include either block in their factory output; the envelope owns
// them so every action gets identical behavior.
func (a *genericAction[T, Client]) Schema(ctx context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	schema := a.schemaFactory(ctx)
	blocks := make(map[string]actionschema.Block, len(schema.Blocks)+2)
	maps.Copy(blocks, schema.Blocks)
	blocks[blockTimeouts] = actiontimeouts.Block(ctx)
	blocks[a.adapter.schemaBlockKey] = a.adapter.schemaBlockFactory()
	schema.Blocks = blocks
	resp.Schema = schema
}

// Invoke implements [action.Action] with a fixed prelude:
//  1. Decode the configuration into T.
//  2. Verify the provider client factory was configured.
//  3. Resolve the scoped client via the adapter.
//  4. Evaluate optional version requirements via [EnforceVersionRequirements].
//  5. Read `timeouts.invoke` and wrap ctx with [context.WithTimeout].
//  6. Delegate to the user-supplied invoke callback.
//
// SendProgress is forwarded to the callback so callers can stream progress
// to Terraform without referencing the framework response struct directly.
func (a *genericAction[T, Client]) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var model T
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	factory := a.Client()
	if factory == nil {
		resp.Diagnostics.AddError(
			a.adapter.errorSummary,
			"Provider not configured: expected configured provider client factory.",
		)
		return
	}

	client, connDiags := a.adapter.getClient(ctx, factory, a.adapter.getConnection(model))
	resp.Diagnostics.Append(connDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if vDiags := EnforceVersionRequirements(ctx, client, &model); vDiags.HasError() {
		resp.Diagnostics.Append(vDiags...)
		return
	}

	defaultTimeout := a.defaultTimeout
	if defaultTimeout <= 0 {
		defaultTimeout = DefaultActionInvokeTimeout
	}
	invokeTimeout, timeoutDiags := model.GetTimeouts().Invoke(ctx, defaultTimeout)
	resp.Diagnostics.Append(timeoutDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, invokeTimeout)
	defer cancel()

	resp.Diagnostics.Append(a.invokeFunc(ctx, client, ActionRequest[T]{
		Config:       model,
		SendProgress: resp.SendProgress,
	})...)
}

// blockTimeouts is the schema block key the action envelope injects for the
// `timeouts` block. Concrete actions must not include a block under this key.
const blockTimeouts = "timeouts"

// NewElasticsearchAction returns an [action.Action] that owns Metadata,
// Configure, Schema (with `elasticsearch_connection` and `timeouts` block
// injection), and the Invoke prelude for the Elasticsearch namespace. The
// concrete action only needs to supply a schema factory (without those two
// blocks) and an invoke callback.
//
// Example:
//
//	type Model struct {
//	    entitycore.ElasticsearchConnectionField
//	    entitycore.ActionTimeoutsField
//	    Repository types.String `tfsdk:"repository"`
//	    // …
//	}
//
//	func NewMyAction() action.Action {
//	    return entitycore.NewElasticsearchAction[Model]("my_action", entitycore.ElasticsearchActionOptions[Model]{
//	        Schema:               getSchema,
//	        Invoke:               invokeMyAction,
//	        DefaultInvokeTimeout: 30 * time.Minute,
//	    })
//	}
func NewElasticsearchAction[T ElasticsearchActionModel](
	name string,
	opts ElasticsearchActionOptions[T],
) action.Action {
	if opts.Schema == nil {
		panic("entitycore: ElasticsearchActionOptions.Schema must not be nil")
	}
	if opts.Invoke == nil {
		panic("entitycore: ElasticsearchActionOptions.Invoke must not be nil")
	}
	return &genericAction[T, *clients.ElasticsearchScopedClient]{
		ActionBase:     NewActionBase(ComponentElasticsearch, name),
		schemaFactory:  opts.Schema,
		invokeFunc:     opts.Invoke,
		defaultTimeout: opts.DefaultInvokeTimeout,
		adapter: actionAdapter[T, *clients.ElasticsearchScopedClient]{
			getConnection: func(model T) types.List { return model.GetElasticsearchConnection() },
			getClient: func(ctx context.Context, factory *clients.ProviderClientFactory, connection types.List) (*clients.ElasticsearchScopedClient, diag.Diagnostics) {
				return factory.GetElasticsearchClient(ctx, connection)
			},
			schemaBlockKey:     blockElasticsearchConnection,
			schemaBlockFactory: providerschema.GetEsActionConnectionBlock,
			errorSummary:       "Elasticsearch action envelope error",
		},
	}
}

// Compile-time interface satisfaction guards.
var (
	_ action.Action              = (*genericAction[elasticsearchActionModelGuard, *clients.ElasticsearchScopedClient])(nil)
	_ action.ActionWithConfigure = (*genericAction[elasticsearchActionModelGuard, *clients.ElasticsearchScopedClient])(nil)
	_ action.Action              = (*genericAction[kibanaActionModelGuard, *clients.KibanaScopedClient])(nil)
	_ action.ActionWithConfigure = (*genericAction[kibanaActionModelGuard, *clients.KibanaScopedClient])(nil)
)

// elasticsearchActionModelGuard is a minimal model that satisfies
// [ElasticsearchActionModel] for compile-time interface assertions.
type elasticsearchActionModelGuard struct {
	ElasticsearchConnectionField
	ActionTimeoutsField
}

// kibanaActionModelGuard is a minimal model that satisfies
// [KibanaActionModel] for compile-time interface assertions.
type kibanaActionModelGuard struct {
	KibanaConnectionField
	ActionTimeoutsField
}
