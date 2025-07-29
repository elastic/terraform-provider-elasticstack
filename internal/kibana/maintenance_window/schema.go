package maintenance_window

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

func (r *MaintenanceWindowResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = getSchema()
}

func getSchema() schema.Schema {
	return schema.Schema{
		Description: "Manages Kibana data views",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Generated ID for the data view.",
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
			"title": schema.StringAttribute{
				Description: "The name of the maintenance window.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the current maintenance window is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"custom_schedule": schema.SingleNestedAttribute{
				Description: "A set schedule over which the maintenance window applies.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"start": schema.StringAttribute{
						Description: "The start date and time of the schedule, provided in ISO 8601 format and set to the UTC timezone. For example: `2025-03-12T12:00:00.000Z`.",
						Required:    true,
					},
					"duration": schema.StringAttribute{
						Description: "The duration of the schedule. It allows values in `<integer><unit>` format. `<unit>` is one of `d`, `h`, `m`, or `s` for hours, minutes, seconds. For example: `1d`, `5h`, `30m`, `5000s`.",
						Required:    true,
						// TODO: Validation
					},
					"timezone": schema.StringAttribute{
						Description: "The timezone of the schedule. The default timezone is UTC.",
						Optional:    true,
						// PlanModifiers: []planmodifier.String{
						// 	stringplanmodifier.UseStateForUnknown(),
						// 	stringplanmodifier.RequiresReplace(),
						// },
					},
					// TODO
					// "recurring": schema.SingleNestedAttribute{
					// 	Description: "A set schedule over which the maintenance window applies.",
					// 	Required:    true,
					// 	Attributes: map[string]schema.Attribute{
					// 		"end": schema.StringAttribute{
					// 			Description: "The start date and time of the schedule, provided in ISO 8601 format and set to the UTC timezone. For example: `2025-03-12T12:00:00.000Z`.",
					// 			Required:    true,
					// 		},
					// 		"every": schema.StringAttribute{
					// 			Description: "The duration of the schedule. It allows values in `<integer><unit>` format. `<unit>` is one of `d`, `h`, `m`, or `s` for hours, minutes, seconds. For example: `1d`, `5h`, `30m`, `5000s`.",
					// 			Required:    true,
					// 			// TODO: Validation
					// 		},
					// 		"on_week_day": schema.StringAttribute{
					// 			Description: "The start date and time of the schedule, provided in ISO 8601 format and set to the UTC timezone. For example: `2025-03-12T12:00:00.000Z`.",
					// 			Required:    true,
					// 		},
					// 		"on_month_day": schema.StringAttribute{
					// 			Description: "The start date and time of the schedule, provided in ISO 8601 format and set to the UTC timezone. For example: `2025-03-12T12:00:00.000Z`.",
					// 			Required:    true,
					// 		},
					// 		"on_month": schema.StringAttribute{
					// 			Description: "The duration of the schedule. It allows values in `<integer><unit>` format. `<unit>` is one of `d`, `h`, `m`, or `s` for hours, minutes, seconds. For example: `1d`, `5h`, `30m`, `5000s`.",
					// 			Required:    true,
					// 			// TODO: Validation
					// 		},
					// 	},
					// },
				},
			},
			"scope": schema.SingleNestedAttribute{
				Description: "An object that narrows the scope of what is affected by this maintenance window.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"alerting": schema.SingleNestedAttribute{
						Description: "A set schedule over which the maintenance window applies.",
						Required:    true,
						Attributes: map[string]schema.Attribute{
							"kql": schema.StringAttribute{
								Description: "A filter written in Kibana Query Language (KQL).",
								Required:    true,
							},
						},
					},
				},
			},
		},
	}
}

// func getDataViewAttrTypes() map[string]attr.Type {
// 	return getSchema().Attributes["data_view"].GetType().(attr.TypeWithAttributeTypes).AttributeTypes()
// }

// func getFieldAttrElemType() attr.Type {
// 	return getDataViewAttrTypes()["field_attrs"].(attr.TypeWithElementType).ElementType()
// }

// func getRuntimeFieldMapElemType() attr.Type {
// 	return getDataViewAttrTypes()["runtime_field_map"].(attr.TypeWithElementType).ElementType()
// }

// func getFieldFormatElemType() attr.Type {
// 	return getDataViewAttrTypes()["field_formats"].(attr.TypeWithElementType).ElementType()
// }

// func getFieldFormatAttrTypes() map[string]attr.Type {
// 	return getFieldFormatElemType().(attr.TypeWithAttributeTypes).AttributeTypes()
// }

// func getFieldFormatParamsAttrTypes() map[string]attr.Type {
// 	return getFieldFormatAttrTypes()["params"].(attr.TypeWithAttributeTypes).AttributeTypes()
// }

// func getFieldFormatParamsColorsElemType() attr.Type {
// 	return getFieldFormatParamsAttrTypes()["colors"].(attr.TypeWithElementType).ElementType()
// }

// func getFieldFormatParamsLookupEntryElemType() attr.Type {
// 	return getFieldFormatParamsAttrTypes()["lookup_entries"].(attr.TypeWithElementType).ElementType()
// }
