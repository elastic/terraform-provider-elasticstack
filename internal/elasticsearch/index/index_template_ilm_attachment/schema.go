package index_template_ilm_attachment

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = getSchema()
}

func getSchema() schema.Schema {
	return schema.Schema{
		Description: "Attaches an ILM policy to a Fleet-managed or externally-managed index template " +
			"by creating/updating the @custom component template with the lifecycle setting. " +
			"**Important:** Do NOT use this resource for index templates already managed by Terraform. " +
			"Instead, set `index.lifecycle.name` directly in the `elasticstack_elasticsearch_index_template` " +
			"or `elasticstack_elasticsearch_component_template` resource settings.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Internal identifier of the resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"index_template": schema.StringAttribute{
				Description: "Name of the index template to attach the ILM policy to. " +
					"For Fleet-managed templates, this is typically the template name (e.g., 'logs-system.syslog').",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"lifecycle_name": schema.StringAttribute{
				Description: "Name of the ILM policy to attach to the index template.",
				Required:    true,
			},
		},
	}
}
