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
	"io"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// agentBuilderGet reduces the Get boilerplate for agent builder entity types.
// apiFn must match the kbapi raw Get method signature: (ctx, id, ...reqEditors).
func agentBuilderGet[T any](
	ctx context.Context,
	spaceID, entityID string,
	apiFn func(context.Context, string, ...kbapi.RequestEditorFn) (*http.Response, error),
) (*T, diag.Diagnostics) {
	rsp, err := apiFn(ctx, entityID, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	body, err := io.ReadAll(rsp.Body)
	_ = rsp.Body.Close()
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	return handleGetResponse[T](rsp.StatusCode, body)
}

// agentBuilderCreate reduces the Create (POST) boilerplate for agent builder entity types.
// apiFn must match the kbapi raw Post method signature: (ctx, body, ...reqEditors).
func agentBuilderCreate[T any, B any](
	ctx context.Context,
	spaceID string,
	reqBody B,
	apiFn func(context.Context, B, ...kbapi.RequestEditorFn) (*http.Response, error),
) (*T, diag.Diagnostics) {
	rsp, err := apiFn(ctx, reqBody, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	body, err := io.ReadAll(rsp.Body)
	_ = rsp.Body.Close()
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	return handleMutateResponse[T](rsp.StatusCode, body)
}

// agentBuilderUpdate reduces the Update (PUT) boilerplate for agent builder entity types.
// apiFn must match the kbapi raw Put method signature: (ctx, id, body, ...reqEditors).
func agentBuilderUpdate[T any, B any](
	ctx context.Context,
	spaceID, entityID string,
	reqBody B,
	apiFn func(context.Context, string, B, ...kbapi.RequestEditorFn) (*http.Response, error),
) (*T, diag.Diagnostics) {
	rsp, err := apiFn(ctx, entityID, reqBody, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	body, err := io.ReadAll(rsp.Body)
	_ = rsp.Body.Close()
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	return handleMutateResponse[T](rsp.StatusCode, body)
}

// agentBuilderDelete reduces the Delete boilerplate for agent builder entity types.
// apiFn is a zero-argument closure that calls the appropriate kbapi Delete method,
// capturing all required parameters (ctx, id, params, reqEditor) in the closure.
func agentBuilderDelete(apiFn func() (*http.Response, error)) diag.Diagnostics {
	rsp, err := apiFn()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	body, err := io.ReadAll(rsp.Body)
	_ = rsp.Body.Close()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return diagutil.HandleStatusResponse(rsp.StatusCode, body, http.StatusOK, http.StatusNotFound)
}

// --- Agent ---

// GetAgent reads a specific agent from the API.
func GetAgent(ctx context.Context, client *Client, spaceID, agentID string) (*models.Agent, diag.Diagnostics) {
	return agentBuilderGet[models.Agent](ctx, spaceID, agentID, client.API.GetAgentBuilderAgentsId)
}

// CreateAgent creates a new agent.
func CreateAgent(ctx context.Context, client *Client, spaceID string, req kbapi.PostAgentBuilderAgentsJSONRequestBody) (*models.Agent, diag.Diagnostics) {
	return agentBuilderCreate[models.Agent](ctx, spaceID, req, client.API.PostAgentBuilderAgents)
}

// UpdateAgent updates an existing agent.
func UpdateAgent(ctx context.Context, client *Client, spaceID, agentID string, req kbapi.PutAgentBuilderAgentsIdJSONRequestBody) (*models.Agent, diag.Diagnostics) {
	return agentBuilderUpdate[models.Agent](ctx, spaceID, agentID, req, client.API.PutAgentBuilderAgentsId)
}

// DeleteAgent deletes an existing agent.
func DeleteAgent(ctx context.Context, client *Client, spaceID, agentID string) diag.Diagnostics {
	return agentBuilderDelete(func() (*http.Response, error) {
		return client.API.DeleteAgentBuilderAgentsId(ctx, agentID, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	})
}

// --- Skill ---

// GetSkill reads a specific skill from the API.
func GetSkill(ctx context.Context, client *Client, spaceID, skillID string) (*models.Skill, diag.Diagnostics) {
	return agentBuilderGet[models.Skill](ctx, spaceID, skillID, client.API.GetAgentBuilderSkillsSkillid)
}

// CreateSkill creates a new skill.
func CreateSkill(ctx context.Context, client *Client, spaceID string, req kbapi.PostAgentBuilderSkillsJSONRequestBody) (*models.Skill, diag.Diagnostics) {
	return agentBuilderCreate[models.Skill](ctx, spaceID, req, client.API.PostAgentBuilderSkills)
}

// UpdateSkill updates an existing skill.
func UpdateSkill(ctx context.Context, client *Client, spaceID, skillID string, req kbapi.PutAgentBuilderSkillsSkillidJSONRequestBody) (*models.Skill, diag.Diagnostics) {
	return agentBuilderUpdate[models.Skill](ctx, spaceID, skillID, req, client.API.PutAgentBuilderSkillsSkillid)
}

// DeleteSkill deletes an existing skill. The API also accepts a `force=true`
// query parameter to cascade the deletion through referencing agents; the
// resource does not expose this in v1 so we always send an empty params
// struct and let 409 Conflict flow through as a normal error diagnostic.
func DeleteSkill(ctx context.Context, client *Client, spaceID, skillID string) diag.Diagnostics {
	return agentBuilderDelete(func() (*http.Response, error) {
		return client.API.DeleteAgentBuilderSkillsSkillid(ctx, skillID, &kbapi.DeleteAgentBuilderSkillsSkillidParams{}, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	})
}

// --- Tool ---

// GetTool reads a specific tool from the API.
func GetTool(ctx context.Context, client *Client, spaceID, toolID string) (*models.Tool, diag.Diagnostics) {
	return agentBuilderGet[models.Tool](ctx, spaceID, toolID, client.API.GetAgentBuilderToolsToolid)
}

// CreateTool creates a new tool.
func CreateTool(ctx context.Context, client *Client, spaceID string, req kbapi.PostAgentBuilderToolsJSONRequestBody) (*models.Tool, diag.Diagnostics) {
	return agentBuilderCreate[models.Tool](ctx, spaceID, req, client.API.PostAgentBuilderTools)
}

// UpdateTool updates an existing tool.
func UpdateTool(ctx context.Context, client *Client, spaceID, toolID string, req kbapi.PutAgentBuilderToolsToolidJSONRequestBody) (*models.Tool, diag.Diagnostics) {
	return agentBuilderUpdate[models.Tool](ctx, spaceID, toolID, req, client.API.PutAgentBuilderToolsToolid)
}

// DeleteTool deletes an existing tool.
func DeleteTool(ctx context.Context, client *Client, spaceID, toolID string) diag.Diagnostics {
	return agentBuilderDelete(func() (*http.Response, error) {
		return client.API.DeleteAgentBuilderToolsToolid(ctx, toolID, nil, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	})
}

// --- Workflow ---

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
func GetWorkflow(ctx context.Context, client *Client, spaceID, workflowID string) (*models.Workflow, diag.Diagnostics) {
	return agentBuilderGet[models.Workflow](ctx, spaceID, workflowID, client.API.GetWorkflowsWorkflowId)
}

// CreateWorkflow creates a new workflow.
func CreateWorkflow(ctx context.Context, client *Client, spaceID string, req kbapi.PostWorkflowsWorkflowJSONRequestBody) (*models.Workflow, diag.Diagnostics) {
	return agentBuilderCreate[models.Workflow](ctx, spaceID, req, client.API.PostWorkflowsWorkflow)
}

// UpdateWorkflow updates an existing workflow. The returned PartialWorkflow
// reflects the PUT response (id, valid, enabled only); callers needing full
// state should rely on the resource envelope's read-after-write refresh.
func UpdateWorkflow(ctx context.Context, client *Client, spaceID, workflowID string, req kbapi.PutWorkflowsWorkflowIdJSONRequestBody) (*PartialWorkflow, diag.Diagnostics) {
	return agentBuilderUpdate[PartialWorkflow](ctx, spaceID, workflowID, req, client.API.PutWorkflowsWorkflowId)
}

// DeleteWorkflow deletes an existing workflow.
func DeleteWorkflow(ctx context.Context, client *Client, spaceID, workflowID string) diag.Diagnostics {
	return agentBuilderDelete(func() (*http.Response, error) {
		return client.API.DeleteWorkflowsWorkflowId(ctx, workflowID, nil, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	})
}
