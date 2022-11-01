package cluster

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceScript() *schema.Resource {
	scriptSchema := map[string]*schema.Schema{
		"script_id": {
			Description: "Identifier for the stored script. Must be unique within the cluster.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"lang": {
			Description:  "Script language. For search templates, use `mustache`.",
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"painless", "expression", "mustache", "java"}, false),
		},
		"source": {
			Description: "For scripts, a string containing the script. For search templates, an object containing the search template.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"params": {
			Description:      "Parameters for the script or search template.",
			Type:             schema.TypeString,
			Optional:         true,
			DiffSuppressFunc: utils.DiffJsonSuppress,
			ValidateFunc:     validation.StringIsJSON,
		},
		"context": {
			Description: "Context in which the script or search template should run.",
			Type:        schema.TypeString,
			Optional:    true,
		},
	}
	utils.AddConnectionSchema(scriptSchema)

	return &schema.Resource{
		Description: "Creates or updates a stored script or search template. See https://www.elastic.co/guide/en/elasticsearch/reference/current/create-stored-script-api.html",

		CreateContext: resourceScriptPut,
		UpdateContext: resourceScriptPut,
		ReadContext:   resourceScriptRead,
		DeleteContext: resourceScriptDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: scriptSchema,
	}
}

func resourceScriptRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client, err := clients.NewApiClient(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	id := d.Id()
	compId, diags := clients.CompositeIdFromStr(id)
	if diags.HasError() {
		return diags
	}

	script, diags := client.GetElasticsearchScript(ctx, compId.ResourceId)
	if script == nil && diags == nil {
		tflog.Warn(ctx, fmt.Sprintf(`Script "%s" not found, removing from state`, compId.ResourceId))
		d.SetId("")
	}
	if diags.HasError() {
		return diags
	}

	if err := d.Set("script_id", compId.ResourceId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("lang", script.Language); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("source", script.Source); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceScriptPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := clients.NewApiClient(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	scriptID := d.Get("script_id").(string)
	id, diags := client.ID(ctx, scriptID)
	if diags.HasError() {
		return diags
	}

	script := models.Script{
		ID:       scriptID,
		Language: d.Get("lang").(string),
		Source:   d.Get("source").(string),
	}
	if paramsJSON, ok := d.GetOk("params"); ok {
		var params map[string]interface{}
		bytes := []byte(paramsJSON.(string))
		err = json.Unmarshal(bytes, &params)
		if err != nil {
			return diag.FromErr(err)
		}
		script.Params = params
	}
	if scriptContext, ok := d.GetOk("context"); ok {
		script.Context = scriptContext.(string)
	}
	if diags := client.PutElasticsearchScript(ctx, &script); diags.HasError() {
		return diags
	}

	d.SetId(id.String())
	return resourceScriptRead(ctx, d, meta)
}

func resourceScriptDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := clients.NewApiClient(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	compId, diags := clients.CompositeIdFromStr(d.Id())
	if diags.HasError() {
		return diags
	}
	return client.DeleteElasticsearchScript(ctx, compId.ResourceId)
}
