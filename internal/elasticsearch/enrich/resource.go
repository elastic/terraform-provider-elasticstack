package enrich

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
)

func NewEnrichPolicyResource() resource.Resource {
	return &enrichPolicyResource{}
}

type enrichPolicyResource struct {
	client *clients.ApiClient
}

func (r *enrichPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_elasticsearch_enrich_policy"
}

func (r *enrichPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func (r *enrichPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetResourceSchema()
}

func GetResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Managing Elasticsearch enrich policies. See: https://www.elastic.co/guide/en/elasticsearch/reference/current/enrich-apis.html",
		Blocks: map[string]schema.Block{
			"elasticsearch_connection": providerschema.GetEsFWConnectionBlock("elasticsearch_connection", false),
		},
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal identifier of the resource",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the enrich policy to manage.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"policy_type": schema.StringAttribute{
				MarkdownDescription: "The type of enrich policy, can be one of geo_match, match, range.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("geo_match", "match", "range"),
				},
			},
			"indices": schema.ListAttribute{
				MarkdownDescription: "Array of one or more source indices used to create the enrich index.",
				ElementType:         types.StringType,
				Required:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"match_field": schema.StringAttribute{
				MarkdownDescription: "Field in source indices used to match incoming documents.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"enrich_fields": schema.ListAttribute{
				MarkdownDescription: "Fields to add to matching incoming documents. These fields must be present in the source indices.",
				ElementType:         types.StringType,
				Required:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"query": schema.StringAttribute{
				MarkdownDescription: "Query used to filter documents in the enrich index. The policy only uses documents matching this query to enrich incoming documents. Defaults to a match_all query.",
				Optional:            true,
				CustomType:          jsontypes.NormalizedType{},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"execute": schema.BoolAttribute{
				MarkdownDescription: "Whether to call the execute API function in order to create the enrich index.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}
