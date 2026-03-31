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
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"
)

func TestAccResourceScript(t *testing.T) {
	scriptID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkScriptDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{"script_id": config.StringVariable(scriptID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "script_id", scriptID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "lang", "painless"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "source", "Math.log(_score * 2) + params['my_modifier']"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "context", "score"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory: acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{"script_id": config.StringVariable(scriptID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "script_id", scriptID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "lang", "painless"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "source", "Math.log(_score * 4) + params['changed_modifier']"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "params", `{"changed_modifier":2}`),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				// Ensure the provider doesn't panic if the script has been deleted outside of the Terraform flow
				PreConfig: func() {
					client, err := clients.NewAcceptanceTestingClient()
					require.NoError(t, err)

					esClient, err := client.GetESClient()
					require.NoError(t, err)

					_, err = esClient.DeleteScript(scriptID)
					require.NoError(t, err)
				},
				ConfigDirectory: acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{"script_id": config.StringVariable(scriptID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "script_id", scriptID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "lang", "painless"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "source", "Math.log(_score * 4) + params['changed_modifier']"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "params", `{"changed_modifier":2}`),
				),
			},
		},
	})
}

func TestAccResourceScriptSearchTemplate(t *testing.T) {
	scriptID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkScriptDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{"script_id": config.StringVariable(scriptID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.search_template_test", "script_id", scriptID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.search_template_test", "lang", "mustache"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.search_template_test", "source", `{"from":"{{from}}","query":{"match":{"message":"{{query_string}}"}},"size":"{{size}}"}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.search_template_test", "params", `{"query_string":"My query string"}`),
				),
			},
		},
	})
}

func checkScriptDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_script" {
			continue
		}

		compID, _ := clients.CompositeIDFromStr(rs.Primary.ID)
		esClient, err := client.GetESClient()
		if err != nil {
			return err
		}
		res, err := esClient.GetScript(compID.ResourceID)
		if err != nil {
			return err
		}

		if res.StatusCode != 404 {
			return fmt.Errorf("script (%s) still exists", compID.ResourceID)
		}
	}
	return nil
}
