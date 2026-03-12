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

package abagent_test

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
	"github.com/hashicorp/go-version"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	providerConfig = `
provider "elasticstack" {
	kibana {}
}
`
)

var minKibanaAgentBuilderAPIVersion = version.Must(version.NewVersion("9.3.0"))

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

func TestAccResourceAgent(t *testing.T) {
	agentID := "test-agent-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceID := "elasticstack_kibana_ab_agent.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckWithWorkflowsEnabled(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				Config: providerConfig + `
resource "elasticstack_kibana_ab_agent" "test" {
	id          = "` + agentID + `"
	name        = "Test Agent"
	description = "A test agent for acceptance testing"
	labels      = ["test", "agent"]
	instructions = <<-EOT
You are a helpful assistant that searches logs.
Use the available tools to help answer questions.
EOT
	tools = ["platform.core.index_explorer"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "id", agentID),
					resource.TestCheckResourceAttr(resourceID, "name", "Test Agent"),
					resource.TestCheckResourceAttr(resourceID, "description", "A test agent for acceptance testing"),
					resource.TestCheckResourceAttr(resourceID, "labels.#", "2"),
					resource.TestCheckResourceAttr(resourceID, "labels.0", "test"),
					resource.TestCheckResourceAttr(resourceID, "labels.1", "agent"),
					resource.TestCheckResourceAttr(resourceID, "tools.#", "1"),
					resource.TestCheckResourceAttr(resourceID, "tools.0", "platform.core.index_explorer"),
					resource.TestCheckResourceAttrSet(resourceID, "instructions"),
				),
			},
			{
				SkipFunc:          versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				ResourceName:      resourceID,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				Config: providerConfig + `
resource "elasticstack_kibana_ab_agent" "test" {
	id          = "` + agentID + `"
	name        = "Updated Test Agent"
	description = "An updated test agent"
	labels      = ["test", "agent", "updated"]
	instructions = <<-EOT
You are an updated helpful assistant.
Use the available tools wisely.
EOT
	tools = ["platform.core.index_explorer", "platform.core.list_indices"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "id", agentID),
					resource.TestCheckResourceAttr(resourceID, "name", "Updated Test Agent"),
					resource.TestCheckResourceAttr(resourceID, "description", "An updated test agent"),
					resource.TestCheckResourceAttr(resourceID, "labels.#", "3"),
					resource.TestCheckResourceAttr(resourceID, "tools.#", "2"),
					resource.TestCheckResourceAttr(resourceID, "tools.0", "platform.core.index_explorer"),
					resource.TestCheckResourceAttr(resourceID, "tools.1", "platform.core.list_indices"),
				),
			},
		},
	})
}

func TestAccResourceAgentNoTools(t *testing.T) {
	agentID := "test-no-tools-agent-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceID := "elasticstack_kibana_ab_agent.test_no_tools"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckWithWorkflowsEnabled(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				Config: providerConfig + `
resource "elasticstack_kibana_ab_agent" "test_no_tools" {
	id          = "` + agentID + `"
	name        = "No Tools Agent"
	description = "An agent without tools"
	instructions = "You are a simple assistant without tools."
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "id", agentID),
					resource.TestCheckResourceAttr(resourceID, "name", "No Tools Agent"),
					resource.TestCheckNoResourceAttr(resourceID, "tools"),
				),
			},
		},
	})
}
