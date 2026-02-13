package ab_workflow

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func (r *WorkflowResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = getSchema()
}

func getSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages Kibana Agent Builder workflows. See the [Workflows API documentation](https://www.elastic.co/guide/en/kibana/current/workflows-api.html) for more information.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "The workflow ID. If not provided, it will be auto-generated. IDs are `workflow-<UUIDv4>`",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"configuration": schema.StringAttribute{
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
