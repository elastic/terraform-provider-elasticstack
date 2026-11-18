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

// Package dynamicsettings holds the shared Terraform model fields for the
// Elasticsearch index dynamic settings. It is anonymously embedded in the
// models of resources that declare indexparent.GetDynamicSettingAttributes()
// so their schemas and models cannot drift apart.
package dynamicsettings

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Model carries one tfsdk-tagged field per indexparent.DynamicSettingsKeys
// entry. Each tag matches the corresponding attribute name in
// indexparent.GetDynamicSettingAttributes().
type Model struct {
	NumberOfReplicas                   types.Int64  `tfsdk:"number_of_replicas"`
	AutoExpandReplicas                 types.String `tfsdk:"auto_expand_replicas"`
	RefreshInterval                    types.String `tfsdk:"refresh_interval"`
	SearchIdleAfter                    types.String `tfsdk:"search_idle_after"`
	MappingTotalFieldsLimit            types.Int64  `tfsdk:"mapping_total_fields_limit"`
	MaxResultWindow                    types.Int64  `tfsdk:"max_result_window"`
	MaxInnerResultWindow               types.Int64  `tfsdk:"max_inner_result_window"`
	MaxRescoreWindow                   types.Int64  `tfsdk:"max_rescore_window"`
	MaxDocvalueFieldsSearch            types.Int64  `tfsdk:"max_docvalue_fields_search"`
	MaxScriptFields                    types.Int64  `tfsdk:"max_script_fields"`
	MaxNgramDiff                       types.Int64  `tfsdk:"max_ngram_diff"`
	MaxShingleDiff                     types.Int64  `tfsdk:"max_shingle_diff"`
	MaxRefreshListeners                types.Int64  `tfsdk:"max_refresh_listeners"`
	AnalyzeMaxTokenCount               types.Int64  `tfsdk:"analyze_max_token_count"`
	HighlightMaxAnalyzedOffset         types.Int64  `tfsdk:"highlight_max_analyzed_offset"`
	MaxTermsCount                      types.Int64  `tfsdk:"max_terms_count"`
	MaxRegexLength                     types.Int64  `tfsdk:"max_regex_length"`
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
}
