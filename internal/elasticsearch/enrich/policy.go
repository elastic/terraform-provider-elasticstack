package enrich

import (
	"context"
	"regexp"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourcePolicy() *schema.Resource {
	policySchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"policy_type": {
			Description:  "The type of enrich policy, can be one of geo_match, match, range.",
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringInSlice([]string{"geo_match", "match", "range"}, false),
		},
		"name": {
			Description: "Name of the enrich policy to manage.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(1, 255),
				validation.StringNotInSlice([]string{".", ".."}, true),
				validation.StringMatch(regexp.MustCompile(`^[^-_+]`), "cannot start with -, _, +"),
				validation.StringMatch(regexp.MustCompile(`^[a-z0-9!$%&'()+.;=@[\]^{}~_-]+$`), "must contain lower case alphanumeric characters and selected punctuation, see: https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-create-data-stream.html#indices-create-data-stream-api-path-params"),
			),
		},
		"indices": {
			Description:  "Array of one or more source indices used to create the enrich index.",
			Type:         schema.TypeList,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.ListOfUniqueStrings,
		},
		"match_field": {
			Description: "Field in source indices used to match incoming documents.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(1, 255),
			),
		},
		"enrich_fields": {
			Description:  "Fields to add to matching incoming documents. These fields must be present in the source indices.",
			Type:         schema.TypeList,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.ListOfUniqueStrings,
		},
		"query": {
			Description:  "Query used to filter documents in the enrich index. The policy only uses documents matching this query to enrich incoming documents. Defaults to a match_all query.",
			Type:         schema.TypeString,
			Optional:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringIsJSON,
		},
	}

	utils.AddConnectionSchema(policySchema)

	return &schema.Resource{
		Description: "Managing Elasticsearch enrich policies, see: https://www.elastic.co/guide/en/elasticsearch/reference/current/enrich-apis.html",

		CreateContext: resourceEnrichPolicyPut,
		UpdateContext: resourceEnrichPolicyPut,
		ReadContext:   resourceEnrichPolicyRead,
		DeleteContext: resourceEnrichPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: policySchema,
	}
}

func resourceEnrichPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}
	name := d.Get("name").(string)
	policy, diags := elasticsearch.GetEnrichPolicy(ctx, client, name)
	if diags.HasError() {
		return diags
	}
	if err := d.Set("name", policy.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("policy_type", policy.Type); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("indices", policy.Indices); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("match_field", policy.MatchField); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("enrich_fields", policy.EnrichFields); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("query", policy.Query); err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceEnrichPolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}
	_, diags = elasticsearch.GetSettings(ctx, client)
	return diags
}

func resourceEnrichPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}
	_, diags = elasticsearch.GetSettings(ctx, client)
	return diags
}
