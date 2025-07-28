package enrich

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
)

func (r *enrichPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetResourceSchema()
}

func GetResourceSchema() resourceschema.Schema {
	return resourceschema.Schema{
		MarkdownDescription: "Managing Elasticsearch enrich policies. See: https://www.elastic.co/guide/en/elasticsearch/reference/current/enrich-apis.html",
		Blocks: map[string]resourceschema.Block{
			"elasticsearch_connection": providerschema.GetEsFWConnectionBlock("elasticsearch_connection", false),
		},
		Attributes: map[string]resourceschema.Attribute{
			"id": resourceschema.StringAttribute{
				MarkdownDescription: "Internal identifier of the resource",
				Computed:            true,
			},
			"name": resourceschema.StringAttribute{
				MarkdownDescription: "Name of the enrich policy to manage.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"policy_type": resourceschema.StringAttribute{
				MarkdownDescription: "The type of enrich policy, can be one of geo_match, match, range.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("geo_match", "match", "range"),
				},
			},
			"indices": resourceschema.ListAttribute{
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
			"match_field": resourceschema.StringAttribute{
				MarkdownDescription: "Field in source indices used to match incoming documents.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"enrich_fields": resourceschema.ListAttribute{
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
			"query": resourceschema.StringAttribute{
				MarkdownDescription: "Query used to filter documents in the enrich index. The policy only uses documents matching this query to enrich incoming documents. Defaults to a match_all query.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^\{.*\}$`), "must be valid JSON"),
				},
			},
			"execute": resourceschema.BoolAttribute{
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

func GetDataSourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Returns information about an enrich policy. See: https://www.elastic.co/guide/en/elasticsearch/reference/current/get-enrich-policy-api.html",
		Blocks: map[string]schema.Block{
			"elasticsearch_connection": providerschema.GetEsFWConnectionBlock("elasticsearch_connection", false),
		},
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal identifier of the resource",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the policy.",
				Required:            true,
			},
			"policy_type": schema.StringAttribute{
				MarkdownDescription: "The type of enrich policy, can be one of geo_match, match, range.",
				Computed:            true,
			},
			"indices": schema.ListAttribute{
				MarkdownDescription: "Array of one or more source indices used to create the enrich index.",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"match_field": schema.StringAttribute{
				MarkdownDescription: "Field in source indices used to match incoming documents.",
				Computed:            true,
			},
			"enrich_fields": schema.ListAttribute{
				MarkdownDescription: "Fields to add to matching incoming documents. These fields must be present in the source indices.",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"query": schema.StringAttribute{
				MarkdownDescription: "Query used to filter documents in the enrich index. The policy only uses documents matching this query to enrich incoming documents. Defaults to a match_all query.",
				Computed:            true,
			},
		},
	}
}