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
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccResourceSLM(t *testing.T) {
	// generate a random policy name
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkSlmDestroy(name),
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{"name": config.StringVariable(name)},
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
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "ignore_unavailable", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm", "include_global_state", "false"),
				),
			},
			{
				ResourceName: "elasticstack_elasticsearch_snapshot_lifecycle.test_slm",
				ImportState:  true,
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
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkSlmDestroy(name),
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{"name": config.StringVariable(name)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm_metadata", "name", name),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm_metadata", "schedule", "0 30 1 * * ?"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm_metadata", "repository", fmt.Sprintf("%s-repo", name)),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm_metadata", "expire_after", "30d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm_metadata", "min_count", "5"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_snapshot_lifecycle.test_slm_metadata", "max_count", "50"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_snapshot_lifecycle.test_slm_metadata", "metadata"),
				),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_snapshot_lifecycle.test_slm_metadata", "metadata"),
				),
			},
		},
	})
}

func checkSlmDestroy(name string) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		client, err := clients.NewAcceptanceTestingClient()
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
