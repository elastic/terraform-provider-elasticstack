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

func DataSourceProcessorGrok() *schema.Resource {
	processorSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"field": {
			Description: "The field to use for grok expression parsing",
			Type:        schema.TypeString,
			Required:    true,
		},
		"patterns": {
			Description: "An ordered list of grok expression to match and extract named captures with. Returns on the first expression in the list that matches.",
			Type:        schema.TypeList,
			Required:    true,
			MinItems:    1,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"pattern_definitions": {
			Description: "A map of pattern-name and pattern tuples defining custom patterns to be used by the current processor. Patterns matching existing names will override the pre-existing definition.",
			Type:        schema.TypeMap,
			Optional:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"ecs_compatibility": {
			Description:  "Must be disabled or v1. If v1, the processor uses patterns with Elastic Common Schema (ECS) field names. **NOTE:** Supported only starting from version of Elasticsearch **7.16.x**.",
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringInSlice([]string{"disabled", "v1"}, false),
		},
		"trace_match": {
			Description: "when true, `_ingest._grok_match_index` will be inserted into your matched document’s metadata with the index into the pattern found in `patterns` that matched.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"ignore_missing": {
			Description: "If `true` and `field` does not exist or is `null`, the processor quietly exits without modifying the document",
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
		Description: "Extracts structured fields out of a single text field within a document. See: https://www.elastic.co/guide/en/elasticsearch/reference/current/grok-processor.html",

		ReadContext: dataSourceProcessorGrokRead,

		Schema: processorSchema,
	}
}

func dataSourceProcessorGrokRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	processor := &models.ProcessorGrok{}

	processor.Field = d.Get("field").(string)
	processor.TraceMatch = d.Get("trace_match").(bool)
	processor.IgnoreMissing = d.Get("ignore_missing").(bool)
	processor.IgnoreFailure = d.Get("ignore_failure").(bool)

	pats := d.Get("patterns").([]interface{})
	patterns := make([]string, len(pats))
	for i, v := range pats {
		patterns[i] = v.(string)
	}
	processor.Patterns = patterns

	if v, ok := d.GetOk("ecs_compatibility"); ok {
		processor.EcsCompatibility = v.(string)
	}
	if v, ok := d.GetOk("pattern_definitions"); ok {
		pd := v.(map[string]interface{})
		defs := make(map[string]string)
		for k, p := range pd {
			defs[k] = p.(string)
		}
		processor.PatternDefinitions = defs
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

	processorJson, err := json.MarshalIndent(map[string]*models.ProcessorGrok{"grok": processor}, "", " ")
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
