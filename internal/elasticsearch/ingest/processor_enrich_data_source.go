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

func DataSourceProcessorEnrich() *schema.Resource {
	processorSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"field": {
			Description: "The field in the input document that matches the policies match_field used to retrieve the enrichment data.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"target_field": {
			Description: "Field added to incoming documents to contain enrich data.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"policy_name": {
			Description: "The name of the enrich policy to use.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"ignore_missing": {
			Description: "If `true` and `field` does not exist or is `null`, the processor quietly exits without modifying the document.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"override": {
			Description: "If processor will update fields with pre-existing non-null-valued field. ",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},
		"max_matches": {
			Description: "The maximum number of matched documents to include under the configured target field. ",
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     1,
		},
		"shape_relation": {
			Description: "A spatial relation operator used to match the geoshape of incoming documents to documents in the enrich index.",
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
		Description: "The enrich processor can enrich documents with data from another index. See: https://www.elastic.co/guide/en/elasticsearch/reference/current/enrich-processor.html",

		ReadContext: dataSourceProcessorEnrichRead,

		Schema: processorSchema,
	}
}

func dataSourceProcessorEnrichRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	processor := &models.ProcessorEnrich{}

	processor.Field = d.Get("field").(string)
	processor.TargetField = d.Get("target_field").(string)
	processor.IgnoreFailure = d.Get("ignore_failure").(bool)
	processor.IgnoreMissing = d.Get("ignore_missing").(bool)
	processor.Override = d.Get("override").(bool)
	processor.PolicyName = d.Get("policy_name").(string)
	processor.MaxMatches = d.Get("max_matches").(int)

	if v, ok := d.GetOk("shape_relation"); ok {
		processor.ShapeRelation = v.(string)
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

	processorJson, err := json.MarshalIndent(map[string]*models.ProcessorEnrich{"enrich": processor}, "", " ")
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
