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
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// KibanaConnectionField is an embeddable struct that provides the
// kibana_connection block field for Kibana entity models used with
// [NewKibanaDataSource] or [NewKibanaResource].
type KibanaConnectionField struct {
	KibanaConnection types.List `tfsdk:"kibana_connection"`
}

// GetKibanaConnection returns the kibana_connection block value.
func (f KibanaConnectionField) GetKibanaConnection() types.List {
	return f.KibanaConnection
}

// ElasticsearchConnectionField is an embeddable struct that provides the
// elasticsearch_connection block field for data source models used with
// [NewElasticsearchDataSource].
type ElasticsearchConnectionField struct {
	ElasticsearchConnection types.List `tfsdk:"elasticsearch_connection"`
}

// GetElasticsearchConnection returns the elasticsearch_connection block value.
func (f ElasticsearchConnectionField) GetElasticsearchConnection() types.List {
	return f.ElasticsearchConnection
}

// KibanaDataSourceModel is the type constraint for models passed to
// [NewKibanaDataSource]. Concrete types must provide value-receiver methods
// GetID, GetResourceID, GetSpaceID, and GetKibanaConnection.
type KibanaDataSourceModel interface {
	GetID() types.String
	GetResourceID() types.String
	GetSpaceID() types.String
	GetKibanaConnection() types.List
}

// ElasticsearchDataSourceModel is the type constraint for models passed to
// [NewElasticsearchDataSource]. Concrete types must provide value-receiver
// methods GetID, GetResourceID, and GetElasticsearchConnection.
type ElasticsearchDataSourceModel interface {
	GetID() types.String
	GetResourceID() types.String
	GetElasticsearchConnection() types.List
}

type elasticsearchDataSourceReadFunc[T ElasticsearchDataSourceModel] func(
	context.Context,
	*clients.ElasticsearchScopedClient,
	string,
	T,
) (T, bool, diag.Diagnostics)

type elasticsearchDataSourcePostReadFunc[T ElasticsearchDataSourceModel] func(
	context.Context,
	*clients.ElasticsearchScopedClient,
	T,
) diag.Diagnostics

// ElasticsearchDataSourceOptions configures [NewElasticsearchDataSource].
// PostRead is optional.
type ElasticsearchDataSourceOptions[T ElasticsearchDataSourceModel] struct {
	Schema   func(context.Context) dsschema.Schema
	Read     elasticsearchDataSourceReadFunc[T]
	PostRead elasticsearchDataSourcePostReadFunc[T]
}

type kibanaDataSourceReadFunc[T KibanaDataSourceModel] func(
	context.Context,
	*clients.KibanaScopedClient,
	string,
	string,
	T,
) (T, bool, diag.Diagnostics)

type kibanaDataSourcePostReadFunc[T KibanaDataSourceModel] func(
	context.Context,
	*clients.KibanaScopedClient,
	T,
) diag.Diagnostics

// KibanaDataSourceOptions configures [NewKibanaDataSource]. PostRead is optional.
type KibanaDataSourceOptions[T KibanaDataSourceModel] struct {
	Schema   func(context.Context) dsschema.Schema
	Read     kibanaDataSourceReadFunc[T]
	PostRead kibanaDataSourcePostReadFunc[T]
}

// VersionRequirement describes a minimum server version that an entity model
// requires before the envelope invokes the concrete lifecycle callback.
type VersionRequirement struct {
	// MinVersion is the minimum server version required.
	MinVersion version.Version
	// ErrorMessage is the human-readable detail added to the
	// "Unsupported server version" diagnostic when the server does not
	// satisfy MinVersion.
	ErrorMessage string
}

// genericKibanaDataSource implements [datasource.DataSource] and
// [datasource.DataSourceWithConfigure] for Kibana-backed data sources. It
// owns config decode, scoped client resolution, identity resolution, and state
// persistence; entity logic is delegated to the readFunc callback.
type genericKibanaDataSource[T KibanaDataSourceModel] struct {
	*DataSourceBase
	schemaFactory func(context.Context) dsschema.Schema
	readFunc      kibanaDataSourceReadFunc[T]
	postReadFunc  kibanaDataSourcePostReadFunc[T]
}

// genericElasticsearchDataSource implements [datasource.DataSource] and
// [datasource.DataSourceWithConfigure] for Elasticsearch-backed data sources.
type genericElasticsearchDataSource[T ElasticsearchDataSourceModel] struct {
	*DataSourceBase
	schemaFactory func(context.Context) dsschema.Schema
	readFunc      elasticsearchDataSourceReadFunc[T]
	postReadFunc  elasticsearchDataSourcePostReadFunc[T]
}

// NewKibanaDataSource returns a [datasource.DataSource] that wraps the
// provided schema and read function with automatic kibana_connection block
// injection, config decode, scoped client resolution, identity resolution,
// centralized not-found handling, and state persistence.
//
// The concrete model T must satisfy [KibanaDataSourceModel] so that the
// connection block and read identity can be decoded and resolved.
func NewKibanaDataSource[T KibanaDataSourceModel](
	component Component,
	name string,
	opts KibanaDataSourceOptions[T],
) datasource.DataSource {
	return &genericKibanaDataSource[T]{
		DataSourceBase: NewDataSourceBase(component, name),
		schemaFactory:  opts.Schema,
		readFunc:       opts.Read,
		postReadFunc:   opts.PostRead,
	}
}

// NewElasticsearchDataSource returns a [datasource.DataSource] that wraps the
// provided schema and read function with automatic elasticsearch_connection
// block injection, config decode, scoped client resolution, identity
// resolution, centralized not-found handling, and state persistence.
//
// The concrete model T must satisfy [ElasticsearchDataSourceModel].
func NewElasticsearchDataSource[T ElasticsearchDataSourceModel](
	component Component,
	name string,
	opts ElasticsearchDataSourceOptions[T],
) datasource.DataSource {
	return &genericElasticsearchDataSource[T]{
		DataSourceBase: NewDataSourceBase(component, name),
		schemaFactory:  opts.Schema,
		readFunc:       opts.Read,
		postReadFunc:   opts.PostRead,
	}
}

// injectConnectionBlockIntoSchema returns a copy of the schema produced by
// schemaFactory with blockKey injected as an extra block. It allocates a new
// Blocks map so each call is independent and safe to reuse the same factory.
func injectConnectionBlockIntoSchema(ctx context.Context, schemaFactory func(context.Context) dsschema.Schema, blockKey string, block dsschema.Block) dsschema.Schema {
	schema := schemaFactory(ctx)
	blocks := make(map[string]dsschema.Block, len(schema.Blocks)+1)
	maps.Copy(blocks, schema.Blocks)
	blocks[blockKey] = block
	schema.Blocks = blocks
	return schema
}

// Schema implements [datasource.DataSource], injecting the connection block.
func (d *genericKibanaDataSource[T]) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = injectConnectionBlockIntoSchema(ctx, d.schemaFactory, blockKibanaConnection, providerschema.GetKbFWConnectionBlock())
}

// Schema implements [datasource.DataSource], injecting the connection block.
func (d *genericElasticsearchDataSource[T]) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = injectConnectionBlockIntoSchema(ctx, d.schemaFactory, blockElasticsearchConnection, providerschema.GetEsFWConnectionBlock())
}

func dataSourceNotFoundDiagnostic(component Component, name string, resourceID string, spaceID string) diag.Diagnostics {
	var diags diag.Diagnostics
	summary := fmt.Sprintf("%s_%s not found", component, name)
	detail := fmt.Sprintf("%s_%s %q was not found", component, name, resourceID)
	if spaceID != "" {
		detail = fmt.Sprintf("%s_%s %q in space %q was not found", component, name, resourceID, spaceID)
	}
	diags.AddError(summary, detail)
	return diags
}

type dataSourceReadParams struct {
	component Component
	name      string
}

// doElasticsearchDataSourceRead is the shared Read orchestration for
// genericElasticsearchDataSource.
func doElasticsearchDataSourceRead[T ElasticsearchDataSourceModel](
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
	params dataSourceReadParams,
	getClient func(context.Context, T) (*clients.ElasticsearchScopedClient, diag.Diagnostics),
	readFunc elasticsearchDataSourceReadFunc[T],
	postReadFunc elasticsearchDataSourcePostReadFunc[T],
) {
	var model T
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
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

	client, diags := getClient(ctx, model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if vDiags := EnforceVersionRequirements(ctx, client, &model); vDiags.HasError() {
		resp.Diagnostics.Append(vDiags...)
		return
	}

	if readFunc == nil {
		resp.Diagnostics.AddError(
			"Elasticsearch envelope configuration error",
			"The read callback passed via ElasticsearchDataSourceOptions must not be nil.",
		)
		return
	}

	result, found, callDiags := readFunc(ctx, client, resourceID, model)
	resp.Diagnostics.Append(callDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !found {
		resp.Diagnostics.Append(dataSourceNotFoundDiagnostic(params.component, params.name, resourceID, "")...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if postReadFunc != nil {
		resp.Diagnostics.Append(postReadFunc(ctx, client, result)...)
	}
}

// doKibanaDataSourceRead is the shared Read orchestration for genericKibanaDataSource.
func doKibanaDataSourceRead[T KibanaDataSourceModel](
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
	params dataSourceReadParams,
	getClient func(context.Context, T) (*clients.KibanaScopedClient, diag.Diagnostics),
	readFunc kibanaDataSourceReadFunc[T],
	postReadFunc kibanaDataSourcePostReadFunc[T],
) {
	var model T
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceID, spaceID := resolveKibanaResourceIdentity(model)
	if resourceID == "" {
		resp.Diagnostics.AddError(
			"Invalid resource identifier",
			"The resolved read identity is empty; cannot read.",
		)
		return
	}

	client, diags := getClient(ctx, model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if vDiags := EnforceVersionRequirements(ctx, client, &model); vDiags.HasError() {
		resp.Diagnostics.Append(vDiags...)
		return
	}

	if readFunc == nil {
		resp.Diagnostics.AddError(
			"Kibana envelope configuration error",
			"The read callback passed via KibanaDataSourceOptions must not be nil.",
		)
		return
	}

	result, found, callDiags := readFunc(ctx, client, resourceID, spaceID, model)
	resp.Diagnostics.Append(callDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !found {
		resp.Diagnostics.Append(dataSourceNotFoundDiagnostic(params.component, params.name, resourceID, spaceID)...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if postReadFunc != nil {
		resp.Diagnostics.Append(postReadFunc(ctx, client, result)...)
	}
}

// Read implements [datasource.DataSource].
func (d *genericKibanaDataSource[T]) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	doKibanaDataSourceRead(ctx, req, resp, dataSourceReadParams{
		component: d.component,
		name:      d.dataSourceName,
	}, func(ctx context.Context, model T) (*clients.KibanaScopedClient, diag.Diagnostics) {
		return d.Client().GetKibanaClient(ctx, model.GetKibanaConnection())
	}, d.readFunc, d.postReadFunc)
}

// Read implements [datasource.DataSource].
func (d *genericElasticsearchDataSource[T]) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	doElasticsearchDataSourceRead(ctx, req, resp, dataSourceReadParams{
		component: d.component,
		name:      d.dataSourceName,
	}, func(ctx context.Context, model T) (*clients.ElasticsearchScopedClient, diag.Diagnostics) {
		return d.Client().GetElasticsearchClient(ctx, model.GetElasticsearchConnection())
	}, d.readFunc, d.postReadFunc)
}
