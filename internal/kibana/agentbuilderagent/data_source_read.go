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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
