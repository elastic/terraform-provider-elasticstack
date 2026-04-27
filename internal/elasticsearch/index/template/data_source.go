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

package template

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// DataSource holds the provider client factory from Configure; Elasticsearch
// access uses the same pattern as other PF data sources in this provider.
type DataSource struct {
	client *clients.ProviderClientFactory
}

func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

// Configure adds the provider-configured client factory to the data source.
func (d *DataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	_ = ctx
	// Add a nil check when handling ProviderData because Terraform sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	factory, diags := clients.ConvertProviderDataToFactory(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	d.client = factory
}

func (d *DataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_elasticsearch_index_template"
}

func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg Model
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, diags := d.client.GetElasticsearchClient(ctx, cfg.ElasticsearchConnection)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := cfg.Name.ValueString()
	out, found, diags := readIndexTemplate(ctx, client, name)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		tflog.Info(ctx, fmt.Sprintf(`Index template "%s" not found; leaving data source attributes unset (legacy SDK behavior)`, name))
		empty := Model{
			ElasticsearchConnection: cfg.ElasticsearchConnection,
			Name:                    cfg.Name,
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &empty)...)
		return
	}

	id, sdkDiags := client.ID(ctx, name)
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out.ElasticsearchConnection = cfg.ElasticsearchConnection
	out.Name = types.StringValue(name)
	out.ID = types.StringValue(id.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}

var (
	_ datasource.DataSource              = &DataSource{}
	_ datasource.DataSourceWithConfigure = &DataSource{}
)
