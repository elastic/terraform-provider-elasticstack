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

// GetAgent reads a specific agent from the API.
func GetAgent(ctx context.Context, client *Client, spaceID, agentID string) (*models.Agent, diag.Diagnostics) {
	resp, err := client.API.GetAgentBuilderAgentsIdWithResponse(ctx, agentID, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	return handleGetResponse[models.Agent](resp.StatusCode(), resp.Body)
}

// CreateAgent creates a new agent.
func CreateAgent(ctx context.Context, client *Client, spaceID string, req kbapi.PostAgentBuilderAgentsJSONRequestBody) (*models.Agent, diag.Diagnostics) {
	resp, err := client.API.PostAgentBuilderAgentsWithResponse(ctx, req, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	return handleMutateResponse[models.Agent](resp.StatusCode(), resp.Body)
}

// UpdateAgent updates an existing agent.
func UpdateAgent(ctx context.Context, client *Client, spaceID string, agentID string, req kbapi.PutAgentBuilderAgentsIdJSONRequestBody) (*models.Agent, diag.Diagnostics) {
	resp, err := client.API.PutAgentBuilderAgentsIdWithResponse(ctx, agentID, req, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	return handleMutateResponse[models.Agent](resp.StatusCode(), resp.Body)
}

// DeleteAgent deletes an existing agent.
func DeleteAgent(ctx context.Context, client *Client, spaceID, agentID string) diag.Diagnostics {
	resp, err := client.API.DeleteAgentBuilderAgentsIdWithResponse(ctx, agentID, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return diagutil.HandleStatusResponse(resp.StatusCode(), resp.Body, http.StatusOK, http.StatusNotFound)
}
