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

package connectors

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// readConnectorFromAPI fetches a connector from the API and populates the given model
// Returns true if the connector was found, false if it doesn't exist
func (r *Resource) readConnectorFromAPI(ctx context.Context, client *clients.APIClient, model *tfModel) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	oapiClient, err := client.GetKibanaOapiClient()
	if err != nil {
		diags.AddError("Failed to get Kibana client", err.Error())
		return false, diags
	}

	compositeID, diagsTemp := model.GetID()
	diags.Append(diagsTemp...)
	if diags.HasError() {
		return false, diags
	}

	connector, diagsTemp := kibanaoapi.GetConnector(ctx, oapiClient, compositeID.ResourceID, compositeID.ClusterID)
	if connector == nil && diagsTemp == nil {
		// Resource not found
		return false, diags
	}
	diags.Append(diagsTemp...)
	if diags.HasError() {
		return false, diags
	}

	diags.Append(model.populateFromAPI(connector, compositeID)...)
	return true, diags
}

func (r *Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state tfModel

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	client, diags := clients.MaybeNewAPIClientFromFrameworkResource(ctx, state.KibanaConnection, r.client)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	exists, diags := r.readConnectorFromAPI(ctx, client, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	if !exists {
		response.State.RemoveResource(ctx)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}
