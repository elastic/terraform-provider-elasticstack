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
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// GetAgentPolicy reads a specific agent policy from the API.
func GetAgentPolicy(ctx context.Context, client *Client, id string, spaceID string) (*kbapi.AgentPolicy, diag.Diagnostics) {
	resp, err := client.API.GetFleetAgentPoliciesAgentpolicyidWithResponse(ctx, id, nil, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &resp.JSON200.Item, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}

// CreateAgentPolicy creates a new agent policy.
func CreateAgentPolicy(ctx context.Context, client *Client, req kbapi.PostFleetAgentPoliciesJSONRequestBody, sysMonitoring bool, spaceID string) (*kbapi.AgentPolicy, diag.Diagnostics) {
	params := kbapi.PostFleetAgentPoliciesParams{
		SysMonitoring: new(sysMonitoring),
	}

	return kibanautil.ConflictRetry(ctx, kibanautil.ConflictMaxAttempts, func() (*kbapi.AgentPolicy, int, diag.Diagnostics) {
		resp, err := client.API.PostFleetAgentPoliciesWithResponse(ctx, &params, req, kibanautil.SpaceAwarePathRequestEditor(spaceID))
		if err != nil {
			return nil, 0, diagutil.FrameworkDiagFromError(err)
		}

		switch resp.StatusCode() {
		case http.StatusOK:
			return &resp.JSON200.Item, resp.StatusCode(), nil
		default:
			return nil, resp.StatusCode(), diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
		}
	})
}

// UpdateAgentPolicy updates an existing agent policy.
func UpdateAgentPolicy(ctx context.Context, client *Client, id string, spaceID string, req kbapi.PutFleetAgentPoliciesAgentpolicyidJSONRequestBody) (*kbapi.AgentPolicy, diag.Diagnostics) {
	return kibanautil.ConflictRetry(ctx, kibanautil.ConflictMaxAttempts, func() (*kbapi.AgentPolicy, int, diag.Diagnostics) {
		resp, err := client.API.PutFleetAgentPoliciesAgentpolicyidWithResponse(ctx, id, nil, req, kibanautil.SpaceAwarePathRequestEditor(spaceID))
		if err != nil {
			return nil, 0, diagutil.FrameworkDiagFromError(err)
		}

		switch resp.StatusCode() {
		case http.StatusOK:
			return &resp.JSON200.Item, resp.StatusCode(), nil
		default:
			return nil, resp.StatusCode(), diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
		}
	})
}

// DeleteAgentPolicy deletes an existing agent policy.
func DeleteAgentPolicy(ctx context.Context, client *Client, id string, spaceID string) diag.Diagnostics {
	body := kbapi.PostFleetAgentPoliciesDeleteJSONRequestBody{
		AgentPolicyId: id,
	}

	_, diags := kibanautil.ConflictRetry(ctx, kibanautil.ConflictMaxAttempts, func() (struct{}, int, diag.Diagnostics) {
		resp, err := client.API.PostFleetAgentPoliciesDeleteWithResponse(ctx, body, kibanautil.SpaceAwarePathRequestEditor(spaceID))
		if err != nil {
			return struct{}{}, 0, diagutil.FrameworkDiagFromError(err)
		}
		return struct{}{}, resp.StatusCode(), handleDeleteResponse(resp.StatusCode(), resp.Body)
	})
	return diags
}
