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

package casetemplate_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/google/uuid"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const resourceName = "elasticstack_kibana_case_template.test"

// The standalone Cases templates API first shipped in 9.5.2 (supported matches
// the production minimum version, unsupported is its complement).
var (
	supported   = version.MustConstraints(version.NewConstraint(">= 9.5.2"))
	unsupported = version.MustConstraints(version.NewConstraint("< 9.5.2"))
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

// captureID records the current template_id; idUnchanged / idChanged later
// assert an in-place update or a replacement respectively.
func captureID(dst *string) resource.TestCheckFunc {
	return func(s *terraform.State) (err error) {
		*dst, err = attrOf(s, "template_id")
		return err
	}
}

func idUnchanged(prev *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		got, err := attrOf(s, "template_id")
		if err == nil && got != *prev {
			err = fmt.Errorf("expected template_id %q to be unchanged, got %q", *prev, got)
		}
		return err
	}
}

// compositeIDMatches asserts id is "<space_id>/<template_id>".
func compositeIDMatches(spaceID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		id, err := attrOf(s, "id")
		if err != nil {
			return err
		}
		templateID, err := attrOf(s, "template_id")
		if err != nil {
			return err
		}
		if want := spaceID + "/" + templateID; id != want {
			return fmt.Errorf("expected id %q, got %q", want, id)
		}
		return nil
	}
}

func kibanaClient(t *testing.T) *kibanaoapi.Client {
	t.Helper()
	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	return client.GetKibanaOapiClient()
}

func TestAccResourceKibanaCaseTemplate(t *testing.T) {
	name := uniqueName()
	vars := nameVars(name)
	var firstID string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { preCheck(t, supported) },
		CheckDestroy: checkTemplatesDestroyed(t),
		Steps: []resource.TestStep{
			// Minimal create: only owner, name and definition_yaml. The
			// read-after-write refresh supplies the final state.
			step("minimal", vars,
				attr("owner", "cases"),
				attr("name", name),
				attr("space_id", "default"),
				attr("template_version", "1"),
				attr("is_enabled", "true"),
				resource.TestCheckResourceAttrSet(resourceName, "template_id"),
				resource.TestMatchResourceAttr(resourceName, "id", regexp.MustCompile(`^default/.+`)),
				compositeIDMatches("default"),
				resource.TestCheckNoResourceAttr(resourceName, "description"),
				resource.TestCheckNoResourceAttr(resourceName, "tags.#"),
				attr("field_count", "1"),
				resource.TestCheckNoResourceAttr(resourceName, "usage_count"),
				resource.TestCheckNoResourceAttr(resourceName, "is_default"),
				captureID(&firstID),
			),
			// Reformat-only edit (key order, comments, indentation): the plan
			// must be empty and no PUT may be sent, so the version is unchanged.
			withPlanCheck(step("reformat", vars, attr("template_version", "1"), attr("field_count", "1"), idUnchanged(&firstID)), plancheck.ExpectEmptyPlan()),
			// Structural edit: a real in-place update that creates version 2.
			withPlanCheck(step("structural_edit", vars,
				resource.TestMatchResourceAttr(resourceName, "definition_yaml", regexp.MustCompile(`severity: high`)),
				attr("template_version", "2"),
				idUnchanged(&firstID),
			), plancheck.ExpectNonEmptyPlan()),
			step("add_description_tags", vars,
				attr("description", "Initial description"),
				attr("tags.#", "2"),
				attr("tags.0", "z"),
				attr("tags.1", "a"),
				attr("template_version", "3"),
				idUnchanged(&firstID),
			),
			// Import after a populated step (description and tags set).
			withImportID(importStep("add_description_tags", vars), idFunc),
			// Duplicate entries and a specific order are preserved by the API.
			step("tags_duplicates", vars,
				attr("tags.#", "3"),
				attr("tags.0", "z"),
				attr("tags.1", "a"),
				attr("tags.2", "z"),
				attr("template_version", "4"),
			),
			// In-place rename.
			step("rename", vars,
				attr("name", name+"_renamed"),
				attr("template_version", "5"),
				idUnchanged(&firstID),
			),
			// is_enabled toggle.
			step("disable", vars,
				attr("is_enabled", "false"),
				attr("template_version", "6"),
				idUnchanged(&firstID),
			),
			// Omitting a previously set description/tags clears them.
			step("remove_description_tags", vars,
				resource.TestCheckNoResourceAttr(resourceName, "description"),
				resource.TestCheckNoResourceAttr(resourceName, "tags.#"),
				attr("template_version", "7"),
			),
			// Explicit empty values stay distinct from null.
			step("empty_description_tags", vars,
				attr("description", ""),
				attr("tags.#", "0"),
				attr("template_version", "8"),
			),
			// Omitting is_enabled on update keeps the current value (false); the
			// server default of true only applies on create. No update is sent.
			withPlanCheck(step("omit_is_enabled", vars, attr("is_enabled", "false"), attr("template_version", "8")), plancheck.ExpectEmptyPlan()),
			// Re-enabling is a real update.
			step("reenable", vars, attr("is_enabled", "true"), attr("template_version", "9")),
			// field_count follows structural field additions and removals.
			step("add_field", vars, attr("field_count", "2"), attr("template_version", "10")),
			step("remove_field", vars, attr("field_count", "1"), attr("template_version", "11")),
			withImportID(importStep("remove_field", vars), idFunc),
		},
	})
}

// importStep builds an import step for the named test-case directory.
func importStep(dir string, vars config.Variables) resource.TestStep {
	s := step(dir, vars)
	s.ResourceName = resourceName
	s.ImportState = true
	return s
}

func withImportID(s resource.TestStep, fn resource.ImportStateIdFunc) resource.TestStep {
	s.ImportStateIdFunc = fn
	s.ImportStateVerify = true
	return s
}

func idFunc(s *terraform.State) (string, error) {
	return attrOf(s, "id")
}

// checkTemplatesDestroyed asserts every template in state is gone after
// destroy: the API answers 404 on read, which the wrapper reports as nil.
func checkTemplatesDestroyed(t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "elasticstack_kibana_case_template" {
				continue
			}
			got, diags := kibanaoapi.GetCaseTemplate(context.Background(), kibanaClient(t), rs.Primary.Attributes["space_id"], rs.Primary.Attributes["template_id"])
			if diags.HasError() {
				return fmt.Errorf("failed to read template %s: %v", rs.Primary.ID, diags)
			}
			if got != nil {
				return fmt.Errorf("template %s still exists after destroy", rs.Primary.ID)
			}
		}
		return nil
	}
}

func withPlanCheck(s resource.TestStep, check plancheck.PlanCheck) resource.TestStep {
	s.ConfigPlanChecks = resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{check}}
	return s
}

// TestAccResourceKibanaCaseTemplateCreateAllFields creates a template with
// every optional attribute set.
func TestAccResourceKibanaCaseTemplateCreateAllFields(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t, supported) },
		Steps: []resource.TestStep{
			step("all_fields", nameVars(uniqueName()),
				attr("description", "All fields"),
				attr("tags.#", "3"),
				attr("tags.0", "z"),
				attr("tags.1", "a"),
				attr("tags.2", "z"),
				attr("is_enabled", "false"),
				attr("template_version", "1"),
				resource.TestCheckResourceAttrSet(resourceName, "template_id"),
			),
		},
	})
}

// TestAccResourceKibanaCaseTemplateInvalidYAML verifies invalid YAML is
// rejected at validation time, before any API call.
func TestAccResourceKibanaCaseTemplateInvalidYAML(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t, supported) },
		Steps:    []resource.TestStep{planErrStep("invalid", `Invalid YAML`)},
	})
}

// TestAccResourceKibanaCaseTemplateInvalidOwner verifies an owner outside the
// allowed set is rejected at validation time (plan only, no API call).
func TestAccResourceKibanaCaseTemplateInvalidOwner(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t, supported) },
		Steps:    []resource.TestStep{planErrStep("invalid", `(?is)Invalid Attribute Value Match.*owner|value must be one of`)},
	})
}

// TestAccResourceKibanaCaseTemplateVersionGate runs only against stacks older
// than 9.5.2 and verifies the resource-level version diagnostic.
func TestAccResourceKibanaCaseTemplateVersionGate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t, unsupported) },
		Steps: []resource.TestStep{
			errStep("create", nameVars(uniqueName()), `(?is)requires\s+Elastic\s+Stack\s+v9\.5\.2\s+or\s+later`),
		},
	})
}

// TestAccResourceKibanaCaseTemplateKibanaConnection verifies a resource with
// kibana_connection uses its scoped client: it succeeds with the real endpoint
// and fails to connect when the block points at an unreachable one, which could
// not happen if the provider-level client were used instead.
func TestAccResourceKibanaCaseTemplateKibanaConnection(t *testing.T) {
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

// TestAccResourceKibanaCaseTemplateSpace creates a template in a named space
// and imports it by its composite id.
func TestAccResourceKibanaCaseTemplateSpace(t *testing.T) {
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
				compositeIDMatches(spaceID),
			),
			withImportID(importStep("create", vars), idFunc),
			withImportError(importStep("create", vars), "not-a-composite-id", `(?i)Wrong resource ID|Invalid import ID`),
		},
	})
}

func withImportError(s resource.TestStep, id, pattern string) resource.TestStep {
	s.ImportStateId = id
	s.ExpectError = regexp.MustCompile(pattern)
	return s
}

// TestAccResourceKibanaCaseTemplateExternalDelete deletes the template outside
// Terraform; refresh must drop it from state and apply must recreate it.
func TestAccResourceKibanaCaseTemplateExternalDelete(t *testing.T) {
	name := uniqueName()
	var firstID string

	recreate := step("create", nameVars(name), idChanged(&firstID))
	recreate.PreConfig = func() { deleteTemplateByID(t, firstID) }

	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t, supported) },
		Steps: []resource.TestStep{
			step("create", nameVars(name), captureID(&firstID)),
			recreate,
		},
	})
}

// TestAccResourceKibanaCaseTemplateDuplicateName verifies a second template
// with the same owner and name (case-insensitive) surfaces the API's 409.
func TestAccResourceKibanaCaseTemplateDuplicateName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t, supported) },
		Steps: []resource.TestStep{
			errStep("duplicate", nameVars(uniqueName()), `(?is)already exists`),
		},
	})
}

// TestAccResourceKibanaCaseTemplateMalformedSchema verifies syntactically valid
// YAML with an invalid field control surfaces the API's 400.
func TestAccResourceKibanaCaseTemplateMalformedSchema(t *testing.T) {
	name := uniqueName()
	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t, supported) },
		Steps: []resource.TestStep{
			errStep("malformed", nameVars(name), `(?is)got HTTP 400`),
		},
	})
}

// TestAccResourceKibanaCaseTemplateMalformedSchemaUpdate verifies the 400 is
// also surfaced when an existing template is updated.
func TestAccResourceKibanaCaseTemplateMalformedSchemaUpdate(t *testing.T) {
	name := uniqueName()
	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t, supported) },
		Steps: []resource.TestStep{
			step("create", nameVars(name)),
			errStep("malformed", nameVars(name), `(?is)got HTTP 400`),
		},
	})
}

// TestAccResourceKibanaCaseTemplateFieldRef creates a template whose
// definition_yaml references a field-library field by $ref. The field library
// API first shipped in 9.6.0.
func TestAccResourceKibanaCaseTemplateFieldRef(t *testing.T) {
	name := uniqueName()
	fieldName := uniqueName() + "_field"
	vars := config.Variables{
		"name":       config.StringVariable(name),
		"field_name": config.StringVariable(fieldName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			preCheck(t, version.MustConstraints(version.NewConstraint(">= 9.6.0-SNAPSHOT")))
		},
		Steps: []resource.TestStep{
			step("create", vars,
				attr("name", name),
				attr("template_version", "1"),
				attr("field_count", "1"),
				resource.TestMatchResourceAttr(resourceName, "definition_yaml", regexp.MustCompile(`\$ref: `+fieldName)),
			),
		},
	})
}

func idChanged(prev *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		got, err := attrOf(s, "template_id")
		if err == nil && got == *prev {
			err = fmt.Errorf("expected template_id to change from %q", *prev)
		}
		return err
	}
}

func deleteTemplateByID(t *testing.T, id string) {
	t.Helper()
	if diags := kibanaoapi.DeleteCaseTemplate(context.Background(), kibanaClient(t), "default", id); diags.HasError() {
		t.Fatalf("failed to delete template %s: %v", id, diags)
	}
}

// TestAccResourceKibanaCaseTemplateOwnerChange verifies changing owner forces
// replacement (destroy before create) and that every owner value round-trips.
func TestAccResourceKibanaCaseTemplateOwnerChange(t *testing.T) {
	name := uniqueName()
	var firstID string
	vars := func(owner string) config.Variables {
		return config.Variables{"name": config.StringVariable(name), "owner": config.StringVariable(owner)}
	}
	replaced := func(owner string) resource.TestStep {
		s := withPlanCheck(step("change", vars(owner), attr("owner", owner), idChanged(&firstID)),
			plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate))
		return s
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t, supported) },
		Steps: []resource.TestStep{
			step("change", vars("cases"), attr("owner", "cases"), captureID(&firstID)),
			replaced("securitySolution"),
			replaced("observability"),
		},
	})
}

// TestAccResourceKibanaCaseTemplateReformatWithChange verifies a reformat-only
// definition_yaml edit combined with a real attribute change is not suppressed.
func TestAccResourceKibanaCaseTemplateReformatWithChange(t *testing.T) {
	name := uniqueName()
	vars := nameVars(name)
	var firstID string

	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t, supported) },
		Steps: []resource.TestStep{
			step("minimal", vars, attr("template_version", "1"), captureID(&firstID)),
			withPlanCheck(step("reformat_rename", vars,
				attr("name", name+"_renamed"),
				attr("template_version", "2"),
				attr("field_count", "1"),
				idUnchanged(&firstID),
			), plancheck.ExpectNonEmptyPlan()),
		},
	})
}

func TestAccResourceKibanaCaseTemplateNameEmpty(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t, supported) },
		Steps:    []resource.TestStep{planErrStep("invalid", `(?is)Invalid Attribute Value Length.*between 1 and 50`)},
	})
}

func TestAccResourceKibanaCaseTemplateNameTooLong(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t, supported) },
		Steps:    []resource.TestStep{planErrStep("invalid", `(?is)Invalid Attribute Value Length.*(at most|between) .*50`)},
	})
}

func TestAccResourceKibanaCaseTemplateDescriptionTooLong(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t, supported) },
		Steps:    []resource.TestStep{planErrStep("invalid", `(?is)Invalid Attribute Value Length.*at most 1000`)},
	})
}

func TestAccResourceKibanaCaseTemplateDefinitionTooLong(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t, supported) },
		Steps:    []resource.TestStep{planErrStep("invalid", `(?is)Invalid Attribute Value Length.*at most 30000`)},
	})
}
