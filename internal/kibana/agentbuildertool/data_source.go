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

package agentbuildertool

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getDataSourceSchema(_ context.Context) dsschema.Schema {
	return dsschema.Schema{
		Description: "Reads an Agent Builder tool by ID. See https://www.elastic.co/guide/en/kibana/current/agent-builder-api.html",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Description: "The tool ID to look up.",
				Required:    true,
			},
			"space_id": dsschema.StringAttribute{
				Description: "An identifier for the space. If space_id is not provided, the default space is used.",
				Optional:    true,
			},
			"tool_id": dsschema.StringAttribute{
				Description: "The ID of the tool.",
				Computed:    true,
			},
			"type": dsschema.StringAttribute{
				Description: "The type of the tool (esql, index_search, workflow, mcp).",
				Computed:    true,
			},
			"description": dsschema.StringAttribute{
				Description: "Description of what the tool does.",
				Computed:    true,
			},
			"tags": dsschema.SetAttribute{
				Description: "Tags for categorizing and organizing tools.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"readonly": dsschema.BoolAttribute{
				Description: "Whether the tool is read-only.",
				Computed:    true,
			},
			"configuration": dsschema.StringAttribute{
				Description: "The tool configuration in JSON format.",
				Computed:    true,
			},
			"include_workflow": dsschema.BoolAttribute{
				Description: "When true, the workflow referenced by this tool will also be included. Only valid when the tool type is `workflow`. Requires Kibana 9.4.0 or above. Defaults to false.",
				Optional:    true,
			},
			"workflow_id": dsschema.StringAttribute{
				Description: "The ID of the referenced workflow. Only populated when `include_workflow` is true.",
				Computed:    true,
			},
			"workflow_configuration_yaml": dsschema.StringAttribute{
				Description: "The YAML configuration of the referenced workflow. Only populated when `include_workflow` is true.",
				Computed:    true,
				CustomType:  customtypes.NormalizedYamlType{},
			},
		},
	}
}

func readToolDataSource(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	resourceID, spaceID string,
	config toolDataSourceModel,
) (toolDataSourceModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	oapiClient := client.GetKibanaOapiClient()

	tool, d := kibanaoapi.GetTool(ctx, oapiClient, spaceID, resourceID)
	diags.Append(d...)
	if diags.HasError() {
		return config, false, diags
	}
	if tool == nil {
		return config, false, diags
	}

	config.SpaceID = types.StringValue(spaceID)
	d = config.populateFromAPI(ctx, tool)
	diags.Append(d...)
	if diags.HasError() {
		return config, false, diags
	}

	if config.IncludeWorkflow.ValueBool() {
		supported, verDiags := client.EnforceMinVersion(ctx, minKibanaAgentBuilderWorkflowAPIVersion)
		diags.Append(verDiags...)
		if diags.HasError() {
			return config, false, diags
		}
		if !supported {
			diags.AddError(
				"Unsupported server version",
				fmt.Sprintf("Exporting workflow configuration requires Elastic Stack v%s or later.", minKibanaAgentBuilderWorkflowAPIVersion),
			)
			return config, false, diags
		}

		if tool.Type != "workflow" {
			diags.AddError(
				"Invalid use of include_workflow",
				fmt.Sprintf("include_workflow is true but the tool type is %q, not \"workflow\".", tool.Type),
			)
			return config, false, diags
		}

		workflowIDRaw, ok := tool.Configuration["workflow_id"]
		if !ok {
			diags.AddError("Missing workflow_id", "Tool configuration does not contain a workflow_id.")
			return config, false, diags
		}
		workflowID, ok := workflowIDRaw.(string)
		if !ok || workflowID == "" {
			diags.AddError("Invalid workflow_id", "workflow_id in tool configuration is not a valid string.")
			return config, false, diags
		}

		workflow, wDiags := kibanaoapi.GetWorkflow(ctx, oapiClient, spaceID, workflowID)
		diags.Append(wDiags...)
		if diags.HasError() {
			return config, false, diags
		}
		if workflow == nil {
			diags.AddError("Workflow not found", fmt.Sprintf("Unable to fetch workflow with ID %s.", workflowID))
			return config, false, diags
		}

		config.WorkflowID = types.StringValue(workflow.ID)
		config.WorkflowConfigurationYaml = customtypes.NewNormalizedYamlValue(workflow.Yaml)
	} else {
		config.WorkflowID = types.StringNull()
		config.WorkflowConfigurationYaml = customtypes.NewNormalizedYamlNull()
	}

	return config, true, diags
}

// NewDataSource is a helper function to simplify the provider implementation.
func NewDataSource() datasource.DataSource {
	return entitycore.NewKibanaDataSource[toolDataSourceModel](
		entitycore.ComponentKibana,
		"agentbuilder_tool",
		entitycore.KibanaDataSourceOptions[toolDataSourceModel]{
			Schema: getDataSourceSchema,
			Read:   readToolDataSource,
		},
	)
}
