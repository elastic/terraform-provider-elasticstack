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

package indexsettings_test

import (
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

// REQ-004: the four non-metadata index.blocks.* booleans round-trip as real
// booleans against live Elasticsearch across the lifecycle: create with all
// four set to true, update flipping all four to false, and removal via null
// reset. index.blocks.metadata is exercised at its explicit-false value only;
// setting it true makes the index's own settings GET return a
// cluster_block_exception, which the resource's read-after-write performs on
// every write, so blocks_metadata=true cannot be driven through this resource
// without tripping an unrelated, pre-existing issue in the create/update
// read-after-write step. That gap is out of scope for this coverage change.
func TestAccResourceIndexSettings_blocksLifecycle(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexSettingsDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_true"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(indexSettingsResourceName, "blocks_read_only", "true"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "blocks_read_only_allow_delete", "true"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "blocks_read", "true"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "blocks_write", "true"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "blocks_metadata", "false"),
					checkIndexSettingValue(indexName, "index.blocks.read_only", "true"),
					checkIndexSettingValue(indexName, "index.blocks.read_only_allow_delete", "true"),
					checkIndexSettingValue(indexName, "index.blocks.read", "true"),
					checkIndexSettingValue(indexName, "index.blocks.write", "true"),
					checkIndexSettingValue(indexName, "index.blocks.metadata", "false"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_false"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectNonEmptyPlan()},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(indexSettingsResourceName, "blocks_read_only", "false"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "blocks_read_only_allow_delete", "false"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "blocks_read", "false"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "blocks_write", "false"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "blocks_metadata", "false"),
					checkIndexSettingValue(indexName, "index.blocks.read_only", "false"),
					checkIndexSettingValue(indexName, "index.blocks.read_only_allow_delete", "false"),
					checkIndexSettingValue(indexName, "index.blocks.read", "false"),
					checkIndexSettingValue(indexName, "index.blocks.write", "false"),
					checkIndexSettingValue(indexName, "index.blocks.metadata", "false"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("index_only"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "blocks_read_only"),
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "blocks_read_only_allow_delete"),
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "blocks_read"),
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "blocks_write"),
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "blocks_metadata"),
					checkIndexSettingAbsent(indexName, "index.blocks.read_only"),
					checkIndexSettingAbsent(indexName, "index.blocks.read_only_allow_delete"),
					checkIndexSettingAbsent(indexName, "index.blocks.read"),
					checkIndexSettingAbsent(indexName, "index.blocks.write"),
					checkIndexSettingAbsent(indexName, "index.blocks.metadata"),
				),
			},
		},
	})
}

// REQ-004: routing_allocation_enable and routing_rebalance_enable, two of the
// four stringvalidator.OneOf-validated dynamic-setting attributes, round-trip
// their valid enum values against live Elasticsearch across create, update to
// a different valid value, and removal via null reset.
//
// search_slowlog_level and indexing_slowlog_level (the other two
// OneOf-validated attributes) are intentionally not exercised here: current
// Elasticsearch no longer accepts index.search.slowlog.level or
// index.indexing.slowlog.level at all ("unknown setting"), so no apply-based
// positive value can round-trip against a live cluster. Their invalid-value
// rejection is still covered by
// TestAccResourceIndexSettings_searchSlowlogLevelRejectsInvalidValue and
// TestAccResourceIndexSettings_indexingSlowlogLevelRejectsInvalidValue, since
// the OneOf validator rejects out-of-enum values at plan time before any API
// call is made.
func TestAccResourceIndexSettings_routingEnableSettingsLifecycle(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexSettingsDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("enums_v1"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(indexSettingsResourceName, "routing_allocation_enable", "primaries"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "routing_rebalance_enable", "primaries"),
					checkIndexSettingValue(indexName, "index.routing.allocation.enable", "primaries"),
					checkIndexSettingValue(indexName, "index.routing.rebalance.enable", "primaries"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("enums_v2"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectNonEmptyPlan()},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(indexSettingsResourceName, "routing_allocation_enable", "new_primaries"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "routing_rebalance_enable", "replicas"),
					checkIndexSettingValue(indexName, "index.routing.allocation.enable", "new_primaries"),
					checkIndexSettingValue(indexName, "index.routing.rebalance.enable", "replicas"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("index_only"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "routing_allocation_enable"),
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "routing_rebalance_enable"),
					checkIndexSettingAbsent(indexName, "index.routing.allocation.enable"),
					checkIndexSettingAbsent(indexName, "index.routing.rebalance.enable"),
				),
			},
		},
	})
}

// REQ-001: out-of-enum values for the four OneOf-validated attributes are
// rejected at plan time, exercising the negative path the validators were
// never asserted against.
func TestAccResourceIndexSettings_routingAllocationEnableRejectsInvalidValue(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory(""),
				ExpectError:              regexp.MustCompile(`(?i)routing_allocation_enable`),
			},
		},
	})
}

func TestAccResourceIndexSettings_routingRebalanceEnableRejectsInvalidValue(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory(""),
				ExpectError:              regexp.MustCompile(`(?i)routing_rebalance_enable`),
			},
		},
	})
}

func TestAccResourceIndexSettings_searchSlowlogLevelRejectsInvalidValue(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory(""),
				ExpectError:              regexp.MustCompile(`(?i)search_slowlog_level`),
			},
		},
	})
}

func TestAccResourceIndexSettings_indexingSlowlogLevelRejectsInvalidValue(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory(""),
				ExpectError:              regexp.MustCompile(`(?i)indexing_slowlog_level`),
			},
		},
	})
}

// REQ-004: the twelve numeric limit settings round-trip against live
// Elasticsearch across create, update to different values, and removal via
// null reset.
func TestAccResourceIndexSettings_numericLimitSettingsLifecycle(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexSettingsDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("numeric_v1"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(indexSettingsResourceName, "max_result_window", "12000"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "max_inner_result_window", "150"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "max_rescore_window", "12000"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "max_docvalue_fields_search", "150"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "max_script_fields", "40"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "max_ngram_diff", "2"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "max_shingle_diff", "4"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "max_refresh_listeners", "1200"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "analyze_max_token_count", "12000"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "highlight_max_analyzed_offset", "1200000"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "max_terms_count", "70000"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "max_regex_length", "1200"),
					checkIndexSettingValue(indexName, "index.max_result_window", "12000"),
					checkIndexSettingValue(indexName, "index.max_inner_result_window", "150"),
					checkIndexSettingValue(indexName, "index.max_rescore_window", "12000"),
					checkIndexSettingValue(indexName, "index.max_docvalue_fields_search", "150"),
					checkIndexSettingValue(indexName, "index.max_script_fields", "40"),
					checkIndexSettingValue(indexName, "index.max_ngram_diff", "2"),
					checkIndexSettingValue(indexName, "index.max_shingle_diff", "4"),
					checkIndexSettingValue(indexName, "index.max_refresh_listeners", "1200"),
					checkIndexSettingValue(indexName, "index.analyze.max_token_count", "12000"),
					checkIndexSettingValue(indexName, "index.highlight.max_analyzed_offset", "1200000"),
					checkIndexSettingValue(indexName, "index.max_terms_count", "70000"),
					checkIndexSettingValue(indexName, "index.max_regex_length", "1200"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("numeric_v2"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectNonEmptyPlan()},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(indexSettingsResourceName, "max_result_window", "15000"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "max_inner_result_window", "200"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "max_rescore_window", "15000"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "max_docvalue_fields_search", "200"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "max_script_fields", "50"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "max_ngram_diff", "3"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "max_shingle_diff", "5"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "max_refresh_listeners", "1500"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "analyze_max_token_count", "15000"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "highlight_max_analyzed_offset", "1500000"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "max_terms_count", "80000"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "max_regex_length", "1500"),
					checkIndexSettingValue(indexName, "index.max_result_window", "15000"),
					checkIndexSettingValue(indexName, "index.max_inner_result_window", "200"),
					checkIndexSettingValue(indexName, "index.max_rescore_window", "15000"),
					checkIndexSettingValue(indexName, "index.max_docvalue_fields_search", "200"),
					checkIndexSettingValue(indexName, "index.max_script_fields", "50"),
					checkIndexSettingValue(indexName, "index.max_ngram_diff", "3"),
					checkIndexSettingValue(indexName, "index.max_shingle_diff", "5"),
					checkIndexSettingValue(indexName, "index.max_refresh_listeners", "1500"),
					checkIndexSettingValue(indexName, "index.analyze.max_token_count", "15000"),
					checkIndexSettingValue(indexName, "index.highlight.max_analyzed_offset", "1500000"),
					checkIndexSettingValue(indexName, "index.max_terms_count", "80000"),
					checkIndexSettingValue(indexName, "index.max_regex_length", "1500"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("index_only"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "max_result_window"),
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "max_inner_result_window"),
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "max_rescore_window"),
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "max_docvalue_fields_search"),
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "max_script_fields"),
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "max_ngram_diff"),
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "max_shingle_diff"),
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "max_refresh_listeners"),
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "analyze_max_token_count"),
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "highlight_max_analyzed_offset"),
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "max_terms_count"),
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "max_regex_length"),
					checkIndexSettingAbsent(indexName, "index.max_result_window"),
					checkIndexSettingAbsent(indexName, "index.max_inner_result_window"),
					checkIndexSettingAbsent(indexName, "index.max_rescore_window"),
					checkIndexSettingAbsent(indexName, "index.max_docvalue_fields_search"),
					checkIndexSettingAbsent(indexName, "index.max_script_fields"),
					checkIndexSettingAbsent(indexName, "index.max_ngram_diff"),
					checkIndexSettingAbsent(indexName, "index.max_shingle_diff"),
					checkIndexSettingAbsent(indexName, "index.max_refresh_listeners"),
					checkIndexSettingAbsent(indexName, "index.analyze.max_token_count"),
					checkIndexSettingAbsent(indexName, "index.highlight.max_analyzed_offset"),
					checkIndexSettingAbsent(indexName, "index.max_terms_count"),
					checkIndexSettingAbsent(indexName, "index.max_regex_length"),
				),
			},
		},
	})
}

// REQ-004: the pipeline and timing-related dynamic-setting attributes
// round-trip against live Elasticsearch across create, update, and removal
// via null reset.
func TestAccResourceIndexSettings_pipelineAndTimingSettingsLifecycle(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexSettingsDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("pipeline_v1"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(indexSettingsResourceName, "auto_expand_replicas", "0-1"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "search_idle_after", "30s"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "gc_deletes", "30s"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "default_pipeline", "pipeline-one"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "final_pipeline", "pipeline-two"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "unassigned_node_left_delayed_timeout", "30s"),
					checkIndexSettingValue(indexName, "index.auto_expand_replicas", "0-1"),
					checkIndexSettingValue(indexName, "index.search.idle.after", "30s"),
					checkIndexSettingValue(indexName, "index.gc_deletes", "30s"),
					checkIndexSettingValue(indexName, "index.default_pipeline", "pipeline-one"),
					checkIndexSettingValue(indexName, "index.final_pipeline", "pipeline-two"),
					checkIndexSettingValue(indexName, "index.unassigned.node_left.delayed_timeout", "30s"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("pipeline_v2"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectNonEmptyPlan()},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(indexSettingsResourceName, "auto_expand_replicas", "0-all"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "search_idle_after", "60s"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "gc_deletes", "90s"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "default_pipeline", "pipeline-three"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "final_pipeline", "pipeline-four"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "unassigned_node_left_delayed_timeout", "45s"),
					checkIndexSettingValue(indexName, "index.auto_expand_replicas", "0-all"),
					checkIndexSettingValue(indexName, "index.search.idle.after", "60s"),
					checkIndexSettingValue(indexName, "index.gc_deletes", "90s"),
					checkIndexSettingValue(indexName, "index.default_pipeline", "pipeline-three"),
					checkIndexSettingValue(indexName, "index.final_pipeline", "pipeline-four"),
					checkIndexSettingValue(indexName, "index.unassigned.node_left.delayed_timeout", "45s"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("index_only"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "auto_expand_replicas"),
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "search_idle_after"),
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "gc_deletes"),
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "default_pipeline"),
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "final_pipeline"),
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "unassigned_node_left_delayed_timeout"),
					checkIndexSettingAbsent(indexName, "index.auto_expand_replicas"),
					checkIndexSettingAbsent(indexName, "index.search.idle.after"),
					checkIndexSettingAbsent(indexName, "index.gc_deletes"),
					checkIndexSettingAbsent(indexName, "index.default_pipeline"),
					checkIndexSettingAbsent(indexName, "index.final_pipeline"),
					checkIndexSettingAbsent(indexName, "index.unassigned.node_left.delayed_timeout"),
				),
			},
		},
	})
}

// REQ-004: the twelve slowlog threshold settings plus
// indexing_slowlog_source round-trip against live Elasticsearch across
// create, update, and removal via null reset.
func TestAccResourceIndexSettings_slowlogSettingsLifecycle(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexSettingsDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("slowlog_v1"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(indexSettingsResourceName, "search_slowlog_threshold_query_warn", "10s"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "search_slowlog_threshold_query_info", "5s"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "search_slowlog_threshold_query_debug", "2s"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "search_slowlog_threshold_query_trace", "500ms"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "search_slowlog_threshold_fetch_warn", "10s"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "search_slowlog_threshold_fetch_info", "5s"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "search_slowlog_threshold_fetch_debug", "2s"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "search_slowlog_threshold_fetch_trace", "500ms"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "indexing_slowlog_threshold_index_warn", "10s"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "indexing_slowlog_threshold_index_info", "5s"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "indexing_slowlog_threshold_index_debug", "2s"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "indexing_slowlog_threshold_index_trace", "500ms"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "indexing_slowlog_source", "1000"),
					checkIndexSettingValue(indexName, "index.search.slowlog.threshold.query.warn", "10s"),
					checkIndexSettingValue(indexName, "index.search.slowlog.threshold.query.info", "5s"),
					checkIndexSettingValue(indexName, "index.search.slowlog.threshold.query.debug", "2s"),
					checkIndexSettingValue(indexName, "index.search.slowlog.threshold.query.trace", "500ms"),
					checkIndexSettingValue(indexName, "index.search.slowlog.threshold.fetch.warn", "10s"),
					checkIndexSettingValue(indexName, "index.search.slowlog.threshold.fetch.info", "5s"),
					checkIndexSettingValue(indexName, "index.search.slowlog.threshold.fetch.debug", "2s"),
					checkIndexSettingValue(indexName, "index.search.slowlog.threshold.fetch.trace", "500ms"),
					checkIndexSettingValue(indexName, "index.indexing.slowlog.threshold.index.warn", "10s"),
					checkIndexSettingValue(indexName, "index.indexing.slowlog.threshold.index.info", "5s"),
					checkIndexSettingValue(indexName, "index.indexing.slowlog.threshold.index.debug", "2s"),
					checkIndexSettingValue(indexName, "index.indexing.slowlog.threshold.index.trace", "500ms"),
					checkIndexSettingValue(indexName, "index.indexing.slowlog.source", "1000"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("slowlog_v2"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectNonEmptyPlan()},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(indexSettingsResourceName, "search_slowlog_threshold_query_warn", "20s"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "search_slowlog_threshold_query_info", "8s"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "search_slowlog_threshold_query_debug", "3s"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "search_slowlog_threshold_query_trace", "750ms"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "search_slowlog_threshold_fetch_warn", "20s"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "search_slowlog_threshold_fetch_info", "8s"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "search_slowlog_threshold_fetch_debug", "3s"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "search_slowlog_threshold_fetch_trace", "750ms"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "indexing_slowlog_threshold_index_warn", "20s"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "indexing_slowlog_threshold_index_info", "8s"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "indexing_slowlog_threshold_index_debug", "3s"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "indexing_slowlog_threshold_index_trace", "750ms"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "indexing_slowlog_source", "false"),
					checkIndexSettingValue(indexName, "index.search.slowlog.threshold.query.warn", "20s"),
					checkIndexSettingValue(indexName, "index.search.slowlog.threshold.query.info", "8s"),
					checkIndexSettingValue(indexName, "index.search.slowlog.threshold.query.debug", "3s"),
					checkIndexSettingValue(indexName, "index.search.slowlog.threshold.query.trace", "750ms"),
					checkIndexSettingValue(indexName, "index.search.slowlog.threshold.fetch.warn", "20s"),
					checkIndexSettingValue(indexName, "index.search.slowlog.threshold.fetch.info", "8s"),
					checkIndexSettingValue(indexName, "index.search.slowlog.threshold.fetch.debug", "3s"),
					checkIndexSettingValue(indexName, "index.search.slowlog.threshold.fetch.trace", "750ms"),
					checkIndexSettingValue(indexName, "index.indexing.slowlog.threshold.index.warn", "20s"),
					checkIndexSettingValue(indexName, "index.indexing.slowlog.threshold.index.info", "8s"),
					checkIndexSettingValue(indexName, "index.indexing.slowlog.threshold.index.debug", "3s"),
					checkIndexSettingValue(indexName, "index.indexing.slowlog.threshold.index.trace", "750ms"),
					checkIndexSettingValue(indexName, "index.indexing.slowlog.source", "false"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("index_only"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "search_slowlog_threshold_query_warn"),
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "search_slowlog_threshold_query_info"),
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "search_slowlog_threshold_query_debug"),
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "search_slowlog_threshold_query_trace"),
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "search_slowlog_threshold_fetch_warn"),
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "search_slowlog_threshold_fetch_info"),
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "search_slowlog_threshold_fetch_debug"),
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "search_slowlog_threshold_fetch_trace"),
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "indexing_slowlog_threshold_index_warn"),
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "indexing_slowlog_threshold_index_info"),
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "indexing_slowlog_threshold_index_debug"),
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "indexing_slowlog_threshold_index_trace"),
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "indexing_slowlog_source"),
					checkIndexSettingAbsent(indexName, "index.search.slowlog.threshold.query.warn"),
					checkIndexSettingAbsent(indexName, "index.search.slowlog.threshold.query.info"),
					checkIndexSettingAbsent(indexName, "index.search.slowlog.threshold.query.debug"),
					checkIndexSettingAbsent(indexName, "index.search.slowlog.threshold.query.trace"),
					checkIndexSettingAbsent(indexName, "index.search.slowlog.threshold.fetch.warn"),
					checkIndexSettingAbsent(indexName, "index.search.slowlog.threshold.fetch.info"),
					checkIndexSettingAbsent(indexName, "index.search.slowlog.threshold.fetch.debug"),
					checkIndexSettingAbsent(indexName, "index.search.slowlog.threshold.fetch.trace"),
					checkIndexSettingAbsent(indexName, "index.indexing.slowlog.threshold.index.warn"),
					checkIndexSettingAbsent(indexName, "index.indexing.slowlog.threshold.index.info"),
					checkIndexSettingAbsent(indexName, "index.indexing.slowlog.threshold.index.debug"),
					checkIndexSettingAbsent(indexName, "index.indexing.slowlog.threshold.index.trace"),
					checkIndexSettingAbsent(indexName, "index.indexing.slowlog.source"),
				),
			},
		},
	})
}
