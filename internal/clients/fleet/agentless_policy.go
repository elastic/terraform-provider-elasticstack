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

// CreateAgentlessPolicy creates a new Fleet agentless policy. The endpoint is
// a bundled create: Kibana provisions a hidden managed agent policy and a
// package policy in one call. The returned item's Id field is the package
// policy ID, which is the identifier used for all subsequent read, update,
// and delete operations (see Decision 4 in the fleet-agentless-policy
// OpenSpec change).
func CreateAgentlessPolicy(ctx context.Context, client *Client, spaceID string, body kbapi.PostFleetAgentlessPoliciesJSONRequestBody) (*kbapi.KibanaHTTPAPIsManagedIntegration, diag.Diagnostics) {
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

// ReadAgentlessPolicyViaPackagePolicy reads the current state of an agentless
// policy via GET /api/fleet/package_policies/{id}. There is no dedicated GET
// endpoint for agentless policies; the package_policies endpoint is the
// documented fallback and works against agentless-created policies (see
// Decision 4). Returns (nil, nil) on HTTP 404, signalling that the policy was
// removed out of band.
func ReadAgentlessPolicyViaPackagePolicy(ctx context.Context, client *Client, spaceID, policyID string) (*kbapi.PackagePolicy, diag.Diagnostics) {
	return GetPackagePolicy(ctx, client, policyID, spaceID)
}

// UpdateAgentlessPolicyViaPackagePolicy updates an agentless policy via
// PUT /api/fleet/package_policies/{id}. There is no dedicated PUT endpoint
// for agentless policies; the package_policies endpoint is the documented
// fallback (see Decision 4). Only the in-place-updatable allowlist of fields
// should be sent by callers (see Decision 3).
func UpdateAgentlessPolicyViaPackagePolicy(ctx context.Context, client *Client, spaceID, policyID string, body kbapi.PackagePolicyRequest) (*kbapi.PackagePolicy, diag.Diagnostics) {
	return UpdatePackagePolicy(ctx, client, policyID, spaceID, body)
}

// DeleteAgentlessPolicy deletes an existing Fleet agentless policy by its
// package policy ID. HTTP 404 is treated as success (idempotent delete).
// When force is true, the request is sent with ?force=true to delete the
// policy even if the underlying agent policy is managed.
//
// The returned bool reports whether the (final, post-retry) response was an
// HTTP 409 Conflict, so callers can offer a force_delete hint without having
// to pattern-match diagutil's generated diagnostic summary text -- see
// internal/fleet/agentlesspolicy/delete.go's conflictHintDiagnostics, which
// used to do exactly that and was flagged in review as brittle against
// wording changes or a switch to a different error-reporting helper.
func DeleteAgentlessPolicy(ctx context.Context, client *Client, spaceID, policyID string, force bool) (isConflict bool, diags diag.Diagnostics) {
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
