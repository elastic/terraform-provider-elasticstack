package agent_test

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

// TestAccDataSourceKibanaExportABAgent tests the export_ab_agent data source (agent only)
func TestAccDataSourceKibanaExportABAgent(t *testing.T) {
	agentID := "test-agent-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckWithWorkflowsEnabled(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				Config: fmt.Sprintf(`
provider "elasticstack" {
	kibana {}
}

resource "elasticstack_kibana_ab_agent" "test" {
	id = "%s"
	name = "Test Agent"
	description = "A test agent for export"
	instructions = <<-EOT
You are a helpful assistant that searches logs.
Use the available tools to help answer questions.
EOT
}

data "elasticstack_kibana_export_ab_agent" "test" {
	id = elasticstack_kibana_ab_agent.test.id
	include_dependencies = false
}
`, agentID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_export_ab_agent.test", "id"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_export_ab_agent.test", "agent"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_export_ab_agent.test", "include_dependencies", "false"),
				),
			},
		},
	})
}

// TestAccDataSourceKibanaExportABAgentWithDependencies tests exporting an agent with its tools and workflows
func TestAccDataSourceKibanaExportABAgentWithDependencies(t *testing.T) {
	agentID := "test-agent-deps-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	esqlToolID := "test-esql-tool-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	workflowToolID := "test-workflow-tool-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckWithWorkflowsEnabled(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				Config: fmt.Sprintf(`
provider "elasticstack" {
	kibana {}
}

resource "elasticstack_kibana_ab_workflow" "test" {
	configuration = <<-EOT
name: New workflow
description: This is a new workflow
enabled: true
triggers:
  - type: manual
inputs:
  - name: message
    type: string
    default: "hello world"
steps:
  - name: hello_world_step
    type: console
    with:
      message: "{{ inputs.message }}"
EOT
}

resource "elasticstack_kibana_ab_tool" "workflow" {
	id = "%s"
	type = "workflow"
	description = "Workflow tool"
	configuration = jsonencode({
		workflow_id = elasticstack_kibana_ab_workflow.test.id
	})
}

resource "elasticstack_kibana_ab_tool" "esql" {
	id = "%s"
	type = "esql"
	description = "Test ES|QL tool"
	tags = ["test", "esql"]
	configuration = jsonencode({
		query = "FROM logs-* | LIMIT ?limit"
		params = {
			limit = {
				type = "integer"
				description = "Maximum number of results to return"
			}
		}
	})
}

resource "elasticstack_kibana_ab_agent" "test" {
	id = "%s"
	name = "Test Agent"
	description = "Agent with tools"
	instructions = "Test instructions"
	tools = [elasticstack_kibana_ab_tool.esql.id, elasticstack_kibana_ab_tool.workflow.id]
}

data "elasticstack_kibana_export_ab_agent" "test" {
	id = elasticstack_kibana_ab_agent.test.id
	include_dependencies = true
}
`, workflowToolID, esqlToolID, agentID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_export_ab_agent.test", "id"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_export_ab_agent.test", "agent"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_export_ab_agent.test", "include_dependencies", "true"),
					// Check that tools are populated
					resource.TestCheckResourceAttr("data.elasticstack_kibana_export_ab_agent.test", "tools.#", "2"),
					// Check that workflows are populated
					resource.TestCheckResourceAttr("data.elasticstack_kibana_export_ab_agent.test", "workflows.#", "1"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_export_ab_agent.test", "workflows.0.yaml"),
				),
			},
		},
	})
}
