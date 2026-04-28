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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan tfModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	if r.Client() == nil {
		response.Diagnostics.AddError("Provider not configured", "Expected configured provider client factory")
		return
	}

	apiClient, diags := r.Client().GetKibanaClient(ctx, plan.KibanaConnection)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	apiModel, diags := plan.toAPIModel()
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	serverVersion, sdkDiags := apiClient.ServerVersion(ctx)
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

	oapi, err := apiClient.GetKibanaOapiClient()
	if err != nil {
		response.Diagnostics.AddError("Failed to get Kibana API client", err.Error())
		return
	}

	indicator, convErr := kibanaoapi.ResponseIndicatorToUpdateIndicator(apiModel.Indicator)
	if convErr != nil {
		response.Diagnostics.AddError("Failed to convert indicator", convErr.Error())
		return
	}

	groupBy := kibanaoapi.TransformGroupBy(apiModel.GroupBy, supportsMultipleGroupBy)
	reqModel := kbapi.SLOsUpdateSloRequest{
		Name:            &apiModel.Name,
		Description:     &apiModel.Description,
		Indicator:       &indicator,
		TimeWindow:      &apiModel.TimeWindow,
		BudgetingMethod: &apiModel.BudgetingMethod,
		Objective:       &apiModel.Objective,
		Settings:        apiModel.Settings,
		GroupBy:         groupBy,
		Tags:            kibanaoapi.TagsToPtr(apiModel.Tags),
		Artifacts:       apiModel.Artifacts,
	}

	desiredEnabled := plan.Enabled

	fwDiags := kibanaoapi.UpdateSlo(ctx, oapi, apiModel.SpaceID, apiModel.SloID, reqModel)
	response.Diagnostics.Append(fwDiags...)
	if response.Diagnostics.HasError() {
		return
	}

	r.readAndPopulate(ctx, apiClient, &plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	r.reconcileSloEnabledAfterWrite(ctx, apiClient, oapi, apiModel.SpaceID, apiModel.SloID, desiredEnabled, &plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, plan)...)
}
