package transform_test

import (
	//"context"
	"fmt"
	//"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	//"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/transform"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceTransform(t *testing.T) {
	transformName := sdkacctest.RandStringFromCharSet(18, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceTransformDestroy,
		ProtoV5ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceTransformCreate(transformName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test", "name", transformName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test", "description", "test description"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test", "source.0.indices.0", "source_index_for_transform"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test", "destination.0.index", "dest_index_for_transform"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test", "frequency", "5m"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test", "pivot.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_transform.test", "latest.#", "0"),
				),
			},
			// {
			// 	Config: testAccResourceTransformUpdate(transformName),
			// 	Check:  resource.ComposeTestCheckFunc(),
			// },
		},
	})
}

func testAccResourceTransformCreate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_transform" "test" {
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

	defer_validation = true
	timeout = "1m"
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
			return fmt.Errorf("Transform (%s) still exists", compId.ResourceId)
		}
	}
	return nil
}
