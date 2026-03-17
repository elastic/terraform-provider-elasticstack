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

package exportsavedobjects_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceKibanaExportSavedObjects(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceKibanaExportSavedObjectsConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_export_saved_objects.test", "id"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_export_saved_objects.test", "exported_objects"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_export_saved_objects.test", "space_id", "default"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_export_saved_objects.test", "exclude_export_details", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_export_saved_objects.test", "include_references_deep", "true"),
				),
			},
		},
	})
}

const testAccDataSourceKibanaExportSavedObjectsConfig = `
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_action_connector" "test" {
	name              = "test-export-connector"
	connector_type_id = ".slack"
	secrets = jsonencode({
	  webhookUrl = "https://example.com"
	})
}

data "elasticstack_kibana_export_saved_objects" "test" {
  space_id = "default"
  exclude_export_details = true
  include_references_deep = true
 objects = [
    {
      type = "action",
      id = elasticstack_kibana_action_connector.test.connector_id
    }
  ]
}
`
