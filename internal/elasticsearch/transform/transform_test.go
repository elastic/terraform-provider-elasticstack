package transform_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var minSupportedDestAliasesVersion = version.Must(version.NewSemver("8.8.0"))

func TestAccResourceTransformWithPivot(t *testing.T) {

	transformNamePivot := sdkacctest.RandStringFromCharSet(18, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceTransformDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceTransformWithPivotCreate(transformNamePivot),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "name", transformNamePivot),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "description", "test description"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "source.0.indices.0", "source_index_for_transform"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "destination.0.index", "dest_index_for_transform"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "frequency", "5m"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "max_page_search_size", "2000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "enabled", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "sync.0.time.0.field", "order_date"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "sync.0.time.0.delay", "20s"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "defer_validation", "true"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "latest"),
				),
			},
			{
				Config: testAccResourceTransformWithPivotUpdate(transformNamePivot),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "name", transformNamePivot),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "description", "yet another test description"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "source.0.indices.0", "source_index_for_transform"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "source.0.indices.1", "additional_index"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "destination.0.index", "dest_index_for_transform_v2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "frequency", "10m"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "max_page_search_size", "2000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "retention_policy.0.time.0.field", "order_date"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "retention_policy.0.time.0.max_age", "7d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "defer_validation", "true"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "latest"),
				),
			},
		},
	})
}

func TestAccResourceTransformWithLatest(t *testing.T) {

	transformNameLatest := sdkacctest.RandStringFromCharSet(20, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceTransformDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceTransformWithLatestCreate(transformNameLatest),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_latest", "name", transformNameLatest),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_latest", "description", "test description (latest)"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_latest", "source.0.indices.0", "source_index_for_transform"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_latest", "destination.0.index", "dest_index_for_transform"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_latest", "frequency", "2m"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_latest", "defer_validation", "true"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_transform.test_latest", "pivot"),
				),
			},
		},
	})
}

func TestAccResourceTransformNoDefer(t *testing.T) {

	transformName := sdkacctest.RandStringFromCharSet(18, sdkacctest.CharSetAlphaNum)
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceTransformDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceTransformNoDeferCreate(transformName, indexName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "name", transformName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "description", "test description"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "source.0.indices.0", indexName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "destination.0.index", "dest_index_for_transform"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "frequency", "5m"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "defer_validation", "false"),
				),
			},
		},
	})
}

func TestAccResourceTransformWithAliases(t *testing.T) {
	transformName := sdkacctest.RandStringFromCharSet(18, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceTransformDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedDestAliasesVersion),
				Config:   testAccResourceTransformWithAliasesCreate(transformName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_aliases", "name", transformName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_aliases", "destination.0.aliases.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_aliases", "destination.0.aliases.0.alias", "test_alias_1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_aliases", "destination.0.aliases.0.move_on_creation", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_aliases", "destination.0.aliases.1.alias", "test_alias_2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_aliases", "destination.0.aliases.1.move_on_creation", "false"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedDestAliasesVersion),
				Config:   testAccResourceTransformWithAliasesUpdate(transformName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_aliases", "name", transformName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_aliases", "destination.0.aliases.#", "3"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_aliases", "destination.0.aliases.0.alias", "test_alias_1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_aliases", "destination.0.aliases.0.move_on_creation", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_aliases", "destination.0.aliases.1.alias", "test_alias_2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_aliases", "destination.0.aliases.1.move_on_creation", "true"),
				),
			},
		},
	})
}

// create a transform referencing a non-existing source index;
// because validations are deferred, this should pass
func testAccResourceTransformWithPivotCreate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_transform" "test_pivot" {
  name = "%s"
  description = "test description"

  source {
    indices = ["source_index_for_transform"]
  }

  destination {
    index = "dest_index_for_transform"
  }

  pivot = jsonencode({
    "group_by": {
      "customer_id": {
        "terms": {
          "field": "customer_id",
          "missing_bucket": true
        }
      }
    },
    "aggregations": {
      "max_price": {
        "max": {
          "field": "taxful_total_price"
        }
      }
    }
  })

  sync {
    time {
      field = "order_date"
      delay = "20s"
    }
  }

  max_page_search_size = 2000
  frequency = "5m"
  enabled = false

  defer_validation = true
  timeout = "1m"
}
  `, name)
}

// update the existing transform, add another source index and start it (enabled = true)
// validations are now unavoidable (at start), so make sure to create the indices _before_ updating the transform
// the tf script below uses implicit dependency, but `depends_on` is also an option
func testAccResourceTransformWithPivotUpdate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test_source_index_1" {
  name = "source_index_for_transform"

  alias {
    name = "test_alias_1"
  }

  mappings = jsonencode({
    properties = {
      field1 = { type = "text" }
    }
  })

  deletion_protection = false
  wait_for_active_shards = "all"
  master_timeout = "1m"
  timeout = "1m"
}

resource "elasticstack_elasticsearch_index" "test_source_index_2" {
  name = "additional_index"

  alias {
    name = "test_alias_2"
  }

  mappings = jsonencode({
    properties = {
      field1 = { type = "text" }
    }
  })

  deletion_protection = false
  wait_for_active_shards = "all"
  master_timeout = "1m"
  timeout = "1m"
}

resource "elasticstack_elasticsearch_transform" "test_pivot" {
  name = "%s"
  description = "yet another test description"

  source {
    indices = [
      elasticstack_elasticsearch_index.test_source_index_1.name,
      elasticstack_elasticsearch_index.test_source_index_2.name
      ]
  }

  destination {
    index = "dest_index_for_transform_v2"
  }

  pivot = jsonencode({
    "group_by": {
      "customer_id": {
        "terms": {
          "field": "customer_id",
          "missing_bucket": true
        }
      }
    },
    "aggregations": {
      "max_price": {
        "max": {
          "field": "taxful_total_price"
        }
      }
    }
  })

  sync {
    time {
      field = "order_date"
      delay = "20s"
    }
  }

  retention_policy {
    time {
      field   = "order_date"
      max_age = "7d"
    }
  }

  max_page_search_size = 2000
  frequency = "10m"
  enabled = true

  defer_validation = true
  timeout = "1m"
}
  `, name)
}

func testAccResourceTransformWithLatestCreate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_transform" "test_latest" {
  name = "%s"
  description = "test description (latest)"

  source {
    indices = ["source_index_for_transform"]
  }

  destination {
    index = "dest_index_for_transform"
  }

  latest = jsonencode({
    "unique_key": ["customer_id"],
    "sort": "order_date"
  })
  frequency = "2m"
  enabled = false

  defer_validation = true
  timeout = "1m"
}
  `, name)
}

func testAccResourceTransformNoDeferCreate(transformName, indexName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test_index" {
  name = "%s"

  alias {
    name = "test_alias_1"
  }

  mappings = jsonencode({
    properties = {
      field1 = { type = "text" }
    }
  })

  deletion_protection = false
  wait_for_active_shards = "all"
  master_timeout = "1m"
  timeout = "1m"
}

resource "elasticstack_elasticsearch_transform" "test_pivot" {
  name = "%s"
  description = "test description"

  source {
    indices = [elasticstack_elasticsearch_index.test_index.name]
  }

  destination {
    index = "dest_index_for_transform"
  }

  pivot = jsonencode({
    "group_by": {
      "customer_id": {
        "terms": {
          "field": "customer_id",
          "missing_bucket": true
        }
      }
    },
    "aggregations": {
      "max_price": {
        "max": {
          "field": "taxful_total_price"
        }
      }
    }
  })
  frequency = "5m"
  enabled = false

  defer_validation = false
  timeout = "1m"
}
  `, indexName, transformName)
}

func testAccResourceTransformWithAliasesCreate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_transform" "test_aliases" {
  name = "%s"
  description = "test transform with aliases"

  source {
    indices = ["source_index"]
  }

  destination {
    index = "dest_index_for_transform"

    aliases {
      alias = "test_alias_1"
      move_on_creation = true
    }

    aliases {
      alias = "test_alias_2"
      move_on_creation = false
    }
  }

  pivot = jsonencode({
    "group_by": {
      "customer_id": {
        "terms": {
          "field": "customer_id"
        }
      }
    },
    "aggregations": {
      "total_sales": {
        "sum": {
          "field": "sales"
        }
      }
    }
  })

  defer_validation = true
}
`, name)
}

func testAccResourceTransformWithAliasesUpdate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_transform" "test_aliases" {
  name = "%s"
  description = "test transform with aliases"

  source {
    indices = ["source_index"]
  }

  destination {
    index = "dest_index_for_transform"

    aliases {
      alias = "test_alias_1"
      move_on_creation = false
    }

    aliases {
      alias = "test_alias_2"
      move_on_creation = true
    }

	aliases {
      alias = "test_alias_3"
	}
  }

  pivot = jsonencode({
    "group_by": {
      "customer_id": {
        "terms": {
          "field": "customer_id"
        }
      }
    },
    "aggregations": {
      "total_sales": {
        "sum": {
          "field": "sales"
        }
      }
    }
  })

  defer_validation = true
}
`, name)
}

func checkResourceTransformDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_transform" {
			continue
		}
		compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)

		esClient, err := client.GetESClient()
		if err != nil {
			return err
		}
		req := esClient.TransformGetTransform.WithTransformID(compId.ResourceId)
		res, err := esClient.TransformGetTransform(req)
		if err != nil {
			return err
		}

		if res.StatusCode != 404 {
			return fmt.Errorf("transform (%s) still exists", compId.ResourceId)
		}
	}
	return nil
}
