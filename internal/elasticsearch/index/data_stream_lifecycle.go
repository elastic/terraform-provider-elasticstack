package index

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceDataStreamLifecycle() *schema.Resource {
	dataStreamLifecycleSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"name": {
			Description: "Name of the data stream. Supports wildcards (*). To target all data streams use * or _all.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"expand_wildcards": {
			Description: `Type of data stream that wildcard patterns can match. Supports comma-separated values, such as open,hidden. Valid values are:

			all, hidden - Match any data stream, including hidden ones.
			open, closed - Matches any non-hidden data stream. Data streams cannot be closed.
			none - Wildcard patterns are not accepted.
			 
			Defaults to open.`,
			Type:             schema.TypeString,
			Default:          "open",
			Optional:         true,
			ValidateDiagFunc: utils.AllowedExpandWildcards,
		},
		"data_retention": {
			Description: "If defined, every document added to this data stream will be stored at least for this time frame. Any time after this duration the document could be deleted. When empty, every document in this data stream will be stored indefinitely",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"enabled": {
			Description: "If defined, it turns data stream lifecycle on/off (true/false) for this data stream. A data stream lifecycle that is disabled (enabled: false) will have no effect on the data stream. Defaults to true.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},
		"downsampling": {
			Description: "An optional array of downsampling configuration objects, each defining an after interval representing when the backing index is meant to be downsampled (the time frame is calculated since the index was rolled over, i.e. generation time) and a fixed_interval representing the downsampling interval (the minimum fixed_interval value is 5m). A maximum number of 10 downsampling rounds can be configured",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"after": {
						Description: "Interval representing when the backing index is meant to be downsampled",
						Type:        schema.TypeString,
						Required:    true,
					},
					"fixed_interval": {
						Description: "The interval at which to aggregate the original time series index. For example, 60m produces a document for each 60 minute (hourly) interval. This follows standard time formatting syntax as used elsewhere in Elasticsearch.",
						Type:        schema.TypeString,
						Required:    true,
					},
				},
			},
		},
		"lifecycles": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"data_retention": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"enabled": {
						Type:     schema.TypeBool,
						Computed: true,
					},
					"downsampling": {
						Type:     schema.TypeList,
						Computed: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"after": {
									Type:     schema.TypeString,
									Computed: true,
								},
								"fixed_interval": {
									Type:     schema.TypeString,
									Computed: true,
								},
							},
						},
					},
				},
			},
		},
	}

	utils.AddConnectionSchema(dataStreamLifecycleSchema)

	return &schema.Resource{
		Description: "Configures the data stream lifecycle for the targeted data streams, see: https://www.elastic.co/guide/en/elasticsearch/reference/current/data-stream-apis.html",

		CreateContext: resourceDataStreamLifecyclePut,
		UpdateContext: resourceDataStreamLifecyclePut,
		ReadContext:   resourceDataStreamLifecycleRead,
		DeleteContext: resourceDataStreamLifecycleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: dataStreamLifecycleSchema,
	}
}

func resourceDataStreamLifecyclePut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}
	dsId := d.Get("name").(string)
	id, diags := client.ID(ctx, dsId)
	if diags.HasError() {
		return diags
	}
	expand_wildcards := d.Get("expand_wildcards").(string)
	var ls models.LifecycleSettings

	ls.DataRetention = d.Get("data_retention").(string)
	ls.Enabled = d.Get("enabled").(bool)

	if v, ok := d.Get("downsampling").([]interface{}); ok && len(v) > 0 {
		ls.Downsampling = make([]models.Downsampling, len(v))
		for i, ds := range v {
			if dsMap, ok := ds.(map[string]interface{}); ok {
				ls.Downsampling[i] = models.Downsampling{
					After:         dsMap["after"].(string),
					FixedInterval: dsMap["fixed_interval"].(string),
				}
			}
		}
	}

	if diags := elasticsearch.PutDataStreamLifecycle(ctx, client, dsId, expand_wildcards, ls); diags.HasError() {
		return diags
	}

	d.SetId(id.String())
	return resourceDataStreamLifecycleRead(ctx, d, meta)
}

func resourceDataStreamLifecycleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}
	id := d.Id()
	compId, diags := clients.CompositeIdFromStr(id)
	if diags.HasError() {
		return diags
	}
	expand_wildcards := d.Get("expand_wildcards").(string)

	ds, diags := elasticsearch.GetDataStreamLifecycle(ctx, client, compId.ResourceId, expand_wildcards)
	if ds == nil && diags == nil && len(*ds) == 0 {
		// no data stream found on ES side
		tflog.Warn(ctx, fmt.Sprintf(`Data stream "%s" not found, removing from state`, compId.ResourceId))
		d.SetId("")
		return diags
	}
	if diags.HasError() {
		return diags
	}

	lifecycles := make([]interface{}, 0)

	for _, lf := range *ds {
		downsampling := make([]interface{}, 0)
		for _, ds := range lf.Lifecycle.Downsampling {
			downsampling = append(downsampling, map[string]interface{}{
				"after":          ds.After,
				"fixed_interval": ds.FixedInterval,
			})
		}
		lifecycleMap := map[string]interface{}{
			"name":           lf.Name,
			"enabled":        lf.Lifecycle.Enabled,
			"data_retention": lf.Lifecycle.DataRetention,
			"downsampling":   downsampling,
		}
		lifecycles = append(lifecycles, lifecycleMap)
	}

	if err := d.Set("lifecycles", lifecycles); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("name", compId.ResourceId); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceDataStreamLifecycleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}
	id := d.Id()
	compId, diags := clients.CompositeIdFromStr(id)
	if diags.HasError() {
		return diags
	}
	expand_wildcards := d.Get("expand_wildcards").(string)
	if diags := elasticsearch.DeleteDataStreamLifecycle(ctx, client, compId.ResourceId, expand_wildcards); diags.HasError() {
		return diags
	}

	return diags
}
