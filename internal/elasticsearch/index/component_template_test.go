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

package index_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccResourceComponentTemplate(t *testing.T) {
	// generate a random username
	templateName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceComponentTemplateDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceComponentTemplateCreate(templateName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_component_template.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_component_template.test", "name", templateName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_component_template.test", "metadata", `{"env":"test"}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_component_template.test", "version", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_component_template.test", "template.0.mappings", `{"properties":{"field1":{"type":"keyword"}}}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_component_template.test", "template.0.settings", `{"index":{"number_of_shards":"3"}}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_component_template.test", "template.0.alias.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_component_template.test", "template.0.alias.*", map[string]string{
						"name":           "my_template_test",
						"filter":         `{"term":{"user":"kimchy"}}`,
						"index_routing":  "ir1",
						"search_routing": "sr1",
						"routing":        "r1",
						"is_hidden":      "false",
						"is_write_index": "true",
					}),
				),
			},
			{
				Config: testAccResourceComponentTemplateUpdate(templateName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_component_template.test", "metadata", `{"env":"production"}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_component_template.test", "version", "2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_component_template.test", "template.0.mappings", `{"properties":{"field1":{"type":"text"}}}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_component_template.test", "template.0.settings", `{"index":{"number_of_shards":"1"}}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_component_template.test", "template.0.alias.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_component_template.test", "template.0.alias.*", map[string]string{
						"name":           "my_template_test",
						"filter":         `{"term":{"user":"elastic"}}`,
						"index_routing":  "ir2",
						"search_routing": "sr2",
						"routing":        "r2",
						"is_hidden":      "true",
						"is_write_index": "false",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_component_template.test", "template.0.alias.*", map[string]string{
						"name":           "second_alias",
						"filter":         "",
						"index_routing":  "",
						"search_routing": "",
						"routing":        "",
						"is_hidden":      "false",
						"is_write_index": "false",
					}),
				),
			},
			{
				ResourceName:      "elasticstack_elasticsearch_component_template.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceComponentTemplateCreate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_component_template" "test" {
  name     = "%s"
  metadata = jsonencode({
    env = "test"
  })
  version = 1

  template {
    alias {
      name = "my_template_test"
      filter = jsonencode({
        term = {
          user = "kimchy"
        }
      })
      index_routing  = "ir1"
      search_routing = "sr1"
      routing        = "r1"
      is_hidden      = false
      is_write_index = true
    }

    mappings = jsonencode({
      properties = {
        field1 = {
          type = "keyword"
        }
      }
    })

    settings = jsonencode({
      number_of_shards = "3"
    })
  }
}`, name)
}

func testAccResourceComponentTemplateUpdate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_component_template" "test" {
  name     = "%s"
  metadata = jsonencode({
    env = "production"
  })
  version = 2

  template {
    alias {
      name = "my_template_test"
      filter = jsonencode({
        term = {
          user = "elastic"
        }
      })
      index_routing  = "ir2"
      search_routing = "sr2"
      routing        = "r2"
      is_hidden      = true
      is_write_index = false
    }

    alias {
      name = "second_alias"
    }

    mappings = jsonencode({
      properties = {
        field1 = {
          type = "text"
        }
      }
    })

    settings = jsonencode({
      number_of_shards = "1"
    })
  }
}`, name)
}

func checkResourceComponentTemplateDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_component_template" {
			continue
		}
		compID, _ := clients.CompositeIDFromStr(rs.Primary.ID)

		esClient, err := client.GetESClient()
		if err != nil {
			return err
		}
		req := esClient.Cluster.GetComponentTemplate.WithName(compID.ResourceID)
		res, err := esClient.Cluster.GetComponentTemplate(req)
		if err != nil {
			return err
		}

		if res.StatusCode != 404 {
			return fmt.Errorf("Component template (%s) still exists", compID.ResourceID)
		}
	}
	return nil
}
