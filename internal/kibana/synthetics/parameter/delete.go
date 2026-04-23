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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// minKibanaPerIDDeleteVersion is the minimum Kibana version that supports
// DELETE /api/synthetics/params/{id}. Earlier versions return 404 for that
// endpoint; use DELETE /api/synthetics/params with {"ids":[...]} instead.
var minKibanaPerIDDeleteVersion = version.Must(version.NewVersion("8.17.0"))

func (r *Resource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var plan tfModelV0
	diags := request.State.Get(ctx, &plan)
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

	resourceID := plan.ID.ValueString()

	compositeID, dg := synthetics.TryReadCompositeID(resourceID)
	response.Diagnostics.Append(dg...)
	if response.Diagnostics.HasError() {
		return
	}

	if compositeID != nil {
		resourceID = compositeID.ResourceID
	}

	// Choose delete endpoint based on Kibana version.
	// DELETE /api/synthetics/params/{id} (DeleteParameterWithResponse) is only
	// supported on Kibana >= 8.17.0; it returns 404 on 8.12.x–8.16.x.
	// DELETE /api/synthetics/params with {"ids":[...]} body (DeleteSyntheticsParamsWithResponse)
	// works on all supported versions (>= 8.12.0), so that is used for older versions.
	kibanaVersion, sdkDiags := apiClient.ServerVersion(ctx)
	response.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if response.Diagnostics.HasError() {
		return
	}

	if kibanaVersion != nil && !kibanaVersion.LessThan(minKibanaPerIDDeleteVersion) {
		// Use the per-ID delete endpoint on Kibana >= 8.17.0.
		deleteResult, err := kibanaClient.API.DeleteParameterWithResponse(ctx, resourceID)
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("Failed to delete parameter `%s`", resourceID), err.Error())
			return
		}
		if deleteResult.StatusCode() != 200 {
			response.Diagnostics.AddError(
				fmt.Sprintf("Unexpected status deleting parameter `%s`", resourceID),
				fmt.Sprintf("API returned status %s", deleteResult.Status()),
			)
		}
		return
	}

	// Use DELETE /api/synthetics/params with {"ids":[...]} body for Kibana < 8.17.0.
	ids := []string{resourceID}
	deleteResult, err := kibanaClient.API.DeleteSyntheticsParamsWithResponse(ctx, kbapi.DeleteSyntheticsParamsJSONRequestBody{
		Ids: &ids,
	})
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("Failed to delete parameter `%s`", resourceID), err.Error())
		return
	}

	if deleteResult.StatusCode() != 200 {
		response.Diagnostics.AddError(
			fmt.Sprintf("Unexpected status deleting parameter `%s`", resourceID),
			fmt.Sprintf("API returned status %s", deleteResult.Status()),
		)
		return
	}

	// Validate that the requested id was actually deleted.
	if deleteResult.JSON200 != nil {
		for _, r := range *deleteResult.JSON200 {
			if r.Id != nil && *r.Id == resourceID {
				if r.Deleted == nil || !*r.Deleted {
					response.Diagnostics.AddError(
						fmt.Sprintf("Parameter `%s` was not deleted", resourceID),
						"Kibana returned deleted=false for the requested parameter id",
					)
				}
				return
			}
		}
		// The response did not include our id — treat as an unexpected error.
		response.Diagnostics.AddError(
			fmt.Sprintf("Parameter `%s` not found in delete response", resourceID),
			"Kibana delete response did not include the requested parameter id",
		)
	}
}
