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

func DataSourceProcessorAppend() *schema.Resource {
	processorSchema := map[string]*schema.Schema{
		"field": {
			Description: "The field to be appended to.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"value": {
			Description: "The value to be appended. ",
			Type:        schema.TypeList,
			Required:    true,
			MinItems:    1,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"allow_duplicates": {
			Description: "If `false`, the processor does not append values already present in the field.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},
		"media_type": {
			Description: "The media type for encoding value. Applies only when value is a template snippet. Must be one of `application/json`, `text/plain`, or `application/x-www-form-urlencoded`.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "application/json",
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
		Description: "Appends one or more values to an existing array if the field already exists and it is an array. Converts a scalar to an array and appends one or more values to it if the field exists and it is a scalar. Creates an array containing the provided values if the field doesn’t exist.",

		ReadContext: dataSourceProcessorAppendRead,

		Schema: processorSchema,
	}
}

func dataSourceProcessorAppendRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	processor := &models.ProcessorAppend{}

	processor.Field = d.Get("field").(string)
	values := make([]string, 0)
	for _, v := range d.Get("value").([]interface{}) {
		values = append(values, v.(string))
	}
	processor.Value = values
	processor.AllowDuplicates = d.Get("allow_duplicates").(bool)
	processor.IgnoreFailure = d.Get("ignore_failure").(bool)
	processor.MediaType = d.Get("media_type").(string)
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

	processorJson, err := json.MarshalIndent(map[string]*models.ProcessorAppend{"append": processor}, "", " ")
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
