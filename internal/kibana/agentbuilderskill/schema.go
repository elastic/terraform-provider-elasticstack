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

package agentbuilderskill

import (
	"context"
	"regexp"

	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *SkillResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = getSchema()
}

func getSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages Kibana Agent Builder skills. Skills are reusable markdown instructions that agents can reference. " +
			"See the [Agent Builder API documentation](https://www.elastic.co/docs/api/doc/kibana/group/endpoint-agent-builder) for more information.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The composite ID of the skill: `<space_id>/<skill_id>`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"skill_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The skill ID. Required; the API does not auto-generate skill IDs.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"space_id": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				Default:             stringdefault.StaticString(defaultSpaceID),
				MarkdownDescription: "An identifier for the Kibana space. If not provided, the default space is used.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Human-readable name for the skill.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"description": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Description of what the skill does.",
			},
			"content": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Skill instructions content as markdown.",
			},
			"tool_ids": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Set of tool IDs from the tool registry that this skill references.",
			},
			"referenced_content": schema.ListNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Ordered list of referenced-content entries. Up to 100 entries; order is preserved.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Name of the referenced content.",
						},
						"relative_path": schema.StringAttribute{
							Required: true,
							MarkdownDescription: "Relative path of the referenced content. Must start with `./` " +
								"(e.g., `./runbooks/standard.md`). Sent to and received from the API as `relativePath`.",
							Validators: []validator.String{
								stringvalidator.RegexMatches(
									regexp.MustCompile(`^\./`),
									"relative_path must start with ./",
								),
							},
						},
						"content": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Content of the reference.",
						},
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"kibana_connection": providerschema.GetKbFWConnectionBlock(),
		},
	}
}
