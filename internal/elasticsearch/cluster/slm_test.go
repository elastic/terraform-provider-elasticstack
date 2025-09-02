package cluster_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccResourceSLM(t *testing.T) {
	// generate a random policy name
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkSlmDestroy(name),
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccSlmCreate(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "name", name),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "schedule", "0 30 1 * * ?"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "repository", fmt.Sprintf("%s-repo", name)),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "expire_after", "30d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "min_count", "5"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "max_count", "50"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "ignore_unavailable", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "include_global_state", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "indices.0", "data-*"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "indices.1", "abc"),
				),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "ignore_unavailable", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "include_global_state", "false"),
				),
			},
			{
				ResourceName: "elasticstack_elasticsearch_snapshot_lifecycle.test_slm",
				ImportState:  true,
				ImportStateCheck: func(is []*terraform.InstanceState) error {
					importedName := is[0].Attributes["name"]
					if importedName != name {
						return fmt.Errorf("expected imported slm policy name [%s] to equal [%s]", importedName, name)
					}

					return nil
				},
			},
		},
	})
}

func testAccSlmCreate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_snapshot_repository" "repo" {
  name = "%s-repo"

  fs {
    location                  = "/tmp/snapshots"
    compress                  = true
    max_restore_bytes_per_sec = "20mb"
  }
}

resource "elasticstack_elasticsearch_snapshot_lifecycle" "test_slm" {
  name = "%s"

  schedule      = "0 30 1 * * ?"
  snapshot_name = "<daily-snap-{now/d}>"
  repository    = elasticstack_elasticsearch_snapshot_repository.repo.name

  indices              = ["data-*", "abc"]
  ignore_unavailable   = false
  include_global_state = false

  expire_after = "30d"
  min_count    = 5
  max_count    = 50
}
	`, name, name)
}

func checkSlmDestroy(name string) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		client, err := clients.NewAcceptanceTestingClient()
		if err != nil {
			return err
		}

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "elasticstack_elasticsearch_snapshot_lifecycle" {
				compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)
				if compId.ResourceId != name {
					continue
				}
			}

			compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)
			esClient, err := client.GetESClient()
			if err != nil {
				return err
			}
			req := esClient.SlmGetLifecycle.WithPolicyID(compId.ResourceId)
			res, err := esClient.SlmGetLifecycle(req)
			if err != nil {
				return err
			}

			if res.StatusCode != 404 {
				return fmt.Errorf("SLM policy (%s) still exists", compId.ResourceId)
			}
		}
		return nil
	}
}
