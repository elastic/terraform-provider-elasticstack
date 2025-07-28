package enrich

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
)

func NewEnrichPolicyDataSource() datasource.DataSource {
	return &enrichPolicyDataSource{}
}

type enrichPolicyDataSource struct {
	client *clients.ApiClient
}

func (d *enrichPolicyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_elasticsearch_enrich_policy"
}

func (d *enrichPolicyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	d.client = client
}

func (d *enrichPolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = GetDataSourceSchema()
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
				MarkdownDescription: "Field from the source indices used to match incoming documents.",
				Computed:            true,
			},
			"enrich_fields": schema.ListAttribute{
				MarkdownDescription: "Fields to add to matching incoming documents. These fields must be present in the source indices.",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"query": schema.StringAttribute{
				MarkdownDescription: "Query used to filter documents in the enrich index for matching.",
				CustomType:          jsontypes.NormalizedType{},
				Computed:            true,
			},
		},
	}
}

func (d *enrichPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data EnrichPolicyData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policyName := data.Name.ValueString()
	client, diags := clients.MaybeNewApiClientFromFrameworkResource(ctx, data.ElasticsearchConnection, d.client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, sdkDiags := client.ID(ctx, policyName)
	resp.Diagnostics.Append(utils.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Id = types.StringValue(id.String())

	// Use the same read logic as the resource
	policy, sdkDiags := elasticsearch.GetEnrichPolicy(ctx, client, policyName)
	resp.Diagnostics.Append(utils.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if policy == nil {
		resp.Diagnostics.AddError("Policy not found", fmt.Sprintf("Enrich policy '%s' not found", policyName))
		return
	}

	// Convert model to framework types using shared function
	data.populateFromPolicy(ctx, policy, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
