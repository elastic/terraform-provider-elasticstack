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
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/datastreamlifecycle"
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
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_template.test", "id"),
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
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "priority", "0"),
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
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "metadata", `{"description":"initial template","owner":"team-a"}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "template.0.mappings", `{"properties":{"log_level":{"type":"keyword"}}}`),
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
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "metadata", `{"description":"updated template","owner":"team-b"}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "template.0.mappings", `{"properties":{"log_level":{"type":"keyword"},"severity":{"type":"integer"}}}`),
				),
			},
			{
				Config: testAccResourceIndexTemplateWithoutMetadataAndMappings(templateName + "unset"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "name", templateName+"unset"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "index_patterns.#", "2"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_index_template.test", "index_patterns.*", fmt.Sprintf("%sunset-a-*", templateName)),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_index_template.test", "index_patterns.*", fmt.Sprintf("%sunset-b-*", templateName)),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index_template.test", "metadata"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index_template.test", "template.0.mappings"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "version", "0"),
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

func testAccResourceIndexTemplateWithoutMetadataAndMappings(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = "%s"
  index_patterns = ["%s-a-*", "%s-b-*"]
}`, name, name, name)
}

func TestAccResourceIndexTemplateLifecycle(t *testing.T) {
	templateName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexTemplateDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(datastreamlifecycle.MinVersion),
				Config:                   testAccResourceIndexTemplateLifecycle(templateName, "30d"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "template.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "template.0.lifecycle.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"elasticstack_elasticsearch_index_template.test",
						"template.0.lifecycle.*",
						map[string]string{
							"data_retention": "30d",
						},
					),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(datastreamlifecycle.MinVersion),
				Config:                   testAccResourceIndexTemplateLifecycle(templateName, "60d"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "template.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "template.0.lifecycle.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"elasticstack_elasticsearch_index_template.test",
						"template.0.lifecycle.*",
						map[string]string{
							"data_retention": "60d",
						},
					),
				),
			},
		},
	})
}

func testAccResourceIndexTemplateLifecycle(name, dataRetention string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = "%s"
  index_patterns = ["%s-*"]

  data_stream {}

  template {
    lifecycle {
      data_retention = "%s"
    }
  }
}`, name, name, dataRetention)
}

func TestAccResourceIndexTemplateAliasFilter(t *testing.T) {
	templateName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceIndexTemplateDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceIndexTemplateAliasFilter(templateName, "filtered_alias_v1", `{"term":{"status":"active"}}`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "template.0.alias.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"elasticstack_elasticsearch_index_template.test",
						"template.0.alias.*",
						map[string]string{
							"name":   "filtered_alias_v1",
							"filter": `{"term":{"status":"active"}}`,
						},
					),
				),
			},
			{
				Config: testAccResourceIndexTemplateAliasFilter(templateName, "filtered_alias_v2", `{"bool":{"must":[{"term":{"service.name":"api"}},{"term":{"status":"active"}}]}}`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "template.0.alias.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"elasticstack_elasticsearch_index_template.test",
						"template.0.alias.*",
						map[string]string{
							"name":   "filtered_alias_v2",
							"filter": `{"bool":{"must":[{"term":{"service.name":"api"}},{"term":{"status":"active"}}]}}`,
						},
					),
				),
			},
			{
				Config: testAccResourceIndexTemplateAliasFilterRemoved(templateName, "filtered_alias_v3"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "template.0.alias.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"elasticstack_elasticsearch_index_template.test",
						"template.0.alias.*",
						map[string]string{
							"name": "filtered_alias_v3",
						},
					),
					testCheckTemplateAliasAttrCleared("elasticstack_elasticsearch_index_template.test", "filtered_alias_v3", "filter"),
				),
			},
		},
	})
}

func testAccResourceIndexTemplateAliasFilter(name, aliasName, filter string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = "%s"
  index_patterns = ["%s-*"]

  template {
    alias {
      name   = "%s"
      filter = jsonencode(%s)
    }
  }
}`, name, name, aliasName, filter)
}

func testAccResourceIndexTemplateAliasFilterRemoved(name, aliasName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = "%s"
  index_patterns = ["%s-*"]

  template {
    alias {
      name = "%s"
    }
  }
}`, name, name, aliasName)
}

func TestAccResourceIndexTemplateAliasDetails(t *testing.T) {
	templateName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceIndexTemplateDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceIndexTemplateWithAliasDetails(templateName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "template.0.alias.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"elasticstack_elasticsearch_index_template.test",
						"template.0.alias.*",
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

func testAccResourceIndexTemplateWithAliasDetails(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = "%s"
  index_patterns = ["%s-*"]

  template {
    alias {
      name           = "detailed_alias"
      is_hidden      = true
      is_write_index = true
      routing        = "shard_1"
      search_routing = "shard_1"
      index_routing  = "shard_1"
    }
  }
}`, name, name)
}

func TestAccResourceIndexTemplateAliasRoutingFromRoutingOnly(t *testing.T) {
	templateName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceIndexTemplateDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceIndexTemplateWithAliasRoutingOnly(templateName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "template.0.alias.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"elasticstack_elasticsearch_index_template.test",
						"template.0.alias.*",
						map[string]string{
							"name":           "routing_only_alias",
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

func testAccResourceIndexTemplateWithAliasRoutingOnly(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = "%s"
  index_patterns = ["%s-*"]

  template {
    alias {
      name    = "routing_only_alias"
      routing = "shard_1"
    }
  }
}`, name, name)
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

func TestAccResourceIndexTemplateEmptyTemplateBlock(t *testing.T) {
	templateName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceIndexTemplateDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceIndexTemplateWithAliasAndSettings(templateName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "template.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "template.0.alias.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "template.0.settings", `{"index":{"number_of_shards":"2"}}`),
				),
			},
			{
				Config: testAccResourceIndexTemplateWithEmptyTemplateBlock(templateName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "template.#", "1"),
					testCheckAttrEmptyOrAbsent("elasticstack_elasticsearch_index_template.test", "template.0.mappings"),
					testCheckAttrEmptyOrAbsent("elasticstack_elasticsearch_index_template.test", "template.0.settings"),
					testCheckNoTemplateAliases("elasticstack_elasticsearch_index_template.test"),
				),
			},
		},
	})
}

func testAccResourceIndexTemplateWithAliasAndSettings(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = "%s"
  index_patterns = ["%s-*"]

  template {
    alias {
      name = "empty_template_alias"
    }

    settings = jsonencode({
      index = {
        number_of_shards = "2"
      }
    })
  }
}`, name, name)
}

func testAccResourceIndexTemplateWithEmptyTemplateBlock(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = "%s"
  index_patterns = ["%s-*"]

  template {}
}`, name, name)
}

func TestAccResourceIndexTemplateDataStreamEmptyObject(t *testing.T) {
	templateName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexTemplateDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedAllowCustomRoutingVersion),
				Config:                   testAccResourceIndexTemplateDataStreamState(templateName, true, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "data_stream.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "data_stream.0.hidden", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "data_stream.0.allow_custom_routing", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedAllowCustomRoutingVersion),
				Config:                   testAccResourceIndexTemplateDataStreamEmptyObject(templateName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "data_stream.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "data_stream.0.hidden", "false"),
					testCheckDataStreamAttrFalseOrAbsent("elasticstack_elasticsearch_index_template.test", "allow_custom_routing"),
				),
			},
		},
	})
}

func testAccResourceIndexTemplateDataStreamState(name string, hidden, allowCustomRouting bool) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = "%s"
  index_patterns = ["%s-*"]

  data_stream {
    hidden               = %t
    allow_custom_routing = %t
  }
}`, name, name, hidden, allowCustomRouting)
}

func testAccResourceIndexTemplateDataStreamEmptyObject(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = "%s"
  index_patterns = ["%s-*"]

  data_stream {}
}`, name, name)
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

func testCheckTemplateAliasAttrCleared(resourceName, aliasName, attrName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		aliasPrefix, err := templateAliasPrefix(s, resourceName, aliasName)
		if err != nil {
			return err
		}

		value, ok := s.RootModule().Resources[resourceName].Primary.Attributes[aliasPrefix+"."+attrName]
		if ok && value != "" {
			return fmt.Errorf("expected %s.%s to be cleared, got %q", aliasPrefix, attrName, value)
		}
		return nil
	}
}

func testCheckNoTemplateAliases(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found in state: %s", resourceName)
		}

		for key, value := range rs.Primary.Attributes {
			if strings.HasPrefix(key, "template.0.alias.") && strings.HasSuffix(key, ".name") && value != "" {
				return fmt.Errorf("expected no template aliases for %s, found %s=%q", resourceName, key, value)
			}
		}
		return nil
	}
}

func testCheckDataStreamAttrFalseOrAbsent(resourceName, attrName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found in state: %s", resourceName)
		}

		value, ok := rs.Primary.Attributes["data_stream.0."+attrName]
		if ok && value != "false" && value != "" {
			return fmt.Errorf("expected data_stream.0.%s to be false or absent, got %q", attrName, value)
		}
		return nil
	}
}

func testCheckAttrEmptyOrAbsent(resourceName, attrName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found in state: %s", resourceName)
		}

		value, ok := rs.Primary.Attributes[attrName]
		if ok && value != "" {
			return fmt.Errorf("expected %s to be empty or absent, got %q", attrName, value)
		}
		return nil
	}
}

func templateAliasPrefix(s *terraform.State, resourceName, aliasName string) (string, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return "", fmt.Errorf("resource not found in state: %s", resourceName)
	}

	for key, value := range rs.Primary.Attributes {
		if strings.HasPrefix(key, "template.0.alias.") && strings.HasSuffix(key, ".name") && value == aliasName {
			return strings.TrimSuffix(key, ".name"), nil
		}
	}

	return "", fmt.Errorf("alias %q not found for %s", aliasName, resourceName)
}
