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

package alertingrules

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ruleAPI is the Kibana surface the read path uses. Tests substitute it so a
// search can be shown not to call the single-rule fetch with the sentinel.
type ruleAPI struct {
	get  func(context.Context, *kibanaoapi.Client, string, string) (*models.AlertingRule, diag.Diagnostics)
	find func(context.Context, *kibanaoapi.Client, string, *string) ([]models.AlertingRule, diag.Diagnostics)
}

func readAlertingRulesDataSource(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	resourceID string,
	spaceID string,
	config alertingRulesDataSourceModel,
) (alertingRulesDataSourceModel, bool, diag.Diagnostics) {
	return readRules(ctx, client.GetKibanaOapiClient(), resourceID, spaceID, config, ruleAPI{
		get:  kibanaoapi.GetAlertingRule,
		find: kibanaoapi.FindAlertingRules,
	})
}

// readRules chooses lookup or search from the configured rule_id. The envelope
// identity argument is ignored: during search it is the sentinel, and a
// composite-looking rule_id must be fetched as a literal ID.
func readRules(
	ctx context.Context,
	apiClient *kibanaoapi.Client,
	_ string,
	spaceID string,
	config alertingRulesDataSourceModel,
	api ruleAPI,
) (alertingRulesDataSourceModel, bool, diag.Diagnostics) {
	spaceID = clients.EffectiveSpaceID(spaceID)
	config.SpaceID = types.StringValue(spaceID)

	if ruleIDConfigured(config.RuleID) {
		rule, diags := api.get(ctx, apiClient, spaceID, config.RuleID.ValueString())
		if diags.HasError() || rule == nil {
			return config, false, diags
		}
		config.ID = clients.CompositeIDValue(spaceID, config.RuleID.ValueString())
		diags = config.setRules(ctx, []models.AlertingRule{*rule})
		return config, !diags.HasError(), diags
	}

	rules, diags := api.find(ctx, apiClient, spaceID, typeutils.ValueStringPointer(config.Filter))
	if diags.HasError() {
		return config, false, diags
	}
	config.ID = types.StringValue(spaceID)
	diags = config.setRules(ctx, rules)
	return config, !diags.HasError(), diags
}
