package prebuilt_rules

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func (r *PrebuiltRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages Elastic prebuilt detection rules. This resource installs and updates Elastic prebuilt rules and timelines. See https://www.elastic.co/guide/en/security/current/prebuilt-rules.html",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of this resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				Description: "An identifier for the space. If space_id is not provided, the default space is used.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("default"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"rules_installed": schema.Int64Attribute{
				Description: "Number of prebuilt rules that are installed.",
				Computed:    true,
			},
			"rules_not_installed": schema.Int64Attribute{
				Description: "Number of prebuilt rules that are not installed.",
				Computed:    true,
			},
			"rules_not_updated": schema.Int64Attribute{
				Description: "Number of prebuilt rules that have updates available.",
				Computed:    true,
			},
			"timelines_installed": schema.Int64Attribute{
				Description: "Number of prebuilt timelines that are installed.",
				Computed:    true,
			},
			"timelines_not_installed": schema.Int64Attribute{
				Description: "Number of prebuilt timelines that are not installed.",
				Computed:    true,
			},
			"timelines_not_updated": schema.Int64Attribute{
				Description: "Number of prebuilt timelines that have updates available.",
				Computed:    true,
			},
		},
	}
}
