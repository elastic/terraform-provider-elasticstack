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

// GetTool reads a specific tool from the API.
func GetTool(ctx context.Context, client *Client, toolID string) (*models.Tool, diag.Diagnostics) {
	resp, err := client.API.GetAgentBuilderToolsToolidWithResponse(ctx, toolID)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		var tool models.Tool
		if err := json.Unmarshal(resp.Body, &tool); err != nil {
			return nil, diagutil.FrameworkDiagFromError(err)
		}
		return &tool, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// CreateTool creates a new tool.
func CreateTool(ctx context.Context, client *Client, req kbapi.PostAgentBuilderToolsJSONRequestBody) (*models.Tool, diag.Diagnostics) {
	resp, err := client.API.PostAgentBuilderToolsWithResponse(ctx, req)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		var tool models.Tool
		if err := json.Unmarshal(resp.Body, &tool); err != nil {
			return nil, diagutil.FrameworkDiagFromError(err)
		}
		return &tool, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// UpdateTool updates an existing tool.
func UpdateTool(ctx context.Context, client *Client, toolID string, req kbapi.PutAgentBuilderToolsToolidJSONRequestBody) (*models.Tool, diag.Diagnostics) {
	resp, err := client.API.PutAgentBuilderToolsToolidWithResponse(ctx, toolID, req)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		var tool models.Tool
		if err := json.Unmarshal(resp.Body, &tool); err != nil {
			return nil, diagutil.FrameworkDiagFromError(err)
		}
		return &tool, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// DeleteTool deletes an existing tool.
func DeleteTool(ctx context.Context, client *Client, toolID string) diag.Diagnostics {
	resp, err := client.API.DeleteAgentBuilderToolsToolidWithResponse(ctx, toolID)
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
