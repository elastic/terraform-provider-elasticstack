package logstash

import (
	"context"
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
		"id": {
			Description: "Internal identifier of the resource",
			Type:        schema.TypeString,
			Computed:    true,
		},
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
			Description: "Optional metadata about the pipeline.",
			Type:        schema.TypeMap,
			Optional:    true,
			Elem: &schema.Schema{
				Type:    schema.TypeString,
				Default: nil,
			},
		},
		"pipeline_settings": {
			Description: "Settings for the pipeline.",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"pipeline_workers": {
						Description: "The number of parallel workers used to run the filter and output stages of the pipeline.",
						Type:        schema.TypeInt,
						Optional:    true,
						Default:     1,
					},
					"pipeline_batch_size": {
						Description: "The maximum number of events an individual worker thread collects before executing filters and outputs.",
						Type:        schema.TypeInt,
						Optional:    true,
						Default:     125,
					},
					"pipeline_batch_delay": {
						Description: "Time in milliseconds to wait for each event before sending an undersized batch to pipeline workers.",
						Type:        schema.TypeInt,
						Optional:    true,
						Default:     50,
					},
					"queue_type": {
						Description:  "The internal queueing model for event buffering. Options are memory for in-memory queueing, or persisted for disk-based acknowledged queueing.",
						Type:         schema.TypeString,
						ValidateFunc: validation.StringInSlice([]string{"memory", "persisted"}, false),
						Optional:     true,
						Default:      "memory",
					},
					"queue_max_bytes_number": {
						Description: "The total capacity of the queue when persistent queues are enabled.",
						Type:        schema.TypeInt,
						Optional:    true,
						Default:     1,
					},
					"queue_max_bytes_units": {
						Description:  "Units for the total capacity of the queue when persistent queues are enabled.",
						Type:         schema.TypeString,
						ValidateFunc: validation.StringInSlice([]string{"b", "kb", "mb", "gb", "tb", "pb"}, false),
						Optional:     true,
						Default:      "gb",
					},
					"queue_checkpoint_writes": {
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
			Optional:    true,
			DefaultFunc: schema.EnvDefaultFunc("ELASTICSEARCH_USERNAME", "api_key"),
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
		LastModified:     utils.FormatStrictDateTime(time.Now()),
		Pipeline:         d.Get("pipeline").(string),
		PipelineMetadata: d.Get("pipeline_metadata").(map[string]interface{}),
		PipelineSettings: expandPipelineSettings(d.Get("pipeline_settings").([]interface{})),
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
	if err := d.Set("pipeline_metadata", logstashPipeline.PipelineMetadata); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("pipeline_settings", flattenPipelineSettings(logstashPipeline.PipelineSettings)); err != nil {
		diag.FromErr(err)
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

func flattenPipelineSettings(pipelineSettings *models.LogstashPipelineSettings) []interface{} {
	settings := make(map[string]interface{})
	settings["pipeline_workers"] = pipelineSettings.PipelineWorkers
	settings["pipeline_batch_size"] = pipelineSettings.PipelineBatchSize
	settings["pipeline_batch_delay"] = pipelineSettings.PipelineBatchDelay
	settings["queue_type"] = pipelineSettings.QueueType
	settings["queue_max_bytes_number"] = pipelineSettings.QueueMaxBytesNumber
	settings["queue_max_bytes_units"] = pipelineSettings.QueueMaxBytesUnits
	settings["queue_checkpoint_writes"] = pipelineSettings.QueueCheckpointWrites

	return []interface{}{settings}
}

func expandPipelineSettings(pipelineSettings []interface{}) *models.LogstashPipelineSettings {
	var settings models.LogstashPipelineSettings
	for _, ps := range pipelineSettings {
		setting := ps.(map[string]interface{})
		settings.PipelineWorkers = setting["pipeline_workers"].(int)
		settings.PipelineBatchSize = setting["pipeline_batch_size"].(int)
		settings.PipelineBatchDelay = setting["pipeline_batch_delay"].(int)
		settings.QueueType = setting["queue_type"].(string)
		settings.QueueMaxBytesNumber = setting["queue_max_bytes_number"].(int)
		settings.QueueMaxBytesUnits = setting["queue_max_bytes_units"].(string)
		settings.QueueCheckpointWrites = setting["queue_checkpoint_writes"].(int)
	}
	return &settings
}
