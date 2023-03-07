package transform_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceTransformWithPivot(t *testing.T) {

	transformNamePivot := sdkacctest.RandStringFromCharSet(18, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceTransformDestroy,
		ProtoV5ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceTransformWithPivotCreate(transformNamePivot),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "name", transformNamePivot),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "description", "test description"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "source.0.indices.0", "source_index_for_transform"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "destination.0.index", "dest_index_for_transform"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "frequency", "5m"),
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
		ProtoV5ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceTransformWithLatestCreate(transformNameLatest),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_latest", "name", transformNameLatest),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_latest", "description", "test description (latest)"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_latest", "source.0.indices.0", "source_index_for_transform"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_latest", "destination.0.index", "dest_index_for_transform"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_latest", "frequency", "2m"),
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
		ProtoV5ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceTransformNoDeferCreate(transformName, indexName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "name", transformName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "description", "test description"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "source.0.indices.0", indexName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "destination.0.index", "dest_index_for_transform"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test_pivot", "frequency", "5m"),
				),
			},
		},
	})
}

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
  frequency = "5m"
	enabled = false

	defer_validation = true
	timeout = "1m"
}
	`, name)
}

func testAccResourceTransformWithPivotUpdate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_transform" "test_pivot" {
  name = "%s"
	description = "yet another test description"

	source {
		indices = ["source_index_for_transform", "additional_index"]
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

  settings {
    setting {
      name  = "index.number_of_replicas"
      value = "2"
    }
  }

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
			return fmt.Errorf("Transform (%s) still exists", compId.ResourceId)
		}
	}
	return nil
}
