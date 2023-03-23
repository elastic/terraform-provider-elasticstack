package enrich

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceEnrichPolicy() *schema.Resource {
	policySchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"name": {
			Description: "The name of the policy.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"policy_type": {
			Description: "The type of enrich policy, can be one of geo_match, match, range.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"indices": {
			Description: "Array of one or more source indices used to create the enrich index.",
			Type:        schema.TypeSet,
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"match_field": {
			Description: "Field in source indices used to match incoming documents.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"enrich_fields": {
			Description: "Fields to add to matching incoming documents. These fields must be present in the source indices.",
			Type:        schema.TypeSet,
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"query": {
			Description: "Query used to filter documents in the enrich index. The policy only uses documents matching this query to enrich incoming documents. Defaults to a match_all query.",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}

	utils.AddConnectionSchema(policySchema)

	return &schema.Resource{
		Description: "Returns information about an enrich policy. See: https://www.elastic.co/guide/en/elasticsearch/reference/current/get-enrich-policy-api.html",
		ReadContext: dataSourceEnrichPolicyRead,
		Schema:      policySchema,
	}
}

func dataSourceEnrichPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}

	policyId := d.Get("name").(string)
	id, diags := client.ID(ctx, policyId)
	if diags.HasError() {
		return diags
	}
	d.SetId(id.String())
	return resourceEnrichPolicyRead(ctx, d, meta)
}
