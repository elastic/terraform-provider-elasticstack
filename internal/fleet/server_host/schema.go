package server_host

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *serverHostResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema.Description = "Creates a new Fleet Server Host."
	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "The ID of this resource.",
			Computed:    true,
		},
		"host_id": schema.StringAttribute{
			Description: "Unique identifier of the Fleet server host.",
			Computed:    true,
			Optional:    true,
		},
		"name": schema.StringAttribute{
			Description: "The name of the Fleet server host.",
			Required:    true,
		},
		"hosts": schema.ListAttribute{
			Description: "A list of hosts.",
			Required:    true,
			ElementType: types.StringType,
			Validators: []validator.List{
				listvalidator.SizeAtLeast(1),
			},
		},
		"default": schema.BoolAttribute{
			Description: "Set as default.",
			Optional:    true,
		},
		"space_ids": schema.ListAttribute{
			Description: "The Kibana space IDs where this server host is available. When set, the server host will be created and managed within the specified space.",
			ElementType: types.StringType,
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.List{
				listplanmodifier.UseStateForUnknown(),
			},
		},
	}
}
