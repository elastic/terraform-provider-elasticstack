package cluster

import (
	"context"
	"fmt"
	"strings"

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
										Summary:  "Not valid value was provided.",
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
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"feature_states": {
						Description: "Feature states to include in the snapshot.",
						Type:        schema.TypeSet,
						Optional:    true,
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
	return resourceSlmRead(ctx, d, meta)
}

func resourceSlmRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}

func resourceSlmDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}
