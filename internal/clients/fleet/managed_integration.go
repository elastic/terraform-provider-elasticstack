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

// CreateManagedIntegration creates a new Fleet managed integration via
// POST /api/fleet/managed_integrations.
func CreateManagedIntegration(ctx context.Context, client *Client, spaceID string, body kbapi.PostFleetManagedIntegrationsJSONRequestBody) (*kbapi.KibanaHTTPAPIsManagedIntegration, diag.Diagnostics) {
	return kibanautil.ConflictRetry(ctx, kibanautil.ConflictMaxAttempts, func() (*kbapi.KibanaHTTPAPIsManagedIntegration, int, diag.Diagnostics) {
		resp, err := client.API.PostFleetManagedIntegrationsWithResponse(ctx, body, kibanautil.SpaceAwarePathRequestEditor(spaceID))
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

// ReadManagedIntegration reads a Fleet managed integration via
// GET /api/fleet/managed_integrations/{id}. Returns (nil, nil) on HTTP 404,
// signalling that the integration was removed out of band.
func ReadManagedIntegration(ctx context.Context, client *Client, spaceID, policyID string) (*kbapi.KibanaHTTPAPIsManagedIntegration, diag.Diagnostics) {
	resp, err := client.API.GetFleetManagedIntegrationsPolicyidWithResponse(ctx, policyID, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	return kibanaoapi.HandleGetTypedResponse(resp.StatusCode(), resp.Body, func() *kbapi.KibanaHTTPAPIsManagedIntegration {
		if resp.JSON200 == nil {
			return nil
		}
		return &resp.JSON200.Item
	})
}

// UpdateManagedIntegration updates a Fleet managed integration via
// PUT /api/fleet/managed_integrations/{id}.
func UpdateManagedIntegration(ctx context.Context, client *Client, spaceID, policyID string, body kbapi.PutFleetManagedIntegrationsPolicyidJSONRequestBody) (*kbapi.KibanaHTTPAPIsManagedIntegration, diag.Diagnostics) {
	return kibanautil.ConflictRetry(ctx, kibanautil.ConflictMaxAttempts, func() (*kbapi.KibanaHTTPAPIsManagedIntegration, int, diag.Diagnostics) {
		resp, err := client.API.PutFleetManagedIntegrationsPolicyidWithResponse(ctx, policyID, body, kibanautil.SpaceAwarePathRequestEditor(spaceID))
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

// DeleteManagedIntegration deletes a Fleet managed integration via
// DELETE /api/fleet/managed_integrations/{id}. HTTP 404 is treated as success
// (idempotent delete). When force is true, the request is sent with
// ?force=true to delete the policy even if the underlying agent policy is
// managed.
//
// The returned bool reports whether the (final, post-retry) response was an
// HTTP 409 Conflict, so callers can offer a force_delete hint without having
// to pattern-match diagutil's generated diagnostic summary text.
func DeleteManagedIntegration(ctx context.Context, client *Client, spaceID, policyID string, force bool) (isConflict bool, diags diag.Diagnostics) {
	params := kbapi.DeleteFleetManagedIntegrationsPolicyidParams{}
	if force {
		t := true
		params.Force = &t
	}

	var lastStatusCode int
	_, diags = kibanautil.ConflictRetry(ctx, kibanautil.ConflictMaxAttempts, func() (struct{}, int, diag.Diagnostics) {
		resp, err := client.API.DeleteFleetManagedIntegrationsPolicyidWithResponse(ctx, policyID, &params, kibanautil.SpaceAwarePathRequestEditor(spaceID))
		if err != nil {
			lastStatusCode = 0
			return struct{}{}, 0, diagutil.FrameworkDiagFromError(err)
		}
		lastStatusCode = resp.StatusCode()
		return struct{}{}, resp.StatusCode(), handleDeleteResponse(resp.StatusCode(), resp.Body)
	})
	return lastStatusCode == http.StatusConflict, diags
}
