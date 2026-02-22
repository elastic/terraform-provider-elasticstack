package index_test

import (
	_ "embed"
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const indexSettingsAnalysisAnalyzerExpected = `{"text_en":{"char_filter":"zero_width_spaces","filter":["lowercase","minimal_english_stemmer"],` +
	`"tokenizer":"standard","type":"custom"}}`

func TestAccResourceIndex(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
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
					resource.TestMatchTypeSetElemNestedAttrs("elasticstack_elasticsearch_index.test", "alias.*", map[string]*regexp.Regexp{
						"name": regexp.MustCompile("test_alias_2"),
					}),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test", "alias.#", "2"),
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

//go:embed testdata/TestAccResourceIndexFromSDK/index.tf
var sdkCreateTestConfig string

func TestAccResourceIndexFromSDK(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
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

	resource.Test(t, resource.TestCase{
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
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "settings.0.setting.0.name", "number_of_replicas"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "settings.0.setting.0.value", "2"),
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
				),
			},
		},
	})
}

func TestAccResourceIndexWithTemplate(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
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
				),
			},
		},
	})
}

func TestAccResourceIndexRemovingField(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
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

func checkResourceIndexDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_index" {
			continue
		}
		compID, _ := clients.CompositeIDFromStr(rs.Primary.ID)

		esClient, err := client.GetESClient()
		if err != nil {
			return err
		}
		res, err := esClient.Indices.Get([]string{compID.ResourceID})
		if err != nil {
			return err
		}

		if res.StatusCode != 404 {
			return fmt.Errorf("Index (%s) still exists", compID.ResourceID)
		}
	}
	return nil
}
