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

// Temporary compatibility wrappers for internal/fleet/agentlesspolicy callers until
// task 8 swaps them to managed_integration.go. Delete this file with task 8.

package fleet

import (
	"context"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// CreateAgentlessPolicy creates a new Fleet agentless policy via the deprecated
// POST /api/fleet/agentless_policies endpoint.
func CreateAgentlessPolicy(
	ctx context.Context,
	client *Client,
	spaceID string,
	body kbapi.PostFleetAgentlessPoliciesJSONRequestBody,
) (*kbapi.KibanaHTTPAPIsManagedIntegration, diag.Diagnostics) {
	return kibanautil.ConflictRetry(ctx, kibanautil.ConflictMaxAttempts, func() (*kbapi.KibanaHTTPAPIsManagedIntegration, int, diag.Diagnostics) {
		resp, err := client.API.PostFleetAgentlessPoliciesWithResponse(ctx, body, kibanautil.SpaceAwarePathRequestEditor(spaceID))
		if err != nil {
			return nil, 0, diagutil.FrameworkDiagFromError(err)
		}

		result, diags := kibanaoapi.HandleMutateTypedResponse(resp.StatusCode(), resp.Body, func() *kbapi.KibanaHTTPAPIsManagedIntegration {
			if resp.JSON200 == nil {
				return nil
			}
			return &resp.JSON200.Item
		})
		return result, resp.StatusCode(), diags
	})
}

// ReadAgentlessPolicyViaPackagePolicy reads an agentless policy via the
// package_policies fallback. Returns (nil, nil) on HTTP 404.
func ReadAgentlessPolicyViaPackagePolicy(
	ctx context.Context,
	client *Client,
	spaceID, policyID string,
) (*kbapi.PackagePolicy, diag.Diagnostics) {
	return GetPackagePolicy(ctx, client, policyID, spaceID)
}

// UpdateAgentlessPolicyViaPackagePolicy updates an agentless policy via the
// package_policies fallback.
func UpdateAgentlessPolicyViaPackagePolicy(
	ctx context.Context,
	client *Client,
	spaceID, policyID string,
	body kbapi.PackagePolicyRequest,
) (*kbapi.PackagePolicy, diag.Diagnostics) {
	return UpdatePackagePolicy(ctx, client, policyID, spaceID, body)
}

// DeleteAgentlessPolicy deletes an agentless policy via the deprecated
// DELETE /api/fleet/agentless_policies/{id} endpoint. Semantics match
// DeleteManagedIntegration: the returned bool reflects the final HTTP status
// observed across retries; transport errors reset it to false.
func DeleteAgentlessPolicy(
	ctx context.Context,
	client *Client,
	spaceID, policyID string,
	force bool,
) (isConflict bool, diags diag.Diagnostics) {
	params := kbapi.DeleteFleetAgentlessPoliciesPolicyidParams{}
	if force {
		t := true
		params.Force = &t
	}

	var lastStatusCode int
	_, diags = kibanautil.ConflictRetry(ctx, kibanautil.ConflictMaxAttempts, func() (struct{}, int, diag.Diagnostics) {
		resp, err := client.API.DeleteFleetAgentlessPoliciesPolicyidWithResponse(ctx, policyID, &params, kibanautil.SpaceAwarePathRequestEditor(spaceID))
		if err != nil {
			lastStatusCode = 0
			return struct{}{}, 0, diagutil.FrameworkDiagFromError(err)
		}
		lastStatusCode = resp.StatusCode()
		return struct{}{}, resp.StatusCode(), handleDeleteResponse(resp.StatusCode(), resp.Body)
	})
	return lastStatusCode == http.StatusConflict, diags
}
