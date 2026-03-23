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

	clientkibana "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan tfModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	if r.client == nil {
		response.Diagnostics.AddError("Provider not configured", "Expected configured API client")
		return
	}

	apiModel, diags := plan.toAPIModel()
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	serverVersion, sdkDiags := r.client.ServerVersion(ctx)
	response.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if response.Diagnostics.HasError() {
		return
	}

	if apiModel.Settings != nil && apiModel.Settings.PreventInitialBackfill != nil {
		if serverVersion.LessThan(SLOSupportsPreventInitialBackfillMinVersion) {
			response.Diagnostics.AddError(
				"Unsupported Elastic Stack version",
				"The 'prevent_initial_backfill' setting requires Elastic Stack version "+SLOSupportsPreventInitialBackfillMinVersion.String()+" or higher.",
			)
			return
		}
	}

	if plan.hasDataViewID() && serverVersion.LessThan(SLOSupportsDataViewIDMinVersion) {
		response.Diagnostics.AddError(
			"Unsupported Elastic Stack version",
			"data_view_id is not supported on Elastic Stack versions < "+SLOSupportsDataViewIDMinVersion.String(),
		)
		return
	}

	supportsGroupBy := serverVersion.GreaterThanOrEqual(SLOSupportsGroupByMinVersion)
	if !supportsGroupBy {
		if len(apiModel.GroupBy) > 0 {
			response.Diagnostics.AddError(
				"Unsupported Elastic Stack version",
				"group_by is not supported in this version of the Elastic Stack. group_by requires "+SLOSupportsGroupByMinVersion.String()+" or higher.",
			)
			return
		}

		// Do not send group_by at all for stacks that don't support it.
		apiModel.GroupBy = nil
	}

	supportsMultipleGroupBy := supportsGroupBy && serverVersion.GreaterThanOrEqual(SLOSupportsMultipleGroupByMinVersion)
	if len(apiModel.GroupBy) > 1 && !supportsMultipleGroupBy {
		response.Diagnostics.AddError(
			"Unsupported Elastic Stack version",
			"multiple group_by fields are not supported in this version of the Elastic Stack. Multiple group_by fields requires "+SLOSupportsMultipleGroupByMinVersion.String(),
		)
		return
	}

	_, sdkDiags = clientkibana.UpdateSlo(ctx, r.client, apiModel, supportsMultipleGroupBy)
	response.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if response.Diagnostics.HasError() {
		return
	}

	r.readAndPopulate(ctx, &plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, plan)...)
}
