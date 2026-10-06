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
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func writeAlias(ctx context.Context, client *clients.ElasticsearchScopedClient, req entitycore.WriteRequest[tfModel]) (entitycore.WriteResult[tfModel], diag.Diagnostics) {
	var diags diag.Diagnostics
	plan := req.Plan

	diags.Append(plan.Validate(ctx)...)
	if diags.HasError() {
		return entitycore.WriteResult[tfModel]{Model: plan}, diags
	}

	if req.Prior == nil {
		id, idDiags := client.ID(ctx, req.WriteID)
		diags.Append(idDiags...)
		if diags.HasError() {
			return entitycore.WriteResult[tfModel]{Model: plan}, diags
		}
		plan.ID = basetypes.NewStringValue(id.String())
	}

	currentIndices, readDiags := elasticsearch.GetAlias(ctx, client, req.WriteID)
	diags.Append(readDiags...)
	if diags.HasError() {
		return entitycore.WriteResult[tfModel]{Model: plan}, diags
	}

	currentConfigs, currentDiags := currentAliasConfigs(req.WriteID, currentIndices)
	diags.Append(currentDiags...)
	if diags.HasError() {
		return entitycore.WriteResult[tfModel]{Model: plan}, diags
	}

	resolveIndexExpression := func(ctx context.Context, expression string) (elasticsearch.ResolvedIndexTargets, diag.Diagnostics) {
		return elasticsearch.ResolveIndexExpression(ctx, client, expression, req.WriteID)
	}
	actions, desiredEmpty, actionDiags := plan.buildResolvedAliasActionsWithOutcome(ctx, req.WriteID, currentConfigs, resolveIndexExpression)
	diags.Append(actionDiags...)
	if diags.HasError() {
		return entitycore.WriteResult[tfModel]{Model: plan}, diags
	}

	if len(actions) > 0 {
		diags.Append(elasticsearch.UpdateAliasesAtomic(ctx, client, actions)...)
		if diags.HasError() {
			return entitycore.WriteResult[tfModel]{Model: plan}, diags
		}
	}

	if desiredEmpty {
		diags.Append(plan.markDesiredEmptyAfterWrite(ctx)...)
		if diags.HasError() {
			return entitycore.WriteResult[tfModel]{Model: plan}, diags
		}
	}

	return entitycore.WriteResult[tfModel]{Model: plan}, diags
}
