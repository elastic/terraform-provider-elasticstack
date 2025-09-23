package agent_configuration

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *resourceAgentConfiguration) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates or updates an APM agent configuration. See https://www.elastic.co/docs/solutions/observability/apm/apm-agent-central-configuration.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Internal identifier of the resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"service_name": schema.StringAttribute{
				Description: "The name of the service.",
				Required:    true,
			},
			"service_environment": schema.StringAttribute{
				Description: "The environment of the service.",
				Optional:    true,
			},
			"agent_name": schema.StringAttribute{
				Description: "The agent name is used by the UI to determine which settings to display.",
				Optional:    true,
			},
			"settings": schema.MapAttribute{
				Description: "Agent configuration settings.",
				Required:    true,
				ElementType: types.StringType,
			},
		},
	}
}
