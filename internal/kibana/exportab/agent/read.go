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

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/exportab"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Read refreshes the Terraform state with the latest data.
func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config dataSourceModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverVersion, sdkDiags := d.client.ServerVersion(ctx)
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if serverVersion.LessThan(minKibanaAgentBuilderAPIVersion) {
		resp.Diagnostics.AddError("Unsupported server version",
			fmt.Sprintf("Agent Builder agents require Elastic Stack v%s or later.", minKibanaAgentBuilderAPIVersion))
		return
	}

	oapiClient, err := d.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError("unable to get Kibana client", err.Error())
		return
	}

	spaceID := "default"
	if typeutils.IsKnown(config.SpaceID) {
		spaceID = config.SpaceID.ValueString()
	}

	agentID := config.ID.ValueString()

	apiResp, err := oapiClient.API.GetAgentBuilderAgentsIdWithResponse(ctx, agentID)
	if err != nil {
		resp.Diagnostics.AddError("API call failed", fmt.Sprintf("Unable to get agent: %v", err))
		return
	}

	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Unexpected API response",
			fmt.Sprintf("Unexpected status code from server: got HTTP %d, response: %s", apiResp.StatusCode(), string(apiResp.Body)),
		)
		return
	}

	var agentData map[string]any
	if err := json.Unmarshal(apiResp.Body, &agentData); err != nil {
		resp.Diagnostics.AddError("JSON parsing failed", fmt.Sprintf("Unable to parse agent response: %v", err))
		return
	}

	agentJSON, err := json.Marshal(agentData)
	if err != nil {
		resp.Diagnostics.AddError("JSON marshaling failed", fmt.Sprintf("Unable to marshal agent to JSON: %v", err))
		return
	}

	compositeID := &clients.CompositeID{ClusterID: spaceID, ResourceID: agentID}

	var state dataSourceModel
	state.ID = types.StringValue(compositeID.String())
	state.SpaceID = types.StringValue(spaceID)
	state.Agent = types.StringValue(string(agentJSON))
	state.IncludeDependencies = config.IncludeDependencies

	includeDeps := false
	if typeutils.IsKnown(config.IncludeDependencies) {
		includeDeps = config.IncludeDependencies.ValueBool()
	}

	state.Tools = []exportab.ToolModel{}
	state.Workflows = []exportab.WorkflowModel{}

	if includeDeps {
		tools := []exportab.ToolModel{}
		workflows := []exportab.WorkflowModel{}

		toolIDs := exportab.ExtractToolIDs(agentData)
		workflowIDSet := make(map[string]bool)

		for _, toolID := range toolIDs {
			tool := exportab.FetchTool(ctx, oapiClient.API, toolID, &resp.Diagnostics)
			if tool == nil {
				continue
			}
			tools = append(tools, *tool)

			if tool.Type.ValueString() == "workflow" {
				workflowID := exportab.ExtractWorkflowIDFromTool(ctx, tool, &resp.Diagnostics)
				if workflowID != "" {
					workflowIDSet[workflowID] = true
				}
			}
		}

		for workflowID := range workflowIDSet {
			workflow := exportab.FetchWorkflow(ctx, oapiClient.API, workflowID, &resp.Diagnostics)
			if workflow != nil {
				workflows = append(workflows, *workflow)
			}
		}

		state.Tools = tools
		state.Workflows = workflows
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
