package integration

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func (r *integrationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema.Version = 1
	resp.Schema.Description = `Installs or uninstalls a Fleet integration package. The Kibana Fleet UI can be
used to view available packages. Additional information for managing integration
packages can be found [here](https://www.elastic.co/guide/en/fleet/current/install-uninstall-integration-assets.html).

To prevent the package from being uninstalled when the resource is destroyed,
set ` + "`skip_destroy` to `true`."
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
		},
		"force": schema.BoolAttribute{
			Description: "Set to true to force the requested action.",
			Optional:    true,
		},
		"prerelease": schema.BoolAttribute{
			Description: "Set to true to allow installation of prerelease (beta, non-GA) packages.",
			Optional:    true,
		},
		"ignore_mapping_update_errors": schema.BoolAttribute{
			Description: "Set to true to ignore mapping update errors during package installation.",
			Optional:    true,
		},
		"skip_data_stream_rollover": schema.BoolAttribute{
			Description: "Set to true to skip data stream rollover during package installation.",
			Optional:    true,
		},
		"ignore_constraints": schema.BoolAttribute{
			Description: "Set to true to ignore constraint errors during package installation.",
			Optional:    true,
		},
		"skip_destroy": schema.BoolAttribute{
			Description: "Set to true if you do not wish the integration package to be uninstalled at destroy time, and instead just remove the integration package from the Terraform state.",
			Optional:    true,
		},
		"space_id": schema.StringAttribute{
			Description: "The Kibana space ID where this integration package should be installed.",
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
	}
}
