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

package spaces

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func updateSpace(ctx context.Context, client *clients.KibanaScopedClient, resourceID, _ string, plan, _ resourceModel) (resourceModel, diag.Diagnostics) {
	oapiClient, err := client.GetKibanaOapiClient()
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError("unable to get Kibana OpenAPI client", err.Error())
		return plan, diags
	}

	features, d := disabledFeaturesSlice(ctx, plan.DisabledFeatures)
	if d.HasError() {
		return plan, d
	}

	body := kbapi.PutSpacesSpaceIdJSONRequestBody{
		Id:               resourceID,
		Name:             plan.Name.ValueString(),
		Description:      optionalStringPtr(plan.Description),
		DisabledFeatures: &features,
		Initials:         optionalStringPtr(plan.Initials),
		Color:            optionalStringPtr(plan.Color),
		ImageUrl:         optionalStringPtr(plan.ImageURL),
	}
	if sol := solutionForPutBody(plan.Solution); sol != nil {
		body.Solution = sol
	}

	_, sdkDiags := kibanaoapi.UpdateSpace(ctx, oapiClient, resourceID, body)
	diags := diagutil.FrameworkDiagsFromSDK(sdkDiags)
	if diags.HasError() {
		return plan, diags
	}

	space, found, fwDiags := fetchSpace(ctx, oapiClient, resourceID)
	diags.Append(fwDiags...)
	if diags.HasError() {
		return plan, diags
	}
	if !found {
		diags.AddError("Update space", "space was not found after update")
		return plan, diags
	}

	return finalizeResourceModelFromAPIResponse(ctx, plan, space)
}
