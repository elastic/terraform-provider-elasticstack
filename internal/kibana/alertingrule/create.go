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

package alertingrule

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func createAlertingRule(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	req entitycore.KibanaWriteRequest[alertingRuleModel],
) (entitycore.KibanaWriteResult[alertingRuleModel], diag.Diagnostics) {
	m := req.Plan
	var diags diag.Diagnostics

	// Convert to API model
	rule, d := m.toAPIModel(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[alertingRuleModel]{}, diags
	}

	oapiClient := client.GetKibanaOapiClient()

	createdRule, createDiags := kibanaoapi.CreateAlertingRule(ctx, oapiClient, req.SpaceID, rule)
	diags.Append(createDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[alertingRuleModel]{}, diags
	}

	// Set identity fields so the envelope can locate the resource for
	// read-after-write. The envelope will re-read full state.
	m.ID = types.StringValue((&clients.CompositeID{
		ClusterID:  req.SpaceID,
		ResourceID: createdRule.RuleID,
	}).String())
	m.RuleID = types.StringValue(createdRule.RuleID)
	m.SpaceID = types.StringValue(req.SpaceID)

	return entitycore.KibanaWriteResult[alertingRuleModel]{Model: m}, diags
}
