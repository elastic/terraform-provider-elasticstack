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
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func readDataSource(ctx context.Context, kbClient *clients.KibanaScopedClient, resourceID string, spaceID string, config enrollmentTokensModel) (enrollmentTokensModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	fleetClient := kbClient.GetFleetClient()

	var tokens []kbapi.EnrollmentApiKey
	policyID := resourceID
	if policyID == "_" {
		policyID = ""
	}

	switch {
	case policyID == "":
		tokens, diags = fleet.GetEnrollmentTokens(ctx, fleetClient, spaceID)
	case spaceID != "" && spaceID != "default":
		tokens, diags = fleet.GetEnrollmentTokensByPolicyInSpace(ctx, fleetClient, policyID, spaceID)
	default:
		tokens, diags = fleet.GetEnrollmentTokensByPolicy(ctx, fleetClient, policyID)
	}
	if diags.HasError() {
		return config, false, diags
	}

	if policyID != "" {
		config.ID = types.StringValue(policyID)
		config.PolicyID = types.StringValue(policyID)
	} else {
		hash, err := typeutils.StringToHash(fleetClient.URL)
		if err != nil {
			diags.AddError(err.Error(), "")
			return config, false, diags
		}
		config.ID = types.StringPointerValue(hash)
	}
	pDiags := (&config).populateFromAPI(ctx, tokens)
	diags.Append(pDiags...)
	if diags.HasError() {
		return config, false, diags
	}

	return config, true, diags
}
