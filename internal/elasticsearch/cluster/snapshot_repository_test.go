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
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccResourceSnapRepoFs(t *testing.T) {
	// generate a random policy name
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkRepoDestroy(name),
		ProtoV6ProviderFactories: acctest.Providers,
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

func TestAccResourceSnapRepoURL(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkRepoDestroy(name),
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccRepoURLCreate(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_url_repo", "name", name),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.0.url", "file:/tmp"),
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

func testAccRepoURLCreate(name string) string {
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
	`, name)
}

func checkRepoDestroy(name string) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		client, err := clients.NewAcceptanceTestingClient()
		if err != nil {
			return err
		}

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "elasticstack_elasticsearch_snapshot_repository" {
				compID, _ := clients.CompositeIDFromStr(rs.Primary.ID)
				if compID.ResourceID != name {
					continue
				}
			}

			compID, _ := clients.CompositeIDFromStr(rs.Primary.ID)
			esClient, err := client.GetESClient()
			if err != nil {
				return err
			}
			req := esClient.Snapshot.GetRepository.WithRepository(compID.ResourceID)
			res, err := esClient.Snapshot.GetRepository(req)
			if err != nil {
				return err
			}

			if res.StatusCode != 404 {
				return fmt.Errorf("Snapshot repository (%s) still exists", compID.ResourceID)
			}
		}
		return nil
	}
}
