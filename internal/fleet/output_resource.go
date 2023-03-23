package fleet

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet/fleetapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceOutput() *schema.Resource {
	outputSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"output_id": {
			Description: "Unique identifier of the output.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"name": {
			Description: "The name of the output.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"type": {
			Description:  "The output type.",
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"elasticsearch", "logstash"}, false),
		},
		"hosts": {
			Description: "A list of hosts.",
			Type:        schema.TypeList,
			Optional:    true,
			MinItems:    1,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"ca_sha256": {
			Description: "Fingerprint of the Elasticsearch CA certificate.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"default_integrations": {
			Description: "Make this output the default for agent integrations.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"default_monitoring": {
			Description: "Make this output the default for agent monitoring.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"config_yaml": {
			Description: "Advanced YAML configuration. YAML settings here will be added to the output section of each agent policy.",
			Type:        schema.TypeString,
			Optional:    true,
			Sensitive:   true,
		},
	}

	return &schema.Resource{
		Description: "Creates a new Fleet Output.",

		CreateContext: resourceOutputCreate,
		ReadContext:   resourceOutputRead,
		UpdateContext: resourceOutputUpdate,
		DeleteContext: resourceOutputDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: outputSchema,
	}
}

func resourceOutputCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fleetClient, diags := getFleetClient(d, meta)
	if diags.HasError() {
		return diags
	}

	req := fleetapi.PostOutputsJSONRequestBody{
		Name: d.Get("name").(string),
		Type: fleetapi.PostOutputsJSONBodyType(d.Get("type").(string)),
	}

	var hosts []string
	if value := d.Get("hosts").([]interface{}); len(value) > 0 {
		for _, v := range value {
			if vStr, ok := v.(string); ok && vStr != "" {
				hosts = append(hosts, vStr)
			}
		}
	}
	if hosts != nil {
		req.Hosts = &hosts
	}
	if value := d.Get("default_integrations").(bool); value {
		req.IsDefault = &value
	}
	if value := d.Get("default_monitoring").(bool); value {
		req.IsDefaultMonitoring = &value
	}
	if value, ok := d.Get("ca_sha256").(string); ok && value != "" {
		req.CaSha256 = &value
	}
	if value, ok := d.Get("config_yaml").(string); ok && value != "" {
		req.ConfigYaml = &value
	}

	host, diags := fleet.CreateOutput(ctx, fleetClient, req)
	if diags.HasError() {
		return diags
	}

	d.SetId(host.Id)
	if err := d.Set("output_id", host.Id); err != nil {
		return diag.FromErr(err)
	}

	return resourceOutputRead(ctx, d, meta)
}

func resourceOutputUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fleetClient, diags := getFleetClient(d, meta)
	if diags.HasError() {
		return diags
	}

	id := d.Get("output_id").(string)
	d.SetId(id)

	req := fleetapi.UpdateOutputJSONRequestBody{
		Name: d.Get("name").(string),
	}

	var hosts []string
	if value := d.Get("hosts").([]interface{}); len(value) > 0 {
		for _, v := range value {
			if vStr, ok := v.(string); ok && vStr != "" {
				hosts = append(hosts, vStr)
			}
		}
	}
	if hosts != nil {
		req.Hosts = &hosts
	}
	if value := d.Get("default_integrations").(bool); value {
		req.IsDefault = &value
	}
	if value := d.Get("default_monitoring").(bool); value {
		req.IsDefaultMonitoring = &value
	}
	if value, ok := d.Get("ca_sha256").(string); ok && value != "" {
		req.CaSha256 = &value
	}
	if value, ok := d.Get("config_yaml").(string); ok && value != "" {
		req.ConfigYaml = &value
	}
	if value, ok := d.Get("type").(string); ok && value != "" {
		req.Type = fleetapi.UpdateOutputJSONBodyType(value)
	}

	_, diags = fleet.UpdateOutput(ctx, fleetClient, id, req)
	if diags.HasError() {
		return diags
	}

	return resourceOutputRead(ctx, d, meta)
}

func resourceOutputRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fleetClient, diags := getFleetClient(d, meta)
	if diags.HasError() {
		return diags
	}

	id := d.Get("output_id").(string)
	d.SetId(id)

	output, diags := fleet.ReadOutput(ctx, fleetClient, id)
	if diags.HasError() {
		return diags
	}

	// Not found.
	if output == nil {
		d.SetId("")
		return nil
	}

	if err := d.Set("name", output.Name); err != nil {
		return diag.FromErr(err)
	}
	if output.Hosts != nil {
		if err := d.Set("hosts", *output.Hosts); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("default_integrations", output.IsDefault); err != nil {
		return diag.FromErr(err)
	}
	if output.IsDefaultMonitoring != nil {
		if err := d.Set("default_monitoring", *output.IsDefaultMonitoring); err != nil {
			return diag.FromErr(err)
		}
	}
	if output.CaSha256 != nil {
		if err := d.Set("ca_sha256", *output.CaSha256); err != nil {
			return diag.FromErr(err)
		}
	}
	if output.ConfigYaml != nil {
		if err := d.Set("config_yaml", *output.ConfigYaml); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceOutputDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fleetClient, diags := getFleetClient(d, meta)
	if diags.HasError() {
		return diags
	}

	id := d.Get("output_id").(string)
	d.SetId(id)

	if diags = fleet.DeleteOutput(ctx, fleetClient, id); diags.HasError() {
		return diags
	}
	d.SetId("")

	return diags
}
