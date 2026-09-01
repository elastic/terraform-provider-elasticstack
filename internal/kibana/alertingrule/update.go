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
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func updateAlertingRule(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	req entitycore.KibanaWriteRequest[alertingRuleModel],
) (entitycore.KibanaWriteResult[alertingRuleModel], diag.Diagnostics) {
	return entitycore.SimpleKibanaUpdate[alertingRuleModel, models.AlertingRule, models.AlertingRule](
		func(plan alertingRuleModel, ctx context.Context, writeID string) (models.AlertingRule, diag.Diagnostics) {
			// Convert to API model, then ensure rule ID and space ID are set from state.
			rule, diags := plan.toAPIModel(ctx)
			if diags.HasError() {
				return rule, diags
			}
			rule.RuleID = writeID
			rule.SpaceID = req.SpaceID
			return rule, diags
		},
		func(ctx context.Context, oapiClient *kibanaoapi.Client, spaceID, _ string, rule models.AlertingRule) (*models.AlertingRule, diag.Diagnostics) {
			return kibanaoapi.UpdateAlertingRule(ctx, oapiClient, spaceID, rule)
		},
		func(plan *alertingRuleModel, ctx context.Context, _ string, _ *models.AlertingRule) diag.Diagnostics {
			// Record the concrete investigation-guide checksum (file-based content)
			// before read-after-write, which preserves it since the API returns no checksum.
			return plan.applyInvestigationGuideChecksum(ctx)
		},
	)(ctx, client, req)
}
