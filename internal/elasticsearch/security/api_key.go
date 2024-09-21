package security

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var APIKeyMinVersion = version.Must(version.NewVersion("8.0.0"))                // Enabled in 8.0
var APIKeyWithRestrictionMinVersion = version.Must(version.NewVersion("8.9.0")) // Enabled in 8.0

func ResourceApiKey() *schema.Resource {
	apikeySchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"key_id": {
			Description: "Unique id for this API key.",
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
		"expiration_timestamp": {
			Description: "Expiration time in milliseconds for the API key. By default, API keys never expire.",
			Type:        schema.TypeInt,
			Computed:    true,
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
		"api_key": {
			Description: "Generated API Key.",
			Type:        schema.TypeString,
			Sensitive:   true,
			Computed:    true,
		},
		"encoded": {
			Description: "API key credentials which is the Base64-encoding of the UTF-8 representation of the id and api_key joined by a colon (:).",
			Type:        schema.TypeString,
			Sensitive:   true,
			Computed:    true,
		},
	}

	utils.AddConnectionSchema(apikeySchema)

	return &schema.Resource{
		Description: "Creates an API key for access without requiring basic authentication. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-create-api-key.html",

		CreateContext: resourceSecurityApiKeyCreate,
		UpdateContext: resourceSecurityApiKeyUpdate,
		ReadContext:   resourceSecurityApiKeyRead,
		DeleteContext: resourceSecurityApiKeyDelete,

		Schema: apikeySchema,
	}
}

func resourceSecurityApiKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}

	nameId := d.Get("name").(string)

	var apikey models.ApiKey
	apikey.Name = nameId

	if v, ok := d.GetOk("expiration"); ok {
		apikey.Expiration = v.(string)
	}

	if v, ok := d.GetOk("role_descriptors"); ok {
		role_descriptors := map[string]models.ApiKeyRoleDescriptor{}
		if err := json.NewDecoder(strings.NewReader(v.(string))).Decode(&role_descriptors); err != nil {
			return diag.FromErr(err)
		}
		apikey.RolesDescriptors = role_descriptors

		var hasRestriction = false
		var keysWithRestrictions []string

		for key, roleDescriptor := range role_descriptors {
			if roleDescriptor.Restriction != nil {
				hasRestriction = true
				keysWithRestrictions = append(keysWithRestrictions, key)
			}
		}

		if hasRestriction {
			isSupported, diags := doesCurrentVersionSupportRestrictionOnApiKey(ctx, client)

			if diags.HasError() {
				return diags
			}

			if !isSupported {
				return diag.Errorf("Specifying `restriction` on an API key role description is not supported in this version of Elasticsearch. Role descriptor(s) %s", strings.Join(keysWithRestrictions, ", "))
			}
		}
	}

	if v, ok := d.GetOk("metadata"); ok {
		metadata := make(map[string]interface{})
		if err := json.NewDecoder(strings.NewReader(v.(string))).Decode(&metadata); err != nil {
			return diag.FromErr(err)
		}
		apikey.Metadata = metadata
	}

	putResponse, diags := elasticsearch.PutApiKey(client, &apikey)

	if diags.HasError() {
		return diags
	}

	id, diags := client.ID(ctx, putResponse.Id)
	if diags.HasError() {
		return diags
	}

	if putResponse.Key != "" {
		if err := d.Set("api_key", putResponse.Key); err != nil {
			return diag.FromErr(err)
		}
	}
	if putResponse.EncodedKey != "" {
		if err := d.Set("encoded", putResponse.EncodedKey); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("key_id", putResponse.Id); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("expiration_timestamp", putResponse.Expiration); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("expiration", apikey.Expiration); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(id.String())
	return resourceSecurityApiKeyRead(ctx, d, meta)
}

func doesCurrentVersionSupportRestrictionOnApiKey(ctx context.Context, client *clients.ApiClient) (bool, diag.Diagnostics) {
	currentVersion, diags := client.ServerVersion(ctx)

	if diags.HasError() {
		return false, diags
	}

	return currentVersion.GreaterThanOrEqual(APIKeyWithRestrictionMinVersion), nil
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
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}
	compId, diags := clients.CompositeIdFromStr(d.Id())
	if diags.HasError() {
		return diags
	}
	id := compId.ResourceId

	apikey, diags := elasticsearch.GetApiKey(client, id)
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
	if err := d.Set("expiration_timestamp", apikey.Expiration); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("key_id", apikey.Id); err != nil {
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
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}
	compId, diags := clients.CompositeIdFromStr(d.Id())
	if diags.HasError() {
		return diags
	}

	if diags := elasticsearch.DeleteApiKey(client, compId.ResourceId); diags.HasError() {
		return diags
	}

	d.SetId("")
	return diags
}
