package logstash

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
			DiffSuppressFunc: utils.DiffJsonSuppress,
			ValidateFunc:     validation.StringIsJSON,
			Optional:         true,
			Default:          "{}",
		},
		"pipeline_settings": {
			Description: "Settings for the pipeline. Supports only flat keys in dot notation.",
			Type:        schema.TypeSet,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"pipeline.workers": {
						Description: "The number of parallel workers used to run the filter and output stages of the pipeline.",
						Type:        schema.TypeInt,
						Optional:    true,
						Default:     1,
					},
					"pipeline.batch.size": {
						Description: "The maximum number of events an individual worker thread collects before executing filters and outputs.",
						Type:        schema.TypeInt,
						Optional:    true,
						Default:     125,
					},
					"pipeline.batch.delay": {
						Description: "Time in milliseconds to wait for each event before sending an undersized batch to pipeline workers.",
						Type:        schema.TypeInt,
						Optional:    true,
						Default:     50,
					},
					"queue.type": {
						Description: "The internal queueing model for event buffering. Options are memory for in-memory queueing, or persisted for disk-based acknowledged queueing.",
						Type:        schema.TypeString,
						Optional:    true,
						Default:     "memory",
					},
					"queue.max_bytes.number": {
						Description: "The total capacity of the queue when persistent queues are enabled.",
						Type:        schema.TypeInt,
						Optional:    true,
						Default:     1,
					},
					"queue.max_bytes.units": {
						Description: "Units for the total capacity of the queue when persistent queues are enabled.",
						Type:        schema.TypeString,
						Optional:    true,
						Default:     "gb",
					},
					"queue.checkpoint.writes": {
						Description: "The maximum number of events written before a checkpoint is forced when persistent queues are enabled.",
						Type:        schema.TypeInt,
						Optional:    true,
						Default:     1024,
					},
				},
			},
		},
		"username": {
			Description: "User who last updated the pipeline.",
			Type:        schema.TypeString,
			Computed:    true,
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

	var logstashPipeline models.LogstashPipeline
	logstashPipeline.PipelineID = pipelineID
	logstashPipeline.Description = d.Get("description").(string)
	logstashPipeline.LastModified = d.Get("last_modified").(string)
	logstashPipeline.Pipeline = d.Get("pipeline").(string)

	if v, ok := d.GetOk("pipeline_settings"); ok {
		pipelineSettings := v.(map[string]interface{})
		settings, diags := expandPipelineSettings(pipelineSettings)
		if diags.HasError() {
			return diags
		}
		logstashPipeline.PipelineSettings = settings
	}

	logstashPipeline.Username = client // How to acheive this?

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

	pipelineMetadata, err := json.Marshal(logstashPipeline.PipelineMetadata)
	if err != nil {
		diag.FromErr(err)
	}

	pipelineSettings := flattenPipelineSettings(logstashPipeline.PipelineSettings)

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
	if err := d.Set("pipeline_metadata", string(pipelineMetadata)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("pipeline_settings", pipelineSettings); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("username", logstashPipeline.Username); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceLogstashPipelineDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client, err := clients.NewApiClient(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	compId, diags := clients.CompositeIdFromStr(d.Id())
	if diags.HasError() {
		return diags
	}

	if diags := client.DeleteLogstashPipeline(ctx, compId.ResourceId); diags.HasError() {
		return diags
	}

	d.SetId("")
	return diags
}

func flattenPipelineSettings(pipelineSettings *models.LogstashPipelineSettings) map[string]interface{} {
	settings := make(map[string]interface{})
	if pipelineSettings != nil {
		settings["pipeline.workers"] = pipelineSettings.PipelineWorkers
		settings["pipeline.batch.size"] = pipelineSettings.PipelineBatchSize
		settings["pipeline.batch.delay"] = pipelineSettings.PipelineBatchDelay
		settings["queue.type"] = pipelineSettings.QueueType
		settings["queue.max_bytes.number"] = pipelineSettings.QueueMaxBytesNumber
		settings["queue.max_bytes.units"] = pipelineSettings.QueueMaxBytesUnits
		settings["queue.checkpoint.writes"] = pipelineSettings.QueueCheckpointWrites
	}
	return settings
}

func expandPipelineSettings(pipelineSettings map[string]interface{}) (*models.LogstashPipelineSettings, diag.Diagnostics) {
	var diags diag.Diagnostics
	settings := models.LogstashPipelineSettings{}

	settings.PipelineWorkers = pipelineSettings["pipeline.workers"].(int)
	settings.PipelineBatchSize = pipelineSettings["pipeline.batch.size"].(int)
	settings.PipelineBatchDelay = pipelineSettings["pipeline.batch.delay"].(int)
	settings.QueueType = pipelineSettings["queue.type"].(string)
	settings.QueueMaxBytesNumber = pipelineSettings["queue.max_bytes.number"].(int)
	settings.QueueMaxBytesUnits = pipelineSettings["queue.max_bytes.units"].(string)
	settings.QueueCheckpointWrites = pipelineSettings["queue.checkpoint.writes"].(int)

	return &settings, diags
}
