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

package index_test

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	esclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const indexSettingsAnalysisAnalyzerExpected = `{"text_en":{"char_filter":"zero_width_spaces","filter":["lowercase","minimal_english_stemmer"],` +
	`"tokenizer":"standard","type":"custom"}}`
const indexAliasFilterExpected = `{"term":{"user.id":"developer"}}`

func TestAccResourceIndex(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test", "name", indexName),
					resource.TestMatchTypeSetElemNestedAttrs("elasticstack_elasticsearch_index.test", "alias.*", map[string]*regexp.Regexp{
						"name": regexp.MustCompile("test_alias_1"),
					}),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_index.test", "alias.*", map[string]string{
						"name":   "test_alias_2",
						"filter": indexAliasFilterExpected,
					}),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test", "alias.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test", "wait_for_active_shards", "all"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test", "master_timeout", "1m"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test", "timeout", "1m"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				ResourceName: "elasticstack_elasticsearch_index.test",
				Destroy:      true,
				ExpectError:  regexp.MustCompile("cannot destroy index without setting deletion_protection=false and running `terraform apply`"),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test", "name", indexName),
					resource.TestMatchTypeSetElemNestedAttrs("elasticstack_elasticsearch_index.test", "alias.*", map[string]*regexp.Regexp{
						"name": regexp.MustCompile("test_alias_1"),
					}),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test", "alias.#", "1"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("zero_replicas"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test", "name", indexName),
					resource.TestMatchTypeSetElemNestedAttrs("elasticstack_elasticsearch_index.test", "alias.*", map[string]*regexp.Regexp{
						"name": regexp.MustCompile("test_alias_1"),
					}),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test", "alias.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test", "number_of_replicas", "0"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("zero_replicas"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				ImportState:  true,
				ResourceName: "elasticstack_elasticsearch_index.test",
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test", "name", indexName),
					resource.TestMatchTypeSetElemNestedAttrs("elasticstack_elasticsearch_index.test", "alias.*", map[string]*regexp.Regexp{
						"name": regexp.MustCompile("test_alias_1"),
					}),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test", "alias.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test", "number_of_replicas", "0"),
				),
			},
		},
	})
}

//go:embed testdata/TestAccResourceIndexFromSDK/main.tf
var sdkCreateTestConfig string

func TestAccResourceIndexFromSDK(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexDestroy,
		Steps: []resource.TestStep{
			{
				// Create the index with the last provider version where the index resource was built on the SDK
				ExternalProviders: map[string]resource.ExternalProvider{
					"elasticstack": {
						Source:            "elastic/elasticstack",
						VersionConstraint: "0.11.3",
					},
				},
				Config: sdkCreateTestConfig,
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "name", indexName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "number_of_shards", "2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "number_of_routing_shards", "2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "codec", "best_compression"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "routing_partition_size", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "shard_check_on_startup", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "sort_field.0", "sort_key"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "sort_order.0", "asc"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "mapping_coerce", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "auto_expand_replicas", "0-5"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "search_idle_after", "30s"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "refresh_interval", "10s"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "max_result_window", "5000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "max_inner_result_window", "2000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "max_rescore_window", "1000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "max_docvalue_fields_search", "1500"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "max_script_fields", "500"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "max_ngram_diff", "100"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "max_shingle_diff", "200"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "max_refresh_listeners", "10"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "analyze_max_token_count", "500000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "highlight_max_analyzed_offset", "1000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "max_terms_count", "10000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "max_regex_length", "1000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "query_default_field.0", "field1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "routing_allocation_enable", "primaries"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "routing_rebalance_enable", "primaries"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "gc_deletes", "30s"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "unassigned_node_left_delayed_timeout", "5m"),
					resource.TestCheckResourceAttr(
						"elasticstack_elasticsearch_index.test_settings",
						"analysis_analyzer",
						indexSettingsAnalysisAnalyzerExpected,
					),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "analysis_char_filter", `{"zero_width_spaces":{"mappings":["\\u200C=\u003e\\u0020"],"type":"mapping"}}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "analysis_filter", `{"minimal_english_stemmer":{"language":"minimal_english","type":"stemmer"}}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "settings.0.setting.0.name", "number_of_replicas"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "settings.0.setting.0.value", "2"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory(""),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "name", indexName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "number_of_shards", "2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "number_of_routing_shards", "2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "codec", "best_compression"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "routing_partition_size", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "shard_check_on_startup", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "sort_field.0", "sort_key"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "sort_order.0", "asc"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "mapping_coerce", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "auto_expand_replicas", "0-5"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "search_idle_after", "30s"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "refresh_interval", "10s"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "max_result_window", "5000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "max_inner_result_window", "2000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "max_rescore_window", "1000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "max_docvalue_fields_search", "1500"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "max_script_fields", "500"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "max_ngram_diff", "100"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "max_shingle_diff", "200"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "max_refresh_listeners", "10"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "analyze_max_token_count", "500000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "highlight_max_analyzed_offset", "1000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "max_terms_count", "10000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "max_regex_length", "1000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "query_default_field.0", "field1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "routing_allocation_enable", "primaries"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "routing_rebalance_enable", "primaries"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "gc_deletes", "30s"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "unassigned_node_left_delayed_timeout", "5m"),
					resource.TestCheckResourceAttr(
						"elasticstack_elasticsearch_index.test_settings",
						"analysis_analyzer",
						indexSettingsAnalysisAnalyzerExpected,
					),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "analysis_char_filter", `{"zero_width_spaces":{"mappings":["\\u200C=\u003e\\u0020"],"type":"mapping"}}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "analysis_filter", `{"minimal_english_stemmer":{"language":"minimal_english","type":"stemmer"}}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "settings.0.setting.0.name", "number_of_replicas"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "settings.0.setting.0.value", "2"),
				),
			},
		},
	})
}

func TestAccResourceIndexSettings(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "name", indexName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "number_of_shards", "2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "number_of_routing_shards", "2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "codec", "best_compression"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "routing_partition_size", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "shard_check_on_startup", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "sort_field.0", "sort_key"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "sort_order.0", "asc"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "mapping_coerce", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "mapping_total_fields_limit", "2000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "auto_expand_replicas", "0-5"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "search_idle_after", "30s"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "refresh_interval", "10s"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "max_result_window", "5000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "max_inner_result_window", "2000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "max_rescore_window", "1000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "max_docvalue_fields_search", "1500"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "max_script_fields", "500"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "max_ngram_diff", "100"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "max_shingle_diff", "200"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "max_refresh_listeners", "10"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "analyze_max_token_count", "500000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "highlight_max_analyzed_offset", "1000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "max_terms_count", "10000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "max_regex_length", "1000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "query_default_field.0", "field1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "routing_allocation_enable", "primaries"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "routing_rebalance_enable", "primaries"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "gc_deletes", "30s"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "unassigned_node_left_delayed_timeout", "5m"),
					resource.TestCheckResourceAttr(
						"elasticstack_elasticsearch_index.test_settings",
						"analysis_analyzer",
						indexSettingsAnalysisAnalyzerExpected,
					),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "analysis_char_filter", `{"zero_width_spaces":{"mappings":["\\u200C=\u003e\\u0020"],"type":"mapping"}}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "analysis_filter", `{"minimal_english_stemmer":{"language":"minimal_english","type":"stemmer"}}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "analysis_tokenizer", `{"path_tokenizer":{"delimiter":"/","type":"path_hierarchy"}}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "analysis_normalizer", `{"lowercase_normalizer":{"filter":["lowercase"],"type":"custom"}}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "settings.0.setting.0.name", "number_of_replicas"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "settings.0.setting.0.value", "2"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index.test_settings", "settings_raw"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "deletion_protection", "false"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "name", indexName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "mapping_total_fields_limit", "3000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "load_fixed_bitset_filters_eagerly", "true"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index.test_settings", "settings_raw"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "deletion_protection", "false"),
				),
			},
		},
	})
}

func TestAccResourceIndexWithTemplate(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test", "name", indexName),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index.test", "default_pipeline"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index.test", "mappings"),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_index.test", "alias.*", map[string]string{
						"name":           fmt.Sprintf("%s-alias", indexName),
						"is_write_index": "true",
					}),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test", "deletion_protection", "false"),
				),
			},
		},
	})
}

func TestAccResourceIndexTemplateNoMappingDrift(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test", "name", indexName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test", "deletion_protection", "false"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccResourceIndexTemplateUserMappingNoDrift(t *testing.T) {

	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test", "name", indexName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test", "deletion_protection", "false"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccResourceIndexRemovingField(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexDestroy,
		Steps: []resource.TestStep{
			// Confirm removing field doesn't produce recreate by using prevent_destroy
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				ExpectNonEmptyPlan: true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("post_update"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccResourceIndexBlocks(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"index_name":   config.StringVariable(indexName),
					"blocks_write": config.BoolVariable(true),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_blocks", "name", indexName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_blocks", "blocks_write", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_blocks", "blocks_read", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_blocks", "blocks_read_only", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_blocks", "blocks_read_only_allow_delete", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_blocks", "blocks_metadata", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_blocks", "deletion_protection", "false"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"index_name":   config.StringVariable(indexName),
					"blocks_write": config.BoolVariable(false),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_blocks", "name", indexName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_blocks", "blocks_write", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_blocks", "blocks_read", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_blocks", "blocks_read_only", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_blocks", "blocks_read_only_allow_delete", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_blocks", "blocks_metadata", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_blocks", "deletion_protection", "false"),
				),
			},
		},
	})
}

func TestAccResourceIndexSlowlog(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_slowlog", "name", indexName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_slowlog", "search_slowlog_threshold_query_warn", "10s"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_slowlog", "search_slowlog_threshold_query_info", "5s"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_slowlog", "search_slowlog_threshold_query_debug", "2s"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_slowlog", "search_slowlog_threshold_query_trace", "500ms"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_slowlog", "search_slowlog_threshold_fetch_warn", "1s"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_slowlog", "search_slowlog_threshold_fetch_info", "800ms"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_slowlog", "search_slowlog_threshold_fetch_debug", "500ms"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_slowlog", "search_slowlog_threshold_fetch_trace", "200ms"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_slowlog", "indexing_slowlog_threshold_index_warn", "10s"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_slowlog", "indexing_slowlog_threshold_index_info", "20ms"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_slowlog", "indexing_slowlog_threshold_index_debug", "10ms"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_slowlog", "indexing_slowlog_threshold_index_trace", "5ms"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_slowlog", "indexing_slowlog_source", "1000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_slowlog", "deletion_protection", "false"),
				),
			},
		},
	})
}

// search_slowlog_level and indexing_slowlog_level are only supported in Elastic Stack 7.x.
var indexingSlowlogLevelVersionConstraint, _ = version.NewConstraint("< 8.0.0")

func TestAccResourceIndexSlowlogLevel(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(indexingSlowlogLevelVersionConstraint),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_slowlog_level", "name", indexName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_slowlog_level", "search_slowlog_level", "info"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_slowlog_level", "indexing_slowlog_level", "warn"),
				),
			},
		},
	})
}

func TestAccResourceIndexPipelines(t *testing.T) {
	pipelineName := sdkacctest.RandomWithPrefix("tf-acc-test")
	indexName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"index_name":    config.StringVariable(indexName),
					"pipeline_name": config.StringVariable(pipelineName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_pipelines", "name", indexName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_pipelines", "default_pipeline", pipelineName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_pipelines", "final_pipeline", pipelineName),
				),
			},
		},
	})
}

func createElasticsearchIndexOOB(t *testing.T, name, body string) {
	t.Helper()
	ctx := context.Background()
	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		t.Fatalf("acceptance elasticsearch client: %v", err)
	}
	typedClient, err := client.GetESClient()
	if err != nil {
		t.Fatalf("get Elasticsearch typed client: %v", err)
	}
	if _, err := typedClient.Indices.Create(name).Raw(strings.NewReader(body)).Do(ctx); err != nil {
		t.Fatalf("Indices.Create(%q): %v", name, err)
	}
}

func deleteElasticsearchIndexOOB(t *testing.T, name string) {
	t.Helper()
	ctx := context.Background()
	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		t.Logf("cleanup: acceptance elasticsearch client: %v", err)
		return
	}
	typedClient, err := client.GetESClient()
	if err != nil {
		t.Logf("cleanup: get Elasticsearch typed client: %v", err)
		return
	}
	if _, err := typedClient.Indices.Delete(name).Do(ctx); err != nil {
		if esclient.IsNotFoundElasticsearchError(err) {
			return
		}
		t.Logf("cleanup: Indices.Delete(%q): %v", name, err)
	}
}

func getElasticsearchIndexState(t *testing.T, indexName string) types.IndexState {
	t.Helper()
	ctx := context.Background()
	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		t.Fatalf("acceptance elasticsearch client: %v", err)
	}
	typedClient, err := client.GetESClient()
	if err != nil {
		t.Fatalf("get Elasticsearch typed client: %v", err)
	}
	resp, err := typedClient.Indices.Get(indexName).Do(ctx)
	if err != nil {
		if esclient.IsNotFoundElasticsearchError(err) {
			t.Fatalf("index %q not found", indexName)
		}
		t.Fatalf("Indices.Get(%q): %v", indexName, err)
	}
	state, ok := resp[indexName]
	if !ok {
		t.Fatalf("index %q not present in response (have %d keys)", indexName, len(resp))
	}
	return state
}

func primaryShardsString(settings *types.IndexSettings) string {
	if settings == nil {
		return ""
	}
	if settings.Index != nil && settings.Index.NumberOfShards != nil {
		return strings.TrimSpace(*settings.Index.NumberOfShards)
	}
	return ""
}

func assertIndexPrimaryShards(t *testing.T, indexName, want string) {
	t.Helper()
	state := getElasticsearchIndexState(t, indexName)
	got := primaryShardsString(state.Settings)
	if got != want {
		t.Fatalf("index %q primary shards: want %q, got %q", indexName, want, got)
	}
}

func assertIndexAliasesExactly(t *testing.T, indexName string, want []string) {
	t.Helper()
	state := getElasticsearchIndexState(t, indexName)
	got := make([]string, 0, len(state.Aliases))
	for k := range state.Aliases {
		got = append(got, k)
	}
	sort.Strings(got)
	if want == nil {
		want = []string{}
	}
	wantSorted := append([]string(nil), want...)
	sort.Strings(wantSorted)
	if !slices.Equal(got, wantSorted) {
		t.Fatalf("index %q aliases: want %v, got %v", indexName, wantSorted, got)
	}
}

func TestAccResourceIndexUseExistingFallthrough(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_use_existing", "name", indexName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_use_existing", "concrete_name", indexName),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index.test_use_existing", "id"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccResourceIndexUseExistingAdoptAliasReconcile(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				PreConfig: func() {
					createElasticsearchIndexOOB(t, indexName, `{
  "settings": { "index": { "number_of_shards": 1 } },
  "aliases": { "legacy_alias": {} }
}`)
				},
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_use_existing", "name", indexName),
					func(_ *terraform.State) error {
						assertIndexAliasesExactly(t, indexName, []string{"new_alias"})
						return nil
					},
				),
			},
		},
	})
}

func TestAccResourceIndexUseExistingTemplateNoMappingDrift(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("step1_template"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "name", indexName),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				PreConfig: func() {
					createElasticsearchIndexOOB(t, indexName, `{"settings":{"index":{"number_of_shards":1}}}`)
				},
				ConfigDirectory: acctest.NamedTestCaseDirectory("step2_adopt"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_adopt_template", "name", indexName),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("step2_adopt"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccResourceIndexUseExistingAdopt(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				PreConfig: func() {
					createElasticsearchIndexOOB(t, indexName, `{"settings":{"index":{"number_of_shards":1}}}`)
				},
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_use_existing", "name", indexName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_use_existing", "concrete_name", indexName),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index.test_use_existing", "id"),
					resource.TestMatchTypeSetElemNestedAttrs("elasticstack_elasticsearch_index.test_use_existing", "alias.*", map[string]*regexp.Regexp{
						"name": regexp.MustCompile("adopt_alias_step1"),
					}),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_alias"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_use_existing", "name", indexName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_use_existing", "alias.#", "1"),
					resource.TestMatchTypeSetElemNestedAttrs("elasticstack_elasticsearch_index.test_use_existing", "alias.*", map[string]*regexp.Regexp{
						"name": regexp.MustCompile("adopt_alias_step1"),
					}),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_index.test_use_existing", "alias.*", map[string]string{
						"filter": indexAliasFilterExpected,
					}),
				),
			},
		},
	})
}

func TestAccResourceIndexUseExistingMismatch(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	t.Cleanup(func() {
		deleteElasticsearchIndexOOB(t, indexName)
	})
	// When TF_ACC is off, PreCheck fatals before the step runs; skip ES verification in that case.
	t.Cleanup(func() {
		if os.Getenv("TF_ACC") != "1" {
			return
		}
		assertIndexPrimaryShards(t, indexName, "1")
		assertIndexAliasesExactly(t, indexName, nil)
	})

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				PreConfig: func() {
					createElasticsearchIndexOOB(t, indexName, `{"settings":{"index":{"number_of_shards":1}}}`)
				},
				ConfigDirectory: acctest.NamedTestCaseDirectory("apply_mismatch"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				ExpectError: regexp.MustCompile(`number_of_shards: configured=2, actual=1`),
			},
		},
	})
}

func TestAccResourceIndexUseExistingDateMath(t *testing.T) {
	// Random label so this test does not fight TestAccResourceIndexDateMath (same <logs-{now/d}> would resolve to one concrete index).
	suffix := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	dateMathName := fmt.Sprintf("<useexist-%s-{now/d}>", suffix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(dateMathName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_date_math_use_existing", "name", dateMathName),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index.test_date_math_use_existing", "concrete_name"),
					resource.TestCheckResourceAttrWith("elasticstack_elasticsearch_index.test_date_math_use_existing", "concrete_name", func(val string) error {
						if val == dateMathName {
							return fmt.Errorf("concrete_name %q must not equal the date math expression", val)
						}
						return nil
					}),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(dateMathName),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccResourceIndexDateMath covers create/read stability and update-path
// regression for date math index names:
//   - Task 3.2: state preserves the configured date math expression in `name` and
//     persists the resolved concrete index in `concrete_name`.
//   - Task 3.3: alias and mapping updates after create from a date math expression
//     target the concrete managed index.
func TestAccResourceIndexDateMath(t *testing.T) {
	// Use a fixed date math expression.  The concrete index name resolved by
	// Elasticsearch will differ from this expression (e.g. logs-2024.01.15).
	dateMathName := `<logs-{now/d}>`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexDestroy,
		Steps: []resource.TestStep{
			{
				// Step 1: create the index using the date math expression and verify
				// that name is preserved as the configured expression while
				// concrete_name holds the resolved concrete index.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(dateMathName),
				},
				Check: resource.ComposeTestCheckFunc(
					// name must remain the configured date math expression.
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_date_math", "name", dateMathName),
					// concrete_name must be set and must differ from the date math expression.
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index.test_date_math", "concrete_name"),
					resource.TestCheckResourceAttrWith("elasticstack_elasticsearch_index.test_date_math", "concrete_name", func(val string) error {
						if val == dateMathName {
							return fmt.Errorf("concrete_name %q must not equal the date math expression", val)
						}
						return nil
					}),
					// alias created during create must be present.
					resource.TestMatchTypeSetElemNestedAttrs("elasticstack_elasticsearch_index.test_date_math", "alias.*", map[string]*regexp.Regexp{
						"name": regexp.MustCompile("date_math_alias_1"),
					}),
				),
			},
			{
				// Step 2: update aliases and mappings — all updates must target the
				// persisted concrete index identity, not the date math expression.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_alias"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(dateMathName),
				},
				Check: resource.ComposeTestCheckFunc(
					// name is still the configured date math expression after the update.
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_date_math", "name", dateMathName),
					// concrete_name is still set (preserved from create).
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index.test_date_math", "concrete_name"),
					// Alias from create is still present after the mappings update.
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_date_math", "alias.#", "1"),
					resource.TestMatchTypeSetElemNestedAttrs("elasticstack_elasticsearch_index.test_date_math", "alias.*", map[string]*regexp.Regexp{
						"name": regexp.MustCompile("date_math_alias_1"),
					}),
				),
			},
		},
	})
}

func checkResourceIndexDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_index" {
			continue
		}
		compID, _ := clients.CompositeIDFromStr(rs.Primary.ID)

		typedClient, err := client.GetESClient()
		if err != nil {
			return err
		}
		_, err = typedClient.Indices.Get(compID.ResourceID).Do(context.Background())
		if err != nil {
			if esclient.IsNotFoundElasticsearchError(err) {
				continue
			}
			return err
		}

		return fmt.Errorf("Index (%s) still exists", compID.ResourceID)
	}
	return nil
}
