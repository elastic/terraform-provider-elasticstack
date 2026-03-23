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

package indices

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &dataSource{}
	_ datasource.DataSourceWithConfigure = &dataSource{}
)

// NewDataSource is a helper function to simplify the provider implementation.
func NewDataSource() datasource.DataSource {
	return &dataSource{}
}

// dataSource is the data source implementation.
type dataSource struct {
	client clients.APIClient
}

// Metadata returns the data source type name.
func (d *dataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_elasticsearch_indices"
}

// Configure adds the provider configured client to the data source.
func (d *dataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	apiClient, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	d.client = *apiClient
}
