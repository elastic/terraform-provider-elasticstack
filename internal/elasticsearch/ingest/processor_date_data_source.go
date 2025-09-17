package ingest

import (
	"context"
	_ "embed"
	"encoding/json"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

//go:embed processor_date_data_source.md
var dateDataSourceDescription string

func DataSourceProcessorDate() *schema.Resource {
	processorSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"field": {
			Description: "The field to get the date from.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"target_field": {
			Description: "The field that will hold the parsed date.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "@timestamp",
		},
		"formats": {
			Description: "An array of the expected date formats.",
			Type:        schema.TypeList,
			Required:    true,
			MinItems:    1,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"timezone": {
			Description: "The timezone to use when parsing the date.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "UTC",
		},
		"locale": {
			Description: "The locale to use when parsing the date, relevant when parsing month names or week days.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "ENGLISH",
		},
		"output_format": {
			Description: "The format to use when writing the date to `target_field`.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "yyyy-MM-dd'T'HH:mm:ss.SSSXXX",
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
		Description: dateDataSourceDescription,
		ReadContext: dataSourceProcessorDateRead,

		Schema: processorSchema,
	}
}

func dataSourceProcessorDateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	processor := &models.ProcessorDate{}

	processor.Field = d.Get("field").(string)
	processor.IgnoreFailure = d.Get("ignore_failure").(bool)
	processor.Timezone = d.Get("timezone").(string)
	processor.Locale = d.Get("locale").(string)
	processor.OutputFormat = d.Get("output_format").(string)

	formats := d.Get("formats").([]interface{})
	res := make([]string, len(formats))
	for i, v := range formats {
		res[i] = v.(string)
	}
	processor.Formats = res

	if v, ok := d.GetOk("target_field"); ok {
		processor.TargetField = v.(string)
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

	processorJson, err := json.MarshalIndent(map[string]*models.ProcessorDate{"date": processor}, "", " ")
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
