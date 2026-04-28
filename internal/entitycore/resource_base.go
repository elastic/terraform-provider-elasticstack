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
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Component is a Terraform type-name namespace segment used when building
// the full resource type name. See package documentation.
type Component string

// Well-known Terraform type-name namespace segments for [ResourceBase.Metadata].
const (
	ComponentElasticsearch Component = "elasticsearch"
	ComponentKibana        Component = "kibana"
	ComponentFleet         Component = "fleet"
	ComponentAPM           Component = "apm"
)

// ResourceBase holds shared Plugin Framework resource wiring: typed naming parts and
// the provider client factory from Configure. Embed *ResourceBase in concrete resources
// to reuse Configure, Metadata, and Client.
type ResourceBase struct {
	component    Component
	resourceName string
	client       *clients.ProviderClientFactory
}

// NewResourceBase returns a [ResourceBase] for the given namespace segment and literal resource name
// suffix. resourceName is not normalized; see package documentation.
func NewResourceBase(component Component, resourceName string) *ResourceBase {
	return &ResourceBase{component: component, resourceName: resourceName}
}

// Configure implements [resource.ResourceWithConfigure], converting provider
// data with [clients.ConvertProviderDataToFactory] and appending diagnostics. If
// the response has error diagnostics, it returns without assigning a new factory,
// leaving any prior successful client unchanged (same pattern as resources such
// as fleet integration and kibana agent builder tool).
func (c *ResourceBase) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	factory, diags := clients.ConvertProviderDataToFactory(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	c.client = factory
}

// Metadata implements the Metadata method of [resource.Resource], setting the Terraform type name to
// "<providerTypeName>_<component>_<resourceName>".
func (c *ResourceBase) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_%s_%s", req.ProviderTypeName, c.component, c.resourceName)
}

// Client returns the client factory from the last successful [ResourceBase.Configure]
// assignment, or nil if none has been stored yet. A nil *ResourceBase (e.g. a partially
// constructed embed) returns nil so callers can surface diagnostics instead of
// panicking.
func (c *ResourceBase) Client() *clients.ProviderClientFactory {
	if c == nil {
		return nil
	}
	return c.client
}
