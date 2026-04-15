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

package agentbuildertool_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/google/uuid"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	minKibanaAgentBuilderAPIVersion         = version.Must(version.NewVersion("9.3.0"))
	minKibanaAgentBuilderWorkflowAPIVersion = version.Must(version.NewVersion("9.4.0-SNAPSHOT"))
	below940Constraints                     = version.MustConstraints(version.NewConstraint(">= 9.3.0, < 9.4.0-SNAPSHOT"))
)

func TestAccResourceAgentBuilderToolEsql(t *testing.T) {
	toolID := "test-esql-tool-" + uuid.New().String()[:8]
	resourceID := "elasticstack_kibana_agentbuilder_tool.test_esql"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderAPIVersion) },
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"tool_id": config.StringVariable(toolID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "tool_id", toolID),
					resource.TestCheckResourceAttr(resourceID, "space_id", "default"),
					resource.TestCheckResourceAttr(resourceID, "type", "esql"),
					resource.TestCheckResourceAttr(resourceID, "description", "Test ES|QL tool"),
					resource.TestCheckResourceAttr(resourceID, "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceID, "tags.*", "test"),
					resource.TestCheckTypeSetElemAttr(resourceID, "tags.*", "esql"),
					resource.TestCheckResourceAttrSet(resourceID, "configuration"),
				),
			},
			{
				// Import by composite id: <tool_id>/<space_id>
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"tool_id": config.StringVariable(toolID),
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
					"tool_id": config.StringVariable(toolID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "tool_id", toolID),
					resource.TestCheckResourceAttr(resourceID, "description", "Updated ES|QL tool"),
					resource.TestCheckResourceAttr(resourceID, "tags.#", "3"),
				),
			},
		},
	})
}

func TestAccResourceAgentBuilderToolEsqlSpace(t *testing.T) {
	toolID := "test-esql-tool-" + uuid.New().String()[:8]
	spaceID := fmt.Sprintf("test-space-%s", uuid.New().String()[:8])
	resourceID := "elasticstack_kibana_agentbuilder_tool.test_esql"
	spaceResourceID := "elasticstack_kibana_space.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderAPIVersion) },
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"tool_id":  config.StringVariable(toolID),
					"space_id": config.StringVariable(spaceID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(spaceResourceID, "space_id", spaceID),
					resource.TestCheckResourceAttr(resourceID, "tool_id", toolID),
					resource.TestCheckResourceAttr(resourceID, "space_id", spaceID),
					resource.TestCheckResourceAttr(resourceID, "type", "esql"),
				),
			},
			{
				// Import by composite id: <tool_id>/<space_id>
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"tool_id":  config.StringVariable(toolID),
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

func TestAccResourceAgentBuilderToolWorkflow(t *testing.T) {
	toolID := "test-workflow-tool-" + uuid.New().String()[:8]
	resourceID := "elasticstack_kibana_agentbuilder_tool.test_workflow"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderAPIVersion) },
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderWorkflowAPIVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"tool_id": config.StringVariable(toolID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "tool_id", toolID),
					resource.TestCheckResourceAttr(resourceID, "space_id", "default"),
					resource.TestCheckResourceAttr(resourceID, "type", "workflow"),
					resource.TestCheckResourceAttr(resourceID, "description", "Test workflow tool"),
					resource.TestCheckResourceAttrSet(resourceID, "configuration"),
				),
			},
			{
				// Import by composite id: <tool_id>/<space_id>
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderWorkflowAPIVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"tool_id": config.StringVariable(toolID),
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

func TestAccDataSourceKibanaAgentBuilderTool(t *testing.T) {
	toolID := "test-tool-" + uuid.New().String()[:8]

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderAPIVersion) },
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"tool_id": config.StringVariable(toolID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_agentbuilder_tool.test", "id"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_agentbuilder_tool.test", "tool_id"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_agentbuilder_tool.test", "type"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_agentbuilder_tool.test", "configuration"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_agentbuilder_tool.test", "tool_id", toolID),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_agentbuilder_tool.test", "type", "esql"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_agentbuilder_tool.test", "space_id", "default"),
				),
			},
		},
	})
}

func TestAccDataSourceKibanaAgentBuilderToolWorkflow(t *testing.T) {
	toolID := "test-workflow-tool-" + uuid.New().String()[:8]

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderAPIVersion) },
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderWorkflowAPIVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"tool_id": config.StringVariable(toolID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_agentbuilder_tool.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_agentbuilder_tool.test", "type", "workflow"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_agentbuilder_tool.test", "workflow_id"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_agentbuilder_tool.test", "workflow_configuration_yaml"),
				),
			},
		},
	})
}

func TestAccDataSourceKibanaAgentBuilderToolWorkflowUnsupportedVersion(t *testing.T) {
	toolID := "test-workflow-tool-" + uuid.New().String()[:8]

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderAPIVersion) },
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(below940Constraints),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read_workflow_unsupported"),
				ConfigVariables: config.Variables{
					"tool_id": config.StringVariable(toolID),
				},
				ExpectError: regexp.MustCompile(`Unsupported server version`),
			},
		},
	})
}

func TestAccDataSourceKibanaAgentBuilderToolSpace(t *testing.T) {
	toolID := "test-tool-" + uuid.New().String()[:8]
	spaceID := fmt.Sprintf("test-space-%s", uuid.New().String()[:8])

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderAPIVersion) },
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"tool_id":  config.StringVariable(toolID),
					"space_id": config.StringVariable(spaceID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_agentbuilder_tool.test", "id"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_agentbuilder_tool.test", "tool_id"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_agentbuilder_tool.test", "configuration"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_agentbuilder_tool.test", "tool_id", toolID),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_agentbuilder_tool.test", "space_id", spaceID),
				),
			},
		},
	})
}
