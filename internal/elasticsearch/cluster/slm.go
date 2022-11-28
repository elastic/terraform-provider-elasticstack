package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceSlm() *schema.Resource {
	slmSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"name": {
			Description: "ID for the snapshot lifecycle policy you want to create or update.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
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
	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}
	slmId := d.Get("name").(string)
	id, diags := client.ID(ctx, slmId)
	if diags.HasError() {
		return diags
	}

	var slm models.SnapshotPolicy
	slm.Id = slmId
	var slmConfig models.SnapshotPolicyConfig
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

	if v, ok := d.GetOk("expand_wildcards"); ok {
		vv := v.(string)
		slmConfig.ExpandWildcards = &vv
	}
	if v, ok := d.GetOk("ignore_unavailable"); ok {
		vv := v.(bool)
		slmConfig.IgnoreUnavailable = &vv
	}
	if v, ok := d.GetOk("include_global_state"); ok {
		vv := v.(bool)
		slmConfig.IncludeGlobalState = &vv
	}
	indices := make([]string, 0)
	if v, ok := d.GetOk("indices"); ok {
		p := v.(*schema.Set)
		for _, e := range p.List() {
			indices = append(indices, e.(string))
		}
	}
	slmConfig.Indices = indices
	states := make([]string, 0)
	if v, ok := d.GetOk("feature_states"); ok {
		p := v.(*schema.Set)
		for _, e := range p.List() {
			states = append(states, e.(string))
		}
	}
	slmConfig.FeatureStates = states
	if v, ok := d.GetOk("metadata"); ok {
		metadata := make(map[string]interface{})
		if err := json.NewDecoder(strings.NewReader(v.(string))).Decode(&metadata); err != nil {
			return diag.FromErr(err)
		}
		slmConfig.Metadata = metadata
	}
	if v, ok := d.GetOk("partial"); ok {
		vv := v.(bool)
		slmConfig.Partial = &vv
	}

	slm.Config = &slmConfig

	if diags := elasticsearch.PutSlm(ctx, client, &slm); diags.HasError() {
		return diags
	}
	d.SetId(id.String())
	return resourceSlmRead(ctx, d, meta)
}

func resourceSlmRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}
	id, diags := clients.CompositeIdFromStr(d.Id())
	if diags.HasError() {
		return diags
	}

	slm, diags := elasticsearch.GetSlm(ctx, client, id.ResourceId)
	if slm == nil && diags == nil {
		tflog.Warn(ctx, fmt.Sprintf(`SLM policy "%s" not found, removing from state`, id.ResourceId))
		d.SetId("")
		return diags
	}
	if diags.HasError() {
		return diags
	}

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

	if c := slm.Config; c != nil {
		if c.ExpandWildcards != nil {
			if err := d.Set("expand_wildcards", *c.ExpandWildcards); err != nil {
				return diag.FromErr(err)
			}
		}

		if c.IncludeGlobalState != nil {
			if err := d.Set("include_global_state", *c.IncludeGlobalState); err != nil {
				return diag.FromErr(err)
			}
		}
		if c.IgnoreUnavailable != nil {
			if err := d.Set("ignore_unavailable", *c.IgnoreUnavailable); err != nil {
				return diag.FromErr(err)
			}
		}
		if c.Partial != nil {
			if err := d.Set("partial", *c.Partial); err != nil {
				return diag.FromErr(err)
			}
		}
		if c.Metadata != nil {
			meta, err := json.Marshal(c.Metadata)
			if err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set("metadata", meta); err != nil {
				return diag.FromErr(err)
			}
		}
		if err := d.Set("indices", c.Indices); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("feature_states", c.FeatureStates); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func resourceSlmDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}
	id, diags := clients.CompositeIdFromStr(d.Id())
	if diags.HasError() {
		return diags
	}
	if diags := elasticsearch.DeleteSlm(ctx, client, id.ResourceId); diags.HasError() {
		return diags
	}
	return diags
}
