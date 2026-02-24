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

package datafeedstate

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *mlDatafeedStateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MLDatafeedStateData
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	readData, diags := r.read(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if readData == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, readData)...)
}

func (r *mlDatafeedStateResource) read(ctx context.Context, data MLDatafeedStateData) (*MLDatafeedStateData, diag.Diagnostics) {
	client, diags := clients.MaybeNewAPIClientFromFrameworkResource(ctx, data.ElasticsearchConnection, r.client)
	if diags.HasError() {
		return nil, diags
	}

	datafeedID := data.DatafeedID.ValueString()
	// Check if the datafeed exists by getting its stats
	datafeedStats, getDiags := elasticsearch.GetDatafeedStats(ctx, client, datafeedID)
	diags.Append(getDiags...)
	if diags.HasError() {
		return nil, diags
	}

	if datafeedStats == nil {
		return nil, diags
	}

	// Update the data with current information
	data.State = types.StringValue(datafeedStats.State)

	// Regenerate composite ID to ensure it's current
	compID, sdkDiags := client.ID(ctx, datafeedID)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return nil, diags
	}

	data.ID = types.StringValue(compID.String())

	diags.Append(data.SetStartAndEndFromAPI(datafeedStats)...)

	return &data, diags
}
