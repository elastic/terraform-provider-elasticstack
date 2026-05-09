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

package parameter

import (
	"context"
	"fmt"
	"net/http"

	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *Resource) readState(ctx context.Context, kibanaClient *kibanaoapi.Client, resourceID string, kibanaConnection types.List, state *tfsdk.State, diagnostics *diag.Diagnostics) {
	getResult, err := kibanaClient.API.GetParameterWithResponse(ctx, resourceID)
	if err != nil {
		diagnostics.AddError(fmt.Sprintf("Failed to get parameter `%s`", resourceID), err.Error())
		return
	}

	if getResult.StatusCode() == http.StatusNotFound {
		state.RemoveResource(ctx)
		return
	}

	unwrapped, unwrapDiags := diagutil.UnwrapJSON200(getResult.JSON200, "synthetics parameter")
	diagnostics.Append(unwrapDiags...)
	if diagnostics.HasError() {
		return
	}

	model := modelV0FromOAPI(*unwrapped)
	model.KibanaConnection = kibanaConnection

	// Set refreshed state
	diags := state.Set(ctx, &model)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return
	}
}

func (r *Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state tfModelV0
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	apiClient, diags := r.Client().GetKibanaClient(ctx, state.KibanaConnection)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	kibanaClient := synthetics.GetKibanaOAPIClientFromScopedClient(apiClient, response.Diagnostics)
	if kibanaClient == nil {
		return
	}

	resourceID := state.ID.ValueString()

	compositeID, dg := synthetics.TryReadCompositeID(resourceID)
	response.Diagnostics.Append(dg...)
	if response.Diagnostics.HasError() {
		return
	}

	if compositeID != nil {
		resourceID = compositeID.ResourceID
	}

	r.readState(ctx, kibanaClient, resourceID, state.KibanaConnection, &response.State, &response.Diagnostics)
}
