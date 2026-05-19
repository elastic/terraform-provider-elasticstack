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

package indexmappings_test

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	esclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const mappingsResourceName = "elasticstack_elasticsearch_index_mappings.test"

func TestAccResourceIndexMappings_basic(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexMappingsDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("apply"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mappingsResourceName, "index", indexName),
					resource.TestCheckResourceAttrSet(mappingsResourceName, "id"),
					checkStateMappingsContainFields("title"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("apply"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccResourceIndexMappings_update(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexMappingsDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("one_field"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("two_fields"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					checkStateMappingsContainFields("title", "body"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("two_fields"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccResourceIndexMappings_drift(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexMappingsDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("apply"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("apply"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				PreConfig: func() {
					addDynamicMappingField(t, indexName, "tags", map[string]any{"type": "keyword"})
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccResourceIndexMappings_allTopLevelKeys(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexMappingsDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("apply"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mappingsResourceName, "index", indexName),
					resource.TestCheckResourceAttrSet(mappingsResourceName, "mappings"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("apply"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccResourceIndexMappings_destroyIsNoop(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexMappingsDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_mappings"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("index_only"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					checkIndexMappingsContainField(indexName, "title"),
				),
			},
		},
	})
}

func TestAccResourceIndexMappings_indexNotFound(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("apply"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				ExpectError: regexp.MustCompile(`Index not found`),
			},
		},
	})
}

func checkStateMappingsContainFields(fields ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[mappingsResourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", mappingsResourceName)
		}

		rawMappings, ok := rs.Primary.Attributes["mappings"]
		if !ok {
			return fmt.Errorf("resource %s has no mappings attribute", mappingsResourceName)
		}

		var mappings map[string]any
		if err := json.Unmarshal([]byte(rawMappings), &mappings); err != nil {
			return fmt.Errorf("failed to unmarshal state mappings: %w", err)
		}

		properties, ok := mappings["properties"].(map[string]any)
		if !ok {
			return fmt.Errorf("state mappings have no properties")
		}

		for _, field := range fields {
			if _, ok := properties[field]; !ok {
				return fmt.Errorf("state mappings missing field %q", field)
			}
		}
		return nil
	}
}

func addDynamicMappingField(t *testing.T, indexName, fieldName string, fieldMapping map[string]any) {
	t.Helper()

	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		t.Fatalf("failed to create Elasticsearch client: %s", err)
	}

	payload := map[string]any{
		"properties": map[string]any{
			fieldName: fieldMapping,
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal dynamic mapping: %s", err)
	}

	diags := esclient.UpdateIndexMappings(context.Background(), client, indexName, string(body))
	if diags.HasError() {
		t.Fatalf("failed to add dynamic mapping field: %v", diags)
	}
}

func checkIndexMappingsContainField(indexName, fieldName string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			return err
		}

		indexState, diags := esclient.GetIndex(context.Background(), client, indexName)
		if diags.HasError() {
			return fmt.Errorf("failed to get index %q: %v", indexName, diags)
		}
		if indexState == nil {
			return fmt.Errorf("index %q not found after destroy", indexName)
		}
		if indexState.Mappings == nil {
			return fmt.Errorf("index %q has no mappings after destroy", indexName)
		}

		mappingBytes, err := json.Marshal(indexState.Mappings)
		if err != nil {
			return err
		}

		var mappings map[string]any
		if err := json.Unmarshal(mappingBytes, &mappings); err != nil {
			return err
		}

		properties, ok := mappings["properties"].(map[string]any)
		if !ok {
			return fmt.Errorf("index %q mappings have no properties", indexName)
		}
		if _, ok := properties[fieldName]; !ok {
			return fmt.Errorf("index %q mappings missing field %q after destroy", indexName, fieldName)
		}
		return nil
	}
}

func checkResourceIndexMappingsDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_index" {
			continue
		}

		compID, compDiags := clients.CompositeIDFromStr(rs.Primary.ID)
		if compDiags.HasError() {
			return fmt.Errorf("failed to parse composite ID %q: %v", rs.Primary.ID, compDiags)
		}
		indexName := compID.ResourceID

		indexState, diags := esclient.GetIndex(context.Background(), client, indexName)
		if diags.HasError() {
			return fmt.Errorf("failed to get index %q: %v", indexName, diags)
		}
		if indexState != nil {
			if delDiags := esclient.DeleteIndex(context.Background(), client, indexName); delDiags.HasError() {
				return fmt.Errorf("failed to delete index %q: %v", indexName, delDiags)
			}
		}
	}
	return nil
}
