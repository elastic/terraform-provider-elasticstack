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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceKibanaConnector(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConnector,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_kibana_action_connector.myconnector", "name", "myconnector"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_action_connector.myconnector", "space_id", "default"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_action_connector.myconnector", "connector_type_id", ".slack"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_action_connector.myconnector", "connector_id"),
				),
			},
		},
	})
}

const testAccDataSourceConnector = `
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_action_connector" "slack" {
	name              = "myconnector"
	connector_type_id = ".slack"
	secrets = jsonencode({
	  webhookUrl = "https://internet.com"
	})
  }

data "elasticstack_kibana_action_connector" "myconnector" {
	name    = elasticstack_kibana_action_connector.slack.name
}
`
