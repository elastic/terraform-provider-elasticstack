package fleet

import (
	"context"

	fleetapi "github.com/elastic/terraform-provider-elasticstack/generated/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceFleetServerHost() *schema.Resource {
	fleetServerHostSchema := map[string]*schema.Schema{
		"host_id": {
			Description: "Unique identifier of the Fleet server host.",
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
		},
		"name": {
			Description: "The name of the Fleet server host.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"hosts": {
			Description: "A list of hosts.",
			Type:        schema.TypeList,
			Required:    true,
			MinItems:    1,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"default": {
			Description: "Set as default.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
	}

	return &schema.Resource{
		Description: "Creates a new Fleet Server Host.",

		CreateContext: resourceFleetServerHostCreate,
		ReadContext:   resourceFleetServerHostRead,
		UpdateContext: resourceFleetServerHostUpdate,
		DeleteContext: resourceFleetServerHostDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: fleetServerHostSchema,
	}
}

func resourceFleetServerHostCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fleetClient, diags := getFleetClient(d, meta)
	if diags.HasError() {
		return diags
	}

	if id := d.Get("host_id").(string); id != "" {
		d.SetId(id)
	}

	req := fleetapi.PostFleetServerHostsJSONRequestBody{
		Name: d.Get("name").(string),
	}

	if value := d.Get("hosts").([]interface{}); len(value) > 0 {
		for _, v := range value {
			if vStr, ok := v.(string); ok && vStr != "" {
				req.HostUrls = append(req.HostUrls, vStr)
			}
		}
	}
	if value := d.Get("default").(bool); value {
		req.IsDefault = &value
	}

	host, diags := fleet.CreateFleetServerHost(ctx, fleetClient, req)
	if diags.HasError() {
		return diags
	}

	d.SetId(host.Id)
	if err := d.Set("host_id", host.Id); err != nil {
		return diag.FromErr(err)
	}

	return resourceFleetServerHostRead(ctx, d, meta)
}

func resourceFleetServerHostUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fleetClient, diags := getFleetClient(d, meta)
	if diags.HasError() {
		return diags
	}

	req := fleetapi.UpdateFleetServerHostsJSONRequestBody{}

	if value, ok := d.Get("name").(string); ok && value != "" {
		req.Name = &value
	}
	var hosts []string
	if value, ok := d.Get("hosts").([]interface{}); ok && len(value) > 0 {
		for _, v := range value {
			if vStr, ok := v.(string); ok && vStr != "" {
				hosts = append(hosts, vStr)
			}
		}
	}
	if hosts != nil {
		req.HostUrls = &hosts
	}
	if value := d.Get("default").(bool); value {
		req.IsDefault = &value
	}

	_, diags = fleet.UpdateFleetServerHost(ctx, fleetClient, d.Id(), req)
	if diags.HasError() {
		return diags
	}

	return resourceFleetServerHostRead(ctx, d, meta)
}

func resourceFleetServerHostRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fleetClient, diags := getFleetClient(d, meta)
	if diags.HasError() {
		return diags
	}

	host, diags := fleet.ReadFleetServerHost(ctx, fleetClient, d.Id())
	if diags.HasError() {
		return diags
	}

	// Not found.
	if host == nil {
		d.SetId("")
		return nil
	}

	if host.Name != nil {
		if err := d.Set("name", *host.Name); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("hosts", host.HostUrls); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("default", host.IsDefault); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFleetServerHostDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fleetClient, diags := getFleetClient(d, meta)
	if diags.HasError() {
		return diags
	}

	if diags = fleet.DeleteFleetServerHost(ctx, fleetClient, d.Id()); diags.HasError() {
		return diags
	}
	d.SetId("")

	return diags
}
