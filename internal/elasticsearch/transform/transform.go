package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceTransform() *schema.Resource {
	transformSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"name": {
			Description: "Name of the transform you wish to create.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(1, 64),
				validation.StringMatch(regexp.MustCompile(`^[a-z0-9_-]+$`), "must contain only lower case alphanumeric characters, hyphens, and underscores"),
				validation.StringMatch(regexp.MustCompile(`^[a-z0-9].*[a-z0-9]$`), "must start and end with a lowercase alphanumeric character"),
			),
		},
		"description": {
			Description: "Free text description of the transform.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"source": {
			Description: "The source of the data for the transform.",
			Type:        schema.TypeList,
			Required:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"indices": {
						Description: "The source indices for the transform.",
						Type:        schema.TypeList,
						Required:    true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"query": {
						Description:      "A query clause that retrieves a subset of data from the source index.",
						Type:             schema.TypeString,
						Optional:         true,
						Default:          `{"match_all":{}}}`,
						DiffSuppressFunc: utils.DiffJsonSuppress,
						ValidateFunc:     validation.StringIsJSON,
					},
					"runtime_mappings": {
						Description: "Definitions of search-time runtime fields that can be used by the transform.",
						Type:        schema.TypeString,
						Optional:    true,
					},
				},
			},
		},
		"destination": {
			Description: "The destination for the transform.",
			Type:        schema.TypeList,
			Required:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"index": {
						Description: "The destination index for the transform.",
						Type:        schema.TypeString,
						Required:    true,
					},
					"pipeline": {
						Description: "The unique identifier for an ingest pipeline.",
						Type:        schema.TypeString,
						Optional:    true,
					},
				},
			},
		},
		"pivot": {
			Description:      "The pivot method transforms the data by aggregating and grouping it. JSON definition expected. Either 'pivot' or 'latest' must be present.",
			Type:             schema.TypeString,
			Optional:         true,
			ExactlyOneOf:     []string{"pivot", "latest"},
			DiffSuppressFunc: utils.DiffJsonSuppress,
			ValidateFunc:     validation.StringIsJSON,
			ForceNew:         true,
		},
		"latest": {
			Description:      "The latest method transforms the data by finding the latest document for each unique key. JSON definition expected. Either 'pivot' or 'latest' must be present.",
			Type:             schema.TypeString,
			Optional:         true,
			ExactlyOneOf:     []string{"pivot", "latest"},
			DiffSuppressFunc: utils.DiffJsonSuppress,
			ValidateFunc:     validation.StringIsJSON,
			ForceNew:         true,
		},
		"frequency": {
			Type:         schema.TypeString,
			Description:  "The interval between checks for changes in the source indices when the transform is running continuously. Defaults to `1m`.",
			Optional:     true,
			Default:      "1m",
			ValidateFunc: utils.StringIsElasticDuration,
		},
		"metadata": {
			Description:      "Defines optional transform metadata.",
			Type:             schema.TypeString,
			Optional:         true,
			ValidateFunc:     validation.StringIsJSON,
			DiffSuppressFunc: utils.DiffJsonSuppress,
		},
		"retention_policy": {
			Description: "Defines a retention policy for the transform.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"time": {
						Description: "Specifies that the transform uses a time field to set the retention policy.",
						Type:        schema.TypeList,
						Required:    true,
						MaxItems:    1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"field": {
									Description: "The date field that is used to calculate the age of the document.",
									Type:        schema.TypeString,
									Required:    true,
								},
								"max_age": {
									Description:  "Specifies the maximum age of a document in the destination index.",
									Type:         schema.TypeString,
									Required:     true,
									ValidateFunc: utils.StringIsElasticDuration,
								},
							},
						},
					},
				},
			},
		},
		"sync": {
			Description: "Defines the properties transforms require to run continuously.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"time": {
						Description: "Specifies that the transform uses a time field to synchronize the source and destination indices.",
						Type:        schema.TypeList,
						Required:    true,
						MaxItems:    1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"field": {
									Description: "The date field that is used to identify new documents in the source.",
									Type:        schema.TypeString,
									Required:    true,
								},
								"delay": {
									Description:  "The time delay between the current time and the latest input data time. The default value is 60s.",
									Type:         schema.TypeString,
									Optional:     true,
									Default:      "60s",
									ValidateFunc: utils.StringIsElasticDuration,
								},
							},
						},
					},
				},
			},
		},
		"settings": {
			Description: "Defines optional transform settings.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"align_checkpoints": {
						Description: "Specifies whether the transform checkpoint ranges should be optimized for performance. Default value is true.",
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     true,
					},
					"dates_as_epoch_millis": {
						Description: "Defines if dates in the output should be written as ISO formatted string (default) or as millis since epoch.",
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     false,
					},
					"deduce_mappings": {
						Description: "Specifies whether the transform should deduce the destination index mappings from the transform config. The default value is true",
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     true,
					},
					"docs_per_second": {
						Description: "Specifies a limit on the number of input documents per second. Default value is null, which disables throttling.",
						Type:        schema.TypeFloat,
						Optional:    true,
					},
					"max_page_search_size": {
						Description: "Defines the initial page size to use for the composite aggregation for each checkpoint. The default value is 500.",
						Type:        schema.TypeInt,
						Optional:    true,
					},
					"num_failure_retries": {
						Description: "Defines the number of retries on a recoverable failure before the transform task is marked as failed. The default value is the cluster-level setting num_transform_failure_retries.",
						Type:        schema.TypeInt,
						Optional:    true,
					},
					"unattended": {
						Description: "In unattended mode, the transform retries indefinitely in case of an error which means the transform never fails. Defaults to false.",
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     false,
					},
				},
			},
		},
		"defer_validation": {
			Type:        schema.TypeBool,
			Description: "When true, deferrable validations are not run upon creation, but rather when the transform is started. This behavior may be desired if the source index does not exist until after the transform is created.",
			Optional:    true,
			Default:     false,
		},
		"timeout": {
			Type:         schema.TypeString,
			Description:  "Period to wait for a response from Elastisearch when performing any management operation. If no response is received before the timeout expires, the operation fails and returns an error. Defaults to `30s`.",
			Optional:     true,
			Default:      "30s",
			ValidateFunc: utils.StringIsDuration,
		},
		"enabled": {
			Type:        schema.TypeBool,
			Description: "Controls wether the transform is started or stopped. Default is `false` (stopped).",
			Optional:    true,
			Default:     false,
		},
	}

	return &schema.Resource{
		Schema:      transformSchema,
		Description: "Manages Elasticsearch transforms. See: https://www.elastic.co/guide/en/elasticsearch/reference/current/transforms.html",

		CreateContext: resourceTransformCreate,
		ReadContext:   resourceTransformRead,
		UpdateContext: resourceTransformUpdate,
		DeleteContext: resourceTransformDelete,
	}
}

func resourceTransformCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}

	transformName := d.Get("name").(string)
	id, diags := client.ID(ctx, transformName)
	if diags.HasError() {
		return diags
	}

	transform, err := getTransformFromResourceData(ctx, d, transformName)
	if err != nil {
		return diag.FromErr(err)
	}

	timeout, err := time.ParseDuration(d.Get("timeout").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	params := models.PutTransformParams{
		DeferValidation: d.Get("defer_validation").(bool),
		Enabled:         d.Get("enabled").(bool),
		Timeout:         timeout,
	}

	if diags := elasticsearch.PutTransform(ctx, client, transform, &params); diags.HasError() {
		return diags
	}

	d.SetId(id.String())
	return resourceTransformRead(ctx, d, meta)
}

func resourceTransformRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}
	compId, diags := clients.CompositeIdFromStr(d.Id())
	if diags.HasError() {
		return diags
	}

	transformName := compId.ResourceId
	if err := d.Set("name", transformName); err != nil {
		return diag.FromErr(err)
	}

	transform, diags := elasticsearch.GetTransform(ctx, client, &transformName)
	if transform == nil && diags == nil {
		tflog.Warn(ctx, fmt.Sprintf(`Transform "%s" not found, removing from state`, compId.ResourceId))
		d.SetId("")
		return diags
	}
	if diags.HasError() {
		return diags
	}

	return diags
}

func resourceTransformUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}

	transformName := d.Get("name").(string)
	_, diags = client.ID(ctx, transformName)
	if diags.HasError() {
		return diags
	}

	updatedTransform, err := getTransformFromResourceData(ctx, d, transformName)
	if err != nil {
		return diag.FromErr(err)
	}

	updatedTransform.Pivot = nil
	updatedTransform.Latest = nil

	params := models.UpdateTransformParams{
		DeferValidation: d.Get("defer_validation").(bool),
	}

	timeout, err := time.ParseDuration(d.Get("timeout").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	params.Timeout = timeout

	params.Enabled = d.Get("enabled").(bool)

	if diags := elasticsearch.UpdateTransform(ctx, client, updatedTransform, &params); diags.HasError() {
		return diags
	}

	return resourceTransformRead(ctx, d, meta)
}

func resourceTransformDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}

	id := d.Id()
	compId, diags := clients.CompositeIdFromStr(id)
	if diags.HasError() {
		return diags
	}

	if diags := elasticsearch.DeleteTransform(ctx, client, &compId.ResourceId); diags.HasError() {
		return diags
	}

	return diags
}

func getTransformFromResourceData(ctx context.Context, d *schema.ResourceData, name string) (*models.Transform, error) {

	var transform models.Transform
	transform.Name = name

	if v, ok := d.GetOk("description"); ok {
		transform.Description = v.(string)
	}

	if v, ok := d.GetOk("source"); ok {
		definedSource := v.([]interface{})[0].(map[string]interface{})

		transform.Source = new(models.TransformSource)
		indices := make([]string, 0)
		for _, i := range definedSource["indices"].([]interface{}) {
			indices = append(indices, i.(string))
		}
		transform.Source.Indices = indices

		if v, ok := definedSource["query"]; ok && len(v.(string)) > 0 {
			var query interface{}
			if err := json.NewDecoder(strings.NewReader(v.(string))).Decode(&query); err != nil {
				return nil, err
			}
			transform.Source.Query = query
		}

		if v, ok := definedSource["runtime_mappings"]; ok && len(v.(string)) > 0 {
			var runtimeMappings interface{}
			if err := json.NewDecoder(strings.NewReader(v.(string))).Decode(&runtimeMappings); err != nil {
				return nil, err
			}
			transform.Source.RuntimeMappings = runtimeMappings
		}
	}

	if v, ok := d.GetOk("destination"); ok {

		definedDestination := v.([]interface{})[0].(map[string]interface{})

		transform.Destination = &models.TransformDestination{
			Index: definedDestination["index"].(string),
		}

		if pipeline, ok := definedDestination["pipeline"]; ok && len(pipeline.(string)) > 0 {
			transform.Destination.Pipeline = pipeline.(string)
		}
	}

	if v, ok := d.GetOk("pivot"); ok {
		var pivot interface{}
		if err := json.NewDecoder(strings.NewReader(v.(string))).Decode(&pivot); err != nil {
			return nil, err
		}
		transform.Pivot = pivot
	}

	if v, ok := d.GetOk("latest"); ok {
		var latest interface{}
		if err := json.NewDecoder(strings.NewReader(v.(string))).Decode(&latest); err != nil {
			return nil, err
		}
		transform.Latest = latest
	}

	if v, ok := d.GetOk("frequency"); ok {
		transform.Frequency = v.(string)
	}

	if v, ok := d.GetOk("metadata"); ok {
		var metadata map[string]interface{}
		if err := json.NewDecoder(strings.NewReader(v.(string))).Decode(&metadata); err != nil {
			return nil, err
		}
		transform.Meta = metadata
	}

	if v, ok := d.GetOk("retention_policy"); ok && v != nil {
		definedRetentionPolicy := v.([]interface{})[0].(map[string]interface{})

		if v, ok := definedRetentionPolicy["time"]; ok {
			retentionTime := models.TransformRetentionPolicyTime{}
			var definedRetentionTime = v.([]interface{})[0].(map[string]interface{})
			if f, ok := definedRetentionTime["field"]; ok {
				retentionTime.Field = f.(string)
			}
			if ma, ok := definedRetentionTime["max_age"]; ok {
				retentionTime.MaxAge = ma.(string)
			}
			transform.RetentionPolicy = &models.TransformRetentionPolicy{
				Time: retentionTime,
			}
		}
	}

	if v, ok := d.GetOk("sync"); ok {
		definedSync := v.([]interface{})[0].(map[string]interface{})

		if v, ok := definedSync["time"]; ok {
			syncTime := models.TransformSyncTime{}
			var definedSyncTime = v.([]interface{})[0].(map[string]interface{})
			if f, ok := definedSyncTime["field"]; ok {
				syncTime.Field = f.(string)
			}
			if d, ok := definedSyncTime["delay"]; ok {
				syncTime.Delay = d.(string)
			}
			transform.Sync = &models.TransformSync{
				Time: syncTime,
			}
		}
	}

	if v, ok := d.GetOk("settings"); ok {
		definedSettings := v.([]interface{})[0].(map[string]interface{})

		settings := models.TransformSettings{}
		if v, ok := definedSettings["align_checkpoints"]; ok {
			ac := v.(bool)
			settings.AlignCheckpoints = &ac
		}
		if v, ok := definedSettings["dates_as_epoch_millis"]; ok {
			dem := v.(bool)
			settings.DatesAsEpochMillis = &dem
		}
		if v, ok := definedSettings["deduce_mappings"]; ok {
			dm := v.(bool)
			settings.DeduceMappings = &dm
		}
		if v, ok := definedSettings["docs_per_second"]; ok {
			dps := v.(float64)
			settings.DocsPerSecond = &dps
		}
		if v, ok := definedSettings["max_page_search_size"]; ok {
			mpss := v.(int)
			settings.MaxPageSearchSize = &mpss
		}
		if v, ok := definedSettings["num_failure_retries"]; ok {
			nfr := v.(int)
			settings.NumFailureRetries = &nfr
		}
		if v, ok := definedSettings["unattended"]; ok {
			u := v.(bool)
			settings.Unattended = &u
		}

		transform.Settings = &settings
	}

	return &transform, nil
}
