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
	agentID := "test-agent-" + uuid.New().String()[:8]
	resourceID := testResourceID

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderAPIVersion) },
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"agent_id": config.StringVariable(agentID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "agent_id", agentID),
					resource.TestCheckResourceAttrSet(resourceID, "id"),
					resource.TestCheckResourceAttr(resourceID, "name", "Test Agent"),
					resource.TestCheckResourceAttr(resourceID, "description", "A test agent for acceptance testing"),
					resource.TestCheckResourceAttr(resourceID, "labels.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceID, "labels.*", "test"),
					resource.TestCheckTypeSetElemAttr(resourceID, "labels.*", "agent"),
					resource.TestCheckResourceAttr(resourceID, "tools.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceID, "tools.*", "platform.core.index_explorer"),
					resource.TestCheckResourceAttrSet(resourceID, "instructions"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
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
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"agent_id": config.StringVariable(agentID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "agent_id", agentID),
					resource.TestCheckResourceAttrSet(resourceID, "id"),
					resource.TestCheckResourceAttr(resourceID, "name", "Updated Test Agent"),
					resource.TestCheckResourceAttr(resourceID, "description", "An updated test agent"),
					resource.TestCheckResourceAttr(resourceID, "labels.#", "3"),
					resource.TestCheckResourceAttr(resourceID, "tools.#", "2"),
				),
			},
		},
	})
}

func TestAccResourceAgentBuilderAgentSpace(t *testing.T) {
	agentID := "test-agent-space-" + uuid.New().String()[:8]
	spaceID := fmt.Sprintf("test-space-%s", uuid.New().String()[:8])
	resourceID := testResourceID

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderAPIVersion) },
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
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
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
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
	agentID := "test-agent-" + uuid.New().String()[:8]

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderAPIVersion) },
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"agent_id": config.StringVariable(agentID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceID, "id"),
					resource.TestCheckResourceAttr(dataSourceID, "agent_id", agentID),
					resource.TestCheckResourceAttr(dataSourceID, "name", "Test Agent"),
					resource.TestCheckResourceAttr(dataSourceID, "include_dependencies", "false"),
					resource.TestCheckResourceAttr(dataSourceID, "tools.#", "0"),
				),
			},
		},
	})
}

func TestAccDataSourceKibanaAgentBuilderAgentWithDependencies(t *testing.T) {
	agentID := "test-agent-deps-" + uuid.New().String()[:8]
	esqlToolID := "test-esql-tool-" + uuid.New().String()[:8]
	workflowToolID := "test-wf-tool-" + uuid.New().String()[:8]

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderWorkflowAPIVersion) },
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"agent_id":         config.StringVariable(agentID),
					"esql_tool_id":     config.StringVariable(esqlToolID),
					"workflow_tool_id": config.StringVariable(workflowToolID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceID, "id"),
					resource.TestCheckResourceAttr(dataSourceID, "agent_id", agentID),
					resource.TestCheckResourceAttr(dataSourceID, "name", "Test Agent With Tools"),
					resource.TestCheckResourceAttr(dataSourceID, "include_dependencies", "true"),
					resource.TestCheckResourceAttr(dataSourceID, "tools.#", "2"),
				),
			},
		},
	})
}

func TestAccDataSourceKibanaAgentBuilderAgentWorkflowTool(t *testing.T) {
	agentID := "test-agent-wft-" + uuid.New().String()[:8]
	workflowToolID := "test-wf-tool-" + uuid.New().String()[:8]

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderWorkflowAPIVersion) },
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"agent_id":         config.StringVariable(agentID),
					"workflow_tool_id": config.StringVariable(workflowToolID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceID, "id"),
					resource.TestCheckResourceAttr(dataSourceID, "agent_id", agentID),
					resource.TestCheckResourceAttr(dataSourceID, "name", "Agent With Workflow Tool"),
					resource.TestCheckResourceAttr(dataSourceID, "include_dependencies", "true"),
					resource.TestCheckResourceAttr(dataSourceID, "tools.#", "1"),
					resource.TestCheckResourceAttrSet(dataSourceID, "tools.0.workflow_id"),
					resource.TestCheckResourceAttrSet(dataSourceID, "tools.0.workflow_configuration_yaml"),
				),
			},
		},
	})
}
