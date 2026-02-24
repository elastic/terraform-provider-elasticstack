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

package enrollmenttokens

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (d *enrollmentTokensDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model enrollmentTokensModel

	diags := req.Config.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := d.client.GetFleetClient()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	var tokens []kbapi.EnrollmentApiKey
	policyID := model.PolicyID.ValueString()
	spaceID := model.SpaceID.ValueString()

	// Query enrollment tokens with space context if needed
	if policyID == "" {
		tokens, diags = fleet.GetEnrollmentTokens(ctx, client, spaceID)
	} else {
		// Get tokens by policy, with space awareness if specified
		if spaceID != "" && spaceID != "default" {
			tokens, diags = fleet.GetEnrollmentTokensByPolicyInSpace(ctx, client, policyID, spaceID)
		} else {
			tokens, diags = fleet.GetEnrollmentTokensByPolicy(ctx, client, policyID)
		}
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if policyID != "" {
		model.ID = types.StringValue(policyID)
	} else {
		hash, err := schemautil.StringToHash(client.URL)
		if err != nil {
			resp.Diagnostics.AddError(err.Error(), "")
			return
		}
		model.ID = types.StringPointerValue(hash)
	}

	diags = model.populateFromAPI(ctx, tokens)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
}
