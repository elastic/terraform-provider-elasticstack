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
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var settingsRequiredVersions map[string]*version.Version

func init() {
	settingsRequiredVersions = make(map[string]*version.Version)

	// capabilities
	settingsRequiredVersions["destination.pipeline"] = version.Must(version.NewVersion("7.3.0"))
	settingsRequiredVersions["destination.aliases"] = version.Must(version.NewVersion("8.8.0"))
	settingsRequiredVersions["frequency"] = version.Must(version.NewVersion("7.3.0"))
	settingsRequiredVersions["latest"] = version.Must(version.NewVersion("7.11.0"))
	settingsRequiredVersions["retention_policy"] = version.Must(version.NewVersion("7.12.0"))
	settingsRequiredVersions["source.runtime_mappings"] = version.Must(version.NewVersion("7.12.0"))
	settingsRequiredVersions["metadata"] = version.Must(version.NewVersion("7.16.0"))

	// settings
	settingsRequiredVersions["docs_per_second"] = version.Must(version.NewVersion("7.8.0"))
	settingsRequiredVersions["max_page_search_size"] = version.Must(version.NewVersion("7.8.0"))
	settingsRequiredVersions["dates_as_epoch_millis"] = version.Must(version.NewVersion("7.11.0"))
	settingsRequiredVersions["align_checkpoints"] = version.Must(version.NewVersion("7.15.0"))
	settingsRequiredVersions["deduce_mappings"] = version.Must(version.NewVersion("8.1.0"))
	settingsRequiredVersions["num_failure_retries"] = version.Must(version.NewVersion("8.4.0"))
	settingsRequiredVersions["unattended"] = version.Must(version.NewVersion("8.5.0"))
}

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
						Default:          `{"match_all":{}}`,
						DiffSuppressFunc: utils.DiffJsonSuppress,
						ValidateFunc:     validation.StringIsJSON,
					},
					"runtime_mappings": {
						Description:      "Definitions of search-time runtime fields that can be used by the transform.",
						Type:             schema.TypeString,
						Optional:         true,
						DiffSuppressFunc: utils.DiffJsonSuppress,
						ValidateFunc:     validation.StringIsJSON,
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
						ValidateFunc: validation.All(
							validation.StringLenBetween(1, 255),
							validation.StringNotInSlice([]string{".", ".."}, true),
							validation.StringMatch(regexp.MustCompile(`^[^-_+]`), "cannot start with -, _, +"),
							validation.StringMatch(regexp.MustCompile(`^[a-z0-9!$%&'()+.;=@[\]^{}~_-]+$`), "must contain lower case alphanumeric characters and selected punctuation, see: https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-create-index.html#indices-create-api-path-params"),
						),
					},
					"aliases": {
						Description: "The aliases that the destination index for the transform should have.",
						Type:        schema.TypeList,
						Optional:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"alias": {
									Description: "The name of the alias.",
									Type:        schema.TypeString,
									Required:    true,
								},
								"move_on_creation": {
									Description: "Whether the destination index should be the only index in this alias. Defaults to false.",
									Type:        schema.TypeBool,
									Optional:    true,
									Default:     false,
								},
							},
						},
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
						Description: "Specifies that the transform uses a time field to set the retention policy. This is currently the only supported option.",
						Type:        schema.TypeList,
						Required:    true,
						MaxItems:    1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"field": {
									Description:  "The date field that is used to calculate the age of the document.",
									Type:         schema.TypeString,
									Required:     true,
									ValidateFunc: validation.StringIsNotWhiteSpace,
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
						Description: "Specifies that the transform uses a time field to synchronize the source and destination indices. This is currently the only supported option.",
						Type:        schema.TypeList,
						Required:    true,
						MaxItems:    1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"field": {
									Description:  "The date field that is used to identify new documents in the source.",
									Type:         schema.TypeString,
									Required:     true,
									ValidateFunc: validation.StringIsNotWhiteSpace,
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
		"align_checkpoints": {
			Description: "Specifies whether the transform checkpoint ranges should be optimized for performance.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"dates_as_epoch_millis": {
			Description: "Defines if dates in the output should be written as ISO formatted string (default) or as millis since epoch.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"deduce_mappings": {
			Description: "Specifies whether the transform should deduce the destination index mappings from the transform config.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"docs_per_second": {
			Description:  "Specifies a limit on the number of input documents per second. Default (unset) value disables throttling.",
			Type:         schema.TypeFloat,
			Optional:     true,
			ValidateFunc: validation.FloatAtLeast(0),
		},
		"max_page_search_size": {
			Description:  "Defines the initial page size to use for the composite aggregation for each checkpoint. Default is 500.",
			Type:         schema.TypeInt,
			Optional:     true,
			ValidateFunc: validation.IntBetween(10, 65536),
		},
		"num_failure_retries": {
			Description:  "Defines the number of retries on a recoverable failure before the transform task is marked as failed. The default value is the cluster-level setting num_transform_failure_retries.",
			Type:         schema.TypeInt,
			Optional:     true,
			ValidateFunc: validation.IntBetween(-1, 100),
		},
		"unattended": {
			Description: "In unattended mode, the transform retries indefinitely in case of an error which means the transform never fails.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"defer_validation": {
			Type:        schema.TypeBool,
			Description: "When true, deferrable validations are not run upon creation, but rather when the transform is started. This behavior may be desired if the source index does not exist until after the transform is created. Default is `false`",
			Optional:    true,
			Default:     false,
		},
		"timeout": {
			Type:         schema.TypeString,
			Description:  "Period to wait for a response from Elasticsearch when performing any management operation. If no response is received before the timeout expires, the operation fails and returns an error. Defaults to `30s`.",
			Optional:     true,
			Default:      "30s",
			ValidateFunc: utils.StringIsDuration,
		},
		"enabled": {
			Type:        schema.TypeBool,
			Description: "Controls whether the transform should be started or stopped. Default is `false` (stopped).",
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

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceTransformCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}

	transformName := d.Get("name").(string)
	id, diags := client.ID(ctx, transformName)
	if diags.HasError() {
		return diags
	}

	serverVersion, diags := client.ServerVersion(ctx)
	if diags.HasError() {
		return diags
	}

	transform, err := getTransformFromResourceData(ctx, d, transformName, serverVersion)
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

	client, diags := clients.NewApiClientFromSDKResource(d, meta)
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

	// actual resource state is established from two sources: the transform definition (model) and the transform stats
	// 1. read transform definition
	transform, diags := elasticsearch.GetTransform(ctx, client, &transformName)
	if transform == nil && diags == nil {
		tflog.Warn(ctx, fmt.Sprintf(`Transform "%s" not found, removing from state`, compId.ResourceId))
		d.SetId("")
		return diags
	}
	if diags.HasError() {
		return diags
	}

	if err := updateResourceDataFromModel(d, transform); err != nil {
		return diag.FromErr(err)
	}

	// 2. read transform stats
	transformStats, diags := elasticsearch.GetTransformStats(ctx, client, &transformName)
	if diags.HasError() {
		return diags
	}

	if err := updateResourceDataFromStats(d, transformStats); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceTransformUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}

	transformName := d.Get("name").(string)
	_, diags = client.ID(ctx, transformName)
	if diags.HasError() {
		return diags
	}

	serverVersion, diags := client.ServerVersion(ctx)
	if diags.HasError() {
		return diags
	}

	updatedTransform, err := getTransformFromResourceData(ctx, d, transformName, serverVersion)
	if err != nil {
		return diag.FromErr(err)
	}

	// pivot and latest cannot be updated; sending them to the API for an update operation would result in an error
	updatedTransform.Pivot = nil
	updatedTransform.Latest = nil

	timeout, err := time.ParseDuration(d.Get("timeout").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	params := models.UpdateTransformParams{
		DeferValidation: d.Get("defer_validation").(bool),
		Timeout:         timeout,
		Enabled:         d.Get("enabled").(bool),
		ApplyEnabled:    d.HasChange("enabled"),
	}

	if diags := elasticsearch.UpdateTransform(ctx, client, updatedTransform, &params); diags.HasError() {
		return diags
	}

	return resourceTransformRead(ctx, d, meta)
}

func resourceTransformDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	client, diags := clients.NewApiClientFromSDKResource(d, meta)
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

func getTransformFromResourceData(ctx context.Context, d *schema.ResourceData, name string, serverVersion *version.Version) (*models.Transform, error) {

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

		if v, ok := definedSource["runtime_mappings"]; ok && len(v.(string)) > 0 && isSettingAllowed(ctx, "source.runtime_mappings", serverVersion) {
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

		if aliases, ok := definedDestination["aliases"].([]interface{}); ok && len(aliases) > 0 && isSettingAllowed(ctx, "destination.aliases", serverVersion) {
			transform.Destination.Aliases = make([]models.TransformAlias, len(aliases))
			for i, alias := range aliases {
				aliasMap := alias.(map[string]interface{})
				transform.Destination.Aliases[i] = models.TransformAlias{
					Alias:          aliasMap["alias"].(string),
					MoveOnCreation: aliasMap["move_on_creation"].(bool),
				}
			}
		}

		if pipeline, ok := definedDestination["pipeline"]; ok && isSettingAllowed(ctx, "destination.pipeline", serverVersion) {
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

	if v, ok := d.GetOk("latest"); ok && isSettingAllowed(ctx, "latest", serverVersion) {
		var latest interface{}
		if err := json.NewDecoder(strings.NewReader(v.(string))).Decode(&latest); err != nil {
			return nil, err
		}
		transform.Latest = latest
	}

	if v, ok := d.GetOk("frequency"); ok && isSettingAllowed(ctx, "frequency", serverVersion) {
		transform.Frequency = v.(string)
	}

	if v, ok := d.GetOk("metadata"); ok && isSettingAllowed(ctx, "metadata", serverVersion) {
		var metadata map[string]interface{}
		if err := json.NewDecoder(strings.NewReader(v.(string))).Decode(&metadata); err != nil {
			return nil, err
		}
		transform.Meta = metadata
	}

	if v, ok := d.GetOk("retention_policy"); ok && isSettingAllowed(ctx, "retention_policy", serverVersion) {
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

	// settings
	settings := models.TransformSettings{}
	setSettings := false

	if v, ok := d.GetOk("align_checkpoints"); ok && isSettingAllowed(ctx, "align_checkpoints", serverVersion) {
		setSettings = true
		ac := v.(bool)
		settings.AlignCheckpoints = &ac
	}
	if v, ok := d.GetOk("dates_as_epoch_millis"); ok && isSettingAllowed(ctx, "dates_as_epoch_millis", serverVersion) {
		setSettings = true
		dem := v.(bool)
		settings.DatesAsEpochMillis = &dem
	}
	if v, ok := d.GetOk("deduce_mappings"); ok && isSettingAllowed(ctx, "deduce_mappings", serverVersion) {
		setSettings = true
		dm := v.(bool)
		settings.DeduceMappings = &dm
	}
	if v, ok := d.GetOk("docs_per_second"); ok && isSettingAllowed(ctx, "docs_per_second", serverVersion) {
		setSettings = true
		dps := v.(float64)
		settings.DocsPerSecond = &dps
	}
	if v, ok := d.GetOk("max_page_search_size"); ok && isSettingAllowed(ctx, "max_page_search_size", serverVersion) {
		setSettings = true
		mpss := v.(int)
		settings.MaxPageSearchSize = &mpss
	}
	if v, ok := d.GetOk("num_failure_retries"); ok && isSettingAllowed(ctx, "num_failure_retries", serverVersion) {
		setSettings = true
		nfr := v.(int)
		settings.NumFailureRetries = &nfr
	}
	if v, ok := d.GetOk("unattended"); ok && isSettingAllowed(ctx, "unattended", serverVersion) {
		setSettings = true
		u := v.(bool)
		settings.Unattended = &u
	}

	if setSettings {
		transform.Settings = &settings
	}

	return &transform, nil
}

func updateResourceDataFromModel(d *schema.ResourceData, transform *models.Transform) error {

	// transform.Description
	if err := d.Set("description", transform.Description); err != nil {
		return err
	}

	// transform.Source
	if err := d.Set("source", flattenSource(transform.Source)); err != nil {
		return err
	}

	// transform.Destination
	if err := d.Set("destination", flattenDestination(transform.Destination)); err != nil {
		return err
	}

	// transform.Pivot
	if transform.Pivot != nil {
		pivot, err := json.Marshal(transform.Pivot)
		if err != nil {
			return err
		}
		if err := d.Set("pivot", string(pivot)); err != nil {
			return err
		}
	}

	// transform.Latest
	if transform.Latest != nil {
		latest, err := json.Marshal(transform.Latest)
		if err != nil {
			return err
		}
		if err := d.Set("latest", string(latest)); err != nil {
			return err
		}
	}

	// transform.Frequency
	if err := d.Set("frequency", transform.Frequency); err != nil {
		return err
	}

	// transform.Sync
	if err := d.Set("sync", flattenSync(transform.Sync)); err != nil {
		return err
	}

	// transform.RetentionPolicy
	if err := d.Set("retention_policy", flattenRetentionPolicy(transform.RetentionPolicy)); err != nil {
		return err
	}

	// transform.Settings
	if transform.Settings != nil && transform.Settings.AlignCheckpoints != nil {
		if err := d.Set("align_checkpoints", *(transform.Settings.AlignCheckpoints)); err != nil {
			return err
		}
	}

	if transform.Settings != nil && transform.Settings.DatesAsEpochMillis != nil {
		if err := d.Set("dates_as_epoch_millis", *(transform.Settings.DatesAsEpochMillis)); err != nil {
			return err
		}
	}

	if transform.Settings != nil && transform.Settings.DeduceMappings != nil {
		if err := d.Set("deduce_mappings", *(transform.Settings.DeduceMappings)); err != nil {
			return err
		}
	}

	if transform.Settings != nil && transform.Settings.DocsPerSecond != nil {
		if err := d.Set("docs_per_second", *(transform.Settings.DocsPerSecond)); err != nil {
			return err
		}
	}

	if transform.Settings != nil && transform.Settings.MaxPageSearchSize != nil {
		if err := d.Set("max_page_search_size", *(transform.Settings.MaxPageSearchSize)); err != nil {
			return err
		}
	}

	if transform.Settings != nil && transform.Settings.NumFailureRetries != nil {
		if err := d.Set("num_failure_retries", *(transform.Settings.NumFailureRetries)); err != nil {
			return err
		}
	}

	if transform.Settings != nil && transform.Settings.Unattended != nil {
		if err := d.Set("unattended", *(transform.Settings.Unattended)); err != nil {
			return err
		}
	}

	// transform.Meta
	if transform.Meta == nil {
		if err := d.Set("metadata", nil); err != nil {
			return err
		}
	} else {
		meta, err := json.Marshal(transform.Meta)
		if err != nil {
			return err
		}

		if err := d.Set("metadata", string(meta)); err != nil {
			return err
		}
	}

	return nil
}

func updateResourceDataFromStats(d *schema.ResourceData, transformStats *models.TransformStats) error {

	// transform.Enabled
	if err := d.Set("enabled", transformStats.IsStarted()); err != nil {
		return err
	}

	return nil
}

func flattenSource(source *models.TransformSource) []interface{} {
	if source == nil {
		return []interface{}{}
	}

	s := make(map[string]interface{})

	if source.Indices != nil {
		s["indices"] = source.Indices
	}

	if source.Query != nil {
		query, err := json.Marshal(source.Query)
		if err != nil {
			return []interface{}{}
		}
		if len(query) > 0 {
			s["query"] = string(query)
		}
	}

	if source.RuntimeMappings != nil {
		rm, err := json.Marshal(source.RuntimeMappings)
		if err != nil {
			return []interface{}{}
		}
		if len(rm) > 0 {
			s["runtime_mappings"] = string(rm)
		}
	}

	return []interface{}{s}
}

func flattenDestination(dest *models.TransformDestination) []interface{} {
	if dest == nil {
		return []interface{}{}
	}

	d := make(map[string]interface{})
	d["index"] = dest.Index

	if len(dest.Aliases) > 0 {
		aliases := make([]interface{}, len(dest.Aliases))
		for i, alias := range dest.Aliases {
			aliasMap := make(map[string]interface{})
			aliasMap["alias"] = alias.Alias
			aliasMap["move_on_creation"] = alias.MoveOnCreation
			aliases[i] = aliasMap
		}
		d["aliases"] = aliases
	}

	if dest.Pipeline != "" {
		d["pipeline"] = dest.Pipeline
	}

	return []interface{}{d}
}

func flattenSync(sync *models.TransformSync) []interface{} {
	if sync == nil {
		return nil
	}

	t := make(map[string]interface{})

	if sync.Time.Delay != "" {
		t["delay"] = sync.Time.Delay
	}

	if sync.Time.Field != "" {
		t["field"] = sync.Time.Field
	}

	s := make(map[string]interface{})
	s["time"] = []interface{}{t}

	return []interface{}{s}
}

func flattenRetentionPolicy(retention *models.TransformRetentionPolicy) []interface{} {
	if retention == nil {
		return []interface{}{}
	}

	t := make(map[string]interface{})

	if retention.Time.MaxAge != "" {
		t["max_age"] = retention.Time.MaxAge
	}

	if retention.Time.Field != "" {
		t["field"] = retention.Time.Field
	}

	r := make(map[string]interface{})
	r["time"] = []interface{}{t}

	return []interface{}{r}
}

func isSettingAllowed(ctx context.Context, settingName string, serverVersion *version.Version) bool {
	if minVersion, ok := settingsRequiredVersions[settingName]; ok {
		if serverVersion.LessThan(minVersion) {
			tflog.Warn(ctx, fmt.Sprintf("Setting [%s] not allowed for Elasticsearch server version %v; min required is %v", settingName, *serverVersion, *minVersion))
			return false
		}
	}

	return true
}
