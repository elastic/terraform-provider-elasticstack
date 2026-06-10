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
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func updateFollowerIndex(
	ctx context.Context,
	client *clients.ElasticsearchScopedClient,
	req entitycore.WriteRequest[Model],
) (entitycore.WriteResult[Model], diag.Diagnostics) {
	var diags diag.Diagnostics
	plan := req.Plan
	prior := req.Prior
	indexName := req.WriteID

	if prior == nil {
		diags.AddError("Internal error", "Update requires prior state")
		return entitycore.WriteResult[Model]{Model: plan}, diags
	}

	diags.Append(executeUpdateOperations(ctx, client, indexName, plan, planUpdateOperations(*prior, plan))...)
	if diags.HasError() {
		return entitycore.WriteResult[Model]{Model: plan}, diags
	}

	return entitycore.WriteResult[Model]{Model: plan}, diags
}

func executeUpdateOperations(
	ctx context.Context,
	client *clients.ElasticsearchScopedClient,
	indexName string,
	plan Model,
	ops []apiOperation,
) diag.Diagnostics {
	var diags diag.Diagnostics

	for _, op := range ops {
		switch op {
		case opPause:
			diags.Append(elasticsearch.PauseFollowerIndex(ctx, client, indexName)...)
		case opUpdateSettings:
			if !typeutils.IsKnown(plan.SettingsRaw) {
				continue
			}
			settings, settingsDiags := parseSettingsRawForUpdate(plan.SettingsRaw.ValueString())
			diags.Append(settingsDiags...)
			if diags.HasError() {
				return diags
			}
			diags.Append(elasticsearch.UpdateIndexSettings(ctx, client, indexName, settings)...)
		case opResume:
			resumeReq := buildResumeFollowRequest(plan)
			diags.Append(elasticsearch.ResumeFollowerIndex(ctx, client, indexName, resumeReq)...)
		default:
			diags.AddError("Internal error", "Unexpected update operation: "+op.String())
		}
		if diags.HasError() {
			return diags
		}
	}

	return diags
}
