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

package kibanaoapi

import (
	"context"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// addInternalOriginHeader adds the required x-elastic-internal-origin header for workflow APIs
func addInternalOriginHeader(_ context.Context, req *http.Request) error {
	req.Header.Set("x-elastic-internal-origin", "Kibana")
	return nil
}

// GetWorkflow reads a specific workflow from the API.
func GetWorkflow(ctx context.Context, client *Client, spaceID string, workflowID string) (*kbapi.WorkflowDetailDto, diag.Diagnostics) {
	resp, err := client.API.GetWorkflowsIdWithResponse(ctx, workflowID, SpaceAwarePathRequestEditor(spaceID), addInternalOriginHeader)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp.JSON200, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// CreateWorkflow creates a new workflow.
func CreateWorkflow(ctx context.Context, client *Client, spaceID string, req kbapi.CreateWorkflowCommand) (*kbapi.WorkflowDetailDto, diag.Diagnostics) {
	resp, err := client.API.PostWorkflowsWithResponse(ctx, req, SpaceAwarePathRequestEditor(spaceID), addInternalOriginHeader)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp.JSON200, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// UpdateWorkflow updates an existing workflow.
func UpdateWorkflow(ctx context.Context, client *Client, spaceID string, workflowID string, req kbapi.UpdateWorkflowCommand) (*kbapi.UpdatedWorkflowResponseDto, diag.Diagnostics) {
	resp, err := client.API.PutWorkflowsIdWithResponse(ctx, workflowID, req, SpaceAwarePathRequestEditor(spaceID), addInternalOriginHeader)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp.JSON200, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// DeleteWorkflow deletes an existing workflow.
func DeleteWorkflow(ctx context.Context, client *Client, spaceID string, workflowID string) diag.Diagnostics {
	resp, err := client.API.DeleteWorkflowsIdWithResponse(ctx, workflowID, SpaceAwarePathRequestEditor(spaceID), addInternalOriginHeader)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return nil
	default:
		return reportUnknownError(resp.StatusCode(), resp.Body)
	}
}
