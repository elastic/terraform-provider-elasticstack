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

func DataSourceProcessorKV() *schema.Resource {
	processorSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"field": {
			Description: "The field to be parsed. Supports template snippets.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"field_split": {
			Description: "Regex pattern to use for splitting key-value pairs.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"value_split": {
			Description: "Regex pattern to use for splitting the key from the value within a key-value pair.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"target_field": {
			Description: "The field to insert the extracted keys into. Defaults to the root of the document.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"include_keys": {
			Description: "List of keys to filter and insert into document. Defaults to including all keys",
			Type:        schema.TypeSet,
			Optional:    true,
			MinItems:    1,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"exclude_keys": {
			Description: "List of keys to exclude from document",
			Type:        schema.TypeSet,
			Optional:    true,
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
		"prefix": {
			Description: "Prefix to be added to extracted keys.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"trim_key": {
			Description: "String of characters to trim from extracted keys.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"trim_value": {
			Description: "String of characters to trim from extracted values.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"strip_brackets": {
			Description: "If `true` strip brackets `()`, `<>`, `[]` as well as quotes `'` and `\"` from extracted values.",
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
		Description: "This processor helps automatically parse messages (or specific event fields) which are of the foo=bar variety. See: https://www.elastic.co/guide/en/elasticsearch/reference/current/kv-processor.html",

		ReadContext: dataSourceProcessorKVRead,

		Schema: processorSchema,
	}
}

func dataSourceProcessorKVRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	processor := &models.ProcessorKV{}

	processor.Field = d.Get("field").(string)
	processor.FieldSplit = d.Get("field_split").(string)
	processor.ValueSplit = d.Get("value_split").(string)
	processor.IgnoreFailure = d.Get("ignore_failure").(bool)
	processor.IgnoreMissing = d.Get("ignore_missing").(bool)
	processor.StripBrackets = d.Get("strip_brackets").(bool)

	if v, ok := d.GetOk("include_keys"); ok {
		kk := v.(*schema.Set)
		keys := make([]string, kk.Len())
		for i, k := range kk.List() {
			keys[i] = k.(string)
		}
		processor.IncludeKeys = keys
	}
	if v, ok := d.GetOk("exclude_keys"); ok {
		kk := v.(*schema.Set)
		keys := make([]string, kk.Len())
		for i, k := range kk.List() {
			keys[i] = k.(string)
		}
		processor.ExcludeKeys = keys
	}
	if v, ok := d.GetOk("target_field"); ok {
		processor.TargetField = v.(string)
	}
	if v, ok := d.GetOk("prefix"); ok {
		processor.Prefix = v.(string)
	}
	if v, ok := d.GetOk("trim_key"); ok {
		processor.TrimKey = v.(string)
	}
	if v, ok := d.GetOk("trim_value"); ok {
		processor.TrimValue = v.(string)
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

	processorJson, err := json.MarshalIndent(map[string]*models.ProcessorKV{"kv": processor}, "", " ")
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
