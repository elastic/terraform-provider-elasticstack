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
	"context"
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	esclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccResourceSnapRepoFs(t *testing.T) {
	// generate a random policy name
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkRepoDestroy(name),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "name", name),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.0.location", "/tmp"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.0.compress", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.0.max_restore_bytes_per_sec", "10mb"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				ResourceName:             "elasticstack_elasticsearch_snapshot_repository.test_fs_repo",
				ImportState:              true,
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
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkRepoDestroy(name),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_url_repo", "name", name),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.0.url", "file:/tmp"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.0.compress", "true"),
				),
			},
		},
	})
}

func TestAccResourceSnapRepoFsExtended(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkRepoDestroy(name),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "name", name),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "verify", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.0.location", "/tmp"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.0.compress", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.0.chunk_size", "1gb"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.0.max_snapshot_bytes_per_sec", "20mb"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.0.max_restore_bytes_per_sec", "10mb"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.0.readonly", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.0.max_number_of_snapshots", "100"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "name", name),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "verify", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.0.compress", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.0.chunk_size", "500mb"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.0.max_snapshot_bytes_per_sec", "40mb"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.0.max_restore_bytes_per_sec", "20mb"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.0.readonly", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.0.max_number_of_snapshots", "50"),
				),
			},
		},
	})
}

func TestAccResourceSnapRepoURLExtended(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkRepoDestroy(name),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_url_repo", "name", name),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_url_repo", "verify", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.0.url", "file:/tmp"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.0.http_max_retries", "3"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.0.http_socket_timeout", "30s"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.0.compress", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.0.max_restore_bytes_per_sec", "10mb"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.0.max_number_of_snapshots", "100"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				ResourceName:             "elasticstack_elasticsearch_snapshot_repository.test_url_repo",
				ImportState:              true,
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

func checkRepoDestroy(name string) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			return err
		}

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "elasticstack_elasticsearch_snapshot_repository" {
				continue
			}

			compID, diags := clients.CompositeIDFromStr(rs.Primary.ID)
			if diags.HasError() {
				return fmt.Errorf("failed to parse snapshot repository composite ID %q: %v", rs.Primary.ID, diags)
			}
			if compID.ResourceID != name {
				continue
			}

			typedClient, err := client.GetESClient()
			if err != nil {
				return err
			}
			res, err := typedClient.Snapshot.GetRepository().Repository(compID.ResourceID).Do(context.Background())
			if err != nil {
				if esclient.IsNotFoundElasticsearchError(err) {
					continue
				}
				return err
			}

			if _, ok := res[compID.ResourceID]; ok {
				return fmt.Errorf("Snapshot repository (%s) still exists", compID.ResourceID)
			}
		}
		return nil
	}
}
