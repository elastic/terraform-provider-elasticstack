package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceIngestPipeline() *schema.Resource {
	pipelineSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"name": {
			Description: "The name of the ingest pipeline.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"description": {
			Description: "Description of the ingest pipeline.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"on_failure": {
			Description: "Processors to run immediately after a processor failure. Each processor supports a processor-level `on_failure` value. If a processor without an `on_failure` value fails, Elasticsearch uses this pipeline-level parameter as a fallback. The processors in this parameter run sequentially in the order specified. Elasticsearch will not attempt to run the pipeline’s remaining processors. See: https://www.elastic.co/guide/en/elasticsearch/reference/current/processors.html. Each record must be a valid JSON document",
			Type:        schema.TypeList,
			Optional:    true,
			MinItems:    1,
			Elem: &schema.Schema{
				Type:             schema.TypeString,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: utils.DiffJsonSuppress,
			},
		},
		"processors": {
			Description: "Processors used to perform transformations on documents before indexing. Processors run sequentially in the order specified. See: https://www.elastic.co/guide/en/elasticsearch/reference/current/processors.html. Each record must be a valid JSON document.",
			Type:        schema.TypeList,
			Required:    true,
			MinItems:    1,
			Elem: &schema.Schema{
				Type:             schema.TypeString,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: utils.DiffJsonSuppress,
			},
		},
		"metadata": {
			Description:      "Optional user metadata about the index template.",
			Type:             schema.TypeString,
			Optional:         true,
			ValidateFunc:     validation.StringIsJSON,
			DiffSuppressFunc: utils.DiffJsonSuppress,
		},
	}

	utils.AddConnectionSchema(pipelineSchema)

	return &schema.Resource{
		Description: "Manages tasks and resources related to ingest pipelines and processors. See: https://www.elastic.co/guide/en/elasticsearch/reference/current/ingest-apis.html",

		CreateContext: resourceIngestPipelineTemplatePut,
		UpdateContext: resourceIngestPipelineTemplatePut,
		ReadContext:   resourceIngestPipelineTemplateRead,
		DeleteContext: resourceIngestPipelineTemplateDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: pipelineSchema,
	}
}

func resourceIngestPipelineTemplatePut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := clients.NewApiClient(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	pipelineId := d.Get("name").(string)
	id, diags := client.ID(ctx, pipelineId)
	if diags.HasError() {
		return diags
	}
	var pipeline models.IngestPipeline
	pipeline.Name = pipelineId
	if v, ok := d.GetOk("description"); ok {
		r := v.(string)
		pipeline.Description = &r
	}
	if v, ok := d.GetOk("on_failure"); ok {
		onFailure := make([]map[string]interface{}, len(v.([]interface{})))
		for i, f := range v.([]interface{}) {
			item := make(map[string]interface{})
			if err := json.NewDecoder(strings.NewReader(f.(string))).Decode(&item); err != nil {
				return diag.FromErr(err)
			}
			onFailure[i] = item
		}
		pipeline.OnFailure = onFailure
	}
	if v, ok := d.GetOk("processors"); ok {
		procs := make([]map[string]interface{}, len(v.([]interface{})))
		for i, f := range v.([]interface{}) {
			item := make(map[string]interface{})
			if err := json.NewDecoder(strings.NewReader(f.(string))).Decode(&item); err != nil {
				return diag.FromErr(err)
			}
			procs[i] = item
		}
		pipeline.Processors = procs
	}
	if v, ok := d.GetOk("metadata"); ok {
		metadata := make(map[string]interface{})
		if err := json.NewDecoder(strings.NewReader(v.(string))).Decode(&metadata); err != nil {
			return diag.FromErr(err)
		}
		pipeline.Metadata = metadata
	}

	if diags := client.PutElasticsearchIngestPipeline(ctx, &pipeline); diags.HasError() {
		return diags
	}

	d.SetId(id.String())
	return resourceIngestPipelineTemplateRead(ctx, d, meta)
}

func resourceIngestPipelineTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	pipeline, diags := client.GetElasticsearchIngestPipeline(ctx, &compId.ResourceId)
	if pipeline == nil && diags == nil {
		tflog.Warn(ctx, fmt.Sprintf(`Injest pipeline "%s" not found, removing from state`, compId.ResourceId))
		d.SetId("")
		return diags
	}
	if diags.HasError() {
		return diags
	}
	if err := d.Set("name", pipeline.Name); err != nil {
		return diag.FromErr(err)
	}
	if desc := pipeline.Description; desc != nil {
		if err := d.Set("description", desc); err != nil {
			return diag.FromErr(err)
		}
	}
	if onFailure := pipeline.OnFailure; onFailure != nil {
		fProcs := make([]string, len(onFailure))
		for i, v := range onFailure {
			res, err := json.Marshal(v)
			if err != nil {
				return diag.FromErr(err)
			}
			fProcs[i] = string(res)
		}

		if err := d.Set("on_failure", fProcs); err != nil {
			return diag.FromErr(err)
		}
	}
	procs := make([]string, len(pipeline.Processors))
	for i, v := range pipeline.Processors {
		res, err := json.Marshal(v)
		if err != nil {
			return diag.FromErr(err)
		}
		procs[i] = string(res)
	}

	if err := d.Set("processors", procs); err != nil {
		return diag.FromErr(err)
	}

	if meta := pipeline.Metadata; meta != nil {
		meta, err := json.Marshal(meta)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("metadata", string(meta)); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func resourceIngestPipelineTemplateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	if diags := client.DeleteElasticsearchIngestPipeline(ctx, &compId.ResourceId); diags.HasError() {
		return diags
	}

	return diags
}
