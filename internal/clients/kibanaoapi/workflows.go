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
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// GetWorkflow reads a specific workflow from the API.
func GetWorkflow(ctx context.Context, client *Client, spaceID string, workflowID string) (*models.Workflow, diag.Diagnostics) {
	resp, err := client.API.GetWorkflowsWorkflowIdWithResponse(ctx, workflowID, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, clientError(err)
	}
	return handleGetResponse[models.Workflow](resp.StatusCode(), resp.Body)
}

// CreateWorkflow creates a new workflow.
func CreateWorkflow(ctx context.Context, client *Client, spaceID string, req kbapi.PostWorkflowsWorkflowJSONRequestBody) (*models.Workflow, diag.Diagnostics) {
	resp, err := client.API.PostWorkflowsWorkflowWithResponse(ctx, req, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, clientError(err)
	}
	return handleMutateResponse[models.Workflow](resp.StatusCode(), resp.Body)
}

// UpdateWorkflow updates an existing workflow.
// The PUT response is partial (id, valid, enabled only); callers must GET afterwards for full state.
func UpdateWorkflow(ctx context.Context, client *Client, spaceID string, workflowID string, req kbapi.PutWorkflowsWorkflowIdJSONRequestBody) diag.Diagnostics {
	resp, err := client.API.PutWorkflowsWorkflowIdWithResponse(ctx, workflowID, req, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return clientError(err)
	}
	return handleStatusResponse(resp.StatusCode(), resp.Body, http.StatusOK)
}

// DeleteWorkflow deletes an existing workflow.
func DeleteWorkflow(ctx context.Context, client *Client, spaceID string, workflowID string) diag.Diagnostics {
	resp, err := client.API.DeleteWorkflowsWorkflowIdWithResponse(ctx, workflowID, nil, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return clientError(err)
	}
	return handleStatusResponse(resp.StatusCode(), resp.Body, http.StatusOK, http.StatusNotFound)
}
