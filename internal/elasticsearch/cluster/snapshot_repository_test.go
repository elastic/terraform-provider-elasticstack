package cluster_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceSnapRepoFs(t *testing.T) {
	// generate a random policy name
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		CheckDestroy:      checkRepoDestroy(name),
		ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccRepoFsCreate(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "name", name),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.0.location", "/tmp"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.0.compress", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.0.max_restore_bytes_per_sec", "10mb"),
				),
			},
			{
				ResourceName: "elasticstack_elasticsearch_snapshot_repository.test_fs_repo",
				ImportState:  true,
				ImportStateCheck: func(is []*terraform.InstanceState) error {
					importedName := is[0].Attributes["name"]
					if importedName != name {
						return fmt.Errorf("expected imported snapshot name [%s] to equal [%s]", importedName, name)
					}

					return nil
				},
			},
		},
	})
}

func TestAccResourceSnapRepoUrl(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		CheckDestroy:      checkRepoDestroy(name),
		ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccRepoUrlCreate(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_url_repo", "name", name),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.0.url", "https://example.com/repo"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.0.compress", "true"),
				),
			},
		},
	})
}

func testAccRepoFsCreate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_snapshot_repository" "test_fs_repo" {
  name = "%s"

  fs {
    location                  = "/tmp"
    compress                  = true
    max_restore_bytes_per_sec = "10mb"
  }
}
	`, name)
}

func testAccRepoUrlCreate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_snapshot_repository" "test_url_repo" {
  name = "%s"

  url {
    url = "https://example.com/repo"
  }
}
	`, name)
}

func checkRepoDestroy(name string) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		client := acctest.Provider.Meta().(*clients.ApiClient)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "elasticstack_elasticsearch_snapshot_repository" {
				compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)
				if compId.ResourceId != name {
					continue
				}
			}

			compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)
			req := client.GetESClient().Snapshot.GetRepository.WithRepository(compId.ResourceId)
			res, err := client.GetESClient().Snapshot.GetRepository(req)
			if err != nil {
				return err
			}

			if res.StatusCode != 404 {
				return fmt.Errorf("Snapshot repository (%s) still exists", compId.ResourceId)
			}
		}
		return nil
	}
}
