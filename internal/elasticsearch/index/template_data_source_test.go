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
	"os"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/datastreamlifecycle"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccIndexTemplateDataSource(t *testing.T) {
	// generate a random role name
	templateName := "test-template-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	templateNameComponent := "test-template-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
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
// aliases (name, filter, index_routing, search_routing, is_hidden, is_write_index), mappings, and settings.
func TestAccIndexTemplateDataSourceTemplate(t *testing.T) {
	templateName := "test-ds-tpl-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
					"mappings":      config.StringVariable(`{"properties":{"log_level":{"type":"keyword"}}}`),
					"settings":      config.StringVariable(`{"index":{"number_of_shards":"1"}}`),
				},
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
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
					"mappings":      config.StringVariable(`{"properties":{"log_level":{"type":"keyword"},"severity":{"type":"integer"}}}`),
					"settings":      config.StringVariable(`{"index":{"number_of_shards":"2"}}`),
				},
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

func TestAccIndexTemplateDataSourceExplicitConnection(t *testing.T) {
	templateName := "test-ds-conn-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	endpoint := indexTemplateDataSourcePrimaryESEndpoint(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { preCheckIndexTemplateDataSourceBasicAuth(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
					"endpoint":      config.StringVariable(endpoint),
					"username":      config.StringVariable(os.Getenv("ELASTICSEARCH_USERNAME")),
					"password":      config.StringVariable(os.Getenv("ELASTICSEARCH_PASSWORD")),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "name", templateName),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.username", os.Getenv("ELASTICSEARCH_USERNAME")),
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.password"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.endpoints.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.endpoints.0", endpoint),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.headers.%", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.headers.XTerraformTest", "basic-auth"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.insecure", "true"),
				),
			},
		},
	})
}

// TestAccIndexTemplateDataSourceDataStream covers data_stream.0.hidden and data_stream.0.allow_custom_routing.
func TestAccIndexTemplateDataSourceDataStream(t *testing.T) {
	templateName := "test-ds-stream-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
					"hidden":        config.BoolVariable(true),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "data_stream.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "data_stream.0.hidden", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedAllowCustomRoutingVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "data_stream.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "data_stream.0.hidden", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "data_stream.0.allow_custom_routing", "true"),
				),
			},
		},
	})
}

// TestAccIndexTemplateDataSourceMetadataVersionID covers metadata, version, and the id attribute.
func TestAccIndexTemplateDataSourceMetadataVersionID(t *testing.T) {
	templateName := "test-ds-meta-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"template_name":    config.StringVariable(templateName),
					"metadata":         config.StringVariable(`{"owner":"team-a","description":"initial"}`),
					"template_version": config.StringVariable("5"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_index_template.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "version", "5"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "metadata", `{"description":"initial","owner":"team-a"}`),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"template_name":    config.StringVariable(templateName),
					"metadata":         config.StringVariable(`{"owner":"team-b","description":"updated"}`),
					"template_version": config.StringVariable("7"),
				},
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

// TestAccIndexTemplateDataSourceCountAssertions covers index_patterns.#, composed_of.#,
// and ignore_missing_component_templates.# with multi-value assertions.
func TestAccIndexTemplateDataSourceCountAssertions(t *testing.T) {
	templateName := "test-ds-count-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	comp1 := templateName + "-comp1@custom"
	comp2 := templateName + "-comp2@custom"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(index.MinSupportedIgnoreMissingComponentTemplateVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
					"component_1":   config.StringVariable(comp1),
					"component_2":   config.StringVariable(comp2),
				},
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

// TestAccIndexTemplateDataSourceLifecycle covers template.0.lifecycle.*.data_retention.
func TestAccIndexTemplateDataSourceLifecycle(t *testing.T) {
	templateName := "test-ds-lifecycle-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(datastreamlifecycle.MinVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
				},
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

func TestAccIndexTemplateDataSourceOrderedComponents(t *testing.T) {
	templateName := "test-ds-order-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	comp1 := templateName + "-comp1@custom"
	comp2 := templateName + "-comp2@custom"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(index.MinSupportedIgnoreMissingComponentTemplateVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
					"component_1":   config.StringVariable(comp1),
					"component_2":   config.StringVariable(comp2),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "composed_of.#", "2"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "composed_of.0", comp1),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "composed_of.1", comp2),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "ignore_missing_component_templates.#", "2"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "ignore_missing_component_templates.0", comp1),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "ignore_missing_component_templates.1", comp2),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(index.MinSupportedIgnoreMissingComponentTemplateVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "composed_of.#", "0"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "ignore_missing_component_templates.#", "0"),
				),
			},
		},
	})
}

func TestAccIndexTemplateDataSourceDataStreamEmptyObject(t *testing.T) {
	templateName := "test-ds-empty-stream-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedAllowCustomRoutingVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "data_stream.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "data_stream.0.hidden", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "data_stream.0.allow_custom_routing", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedAllowCustomRoutingVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "data_stream.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "data_stream.0.hidden", "false"),
					testCheckDataStreamAttrFalseOrAbsent("data.elasticstack_elasticsearch_index_template.test", "allow_custom_routing"),
				),
			},
		},
	})
}

func TestAccIndexTemplateDataSourceOptionalFieldRemoval(t *testing.T) {
	templateName := "test-ds-remove-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(datastreamlifecycle.MinVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "metadata", `{"description":"initial","owner":"team-a"}`),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "version", "7"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "template.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "template.0.alias.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "template.0.mappings", `{"properties":{"log_level":{"type":"keyword"}}}`),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "template.0.settings", `{"index":{"number_of_shards":"1"}}`),
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
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(datastreamlifecycle.MinVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "template.#", "1"),
					testCheckDataSourceAttrEmptyOrAbsent("data.elasticstack_elasticsearch_index_template.test", "metadata"),
					testCheckAttrZeroOrAbsent("data.elasticstack_elasticsearch_index_template.test", "version"),
					testCheckDataSourceAttrEmptyOrAbsent("data.elasticstack_elasticsearch_index_template.test", "template.0.mappings"),
					testCheckDataSourceAttrEmptyOrAbsent("data.elasticstack_elasticsearch_index_template.test", "template.0.settings"),
					testCheckAttrZeroOrAbsent("data.elasticstack_elasticsearch_index_template.test", "template.0.lifecycle.#"),
					testCheckDataSourceTemplateLifecycleAttrCleared("data.elasticstack_elasticsearch_index_template.test", "data_retention"),
					testCheckNoTemplateAliases("data.elasticstack_elasticsearch_index_template.test"),
				),
			},
		},
	})
}

func preCheckIndexTemplateDataSourceBasicAuth(t *testing.T) {
	acctest.PreCheck(t)
	if os.Getenv("ELASTICSEARCH_USERNAME") == "" || os.Getenv("ELASTICSEARCH_PASSWORD") == "" {
		t.Skip("ELASTICSEARCH_USERNAME and ELASTICSEARCH_PASSWORD must be set for explicit basic auth coverage")
	}
}

func indexTemplateDataSourcePrimaryESEndpoint(t *testing.T) string {
	for endpoint := range strings.SplitSeq(os.Getenv("ELASTICSEARCH_ENDPOINTS"), ",") {
		if trimmed := strings.TrimSpace(endpoint); trimmed != "" {
			return trimmed
		}
	}

	t.Fatal("ELASTICSEARCH_ENDPOINTS must contain at least one endpoint")
	return ""
}

func testCheckDataSourceAttrEmptyOrAbsent(resourceName, attrName string) resource.TestCheckFunc {
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

func testCheckDataSourceTemplateLifecycleAttrCleared(resourceName, attrName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found in state: %s", resourceName)
		}

		for key, value := range rs.Primary.Attributes {
			if strings.HasPrefix(key, "template.0.lifecycle.") && strings.HasSuffix(key, "."+attrName) && value != "" {
				return fmt.Errorf("expected lifecycle attribute %s to be cleared, got %q", key, value)
			}
		}
		return nil
	}
}
