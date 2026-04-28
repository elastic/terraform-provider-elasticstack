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

// testAccAgentBuilderEsqlResourceName is the address of the ESQL tool used in acc configs.
const testAccAgentBuilderEsqlResourceName = "elasticstack_kibana_agentbuilder_tool.test_esql"

// testAccAgentBuilderWorkflowResourceName is the address of the workflow tool used in acc configs.
const testAccAgentBuilderWorkflowResourceName = "elasticstack_kibana_agentbuilder_tool.test_workflow"

// testAccAgentBuilderToolDataSourceName is the address of the data source used in acc configs.
const testAccAgentBuilderToolDataSourceName = "data.elasticstack_kibana_agentbuilder_tool.test"

func TestAccResourceAgentBuilderToolEsql(t *testing.T) {
	toolID := "test-esql-tool-" + uuid.New().String()[:8]
	resourceID := testAccAgentBuilderEsqlResourceName

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
			{
				// Import after update to verify the post-update state round-trips (covers Configure + metadata after CRUD)
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"tool_id": config.StringVariable(toolID),
				},
				ResourceName:      resourceID,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return s.RootModule().Resources[resourceID].Primary.ID, nil
				},
			},
		},
	})
}

// TestAccResourceAgentBuilderToolEsqlKibanaConnection exercises a scoped
// kibana_connection block (Kibana client from r.Client) with import round-trip.
func TestAccResourceAgentBuilderToolEsqlKibanaConnection(t *testing.T) {
	toolID := "test-kb-conn-tool-" + uuid.New().String()[:8]
	resourceID := testAccAgentBuilderEsqlResourceName

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheckWithExplicitKibanaEndpoint(t)
			acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderAPIVersion)
		},
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: acctest.KibanaConnectionVariables(config.Variables{
					"tool_id": config.StringVariable(toolID),
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "tool_id", toolID),
					resource.TestCheckResourceAttr(resourceID, "space_id", "default"),
					resource.TestCheckResourceAttr(resourceID, "type", "esql"),
					resource.TestCheckResourceAttr(resourceID, "kibana_connection.#", "1"),
					resource.TestCheckResourceAttrSet(resourceID, "kibana_connection.0.endpoints.0"),
					resource.TestCheckResourceAttrSet(resourceID, "configuration"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: acctest.KibanaConnectionVariables(config.Variables{
					"tool_id": config.StringVariable(toolID),
				}),
				ResourceName:            resourceID,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"kibana_connection"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return s.RootModule().Resources[resourceID].Primary.ID, nil
				},
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: acctest.KibanaConnectionVariables(config.Variables{
					"tool_id": config.StringVariable(toolID),
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "tool_id", toolID),
					resource.TestCheckResourceAttr(resourceID, "description", "Updated ES|QL tool (kibana_connection)"),
					resource.TestCheckResourceAttr(resourceID, "tags.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceID, "tags.*", "updated"),
					resource.TestCheckResourceAttr(resourceID, "kibana_connection.#", "1"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: acctest.KibanaConnectionVariables(config.Variables{
					"tool_id": config.StringVariable(toolID),
				}),
				ResourceName:            resourceID,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"kibana_connection"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return s.RootModule().Resources[resourceID].Primary.ID, nil
				},
			},
		},
	})
}

func TestAccResourceAgentBuilderToolEsqlSpace(t *testing.T) {
	toolID := "test-esql-tool-" + uuid.New().String()[:8]
	spaceID := fmt.Sprintf("test-space-%s", uuid.New().String()[:8])
	resourceID := testAccAgentBuilderEsqlResourceName
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
	resourceID := testAccAgentBuilderWorkflowResourceName

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
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderWorkflowAPIVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"tool_id": config.StringVariable(toolID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "tool_id", toolID),
					resource.TestCheckResourceAttr(resourceID, "description", "Updated workflow tool"),
					resource.TestCheckResourceAttr(resourceID, "tags.#", "3"),
				),
			},
			{
				// Import after update (same pattern as TestAccResourceAgentBuilderToolEsql)
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderWorkflowAPIVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"tool_id": config.StringVariable(toolID),
				},
				ResourceName:      resourceID,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return s.RootModule().Resources[resourceID].Primary.ID, nil
				},
			},
		},
	})
}

// TestAccResourceAgentBuilderToolWorkflowKibanaConnection mirrors
// TestAccResourceAgentBuilderToolWorkflow with entity-local kibana_connection
// (9.4+ workflow tool API) and the ESQL kibana_connection import pattern.
func TestAccResourceAgentBuilderToolWorkflowKibanaConnection(t *testing.T) {
	toolID := "test-wf-kb-conn-" + uuid.New().String()[:8]
	resourceID := testAccAgentBuilderWorkflowResourceName

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheckWithExplicitKibanaEndpoint(t)
			acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderAPIVersion)
		},
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderWorkflowAPIVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: acctest.KibanaConnectionVariables(config.Variables{
					"tool_id": config.StringVariable(toolID),
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "tool_id", toolID),
					resource.TestCheckResourceAttr(resourceID, "type", "workflow"),
					resource.TestCheckResourceAttr(resourceID, "kibana_connection.#", "1"),
					resource.TestCheckResourceAttrSet(resourceID, "kibana_connection.0.endpoints.0"),
					resource.TestCheckResourceAttrSet(resourceID, "configuration"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderWorkflowAPIVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: acctest.KibanaConnectionVariables(config.Variables{
					"tool_id": config.StringVariable(toolID),
				}),
				ResourceName:            resourceID,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"kibana_connection"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return s.RootModule().Resources[resourceID].Primary.ID, nil
				},
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderWorkflowAPIVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: acctest.KibanaConnectionVariables(config.Variables{
					"tool_id": config.StringVariable(toolID),
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "tool_id", toolID),
					resource.TestCheckResourceAttr(resourceID, "description", "Updated workflow tool (kibana_connection)"),
					resource.TestCheckResourceAttr(resourceID, "tags.#", "4"),
					resource.TestCheckTypeSetElemAttr(resourceID, "tags.*", "updated"),
					resource.TestCheckResourceAttr(resourceID, "kibana_connection.#", "1"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderWorkflowAPIVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: acctest.KibanaConnectionVariables(config.Variables{
					"tool_id": config.StringVariable(toolID),
				}),
				ResourceName:            resourceID,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"kibana_connection"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return s.RootModule().Resources[resourceID].Primary.ID, nil
				},
			},
		},
	})
}

func TestAccDataSourceKibanaAgentBuilderTool(t *testing.T) {
	toolID := "test-tool-" + uuid.New().String()[:8]
	dsID := testAccAgentBuilderToolDataSourceName

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
					// id round-trip: exported composite ID must match the backing resource
					resource.TestCheckResourceAttrPair(dsID, "id",
						"elasticstack_kibana_agentbuilder_tool.test", "id"),
					// exact tool_id and type/space assertions
					resource.TestCheckResourceAttr(dsID, "tool_id", toolID),
					resource.TestCheckResourceAttr(dsID, "type", "esql"),
					resource.TestCheckResourceAttr(dsID, "space_id", "default"),
					// computed fields from the backing resource
					resource.TestCheckResourceAttr(dsID, "description", "Test ESQL tool"),
					resource.TestCheckResourceAttr(dsID, "tags.#", "1"),
					resource.TestCheckTypeSetElemAttr(dsID, "tags.*", "test"),
					resource.TestCheckResourceAttr(dsID, "readonly", "false"),
					// configuration contains the expected ESQL query fragment
					resource.TestMatchResourceAttr(dsID, "configuration",
						regexp.MustCompile(`FROM logs \| LIMIT 10`)),
					// workflow-only fields must be absent when include_workflow is not set
					resource.TestCheckNoResourceAttr(dsID, "workflow_id"),
					resource.TestCheckNoResourceAttr(dsID, "workflow_configuration_yaml"),
				),
			},
		},
	})
}

func TestAccDataSourceKibanaAgentBuilderToolWorkflow(t *testing.T) {
	toolID := "test-workflow-tool-" + uuid.New().String()[:8]
	dsID := testAccAgentBuilderToolDataSourceName

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
					// id round-trip: composite ID must match the backing resource
					resource.TestCheckResourceAttrPair(dsID, "id",
						"elasticstack_kibana_agentbuilder_tool.test", "id"),
					// exact tool_id, type, space_id assertions
					resource.TestCheckResourceAttr(dsID, "tool_id", toolID),
					resource.TestCheckResourceAttr(dsID, "type", "workflow"),
					resource.TestCheckResourceAttr(dsID, "space_id", "default"),
					resource.TestCheckResourceAttr(dsID, "description", "Workflow tool"),
					// workflow_id must equal the backing workflow resource's workflow_id
					resource.TestCheckResourceAttrPair(dsID, "workflow_id",
						"elasticstack_kibana_agentbuilder_workflow.test", "workflow_id"),
					// workflow YAML is populated when include_workflow = true
					resource.TestCheckResourceAttrSet(dsID, "workflow_configuration_yaml"),
					// configuration contains a workflow_id JSON key
					resource.TestMatchResourceAttr(dsID, "configuration",
						regexp.MustCompile(`"workflow_id"`)),
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
	dsID := testAccAgentBuilderToolDataSourceName

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
					// id round-trip: composite ID must match the backing resource
					resource.TestCheckResourceAttrPair(dsID, "id",
						"elasticstack_kibana_agentbuilder_tool.test", "id"),
					// exact tool_id and space_id
					resource.TestCheckResourceAttr(dsID, "tool_id", toolID),
					resource.TestCheckResourceAttr(dsID, "space_id", spaceID),
					// computed type/readonly
					resource.TestCheckResourceAttr(dsID, "type", "esql"),
					resource.TestCheckResourceAttr(dsID, "readonly", "false"),
					// configuration is populated
					resource.TestCheckResourceAttrSet(dsID, "configuration"),
				),
			},
		},
	})
}

// TestAccDataSourceKibanaAgentBuilderToolKibanaConnection exercises the data source
// with an entity-local kibana_connection block, mirroring the resource-level connection tests.
func TestAccDataSourceKibanaAgentBuilderToolKibanaConnection(t *testing.T) {
	toolID := "test-tool-ds-conn-" + uuid.New().String()[:8]
	dsID := testAccAgentBuilderToolDataSourceName

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheckWithExplicitKibanaEndpoint(t)
			acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderAPIVersion)
		},
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: acctest.KibanaConnectionVariables(config.Variables{
					"tool_id": config.StringVariable(toolID),
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dsID, "tool_id", toolID),
					resource.TestCheckResourceAttr(dsID, "space_id", "default"),
					resource.TestCheckResourceAttr(dsID, "type", "esql"),
					resource.TestCheckResourceAttr(dsID, "description", "Test ESQL tool (DS kibana_connection)"),
					resource.TestCheckResourceAttr(dsID, "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr(dsID, "tags.*", "test"),
					resource.TestCheckTypeSetElemAttr(dsID, "tags.*", "ds-conn"),
					resource.TestCheckResourceAttr(dsID, "kibana_connection.#", "1"),
					resource.TestCheckResourceAttrSet(dsID, "kibana_connection.0.endpoints.0"),
					resource.TestCheckResourceAttrSet(dsID, "configuration"),
				),
			},
		},
	})
}
