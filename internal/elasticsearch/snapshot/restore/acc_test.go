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

package restore_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	esclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func actionTerraformVersionChecks() []tfversion.TerraformVersionCheck {
	return []tfversion.TerraformVersionCheck{
		tfversion.SkipBelow(tfversion.Version1_14_0),
	}
}

func TestAccActionSnapshotRestore(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	snapshotName := fmt.Sprintf("%s-snap", name)
	restoredIndex := fmt.Sprintf("restored-%s-idx", name)

	resource.Test(t, resource.TestCase{
		PreCheck:               func() { acctest.PreCheck(t) },
		TerraformVersionChecks: actionTerraformVersionChecks(),
		CheckDestroy:           checkSnapshotRestoreDestroy(name, snapshotName, restoredIndex),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("sync/bootstrap"),
				ConfigVariables: config.Variables{
					"name":          config.StringVariable(name),
					"snapshot_name": config.StringVariable(snapshotName),
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("sync/restore"),
				ConfigVariables: config.Variables{
					"name":          config.StringVariable(name),
					"snapshot_name": config.StringVariable(snapshotName),
				},
				Check: resource.ComposeTestCheckFunc(
					checkIndexExists(restoredIndex),
				),
			},
		},
	})
}

func TestAccActionSnapshotRestoreAsync(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	snapshotName := fmt.Sprintf("%s-snap-async", name)
	restoredIndex := fmt.Sprintf("restored-%s-idx", name)

	resource.Test(t, resource.TestCase{
		PreCheck:               func() { acctest.PreCheck(t) },
		TerraformVersionChecks: actionTerraformVersionChecks(),
		CheckDestroy:           checkSnapshotRestoreDestroy(name, snapshotName, restoredIndex),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("async/bootstrap"),
				ConfigVariables: config.Variables{
					"name":          config.StringVariable(name),
					"snapshot_name": config.StringVariable(snapshotName),
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("async/restore"),
				ConfigVariables: config.Variables{
					"name":          config.StringVariable(name),
					"snapshot_name": config.StringVariable(snapshotName),
				},
				Check: resource.ComposeTestCheckFunc(
					waitForRestoredIndex(restoredIndex),
				),
			},
		},
	})
}

func TestAccActionSnapshotRestore_ConflictWithoutRename(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	snapshotName := fmt.Sprintf("%s-snap-conflict", name)

	resource.Test(t, resource.TestCase{
		PreCheck:               func() { acctest.PreCheck(t) },
		TerraformVersionChecks: actionTerraformVersionChecks(),
		CheckDestroy:           checkSnapshotRestoreDestroy(name, snapshotName, ""),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("bootstrap"),
				ConfigVariables: config.Variables{
					"name":          config.StringVariable(name),
					"snapshot_name": config.StringVariable(snapshotName),
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("conflict"),
				ConfigVariables: config.Variables{
					"name":          config.StringVariable(name),
					"snapshot_name": config.StringVariable(snapshotName),
				},
				ExpectError: regexp.MustCompile(`(?i)(index.*already exists|resource_already_exists|cannot restore|snapshot_restore_exception)`),
			},
		},
	})
}

func checkIndexExists(indexName string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			return err
		}

		exists, err := client.GetESClient().Indices.Exists(indexName).Do(context.Background())
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("expected restored index %q to exist", indexName)
		}
		return nil
	}
}

func waitForRestoredIndex(indexName string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			return err
		}

		typedClient := client.GetESClient()
		ctx := context.Background()
		deadline := time.Now().Add(5 * time.Minute)

		for time.Now().Before(deadline) {
			exists, err := typedClient.Indices.Exists(indexName).Do(ctx)
			if err != nil {
				return err
			}
			if exists {
				return nil
			}
			time.Sleep(2 * time.Second)
		}

		return fmt.Errorf("restored index %q did not appear within timeout", indexName)
	}
}

func checkSnapshotRestoreDestroy(name, snapshotName, restoredIndex string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			return err
		}

		repoName := fmt.Sprintf("%s-repo", name)
		indexName := fmt.Sprintf("%s-idx", name)
		typedClient := client.GetESClient()
		ctx := context.Background()

		for _, rs := range s.RootModule().Resources {
			if rs.Type == "elasticstack_elasticsearch_snapshot_repository" {
				_, err := typedClient.Snapshot.DeleteRepository(rs.Primary.Attributes["name"]).Do(ctx)
				if err != nil && !esclient.IsNotFoundElasticsearchError(err) {
					return err
				}
			}
		}

		indices := []string{indexName}
		if restoredIndex != "" {
			indices = append(indices, restoredIndex)
		}
		for _, index := range indices {
			exists, err := typedClient.Indices.Exists(index).Do(ctx)
			if err != nil {
				return err
			}
			if exists {
				_, err = typedClient.Indices.Delete(index).Do(ctx)
				if err != nil {
					return err
				}
			}
		}

		if snapshotName != "" {
			_, err = typedClient.Snapshot.Delete(repoName, snapshotName).Do(ctx)
			if err != nil && !esclient.IsNotFoundElasticsearchError(err) {
				return err
			}
		}

		return nil
	}
}
