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

package monitor

import (
	"context"
	"fmt"
	"slices"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func kibanaSpacesForUpdate(plan types.List, prior *tfModelV0) types.List {
	if plan.IsNull() && prior != nil && !prior.KibanaSpaces.IsNull() {
		return typeutils.StringsToListMust([]string{})
	}
	return plan
}

func kibanaSpacesForEndpoint(spaces types.List, spaceID string) types.List {
	if !typeutils.IsKnown(spaces) {
		return spaces
	}

	configured := make([]string, 0, len(spaces.Elements()))
	for _, space := range spaces.Elements() {
		configured = append(configured, space.(types.String).ValueString())
	}
	if len(configured) == 0 || slices.Contains(configured, "*") {
		return spaces
	}

	if spaceID == "" {
		spaceID = "default"
	}
	if slices.Contains(configured, spaceID) {
		return spaces
	}

	return typeutils.StringsToListMust(append(configured, spaceID))
}

func updateMonitor(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	req entitycore.KibanaWriteRequest[tfModelV0],
) (entitycore.KibanaWriteResult[tfModelV0], diag.Diagnostics) {
	planModel := req.Plan
	var diags diag.Diagnostics

	diags.Append(planModel.enforceVersionConstraints(ctx, client)...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[tfModelV0]{}, diags
	}

	requestModel := planModel
	requestModel.KibanaSpaces = kibanaSpacesForUpdate(planModel.KibanaSpaces, req.Prior)
	requestModel.KibanaSpaces = kibanaSpacesForEndpoint(requestModel.KibanaSpaces, req.SpaceID)
	input, apiDiags := requestModel.toKibanaAPIRequest(ctx)
	diags.Append(apiDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[tfModelV0]{}, diags
	}

	oapiClient := client.GetKibanaOapiClient()

	spaceID := req.SpaceID
	monitorID := req.WriteID
	result, updateDiags := kibanaoapi.UpdateMonitor(ctx, oapiClient, spaceID, monitorID, *input)
	diags.Append(updateDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[tfModelV0]{}, diags
	}

	if result == nil {
		diags.AddError(
			fmt.Sprintf("Failed to update Kibana monitor `%s`, space %s", planModel.Name.ValueString(), spaceID),
			"empty response from API",
		)
		return entitycore.KibanaWriteResult[tfModelV0]{}, diags
	}

	updatedPlan, modelDiags := planModel.toModelV0(ctx, result, spaceID)
	diags.Append(modelDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[tfModelV0]{}, diags
	}
	planModel = *updatedPlan

	return entitycore.KibanaWriteResult[tfModelV0]{Model: planModel}, diags
}
