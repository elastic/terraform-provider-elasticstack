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
var shrinkAllowWriteVersionLimit = version.Must(version.NewVersion("8.14.0"))

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
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.0.min_age", "1h"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.0.set_priority.0.priority", "10"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.0.rollover.0.max_age", "1d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.0.readonly.0.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "delete.0.min_age", "0ms"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "delete.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "delete.0.delete.0.delete_searchable_snapshot", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.#", "0"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "cold.#", "0"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "frozen.#", "0"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_lifecycle.test", "modified_date"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("remove_actions"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.0.min_age", "1h"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.0.set_priority.0.priority", "0"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.0.rollover.0.max_age", "2d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.0.readonly.#", "0"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "delete.#", "0"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.0.min_age", "0ms"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.0.set_priority.0.priority", "60"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.0.readonly.0.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.0.allocate.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.0.allocate.0.number_of_replicas", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.0.allocate.0.exclude", `{"box_type":"hot"}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.0.shrink.0.max_primary_shard_size", "50gb"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "cold.#", "0"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "frozen.#", "0"),
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
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.0.min_age", "0ms"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.0.set_priority.0.priority", "60"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.0.readonly.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.0.allocate.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.0.allocate.0.number_of_replicas", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.0.allocate.0.total_shards_per_node", "200"),
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
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.0.min_age", "1h"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.0.set_priority.0.priority", "10"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.0.downsample.0.fixed_interval", "1d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.0.rollover.0.max_age", "1d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.0.readonly.0.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "delete.0.min_age", "0ms"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "delete.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.#", "0"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "cold.#", "0"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "frozen.#", "0"),
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
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.0.min_age", "1h"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.0.set_priority.0.priority", "10"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.0.downsample.0.fixed_interval", "1d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.0.downsample.0.wait_timeout", "1d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.0.rollover.0.max_age", "1d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.0.readonly.0.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "delete.0.min_age", "0ms"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "delete.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "warm.#", "0"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "cold.#", "0"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "frozen.#", "0"),
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
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.0.rollover.0.max_primary_shard_docs", "5000"),
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
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.0.rollover.0.max_age", "7d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.0.rollover.0.max_docs", "10000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.0.rollover.0.max_size", "100gb"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.0.rollover.0.max_primary_shard_docs", "5000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.0.rollover.0.max_primary_shard_size", "50gb"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.0.rollover.0.min_age", "3d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.0.rollover.0.min_docs", "1000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.0.rollover.0.min_size", "50gb"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.0.rollover.0.min_primary_shard_docs", "500"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_rollover", "hot.0.rollover.0.min_primary_shard_size", "25gb"),
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
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_cold", "cold.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_cold", "cold.0.min_age", "30d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_cold", "cold.0.set_priority.0.priority", "0"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_cold", "cold.0.readonly.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_cold", "cold.0.readonly.0.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_cold", "cold.0.allocate.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_cold", "cold.0.allocate.0.number_of_replicas", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_cold", "cold.0.allocate.0.include", `{"box_type":"cold"}`),
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
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_forcemerge", "warm.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_forcemerge", "warm.0.forcemerge.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_forcemerge", "warm.0.forcemerge.0.max_num_segments", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_forcemerge", "warm.0.forcemerge.0.index_codec", "best_compression"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_lifecycle.test_forcemerge", "modified_date"),
				),
			},
		},
	})
}

//go:embed testdata/TestAccResourceILMFromSDK/create/resource.tf
var sdkILMCreateConfig string

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
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.0.min_age", "1h"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "hot.0.set_priority.0.priority", "10"),
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
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_migrate", "warm.0.migrate.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_migrate", "warm.0.migrate.0.enabled", "false"),
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
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(shrinkAllowWriteVersionLimit),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("shrink_write"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_shrink", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_shrink", "hot.0.shrink.0.allow_write_after_shrink", "true"),
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
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_ro", "hot.0.readonly.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test_ro", "hot.0.readonly.0.enabled", "false"),
				),
			},
		},
	})
}
