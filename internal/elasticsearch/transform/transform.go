package transform

import (
	"context"
	"encoding/json"
	"fmt"
	//"reflect"
	"regexp"
	//"strconv"
	"strings"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	//"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	//"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
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
			Description:      "The pivot method transforms the data by aggregating and grouping it.",
			Type:             schema.TypeString,
			Optional:         true,
			AtLeastOneOf:     []string{"pivot", "latest"},
			DiffSuppressFunc: utils.DiffJsonSuppress,
			ValidateFunc:     validation.StringIsJSON,
			ForceNew:         true,
		},
		"latest": {
			Description:      "The latest method transforms the data by finding the latest document for each unique key.",
			Type:             schema.TypeString,
			Optional:         true,
			AtLeastOneOf:     []string{"pivot", "latest"},
			DiffSuppressFunc: utils.DiffJsonSuppress,
			ValidateFunc:     validation.StringIsJSON,
			ForceNew:         true,
		},
		"frequency": {
			Type:         schema.TypeString,
			Description:  "The interval between checks for changes in the source indices when the transform is running continuously. Defaults to `1m`.",
			Optional:     true,
			Default:      "1m",
			ValidateFunc: utils.StringIsDuration,
		},
		"metadata": {
			Description:      "Defines optional transform metadata.",
			Type:             schema.TypeString,
			Optional:         true,
			ValidateFunc:     validation.StringIsJSON,
			DiffSuppressFunc: utils.DiffJsonSuppress,
		},
		"defer_validation": {
			Type:        schema.TypeBool,
			Description: "When true, deferrable validations are not run upon creation, but rather when the transform is started. This behavior may be desired if the source index does not exist until after the transform is created.",
			Optional:    true,
			Default:     false,
		},
		"timeout": {
			Type:         schema.TypeString,
			Description:  "Period to wait for a response. If no response is received before the timeout expires, the request fails and returns an error. Defaults to `30s`.",
			Optional:     true,
			Default:      "30s",
			ValidateFunc: utils.StringIsDuration,
		},
	}

	utils.AddConnectionSchema(transformSchema)

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
	fmt.Println("entering resourceTransformCreate")

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

	params := models.PutTransformParams{
		DeferValidation: d.Get("defer_validation").(bool),
	}

	timeout, err := time.ParseDuration(d.Get("timeout").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	params.Timeout = timeout

	if diags := elasticsearch.PutTransform(ctx, client, transform, &params); diags.HasError() {
		return diags
	}

	d.SetId(id.String())
	return resourceTransformRead(ctx, d, meta)
}

func resourceTransformRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fmt.Println("entering resourceTransformRead")
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
	fmt.Println("entering resourceTransformUpdate")

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

	if diags := elasticsearch.UpdateTransform(ctx, client, updatedTransform, &params); diags.HasError() {
		return diags
	}

	return resourceTransformRead(ctx, d, meta)
}

func resourceTransformDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fmt.Println("entering resourceTransformDelete")
	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}

	id := d.Id()
	compId, diags := clients.CompositeIdFromStr(id)
	if diags.HasError() {
		return diags
	}

	if diags := elasticsearch.DeleteTransform(ctx, client, compId.ResourceId); diags.HasError() {
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

		indices := make([]string, 0)
		for _, i := range definedSource["indices"].([]interface{}) {
			indices = append(indices, i.(string))
		}
		transform.Source = models.TransformSource{
			Indices: indices,
		}

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
		transform.Destination = models.TransformDestination{
			Index: definedDestination["index"].(string),
		}

		if pipeline, ok := definedDestination["pipeline"]; ok {
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

	if v, ok := d.GetOk("metadata"); ok {
		metadata := make(map[string]interface{})
		if err := json.NewDecoder(strings.NewReader(v.(string))).Decode(&metadata); err != nil {
			return nil, err
		}
		transform.Meta = metadata
	}

	return &transform, nil
}
