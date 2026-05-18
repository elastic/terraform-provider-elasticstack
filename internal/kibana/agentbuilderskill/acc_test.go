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

package agentbuilderskill_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/google/uuid"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const (
	testResourceID = "elasticstack_kibana_agentbuilder_skill.test"
	dataSourceID   = "data.elasticstack_kibana_agentbuilder_skill.test"
)

var minKibanaAgentBuilderSkillsAPIVersion = version.Must(version.NewVersion("9.4.0-SNAPSHOT"))

func TestAccResourceAgentBuilderSkill(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaAgentBuilderSkillsAPIVersion, versionutils.FlavorAny)

	skillID := "test-skill-" + uuid.New().String()[:8]

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"skill_id": config.StringVariable(skillID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceID, "skill_id", skillID),
					resource.TestCheckResourceAttr(testResourceID, "id", "default/"+skillID),
					resource.TestCheckResourceAttr(testResourceID, "space_id", "default"),
					resource.TestCheckResourceAttr(testResourceID, "name", "Test Skill"),
					resource.TestCheckResourceAttr(testResourceID, "description", "A test skill for acceptance testing"),
					resource.TestCheckResourceAttr(testResourceID, "content", "Always be helpful and accurate."),
					resource.TestCheckNoResourceAttr(testResourceID, "tool_ids"),
					resource.TestCheckNoResourceAttr(testResourceID, "referenced_content"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"skill_id": config.StringVariable(skillID),
				},
				ResourceName: testResourceID,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return s.RootModule().Resources[testResourceID].Primary.ID, nil
				},
				ImportStateVerify: true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"skill_id": config.StringVariable(skillID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceID, "skill_id", skillID),
					resource.TestCheckResourceAttr(testResourceID, "name", "Updated Test Skill"),
					resource.TestCheckResourceAttr(testResourceID, "description", "Updated description"),
					resource.TestCheckResourceAttr(testResourceID, "content", "Be precise and cite sources."),
					resource.TestCheckResourceAttr(testResourceID, "tool_ids.#", "1"),
					resource.TestCheckTypeSetElemAttr(testResourceID, "tool_ids.*", "platform.core.index_explorer"),
					resource.TestCheckResourceAttr(testResourceID, "referenced_content.#", "2"),
					resource.TestCheckResourceAttr(testResourceID, "referenced_content.0.name", "Runbook"),
					resource.TestCheckResourceAttr(testResourceID, "referenced_content.0.relative_path", "./runbooks/standard.md"),
					resource.TestCheckResourceAttr(testResourceID, "referenced_content.0.content", "First entry"),
					resource.TestCheckResourceAttr(testResourceID, "referenced_content.1.name", "Glossary"),
					resource.TestCheckResourceAttr(testResourceID, "referenced_content.1.relative_path", "./reference/glossary.md"),
					resource.TestCheckResourceAttr(testResourceID, "referenced_content.1.content", "Second entry"),
				),
			},
			{
				// Import after update to verify referenced_content ordering and tool_ids round-trip cleanly.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"skill_id": config.StringVariable(skillID),
				},
				ResourceName: testResourceID,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return s.RootModule().Resources[testResourceID].Primary.ID, nil
				},
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceAgentBuilderSkillFull(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaAgentBuilderSkillsAPIVersion, versionutils.FlavorAny)

	skillID := "test-skill-" + uuid.New().String()[:8]

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_full"),
				ConfigVariables: config.Variables{
					"skill_id": config.StringVariable(skillID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceID, "skill_id", skillID),
					resource.TestCheckResourceAttr(testResourceID, "tool_ids.#", "1"),
					resource.TestCheckResourceAttr(testResourceID, "referenced_content.#", "1"),
					resource.TestCheckResourceAttr(testResourceID, "referenced_content.0.name", "Initial"),
					resource.TestCheckResourceAttr(testResourceID, "referenced_content.0.relative_path", "./initial/path.md"),
				),
			},
			{
				// Import after full create to verify tool_ids + referenced_content round-trip.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_full"),
				ConfigVariables: config.Variables{
					"skill_id": config.StringVariable(skillID),
				},
				ResourceName: testResourceID,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return s.RootModule().Resources[testResourceID].Primary.ID, nil
				},
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceAgentBuilderSkillSpace(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaAgentBuilderSkillsAPIVersion, versionutils.FlavorAny)

	skillID := "test-skill-space-" + uuid.New().String()[:8]
	spaceID := fmt.Sprintf("test-space-%s", uuid.New().String()[:8])

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"skill_id": config.StringVariable(skillID),
					"space_id": config.StringVariable(spaceID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test", "space_id", spaceID),
					resource.TestCheckResourceAttr(testResourceID, "skill_id", skillID),
					resource.TestCheckResourceAttr(testResourceID, "space_id", spaceID),
					resource.TestCheckResourceAttr(testResourceID, "name", "Space Skill"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"skill_id": config.StringVariable(skillID),
					"space_id": config.StringVariable(spaceID),
				},
				ResourceName: testResourceID,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return s.RootModule().Resources[testResourceID].Primary.ID, nil
				},
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceAgentBuilderSkillKibanaConnection(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaAgentBuilderSkillsAPIVersion, versionutils.FlavorAny)

	skillID := "test-skill-kbconn-" + uuid.New().String()[:8]

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheckWithExplicitKibanaEndpoint(t)
			acctest.PreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: acctest.KibanaConnectionVariables(config.Variables{
					"skill_id": config.StringVariable(skillID),
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceID, "skill_id", skillID),
					resource.TestCheckResourceAttr(testResourceID, "space_id", "default"),
					resource.TestCheckResourceAttr(testResourceID, "kibana_connection.#", "1"),
					resource.TestCheckResourceAttrSet(testResourceID, "kibana_connection.0.endpoints.0"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: acctest.KibanaConnectionVariables(config.Variables{
					"skill_id": config.StringVariable(skillID),
				}),
				ResourceName:            testResourceID,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"kibana_connection"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return s.RootModule().Resources[testResourceID].Primary.ID, nil
				},
			},
		},
	})
}

func TestAccDataSourceKibanaAgentBuilderSkill(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaAgentBuilderSkillsAPIVersion, versionutils.FlavorAny)

	skillID := "test-skill-" + uuid.New().String()[:8]

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"skill_id": config.StringVariable(skillID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceID, "id", testResourceID, "id"),
					resource.TestCheckResourceAttrPair(dataSourceID, "skill_id", testResourceID, "skill_id"),
					resource.TestCheckResourceAttrPair(dataSourceID, "name", testResourceID, "name"),
					resource.TestCheckResourceAttrPair(dataSourceID, "description", testResourceID, "description"),
					resource.TestCheckResourceAttrPair(dataSourceID, "content", testResourceID, "content"),
					resource.TestCheckResourceAttr(dataSourceID, "space_id", "default"),
					resource.TestCheckResourceAttr(dataSourceID, "referenced_content.#", "1"),
					resource.TestCheckResourceAttr(dataSourceID, "referenced_content.0.name", "Exported"),
					resource.TestCheckResourceAttr(dataSourceID, "referenced_content.0.relative_path", "./exported/path.md"),
				),
			},
		},
	})
}

func TestAccDataSourceKibanaAgentBuilderSkillSpace(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaAgentBuilderSkillsAPIVersion, versionutils.FlavorAny)

	skillID := "test-skill-space-ds-" + uuid.New().String()[:8]
	spaceID := fmt.Sprintf("test-space-%s", uuid.New().String()[:8])

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"skill_id": config.StringVariable(skillID),
					"space_id": config.StringVariable(spaceID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceID, "id", testResourceID, "id"),
					resource.TestCheckResourceAttrPair(dataSourceID, "skill_id", testResourceID, "skill_id"),
					resource.TestCheckResourceAttr(dataSourceID, "space_id", spaceID),
					resource.TestCheckResourceAttrPair(dataSourceID, "name", testResourceID, "name"),
				),
			},
		},
	})
}

func TestAccDataSourceKibanaAgentBuilderSkillKibanaConnection(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaAgentBuilderSkillsAPIVersion, versionutils.FlavorAny)

	skillID := "test-skill-ds-kbconn-" + uuid.New().String()[:8]

	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttrPair(dataSourceID, "id", testResourceID, "id"),
		resource.TestCheckResourceAttr(dataSourceID, "skill_id", skillID),
		resource.TestCheckResourceAttr(dataSourceID, "space_id", "default"),
		resource.TestCheckResourceAttr(dataSourceID, "name", "Skill datasource kibana_connection"),
		resource.TestCheckResourceAttr(dataSourceID, "kibana_connection.#", "1"),
		resource.TestCheckResourceAttr(dataSourceID, "kibana_connection.0.endpoints.#", "1"),
		resource.TestCheckResourceAttr(dataSourceID, "kibana_connection.0.endpoints.0", strings.TrimSpace(os.Getenv("KIBANA_ENDPOINT"))),
		resource.TestCheckResourceAttr(dataSourceID, "kibana_connection.0.insecure", "false"),
	}
	checks = append(checks, acctest.KibanaConnectionAuthChecks(dataSourceID)...)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheckWithExplicitKibanaEndpoint(t)
			acctest.PreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: acctest.KibanaConnectionVariables(config.Variables{
					"skill_id": config.StringVariable(skillID),
				}),
				Check: resource.ComposeAggregateTestCheckFunc(checks...),
			},
		},
	})
}
