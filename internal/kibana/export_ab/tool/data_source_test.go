package tool_test

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

// TestAccDataSourceKibanaExportABTool tests the export_ab_tool data source
func TestAccDataSourceKibanaExportABTool(t *testing.T) {
	toolID := "test-tool-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

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

resource "elasticstack_kibana_ab_tool" "test" {
	id = "%s"
	type = "esql"
	description = "Test ESQL tool"
	tags = ["test"]
	configuration = jsonencode({
		query = "FROM logs | LIMIT 10"
		params = {}
	})
}

data "elasticstack_kibana_export_ab_tool" "test" {
	id = elasticstack_kibana_ab_tool.test.id
}
`, toolID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_export_ab_tool.test", "id"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_export_ab_tool.test", "tool_id"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_export_ab_tool.test", "type"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_export_ab_tool.test", "configuration"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_export_ab_tool.test", "tool_id", toolID),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_export_ab_tool.test", "type", "esql"),
				),
			},
		},
	})
}

// TestAccDataSourceKibanaExportABToolWorkflow tests exporting a workflow-type tool
func TestAccDataSourceKibanaExportABToolWorkflow(t *testing.T) {
	toolID := "test-workflow-tool-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

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
name: Test Workflow
description: A test workflow for tool export
enabled: true
triggers:
  - type: manual
inputs:
  - name: data
    type: string
    default: "test"
steps:
  - name: process_step
    type: console
    with:
      message: "{{ inputs.data }}"
EOT
}

resource "elasticstack_kibana_ab_tool" "test" {
	id = "%s"
	type = "workflow"
	description = "Workflow tool"
	configuration = jsonencode({
		workflow_id = elasticstack_kibana_ab_workflow.test.id
	})
}

data "elasticstack_kibana_export_ab_tool" "test" {
	id = elasticstack_kibana_ab_tool.test.id
}
`, toolID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_export_ab_tool.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_export_ab_tool.test", "type", "workflow"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_export_ab_tool.test", "configuration"),
				),
			},
		},
	})
}
