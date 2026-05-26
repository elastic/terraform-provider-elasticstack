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

// Elasticsearch index setting keys used by the resource schema, models, and
// API translation layer. Centralised here so the same string literal does not
// drift between schema attribute names, model field tags, and API mapping
// helpers.
const (
	settingCodec                            = "codec"
	settingLoadFixedBitsetFiltersEagerly    = "load_fixed_bitset_filters_eagerly"
	settingMappingCoerce                    = "mapping.coerce"
	settingNumberOfReplicas                 = "number_of_replicas"
	settingAutoExpandReplicas               = "auto_expand_replicas"
	settingMaxResultWindow                  = "max_result_window"
	settingMaxInnerResultWindow             = "max_inner_result_window"
	settingMaxRescoreWindow                 = "max_rescore_window"
	settingMaxDocvalueFieldsSearch          = "max_docvalue_fields_search"
	settingMaxScriptFields                  = "max_script_fields"
	settingMaxNgramDiff                     = "max_ngram_diff"
	settingMaxShingleDiff                   = "max_shingle_diff"
	settingMaxRefreshListeners              = "max_refresh_listeners"
	settingMaxTermsCount                    = "max_terms_count"
	settingMaxRegexLength                   = "max_regex_length"
	settingGCDeletes                        = "gc_deletes"
	settingDefaultPipeline                  = "default_pipeline"
	settingFinalPipeline                    = "final_pipeline"
	settingNumberOfShards                   = "number_of_shards"
	settingNumberOfRoutingShards            = "number_of_routing_shards"
	settingRoutingPartitionSize             = "routing_partition_size"
	settingShardCheckOnStartup              = "shard.check_on_startup"
	settingSortField                        = "sort.field"
	settingSortOrder                        = "sort.order"
	settingSortMissing                      = "sort.missing"
	settingSortMode                         = "sort.mode"
	settingRefreshInterval                  = "refresh_interval"
	settingSearchIdleAfter                  = "search.idle.after"
	settingQueryDefaultField                = "query.default_field"
	settingRoutingAllocationEnable          = "routing.allocation.enable"
	settingRoutingRebalanceEnable           = "routing.rebalance.enable"
	settingUnassignedNodeLeftDelayedTimeout = "unassigned.node_left.delayed_timeout"
)

// Terraform schema attribute keys reused by the index resource and its sort
// support. These names match the Elasticsearch index sort and alias APIs.
const (
	attrName    = "name"
	attrSetting = "setting"
	attrField   = "field"
	attrOrder   = "order"
	attrMissing = "missing"
	attrMode    = "mode"
)

// Sort order and tie-breaker tokens used in index sort plan modifiers and
// flattened state.
const (
	sortOrderAsc    = "asc"
	sortOrderDesc   = "desc"
	sortMissingLast = "_last"
	sortModeMax     = "max"
	sortModeMin     = "min"
)
