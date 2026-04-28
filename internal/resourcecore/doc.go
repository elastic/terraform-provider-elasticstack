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

// Package resourcecore provides a shared embedded core for Plugin Framework resources:
// [Core] centralizes [resource.ResourceWithConfigure] Configure wiring, the Metadata method
// required by [resource.Resource], and stores the configured [*clients.ProviderClientFactory]
// for use via [Core.Client].
//
// Component is a typed Terraform resource type-name namespace segment (for example
// "elasticsearch", "kibana"). It is not a client-resolution kind: the same API family
// can use different component strings for Terraform naming, such as APM resources
// using the "apm" segment while calling Kibana APIs.
//
// The resourceName argument to [New] is the final literal suffix segment in the
// Terraform type name, joined without normalization. Callers must preserve existing
// spellings for compatibility (for example "agentbuilder_tool" versus "agent_builder_tool").
package resourcecore
