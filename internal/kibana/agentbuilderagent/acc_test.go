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

package agentbuilderagent_test

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
	testResourceID = "elasticstack_kibana_agentbuilder_agent.test"
	dataSourceID   = "data.elasticstack_kibana_agentbuilder_agent.test"
)

var (
	minKibanaAgentBuilderAPIVersion         = version.Must(version.NewVersion("9.3.0"))
	minKibanaAgentBuilderWorkflowAPIVersion = version.Must(version.NewVersion("9.4.0-SNAPSHOT"))
)

func TestAccResourceAgentBuilderAgent(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaAgentBuilderAPIVersion, versionutils.FlavorAny)

	agentID := "test-agent-" + uuid.New().String()[:8]
	resourceID := testResourceID

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderAPIVersion) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"agent_id": config.StringVariable(agentID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "agent_id", agentID),
					resource.TestCheckResourceAttr(resourceID, "id", "default/"+agentID),
					resource.TestCheckResourceAttr(resourceID, "name", "Test Agent"),
					resource.TestCheckResourceAttr(resourceID, "description", "A test agent for acceptance testing"),
					resource.TestCheckResourceAttr(resourceID, "labels.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceID, "labels.*", "test"),
					resource.TestCheckTypeSetElemAttr(resourceID, "labels.*", "agent"),
					resource.TestCheckResourceAttr(resourceID, "tools.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceID, "tools.*", "platform.core.index_explorer"),
					resource.TestCheckResourceAttr(resourceID, "instructions", "You are a helpful assistant that searches logs. Use the available tools to help answer questions."),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"agent_id": config.StringVariable(agentID),
				},
				ResourceName: resourceID,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return s.RootModule().Resources[resourceID].Primary.ID, nil
				},
				ImportStateVerify: true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"agent_id": config.StringVariable(agentID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "agent_id", agentID),
					resource.TestCheckResourceAttr(resourceID, "id", "default/"+agentID),
					resource.TestCheckResourceAttr(resourceID, "name", "Updated Test Agent"),
					resource.TestCheckResourceAttr(resourceID, "description", "An updated test agent"),
					resource.TestCheckResourceAttr(resourceID, "labels.#", "3"),
					resource.TestCheckResourceAttr(resourceID, "tools.#", "2"),
					resource.TestCheckResourceAttr(resourceID, "instructions", "You are an updated helpful assistant. Use the available tools wisely."),
				),
			},
		},
	})
}

func TestAccResourceAgentBuilderAgentSpace(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaAgentBuilderAPIVersion, versionutils.FlavorAny)

	agentID := "test-agent-space-" + uuid.New().String()[:8]
	spaceID := fmt.Sprintf("test-space-%s", uuid.New().String()[:8])
	resourceID := testResourceID

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderAPIVersion) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"agent_id": config.StringVariable(agentID),
					"space_id": config.StringVariable(spaceID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "agent_id", agentID),
					resource.TestCheckResourceAttrSet(resourceID, "id"),
					resource.TestCheckResourceAttr(resourceID, "space_id", spaceID),
					resource.TestCheckResourceAttr(resourceID, "name", "Space Agent"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"agent_id": config.StringVariable(agentID),
					"space_id": config.StringVariable(spaceID),
				},
				ResourceName: resourceID,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return s.RootModule().Resources[resourceID].Primary.ID, nil
				},
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDataSourceKibanaAgentBuilderAgent(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaAgentBuilderAPIVersion, versionutils.FlavorAny)

	agentID := "test-agent-" + uuid.New().String()[:8]
	spaceAgentID := "test-agent-space-" + uuid.New().String()[:8]
	spaceID := fmt.Sprintf("test-space-%s", uuid.New().String()[:8])

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderAPIVersion) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"agent_id": config.StringVariable(agentID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceID, "id", "default/"+agentID),
					resource.TestCheckResourceAttr(dataSourceID, "agent_id", agentID),
					resource.TestCheckResourceAttr(dataSourceID, "space_id", "default"),
					resource.TestCheckResourceAttr(dataSourceID, "name", "Test Agent"),
					resource.TestCheckResourceAttr(dataSourceID, "description", "A test agent for export"),
					resource.TestCheckResourceAttr(dataSourceID, "instructions", "You are a helpful assistant."),
					resource.TestCheckResourceAttr(dataSourceID, "include_dependencies", "false"),
					resource.TestCheckResourceAttr(dataSourceID, "tools.#", "0"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read_space"),
				ConfigVariables: config.Variables{
					"agent_id": config.StringVariable(spaceAgentID),
					"space_id": config.StringVariable(spaceID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceID, "id", spaceID+"/"+spaceAgentID),
					resource.TestCheckResourceAttr(dataSourceID, "agent_id", spaceAgentID),
					resource.TestCheckResourceAttr(dataSourceID, "space_id", spaceID),
					resource.TestCheckResourceAttr(dataSourceID, "name", "Space Agent"),
					resource.TestCheckResourceAttr(dataSourceID, "description", "A space-scoped agent for export"),
					resource.TestCheckResourceAttr(dataSourceID, "avatar_color", "#BFDBFF"),
					resource.TestCheckResourceAttr(dataSourceID, "avatar_symbol", "SA"),
					resource.TestCheckResourceAttr(dataSourceID, "labels.#", "2"),
					resource.TestCheckTypeSetElemAttr(dataSourceID, "labels.*", "agent"),
					resource.TestCheckTypeSetElemAttr(dataSourceID, "labels.*", "space"),
					resource.TestCheckResourceAttr(dataSourceID, "instructions", "You are a helpful assistant for a Kibana space."),
					resource.TestCheckResourceAttr(dataSourceID, "include_dependencies", "false"),
					resource.TestCheckResourceAttr(dataSourceID, "tools.#", "0"),
				),
			},
		},
	})
}

func TestAccDataSourceKibanaAgentBuilderAgentKibanaConnection(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaAgentBuilderAPIVersion, versionutils.FlavorAny)

	agentID := "test-agent-kbconn-" + uuid.New().String()[:8]

	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr(dataSourceID, "id", "default/"+agentID),
		resource.TestCheckResourceAttr(dataSourceID, "agent_id", agentID),
		resource.TestCheckResourceAttr(dataSourceID, "space_id", "default"),
		resource.TestCheckResourceAttr(dataSourceID, "name", "Test Agent (kibana_connection)"),
		resource.TestCheckResourceAttr(dataSourceID, "kibana_connection.#", "1"),
		resource.TestCheckResourceAttr(dataSourceID, "kibana_connection.0.endpoints.#", "1"),
		resource.TestCheckResourceAttr(dataSourceID, "kibana_connection.0.endpoints.0", strings.TrimSpace(os.Getenv("KIBANA_ENDPOINT"))),
		resource.TestCheckResourceAttr(dataSourceID, "kibana_connection.0.insecure", "false"),
	}
	checks = append(checks, acctest.KibanaConnectionAuthChecks(dataSourceID)...)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheckWithExplicitKibanaEndpoint(t)
			acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderAPIVersion)
		},
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: acctest.KibanaConnectionVariables(config.Variables{
					"agent_id": config.StringVariable(agentID),
				}),
				Check: resource.ComposeAggregateTestCheckFunc(checks...),
			},
		},
	})
}

func TestAccDataSourceKibanaAgentBuilderAgentWithDependencies(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaAgentBuilderAPIVersion, versionutils.FlavorAny)

	agentID := "test-agent-deps-" + uuid.New().String()[:8]
	esqlToolID := "test-esql-tool-" + uuid.New().String()[:8]
	workflowToolID := "test-wf-tool-" + uuid.New().String()[:8]

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderWorkflowAPIVersion) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"agent_id":         config.StringVariable(agentID),
					"esql_tool_id":     config.StringVariable(esqlToolID),
					"workflow_tool_id": config.StringVariable(workflowToolID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceID, "id", "default/"+agentID),
					resource.TestCheckResourceAttr(dataSourceID, "agent_id", agentID),
					resource.TestCheckResourceAttr(dataSourceID, "space_id", "default"),
					resource.TestCheckResourceAttr(dataSourceID, "name", "Test Agent With Tools"),
					resource.TestCheckResourceAttr(dataSourceID, "include_dependencies", "true"),
					resource.TestCheckResourceAttr(dataSourceID, "tools.#", "2"),
					resource.TestCheckResourceAttr(dataSourceID, "tools.0.id", "default/"+esqlToolID),
					resource.TestCheckResourceAttr(dataSourceID, "tools.0.space_id", "default"),
					resource.TestCheckResourceAttr(dataSourceID, "tools.0.tool_id", esqlToolID),
					resource.TestCheckResourceAttr(dataSourceID, "tools.0.type", "esql"),
					resource.TestCheckResourceAttr(dataSourceID, "tools.0.description", "Test ES|QL tool"),
					resource.TestCheckResourceAttr(dataSourceID, "tools.0.readonly", "false"),
					resource.TestCheckResourceAttr(dataSourceID, "tools.0.tags.#", "2"),
					resource.TestCheckTypeSetElemAttr(dataSourceID, "tools.0.tags.*", "esql"),
					resource.TestCheckTypeSetElemAttr(dataSourceID, "tools.0.tags.*", "test"),
					resource.TestCheckResourceAttrPair(dataSourceID, "tools.0.configuration", "elasticstack_kibana_agentbuilder_tool.esql", "configuration"),
					resource.TestCheckResourceAttr(dataSourceID, "tools.1.id", "default/"+workflowToolID),
					resource.TestCheckResourceAttr(dataSourceID, "tools.1.space_id", "default"),
					resource.TestCheckResourceAttr(dataSourceID, "tools.1.tool_id", workflowToolID),
					resource.TestCheckResourceAttr(dataSourceID, "tools.1.type", "workflow"),
					resource.TestCheckResourceAttr(dataSourceID, "tools.1.description", "Workflow tool"),
					resource.TestCheckResourceAttr(dataSourceID, "tools.1.readonly", "false"),
					resource.TestCheckResourceAttrPair(dataSourceID, "tools.1.configuration", "elasticstack_kibana_agentbuilder_tool.workflow", "configuration"),
					resource.TestCheckResourceAttrPair(dataSourceID, "tools.1.workflow_id", "elasticstack_kibana_agentbuilder_workflow.test", "workflow_id"),
					resource.TestCheckResourceAttrPair(dataSourceID, "tools.1.workflow_configuration_yaml", "elasticstack_kibana_agentbuilder_workflow.test", "configuration_yaml"),
				),
			},
		},
	})
}

func TestAccResourceAgentBuilderAgentAvatar(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaAgentBuilderAPIVersion, versionutils.FlavorAny)

	agentID := "test-agent-avatar-" + uuid.New().String()[:8]

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderAPIVersion) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_avatar"),
				ConfigVariables: config.Variables{
					"agent_id": config.StringVariable(agentID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceID, "avatar_color", "#BFDBFF"),
					resource.TestCheckResourceAttr(testResourceID, "avatar_symbol", "TA"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_no_avatar"),
				ConfigVariables: config.Variables{
					"agent_id": config.StringVariable(agentID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceID, "avatar_color", ""),
					resource.TestCheckResourceAttr(testResourceID, "avatar_symbol", ""),
				),
			},
		},
	})
}

func TestAccResourceAgentBuilderAgentKibanaConnection(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaAgentBuilderAPIVersion, versionutils.FlavorAny)

	agentID := "test-agent-kbconn-" + uuid.New().String()[:8]

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheckWithExplicitKibanaEndpoint(t)
			acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderAPIVersion)
		},
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: acctest.KibanaConnectionVariables(config.Variables{
					"agent_id": config.StringVariable(agentID),
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceID, "kibana_connection.#", "1"),
					resource.TestCheckResourceAttr(testResourceID, "kibana_connection.0.insecure", "false"),
				),
			},
		},
	})
}

func TestAccResourceAgentBuilderAgentSkillIds(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaAgentBuilderWorkflowAPIVersion, versionutils.FlavorAny)

	agentID := "test-agent-skills-" + uuid.New().String()[:8]
	skillID := "test-skill-" + uuid.New().String()[:8]

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderWorkflowAPIVersion) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"agent_id": config.StringVariable(agentID),
					"skill_id": config.StringVariable(skillID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceID, "agent_id", agentID),
					resource.TestCheckResourceAttr(testResourceID, "name", "Test Agent With Skills"),
					resource.TestCheckResourceAttr(testResourceID, "skill_ids.#", "1"),
					resource.TestCheckTypeSetElemAttr(testResourceID, "skill_ids.*", skillID),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"agent_id": config.StringVariable(agentID),
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
				// Update clears skill_ids to verify the field can be removed.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"agent_id": config.StringVariable(agentID),
					"skill_id": config.StringVariable(skillID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceID, "agent_id", agentID),
					resource.TestCheckResourceAttr(testResourceID, "name", "Test Agent With Skills Updated"),
					resource.TestCheckNoResourceAttr(testResourceID, "skill_ids"),
				),
			},
		},
	})
}

func TestAccDataSourceKibanaAgentBuilderAgentSkillIds(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaAgentBuilderWorkflowAPIVersion, versionutils.FlavorAny)

	agentID := "test-agent-skills-ds-" + uuid.New().String()[:8]
	skillID := "test-skill-" + uuid.New().String()[:8]

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderWorkflowAPIVersion) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"agent_id": config.StringVariable(agentID),
					"skill_id": config.StringVariable(skillID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceID, "id", testResourceID, "id"),
					resource.TestCheckResourceAttr(dataSourceID, "agent_id", agentID),
					resource.TestCheckResourceAttr(dataSourceID, "space_id", "default"),
					resource.TestCheckResourceAttr(dataSourceID, "name", "Test Agent With Skills"),
					resource.TestCheckResourceAttr(dataSourceID, "skill_ids.#", "1"),
					resource.TestCheckTypeSetElemAttr(dataSourceID, "skill_ids.*", skillID),
				),
			},
		},
	})
}

func TestAccDataSourceKibanaAgentBuilderAgentWorkflowTool(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaAgentBuilderAPIVersion, versionutils.FlavorAny)

	agentID := "test-agent-wft-" + uuid.New().String()[:8]
	workflowToolID := "test-wf-tool-" + uuid.New().String()[:8]

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderWorkflowAPIVersion) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"agent_id":         config.StringVariable(agentID),
					"workflow_tool_id": config.StringVariable(workflowToolID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceID, "id", "default/"+agentID),
					resource.TestCheckResourceAttr(dataSourceID, "agent_id", agentID),
					resource.TestCheckResourceAttr(dataSourceID, "space_id", "default"),
					resource.TestCheckResourceAttr(dataSourceID, "name", "Agent With Workflow Tool"),
					resource.TestCheckResourceAttr(dataSourceID, "include_dependencies", "true"),
					resource.TestCheckResourceAttr(dataSourceID, "tools.#", "1"),
					resource.TestCheckResourceAttr(dataSourceID, "tools.0.id", "default/"+workflowToolID),
					resource.TestCheckResourceAttr(dataSourceID, "tools.0.space_id", "default"),
					resource.TestCheckResourceAttr(dataSourceID, "tools.0.tool_id", workflowToolID),
					resource.TestCheckResourceAttr(dataSourceID, "tools.0.type", "workflow"),
					resource.TestCheckResourceAttr(dataSourceID, "tools.0.description", "Workflow tool"),
					resource.TestCheckResourceAttr(dataSourceID, "tools.0.readonly", "false"),
					resource.TestCheckResourceAttrPair(dataSourceID, "tools.0.configuration", "elasticstack_kibana_agentbuilder_tool.workflow", "configuration"),
					resource.TestCheckResourceAttrPair(dataSourceID, "tools.0.workflow_id", "elasticstack_kibana_agentbuilder_workflow.test", "workflow_id"),
					resource.TestCheckResourceAttrPair(dataSourceID, "tools.0.workflow_configuration_yaml", "elasticstack_kibana_agentbuilder_workflow.test", "configuration_yaml"),
				),
			},
		},
	})
}
