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

package agentdownloadsource

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	fleetutils "github.com/elastic/terraform-provider-elasticstack/internal/fleet"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state model

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiClient, apiClientDiags := r.Client().GetKibanaClient(ctx, state.KibanaConnection)
	resp.Diagnostics.Append(apiClientDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := apiClient.GetFleetClient()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}
	resp.Diagnostics.Append(r.assertVersionSupported(ctx, apiClient)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sourceID := state.SourceID.ValueString()

	// Read the existing spaces from state to determine where to delete.
	spaceID, diags := fleetutils.GetOperationalSpaceFromState(ctx, req.State)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If deleting the default download source, promote another source to default
	// first so Fleet is never left without one. Without this, concurrent agent
	// policy creates fail with HTTP 400 "Default download source host is not setup".
	if state.Default.ValueBool() {
		listResp, listDiags := fleet.ListAgentDownloadSources(ctx, client, spaceID)
		resp.Diagnostics.Append(listDiags...)
		if !resp.Diagnostics.HasError() && listResp.JSON200 != nil {
			for _, s := range listResp.JSON200.Items {
				if s.Id == sourceID {
					continue
				}
				isDefault := true
				updateReq := kbapi.PutFleetAgentDownloadSourcesSourceidJSONRequestBody{
					Host:      s.Host,
					Name:      s.Name,
					IsDefault: &isDefault,
					ProxyId:   s.ProxyId,
				}
				_, updateDiags := fleet.UpdateAgentDownloadSource(ctx, client, s.Id, spaceID, updateReq)
				resp.Diagnostics.Append(updateDiags...)
				break
			}
		}
		if resp.Diagnostics.HasError() {
			return
		}
	}

	diags = fleet.DeleteAgentDownloadSource(ctx, client, sourceID, spaceID)
	resp.Diagnostics.Append(diags...)
}
