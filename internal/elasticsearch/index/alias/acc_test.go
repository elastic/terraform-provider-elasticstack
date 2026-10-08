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
	"context"
	_ "embed"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"

	estypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	esclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/aliasutil"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
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
			// Step 4: verify filter removal after transitioning back to update config (no filter)
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
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.filter"),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.*", map[string]string{
						"name": indexName,
					}),
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
			// Step 1: index_routing and search_routing on write and read indices
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_routing"),
				ConfigVariables: map[string]config.Variable{
					"alias_name":  config.StringVariable(aliasName),
					"index_name":  config.StringVariable(indexName),
					"index_name2": config.StringVariable(indexName2),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.name", indexName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.index_routing", "wir1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.search_routing", "wsr1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.*", map[string]string{
						"name":           indexName2,
						"index_routing":  "rir1",
						"search_routing": "rsr1",
					}),
				),
			},
			// Step 2: remove all routing attributes; verify absence
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("no_routing"),
				ConfigVariables: map[string]config.Variable{
					"alias_name":  config.StringVariable(aliasName),
					"index_name":  config.StringVariable(indexName),
					"index_name2": config.StringVariable(indexName2),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "name", aliasName),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.routing"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.index_routing"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.search_routing"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.*", map[string]string{
						"name": indexName2,
					}),
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

func TestAccResourceAliasNoWriteIndex(t *testing.T) {
	aliasName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlpha)
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAliasDestroy,
		Steps: []resource.TestStep{
			// Alias with only read_indices and no write_index
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read_only"),
				ConfigVariables: map[string]config.Variable{
					"alias_name": config.StringVariable(aliasName),
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "name", aliasName),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.*", map[string]string{
						"name": indexName,
					}),
				),
			},
		},
	})
}

//go:embed testdata/TestAccResourceAliasExistingStateWithoutConcreteIndices/previous/main.tf
var aliasExistingStateWithoutConcreteIndicesConfig string

func TestAccResourceAliasExistingStateWithoutConcreteIndices(t *testing.T) {
	suffix := strings.ToLower(sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlpha))
	aliasName := fmt.Sprintf("alias-existing-state-%s", suffix)
	indexName := fmt.Sprintf("existing-state-target-%s", suffix)
	createAliasTestIndex(context.Background(), t, indexName)
	t.Cleanup(func() { deleteAliasTestIndex(context.Background(), t, indexName) })

	variables := config.Variables{
		"alias_name": config.StringVariable(aliasName),
		"index_name": config.StringVariable(indexName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAliasDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"elasticstack": {
						Source:            "elastic/elasticstack",
						VersionConstraint: "0.16.5",
					},
				},
				Config:          aliasExistingStateWithoutConcreteIndicesConfig,
				ConfigVariables: variables,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("upgrade"),
				ConfigVariables:          variables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccResourceAliasImportWithWildcardConfiguration(t *testing.T) {
	suffix := strings.ToLower(sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlpha))
	aliasName := fmt.Sprintf("alias-import-wildcard-%s", suffix)
	indexName := fmt.Sprintf("import-wildcard-target-%s", suffix)
	readIndicesPattern := fmt.Sprintf("import-wildcard-target-*%s", suffix)
	ctx := context.Background()
	createAliasTestIndex(ctx, t, indexName)
	t.Cleanup(func() { deleteAliasTestIndex(ctx, t, indexName) })

	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		t.Fatalf("acceptance Elasticsearch client: %v", err)
	}
	diags := esclient.UpdateAliasesAtomic(ctx, client, []esclient.AliasAction{{
		Type:  "add",
		Index: indexName,
		Alias: aliasName,
	}})
	if diags.HasError() {
		t.Fatalf("create alias %q: %s", aliasName, diags.Errors()[0].Detail())
	}

	variables := config.Variables{
		"alias_name":           config.StringVariable(aliasName),
		"read_indices_pattern": config.StringVariable(readIndicesPattern),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAliasDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("wildcard"),
				ConfigVariables:          variables,
				ImportState:              true,
				ImportStateId:            aliasName,
				ImportStatePersist:       true,
				ResourceName:             "elasticstack_elasticsearch_index_alias.test_alias",
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.*", map[string]string{
						"name":               indexName,
						"concrete_indices.#": "1",
					}),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("wildcard"),
				ConfigVariables:          variables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("elasticstack_elasticsearch_index_alias.test_alias", plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccResourceAliasWildcardReadIndices(t *testing.T) {
	suffix := sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlpha)
	aliasName := fmt.Sprintf("alias-test-traces-apm-%s", suffix)
	indexName1 := fmt.Sprintf("test-traces-apm-default-%s", suffix)
	indexName2 := fmt.Sprintf("test-traces-apm.rum-default-%s", suffix)
	indexName3 := fmt.Sprintf("test-traces-apm.logs-default-%s", suffix)
	readIndicesPattern := fmt.Sprintf("test-traces-apm*-%s", suffix)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAliasDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("wildcard_create"),
				ConfigVariables: config.Variables{
					"alias_name":           config.StringVariable(aliasName),
					"index_name1":          config.StringVariable(indexName1),
					"index_name2":          config.StringVariable(indexName2),
					"index_name3":          config.StringVariable(indexName3),
					"read_indices_pattern": config.StringVariable(readIndicesPattern),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.*", map[string]string{
						"name":               readIndicesPattern,
						"concrete_indices.#": "2",
					}),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("wildcard_target_added"),
				ConfigVariables: config.Variables{
					"alias_name":           config.StringVariable(aliasName),
					"index_name1":          config.StringVariable(indexName1),
					"index_name2":          config.StringVariable(indexName2),
					"index_name3":          config.StringVariable(indexName3),
					"read_indices_pattern": config.StringVariable(readIndicesPattern),
				},
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.*", map[string]string{
						"name":               readIndicesPattern,
						"concrete_indices.#": "2",
					}),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("wildcard_target_added"),
				ConfigVariables: config.Variables{
					"alias_name":           config.StringVariable(aliasName),
					"index_name1":          config.StringVariable(indexName1),
					"index_name2":          config.StringVariable(indexName2),
					"index_name3":          config.StringVariable(indexName3),
					"read_indices_pattern": config.StringVariable(readIndicesPattern),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("elasticstack_elasticsearch_index_alias.test_alias", plancheck.ResourceActionUpdate),
					},
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("wildcard_update"),
				ConfigVariables: config.Variables{
					"alias_name":           config.StringVariable(aliasName),
					"index_name1":          config.StringVariable(indexName1),
					"index_name2":          config.StringVariable(indexName2),
					"index_name3":          config.StringVariable(indexName3),
					"read_indices_pattern": config.StringVariable(readIndicesPattern),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.*", map[string]string{
						"name":               readIndicesPattern,
						"concrete_indices.#": "3",
					}),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("wildcard_update"),
				ConfigVariables: config.Variables{
					"alias_name":           config.StringVariable(aliasName),
					"index_name1":          config.StringVariable(indexName1),
					"index_name2":          config.StringVariable(indexName2),
					"index_name3":          config.StringVariable(indexName3),
					"read_indices_pattern": config.StringVariable(readIndicesPattern),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccResourceAliasVirtualStateBecomesConcrete(t *testing.T) {
	suffix := sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlpha)
	aliasName := fmt.Sprintf("alias-virtual-%s", suffix)
	indexName := fmt.Sprintf("virtual-target-%s", suffix)
	readIndicesPattern := fmt.Sprintf("virtual-target-*%s", suffix)
	t.Cleanup(func() { deleteAliasTestIndex(context.Background(), t, indexName) })

	variables := config.Variables{
		"alias_name":           config.StringVariable(aliasName),
		"read_indices_pattern": config.StringVariable(readIndicesPattern),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAliasDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("virtual"),
				ConfigVariables:          variables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						createAliasTestIndexAfterPlan{t: t, indexName: indexName},
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.*", map[string]string{
						"name":               readIndicesPattern,
						"concrete_indices.#": "0",
					}),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("virtual"),
				ConfigVariables:          variables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("elasticstack_elasticsearch_index_alias.test_alias", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.*", map[string]string{
						"name":               readIndicesPattern,
						"concrete_indices.#": "1",
					}),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("virtual"),
				ConfigVariables:          variables,
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       false,
			},
		},
	})
}

func TestAccResourceAliasNoMatchExpression(t *testing.T) {
	suffix := sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlpha)
	aliasName := fmt.Sprintf("alias-no-match-expression-%s", suffix)
	readIndicesExpression := fmt.Sprintf("missing-target-%s*", suffix)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAliasDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("no_match"),
				ConfigVariables: config.Variables{
					"alias_name":              config.StringVariable(aliasName),
					"read_indices_expression": config.StringVariable(readIndicesExpression),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.*", map[string]string{
						"name":               readIndicesExpression,
						"concrete_indices.#": "0",
					}),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("no_match"),
				ConfigVariables: config.Variables{
					"alias_name":              config.StringVariable(aliasName),
					"read_indices_expression": config.StringVariable(readIndicesExpression),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccResourceAliasCommaAndExclusionExpressions(t *testing.T) {
	suffix := sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlpha)
	aliasName := fmt.Sprintf("alias-expression-variants-%s", suffix)
	commaIndexName1 := fmt.Sprintf("comma-target-one-%s", suffix)
	commaIndexName2 := fmt.Sprintf("comma-target-two-%s", suffix)
	exclusionIndexName := fmt.Sprintf("exclusion-target-keep-%s", suffix)
	excludedIndexName := fmt.Sprintf("exclusion-target-skip-%s", suffix)
	commaExpression := fmt.Sprintf("%s,%s", commaIndexName1, commaIndexName2)
	exclusionExpression := fmt.Sprintf("exclusion-target-*-%s,-%s", suffix, excludedIndexName)
	for _, indexName := range []string{commaIndexName1, commaIndexName2, exclusionIndexName, excludedIndexName} {
		createAliasTestIndex(context.Background(), t, indexName)
		t.Cleanup(func() { deleteAliasTestIndex(context.Background(), t, indexName) })
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAliasDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"alias_name":           config.StringVariable(aliasName),
					"comma_expression":     config.StringVariable(commaExpression),
					"exclusion_expression": config.StringVariable(exclusionExpression),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.*", map[string]string{
						"name":               commaExpression,
						"concrete_indices.#": "2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.*", map[string]string{
						"name":               exclusionExpression,
						"concrete_indices.#": "1",
					}),
					assertAliasTestMemberPresent(aliasName, exclusionIndexName),
					assertAliasTestMemberAbsent(aliasName, excludedIndexName),
				),
			},
		},
	})
}

func TestAccResourceAliasBroadExpressionExcludesWriteIndex(t *testing.T) {
	suffix := strings.ToLower(sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlpha))
	aliasName := fmt.Sprintf("alias-write-index-exclusion-%s", suffix)
	writeIndexName := fmt.Sprintf("write-target-current-%s", suffix)
	readIndexName := fmt.Sprintf("write-target-previous-%s", suffix)
	broadExpression := fmt.Sprintf("write-target-*-%s", suffix)
	exclusionExpression := fmt.Sprintf("%s,-%s", broadExpression, writeIndexName)
	for _, indexName := range []string{writeIndexName, readIndexName} {
		createAliasTestIndex(context.Background(), t, indexName)
		t.Cleanup(func() { deleteAliasTestIndex(context.Background(), t, indexName) })
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAliasDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("rejected"),
				ConfigVariables: config.Variables{
					"alias_name":              config.StringVariable(aliasName),
					"write_index_name":        config.StringVariable(writeIndexName),
					"read_indices_expression": config.StringVariable(broadExpression),
				},
				ExpectError: regexp.MustCompile(`(?s)Invalid Configuration.*resolves to write index`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("excluded"),
				ConfigVariables: config.Variables{
					"alias_name":              config.StringVariable(aliasName),
					"write_index_name":        config.StringVariable(writeIndexName),
					"read_indices_expression": config.StringVariable(exclusionExpression),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.name", writeIndexName),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.*", map[string]string{
						"name":               exclusionExpression,
						"concrete_indices.#": "1",
					}),
				),
			},
		},
	})
}

func TestAccResourceAliasHiddenAndClosedTargets(t *testing.T) {
	suffix := strings.ToLower(sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlpha))
	aliasName := fmt.Sprintf("alias-hidden-closed-%s", suffix)
	hiddenIndexName := fmt.Sprintf("hidden-target-%s", suffix)
	closedIndexName := fmt.Sprintf("closed-target-%s", suffix)
	readIndicesExpression := fmt.Sprintf("*-target-%s", suffix)
	ctx := context.Background()
	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		t.Fatalf("acceptance Elasticsearch client: %v", err)
	}
	hidden := "true"
	if _, err := client.GetESClient().Indices.Create(hiddenIndexName).Settings(&estypes.IndexSettings{Hidden: &hidden}).Do(ctx); err != nil {
		t.Fatalf("create hidden index %q: %v", hiddenIndexName, err)
	}
	if _, err := client.GetESClient().Indices.Create(closedIndexName).Do(ctx); err != nil {
		t.Fatalf("create closed index %q: %v", closedIndexName, err)
	}
	if _, err := client.GetESClient().Indices.Close(closedIndexName).Do(ctx); err != nil {
		t.Fatalf("close index %q: %v", closedIndexName, err)
	}
	t.Cleanup(func() { deleteAliasTestIndex(ctx, t, hiddenIndexName) })
	t.Cleanup(func() { deleteAliasTestIndex(ctx, t, closedIndexName) })

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAliasDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"alias_name":              config.StringVariable(aliasName),
					"read_indices_expression": config.StringVariable(readIndicesExpression),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.*", map[string]string{
						"name":               readIndicesExpression,
						"concrete_indices.#": "2",
					}),
				),
			},
		},
	})
}

func TestAccResourceAliasPopulatedToNoMatch(t *testing.T) {
	suffix := sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlpha)
	aliasName := fmt.Sprintf("alias-no-match-%s", suffix)
	indexName := fmt.Sprintf("populated-target-%s", suffix)
	readIndicesPattern := fmt.Sprintf("populated-target-*%s", suffix)
	variables := config.Variables{
		"alias_name":           config.StringVariable(aliasName),
		"index_name":           config.StringVariable(indexName),
		"read_indices_pattern": config.StringVariable(readIndicesPattern),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAliasDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          variables,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.*", map[string]string{
						"name":               readIndicesPattern,
						"concrete_indices.#": "1",
					}),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("no_match"),
				ConfigVariables:          variables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectNonEmptyPlan()},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.*", map[string]string{
						"name":               "no-match-" + readIndicesPattern,
						"concrete_indices.#": "0",
					}),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("no_match"),
				ConfigVariables:          variables,
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       false,
			},
		},
	})
}

func TestAccResourceAliasWriteOnlyToEmpty(t *testing.T) {
	suffix := sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlpha)
	aliasName := fmt.Sprintf("alias-write-only-empty-%s", suffix)
	indexName := fmt.Sprintf("write-only-target-%s", suffix)
	variables := config.Variables{
		"alias_name": config.StringVariable(aliasName),
		"index_name": config.StringVariable(indexName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAliasDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          variables,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.name", indexName),
					assertAliasTestMemberPresent(aliasName, indexName),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("empty"),
				ConfigVariables:          variables,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index"),
					assertAliasTestMemberAbsent(aliasName, indexName),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("empty"),
				ConfigVariables:          variables,
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       false,
			},
		},
	})
}

func TestAccResourceAliasUnconfiguredMemberPlansRemoval(t *testing.T) {
	suffix := sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlpha)
	aliasName := fmt.Sprintf("alias-unconfigured-member-%s", suffix)
	configuredIndexName := fmt.Sprintf("configured-target-%s", suffix)
	unconfiguredIndexName := fmt.Sprintf("unconfigured-target-%s", suffix)
	readIndicesPattern := fmt.Sprintf("configured-target-*%s", suffix)
	variables := config.Variables{
		"alias_name":            config.StringVariable(aliasName),
		"configured_index_name": config.StringVariable(configuredIndexName),
		"read_indices_pattern":  config.StringVariable(readIndicesPattern),
	}
	t.Cleanup(func() { deleteAliasTestIndex(context.Background(), t, unconfiguredIndexName) })

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAliasDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("unconfigured_member"),
				ConfigVariables:          variables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						addAliasTestMemberAfterPlan{t: t, aliasName: aliasName, indexName: unconfiguredIndexName},
					},
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("unconfigured_member"),
				ConfigVariables:          variables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("elasticstack_elasticsearch_index_alias.test_alias", plancheck.ResourceActionUpdate),
					},
				},
				Check: assertAliasTestMemberAbsent(aliasName, unconfiguredIndexName),
			},
		},
	})
}

func TestAccResourceAliasSettingsDriftIsRepairedDuringMembershipUpdate(t *testing.T) {
	suffix := sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlpha)
	aliasName := fmt.Sprintf("alias-settings-drift-%s", suffix)
	indexName1 := fmt.Sprintf("settings-drift-target-one-%s", suffix)
	indexName2 := fmt.Sprintf("settings-drift-target-two-%s", suffix)
	readIndicesPattern := fmt.Sprintf("settings-drift-target-*%s", suffix)
	variables := config.Variables{
		"alias_name":           config.StringVariable(aliasName),
		"index_name1":          config.StringVariable(indexName1),
		"index_name2":          config.StringVariable(indexName2),
		"read_indices_pattern": config.StringVariable(readIndicesPattern),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAliasDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("settings_drift_create"),
				ConfigVariables:          variables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						removeAliasTestMemberFilterAfterPlan{t: t, aliasName: aliasName, indexName: indexName1},
					},
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("settings_drift_create"),
				ConfigVariables:          variables,
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       false,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("settings_drift_add_target"),
				ConfigVariables:          variables,
				ExpectNonEmptyPlan:       true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("settings_drift_add_target"),
				ConfigVariables:          variables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectNonEmptyPlan()},
				},
				Check: assertAliasTestMemberFilter(aliasName, indexName1),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("settings_drift_add_target"),
				ConfigVariables:          variables,
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       false,
			},
		},
	})
}

func TestAccResourceAliasTargetCreatedBetweenPlanAndApply(t *testing.T) {
	suffix := sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlpha)
	aliasName := fmt.Sprintf("alias-plan-apply-race-%s", suffix)
	indexName1 := fmt.Sprintf("plan-apply-target-one-%s", suffix)
	indexName2 := fmt.Sprintf("plan-apply-target-two-%s", suffix)
	readIndicesPattern := fmt.Sprintf("plan-apply-target-*%s", suffix)
	variables := config.Variables{
		"alias_name":           config.StringVariable(aliasName),
		"index_name1":          config.StringVariable(indexName1),
		"read_indices_pattern": config.StringVariable(readIndicesPattern),
	}
	t.Cleanup(func() { deleteAliasTestIndex(context.Background(), t, indexName2) })

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAliasDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("plan_apply_race_create"),
				ConfigVariables:          variables,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("plan_apply_race_update"),
				ConfigVariables:          variables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						createAliasTestIndexBeforeApply{t: t, indexName: indexName2},
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.*", map[string]string{
						"name":               readIndicesPattern,
						"concrete_indices.#": "2",
					}),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("plan_apply_race_update"),
				ConfigVariables:          variables,
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       false,
			},
		},
	})
}

type createAliasTestIndexAfterPlan struct {
	t         *testing.T
	indexName string
}

func (c createAliasTestIndexAfterPlan) CheckPlan(ctx context.Context, _ plancheck.CheckPlanRequest, _ *plancheck.CheckPlanResponse) {
	createAliasTestIndex(ctx, c.t, c.indexName)
}

type createAliasTestIndexBeforeApply struct {
	t         *testing.T
	indexName string
}

func (c createAliasTestIndexBeforeApply) CheckPlan(ctx context.Context, _ plancheck.CheckPlanRequest, _ *plancheck.CheckPlanResponse) {
	createAliasTestIndex(ctx, c.t, c.indexName)
}

type addAliasTestMemberAfterPlan struct {
	t         *testing.T
	aliasName string
	indexName string
}

func (c addAliasTestMemberAfterPlan) CheckPlan(ctx context.Context, _ plancheck.CheckPlanRequest, _ *plancheck.CheckPlanResponse) {
	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		c.t.Fatalf("acceptance Elasticsearch client: %v", err)
	}
	if _, err := client.GetESClient().Indices.Create(c.indexName).Do(ctx); err != nil {
		c.t.Fatalf("create index %q: %v", c.indexName, err)
	}
	diags := esclient.UpdateAliasesAtomic(ctx, client, []esclient.AliasAction{{
		Type:  "add",
		Index: c.indexName,
		Alias: c.aliasName,
	}})
	if diags.HasError() {
		c.t.Fatalf("add alias member %q: %s", c.indexName, diags.Errors()[0].Detail())
	}
}

type removeAliasTestMemberFilterAfterPlan struct {
	t         *testing.T
	aliasName string
	indexName string
}

func (c removeAliasTestMemberFilterAfterPlan) CheckPlan(ctx context.Context, _ plancheck.CheckPlanRequest, _ *plancheck.CheckPlanResponse) {
	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		c.t.Fatalf("acceptance Elasticsearch client: %v", err)
	}
	diags := esclient.UpdateAliasesAtomic(ctx, client, []esclient.AliasAction{
		{
			Type:  "remove",
			Index: c.indexName,
			Alias: c.aliasName,
		},
		{
			Type:  "add",
			Index: c.indexName,
			Alias: c.aliasName,
		},
	})
	if diags.HasError() {
		c.t.Fatalf("remove alias member filter %q: %s", c.indexName, diags.Errors()[0].Detail())
	}
	indices, readDiags := esclient.GetAlias(ctx, client, c.aliasName)
	if readDiags.HasError() {
		c.t.Fatalf("get alias %q: %s", c.aliasName, readDiags.Errors()[0].Detail())
	}
	alias, found := indices[c.indexName].Aliases[c.aliasName]
	if !found {
		c.t.Fatalf("alias %q is not attached to %q", c.aliasName, c.indexName)
	}
	if alias.Filter != nil {
		c.t.Fatalf("alias filter on %q = %#v, want nil", c.indexName, alias.Filter)
	}
}

func createAliasTestIndex(ctx context.Context, t *testing.T, indexName string) {
	t.Helper()

	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		t.Fatalf("acceptance Elasticsearch client: %v", err)
	}
	if _, err := client.GetESClient().Indices.Create(indexName).Do(ctx); err != nil {
		t.Fatalf("create index %q: %v", indexName, err)
	}
}

func assertAliasTestMemberFilter(aliasName, indexName string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		ctx := context.Background()
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			return fmt.Errorf("acceptance Elasticsearch client: %w", err)
		}
		indices, diags := esclient.GetAlias(ctx, client, aliasName)
		if diags.HasError() {
			return fmt.Errorf("get alias %q: %s", aliasName, diags.Errors()[0].Detail())
		}
		alias, found := indices[indexName].Aliases[aliasName]
		if !found {
			return fmt.Errorf("alias %q is not attached to %q", aliasName, indexName)
		}
		filter, filterDiags := aliasutil.NormalizeAliasFilterAnyToMap(alias.Filter)
		if filterDiags.HasError() {
			return fmt.Errorf("normalize alias filter: %s", filterDiags.Errors()[0].Detail())
		}
		expected := map[string]any{"term": map[string]any{"status": "published"}}
		if !reflect.DeepEqual(filter, expected) {
			return fmt.Errorf("alias filter = %#v, want %#v", filter, expected)
		}
		return nil
	}
}

func assertAliasTestMemberAbsent(aliasName, indexName string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		ctx := context.Background()
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			return fmt.Errorf("acceptance Elasticsearch client: %w", err)
		}
		indices, diags := esclient.GetAlias(ctx, client, aliasName)
		if diags.HasError() {
			return fmt.Errorf("get alias %q: %s", aliasName, diags.Errors()[0].Detail())
		}
		if _, found := indices[indexName].Aliases[aliasName]; found {
			return fmt.Errorf("alias %q is still attached to %q", aliasName, indexName)
		}
		return nil
	}
}

func assertAliasTestMemberPresent(aliasName, indexName string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		ctx := context.Background()
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			return fmt.Errorf("acceptance Elasticsearch client: %w", err)
		}
		indices, diags := esclient.GetAlias(ctx, client, aliasName)
		if diags.HasError() {
			return fmt.Errorf("get alias %q: %s", aliasName, diags.Errors()[0].Detail())
		}
		if _, found := indices[indexName].Aliases[aliasName]; !found {
			return fmt.Errorf("alias %q is not attached to %q", aliasName, indexName)
		}
		return nil
	}
}

func deleteAliasTestIndex(ctx context.Context, t *testing.T, indexName string) {
	t.Helper()

	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		t.Logf("cleanup Elasticsearch client: %v", err)
		return
	}
	if _, err := client.GetESClient().Indices.Delete(indexName).Do(ctx); err != nil && !esclient.IsNotFoundElasticsearchError(err) {
		t.Logf("cleanup index %q: %v", indexName, err)
	}
}

func checkResourceAliasDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
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

		typedClient := client.GetESClient()

		_, err = typedClient.Indices.GetAlias().Name(aliasName).Do(context.Background())
		if err != nil {
			if esclient.IsNotFoundElasticsearchError(err) {
				continue
			}
			return err
		}

		return fmt.Errorf("Alias (%s) still exists", aliasName)
	}
	return nil
}
