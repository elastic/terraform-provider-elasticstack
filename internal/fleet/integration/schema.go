package integration

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func (r *integrationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema.Description = "Manage installation of a Fleet integration package."
	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "The ID of this resource.",
			Computed:    true,
		},
		"name": schema.StringAttribute{
			Description: "The integration package name.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"version": schema.StringAttribute{
			Description: "The integration package version.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"force": schema.BoolAttribute{
			Description: "Set to true to force the requested action.",
			Optional:    true,
		},
		"skip_destroy": schema.BoolAttribute{
			Description: "Set to true if you do not wish the integration package to be uninstalled at destroy time, and instead just remove the integration package from the Terraform state.",
			Optional:    true,
		},
	}
}
