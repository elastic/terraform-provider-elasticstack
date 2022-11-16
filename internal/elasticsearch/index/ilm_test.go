package index_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/go-version"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var totalShardsPerNodeVersionLimit = version.Must(version.NewVersion("7.16.0"))

func TestAccResourceILM(t *testing.T) {
	// generate a random policy name
	policyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		CheckDestroy:      checkResourceILMDestroy,
		ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceILMCreate(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.0.min_age", "1h"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.0.set_priority.0.priority", "10"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.0.rollover.0.max_age", "1d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.0.readonly.0.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "delete.0.min_age", "0ms"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "delete.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.#", "0"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "cold.#", "0"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "frozen.#", "0"),
				),
			},
			{
				Config: testAccResourceILMRemoveActions(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.0.min_age", "1h"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.0.set_priority.0.priority", "0"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.0.rollover.0.max_age", "2d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.0.readonly.#", "0"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "delete.#", "0"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.0.min_age", "0ms"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.0.set_priority.0.priority", "60"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.0.readonly.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.0.allocate.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.0.allocate.0.number_of_replicas", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "cold.#", "0"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "frozen.#", "0"),
				),
			},
			{
				SkipFunc: serverVersionLessThanTotalShardsPerNodeLimit,
				Config:   testAccResourceILMTotalShardsPerNode(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.0.min_age", "0ms"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.0.set_priority.0.priority", "60"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.0.readonly.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.0.allocate.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.0.allocate.0.number_of_replicas", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.0.allocate.0.total_shards_per_node", "200"),
				),
			},
		},
	})
}
func serverVersionLessThanTotalShardsPerNodeLimit() (bool, error) {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return false, err
	}
	serverVersion, diags := client.ServerVersion(context.Background())
	if diags.HasError() {
		return false, fmt.Errorf("failed to parse the elasticsearch version %v", diags)
	}

	return serverVersion.LessThan(totalShardsPerNodeVersionLimit), nil
}

func testAccResourceILMCreate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_lifecycle" "test" {
  name = "%s"

  hot {
    min_age = "1h"

    set_priority {
      priority = 10
    }

    rollover {
      max_age = "1d"
    }

    readonly {}
  }

  delete {
    delete {}
  }
}
 `, name)
}

func testAccResourceILMRemoveActions(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_lifecycle" "test" {
  name = "%s"

  hot {
    min_age = "1h"

    set_priority {
      priority = 0
    }

    rollover {
      max_age = "2d"
    }
  }

  warm {
    min_age = "0ms"
    set_priority {
      priority = 60
    }
    readonly {}
    allocate {
      exclude = jsonencode({
        box_type = "hot"
      })
      number_of_replicas = 1
    }
  }

}
 `, name)
}

func testAccResourceILMTotalShardsPerNode(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_lifecycle" "test" {
  name = "%s"

  hot {
    min_age = "1h"

    set_priority {
      priority = 0
    }

    rollover {
      max_age = "2d"
    }
  }

  warm {
    min_age = "0ms"
    set_priority {
      priority = 60
    }
    readonly {}
    allocate {
      exclude = jsonencode({
        box_type = "hot"
      })
      number_of_replicas = 1
      total_shards_per_node = 200
    }
  }

}
 `, name)
}

func checkResourceILMDestroy(s *terraform.State) error {
	client := acctest.Provider.Meta().(*clients.ApiClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_index_lifecycle" {
			continue
		}
		compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)

		req := client.GetESClient().ILM.GetLifecycle.WithPolicy(compId.ResourceId)
		res, err := client.GetESClient().ILM.GetLifecycle(req)
		if err != nil {
			return err
		}

		if res.StatusCode != 404 {
			return fmt.Errorf("ILM policy (%s) still exists", compId.ResourceId)
		}
	}
	return nil
}
