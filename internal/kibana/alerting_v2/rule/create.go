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

package rule

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/alerting_v2/utils"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const ruleNoun = "rule"

func createRule(ctx context.Context, client *clients.KibanaScopedClient, req entitycore.KibanaWriteRequest[Model]) (entitycore.KibanaWriteResult[Model], diag.Diagnostics) {
	plan := req.Plan
	var diags diag.Diagnostics

	ruleID := plan.RuleID.ValueString()
	if !typeutils.IsKnown(plan.RuleID) {
		ruleID = uuid.NewString()
	}
	importID := clients.CompositeIDValue(req.SpaceID, ruleID).ValueString()

	existing, getDiags := kibanaoapi.GetAlertingV2Rule(ctx, client.GetKibanaOapiClient(), req.SpaceID, ruleID)
	diags.Append(getDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[Model]{}, diags
	}
	if existing != nil {
		diags.Append(utils.AlreadyExistsDiagnostic(ruleNoun, importID))
		return entitycore.KibanaWriteResult[Model]{}, diags
	}

	created, written, putDiags := putRule(ctx, client, req.SpaceID, ruleID, plan)
	diags.Append(putDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[Model]{}, diags
	}
	if !created {
		diags.Append(utils.AlreadyReplacedDiagnostic(ruleNoun, importID))
		return entitycore.KibanaWriteResult[Model]{}, diags
	}

	return entitycore.KibanaWriteResult[Model]{Model: written}, diags
}

// putRule issues the create-or-replace write. created reports whether the server answered 201 (created) rather than 200 (replaced).
func putRule(ctx context.Context, client *clients.KibanaScopedClient, spaceID, ruleID string, plan Model) (created bool, written Model, diags diag.Diagnostics) {
	body, bodyDiags := plan.toAPIPutBody(ctx)
	diags.Append(bodyDiags...)
	if diags.HasError() {
		return false, plan, diags
	}

	created, putDiags := kibanaoapi.PutAlertingV2Rule(ctx, client.GetKibanaOapiClient(), spaceID, ruleID, body)
	diags.Append(putDiags...)
	if diags.HasError() {
		return false, plan, diags
	}

	plan.ID, plan.SpaceID = clients.CompositeIdentityValues(spaceID, ruleID)
	plan.RuleID = types.StringValue(ruleID)
	return created, plan, diags
}
