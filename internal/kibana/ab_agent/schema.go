package ab_agent

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *AgentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = getSchema()
}

func getSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages Kibana Agent Builder agents. See the [Agent Builder API documentation](https://www.elastic.co/guide/en/kibana/current/agent-builder-api.html) for more information.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The agent ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The agent name.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The agent description.",
			},
			"avatar_color": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Hex color code for the agent avatar (e.g., `#BFDBFF`).",
			},
			"avatar_symbol": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Symbol or initials for the agent avatar (e.g., `SI`).",
			},
			"labels": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "List of labels for the agent.",
			},
			"tools": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "List of tool IDs that the agent can use.",
			},
			"instructions": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Instructions for the agent's behavior.",
			},
		},
	}
}
