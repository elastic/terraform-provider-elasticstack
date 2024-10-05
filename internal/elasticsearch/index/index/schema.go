package index

import (
	"context"
	"regexp"

	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/planmodifiers"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = getSchema()
}

func getSchema() schema.Schema {
	return schema.Schema{
		Description: "Creates Elasticsearch indices. See: https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-create-index.html",
		Blocks: map[string]schema.Block{
			"elasticsearch_connection": providerschema.GetEsFWConnectionBlock("elasticsearch_connection", false),
			"alias": schema.SetNestedBlock{
				Description: "Aliases for the index.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Index alias name.",
							Required:    true,
						},
						"filter": schema.StringAttribute{
							Description: "Query used to limit documents the alias can access.",
							Optional:    true,
							CustomType:  jsontypes.NormalizedType{},
						},
						"index_routing": schema.StringAttribute{
							Description: "Value used to route indexing operations to a specific shard. If specified, this overwrites the `routing` value for indexing operations.",
							Optional:    true,
							Computed:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
								planmodifiers.StringUseDefaultIfUnknown(""),
							},
						},
						"is_hidden": schema.BoolAttribute{
							Description: "If true, the alias is hidden.",
							Optional:    true,
							Computed:    true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
								planmodifiers.BoolUseDefaultIfUnknown(false),
							},
						},
						"is_write_index": schema.BoolAttribute{
							Description: "If true, the index is the write index for the alias.",
							Optional:    true,
							Computed:    true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
								planmodifiers.BoolUseDefaultIfUnknown(false),
							},
						},
						"routing": schema.StringAttribute{
							Description: "Value used to route indexing and search operations to a specific shard.",
							Optional:    true,
							Computed:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
								planmodifiers.StringUseDefaultIfUnknown(""),
							},
						},
						"search_routing": schema.StringAttribute{
							Description: "Value used to route search operations to a specific shard. If specified, this overwrites the routing value for search operations.",
							Optional:    true,
							Computed:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
								planmodifiers.StringUseDefaultIfUnknown(""),
							},
						},
					},
				},
			},
			"settings": schema.ListNestedBlock{
				Description: `DEPRECATED: Please use dedicated setting field. Configuration options for the index. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/index-modules.html#index-modules-settings.
**NOTE:** Static index settings (see: https://www.elastic.co/guide/en/elasticsearch/reference/current/index-modules.html#_static_index_settings) can be only set on the index creation and later cannot be removed or updated - _apply_ will return error`,
				DeprecationMessage: "Using settings makes it easier to misconfigure.  Use dedicated field for the each setting instead.",
				Validators: []validator.List{
					listvalidator.SizeBetween(1, 1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"setting": schema.SetNestedBlock{
							Description: "Defines the setting for the index.",
							Validators: []validator.Set{
								setvalidator.SizeAtLeast(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Description: "The name of the setting to set and track.",
										Required:    true,
									},
									"value": schema.StringAttribute{
										Description: "The value of the setting to set and track.",
										Required:    true,
									},
								},
							},
						},
					},
				},
			},
		},
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Internal identifier of the resource",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the index you wish to create.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
					stringvalidator.NoneOf(".", ".."),
					stringvalidator.RegexMatches(regexp.MustCompile(`^[^-_+]`), "cannot start with -, _, +"),
					stringvalidator.RegexMatches(regexp.MustCompile(`^[a-z0-9!$%&'()+.;=@[\]^{}~_-]+$`), "must contain lower case alphanumeric characters and selected punctuation, see: https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-create-index.html#indices-create-api-path-params"),
				},
			},
			// Static settings that can only be set on creation
			"number_of_shards": schema.Int64Attribute{
				Description: "Number of shards for the index. This can be set only on creation.",
				Optional:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"number_of_routing_shards": schema.Int64Attribute{
				Description: "Value used with number_of_shards to route documents to a primary shard. This can be set only on creation.",
				Optional:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"codec": schema.StringAttribute{
				Description: "The `default` value compresses stored data with LZ4 compression, but this can be set to `best_compression` which uses DEFLATE for a higher compression ratio. This can be set only on creation.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("best_compression"),
				},
			},
			"routing_partition_size": schema.Int64Attribute{
				Description: "The number of shards a custom routing value can go to. This can be set only on creation.",
				Optional:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"load_fixed_bitset_filters_eagerly": schema.BoolAttribute{
				Description: "Indicates whether cached filters are pre-loaded for nested queries. This can be set only on creation.",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"shard_check_on_startup": schema.StringAttribute{
				Description: "Whether or not shards should be checked for corruption before opening. When corruption is detected, it will prevent the shard from being opened. Accepts `false`, `true`, `checksum`.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("false", "true", "checksum"),
				},
			},
			"sort_field": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "The field to sort shards in this index by.",
				Optional:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
			},
			// sort_order can't be set type since it can have dup strings like ["asc", "asc"]
			"sort_order": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "The direction to sort shards in. Accepts `asc`, `desc`.",
				Optional:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"mapping_coerce": schema.BoolAttribute{
				Description: "Set index level coercion setting that is applied to all mapping types.",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			// Dynamic settings that can be changed at runtime
			"number_of_replicas": schema.Int64Attribute{
				Description: "Number of shard replicas.",
				Optional:    true,
			},
			"auto_expand_replicas": schema.StringAttribute{
				Description: "Set the number of replicas to the node count in the cluster. Set to a dash delimited lower and upper bound (e.g. 0-5) or use all for the upper bound (e.g. 0-all)",
				Optional:    true,
			},
			"search_idle_after": schema.StringAttribute{
				Description: "How long a shard can not receive a search or get request until it’s considered search idle.",
				Optional:    true,
			},
			"refresh_interval": schema.StringAttribute{
				Description: "How often to perform a refresh operation, which makes recent changes to the index visible to search. Can be set to `-1` to disable refresh.",
				Optional:    true,
			},
			"max_result_window": schema.Int64Attribute{
				Description: "The maximum value of `from + size` for searches to this index.",
				Optional:    true,
			},
			"max_inner_result_window": schema.Int64Attribute{
				Description: "The maximum value of `from + size` for inner hits definition and top hits aggregations to this index.",
				Optional:    true,
			},
			"max_rescore_window": schema.Int64Attribute{
				Description: "The maximum value of `window_size` for `rescore` requests in searches of this index.",
				Optional:    true,
			},
			"max_docvalue_fields_search": schema.Int64Attribute{
				Description: "The maximum number of `docvalue_fields` that are allowed in a query.",
				Optional:    true,
			},
			"max_script_fields": schema.Int64Attribute{
				Description: "The maximum number of `script_fields` that are allowed in a query.",
				Optional:    true,
			},
			"max_ngram_diff": schema.Int64Attribute{
				Description: "The maximum allowed difference between min_gram and max_gram for NGramTokenizer and NGramTokenFilter.",
				Optional:    true,
			},
			"max_shingle_diff": schema.Int64Attribute{
				Description: "The maximum allowed difference between max_shingle_size and min_shingle_size for ShingleTokenFilter.",
				Optional:    true,
			},
			"max_refresh_listeners": schema.Int64Attribute{
				Description: "Maximum number of refresh listeners available on each shard of the index.",
				Optional:    true,
			},
			"analyze_max_token_count": schema.Int64Attribute{
				Description: "The maximum number of tokens that can be produced using _analyze API.",
				Optional:    true,
			},
			"highlight_max_analyzed_offset": schema.Int64Attribute{
				Description: "The maximum number of characters that will be analyzed for a highlight request.",
				Optional:    true,
			},
			"max_terms_count": schema.Int64Attribute{
				Description: "The maximum number of terms that can be used in Terms Query.",
				Optional:    true,
			},
			"max_regex_length": schema.Int64Attribute{
				Description: "The maximum length of regex that can be used in Regexp Query.",
				Optional:    true,
			},
			"query_default_field": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "Wildcard (*) patterns matching one or more fields. Defaults to '*', which matches all fields eligible for term-level queries, excluding metadata fields.",
				Optional:    true,
			},
			"routing_allocation_enable": schema.StringAttribute{
				Description: "Controls shard allocation for this index. It can be set to: `all` , `primaries` , `new_primaries` , `none`.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("all", "primaries", "new_primaries", "none"),
				},
			},
			"routing_rebalance_enable": schema.StringAttribute{
				Description: "Enables shard rebalancing for this index. It can be set to: `all`, `primaries` , `replicas` , `none`.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("all", "primaries", "replicas", "none"),
				},
			},
			"gc_deletes": schema.StringAttribute{
				Description: "The length of time that a deleted document's version number remains available for further versioned operations.",
				Optional:    true,
			},
			"blocks_read_only": schema.BoolAttribute{
				Description: "Set to `true` to make the index and index metadata read only, `false` to allow writes and metadata changes.",
				Optional:    true,
			},
			"blocks_read_only_allow_delete": schema.BoolAttribute{
				Description: "Identical to `index.blocks.read_only` but allows deleting the index to free up resources.",
				Optional:    true,
			},
			"blocks_read": schema.BoolAttribute{
				Description: "Set to `true` to disable read operations against the index.",
				Optional:    true,
			},
			"blocks_write": schema.BoolAttribute{
				Description: "Set to `true` to disable data write operations against the index. This setting does not affect metadata.",
				Optional:    true,
			},
			"blocks_metadata": schema.BoolAttribute{
				Description: "Set to `true` to disable index metadata reads and writes.",
				Optional:    true,
			},
			"default_pipeline": schema.StringAttribute{
				Description: "The default ingest node pipeline for this index. Index requests will fail if the default pipeline is set and the pipeline does not exist.",
				Optional:    true,
			},
			"final_pipeline": schema.StringAttribute{
				Description: "Final ingest pipeline for the index. Indexing requests will fail if the final pipeline is set and the pipeline does not exist. The final pipeline always runs after the request pipeline (if specified) and the default pipeline (if it exists). The special pipeline name _none indicates no ingest pipeline will run.",
				Optional:    true,
			},
			"unassigned_node_left_delayed_timeout": schema.StringAttribute{
				Description: "Time to delay the allocation of replica shards which become unassigned because a node has left, in time units, e.g. `10s`",
				Optional:    true,
			},
			"search_slowlog_threshold_query_warn": schema.StringAttribute{
				Description: "Set the cutoff for shard level slow search logging of slow searches in the query phase, in time units, e.g. `10s`",
				Optional:    true,
			},
			"search_slowlog_threshold_query_info": schema.StringAttribute{
				Description: "Set the cutoff for shard level slow search logging of slow searches in the query phase, in time units, e.g. `5s`",
				Optional:    true,
			},
			"search_slowlog_threshold_query_debug": schema.StringAttribute{
				Description: "Set the cutoff for shard level slow search logging of slow searches in the query phase, in time units, e.g. `2s`",
				Optional:    true,
			},
			"search_slowlog_threshold_query_trace": schema.StringAttribute{
				Description: "Set the cutoff for shard level slow search logging of slow searches in the query phase, in time units, e.g. `500ms`",
				Optional:    true,
			},
			"search_slowlog_threshold_fetch_warn": schema.StringAttribute{
				Description: "Set the cutoff for shard level slow search logging of slow searches in the fetch phase, in time units, e.g. `10s`",
				Optional:    true,
			},
			"search_slowlog_threshold_fetch_info": schema.StringAttribute{
				Description: "Set the cutoff for shard level slow search logging of slow searches in the fetch phase, in time units, e.g. `5s`",
				Optional:    true,
			},
			"search_slowlog_threshold_fetch_debug": schema.StringAttribute{
				Description: "Set the cutoff for shard level slow search logging of slow searches in the fetch phase, in time units, e.g. `2s`",
				Optional:    true,
			},
			"search_slowlog_threshold_fetch_trace": schema.StringAttribute{
				Description: "Set the cutoff for shard level slow search logging of slow searches in the fetch phase, in time units, e.g. `500ms`",
				Optional:    true,
			},
			"search_slowlog_level": schema.StringAttribute{
				Description: "Set which logging level to use for the search slow log, can be: `warn`, `info`, `debug`, `trace`",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("warn", "info", "debug", "trace"),
				},
			},
			"indexing_slowlog_threshold_index_warn": schema.StringAttribute{
				Description: "Set the cutoff for shard level slow search logging of slow searches for indexing queries, in time units, e.g. `10s`",
				Optional:    true,
			},
			"indexing_slowlog_threshold_index_info": schema.StringAttribute{
				Description: "Set the cutoff for shard level slow search logging of slow searches for indexing queries, in time units, e.g. `5s`",
				Optional:    true,
			},
			"indexing_slowlog_threshold_index_debug": schema.StringAttribute{
				Description: "Set the cutoff for shard level slow search logging of slow searches for indexing queries, in time units, e.g. `2s`",
				Optional:    true,
			},
			"indexing_slowlog_threshold_index_trace": schema.StringAttribute{
				Description: "Set the cutoff for shard level slow search logging of slow searches for indexing queries, in time units, e.g. `500ms`",
				Optional:    true,
			},
			"indexing_slowlog_level": schema.StringAttribute{
				Description: "Set which logging level to use for the search slow log, can be: `warn`, `info`, `debug`, `trace`",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("warn", "info", "debug", "trace"),
				},
			},
			"indexing_slowlog_source": schema.StringAttribute{
				Description: "Set the number of characters of the `_source` to include in the slowlog lines, `false` or `0` will skip logging the source entirely and setting it to `true` will log the entire source regardless of size. The original `_source` is reformatted by default to make sure that it fits on a single log line.",
				Optional:    true,
			},
			// To change analyzer setting, the index must be closed, updated, and then reopened but it can't be handled in terraform.
			// We raise error when they are tried to be updated instead of setting ForceNew not to have unexpected deletion.
			"analysis_analyzer": schema.StringAttribute{
				Description: "A JSON string describing the analyzers applied to the index.",
				Optional:    true,
				CustomType:  jsontypes.NormalizedType{},
				Validators: []validator.String{
					index.StringIsJSONObject{},
				},
			},
			"analysis_tokenizer": schema.StringAttribute{
				Description: "A JSON string describing the tokenizers applied to the index.",
				Optional:    true,
				CustomType:  jsontypes.NormalizedType{},
				Validators: []validator.String{
					index.StringIsJSONObject{},
				},
			},
			"analysis_char_filter": schema.StringAttribute{
				Description: "A JSON string describing the char_filters applied to the index.",
				Optional:    true,
				CustomType:  jsontypes.NormalizedType{},
				Validators: []validator.String{
					index.StringIsJSONObject{},
				},
			},
			"analysis_filter": schema.StringAttribute{
				Description: "A JSON string describing the filters applied to the index.",
				Optional:    true,
				CustomType:  jsontypes.NormalizedType{},
				Validators: []validator.String{
					index.StringIsJSONObject{},
				},
			},
			"analysis_normalizer": schema.StringAttribute{
				Description: "A JSON string describing the normalizers applied to the index.",
				Optional:    true,
				CustomType:  jsontypes.NormalizedType{},
				Validators: []validator.String{
					index.StringIsJSONObject{},
				},
			},
			"mappings": schema.StringAttribute{
				Description: `Mapping for fields in the index.
			If specified, this mapping can include: field names, [field data types](https://www.elastic.co/guide/en/elasticsearch/reference/current/mapping-types.html), [mapping parameters](https://www.elastic.co/guide/en/elasticsearch/reference/current/mapping-params.html).
			**NOTE:**
			- Changing datatypes in the existing _mappings_ will force index to be re-created.
			- Removing field will be ignored by default same as elasticsearch. You need to recreate the index to remove field completely.
			`,
				Optional:   true,
				Computed:   true,
				CustomType: jsontypes.NormalizedType{},
				Validators: []validator.String{
					index.StringIsJSONObject{},
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					mappingsPlanModifier{},
				},
			},
			"settings_raw": schema.StringAttribute{
				Description: "All raw settings fetched from the cluster.",
				Computed:    true,
				CustomType:  jsontypes.NormalizedType{},
				// TODO: Plan modifier. Use state if no other settings have been modified
			},
			"deletion_protection": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether to allow Terraform to destroy the index. Unless this field is set to false in Terraform state, a terraform destroy or terraform apply command that deletes the instance will fail.",
				PlanModifiers: []planmodifier.Bool{
					planmodifiers.BoolUseDefaultIfUnknown(true),
				},
			},
			"wait_for_active_shards": schema.StringAttribute{
				Description: "The number of shard copies that must be active before proceeding with the operation. Set to `all` or any positive integer up to the total number of shards in the index (number_of_replicas+1). Default: `1`, the primary shard. This value is ignored when running against Serverless projects.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					planmodifiers.StringUseDefaultIfUnknown("1"),
				},
			},
			"master_timeout": schema.StringAttribute{
				Description: "Period to wait for a connection to the master node. If no response is received before the timeout expires, the request fails and returns an error. Defaults to `30s`. This value is ignored when running against Serverless projects.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					planmodifiers.StringUseDefaultIfUnknown("30s"),
				},
				CustomType: customtypes.DurationType{},
			},
			"timeout": schema.StringAttribute{
				Description: "Period to wait for a response. If no response is received before the timeout expires, the request fails and returns an error. Defaults to `30s`.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					planmodifiers.StringUseDefaultIfUnknown("30s"),
				},
				CustomType: customtypes.DurationType{},
			},
		},
	}
}

func aliasElementType() attr.Type {
	return getSchema().Blocks["alias"].Type().(attr.TypeWithElementType).ElementType()
}

func settingsElementType() attr.Type {
	return getSchema().Blocks["settings"].Type().(attr.TypeWithElementType).ElementType()
}

func settingElementType() attr.Type {
	return getSchema().Blocks["settings"].GetNestedObject().GetBlocks()["setting"].Type().(attr.TypeWithElementType).ElementType()
}
