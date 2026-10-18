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

package agentpolicyds

import (
	"context"
	"fmt"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// The selector is read from config. The envelope's resource id is the "_"
// placeholder when policy_id is unset, and "_" is itself a valid policy_id.
func readDataSource(ctx context.Context, kbClient *clients.KibanaScopedClient, _ string, spaceID string, config agentPolicyModel) (agentPolicyModel, bool, diag.Diagnostics) {
	fleetClient := kbClient.GetFleetClient()

	switch {
	case configuredString(config.PolicyID):
		policy, diags := fleet.GetAgentPolicy(ctx, fleetClient, config.PolicyID.ValueString(), spaceID)
		if diags.HasError() || policy == nil {
			return config, false, diags
		}
		return applyPolicy(ctx, config, policy)
	case configuredString(config.Name):
		return readByName(ctx, fleetClient, spaceID, config)
	default:
		return config, false, nil
	}
}

func readByName(ctx context.Context, fleetClient *fleet.Client, spaceID string, config agentPolicyModel) (agentPolicyModel, bool, diag.Diagnostics) {
	name := config.Name.ValueString()
	matches, diags := fleet.GetAgentPolicyByName(ctx, fleetClient, name, spaceID)
	if diags.HasError() {
		return config, false, diags
	}

	switch len(matches) {
	case 0:
		diags.AddError(
			"Agent policy not found",
			fmt.Sprintf("no agent policy found matching name %q in space %q", name, spaceID),
		)
		return config, false, diags
	case 1:
		// The list item omits empty collections that the get endpoint returns as [].
		policy, getDiags := fleet.GetAgentPolicy(ctx, fleetClient, matches[0].Id, spaceID)
		diags.Append(getDiags...)
		if diags.HasError() {
			return config, false, diags
		}
		if policy == nil {
			diags.AddError(
				"Agent policy not found",
				fmt.Sprintf("no agent policy found matching name %q in space %q", name, spaceID),
			)
			return config, false, diags
		}
		return applyPolicy(ctx, config, policy)
	default:
		diags.AddError(
			"Multiple agent policies match name",
			fmt.Sprintf("multiple agent policies match name %q in space %q: %s", name, spaceID, policyIDList(matches)),
		)
		return config, false, diags
	}
}

func applyPolicy(ctx context.Context, config agentPolicyModel, policy *kbapi.KibanaHTTPAPIsAgentPolicyResponse) (agentPolicyModel, bool, diag.Diagnostics) {
	diags := (&config).populateFromAPI(ctx, policy)
	if diags.HasError() {
		return config, false, diags
	}
	return config, true, nil
}

func policyIDList(matches []*kbapi.KibanaHTTPAPIsAgentPolicyResponse) string {
	parts := make([]string, len(matches))
	for i, match := range matches {
		parts[i] = fmt.Sprintf("policy_id %q", match.Id)
	}
	return strings.Join(parts, ", ")
}

func configuredString(value types.String) bool {
	return typeutils.IsKnown(value) && value.ValueString() != ""
}
