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

//go:embed processor_community_id_data_source.md
var communityIdDataSourceDescription string

func DataSourceProcessorCommunityId() *schema.Resource {
	processorSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"source_ip": {
			Description: "Field containing the source IP address.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"source_port": {
			Description: "Field containing the source port.",
			Type:        schema.TypeInt,
			Optional:    true,
		},
		"destination_ip": {
			Description: "Field containing the destination IP address.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"destination_port": {
			Description: "Field containing the destination port.",
			Type:        schema.TypeInt,
			Optional:    true,
		},
		"iana_number": {
			Description: "Field containing the IANA number.",
			Type:        schema.TypeInt,
			Optional:    true,
		},
		"icmp_type": {
			Description: "Field containing the ICMP type.",
			Type:        schema.TypeInt,
			Optional:    true,
		},
		"icmp_code": {
			Description: "Field containing the ICMP code.",
			Type:        schema.TypeInt,
			Optional:    true,
		},
		"seed": {
			Description:  "Seed for the community ID hash. Must be between 0 and 65535 (inclusive). The seed can prevent hash collisions between network domains, such as a staging and production network that use the same addressing scheme.",
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      0,
			ValidateFunc: validation.IntBetween(0, 65535),
		},
		"transport": {
			Description: "Field containing the transport protocol. Used only when the `iana_number` field is not present.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"target_field": {
			Description: "Output field for the community ID.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"ignore_missing": {
			Description: "If `true` and `field` does not exist or is `null`, the processor quietly exits without modifying the document.",
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
		Description: communityIdDataSourceDescription,

		ReadContext: dataSourceProcessorCommunityIdRead,

		Schema: processorSchema,
	}
}

func dataSourceProcessorCommunityIdRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	processor := &models.ProcessorCommunityId{}

	processor.IgnoreFailure = d.Get("ignore_failure").(bool)
	processor.IgnoreMissing = d.Get("ignore_missing").(bool)
	seed := d.Get("seed").(int)
	processor.Seed = &seed

	if v, ok := d.GetOk("source_ip"); ok {
		processor.SourceIp = v.(string)
	}
	if v, ok := d.GetOk("source_port"); ok {
		port := v.(int)
		processor.SourcePort = &port
	}
	if v, ok := d.GetOk("destination_ip"); ok {
		processor.DestinationIp = v.(string)
	}
	if v, ok := d.GetOk("destination_port"); ok {
		port := v.(int)
		processor.DestinationPort = &port
	}
	if v, ok := d.GetOk("iana_number"); ok {
		processor.IanaNumber = v.(string)
	}
	if v, ok := d.GetOk("icmp_type"); ok {
		num := v.(int)
		processor.IcmpType = &num
	}
	if v, ok := d.GetOk("icmp_code"); ok {
		num := v.(int)
		processor.IcmpCode = &num
	}
	if v, ok := d.GetOk("transport"); ok {
		processor.Transport = v.(string)
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

	processorJson, err := json.MarshalIndent(map[string]*models.ProcessorCommunityId{"community_id": processor}, "", " ")
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
