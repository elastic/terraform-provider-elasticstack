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

package queryrulesets_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	esclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/queryrulesets"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const queryRulesetResourceAddr = "elasticstack_elasticsearch_query_ruleset.test"

// TestAccResourceQueryRuleset covers basic CRUD:
//   - Step 1 (create): one pinned and one exclude rule with ids actions.
//   - Step 2 (update): modify a criterion on rule-2 and add rule-3.
func TestAccResourceQueryRuleset(t *testing.T) {
	rulesetID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	vars := config.Variables{"ruleset_id": config.StringVariable(rulesetID)}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkQueryRulesetDestroy(rulesetID),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(queryrulesets.MinSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(queryRulesetResourceAddr, "id"),
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "ruleset_id", rulesetID),
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "rules.#", "2"),
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "rules.0.rule_id", "rule-1"),
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "rules.0.type", "pinned"),
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "rules.0.priority", "1"),
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "rules.0.criteria.#", "1"),
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "rules.0.criteria.0.type", "exact"),
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "rules.0.criteria.0.metadata", "query"),
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "rules.0.criteria.0.values", `["laptop","notebook"]`),
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "rules.0.actions.ids.#", "2"),
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "rules.0.actions.ids.0", "doc-1"),
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "rules.0.actions.ids.1", "doc-2"),
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "rules.1.rule_id", "rule-2"),
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "rules.1.type", "exclude"),
					resource.TestCheckNoResourceAttr(queryRulesetResourceAddr, "rules.1.priority"),
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "rules.1.criteria.0.type", "contains"),
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "rules.1.criteria.0.values", `["deprecated"]`),
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "rules.1.actions.ids.#", "1"),
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "rules.1.actions.ids.0", "doc-old"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(queryrulesets.MinSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "rules.#", "3"),
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "rules.0.rule_id", "rule-1"),
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "rules.1.rule_id", "rule-2"),
					resource.TestCheckNoResourceAttr(queryRulesetResourceAddr, "rules.1.priority"),
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "rules.1.criteria.0.values", `["deprecated","obsolete"]`),
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "rules.2.rule_id", "rule-3"),
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "rules.2.type", "pinned"),
					resource.TestCheckNoResourceAttr(queryRulesetResourceAddr, "rules.2.priority"),
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "rules.2.criteria.0.type", "always"),
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "rules.2.actions.ids.0", "doc-3"),
				),
			},
		},
	})
}

// TestAccResourceQueryRulesetOrdering verifies declaration order is preserved and
// a subsequent plan against the same config shows no diff.
func TestAccResourceQueryRulesetOrdering(t *testing.T) {
	rulesetID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	vars := config.Variables{"ruleset_id": config.StringVariable(rulesetID)}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkQueryRulesetDestroy(rulesetID),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(queryrulesets.MinSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "rules.#", "3"),
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "rules.0.rule_id", "rule-1"),
					resource.TestCheckNoResourceAttr(queryRulesetResourceAddr, "rules.0.priority"),
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "rules.1.rule_id", "rule-2"),
					resource.TestCheckNoResourceAttr(queryRulesetResourceAddr, "rules.1.priority"),
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "rules.2.rule_id", "rule-3"),
					resource.TestCheckNoResourceAttr(queryRulesetResourceAddr, "rules.2.priority"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(queryrulesets.MinSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       false,
			},
		},
	})
}

// TestAccResourceQueryRulesetNumericCriteriaValues verifies numeric criteria values
// round-trip as a normalized JSON array string in state.
func TestAccResourceQueryRulesetNumericCriteriaValues(t *testing.T) {
	rulesetID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	vars := config.Variables{"ruleset_id": config.StringVariable(rulesetID)}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkQueryRulesetDestroy(rulesetID),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(queryrulesets.MinSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "rules.0.criteria.0.type", "gt"),
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "rules.0.criteria.0.metadata", "popularity"),
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "rules.0.criteria.0.values", "[100]"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(queryrulesets.MinSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       false,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(queryrulesets.MinSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "rules.0.criteria.0.values", "[200]"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(queryrulesets.MinSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          vars,
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       false,
			},
		},
	})
}

// TestAccResourceQueryRulesetActionsDocs verifies a rule using actions.docs instead of ids.
func TestAccResourceQueryRulesetActionsDocs(t *testing.T) {
	rulesetID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	vars := config.Variables{"ruleset_id": config.StringVariable(rulesetID)}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkQueryRulesetDestroy(rulesetID),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(queryrulesets.MinSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "rules.0.actions.docs.#", "1"),
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "rules.0.actions.docs.0._index", "my-index"),
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "rules.0.actions.docs.0._id", "42"),
					resource.TestCheckNoResourceAttr(queryRulesetResourceAddr, "rules.0.actions.ids.#"),
				),
			},
		},
	})
}

// TestAccResourceQueryRulesetCriteriaAlways verifies an always criterion without metadata or values.
func TestAccResourceQueryRulesetCriteriaAlways(t *testing.T) {
	rulesetID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	vars := config.Variables{"ruleset_id": config.StringVariable(rulesetID)}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkQueryRulesetDestroy(rulesetID),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(queryrulesets.MinSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "rules.0.criteria.0.type", "always"),
					resource.TestCheckNoResourceAttr(queryRulesetResourceAddr, "rules.0.criteria.0.metadata"),
					resource.TestCheckNoResourceAttr(queryRulesetResourceAddr, "rules.0.criteria.0.values"),
				),
			},
		},
	})
}

// TestAccResourceQueryRulesetImport verifies import by composite ID and a clean post-import plan.
// Rules are declared in rule_id ascending order so import sorting matches config order.
func TestAccResourceQueryRulesetImport(t *testing.T) {
	rulesetID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	vars := config.Variables{"ruleset_id": config.StringVariable(rulesetID)}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkQueryRulesetDestroy(rulesetID),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(queryrulesets.MinSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(queryRulesetResourceAddr, "id"),
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "ruleset_id", rulesetID),
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "rules.#", "2"),
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "rules.0.rule_id", "rule-a"),
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "rules.1.rule_id", "rule-b"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(queryrulesets.MinSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				ResourceName:             queryRulesetResourceAddr,
				ImportStateIdFunc:        queryRulesetImportID(queryRulesetResourceAddr),
				ImportState:              true,
				ImportStateVerify:        true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(queryrulesets.MinSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       false,
			},
		},
	})
}

// TestAccDataSourceQueryRuleset verifies the data source reflects a resource-managed ruleset.
func TestAccDataSourceQueryRuleset(t *testing.T) {
	rulesetID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	vars := config.Variables{"ruleset_id": config.StringVariable(rulesetID)}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(queryrulesets.MinSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_query_ruleset.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_query_ruleset.test", "ruleset_id", rulesetID),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_query_ruleset.test", "rules.#", "2"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_query_ruleset.test", "rules.0.rule_id", "rule-1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_query_ruleset.test", "rules.0.type", "pinned"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_query_ruleset.test", "rules.0.priority", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_query_ruleset.test", "rules.0.criteria.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_query_ruleset.test", "rules.0.criteria.0.type", "exact"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_query_ruleset.test", "rules.0.criteria.0.metadata", "query"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_query_ruleset.test", "rules.0.criteria.0.values", `["laptop","notebook"]`),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_query_ruleset.test", "rules.0.actions.ids.#", "2"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_query_ruleset.test", "rules.0.actions.ids.0", "doc-1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_query_ruleset.test", "rules.0.actions.ids.1", "doc-2"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_query_ruleset.test", "rules.1.rule_id", "rule-2"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_query_ruleset.test", "rules.1.type", "exclude"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_query_ruleset.test", "rules.1.criteria.0.type", "contains"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_query_ruleset.test", "rules.1.criteria.0.values", `["deprecated"]`),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_query_ruleset.test", "rules.1.actions.ids.0", "doc-old"),
				),
			},
		},
	})
}

// TestAccDataSourceQueryRulesetDocs verifies the data source correctly surfaces a docs-based ruleset.
func TestAccDataSourceQueryRulesetDocs(t *testing.T) {
	rulesetID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	vars := config.Variables{"ruleset_id": config.StringVariable(rulesetID)}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(queryrulesets.MinSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_query_ruleset.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_query_ruleset.test", "ruleset_id", rulesetID),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_query_ruleset.test", "rules.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_query_ruleset.test", "rules.0.rule_id", "docs-rule"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_query_ruleset.test", "rules.0.type", "pinned"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_query_ruleset.test", "rules.0.actions.docs.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_query_ruleset.test", "rules.0.actions.docs.0._index", "my-index"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_query_ruleset.test", "rules.0.actions.docs.0._id", "42"),
				),
			},
		},
	})
}

// TestAccResourceQueryRulesetNotFound verifies that when a ruleset is deleted outside
// Terraform, refresh removes the resource from state and the next plan shows recreation.
func TestAccResourceQueryRulesetNotFound(t *testing.T) {
	// Guard the whole test so PreConfig never runs on unsupported versions.
	// Step-level SkipFunc alone is not sufficient because PreConfig executes
	// regardless of SkipFunc when an earlier step was skipped.
	notSupported, err := versionutils.CheckIfVersionIsUnsupported(queryrulesets.MinSupportedVersion)()
	if err != nil {
		t.Fatalf("could not determine server version: %v", err)
	}
	if notSupported {
		t.Skipf("skipping: requires Elasticsearch >= %s (Query Rules API)", queryrulesets.MinSupportedVersion)
	}

	rulesetID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	vars := config.Variables{"ruleset_id": config.StringVariable(rulesetID)}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(queryrulesets.MinSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(queryRulesetResourceAddr, "ruleset_id", rulesetID),
					resource.TestCheckResourceAttrSet(queryRulesetResourceAddr, "id"),
				),
			},
			{
				PreConfig: func() {
					ctx := context.Background()
					client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
					if err != nil {
						t.Fatalf("acceptance elasticsearch client: %v", err)
					}
					diags := esclient.DeleteQueryRuleset(ctx, client, rulesetID)
					if diags.HasError() {
						t.Fatalf("delete query ruleset %q before refresh: %s", rulesetID, diags[0].Summary())
					}
				},
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(queryrulesets.MinSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       true,
				Check:                    testAccCheckQueryRulesetAbsentFromState(queryRulesetResourceAddr),
			},
		},
	})
}

func queryRulesetImportID(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource %s not found in state", resourceName)
		}
		return rs.Primary.ID, nil
	}
}

func checkQueryRulesetDestroy(rulesetID string) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			return err
		}

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "elasticstack_elasticsearch_query_ruleset" {
				continue
			}
			compID, _ := clients.CompositeIDFromStr(rs.Primary.ID)
			if compID.ResourceID != rulesetID {
				continue
			}

			resp, diags := esclient.GetQueryRuleset(context.Background(), client, rulesetID)
			if diags.HasError() {
				return fmt.Errorf("error checking query ruleset deletion: %s", diags[0].Summary())
			}
			if resp != nil {
				return fmt.Errorf("query ruleset (%s) still exists", rulesetID)
			}
		}
		return nil
	}
}

func testAccCheckQueryRulesetAbsentFromState(addr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if _, ok := s.RootModule().Resources[addr]; ok {
			return fmt.Errorf("expected %q to be absent from state after refresh (ruleset deleted out-of-band)", addr)
		}
		return nil
	}
}
