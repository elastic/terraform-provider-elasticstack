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

package slo

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state tfModel

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if r.Client() == nil {
		response.Diagnostics.AddError("Provider not configured", "Expected configured provider client factory")
		return
	}

	apiClient, diags := r.Client().GetKibanaClient(ctx, state.KibanaConnection)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	exists, diags := r.readSloFromAPI(ctx, apiClient, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	if !exists {
		response.State.RemoveResource(ctx)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func (r *Resource) readSloFromAPI(ctx context.Context, apiClient *clients.KibanaScopedClient, state *tfModel) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	compID, idDiags := clients.CompositeIDFromStrFw(state.ID.ValueString())
	diags.Append(idDiags...)
	if diags.HasError() {
		return false, diags
	}

	oapi, err := apiClient.GetKibanaOapiClient()
	if err != nil {
		diags.AddError("Failed to get Kibana API client", err.Error())
		return false, diags
	}

	// CompositeID stores spaceID as ClusterID and sloID as ResourceID (see create.go).
	res, fwDiags := kibanaoapi.GetSlo(ctx, oapi, compID.ClusterID, compID.ResourceID)
	diags.Append(fwDiags...)
	if diags.HasError() {
		return false, diags
	}
	if res == nil {
		return false, diags
	}

	apiModel := kibanaoapi.SloResponseToModel(compID.ClusterID, res)
	state.ID = types.StringValue((&clients.CompositeID{ClusterID: apiModel.SpaceID, ResourceID: apiModel.SloID}).String())
	diags.Append(state.populateFromAPI(apiModel)...)
	if diags.HasError() {
		return true, diags
	}

	return true, diags
}

func (r *Resource) readAndPopulate(ctx context.Context, apiClient *clients.KibanaScopedClient, plan *tfModel, diags *diag.Diagnostics) {
	exists, readDiags := r.readSloFromAPI(ctx, apiClient, plan)
	diags.Append(readDiags...)
	if diags.HasError() {
		return
	}
	if !exists {
		diags.AddError("SLO not found", "SLO was created/updated but could not be found afterwards")
		return
	}
}
