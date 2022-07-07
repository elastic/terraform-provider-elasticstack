package index_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceIndex(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		CheckDestroy:      checkResourceIndexDestroy,
		ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceIndexCreate(indexName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test", "name", indexName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test", "alias.0.name", "test_alias_1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test", "alias.1.name", "test_alias_2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test", "alias.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test", "settings.0.setting.#", "2"),
				),
			},
			{
				Config: testAccResourceIndexUpdate(indexName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test", "name", indexName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test", "alias.0.name", "test_alias_1"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index.test", "alias.1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test", "alias.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test", "settings.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceIndexSettings(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		CheckDestroy:      checkResourceIndexDestroy,
		ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceIndexSettingsCreate(indexName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "name", indexName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "number_of_shards", "2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "number_of_routing_shards", "2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "codec", "best_compression"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "routing_partition_size", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "shard_check_on_startup", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "sort_field.0", "sort_key"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings", "sort_order.0", "asc"),
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

func TestAccResourceIndexSettingsMigration(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		CheckDestroy:      checkResourceIndexDestroy,
		ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceIndexSettingsMigrationCreate(indexName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings_migration", "name", indexName),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index.test_settings_migration", "number_of_replicas"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings_migration", "settings.0.setting.0.name", "number_of_replicas"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings_migration", "settings.0.setting.0.value", "2"),
				),
			},
			{
				Config: testAccResourceIndexSettingsMigrationUpdate(indexName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings_migration", "name", indexName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index.test_settings_migration", "number_of_replicas", "1"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index.test_settings_migration", "settings.#"),
				),
			},
		},
	})
}

func TestAccResourceIndexSettingsConflict(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		CheckDestroy:      checkResourceIndexDestroy,
		ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceIndexSettingsConflict(indexName),
				ExpectError: regexp.MustCompile("setting 'number_of_shards' is already defined by the other field, please remove it from `settings` to avoid unexpected settings"),
			},
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

  alias {
    name = "test_alias_1"
  }
  alias {
    name = "test_alias_2"
    filter = jsonencode({
      term = { "user.id" = "developer" }
    })
  }

  mappings = jsonencode({
    properties = {
      field1 = { type = "text" }
    }
  })

  settings {
    setting {
      name  = "index.number_of_replicas"
      value = "2"
    }
    setting {
      name  = "index.search.idle.after"
      value = "20s"
    }
  }
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

  alias {
    name = "test_alias_1"
  }

  mappings = jsonencode({
    properties = {
      field1 = { type = "text" }
    }
  })
}
	`, name)
}

func testAccResourceIndexSettingsCreate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test_settings" {
  name = "%s"

  mappings = jsonencode({
    properties = {
      field1    = { type = "text" }
      sort_key = { type = "keyword" }
    }
  })

  number_of_shards = 2
  number_of_routing_shards = 2
  codec = "best_compression"
  routing_partition_size = 1
  shard_check_on_startup = "false"
  sort_field = ["sort_key"]
  sort_order = ["asc"]
  auto_expand_replicas =  "0-5"
  search_idle_after = "30s"
  refresh_interval = "10s"
  max_result_window = 5000
  max_inner_result_window = 2000
  max_rescore_window = 1000
  max_docvalue_fields_search = 1500
  max_script_fields = 500
  max_ngram_diff = 100
  max_shingle_diff = 200
  max_refresh_listeners = 10
  analyze_max_token_count = 500000
  highlight_max_analyzed_offset = 1000
  max_terms_count = 10000
  max_regex_length = 1000
  query_default_field = ["field1"]
  routing_allocation_enable = "primaries"
  routing_rebalance_enable = "primaries"
  gc_deletes = "30s"

  analysis_char_filter = jsonencode({
    zero_width_spaces = {
      type     = "mapping"
      mappings = ["\\u200C=>\\u0020"] 
    }
  })
  analysis_filter = jsonencode({
    minimal_english_stemmer = {
      type     = "stemmer"
      language = "minimal_english"
    }
  })
  analysis_analyzer = jsonencode({
    text_en = { 
      type = "custom" 
      tokenizer = "standard"
      char_filter = "zero_width_spaces"
      filter = ["lowercase", "minimal_english_stemmer"]
    }
  })

  settings {
    setting {
      name  = "number_of_replicas"
      value = "2"
    }
  }
}
	`, name)
}

func testAccResourceIndexSettingsMigrationCreate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test_settings_migration" {
  name = "%s"

  settings {
    setting {
      name  = "number_of_replicas"
      value = "2"
    }
  }
}
	`, name)
}

func testAccResourceIndexSettingsMigrationUpdate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test_settings_migration" {
  name = "%s"

  number_of_replicas = 1
}
	`, name)
}

func testAccResourceIndexSettingsConflict(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test_settings_conflict" {
  name = "%s"

  mappings = jsonencode({
    properties = {
      field1    = { type = "text" }
    }
  })

  number_of_shards = 2

  settings {
    setting {
      name  = "number_of_shards"
      value = "3"
    }
  }
}
	`, name)
}

func checkResourceIndexDestroy(s *terraform.State) error {
	client := acctest.Provider.Meta().(*clients.ApiClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_index" {
			continue
		}
		compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)

		res, err := client.GetESClient().Indices.Get([]string{compId.ResourceId})
		if err != nil {
			return err
		}

		if res.StatusCode != 404 {
			return fmt.Errorf("Index (%s) still exists", compId.ResourceId)
		}
	}
	return nil
}
