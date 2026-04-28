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

package kibana_test

import (
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceKibanaConnector(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_kibana_action_connector.myconnector", "name", "myconnector"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_action_connector.myconnector", "space_id", "default"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_action_connector.myconnector", "connector_type_id", ".slack"),
					resource.TestCheckResourceAttrPair(
						"data.elasticstack_kibana_action_connector.myconnector", "connector_id",
						"elasticstack_kibana_action_connector.slack", "connector_id",
					),
					resource.TestMatchResourceAttr("data.elasticstack_kibana_action_connector.myconnector", "config", regexp.MustCompile(`^(\{\})?$`)),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_action_connector.myconnector", "is_deprecated", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_action_connector.myconnector", "is_missing_secrets", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_action_connector.myconnector", "is_preconfigured", "false"),
				),
			},
		},
	})
}

func TestAccDataSourceKibanaConnector_customSpace(t *testing.T) {
	spaceID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	connectorName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"space_id":       config.StringVariable(spaceID),
					"connector_name": config.StringVariable(connectorName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_kibana_action_connector.test", "space_id", spaceID),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_action_connector.test", "connector_type_id", ".index"),
					resource.TestCheckResourceAttrPair(
						"data.elasticstack_kibana_action_connector.test", "connector_id",
						"elasticstack_kibana_action_connector.test", "connector_id",
					),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_action_connector.test", "is_deprecated", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_action_connector.test", "is_missing_secrets", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_action_connector.test", "is_preconfigured", "false"),
				),
			},
		},
	})
}

func TestAccDataSourceKibanaConnector_duplicateName(t *testing.T) {
	connectorName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("duplicate"),
				ConfigVariables: config.Variables{
					"connector_name": config.StringVariable(connectorName),
				},
				ExpectError: regexp.MustCompile(`multiple connectors found`),
			},
		},
	})
}

func TestAccDataSourceKibanaConnector_kibanaConnection(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckWithExplicitKibanaEndpoint(t)
		},
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("kibana_connection"),
				ConfigVariables:          acctest.KibanaConnectionVariables(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_kibana_action_connector.test", "name", "kbconn_connector"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_action_connector.test", "connector_type_id", ".slack"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_action_connector.test", "connector_id"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_action_connector.test", "kibana_connection.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_action_connector.test", "kibana_connection.0.endpoints.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_action_connector.test", "kibana_connection.0.insecure", "false"),
				),
			},
		},
	})
}
