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

func DataSourceProcessorDateIndexName() *schema.Resource {
	processorSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"field": {
			Description: "The field to get the date or timestamp from.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"index_name_prefix": {
			Description: "A prefix of the index name to be prepended before the printed date.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"date_rounding": {
			Description:  "How to round the date when formatting the date into the index name.",
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"y", "M", "w", "d", "h", "m", "s"}, false),
		},
		"date_formats": {
			Description: "An array of the expected date formats for parsing dates / timestamps in the document being preprocessed.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"timezone": {
			Description: "The timezone to use when parsing the date and when date math index supports resolves expressions into concrete index names.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "UTC",
		},
		"locale": {
			Description: "The locale to use when parsing the date from the document being preprocessed, relevant when parsing month names or week days.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "ENGLISH",
		},
		"index_name_format": {
			Description: "The format to be used when printing the parsed date into the index name.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "yyyy-MM-dd",
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
		Description: "The purpose of this processor is to point documents to the right time based index based on a date or timestamp field in a document by using the date math index name support. See: https://www.elastic.co/guide/en/elasticsearch/reference/current/date-index-name-processor.html",

		ReadContext: dataSourceProcessorDateIndexNameRead,

		Schema: processorSchema,
	}
}

func dataSourceProcessorDateIndexNameRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	processor := &models.ProcessorDateIndexName{}

	processor.Field = d.Get("field").(string)
	processor.IgnoreFailure = d.Get("ignore_failure").(bool)
	processor.Timezone = d.Get("timezone").(string)
	processor.Locale = d.Get("locale").(string)
	processor.IndexNameFormat = d.Get("index_name_format").(string)
	processor.DateRounding = d.Get("date_rounding").(string)

	if v, ok := d.GetOk("date_formats"); ok {
		formats := v.([]interface{})
		res := make([]string, len(formats))
		for i, v := range formats {
			res[i] = v.(string)
		}
		processor.DateFormats = res
	}

	if v, ok := d.GetOk("index_name_prefix"); ok {
		processor.IndexNamePrefix = v.(string)
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

	processorJson, err := json.MarshalIndent(map[string]*models.ProcessorDateIndexName{"date_index_name": processor}, "", " ")
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
