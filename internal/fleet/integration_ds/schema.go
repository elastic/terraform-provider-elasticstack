package integration_ds

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func (d *integrationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema.Description = "Retrieves the latest version of an integration package in Fleet."
	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "The ID of this resource.",
			Computed:    true,
		},
		"name": schema.StringAttribute{
			Description: "The integration package name.",
			Required:    true,
		},
		"prerelease": schema.BoolAttribute{
			Description: "Include prerelease packages.",
			Optional:    true,
		},
		"version": schema.StringAttribute{
			Description: "The integration package version.",
			Computed:    true,
		},
	}
}
