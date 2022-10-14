package logstash

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceLogstashPipeline() *schema.Resource {
	logstashPipelineSchema := map[string]*schema.Schema{
		"pipeline_id": {
			Description: "Identifier for the pipeline.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"description": {
			Description: "Description of the pipeline.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"last_modified": {
			Description: "Date the pipeline was last updated.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"pipeline": {
			Description: "Configuration for the pipeline.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"pipeline_metadata": {
			Description:      "Optional metadata about the pipeline.",
			Type:             schema.TypeString,
			ValidateFunc:     validation.StringIsJSON,
			DiffSuppressFunc: utils.DiffJsonSuppress,
			Optional:         true,
			Default:          "{}",
		},
		"pipeline_settings": {
			Description:      "Settings for the pipeline. Supports only flat keys in dot notation.",
			Type:             schema.TypeString,
			ValidateFunc:     validation.StringIsJSON,
			DiffSuppressFunc: utils.DiffJsonSuppress,
			Optional:         true,
			Default:          "{}",
		},
		"username": {
			Description: "User who last updated the pipeline.",
			Type:        schema.TypeString,
			Optional:    true,
			DefaultFunc: schema.EnvDefaultFunc("ELASTICSEARCH_USERNAME", "unknown_user"),
		},
	}

	utils.AddConnectionSchema(logstashPipelineSchema)

	return &schema.Resource{
		Description: "Manage Logstash Pipelines via Centralized Pipeline Management. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/logstash-apis.html",

		CreateContext: resourceLogstashPipelinePut,
		UpdateContext: resourceLogstashPipelinePut,
		ReadContext:   resourceLogstashPipelineRead,
		DeleteContext: resourceLogstashPipelineDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: logstashPipelineSchema,
	}
}

func resourceLogstashPipelinePut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := clients.NewApiClient(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	pipelineID := d.Get("pipeline_id").(string)
	id, diags := client.ID(ctx, pipelineID)
	if diags.HasError() {
		return diags
	}

	logstashPipeline := models.LogstashPipeline{
		PipelineID:       pipelineID,
		Description:      d.Get("description").(string),
		LastModified:     formatStrictDateTime(time.Now()),
		Pipeline:         d.Get("pipeline").(string),
		PipelineMetadata: json.RawMessage(d.Get("pipeline_metadata").(string)),
		PipelineSettings: json.RawMessage(d.Get("pipeline_settings").(string)),
		Username:         d.Get("username").(string),
	}

	if diags := client.PutLogstashPipeline(ctx, &logstashPipeline); diags.HasError() {
		return diags
	}

	d.SetId(id.String())
	return resourceLogstashPipelineRead(ctx, d, meta)
}

func resourceLogstashPipelineRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := clients.NewApiClient(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	resourceID, diags := clients.ResourceIDFromStr(d.Id())
	if diags.HasError() {
		return diags
	}

	logstashPipeline, diags := client.GetLogstashPipeline(ctx, resourceID)
	if logstashPipeline == nil && diags == nil {
		d.SetId("")
		return diags
	}
	if diags.HasError() {
		return diags
	}

	if err := d.Set("pipeline_id", logstashPipeline.PipelineID); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("description", logstashPipeline.Description); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("last_modified", logstashPipeline.LastModified); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("pipeline", logstashPipeline.Pipeline); err != nil {
		return diag.FromErr(err)
	}

	pipelineMetadata, err := json.Marshal(logstashPipeline.PipelineMetadata)
	if err != nil {
		diag.FromErr(err)
	}
	if err := d.Set("pipeline_metadata", string(pipelineMetadata)); err != nil {
		return diag.FromErr(err)
	}

	pipelineSettings, err := json.Marshal(logstashPipeline.PipelineSettings)
	if err != nil {
		diag.FromErr(err)
	}
	if err := d.Set("pipeline_settings", string(pipelineSettings)); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("username", logstashPipeline.Username); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceLogstashPipelineDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := clients.NewApiClient(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	resourceID, diags := clients.ResourceIDFromStr(d.Id())
	if diags.HasError() {
		return diags
	}

	if diags := client.DeleteLogstashPipeline(ctx, resourceID); diags.HasError() {
		return diags
	}
	return nil
}

func formatStrictDateTime(t time.Time) string {
	strictDateTime := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d.%03dZ", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond())
	return strictDateTime
}
