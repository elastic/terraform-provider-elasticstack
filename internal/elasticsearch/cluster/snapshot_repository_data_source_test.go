package cluster_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceSnapRepoFs(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSnapRepoFs(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "name", name),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "gcs.#", "0"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "url.#", "0"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.0.location", "/tmp"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.0.compress", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.0.max_restore_bytes_per_sec", "10mb"),
				),
			},
		},
	})
}

func testAccDataSourceSnapRepoFs(name string) string {
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

data "elasticstack_elasticsearch_snapshot_repository" "test_fs_repo" {
  name = resource.elasticstack_elasticsearch_snapshot_repository.test_fs_repo.name
}
	`, name)
}

func TestAccDataSourceSnapRepoUrl(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSnapRepoUrl(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_url_repo", "name", name),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_url_repo", "s3.#", "0"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_url_repo", "fs.#", "0"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.0.url", "https://example.com/repo"),
				),
			},
		},
	})
}

func testAccDataSourceSnapRepoUrl(name string) string {
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

data "elasticstack_elasticsearch_snapshot_repository" "test_url_repo" {
  name = resource.elasticstack_elasticsearch_snapshot_repository.test_url_repo.name
}
	`, name)
}
