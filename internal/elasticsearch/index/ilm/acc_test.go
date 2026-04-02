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

package ilm_test

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/ilm"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var totalShardsPerNodeVersionLimit = version.Must(version.NewVersion("7.16.0"))
var downsampleNoTimeoutVersionLimit = version.Must(version.NewVersion("8.5.0"))
var downsampleVersionLimit = version.Must(version.NewVersion("8.10.0"))
var allowWriteAfterShrinkVersionLimit = version.Must(version.NewVersion("8.14.0"))

func TestAccResourceILM(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceILMDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.min_age", "1h"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.set_priority.priority", "10"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.rollover.max_age", "1d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.readonly.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "delete.min_age", "0ms"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_lifecycle.test", "hot.min_age"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_lifecycle.test", "delete.min_age"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "delete.delete.delete_searchable_snapshot", "true"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.min_age"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "cold.min_age"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "frozen.min_age"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_lifecycle.test", "id"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_lifecycle.test", "modified_date"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				ImportState:       true,
				ImportStateVerify: true,
				ResourceName:      "elasticstack_elasticsearch_index_lifecycle.test",
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("remove_actions"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.min_age", "1h"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.set_priority.priority", "0"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.rollover.max_age", "2d"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.readonly.enabled"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_lifecycle.test", "hot.min_age"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "delete.min_age"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_lifecycle.test", "warm.min_age"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.min_age", "0ms"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.set_priority.priority", "60"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.readonly.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.allocate.number_of_replicas", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.allocate.exclude", `{"box_type":"hot"}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.shrink.max_primary_shard_size", "50gb"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "cold.min_age"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "frozen.min_age"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_lifecycle.test", "modified_date"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(totalShardsPerNodeVersionLimit),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("total_shards_per_node"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "name", policyName),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_lifecycle.test", "warm.min_age"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.min_age", "0ms"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.set_priority.priority", "60"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.readonly.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.allocate.number_of_replicas", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.allocate.total_shards_per_node", "200"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(downsampleNoTimeoutVersionLimit),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("downsample_no_timeout"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.min_age", "1h"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.set_priority.priority", "10"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.downsample.fixed_interval", "1d"),
					checkILMDownsampleDefaultWaitTimeout("elasticstack_elasticsearch_index_lifecycle.test", "hot.downsample.wait_timeout"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.rollover.max_age", "1d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.readonly.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "delete.min_age", "0ms"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_lifecycle.test", "hot.min_age"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_lifecycle.test", "delete.min_age"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.min_age"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "cold.min_age"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "frozen.min_age"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(downsampleVersionLimit),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("downsample"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.min_age", "1h"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.set_priority.priority", "10"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.downsample.fixed_interval", "1d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.downsample.wait_timeout", "1d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.rollover.max_age", "1d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.readonly.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "delete.min_age", "0ms"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_lifecycle.test", "hot.min_age"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_lifecycle.test", "delete.min_age"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.min_age"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "cold.min_age"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "frozen.min_age"),
				),
			},
		},
	})
}

func TestAccResourceILMRolloverConditions(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceILMDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(ilm.MaxPrimaryShardDocsMinSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("max_primary_shard_docs"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.rollover.max_primary_shard_docs", "5000"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(ilm.RolloverMinConditionsMinSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("rollover_conditions"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.rollover.max_age", "7d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.rollover.max_docs", "10000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.rollover.max_size", "100gb"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.rollover.max_primary_shard_docs", "5000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.rollover.max_primary_shard_size", "50gb"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.rollover.min_age", "3d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.rollover.min_docs", "1000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.rollover.min_size", "50gb"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.rollover.min_primary_shard_docs", "500"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.rollover.min_primary_shard_size", "25gb"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(ilm.RolloverMinConditionsMinSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.rollover.max_age", "14d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.rollover.max_docs", "15000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.rollover.max_size", "150gb"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.rollover.max_primary_shard_docs", "8000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.rollover.max_primary_shard_size", "75gb"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.rollover.min_age", "5d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.rollover.min_docs", "2000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.rollover.min_size", "60gb"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.rollover.min_primary_shard_docs", "750"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.rollover.min_primary_shard_size", "30gb"),
				),
			},
		},
	})
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
		compID, _ := clients.CompositeIDFromStr(rs.Primary.ID)

		esClient, err := client.GetESClient()
		if err != nil {
			return err
		}
		req := esClient.ILM.GetLifecycle.WithPolicy(compID.ResourceID)
		res, err := esClient.ILM.GetLifecycle(req)
		if err != nil {
			return err
		}

		if res.StatusCode != 404 {
			return fmt.Errorf("ILM policy (%s) still exists", compID.ResourceID)
		}
	}
	return nil
}

func TestAccResourceILMMetadata(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceILMDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
					"metadata":    config.StringVariable(`{"managed_by":"terraform"}`),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_meta", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_meta", "metadata", `{"managed_by":"terraform"}`),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_lifecycle.test_meta", "modified_date"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
					"metadata":    config.StringVariable(`{"managed_by":"terraform","version":"2"}`),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_meta", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_meta", "metadata", `{"managed_by":"terraform","version":"2"}`),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_lifecycle.test_meta", "modified_date"),
				),
			},
		},
	})
}

func TestAccResourceILMColdPhase(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceILMDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_cold", "name", policyName),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_lifecycle.test_cold", "cold.min_age"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_cold", "cold.min_age", "30d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_cold", "cold.set_priority.priority", "0"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_cold", "cold.readonly.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_cold", "cold.allocate.number_of_replicas", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_cold", "cold.allocate.include", `{"box_type":"cold"}`),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_lifecycle.test_cold", "modified_date"),
				),
			},
		},
	})
}

func TestAccResourceILMForcemerge(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceILMDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_forcemerge", "name", policyName),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_lifecycle.test_forcemerge", "warm.min_age"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_forcemerge", "warm.forcemerge.max_num_segments", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_forcemerge", "warm.forcemerge.index_codec", "best_compression"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_lifecycle.test_forcemerge", "modified_date"),
				),
			},
		},
	})
}

//go:embed testdata/TestAccResourceILMFromSDK/create/resource.tf
var sdkILMCreateConfig string

func TestAccResourceILMFrozenPhase(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	repositoryName := fmt.Sprintf("%s-repo", policyName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceILMDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name":     config.StringVariable(policyName),
					"repository_name": config.StringVariable(repositoryName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_frozen", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_frozen", "frozen.min_age", "30d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_frozen", "frozen.searchable_snapshot.snapshot_repository", repositoryName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_frozen", "frozen.searchable_snapshot.force_merge_index", "false"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_lifecycle.test_frozen", "modified_date"),
				),
			},
		},
	})
}

func TestAccResourceILMDeleteWaitForSnapshot(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	repositoryName := fmt.Sprintf("%s-repo", policyName)
	slmPolicyName := fmt.Sprintf("%s-slm", policyName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceILMDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name":     config.StringVariable(policyName),
					"repository_name": config.StringVariable(repositoryName),
					"slm_policy_name": config.StringVariable(slmPolicyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_delete_snapshot", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_delete_snapshot", "delete.wait_for_snapshot.policy", slmPolicyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_delete_snapshot", "delete.delete.delete_searchable_snapshot", "false"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_lifecycle.test_delete_snapshot", "modified_date"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("restore_default"),
				ConfigVariables: config.Variables{
					"policy_name":     config.StringVariable(policyName),
					"repository_name": config.StringVariable(repositoryName),
					"slm_policy_name": config.StringVariable(slmPolicyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_delete_snapshot", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_delete_snapshot", "delete.wait_for_snapshot.policy", slmPolicyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_delete_snapshot", "delete.delete.delete_searchable_snapshot", "true"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_lifecycle.test_delete_snapshot", "modified_date"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("remove_wait_for_snapshot"),
				ConfigVariables: config.Variables{
					"policy_name":     config.StringVariable(policyName),
					"repository_name": config.StringVariable(repositoryName),
					"slm_policy_name": config.StringVariable(slmPolicyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_delete_snapshot", "name", policyName),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_delete_snapshot", "delete.wait_for_snapshot.policy"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_delete_snapshot", "delete.delete.delete_searchable_snapshot", "true"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_lifecycle.test_delete_snapshot", "modified_date"),
				),
			},
		},
	})
}

func TestAccResourceILMConnectionOverride(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	endpoint := ilmPrimaryESEndpoint()
	apiKey := os.Getenv("ELASTICSEARCH_API_KEY")
	username := os.Getenv("ELASTICSEARCH_USERNAME")
	password := os.Getenv("ELASTICSEARCH_PASSWORD")

	if username == "" {
		username = "elastic"
	}
	if password == "" {
		password = "password"
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceILMDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
					"endpoints":   config.ListVariable(config.StringVariable(endpoint)),
					"api_key":     config.StringVariable(apiKey),
					"username":    config.StringVariable(username),
					"password":    config.StringVariable(password),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_conn", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_conn", "hot.rollover.max_age", "7d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_conn", "delete.delete.delete_searchable_snapshot", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_conn", "elasticsearch_connection.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_conn", "elasticsearch_connection.0.endpoints.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_conn", "elasticsearch_connection.0.endpoints.0", endpoint),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_conn", "elasticsearch_connection.0.insecure", "true"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_lifecycle.test_conn", "modified_date"),
				),
			},
		},
	})
}

func TestAccResourceILMSearchableSnapshotPhases(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	repositoryName := fmt.Sprintf("%s-repo", policyName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceILMDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name":     config.StringVariable(policyName),
					"repository_name": config.StringVariable(repositoryName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_searchable_snapshot", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_searchable_snapshot", "hot.searchable_snapshot.snapshot_repository", repositoryName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_searchable_snapshot", "hot.searchable_snapshot.force_merge_index", "false"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_searchable_snapshot", "cold.searchable_snapshot.snapshot_repository"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_lifecycle.test_searchable_snapshot", "modified_date"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"policy_name":     config.StringVariable(policyName),
					"repository_name": config.StringVariable(repositoryName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_searchable_snapshot", "name", policyName),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_searchable_snapshot", "hot.searchable_snapshot.snapshot_repository"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_searchable_snapshot", "cold.searchable_snapshot.snapshot_repository", repositoryName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_searchable_snapshot", "cold.searchable_snapshot.force_merge_index", "true"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_lifecycle.test_searchable_snapshot", "modified_date"),
				),
			},
		},
	})
}

func TestAccResourceILMHotActions(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceILMDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(allowWriteAfterShrinkVersionLimit),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("number_of_shards"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_hot_actions", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_hot_actions", "hot.forcemerge.max_num_segments", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_hot_actions", "hot.forcemerge.index_codec", "best_compression"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_hot_actions", "hot.shrink.number_of_shards", "1"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_hot_actions", "hot.shrink.max_primary_shard_size"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_hot_actions", "hot.shrink.allow_write_after_shrink", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_hot_actions", "hot.unfollow.enabled", "true"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_lifecycle.test_hot_actions", "modified_date"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(allowWriteAfterShrinkVersionLimit),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("max_primary_shard_size"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_hot_actions", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_hot_actions", "hot.forcemerge.max_num_segments", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_hot_actions", "hot.forcemerge.index_codec", "best_compression"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_hot_actions", "hot.shrink.number_of_shards"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_hot_actions", "hot.shrink.max_primary_shard_size", "50gb"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_hot_actions", "hot.shrink.allow_write_after_shrink", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_hot_actions", "hot.unfollow.enabled", "true"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_lifecycle.test_hot_actions", "modified_date"),
				),
			},
		},
	})
}

func TestAccResourceILMWarmDownsampleAndShrink(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceILMDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(allowWriteAfterShrinkVersionLimit),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_warm_actions", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_warm_actions", "warm.allocate.number_of_replicas", "2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_warm_actions", "warm.allocate.total_shards_per_node", "5"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_warm_actions", "warm.allocate.include", `{"box_type":"warm"}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_warm_actions", "warm.allocate.require", `{"storage":"fast"}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_warm_actions", "warm.downsample.fixed_interval", "1d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_warm_actions", "warm.downsample.wait_timeout", "12h"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_warm_actions", "warm.shrink.number_of_shards", "1"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_warm_actions", "warm.shrink.max_primary_shard_size"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_warm_actions", "warm.shrink.allow_write_after_shrink", "true"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_lifecycle.test_warm_actions", "modified_date"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(allowWriteAfterShrinkVersionLimit),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_warm_actions", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_warm_actions", "warm.allocate.number_of_replicas", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_warm_actions", "warm.allocate.total_shards_per_node", "-1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_warm_actions", "warm.allocate.exclude", `{"box_type":"hot"}`),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_warm_actions", "warm.allocate.include"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_warm_actions", "warm.allocate.require"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_warm_actions", "warm.downsample.fixed_interval", "2d"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_warm_actions", "warm.shrink.number_of_shards"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_warm_actions", "warm.shrink.max_primary_shard_size", "50gb"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_warm_actions", "warm.shrink.allow_write_after_shrink", "false"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_lifecycle.test_warm_actions", "modified_date"),
				),
			},
		},
	})
}

func TestAccResourceILMColdAllocateAndDownsample(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceILMDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(downsampleVersionLimit),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_cold_actions", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_cold_actions", "cold.allocate.number_of_replicas", "2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_cold_actions", "cold.allocate.total_shards_per_node", "4"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_cold_actions", "cold.allocate.include", `{"box_type":"cold"}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_cold_actions", "cold.allocate.exclude", `{"box_type":"warm"}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_cold_actions", "cold.allocate.require", `{"data":"cold"}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_cold_actions", "cold.downsample.fixed_interval", "1d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_cold_actions", "cold.downsample.wait_timeout", "12h"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_lifecycle.test_cold_actions", "modified_date"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(downsampleVersionLimit),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_cold_actions", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_cold_actions", "cold.allocate.number_of_replicas", "0"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_cold_actions", "cold.allocate.total_shards_per_node", "-1"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_cold_actions", "cold.allocate.include"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_cold_actions", "cold.allocate.exclude"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_cold_actions", "cold.allocate.require"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_cold_actions", "cold.downsample.fixed_interval", "2d"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_lifecycle.test_cold_actions", "modified_date"),
				),
			},
		},
	})
}

func TestAccResourceILMPhaseActionToggles(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceILMDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_toggles", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_toggles", "hot.readonly.enabled", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_toggles", "hot.unfollow.enabled", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_toggles", "warm.readonly.enabled", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_toggles", "warm.migrate.enabled", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_toggles", "warm.unfollow.enabled", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_toggles", "cold.readonly.enabled", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_toggles", "cold.migrate.enabled", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_toggles", "cold.freeze.enabled", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_toggles", "cold.unfollow.enabled", "false"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_toggles", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_toggles", "hot.readonly.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_toggles", "hot.unfollow.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_toggles", "warm.readonly.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_toggles", "warm.migrate.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_toggles", "warm.unfollow.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_toggles", "cold.readonly.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_toggles", "cold.migrate.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_toggles", "cold.freeze.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_toggles", "cold.unfollow.enabled", "true"),
				),
			},
		},
	})
}

func TestAccResourceILMFromSDK(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceILMDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"elasticstack": {
						Source:            "elastic/elasticstack",
						VersionConstraint: "0.14.3",
					},
				},
				Config: sdkILMCreateConfig,
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "name", policyName),
					// SDK provider (<1) uses list-shaped nested blocks in state (hot.0.*).
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.0.min_age", "1h"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.0.set_priority.0.priority", "10"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.min_age", "1h"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.set_priority.priority", "10"),
				),
			},
		},
	})
}

func TestAccResourceILM_importBasic(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceILMDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				ResourceName: "elasticstack_elasticsearch_index_lifecycle.test",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["elasticstack_elasticsearch_index_lifecycle.test"]
					if !ok {
						return "", fmt.Errorf("expected resource not found")
					}
					return rs.Primary.ID, nil
				},
			},
		},
	})
}

func TestAccResourceILM_warmMigrateDisabled(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceILMDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("migrate_warm"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_migrate", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_migrate", "warm.migrate.enabled", "false"),
				),
			},
		},
	})
}

func TestAccResourceILM_shrinkAllowWriteAfterShrink(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceILMDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(allowWriteAfterShrinkVersionLimit),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("shrink_write"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_shrink", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_shrink", "hot.shrink.allow_write_after_shrink", "true"),
				),
			},
		},
	})
}

func TestAccResourceILM_hotReadonlyDisabled(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceILMDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("readonly_disabled"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_ro", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_ro", "hot.readonly.enabled", "false"),
				),
			},
		},
	})
}

func ilmPrimaryESEndpoint() string {
	for endpoint := range strings.SplitSeq(os.Getenv("ELASTICSEARCH_ENDPOINTS"), ",") {
		endpoint = strings.TrimSpace(endpoint)
		if endpoint != "" {
			return endpoint
		}
	}

	return "http://localhost:9200"
}

func checkILMDownsampleDefaultWaitTimeout(resourceName, attribute string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		versionLacksDefaultExposure, err := versionutils.CheckIfVersionIsUnsupported(downsampleVersionLimit)()
		if err != nil {
			return err
		}
		if versionLacksDefaultExposure {
			return nil
		}

		return resource.TestCheckResourceAttr(resourceName, attribute, "1d")(s)
	}
}
