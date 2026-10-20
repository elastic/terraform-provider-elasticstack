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

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/kbschema"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getResourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages Kibana Agent Builder agents. See the [Agent Builder API documentation](https://www.elastic.co/docs/api/doc/kibana/group/endpoint-agent-builder) for more information.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The composite ID of the agent: `<space_id>/<agent_id>`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			attrAgentID: schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The agent ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			attrSpaceID: kbschema.ResourceSpaceIDAttribute(),
			attrName: schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The agent name.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			attrDescription: schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The agent description.",
			},
			attrAvatarColor: schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Hex color code for the agent avatar (e.g., `#BFDBFF`).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			attrAvatarSymbol: schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Symbol or initials for the agent avatar (e.g., `SI`).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			attrLabels: schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Set of labels for the agent.",
			},
			attrTools: schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Set of tool IDs that the agent can use.",
			},
			"skill_ids": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Set of skill IDs to assign to the agent. Requires Elastic Stack 9.4.0 or later.",
			},
			attrAccessControl: schema.SingleNestedAttribute{
				Optional: true,
				Computed: true,
				MarkdownDescription: "Access control for the agent. Requires Elastic Stack 9.5.0 or later when set. " +
					"`access_mode` is one of `public`, `shared`, or `private`. When omitted, the server-assigned mode is " +
					"read back and not managed by Terraform; removing the block from configuration relinquishes management " +
					"of the mode (the server keeps its current mode) rather than resetting it.",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					attrAccessMode: schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "The access mode: `public`, `shared`, or `private`.",
						Validators: []validator.String{
							stringvalidator.OneOf("public", "shared", "private"),
						},
					},
				},
			},
			"instructions": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional system instructions that define the agent behavior.",
			},
		},
	}
}
