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

package autofollow

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func createAutoFollowPattern(
	ctx context.Context,
	client *clients.ElasticsearchScopedClient,
	req entitycore.WriteRequest[Model],
) (entitycore.WriteResult[Model], diag.Diagnostics) {
	var diags diag.Diagnostics
	plan := req.Plan
	name := req.WriteID

	diags.Append(executeOperations(ctx, client, name, plan, planCreateOperations(plan))...)
	if diags.HasError() {
		return entitycore.WriteResult[Model]{Model: plan}, diags
	}

	id, idDiags := client.ID(ctx, name)
	diags.Append(idDiags...)
	if diags.HasError() {
		return entitycore.WriteResult[Model]{Model: plan}, diags
	}

	plan.ID = types.StringValue(id.String())

	return entitycore.WriteResult[Model]{Model: plan}, diags
}

func executeOperations(
	ctx context.Context,
	client *clients.ElasticsearchScopedClient,
	name string,
	plan Model,
	ops []apiOperation,
) diag.Diagnostics {
	var diags diag.Diagnostics

	for _, op := range ops {
		switch op {
		case opPut:
			putReq, buildDiags := buildPutAutoFollowPatternRequest(ctx, plan)
			diags.Append(buildDiags...)
			if diags.HasError() {
				return diags
			}
			diags.Append(elasticsearch.PutAutoFollowPattern(ctx, client, name, putReq)...)
		case opPause:
			diags.Append(elasticsearch.PauseAutoFollowPattern(ctx, client, name)...)
		case opResume:
			diags.Append(elasticsearch.ResumeAutoFollowPattern(ctx, client, name)...)
		default:
			diags.AddError("Internal error", "Unexpected operation: "+op.String())
		}
		if diags.HasError() {
			return diags
		}
	}

	return diags
}
