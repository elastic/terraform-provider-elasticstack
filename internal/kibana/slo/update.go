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
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
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

	response.Diagnostics.Append(entitycore.EnforceVersionRequirements(ctx, apiClient, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	supportsMultipleGroupBy := resolveGroupBySupport(ctx, apiClient, &apiModel, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	oapi, d := apiClient.GetKibanaOapiClient()
	response.Diagnostics.Append(d...)
	if response.Diagnostics.HasError() {
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
		Tags:            typeutils.SliceRef(apiModel.Tags),
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
