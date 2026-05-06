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
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// readMLDatafeedState is the envelope read callback. It reads datafeed stats
// and returns the updated model. During import, computed attributes without
// a stored value are set to their zero/null defaults.
func readMLDatafeedState(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, state MLDatafeedStateData) (MLDatafeedStateData, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Check if the datafeed exists by getting its stats
	datafeedStats, getDiags := elasticsearch.GetDatafeedStats(ctx, client, resourceID)
	diags.Append(getDiags...)
	if diags.HasError() {
		return state, false, diags
	}

	if datafeedStats == nil {
		return state, false, diags
	}

	// Update the data with current information
	state.State = types.StringValue(datafeedStats.State.String())

	// Regenerate composite ID to ensure it's current
	compID, sdkDiags := client.ID(ctx, resourceID)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return state, false, diags
	}

	state.ID = types.StringValue(compID.String())

	diags.Append(state.SetStartAndEndFromAPI(datafeedStats)...)
	if diags.HasError() {
		return state, false, diags
	}

	// Set defaults for computed attributes if they're not already set (e.g., during import)
	if state.Force.IsNull() {
		state.Force = types.BoolValue(false)
	}
	if state.Timeout.IsNull() {
		state.Timeout = customtypes.NewDurationValue("30s")
	}

	return state, true, diags
}

// read is the internal helper used by the update path to read datafeed stats
// given a model that already has the ElasticsearchConnection populated.
func (r *mlDatafeedStateResource) read(ctx context.Context, data MLDatafeedStateData) (*MLDatafeedStateData, diag.Diagnostics) {
	client, diags := r.Client().GetElasticsearchClient(ctx, data.ElasticsearchConnection)
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
	data.State = types.StringValue(datafeedStats.State.String())

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
