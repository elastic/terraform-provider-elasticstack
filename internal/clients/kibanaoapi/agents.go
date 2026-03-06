package kibanaoapi

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// GetAgent reads a specific agent from the API.
func GetAgent(ctx context.Context, client *Client, agentID string) (*models.Agent, diag.Diagnostics) {
	resp, err := client.API.GetAgentBuilderAgentsIdWithResponse(ctx, agentID)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		var agent models.Agent
		if err := json.Unmarshal(resp.Body, &agent); err != nil {
			return nil, diagutil.FrameworkDiagFromError(err)
		}
		return &agent, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// CreateAgent creates a new agent.
func CreateAgent(ctx context.Context, client *Client, req kbapi.PostAgentBuilderAgentsJSONRequestBody) (*models.Agent, diag.Diagnostics) {
	resp, err := client.API.PostAgentBuilderAgentsWithResponse(ctx, req)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		var agent models.Agent
		if err := json.Unmarshal(resp.Body, &agent); err != nil {
			return nil, diagutil.FrameworkDiagFromError(err)
		}
		return &agent, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// UpdateAgent updates an existing agent.
func UpdateAgent(ctx context.Context, client *Client, agentID string, req kbapi.PutAgentBuilderAgentsIdJSONRequestBody) (*models.Agent, diag.Diagnostics) {
	resp, err := client.API.PutAgentBuilderAgentsIdWithResponse(ctx, agentID, req)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		var agent models.Agent
		if err := json.Unmarshal(resp.Body, &agent); err != nil {
			return nil, diagutil.FrameworkDiagFromError(err)
		}
		return &agent, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// DeleteAgent deletes an existing agent.
func DeleteAgent(ctx context.Context, client *Client, agentID string) diag.Diagnostics {
	resp, err := client.API.DeleteAgentBuilderAgentsIdWithResponse(ctx, agentID)
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
