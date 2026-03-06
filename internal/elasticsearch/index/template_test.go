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
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var minSupportedAllowCustomRoutingVersion = version.Must(version.NewVersion("8.0.0"))

func TestAccResourceIndexTemplate(t *testing.T) {
	// generate random template name
	templateName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	templateNameComponent := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexTemplateDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckTypeSetElemAttr(
						"elasticstack_elasticsearch_index_template.test",
						"index_patterns.*",
						fmt.Sprintf("%s-logs-*", templateName),
					),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "priority", "42"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "template.0.alias.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"elasticstack_elasticsearch_index_template.test",
						"template.0.alias.*",
						map[string]string{"name": "my_template_test"},
					),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "template.0.settings", `{"index":{"number_of_shards":"3"}}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test2", "name", fmt.Sprintf("%s-stream", templateName)),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test2", "data_stream.0.hidden", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckTypeSetElemAttr(
						"elasticstack_elasticsearch_index_template.test",
						"index_patterns.*",
						fmt.Sprintf("%s-logs-*", templateName),
					),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "template.0.alias.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"elasticstack_elasticsearch_index_template.test",
						"template.0.alias.*",
						map[string]string{"name": "my_template_test"},
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						"elasticstack_elasticsearch_index_template.test",
						"template.0.alias.*",
						map[string]string{"name": "alias2"},
					),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "template.0.settings", `{"index":{"number_of_shards":"3"}}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test2", "name", fmt.Sprintf("%s-stream", templateName)),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test2", "data_stream.0.hidden", "false"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(index.MinSupportedIgnoreMissingComponentTemplateVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_ignore_component"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateNameComponent),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test_component", "name", templateNameComponent),
					resource.TestCheckTypeSetElemAttr(
						"elasticstack_elasticsearch_index_template.test_component",
						"index_patterns.*",
						fmt.Sprintf("%s-logscomponent-*", templateNameComponent),
					),
					resource.TestCheckTypeSetElemAttr(
						"elasticstack_elasticsearch_index_template.test_component",
						"composed_of.*",
						fmt.Sprintf("%s-logscomponent@custom", templateNameComponent),
					),
					resource.TestCheckTypeSetElemAttr(
						"elasticstack_elasticsearch_index_template.test_component",
						"ignore_missing_component_templates.*",
						fmt.Sprintf("%s-logscomponent@custom", templateNameComponent),
					),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(index.MinSupportedIgnoreMissingComponentTemplateVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_ignore_component"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateNameComponent),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test_component", "name", templateNameComponent),
					resource.TestCheckTypeSetElemAttr(
						"elasticstack_elasticsearch_index_template.test_component",
						"index_patterns.*",
						fmt.Sprintf("%s-logscomponent-*", templateNameComponent),
					),
					resource.TestCheckTypeSetElemAttr(
						"elasticstack_elasticsearch_index_template.test_component",
						"composed_of.*",
						fmt.Sprintf("%s-logscomponent-updated@custom", templateNameComponent),
					),
					resource.TestCheckTypeSetElemAttr(
						"elasticstack_elasticsearch_index_template.test_component",
						"ignore_missing_component_templates.*",
						fmt.Sprintf("%s-logscomponent-updated@custom", templateNameComponent),
					),
				),
			},
		},
	})
}

func checkResourceIndexTemplateDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_index_template" {
			continue
		}
		compID, _ := clients.CompositeIDFromStr(rs.Primary.ID)

		esClient, err := client.GetESClient()
		if err != nil {
			return err
		}
		req := esClient.Indices.GetIndexTemplate.WithName(compID.ResourceID)
		res, err := esClient.Indices.GetIndexTemplate(req)
		if err != nil {
			return err
		}

		if res.StatusCode != 404 {
			return fmt.Errorf("Index template (%s) still exists", compID.ResourceID)
		}
	}
	return nil
}

func TestAccResourceIndexTemplateMetadataAndMappings(t *testing.T) {
	templateName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceIndexTemplateDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceIndexTemplateWithMetadataAndMappings(templateName, `{"description":"initial template","owner":"team-a"}`, `{"properties":{"log_level":{"type":"keyword"}}}`, 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "version", "1"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_template.test", "metadata"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_template.test", "template.0.mappings"),
				),
			},
			{
				Config: testAccResourceIndexTemplateWithMetadataAndMappings(
					templateName,
					`{"description":"updated template","owner":"team-b"}`,
					`{"properties":{"log_level":{"type":"keyword"},"severity":{"type":"integer"}}}`,
					2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "version", "2"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_template.test", "metadata"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_template.test", "template.0.mappings"),
				),
			},
		},
	})
}

func testAccResourceIndexTemplateWithMetadataAndMappings(name, metadata, mappings string, ver int) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = "%s"
  index_patterns = ["%s-*"]
  version        = %d
  metadata       = jsonencode(%s)

  template {
    mappings = jsonencode(%s)
  }
}`, name, name, ver, metadata, mappings)
}

func TestAccResourceIndexTemplateDataStreamCustomRouting(t *testing.T) {
	templateName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexTemplateDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedAllowCustomRoutingVersion),
				Config:                   testAccResourceIndexTemplateDataStreamCustomRouting(templateName, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "data_stream.0.allow_custom_routing", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedAllowCustomRoutingVersion),
				Config:                   testAccResourceIndexTemplateDataStreamCustomRouting(templateName, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "data_stream.0.allow_custom_routing", "false"),
				),
			},
		},
	})
}

func testAccResourceIndexTemplateDataStreamCustomRouting(name string, allowCustomRouting bool) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = "%s"
  index_patterns = ["%s-*"]

  data_stream {
    allow_custom_routing = %t
  }
}`, name, name, allowCustomRouting)
}

func TestAccResourceIndexTemplateEmptyCollections(t *testing.T) {
	templateName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexTemplateDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(index.MinSupportedIgnoreMissingComponentTemplateVersion),
				Config:                   testAccResourceIndexTemplateWithCollections(templateName, templateName+"-comp@custom", templateName+"-comp@custom"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_index_template.test", "composed_of.*", templateName+"-comp@custom"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_index_template.test", "ignore_missing_component_templates.*", templateName+"-comp@custom"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(index.MinSupportedIgnoreMissingComponentTemplateVersion),
				Config:                   testAccResourceIndexTemplateWithEmptyCollections(templateName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "composed_of.#", "0"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "ignore_missing_component_templates.#", "0"),
				),
			},
		},
	})
}

func testAccResourceIndexTemplateWithCollections(name, composedOf, ignoreMissing string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = "%s"
  index_patterns = ["%s-*"]

  composed_of                         = ["%s"]
  ignore_missing_component_templates  = ["%s"]
}`, name, name, composedOf, ignoreMissing)
}

func testAccResourceIndexTemplateWithEmptyCollections(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = "%s"
  index_patterns = ["%s-*"]

  composed_of                         = []
  ignore_missing_component_templates  = []
}`, name, name)
}
