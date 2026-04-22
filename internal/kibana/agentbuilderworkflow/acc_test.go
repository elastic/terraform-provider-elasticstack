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

package agentbuilderworkflow_test

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
	minKibanaAgentBuilderAPIVersion = version.Must(version.NewVersion("9.4.0-SNAPSHOT"))
)

func TestAccResourceAgentBuilderWorkflow(t *testing.T) {
	// workflow IDs are workflow-<UUIDv4>
	workflowUUID := uuid.New()
	workflowID := "workflow-" + workflowUUID.String()
	resourceID := "elasticstack_kibana_agentbuilder_workflow.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderAPIVersion) },
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"workflow_id": config.StringVariable(workflowID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "workflow_id", workflowID),
					resource.TestCheckResourceAttr(resourceID, "space_id", "default"),
					resource.TestCheckResourceAttr(resourceID, "name", "Test Workflow"),
					resource.TestCheckResourceAttr(resourceID, "description", "A test workflow for acceptance testing"),
					resource.TestCheckResourceAttr(resourceID, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceID, "valid", "true"),
					resource.TestCheckResourceAttrSet(resourceID, "configuration_yaml"),
				),
			},
			{
				// Verify whitespace/indentation/blank-line changes are detected as updates.
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_whitespace"),
				ConfigVariables: config.Variables{
					"workflow_id": config.StringVariable(workflowID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "workflow_id", workflowID),
					resource.TestCheckResourceAttr(resourceID, "name", "Test Workflow"),
				),
			},
			{
				// Verify map key reordering is detected as an update.
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_reordered"),
				ConfigVariables: config.Variables{
					"workflow_id": config.StringVariable(workflowID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "workflow_id", workflowID),
					resource.TestCheckResourceAttr(resourceID, "name", "Test Workflow"),
				),
			},
			{
				// Import by composite id: <workflow_id>/<space_id>
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"workflow_id": config.StringVariable(workflowID),
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
					"workflow_id": config.StringVariable(workflowID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "workflow_id", workflowID),
					resource.TestCheckResourceAttr(resourceID, "name", "Updated Test Workflow"),
					resource.TestCheckResourceAttr(resourceID, "description", "An updated test workflow"),
					resource.TestCheckResourceAttr(resourceID, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceID, "valid", "true"),
				),
			},
		},
	})
}

func TestAccResourceAgentBuilderWorkflowSpace(t *testing.T) {
	workflowUUID := uuid.New()
	workflowID := "workflow-" + workflowUUID.String()
	spaceID := fmt.Sprintf("test-space-%s", uuid.New().String()[:8])
	resourceID := "elasticstack_kibana_agentbuilder_workflow.test_space"
	spaceResourceID := "elasticstack_kibana_space.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderAPIVersion) },
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"workflow_id": config.StringVariable(workflowID),
					"space_id":    config.StringVariable(spaceID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(spaceResourceID, "space_id", spaceID),
					resource.TestCheckResourceAttr(resourceID, "workflow_id", workflowID),
					resource.TestCheckResourceAttr(resourceID, "space_id", spaceID),
					resource.TestCheckResourceAttr(resourceID, "name", "Space Test Workflow"),
					resource.TestCheckResourceAttr(resourceID, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceID, "valid", "true"),
				),
			},
			{
				// Import by composite id: <workflow_id>/<space_id>
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"workflow_id": config.StringVariable(workflowID),
					"space_id":    config.StringVariable(spaceID),
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

func TestAccResourceAgentBuilderWorkflowInvalidCreate(t *testing.T) {
	workflowUUID := uuid.New()
	workflowID := "workflow-" + workflowUUID.String()

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderAPIVersion) },
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_invalid"),
				ConfigVariables: config.Variables{
					"workflow_id": config.StringVariable(workflowID),
				},
				ExpectError: regexp.MustCompile(`(?i)invalid workflow`),
			},
		},
	})
}

func TestAccResourceAgentBuilderWorkflowInvalidUpdate(t *testing.T) {
	workflowUUID := uuid.New()
	workflowID := "workflow-" + workflowUUID.String()
	resourceID := "elasticstack_kibana_agentbuilder_workflow.test_invalid"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderAPIVersion) },
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_valid"),
				ConfigVariables: config.Variables{
					"workflow_id": config.StringVariable(workflowID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "workflow_id", workflowID),
					resource.TestCheckResourceAttr(resourceID, "valid", "true"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_invalid"),
				ConfigVariables: config.Variables{
					"workflow_id": config.StringVariable(workflowID),
				},
				ExpectError: regexp.MustCompile(`(?i)invalid workflow`),
			},
		},
	})
}

func TestAccResourceAgentBuilderWorkflowAutoGeneratedID(t *testing.T) {
	resourceID := "elasticstack_kibana_agentbuilder_workflow.test_auto"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderAPIVersion) },
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceID, "workflow_id"),
					resource.TestCheckResourceAttr(resourceID, "name", "Auto ID Workflow"),
					resource.TestCheckResourceAttr(resourceID, "enabled", "true"),
				),
			},
		},
	})
}

func TestAccDataSourceKibanaAgentBuilderWorkflow(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderAPIVersion) },
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_agentbuilder_workflow.test", "id"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_agentbuilder_workflow.test", "workflow_id"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_agentbuilder_workflow.test", "configuration_yaml"),
				),
			},
		},
	})
}

func TestAccDataSourceKibanaAgentBuilderWorkflowSpace(t *testing.T) {
	spaceID := fmt.Sprintf("test-space-%s", uuid.New().String()[:8])

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheckWithWorkflowsEnabled(t, minKibanaAgentBuilderAPIVersion) },
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_agentbuilder_workflow.test", "id"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_agentbuilder_workflow.test", "workflow_id"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_agentbuilder_workflow.test", "configuration_yaml"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_agentbuilder_workflow.test", "space_id", spaceID),
				),
			},
		},
	})
}
