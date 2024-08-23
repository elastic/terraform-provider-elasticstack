package indices

import "github.com/hashicorp/terraform-plugin-framework/types"

// dataSourceModel maps the data source schema data.
type dataSourceModel struct {
	ID      types.String `tfsdk:"id"`
	Search  types.String `tfsdk:"search"`
	Indices []indexModel `tfsdk:"indices"`
}

// indexModel maps indices schema data.
type indexModel struct {
	ID                                 types.String `tfsdk:"id"`
	Name                               types.String `tfsdk:"name"`
	NumberOfShards                     types.Int32  `tfsdk:"number_of_shards"`
	NumberOfRoutingShards              types.Int32  `tfsdk:"number_of_routing_shards"`
	Codec                              types.String `tfsdk:"codec"`
	RoutingPartitionSize               types.Int32  `tfsdk:"routing_partition_size"`
	LoadFixedBitsetFiltersEagerly      types.Bool   `tfsdk:"load_fixed_bitset_filters_eagerly"`
	ShardCheckOnStartup                types.String `tfsdk:"shard_check_on_startup"`
	SortField                          types.Set    `tfsdk:"sort_field"`
	SortOrder                          types.List   `tfsdk:"sort_order"`
	MappingCoerce                      types.Bool   `tfsdk:"mapping_coerce"`
	NumberOfReplicas                   types.Int32  `tfsdk:"number_of_replicas"`
	AutoExpandReplicas                 types.String `tfsdk:"auto_expand_replicas"`
	SearchIdleAfter                    types.String `tfsdk:"search_idle_after"`
	RefreshInterval                    types.String `tfsdk:"refresh_interval"`
	MaxResultWindow                    types.Int32  `tfsdk:"max_result_window"`
	MaxInnerResultWindow               types.Int32  `tfsdk:"max_inner_result_window"`
	MaxRescoreWindow                   types.Int32  `tfsdk:"max_rescore_window"`
	MaxDocvalueFieldsSearch            types.Int32  `tfsdk:"max_docvalue_fields_search"`
	MaxScriptFields                    types.Int32  `tfsdk:"max_script_fields"`
	MaxNgramDiff                       types.Int32  `tfsdk:"max_ngram_diff"`
	MaxShingleDiff                     types.Int32  `tfsdk:"max_shingle_diff"`
	MaxRefreshListeners                types.Int32  `tfsdk:"max_refresh_listeners"`
	AnalyzeMaxTokenCount               types.Int32  `tfsdk:"analyze_max_token_count"`
	HighlightMaxAnalyzedOffset         types.Int32  `tfsdk:"highlight_max_analyzed_offset"`
	MaxTermsCount                      types.Int32  `tfsdk:"max_terms_count"`
	MaxRegexLength                     types.Int32  `tfsdk:"max_regex_length"`
	QueryDefaultField                  types.Set    `tfsdk:"query_default_field"`
	RoutingAllocationEnable            types.String `tfsdk:"routing_allocation_enable"`
	RoutingRebalanceEnable             types.String `tfsdk:"routing_rebalance_enable"`
	GCDeletes                          types.String `tfsdk:"gc_deletes"`
	BlocksReadOnly                     types.Bool   `tfsdk:"blocks_read_only"`
	BlocksReadOnlyAllowDelete          types.Bool   `tfsdk:"blocks_read_only_allow_delete"`
	BlocksRead                         types.Bool   `tfsdk:"blocks_read"`
	BlocksWrite                        types.Bool   `tfsdk:"blocks_write"`
	BlocksMetadata                     types.Bool   `tfsdk:"blocks_metadata"`
	DefaultPipeline                    types.String `tfsdk:"default_pipeline"`
	FinalPipeline                      types.String `tfsdk:"final_pipeline"`
	UnassignedNodeLeftDelayedTimeout   types.String `tfsdk:"unassigned_node_left_delayed_timeout"`
	SearchSlowlogThresholdQueryWarn    types.String `tfsdk:"search_slowlog_threshold_query_warn"`
	SearchSlowlogThresholdQueryInfo    types.String `tfsdk:"search_slowlog_threshold_query_info"`
	SearchSlowlogThresholdQueryDebug   types.String `tfsdk:"search_slowlog_threshold_query_debug"`
	SearchSlowlogThresholdQueryTrace   types.String `tfsdk:"search_slowlog_threshold_query_trace"`
	SearchSlowlogThresholdFetchWarn    types.String `tfsdk:"search_slowlog_threshold_fetch_warn"`
	SearchSlowlogThresholdFetchInfo    types.String `tfsdk:"search_slowlog_threshold_fetch_info"`
	SearchSlowlogThresholdFetchDebug   types.String `tfsdk:"search_slowlog_threshold_fetch_debug"`
	SearchSlowlogThresholdFetchTrace   types.String `tfsdk:"search_slowlog_threshold_fetch_trace"`
	SearchSlowlogLevel                 types.String `tfsdk:"search_slowlog_level"`
	IndexingSlowlogThresholdIndexWarn  types.String `tfsdk:"indexing_slowlog_threshold_index_warn"`
	IndexingSlowlogThresholdIndexInfo  types.String `tfsdk:"indexing_slowlog_threshold_index_info"`
	IndexingSlowlogThresholdIndexDebug types.String `tfsdk:"indexing_slowlog_threshold_index_debug"`
	IndexingSlowlogThresholdIndexTrace types.String `tfsdk:"indexing_slowlog_threshold_index_trace"`
	IndexingSlowlogLevel               types.String `tfsdk:"indexing_slowlog_level"`
	IndexingSlowlogSource              types.String `tfsdk:"indexing_slowlog_source"`
	AnalysisAnalyzer                   types.String `tfsdk:"analysis_analyzer"`
	AnalysisTokenizer                  types.String `tfsdk:"analysis_tokenizer"`
	AnalysisCharFilter                 types.String `tfsdk:"analysis_char_filter"`
	AnalysisFilter                     types.String `tfsdk:"analysis_filter"`
	AnalysisNormalizer                 types.String `tfsdk:"analysis_normalizer"`
	DeletionProtection                 types.Bool   `tfsdk:"deletion_protection"`
	IncludeTypeName                    types.Bool   `tfsdk:"include_type_name"`
	WaitForActiveShards                types.String `tfsdk:"wait_for_active_shards"`
	MasterTimeout                      types.String `tfsdk:"master_timeout"`
	Timeout                            types.String `tfsdk:"timeout"`
	SettingsRaw                        types.String `tfsdk:"settings_raw"`
	Alias                              []aliasModel `tfsdk:"alias"`
	Mappings                           types.String `tfsdk:"mappings"`
}

// aliasModel maps alias schema data.
type aliasModel struct {
	Name          types.String `tfsdk:"name"`
	Filter        types.String `tfsdk:"filter"`
	IndexRouting  types.String `tfsdk:"index_routing"`
	IsHidden      types.Bool   `tfsdk:"is_hidden"`
	IsWriteIndex  types.Bool   `tfsdk:"is_write_index"`
	Routing       types.String `tfsdk:"routing"`
	SearchRouting types.String `tfsdk:"search_routing"`
}
