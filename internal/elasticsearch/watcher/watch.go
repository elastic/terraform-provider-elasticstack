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
		"trigger": {
			Description:      "The trigger that defines when the watch should run.",
			Type:             schema.TypeString,
			ValidateFunc:     validation.StringIsJSON,
			DiffSuppressFunc: utils.DiffJsonSuppress,
			Required:         true,
		},
		"input": {
			Description:      "The input that defines the input that loads the data for the watch.",
			Type:             schema.TypeString,
			ValidateFunc:     validation.StringIsJSON,
			DiffSuppressFunc: utils.DiffJsonSuppress,
			Optional:         true,
			Default:          "{\"none\":{}}",
		},
		"condition": {
			Description:      "The condition that defines if the actions should be run.",
			Type:             schema.TypeString,
			ValidateFunc:     validation.StringIsJSON,
			DiffSuppressFunc: utils.DiffJsonSuppress,
			Optional:         true,
			Default:          "{\"always\":{}}",
		},
		"actions": {
			Description:      "The list of actions that will be run if the condition matches.",
			Type:             schema.TypeString,
			ValidateFunc:     validation.StringIsJSON,
			DiffSuppressFunc: utils.DiffJsonSuppress,
			Optional:         true,
			Default:          "{}",
		},
		"metadata": {
			Description:      "Metadata json that will be copied into the history entries.",
			Type:             schema.TypeString,
			ValidateFunc:     validation.StringIsJSON,
			DiffSuppressFunc: utils.DiffJsonSuppress,
			Optional:         true,
			Default:          "{}",
		},
		"transform": {
			Description:      "Processes the watch payload to prepare it for the watch actions.",
			Type:             schema.TypeString,
			ValidateFunc:     validation.StringIsJSON,
			DiffSuppressFunc: utils.DiffJsonSuppress,
			Optional:         true,
		},
		"throttle_period_in_millis": {
			Description: "Minimum time in milliseconds between actions being run. Defaults to 5000.",
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     5000,
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
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}

	watchID := d.Get("watch_id").(string)
	id, diags := client.ID(ctx, watchID)
	if diags.HasError() {
		return diags
	}

	var watch models.PutWatch
	watch.WatchID = watchID
	watch.Active = d.Get("active").(bool)

	var trigger map[string]interface{}
	if err := json.Unmarshal([]byte(d.Get("trigger").(string)), &trigger); err != nil {
		return diag.FromErr(err)
	}
	watch.Body.Trigger = trigger

	var input map[string]interface{}
	if err := json.Unmarshal([]byte(d.Get("input").(string)), &input); err != nil {
		return diag.FromErr(err)
	}
	watch.Body.Input = input

	var condition map[string]interface{}
	if err := json.Unmarshal([]byte(d.Get("condition").(string)), &condition); err != nil {
		return diag.FromErr(err)
	}
	watch.Body.Condition = condition

	var actions map[string]interface{}
	if err := json.Unmarshal([]byte(d.Get("actions").(string)), &actions); err != nil {
		return diag.FromErr(err)
	}
	watch.Body.Actions = actions

	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(d.Get("metadata").(string)), &metadata); err != nil {
		return diag.FromErr(err)
	}
	watch.Body.Metadata = metadata

	if transformJSON, ok := d.GetOk("transform"); ok {
		var transform map[string]interface{}
		if err := json.Unmarshal([]byte(transformJSON.(string)), &transform); err != nil {
			return diag.FromErr(err)
		}
		watch.Body.Transform = transform
	}

	watch.Body.Throttle_period_in_millis = d.Get("throttle_period_in_millis").(int)

	if diags := elasticsearch.PutWatch(ctx, client, &watch); diags.HasError() {
		return diags
	}

	d.SetId(id.String())
	return resourceWatchRead(ctx, d, meta)
}

func resourceWatchRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
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

	trigger, err := json.Marshal(watch.Body.Trigger)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("trigger", string(trigger)); err != nil {
		return diag.FromErr(err)
	}

	input, err := json.Marshal(watch.Body.Input)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("input", string(input)); err != nil {
		return diag.FromErr(err)
	}

	condition, err := json.Marshal(watch.Body.Condition)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("condition", string(condition)); err != nil {
		return diag.FromErr(err)
	}

	actions, err := json.Marshal(watch.Body.Actions)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("actions", string(actions)); err != nil {
		return diag.FromErr(err)
	}

	metadata, err := json.Marshal(watch.Body.Metadata)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("metadata", string(metadata)); err != nil {
		return diag.FromErr(err)
	}

	if watch.Body.Transform != nil {
		transform, err := json.Marshal(watch.Body.Transform)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("transform", string(transform)); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := d.Set("throttle_period_in_millis", watch.Body.Throttle_period_in_millis); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceWatchDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
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
