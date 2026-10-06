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

package alias

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// applyResolvedAliasConfig validates the plan, resolves it against the current
// alias state, and atomically applies the resulting alias actions. It is shared
// by createAlias and updateAlias, which differ only in how they obtain the
// resource ID before delegating here.
func applyResolvedAliasConfig(ctx context.Context, client *clients.ElasticsearchScopedClient, plan tfModel, aliasName string) (tfModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	diags.Append(plan.Validate(ctx)...)
	if diags.HasError() {
		return plan, diags
	}

	currentIndices, readDiags := elasticsearch.GetAlias(ctx, client, aliasName)
	diags.Append(readDiags...)
	if diags.HasError() {
		return plan, diags
	}

	currentConfigs, currentDiags := currentAliasConfigs(aliasName, currentIndices)
	diags.Append(currentDiags...)
	if diags.HasError() {
		return plan, diags
	}

	resolveIndexExpression := func(ctx context.Context, expression string) (elasticsearch.ResolvedIndexTargets, diag.Diagnostics) {
		return elasticsearch.ResolveIndexExpression(ctx, client, expression, aliasName)
	}
	actions, desiredEmpty, actionDiags := plan.buildResolvedAliasActionsWithOutcome(ctx, aliasName, currentConfigs, resolveIndexExpression)
	diags.Append(actionDiags...)
	if diags.HasError() {
		return plan, diags
	}

	if len(actions) > 0 {
		diags.Append(elasticsearch.UpdateAliasesAtomic(ctx, client, actions)...)
		if diags.HasError() {
			return plan, diags
		}
	}

	if desiredEmpty {
		diags.Append(plan.markDesiredEmptyAfterWrite(ctx)...)
		if diags.HasError() {
			return plan, diags
		}
	}

	return plan, diags
}
