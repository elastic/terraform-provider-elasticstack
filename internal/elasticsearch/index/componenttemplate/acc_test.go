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
	"context"
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	esclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccResourceComponentTemplate(t *testing.T) {
	templateName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceComponentTemplateDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(templateName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_component_template.test", "name", templateName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_component_template.test", "template.alias.0.name", "my_template_test"),
					testAccCheckResourceAttrIndexSettingsSemantic("elasticstack_elasticsearch_component_template.test", `{"index":{"number_of_shards":"3"}}`),
				),
			},
		},
	})
}

func TestAccResourceComponentTemplateAliasDetails(t *testing.T) {
	templateName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceComponentTemplateDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(templateName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_component_template.test", "name", templateName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_component_template.test", "template.alias.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"elasticstack_elasticsearch_component_template.test",
						"template.alias.*",
						map[string]string{
							"name":           "detailed_alias",
							"is_hidden":      "true",
							"is_write_index": "true",
							"routing":        "shard_1",
							"search_routing": "shard_1",
							"index_routing":  "shard_1",
						},
					),
				),
			},
		},
	})
}

func checkResourceComponentTemplateDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_component_template" {
			continue
		}
		compID, _ := clients.CompositeIDFromStr(rs.Primary.ID)

		typedClient, err := client.GetESClient()
		if err != nil {
			return err
		}
		_, err = typedClient.Cluster.GetComponentTemplate().Name(compID.ResourceID).Do(context.Background())
		if err != nil {
			if esclient.IsNotFoundElasticsearchError(err) {
				continue
			}
			return err
		}

		return fmt.Errorf("Component template (%s) still exists", compID.ResourceID)
	}
	return nil
}

// testAccCheckResourceAttrIndexSettingsSemantic asserts template.settings matches the expected
// effective index settings JSON using the same rules as DiffIndexSettingSuppress /
// IndexSettingsValue.SemanticallyEqual.
func testAccCheckResourceAttrIndexSettingsSemantic(addr, want string) resource.TestCheckFunc {
	const attr = "template.settings"
	return func(s *terraform.State) error {
		ctx := context.Background()
		rs, ok := s.RootModule().Resources[addr]
		if !ok {
			return fmt.Errorf("resource not found: %s", addr)
		}
		got, ok := rs.Primary.Attributes[attr]
		if !ok {
			return fmt.Errorf("%s: attribute %q not found in state", addr, attr)
		}
		a := customtypes.NewIndexSettingsValue(want)
		b := customtypes.NewIndexSettingsValue(got)
		eq, diags := a.SemanticallyEqual(ctx, b)
		if diags.HasError() {
			return fmt.Errorf("%s: %v", addr, diags)
		}
		if !eq {
			return fmt.Errorf("%s: %s = %q, expected semantically equivalent to %q", addr, attr, got, want)
		}
		return nil
	}
}
