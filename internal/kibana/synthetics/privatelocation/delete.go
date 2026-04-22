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

	compositeID, dg := tryReadCompositeID(resourceID)
	response.Diagnostics.Append(dg...)
	if response.Diagnostics.HasError() {
		return
	}

	if compositeID != nil {
		resourceID = compositeID.ResourceID
	}

	spaceID := effectiveSpaceID(plan.SpaceID, compositeID)

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

	dg = kibanaoapi.DeletePrivateLocation(ctx, kibanaClient, spaceID, resourceID)
	response.Diagnostics.Append(dg...)
}
