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

func TestAccResourceIndex(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceIndexDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceIndexCreate(indexName),
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
				Config:       testAccResourceIndexCreate(indexName),
				ResourceName: "elasticstack_elasticsearch_index.test",
				Destroy:      true,
				ExpectError:  regexp.MustCompile("cannot destroy index without setting deletion_protection=false and running `terraform apply`"),
			},
			{
				Config: testAccResourceIndexUpdate(indexName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test", "name", indexName),
					resource.TestMatchTypeSetElemNestedAttrs("elasticstack_elasticsearch_index.test", "alias.*", map[string]*regexp.Regexp{
						"name": regexp.MustCompile("test_alias_1"),
					}),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test", "alias.#", "1"),
				),
			},
			{
				Config: testAccResourceIndexZeroReplicas(indexName),
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
				Config:       testAccResourceIndexZeroReplicas(indexName),
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
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "analysis_analyzer", `{"text_en":{"char_filter":"zero_width_spaces","filter":["lowercase","minimal_english_stemmer"],"tokenizer":"standard","type":"custom"}}`),
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
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "analysis_analyzer", `{"text_en":{"char_filter":"zero_width_spaces","filter":["lowercase","minimal_english_stemmer"],"tokenizer":"standard","type":"custom"}}`),
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
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "analysis_analyzer", `{"text_en":{"char_filter":"zero_width_spaces","filter":["lowercase","minimal_english_stemmer"],"tokenizer":"standard","type":"custom"}}`),
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
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceIndexDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceIndexWithTemplate(indexName),
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
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceIndexDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			// Confirm removing field doesn't produce recreate by using prevent_destroy
			{Config: testAccResourceIndexRemovingFieldCreate(indexName)},
			{Config: testAccResourceIndexRemovingFieldUpdate(indexName), ExpectNonEmptyPlan: true},
			{Config: testAccResourceIndexRemovingFieldPostUpdate(indexName), ExpectNonEmptyPlan: true},
		},
	})
}

func testAccResourceIndexCreate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test" {
  name = "%s"

  alias = [
    {
      name = "test_alias_1"
    },
    {
      name = "test_alias_2"
      filter = jsonencode({
        term = { "user.id" = "developer" }
      })
    }
  ]

  mappings = jsonencode({
    properties = {
      field1 = { type = "text" }
    }
  })

	wait_for_active_shards = "all"
	master_timeout = "1m"
	timeout = "1m"
}
	`, name)
}

func testAccResourceIndexUpdate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test" {
  name = "%s"

  alias = [
    {
      name = "test_alias_1"
    }
  ]

  mappings = jsonencode({
    properties = {
      field1 = { type = "text" }
    }
  })

  deletion_protection = false
}
	`, name)
}

func testAccResourceIndexZeroReplicas(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test" {
  name = "%s"
  number_of_replicas = 0

  alias = [
    {
      name = "test_alias_1"
    }
  ]

  mappings = jsonencode({
    properties = {
      field1 = { type = "text" }
    }
  })

  deletion_protection = false
}
	`, name)
}

func testAccResourceIndexRemovingFieldCreate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test_settings_removing_field" {
  name = "%s"

  mappings = jsonencode({
    properties = {
      field1    = { type = "text" }
      field2    = { type = "text" }
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}
	`, name)
}

func testAccResourceIndexRemovingFieldUpdate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test_settings_removing_field" {
  name = "%s"

  mappings = jsonencode({
    properties = {
      field1    = { type = "text" }
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}
	`, name)
}

func testAccResourceIndexRemovingFieldPostUpdate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test_settings_removing_field" {
  name = "%s"

  mappings = jsonencode({
    properties = {
      field1    = { type = "text" }
    }
  })
  deletion_protection = false
}
	`, name)
}

func testAccResourceIndexWithTemplate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = "%s"
  index_patterns = ["%s"]
  template {
    settings = jsonencode({
      default_pipeline = ".fleet_final_pipeline-1"
      lifecycle        = { name = ".monitoring-8-ilm-policy" }
    })
	mappings = jsonencode({
      dynamic_templates = [
        {
          strings_as_ip = {
            match_mapping_type = "string",
            match              = "ip*",
            runtime = {
              type = "ip"
            }
          }
        }
      ]
	})
  }
}

resource "elasticstack_elasticsearch_index" "test" {
  name                = "%s"
  deletion_protection = false
  alias = [
    {
      name           = "%s-alias"
      is_write_index = true
    }
  ]
  lifecycle {
    ignore_changes = [mappings]
  }
  depends_on = [elasticstack_elasticsearch_index_template.test]
}
`, name, name, name, name)
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
		compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)

		esClient, err := client.GetESClient()
		if err != nil {
			return err
		}
		res, err := esClient.Indices.Get([]string{compId.ResourceId})
		if err != nil {
			return err
		}

		if res.StatusCode != 404 {
			return fmt.Errorf("Index (%s) still exists", compId.ResourceId)
		}
	}
	return nil
}
