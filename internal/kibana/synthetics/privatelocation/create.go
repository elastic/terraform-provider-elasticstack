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

package privatelocation

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {

	var plan tfModelV0
	diags := request.Plan.Get(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	apiClient, diags := r.client.GetKibanaClient(ctx, plan.KibanaConnection)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	kibanaClient := synthetics.GetKibanaOAPIClientFromScopedClient(apiClient, response.Diagnostics)
	if kibanaClient == nil {
		return
	}

	spaceID := plan.SpaceID.ValueString()

	if requiresSpaceIDMinVersion(spaceID) {
		supported, sdkDiags := apiClient.EnforceMinVersion(ctx, MinVersionSpaceID)
		response.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
		if response.Diagnostics.HasError() {
			return
		}
		if !supported {
			response.Diagnostics.AddError(
				"Unsupported server version",
				fmt.Sprintf("Synthetics private locations in a non-default Kibana space require Elastic Stack %s or later.", MinVersionSpaceID),
			)
			return
		}
	}

	// Preserve the planned geo values before the API call. The Kibana API stores geo
	// coordinates as float32 and returns float32-precision values on read (e.g.
	// 42.42 → 42.41999816894531). If we blindly set state from the API response, the
	// state value differs from the plan value, which Terraform rejects as an
	// inconsistent result. We use the planned values (from config) so state matches
	// the plan. The Float32PrecisionType custom type handles subsequent semantic
	// equality checks so that subsequent plans detect no diff.
	plannedGeo := plan.Geo

	body := privateLocationToCreateBody(plan)
	result, dg := kibanaoapi.CreatePrivateLocation(ctx, kibanaClient, spaceID, body)
	response.Diagnostics.Append(dg...)
	if response.Diagnostics.HasError() {
		return
	}

	plan = privateLocationFromAPI(*result, spaceID, plan.KibanaConnection)
	// Restore geo from plan to keep state consistent with what was planned.
	plan.Geo = plannedGeo

	diags = response.State.Set(ctx, plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
}
