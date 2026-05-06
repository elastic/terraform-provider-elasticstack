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

package apikey

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// readAPIKey is the package-level read callback shared with the envelope and
// the concrete Read override.
func readAPIKey(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, state tfModel) (tfModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	apiKey, apiKeyDiags := elasticsearch.GetAPIKey(ctx, client, resourceID)
	diags.Append(apiKeyDiags...)
	if diags.HasError() {
		return state, false, diags
	}
	if apiKey == nil {
		return state, false, diags
	}

	ver, sdkDiags := client.ServerVersion(ctx)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return state, false, diags
	}

	diags.Append(state.populateFromAPI(apiKey, ver)...)
	if diags.HasError() {
		return state, false, diags
	}

	return state, true, diags
}

// Read overrides the envelope's Read to add private-state cluster-version caching.
func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var stateModel tfModel
	resp.Diagnostics.Append(req.State.Get(ctx, &stateModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	compID, diags := clients.CompositeIDFromStrFw(stateModel.GetID().ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, connDiags := r.Client().GetElasticsearchClient(ctx, stateModel.GetElasticsearchConnection())
	resp.Diagnostics.Append(connDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	finalModel, found, readDiags := readAPIKey(ctx, client, compID.ResourceID, stateModel)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, finalModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(saveClusterVersion(ctx, client, resp.Private)...)
}
