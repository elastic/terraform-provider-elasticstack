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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
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
)

func preCheckWithWorkflowsEnabled(t *testing.T) {
	acctest.PreCheck(t)

	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		t.Fatalf("Failed to create API client: %v", err)
	}

	serverVersion, diags := client.ServerVersion(context.Background())
	if diags.HasError() {
		t.Fatalf("Failed to get server version: %v", diags)
	}
	if serverVersion.LessThan(minKibanaAgentBuilderAPIVersion) {
		t.Skipf("Skipping test: server version %s is below minimum %s", serverVersion, minKibanaAgentBuilderAPIVersion)
	}

	kibanaClient, err := client.GetKibanaOapiClient()
	if err != nil {
		t.Fatalf("Failed to get Kibana client: %v", err)
	}

	// Try the internal settings API endpoint
	settingsURL := fmt.Sprintf("%s/internal/kibana/settings/workflows:ui:enabled", kibanaClient.URL)
	body := map[string]any{
		"value": true,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Failed to marshal body: %v", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", settingsURL, bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatalf("Failed to create POST request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("kbn-xsrf", "true")
	req.Header.Set("x-elastic-internal-origin", "Kibana")

	resp, err := kibanaClient.HTTP.Do(req)
	if err != nil {
		t.Fatalf("Failed to enable workflows: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("Failed to enable workflows (status %d): %s. Make sure workflows are enabled in kibana.yml with 'xpack.aiAssistant.workflows.enabled: true'", resp.StatusCode, string(respBody))
	}
}

func TestAccResourceAgentBuilderToolEsql(t *testing.T) {
	toolID := "test-esql-tool-" + uuid.New().String()[:8]
	resourceID := "elasticstack_kibana_agentbuilder_tool.test_esql"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheckWithWorkflowsEnabled(t) },
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
					resource.TestCheckResourceAttr(resourceID, "tags.0", "test"),
					resource.TestCheckResourceAttr(resourceID, "tags.1", "esql"),
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
		PreCheck: func() { preCheckWithWorkflowsEnabled(t) },
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
		PreCheck: func() { preCheckWithWorkflowsEnabled(t) },
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
