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
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.name", indexName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.is_hidden", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.#", "0"),
				),
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
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.name", indexName2),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.#", "1"),
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
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.name", indexName),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_alias.test_alias", "write_index.filter"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.index_routing", "write-routing"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.#", "1"),
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
