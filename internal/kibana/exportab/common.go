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

package exportab

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
func addInternalOriginHeader(_ context.Context, req *http.Request) error {
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

// ExtractToolIDs extracts tool IDs from the agent configuration
func ExtractToolIDs(agentData map[string]any) []string {
	var toolIDs []string

	if config, ok := agentData["configuration"].(map[string]any); ok {
		if tools, ok := config["tools"].([]any); ok {
			for _, tool := range tools {
				if toolMap, ok := tool.(map[string]any); ok {
					if ids, ok := toolMap["tool_ids"].([]any); ok {
						for _, id := range ids {
							if idStr, ok := id.(string); ok {
								toolIDs = append(toolIDs, idStr)
							}
						}
					}
				}
			}
		}
	}

	return toolIDs
}

// FetchTool fetches and parses a tool by ID
func FetchTool(ctx context.Context, client *kbapi.ClientWithResponses, toolID string, diagnostics *diag.Diagnostics) *ToolModel {
	toolResp, err := client.GetAgentBuilderToolsToolidWithResponse(ctx, toolID)
	if err != nil {
		diagnostics.AddWarning("Tool fetch failed", fmt.Sprintf("Unable to get tool %s: %v", toolID, err))
		return nil
	}

	if toolResp.StatusCode() != http.StatusOK {
		diagnostics.AddWarning("Tool fetch failed", fmt.Sprintf("Unable to get tool %s: HTTP %d", toolID, toolResp.StatusCode()))
		return nil
	}

	var toolData map[string]any
	if err := json.Unmarshal(toolResp.Body, &toolData); err != nil {
		diagnostics.AddWarning("Tool parse failed", fmt.Sprintf("Unable to parse tool %s: %v", toolID, err))
		return nil
	}

	toolIDVal := ""
	if id, ok := toolData["id"].(string); ok {
		toolIDVal = id
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
	if tagsData, ok := toolData["tags"].([]any); ok {
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
	if config, ok := toolData["configuration"].(map[string]any); ok {
		configBytes, err := json.Marshal(config)
		if err != nil {
			diagnostics.AddWarning("Configuration marshal failed", fmt.Sprintf("Unable to marshal configuration for tool %s: %v", toolID, err))
			configJSON = "{}"
		} else {
			configJSON = string(configBytes)
		}
	}

	tagsList, listDiags := types.ListValueFrom(ctx, types.StringType, tags)
	diagnostics.Append(listDiags...)

	tool := &ToolModel{
		ID:            types.StringValue(toolIDVal),
		Type:          types.StringValue(toolType),
		Description:   types.StringValue(description),
		Tags:          tagsList,
		ReadOnly:      types.BoolValue(readOnly),
		Configuration: types.StringValue(configJSON),
	}

	return tool
}

// ExtractWorkflowIDFromTool extracts the workflow_id from a workflow-type tool's configuration
func ExtractWorkflowIDFromTool(_ context.Context, tool *ToolModel, diagnostics *diag.Diagnostics) string {
	var config map[string]any
	if err := json.Unmarshal([]byte(tool.Configuration.ValueString()), &config); err != nil {
		diagnostics.AddWarning(
			"Configuration parse failed",
			fmt.Sprintf("Unable to parse configuration for workflow tool %s: %v", tool.ID.ValueString(), err),
		)
		return ""
	}

	workflowID, ok := config["workflow_id"].(string)
	if !ok || workflowID == "" {
		diagnostics.AddWarning(
			"Workflow ID missing",
			fmt.Sprintf("Tool %s is type 'workflow' but does not have a valid workflow_id in its configuration", tool.ID.ValueString()),
		)
		return ""
	}

	return workflowID
}

// FetchWorkflow fetches and parses a workflow by ID
func FetchWorkflow(ctx context.Context, client *kbapi.ClientWithResponses, workflowID string, diagnostics *diag.Diagnostics) *WorkflowModel {
	workflowResp, err := client.GetWorkflowsIdWithResponse(ctx, workflowID, addInternalOriginHeader)
	if err != nil {
		diagnostics.AddWarning("Workflow fetch failed", fmt.Sprintf("Unable to get workflow %s: %v", workflowID, err))
		return nil
	}

	if workflowResp.StatusCode() != http.StatusOK {
		diagnostics.AddWarning("Workflow fetch failed", fmt.Sprintf("Unable to get workflow %s: HTTP %d", workflowID, workflowResp.StatusCode()))
		return nil
	}

	if workflowResp.JSON200 == nil {
		diagnostics.AddWarning("Workflow parse failed", fmt.Sprintf("Workflow %s returned nil data", workflowID))
		return nil
	}

	return &WorkflowModel{
		ID:   types.StringValue(workflowResp.JSON200.Id),
		Yaml: types.StringValue(workflowResp.JSON200.Yaml),
	}
}
