package watcher

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceWatch() *schema.Resource {
	watchSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource.",
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
			ValidateFunc:     validateWatchBody,
			DiffSuppressFunc: utils.DiffJsonSuppress,
			Required:         true,
		},
	}

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
	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}

	watchID := d.Get("watch_id").(string)
	id, diags := client.ID(ctx, watchID)
	if diags.HasError() {
		return diags
	}

	var watchBody models.WatchBody

	if err := json.Unmarshal([]byte(d.Get("body").(string)), &watchBody); err != nil {
		return diag.FromErr(err)
	}

	var watch models.Watch
	watch.WatchID = watchID
	watch.Status.State.Active = d.Get("active").(bool)
	watch.Body = watchBody

	if diags := elasticsearch.PutWatch(ctx, client, &watch); diags.HasError() {
		return diags
	}

	d.SetId(id.String())
	return resourceWatchRead(ctx, d, meta)
}

func resourceWatchRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}
	resourceID, diags := clients.ResourceIDFromStr(d.Id())
	if diags.HasError() {
		return diags
	}

	watch, diags := elasticsearch.GetWatch(ctx, client, resourceID)
	if watch == nil && diags == nil {
		tflog.Warn(ctx, fmt.Sprintf(`Watch "%s" not found, removing from state`, resourceID))
		d.SetId("")
		return diags
	}
	if diags.HasError() {
		return diags
	}

	if err := d.Set("watch_id", watch.WatchID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("active", watch.Status.State.Active); err != nil {
		return diag.FromErr(err)
	}

	body, err := json.Marshal(watch.Body)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("body", string(body)); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceWatchDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}
	resourceID, diags := clients.ResourceIDFromStr(d.Id())
	if diags.HasError() {
		return diags
	}

	if diags := elasticsearch.DeleteWatch(ctx, client, resourceID); diags.HasError() {
		return diags
	}
	return nil
}

func validateWatchBody(i interface{}, k string) (warnings []string, errors []error) {
	warnings, errors = validation.StringIsJSON(i, k)

	var watchBody models.WatchBody
	if err := json.Unmarshal([]byte(i.(string)), &watchBody); err != nil {
		panic(err)
	}

	if watchBody.Actions == nil {
		errors = append(errors, fmt.Errorf("watch field must be declared: actions"))
		return warnings, errors
	}
	if watchBody.Condition == nil {
		errors = append(errors, fmt.Errorf("watch field must be declared: condition"))
		return warnings, errors
	}
	if watchBody.Input == nil {
		errors = append(errors, fmt.Errorf("watch field must be declared: input"))
		return warnings, errors
	}
	if watchBody.Trigger == nil {
		errors = append(errors, fmt.Errorf("watch field must be declared: trigger"))
		return warnings, errors
	}
	if len(watchBody.Metadata) == 0 && watchBody.Metadata != nil {
		errors = append(errors, fmt.Errorf("metadata field should not be an empty map"))
		return warnings, errors
	}

	return warnings, errors
}
