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

func DataSourceProcessorScript() *schema.Resource {
	processorSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"lang": {
			Description: "Script language.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"script_id": {
			Description:   "ID of a stored script. If no `source` is specified, this parameter is required.",
			Type:          schema.TypeString,
			Optional:      true,
			ConflictsWith: []string{"source"},
			ExactlyOneOf:  []string{"script_id", "source"},
		},
		"source": {
			Description:   "Inline script. If no id is specified, this parameter is required.",
			Type:          schema.TypeString,
			Optional:      true,
			ConflictsWith: []string{"script_id"},
			ExactlyOneOf:  []string{"script_id", "source"},
		},
		"params": {
			Description:      "Object containing parameters for the script.",
			Type:             schema.TypeString,
			Optional:         true,
			ValidateFunc:     validation.StringIsJSON,
			DiffSuppressFunc: utils.DiffJsonSuppress,
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
		Description: "Helper data source which can be used to create the configuration for a script processor. This processor runs an inline or stored script on incoming documents. See: https://www.elastic.co/guide/en/elasticsearch/reference/current/script-processor.html",

		ReadContext: dataSourceProcessorScriptRead,

		Schema: processorSchema,
	}
}

func dataSourceProcessorScriptRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	processor := &models.ProcessorScript{}

	processor.IgnoreFailure = d.Get("ignore_failure").(bool)

	if v, ok := d.GetOk("lang"); ok {
		processor.Lang = v.(string)
	}
	if v, ok := d.GetOk("script_id"); ok {
		processor.ScriptId = v.(string)
	}
	if v, ok := d.GetOk("source"); ok {
		processor.Source = v.(string)
	}
	if v, ok := d.GetOk("params"); ok {
		params := make(map[string]interface{})
		if err := json.NewDecoder(strings.NewReader(v.(string))).Decode(&params); err != nil {
			return diag.FromErr(err)
		}
		processor.Params = params
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

	processorJson, err := json.MarshalIndent(map[string]*models.ProcessorScript{"script": processor}, "", " ")
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
