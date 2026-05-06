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

package sourcemap_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const testAccApmSourceMapResourceName = "elasticstack_apm_source_map.test"

// TestAccResourceApmSourceMap_json tests creating a source map using
// sourcemap.json, asserts id is set and non-empty, and confirms clean destroy.
func TestAccResourceApmSourceMap_json(t *testing.T) {
	serviceName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory(""),
				ConfigVariables: config.Variables{
					"service_name":    config.StringVariable(serviceName),
					"service_version": config.StringVariable("1.0.0"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testAccApmSourceMapResourceName, "id"),
					resource.TestCheckResourceAttr(testAccApmSourceMapResourceName, "service_name", serviceName),
					resource.TestCheckResourceAttr(testAccApmSourceMapResourceName, "service_version", "1.0.0"),
					resource.TestCheckResourceAttr(testAccApmSourceMapResourceName, "bundle_filepath", "/static/js/test.min.js"),
					testCheckApmSourceMapIDNonEmpty(testAccApmSourceMapResourceName),
				),
			},
		},
	})
}

// TestAccResourceApmSourceMap_binary tests creating a source map using
// sourcemap.binary (base64-encoded content) and asserts id is set and non-empty.
func TestAccResourceApmSourceMap_binary(t *testing.T) {
	serviceName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory(""),
				ConfigVariables: config.Variables{
					"service_name":    config.StringVariable(serviceName),
					"service_version": config.StringVariable("1.0.0"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testAccApmSourceMapResourceName, "id"),
					resource.TestCheckResourceAttr(testAccApmSourceMapResourceName, "service_name", serviceName),
					resource.TestCheckResourceAttr(testAccApmSourceMapResourceName, "service_version", "1.0.0"),
					resource.TestCheckResourceAttr(testAccApmSourceMapResourceName, "bundle_filepath", "/static/js/test.min.js"),
					testCheckApmSourceMapIDNonEmpty(testAccApmSourceMapResourceName),
				),
			},
		},
	})
}

// TestAccResourceApmSourceMap_import tests importing a source map created in a
// named Kibana space using the composite import ID "<space_id>/<artifact_id>".
// It also verifies that a plain (no-slash) import ID works and leaves space_id unset.
func TestAccResourceApmSourceMap_import(t *testing.T) {
	serviceName := sdkacctest.RandomWithPrefix("tf-acc-test")
	suffix := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)
	spaceID := fmt.Sprintf("apm-import-%s", suffix)

	resourceName := testAccApmSourceMapResourceName

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// Step 1: create the resource in a named space.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory(""),
				ConfigVariables: config.Variables{
					"service_name":    config.StringVariable(serviceName),
					"service_version": config.StringVariable("1.0.0"),
					"space_id":        config.StringVariable(spaceID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "space_id", spaceID),
					resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
					resource.TestCheckResourceAttr(resourceName, "service_version", "1.0.0"),
					resource.TestCheckResourceAttr(resourceName, "bundle_filepath", "/static/js/test.min.js"),
				),
			},
			// Step 2: import using composite "<space_id>/<artifact_id>".
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory(""),
				ConfigVariables: config.Variables{
					"service_name":    config.StringVariable(serviceName),
					"service_version": config.StringVariable("1.0.0"),
					"space_id":        config.StringVariable(spaceID),
				},
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"sourcemap", "kibana_connection"},
				ImportStateIdFunc:       testAccApmSourceMapCompositeImportID(resourceName),
			},
			// Step 3: import using just the artifact id (no space prefix) —
			// results in space_id being unset (default space semantics).
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory(""),
				ConfigVariables: config.Variables{
					"service_name":    config.StringVariable(serviceName),
					"service_version": config.StringVariable("1.0.0"),
					"space_id":        config.StringVariable(spaceID),
				},
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: false, // space_id will differ; we verify manually
				ImportStateIdFunc: testAccApmSourceMapPlainImportID(resourceName),
				// After a plain import the space_id attribute is not set.
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					if len(states) != 1 {
						return fmt.Errorf("expected 1 imported state, got %d", len(states))
					}
					if v, ok := states[0].Attributes["space_id"]; ok && v != "" {
						return fmt.Errorf("expected space_id to be empty after plain import, got %q", v)
					}
					return nil
				},
			},
		},
	})
}

// TestAccResourceApmSourceMap_plainImport tests that a source map created in the
// default space (no space_id) can be imported using just the artifact id (no
// space prefix), and that space_id is empty after the import.
func TestAccResourceApmSourceMap_plainImport(t *testing.T) {
	serviceName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resourceName := testAccApmSourceMapResourceName

	vars := config.Variables{
		"service_name":    config.StringVariable(serviceName),
		"service_version": config.StringVariable("1.0.0"),
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// Step 1: create the resource in the default space (no space_id).
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory(""),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckNoResourceAttr(resourceName, "space_id"),
					resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
					resource.TestCheckResourceAttr(resourceName, "service_version", "1.0.0"),
					resource.TestCheckResourceAttr(resourceName, "bundle_filepath", "/static/js/test.min.js"),
				),
			},
			// Step 2: import using just the artifact id (plain, no space prefix).
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory(""),
				ConfigVariables:          vars,
				ResourceName:             resourceName,
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateVerifyIgnore:  []string{"sourcemap", "kibana_connection"},
				ImportStateIdFunc:        testAccApmSourceMapPlainImportID(resourceName),
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					if len(states) != 1 {
						return fmt.Errorf("expected 1 imported state, got %d", len(states))
					}
					if v, ok := states[0].Attributes["space_id"]; ok && v != "" {
						return fmt.Errorf("expected space_id to be empty after plain import, got %q", v)
					}
					return nil
				},
			},
		},
	})
}

// TestAccResourceApmSourceMap_space verifies that creating a source map with a
// non-default space_id routes all CRUD operations to that space.
func TestAccResourceApmSourceMap_space(t *testing.T) {
	serviceName := sdkacctest.RandomWithPrefix("tf-acc-test")
	suffix := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)
	spaceID := fmt.Sprintf("apm-sm-%s", suffix)

	resourceName := testAccApmSourceMapResourceName

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory(""),
				ConfigVariables: config.Variables{
					"service_name":    config.StringVariable(serviceName),
					"service_version": config.StringVariable("1.0.0"),
					"space_id":        config.StringVariable(spaceID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "space_id", spaceID),
					resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
					resource.TestCheckResourceAttr(resourceName, "service_version", "1.0.0"),
					resource.TestCheckResourceAttr(resourceName, "bundle_filepath", "/static/js/test.min.js"),
				),
			},
		},
	})
}

// TestAccResourceApmSourceMap_validationNeitherSet verifies that applying a config
// with none of sourcemap.json, sourcemap.binary, or sourcemap.file.path returns a validation diagnostic.
func TestAccResourceApmSourceMap_validationNeitherSet(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory(""),
				ExpectError:              regexp.MustCompile(`(?i)one \(and only one\)`),
			},
		},
	})
}

// TestAccResourceApmSourceMap_validationBothSet verifies that applying a config
// with both sourcemap.json and sourcemap.binary returns a validation diagnostic.
func TestAccResourceApmSourceMap_validationBothSet(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory(""),
				ExpectError:              regexp.MustCompile(`(?i)one \(and only one\)`),
			},
		},
	})
}

// TestAccResourceApmSourceMap_validationEmptyString verifies that setting
// sourcemap.json to an empty string triggers the LengthAtLeast validation.
func TestAccResourceApmSourceMap_validationEmptyString(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory(""),
				ExpectError:              regexp.MustCompile(`(?i)at least 1`),
			},
		},
	})
}

// TestAccResourceApmSourceMap_binaryInvalidBase64 verifies that setting
// sourcemap.binary to a non-base64 string causes an error at apply time.
func TestAccResourceApmSourceMap_binaryInvalidBase64(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory(""),
				ExpectError:              regexp.MustCompile(`(?i)(base64|decod)`),
			},
		},
	})
}

// TestAccResourceApmSourceMap_requireReplace verifies that changing service_version
// produces a ResourceActionDestroyBeforeCreate plan action, not an in-place update.
func TestAccResourceApmSourceMap_requireReplace(t *testing.T) {
	serviceName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// Step 1: apply with service_version = "1.0.0".
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("step1"),
				ConfigVariables: config.Variables{
					"service_name": config.StringVariable(serviceName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testAccApmSourceMapResourceName, "service_version", "1.0.0"),
					resource.TestCheckResourceAttrSet(testAccApmSourceMapResourceName, "id"),
				),
			},
			// Step 2: plan with service_version = "1.1.0" — must show replacement.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("step2"),
				ConfigVariables: config.Variables{
					"service_name": config.StringVariable(serviceName),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							testAccApmSourceMapResourceName,
							plancheck.ResourceActionDestroyBeforeCreate,
						),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testAccApmSourceMapResourceName, "service_version", "1.1.0"),
				),
			},
		},
	})
}

// testCheckApmSourceMapIDNonEmpty is a check function that asserts the id
// attribute is set and non-empty.
func testCheckApmSourceMapIDNonEmpty(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[resourceName]
		if !ok || res.Primary == nil {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		id := res.Primary.Attributes["id"]
		if id == "" {
			return fmt.Errorf("expected non-empty id for %s, got empty string", resourceName)
		}
		return nil
	}
}

// testAccApmSourceMapCompositeImportID returns an ImportStateIdFunc that builds
// the "<space_id>/<artifact_id>" composite import ID from state.
func testAccApmSourceMapCompositeImportID(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		res, ok := s.RootModule().Resources[resourceName]
		if !ok || res.Primary == nil {
			return "", fmt.Errorf("resource %s not found in state", resourceName)
		}
		spaceID := res.Primary.Attributes["space_id"]
		id := res.Primary.Attributes["id"]
		if spaceID == "" {
			return "", fmt.Errorf("space_id is empty in state for %s", resourceName)
		}
		if id == "" {
			return "", fmt.Errorf("id is empty in state for %s", resourceName)
		}
		return spaceID + "/" + id, nil
	}
}

// testAccApmSourceMapPlainImportID returns an ImportStateIdFunc that uses only
// the artifact id (no space prefix) from state.
func testAccApmSourceMapPlainImportID(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		res, ok := s.RootModule().Resources[resourceName]
		if !ok || res.Primary == nil {
			return "", fmt.Errorf("resource %s not found in state", resourceName)
		}
		id := res.Primary.Attributes["id"]
		if id == "" {
			return "", fmt.Errorf("id is empty in state for %s", resourceName)
		}
		return id, nil
	}
}
