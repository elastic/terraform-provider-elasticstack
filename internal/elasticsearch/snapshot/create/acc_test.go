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

package create_test

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
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

func TestAccActionSnapshotCreate(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	snapshotName := fmt.Sprintf("%s-snap", name)

	resource.Test(t, resource.TestCase{
		PreCheck:               func() { acctest.PreCheck(t) },
		TerraformVersionChecks: actionTerraformVersionChecks(),
		CheckDestroy:           checkSnapshotCreateDestroy(name, snapshotName),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("sync"),
				ConfigVariables: config.Variables{
					"name":          config.StringVariable(name),
					"snapshot_name": config.StringVariable(snapshotName),
				},
				Check: resource.ComposeTestCheckFunc(
					checkSnapshotExists(name, snapshotName, false),
				),
			},
		},
	})
}

func TestAccActionSnapshotCreateWithMetadata(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	snapshotName := fmt.Sprintf("%s-snap-meta", name)

	resource.Test(t, resource.TestCase{
		PreCheck:               func() { acctest.PreCheck(t) },
		TerraformVersionChecks: actionTerraformVersionChecks(),
		CheckDestroy:           checkSnapshotCreateDestroy(name, snapshotName),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("sync"),
				ConfigVariables: config.Variables{
					"name":          config.StringVariable(name),
					"snapshot_name": config.StringVariable(snapshotName),
				},
				Check: resource.ComposeTestCheckFunc(
					checkSnapshotExists(name, snapshotName, true),
				),
			},
		},
	})
}

func TestAccActionSnapshotCreateAsync(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	snapshotName := fmt.Sprintf("%s-snap-async", name)

	resource.Test(t, resource.TestCase{
		PreCheck:               func() { acctest.PreCheck(t) },
		TerraformVersionChecks: actionTerraformVersionChecks(),
		CheckDestroy:           checkSnapshotCreateDestroy(name, snapshotName),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("async"),
				ConfigVariables: config.Variables{
					"name":          config.StringVariable(name),
					"snapshot_name": config.StringVariable(snapshotName),
				},
				Check: resource.ComposeTestCheckFunc(
					waitForSnapshotSuccess(name, snapshotName),
				),
			},
		},
	})
}

func TestAccActionSnapshotCreate_DuplicateName(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	snapshotName := fmt.Sprintf("%s-snap-dup", name)

	resource.Test(t, resource.TestCase{
		PreCheck:               func() { acctest.PreCheck(t) },
		TerraformVersionChecks: actionTerraformVersionChecks(),
		CheckDestroy:           checkSnapshotCreateDestroy(name, snapshotName),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("bootstrap"),
				ConfigVariables: config.Variables{
					"name":          config.StringVariable(name),
					"snapshot_name": config.StringVariable(snapshotName),
				},
				Check: resource.ComposeTestCheckFunc(
					checkSnapshotExists(name, snapshotName, false),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("duplicate"),
				ConfigVariables: config.Variables{
					"name":          config.StringVariable(name),
					"snapshot_name": config.StringVariable(snapshotName),
				},
				ExpectError: regexp.MustCompile(`(?i)(snapshot.*already exists|already exists)`),
			},
		},
	})
}

func checkSnapshotExists(name, snapshotName string, expectMetadata bool) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			return err
		}

		repoName := fmt.Sprintf("%s-repo", name)
		typedClient := client.GetESClient()
		resp, err := typedClient.Snapshot.Get(repoName, snapshotName).Do(context.Background())
		if err != nil {
			return err
		}

		var snapshotInfo *types.SnapshotInfo
		for i := range resp.Snapshots {
			if resp.Snapshots[i].Snapshot == snapshotName {
				snapshotInfo = &resp.Snapshots[i]
				break
			}
		}
		if snapshotInfo == nil {
			return fmt.Errorf("snapshot %q not found in repository %q", snapshotName, repoName)
		}

		if expectMetadata {
			meta := snapshotInfo.Metadata
			if meta == nil {
				return fmt.Errorf("expected metadata on snapshot %q", snapshotName)
			}
			createdBy, ok := meta["created_by"]
			if !ok {
				return fmt.Errorf("expected created_by metadata on snapshot %q", snapshotName)
			}
			var createdByValue string
			if err := json.Unmarshal(createdBy, &createdByValue); err != nil {
				return fmt.Errorf("failed to decode created_by metadata: %w", err)
			}
			if createdByValue != "terraform" {
				return fmt.Errorf("expected created_by=terraform, got %q", createdByValue)
			}
		}

		return nil
	}
}

func waitForSnapshotSuccess(name, snapshotName string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			return err
		}

		repoName := fmt.Sprintf("%s-repo", name)
		typedClient := client.GetESClient()
		ctx := context.Background()
		deadline := time.Now().Add(5 * time.Minute)

		for time.Now().Before(deadline) {
			statusResp, err := typedClient.Snapshot.Status().Repository(repoName).Snapshot(snapshotName).Do(ctx)
			if err != nil {
				return err
			}

			for _, snap := range statusResp.Snapshots {
				if snap.Snapshot == snapshotName && snap.State == "SUCCESS" {
					return nil
				}
			}

			time.Sleep(2 * time.Second)
		}

		return fmt.Errorf("snapshot %q did not reach SUCCESS within timeout", snapshotName)
	}
}

func checkSnapshotCreateDestroy(name, snapshotName string) func(*terraform.State) error {
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

		exists, err := typedClient.Indices.Exists(indexName).Do(ctx)
		if err != nil {
			return err
		}
		if exists {
			_, err = typedClient.Indices.Delete(indexName).Do(ctx)
			if err != nil {
				return err
			}
		}

		restoredName := fmt.Sprintf("restored-%s-idx", name)
		exists, err = typedClient.Indices.Exists(restoredName).Do(ctx)
		if err != nil {
			return err
		}
		if exists {
			_, err = typedClient.Indices.Delete(restoredName).Do(ctx)
			if err != nil {
				return err
			}
		}

		_, err = typedClient.Snapshot.Delete(repoName, snapshotName).Do(ctx)
		if err != nil && !esclient.IsNotFoundElasticsearchError(err) {
			return err
		}

		return nil
	}
}
