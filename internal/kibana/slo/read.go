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
	clientkibana "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
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

	if r.client == nil {
		response.Diagnostics.AddError("Provider not configured", "Expected configured API client")
		return
	}

	exists, diags := r.readSloFromAPI(ctx, &state)
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

func (r *Resource) readSloFromAPI(ctx context.Context, state *tfModel) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	compID, idDiags := clients.CompositeIDFromStrFw(state.ID.ValueString())
	diags.Append(idDiags...)
	if diags.HasError() {
		return false, diags
	}

	apiModel, sdkDiags := clientkibana.GetSlo(ctx, r.client, compID.ResourceID, compID.ClusterID)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return false, diags
	}
	if apiModel == nil {
		return false, diags
	}

	state.ID = types.StringValue((&clients.CompositeID{ClusterID: apiModel.SpaceID, ResourceID: apiModel.SloID}).String())
	diags.Append(state.populateFromAPI(apiModel)...)
	if diags.HasError() {
		return true, diags
	}

	return true, diags
}

func (r *Resource) readAndPopulate(ctx context.Context, plan *tfModel, diags *diag.Diagnostics) {
	exists, readDiags := r.readSloFromAPI(ctx, plan)
	diags.Append(readDiags...)
	if diags.HasError() {
		return
	}
	if !exists {
		diags.AddError("SLO not found", "SLO was created/updated but could not be found afterwards")
		return
	}
}
