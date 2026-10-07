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

package index

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Elasticsearch index settings key constants shared across the index resource
// and indices data source packages.
const (
	SettingCodec                            = "codec"
	SettingLoadFixedBitsetFiltersEagerly    = "load_fixed_bitset_filters_eagerly"
	SettingMappingCoerce                    = "mapping.coerce"
	SettingNumberOfReplicas                 = "number_of_replicas"
	SettingAutoExpandReplicas               = "auto_expand_replicas"
	SettingMaxResultWindow                  = "max_result_window"
	SettingMaxInnerResultWindow             = "max_inner_result_window"
	SettingMaxRescoreWindow                 = "max_rescore_window"
	SettingMaxDocvalueFieldsSearch          = "max_docvalue_fields_search"
	SettingMaxScriptFields                  = "max_script_fields"
	SettingMaxNgramDiff                     = "max_ngram_diff"
	SettingMaxShingleDiff                   = "max_shingle_diff"
	SettingMaxRefreshListeners              = "max_refresh_listeners"
	SettingMaxTermsCount                    = "max_terms_count"
	SettingMaxRegexLength                   = "max_regex_length"
	SettingGCDeletes                        = "gc_deletes"
	SettingDefaultPipeline                  = "default_pipeline"
	SettingFinalPipeline                    = "final_pipeline"
	SettingNumberOfShards                   = "number_of_shards"
	SettingNumberOfRoutingShards            = "number_of_routing_shards"
	SettingRoutingPartitionSize             = "routing_partition_size"
	SettingShardCheckOnStartup              = "shard.check_on_startup"
	SettingSortField                        = "sort.field"
	SettingSortOrder                        = "sort.order"
	SettingSortMissing                      = "sort.missing"
	SettingSortMode                         = "sort.mode"
	SettingRefreshInterval                  = "refresh_interval"
	SettingSearchIdleAfter                  = "search.idle.after"
	SettingQueryDefaultField                = "query.default_field"
	SettingRoutingAllocationEnable          = "routing.allocation.enable"
	SettingRoutingRebalanceEnable           = "routing.rebalance.enable"
	SettingUnassignedNodeLeftDelayedTimeout = "unassigned.node_left.delayed_timeout"
)

// StaticSettingsKeys are index settings that can only be set at creation time.
var StaticSettingsKeys = []string{
	SettingNumberOfShards,
	SettingNumberOfRoutingShards,
	SettingCodec,
	SettingRoutingPartitionSize,
	SettingLoadFixedBitsetFiltersEagerly,
	SettingShardCheckOnStartup,
	SettingSortField,
	SettingSortOrder,
	SettingSortMissing,
	SettingSortMode,
	SettingMappingCoerce,
}

// DynamicSettingsKeys are index settings that can be changed at runtime.
var DynamicSettingsKeys = []string{
	SettingNumberOfReplicas,
	SettingAutoExpandReplicas,
	SettingRefreshInterval,
	SettingSearchIdleAfter,
	"mapping.total_fields.limit",
	SettingMaxResultWindow,
	SettingMaxInnerResultWindow,
	SettingMaxRescoreWindow,
	SettingMaxDocvalueFieldsSearch,
	SettingMaxScriptFields,
	SettingMaxNgramDiff,
	SettingMaxShingleDiff,
	"blocks.read_only",
	"blocks.read_only_allow_delete",
	"blocks.read",
	"blocks.write",
	"blocks.metadata",
	SettingMaxRefreshListeners,
	"analyze.max_token_count",
	"highlight.max_analyzed_offset",
	SettingMaxTermsCount,
	SettingMaxRegexLength,
	SettingQueryDefaultField,
	SettingRoutingAllocationEnable,
	SettingRoutingRebalanceEnable,
	SettingGCDeletes,
	SettingDefaultPipeline,
	SettingFinalPipeline,
	SettingUnassignedNodeLeftDelayedTimeout,
	"search.slowlog.threshold.query.warn",
	"search.slowlog.threshold.query.info",
	"search.slowlog.threshold.query.debug",
	"search.slowlog.threshold.query.trace",
	"search.slowlog.threshold.fetch.warn",
	"search.slowlog.threshold.fetch.info",
	"search.slowlog.threshold.fetch.debug",
	"search.slowlog.threshold.fetch.trace",
	"search.slowlog.level",
	"indexing.slowlog.threshold.index.warn",
	"indexing.slowlog.threshold.index.info",
	"indexing.slowlog.threshold.index.debug",
	"indexing.slowlog.threshold.index.trace",
	"indexing.slowlog.level",
	"indexing.slowlog.source",
}

// AllSettingsKeys is the concatenation of StaticSettingsKeys and DynamicSettingsKeys.
var AllSettingsKeys = func() []string {
	all := make([]string, 0, len(StaticSettingsKeys)+len(DynamicSettingsKeys))
	all = append(all, StaticSettingsKeys...)
	all = append(all, DynamicSettingsKeys...)
	return all
}()

// dynamicSettingsDescriptions mirrors the descriptions previously hand-declared
// in the index resource schema; keys are the Terraform attribute names derived
// from the Elasticsearch settings keys.
const (
	finalPipelineDescription = "Final ingest pipeline for the index. Indexing requests will fail if the final pipeline is set and the pipeline does not exist. " +
		"The final pipeline always runs after the request pipeline (if specified) and the default pipeline (if it exists). " +
		"The special pipeline name `_none` indicates no ingest pipeline will run.\n"
	indexingSlowlogSourceDescription = "Set the number of characters of the `_source` to include in the slowlog lines. " +
		"`false` or `0` skips logging the source entirely; `true` logs the entire source regardless of size. " +
		"The original `_source` is reformatted by default to make sure that it fits on a single log line.\n"
)

// GetDynamicSettingAttributes returns the shared schema attributes for every
// DynamicSettingsKeys entry, keyed by the Terraform attribute name (dotted
// settings keys flattened to underscores). Consumed by the index resource and
// the index settings resource so the two schemas cannot drift apart.
func GetDynamicSettingAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		SettingNumberOfReplicas: schema.Int64Attribute{
			Description: "Number of shard replicas.",
			Optional:    true,
		},
		SettingAutoExpandReplicas: schema.StringAttribute{
			Description: "Set the number of replicas to the node count in the cluster. Set to a dash delimited lower and upper bound (e.g. 0-5) or use all for the upper bound (e.g. 0-all)",
			Optional:    true,
		},
		SettingRefreshInterval: schema.StringAttribute{
			Description: "How often to perform a refresh operation, which makes recent changes to the index visible to search. Can be set to `-1` to disable refresh.",
			Optional:    true,
		},
		"search_idle_after": schema.StringAttribute{
			Description: "How long a shard can not receive a search or get request until it’s considered search idle.",
			Optional:    true,
		},
		"mapping_total_fields_limit": schema.Int64Attribute{
			Description: "The maximum number of fields in an index. Field type parameters count towards this limit. The default value is 1000.",
			Optional:    true,
		},
		SettingMaxResultWindow: schema.Int64Attribute{
			Description: "The maximum value of `from + size` for searches to this index.",
			Optional:    true,
		},
		SettingMaxInnerResultWindow: schema.Int64Attribute{
			Description: "The maximum value of `from + size` for inner hits definition and top hits aggregations to this index.",
			Optional:    true,
		},
		SettingMaxRescoreWindow: schema.Int64Attribute{
			Description: "The maximum value of `window_size` for `rescore` requests in searches of this index.",
			Optional:    true,
		},
		SettingMaxDocvalueFieldsSearch: schema.Int64Attribute{
			Description: "The maximum number of `docvalue_fields` that are allowed in a query.",
			Optional:    true,
		},
		SettingMaxScriptFields: schema.Int64Attribute{
			Description: "The maximum number of `script_fields` that are allowed in a query.",
			Optional:    true,
		},
		SettingMaxNgramDiff: schema.Int64Attribute{
			Description: "The maximum allowed difference between min_gram and max_gram for NGramTokenizer and NGramTokenFilter.",
			Optional:    true,
		},
		SettingMaxShingleDiff: schema.Int64Attribute{
			Description: "The maximum allowed difference between max_shingle_size and min_shingle_size for ShingleTokenFilter.",
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
		SettingMaxRefreshListeners: schema.Int64Attribute{
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
		SettingMaxTermsCount: schema.Int64Attribute{
			Description: "The maximum number of terms that can be used in Terms Query.",
			Optional:    true,
		},
		SettingMaxRegexLength: schema.Int64Attribute{
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
		SettingGCDeletes: schema.StringAttribute{
			Description: "The length of time that a deleted document's version number remains available for further versioned operations.",
			Optional:    true,
		},
		SettingDefaultPipeline: schema.StringAttribute{
			Description: "The default ingest node pipeline for this index. Index requests will fail if the default pipeline is set and the pipeline does not exist.",
			Optional:    true,
		},
		SettingFinalPipeline: schema.StringAttribute{
			Description: finalPipelineDescription,
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
			Description: indexingSlowlogSourceDescription,
			Optional:    true,
		},
	}
}
