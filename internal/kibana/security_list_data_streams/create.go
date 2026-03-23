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

package securitylistdatastreams

import (
	"context"

	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *securityListDataStreamsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get Kibana client
	client, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Kibana client", err.Error())
		return
	}

	// Create the list data streams
	spaceID := plan.SpaceID.ValueString()
	_, diags := kibanaoapi.CreateListIndex(ctx, client, spaceID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the data streams to get the actual state
	listIndex, listItemIndex, diags := kibanaoapi.ReadListIndex(ctx, client, spaceID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !listIndex || !listItemIndex {
		resp.Diagnostics.AddError(
			"Failed to verify list data streams",
			"List data streams were created but could not be verified",
		)
		return
	}

	// Populate state using the fromAPIResponse helper method
	plan.fromAPIResponse(spaceID, listIndex, listItemIndex)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
