package import_saved_objects

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &Resource{}
var _ resource.ResourceWithConfigure = &Resource{}

// TODO - Uncomment these lines when we're using a kibana client which supports create_new_copies and compatibility_mode
// create_new_copies and compatibility_mode aren't supported by the current version of the Kibana client
// We can add these ourselves once https://github.com/elastic/terraform-provider-elasticstack/pull/372 is merged

// var _ resource.ResourceWithConfigValidators = &Resource{}

// func (r *Resource) ConfigValidators(context.Context) []resource.ConfigValidator {
// 	return []resource.ConfigValidator{
// 		resourcevalidator.Conflicting(
// 			path.MatchRoot("create_new_copies"),
// 			path.MatchRoot("overwrite"),
// 			path.MatchRoot("compatibility_mode"),
// 		),
// 	}
// }

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Imports saved objects from the referenced file",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Generated ID for the import.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				Description: "An identifier for the space. If space_id is not provided, the default space is used.",
				Optional:    true,
			},
			"ignore_import_errors": schema.BoolAttribute{
				Description: "If set to true, errors during the import process will not fail the configuration application",
				Optional:    true,
			},
			// create_new_copies and compatibility_mode aren't supported by the current version of the Kibana client
			// We can add these ourselves once https://github.com/elastic/terraform-provider-elasticstack/pull/372 is merged
			// "create_new_copies": schema.BoolAttribute{
			// 	Description: "Creates copies of saved objects, regenerates each object ID, and resets the origin. When used, potential conflict errors are avoided.",
			// 	Optional:    true,
			// },
			"overwrite": schema.BoolAttribute{
				Description: "Overwrites saved objects when they already exist. When used, potential conflict errors are automatically resolved by overwriting the destination object.",
				Optional:    true,
			},
			// "compatibility_mode": schema.BoolAttribute{
			// 	Description: "Applies various adjustments to the saved objects that are being imported to maintain compatibility between different Kibana versions. Use this option only if you encounter issues with imported saved objects.",
			// 	Optional:    true,
			// },
			"file_contents": schema.StringAttribute{
				Description: "The contents of the exported saved objects file.",
				Required:    true,
			},

			"success": schema.BoolAttribute{
				Description: "Indicates when the import was successfully completed. When set to false, some objects may not have been created. For additional information, refer to the errors and success_results properties.",
				Computed:    true,
			},
			"success_count": schema.Int64Attribute{
				Description: "Indicates the number of successfully imported records.",
				Computed:    true,
			},
			"errors": schema.ListAttribute{
				Computed: true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":    types.StringType,
						"type":  types.StringType,
						"title": types.StringType,
						"error": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"type": types.StringType,
							},
						},
						"meta": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"icon":  types.StringType,
								"title": types.StringType,
							},
						},
					},
				},
			},
			"success_results": schema.ListAttribute{
				Computed: true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":             types.StringType,
						"type":           types.StringType,
						"destination_id": types.StringType,
						"meta": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"icon":  types.StringType,
								"title": types.StringType,
							},
						},
					},
				},
			},
		},
	}
}

type Resource struct {
	client *clients.ApiClient
}

func resourceReady(r *Resource, dg *diag.Diagnostics) bool {
	if r.client == nil {
		dg.AddError(
			"Unconfigured Client",
			"Expected configured client. Please report this issue to the provider developers.",
		)

		return false
	}
	return true
}

func (r *Resource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(request.ProviderData)
	response.Diagnostics.Append(diags...)
	r.client = client
}

func (r *Resource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_kibana_import_saved_objects"
}

type modelV0 struct {
	ID                 types.String `tfsdk:"id"`
	SpaceID            types.String `tfsdk:"space_id"`
	IgnoreImportErrors types.Bool   `tfsdk:"ignore_import_errors"`
	// CreateNewCopies    types.Bool   `tfsdk:"create_new_copies"`
	Overwrite types.Bool `tfsdk:"overwrite"`
	// CompatibilityMode  types.Bool   `tfsdk:"compatibility_mode"`
	FileContents   types.String `tfsdk:"file_contents"`
	Success        types.Bool   `tfsdk:"success"`
	SuccessCount   types.Int64  `tfsdk:"success_count"`
	Errors         types.List   `tfsdk:"errors"`
	SuccessResults types.List   `tfsdk:"success_results"`
}
