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
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func updateSlo(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	req entitycore.KibanaWriteRequest[tfModel],
) (entitycore.KibanaWriteResult[tfModel], diag.Diagnostics) {
	planModel := req.Plan
	var diags diag.Diagnostics

	diags.Append(entitycore.EnforceVersionRequirements(ctx, client, &planModel)...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[tfModel]{}, diags
	}

	apiModel, apiDiags := planModel.toAPIModel()
	diags.Append(apiDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[tfModel]{}, diags
	}

	supportsMultipleGroupBy := resolveGroupBySupport(ctx, client, &apiModel, &diags)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[tfModel]{}, diags
	}

	oapi := client.GetKibanaOapiClient()

	indicator, convErr := kibanaoapi.ResponseIndicatorToUpdateIndicator(apiModel.Indicator)
	if convErr != nil {
		diags.AddError("Failed to convert indicator", convErr.Error())
		return entitycore.KibanaWriteResult[tfModel]{}, diags
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

	desiredEnabled := planModel.Enabled

	fwDiags := kibanaoapi.UpdateSlo(ctx, oapi, req.SpaceID, apiModel.SloID, reqModel)
	diags.Append(fwDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[tfModel]{}, diags
	}

	// Read back to populate computed fields.
	readSloAndPopulate(ctx, client, &planModel, &diags)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[tfModel]{}, diags
	}

	reconcileSloEnabled(ctx, client, oapi, req.SpaceID, apiModel.SloID, desiredEnabled, &planModel, &diags)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[tfModel]{}, diags
	}

	return entitycore.KibanaWriteResult[tfModel]{Model: planModel}, diags
}
