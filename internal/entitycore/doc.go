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
//	        getDataSourceSchema, // func(ctx context.Context) datasource.Schema, without kibana_connection block
//	        readMyEntity,
//	    )
//	}
//
// # Resource patterns
//
// Resources have three patterns:
//
//  1. **Struct-based embedding** — embed [*ResourceBase] and implement [resource.Resource]
//     directly. This is the right choice when Create and Update flows diverge
//     significantly from a uniform shape.
//
//  2. **Elasticsearch resource envelope** — use [NewElasticsearchResource] for
//     Elasticsearch-backed resources whose Create and Update flows match a common
//     shape: decode plan, resolve the scoped client from the connection block,
//     run a mutating API call using the plan-safe write identity from
//     [ElasticsearchResourceModel.GetResourceID], and persist the callback's
//     returned model. The model must satisfy [ElasticsearchResourceModel]
//     (value-receiver GetID for composite state ID, GetResourceID for the write
//     key such as name or username, and GetElasticsearchConnection). Supply a
//     schema factory (without elasticsearch_connection block), read and delete
//     callbacks, and required create and update callbacks
//     ([ElasticsearchCreateFunc], [ElasticsearchUpdateFunc]); pass the same
//     function for both when behavior matches. The envelope injects the
//     connection block, parses composite IDs for Read and Delete only, resolves
//     the client, and owns state persistence. It does not implement ImportState;
//     concrete resources add that when needed. Resources that still override
//     Create or Update (for example when the update path needs Config or prior
//     state in addition to Plan) may pass [PlaceholderElasticsearchWriteCallbacks]
//     until their logic is migrated into envelope callbacks. Constructor shape and
//     callback types are defined on [NewElasticsearchResource] in resource_envelope.go.
//
//  3. **Kibana resource envelope** — use [NewKibanaResource] for Kibana-backed
//     resources whose Create, Read, Update, and Delete flows match a common shape.
//     The model must satisfy [KibanaResourceModel] (value-receiver GetID for
//     composite or plain state ID, GetResourceID for the write key such as name
//     or API-assigned UUID, GetSpaceID for the Kibana space, and
//     GetKibanaConnection). Supply a schema factory (without kibana_connection
//     block), read and delete callbacks, and required create and update callbacks
//     ([KibanaCreateFunc], [KibanaUpdateFunc]). The envelope injects the
//     kibana_connection block, resolves resource identity via composite-ID-or-fallback
//     for Read, Update, and Delete, validates spaceID for Create, resolves the
//     scoped Kibana client, and owns state persistence. Create callbacks receive
//     the plan model (and can call plan.GetResourceID() for user-ID resources);
//     Update callbacks receive both plan and prior state. It does not implement
//     ImportState; concrete resources add that when needed. Resources that override
//     Create or Update may pass [PlaceholderKibanaWriteCallbacks] until their logic
//     is migrated into envelope callbacks. Constructor shape and callback types are
//     defined on [NewKibanaResource] in kibana_resource_envelope.go.
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
