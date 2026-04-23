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

package agentbuilderapi

import (
	"context"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// WorkflowsAPI implements the ResourceAPI interface for Agent Builder workflows.
type WorkflowsAPI struct{}

// Create creates a new workflow and returns its ID.
func (w *WorkflowsAPI) Create(
	ctx context.Context,
	client *kibanaoapi.Client,
	spaceID string,
	req kbapi.PostWorkflowsWorkflowJSONRequestBody,
) (string, diag.Diagnostics) {
	resp, err := client.API.PostWorkflowsWorkflowWithResponse(ctx, req, kibanaoapi.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return "", diagutil.FrameworkDiagFromError(err)
	}

	workflow, diags := handleMutateResponse[models.Workflow](resp.StatusCode(), resp.Body)
	if diags.HasError() {
		return "", diags
	}

	return workflow.ID, diags
}

// Get retrieves a workflow by ID. Returns (nil, false, nil) if not found.
func (w *WorkflowsAPI) Get(
	ctx context.Context,
	client *kibanaoapi.Client,
	spaceID string,
	workflowID string,
) (*models.Workflow, bool, diag.Diagnostics) {
	resp, err := client.API.GetWorkflowsWorkflowIdWithResponse(ctx, workflowID, kibanaoapi.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, false, diagutil.FrameworkDiagFromError(err)
	}

	return handleGetResponse[models.Workflow](resp.StatusCode(), resp.Body)
}

// Update updates an existing workflow. Note that the PUT response is partial;
// callers should follow up with a Get to retrieve the full workflow state.
func (w *WorkflowsAPI) Update(
	ctx context.Context,
	client *kibanaoapi.Client,
	spaceID string,
	workflowID string,
	req kbapi.PutWorkflowsWorkflowIdJSONRequestBody,
) diag.Diagnostics {
	resp, err := client.API.PutWorkflowsWorkflowIdWithResponse(ctx, workflowID, req, kibanaoapi.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	return handleStatusResponse(resp.StatusCode(), resp.Body, http.StatusOK)
}

// Delete deletes a workflow by ID.
func (w *WorkflowsAPI) Delete(
	ctx context.Context,
	client *kibanaoapi.Client,
	spaceID string,
	workflowID string,
) diag.Diagnostics {
	resp, err := client.API.DeleteWorkflowsWorkflowIdWithResponse(ctx, workflowID, nil, kibanaoapi.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	return handleStatusResponse(resp.StatusCode(), resp.Body, http.StatusOK, http.StatusNotFound)
}
