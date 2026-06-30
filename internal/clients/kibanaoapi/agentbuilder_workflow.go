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
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// PartialWorkflow captures the subset of fields the workflow PUT endpoint
// returns (id, valid, enabled). Callers use it to validate the update result
// without issuing an extra GET; full state is refreshed by the resource
// envelope's read-after-write step.
type PartialWorkflow struct {
	ID      string `json:"id"`
	Valid   bool   `json:"valid"`
	Enabled bool   `json:"enabled"`
}

// GetWorkflow reads a specific workflow from the API.
func GetWorkflow(ctx context.Context, client *Client, spaceID string, workflowID string) (*models.Workflow, diag.Diagnostics) {
	resp, err := client.API.GetWorkflowsWorkflowIdWithResponse(ctx, workflowID, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	return HandleGetRawResponse[models.Workflow](resp.StatusCode(), resp.Body)
}

// CreateWorkflow creates a new workflow.
func CreateWorkflow(ctx context.Context, client *Client, spaceID string, req kbapi.PostWorkflowsWorkflowJSONRequestBody) (*models.Workflow, diag.Diagnostics) {
	resp, err := client.API.PostWorkflowsWorkflowWithResponse(ctx, req, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	return HandleMutateRawResponse[models.Workflow](resp.StatusCode(), resp.Body)
}

// UpdateWorkflow updates an existing workflow. The returned PartialWorkflow
// reflects the PUT response (id, valid, enabled only); callers needing full
// state should rely on the resource envelope's read-after-write refresh.
func UpdateWorkflow(ctx context.Context, client *Client, spaceID string, workflowID string, req kbapi.PutWorkflowsWorkflowIdJSONRequestBody) (*PartialWorkflow, diag.Diagnostics) {
	resp, err := client.API.PutWorkflowsWorkflowIdWithResponse(ctx, workflowID, req, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	return HandleMutateRawResponse[PartialWorkflow](resp.StatusCode(), resp.Body)
}

// DeleteWorkflow deletes an existing workflow.
func DeleteWorkflow(ctx context.Context, client *Client, spaceID string, workflowID string) diag.Diagnostics {
	resp, err := client.API.DeleteWorkflowsWorkflowIdWithResponse(ctx, workflowID, nil, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return diagutil.HandleStatusResponse(resp.StatusCode(), resp.Body, http.StatusOK, http.StatusNotFound)
}
