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

package resourcecore

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Component is a Terraform type-name namespace segment used when building
// the full resource type name. See package documentation.
type Component string

// Well-known Terraform type-name namespace segments for [Core.Metadata].
const (
	ComponentElasticsearch Component = "elasticsearch"
	ComponentKibana        Component = "kibana"
	ComponentFleet         Component = "fleet"
	ComponentAPM           Component = "apm"
)

// Core holds shared Plugin Framework resource wiring: typed naming parts and
// the provider client factory from Configure. Embed *Core in concrete resources
// to reuse Configure, Metadata, and Client.
type Core struct {
	component    Component
	resourceName string
	client       *clients.ProviderClientFactory
}

// New returns a [Core] for the given namespace segment and literal resource name
// suffix. resourceName is not normalized; see package documentation.
func New(component Component, resourceName string) *Core {
	return &Core{component: component, resourceName: resourceName}
}

// Configure implements [resource.ResourceWithConfigure], converting provider
// data with [clients.ConvertProviderDataToFactory] and appending diagnostics. If
// the response has error diagnostics, it returns without assigning a new factory,
// leaving any prior successful client unchanged (same pattern as resources such
// as fleet integration and kibana agent builder tool).
func (c *Core) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	factory, diags := clients.ConvertProviderDataToFactory(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	c.client = factory
}

// Metadata implements the Metadata method of [resource.Resource], setting the Terraform type name to
// "<providerTypeName>_<component>_<resourceName>".
func (c *Core) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_%s_%s", req.ProviderTypeName, c.component, c.resourceName)
}

// Client returns the client factory from the last successful [Core.Configure]
// assignment, or nil if none has been stored yet. A nil *Core (e.g. a partially
// constructed embed) returns nil so callers can surface diagnostics instead of
// panicking.
func (c *Core) Client() *clients.ProviderClientFactory {
	if c == nil {
		return nil
	}
	return c.client
}
