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

var minVersion = version.Must(version.NewVersion("9.6.0-SNAPSHOT"))

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
					// Omitted optional fields are absent from the API response and stay null.
					resource.TestCheckNoResourceAttr(resourceName, "description"),
					resource.TestCheckNoResourceAttr(resourceName, "is_global"),
					resource.TestCheckNoResourceAttr(resourceName, "display_order"),
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
					resource.TestCheckResourceAttrSet(resourceName, "display_order"),
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
					resource.TestCheckResourceAttrSet(resourceName, "display_order"),
					sameID(&firstID, true),
				),
			},
			{
				// Omitting description and is_global from config does not clear them:
				// the API preserves stored values on omission (verified live on 9.6),
				// and the plan keeps the prior state values.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("remove_description"),
				ConfigVariables:          nameVars(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "updated description"),
					resource.TestCheckResourceAttr(resourceName, "is_global", "false"),
					sameID(&firstID, true),
				),
			},
			{
				// An explicit empty string clears the description.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("clear_description"),
				ConfigVariables:          nameVars(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					sameID(&firstID, true),
				),
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
				// Reformat-only YAML (key order, blank line, quoting). Verified live: the
				// plan still shows an in-place update (the framework does not suppress the
				// config-vs-state diff via StringSemanticEquals), which is the documented
				// v1 limitation. The apply must succeed without replacement or an
				// inconsistent-result error, and a follow-up plan must be empty.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("reformat_yaml"),
				ConfigVariables:          nameVars(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					sameID(&firstID, true),
				),
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
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_yaml"),
				ConfigVariables:          nameVars(name),
				ResourceName:             resourceName,
				ImportState:              true,
				ImportStateId:            "not-a-composite-id",
				ExpectError:              regexp.MustCompile(`(?i)Wrong resource ID|Invalid import ID`),
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

// TestAccResourceKibanaCaseFieldDefinitionCreateAllFields applies the
// all_fields config as the first step so name, description and is_global are
// sent on POST.
func TestAccResourceKibanaCaseFieldDefinitionCreateAllFields(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersion, versionutils.FlavorAny)

	name := uniqueName()
	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_fields"),
				ConfigVariables:          nameVars(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "owner", "cases"),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "description", "initial description"),
					resource.TestCheckResourceAttr(resourceName, "is_global", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "field_definition_id"),
					resource.TestCheckResourceAttrSet(resourceName, "display_order"),
				),
			},
		},
	})
}

// TestAccResourceKibanaCaseFieldDefinitionInvalidOwner verifies an owner outside
// the allowed set is rejected at validation time (plan only, no API call).
func TestAccResourceKibanaCaseFieldDefinitionInvalidOwner(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersion, versionutils.FlavorAny)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("invalid"),
				PlanOnly:                 true,
				ExpectError:              regexp.MustCompile(`(?is)Invalid Attribute Value Match.*owner|value must be one of`),
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
				ExpectError: regexp.MustCompile(`(?is)must match the name in the YAML`),
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

func TestAccResourceKibanaCaseFieldDefinitionOwnerChange(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersion, versionutils.FlavorAny)

	name := uniqueName()
	var firstID string
	vars := func(owner string) config.Variables {
		return config.Variables{"name": config.StringVariable(name), "owner": config.StringVariable(owner)}
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars("cases"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "owner", "cases"),
					captureID(&firstID),
				),
			},
			{
				// owner is immutable server-side; changing it must replace the resource.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("change"),
				ConfigVariables:          vars("observability"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "owner", "observability"),
					sameID(&firstID, false),
				),
			},
		},
	})
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
// 409 ("referenced by N active template(s)"). The template shape (a `$ref`
// field entry, response key `templateId`) was verified against a live 9.6 stack.
func TestAccResourceKibanaCaseFieldDefinitionDestroyBlocked(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersion, versionutils.FlavorAny)

	name := uniqueName()
	var templateID string

	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          nameVars(name),
			},
			{
				PreConfig: func() {
					templateID = createTemplateReferencing(t, name)
					t.Cleanup(func() { deleteTemplate(t, templateID) })
				},
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          nameVars(name),
				Destroy:                  true,
				ExpectError:              regexp.MustCompile(`(?i)referenced by .* active template`),
			},
			{
				// Remove the reference so the harness's final destroy can succeed.
				PreConfig:                func() { deleteTemplate(t, templateID) },
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          nameVars(name),
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
				ExpectError:              regexp.MustCompile(`(?is)requires\s+Elastic\s+Stack\s+v9\.6\.0\s+or\s+later`),
			},
		},
	})
}

// TestAccResourceKibanaCaseFieldDefinitionInvalidYAML verifies invalid YAML is
// rejected at validation time, before any API call (the provider has no
// reachable-API requirement for validation).
func TestAccResourceKibanaCaseFieldDefinitionInvalidYAML(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersion, versionutils.FlavorAny)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("invalid"),
				PlanOnly:                 true,
				ExpectError:              regexp.MustCompile(`Invalid YAML`),
			},
		},
	})
}

// TestAccResourceKibanaCaseFieldDefinitionKibanaConnection verifies a resource
// with kibana_connection uses its scoped client: it succeeds with the real
// endpoint and fails to connect when the block points at an unreachable one,
// which could not happen if the provider-level client were used instead.
func TestAccResourceKibanaCaseFieldDefinitionKibanaConnection(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersion, versionutils.FlavorAny)

	name := uniqueName()
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheckWithExplicitKibanaEndpoint(t)
			preCheck(t)
		},
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          acctest.KibanaConnectionVariables(nameVars(name)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "kibana_connection.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", name),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: acctest.KibanaConnectionVariables(nameVars(name), config.Variables{
					"kibana_endpoints": config.ListVariable(config.StringVariable("http://127.0.0.1:1")),
				}),
				ExpectError: regexp.MustCompile(`(?is)connection refused|dial|unable|failed`),
			},
		},
	})
}
