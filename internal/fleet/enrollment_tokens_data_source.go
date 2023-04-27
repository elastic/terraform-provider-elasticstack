package fleet

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceEnrollmentTokens() *schema.Resource {
	enrollmentTokenSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the data source.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"policy_id": {
			Description: "The identifier of the target agent policy. When provided, only the enrollment tokens associated with this agent policy will be selected. Omit this value to select all enrollment tokens.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"tokens": {
			Description: "A list of enrollment tokens.",
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"key_id": {
						Description: "The unique identifier of the enrollment token.",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"api_key": {
						Description: "The API key.",
						Type:        schema.TypeString,
						Computed:    true,
						Sensitive:   true,
					},
					"api_key_id": {
						Description: "The API key identifier.",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"created_at": {
						Description: "The time at which the enrollment token was created.",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"name": {
						Description: "The name of the enrollment token.",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"active": {
						Description: "Indicates if the enrollment token is active.",
						Type:        schema.TypeBool,
						Computed:    true,
					},
					"policy_id": {
						Description: "The identifier of the associated agent policy.",
						Type:        schema.TypeString,
						Computed:    true,
					},
				},
			},
		},
	}

	return &schema.Resource{
		Description: "Retrieves Elasticsearch API keys used to enroll Elastic Agents in Fleet. See: https://www.elastic.co/guide/en/fleet/current/fleet-enrollment-tokens.html",

		ReadContext: dataSourceEnrollmentTokensRead,

		Schema: enrollmentTokenSchema,
	}
}

func dataSourceEnrollmentTokensRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fleetClient, diags := getFleetClient(d, meta)
	if diags.HasError() {
		return diags
	}

	policyID := d.Get("policy_id").(string)

	allTokens, diags := fleet.AllEnrollmentTokens(ctx, fleetClient)
	if diags.HasError() {
		return diags
	}

	var enrollmentTokens []map[string]any
	for _, v := range allTokens {
		if policyID != "" && v.PolicyId != nil && *v.PolicyId != policyID {
			continue
		}

		keyData := map[string]any{
			"api_key":    v.ApiKey,
			"api_key_id": v.ApiKeyId,
			"created_at": v.CreatedAt,
			"active":     v.Active,
		}
		if v.Name != nil {
			keyData["name"] = *v.Name
		}
		if v.PolicyId != nil {
			keyData["policy_id"] = *v.PolicyId
		}

		enrollmentTokens = append(enrollmentTokens, keyData)
	}

	if enrollmentTokens != nil {
		if err := d.Set("tokens", enrollmentTokens); err != nil {
			return diag.FromErr(err)
		}
	}

	if policyID != "" {
		d.SetId(policyID)
	} else {
		hash, err := utils.StringToHash(fleetClient.URL)
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(*hash)
	}

	return diags
}
