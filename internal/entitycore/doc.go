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

// Package entitycore provides a shared embedded core for Plugin Framework entities:
// [ResourceBase] centralizes [resource.ResourceWithConfigure] Configure wiring, the Metadata method
// required by [resource.Resource], and stores the configured [*clients.ProviderClientFactory]
// for use via [ResourceBase.Client]. Data sources use envelope generics.
//
// # Data source patterns
//
// Plugin Framework data sources use **envelope generics** — [NewKibanaDataSource] or
// [NewElasticsearchDataSource] — which eliminate Read orchestration boilerplate.
// The constructor owns config decode, scoped client resolution, and state persistence.
// The concrete package provides only a schema factory (without connection blocks),
// a model that embeds [KibanaConnectionField] or [ElasticsearchConnectionField], and
// a pure read function that performs the entity-specific API call and model mapping.
//
// Example envelope data source:
//
//	type myModel struct {
//	    entitycore.KibanaConnectionField
//	    ID types.String `tfsdk:"id"`
//	}
//
//	func readMyEntity(ctx context.Context, client *clients.KibanaScopedClient, model myModel) (myModel, diag.Diagnostics) {
//	    // API call and model population …
//	    return model, nil
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
//
// # Resource patterns
//
// Resources have the same two patterns:
//
//  1. **Struct-based embedding** — embed [*ResourceBase] and implement [resource.Resource]
//     directly. This is the right choice when Create and Update flows diverge
//     significantly from a uniform shape.
//
//  2. **Envelope generics** — use [NewElasticsearchResource] to eliminate duplicated
//     Read/Delete/Schema preludes for Elasticsearch-backed resources.
//     The model must satisfy [ElasticsearchResourceModel] (value-receiver GetID and
//     GetElasticsearchConnection). The concrete resource provides a schema factory
//     (without elasticsearch_connection block), a read callback returning
//     (T, bool, diag.Diagnostics), and a delete callback. The envelope injects the
//     connection block, parses composite IDs, resolves the scoped client, and
//     persists state. Concrete resources keep Create, Update, and ImportState;
//     the envelope's default Create and Update return diagnostics so forgotten
//     overrides fail loudly during runtime instead of silently no-oping.
//
//     Example envelope resource:
//
//     type myResource struct {
//     *entitycore.ElasticsearchResource[Data]
//     }
//
//     func readMyEntity(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, state Data) (Data, bool, diag.Diagnostics) {
//     // API call and model population …
//     return state, true, nil
//     }
//
//     func deleteMyEntity(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, state Data) diag.Diagnostics {
//     // API call to delete …
//     return nil
//     }
//
//     func newMyResource() *myResource {
//     return &myResource{
//     ElasticsearchResource: entitycore.NewElasticsearchResource[Data](
//     entitycore.ComponentElasticsearch,
//     "my_entity",
//     getResourceSchema, // returns resource.Schema without elasticsearch_connection block
//     readMyEntity,
//     deleteMyEntity,
//     ),
//     }
//     }
//
// Component is a typed Terraform resource type-name namespace segment (for example
// "elasticsearch", "kibana"). It is not a client-resolution kind: the same API family
// can use different component strings for Terraform naming, such as APM resources
// using the "apm" segment while calling Kibana APIs.
//
// The resourceName argument to [NewResourceBase] is the final literal suffix segment in the
// Terraform type name, joined without normalization. Callers must preserve existing
// spellings for compatibility (for example "agentbuilder_tool" versus "agent_builder_tool").
package entitycore
