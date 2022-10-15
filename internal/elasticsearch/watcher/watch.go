package watcher

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceWatch() *schema.Resource {
	watchSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"watch_id": {
			Description: "Identifier for the watch.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"active": {
			Description: "Defines whether the watch is active or inactive by default. The default value is true, which means the watch is active by default.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},
		"body": {
			Description:      "Configuration for the pipeline.",
			Type:             schema.TypeString,
			ValidateFunc:     validation.StringIsJSON,
			DiffSuppressFunc: utils.DiffJsonSuppress,
			Required:         true,
		},
	}

	utils.AddConnectionSchema(watchSchema)

	return &schema.Resource{
		Description: "Manage Watches. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/watcher-api.html",

		CreateContext: resourceWatchPut,
		UpdateContext: resourceWatchPut,
		ReadContext:   resourceWatchRead,
		DeleteContext: resourceWatchDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: watchSchema,
	}
}

func resourceWatchPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := clients.NewApiClient(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	watchID := d.Get("watch_id").(string)
	id, diags := client.ID(ctx, watchID)
	if diags.HasError() {
		return diags
	}

	watchBody := make(map[string]interface{})
	if err := json.NewDecoder(strings.NewReader(d.Get("body").(string))).Decode(&watchBody); err != nil {
		return diag.FromErr(err)
	}

	watch := models.Watch{
		WatchID: watchID,
		Active:  d.Get("active").(bool),
		Body:    watchBody,
	}

	if diags := client.PutWatch(ctx, &watch); diags.HasError() {
		return diags
	}

	d.SetId(id.String())
	return resourceWatchRead(ctx, d, meta)
}

func resourceWatchRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := clients.NewApiClient(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	resourceID, diags := clients.ResourceIDFromStr(d.Id())
	if diags.HasError() {
		return diags
	}

	watch, diags := client.GetWatch(ctx, resourceID)
	if watch == nil && diags == nil {
		d.SetId("")
		return diags
	}
	if diags.HasError() {
		return diags
	}

	if err := d.Set("watch_id", watch.WatchID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("active", watch.Active); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("body", watch.Body); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceWatchDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := clients.NewApiClient(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	resourceID, diags := clients.ResourceIDFromStr(d.Id())
	if diags.HasError() {
		return diags
	}

	if diags := client.DeleteWatch(ctx, resourceID); diags.HasError() {
		return diags
	}
	return nil
}
