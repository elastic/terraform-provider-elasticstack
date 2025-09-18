package ingest

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceProcessorUserAgent() *schema.Resource {
	processorSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"field": {
			Description: "The field containing the user agent string.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"target_field": {
			Description: "The field that will be filled with the user agent details.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"regex_file": {
			Description: "The name of the file in the `config/ingest-user-agent` directory containing the regular expressions for parsing the user agent string.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"properties": {
			Description: "Controls what properties are added to `target_field`.",
			Type:        schema.TypeSet,
			Optional:    true,
			MinItems:    1,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"extract_device_type": {
			Description: "Extracts device type from the user agent string on a best-effort basis. Supported only starting from Elasticsearch version **8.0**",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"ignore_missing": {
			Description: "If `true` and `field` does not exist or is `null`, the processor quietly exits without modifying the document.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"json": {
			Description: "JSON representation of this data source.",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}

	return &schema.Resource{
		Description: "Helper data source which can be used to create the configuration for a user agent processor. This processor extracts details from the user agent string a browser sends with its web requests. See: https://www.elastic.co/guide/en/elasticsearch/reference/current/user-agent-processor.html",

		ReadContext: dataSourceProcessorUserAgentRead,

		Schema: processorSchema,
	}
}

func dataSourceProcessorUserAgentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	processor := &models.ProcessorUserAgent{}

	processor.Field = d.Get("field").(string)
	processor.IgnoreMissing = d.Get("ignore_missing").(bool)

	if v, ok := d.GetOk("target_field"); ok {
		processor.TargetField = v.(string)
	}
	if v, ok := d.GetOk("regex_file"); ok {
		processor.RegexFile = v.(string)
	}
	if v, ok := d.GetOk("properties"); ok {
		props := v.(*schema.Set)
		properties := make([]string, props.Len())
		for i, p := range props.List() {
			properties[i] = p.(string)
		}
		processor.Properties = properties
	}
	if v, ok := d.GetOk("extract_device_type"); ok {
		dev := v.(bool)
		processor.ExtractDeviceType = &dev
	}

	processorJson, err := json.MarshalIndent(map[string]*models.ProcessorUserAgent{"user_agent": processor}, "", " ")
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
