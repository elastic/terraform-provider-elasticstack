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
	return HandleGetRawResponse[models.Agent](resp.StatusCode(), resp.Body)
}

// CreateAgent creates a new agent.
func CreateAgent(ctx context.Context, client *Client, spaceID string, req kbapi.PostAgentBuilderAgentsJSONRequestBody) (*models.Agent, diag.Diagnostics) {
	resp, err := client.API.PostAgentBuilderAgentsWithResponse(ctx, req, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	return HandleMutateRawResponse[models.Agent](resp.StatusCode(), resp.Body)
}

// UpdateAgent updates an existing agent.
func UpdateAgent(ctx context.Context, client *Client, spaceID string, agentID string, req kbapi.PutAgentBuilderAgentsIdJSONRequestBody) (*models.Agent, diag.Diagnostics) {
	resp, err := client.API.PutAgentBuilderAgentsIdWithResponse(ctx, agentID, req, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	return HandleMutateRawResponse[models.Agent](resp.StatusCode(), resp.Body)
}

// DeleteAgent deletes an existing agent.
func DeleteAgent(ctx context.Context, client *Client, spaceID, agentID string) diag.Diagnostics {
	resp, err := client.API.DeleteAgentBuilderAgentsIdWithResponse(ctx, agentID, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return diagutil.HandleStatusResponse(resp.StatusCode(), resp.Body, http.StatusOK, http.StatusNotFound)
}

// GetSkill reads a specific skill from the API.
func GetSkill(ctx context.Context, client *Client, spaceID, skillID string) (*models.Skill, diag.Diagnostics) {
	resp, err := client.API.GetAgentBuilderSkillsSkillidWithResponse(ctx, skillID, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	return HandleGetRawResponse[models.Skill](resp.StatusCode(), resp.Body)
}

// CreateSkill creates a new skill.
func CreateSkill(ctx context.Context, client *Client, spaceID string, req kbapi.PostAgentBuilderSkillsJSONRequestBody) (*models.Skill, diag.Diagnostics) {
	resp, err := client.API.PostAgentBuilderSkillsWithResponse(ctx, req, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	return HandleMutateRawResponse[models.Skill](resp.StatusCode(), resp.Body)
}

// UpdateSkill updates an existing skill.
func UpdateSkill(ctx context.Context, client *Client, spaceID, skillID string, req kbapi.PutAgentBuilderSkillsSkillidJSONRequestBody) (*models.Skill, diag.Diagnostics) {
	resp, err := client.API.PutAgentBuilderSkillsSkillidWithResponse(ctx, skillID, req, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	return HandleMutateRawResponse[models.Skill](resp.StatusCode(), resp.Body)
}

// DeleteSkill deletes an existing skill. The API also accepts a force=true
// query parameter to cascade the deletion through referencing agents; the
// resource does not expose this in v1 so we always send an empty params
// struct and let 409 Conflict flow through as a normal error diagnostic.
func DeleteSkill(ctx context.Context, client *Client, spaceID, skillID string) diag.Diagnostics {
	resp, err := client.API.DeleteAgentBuilderSkillsSkillidWithResponse(ctx, skillID, &kbapi.DeleteAgentBuilderSkillsSkillidParams{}, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return diagutil.HandleStatusResponse(resp.StatusCode(), resp.Body, http.StatusOK, http.StatusNotFound)
}

// GetTool reads a specific tool from the API.
func GetTool(ctx context.Context, client *Client, spaceID string, toolID string) (*models.Tool, diag.Diagnostics) {
	resp, err := client.API.GetAgentBuilderToolsToolidWithResponse(ctx, toolID, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	return HandleGetRawResponse[models.Tool](resp.StatusCode(), resp.Body)
}

// CreateTool creates a new tool.
func CreateTool(ctx context.Context, client *Client, spaceID string, req kbapi.PostAgentBuilderToolsJSONRequestBody) (*models.Tool, diag.Diagnostics) {
	resp, err := client.API.PostAgentBuilderToolsWithResponse(ctx, req, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	return HandleMutateRawResponse[models.Tool](resp.StatusCode(), resp.Body)
}

// UpdateTool updates an existing tool.
func UpdateTool(ctx context.Context, client *Client, spaceID string, toolID string, req kbapi.PutAgentBuilderToolsToolidJSONRequestBody) (*models.Tool, diag.Diagnostics) {
	resp, err := client.API.PutAgentBuilderToolsToolidWithResponse(ctx, toolID, req, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	return HandleMutateRawResponse[models.Tool](resp.StatusCode(), resp.Body)
}

// DeleteTool deletes an existing tool.
func DeleteTool(ctx context.Context, client *Client, spaceID string, toolID string) diag.Diagnostics {
	resp, err := client.API.DeleteAgentBuilderToolsToolidWithResponse(ctx, toolID, nil, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return diagutil.HandleStatusResponse(resp.StatusCode(), resp.Body, http.StatusOK, http.StatusNotFound)
}

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
