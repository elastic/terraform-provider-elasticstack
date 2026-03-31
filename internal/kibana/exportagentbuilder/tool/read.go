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

package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
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

	supported, sdkDiags := d.client.EnforceMinVersion(ctx, minKibanaAgentBuilderAPIVersion)
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !supported {
		resp.Diagnostics.AddError("Unsupported server version",
			fmt.Sprintf("Agent Builder tools require Elastic Stack v%s or later.", minKibanaAgentBuilderAPIVersion))
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

	toolID := config.ID.ValueString()
	if compID, compDiags := clients.CompositeIDFromStrFw(toolID); !compDiags.HasError() {
		toolID = compID.ResourceID
		if !typeutils.IsKnown(config.SpaceID) {
			spaceID = compID.ClusterID
		}
	}

	tool, diags := kibanaoapi.GetTool(ctx, oapiClient, spaceID, toolID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if tool == nil {
		resp.Diagnostics.AddError("Tool not found", fmt.Sprintf("Unable to fetch tool with ID %s", toolID))
		return
	}

	compositeID := &clients.CompositeID{ClusterID: spaceID, ResourceID: toolID}

	var state dataSourceModel
	state.ID = types.StringValue(compositeID.String())
	state.SpaceID = types.StringValue(spaceID)
	state.ToolID = types.StringValue(tool.ID)
	state.Type = types.StringValue(tool.Type)

	if tool.Description != nil {
		state.Description = types.StringValue(*tool.Description)
	} else {
		state.Description = types.StringNull()
	}

	if len(tool.Tags) > 0 {
		tags, tagDiags := types.ListValueFrom(ctx, types.StringType, tool.Tags)
		resp.Diagnostics.Append(tagDiags...)
		state.Tags = tags
	} else {
		state.Tags = types.ListNull(types.StringType)
	}

	state.ReadOnly = types.BoolValue(tool.ReadOnly)

	if tool.Configuration != nil {
		configJSON, err := json.Marshal(tool.Configuration)
		if err != nil {
			resp.Diagnostics.AddError("Configuration Error", "Failed to marshal configuration to JSON: "+err.Error())
			return
		}
		state.Configuration = types.StringValue(string(configJSON))
	} else {
		state.Configuration = types.StringNull()
	}

	if config.IncludeWorkflow.ValueBool() {
		supported, sdkDiags := d.client.EnforceMinVersion(ctx, minKibanaAgentBuilderWorkflowAPIVersion)
		resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if !supported {
			resp.Diagnostics.AddError(
				"Unsupported server version",
				fmt.Sprintf("Exporting workflow configuration requires Elastic Stack v%s or later.", minKibanaAgentBuilderWorkflowAPIVersion),
			)
			return
		}

		if tool.Type != "workflow" {
			resp.Diagnostics.AddError(
				"Invalid use of include_workflow",
				fmt.Sprintf("include_workflow is true but the tool type is %q, not \"workflow\".", tool.Type),
			)
			return
		}

		workflowIDRaw, ok := tool.Configuration["workflow_id"]
		if !ok {
			resp.Diagnostics.AddError("Missing workflow_id", "Tool configuration does not contain a workflow_id.")
			return
		}
		workflowID, ok := workflowIDRaw.(string)
		if !ok || workflowID == "" {
			resp.Diagnostics.AddError("Invalid workflow_id", "workflow_id in tool configuration is not a valid string.")
			return
		}

		workflow, wDiags := kibanaoapi.GetWorkflow(ctx, oapiClient, spaceID, workflowID)
		resp.Diagnostics.Append(wDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		if workflow == nil {
			resp.Diagnostics.AddError("Workflow not found", fmt.Sprintf("Unable to fetch workflow with ID %s.", workflowID))
			return
		}

		state.WorkflowID = types.StringValue(workflow.ID)
		state.WorkflowConfigurationYaml = customtypes.NewNormalizedYamlValue(workflow.Yaml)
	} else {
		state.WorkflowID = types.StringNull()
		state.WorkflowConfigurationYaml = customtypes.NewNormalizedYamlNull()
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
