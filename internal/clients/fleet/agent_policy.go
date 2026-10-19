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
	"fmt"
	"net/http"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// GetAgentPolicy reads a specific agent policy from the API.
func GetAgentPolicy(ctx context.Context, client *Client, id string, spaceID string) (*kbapi.KibanaHTTPAPIsAgentPolicyResponse, diag.Diagnostics) {
	resp, err := client.API.GetFleetAgentPoliciesAgentpolicyidWithResponse(ctx, id, nil, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	return kibanaoapi.HandleGetTypedResponse(resp.StatusCode(), resp.Body, func() *kbapi.KibanaHTTPAPIsAgentPolicyResponse { return &resp.JSON200.Item })
}

const agentPolicyByNamePerPage float32 = 100

// GetAgentPolicyByName returns every agent policy in spaceID whose name equals name.
// An empty slice means nothing matched and is not an error.
func GetAgentPolicyByName(ctx context.Context, client *Client, name string, spaceID string) ([]*kbapi.KibanaHTTPAPIsAgentPolicyResponse, diag.Diagnostics) {
	kuery := fmt.Sprintf("ingest-agent-policies.name:\"%s\"", escapeKueryQuotedLiteral(name))
	sortField := "name"
	perPage := agentPolicyByNamePerPage

	seenIDs := make(map[string]struct{})
	var expectedTotal float32
	var seenCount float32

	collected, diags := kibanaoapi.CollectAllPages(func(page int) ([]kbapi.KibanaHTTPAPIsAgentPolicyResponse, int, diag.Diagnostics) {
		pageNum := float32(page)
		params := &kbapi.GetFleetAgentPoliciesParams{
			Kuery:     &kuery,
			SortField: &sortField,
			Page:      new(pageNum),
			PerPage:   new(perPage),
		}
		resp, err := client.API.GetFleetAgentPoliciesWithResponse(ctx, params, kibanautil.SpaceAwarePathRequestEditor(spaceID))
		if err != nil {
			return nil, 0, diagutil.FrameworkDiagFromError(err)
		}
		if diags := diagutil.HandleStatusResponse(resp.StatusCode(), resp.Body, http.StatusOK); diags.HasError() {
			return nil, 0, diags
		}
		pageResult, diags := diagutil.UnwrapJSON200(resp.JSON200, "agent policy list")
		if diags.HasError() {
			return nil, 0, diags
		}

		if page > 1 && pageResult.Total != expectedTotal {
			return nil, 0, inconsistentAgentPolicyList(fmt.Sprintf("total changed from %g to %g while paging agent policies", expectedTotal, pageResult.Total))
		}
		expectedTotal = pageResult.Total

		for _, item := range pageResult.Items {
			if _, seen := seenIDs[item.Id]; seen {
				return nil, 0, inconsistentAgentPolicyList(fmt.Sprintf("agent policy id %q was returned more than once", item.Id))
			}
			seenIDs[item.Id] = struct{}{}
		}

		seenCount += float32(len(pageResult.Items))
		if seenCount < expectedTotal && float32(len(pageResult.Items)) < perPage {
			return nil, 0, inconsistentAgentPolicyList(fmt.Sprintf("page %g returned %d items, fewer than perPage %g, before reaching total %g", pageNum, len(pageResult.Items), perPage, expectedTotal))
		}

		return pageResult.Items, int(pageResult.Total), nil
	})
	if diags.HasError() {
		return nil, diags
	}

	matches := make([]*kbapi.KibanaHTTPAPIsAgentPolicyResponse, 0)
	for i := range collected {
		if collected[i].Name == name {
			matches = append(matches, &collected[i])
		}
	}
	return matches, nil
}

func inconsistentAgentPolicyList(detail string) diag.Diagnostics {
	return diag.Diagnostics{diag.NewErrorDiagnostic("Inconsistent agent policy list", detail)}
}

// escapeKueryQuotedLiteral escapes value for a double-quoted KQL literal.
// Backslashes are escaped before quotes so the backslashes added for quotes are not doubled.
func escapeKueryQuotedLiteral(value string) string {
	escaped := strings.ReplaceAll(value, `\`, `\\`)
	return strings.ReplaceAll(escaped, `"`, `\"`)
}

// CreateAgentPolicy creates a new agent policy.
func CreateAgentPolicy(
	ctx context.Context,
	client *Client,
	req kbapi.PostFleetAgentPoliciesJSONRequestBody,
	sysMonitoring bool,
	spaceID string,
) (*kbapi.KibanaHTTPAPIsAgentPolicyResponse, diag.Diagnostics) {
	params := kbapi.PostFleetAgentPoliciesParams{
		SysMonitoring: new(sysMonitoring),
	}

	return kibanautil.ConflictRetry(ctx, kibanautil.ConflictMaxAttempts, func() (*kbapi.KibanaHTTPAPIsAgentPolicyResponse, int, diag.Diagnostics) {
		resp, err := client.API.PostFleetAgentPoliciesWithResponse(ctx, &params, req, kibanautil.SpaceAwarePathRequestEditor(spaceID))
		if err != nil {
			return nil, 0, diagutil.FrameworkDiagFromError(err)
		}

		result, diags := kibanaoapi.HandleMutateTypedResponse(resp.StatusCode(), resp.Body, func() *kbapi.KibanaHTTPAPIsAgentPolicyResponse { return &resp.JSON200.Item })
		return result, resp.StatusCode(), diags
	})
}

// UpdateAgentPolicy updates an existing agent policy.
func UpdateAgentPolicy(
	ctx context.Context,
	client *Client,
	id string,
	spaceID string,
	req kbapi.PutFleetAgentPoliciesAgentpolicyidJSONRequestBody,
) (*kbapi.KibanaHTTPAPIsAgentPolicyResponse, diag.Diagnostics) {
	return kibanautil.ConflictRetry(ctx, kibanautil.ConflictMaxAttempts, func() (*kbapi.KibanaHTTPAPIsAgentPolicyResponse, int, diag.Diagnostics) {
		resp, err := client.API.PutFleetAgentPoliciesAgentpolicyidWithResponse(ctx, id, nil, req, kibanautil.SpaceAwarePathRequestEditor(spaceID))
		if err != nil {
			return nil, 0, diagutil.FrameworkDiagFromError(err)
		}

		result, diags := kibanaoapi.HandleMutateTypedResponse(resp.StatusCode(), resp.Body, func() *kbapi.KibanaHTTPAPIsAgentPolicyResponse { return &resp.JSON200.Item })
		return result, resp.StatusCode(), diags
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
