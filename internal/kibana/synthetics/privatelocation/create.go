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

	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {

	kibanaClient := synthetics.GetKibanaClient(r, response.Diagnostics)
	if kibanaClient == nil {
		return
	}

	var plan tfModelV0
	diags := request.Plan.Get(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	input := plan.toPrivateLocationConfig()
	spaceID := plan.SpaceID.ValueString()

	if requiresSpaceIDMinVersion(spaceID) {
		supported, sdkDiags := r.client.EnforceMinVersion(ctx, MinVersionSpaceID)
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

	result, err := kibanaClient.KibanaSynthetics.PrivateLocation.Create(ctx, spaceID, input)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("Failed to create private location `%s`", input.Label), err.Error())
		return
	}

	plan = toModelV0(*result, spaceID)

	diags = response.State.Set(ctx, plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
}
