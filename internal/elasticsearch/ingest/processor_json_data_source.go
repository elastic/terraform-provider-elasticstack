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

func DataSourceProcessorJson() *schema.Resource {
	processorSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"field": {
			Description: "The field to be parsed.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"target_field": {
			Description: "The field that the converted structured object will be written into. Any existing content in this field will be overwritten.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"add_to_root": {
			Description: "Flag that forces the parsed JSON to be added at the top level of the document. `target_field` must not be set when this option is chosen.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"add_to_root_conflict_strategy": {
			Description:  "When set to `replace`, root fields that conflict with fields from the parsed JSON will be overridden. When set to `merge`, conflicting fields will be merged. Only applicable if `add_to_root` is set to `true`.",
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringInSlice([]string{"replace", "merge"}, false),
		},
		"allow_duplicate_keys": {
			Description: "When set to `true`, the JSON parser will not fail if the JSON contains duplicate keys. Instead, the last encountered value for any duplicate key wins.",
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
		Description: "Helper data source which can be used to create the configuration for a JSON processor. This processor converts a JSON string into a structured JSON object. See: https://www.elastic.co/guide/en/elasticsearch/reference/current/json-processor.html",

		ReadContext: dataSourceProcessorJsonRead,

		Schema: processorSchema,
	}
}

func dataSourceProcessorJsonRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	processor := &models.ProcessorJson{}

	processor.Field = d.Get("field").(string)
	processor.IgnoreFailure = d.Get("ignore_failure").(bool)

	if v, ok := d.GetOk("add_to_root_conflict_strategy"); ok {
		processor.AddToRootConflictStrategy = v.(string)
	}
	if v, ok := d.GetOk("add_to_root"); ok {
		ar := v.(bool)
		processor.AddToRoot = &ar
	}
	if v, ok := d.GetOk("allow_duplicate_keys"); ok {
		ar := v.(bool)
		processor.AllowDuplicateKeys = &ar
	}
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

	processorJson, err := json.MarshalIndent(map[string]*models.ProcessorJson{"json": processor}, "", " ")
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
