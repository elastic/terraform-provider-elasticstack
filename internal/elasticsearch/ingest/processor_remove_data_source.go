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

func DataSourceProcessorRemove() *schema.Resource {
	processorSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"field": {
			Description: "Fields to be removed.",
			Type:        schema.TypeSet,
			Required:    true,
			MinItems:    1,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"ignore_missing": {
			Description: "If `true` and `field` does not exist or is `null`, the processor quietly exits without modifying the document.",
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
		Description: "Helper data source which can be used to create the configuration for a remove processor. This processor removes existing fields. See: https://www.elastic.co/guide/en/elasticsearch/reference/current/remove-processor.html",

		ReadContext: dataSourceProcessorRemoveRead,

		Schema: processorSchema,
	}
}

func dataSourceProcessorRemoveRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	processor := &models.ProcessorRemove{}

	processor.IgnoreFailure = d.Get("ignore_failure").(bool)
	processor.IgnoreMissing = d.Get("ignore_missing").(bool)

	fields := d.Get("field").(*schema.Set)
	flds := make([]string, fields.Len())
	for i, f := range fields.List() {
		flds[i] = f.(string)
	}
	processor.Field = flds

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

	processorJson, err := json.MarshalIndent(map[string]*models.ProcessorRemove{"remove": processor}, "", " ")
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
