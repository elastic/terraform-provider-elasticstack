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
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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

// DataSourceBase holds shared Plugin Framework data source wiring: typed naming
// parts and the provider client factory from Configure. Embed *DataSourceBase in
// concrete data sources to reuse Configure, Metadata, and Client.
type DataSourceBase struct {
	component      Component
	dataSourceName string
	client         *clients.ProviderClientFactory
}

// NewDataSourceBase returns a [DataSourceBase] for the given namespace segment
// and literal data source name suffix. dataSourceName is not normalized; see
// package documentation.
func NewDataSourceBase(component Component, dataSourceName string) *DataSourceBase {
	return &DataSourceBase{component: component, dataSourceName: dataSourceName}
}

// Configure implements [datasource.DataSourceWithConfigure], converting provider
// data with [clients.ConvertProviderDataToFactory] and appending diagnostics. If
// the response has error diagnostics, it returns without assigning a new factory,
// leaving any prior successful client unchanged.
func (d *DataSourceBase) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	factory, diags := clients.ConvertProviderDataToFactory(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	d.client = factory
}

// Metadata implements the Metadata method of [datasource.DataSource], setting
// the Terraform type name to "<providerTypeName>_<component>_<dataSourceName>".
func (d *DataSourceBase) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_%s_%s", req.ProviderTypeName, d.component, d.dataSourceName)
}

// Client returns the client factory from the last successful Configure call,
// or nil if none has been stored yet. A nil *DataSourceBase returns nil so
// callers can surface diagnostics instead of panicking.
func (d *DataSourceBase) Client() *clients.ProviderClientFactory {
	if d == nil {
		return nil
	}
	return d.client
}

// configureFactoryFromProviderData converts provider data to a client factory.
// Ephemeral envelopes share this helper because their Configure receiver types
// differ from [ResourceBase.Configure] and cannot embed [ResourceBase] directly.
func configureFactoryFromProviderData(providerData any) (*clients.ProviderClientFactory, diag.Diagnostics) {
	return clients.ConvertProviderDataToFactory(providerData)
}

// EphemeralBase holds shared Plugin Framework ephemeral resource wiring:
// typed naming parts and the provider client factory from Configure.
type EphemeralBase struct {
	component     Component
	ephemeralName string
	client        *clients.ProviderClientFactory
}

// NewEphemeralBase returns an [EphemeralBase] for the given namespace segment
// and literal ephemeral resource name suffix.
func NewEphemeralBase(component Component, ephemeralName string) *EphemeralBase {
	return &EphemeralBase{component: component, ephemeralName: ephemeralName}
}

// Metadata sets the Terraform type name to "<providerTypeName>_<component>_<ephemeralName>".
func (e *EphemeralBase) Metadata(providerTypeName string) string {
	return fmt.Sprintf("%s_%s_%s", providerTypeName, e.component, e.ephemeralName)
}

// Client returns the client factory from the last successful Configure call,
// or nil if none has been stored yet.
func (e *EphemeralBase) Client() *clients.ProviderClientFactory {
	if e == nil {
		return nil
	}
	return e.client
}

// SetClient assigns the configured client factory when Configure succeeds.
func (e *EphemeralBase) SetClient(factory *clients.ProviderClientFactory) {
	if e == nil {
		return
	}
	e.client = factory
}
