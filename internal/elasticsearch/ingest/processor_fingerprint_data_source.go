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

func DataSourceProcessorFingerprint() *schema.Resource {
	processorSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"fields": {
			Description: "Array of fields to include in the fingerprint.",
			Type:        schema.TypeList,
			Required:    true,
			MinItems:    1,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"target_field": {
			Description: "Output field for the fingerprint.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "fingerprint",
		},
		"salt": {
			Description: "Salt value for the hash function.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"method": {
			Description:  "The hash method used to compute the fingerprint.",
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "SHA-1",
			ValidateFunc: validation.StringInSlice([]string{"MD5", "SHA-1", "SHA-256", "SHA-512", "MurmurHash3"}, false),
		},
		"ignore_missing": {
			Description: "If `true`, the processor ignores any missing `fields`. If all fields are missing, the processor silently exits without modifying the document.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
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
		Description: "Helper data source which can be used to create the configuration for a fingerprint processor. This processor computes a hash of the document’s content. See: https://www.elastic.co/guide/en/elasticsearch/reference/current/fingerprint-processor.html",

		ReadContext: dataSourceProcessorFingerprintRead,

		Schema: processorSchema,
	}
}

func dataSourceProcessorFingerprintRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	processor := &models.ProcessorFingerprint{}

	processor.IgnoreFailure = d.Get("ignore_failure").(bool)
	processor.IgnoreMissing = d.Get("ignore_missing").(bool)
	processor.Method = d.Get("method").(string)
	processor.TargetField = d.Get("target_field").(string)

	fields := d.Get("fields").([]interface{})
	flds := make([]string, len(fields))
	for i, v := range fields {
		flds[i] = v.(string)
	}
	processor.Fields = flds

	if v, ok := d.GetOk("salt"); ok {
		processor.Salt = v.(string)
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

	processorJson, err := json.MarshalIndent(map[string]*models.ProcessorFingerprint{"fingerprint": processor}, "", " ")
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("json", string(processorJson)); err != nil {
		return diag.FromErr(err)
	}

	hash, err := utils.StringToHash(string(processorJson))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(*hash)

	return diags
}
