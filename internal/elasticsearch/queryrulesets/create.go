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

package queryrulesets

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// upsertQueryRuleset handles both Create and Update. The Elasticsearch query ruleset
// PUT API is idempotent, so the same callback serves both lifecycle methods.
func upsertQueryRuleset(
	ctx context.Context,
	client *clients.ElasticsearchScopedClient,
	req entitycore.WriteRequest[QueryRulesetData],
) (entitycore.WriteResult[QueryRulesetData], diag.Diagnostics) {
	var diags diag.Diagnostics
	data := req.Plan
	rulesetID := req.WriteID

	rules := data.toAPIRules(ctx, &diags)
	if diags.HasError() {
		return entitycore.WriteResult[QueryRulesetData]{Model: data}, diags
	}

	if putDiags := elasticsearch.PutQueryRuleset(ctx, client, rulesetID, rules); putDiags.HasError() {
		diags.Append(putDiags...)
		return entitycore.WriteResult[QueryRulesetData]{Model: data}, diags
	}

	id, idDiags := client.ID(ctx, rulesetID)
	diags.Append(idDiags...)
	if diags.HasError() {
		return entitycore.WriteResult[QueryRulesetData]{Model: data}, diags
	}

	data.ID = types.StringValue(id.String())

	return entitycore.WriteResult[QueryRulesetData]{Model: data}, diags
}
