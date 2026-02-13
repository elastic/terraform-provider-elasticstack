package export_ab

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// addInternalOriginHeader adds the required x-elastic-internal-origin header for workflow APIs
func addInternalOriginHeader(ctx context.Context, req *http.Request) error {
	req.Header.Set("x-elastic-internal-origin", "Kibana")
	return nil
}

// ToolModel maps tool data
type ToolModel struct {
	ID            types.String `tfsdk:"id"`
	Type          types.String `tfsdk:"type"`
	Description   types.String `tfsdk:"description"`
	Tags          types.List   `tfsdk:"tags"`
	ReadOnly      types.Bool   `tfsdk:"readonly"`
	Configuration types.String `tfsdk:"configuration"`
}

// WorkflowModel maps workflow data
type WorkflowModel struct {
	ID   types.String `tfsdk:"id"`
	Yaml types.String `tfsdk:"yaml"`
}

// ExtractToolIds extracts tool IDs from the agent configuration
func ExtractToolIds(agentData map[string]interface{}) []string {
	var toolIds []string

	if config, ok := agentData["configuration"].(map[string]interface{}); ok {
		if tools, ok := config["tools"].([]interface{}); ok {
			for _, tool := range tools {
				if toolMap, ok := tool.(map[string]interface{}); ok {
					if ids, ok := toolMap["tool_ids"].([]interface{}); ok {
						for _, id := range ids {
							if idStr, ok := id.(string); ok {
								toolIds = append(toolIds, idStr)
							}
						}
					}
				}
			}
		}
	}

	return toolIds
}

// FetchTool fetches and parses a tool by ID
func FetchTool(ctx context.Context, client *kbapi.ClientWithResponses, toolId string, diagnostics *diag.Diagnostics) *ToolModel {
	toolResp, err := client.GetAgentBuilderToolsToolidWithResponse(ctx, toolId)
	if err != nil {
		diagnostics.AddWarning("Tool fetch failed", fmt.Sprintf("Unable to get tool %s: %v", toolId, err))
		return nil
	}

	if toolResp.StatusCode() != http.StatusOK {
		diagnostics.AddWarning("Tool fetch failed", fmt.Sprintf("Unable to get tool %s: HTTP %d", toolId, toolResp.StatusCode()))
		return nil
	}

	var toolData map[string]interface{}
	if err := json.Unmarshal(toolResp.Body, &toolData); err != nil {
		diagnostics.AddWarning("Tool parse failed", fmt.Sprintf("Unable to parse tool %s: %v", toolId, err))
		return nil
	}

	toolIdVal := ""
	if id, ok := toolData["id"].(string); ok {
		toolIdVal = id
	}

	toolType := ""
	if t, ok := toolData["type"].(string); ok {
		toolType = t
	}

	description := ""
	if desc, ok := toolData["description"].(string); ok {
		description = desc
	}

	var tags []string
	if tagsData, ok := toolData["tags"].([]interface{}); ok {
		for _, tag := range tagsData {
			if tagStr, ok := tag.(string); ok {
				tags = append(tags, tagStr)
			}
		}
	}

	readOnly := false
	if ro, ok := toolData["readonly"].(bool); ok {
		readOnly = ro
	}

	// Extract and marshal configuration
	configJSON := "{}"
	if config, ok := toolData["configuration"].(map[string]interface{}); ok {
		configBytes, err := json.Marshal(config)
		if err != nil {
			diagnostics.AddWarning("Configuration marshal failed", fmt.Sprintf("Unable to marshal configuration for tool %s: %v", toolId, err))
			configJSON = "{}"
		} else {
			configJSON = string(configBytes)
		}
	}

	tagsList, listDiags := types.ListValueFrom(ctx, types.StringType, tags)
	diagnostics.Append(listDiags...)

	tool := &ToolModel{
		ID:            types.StringValue(toolIdVal),
		Type:          types.StringValue(toolType),
		Description:   types.StringValue(description),
		Tags:          tagsList,
		ReadOnly:      types.BoolValue(readOnly),
		Configuration: types.StringValue(configJSON),
	}

	return tool
}

// ExtractWorkflowIdFromTool extracts the workflow_id from a workflow-type tool's configuration
func ExtractWorkflowIdFromTool(ctx context.Context, tool *ToolModel, diagnostics *diag.Diagnostics) string {
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(tool.Configuration.ValueString()), &config); err != nil {
		diagnostics.AddWarning(
			"Configuration parse failed",
			fmt.Sprintf("Unable to parse configuration for workflow tool %s: %v", tool.ID.ValueString(), err),
		)
		return ""
	}

	workflowId, ok := config["workflow_id"].(string)
	if !ok || workflowId == "" {
		diagnostics.AddWarning(
			"Workflow ID missing",
			fmt.Sprintf("Tool %s is type 'workflow' but does not have a valid workflow_id in its configuration", tool.ID.ValueString()),
		)
		return ""
	}

	return workflowId
}

// FetchWorkflow fetches and parses a workflow by ID
func FetchWorkflow(ctx context.Context, client *kbapi.ClientWithResponses, workflowId string, diagnostics *diag.Diagnostics) *WorkflowModel {
	workflowResp, err := client.GetWorkflowsIdWithResponse(ctx, workflowId, addInternalOriginHeader)
	if err != nil {
		diagnostics.AddWarning("Workflow fetch failed", fmt.Sprintf("Unable to get workflow %s: %v", workflowId, err))
		return nil
	}

	if workflowResp.StatusCode() != http.StatusOK {
		diagnostics.AddWarning("Workflow fetch failed", fmt.Sprintf("Unable to get workflow %s: HTTP %d", workflowId, workflowResp.StatusCode()))
		return nil
	}

	if workflowResp.JSON200 == nil {
		diagnostics.AddWarning("Workflow parse failed", fmt.Sprintf("Workflow %s returned nil data", workflowId))
		return nil
	}

	return &WorkflowModel{
		ID:   types.StringValue(workflowResp.JSON200.Id),
		Yaml: types.StringValue(workflowResp.JSON200.Yaml),
	}
}
