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

package securitylist_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const securityListResourceName = "elasticstack_kibana_security_list.test"

// attrOf returns the current value of a security list attribute from state,
// used to capture pre-update values that are later asserted to have changed.
func attrOf(s *terraform.State, key string) (string, error) {
	rs, ok := s.RootModule().Resources[securityListResourceName]
	if !ok {
		return "", fmt.Errorf("resource %s not found in state", securityListResourceName)
	}
	return rs.Primary.Attributes[key], nil
}

func captureAttr(key string, dst *string) resource.TestCheckFunc {
	return func(s *terraform.State) (err error) {
		*dst, err = attrOf(s, key)
		return err
	}
}

func attrChanged(key string, prev *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		got, err := attrOf(s, key)
		if err != nil {
			return err
		}
		if got == *prev {
			return fmt.Errorf("expected %s to change from %q after update", key, *prev)
		}
		return nil
	}
}

func TestAccResourceSecurityList(t *testing.T) {
	listID := "test-list-" + uuid.New().String()
	spaceID := "test-space-" + uuid.New().String()[:8]
	var preUpdateUpdatedAt, preUpdateVersionID string
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers, // Create
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"space_id":    config.StringVariable(spaceID),
					"list_id":     config.StringVariable(listID),
					"name":        config.StringVariable("Test Security List"),
					"description": config.StringVariable("A test security list for IP addresses"),
					"type":        config.StringVariable("ip"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list.test", "space_id", spaceID),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list.test", "list_id", listID),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list.test", "name", "Test Security List"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list.test", "description", "A test security list for IP addresses"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list.test", "type", "ip"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list.test", "version"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list.test", "created_at"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list.test", "created_by"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list.test", "updated_at"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list.test", "updated_by"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list.test", "version_id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list.test", "immutable", "false"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list.test", "tie_breaker_id"),
					captureAttr("updated_at", &preUpdateUpdatedAt),
					captureAttr("version_id", &preUpdateVersionID),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers, // Update
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"space_id":    config.StringVariable(spaceID),
					"list_id":     config.StringVariable(listID),
					"name":        config.StringVariable("Updated Security List"),
					"description": config.StringVariable("An updated test security list"),
					"type":        config.StringVariable("ip"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list.test", "name", "Updated Security List"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list.test", "description", "An updated test security list"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list.test", "version"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list.test", "updated_by"),
					attrChanged("updated_at", &preUpdateUpdatedAt),
					attrChanged("version_id", &preUpdateVersionID),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers, // Import
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"space_id":    config.StringVariable(spaceID),
					"list_id":     config.StringVariable(listID),
					"name":        config.StringVariable("Updated Security List"),
					"description": config.StringVariable("An updated test security list"),
					"type":        config.StringVariable("ip"),
				},
				ResourceName:      "elasticstack_kibana_security_list.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceSecurityList_KeywordType(t *testing.T) {
	listID := "keyword-list-" + uuid.New().String()
	spaceID := "test-space-" + uuid.New().String()[:8]
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("keyword_type"),
				ConfigVariables: config.Variables{
					"space_id":    config.StringVariable(spaceID),
					"list_id":     config.StringVariable(listID),
					"name":        config.StringVariable("Keyword Security List"),
					"description": config.StringVariable("A test security list for keywords"),
					"type":        config.StringVariable("keyword"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list.test", "type", "keyword"),
				),
			},
		},
	})
}

func TestAccResourceSecurityList_WithMetadata(t *testing.T) {
	listID := "meta-list-" + uuid.New().String()
	spaceID := "test-space-" + uuid.New().String()[:8]
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers, // Create with metadata
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_with_meta"),
				ConfigVariables: config.Variables{
					"space_id":    config.StringVariable(spaceID),
					"list_id":     config.StringVariable(listID),
					"name":        config.StringVariable("List with Metadata"),
					"description": config.StringVariable("A test list with metadata"),
					"type":        config.StringVariable("ip"),
					"meta":        config.StringVariable(`{"author":"test-user","category":"network","tags":["production","firewall"]}`),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list.test", "name", "List with Metadata"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list.test", "type", "ip"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list.test", "meta", `{"author":"test-user","category":"network","tags":["production","firewall"]}`),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers, // Update metadata
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_meta"),
				ConfigVariables: config.Variables{
					"space_id":    config.StringVariable(spaceID),
					"list_id":     config.StringVariable(listID),
					"name":        config.StringVariable("List with Metadata"),
					"description": config.StringVariable("A test list with updated metadata"),
					"type":        config.StringVariable("ip"),
					"meta":        config.StringVariable(`{"author":"updated-user","category":"security","tags":["staging","api"],"version":"2.0"}`),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list.test", "description", "A test list with updated metadata"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list.test", "meta", `{"author":"updated-user","category":"security","tags":["staging","api"],"version":"2.0"}`),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers, // Remove metadata
				ConfigDirectory:          acctest.NamedTestCaseDirectory("remove_meta"),
				ConfigVariables: config.Variables{
					"space_id":    config.StringVariable(spaceID),
					"list_id":     config.StringVariable(listID),
					"name":        config.StringVariable("List with Metadata"),
					"description": config.StringVariable("A test list with metadata removed"),
					"type":        config.StringVariable("ip"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("elasticstack_kibana_security_list.test", "meta"),
				),
			},
		},
	})
}

// TestAccResourceSecurityList_InvalidType asserts that an unsupported `type`
// value is rejected by the stringvalidator.OneOf validator before any API
// call is made.
func TestAccResourceSecurityList_InvalidType(t *testing.T) {
	listID := "invalid-type-list-" + uuid.New().String()
	spaceID := "test-space-" + uuid.New().String()[:8]
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"space_id":    config.StringVariable(spaceID),
					"list_id":     config.StringVariable(listID),
					"name":        config.StringVariable("Invalid Type Security List"),
					"description": config.StringVariable("A test security list with an invalid type"),
					"type":        config.StringVariable("not_a_real_type"),
				},
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?is)Invalid Attribute Value Match.*type|value must be one of`),
			},
		},
	})
}

// TestAccResourceSecurityList_GeneratedListID asserts that omitting list_id
// exercises the server-generated/Computed path: Kibana assigns the ID and the
// provider surfaces it back into state.
func TestAccResourceSecurityList_GeneratedListID(t *testing.T) {
	spaceID := "test-space-" + uuid.New().String()[:8]
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("generated"),
				ConfigVariables: config.Variables{
					"space_id":    config.StringVariable(spaceID),
					"name":        config.StringVariable("Generated List ID Security List"),
					"description": config.StringVariable("A test security list with a server-generated list_id"),
					"type":        config.StringVariable("ip"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(securityListResourceName, "list_id"),
					resource.TestCheckResourceAttrSet(securityListResourceName, "id"),
				),
			},
		},
	})
}

// TestAccResourceSecurityList_ReplaceOnChange asserts that changing `type`,
// `list_id`, or `space_id` forces a destroy-before-create replacement, and
// exercises additional `type` values beyond the `ip`/`keyword` cases covered
// by the other tests in this file.
func TestAccResourceSecurityList_ReplaceOnChange(t *testing.T) {
	listIDA := "replace-list-a-" + uuid.New().String()
	listIDB := "replace-list-b-" + uuid.New().String()
	spaceIDA := "test-space-a-" + uuid.New().String()[:8]
	spaceIDB := "test-space-b-" + uuid.New().String()[:8]

	baseVars := func(spaceID, listID, listType string) config.Variables {
		return config.Variables{
			"space_id":    config.StringVariable(spaceID),
			"list_id":     config.StringVariable(listID),
			"name":        config.StringVariable("Replace On Change Security List"),
			"description": config.StringVariable("A test security list exercising RequiresReplace"),
			"type":        config.StringVariable(listType),
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Step 1: create with type = "long".
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          baseVars(spaceIDA, listIDA, "long"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(securityListResourceName, "space_id", spaceIDA),
					resource.TestCheckResourceAttr(securityListResourceName, "list_id", listIDA),
					resource.TestCheckResourceAttr(securityListResourceName, "type", "long"),
				),
			},
			{
				// Step 2: changing type to "date" (same space, same list_id) must force a replace.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          baseVars(spaceIDA, listIDA, "date"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(securityListResourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(securityListResourceName, "type", "date"),
				),
			},
			{
				// Step 3: changing list_id (same space, same type) must force a replace.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          baseVars(spaceIDA, listIDB, "date"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(securityListResourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(securityListResourceName, "list_id", listIDB),
				),
			},
			{
				// Step 4: changing space_id (same list_id, same type) must force a replace.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          baseVars(spaceIDB, listIDB, "date"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(securityListResourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(securityListResourceName, "space_id", spaceIDB),
				),
			},
		},
	})
}
