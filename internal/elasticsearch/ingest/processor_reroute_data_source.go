package ingest

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func DataSourceProcessorReroute() *schema.Resource {
	processorSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"destination": {
			Description: "The destination for the rerouted documents. It can be an index, an alias, or a data stream. It cannot be used with `dataset` or `namespace`.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"dataset": {
			Description: "The dataset for the rerouted documents. It can be a static value or a field reference. It cannot be used with `destination`.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"namespace": {
			Description: "The namespace for the rerouted documents. It can be a static value or a field reference. It cannot be used with `destination`.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"description": {
			Description: "Description of the processor.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"if": {
			Description: "Conditionally execute the processor.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"ignore_failure": {
			Description: "Ignore failures for the processor.",
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
		Description: "The reroute processor allows to route a document to another target index or data stream. See: https://www.elastic.co/guide/en/elasticsearch/reference/current/reroute-processor.html",

		ReadContext: dataSourceProcessorRerouteRead,

		Schema: processorSchema,
	}
}

func dataSourceProcessorRerouteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	processor := &models.ProcessorReroute{}

	if v, ok := d.GetOk("destination"); ok {
		processor.Destination = v.(string)
	}
	if v, ok := d.GetOk("dataset"); ok {
		dataset := v.([]interface{})
		processor.Dataset = make([]string, len(dataset))
		for i, ds := range dataset {
			processor.Dataset[i] = ds.(string)
		}
	}
	if v, ok := d.GetOk("namespace"); ok {
		namespace := v.([]interface{})
		processor.Namespace = make([]string, len(namespace))
		for i, ns := range namespace {
			processor.Namespace[i] = ns.(string)
		}
	}
	if v, ok := d.GetOk("description"); ok {
		processor.Description = v.(string)
	}
	if v, ok := d.GetOk("if"); ok {
		processor.If = v.(string)
	}
	processor.IgnoreFailure = d.Get("ignore_failure").(bool)
	if v, ok := d.GetOk("on_failure"); ok {
		onFailure := v.([]interface{})
		processor.OnFailure = make([]map[string]interface{}, len(onFailure))
		for i, f := range onFailure {
			var failure map[string]interface{}
			if err := json.Unmarshal([]byte(f.(string)), &failure); err != nil {
				return diag.FromErr(err)
			}
			processor.OnFailure[i] = failure
		}
	}
	if v, ok := d.GetOk("tag"); ok {
		processor.Tag = v.(string)
	}

	processorJson, err := json.MarshalIndent(map[string]*models.ProcessorReroute{"reroute": processor}, "", "  ")
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
