package cluster

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceSlm() *schema.Resource {
	slmSchema := map[string]*schema.Schema{
		"name": {
			Description: "ID for the snapshot lifecycle policy you want to create or update.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"config": {
			Description: "Configuration for each snapshot created by the policy.",
			Type:        schema.TypeList,
			Required:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"expand_wildcards": {
						Description: "Determines how wildcard patterns in the `indices` parameter match data streams and indices. Supports comma-separated values, such as `closed,hidden`.",
						Type:        schema.TypeString,
						Optional:    true,
						Default:     "open,hidden",
						ValidateDiagFunc: func(value interface{}, path cty.Path) diag.Diagnostics {
							validValues := []string{"all", "open", "closed", "hidden", "none"}

							var diags diag.Diagnostics
							for _, pv := range strings.Split(value.(string), ",") {
								found := false
								for _, vv := range validValues {
									if vv == strings.TrimSpace(pv) {
										found = true
										break
									}
								}
								if !found {
									diags = append(diags, diag.Diagnostic{
										Severity: diag.Error,
										Summary:  "Invalid value was provided.",
										Detail:   fmt.Sprintf(`"%s" is not valid value for this field.`, pv),
									})
									return diags
								}
							}
							return diags
						},
					},
					"ignore_unavailable": {
						Description: "If `false`, the snapshot fails if any data stream or index in indices is missing or closed. If `true`, the snapshot ignores missing or closed data streams and indices.",
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     false,
					},
					"include_global_state": {
						Description: "If `true`, include the cluster state in the snapshot.",
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     true,
					},
					"indices": {
						Description: "Comma-separated list of data streams and indices to include in the snapshot.",
						Type:        schema.TypeSet,
						Optional:    true,
						Computed:    true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"feature_states": {
						Description: "Feature states to include in the snapshot.",
						Type:        schema.TypeSet,
						Optional:    true,
						Computed:    true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"metadata": {
						Description:      "Attaches arbitrary metadata to the snapshot.",
						Type:             schema.TypeString,
						Optional:         true,
						Computed:         true,
						ValidateFunc:     validation.StringIsJSON,
						DiffSuppressFunc: utils.DiffJsonSuppress,
					},
					"partial": {
						Description: "If `false`, the entire snapshot will fail if one or more indices included in the snapshot do not have all primary shards available.",
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     false,
					},
				},
			},
		},
		"snapshot_name": {
			Description: "Name automatically assigned to each snapshot created by the policy.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "<snap-{now/d}>",
		},
		"repository": {
			Description: "Repository used to store snapshots created by this policy.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"expire_after": {
			Description: "Time period after which a snapshot is considered expired and eligible for deletion.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"max_count": {
			Description: "Maximum number of snapshots to retain, even if the snapshots have not yet expired.",
			Type:        schema.TypeInt,
			Optional:    true,
		},
		"min_count": {
			Description: "Minimum number of snapshots to retain, even if the snapshots have expired.",
			Type:        schema.TypeInt,
			Optional:    true,
		},
		"schedule": {
			Description: "Periodic or absolute schedule at which the policy creates snapshots.",
			Type:        schema.TypeString,
			Required:    true,
		},
	}

	utils.AddConnectionSchema(slmSchema)

	return &schema.Resource{
		Description: "Creates or updates a snapshot lifecycle policy. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/slm-api-put-policy.html",

		CreateContext: resourceSlmPut,
		UpdateContext: resourceSlmPut,
		ReadContext:   resourceSlmRead,
		DeleteContext: resourceSlmDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: slmSchema,
	}
}

func resourceSlmPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client, err := clients.NewApiClient(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	slmId := d.Get("name").(string)
	id, diags := client.ID(slmId)
	if diags.HasError() {
		return diags
	}

	var slm models.SnapshotPolicy
	slmRetention := models.SnapshortRetention{}

	slm.Repository = d.Get("repository").(string)
	slm.Schedule = d.Get("schedule").(string)
	if v, ok := d.GetOk("snapshot_name"); ok {
		slm.Name = v.(string)
	}
	if v, ok := d.GetOk("expire_after"); ok {
		vv := v.(string)
		slmRetention.ExpireAfter = &vv
	}
	if v, ok := d.GetOk("max_count"); ok {
		vv := v.(int)
		slmRetention.MaxCount = &vv
	}
	if v, ok := d.GetOk("min_count"); ok {
		vv := v.(int)
		slmRetention.MinCount = &vv
	}
	slm.Retention = &slmRetention

	// config is required, so we know the block exists
	config := d.Get("config").([]interface{})[0].(map[string]interface{})
	if v := config["expand_wildcards"]; v != nil {
		vv := v.(string)
		slm.Config.ExpandWildcards = &vv
	}
	if v := config["ignore_unavailable"]; v != nil {
		vv := v.(bool)
		slm.Config.IgnoreUnavailable = &vv
	}
	if v := config["include_global_state"]; v != nil {
		vv := v.(bool)
		slm.Config.IncludeGlobalState = &vv
	}
	if v := config["indices"]; v != nil {
		p := v.(*schema.Set)
		indices := make([]string, p.Len())
		for i, e := range p.List() {
			indices[i] = e.(string)
		}
		slm.Config.Indices = indices
	}
	if v := config["feature_states"]; v != nil {
		p := v.(*schema.Set)
		states := make([]string, p.Len())
		for i, e := range p.List() {
			states[i] = e.(string)
		}
		slm.Config.FeatureStates = states
	}
	if v := config["metadata"]; v != nil && v.(string) != "" {
		metadata := make(map[string]interface{})
		if err := json.NewDecoder(strings.NewReader(v.(string))).Decode(&metadata); err != nil {
			return diag.FromErr(err)
		}
		slm.Config.Metadata = metadata
	}
	if v := config["partial"]; v != nil {
		vv := v.(bool)
		slm.Config.Partial = &vv
	}

	slmBytes, err := json.Marshal(slm)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] sending SLM to ES API: %s", slmBytes)
	req := client.SlmPutLifecycle.WithBody(bytes.NewReader(slmBytes))
	res, err := client.SlmPutLifecycle(slmId, req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to create or update the SLM"); diags.HasError() {
		return diags
	}

	d.SetId(id.String())
	return resourceSlmRead(ctx, d, meta)
}

func resourceSlmRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client, err := clients.NewApiClient(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	id, diags := clients.CompositeIdFromStr(d.Id())
	if diags.HasError() {
		return diags
	}

	type SlmReponse = map[string]struct {
		Policy models.SnapshotPolicy `json:"policy"`
	}
	var slmResponse SlmReponse

	req := client.SlmGetLifecycle.WithPolicyID(id.ResourceId)
	res, err := client.SlmGetLifecycle(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to get SLM policy from ES API"); diags.HasError() {
		return diags
	}
	if err := json.NewDecoder(res.Body).Decode(&slmResponse); err != nil {
		return diag.FromErr(err)
	}
	if _, ok := slmResponse[id.ResourceId]; !ok {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to find the SLM policy in the response",
			Detail:   fmt.Sprintf(`Unable to find "%s" policy in the ES API response.`, id.ResourceId),
		})
		return diags
	}
	slm := slmResponse[id.ResourceId].Policy

	if err := d.Set("snapshot_name", slm.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("repository", slm.Repository); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("schedule", slm.Schedule); err != nil {
		return diag.FromErr(err)
	}
	if slm.Retention != nil {
		if v := slm.Retention.ExpireAfter; v != nil {
			if err := d.Set("expire_after", *v); err != nil {
				return diag.FromErr(err)
			}
		}
		if v := slm.Retention.MaxCount; v != nil {
			if err := d.Set("max_count", *v); err != nil {
				return diag.FromErr(err)
			}
		}
		if v := slm.Retention.MinCount; v != nil {
			if err := d.Set("min_count", *v); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	config, err := flatteSlmConfig(slm.Config)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("config", config); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func flatteSlmConfig(c models.SnapshotPolicyConfig) ([]interface{}, error) {
	config := make([]interface{}, 1)
	configMap := make(map[string]interface{})

	if c.ExpandWildcards != nil {
		configMap["expand_wildcards"] = *c.ExpandWildcards
	}
	if c.IncludeGlobalState != nil {
		configMap["include_global_state"] = *c.IncludeGlobalState
	}
	if c.IgnoreUnavailable != nil {
		configMap["ignore_unavailable"] = *c.IgnoreUnavailable
	}
	if c.Partial != nil {
		configMap["partial"] = *c.Partial
	}
	if c.Metadata != nil {
		meta, err := json.Marshal(c.Metadata)
		if err != nil {
			return nil, err
		}
		configMap["partial"] = string(meta)
	}
	configMap["indices"] = c.Indices
	configMap["feature_states"] = c.FeatureStates

	config[0] = configMap
	return config, nil
}

func resourceSlmDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client, err := clients.NewApiClient(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	id, diags := clients.CompositeIdFromStr(d.Id())
	if diags.HasError() {
		return diags
	}

	res, err := client.SlmDeleteLifecycle(id.ResourceId)
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to delete SLM policy: %s", id.ResourceId)); diags.HasError() {
		return diags
	}

	d.SetId("")
	return diags
}
