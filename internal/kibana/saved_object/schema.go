package saved_object

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &Resource{}
var _ resource.ResourceWithConfigure = &Resource{}
var _ resource.ResourceWithModifyPlan = &Resource{}
var _ resource.ResourceWithConfigValidators = &Resource{}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Import a Kibana saved object",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Extracted ID from the object",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			// This is needed as the user provided object cannot be cleaned up
			// see https://discuss.hashicorp.com/t/using-modifyplan-with-a-custom-provider/47690
			// We thus store a copy of the object with some fields removed
			"imported": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Kibana object imported.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				Description: "An identifier for the space. If space_id is not provided, the default space is used.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"object": schema.StringAttribute{
				Description: "Kibana object to import in JSON format",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Extracted type from the object",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

type Resource struct {
	client *clients.ApiClient
}

func (r *Resource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(request.ProviderData)
	response.Diagnostics.Append(diags...)
	r.client = client
}

func (r *Resource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_kibana_saved_object"
}

func (r *Resource) ConfigValidators(context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		&KibanaSavedObjectValidator{},
	}
}
