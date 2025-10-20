package default_data_view

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func (r *DefaultDataViewResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = getSchema()
}

func getSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages the default Kibana data view. See the [Kibana Data Views API documentation](https://www.elastic.co/docs/api/doc/kibana/v8/operation/operation-setdefaultdatailviewdefault) for more information.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Internal identifier of the resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"data_view_id": schema.StringAttribute{
				MarkdownDescription: "The data view identifier to set as default. NOTE: The API does not validate whether it is a valid identifier. Use `null` to unset the default data view.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"force": schema.BoolAttribute{
				MarkdownDescription: "Update an existing default data view identifier. If set to false and a default data view already exists, the operation will fail.",
				Optional:            true,
			},
			"skip_delete": schema.BoolAttribute{
				MarkdownDescription: "If set to true, the default data view will not be unset when the resource is destroyed. The existing default data view will remain unchanged.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"space_id": schema.StringAttribute{
				MarkdownDescription: "The Kibana space ID to set the default data view in. Defaults to `default`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("default"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}
