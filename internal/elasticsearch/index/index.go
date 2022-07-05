package index

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceIndex() *schema.Resource {
	indexSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"name": {
			Description: "Name of the index you wish to create.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(1, 255),
				validation.StringNotInSlice([]string{".", ".."}, true),
				validation.StringMatch(regexp.MustCompile(`^[^-_+]`), "cannot start with -, _, +"),
				validation.StringMatch(regexp.MustCompile(`^[a-z0-9!$%&'()+.;=@[\]^{}~_-]+$`), "must contain lower case alphanumeric characters and selected punctuation, see: https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-create-index.html#indices-create-api-path-params"),
			),
		},
		"alias": {
			Description: "Aliases for the index.",
			Type:        schema.TypeSet,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Description: "Index alias name.",
						Type:        schema.TypeString,
						Required:    true,
					},
					"filter": {
						Description:      "Query used to limit documents the alias can access.",
						Type:             schema.TypeString,
						Optional:         true,
						Default:          "",
						DiffSuppressFunc: utils.DiffJsonSuppress,
						ValidateFunc:     validation.StringIsJSON,
					},
					"index_routing": {
						Description: "Value used to route indexing operations to a specific shard. If specified, this overwrites the `routing` value for indexing operations.",
						Type:        schema.TypeString,
						Optional:    true,
						Default:     "",
					},
					"is_hidden": {
						Description: "If true, the alias is hidden.",
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     false,
					},
					"is_write_index": {
						Description: "If true, the index is the write index for the alias.",
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     false,
					},
					"routing": {
						Description: "Value used to route indexing and search operations to a specific shard.",
						Type:        schema.TypeString,
						Optional:    true,
						Default:     "",
					},
					"search_routing": {
						Description: "Value used to route search operations to a specific shard. If specified, this overwrites the routing value for search operations.",
						Type:        schema.TypeString,
						Optional:    true,
						Default:     "",
					},
				},
			},
		},
		"mappings": {
			Description: `Mapping for fields in the index.
If specified, this mapping can include: field names, [field data types](https://www.elastic.co/guide/en/elasticsearch/reference/current/mapping-types.html), [mapping parameters](https://www.elastic.co/guide/en/elasticsearch/reference/current/mapping-params.html).
**NOTE:** changing datatypes in the existing _mappings_ will force index to be re-created.`,
			Type:             schema.TypeString,
			Optional:         true,
			DiffSuppressFunc: utils.DiffJsonSuppress,
			ValidateFunc:     validation.StringIsJSON,
			Default:          "{}",
		},
		"settings": {
			Description: `Configuration options for the index. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/index-modules.html#index-modules-settings.
**NOTE:** Static index settings (see: https://www.elastic.co/guide/en/elasticsearch/reference/current/index-modules.html#_static_index_settings) can be only set on the index creation and later cannot be removed or updated - _apply_ will return error`,
			Type:     schema.TypeList,
			MaxItems: 1,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"setting": {
						Description: "Defines the setting for the index.",
						Type:        schema.TypeSet,
						Required:    true,
						MinItems:    1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"name": {
									Description: "The name of the setting to set and track.",
									Type:        schema.TypeString,
									Required:    true,
								},
								"value": {
									Description: "The value of the setting to set and track.",
									Type:        schema.TypeString,
									Required:    true,
								},
							},
						},
					},
				},
			},
		},
		"settings_raw": {
			Description: "All raw settings fetched from the cluster.",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}

	utils.AddConnectionSchema(indexSchema)

	return &schema.Resource{
		Description: "Creates Elasticsearch indices. See: https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-create-index.html",

		CreateContext: resourceIndexCreate,
		UpdateContext: resourceIndexUpdate,
		ReadContext:   resourceIndexRead,
		DeleteContext: resourceIndexDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				// default settings populated by Elasticsearch, which we do not support and should ignore
				var ignoredDefaults = map[string]struct{}{
					"index.creation_date":   struct{}{},
					"index.provided_name":   struct{}{},
					"index.uuid":            struct{}{},
					"index.version.created": struct{}{},
				}

				// first populate what we can with Read
				diags := resourceIndexRead(ctx, d, m)
				if diags.HasError() {
					return nil, fmt.Errorf("Unable to import requested index")
				}

				client, err := clients.NewApiClient(d, m)
				if err != nil {
					return nil, err
				}
				compId, diags := clients.CompositeIdFromStr(d.Id())
				if diags.HasError() {
					return nil, fmt.Errorf("Failed to parse provided ID")
				}
				indexName := compId.ResourceId
				index, diags := client.GetElasticsearchIndex(indexName)
				if diags.HasError() {
					return nil, fmt.Errorf("Failed to get an ES Index")
				}
				// check the settings and import those as well
				if index.Settings != nil {
					settings := make(map[string]interface{})
					result := make([]interface{}, 0)
					for k, v := range index.Settings {
						if _, ok := ignoredDefaults[k]; ok {
							continue
						}
						setting := make(map[string]interface{})
						setting["name"] = k
						setting["value"] = v
						result = append(result, setting)
					}
					settings["setting"] = result

					if err := d.Set("settings", []interface{}{settings}); err != nil {
						return nil, err
					}
				}
				return []*schema.ResourceData{d}, nil
			},
		},

		CustomizeDiff: customdiff.ForceNewIfChange("mappings", func(ctx context.Context, old, new, meta interface{}) bool {
			o := make(map[string]interface{})
			if err := json.NewDecoder(strings.NewReader(old.(string))).Decode(&o); err != nil {
				return true
			}
			n := make(map[string]interface{})
			if err := json.NewDecoder(strings.NewReader(new.(string))).Decode(&n); err != nil {
				return true
			}
			log.Printf("[TRACE] mappings custom diff old = %+v new = %+v", o, n)

			var isForceable func(map[string]interface{}, map[string]interface{}) bool
			isForceable = func(old, new map[string]interface{}) bool {
				for k, v := range old {
					oldFieldSettings := v.(map[string]interface{})
					if newFieldSettings, ok := new[k]; ok {
						newSettings := newFieldSettings.(map[string]interface{})
						// check if the "type" field exists and match with new one
						if s, ok := oldFieldSettings["type"]; ok {
							if ns, ok := newSettings["type"]; ok {
								if !reflect.DeepEqual(s, ns) {
									return true
								}
								continue
							} else {
								return true
							}
						}

						// if we have "mapping" field, let's call ourself to check again
						if s, ok := oldFieldSettings["properties"]; ok {
							if ns, ok := newSettings["properties"]; ok {
								if isForceable(s.(map[string]interface{}), ns.(map[string]interface{})) {
									return true
								}
								continue
							} else {
								return true
							}
						}
					} else {
						// if the key not found in the new props, force new resource
						return true
					}
				}
				return false
			}

			// if old defined we must check if the type of the existing fields were changed
			if oldProps, ok := o["properties"]; ok {
				newProps, ok := n["properties"]
				// if the old has props but new one not, immediately force new resource
				if !ok {
					return true
				}
				return isForceable(oldProps.(map[string]interface{}), newProps.(map[string]interface{}))
			}

			// if all check passed, we can update the map
			return false
		}),

		Schema: indexSchema,
	}
}

func resourceIndexCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := clients.NewApiClient(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	indexName := d.Get("name").(string)
	id, diags := client.ID(indexName)
	if diags.HasError() {
		return diags
	}
	var index models.Index
	index.Name = indexName

	if v, ok := d.GetOk("alias"); ok {
		aliases := v.(*schema.Set)
		als, diags := ExpandIndexAliases(aliases)
		if diags.HasError() {
			return diags
		}
		index.Aliases = als
	}

	if v, ok := d.GetOk("mappings"); ok {
		maps := make(map[string]interface{})
		if v.(string) != "" {
			if err := json.Unmarshal([]byte(v.(string)), &maps); err != nil {
				return diag.FromErr(err)
			}
		}
		index.Mappings = maps
	}

	if v, ok := d.GetOk("settings"); ok {
		// we know at this point we have 1 and only 1 `settings` block defined
		managed_settings := v.([]interface{})[0].(map[string]interface{})["setting"].(*schema.Set)
		sets := make(map[string]interface{}, managed_settings.Len())
		for _, s := range managed_settings.List() {
			setting := s.(map[string]interface{})
			sets[setting["name"].(string)] = setting["value"]
		}
		index.Settings = sets
	}

	if diags := client.PutElasticsearchIndex(&index); diags.HasError() {
		return diags
	}

	d.SetId(id.String())
	return resourceIndexRead(ctx, d, meta)
}

// Because of limitation of ES API we must handle changes to aliases, mappings and settings separately
func resourceIndexUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := clients.NewApiClient(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	indexName := d.Get("name").(string)

	// aliases
	if d.HasChange("alias") {
		oldAliases, newAliases := d.GetChange("alias")
		eold, diags := ExpandIndexAliases(oldAliases.(*schema.Set))
		if diags.HasError() {
			return diags
		}
		enew, diags := ExpandIndexAliases(newAliases.(*schema.Set))
		if diags.HasError() {
			return diags
		}

		aliasesToDelete := make([]string, 0)
		// iterate old aliases and decide which aliases to be deleted
		for k := range eold {
			if _, ok := enew[k]; !ok {
				// delete the alias
				aliasesToDelete = append(aliasesToDelete, k)
			}
		}
		if len(aliasesToDelete) > 0 {
			if diags := client.DeleteElasticsearchIndexAlias(indexName, aliasesToDelete); diags.HasError() {
				return diags
			}
		}

		// keep new aliases up-to-date
		for _, v := range enew {
			if diags := client.UpdateElasticsearchIndexAlias(indexName, &v); diags.HasError() {
				return diags
			}
		}
	}

	// settings
	if d.HasChange("settings") {
		oldSettings, newSettings := d.GetChange("settings")
		os := flattenIndexSettings(oldSettings.([]interface{}))
		ns := flattenIndexSettings(newSettings.([]interface{}))
		log.Printf("[TRACE] Change in the settings detected old settings = %+v, new  settings = %+v", os, ns)
		// make sure to add setting to the new map which were removed
		for k, ov := range os {
			if _, ok := ns[k]; !ok {
				ns[k] = nil
			}
			// remove the keys if the new value matches old one
			// we need to update only changed settings
			if nv, ok := ns[k]; ok && nv == ov {
				delete(ns, k)
			}
		}
		log.Printf("[TRACE] settings to update: %+v", ns)
		if diags := client.UpdateElasticsearchIndexSettings(indexName, ns); diags.HasError() {
			return diags
		}
	}

	// mappings
	if d.HasChange("mappings") {
		// at this point we know there are mappings defined and there is a change which we can apply
		mappings := d.Get("mappings").(string)
		if diags := client.UpdateElasticsearchIndexMappings(indexName, mappings); diags.HasError() {
			return diags
		}
	}

	return resourceIndexRead(ctx, d, meta)
}

func flattenIndexSettings(settings []interface{}) map[string]interface{} {
	ns := make(map[string]interface{})
	if len(settings) > 0 {
		s := settings[0].(map[string]interface{})["setting"].(*schema.Set)
		for _, v := range s.List() {
			vv := v.(map[string]interface{})
			ns[vv["name"].(string)] = vv["value"]
		}
	}
	return ns
}

func resourceIndexRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client, err := clients.NewApiClient(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	compId, diags := clients.CompositeIdFromStr(d.Id())
	if diags.HasError() {
		return diags
	}
	indexName := compId.ResourceId

	if err := d.Set("name", indexName); err != nil {
		return diag.FromErr(err)
	}

	index, diags := client.GetElasticsearchIndex(indexName)
	if index == nil && diags == nil {
		// no index found on ES side
		d.SetId("")
		return diags
	}
	if diags.HasError() {
		return diags
	}
	log.Printf("[TRACE] read the index data: %+v", index)

	if index.Aliases != nil {
		aliases, diags := FlattenIndexAliases(index.Aliases)
		if diags.HasError() {
			return diags
		}
		if err := d.Set("alias", aliases); err != nil {
			diag.FromErr(err)
		}
	}
	if index.Mappings != nil {
		m, err := json.Marshal(index.Mappings)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("mappings", string(m)); err != nil {
			return diag.FromErr(err)
		}
	}
	if index.Settings != nil {
		s, err := json.Marshal(index.Settings)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("settings_raw", string(s)); err != nil {
			return diag.FromErr(err)
		}
	}
	return diags
}

func resourceIndexDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	if diags := client.DeleteElasticsearchIndex(compId.ResourceId); diags.HasError() {
		return diags
	}
	d.SetId("")
	return diags
}
