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

package ilm

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// updateILM is the envelope Update callback. It expands the plan into a
// Policy, applies version-gating, and PUTs the ILM policy. The ILM PUT is
// idempotent for both create and update. The envelope invokes readILM after
// this returns and sets state from the read result.
func updateILM(ctx context.Context, client *clients.ElasticsearchScopedClient, req entitycore.WriteRequest[tfModel]) (entitycore.WriteResult[tfModel], diag.Diagnostics) {
	var diags diag.Diagnostics
	plan := req.Plan

	sv, sdkDiags := client.ServerVersion(ctx)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return entitycore.WriteResult[tfModel]{}, diags
	}

	policy, policyDiags := policyFromModel(ctx, &plan, sv)
	diags.Append(policyDiags...)
	if diags.HasError() {
		return entitycore.WriteResult[tfModel]{}, diags
	}
	policy.Name = plan.Name.ValueString()

	diags.Append(elasticsearch.PutIlm(ctx, client, policy)...)
	if diags.HasError() {
		return entitycore.WriteResult[tfModel]{}, diags
	}

	return entitycore.WriteResult[tfModel]{Model: plan}, diags
}
