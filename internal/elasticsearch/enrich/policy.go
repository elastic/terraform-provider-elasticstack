package enrich

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceEnrichPolicy() *schema.Resource {
	policySchema := map[string]*schema.Schema{
		"name": {
			Description: "Name of the enrich policy to manage.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"policy_type": {
			Description:  "The type of enrich policy, can be one of geo_match, match, range.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringInSlice([]string{"geo_match", "match", "range"}, false),
		},
		"indices": {
			Description: "Array of one or more source indices used to create the enrich index.",
			Type:        schema.TypeSet,
			Required:    true,
			ForceNew:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
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
			Description: "Fields to add to matching incoming documents. These fields must be present in the source indices.",
			Type:        schema.TypeSet,
			Required:    true,
			ForceNew:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"query": {
			Description:      "Query used to filter documents in the enrich index. The policy only uses documents matching this query to enrich incoming documents. Defaults to a match_all query.",
			Type:             schema.TypeString,
			Optional:         true,
			ForceNew:         true,
			ValidateFunc:     validation.StringIsJSON,
			DiffSuppressFunc: utils.DiffJsonSuppress,
		},
		"execute": {
			Description: "Whether to call the execute API function in order to create the enrich index.",
			Type:        schema.TypeBool,
			Optional:    true,
			ForceNew:    true,
			Default:     true,
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
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}

	compName, diags := clients.CompositeIdFromStr(d.Id())
	if diags.HasError() {
		return diags
	}
	policy, diags := elasticsearch.GetEnrichPolicy(ctx, client, compName.ResourceId)
	if policy == nil && diags == nil {
		tflog.Warn(ctx, fmt.Sprintf(`Enrich policy "%s" not found, removing from state`, compName.ResourceId))
		d.SetId("")
		return diags
	}
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
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}

	name := d.Get("name").(string)
	id, diags := client.ID(ctx, name)
	if diags.HasError() {
		return diags
	}
	policy := &models.EnrichPolicy{
		Type:         d.Get("policy_type").(string),
		Name:         name,
		Indices:      utils.ExpandStringSet(d.Get("indices").(*schema.Set)),
		MatchField:   d.Get("match_field").(string),
		EnrichFields: utils.ExpandStringSet(d.Get("enrich_fields").(*schema.Set)),
	}

	if query, ok := d.GetOk("query"); ok {
		policy.Query = query.(string)
	}

	if diags = elasticsearch.PutEnrichPolicy(ctx, client, policy); diags.HasError() {
		return diags
	}
	d.SetId(id.String())
	if d.Get("execute").(bool) {
		diags := elasticsearch.ExecuteEnrichPolicy(ctx, client, name)
		if diags.HasError() {
			return diags
		}
	}
	return resourceEnrichPolicyRead(ctx, d, meta)
}

func resourceEnrichPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}

	compName, diags := clients.CompositeIdFromStr(d.Id())
	if diags.HasError() {
		return diags
	}
	return elasticsearch.DeleteEnrichPolicy(ctx, client, compName.ResourceId)
}
