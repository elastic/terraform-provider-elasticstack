package spaces

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Schema defines the schema for the data source.
func (d *dataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve and get information about all existing Kibana spaces. See https://www.elastic.co/guide/en/kibana/master/spaces-api-get-all.html",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Generated ID for the spaces.",
				Computed:    true,
			},
			"spaces": schema.ListNestedAttribute{
				Description: "The list of spaces.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Internal identifier of the resource.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The display name for the space.",
							Required:    true,
						},
						"description": schema.StringAttribute{
							Description: "The description for the space.",
							Optional:    true,
						},
						"disabled_features": schema.ListAttribute{
							Description: "The list of disabled features for the space. To get a list of available feature IDs, use the Features API (https://www.elastic.co/guide/en/kibana/master/features-api-get.html).",
							ElementType: types.StringType,
							Computed:    true,
						},
						"initials": schema.StringAttribute{
							Description: "The initials shown in the space avatar. By default, the initials are automatically generated from the space name. Initials must be 1 or 2 characters.",
							Computed:    true,
						},
						"color": schema.StringAttribute{
							Description: "The hexadecimal color code used in the space avatar. By default, the color is automatically generated from the space name.",
							Computed:    true,
						},
						"image_url": schema.StringAttribute{
							Description: "The data-URL encoded image to display in the space avatar.",
							Optional:    true,
						},
						"solution": schema.StringAttribute{
							Description: "The solution view for the space. Valid options are `security`, `oblt`, `es`, or `classic`.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}
