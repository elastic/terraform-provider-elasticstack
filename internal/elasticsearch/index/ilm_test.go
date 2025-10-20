package index_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var totalShardsPerNodeVersionLimit = version.Must(version.NewVersion("7.16.0"))
var downsampleNoTimeoutVersionLimit = version.Must(version.NewVersion("8.5.0"))
var downsampleVersionLimit = version.Must(version.NewVersion("8.10.0"))

func TestAccResourceILM(t *testing.T) {
	// generate a random policy name
	policyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceILMDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
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
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(totalShardsPerNodeVersionLimit),
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
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(downsampleNoTimeoutVersionLimit),
				Config:   testAccResourceILMDownsampleNoTimeout(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.0.min_age", "1h"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.0.set_priority.0.priority", "10"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.0.downsample.0.fixed_interval", "1d"),
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
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(downsampleVersionLimit),
				Config:   testAccResourceILMDownsample(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.0.min_age", "1h"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.0.set_priority.0.priority", "10"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.0.downsample.0.fixed_interval", "1d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.0.downsample.0.wait_timeout", "1d"),
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
		},
	})
}

func TestAccResourceILMRolloverConditions(t *testing.T) {
	// generate a random policy name
	policyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceILMDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(index.MaxPrimaryShardDocsMinSupportedVersion),
				Config:   testAccResourceILMCreateWithMaxPrimaryShardDocs(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.0.rollover.0.max_primary_shard_docs", "5000"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(index.RolloverMinConditionsMinSupportedVersion),
				Config:   testAccResourceILMCreateWithRolloverConditions(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.0.rollover.0.max_age", "7d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.0.rollover.0.max_docs", "10000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.0.rollover.0.max_size", "100gb"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.0.rollover.0.max_primary_shard_docs", "5000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.0.rollover.0.max_primary_shard_size", "50gb"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.0.rollover.0.min_age", "3d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.0.rollover.0.min_docs", "1000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.0.rollover.0.min_size", "50gb"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.0.rollover.0.min_primary_shard_docs", "500"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.0.rollover.0.min_primary_shard_size", "25gb"),
				),
			},
		},
	})
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
    shrink {
      max_primary_shard_size = "50gb"
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

func testAccResourceILMCreateWithRolloverConditions(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_lifecycle" "test_rollover" {
  name = "%s"

  hot {
    rollover {
      max_age = "7d"
      max_docs = 10000
      max_size = "100gb"
      max_primary_shard_docs = 5000
      max_primary_shard_size = "50gb"
      min_age = "3d"
      min_docs = 1000
      min_size = "50gb"
      min_primary_shard_docs = 500
      min_primary_shard_size = "25gb"
    }

    readonly {}
  }

  delete {
    delete {}
  }
}
 `, name)
}

func testAccResourceILMCreateWithMaxPrimaryShardDocs(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_lifecycle" "test_rollover" {
  name = "%s"

  hot {
    rollover {
      max_primary_shard_docs = 5000
    }

    readonly {}
  }

  delete {
    delete {}
  }
}
 `, name)
}

func testAccResourceILMDownsampleNoTimeout(name string) string {
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

    downsample {
      fixed_interval = "1d"
    }

    readonly {}
  }

  delete {
    delete {}
  }
}
 `, name)
}

func testAccResourceILMDownsample(name string) string {
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

    downsample {
      fixed_interval = "1d"
      wait_timeout = "1d"
    }

    readonly {}
  }

  delete {
    delete {}
  }
}
 `, name)
}

func checkResourceILMDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_index_lifecycle" {
			continue
		}
		compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)

		esClient, err := client.GetESClient()
		if err != nil {
			return err
		}
		req := esClient.ILM.GetLifecycle.WithPolicy(compId.ResourceId)
		res, err := esClient.ILM.GetLifecycle(req)
		if err != nil {
			return err
		}

		if res.StatusCode != 404 {
			return fmt.Errorf("ILM policy (%s) still exists", compId.ResourceId)
		}
	}
	return nil
}
