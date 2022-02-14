package ingest

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func DataSourceProcessorSetSecurityUser() *schema.Resource {
	processorSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"field": {
			Description: "The field to store the user information into.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"properties": {
			Description: "Controls what user related properties are added to the `field`.",
			Type:        schema.TypeSet,
			Optional:    true,
			MinItems:    1,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"description": {
			Description: "Description of the processor. ",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"if": {
			Description: "Conditionally execute the processor",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"ignore_failure": {
			Description: "Ignore failures for the processor. ",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"on_failure": {
			Description: "Handle failures for the processor.",
			Type:        schema.TypeList,
			Optional:    true,
			MinItems:    1,
			Elem: &schema.Schema{
				Type:             schema.TypeString,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: utils.DiffJsonSuppress,
			},
		},
		"tag": {
			Description: "Identifier for the processor.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"json": {
			Description: "JSON representation of this data source.",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}

	return &schema.Resource{
		Description: "Sets user-related details (such as username, roles, email, full_name, metadata, api_key, realm and authentication_type) from the current authenticated user to the current document by pre-processing the ingest. See: https://www.elastic.co/guide/en/elasticsearch/reference/current/ingest-node-set-security-user-processor.html",

		ReadContext: dataSourceProcessorSetSecurityUserRead,

		Schema: processorSchema,
	}
}

func dataSourceProcessorSetSecurityUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	processor := &models.ProcessorSetSecurityUser{}

	processor.Field = d.Get("field").(string)
	processor.IgnoreFailure = d.Get("ignore_failure").(bool)

	if v, ok := d.GetOk("properties"); ok {
		props := v.(*schema.Set)
		properties := make([]string, props.Len())
		for i, p := range props.List() {
			properties[i] = p.(string)
		}
		processor.Properties = properties
	}
	if v, ok := d.GetOk("description"); ok {
		processor.Description = v.(string)
	}
	if v, ok := d.GetOk("if"); ok {
		processor.If = v.(string)
	}
	if v, ok := d.GetOk("tag"); ok {
		processor.Tag = v.(string)
	}
	if v, ok := d.GetOk("on_failure"); ok {
		onFailure := make([]map[string]interface{}, len(v.([]interface{})))
		for i, f := range v.([]interface{}) {
			item := make(map[string]interface{})
			if err := json.NewDecoder(strings.NewReader(f.(string))).Decode(&item); err != nil {
				return diag.FromErr(err)
			}
			onFailure[i] = item
		}
		processor.OnFailure = onFailure
	}

	processorJson, err := json.MarshalIndent(map[string]*models.ProcessorSetSecurityUser{"set_security_user": processor}, "", " ")
	if err != nil {
		diag.FromErr(err)
	}
	if err := d.Set("json", string(processorJson)); err != nil {
		diag.FromErr(err)
	}

	hash, err := utils.StringToHash(string(processorJson))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(*hash)

	return diags
}
