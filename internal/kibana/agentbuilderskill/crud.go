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

package agentbuilderskill

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func createSkill(ctx context.Context, client *clients.KibanaScopedClient, req entitycore.KibanaWriteRequest[skillModel]) (entitycore.KibanaWriteResult[skillModel], diag.Diagnostics) {
	plan := req.Plan
	var diags diag.Diagnostics

	body, d := plan.toAPICreateModel(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[skillModel]{}, diags
	}

	oapiClient := client.GetKibanaOapiClient()

	_, d = kibanaoapi.CreateSkill(ctx, oapiClient, req.SpaceID, body)
	diags.Append(d...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[skillModel]{}, diags
	}

	// SpaceID is set explicitly so the returned model carries the resolved
	// space for the envelope's read-after-write step.
	plan.SpaceID = types.StringValue(req.SpaceID)

	return entitycore.KibanaWriteResult[skillModel]{Model: plan}, diags
}

func readSkill(ctx context.Context, client *clients.KibanaScopedClient, resourceID string, spaceID string, prior skillModel) (skillModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	oapiClient := client.GetKibanaOapiClient()

	skill, d := kibanaoapi.GetSkill(ctx, oapiClient, spaceID, resourceID)
	diags.Append(d...)
	if diags.HasError() {
		return prior, false, diags
	}

	if skill == nil {
		return prior, false, diags
	}

	diags.Append(prior.populateFromAPI(ctx, spaceID, skill)...)
	return prior, true, diags
}

func updateSkill(ctx context.Context, client *clients.KibanaScopedClient, req entitycore.KibanaWriteRequest[skillModel]) (entitycore.KibanaWriteResult[skillModel], diag.Diagnostics) {
	plan := req.Plan
	var diags diag.Diagnostics

	body, d := plan.toAPIUpdateModel(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[skillModel]{}, diags
	}

	oapiClient := client.GetKibanaOapiClient()

	_, d = kibanaoapi.UpdateSkill(ctx, oapiClient, req.SpaceID, req.WriteID, body)
	diags.Append(d...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[skillModel]{}, diags
	}

	// SpaceID is set explicitly so the returned model carries the resolved
	// space for the envelope's read-after-write step.
	plan.SpaceID = types.StringValue(req.SpaceID)

	return entitycore.KibanaWriteResult[skillModel]{Model: plan}, diags
}

func deleteSkill(ctx context.Context, client *clients.KibanaScopedClient, resourceID string, spaceID string, _ skillModel) diag.Diagnostics {
	oapiClient := client.GetKibanaOapiClient()
	return kibanaoapi.DeleteSkill(ctx, oapiClient, spaceID, resourceID)
}
