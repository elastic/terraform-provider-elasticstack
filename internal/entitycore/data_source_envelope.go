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
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// KibanaConnectionField is an embeddable struct that provides the
// kibana_connection block field for data source models used with
// [NewKibanaDataSource].
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
// [NewKibanaDataSource]. It is satisfied by any struct that embeds
// [KibanaConnectionField] (or otherwise provides a GetKibanaConnection method).
type KibanaDataSourceModel interface {
	GetKibanaConnection() types.List
}

// ElasticsearchDataSourceModel is the type constraint for models passed to
// [NewElasticsearchDataSource]. It is satisfied by any struct that embeds
// [ElasticsearchConnectionField] (or otherwise provides a
// GetElasticsearchConnection method).
type ElasticsearchDataSourceModel interface {
	GetElasticsearchConnection() types.List
}

// genericKibanaDataSource implements [datasource.DataSource] and
// [datasource.DataSourceWithConfigure] for Kibana-backed data sources. It
// owns config decode, scoped client resolution, and state persistence; entity
// logic is delegated to the readFunc callback.
type genericKibanaDataSource[T KibanaDataSourceModel] struct {
	component      Component
	dataSourceName string
	client         *clients.ProviderClientFactory
	schemaFactory  func() dsschema.Schema
	readFunc       func(context.Context, *clients.KibanaScopedClient, T) (T, diag.Diagnostics)
}

// genericElasticsearchDataSource implements [datasource.DataSource] and
// [datasource.DataSourceWithConfigure] for Elasticsearch-backed data sources.
type genericElasticsearchDataSource[T ElasticsearchDataSourceModel] struct {
	component      Component
	dataSourceName string
	client         *clients.ProviderClientFactory
	schemaFactory  func() dsschema.Schema
	readFunc       func(context.Context, *clients.ElasticsearchScopedClient, T) (T, diag.Diagnostics)
}

// NewKibanaDataSource returns a [datasource.DataSource] that wraps the
// provided schema and read function with automatic kibana_connection block
// injection, config decode, scoped client resolution, and state persistence.
//
// The concrete model T must embed [KibanaConnectionField] so that the
// connection block can be decoded alongside entity attributes.
//
// Example usage (package doc):
//
//	type myModel struct {
//	    entitycore.KibanaConnectionField
//	    ID types.String `tfsdk:"id"`
//	}
//
//	func NewDataSource() datasource.DataSource {
//	    return entitycore.NewKibanaDataSource[myModel](
//	        entitycore.ComponentKibana,
//	        "my_entity",
//	        getDataSourceSchema, // returns datasource.Schema without kibana_connection block
//	        readMyEntity,
//	    )
//	}
func NewKibanaDataSource[T KibanaDataSourceModel](
	component Component,
	name string,
	schemaFactory func() dsschema.Schema,
	readFunc func(context.Context, *clients.KibanaScopedClient, T) (T, diag.Diagnostics),
) datasource.DataSource {
	return &genericKibanaDataSource[T]{
		component:      component,
		dataSourceName: name,
		schemaFactory:  schemaFactory,
		readFunc:       readFunc,
	}
}

// NewElasticsearchDataSource returns a [datasource.DataSource] that wraps the
// provided schema and read function with automatic elasticsearch_connection
// block injection, config decode, scoped client resolution, and state
// persistence.
//
// The concrete model T must embed [ElasticsearchConnectionField].
func NewElasticsearchDataSource[T ElasticsearchDataSourceModel](
	component Component,
	name string,
	schemaFactory func() dsschema.Schema,
	readFunc func(context.Context, *clients.ElasticsearchScopedClient, T) (T, diag.Diagnostics),
) datasource.DataSource {
	return &genericElasticsearchDataSource[T]{
		component:      component,
		dataSourceName: name,
		schemaFactory:  schemaFactory,
		readFunc:       readFunc,
	}
}

// Configure implements [datasource.DataSourceWithConfigure].
func (d *genericKibanaDataSource[T]) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	factory, diags := clients.ConvertProviderDataToFactory(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	d.client = factory
}

// Configure implements [datasource.DataSourceWithConfigure].
func (d *genericElasticsearchDataSource[T]) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	factory, diags := clients.ConvertProviderDataToFactory(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	d.client = factory
}

// Metadata implements [datasource.DataSource].
func (d *genericKibanaDataSource[T]) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_%s_%s", req.ProviderTypeName, d.component, d.dataSourceName)
}

// Metadata implements [datasource.DataSource].
func (d *genericElasticsearchDataSource[T]) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_%s_%s", req.ProviderTypeName, d.component, d.dataSourceName)
}

// Schema implements [datasource.DataSource], injecting the connection block.
func (d *genericKibanaDataSource[T]) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	schema := d.schemaFactory()
	blocks := make(map[string]dsschema.Block, len(schema.Blocks)+1)
	maps.Copy(blocks, schema.Blocks)
	blocks["kibana_connection"] = providerschema.GetKbFWConnectionBlock()
	schema.Blocks = blocks
	resp.Schema = schema
}

// Schema implements [datasource.DataSource], injecting the connection block.
func (d *genericElasticsearchDataSource[T]) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	schema := d.schemaFactory()
	blocks := make(map[string]dsschema.Block, len(schema.Blocks)+1)
	maps.Copy(blocks, schema.Blocks)
	blocks["elasticsearch_connection"] = providerschema.GetEsFWConnectionBlock()
	schema.Blocks = blocks
	resp.Schema = schema
}

// Read implements [datasource.DataSource].
func (d *genericKibanaDataSource[T]) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model T
	diags := req.Config.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, diags := d.client.GetKibanaClient(ctx, model.GetKibanaConnection())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, diags := d.readFunc(ctx, client, model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &result)
	resp.Diagnostics.Append(diags...)
}

// Read implements [datasource.DataSource].
func (d *genericElasticsearchDataSource[T]) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model T
	diags := req.Config.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, diags := d.client.GetElasticsearchClient(ctx, model.GetElasticsearchConnection())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, diags := d.readFunc(ctx, client, model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &result)
	resp.Diagnostics.Append(diags...)
}
