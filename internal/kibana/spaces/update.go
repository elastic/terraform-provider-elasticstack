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
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func updateSpace(ctx context.Context, client *clients.KibanaScopedClient, req entitycore.KibanaWriteRequest[resourceModel]) (entitycore.KibanaWriteResult[resourceModel], diag.Diagnostics) {
	plan := req.Plan
	oapiClient, err := client.GetKibanaOapiClient()
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError("unable to get Kibana OpenAPI client", err.Error())
		return entitycore.KibanaWriteResult[resourceModel]{Model: plan}, diags
	}

	features, d := disabledFeaturesSlice(ctx, plan.DisabledFeatures)
	if d.HasError() {
		return entitycore.KibanaWriteResult[resourceModel]{Model: plan}, d
	}

	body := kbapi.PutSpacesSpaceIdJSONRequestBody{
		Id:               req.WriteID,
		Name:             plan.Name.ValueString(),
		Description:      typeutils.OptStringPtr(plan.Description),
		DisabledFeatures: &features,
		Initials:         typeutils.OptStringPtr(plan.Initials),
		Color:            typeutils.OptStringPtr(plan.Color),
		ImageUrl:         typeutils.OptStringPtr(plan.ImageURL),
	}
	if sol := solutionForPutBody(plan.Solution); sol != nil {
		body.Solution = sol
	}

	_, apiDiags := kibanaoapi.UpdateSpace(ctx, oapiClient, req.WriteID, body)
	var diags diag.Diagnostics
	diags.Append(apiDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[resourceModel]{Model: plan}, diags
	}

	// The envelope's read-after-write step refreshes the model and surfaces
	// "not found" when the space disappears between update and read. Only the
	// resource identity needs to be set here so the read can resolve it.
	plan.ID = plan.SpaceID
	return entitycore.KibanaWriteResult[resourceModel]{Model: plan}, diags
}
