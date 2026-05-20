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
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// getDataSourceSchema returns the schema for the agent data source.
// The kibana_connection block is injected by the envelope and must not be
// declared here.
func getDataSourceSchema(_ context.Context) dsschema.Schema {
	return dsschema.Schema{
		Description: "Export an Agent Builder agent by ID, optionally including its tools and workflows. See https://www.elastic.co/docs/api/doc/kibana/operation/operation-get-agent-builder-agents-id",
		MarkdownDescription: "Export an Agent Builder agent by ID, optionally including its tools and workflows. " +
			"See the [Agent Builder API documentation](https://www.elastic.co/guide/en/kibana/current/agent-builder-api.html).",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The composite ID of the agent: `<space_id>/<agent_id>`.",
			},
			"agent_id": dsschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The agent ID.",
			},
			"space_id": dsschema.StringAttribute{
				Description: "An identifier for the space. If space_id is not provided, the default space is used.",
				Optional:    true,
			},
			"include_dependencies": dsschema.BoolAttribute{
				Description:         "If true, exports the agent along with its tools and workflows. If omitted, false is used (tool rows only list id, space_id, and tool_id unless this is true).",
				MarkdownDescription: "If `true`, exports the agent along with its tools and workflows. If omitted, `false` is used (tool rows only list `id`, `space_id`, and `tool_id` unless this is `true`).",
				Optional:            true,
			},
			"name": dsschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The agent name.",
			},
			"description": dsschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The agent description.",
			},
			"avatar_color": dsschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Hex color code for the agent avatar (e.g., `#BFDBFF`).",
			},
			"avatar_symbol": dsschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Symbol or initials for the agent avatar (e.g., `SI`).",
			},
			"labels": dsschema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "List of labels for the agent.",
			},
			"instructions": dsschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Optional system instructions that define the agent behavior.",
			},
			"skill_ids": dsschema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Set of skill IDs assigned to the agent. Requires Elastic Stack 9.4.0 or later.",
			},
			"tools": dsschema.ListNestedAttribute{
				Description: "Tools attached to the agent. When include_dependencies is true, each entry includes full tool data and workflow YAML for workflow-type tools. " +
					"When false, only id (composite space/tool), space_id, and tool_id are set.",
				Computed: true,
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: map[string]dsschema.Attribute{
						"id": dsschema.StringAttribute{
							Description: "The tool ID to look up.",
							Computed:    true,
						},
						"space_id": dsschema.StringAttribute{
							Description: "An identifier for the space. If space_id is not provided, the default space is used.",
							Computed:    true,
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
						"workflow_id": dsschema.StringAttribute{
							Description: "The ID of the referenced workflow. Only populated for workflow-type tools. Requires Elastic Stack v9.4.0 or later.",
							Computed:    true,
						},
						"workflow_configuration_yaml": dsschema.StringAttribute{
							Description: "The YAML configuration of the referenced workflow. Only populated for workflow-type tools. Requires Elastic Stack v9.4.0 or later.",
							Computed:    true,
							CustomType:  customtypes.NormalizedYamlType{},
						},
					},
				},
			},
		},
	}
}
