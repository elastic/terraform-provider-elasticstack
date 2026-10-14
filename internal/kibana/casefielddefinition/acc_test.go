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

package casefielddefinition_test

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/google/uuid"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const resourceName = "elasticstack_kibana_case_field_definition.test"

var minVersion = version.Must(version.NewVersion("9.6.0"))

func uniqueName() string {
	return "tf_acc_" + uuid.New().String()[:8]
}

func nameVars(name string) config.Variables {
	return config.Variables{"name": config.StringVariable(name)}
}

func idFunc(s *terraform.State) (string, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return "", fmt.Errorf("resource %s not found in state", resourceName)
	}
	return rs.Primary.ID, nil
}

// captureID records the current field_definition_id; sameID later asserts it is
// unchanged (in-place update) or different (replacement).
func captureID(dst *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		*dst = rs.Primary.Attributes["field_definition_id"]
		return nil
	}
}

func sameID(prev *string, same bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		got := rs.Primary.Attributes["field_definition_id"]
		if same && got != *prev {
			return fmt.Errorf("expected field_definition_id %q to be unchanged, got %q", *prev, got)
		}
		if !same && got == *prev {
			return fmt.Errorf("expected field_definition_id to change from %q", *prev)
		}
		return nil
	}
}

func preCheck(t *testing.T) {
	acctest.PreCheck(t)
}

func TestAccResourceKibanaCaseFieldDefinition(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersion, versionutils.FlavorAny)

	name := uniqueName()
	newName := name + "_renamed"
	var firstID string

	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t) },
		Steps: []resource.TestStep{
			{
				// Minimal create: only owner and definition (YAML name present).
				// Also serves as the probe for omitted description/is_global:
				// the read-after-write refresh must leave a clean plan and
				// nil response pointers must not panic.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("minimal"),
				ConfigVariables:          nameVars(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "owner", "cases"),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "space_id", "default"),
					resource.TestCheckResourceAttrSet(resourceName, "field_definition_id"),
					resource.TestCheckResourceAttrSet(resourceName, "display_order"),
					resource.TestMatchResourceAttr(resourceName, "id", regexp.MustCompile(`^default/.+`)),
					captureID(&firstID),
				),
			},
			{
				// Create-with-all-fields equivalent: adopt top-level name, description, is_global.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_fields"),
				ConfigVariables:          nameVars(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "description", "initial description"),
					resource.TestCheckResourceAttr(resourceName, "is_global", "true"),
					sameID(&firstID, true),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_fields"),
				ConfigVariables:          nameVars(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "updated description"),
					resource.TestCheckResourceAttr(resourceName, "is_global", "false"),
					sameID(&firstID, true),
				),
			},
			{
				// Remove description from config. The assertion is limited to a
				// clean follow-up plan; whether the server clears or preserves
				// the stored value is the open question recorded in design.md.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("remove_description"),
				ConfigVariables:          nameVars(name),
				Check:                    sameID(&firstID, true),
			},
			{
				// Non-identity YAML change updates in place.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_yaml"),
				ConfigVariables:          nameVars(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "definition", regexp.MustCompile(`Renamed Label`)),
					sameID(&firstID, true),
				),
			},
			{
				// Semantically identical YAML (key order, blank line, quoting) must plan clean.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("reformat_yaml"),
				ConfigVariables:          nameVars(name),
				PlanOnly:                 true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_yaml"),
				ConfigVariables:          nameVars(name),
				ResourceName:             resourceName,
				ImportState:              true,
				ImportStateIdFunc:        idFunc,
				ImportStateVerify:        true,
			},
			{
				// Changing top-level name together with the YAML name replaces the resource.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("rename"),
				ConfigVariables: config.Variables{
					"name":     config.StringVariable(name),
					"new_name": config.StringVariable(newName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", newName),
					sameID(&firstID, false),
				),
			},
		},
	})
}

func TestAccResourceKibanaCaseFieldDefinitionNameMismatch(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersion, versionutils.FlavorAny)

	name := uniqueName()
	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          nameVars(name),
			},
			{
				// Only the YAML-internal name changes: the API rejects the mismatch with a 400.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("mismatch"),
				ConfigVariables: config.Variables{
					"name":       config.StringVariable(name),
					"other_name": config.StringVariable(name + "_other"),
				},
				ExpectError: regexp.MustCompile(`(?is)name`),
			},
		},
	})
}

func TestAccResourceKibanaCaseFieldDefinitionTypeChange(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersion, versionutils.FlavorAny)

	name := uniqueName()
	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          nameVars(name),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("change_type"),
				ConfigVariables:          nameVars(name),
				ExpectError:              regexp.MustCompile(`field_identity_immutable`),
			},
		},
	})
}

func TestAccResourceKibanaCaseFieldDefinitionSpace(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersion, versionutils.FlavorAny)

	name := uniqueName()
	spaceID := "tf-acc-" + uuid.New().String()[:8]
	vars := config.Variables{
		"name":     config.StringVariable(name),
		"space_id": config.StringVariable(spaceID),
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "space_id", spaceID),
					resource.TestMatchResourceAttr(resourceName, "id", regexp.MustCompile("^"+spaceID+"/.+")),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				ResourceName:             resourceName,
				ImportState:              true,
				ImportStateIdFunc:        idFunc,
				ImportStateVerify:        true,
			},
		},
	})
}

func deleteFieldDefinitionByID(t *testing.T, id string) {
	t.Helper()
	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	if diags := kibanaoapi.DeleteCaseFieldDefinition(context.Background(), client.GetKibanaOapiClient(), "default", id); diags.HasError() {
		t.Fatalf("failed to delete field definition %s: %v", id, diags)
	}
}

func TestAccResourceKibanaCaseFieldDefinitionExternalDelete(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersion, versionutils.FlavorAny)

	name := uniqueName()
	var firstID string

	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          nameVars(name),
				Check:                    captureID(&firstID),
			},
			{
				// Delete outside Terraform; the refresh must drop the resource from
				// state and the apply must recreate it.
				PreConfig:                func() { deleteFieldDefinitionByID(t, firstID) },
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          nameVars(name),
				Check:                    sameID(&firstID, false),
			},
		},
	})
}

// TestAccResourceKibanaCaseFieldDefinitionDestroyBlocked links the field into a
// case template through the Kibana API, then expects destroy to fail with a
// 409. NOTE: the template YAML (`$ref` entry) and response shape follow the
// OpenAPI description and have not been verified against a live 9.6+ stack.
func TestAccResourceKibanaCaseFieldDefinitionDestroyBlocked(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersion, versionutils.FlavorAny)

	name := uniqueName()
	var fieldID, templateID string

	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          nameVars(name),
				Check:                    captureID(&fieldID),
			},
			{
				PreConfig: func() {
					templateID = createTemplateReferencing(t, name)
					t.Cleanup(func() {
						deleteTemplate(t, templateID)
						deleteFieldDefinitionByID(t, fieldID)
					})
				},
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          nameVars(name),
				Destroy:                  true,
				ExpectError:              regexp.MustCompile(`(?i)409|conflict|referenc|linked|in use`),
			},
		},
	})
}

func createTemplateReferencing(t *testing.T, fieldName string) string {
	t.Helper()
	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	tplName := fieldName + "_tpl"
	yaml := fmt.Sprintf("name: %s\nfields:\n  - $ref: %s\n", tplName, fieldName)
	resp, err := client.GetKibanaOapiClient().API.CreateCaseTemplateDefaultSpaceWithResponse(
		context.Background(),
		&kbapi.CreateCaseTemplateDefaultSpaceParams{},
		kbapi.CasesTemplateWriteRequest{Owner: kbapi.CasesOwnerCases, Name: &tplName, Definition: yaml},
		kibanautil.SpaceAwarePathRequestEditor("default"),
	)
	if err != nil {
		t.Fatalf("failed to create template: %v", err)
	}
	if resp.StatusCode() != 200 {
		t.Fatalf("failed to create template: %d %s", resp.StatusCode(), string(resp.Body))
	}
	var body map[string]any
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		t.Fatalf("failed to decode template response: %v", err)
	}
	for _, key := range []string{"templateId", "id"} {
		if v, ok := body[key].(string); ok && v != "" {
			return v
		}
	}
	t.Fatalf("template response did not include an id: %s", string(resp.Body))
	return ""
}

func deleteTemplate(t *testing.T, id string) {
	t.Helper()
	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		t.Errorf("failed to create client: %v", err)
		return
	}
	if _, err := client.GetKibanaOapiClient().API.DeleteCaseTemplateDefaultSpaceWithResponse(context.Background(), id); err != nil {
		t.Errorf("failed to delete template %s: %v", id, err)
	}
}

// TestAccResourceKibanaCaseFieldDefinitionVersionGate runs only against stacks
// older than 9.6.0 and verifies the resource-level version diagnostic.
func TestAccResourceKibanaCaseFieldDefinitionVersionGate(t *testing.T) {
	versionutils.SkipIfUnsupportedConstraints(t, version.MustConstraints(version.NewConstraint("< 9.6.0")), versionutils.FlavorAny)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          nameVars(uniqueName()),
				ExpectError:              regexp.MustCompile(`(?i)requires Elastic Stack v9\.5\.0 or later`),
			},
		},
	})
}
