package api_key

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
)

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = r.getSchema()
}

func (r *Resource) getSchema() schema.Schema {
	return schema.Schema{
		Description: "Creates an API key for access without requiring basic authentication. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-create-api-key.html",
		Blocks: map[string]schema.Block{
			"elasticsearch_connection": providerschema.GetEsFWConnectionBlock("elasticsearch_connection", false),
		},
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Internal identifier of the resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key_id": schema.StringAttribute{
				Description: "Unique id for this API key.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Specifies the name for this API key.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 1024),
					stringvalidator.RegexMatches(regexp.MustCompile(`^([[:graph:]]| )+$`), "must contain alphanumeric characters (a-z, A-Z, 0-9), spaces, punctuation, and printable symbols in the Basic Latin (ASCII) block. Leading or trailing whitespace is not allowed"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role_descriptors": schema.StringAttribute{
				Description: "Role descriptors for this API key.",
				CustomType:  jsontypes.NormalizedType{},
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					r.requiresReplaceIfUpdateNotSupported(),
				},
			},
			"expiration": schema.StringAttribute{
				Description: "Expiration time for the API key. By default, API keys never expire.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"expiration_timestamp": schema.Int64Attribute{
				Description: "Expiration time in milliseconds for the API key. By default, API keys never expire.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"metadata": schema.StringAttribute{
				Description: "Arbitrary metadata that you want to associate with the API key.",
				Optional:    true,
				Computed:    true,
				CustomType:  jsontypes.NormalizedType{},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					r.requiresReplaceIfUpdateNotSupported(),
				},
			},
			"api_key": schema.StringAttribute{
				Description: "Generated API Key.",
				Sensitive:   true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"encoded": schema.StringAttribute{
				Description: "API key credentials which is the Base64-encoding of the UTF-8 representation of the id and api_key joined by a colon (:).",
				Sensitive:   true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *Resource) requiresReplaceIfUpdateNotSupported() planmodifier.String {
	return stringplanmodifier.RequiresReplaceIf(
		func(ctx context.Context, res planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
			version, _ := r.clusterVersionOfLastRead(ctx, res.Private)

			resp.RequiresReplace = version != nil && version.LessThan(MinVersionWithUpdate)
		},
		"Requires replace if the server does not support update",
		"Requries replace if the server does not support update",
	)
}
