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
)

// DataSourceBase holds shared Plugin Framework data source wiring: typed naming parts and
// the provider client factory from Configure. Embed *DataSourceBase in concrete data sources
// to reuse Configure, Metadata, and Client.
type DataSourceBase struct {
	component      Component
	dataSourceName string
	client         *clients.ProviderClientFactory
}

// NewDataSourceBase returns a [DataSourceBase] for the given namespace segment and literal data source name
// suffix. dataSourceName is not normalized; see package documentation.
func NewDataSourceBase(component Component, dataSourceName string) *DataSourceBase {
	return &DataSourceBase{component: component, dataSourceName: dataSourceName}
}

// Configure implements [datasource.DataSourceWithConfigure], converting provider
// data with [clients.ConvertProviderDataToFactory] and appending diagnostics. If
// the response has error diagnostics, it returns without assigning a new factory,
// leaving any prior successful client unchanged.
func (b *DataSourceBase) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	factory, diags := clients.ConvertProviderDataToFactory(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	b.client = factory
}

// Metadata implements the Metadata method of [datasource.DataSource], setting the Terraform type name to
// "<providerTypeName>_<component>_<dataSourceName>".
func (b *DataSourceBase) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_%s_%s", req.ProviderTypeName, b.component, b.dataSourceName)
}

// Client returns the client factory from the last successful [DataSourceBase.Configure]
// assignment, or nil if none has been stored yet. A nil *DataSourceBase returns nil so
// callers can surface diagnostics instead of panicking.
func (b *DataSourceBase) Client() *clients.ProviderClientFactory {
	if b == nil {
		return nil
	}
	return b.client
}
