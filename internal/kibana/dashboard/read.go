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

package dashboard

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var stateModel dashboardModel

	diags := req.State.Get(ctx, &stateModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	readModel, diags := r.read(ctx, stateModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if readModel == nil {
		// Dashboard not found, remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, *readModel)...)
}

func (r *Resource) read(ctx context.Context, stateModel dashboardModel) (*dashboardModel, diag.Diagnostics) {
	// Parse composite ID
	composite, diags := clients.CompositeIDFromStrFw(stateModel.ID.ValueString())
	if diags.HasError() {
		return nil, diags
	}

	dashboardID := composite.ResourceID
	spaceID := composite.ClusterID

	// Get the Kibana client
	kibanaClient, err := r.client.GetKibanaOapiClient()
	if err != nil {
		diags.AddError("Unable to get Kibana client", err.Error())
		return nil, diags
	}

	// Get the dashboard
	getResp, getDiags := kibanaoapi.GetDashboard(ctx, kibanaClient, spaceID, dashboardID)
	diags.Append(getDiags...)
	if diags.HasError() {
		return nil, diags
	}

	if getResp == nil {
		return nil, diags
	}

	if getResp.JSON200 == nil {
		diags.AddError("Empty response when getting dashboard", "GET dashboard was successful, however contained an empty response")
		return nil, diags
	}

	diags.Append(stateModel.populateFromAPI(ctx, getResp, dashboardID, spaceID)...)
	return &stateModel, diags
}
