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

package agentbuilderworkflow

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func (r *WorkflowResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = getSchema()
}

func getSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages Kibana Agent Builder workflows. See the [Workflows documentation](https://www.elastic.co/docs/explore-analyze/workflows) for more information.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The composite ID of the workflow: `<workflow_id>/<space_id>`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"workflow_id": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "The workflow ID. If not provided, it will be auto-generated. IDs are `workflow-<UUIDv4>`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				Default:             stringdefault.StaticString("default"),
				MarkdownDescription: "An identifier for the Kibana space. If not provided, the default space is used.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"configuration_yaml": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The YAML configuration for the workflow.",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The workflow name (extracted from YAML configuration).",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The workflow description (extracted from YAML configuration).",
			},
			"enabled": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the workflow is enabled (extracted from YAML configuration).",
			},
			"valid": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the workflow configuration is valid.",
			},
		},
	}
}
