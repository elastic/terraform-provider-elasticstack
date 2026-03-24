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

package alias_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccResourceAlias(t *testing.T) {
	// generate random names
	aliasName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlpha)
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlpha)
	indexName2 := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
		},
		CheckDestroy: checkResourceAliasDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: map[string]config.Variable{
					"alias_name":  config.StringVariable(aliasName),
					"index_name":  config.StringVariable(indexName),
					"index_name2": config.StringVariable(indexName2),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_alias.test_alias", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.name", indexName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.is_hidden", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.#", "0"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ResourceName:             "elasticstack_elasticsearch_index_alias.test_alias",
				ImportState:              true,
				ImportStateVerify:        true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: map[string]config.Variable{
					"alias_name":  config.StringVariable(aliasName),
					"index_name":  config.StringVariable(indexName),
					"index_name2": config.StringVariable(indexName2),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_alias.test_alias", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.name", indexName2),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.*", map[string]string{
						"name": indexName,
					}),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_filter"),
				ConfigVariables: map[string]config.Variable{
					"alias_name":  config.StringVariable(aliasName),
					"index_name":  config.StringVariable(indexName),
					"index_name2": config.StringVariable(indexName2),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_alias.test_alias", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.name", indexName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.filter", `{"term":{"status":"published"}}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.index_routing", "write-routing"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.*", map[string]string{
						"name":   indexName2,
						"filter": `{"term":{"status":"draft"}}`,
					}),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_filter"),
				ConfigVariables: map[string]config.Variable{
					"alias_name":  config.StringVariable(aliasName),
					"index_name":  config.StringVariable(indexName),
					"index_name2": config.StringVariable(indexName2),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_alias.test_alias", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.name", indexName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.filter", `{"term":{"status":"review"}}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.index_routing", "write-routing"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.*", map[string]string{
						"name":   indexName2,
						"filter": `{"term":{"status":"archived"}}`,
					}),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("remove_filter"),
				ConfigVariables: map[string]config.Variable{
					"alias_name":  config.StringVariable(aliasName),
					"index_name":  config.StringVariable(indexName),
					"index_name2": config.StringVariable(indexName2),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_alias.test_alias", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.name", indexName),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.filter"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.index_routing", "write-routing"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.#", "1"),
					testCheckReadIndexAttrs("elasticstack_elasticsearch_index_alias.test_alias", indexName2, map[string]string{}, []string{"filter"}),
				),
			},
		},
	})
}

func TestAccResourceAliasIssue1750(t *testing.T) {
	aliasName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlpha)
	writeIndexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlpha)
	readIndexName1 := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlpha)
	readIndexName2 := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAliasDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("repro"),
				ConfigVariables: config.Variables{
					"aliases": config.ListVariable(
						config.ObjectVariable(map[string]config.Variable{
							"name": config.StringVariable(aliasName),
							"write_index": config.ObjectVariable(map[string]config.Variable{
								"name": config.StringVariable(writeIndexName),
							}),
							"read_indices": config.SetVariable(
								config.ObjectVariable(map[string]config.Variable{
									"name": config.StringVariable(readIndexName1),
								}),
								config.ObjectVariable(map[string]config.Variable{
									"name": config.StringVariable(readIndexName2),
								}),
							),
						}),
					),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"elasticstack_elasticsearch_index_alias.this.0",
						"name",
						aliasName,
					),
					resource.TestCheckResourceAttr(
						"elasticstack_elasticsearch_index_alias.this.0",
						"write_index.name",
						writeIndexName,
					),
					resource.TestCheckResourceAttr(
						"elasticstack_elasticsearch_index_alias.this.0",
						"read_indices.#",
						"2",
					),
				),
			},
		},
	})
}

func TestAccResourceAliasWriteIndex(t *testing.T) {
	// generate random names
	aliasName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlpha)
	indexName1 := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlpha)
	indexName2 := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlpha)
	indexName3 := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
		},
		CheckDestroy: checkResourceAliasDestroy,
		Steps: []resource.TestStep{
			// Case 1: Single index with is_write_index=true
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("single"),
				ConfigVariables: map[string]config.Variable{
					"alias_name":  config.StringVariable(aliasName),
					"index_name1": config.StringVariable(indexName1),
					"index_name2": config.StringVariable(indexName2),
					"index_name3": config.StringVariable(indexName3),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.name", indexName1),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.#", "0"),
				),
			},
			// Case 2: Add new index with is_write_index=true, existing becomes read index
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("switch"),
				ConfigVariables: map[string]config.Variable{
					"alias_name":  config.StringVariable(aliasName),
					"index_name1": config.StringVariable(indexName1),
					"index_name2": config.StringVariable(indexName2),
					"index_name3": config.StringVariable(indexName3),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.name", indexName2),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.#", "1"),
				),
			},
			// Case 3: Add third index as write index
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("triple"),
				ConfigVariables: map[string]config.Variable{
					"alias_name":  config.StringVariable(aliasName),
					"index_name1": config.StringVariable(indexName1),
					"index_name2": config.StringVariable(indexName2),
					"index_name3": config.StringVariable(indexName3),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.name", indexName3),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.#", "2"),
				),
			},
			// Case 4: Remove initial index, keep two indices with one as write index
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("remove_first"),
				ConfigVariables: map[string]config.Variable{
					"alias_name":  config.StringVariable(aliasName),
					"index_name1": config.StringVariable(indexName1),
					"index_name2": config.StringVariable(indexName2),
					"index_name3": config.StringVariable(indexName3),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.name", indexName3),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceAliasDataStream(t *testing.T) {
	// generate random names
	aliasName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlpha)
	dsName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAliasDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: map[string]config.Variable{
					"alias_name": config.StringVariable(aliasName),
					"ds_name":    config.StringVariable(dsName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.name", dsName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceAliasRouting(t *testing.T) {
	aliasName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlpha)
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlpha)
	indexName2 := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAliasDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_routing"),
				ConfigVariables: map[string]config.Variable{
					"alias_name":  config.StringVariable(aliasName),
					"index_name":  config.StringVariable(indexName),
					"index_name2": config.StringVariable(indexName2),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_alias.test_alias", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.name", indexName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.routing", "wr1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.index_routing", "wir1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.search_routing", "wsr1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.*", map[string]string{
						"name":           indexName2,
						"routing":        "rr1",
						"index_routing":  "rir1",
						"search_routing": "rsr1",
					}),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_routing"),
				ConfigVariables: map[string]config.Variable{
					"alias_name":  config.StringVariable(aliasName),
					"index_name":  config.StringVariable(indexName),
					"index_name2": config.StringVariable(indexName2),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_alias.test_alias", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.name", indexName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.routing", "wr2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.index_routing", "wir2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.search_routing", "wsr2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.*", map[string]string{
						"name":           indexName2,
						"routing":        "rr2",
						"index_routing":  "rir2",
						"search_routing": "rsr2",
					}),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("remove_routing"),
				ConfigVariables: map[string]config.Variable{
					"alias_name":  config.StringVariable(aliasName),
					"index_name":  config.StringVariable(indexName),
					"index_name2": config.StringVariable(indexName2),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_alias.test_alias", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.name", indexName),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.routing"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.index_routing"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.search_routing"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.#", "1"),
					testCheckReadIndexAttrs("elasticstack_elasticsearch_index_alias.test_alias", indexName2, map[string]string{}, []string{"routing", "index_routing", "search_routing"}),
				),
			},
		},
	})
}

func TestAccResourceAliasIsHidden(t *testing.T) {
	aliasName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlpha)
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlpha)
	indexName2 := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAliasDestroy,
		Steps: []resource.TestStep{
			// Step 1: set is_hidden = true on both write and read indices
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("hidden"),
				ConfigVariables: map[string]config.Variable{
					"alias_name":  config.StringVariable(aliasName),
					"index_name":  config.StringVariable(indexName),
					"index_name2": config.StringVariable(indexName2),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.name", indexName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.is_hidden", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.*", map[string]string{
						"name":      indexName2,
						"is_hidden": "true",
					}),
				),
			},
			// Step 2: update is_hidden back to false
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("visible"),
				ConfigVariables: map[string]config.Variable{
					"alias_name":  config.StringVariable(aliasName),
					"index_name":  config.StringVariable(indexName),
					"index_name2": config.StringVariable(indexName2),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.name", indexName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.is_hidden", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.*", map[string]string{
						"name":      indexName2,
						"is_hidden": "false",
					}),
				),
			},
		},
	})
}

func checkResourceAliasDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_index_alias" {
			continue
		}

		// Handle the case where ID might not be in the expected format
		aliasName := rs.Primary.ID
		if compID, err := clients.CompositeIDFromStr(rs.Primary.ID); err == nil {
			aliasName = compID.ResourceID
		}

		esClient, err := client.GetESClient()
		if err != nil {
			return err
		}

		res, err := esClient.Indices.GetAlias(
			esClient.Indices.GetAlias.WithName(aliasName),
		)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if res.StatusCode != 404 {
			return fmt.Errorf("Alias (%s) still exists", aliasName)
		}
	}
	return nil
}

func testCheckReadIndexAttrs(resourceName, indexName string, expected map[string]string, absent []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found in state: %s", resourceName)
		}

		var prefix string
		for key, value := range rs.Primary.Attributes {
			if strings.HasPrefix(key, "read_indices.") && strings.HasSuffix(key, ".name") && value == indexName {
				prefix = strings.TrimSuffix(key, ".name")
				break
			}
		}

		if prefix == "" {
			return fmt.Errorf("read index %q not found in state for resource %s", indexName, resourceName)
		}

		for attr, want := range expected {
			key := prefix + "." + attr
			got, ok := rs.Primary.Attributes[key]
			if !ok {
				return fmt.Errorf("expected attribute %q to be set for read index %q", key, indexName)
			}
			if got != want {
				return fmt.Errorf("expected attribute %q for read index %q to be %q, got %q", key, indexName, want, got)
			}
		}

		for _, attr := range absent {
			key := prefix + "." + attr
			if got, ok := rs.Primary.Attributes[key]; ok {
				return fmt.Errorf("expected attribute %q for read index %q to be absent, got %q", key, indexName, got)
			}
		}

		return nil
	}
}
