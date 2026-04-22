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

package agentbuilderagent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Read refreshes the Terraform state with the latest data.
func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config agentDataSourceModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	kbClient, diags := d.client.GetKibanaClient(ctx, config.KibanaConnection)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	supported, sdkDiags := kbClient.EnforceMinVersion(ctx, minKibanaAgentBuilderAPIVersion)
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !supported {
		resp.Diagnostics.AddError("Unsupported server version",
			fmt.Sprintf("Agent Builder agents require Elastic Stack v%s or later.", minKibanaAgentBuilderAPIVersion))
		return
	}

	serverVersion, sdkDiags := kbClient.ServerVersion(ctx)
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}
	supportsAdvancedConfig := !serverVersion.LessThan(minVersionAdvancedAgentConfig)

	if !typeutils.IsKnown(config.AgentID) || config.AgentID.ValueString() == "" {
		resp.Diagnostics.AddError("Invalid configuration", "agent_id must be set.")
		return
	}

	// Datasource BoolAttribute has no schema Default in this framework version; treat unset as false.
	includeDeps := false
	if typeutils.IsKnown(config.IncludeDependencies) {
		includeDeps = config.IncludeDependencies.ValueBool()
	}

	client, err := kbClient.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError("unable to get Kibana client", err.Error())
		return
	}

	spaceID := "default"
	if typeutils.IsKnown(config.SpaceID) && config.SpaceID.ValueString() != "" {
		spaceID = config.SpaceID.ValueString()
	}

	agentID := config.AgentID.ValueString()
	if compID, idDiags := clients.CompositeIDFromStrFw(agentID); !idDiags.HasError() {
		agentID = compID.ResourceID
		if !typeutils.IsKnown(config.SpaceID) || config.SpaceID.ValueString() == "" {
			spaceID = compID.ClusterID
		}
	}

	agent, agentDiags := kibanaoapi.GetAgent(ctx, client, spaceID, agentID)
	resp.Diagnostics.Append(agentDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if agent == nil {
		resp.Diagnostics.AddError("Agent not found", fmt.Sprintf("Unable to fetch agent with ID %s", agentID))
		return
	}

	diags = config.populateFromAPI(ctx, spaceID, agent)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	toolIDs := agentToolIDsInOrder(agent)
	switch {
	case len(toolIDs) == 0:
		config.Tools = nil
	case !includeDeps:
		config.Tools = make([]toolModel, 0, len(toolIDs))
		for _, tid := range toolIDs {
			config.Tools = append(config.Tools, toolModelFromToolRef(spaceID, tid))
		}
	default:
		// Fetch each tool and track workflow IDs for workflow-type tools.
		// These are "tool-embedded" workflows whose YAML is surfaced on the tool itself.
		toolWorkflowIDSet := make(map[string]struct{})
		toolsByID := make(map[string]*models.Tool)

		for _, toolID := range toolIDs {
			tool, toolDiags := kibanaoapi.GetTool(ctx, client, spaceID, toolID)
			resp.Diagnostics.Append(toolDiags...)
			if resp.Diagnostics.HasError() {
				return
			}
			if tool == nil {
				continue
			}
			toolsByID[toolID] = tool

			if tool.Type == "workflow" {
				if workflowID, ok := tool.Configuration["workflow_id"].(string); ok && workflowID != "" {
					toolWorkflowIDSet[workflowID] = struct{}{}
				}
			}
		}

		// Fetch workflows referenced by tools (for embedding YAML into the tool model).
		// The workflow API requires 9.4+, so error if workflow-type tools were found on an older server.
		workflowsByID := make(map[string]*models.Workflow)
		if len(toolWorkflowIDSet) > 0 && !supportsAdvancedConfig {
			resp.Diagnostics.AddError(
				"Unsupported server version",
				fmt.Sprintf(
					"This agent has workflow-type tools whose configuration cannot be exported: "+
						"the workflow API requires Elastic Stack v%s or later.",
					minVersionAdvancedAgentConfig,
				),
			)
			return
		}
		for workflowID := range toolWorkflowIDSet {
			workflow, wDiags := kibanaoapi.GetWorkflow(ctx, client, spaceID, workflowID)
			resp.Diagnostics.Append(wDiags...)
			if resp.Diagnostics.HasError() {
				return
			}
			if workflow != nil {
				workflowsByID[workflowID] = workflow
			}
		}

		// Convert tools to state models (same order as on the agent).
		config.Tools = make([]toolModel, 0, len(toolIDs))
		for _, toolID := range toolIDs {
			tool := toolsByID[toolID]
			if tool == nil {
				continue
			}
			tm, tmDiags := toolModelFromAPI(ctx, spaceID, tool, workflowsByID)
			resp.Diagnostics.Append(tmDiags...)
			if resp.Diagnostics.HasError() {
				return
			}
			config.Tools = append(config.Tools, tm)
		}
	}

	config.IncludeDependencies = types.BoolValue(includeDeps)

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}

// agentToolIDsInOrder returns unique tool IDs from the agent configuration, preserving first-seen order.
func agentToolIDsInOrder(agent *models.Agent) []string {
	if agent == nil {
		return nil
	}
	seen := make(map[string]struct{})
	var out []string
	for _, tc := range agent.Configuration.Tools {
		for _, id := range tc.ToolIDs {
			if id == "" {
				continue
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			out = append(out, id)
		}
	}
	return out
}

// toolModelFromToolRef builds a minimal tool row (composite id, space, tool id) without calling the tools API.
func toolModelFromToolRef(spaceID, toolID string) toolModel {
	return toolModel{
		ID:                        types.StringValue((&clients.CompositeID{ClusterID: spaceID, ResourceID: toolID}).String()),
		SpaceID:                   types.StringValue(spaceID),
		ToolID:                    types.StringValue(toolID),
		Type:                      types.StringNull(),
		Description:               types.StringNull(),
		Tags:                      types.SetNull(types.StringType),
		ReadOnly:                  types.BoolNull(),
		Configuration:             types.StringNull(),
		WorkflowID:                types.StringNull(),
		WorkflowConfigurationYaml: customtypes.NewNormalizedYamlNull(),
	}
}

// toolModelFromAPI converts a models.Tool (and optionally its workflow) into a toolModel.
func toolModelFromAPI(ctx context.Context, spaceID string, tool *models.Tool, workflowsByID map[string]*models.Workflow) (toolModel, diag.Diagnostics) {
	var tm toolModel
	var diags diag.Diagnostics

	tm.ID = types.StringValue((&clients.CompositeID{ClusterID: spaceID, ResourceID: tool.ID}).String())
	tm.SpaceID = types.StringValue(spaceID)
	tm.ToolID = types.StringValue(tool.ID)
	tm.Type = types.StringValue(tool.Type)

	if tool.Description != nil {
		tm.Description = types.StringValue(*tool.Description)
	} else {
		tm.Description = types.StringNull()
	}

	if len(tool.Tags) > 0 {
		tags, tagDiags := types.SetValueFrom(ctx, types.StringType, tool.Tags)
		diags.Append(tagDiags...)
		tm.Tags = tags
	} else {
		tm.Tags = types.SetNull(types.StringType)
	}

	tm.ReadOnly = types.BoolValue(tool.ReadOnly)

	if tool.Configuration != nil {
		configJSON, err := json.Marshal(tool.Configuration)
		if err != nil {
			diags.AddError("Configuration Error", "Failed to marshal configuration to JSON: "+err.Error())
			return tm, diags
		}
		tm.Configuration = types.StringValue(string(configJSON))
	} else {
		tm.Configuration = types.StringNull()
	}

	if tool.Type == "workflow" {
		if workflowID, ok := tool.Configuration["workflow_id"].(string); ok && workflowID != "" {
			tm.WorkflowID = types.StringValue(workflowID)
			if workflow, found := workflowsByID[workflowID]; found {
				tm.WorkflowConfigurationYaml = customtypes.NewNormalizedYamlValue(workflow.Yaml)
			} else {
				tm.WorkflowConfigurationYaml = customtypes.NewNormalizedYamlNull()
			}
		} else {
			tm.WorkflowID = types.StringNull()
			tm.WorkflowConfigurationYaml = customtypes.NewNormalizedYamlNull()
		}
	} else {
		tm.WorkflowID = types.StringNull()
		tm.WorkflowConfigurationYaml = customtypes.NewNormalizedYamlNull()
	}

	return tm, diags
}
