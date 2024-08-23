package indices

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Schema defines the schema for the data source.
func (d *dataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages Elasticsearch indices",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Generated ID for the indices.",
				Computed:    true,
			},
			"search": schema.StringAttribute{
				Description: "Comma-separated list of indices to resolve by their name. Supports wildcards `*`.",
				Optional:    true,
			},
			"indices": schema.ListNestedAttribute{
				Description: "The list of indices.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Internal identifier of the resource.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the index.",
							Required:    true,
						},
						"number_of_shards": schema.Int32Attribute{
							Description: "Number of shards for the index.",
							Optional:    true,
						},
						"number_of_routing_shards": schema.Int32Attribute{
							Description: "Value used with number_of_shards to route documents to a primary shard.",
							Optional:    true,
						},
						"codec": schema.StringAttribute{
							Description: "The `default` value compresses stored data with LZ4 compression, but this can be set to `best_compression` which uses DEFLATE for a higher compression ratio.",
							Optional:    true,
						},
						"routing_partition_size": schema.Int32Attribute{
							Description: "The number of shards a custom routing value can go to.",
							Optional:    true,
						},
						"load_fixed_bitset_filters_eagerly": schema.BoolAttribute{
							Description: "Indicates whether cached filters are pre-loaded for nested queries.",
							Optional:    true,
						},
						"shard_check_on_startup": schema.StringAttribute{
							Description: "Whether or not shards should be checked for corruption before opening. When corruption is detected, it will prevent the shard from being opened. Accepts `false`, `true`, `checksum`.",
							Optional:    true,
						},
						"sort_field": schema.SetAttribute{
							ElementType: types.StringType,
							Description: "The field to sort shards in this index by.",
							Optional:    true,
						},
						"sort_order": schema.ListAttribute{
							ElementType: types.StringType,
							Description: "The direction to sort shards in. Accepts `asc`, `desc`.",
							Optional:    true,
						},
						"mapping_coerce": schema.BoolAttribute{
							Description: "The index level coercion setting that is applied to all mapping types.",
							Optional:    true,
						},
						"number_of_replicas": schema.Int32Attribute{
							Description: "Number of shard replicas.",
							Computed:    true,
						},
						"auto_expand_replicas": schema.StringAttribute{
							Description: "The number of replicas to the node count in the cluster. Set to a dash delimited lower and upper bound (e.g. 0-5) or use all for the upper bound (e.g. 0-all)",
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
						"max_result_window": schema.Int32Attribute{
							Description: "The maximum value of `from + size` for searches to this index.",
							Optional:    true,
						},
						"max_inner_result_window": schema.Int32Attribute{
							Description: "The maximum value of `from + size` for inner hits definition and top hits aggregations to this index.",
							Optional:    true,
						},
						"max_rescore_window": schema.Int32Attribute{
							Description: "The maximum value of `window_size` for `rescore` requests in searches of this index.",
							Optional:    true,
						},
						"max_docvalue_fields_search": schema.Int32Attribute{
							Description: "The maximum number of `docvalue_fields` that are allowed in a query.",
							Optional:    true,
						},
						"max_script_fields": schema.Int32Attribute{
							Description: "The maximum number of `script_fields` that are allowed in a query.",
							Optional:    true,
						},
						"max_ngram_diff": schema.Int32Attribute{
							Description: "The maximum allowed difference between min_gram and max_gram for NGramTokenizer and NGramTokenFilter.",
							Optional:    true,
						},
						"max_shingle_diff": schema.Int32Attribute{
							Description: "The maximum allowed difference between max_shingle_size and min_shingle_size for ShingleTokenFilter.",
							Optional:    true,
						},
						"max_refresh_listeners": schema.Int32Attribute{
							Description: "Maximum number of refresh listeners available on each shard of the index.",
							Optional:    true,
						},
						"analyze_max_token_count": schema.Int32Attribute{
							Description: "The maximum number of tokens that can be produced using _analyze API.",
							Optional:    true,
						},
						"highlight_max_analyzed_offset": schema.Int32Attribute{
							Description: "The maximum number of characters that will be analyzed for a highlight request.",
							Optional:    true,
						},
						"max_terms_count": schema.Int32Attribute{
							Description: "The maximum number of terms that can be used in Terms Query.",
							Optional:    true,
						},
						"max_regex_length": schema.Int32Attribute{
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
						},
						"routing_rebalance_enable": schema.StringAttribute{
							Description: "Enables shard rebalancing for this index. It can be set to: `all`, `primaries` , `replicas` , `none`.",
							Optional:    true,
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
						},
						"indexing_slowlog_source": schema.StringAttribute{
							Description: "Set the number of characters of the `_source` to include in the slowlog lines, `false` or `0` will skip logging the source entirely and setting it to `true` will log the entire source regardless of size. The original `_source` is reformatted by default to make sure that it fits on a single log line.",
							Optional:    true,
						},
						"analysis_analyzer": schema.StringAttribute{
							Description: "A JSON string describing the analyzers applied to the index.",
							Optional:    true,
						},
						"analysis_tokenizer": schema.StringAttribute{
							Description: "A JSON string describing the tokenizers applied to the index.",
							Optional:    true,
						},
						"analysis_char_filter": schema.StringAttribute{
							Description: "A JSON string describing the char_filters applied to the index.",
							Optional:    true,
						},
						"analysis_filter": schema.StringAttribute{
							Description: "A JSON string describing the filters applied to the index.",
							Optional:    true,
						},
						"analysis_normalizer": schema.StringAttribute{
							Description: "A JSON string describing the normalizers applied to the index.",
							Optional:    true,
						},
						"deletion_protection": schema.BoolAttribute{
							Description: "Whether to allow Terraform to destroy the index. Unless this field is set to false in Terraform state, a terraform destroy or terraform apply command that deletes the instance will fail.",
							Optional:    true,
						},
						"include_type_name": schema.BoolAttribute{
							Description: "If true, a mapping type is expected in the body of mappings. Defaults to false. Supported for Elasticsearch 7.x.",
							Optional:    true,
						},
						"wait_for_active_shards": schema.StringAttribute{
							Description: "The number of shard copies that must be active before proceeding with the operation. Set to `all` or any positive integer up to the total number of shards in the index (number_of_replicas+1). Default: `1`, the primary shard. This value is ignored when running against Serverless projects.",
							Optional:    true,
						},
						"master_timeout": schema.StringAttribute{
							Description: "Period to wait for a connection to the master node. If no response is received before the timeout expires, the request fails and returns an error. Defaults to `30s`. This value is ignored when running against Serverless projects.",
							Optional:    true,
						},
						"timeout": schema.StringAttribute{
							Description: "Period to wait for a response. If no response is received before the timeout expires, the request fails and returns an error. Defaults to `30s`.",
							Optional:    true,
						},
						"settings_raw": schema.StringAttribute{
							Description: "All raw settings fetched from the cluster.",
							Computed:    true,
						},
						"alias": schema.SetNestedAttribute{
							Description: "Aliases for the index.",
							Optional:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Description: "Index alias name.",
										Required:    true,
									},
									"filter": schema.StringAttribute{
										Description: "Query used to limit documents the alias can access.",
										Optional:    true,
									},
									"index_routing": schema.StringAttribute{
										Description: "Value used to route indexing operations to a specific shard. If specified, this overwrites the `routing` value for indexing operations.",
										Optional:    true,
									},
									"is_hidden": schema.BoolAttribute{
										Description: "If true, the alias is hidden.",
										Optional:    true,
									},
									"is_write_index": schema.BoolAttribute{
										Description: "If true, the index is the write index for the alias.",
										Optional:    true,
									},
									"routing": schema.StringAttribute{
										Description: "Value used to route indexing and search operations to a specific shard.",
										Optional:    true,
									},
									"search_routing": schema.StringAttribute{
										Description: "Value used to route search operations to a specific shard. If specified, this overwrites the routing value for search operations.",
										Optional:    true,
									},
								},
							},
						},
						"mappings": schema.StringAttribute{
							Description: `Mapping for fields in the index.
If specified, this mapping can include: field names, [field data types](https://www.elastic.co/guide/en/elasticsearch/reference/current/mapping-types.html), [mapping parameters](https://www.elastic.co/guide/en/elasticsearch/reference/current/mapping-params.html).
**NOTE:**
- Changing datatypes in the existing _mappings_ will force index to be re-created.
- Removing field will be ignored by default same as elasticsearch. You need to recreate the index to remove field completely.`,
							Optional: true,
						},
					},
				},
			},
		},
	}
}