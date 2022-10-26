package index

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var (
	staticSettingsKeys = map[string]schema.ValueType{
		"number_of_shards":                  schema.TypeInt,
		"number_of_routing_shards":          schema.TypeInt,
		"codec":                             schema.TypeString,
		"routing_partition_size":            schema.TypeInt,
		"load_fixed_bitset_filters_eagerly": schema.TypeBool,
		"shard.check_on_startup":            schema.TypeString,
		"sort.field":                        schema.TypeSet,
		"sort.order":                        schema.TypeSet,
	}
	dynamicsSettingsKeys = map[string]schema.ValueType{
		"number_of_replicas":                     schema.TypeInt,
		"auto_expand_replicas":                   schema.TypeString,
		"refresh_interval":                       schema.TypeString,
		"search.idle.after":                      schema.TypeString,
		"max_result_window":                      schema.TypeInt,
		"max_inner_result_window":                schema.TypeInt,
		"max_rescore_window":                     schema.TypeInt,
		"max_docvalue_fields_search":             schema.TypeInt,
		"max_script_fields":                      schema.TypeInt,
		"max_ngram_diff":                         schema.TypeInt,
		"max_shingle_diff":                       schema.TypeInt,
		"blocks.read_only":                       schema.TypeBool,
		"blocks.read_only_allow_delete":          schema.TypeBool,
		"blocks.read":                            schema.TypeBool,
		"blocks.write":                           schema.TypeBool,
		"blocks.metadata":                        schema.TypeBool,
		"max_refresh_listeners":                  schema.TypeInt,
		"analyze.max_token_count":                schema.TypeInt,
		"highlight.max_analyzed_offset":          schema.TypeInt,
		"max_terms_count":                        schema.TypeInt,
		"max_regex_length":                       schema.TypeInt,
		"query.default_field":                    schema.TypeSet,
		"routing.allocation.enable":              schema.TypeString,
		"routing.rebalance.enable":               schema.TypeString,
		"gc_deletes":                             schema.TypeString,
		"default_pipeline":                       schema.TypeString,
		"final_pipeline":                         schema.TypeString,
		"search.slowlog.threshold.query.warn":    schema.TypeString,
		"search.slowlog.threshold.query.info":    schema.TypeString,
		"search.slowlog.threshold.query.debug":   schema.TypeString,
		"search.slowlog.threshold.query.trace":   schema.TypeString,
		"search.slowlog.threshold.fetch.warn":    schema.TypeString,
		"search.slowlog.threshold.fetch.info":    schema.TypeString,
		"search.slowlog.threshold.fetch.debug":   schema.TypeString,
		"search.slowlog.threshold.fetch.trace":   schema.TypeString,
		"search.slowlog.level":                   schema.TypeString,
		"indexing.slowlog.threshold.index.warn":  schema.TypeString,
		"indexing.slowlog.threshold.index.info":  schema.TypeString,
		"indexing.slowlog.threshold.index.debug": schema.TypeString,
		"indexing.slowlog.threshold.index.trace": schema.TypeString,
		"indexing.slowlog.level":                 schema.TypeString,
		"indexing.slowlog.source":                schema.TypeString,
	}
	allSettingsKeys = map[string]schema.ValueType{}
)

func init() {
	for k, v := range staticSettingsKeys {
		allSettingsKeys[k] = v
	}
	for k, v := range dynamicsSettingsKeys {
		allSettingsKeys[k] = v
	}
}

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
		// Static settings that can only be set on creation
		"number_of_shards": {
			Type:        schema.TypeInt,
			Description: "Number of shards for the index. This can be set only on creation.",
			ForceNew:    true,
			Optional:    true,
		},
		"number_of_routing_shards": {
			Type:        schema.TypeInt,
			Description: "Value used with number_of_shards to route documents to a primary shard. This can be set only on creation.",
			ForceNew:    true,
			Optional:    true,
		},
		"codec": {
			Type:         schema.TypeString,
			Description:  "The `default` value compresses stored data with LZ4 compression, but this can be set to `best_compression` which uses DEFLATE for a higher compression ratio. This can be set only on creation.",
			ForceNew:     true,
			Optional:     true,
			ValidateFunc: validation.StringInSlice([]string{"best_compression"}, false),
		},
		"routing_partition_size": {
			Type:        schema.TypeInt,
			Description: "The number of shards a custom routing value can go to. This can be set only on creation.",
			ForceNew:    true,
			Optional:    true,
		},
		"load_fixed_bitset_filters_eagerly": {
			Type:        schema.TypeBool,
			Description: "Indicates whether cached filters are pre-loaded for nested queries. This can be set only on creation.",
			ForceNew:    true,
			Optional:    true,
		},
		"shard_check_on_startup": {
			Type:         schema.TypeString,
			Description:  "Whether or not shards should be checked for corruption before opening. When corruption is detected, it will prevent the shard from being opened. Accepts `false`, `true`, `checksum`.",
			ForceNew:     true,
			Optional:     true,
			ValidateFunc: validation.StringInSlice([]string{"false", "true", "checksum"}, false),
		},
		"sort_field": {
			Type:        schema.TypeSet,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "The field to sort shards in this index by.",
			ForceNew:    true,
			Optional:    true,
		},
		// sort_order can't be set type since it can have dup strings like ["asc", "asc"]
		"sort_order": {
			Type:        schema.TypeList,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "The direction to sort shards in. Accepts `asc`, `desc`.",
			ForceNew:    true,
			Optional:    true,
		},
		// Dynamic settings that can be changed at runtime
		"number_of_replicas": {
			Type:        schema.TypeInt,
			Description: "Number of shard replicas.",
			Optional:    true,
			Computed:    true,
		},
		"auto_expand_replicas": {
			Type:        schema.TypeString,
			Description: "Set the number of replicas to the node count in the cluster. Set to a dash delimited lower and upper bound (e.g. 0-5) or use all for the upper bound (e.g. 0-all)",
			Optional:    true,
		},
		"search_idle_after": {
			Type:        schema.TypeString,
			Description: "How long a shard can not receive a search or get request until it’s considered search idle.",
			Optional:    true,
		},
		"refresh_interval": {
			Type:        schema.TypeString,
			Description: "How often to perform a refresh operation, which makes recent changes to the index visible to search. Can be set to `-1` to disable refresh.",
			Optional:    true,
		},
		"max_result_window": {
			Type:        schema.TypeInt,
			Description: "The maximum value of `from + size` for searches to this index.",
			Optional:    true,
		},
		"max_inner_result_window": {
			Type:        schema.TypeInt,
			Description: "The maximum value of `from + size` for inner hits definition and top hits aggregations to this index.",
			Optional:    true,
		},
		"max_rescore_window": {
			Type:        schema.TypeInt,
			Description: "The maximum value of `window_size` for `rescore` requests in searches of this index.",
			Optional:    true,
		},
		"max_docvalue_fields_search": {
			Type:        schema.TypeInt,
			Description: "The maximum number of `docvalue_fields` that are allowed in a query.",
			Optional:    true,
		},
		"max_script_fields": {
			Type:        schema.TypeInt,
			Description: "The maximum number of `script_fields` that are allowed in a query.",
			Optional:    true,
		},
		"max_ngram_diff": {
			Type:        schema.TypeInt,
			Description: "The maximum allowed difference between min_gram and max_gram for NGramTokenizer and NGramTokenFilter.",
			Optional:    true,
		},
		"max_shingle_diff": {
			Type:        schema.TypeInt,
			Description: "The maximum allowed difference between max_shingle_size and min_shingle_size for ShingleTokenFilter.",
			Optional:    true,
		},
		"max_refresh_listeners": {
			Type:        schema.TypeInt,
			Description: "Maximum number of refresh listeners available on each shard of the index.",
			Optional:    true,
		},
		"analyze_max_token_count": {
			Type:        schema.TypeInt,
			Description: "The maximum number of tokens that can be produced using _analyze API.",
			Optional:    true,
		},
		"highlight_max_analyzed_offset": {
			Type:        schema.TypeInt,
			Description: "The maximum number of characters that will be analyzed for a highlight request.",
			Optional:    true,
		},
		"max_terms_count": {
			Type:        schema.TypeInt,
			Description: "The maximum number of terms that can be used in Terms Query.",
			Optional:    true,
		},
		"max_regex_length": {
			Type:        schema.TypeInt,
			Description: "The maximum length of regex that can be used in Regexp Query.",
			Optional:    true,
		},
		"query_default_field": {
			Type:        schema.TypeSet,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "Wildcard (*) patterns matching one or more fields. Defaults to '*', which matches all fields eligible for term-level queries, excluding metadata fields.",
			Optional:    true,
		},
		"routing_allocation_enable": {
			Type:         schema.TypeString,
			Description:  "Controls shard allocation for this index. It can be set to: `all` , `primaries` , `new_primaries` , `none`.",
			Optional:     true,
			ValidateFunc: validation.StringInSlice([]string{"all", "primaries", "new_primaries", "none"}, false),
		},
		"routing_rebalance_enable": {
			Type:         schema.TypeString,
			Description:  "Enables shard rebalancing for this index. It can be set to: `all`, `primaries` , `replicas` , `none`.",
			Optional:     true,
			ValidateFunc: validation.StringInSlice([]string{"all", "primaries", "replicas", "none"}, false),
		},
		"gc_deletes": {
			Type:        schema.TypeString,
			Description: "The length of time that a deleted document's version number remains available for further versioned operations.",
			Optional:    true,
		},
		"blocks_read_only": {
			Type:        schema.TypeBool,
			Description: "Set to `true` to make the index and index metadata read only, `false` to allow writes and metadata changes.",
			Optional:    true,
		},
		"blocks_read_only_allow_delete": {
			Type:        schema.TypeBool,
			Description: "Identical to `index.blocks.read_only` but allows deleting the index to free up resources.",
			Optional:    true,
		},
		"blocks_read": {
			Type:        schema.TypeBool,
			Description: "Set to `true` to disable read operations against the index.",
			Optional:    true,
		},
		"blocks_write": {
			Type:        schema.TypeBool,
			Description: "Set to `true` to disable data write operations against the index. This setting does not affect metadata.",
			Optional:    true,
		},
		"blocks_metadata": {
			Type:        schema.TypeBool,
			Description: "Set to `true` to disable index metadata reads and writes.",
			Optional:    true,
		},
		"default_pipeline": {
			Type:        schema.TypeString,
			Description: "The default ingest node pipeline for this index. Index requests will fail if the default pipeline is set and the pipeline does not exist.",
			Optional:    true,
		},
		"final_pipeline": {
			Type:        schema.TypeString,
			Description: "Final ingest pipeline for the index. Indexing requests will fail if the final pipeline is set and the pipeline does not exist. The final pipeline always runs after the request pipeline (if specified) and the default pipeline (if it exists). The special pipeline name _none indicates no ingest pipeline will run.",
			Optional:    true,
		},
		"search_slowlog_threshold_query_warn": {
			Type:        schema.TypeString,
			Description: "Set the cutoff for shard level slow search logging of slow searches in the query phase, in time units, e.g. `10s`",
			Optional:    true,
		},
		"search_slowlog_threshold_query_info": {
			Type:        schema.TypeString,
			Description: "Set the cutoff for shard level slow search logging of slow searches in the query phase, in time units, e.g. `5s`",
			Optional:    true,
		},
		"search_slowlog_threshold_query_debug": {
			Type:        schema.TypeString,
			Description: "Set the cutoff for shard level slow search logging of slow searches in the query phase, in time units, e.g. `2s`",
			Optional:    true,
		},
		"search_slowlog_threshold_query_trace": {
			Type:        schema.TypeString,
			Description: "Set the cutoff for shard level slow search logging of slow searches in the query phase, in time units, e.g. `500ms`",
			Optional:    true,
		},
		"search_slowlog_threshold_fetch_warn": {
			Type:        schema.TypeString,
			Description: "Set the cutoff for shard level slow search logging of slow searches in the fetch phase, in time units, e.g. `10s`",
			Optional:    true,
		},
		"search_slowlog_threshold_fetch_info": {
			Type:        schema.TypeString,
			Description: "Set the cutoff for shard level slow search logging of slow searches in the fetch phase, in time units, e.g. `5s`",
			Optional:    true,
		},
		"search_slowlog_threshold_fetch_debug": {
			Type:        schema.TypeString,
			Description: "Set the cutoff for shard level slow search logging of slow searches in the fetch phase, in time units, e.g. `2s`",
			Optional:    true,
		},
		"search_slowlog_threshold_fetch_trace": {
			Type:        schema.TypeString,
			Description: "Set the cutoff for shard level slow search logging of slow searches in the fetch phase, in time units, e.g. `500ms`",
			Optional:    true,
		},
		"search_slowlog_level": {
			Type:         schema.TypeString,
			Description:  "Set which logging level to use for the search slow log, can be: `warn`, `info`, `debug`, `trace`",
			Optional:     true,
			ValidateFunc: validation.StringInSlice([]string{"warn", "info", "debug", "trace"}, false),
		},
		"indexing_slowlog_threshold_index_warn": {
			Type:        schema.TypeString,
			Description: "Set the cutoff for shard level slow search logging of slow searches for indexing queries, in time units, e.g. `10s`",
			Optional:    true,
		},
		"indexing_slowlog_threshold_index_info": {
			Type:        schema.TypeString,
			Description: "Set the cutoff for shard level slow search logging of slow searches for indexing queries, in time units, e.g. `5s`",
			Optional:    true,
		},
		"indexing_slowlog_threshold_index_debug": {
			Type:        schema.TypeString,
			Description: "Set the cutoff for shard level slow search logging of slow searches for indexing queries, in time units, e.g. `2s`",
			Optional:    true,
		},
		"indexing_slowlog_threshold_index_trace": {
			Type:        schema.TypeString,
			Description: "Set the cutoff for shard level slow search logging of slow searches for indexing queries, in time units, e.g. `500ms`",
			Optional:    true,
		},
		"indexing_slowlog_level": {
			Type:         schema.TypeString,
			Description:  "Set which logging level to use for the search slow log, can be: `warn`, `info`, `debug`, `trace`",
			Optional:     true,
			ValidateFunc: validation.StringInSlice([]string{"warn", "info", "debug", "trace"}, false),
		},
		"indexing_slowlog_source": {
			Type:        schema.TypeString,
			Description: "Set the number of characters of the `_source` to include in the slowlog lines, `false` or `0` will skip logging the source entirely and setting it to `true` will log the entire source regardless of size. The original `_source` is reformatted by default to make sure that it fits on a single log line.",
			Optional:    true,
		},
		// To change analyzer setting, the index must be closed, updated, and then reopened but it can't be handled in terraform.
		// We raise error when they are tried to be updated instead of setting ForceNew not to have unexpected deletion.
		"analysis_analyzer": {
			Type:         schema.TypeString,
			Description:  "A JSON string describing the analyzers applied to the index.",
			Optional:     true,
			ValidateFunc: validation.StringIsJSON,
		},
		"analysis_tokenizer": {
			Type:         schema.TypeString,
			Description:  "A JSON string describing the tokenizers applied to the index.",
			Optional:     true,
			ValidateFunc: validation.StringIsJSON,
		},
		"analysis_char_filter": {
			Type:         schema.TypeString,
			Description:  "A JSON string describing the char_filters applied to the index.",
			Optional:     true,
			ValidateFunc: validation.StringIsJSON,
		},
		"analysis_filter": {
			Type:         schema.TypeString,
			Description:  "A JSON string describing the filters applied to the index.",
			Optional:     true,
			ValidateFunc: validation.StringIsJSON,
		},
		"analysis_normalizer": {
			Type:         schema.TypeString,
			Description:  "A JSON string describing the normalizers applied to the index.",
			Optional:     true,
			ValidateFunc: validation.StringIsJSON,
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
		// Deprecated: individual setting field should be used instead
		"settings": {
			Description: `DEPRECATED: Please use dedicated setting field. Configuration options for the index. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/index-modules.html#index-modules-settings.
**NOTE:** Static index settings (see: https://www.elastic.co/guide/en/elasticsearch/reference/current/index-modules.html#_static_index_settings) can be only set on the index creation and later cannot be removed or updated - _apply_ will return error`,
			Type:       schema.TypeList,
			MaxItems:   1,
			Optional:   true,
			Deprecated: "Using settings makes it easier to misconfigure.  Use dedicated field for the each setting instead.",
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
				// first populate what we can with Read
				diags := resourceIndexRead(ctx, d, m)
				if diags.HasError() {
					return nil, fmt.Errorf("unable to import requested index")
				}

				client, err := clients.NewApiClient(d, m)
				if err != nil {
					return nil, err
				}
				compId, diags := clients.CompositeIdFromStr(d.Id())
				if diags.HasError() {
					return nil, fmt.Errorf("failed to parse provided ID")
				}
				indexName := compId.ResourceId
				index, diags := client.GetElasticsearchIndex(ctx, indexName)
				if diags.HasError() {
					return nil, fmt.Errorf("failed to get an ES Index")
				}

				// check the settings and import those as well
				if index.Settings != nil {
					for key, typ := range allSettingsKeys {
						var value interface{}
						if v, ok := index.Settings[key]; ok {
							value = v
						} else if v, ok := index.Settings["index."+key]; ok {
							value = v
						} else {
							tflog.Warn(ctx, fmt.Sprintf("setting '%s' is not currently managed by terraform provider and has been ignored", key))
							continue
						}
						switch typ {
						case schema.TypeInt:
							v, err := strconv.Atoi(value.(string))
							if err != nil {
								return nil, fmt.Errorf("failed to convert setting '%s' value %v to int: %w", key, value, err)
							}
							value = v
						case schema.TypeBool:
							v, err := strconv.ParseBool(value.(string))
							if err != nil {
								return nil, fmt.Errorf("failed to convert setting '%s' value %v to bool: %w", key, value, err)
							}
							value = v
						}
						if err := d.Set(utils.ConvertSettingsKeyToTFFieldKey(key), value); err != nil {
							return nil, err
						}
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
			tflog.Trace(ctx, "mappings custom diff old = %+v new = %+v", o, n)

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
	id, diags := client.ID(ctx, indexName)
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

	index.Settings = map[string]interface{}{}
	if settings := utils.ExpandIndividuallyDefinedSettings(ctx, d, allSettingsKeys); len(settings) > 0 {
		index.Settings = settings
	}

	analysis := map[string]interface{}{}
	if analyzerJSON, ok := d.GetOk("analysis_analyzer"); ok {
		var analyzer map[string]interface{}
		bytes := []byte(analyzerJSON.(string))
		err = json.Unmarshal(bytes, &analyzer)
		if err != nil {
			return diag.FromErr(err)
		}
		analysis["analyzer"] = analyzer
	}
	if tokenizerJSON, ok := d.GetOk("analysis_tokenizer"); ok {
		var tokenizer map[string]interface{}
		bytes := []byte(tokenizerJSON.(string))
		err = json.Unmarshal(bytes, &tokenizer)
		if err != nil {
			return diag.FromErr(err)
		}
		analysis["tokenizer"] = tokenizer
	}
	if charFilterJSON, ok := d.GetOk("analysis_char_filter"); ok {
		var filter map[string]interface{}
		bytes := []byte(charFilterJSON.(string))
		if err = json.Unmarshal(bytes, &filter); err != nil {
			return diag.FromErr(err)
		}
		analysis["char_filter"] = filter
	}
	if filterJSON, ok := d.GetOk("analysis_filter"); ok {
		var filter map[string]interface{}
		bytes := []byte(filterJSON.(string))
		err = json.Unmarshal(bytes, &filter)
		if err != nil {
			return diag.FromErr(err)
		}
		analysis["filter"] = filter
	}
	if normalizerJSON, ok := d.GetOk("analysis_normalizer"); ok {
		var normalizer map[string]interface{}
		bytes := []byte(normalizerJSON.(string))
		err = json.Unmarshal(bytes, &normalizer)
		if err != nil {
			return diag.FromErr(err)
		}
		analysis["normalizer"] = normalizer
	}
	if len(analysis) > 0 {
		index.Settings["analysis"] = analysis
	}

	if v, ok := d.GetOk("settings"); ok {
		// we know at this point we have 1 and only 1 `settings` block defined
		managedSettings := v.([]interface{})[0].(map[string]interface{})["setting"].(*schema.Set)
		for _, s := range managedSettings.List() {
			setting := s.(map[string]interface{})
			name := setting["name"].(string)
			if _, ok := index.Settings[name]; ok {
				return diag.FromErr(fmt.Errorf("setting '%s' is already defined by the other field, please remove it from `settings` to avoid unexpected settings", name))
			}
			index.Settings[name] = setting["value"]
		}
	}

	if diags := client.PutElasticsearchIndex(ctx, &index); diags.HasError() {
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
			if diags := client.DeleteElasticsearchIndexAlias(ctx, indexName, aliasesToDelete); diags.HasError() {
				return diags
			}
		}

		// keep new aliases up-to-date
		for _, v := range enew {
			if diags := client.UpdateElasticsearchIndexAlias(ctx, indexName, &v); diags.HasError() {
				return diags
			}
		}
	}

	// settings
	updatedSettings := make(map[string]interface{})
	for key := range dynamicsSettingsKeys {
		fieldKey := utils.ConvertSettingsKeyToTFFieldKey(key)
		if d.HasChange(fieldKey) {
			updatedSettings[key] = d.Get(fieldKey)
		}
	}
	if d.HasChange("settings") {
		oldSettings, newSettings := d.GetChange("settings")
		os := flattenIndexSettings(oldSettings.([]interface{}))
		ns := flattenIndexSettings(newSettings.([]interface{}))
		tflog.Trace(ctx, fmt.Sprintf("Change in the settings detected old settings = %+v, new  settings = %+v", os, ns))
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
		for k, v := range ns {
			if _, ok := updatedSettings[k]; ok && v != nil {
				return diag.FromErr(fmt.Errorf("setting '%s' is already updated by the other field, please remove it from `settings` to avoid unexpected settings", k))
			} else {
				updatedSettings[k] = v
			}
		}
	}
	if len(updatedSettings) > 0 {
		tflog.Trace(ctx, fmt.Sprintf("settings to update: %+v", updatedSettings))
		if diags := client.UpdateElasticsearchIndexSettings(ctx, indexName, updatedSettings); diags.HasError() {
			return diags
		}
	}

	// mappings
	if d.HasChange("mappings") {
		// at this point we know there are mappings defined and there is a change which we can apply
		mappings := d.Get("mappings").(string)
		if diags := client.UpdateElasticsearchIndexMappings(ctx, indexName, mappings); diags.HasError() {
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

	index, diags := client.GetElasticsearchIndex(ctx, indexName)
	if index == nil && diags == nil {
		// no index found on ES side
		d.SetId("")
		return diags
	}
	if diags.HasError() {
		return diags
	}

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
	// TODO: We ideally should set read settings to each field to detect changes
	// But for now, setting it will cause unexpected diff for the existing clients which use `settings`
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
	if diags := client.DeleteElasticsearchIndex(ctx, compId.ResourceId); diags.HasError() {
		return diags
	}
	d.SetId("")
	return diags
}
