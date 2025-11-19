package exception_container

import (
	"context"
	_ "embed"

	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
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
var exceptionContainerResourceDescription string

func (r *exceptionContainerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetSchema()
}

func GetSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: exceptionContainerResourceDescription,
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
				MarkdownDescription: "The exception list's human readable string identifier, e.g. `my-exception-list`.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the exception list.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Describes the exception list.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of exception list. Supported types: detection, endpoint, endpoint_blocklists, endpoint_events, endpoint_host_isolation_exceptions, endpoint_trusted_apps, endpoint_trusted_devices, rule_default.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("detection", "endpoint", "endpoint_blocklists", "endpoint_events", "endpoint_host_isolation_exceptions", "endpoint_trusted_apps", "endpoint_trusted_devices", "rule_default"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"namespace_type": schema.StringAttribute{
				MarkdownDescription: "Determines whether the exception container is available in all Kibana spaces or just the space in which it is created. Values: `single` (only available in the space in which it is created) or `agnostic` (available in all Kibana spaces).",
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
			"tags": schema.ListAttribute{
				MarkdownDescription: "String array containing words and phrases to help categorize exception containers.",
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
