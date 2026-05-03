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
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccResourceSLM(t *testing.T) {
	// generate a random policy name
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkSlmDestroy(name),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "name", name),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "schedule", "0 30 1 * * ?"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "repository", fmt.Sprintf("%s-repo", name)),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "expire_after", "30d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "min_count", "5"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "max_count", "50"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "ignore_unavailable", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "include_global_state", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "indices.0", "data-*"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "indices.1", "abc"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				ResourceName:             "elasticstack_elasticsearch_snapshot_lifecycle.test_slm",
				ImportState:              true,
				ImportStateCheck: func(is []*terraform.InstanceState) error {
					importedName := is[0].Attributes["name"]
					if importedName != name {
						return fmt.Errorf("expected imported slm policy name [%s] to equal [%s]", importedName, name)
					}

					return nil
				},
			},
		},
	})
}

func TestAccResourceSLMWithMetadata(t *testing.T) {
	// generate a random policy name
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkSlmDestroy(name),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm_metadata", "name", name),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm_metadata", "schedule", "0 30 1 * * ?"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm_metadata", "repository", fmt.Sprintf("%s-repo", name)),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm_metadata", "expire_after", "30d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm_metadata", "min_count", "5"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm_metadata", "max_count", "50"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm_metadata", "metadata", `{"created_by":"terraform","purpose":"daily backup"}`),
				),
			},
		},
	})
}

func TestAccResourceSLMExtended(t *testing.T) {
	// generate a random policy name
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkSlmDestroy(name),
		Steps: []resource.TestStep{
			{
				// Step 1: create with expand_wildcards, feature_states, partial, metadata
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "name", name),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "schedule", "0 30 1 * * ?"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "snapshot_name", "<daily-snap-{now/d}>"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "repository", fmt.Sprintf("%s-repo", name)),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "expand_wildcards", "closed,hidden"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "indices.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "indices.0", "data-*"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "feature_states.#", "1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "feature_states.*", "kibana"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "partial", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "ignore_unavailable", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "include_global_state", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "expire_after", "30d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "min_count", "5"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "max_count", "50"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "metadata", `{"created_by":"terraform","purpose":"daily backup"}`),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				ResourceName:             "elasticstack_elasticsearch_snapshot_lifecycle.test_slm",
				ImportState:              true,
				ImportStateCheck: func(is []*terraform.InstanceState) error {
					attrs := is[0].Attributes
					if got := attrs["expand_wildcards"]; got != "closed,hidden" {
						return fmt.Errorf("expected imported expand_wildcards [closed,hidden], got [%s]", got)
					}
					if got := attrs["min_count"]; got != "5" {
						return fmt.Errorf("expected imported min_count [5], got [%s]", got)
					}
					if got := attrs["feature_states.#"]; got != "1" {
						return fmt.Errorf("expected imported feature_states count [1], got [%s]", got)
					}
					for attr, value := range attrs {
						if strings.HasPrefix(attr, "feature_states.") && attr != "feature_states.#" && value == "kibana" {
							return nil
						}
					}
					return fmt.Errorf("expected imported feature_states to contain [kibana]")
				},
			},
			{
				// Step 2: update schedule, retention, indices, feature_states, and boolean flags
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "name", name),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "schedule", "0 30 2 * * ?"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "expand_wildcards", "all"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "indices.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "indices.0", "data-*"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "indices.1", "metrics-*"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "feature_states.#", "0"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "partial", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "ignore_unavailable", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "include_global_state", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "expire_after", "60d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "min_count", "3"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "max_count", "30"),
				),
			},
			{
				// Step 3: omit optional attributes with defaults and assert default values are applied
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("defaults"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "name", name),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "snapshot_name", "<snap-{now/d}>"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "expand_wildcards", "open,hidden"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "partial", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "ignore_unavailable", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "include_global_state", "true"),
				),
			},
		},
	})
}

func checkSlmDestroy(name string) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			return err
		}

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "elasticstack_elasticsearch_snapshot_lifecycle" {
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
			req := esClient.SlmGetLifecycle.WithPolicyID(compID.ResourceID)
			res, err := esClient.SlmGetLifecycle(req)
			if err != nil {
				return err
			}

			if res.StatusCode != 404 {
				return fmt.Errorf("SLM policy (%s) still exists", compID.ResourceID)
			}
		}
		return nil
	}
}
