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

package componenttemplate_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccReproduceIssue3124ComponentTemplate is the regression test for
// https://github.com/elastic/terraform-provider-elasticstack/issues/3124 as it
// applies to component templates.
//
// Component templates suffer from the same class of bug as index templates: the
// read path decoded the response through *types.ClusterComponentTemplate, whose
// nested IndexSettings struct (and its hand-coded UnmarshalJSON) silently drops
// any settings sub-key it does not model (e.g. index.search.slowlog.include)
// and coerces string-encoded scalars such as
// index.lifecycle.parse_origination_date "true" into a typed bool.
//
// The provider now decodes component template responses through
// internal/models.ComponentTemplate (settings as map[string]any) so unmodeled
// fields and string-encoded scalars survive the refresh.
func TestAccReproduceIssue3124ComponentTemplate(t *testing.T) {
	templateName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	addr := "elasticstack_elasticsearch_component_template.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceComponentTemplateDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				Config:                   testAccReproduceIssue3124ComponentTemplateConfig(templateName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceAttrIndexSettingsSemantic(addr, `{
						"index": {
							"number_of_shards": "1",
							"number_of_replicas": "0",
							"search": {
								"slowlog": {
									"include": {
										"user": "true"
									},
									"threshold": {
										"query": {
											"warn": "10s"
										}
									}
								}
							},
							"lifecycle": {
								"parse_origination_date": "true"
							}
						}
					}`),
				),
			},
			{
				// Refresh must not introduce drift: a second plan must be empty.
				ProtoV6ProviderFactories: acctest.Providers,
				Config:                   testAccReproduceIssue3124ComponentTemplateConfig(templateName),
				PlanOnly:                 true,
			},
		},
	})
}

func testAccReproduceIssue3124ComponentTemplateConfig(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_component_template" "test" {
  name = %q

  template {
    settings = jsonencode({
      index = {
        number_of_shards   = "1"
        number_of_replicas = "0"
        search = {
          slowlog = {
            include = {
              user = "true"
            }
            threshold = {
              query = {
                warn = "10s"
              }
            }
          }
        }
        lifecycle = {
          parse_origination_date = "true"
        }
      }
    })
  }
}
`, name)
}
