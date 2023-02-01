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

func DataSourceProcessorSet() *schema.Resource {
	processorSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"field": {
			Description: "The field to insert, upsert, or update.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"value": {
			Description:   "The value to be set for the field. Supports template snippets. May specify only one of `value` or `copy_from`.",
			Type:          schema.TypeString,
			Optional:      true,
			ConflictsWith: []string{"copy_from"},
			ExactlyOneOf:  []string{"copy_from", "value"},
		},
		"copy_from": {
			Description:   "The origin field which will be copied to `field`, cannot set `value` simultaneously.",
			Type:          schema.TypeString,
			Optional:      true,
			ConflictsWith: []string{"value"},
			ExactlyOneOf:  []string{"copy_from", "value"},
		},
		"override": {
			Description: "If processor will update fields with pre-existing non-null-valued field.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},
		"ignore_empty_value": {
			Description: "If `true` and `value` is a template snippet that evaluates to `null` or the empty string, the processor quietly exits without modifying the document",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"media_type": {
			Description: "The media type for encoding value.",
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
		Description: "Sets one field and associates it with the specified value. See: https://www.elastic.co/guide/en/elasticsearch/reference/current/set-processor.html",

		ReadContext: dataSourceProcessorSetRead,

		Schema: processorSchema,
	}
}

func dataSourceProcessorSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	processor := &models.ProcessorSet{}

	processor.Field = d.Get("field").(string)
	processor.IgnoreFailure = d.Get("ignore_failure").(bool)
	processor.Override = d.Get("override").(bool)
	processor.IgnoreEmptyValue = d.Get("ignore_empty_value").(bool)

	if v, ok := d.GetOk("value"); ok {
		processor.Value = v.(string)
	}
	if v, ok := d.GetOk("copy_from"); ok {
		processor.CopyFrom = v.(string)
	}
	if v, ok := d.GetOk("media_type"); ok {
		processor.MediaType = v.(string)
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

	processorJson, err := json.MarshalIndent(map[string]*models.ProcessorSet{"set": processor}, "", " ")
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
