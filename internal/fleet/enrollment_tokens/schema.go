package enrollment_tokens

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func (d *enrollmentTokensDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = getSchema()
}

func getSchema() schema.Schema {
	return schema.Schema{
		Description: "Retrieves Elasticsearch API keys used to enroll Elastic Agents in Fleet. See the [Fleet enrollment tokens documentation](https://www.elastic.co/guide/en/fleet/current/fleet-enrollment-tokens.html) for more details.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
			},
			"policy_id": schema.StringAttribute{
				Description: "The identifier of the target agent policy. When provided, only the enrollment tokens associated with this agent policy will be selected. Omit this value to select all enrollment tokens.",
				Optional:    true,
			},
			"tokens": schema.ListNestedAttribute{
				Description: "A list of enrollment tokens.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key_id": schema.StringAttribute{
							Description: "The unique identifier of the enrollment token.",
							Computed:    true,
						},
						"api_key": schema.StringAttribute{
							Description: "The API key.",
							Computed:    true,
							Sensitive:   true,
						},
						"api_key_id": schema.StringAttribute{
							Description: "The API key identifier.",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "The time at which the enrollment token was created.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the enrollment token.",
							Computed:    true,
						},
						"active": schema.BoolAttribute{
							Description: "Indicates if the enrollment token is active.",
							Computed:    true,
						},
						"policy_id": schema.StringAttribute{
							Description: "The identifier of the associated agent policy.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func getTokenType() attr.Type {
	return getSchema().Attributes["tokens"].GetType().(attr.TypeWithElementType).ElementType()
}
