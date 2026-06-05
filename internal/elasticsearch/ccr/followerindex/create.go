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

package followerindex

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func createFollowerIndex(
	ctx context.Context,
	client *clients.ElasticsearchScopedClient,
	req entitycore.WriteRequest[Model],
) (entitycore.WriteResult[Model], diag.Diagnostics) {
	var diags diag.Diagnostics
	plan := req.Plan
	indexName := req.WriteID

	followReq, buildDiags := buildFollowRequest(plan)
	diags.Append(buildDiags...)
	if diags.HasError() {
		return entitycore.WriteResult[Model]{Model: plan}, diags
	}

	diags.Append(elasticsearch.CreateFollowerIndex(ctx, client, indexName, followReq)...)
	if diags.HasError() {
		return entitycore.WriteResult[Model]{Model: plan}, diags
	}

	// Capture tuning parameters while the follower is still active. Paused followers
	// omit Parameters from GET /_ccr/info, which would leave Computed attrs unknown.
	follower, getDiags := elasticsearch.GetFollowerIndex(ctx, client, indexName)
	diags.Append(getDiags...)
	if diags.HasError() {
		return entitycore.WriteResult[Model]{Model: plan}, diags
	}
	if follower != nil {
		plan = mapFollowerIndexToModel(follower, plan)
	}

	if plan.Status.ValueString() == statusPaused {
		diags.Append(elasticsearch.PauseFollowerIndex(ctx, client, indexName)...)
		if diags.HasError() {
			return entitycore.WriteResult[Model]{Model: plan}, diags
		}
	}

	id, idDiags := client.ID(ctx, indexName)
	diags.Append(idDiags...)
	if diags.HasError() {
		return entitycore.WriteResult[Model]{Model: plan}, diags
	}

	plan.ID = types.StringValue(id.String())

	return entitycore.WriteResult[Model]{Model: plan}, diags
}
