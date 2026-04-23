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

// AgentsAPI implements the ResourceAPI interface for Agent Builder agents.
type AgentsAPI struct{}

// Create creates a new agent and returns its ID.
func (a *AgentsAPI) Create(
	ctx context.Context,
	client *kibanaoapi.Client,
	spaceID string,
	req kbapi.PostAgentBuilderAgentsJSONRequestBody,
) (string, diag.Diagnostics) {
	resp, err := client.API.PostAgentBuilderAgentsWithResponse(ctx, req, kibanaoapi.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return "", diagutil.FrameworkDiagFromError(err)
	}

	agent, diags := handleMutateResponse[models.Agent](resp.StatusCode(), resp.Body)
	if diags.HasError() {
		return "", diags
	}

	return agent.ID, diags
}

// Get retrieves an agent by ID. Returns (nil, false, nil) if not found.
func (a *AgentsAPI) Get(
	ctx context.Context,
	client *kibanaoapi.Client,
	spaceID string,
	agentID string,
) (*models.Agent, bool, diag.Diagnostics) {
	resp, err := client.API.GetAgentBuilderAgentsIdWithResponse(ctx, agentID, kibanaoapi.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, false, diagutil.FrameworkDiagFromError(err)
	}

	return handleGetResponse[models.Agent](resp.StatusCode(), resp.Body)
}

// Update updates an existing agent.
func (a *AgentsAPI) Update(
	ctx context.Context,
	client *kibanaoapi.Client,
	spaceID string,
	agentID string,
	req kbapi.PutAgentBuilderAgentsIdJSONRequestBody,
) diag.Diagnostics {
	resp, err := client.API.PutAgentBuilderAgentsIdWithResponse(ctx, agentID, req, kibanaoapi.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	return handleStatusResponse(resp.StatusCode(), resp.Body, http.StatusOK)
}

// Delete deletes an agent by ID.
func (a *AgentsAPI) Delete(
	ctx context.Context,
	client *kibanaoapi.Client,
	spaceID string,
	agentID string,
) diag.Diagnostics {
	resp, err := client.API.DeleteAgentBuilderAgentsIdWithResponse(ctx, agentID, kibanaoapi.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	return handleStatusResponse(resp.StatusCode(), resp.Body, http.StatusOK, http.StatusNotFound)
}
