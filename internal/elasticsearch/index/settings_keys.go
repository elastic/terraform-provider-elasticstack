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
