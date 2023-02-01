package ingest

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceProcessorGeoip() *schema.Resource {
	processorSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"field": {
			Description: "The field to get the ip address from for the geographical lookup.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"target_field": {
			Description: "The field that will hold the geographical information looked up from the MaxMind database.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "geoip",
		},
		"database_file": {
			Description: "The database filename referring to a database the module ships with (GeoLite2-City.mmdb, GeoLite2-Country.mmdb, or GeoLite2-ASN.mmdb) or a custom database in the `ingest-geoip` config directory.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"properties": {
			Description: "Controls what properties are added to the `target_field` based on the geoip lookup.",
			Type:        schema.TypeSet,
			Optional:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"ignore_missing": {
			Description: "If `true` and `field` does not exist, the processor quietly exits without modifying the document.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"first_only": {
			Description: "If `true` only first found geoip data will be returned, even if field contains array.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},
		"json": {
			Description: "JSON representation of this data source.",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}

	return &schema.Resource{
		Description: "The geoip processor adds information about the geographical location of an IPv4 or IPv6 address. See: https://www.elastic.co/guide/en/elasticsearch/reference/current/geoip-processor.html",

		ReadContext: dataSourceProcessorGeoipRead,

		Schema: processorSchema,
	}
}

func dataSourceProcessorGeoipRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	processor := &models.ProcessorGeoip{}

	processor.IgnoreMissing = d.Get("ignore_missing").(bool)
	processor.FirstOnly = d.Get("first_only").(bool)
	processor.Field = d.Get("field").(string)
	processor.TargetField = d.Get("target_field").(string)

	if v, ok := d.GetOk("properties"); ok {
		props := v.(*schema.Set)
		properties := make([]string, props.Len())
		for i, p := range props.List() {
			properties[i] = p.(string)
		}
		processor.Properties = properties
	}

	if v, ok := d.GetOk("database_file"); ok {
		processor.DatabaseFile = v.(string)
	}

	processorJson, err := json.MarshalIndent(map[string]*models.ProcessorGeoip{"geoip": processor}, "", " ")
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
