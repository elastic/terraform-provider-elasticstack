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
		MarkdownDescription: "Manages Kibana [data views](https://www.elastic.co/guide/en/kibana/current/data-views-api.html)",
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
									Required:            true,
									MarkdownDescription: "The ID of the field format. Valid values include: `boolean`, `color`, `date`, `duration`, `number`, `percent`, `relative_date`, `static_lookup`, `string`, `truncate`, `url`.",
								},
								"params": schema.SingleNestedAttribute{
									Optional: true,
									Attributes: map[string]schema.Attribute{
										"pattern": schema.StringAttribute{
											Optional:            true,
											MarkdownDescription: "Pattern for formatting the field value.",
										},
										"urltemplate": schema.StringAttribute{
											Optional:            true,
											MarkdownDescription: "URL template for the field value.",
										},
										"labeltemplate": schema.StringAttribute{
											Optional:            true,
											MarkdownDescription: "Label template for the field value.",
										},
										"input_format": schema.StringAttribute{
											Optional:            true,
											MarkdownDescription: "Input format for duration fields (e.g., `hours`, `minutes`).",
										},
										"output_format": schema.StringAttribute{
											Optional:            true,
											MarkdownDescription: "Output format for duration fields (e.g., `humanizePrecise`, `humanize`).",
										},
										"output_precision": schema.Int64Attribute{
											Optional:            true,
											MarkdownDescription: "Precision for duration output.",
										},
										"include_space_with_suffix": schema.BoolAttribute{
											Optional:            true,
											MarkdownDescription: "Whether to include a space before the suffix in duration format.",
										},
										"use_short_suffix": schema.BoolAttribute{
											Optional:            true,
											MarkdownDescription: "Whether to use short suffixes in duration format.",
										},
										"timezone": schema.StringAttribute{
											Optional:            true,
											MarkdownDescription: "Timezone for date formatting (e.g., `America/New_York`).",
										},
										"field_type": schema.StringAttribute{
											Optional:            true,
											MarkdownDescription: "Field type for color formatting (e.g., `string`, `number`).",
										},
										"colors": schema.ListNestedAttribute{
											Optional:            true,
											MarkdownDescription: "Color rules for the field.",
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"range": schema.StringAttribute{
														Optional:            true,
														MarkdownDescription: "Range for the color rule (e.g., `-Infinity:Infinity`).",
													},
													"regex": schema.StringAttribute{
														Optional:            true,
														MarkdownDescription: "Regex pattern for the color rule.",
													},
													"text": schema.StringAttribute{
														Optional:            true,
														MarkdownDescription: "Text color in hex format.",
													},
													"background": schema.StringAttribute{
														Optional:            true,
														MarkdownDescription: "Background color in hex format.",
													},
												},
											},
										},
										"field_length": schema.Int64Attribute{
											Optional:            true,
											MarkdownDescription: "Length to truncate the field value.",
										},
										"transform": schema.StringAttribute{
											Optional:            true,
											MarkdownDescription: "Transform to apply to string fields (e.g., `upper`, `lower`).",
										},
										"lookup_entries": schema.ListNestedAttribute{
											Optional:            true,
											MarkdownDescription: "Key-value pairs for static lookup.",
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"key": schema.StringAttribute{
														Required:            true,
														MarkdownDescription: "Key for the lookup entry.",
													},
													"value": schema.StringAttribute{
														Required:            true,
														MarkdownDescription: "Value for the lookup entry.",
													},
												},
											},
										},
										"unknown_key_value": schema.StringAttribute{
											Optional:            true,
											MarkdownDescription: "Value to display when key is not found in lookup.",
										},
										"type": schema.StringAttribute{
											Optional:            true,
											MarkdownDescription: "Type of URL format (e.g., `a`, `img`, `audio`).",
										},
										"width": schema.Int64Attribute{
											Optional:            true,
											MarkdownDescription: "Width for image type URLs.",
										},
										"height": schema.Int64Attribute{
											Optional:            true,
											MarkdownDescription: "Height for image type URLs.",
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

func getFieldFormatParamsColorsElemType() attr.Type {
	return getFieldFormatParamsAttrTypes()["colors"].(attr.TypeWithElementType).ElementType()
}

func getFieldFormatParamsLookupEntryElemType() attr.Type {
	return getFieldFormatParamsAttrTypes()["lookup_entries"].(attr.TypeWithElementType).ElementType()
}
