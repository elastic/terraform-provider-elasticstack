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
	"regexp"
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

	resource.ParallelTest(t, resource.TestCase{
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
					resource.TestMatchResourceAttr("elasticstack_elasticsearch_index_template.test", "id", regexp.MustCompile(fmt.Sprintf(".+/%s$", templateName))),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "index_patterns.#", "1"),
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
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "index_patterns.#", "3"),
					resource.TestCheckTypeSetElemAttr(
						"elasticstack_elasticsearch_index_template.test",
						"index_patterns.*",
						fmt.Sprintf("%s-logs-*", templateName),
					),
					resource.TestCheckTypeSetElemAttr(
						"elasticstack_elasticsearch_index_template.test",
						"index_patterns.*",
						fmt.Sprintf("%s-metrics-*", templateName),
					),
					resource.TestCheckTypeSetElemAttr(
						"elasticstack_elasticsearch_index_template.test",
						"index_patterns.*",
						fmt.Sprintf("%s-traces-*", templateName),
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
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "template.0.settings", `{"index":{"number_of_replicas":"0","number_of_shards":"1"}}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test2", "name", fmt.Sprintf("%s-stream", templateName)),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test2", "data_stream.0.hidden", "false"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_final"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "index_patterns.#", "1"),
					resource.TestCheckTypeSetElemAttr(
						"elasticstack_elasticsearch_index_template.test",
						"index_patterns.*",
						fmt.Sprintf("%s-archive-*", templateName),
					),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "template.0.settings", `{"index":{"number_of_replicas":"0","number_of_shards":"1"}}`),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_final"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
				},
				ResourceName:      "elasticstack_elasticsearch_index_template.test",
				ImportState:       true,
				ImportStateVerify: true,
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
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test_component", "composed_of.#", "3"),
					resource.TestCheckTypeSetElemAttr(
						"elasticstack_elasticsearch_index_template.test_component",
						"composed_of.*",
						fmt.Sprintf("%s-logscomponent-a@custom", templateNameComponent),
					),
					resource.TestCheckTypeSetElemAttr(
						"elasticstack_elasticsearch_index_template.test_component",
						"composed_of.*",
						fmt.Sprintf("%s-logscomponent-b@custom", templateNameComponent),
					),
					resource.TestCheckTypeSetElemAttr(
						"elasticstack_elasticsearch_index_template.test_component",
						"composed_of.*",
						fmt.Sprintf("%s-logscomponent-c@custom", templateNameComponent),
					),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test_component", "ignore_missing_component_templates.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"elasticstack_elasticsearch_index_template.test_component",
						"ignore_missing_component_templates.*",
						fmt.Sprintf("%s-logscomponent-b@custom", templateNameComponent),
					),
					resource.TestCheckTypeSetElemAttr(
						"elasticstack_elasticsearch_index_template.test_component",
						"ignore_missing_component_templates.*",
						fmt.Sprintf("%s-logscomponent-c@custom", templateNameComponent),
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
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test_component", "composed_of.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"elasticstack_elasticsearch_index_template.test_component",
						"composed_of.*",
						fmt.Sprintf("%s-logscomponent-a@custom", templateNameComponent),
					),
					resource.TestCheckTypeSetElemAttr(
						"elasticstack_elasticsearch_index_template.test_component",
						"composed_of.*",
						fmt.Sprintf("%s-logscomponent-c@custom", templateNameComponent),
					),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test_component", "ignore_missing_component_templates.#", "1"),
					resource.TestCheckTypeSetElemAttr(
						"elasticstack_elasticsearch_index_template.test_component",
						"ignore_missing_component_templates.*",
						fmt.Sprintf("%s-logscomponent-c@custom", templateNameComponent),
					),
				),
			},
		},
	})
}

func TestAccResourceIndexTemplateWithExplicitConnection(t *testing.T) {
	templateName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	endpoints := indexTemplateESEndpoints()
	if len(endpoints) == 0 {
		t.Fatal("ELASTICSEARCH_ENDPOINTS must contain at least one endpoint")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexTemplateDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				Config:                   testAccResourceIndexTemplateWithExplicitConnection(templateName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("elasticstack_elasticsearch_index_template.test", "id", regexp.MustCompile(fmt.Sprintf(".+/%s$", templateName))),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "elasticsearch_connection.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "elasticsearch_connection.0.endpoints.#", fmt.Sprintf("%d", len(endpoints))),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "elasticsearch_connection.0.endpoints.0", endpoints[0]),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "elasticsearch_connection.0.insecure", "true"),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexTemplateDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"template_name":    config.StringVariable(templateName),
					"metadata":         config.StringVariable(`{"description":"initial template","owner":"team-a"}`),
					"mappings":         config.StringVariable(`{"properties":{"log_level":{"type":"keyword"}}}`),
					"template_version": config.IntegerVariable(1),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "version", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "metadata", `{"description":"initial template","owner":"team-a"}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "template.0.mappings", `{"properties":{"log_level":{"type":"keyword"}}}`),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"template_name":    config.StringVariable(templateName),
					"metadata":         config.StringVariable(`{"description":"updated template","owner":"team-b"}`),
					"mappings":         config.StringVariable(`{"properties":{"log_level":{"type":"keyword"},"severity":{"type":"integer"}}}`),
					"template_version": config.IntegerVariable(2),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "version", "2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "metadata", `{"description":"updated template","owner":"team-b"}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "template.0.mappings", `{"properties":{"log_level":{"type":"keyword"},"severity":{"type":"integer"}}}`),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("unset"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName + "unset"),
				},
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

func TestAccResourceIndexTemplateLifecycle(t *testing.T) {
	templateName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexTemplateDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(datastreamlifecycle.MinVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"template_name":  config.StringVariable(templateName),
					"data_retention": config.StringVariable("30d"),
				},
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
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"template_name":  config.StringVariable(templateName),
					"data_retention": config.StringVariable("60d"),
				},
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

func TestAccResourceIndexTemplateAliasFilter(t *testing.T) {
	templateName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexTemplateDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
					"alias_name":    config.StringVariable("filtered_alias_v1"),
					"filter":        config.StringVariable(`{"term":{"status":"active"}}`),
				},
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
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
					"alias_name":    config.StringVariable("filtered_alias_v2"),
					"filter":        config.StringVariable(`{"bool":{"must":[{"term":{"service.name":"api"}},{"term":{"status":"active"}}]}}`),
				},
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
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("remove"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
					"alias_name":    config.StringVariable("filtered_alias_v3"),
				},
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
					testCheckTemplateAliasAttrCleared("filtered_alias_v3", "filter"),
				),
			},
		},
	})
}

func TestAccResourceIndexTemplateAliasDetails(t *testing.T) {
	templateName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
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

func TestAccResourceIndexTemplateAliasLifecycleRemoval(t *testing.T) {
	templateName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexTemplateDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(datastreamlifecycle.MinVersion),
				Config:                   testAccResourceIndexTemplateAliasLifecycleConfig(templateName, "detailed_alias_initial", "30d", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "template.0.alias.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"elasticstack_elasticsearch_index_template.test",
						"template.0.alias.*",
						map[string]string{
							"name":           "detailed_alias_initial",
							"is_hidden":      "true",
							"is_write_index": "true",
							"routing":        "shard_1",
							"search_routing": "shard_1",
							"index_routing":  "shard_1",
						},
					),
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
				Config:                   testAccResourceIndexTemplateAliasLifecycleConfig(templateName, "detailed_alias_reset", "", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "template.0.alias.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"elasticstack_elasticsearch_index_template.test",
						"template.0.alias.*",
						map[string]string{
							"name": "detailed_alias_reset",
						},
					),
					testCheckTemplateAliasBoolAttrFalseOrAbsent("detailed_alias_reset", "is_hidden"),
					testCheckTemplateAliasBoolAttrFalseOrAbsent("detailed_alias_reset", "is_write_index"),
					testCheckTemplateAliasAttrCleared("detailed_alias_reset", "routing"),
					testCheckTemplateAliasAttrCleared("detailed_alias_reset", "search_routing"),
					testCheckTemplateAliasAttrCleared("detailed_alias_reset", "index_routing"),
					testCheckTemplateLifecycleAttrCleared("data_retention"),
				),
			},
		},
	})
}

func TestAccResourceIndexTemplateAliasRoutingFromRoutingOnly(t *testing.T) {
	templateName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
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

func TestAccResourceIndexTemplateDataStreamCustomRouting(t *testing.T) {
	templateName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexTemplateDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedAllowCustomRoutingVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"template_name":        config.StringVariable(templateName),
					"allow_custom_routing": config.BoolVariable(true),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "data_stream.0.allow_custom_routing", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedAllowCustomRoutingVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"template_name":        config.StringVariable(templateName),
					"allow_custom_routing": config.BoolVariable(false),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "data_stream.0.allow_custom_routing", "false"),
				),
			},
		},
	})
}

func TestAccResourceIndexTemplateEmptyTemplateBlock(t *testing.T) {
	templateName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
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
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "template.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "template.0.alias.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "template.0.settings", `{"index":{"number_of_shards":"2"}}`),
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
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "template.#", "1"),
					testCheckAttrEmptyOrAbsent("elasticstack_elasticsearch_index_template.test", "template.0.mappings"),
					testCheckAttrEmptyOrAbsent("elasticstack_elasticsearch_index_template.test", "template.0.settings"),
					testCheckNoTemplateAliases("elasticstack_elasticsearch_index_template.test"),
				),
			},
		},
	})
}

func TestAccResourceIndexTemplateDataStreamEmptyObject(t *testing.T) {
	templateName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexTemplateDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedAllowCustomRoutingVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"template_name":        config.StringVariable(templateName),
					"hidden":               config.BoolVariable(true),
					"allow_custom_routing": config.BoolVariable(true),
				},
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
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
				},
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

func TestAccResourceIndexTemplateEmptyCollections(t *testing.T) {
	templateName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexTemplateDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(index.MinSupportedIgnoreMissingComponentTemplateVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"template_name":  config.StringVariable(templateName),
					"component_name": config.StringVariable(templateName + "-comp@custom"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_index_template.test", "composed_of.*", templateName+"-comp@custom"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_index_template.test", "ignore_missing_component_templates.*", templateName+"-comp@custom"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(index.MinSupportedIgnoreMissingComponentTemplateVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "composed_of.#", "0"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "ignore_missing_component_templates.#", "0"),
				),
			},
		},
	})
}

func testCheckTemplateAliasAttrCleared(aliasName, attrName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		const resourceName = "elasticstack_elasticsearch_index_template.test"
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

func testCheckTemplateAliasBoolAttrFalseOrAbsent(aliasName, attrName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		const resourceName = "elasticstack_elasticsearch_index_template.test"
		aliasPrefix, err := templateAliasPrefix(s, resourceName, aliasName)
		if err != nil {
			return err
		}

		value, ok := s.RootModule().Resources[resourceName].Primary.Attributes[aliasPrefix+"."+attrName]
		if ok && value != "" && value != "false" {
			return fmt.Errorf("expected %s.%s to be false or absent, got %q", aliasPrefix, attrName, value)
		}
		return nil
	}
}

func testCheckTemplateLifecycleAttrCleared(attrName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		const resourceName = "elasticstack_elasticsearch_index_template.test"
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

func testAccResourceIndexTemplateWithExplicitConnection(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = %q
  index_patterns = [%q]

  elasticsearch_connection {
    %s
    insecure = true
  }

  template {
    settings = jsonencode({
      number_of_shards = "1"
    })
  }
}`, name, name+"-connection-*", buildIndexTemplateESConnectionBlock())
}

func testAccResourceIndexTemplateAliasLifecycleConfig(name, aliasName, dataRetention string, includeAliasDetails bool) string {
	lifecycleBlock := ""
	if dataRetention != "" {
		lifecycleBlock = fmt.Sprintf(`
    lifecycle {
      data_retention = %q
    }`, dataRetention)
	}

	aliasFields := ""
	if includeAliasDetails {
		aliasFields = `
      is_hidden      = true
      is_write_index = true
      routing        = "shard_1"
      search_routing = "shard_1"
      index_routing  = "shard_1"`
	}

	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = %q
  index_patterns = [%q]

  data_stream {}

  template {
    alias {
      name = %q%s
    }%s
  }
}`, name, name+"-*", aliasName, aliasFields, lifecycleBlock)
}

func buildIndexTemplateESConnectionBlock() string {
	endpoints := indexTemplateESEndpoints()
	quoted := make([]string, 0, len(endpoints))
	for _, endpoint := range endpoints {
		quoted = append(quoted, fmt.Sprintf("%q", endpoint))
	}
	endpointList := strings.Join(quoted, ", ")

	if apiKey := os.Getenv("ELASTICSEARCH_API_KEY"); apiKey != "" {
		return fmt.Sprintf(`endpoints = [%s]
    api_key   = %q`, endpointList, apiKey)
	}

	return fmt.Sprintf(`endpoints = [%s]
    username  = %q
    password  = %q`, endpointList, os.Getenv("ELASTICSEARCH_USERNAME"), os.Getenv("ELASTICSEARCH_PASSWORD"))
}

func indexTemplateESEndpoints() []string {
	rawEndpoints := os.Getenv("ELASTICSEARCH_ENDPOINTS")
	parts := strings.Split(rawEndpoints, ",")
	endpoints := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			endpoints = append(endpoints, part)
		}
	}
	return endpoints
}
