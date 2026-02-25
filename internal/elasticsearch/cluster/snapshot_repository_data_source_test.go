// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package cluster_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceSnapRepoMissing(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_snapshot_repository" "test_fs_repo" {
  name = "%s"
}`, name),
			},
		},
	})
}

func TestAccDataSourceSnapRepoFs(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
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

func TestAccDataSourceSnapRepoURL(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSnapRepoURL(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_url_repo", "name", name),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_url_repo", "s3.#", "0"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_url_repo", "fs.#", "0"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.0.url", "file:/tmp"),
				),
			},
		},
	})
}

func testAccDataSourceSnapRepoURL(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_snapshot_repository" "test_url_repo" {
  name = "%s"

  url {
    url = "file:/tmp"
  }
}

data "elasticstack_elasticsearch_snapshot_repository" "test_url_repo" {
  name = resource.elasticstack_elasticsearch_snapshot_repository.test_url_repo.name
}
	`, name)
}
