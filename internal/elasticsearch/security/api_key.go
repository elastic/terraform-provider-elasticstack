package security

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceApiKey() *schema.Resource {
	apikeySchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"name": {
			Description: "Specifies the name for this API key.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(1, 1024),
				validation.StringMatch(regexp.MustCompile(`^([[:graph:]]| )+$`), "must contain alphanumeric characters (a-z, A-Z, 0-9), spaces, punctuation, and printable symbols in the Basic Latin (ASCII) block. Leading or trailing whitespace is not allowed"),
			),
		},
		"role_descriptors": {
			Description:      "Role descriptors for this API key.",
			Type:             schema.TypeString,
			Optional:         true,
			ForceNew:         true,
			ValidateFunc:     validation.StringIsJSON,
			DiffSuppressFunc: utils.DiffJsonSuppress,
		},
		"expiration": {
			Description: "Expiration time for the API key. By default, API keys never expire.",
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
		},
		"metadata": {
			Description:      "Arbitrary metadata that you want to associate with the API key.",
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			ForceNew:         true,
			ValidateFunc:     validation.StringIsJSON,
			DiffSuppressFunc: utils.DiffJsonSuppress,
		},
	}

	utils.AddConnectionSchema(apikeySchema)

	return &schema.Resource{
		Description: "Creates an API key for access without requiring basic authentication. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-create-api-key.html",

		CreateContext: resourceSecurityApiKeyCreate,
		UpdateContext: resourceSecurityApiKeyUpdate,
		ReadContext:   resourceSecurityApiKeyRead,
		DeleteContext: resourceSecurityApiKeyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: apikeySchema,
	}
}

func resourceSecurityApiKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := clients.NewApiClient(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	nameId := d.Get("name").(string)

	var apikey models.ApiKey
	apikey.Name = nameId

	if v, ok := d.GetOk("expiration"); ok {
		apikey.Expiration = v.(string)
	}

	if v, ok := d.GetOk("role_descriptors"); ok {
		role_descriptors := make(map[string]models.Role)
		if err := json.NewDecoder(strings.NewReader(v.(string))).Decode(&role_descriptors); err != nil {
			return diag.FromErr(err)
		}
		apikey.RolesDescriptors = role_descriptors
	}

	if v, ok := d.GetOk("metadata"); ok {
		metadata := make(map[string]interface{})
		if err := json.NewDecoder(strings.NewReader(v.(string))).Decode(&metadata); err != nil {
			return diag.FromErr(err)
		}
		apikey.Metadata = metadata
	}

	putResponse, diags := client.PutElasticsearchApiKey(&apikey)

	if diags.HasError() {
		return diags
	}

	id, diags := client.ID(ctx, putResponse.Id)
	if diags.HasError() {
		return diags
	}

	if putResponse.ApiKey != "" {
		apikey.ApiKey = putResponse.ApiKey
	}
	if putResponse.EncodedApiKey != "" {
		apikey.EncodedApiKey = putResponse.EncodedApiKey
	}

	d.SetId(id.String())
	return resourceSecurityApiKeyRead(ctx, d, meta)
}

func resourceSecurityApiKeyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{diag.Diagnostic{
		Severity: diag.Error,
		Summary:  `Cannot update API Key`,
		Detail:   `update not currently supported.`,
	}}

	return diags
}

func resourceSecurityApiKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client, err := clients.NewApiClient(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	compId, diags := clients.CompositeIdFromStr(d.Id())
	if diags.HasError() {
		return diags
	}
	id := compId.ResourceId

	apikey, diags := client.GetElasticsearchApiKey(id)
	if apikey == nil && diags == nil {
		d.SetId("")
		return diags
	}
	if diags.HasError() {
		return diags
	}

	metadata, err := json.Marshal(apikey.Metadata)
	if err != nil {
		return diag.FromErr(err)
	}

	// set the fields
	if err := d.Set("name", apikey.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("expiration", apikey.Expiration); err != nil {
		return diag.FromErr(err)
	}

	if apikey.RolesDescriptors != nil {
		rolesDescriptors, err := json.Marshal(apikey.RolesDescriptors)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("role_descriptors", string(rolesDescriptors)); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := d.Set("metadata", string(metadata)); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceSecurityApiKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// client, err := clients.NewApiClient(d, meta)
	// if err != nil {
	// 	return diag.FromErr(err)
	// }
	// compId, diags := clients.CompositeIdFromStr(d.Id())
	// if diags.HasError() {
	// 	return diags
	// }

	// if diags := client.DeleteElasticsearchApiKey(compId.ResourceId); diags.HasError() {
	// 	return diags
	// }

	// d.SetId("")
	return diags
}
