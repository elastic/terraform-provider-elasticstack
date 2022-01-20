package index_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceIndex(t *testing.T) {
	// generate renadom index name
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.UnitTest(t, resource.TestCase{
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
