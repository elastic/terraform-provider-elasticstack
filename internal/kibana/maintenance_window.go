package kibana

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceMaintenanceWindow() *schema.Resource {
	apikeySchema := map[string]*schema.Schema{
		"id": {
			Description: "A UUID v1 or v4 to use instead of a randomly generated ID.",
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
			ForceNew:    true,
		},
		"title": {
			Description: "The name of the maintenance window.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"enabled": {
			Description: "Whether the current maintenance window is enabled.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"start": {
			Description: "The start date of the maintenance window.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"duration": {
			Description: "How long the maintenance window should run from the start date.",
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     "default",
			ForceNew:    true,
		},
	}

	return &schema.Resource{
		Description: "Creates a Kibana Maintenance Window.",

		CreateContext: resourceMaintenanceWindowCreate,
		UpdateContext: resourceMaintenanceWindowUpdate,
		ReadContext:   resourceMaintenanceWindowRead,
		DeleteContext: resourceMaintenanceWindowDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: apikeySchema,
	}
}

func getMaintenanceWindowFromResourceData(d *schema.ResourceData, serverVersion *version.Version) (models.MaintenanceWindow, diag.Diagnostics) {
	var diags diag.Diagnostics
	maintenanceWindow := models.MaintenanceWindow{
		Title:    d.Get("title").(string),
		Enabled:  d.Get("enabled").(bool),
		Start:    d.Get("start").(string),
		Duration: d.Get("duration").(int),
	}

	// Explicitly set maintenance window id if provided, otherwise we'll use the autogenerated ID from the Kibana API response
	if maintenanceWindowID := getOrNilString("id", d); maintenanceWindowID != nil && *maintenanceWindowID != "" {
		maintenanceWindow.Id = *maintenanceWindowID
	}

	return maintenanceWindow, diags
}

func resourceMaintenanceWindowCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}

	serverVersion, diags := client.ServerVersion(ctx)
	if diags.HasError() {
		return diags
	}

	maintenanceWindow, diags := getMaintenanceWindowFromResourceData(d, serverVersion)
	if diags.HasError() {
		return diags
	}

	res, diags := kibana.CreateMaintenanceWindow(ctx, client, maintenanceWindow)

	if diags.HasError() {
		return diags
	}

	d.SetId(res.MaintenanceWindowID)

	return resourceRuleRead(ctx, d, meta)
}

func resourceMaintenanceWindowUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}

	serverVersion, diags := client.ServerVersion(ctx)
	if diags.HasError() {
		return diags
	}

	maintenanceWindow, diags := getMaintenanceWindowFromResourceData(d, serverVersion)
	if diags.HasError() {
		return diags
	}

	// DO NOTHING
	// res, diags := kibana.UpdateAlertingRule(ctx, client, maintenanceWindow)
	d.SetId(maintenanceWindow.MaintenanceWindowID)

	return resourceRuleRead(ctx, d, meta)
}

func resourceMaintenanceWindowRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}

	maintenanceWindow, diags := kibana.GetMaintenanceWindow(ctx, client, d.Id())

	if maintenanceWindow == nil && diags == nil {
		d.SetId("")
		return diags
	}
	if diags.HasError() {
		return diags
	}

	// set the fields
	if err := d.Set("id", maintenanceWindow.Id); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("title", maintenanceWindow.Title); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("enabled", maintenanceWindow.Enabled); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("start", maintenanceWindow.Start); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("duration", maintenanceWindow.Duration); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceMaintenanceWindowDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}

	if diags = kibana.DeleteMaintenanceWindow(ctx, client, d.Id()); diags.HasError() {
		return diags
	}

	d.SetId("")
	return diags
}
