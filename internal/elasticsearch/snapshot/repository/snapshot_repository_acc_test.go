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

package repository_test

import (
	"context"
	_ "embed"
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	esclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

//go:embed testdata/TestAccResourceSnapRepoMigration/main.tf
var snapRepoMigrationConfig string

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
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "verify", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.location", "/tmp"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.compress", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.max_restore_bytes_per_sec", "10mb"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.readonly", "false"),
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
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_url_repo", "verify", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.url", "file:/tmp"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.compress", "true"),
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
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.location", "/tmp"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.compress", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.chunk_size", "1gb"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.max_snapshot_bytes_per_sec", "20mb"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.max_restore_bytes_per_sec", "10mb"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.readonly", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.max_number_of_snapshots", "100"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "name", name),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "verify", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.compress", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.chunk_size", "500mb"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.max_snapshot_bytes_per_sec", "40mb"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.max_restore_bytes_per_sec", "20mb"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.readonly", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.max_number_of_snapshots", "50"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("replace"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							"elasticstack_elasticsearch_snapshot_repository.test_fs_repo",
							plancheck.ResourceActionReplace,
						),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_fs_repo", "fs.location", "/tmp/replace"),
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
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.url", "file:/tmp"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.http_max_retries", "3"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.http_socket_timeout", "30s"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.compress", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.max_snapshot_bytes_per_sec", "5mb"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.max_restore_bytes_per_sec", "10mb"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.readonly", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.max_number_of_snapshots", "100"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_url_repo", "name", name),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_url_repo", "verify", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.http_max_retries", "7"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.http_socket_timeout", "45s"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.compress", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.max_snapshot_bytes_per_sec", "40mb"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.max_restore_bytes_per_sec", "20mb"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.readonly", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.max_number_of_snapshots", "50"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
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
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("replace"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							"elasticstack_elasticsearch_snapshot_repository.test_url_repo",
							plancheck.ResourceActionReplace,
						),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_repository.test_url_repo", "url.url", "file:/tmp/replace"),
				),
			},
		},
	})
}

func TestAccResourceSnapRepoS3(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "elasticstack_elasticsearch_snapshot_repository.test_s3_repo"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkRepoDestroy(name),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "verify", "false"),
					resource.TestCheckResourceAttr(resourceName, "s3.bucket", "test-bucket"),
					resource.TestCheckResourceAttr(resourceName, "s3.endpoint", "https://minio.example.com:9000"),
					resource.TestCheckResourceAttr(resourceName, "s3.path_style_access", "true"),
					resource.TestCheckResourceAttr(resourceName, "s3.client", "default"),
					resource.TestCheckResourceAttr(resourceName, "s3.canned_acl", "private"),
					resource.TestCheckResourceAttr(resourceName, "s3.storage_class", "standard"),
					resource.TestCheckResourceAttr(resourceName, "s3.server_side_encryption", "false"),
					resource.TestCheckResourceAttr(resourceName, "s3.base_path", "snapshots"),
					resource.TestCheckResourceAttr(resourceName, "s3.buffer_size", "5mb"),
					resource.TestCheckResourceAttr(resourceName, "s3.chunk_size", "1gb"),
					resource.TestCheckResourceAttr(resourceName, "s3.compress", "false"),
					resource.TestCheckResourceAttr(resourceName, "s3.max_snapshot_bytes_per_sec", "20mb"),
					resource.TestCheckResourceAttr(resourceName, "s3.max_restore_bytes_per_sec", "10mb"),
					resource.TestCheckResourceAttr(resourceName, "s3.readonly", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       false,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "s3.endpoint", "https://minio-alt.example.com:9000"),
					resource.TestCheckResourceAttr(resourceName, "s3.path_style_access", "false"),
					resource.TestCheckResourceAttr(resourceName, "s3.client", "secondary"),
					resource.TestCheckResourceAttr(resourceName, "s3.canned_acl", "public-read"),
					resource.TestCheckResourceAttr(resourceName, "s3.storage_class", "reduced_redundancy"),
					resource.TestCheckResourceAttr(resourceName, "s3.server_side_encryption", "true"),
					resource.TestCheckResourceAttr(resourceName, "s3.base_path", "snapshots/v2"),
					resource.TestCheckResourceAttr(resourceName, "s3.buffer_size", "10mb"),
					resource.TestCheckResourceAttr(resourceName, "s3.chunk_size", "500mb"),
					resource.TestCheckResourceAttr(resourceName, "s3.compress", "true"),
					resource.TestCheckResourceAttr(resourceName, "s3.max_snapshot_bytes_per_sec", "40mb"),
					resource.TestCheckResourceAttr(resourceName, "s3.max_restore_bytes_per_sec", "20mb"),
					resource.TestCheckResourceAttr(resourceName, "s3.readonly", "false"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				ResourceName:             resourceName,
				ImportState:              true,
				ImportStateCheck: func(is []*terraform.InstanceState) error {
					importedName := is[0].Attributes["name"]
					if importedName != name {
						return fmt.Errorf("expected imported snapshot name [%s] to equal [%s]", importedName, name)
					}
					return nil
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("replace"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							resourceName,
							plancheck.ResourceActionReplace,
						),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "s3.bucket", "test-bucket-replaced"),
				),
			},
		},
	})
}

func TestAccResourceSnapRepoGcs(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "elasticstack_elasticsearch_snapshot_repository.test_gcs_repo"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkRepoDestroy(name),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "verify", "false"),
					resource.TestCheckResourceAttr(resourceName, "gcs.bucket", "test-gcs-bucket"),
					resource.TestCheckResourceAttr(resourceName, "gcs.client", "default"),
					resource.TestCheckResourceAttr(resourceName, "gcs.base_path", "snapshots"),
					resource.TestCheckResourceAttr(resourceName, "gcs.compress", "false"),
					resource.TestCheckResourceAttr(resourceName, "gcs.chunk_size", "1gb"),
					resource.TestCheckResourceAttr(resourceName, "gcs.max_snapshot_bytes_per_sec", "20mb"),
					resource.TestCheckResourceAttr(resourceName, "gcs.max_restore_bytes_per_sec", "10mb"),
					resource.TestCheckResourceAttr(resourceName, "gcs.readonly", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "verify", "false"),
					resource.TestCheckResourceAttr(resourceName, "gcs.bucket", "test-gcs-bucket"),
					resource.TestCheckResourceAttr(resourceName, "gcs.client", "secondary"),
					resource.TestCheckResourceAttr(resourceName, "gcs.base_path", "snapshots/v2"),
					resource.TestCheckResourceAttr(resourceName, "gcs.compress", "true"),
					resource.TestCheckResourceAttr(resourceName, "gcs.chunk_size", "500mb"),
					resource.TestCheckResourceAttr(resourceName, "gcs.max_snapshot_bytes_per_sec", "40mb"),
					resource.TestCheckResourceAttr(resourceName, "gcs.max_restore_bytes_per_sec", "20mb"),
					resource.TestCheckResourceAttr(resourceName, "gcs.readonly", "false"),
				),
			},
		},
	})
}

func TestAccResourceSnapRepoAzure(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "elasticstack_elasticsearch_snapshot_repository.test_azure_repo"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkRepoDestroy(name),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "verify", "false"),
					resource.TestCheckResourceAttr(resourceName, "azure.container", "test-azure-container"),
					resource.TestCheckResourceAttr(resourceName, "azure.client", "default"),
					resource.TestCheckResourceAttr(resourceName, "azure.base_path", "snapshots"),
					resource.TestCheckResourceAttr(resourceName, "azure.location_mode", "primary_only"),
					resource.TestCheckResourceAttr(resourceName, "azure.compress", "false"),
					resource.TestCheckResourceAttr(resourceName, "azure.chunk_size", "1gb"),
					resource.TestCheckResourceAttr(resourceName, "azure.max_snapshot_bytes_per_sec", "20mb"),
					resource.TestCheckResourceAttr(resourceName, "azure.max_restore_bytes_per_sec", "10mb"),
					resource.TestCheckResourceAttr(resourceName, "azure.readonly", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "verify", "false"),
					resource.TestCheckResourceAttr(resourceName, "azure.container", "test-azure-container"),
					resource.TestCheckResourceAttr(resourceName, "azure.client", "secondary"),
					resource.TestCheckResourceAttr(resourceName, "azure.base_path", "snapshots/v2"),
					resource.TestCheckResourceAttr(resourceName, "azure.location_mode", "secondary_only"),
					resource.TestCheckResourceAttr(resourceName, "azure.compress", "true"),
					resource.TestCheckResourceAttr(resourceName, "azure.chunk_size", "500mb"),
					resource.TestCheckResourceAttr(resourceName, "azure.max_snapshot_bytes_per_sec", "40mb"),
					resource.TestCheckResourceAttr(resourceName, "azure.max_restore_bytes_per_sec", "20mb"),
					resource.TestCheckResourceAttr(resourceName, "azure.readonly", "false"),
				),
			},
		},
	})
}

// hdfs is intentionally not covered by an acceptance test: it requires the
// repository-hdfs plugin, which is not bundled with Elasticsearch and is not
// installed on the standard acceptance-test cluster.

func TestAccResourceSnapRepoNegativeExactlyOneOf(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				ExpectError:              regexp.MustCompile(`attributes specified when one \(and only one\) of`),
			},
		},
	})
}

func TestAccResourceSnapRepoNegativeAlsoRequires(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				ExpectError:              regexp.MustCompile(`must be specified when`),
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

			typedClient := client.GetESClient()
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

// TestAccResourceSnapRepoFsIssue3709 verifies that explicitly setting empty
// strings for optional fs block settings no longer triggers an "inconsistent
// result after apply" error (issue #3709). The read path preserves the prior
// state value when the API omits the key, so the post-apply state matches the
// user's configured empty strings.
func TestAccResourceSnapRepoFsIssue3709(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "elasticstack_elasticsearch_snapshot_repository.test_fs_repo"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkRepoDestroy(name),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "fs.location", "/tmp"),
					resource.TestCheckResourceAttr(resourceName, "fs.compress", "true"),
					resource.TestCheckResourceAttr(resourceName, "fs.chunk_size", ""),
					resource.TestCheckResourceAttr(resourceName, "fs.max_restore_bytes_per_sec", ""),
					resource.TestCheckResourceAttr(resourceName, "fs.max_snapshot_bytes_per_sec", ""),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       false,
			},
		},
	})
}

func TestAccResourceSnapRepoMigration(t *testing.T) {
	name := "repo-migration-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "elasticstack_elasticsearch_snapshot_repository.test_migration"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"elasticstack": {
						Source:            "elastic/elasticstack",
						VersionConstraint: "0.14.5",
					},
				},
				Config: snapRepoMigrationConfig,
				ConfigVariables: config.Variables{
					"name": config.StringVariable(name),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "fs.0.location", "/tmp"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory(""),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(name),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "fs.location", "/tmp"),
				),
			},
		},
	})
}
