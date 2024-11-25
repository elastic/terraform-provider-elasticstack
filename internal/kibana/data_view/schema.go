package data_view

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *DataViewResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"override": schema.BoolAttribute{
				Description: "Overrides an existing data view if a data view with the provided title already exists.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"data_view": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"title": schema.StringAttribute{
						Description: "Comma-separated list of data streams, indices, and aliases that you want to search. Supports wildcards (*).",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.LengthAtLeast(1),
						},
					},
					"name": schema.StringAttribute{
						Description: "The Data view name.",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"id": schema.StringAttribute{
						Description: "Saved object ID.",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
							stringplanmodifier.RequiresReplace(),
						},
					},
					"time_field_name": schema.StringAttribute{
						Description: "Timestamp field name, which you use for time-based Data views.",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"source_filters": schema.ListAttribute{
						Description: "List of field names you want to filter out in Discover.",
						ElementType: types.StringType,
						Optional:    true,
					},
					"field_attrs": schema.MapNestedAttribute{
						Description: "Map of field attributes by field name.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"custom_label": schema.StringAttribute{
									Description: "Custom label for the field.",
									Optional:    true,
								},
								"count": schema.Int64Attribute{
									Description: "Popularity count for the field.",
									Optional:    true,
								},
							},
						},
						Optional: true,
						PlanModifiers: []planmodifier.Map{
							mapplanmodifier.RequiresReplace(),
						},
					},
					"runtime_field_map": schema.MapNestedAttribute{
						Description: "Map of runtime field definitions by field name.",
						Optional:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"type": schema.StringAttribute{
									MarkdownDescription: "Mapping type of the runtime field. For more information, check [Field data types](https://www.elastic.co/guide/en/elasticsearch/reference/8.11/mapping-types.html).",
									Required:            true,
								},
								"script_source": schema.StringAttribute{
									Description: "Script of the runtime field.",
									Required:    true,
								},
							},
						},
					},
					"field_formats": schema.MapNestedAttribute{
						Description: "Map of field formats by field name.",
						Optional:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									Required: true,
								},
								"params": schema.SingleNestedAttribute{
									Optional: true,
									Attributes: map[string]schema.Attribute{
										"pattern": schema.StringAttribute{
											Optional: true,
										},
										"urltemplate": schema.StringAttribute{
											Optional: true,
										},
										"labeltemplate": schema.StringAttribute{
											Optional: true,
										},
									},
								},
							},
						},
					},
					"allow_no_index": schema.BoolAttribute{
						Description: "Allows the Data view saved object to exist before the data is available.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.RequiresReplace(),
						},
					},
					"namespaces": schema.ListAttribute{
						Description: "Array of space IDs for sharing the Data view between multiple spaces.",
						ElementType: types.StringType,
						Optional:    true,
						PlanModifiers: []planmodifier.List{
							listplanmodifier.RequiresReplace(),
						},
					},
				},
			},
		},
	}
}

func getDataViewAttrTypes() map[string]attr.Type {
	return getSchema().Attributes["data_view"].GetType().(attr.TypeWithAttributeTypes).AttributeTypes()
}

func getFieldAttrElemType() attr.Type {
	return getDataViewAttrTypes()["field_attrs"].(attr.TypeWithElementType).ElementType()
}

func getRuntimeFieldMapElemType() attr.Type {
	return getDataViewAttrTypes()["runtime_field_map"].(attr.TypeWithElementType).ElementType()
}

func getFieldFormatElemType() attr.Type {
	return getDataViewAttrTypes()["field_formats"].(attr.TypeWithElementType).ElementType()
}

func getFieldFormatAttrTypes() map[string]attr.Type {
	return getFieldFormatElemType().(attr.TypeWithAttributeTypes).AttributeTypes()
}

func getFieldFormatParamsAttrTypes() map[string]attr.Type {
	return getFieldFormatAttrTypes()["params"].(attr.TypeWithAttributeTypes).AttributeTypes()
}
