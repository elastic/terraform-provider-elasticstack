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

func DataSourceProcessorCSV() *schema.Resource {
	processorSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"field": {
			Description: "The field to extract data from.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"target_fields": {
			Description: "The array of fields to assign extracted values to.",
			Type:        schema.TypeList,
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
		"separator": {
			Description: "Separator used in CSV, has to be single character string.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     ",",
		},
		"quote": {
			Description: "Quote used in CSV, has to be single character string",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     `"`,
		},
		"trim": {
			Description: "Trim whitespaces in unquoted fields.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"empty_value": {
			Description: "Value used to fill empty fields, empty fields will be skipped if this is not provided.",
			Type:        schema.TypeString,
			Optional:    true,
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
		Description: "Extracts fields from CSV line out of a single text field within a document. Any empty field in CSV will be skipped. See: https://www.elastic.co/guide/en/elasticsearch/reference/current/csv-processor.html",

		ReadContext: dataSourceProcessorCSVRead,

		Schema: processorSchema,
	}
}

func dataSourceProcessorCSVRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	processor := &models.ProcessorCSV{}

	processor.Field = d.Get("field").(string)
	processor.IgnoreFailure = d.Get("ignore_failure").(bool)
	processor.IgnoreMissing = d.Get("ignore_missing").(bool)
	processor.Separator = d.Get("separator").(string)
	processor.Quote = d.Get("quote").(string)
	processor.Trim = d.Get("trim").(bool)

	tFields := d.Get("target_fields").([]interface{})
	targets := make([]string, len(tFields))
	for i, v := range tFields {
		targets[i] = v.(string)
	}
	processor.TargetFields = targets

	if v, ok := d.GetOk("empty_value"); ok {
		processor.EmptyValue = v.(string)
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

	processorJson, err := json.MarshalIndent(map[string]*models.ProcessorCSV{"csv": processor}, "", " ")
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
