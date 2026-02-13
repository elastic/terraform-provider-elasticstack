package ab_agent_test

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

var (
	minKibanaAgentBuilderAPIVersion = version.Must(version.NewVersion("9.3.0"))
)

func preCheckWithWorkflowsEnabled(t *testing.T) {
	acctest.PreCheck(t)

	// Enable workflows via Kibana API
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		t.Fatalf("Failed to create API client: %v", err)
	}

	kibanaClient, err := client.GetKibanaOapiClient()
	if err != nil {
		t.Fatalf("Failed to get Kibana client: %v", err)
	}

	// Try the internal settings API endpoint
	settingsURL := fmt.Sprintf("%s/internal/kibana/settings/workflows:ui:enabled", kibanaClient.URL)
	body := map[string]interface{}{
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
	resourceId := "elasticstack_kibana_ab_agent.test"

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
					resource.TestCheckResourceAttr(resourceId, "id", agentID),
					resource.TestCheckResourceAttr(resourceId, "name", "Test Agent"),
					resource.TestCheckResourceAttr(resourceId, "description", "A test agent for acceptance testing"),
					resource.TestCheckResourceAttr(resourceId, "labels.#", "2"),
					resource.TestCheckResourceAttr(resourceId, "labels.0", "test"),
					resource.TestCheckResourceAttr(resourceId, "labels.1", "agent"),
					resource.TestCheckResourceAttr(resourceId, "tools.#", "1"),
					resource.TestCheckResourceAttr(resourceId, "tools.0", "platform.core.index_explorer"),
					resource.TestCheckResourceAttrSet(resourceId, "instructions"),
				),
			},
			{
				SkipFunc:          versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				ResourceName:      resourceId,
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
					resource.TestCheckResourceAttr(resourceId, "id", agentID),
					resource.TestCheckResourceAttr(resourceId, "name", "Updated Test Agent"),
					resource.TestCheckResourceAttr(resourceId, "description", "An updated test agent"),
					resource.TestCheckResourceAttr(resourceId, "labels.#", "3"),
					resource.TestCheckResourceAttr(resourceId, "tools.#", "2"),
					resource.TestCheckResourceAttr(resourceId, "tools.0", "platform.core.index_explorer"),
					resource.TestCheckResourceAttr(resourceId, "tools.1", "platform.core.list_indices"),
				),
			},
		},
	})
}

func TestAccResourceAgentNoTools(t *testing.T) {
	agentID := "test-no-tools-agent-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceId := "elasticstack_kibana_ab_agent.test_no_tools"

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
					resource.TestCheckResourceAttr(resourceId, "id", agentID),
					resource.TestCheckResourceAttr(resourceId, "name", "No Tools Agent"),
					resource.TestCheckNoResourceAttr(resourceId, "tools"),
				),
			},
		},
	})
}
