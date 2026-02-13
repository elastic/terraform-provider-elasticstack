package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/export_ab"
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

	oapiClient, err := d.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError("unable to get Kibana client", err.Error())
		return
	}

	spaceId := "default"
	if !config.SpaceID.IsNull() && !config.SpaceID.IsUnknown() {
		spaceId = config.SpaceID.ValueString()
	}

	agentId := config.ID.ValueString()

	apiResp, err := oapiClient.API.GetAgentBuilderAgentsIdWithResponse(ctx, agentId)
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

	var agentData map[string]interface{}
	if err := json.Unmarshal(apiResp.Body, &agentData); err != nil {
		resp.Diagnostics.AddError("JSON parsing failed", fmt.Sprintf("Unable to parse agent response: %v", err))
		return
	}

	agentJSON, err := json.Marshal(agentData)
	if err != nil {
		resp.Diagnostics.AddError("JSON marshaling failed", fmt.Sprintf("Unable to marshal agent to JSON: %v", err))
		return
	}

	compositeID := &clients.CompositeId{ClusterId: spaceId, ResourceId: agentId}

	var state dataSourceModel
	state.ID = types.StringValue(compositeID.String())
	state.SpaceID = types.StringValue(spaceId)
	state.Agent = types.StringValue(string(agentJSON))
	state.IncludeDependencies = config.IncludeDependencies

	includeDeps := false
	if !config.IncludeDependencies.IsNull() && !config.IncludeDependencies.IsUnknown() {
		includeDeps = config.IncludeDependencies.ValueBool()
	}

	state.Tools = []export_ab.ToolModel{}
	state.Workflows = []export_ab.WorkflowModel{}

	if includeDeps {
		tools := []export_ab.ToolModel{}
		workflows := []export_ab.WorkflowModel{}

		toolIds := export_ab.ExtractToolIds(agentData)
		workflowIdSet := make(map[string]bool)

		for _, toolId := range toolIds {
			tool := export_ab.FetchTool(ctx, oapiClient.API, toolId, &resp.Diagnostics)
			if tool == nil {
				continue
			}
			tools = append(tools, *tool)

			if tool.Type.ValueString() == "workflow" {
				workflowId := export_ab.ExtractWorkflowIdFromTool(ctx, tool, &resp.Diagnostics)
				if workflowId != "" {
					workflowIdSet[workflowId] = true
				}
			}
		}

		for workflowId := range workflowIdSet {
			workflow := export_ab.FetchWorkflow(ctx, oapiClient.API, workflowId, &resp.Diagnostics)
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
