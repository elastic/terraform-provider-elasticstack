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
	"net/http"
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

// The Cases field library API first shipped in 9.6.0 (supported matches the
// production minimum version, unsupported is its complement).
var (
	supported   = version.MustConstraints(version.NewConstraint(">= 9.6.0-SNAPSHOT"))
	unsupported = version.MustConstraints(version.NewConstraint("< 9.6.0"))
)

// preCheck skips the test unless the stack version satisfies constraints, then
// runs the standard acceptance pre-check.
func preCheck(t *testing.T, constraints version.Constraints) {
	versionutils.SkipIfUnsupportedConstraints(t, constraints, versionutils.FlavorAny)
	acctest.PreCheck(t)
}

func uniqueName() string {
	return "tf_acc_" + uuid.New().String()[:8]
}

func nameVars(name string) config.Variables {
	return config.Variables{"name": config.StringVariable(name)}
}

// step builds an apply step for the named test-case directory; checks are
// aggregated.
func step(dir string, vars config.Variables, checks ...resource.TestCheckFunc) resource.TestStep {
	s := resource.TestStep{
		ProtoV6ProviderFactories: acctest.Providers,
		ConfigDirectory:          acctest.NamedTestCaseDirectory(dir),
		ConfigVariables:          vars,
	}
	if len(checks) > 0 {
		s.Check = resource.ComposeAggregateTestCheckFunc(checks...)
	}
	return s
}

// errStep builds a step that is expected to fail with an error matching pattern.
func errStep(dir string, vars config.Variables, pattern string) resource.TestStep {
	s := step(dir, vars)
	s.ExpectError = regexp.MustCompile(pattern)
	return s
}

// planErrStep builds a plan-only step that fails validation before any API call.
func planErrStep(dir, pattern string) resource.TestStep {
	s := errStep(dir, nil, pattern)
	s.PlanOnly = true
	return s
}

// importStep builds an import step for the named test-case directory.
func importStep(dir string, vars config.Variables) resource.TestStep {
	s := step(dir, vars)
	s.ResourceName = resourceName
	s.ImportState = true
	return s
}

func attr(key, value string) resource.TestCheckFunc {
	return resource.TestCheckResourceAttr(resourceName, key, value)
}

func attrOf(s *terraform.State, key string) (string, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return "", fmt.Errorf("resource %s not found in state", resourceName)
	}
	if key == "id" {
		return rs.Primary.ID, nil
	}
	return rs.Primary.Attributes[key], nil
}

func idFunc(s *terraform.State) (string, error) {
	return attrOf(s, "id")
}

// captureID records the current field_definition_id; idUnchanged / idChanged
// later assert an in-place update or a replacement respectively.
func captureID(dst *string) resource.TestCheckFunc {
	return func(s *terraform.State) (err error) {
		*dst, err = attrOf(s, "field_definition_id")
		return err
	}
}

func idUnchanged(prev *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		got, err := attrOf(s, "field_definition_id")
		if err == nil && got != *prev {
			err = fmt.Errorf("expected field_definition_id %q to be unchanged, got %q", *prev, got)
		}
		return err
	}
}

func idChanged(prev *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		got, err := attrOf(s, "field_definition_id")
		if err == nil && got == *prev {
			err = fmt.Errorf("expected field_definition_id to change from %q", *prev)
		}
		return err
	}
}

// allFieldsChecks are the assertions for the all_fields config (top-level name,
// description, is_global), shared by the update and create-with-all-fields tests.
func allFieldsChecks(name string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		attr("owner", "cases"),
		attr("name", name),
		attr("description", "initial description"),
		attr("is_global", "true"),
		resource.TestCheckResourceAttrSet(resourceName, "field_definition_id"),
		resource.TestCheckResourceAttrSet(resourceName, "display_order"),
	}
}

func TestAccResourceKibanaCaseFieldDefinition(t *testing.T) {
	name := uniqueName()
	vars := nameVars(name)
	var firstID string

	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t, supported) },
		Steps: []resource.TestStep{
			// Minimal create: only owner and definition (YAML name present).
			// Also serves as the probe for omitted description/is_global:
			// the read-after-write refresh must leave a clean plan and
			// nil response pointers must not panic.
			step("minimal", vars,
				attr("owner", "cases"),
				attr("name", name),
				attr("space_id", "default"),
				resource.TestCheckResourceAttrSet(resourceName, "field_definition_id"),
				// Omitted optional fields are absent from the API response and stay null.
				resource.TestCheckNoResourceAttr(resourceName, "description"),
				resource.TestCheckNoResourceAttr(resourceName, "is_global"),
				resource.TestCheckNoResourceAttr(resourceName, "display_order"),
				resource.TestMatchResourceAttr(resourceName, "id", regexp.MustCompile(`^default/.+`)),
				captureID(&firstID),
			),
			// Adopt top-level name, description, is_global.
			step("all_fields", vars, append(allFieldsChecks(name), idUnchanged(&firstID))...),
			step("update_fields", vars,
				attr("description", "updated description"),
				attr("is_global", "false"),
				resource.TestCheckResourceAttrSet(resourceName, "display_order"),
				idUnchanged(&firstID),
			),
			// Omitting description and is_global from config does not clear them:
			// the API preserves stored values on omission (verified live on 9.6),
			// and the plan keeps the prior state values.
			step("remove_description", vars,
				attr("description", "updated description"),
				attr("is_global", "false"),
				idUnchanged(&firstID),
			),
			// An explicit empty string clears the description.
			step("clear_description", vars, attr("description", ""), idUnchanged(&firstID)),
			// Non-identity YAML change updates in place.
			step("update_yaml", vars,
				resource.TestMatchResourceAttr(resourceName, "definition", regexp.MustCompile(`Renamed Label`)),
				idUnchanged(&firstID),
			),
			// Reformat-only YAML (key order, blank line, quoting). Verified live: the
			// plan still shows an in-place update (the framework does not suppress the
			// config-vs-state diff via StringSemanticEquals), which is the documented
			// v1 limitation. The apply must succeed without replacement or an
			// inconsistent-result error, and a follow-up plan must be empty.
			step("reformat_yaml", vars, attr("name", name), idUnchanged(&firstID)),
			withImportID(importStep("update_yaml", vars), idFunc),
			withImportError(importStep("update_yaml", vars), "not-a-composite-id", `(?i)Wrong resource ID|Invalid import ID`),
			// Changing top-level name together with the YAML name replaces the resource.
			step("rename", config.Variables{
				"name":     config.StringVariable(name),
				"new_name": config.StringVariable(name + "_renamed"),
			}, attr("name", name+"_renamed"), idChanged(&firstID)),
		},
	})
}

func withImportID(s resource.TestStep, fn resource.ImportStateIdFunc) resource.TestStep {
	s.ImportStateIdFunc = fn
	s.ImportStateVerify = true
	return s
}

func withImportError(s resource.TestStep, id, pattern string) resource.TestStep {
	s.ImportStateId = id
	s.ExpectError = regexp.MustCompile(pattern)
	return s
}

// TestAccResourceKibanaCaseFieldDefinitionCreateAllFields applies the
// all_fields config as the first step so name, description and is_global are
// sent on POST.
func TestAccResourceKibanaCaseFieldDefinitionCreateAllFields(t *testing.T) {
	name := uniqueName()
	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t, supported) },
		Steps:    []resource.TestStep{step("all_fields", nameVars(name), allFieldsChecks(name)...)},
	})
}

// TestAccResourceKibanaCaseFieldDefinitionInvalidOwner verifies an owner outside
// the allowed set is rejected at validation time (plan only, no API call).
func TestAccResourceKibanaCaseFieldDefinitionInvalidOwner(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t, supported) },
		Steps:    []resource.TestStep{planErrStep("invalid", `(?is)Invalid Attribute Value Match.*owner|value must be one of`)},
	})
}

func TestAccResourceKibanaCaseFieldDefinitionNameMismatch(t *testing.T) {
	name := uniqueName()
	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t, supported) },
		Steps: []resource.TestStep{
			step("create", nameVars(name)),
			// Only the YAML-internal name changes: the API rejects the mismatch with a 400.
			errStep("mismatch", config.Variables{
				"name":       config.StringVariable(name),
				"other_name": config.StringVariable(name + "_other"),
			}, `(?is)must match the name in the YAML`),
		},
	})
}

func TestAccResourceKibanaCaseFieldDefinitionTypeChange(t *testing.T) {
	name := uniqueName()
	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t, supported) },
		Steps: []resource.TestStep{
			step("create", nameVars(name)),
			errStep("change_type", nameVars(name), `field_identity_immutable`),
		},
	})
}

func TestAccResourceKibanaCaseFieldDefinitionSpace(t *testing.T) {
	spaceID := "tf-acc-" + uuid.New().String()[:8]
	vars := config.Variables{
		"name":     config.StringVariable(uniqueName()),
		"space_id": config.StringVariable(spaceID),
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t, supported) },
		Steps: []resource.TestStep{
			step("create", vars,
				attr("space_id", spaceID),
				resource.TestMatchResourceAttr(resourceName, "id", regexp.MustCompile("^"+spaceID+"/.+")),
			),
			withImportID(importStep("create", vars), idFunc),
		},
	})
}

func TestAccResourceKibanaCaseFieldDefinitionOwnerChange(t *testing.T) {
	name := uniqueName()
	var firstID string
	vars := func(owner string) config.Variables {
		return config.Variables{"name": config.StringVariable(name), "owner": config.StringVariable(owner)}
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t, supported) },
		Steps: []resource.TestStep{
			step("create", vars("cases"), attr("owner", "cases"), captureID(&firstID)),
			// owner is immutable server-side; changing it must replace the resource.
			step("change", vars("observability"), attr("owner", "observability"), idChanged(&firstID)),
			// the third allowed owner value must also round-trip.
			step("change", vars("securitySolution"), attr("owner", "securitySolution"), idChanged(&firstID)),
		},
	})
}

func TestAccResourceKibanaCaseFieldDefinitionExternalDelete(t *testing.T) {
	name := uniqueName()
	var firstID string

	// Delete outside Terraform; the refresh must drop the resource from
	// state and the apply must recreate it.
	recreate := step("create", nameVars(name), idChanged(&firstID))
	recreate.PreConfig = func() { deleteFieldDefinitionByID(t, firstID) }

	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t, supported) },
		Steps: []resource.TestStep{
			step("create", nameVars(name), captureID(&firstID)),
			recreate,
		},
	})
}

// TestAccResourceKibanaCaseFieldDefinitionDestroyBlocked links the field into a
// case template through the Kibana API, then expects destroy to fail with a
// 409 ("referenced by N active template(s)"). The template shape (a `$ref`
// field entry, response key `templateId`) was verified against a live 9.6 stack.
func TestAccResourceKibanaCaseFieldDefinitionDestroyBlocked(t *testing.T) {
	name := uniqueName()
	var templateID string

	blocked := errStep("create", nameVars(name), `(?i)referenced by .* active template`)
	blocked.Destroy = true
	blocked.PreConfig = func() {
		templateID = createTemplateReferencing(t, name)
		t.Cleanup(func() { deleteTemplate(t, templateID) })
	}

	// Remove the reference so the harness's final destroy can succeed.
	unblocked := step("create", nameVars(name))
	unblocked.PreConfig = func() { deleteTemplate(t, templateID) }

	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t, supported) },
		Steps:    []resource.TestStep{step("create", nameVars(name)), blocked, unblocked},
	})
}

// TestAccResourceKibanaCaseFieldDefinitionVersionGate runs only against stacks
// older than 9.6.0 and verifies the resource-level version diagnostic.
func TestAccResourceKibanaCaseFieldDefinitionVersionGate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t, unsupported) },
		Steps: []resource.TestStep{
			errStep("create", nameVars(uniqueName()), `(?is)requires\s+Elastic\s+Stack\s+v9\.6\.0\s+or\s+later`),
		},
	})
}

// TestAccResourceKibanaCaseFieldDefinitionInvalidYAML verifies invalid YAML is
// rejected at validation time, before any API call (the provider has no
// reachable-API requirement for validation).
func TestAccResourceKibanaCaseFieldDefinitionInvalidYAML(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t, supported) },
		Steps:    []resource.TestStep{planErrStep("invalid", `Invalid YAML`)},
	})
}

// TestAccResourceKibanaCaseFieldDefinitionDefinitionTooLong verifies a
// `definition` longer than 30000 characters is rejected at validation time,
// before any API call. The YAML itself is valid (the excess length is a
// trailing comment) so only the length validator fires.
func TestAccResourceKibanaCaseFieldDefinitionDefinitionTooLong(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t, supported) },
		Steps:    []resource.TestStep{planErrStep("too_long", `(?is)Invalid Attribute Value Length.*at most 30000`)},
	})
}

// TestAccResourceKibanaCaseFieldDefinitionKibanaConnection verifies a resource
// with kibana_connection uses its scoped client: it succeeds with the real
// endpoint and fails to connect when the block points at an unreachable one,
// which could not happen if the provider-level client were used instead.
func TestAccResourceKibanaCaseFieldDefinitionKibanaConnection(t *testing.T) {
	name := uniqueName()
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheckWithExplicitKibanaEndpoint(t)
			preCheck(t, supported)
		},
		Steps: []resource.TestStep{
			step("create", acctest.KibanaConnectionVariables(nameVars(name)),
				attr("kibana_connection.#", "1"),
				attr("name", name),
			),
			errStep("create", acctest.KibanaConnectionVariables(nameVars(name), config.Variables{
				"kibana_endpoints": config.ListVariable(config.StringVariable("http://127.0.0.1:1")),
			}), `(?is)connection refused|dial|unable|failed`),
		},
	})
}

func kibanaClient(t *testing.T) *kibanaoapi.Client {
	t.Helper()
	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	return client.GetKibanaOapiClient()
}

func deleteFieldDefinitionByID(t *testing.T, id string) {
	t.Helper()
	if diags := kibanaoapi.DeleteCaseFieldDefinition(context.Background(), kibanaClient(t), "default", id); diags.HasError() {
		t.Fatalf("failed to delete field definition %s: %v", id, diags)
	}
}

func createTemplateReferencing(t *testing.T, fieldName string) string {
	t.Helper()
	tplName := fieldName + "_tpl"
	yaml := fmt.Sprintf("name: %s\nfields:\n  - $ref: %s\n", tplName, fieldName)
	resp, err := kibanaClient(t).API.CreateCaseTemplateDefaultSpaceWithResponse(
		context.Background(),
		&kbapi.CreateCaseTemplateDefaultSpaceParams{},
		kbapi.CasesTemplateWriteRequest{Owner: kbapi.CasesOwnerCases, Name: &tplName, Definition: yaml},
		kibanautil.SpaceAwarePathRequestEditor("default"),
	)
	if err != nil {
		t.Fatalf("failed to create template: %v", err)
	}
	if resp.StatusCode() != http.StatusOK {
		t.Fatalf("failed to create template: %d %s", resp.StatusCode(), string(resp.Body))
	}
	var body struct {
		TemplateID string `json:"templateId"`
	}
	if err := json.Unmarshal(resp.Body, &body); err != nil || body.TemplateID == "" {
		t.Fatalf("template response did not include a templateId (err=%v): %s", err, string(resp.Body))
	}
	return body.TemplateID
}

// deleteTemplate is best-effort cleanup (also run via t.Cleanup after an
// explicit delete), so it deliberately ignores the response status.
func deleteTemplate(t *testing.T, id string) {
	t.Helper()
	if _, err := kibanaClient(t).API.DeleteCaseTemplateDefaultSpaceWithResponse(context.Background(), id); err != nil {
		t.Errorf("failed to delete template %s: %v", id, err)
	}
}
