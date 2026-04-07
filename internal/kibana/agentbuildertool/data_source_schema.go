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

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (d *ToolDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
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
