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

package synonyms_test

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	esclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// minSupportedVersion is the minimum Elasticsearch version that supports the Synonyms API.
var minSupportedVersion = version.Must(version.NewVersion("8.10.0"))

// TestAccResourceSynonymSet covers basic CRUD and rule ordering:
//   - Step 1 (create): verify state reflects the two initial rules with explicit IDs and
//     correct synonyms strings, preserving declaration order.
//   - Step 2 (update): add a third rule, modify an existing rule's synonyms, verify the
//     updated state and that the original rule order is preserved.
func TestAccResourceSynonymSet(t *testing.T) {
	synonymSetID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkSynonymSetDestroy(synonymSetID),
		Steps: []resource.TestStep{
			// Step 1: Create with two explicit-ID rules.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"synonym_set_id": config.StringVariable(synonymSetID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_synonym_set.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_synonym_set.test", "synonym_set_id", synonymSetID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_synonym_set.test", "synonyms_set.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_synonym_set.test", "synonyms_set.0.id", "rule-1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_synonym_set.test", "synonyms_set.0.synonyms", "i-pod, i pod => ipod"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_synonym_set.test", "synonyms_set.1.id", "rule-2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_synonym_set.test", "synonyms_set.1.synonyms", "universe, cosmos"),
				),
			},
			// Step 2: Update — modify rule-2's synonyms and add a new rule-3.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          config.Variables{"synonym_set_id": config.StringVariable(synonymSetID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_synonym_set.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_synonym_set.test", "synonym_set_id", synonymSetID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_synonym_set.test", "synonyms_set.#", "3"),
					// Rule ordering must be preserved: rule-1 first, rule-2 second, rule-3 third.
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_synonym_set.test", "synonyms_set.0.id", "rule-1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_synonym_set.test", "synonyms_set.0.synonyms", "i-pod, i pod => ipod"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_synonym_set.test", "synonyms_set.1.id", "rule-2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_synonym_set.test", "synonyms_set.1.synonyms", "universe, cosmos, world"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_synonym_set.test", "synonyms_set.2.id", "rule-3"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_synonym_set.test", "synonyms_set.2.synonyms", "laptop, notebook"),
				),
			},
		},
	})
}

// TestAccResourceSynonymSetOptionalRuleID verifies that when a rule's `id` is
// omitted from the config, the provider generates a stable UUID. A subsequent
// plan against the same config must show no diff.
func TestAccResourceSynonymSetOptionalRuleID(t *testing.T) {
	synonymSetID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkSynonymSetDestroy(synonymSetID),
		Steps: []resource.TestStep{
			// Step 1: Create — first rule has no explicit id (provider must generate one).
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"synonym_set_id": config.StringVariable(synonymSetID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_synonym_set.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_synonym_set.test", "synonyms_set.#", "2"),
					// The first rule had no id in config; the provider must have generated one.
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_synonym_set.test", "synonyms_set.0.id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_synonym_set.test", "synonyms_set.0.synonyms", "quick, fast, speedy"),
					// The second rule has an explicit id.
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_synonym_set.test", "synonyms_set.1.id", "explicit-rule"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_synonym_set.test", "synonyms_set.1.synonyms", "slow, sluggish"),
				),
			},
			// Step 2: Re-apply the same config and verify there is no diff. The
			// provider must not regenerate the UUID on every plan.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"synonym_set_id": config.StringVariable(synonymSetID)},
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       false,
			},
		},
	})
}

// TestAccResourceSynonymSetImport verifies that a synonym set resource can be
// imported by its composite ID (<cluster_uuid>/<synonym_set_id>) and that the
// resulting state matches the pre-import state. A subsequent plan must show no diff.
func TestAccResourceSynonymSetImport(t *testing.T) {
	synonymSetID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkSynonymSetDestroy(synonymSetID),
		Steps: []resource.TestStep{
			// Step 1: Create the resource.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"synonym_set_id": config.StringVariable(synonymSetID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_synonym_set.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_synonym_set.test", "synonym_set_id", synonymSetID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_synonym_set.test", "synonyms_set.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_synonym_set.test", "synonyms_set.0.id", "rule-a"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_synonym_set.test", "synonyms_set.0.synonyms", "dog, hound, canine"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_synonym_set.test", "synonyms_set.1.id", "rule-b"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_synonym_set.test", "synonyms_set.1.synonyms", "cat, feline"),
				),
			},
			// Step 2: Import by composite ID retrieved from state and verify state matches.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"synonym_set_id": config.StringVariable(synonymSetID)},
				ResourceName:             "elasticstack_elasticsearch_synonym_set.test",
				ImportStateIdFunc:        synonymSetImportID("elasticstack_elasticsearch_synonym_set.test"),
				ImportState:              true,
				ImportStateVerify:        true,
			},
		},
	})
}

// TestAccDataSourceSynonymSetNotFound verifies that looking up a non-existent synonym set
// produces a clear "Synonym set not found" error diagnostic.
func TestAccDataSourceSynonymSetNotFound(t *testing.T) {
	nonExistentID := "does-not-exist-" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("not-found"),
				ConfigVariables:          config.Variables{"synonym_set_id": config.StringVariable(nonExistentID)},
				ExpectError:              regexp.MustCompile(`Synonym set not found`),
			},
		},
	})
}

// TestAccDataSourceSynonymSetCountVariety verifies that the data source correctly reflects
// synonym sets with different numbers of rules (1 rule, then 3 rules).
func TestAccDataSourceSynonymSetCountVariety(t *testing.T) {
	synonymSetID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// Step 1: read a synonym set with exactly 1 rule.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("one-rule"),
				ConfigVariables:          config.Variables{"synonym_set_id": config.StringVariable(synonymSetID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_synonym_set.test", "synonyms_set.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_synonym_set.test", "synonyms_set.0.id", "only-rule"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_synonym_set.test", "synonyms_set.0.synonyms", "big, large"),
				),
			},
			// Step 2: update the same synonym set to 3 rules and verify the count.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("three-rules"),
				ConfigVariables:          config.Variables{"synonym_set_id": config.StringVariable(synonymSetID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_synonym_set.test", "synonyms_set.#", "3"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_synonym_set.test", "synonyms_set.0.id", "rule-a"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_synonym_set.test", "synonyms_set.0.synonyms", "cat, feline"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_synonym_set.test", "synonyms_set.1.id", "rule-b"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_synonym_set.test", "synonyms_set.1.synonyms", "dog, hound, canine"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_synonym_set.test", "synonyms_set.2.id", "rule-c"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_synonym_set.test", "synonyms_set.2.synonyms", "happy, joyful, cheerful"),
				),
			},
		},
	})
}

// TestAccDataSourceSynonymSet verifies that the synonym set data source reads a
// resource created by the synonym set resource and reflects all attributes correctly.
func TestAccDataSourceSynonymSet(t *testing.T) {
	synonymSetID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables:          config.Variables{"synonym_set_id": config.StringVariable(synonymSetID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_synonym_set.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_synonym_set.test", "synonym_set_id", synonymSetID),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_synonym_set.test", "synonyms_set.#", "2"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_synonym_set.test", "synonyms_set.0.id", "rule-1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_synonym_set.test", "synonyms_set.0.synonyms", "i-pod, i pod => ipod"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_synonym_set.test", "synonyms_set.1.id", "rule-2"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_synonym_set.test", "synonyms_set.1.synonyms", "universe, cosmos"),
				),
			},
		},
	})
}

// TestAccResourceSynonymSetDeleteWhileInUse verifies that attempting to delete a
// synonym set referenced by an index analyzer returns a clear error diagnostic
// (REQ-008) rather than silently failing or panicking.
func TestAccResourceSynonymSetDeleteWhileInUse(t *testing.T) {
	// Guard the whole test here so PreConfig never runs on unsupported versions.
	// Step-level SkipFunc alone is not sufficient because PreConfig executes
	// regardless of SkipFunc, and createIndexWithSynonymFilter would fail on
	// ES < 8.10.0 where the synonyms_set filter option does not exist.
	notSupported, err := versionutils.CheckIfVersionIsUnsupported(minSupportedVersion)()
	if err != nil {
		t.Fatalf("could not determine server version: %v", err)
	}
	if notSupported {
		t.Skipf("skipping: requires Elasticsearch >= %s (Synonyms API)", minSupportedVersion)
	}

	synonymSetID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	indexName := "test-synonym-in-use-" + sdkacctest.RandStringFromCharSet(6, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkSynonymSetDestroy(synonymSetID),
		Steps: []resource.TestStep{
			// Step 1: Create the synonym set.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"synonym_set_id": config.StringVariable(synonymSetID)},
			},
			// Step 2: Create an ES index outside Terraform with a synonym token filter
			// referencing the set, then try to destroy it — must fail with a clear error.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				PreConfig: func() {
					if err := createIndexWithSynonymFilter(indexName, synonymSetID); err != nil {
						t.Fatalf("failed to create index %s: %s", indexName, err)
					}
				},
				ConfigDirectory: acctest.NamedTestCaseDirectory("destroy"),
				ExpectError:     regexp.MustCompile(`Cannot delete synonym set`),
			},
			// Step 3: Remove the blocking index so Terraform can successfully destroy the synonym set.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				PreConfig: func() {
					if err := deleteIndexForTest(indexName); err != nil {
						t.Fatalf("failed to delete index %s: %s", indexName, err)
					}
				},
				ConfigDirectory: acctest.NamedTestCaseDirectory("destroy"),
			},
		},
	})
}

func createIndexWithSynonymFilter(indexName, synonymSetID string) error {
	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		return err
	}
	body := fmt.Sprintf(`{
		"settings": {
			"index": {
				"analysis": {
					"filter": {
						"synonym_filter": {
							"type": "synonym",
							"synonyms_set": %q,
							"updateable": true
						}
					},
					"analyzer": {
						"synonym_analyzer": {
							"tokenizer": "standard",
							"filter": ["lowercase", "synonym_filter"]
						}
					}
				}
			}
		}
	}`, synonymSetID)
	_, err = client.GetESClient().Indices.Create(indexName).Raw(bytes.NewReader([]byte(body))).Do(context.Background())
	return err
}

func deleteIndexForTest(indexName string) error {
	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		return err
	}
	_, err = client.GetESClient().Indices.Delete(indexName).Do(context.Background())
	return err
}

// synonymSetImportID returns an ImportStateIdFunc that retrieves the composite
// resource ID (<cluster_uuid>/<synonym_set_id>) from the Terraform state. This
// matches the ID format used by the synonym set resource.
func synonymSetImportID(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource %s not found in state", resourceName)
		}
		return rs.Primary.ID, nil
	}
}

// checkSynonymSetDestroy returns a check function that verifies the synonym set
// has been deleted from Elasticsearch after the test completes.
func checkSynonymSetDestroy(synonymSetID string) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			return err
		}

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "elasticstack_elasticsearch_synonym_set" {
				continue
			}
			compID, _ := clients.CompositeIDFromStr(rs.Primary.ID)
			if compID.ResourceID != synonymSetID {
				continue
			}

			rules, diags := esclient.GetSynonymSet(context.Background(), client, synonymSetID)
			if diags.HasError() {
				return fmt.Errorf("error checking synonym set deletion: %s", diags[0].Summary())
			}
			if rules != nil {
				return fmt.Errorf("synonym set (%s) still exists", synonymSetID)
			}
		}
		return nil
	}
}
