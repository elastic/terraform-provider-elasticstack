package logstash

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var (
	allSettingsKeys = map[string]schema.ValueType{
		"pipeline.batch.delay":         schema.TypeInt,
		"pipeline.batch.size":          schema.TypeInt,
		"pipeline.ecs_compatibility":   schema.TypeString,
		"pipeline.ordered":             schema.TypeString,
		"pipeline.plugin_classloaders": schema.TypeBool,
		"pipeline.unsafe_shutdown":     schema.TypeBool,
		"pipeline.workers":             schema.TypeInt,
		"queue.checkpoint.acks":        schema.TypeInt,
		"queue.checkpoint.retry":       schema.TypeBool,
		"queue.checkpoint.writes":      schema.TypeInt,
		"queue.drain":                  schema.TypeBool,
		"queue.max_bytes.number":       schema.TypeInt,
		"queue.max_bytes.units":        schema.TypeString,
		"queue.max_events":             schema.TypeInt,
		"queue.page_capacity":          schema.TypeString,
		"queue.type":                   schema.TypeString,
	}
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
		// Pipeline Settings
		"pipeline_batch_delay": {
			Description: "Time in milliseconds to wait for each event before sending an undersized batch to pipeline workers.",
			Type:        schema.TypeInt,
			Optional:    true,
		},
		"pipeline_batch_size": {
			Description: "The maximum number of events an individual worker thread collects before executing filters and outputs.",
			Type:        schema.TypeInt,
			Optional:    true,
		},
		"pipeline_ecs_compatibility": {
			Description:  "Sets the pipeline default value for ecs_compatibility, a setting that is available to plugins that implement an ECS compatibility mode for use with the Elastic Common Schema.",
			Type:         schema.TypeString,
			ValidateFunc: validation.StringInSlice([]string{"disabled", "v1", "v8"}, false),
			Optional:     true,
		},
		"pipeline_ordered": {
			Description:  "Set the pipeline event ordering.",
			Type:         schema.TypeString,
			ValidateFunc: validation.StringInSlice([]string{"auto", "true", "false"}, false),
			Optional:     true,
		},
		"pipeline_plugin_classloaders": {
			Description: "(Beta) Load Java plugins in independent classloaders to isolate their dependencies.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"pipeline_unsafe_shutdown": {
			Description: "Forces Logstash to exit during shutdown even if there are still inflight events in memory.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"pipeline_workers": {
			Description:  "The number of parallel workers used to run the filter and output stages of the pipeline.",
			Type:         schema.TypeInt,
			Optional:     true,
			ValidateFunc: validation.IntAtLeast(1),
		},
		"queue_checkpoint_acks": {
			Description: "The maximum number of ACKed events before forcing a checkpoint when persistent queues are enabled.",
			Type:        schema.TypeInt,
			Optional:    true,
		},
		"queue_checkpoint_retry": {
			Description: "When enabled, Logstash will retry four times per attempted checkpoint write for any checkpoint writes that fail. Any subsequent errors are not retried.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"queue_checkpoint_writes": {
			Description: "The maximum number of written events before forcing a checkpoint when persistent queues are enabled.",
			Type:        schema.TypeInt,
			Optional:    true,
		},
		"queue_drain": {
			Description: "When enabled, Logstash waits until the persistent queue is drained before shutting down.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"queue_max_bytes_number": {
			Description: "The total capacity of the queue when persistent queues are enabled.",
			Type:        schema.TypeInt,
			Optional:    true,
		},
		"queue_max_bytes_units": {
			Description:  "Units for the total capacity of the queue when persistent queues are enabled.",
			Type:         schema.TypeString,
			ValidateFunc: validation.StringInSlice([]string{"b", "kb", "mb", "gb", "tb", "pb"}, false),
			Optional:     true,
		},
		"queue_max_events": {
			Description: "The maximum number of unread events in the queue when persistent queues are enabled.",
			Type:        schema.TypeInt,
			Optional:    true,
		},
		"queue_page_capacity": {
			Description: "The size of the page data files used when persistent queues are enabled. The queue data consists of append-only data files separated into pages.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"queue_type": {
			Description:  "The internal queueing model for event buffering. Options are memory for in-memory queueing, or persisted for disk-based acknowledged queueing.",
			Type:         schema.TypeString,
			ValidateFunc: validation.StringInSlice([]string{"memory", "persisted"}, false),
			Optional:     true,
		},
		// Pipeline Settings - End
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

	var logstashPipeline models.LogstashPipeline

	logstashPipeline.PipelineID = pipelineID
	logstashPipeline.Description = d.Get("description").(string)
	logstashPipeline.LastModified = utils.FormatStrictDateTime(time.Now().UTC())
	logstashPipeline.Pipeline = d.Get("pipeline").(string)
	logstashPipeline.PipelineMetadata = d.Get("pipeline_metadata").(map[string]interface{})

	logstashPipeline.PipelineSettings = map[string]interface{}{}
	if settings := utils.ExpandIndividuallyDefinedSettings(ctx, d, allSettingsKeys); len(settings) > 0 {
		logstashPipeline.PipelineSettings = settings
	}

	logstashPipeline.Username = d.Get("username").(string)

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
	for key, typ := range allSettingsKeys {
		var value interface{}
		if v, ok := logstashPipeline.PipelineSettings[key]; ok {
			value = v
		} else {
			tflog.Warn(ctx, fmt.Sprintf("setting '%s' is not currently managed by terraform provider and has been ignored", key))
			continue
		}
		switch typ {
		case schema.TypeInt:
			value = int(math.Round(value.(float64)))
		}
		if err := d.Set(utils.ConvertSettingsKeyToTFFieldKey(key), value); err != nil {
			return diag.FromErr(err)
		}
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
