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

package agentbuilderagent

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/agentbuilder"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func createAgent(ctx context.Context, client *clients.KibanaScopedClient, req entitycore.KibanaWriteRequest[agentModel]) (entitycore.KibanaWriteResult[agentModel], diag.Diagnostics) {
	plan := req.Plan
	var diags diag.Diagnostics

	supportsSkillIDs, d := client.EnforceMinVersion(ctx, agentbuilder.MinExtendedAPIVersion)
	diags.Append(d...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[agentModel]{}, diags
	}

	body, d := plan.toAPICreateModel(ctx, supportsSkillIDs)
	diags.Append(d...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[agentModel]{}, diags
	}

	oapiClient := client.GetKibanaOapiClient()

	_, d = kibanaoapi.CreateAgent(ctx, oapiClient, req.SpaceID, body)
	diags.Append(d...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[agentModel]{}, diags
	}

	// SpaceID is set explicitly so the returned model carries the resolved
	// space for the envelope's read-after-write step.
	plan.SpaceID = types.StringValue(req.SpaceID)

	return entitycore.KibanaWriteResult[agentModel]{Model: plan}, diags
}

func readAgent(ctx context.Context, client *clients.KibanaScopedClient, resourceID string, spaceID string, prior agentModel) (agentModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	oapiClient := client.GetKibanaOapiClient()

	agent, d := kibanaoapi.GetAgent(ctx, oapiClient, spaceID, resourceID)
	diags.Append(d...)
	if diags.HasError() {
		return prior, false, diags
	}

	if agent == nil {
		return prior, false, diags
	}

	diags.Append(prior.populateFromAPI(ctx, spaceID, agent)...)
	return prior, true, diags
}

func updateAgent(ctx context.Context, client *clients.KibanaScopedClient, req entitycore.KibanaWriteRequest[agentModel]) (entitycore.KibanaWriteResult[agentModel], diag.Diagnostics) {
	plan := req.Plan
	var diags diag.Diagnostics

	supportsSkillIDs, d := client.EnforceMinVersion(ctx, agentbuilder.MinExtendedAPIVersion)
	diags.Append(d...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[agentModel]{}, diags
	}

	body, d := plan.toAPIUpdateModel(ctx, supportsSkillIDs)
	diags.Append(d...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[agentModel]{}, diags
	}

	oapiClient := client.GetKibanaOapiClient()

	_, d = kibanaoapi.UpdateAgent(ctx, oapiClient, req.SpaceID, req.WriteID, body)
	diags.Append(d...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[agentModel]{}, diags
	}

	// SpaceID is set explicitly so the returned model carries the resolved
	// space for the envelope's read-after-write step.
	plan.SpaceID = types.StringValue(req.SpaceID)

	return entitycore.KibanaWriteResult[agentModel]{Model: plan}, diags
}

func deleteAgent(ctx context.Context, client *clients.KibanaScopedClient, resourceID string, spaceID string, _ agentModel) diag.Diagnostics {
	oapiClient := client.GetKibanaOapiClient()
	return kibanaoapi.DeleteAgent(ctx, oapiClient, spaceID, resourceID)
}
