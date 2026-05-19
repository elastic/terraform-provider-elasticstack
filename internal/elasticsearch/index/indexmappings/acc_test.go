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
	"strings"
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

var indexMappingsIDRegexp = regexp.MustCompile(`^[A-Za-z0-9_-]+/.+$`)

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
					resource.TestMatchResourceAttr(mappingsResourceName, "id", indexMappingsIDRegexp),
					checkStateMappingsProperties([]string{"title"}, nil),
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
				Check: resource.ComposeTestCheckFunc(
					checkStateMappingsProperties([]string{"title"}, nil),
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
						plancheck.ExpectNonEmptyPlan(),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					checkStateMappingsProperties([]string{"title", "body"}, nil),
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
				Check: resource.ComposeTestCheckFunc(
					checkStateMappingsProperties([]string{"title"}, []string{"tags"}),
				),
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
					checkStateMappingsTopLevelKeys("dynamic", "_source", "dynamic_templates", "runtime", "properties"),
					checkStateMappingsDynamic(false),
					checkStateMappingsSourceEnabled(true),
					checkStateMappingsRuntimeFields("day_of_week"),
					checkStateMappingsDynamicTemplates(1),
					checkStateMappingsProperties([]string{"title"}, nil),
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

func TestAccResourceIndexMappings_import(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexMappingsDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("precondition"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("full"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				ResourceName:       mappingsResourceName,
				ImportState:        true,
				ImportStatePersist: true,
				ImportStateIdFunc:  importStateIDForIndexName(indexName),
				// First import: no prior mappings resource in state for ImportStateVerify.
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(mappingsResourceName, "id", indexMappingsIDRegexp),
					resource.TestCheckResourceAttr(mappingsResourceName, "index", indexName),
					checkStateMappingsProperties([]string{"title", "body"}, nil),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("full"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				ResourceName:            mappingsResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIDForIndexName(indexName),
				ImportStateVerifyIgnore: []string{"id"},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("narrow"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
					},
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("narrow"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					checkStateMappingsProperties([]string{"title"}, []string{"body"}),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("narrow"),
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

func TestAccResourceIndexMappings_indexDeletedOutOfBand(t *testing.T) {
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
					deleteIndexOutOfBand(t, indexName)
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
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

func stateMappingsFromResource(s *terraform.State) (map[string]any, error) {
	rs, ok := s.RootModule().Resources[mappingsResourceName]
	if !ok {
		return nil, fmt.Errorf("resource %s not found in state", mappingsResourceName)
	}

	rawMappings, ok := rs.Primary.Attributes["mappings"]
	if !ok {
		return nil, fmt.Errorf("resource %s has no mappings attribute", mappingsResourceName)
	}

	var mappings map[string]any
	if err := json.Unmarshal([]byte(rawMappings), &mappings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state mappings: %w", err)
	}
	return mappings, nil
}

func checkStateMappingsProperties(present, absent []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		mappings, err := stateMappingsFromResource(s)
		if err != nil {
			return err
		}

		properties, ok := mappings["properties"].(map[string]any)
		if !ok {
			return fmt.Errorf("state mappings have no properties")
		}

		for _, field := range present {
			if _, ok := properties[field]; !ok {
				return fmt.Errorf("state mappings missing field %q", field)
			}
		}
		for _, field := range absent {
			if _, ok := properties[field]; ok {
				return fmt.Errorf("state mappings unexpectedly contain field %q", field)
			}
		}
		return nil
	}
}

func checkStateMappingsTopLevelKeys(keys ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		mappings, err := stateMappingsFromResource(s)
		if err != nil {
			return err
		}
		for _, key := range keys {
			if _, ok := mappings[key]; !ok {
				return fmt.Errorf("state mappings missing top-level key %q", key)
			}
		}
		return nil
	}
}

func checkStateMappingsDynamic(expected bool) resource.TestCheckFunc {
	return checkStateMappingsBoolAt(expected, "dynamic")
}

func checkStateMappingsSourceEnabled(expected bool) resource.TestCheckFunc {
	return checkStateMappingsBoolAt(expected, "_source", "enabled")
}

// checkStateMappingsBoolAt asserts that the value at the given nested path in
// state mappings is the expected bool. Elasticsearch echoes some boolean
// fields as JSON strings; both forms are accepted.
func checkStateMappingsBoolAt(expected bool, path ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		mappings, err := stateMappingsFromResource(s)
		if err != nil {
			return err
		}
		pathStr := strings.Join(path, ".")
		node := any(mappings)
		for i, key := range path {
			m, ok := node.(map[string]any)
			if !ok {
				return fmt.Errorf("state mappings %q is not an object", strings.Join(path[:i], "."))
			}
			v, ok := m[key]
			if !ok {
				return fmt.Errorf("state mappings missing key %q", strings.Join(path[:i+1], "."))
			}
			node = v
		}
		switch v := node.(type) {
		case bool:
			if v != expected {
				return fmt.Errorf("state mappings %s = %v, want %v", pathStr, v, expected)
			}
		case string:
			want := "false"
			if expected {
				want = "true"
			}
			if v != want {
				return fmt.Errorf("state mappings %s = %q, want %q", pathStr, v, want)
			}
		default:
			return fmt.Errorf("state mappings %s has unexpected type %T", pathStr, node)
		}
		return nil
	}
}

func checkStateMappingsRuntimeFields(names ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		mappings, err := stateMappingsFromResource(s)
		if err != nil {
			return err
		}
		runtime, ok := mappings["runtime"].(map[string]any)
		if !ok {
			return fmt.Errorf("state mappings runtime is not an object")
		}
		for _, name := range names {
			if _, ok := runtime[name]; !ok {
				return fmt.Errorf("state mappings runtime missing field %q", name)
			}
		}
		return nil
	}
}

func checkStateMappingsDynamicTemplates(minCount int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		mappings, err := stateMappingsFromResource(s)
		if err != nil {
			return err
		}
		templates, ok := mappings["dynamic_templates"].([]any)
		if !ok {
			return fmt.Errorf("state mappings dynamic_templates is not an array")
		}
		if len(templates) < minCount {
			return fmt.Errorf("state mappings dynamic_templates length %d, want at least %d", len(templates), minCount)
		}
		return nil
	}
}

func importStateIDForIndexName(indexName string) resource.ImportStateIdFunc {
	return func(_ *terraform.State) (string, error) {
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			return "", err
		}

		id, diags := client.ID(context.Background(), indexName)
		if diags.HasError() {
			return "", fmt.Errorf("failed to build import ID for index %q: %v", indexName, diags)
		}
		return id.String(), nil
	}
}

func deleteIndexOutOfBand(t *testing.T, indexName string) {
	t.Helper()

	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		t.Fatalf("failed to create Elasticsearch client: %s", err)
	}

	diags := esclient.DeleteIndex(context.Background(), client, indexName)
	if diags.HasError() {
		t.Fatalf("failed to delete index %q out of band: %v", indexName, diags)
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

	indexNames := make(map[string]struct{})
	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "elasticstack_elasticsearch_index":
			compID, compDiags := clients.CompositeIDFromStr(rs.Primary.ID)
			if compDiags.HasError() {
				return fmt.Errorf("failed to parse composite ID %q: %v", rs.Primary.ID, compDiags)
			}
			indexNames[compID.ResourceID] = struct{}{}
		case "elasticstack_elasticsearch_index_mappings":
			compID, compDiags := clients.CompositeIDFromStr(rs.Primary.ID)
			if compDiags.HasError() {
				return fmt.Errorf("failed to parse composite ID %q: %v", rs.Primary.ID, compDiags)
			}
			indexNames[compID.ResourceID] = struct{}{}
		}
	}

	for indexName := range indexNames {
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
