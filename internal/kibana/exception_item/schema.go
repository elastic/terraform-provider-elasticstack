package exception_item

import (
	"context"
	_ "embed"

	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

//go:embed resource-description.md
var exceptionItemResourceDescription string

func (r *exceptionItemResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetSchema()
}

func GetSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: exceptionItemResourceDescription,
		Blocks: map[string]schema.Block{
			"kibana_connection": providerschema.GetKbFWConnectionBlock(),
		},
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal identifier of the resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				MarkdownDescription: "An identifier for the space. If space_id is not provided, the default space is used.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("default"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"list_id": schema.StringAttribute{
				MarkdownDescription: "The exception list's human readable string identifier that this item belongs to.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"item_id": schema.StringAttribute{
				MarkdownDescription: "Human readable string identifier for the exception item, e.g. `my-exception-item`.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Exception item name.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Describes the exception item.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of exception item. Currently only `simple` is supported.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("simple"),
				Validators: []validator.String{
					stringvalidator.OneOf("simple"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"namespace_type": schema.StringAttribute{
				MarkdownDescription: "Determines whether the exception item is available in all Kibana spaces or just the space in which it is created. Values: `single` (only available in the space in which it is created) or `agnostic` (available in all Kibana spaces).",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("single"),
				Validators: []validator.String{
					stringvalidator.OneOf("single", "agnostic"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"entries": schema.StringAttribute{
				MarkdownDescription: "The query (fields, values, and logic) used to prevent rules from generating alerts. Must be a JSON array of exception entry objects.",
				CustomType:          jsontypes.NormalizedType{},
				Required:            true,
			},
			"comments": schema.StringAttribute{
				MarkdownDescription: "Array of comment objects. Must be a JSON array of comment objects with a `comment` field.",
				CustomType:          jsontypes.NormalizedType{},
				Optional:            true,
			},
			"expire_time": schema.StringAttribute{
				MarkdownDescription: "The exception item's expiration date, in ISO format. This field is only available for regular exception items, not endpoint exceptions.",
				Optional:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "String array containing words and phrases to help categorize exception items.",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"os_types": schema.ListAttribute{
				MarkdownDescription: "Use this field to specify the operating system. Valid values: linux, macos, windows.",
				ElementType:         types.StringType,
				Optional:            true,
			},
		},
	}
}
