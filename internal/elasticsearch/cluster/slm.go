// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/tfsdkutils"
	schemautil "github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const slmDefaultExpandWildcards = "open,hidden"

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
			Default:     slmDefaultExpandWildcards,
			ValidateDiagFunc: func(value any, _ cty.Path) diag.Diagnostics {
				validValues := []string{"all", "open", "closed", "hidden", "none"}

				var diags diag.Diagnostics
				for pv := range strings.SplitSeq(value.(string), ",") {
					found := slices.Contains(validValues, strings.TrimSpace(pv))
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
			Type:        schema.TypeList,
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
			DiffSuppressFunc: tfsdkutils.DiffJSONSuppress,
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

	schemautil.AddConnectionSchema(slmSchema)

	return &schema.Resource{
		Description: slmResourceDescription,

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

func resourceSlmPut(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	factory, diags := clients.ConvertMetaToFactory(meta)
	if diags.HasError() {
		return diags
	}
	client, diags := factory.GetElasticsearchClientFromSDK(d)
	if diags.HasError() {
		return diags
	}
	slmID := d.Get("name").(string)
	id, diags := client.ID(ctx, slmID)
	if diags.HasError() {
		return diags
	}

	var slm elasticsearch.SlmPolicy
	var slmConfig elasticsearch.SlmConfig
	var slmRetention elasticsearch.SlmRetention

	slm.Name = slmID
	slm.Repository = d.Get("repository").(string)
	slm.Schedule = d.Get("schedule").(string)
	if v, ok := d.GetOk("snapshot_name"); ok {
		slm.Name = v.(string)
	}
	if v, ok := d.GetOk("expire_after"); ok {
		expireAfter := v.(string)
		slmRetention.ExpireAfter = &expireAfter
	}
	if v, ok := d.GetOk("max_count"); ok {
		maxCount := v.(int)
		slmRetention.MaxCount = &maxCount
	}
	if v, ok := d.GetOk("min_count"); ok {
		minCount := v.(int)
		slmRetention.MinCount = &minCount
	}
	if slmRetention.ExpireAfter != nil || slmRetention.MaxCount != nil || slmRetention.MinCount != nil {
		slm.Retention = &slmRetention
	}

	slmConfig.ExpandWildcards = d.Get("expand_wildcards").(string)
	vvIgnore := d.Get("ignore_unavailable").(bool)
	slmConfig.IgnoreUnavailable = &vvIgnore
	vvInclude := d.Get("include_global_state").(bool)
	slmConfig.IncludeGlobalState = &vvInclude
	indices := make([]string, 0)
	if v, ok := d.GetOk("indices"); ok {
		list := v.([]any)
		for _, idx := range list {
			indices = append(indices, idx.(string))
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
		metadata := make(map[string]any)
		if err := json.NewDecoder(strings.NewReader(v.(string))).Decode(&metadata); err != nil {
			return diag.FromErr(err)
		}
		metaRaw := make(types.Metadata)
		for k, val := range metadata {
			data, err := json.Marshal(val)
			if err != nil {
				return diag.FromErr(err)
			}
			metaRaw[k] = data
		}
		slmConfig.Metadata = metaRaw
	}
	vvPartial := d.Get("partial").(bool)
	slmConfig.Partial = &vvPartial

	slm.Config = &slmConfig

	if diags := elasticsearch.PutSlm(ctx, client, slmID, &slm); diags.HasError() {
		return diags
	}
	d.SetId(id.String())
	return resourceSlmRead(ctx, d, meta)
}

func resourceSlmRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	factory, diags := clients.ConvertMetaToFactory(meta)
	if diags.HasError() {
		return diags
	}
	client, diags := factory.GetElasticsearchClientFromSDK(d)
	if diags.HasError() {
		return diags
	}
	id, diags := clients.CompositeIDFromStr(d.Id())
	if diags.HasError() {
		return diags
	}

	slm, diags := elasticsearch.GetSlm(ctx, client, id.ResourceID)
	if slm == nil && diags == nil {
		tflog.Warn(ctx, fmt.Sprintf(`SLM policy "%s" not found, removing from state`, id.ResourceID))
		d.SetId("")
		return diags
	}
	if diags.HasError() {
		return diags
	}

	if err := d.Set("name", id.ResourceID); err != nil {
		return diag.FromErr(err)
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

	var expireAfter string
	var maxCount int
	var minCount int
	if slm.Retention != nil {
		if slm.Retention.ExpireAfter != nil {
			expireAfter = *slm.Retention.ExpireAfter
		}
		if slm.Retention.MaxCount != nil {
			maxCount = *slm.Retention.MaxCount
		}
		if slm.Retention.MinCount != nil {
			minCount = *slm.Retention.MinCount
		}
	}
	if err := d.Set("expire_after", expireAfter); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("max_count", maxCount); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("min_count", minCount); err != nil {
		return diag.FromErr(err)
	}

	expandWildcards := slmDefaultExpandWildcards
	includeGlobalState := true
	ignoreUnavailable := false
	partial := false
	var metadata string
	var indices []string
	var featureStates []string

	if c := slm.Config; c != nil {
		if c.ExpandWildcards != "" {
			expandWildcards = c.ExpandWildcards
		}
		if c.IncludeGlobalState != nil {
			includeGlobalState = *c.IncludeGlobalState
		}
		if c.IgnoreUnavailable != nil {
			ignoreUnavailable = *c.IgnoreUnavailable
		}
		if c.Partial != nil {
			partial = *c.Partial
		}
		if c.Metadata != nil {
			meta := make(map[string]any)
			for k, v := range c.Metadata {
				var val any
				if err := json.Unmarshal(v, &val); err != nil {
					return diag.FromErr(fmt.Errorf("failed to unmarshal metadata key %q: %w", k, err))
				}
				meta[k] = val
			}
			metaBytes, err := json.Marshal(meta)
			if err != nil {
				return diag.FromErr(err)
			}
			metadata = string(metaBytes)
		}
		indices = c.Indices
		featureStates = c.FeatureStates
	}
	if err := d.Set("expand_wildcards", expandWildcards); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("include_global_state", includeGlobalState); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("ignore_unavailable", ignoreUnavailable); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("partial", partial); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("metadata", metadata); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("indices", indices); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("feature_states", featureStates); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceSlmDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	factory, diags := clients.ConvertMetaToFactory(meta)
	if diags.HasError() {
		return diags
	}
	client, diags := factory.GetElasticsearchClientFromSDK(d)
	if diags.HasError() {
		return diags
	}
	id, diags := clients.CompositeIDFromStr(d.Id())
	if diags.HasError() {
		return diags
	}
	if diags := elasticsearch.DeleteSlm(ctx, client, id.ResourceID); diags.HasError() {
		return diags
	}
	return diags
}
