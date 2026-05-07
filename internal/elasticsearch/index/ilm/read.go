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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// readILM is the envelope read callback. It fetches the ILM policy from Elasticsearch,
// maps it into tfModel, copies ID and ElasticsearchConnection from the prior state, and
// returns (model, true, nil). Returns (_, false, nil) when the policy is not found.
func readILM(ctx context.Context, client *clients.ElasticsearchScopedClient, policyName string, prior tfModel) (tfModel, bool, diag.Diagnostics) {
	ilmDef, diags := elasticsearch.GetIlm(ctx, client, policyName)
	if diags.HasError() {
		return tfModel{}, false, diags
	}
	if ilmDef == nil {
		tflog.Warn(ctx, "ILM policy not found during read, removing from state", map[string]any{"policy_name": policyName})
		return tfModel{}, false, diags
	}

	out, diags := readPolicyIntoModel(ctx, ilmDef, &prior, policyName)
	if diags.HasError() {
		return tfModel{}, false, diags
	}

	out.ID = prior.ID
	out.ElasticsearchConnection = prior.ElasticsearchConnection

	return *out, true, diags
}
