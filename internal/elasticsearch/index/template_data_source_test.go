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
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/datastreamlifecycle"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIndexTemplateDataSource(t *testing.T) {
	// generate a random role name
	templateName := "test-template-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	templateNameComponent := "test-template-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckTypeSetElemAttr(
						"data.elasticstack_elasticsearch_index_template.test",
						"index_patterns.*",
						fmt.Sprintf("tf-acc-%s-*", templateName),
					),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "priority", "100"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(index.MinSupportedIgnoreMissingComponentTemplateVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("ignore_component"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateNameComponent),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_component", "name", templateNameComponent),
					resource.TestCheckTypeSetElemAttr(
						"data.elasticstack_elasticsearch_index_template.test_component",
						"index_patterns.*",
						fmt.Sprintf("tf-acc-component-%s-*", templateNameComponent),
					),
					resource.TestCheckTypeSetElemAttr(
						"data.elasticstack_elasticsearch_index_template.test_component",
						"composed_of.*",
						fmt.Sprintf("%s-logscomponent@custom", templateNameComponent),
					),
					resource.TestCheckTypeSetElemAttr(
						"data.elasticstack_elasticsearch_index_template.test_component",
						"ignore_missing_component_templates.*",
						fmt.Sprintf("%s-logscomponent@custom", templateNameComponent),
					),
				),
			},
		},
	})
}

// TestAccIndexTemplateDataSourceTemplate covers the template subtree:
// aliases (name, filter, routing, index_routing, search_routing, is_hidden, is_write_index), mappings, and settings.
func TestAccIndexTemplateDataSourceTemplate(t *testing.T) {
	templateName := "test-ds-tpl-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccIndexTemplateDataSourceTemplateConfig(
					templateName,
					`{"properties":{"log_level":{"type":"keyword"}}}`,
					`{"index":{"number_of_shards":"1"}}`,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "template.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "template.0.alias.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.elasticstack_elasticsearch_index_template.test",
						"template.0.alias.*",
						map[string]string{
							"name":           "my_alias",
							"filter":         `{"term":{"status":"active"}}`,
							"index_routing":  "shard_1",
							"routing":        "shard_1",
							"search_routing": "shard_1",
							"is_hidden":      "false",
							"is_write_index": "true",
						},
					),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "template.0.mappings", `{"properties":{"log_level":{"type":"keyword"}}}`),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "template.0.settings", `{"index":{"number_of_shards":"1"}}`),
				),
			},
			{
				Config: testAccIndexTemplateDataSourceTemplateConfig(
					templateName,
					`{"properties":{"log_level":{"type":"keyword"},"severity":{"type":"integer"}}}`,
					`{"index":{"number_of_shards":"2"}}`,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "template.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "template.0.alias.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.elasticstack_elasticsearch_index_template.test",
						"template.0.alias.*",
						map[string]string{
							"name":           "my_alias",
							"filter":         `{"term":{"status":"active"}}`,
							"index_routing":  "shard_1",
							"routing":        "shard_1",
							"search_routing": "shard_1",
							"is_hidden":      "false",
							"is_write_index": "true",
						},
					),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "template.0.mappings", `{"properties":{"log_level":{"type":"keyword"},"severity":{"type":"integer"}}}`),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "template.0.settings", `{"index":{"number_of_shards":"2"}}`),
				),
			},
		},
	})
}

func testAccIndexTemplateDataSourceTemplateConfig(name, mappings, settings string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = "%[1]s"
  index_patterns = ["%[1]s-*"]

  template {
    alias {
      name           = "my_alias"
      filter         = jsonencode({ term = { status = "active" } })
      index_routing  = "shard_1"
      routing        = "shard_1"
      search_routing = "shard_1"
      is_hidden      = false
      is_write_index = true
    }

    mappings = jsonencode(%[2]s)
    settings = jsonencode(%[3]s)
  }
}

data "elasticstack_elasticsearch_index_template" "test" {
  name = elasticstack_elasticsearch_index_template.test.name
}
`, name, mappings, settings)
}

// TestAccIndexTemplateDataSourceDataStream covers data_stream.0.hidden and data_stream.0.allow_custom_routing.
func TestAccIndexTemplateDataSourceDataStream(t *testing.T) {
	templateName := "test-ds-stream-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccIndexTemplateDataSourceDataStreamConfig(templateName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "data_stream.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "data_stream.0.hidden", "true"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedAllowCustomRoutingVersion),
				Config:   testAccIndexTemplateDataSourceDataStreamWithCustomRoutingConfig(templateName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "data_stream.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "data_stream.0.hidden", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "data_stream.0.allow_custom_routing", "true"),
				),
			},
		},
	})
}

func testAccIndexTemplateDataSourceDataStreamConfig(name string, hidden bool) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = "%[1]s"
  index_patterns = ["%[1]s-*"]

  data_stream {
    hidden = %[2]t
  }
}

data "elasticstack_elasticsearch_index_template" "test" {
  name = elasticstack_elasticsearch_index_template.test.name
}
`, name, hidden)
}

func testAccIndexTemplateDataSourceDataStreamWithCustomRoutingConfig(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = "%[1]s"
  index_patterns = ["%[1]s-*"]

  data_stream {
    hidden               = false
    allow_custom_routing = true
  }
}

data "elasticstack_elasticsearch_index_template" "test" {
  name = elasticstack_elasticsearch_index_template.test.name
}
`, name)
}

// TestAccIndexTemplateDataSourceMetadataVersionID covers metadata, version, and the id attribute.
func TestAccIndexTemplateDataSourceMetadataVersionID(t *testing.T) {
	templateName := "test-ds-meta-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccIndexTemplateDataSourceMetadataVersionConfig(templateName, `{"owner":"team-a","description":"initial"}`, 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_index_template.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "version", "5"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "metadata", `{"description":"initial","owner":"team-a"}`),
				),
			},
			{
				Config: testAccIndexTemplateDataSourceMetadataVersionConfig(templateName, `{"owner":"team-b","description":"updated"}`, 7),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_index_template.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "version", "7"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "metadata", `{"description":"updated","owner":"team-b"}`),
				),
			},
		},
	})
}

func testAccIndexTemplateDataSourceMetadataVersionConfig(name, metadata string, ver int) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = "%[1]s"
  index_patterns = ["%[1]s-*"]
  version        = %[3]d
  metadata       = jsonencode(%[2]s)
}

data "elasticstack_elasticsearch_index_template" "test" {
  name = elasticstack_elasticsearch_index_template.test.name
}
`, name, metadata, ver)
}

// TestAccIndexTemplateDataSourceCountAssertions covers index_patterns.#, composed_of.#,
// and ignore_missing_component_templates.# with multi-value assertions.
func TestAccIndexTemplateDataSourceCountAssertions(t *testing.T) {
	templateName := "test-ds-count-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	comp1 := templateName + "-comp1@custom"
	comp2 := templateName + "-comp2@custom"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(index.MinSupportedIgnoreMissingComponentTemplateVersion),
				Config:   testAccIndexTemplateDataSourceCountConfig(templateName, comp1, comp2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "index_patterns.#", "2"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_index_template.test", "index_patterns.*", fmt.Sprintf("%s-a-*", templateName)),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_index_template.test", "index_patterns.*", fmt.Sprintf("%s-b-*", templateName)),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "composed_of.#", "2"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_index_template.test", "composed_of.*", comp1),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_index_template.test", "composed_of.*", comp2),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "ignore_missing_component_templates.#", "2"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_index_template.test", "ignore_missing_component_templates.*", comp1),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_index_template.test", "ignore_missing_component_templates.*", comp2),
				),
			},
		},
	})
}

func testAccIndexTemplateDataSourceCountConfig(name, comp1, comp2 string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = "%[1]s"
  index_patterns = ["%[1]s-a-*", "%[1]s-b-*"]

  composed_of                        = ["%[2]s", "%[3]s"]
  ignore_missing_component_templates = ["%[2]s", "%[3]s"]
}

data "elasticstack_elasticsearch_index_template" "test" {
  name = elasticstack_elasticsearch_index_template.test.name
}
`, name, comp1, comp2)
}

// TestAccIndexTemplateDataSourceLifecycle covers template.0.lifecycle.*.data_retention.
func TestAccIndexTemplateDataSourceLifecycle(t *testing.T) {
	templateName := "test-ds-lifecycle-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(datastreamlifecycle.MinVersion),
				Config:   testAccIndexTemplateDataSourceLifecycleConfig(templateName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "template.0.lifecycle.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.elasticstack_elasticsearch_index_template.test",
						"template.0.lifecycle.*",
						map[string]string{
							"data_retention": "30d",
						},
					),
				),
			},
		},
	})
}

func testAccIndexTemplateDataSourceLifecycleConfig(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = "%[1]s"
  index_patterns = ["%[1]s-*"]

  data_stream {}

  template {
    lifecycle {
      data_retention = "30d"
    }
  }
}

data "elasticstack_elasticsearch_index_template" "test" {
  name = elasticstack_elasticsearch_index_template.test.name
}
`, name)
}
