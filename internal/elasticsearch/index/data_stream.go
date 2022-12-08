package index

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceDataStream() *schema.Resource {
	dataStreamSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"name": {
			Description: "Name of the data stream to create.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(1, 255),
				validation.StringNotInSlice([]string{".", ".."}, true),
				validation.StringMatch(regexp.MustCompile(`^[^-_+]`), "cannot start with -, _, +"),
				validation.StringMatch(regexp.MustCompile(`^[a-z0-9!$%&'()+.;=@[\]^{}~_-]+$`), "must contain lower case alphanumeric characters and selected punctuation, see: https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-create-data-stream.html#indices-create-data-stream-api-path-params"),
			),
		},
		"timestamp_field": {
			Description: "Contains information about the data stream’s @timestamp field.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"indices": {
			Description: "Array of objects containing information about the data stream’s backing indices. The last item in this array contains information about the stream’s current write index.",
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"index_name": {
						Description: "Name of the backing index.",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"index_uuid": {
						Description: "Universally unique identifier (UUID) for the index.",
						Type:        schema.TypeString,
						Computed:    true,
					},
				},
			},
		},
		"generation": {
			Description: "Current generation for the data stream.",
			Type:        schema.TypeInt,
			Computed:    true,
		},
		"metadata": {
			Description: "Custom metadata for the stream, copied from the _meta object of the stream’s matching index template.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"status": {
			Description: "Health status of the data stream.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"template": {
			Description: "Name of the index template used to create the data stream’s backing indices.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"ilm_policy": {
			Description: "Name of the current ILM lifecycle policy in the stream’s matching index template.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"hidden": {
			Description: "If `true`, the data stream is hidden.",
			Type:        schema.TypeBool,
			Computed:    true,
		},
		"system": {
			Description: "If `true`, the data stream is created and managed by an Elastic stack component and cannot be modified through normal user interaction.",
			Type:        schema.TypeBool,
			Computed:    true,
		},
		"replicated": {
			Description: "If `true`, the data stream is created and managed by cross-cluster replication and the local cluster can not write into this data stream or change its mappings.",
			Type:        schema.TypeBool,
			Computed:    true,
		},
	}

	utils.AddConnectionSchema(dataStreamSchema)

	return &schema.Resource{
		Description: "Managing Elasticsearch data streams, see: https://www.elastic.co/guide/en/elasticsearch/reference/current/data-stream-apis.html",

		CreateContext: resourceDataStreamPut,
		UpdateContext: resourceDataStreamPut,
		ReadContext:   resourceDataStreamRead,
		DeleteContext: resourceDataStreamDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: dataStreamSchema,
	}
}

func resourceDataStreamPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}
	dsId := d.Get("name").(string)
	id, diags := client.ID(ctx, dsId)
	if diags.HasError() {
		return diags
	}

	if diags := elasticsearch.PutDataStream(ctx, client, dsId); diags.HasError() {
		return diags
	}

	d.SetId(id.String())
	return resourceDataStreamRead(ctx, d, meta)
}

func resourceDataStreamRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}
	id := d.Id()
	compId, diags := clients.CompositeIdFromStr(id)
	if diags.HasError() {
		return diags
	}

	ds, diags := elasticsearch.GetDataStream(ctx, client, compId.ResourceId)
	if ds == nil && diags == nil {
		// no data stream found on ES side
		tflog.Warn(ctx, fmt.Sprintf(`Data stream "%s" not found, removing from state`, compId.ResourceId))
		d.SetId("")
		return diags
	}
	if diags.HasError() {
		return diags
	}

	if err := d.Set("name", ds.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("timestamp_field", ds.TimestampField.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("generation", ds.Generation); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("status", ds.Status); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("template", ds.Template); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("ilm_policy", ds.IlmPolicy); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("hidden", ds.Hidden); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("system", ds.System); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("replicated", ds.Replicated); err != nil {
		return diag.FromErr(err)
	}
	if ds.Meta != nil {
		metadata, err := json.Marshal(ds.Meta)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("metadata", string(metadata)); err != nil {
			return diag.FromErr(err)
		}
	}

	indices := make([]interface{}, len(ds.Indices))
	for i, idx := range ds.Indices {
		index := make(map[string]interface{})
		index["index_name"] = idx.IndexName
		index["index_uuid"] = idx.IndexUUID
		indices[i] = index
	}
	if err := d.Set("indices", indices); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceDataStreamDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}
	id := d.Id()
	compId, diags := clients.CompositeIdFromStr(id)
	if diags.HasError() {
		return diags
	}
	if diags := elasticsearch.DeleteDataStream(ctx, client, compId.ResourceId); diags.HasError() {
		return diags
	}

	return diags
}
