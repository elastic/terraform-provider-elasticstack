package ingest

import (
	"context"
	"encoding/json"
	"strings"

	_ "embed"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

//go:embed processor_network_direction_data_source.md
var networkDirectionDataSourceDescription string

func DataSourceProcessorNetworkDirection() *schema.Resource {
	processorSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"source_ip": {
			Description: "Field containing the source IP address.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"destination_ip": {
			Description: "Field containing the destination IP address.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"target_field": {
			Description: "Output field for the network direction.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"internal_networks": {
			Description: "List of internal networks.",
			Type:        schema.TypeSet,
			MinItems:    1,
			Optional:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			ConflictsWith: []string{"internal_networks_field"},
			ExactlyOneOf:  []string{"internal_networks", "internal_networks_field"},
		},
		"internal_networks_field": {
			Description:   "A field on the given document to read the internal_networks configuration from.",
			Type:          schema.TypeString,
			Optional:      true,
			ConflictsWith: []string{"internal_networks"},
			ExactlyOneOf:  []string{"internal_networks", "internal_networks_field"},
		},
		"ignore_missing": {
			Description: "If `true` and `field` does not exist or is `null`, the processor quietly exits without modifying the document.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
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
		Description: networkDirectionDataSourceDescription,
		ReadContext: dataSourceProcessorNetworkDirectionRead,

		Schema: processorSchema,
	}
}

func dataSourceProcessorNetworkDirectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	processor := &models.ProcessorNetworkDirection{}

	processor.IgnoreFailure = d.Get("ignore_failure").(bool)
	processor.IgnoreMissing = d.Get("ignore_missing").(bool)

	if v, ok := d.GetOk("source_ip"); ok {
		processor.SourceIp = v.(string)
	}
	if v, ok := d.GetOk("destination_ip"); ok {
		processor.DestinationIp = v.(string)
	}
	if v, ok := d.GetOk("internal_networks"); ok {
		nets := v.(*schema.Set)
		networks := make([]string, nets.Len())
		for i, n := range nets.List() {
			networks[i] = n.(string)
		}
		processor.InternalNetworks = networks
	}
	if v, ok := d.GetOk("internal_networks_field"); ok {
		processor.InternalNetworksField = v.(string)
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

	processorJson, err := json.MarshalIndent(map[string]*models.ProcessorNetworkDirection{"network_direction": processor}, "", " ")
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
